package bff

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/byuoitav/common/structs"
	"github.com/byuoitav/ui/log"
	"go.uber.org/zap"
)

// TODO send function
type Client struct {
	buildingID string
	roomID     string

	room     structs.Room
	state    structs.PublicRoom
	uiConfig structs.UIConfig

	httpClient *http.Client

	Out chan []byte

	*zap.Logger
}

func RegisterClient(ctx context.Context, roomID, controlGroupID, name string) (*Client, error) {
	log.P.Info("Registering client", zap.String("roomID", roomID), zap.String("controlGroupID", controlGroupID), zap.String("name", name))

	split := strings.Split(roomID, "-")
	if len(split) != 2 {
		return nil, fmt.Errorf("invalid roomID %q - must match format BLDG-ROOM", roomID)
	}

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	c := &Client{
		buildingID: split[0],
		roomID:     roomID,
		httpClient: &http.Client{},
		Out:        make(chan []byte, 8),
		Logger:     log.P.Named(name),
	}

	errCh := make(chan error, 3)
	doneCh := make(chan struct{})

	wg := sync.WaitGroup{}
	wg.Add(3)

	go func() {
		defer close(doneCh)
		wg.Wait()
	}()

	go func() {
		var err error
		defer wg.Done()

		c.state, err = GetRoomState(ctx, c.httpClient, c.roomID)
		if err != nil {
			c.Warn("unable to get room state", zap.Error(err))
			errCh <- fmt.Errorf("unable to get ui config: %v", err)
		}

		c.Debug("Successfully got room state")
	}()

	go func() {
		var err error
		defer wg.Done()

		c.room, err = GetRoomConfig(ctx, c.httpClient, c.roomID)
		if err != nil {
			c.Warn("unable to get room config", zap.Error(err))
			errCh <- fmt.Errorf("unable to get room config: %v", err)
		}

		c.Debug("Successfully got room config")
	}()

	go func() {
		var err error
		defer wg.Done()

		c.uiConfig, err = GetUIConfig(ctx, c.httpClient, c.roomID)
		if err != nil {
			c.Warn("unable to get ui config", zap.Error(err))
			errCh <- fmt.Errorf("unable to get ui config: %v", err)
		}

		c.Debug("Successfully got ui config")
	}()

	select {
	case err := <-errCh:
		return nil, fmt.Errorf("unable to get room information: %v", err)
	case <-ctx.Done():
		return nil, fmt.Errorf("unable to get room information: all requests timed out")
	case <-doneCh:
	}

	c.Info("Got all initial information, sending room to client")

	// write the inital room info
	room := c.GetRoom()
	buf, err := json.Marshal(room)
	if err != nil {
		return nil, fmt.Errorf("unable to marshal room: %s", err)
	}

	c.Out <- buf
	return c, nil
}

func (c *Client) GetRoom() Room {
	room := Room{
		ID:                   ID(c.roomID),
		Name:                 c.room.Name,
		ControlGroups:        make(map[string]ControlGroup),
		SelectedControlGroup: "", // TODO where is this saved? c?
	}

	for _, preset := range c.uiConfig.Presets {
		cg := ControlGroup{
			ID:   ID(preset.Name),
			Name: preset.Name,
			Support: Support{
				HelpRequested: false, // This info also needs to be saved...
				HelpMessage:   "Request Help",
				HelpEnabled:   true,
			},
		}

		for _, name := range preset.Displays {
			config := GetDeviceConfigByName(c.room.Devices, name)
			state := GetDisplayStateByName(c.state.Displays, name)

			d := Display{
				ID:    ID(config.ID),
				Input: ID(state.Input),
			}

			// TODO outputs when we do sharing
			d.Outputs = append(d.Outputs, IconPair{
				Name: config.Name,
				Icon: Icon{"tv"}, // TODO get this from the ui config
			})

			if state.Blanked != nil {
				d.Blanked = *state.Blanked
			}

			cg.Displays = append(cg.Displays, d)
		}

		// add a blank input as the first input
		cg.Inputs = append(cg.Inputs, Input{
			ID: ID("blank"),
			IconPair: IconPair{
				Name: "Blank",
				Icon: Icon{"blank_icon"},
			},
			Disabled: false,
		})

		for _, name := range preset.Inputs {
			config := GetDeviceConfigByName(c.room.Devices, name)

			i := Input{
				ID: ID(config.ID),
				IconPair: IconPair{
					Name: config.DisplayName,
					Icon: Icon{"settings_input_hdmi"},
				},
				Disabled: false, // TODO look at the current displays reachable inputs to determine
			}

			// TODO subinputs

			cg.Inputs = append(cg.Inputs, i)
		}

		for group, audioDevices := range preset.AudioGroups {
			ag := AudioGroup{
				ID:    ID(group),
				Name:  group,
				Muted: true,
			}

			for _, name := range audioDevices {
				config := GetDeviceConfigByName(c.room.Devices, name)
				state := GetAudioDeviceStateByName(c.state.AudioDevices, name)

				ad := AudioDevice{
					ID: ID(config.ID),
					IconPair: IconPair{
						Name: config.DisplayName,
						Icon: Icon{"mic"}, // TODO
					},
				}

				if state.Volume != nil {
					ad.Level = *state.Volume
				}

				if state.Muted != nil {
					ad.Muted = *state.Muted
				}

				// set all muted if one of them isn't
				if !ad.Muted {
					ag.Muted = false
				}

				ag.AudioDevices = append(ag.AudioDevices, ad)
			}

			cg.AudioGroups = append(cg.AudioGroups, ag)
		}

		room.ControlGroups[string(cg.ID)] = cg
		// TODO PresentGroups
	}

	return room
}
