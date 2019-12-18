package bff

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	//"github.com/byuoitav/av-api/base"
	"github.com/byuoitav/central-event-system/hub/base"
	"github.com/byuoitav/common/structs"
	"github.com/byuoitav/device-monitoring/messenger"
	"go.uber.org/zap"
)

func (c *Client) HandleEvents() {
	mess, err := messenger.BuildMessenger(os.Getenv("HUB_ADDRESS"), base.Messenger, 1)
	if err != nil {
		fmt.Printf("%s", err)
	}

	mess.SubscribeToRooms(c.roomID)

	for {
		event := mess.ReceiveEvent()
		isCoreState := false

		for _, tag := range event.EventTags {
			if tag == "core-state" {
				isCoreState = true
			}
		}

		if !isCoreState {
			continue
		}

		state := c.state
		var changed bool
		var newstate structs.PublicRoom
		switch event.Key {
		case "volume":
			newstate, changed = handleVolume(state, event.Value, event.TargetDevice.DeviceID)
		case "muted":
			newstate, changed = handleMuted(state, event.Value, event.TargetDevice.DeviceID)
		case "power":
			newstate, changed = handlePower(state, event.Value, event.TargetDevice.DeviceID)
		case "input":
			newstate, changed = handleInput(state, event.Value, event.TargetDevice.DeviceID)
		case "blanked":
			newstate, changed = handleBlanked(state, event.Value, event.TargetDevice.DeviceID)
		default:
			continue
		}

		if !changed {
			continue
		}

		c.state = newstate
		msg, err := JSONMessage("room", c.GetRoom())
		if err != nil {
			c.Warn("failed to create JSON message with new room state", zap.Error(err))
		}
		c.Out <- msg

	}

	// TODO close messenger when it's done!
}

func handleVolume(state structs.PublicRoom, volume, targetDevice string) (structs.PublicRoom, bool) {
	intVolume, _ := strconv.Atoi(volume)
	deviceId := getDeviceId(targetDevice)
	changed := false

	for i := range state.AudioDevices {
		if state.AudioDevices[i].Name != deviceId {
			continue
		}

		if state.AudioDevices[i].Volume == &intVolume {
			continue
		}

		state.AudioDevices[i].Volume = &intVolume
		changed = true
	}

	return state, changed
}

func handleMuted(state structs.PublicRoom, muted, targetDevice string) (structs.PublicRoom, bool) {
	isMuted, _ := strconv.ParseBool(muted)
	deviceId := getDeviceId(targetDevice)
	changed := false

	for i := range state.AudioDevices {
		if state.AudioDevices[i].Name != deviceId {
			continue
		}

		if state.AudioDevices[i].Muted == &isMuted {
			continue
		}

		state.AudioDevices[i].Muted = &isMuted
		changed = true
	}

	return state, changed
}
func handlePower(state structs.PublicRoom, power, targetDevice string) (structs.PublicRoom, bool) {
	deviceId := getDeviceId(targetDevice)
	changed := false

	for i := range state.Displays {
		if state.Displays[i].Name != deviceId {
			continue
		}

		if state.Displays[i].Power == power {
			continue
		}

		state.Displays[i].Power = power
		changed = true
	}

	return state, changed
}
func handleInput(state structs.PublicRoom, input, targetDevice string) (structs.PublicRoom, bool) {
	deviceId := getDeviceId(targetDevice)
	changed := false

	for i := range state.Displays {
		if state.Displays[i].Name != deviceId {
			continue
		}

		if state.Displays[i].Input == input {
			continue
		}

		state.Displays[i].Input = input
		changed = true
	}

	return state, changed
}
func handleBlanked(state structs.PublicRoom, blanked, targetDevice string) (structs.PublicRoom, bool) {
	isBlanked, _ := strconv.ParseBool(blanked)
	deviceId := getDeviceId(targetDevice)
	changed := false

	for i := range state.Displays {
		if state.Displays[i].Name != deviceId {
			continue
		}

		if state.Displays[i].Blanked == &isBlanked {
			continue
		}

		state.Displays[i].Blanked = &isBlanked
		changed = true
	}

	return state, changed
}

func getDeviceId(targetDevice string) string {
	splitDevice := strings.Split(targetDevice, "-")
	return splitDevice[2]
}
