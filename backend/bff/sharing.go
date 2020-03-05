package bff

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/byuoitav/common/structs"
	"go.uber.org/zap"
)

type shareState int

const (
	stateCantShare shareState = iota
	stateCanShare
	stateIsMaster
	stateIsActiveMinion
	stateIsInactiveMinion
)

const (
	lazSharing = "-sharing-"
)

// the types of objects that are stored in lazarette for each display group
type lazShareData struct {
	state shareState

	// if they are in a group, this is who that leader is
	master ID
}

func (c *Client) getActiveAndInactiveForDisplayGroup(group ID) ([]ID, []ID) {
	var active []ID
	var inactive []ID

	c.lazs.Range(func(key interface{}, value interface{}) bool {
		skey, ok := key.(string)
		if !ok || !strings.HasPrefix(skey, lazSharing) {
			return true
		}

		shareData, ok := value.(lazShareData)
		if !ok {
			return true
		}

		// make sure it's in this group
		if shareData.master != group {
			return true
		}

		// strip prefix off of id
		id := strings.TrimPrefix(skey, lazSharing)

		// add it to active/inactive list
		switch shareData.state {
		case stateIsActiveMinion:
			active = append(active, ID(id))
		case stateIsInactiveMinion:
			inactive = append(inactive, ID(id))
		}

		return true
	})

	return active, inactive
}

func (c *Client) getShareData(group ID) (lazShareData, error) {
	var data lazShareData

	idata, ok := c.lazs.Load(lazSharing + group)
	if !ok {
		return data, errors.New("no share data found")
	}

	data, ok = idata.(lazShareData)
	if !ok {
		return data, errors.New("unexpected type")
	}

	return data, nil
}

// SetSharing .
type SetSharing struct {
}

// SetSharingMessage is a message enabling or disabling sharing
type SetSharingMessage struct {
	Master  ID       `json:"master"`
	Status  bool     `json:"status"`
	Options []string `json:"options,omitempty"`
}

func (ss SetSharing) Do(c *Client, data []byte) {
	var msg SetSharingMessage
	if err := json.Unmarshal(data, &msg); err != nil {
		c.Out <- ErrorMessage(fmt.Errorf("invalid value for setSharing: %s", err))
		return
	}

	// TODO we may need some kind of lock in lazarette to handle the case of
	// two people starting a share at the exact same time,
	// right now, whoever's request gets put into lazarette second will win

	if msg.Status {
		ss.On(c, msg)
	} else {
		// ss.Off(c, msg)
	}
}

func (ss SetSharing) On(c *Client, msg SetSharingMessage) {
	room := c.GetRoom()
	cg := room.ControlGroups[c.selectedControlGroupID]

	// create a list of all display groups across all control groups
	var dgroups []DisplayGroup
	for _, v := range room.ControlGroups {
		dgroups = append(dgroups, v.DisplayGroups...)
	}

	// get the current input that the master is on
	var input string
	for i := range cg.DisplayGroups {
		if cg.DisplayGroups[i].ID == msg.Master {
			input = cg.DisplayGroups[i].Input.GetName()
		}
	}

	if len(input) == 0 {
		// handle
	}

	// blanked?
	// validate that options are valid
	// validate that master group id is valid
	// validate that inputs are valid for minions

	var state structs.PublicRoom
	for _, minion := range msg.Options {
		c.lazUpdates <- lazMessage{
			Key: lazSharing + minion,
			Data: lazShareData{
				state:  stateIsActiveMinion,
				master: msg.Master,
			},
		}

		mgroup, err := GetDisplayGroupByID(dgroups, ID(minion))
		if err != nil {
			// handle
		}

		for _, disp := range mgroup.Displays {
			state.Displays = append(state.Displays, structs.Display{
				PublicDevice: structs.PublicDevice{
					Name:  disp.ID.GetName(),
					Input: input,
				},
			})
		}
	}

	// update the masters lazarette data
	c.lazUpdates <- lazMessage{
		Key: lazSharing + string(msg.Master),
		Data: lazShareData{
			state: stateIsMaster,
		},
	}

	// send the api request
	if err := c.SendAPIRequest(context.TODO(), state); err != nil {
		c.Warn("failed to set sharing", zap.Error(err))
		c.Out <- ErrorMessage(fmt.Errorf("failed to set sharing: %w", err))
	}
}
