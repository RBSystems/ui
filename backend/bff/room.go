package bff

func (c *Client) GetRoom() Room {
	room := Room{
		ID:                   ID(c.roomID),
		Name:                 c.room.Name,
		ControlGroups:        make(map[string]ControlGroup),
		SelectedControlGroup: ID(c.selectedControlGroupID),
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
			outputIcon := "tv"

			for _, IOconfig := range c.uiConfig.OutputConfiguration {
				if config.Name != IOconfig.Name {
					continue
				}

				outputIcon = IOconfig.Icon
			}

			// figure out what the current input for this display is
			// we are assuming that input is roomid - input name
			// unless it's blanked, then the "input" is blank
			curInput := c.roomID + "-" + state.Input
			if state.Blanked != nil && *state.Blanked {
				curInput = "blank"
			}

			d := Display{
				ID:    ID(config.ID),
				Input: ID(curInput),
			}

			// TODO outputs when we do sharing
			d.Outputs = append(d.Outputs, IconPair{
				ID:   ID(config.ID),
				Name: config.DisplayName,
				Icon: Icon{outputIcon},
			})

			cg.Displays = append(cg.Displays, d)
		}

		// add a blank input as the first input
		cg.Inputs = append(cg.Inputs, Input{
			ID: ID("blank"),
			IconPair: IconPair{
				Name: "Blank",
				Icon: Icon{"crop_landscape"},
			},
			Disabled: false,
		})

		for _, name := range preset.Inputs {
			config := GetDeviceConfigByName(c.room.Devices, name)
			inputIcon := "settings_input_hdmi"

			for _, IOconfig := range c.uiConfig.InputConfiguration {
				if config.Name != IOconfig.Name {
					continue
				}

				inputIcon = IOconfig.Icon
			}

			i := Input{
				ID: ID(config.ID),
				IconPair: IconPair{
					Name: config.DisplayName,
					Icon: Icon{inputIcon},
				},
				Disabled: false, // TODO look at the current displays reachable inputs to determine
			}

			// TODO subinputs

			cg.Inputs = append(cg.Inputs, i)
		}

		if len(preset.AudioGroups) > 0 {
			for group, audioDevices := range preset.AudioGroups {
				ag := AudioGroup{
					ID:    ID(group),
					Name:  group,
					Muted: true,
				}

				for _, name := range audioDevices {
					config := GetDeviceConfigByName(c.room.Devices, name)
					state := GetAudioDeviceStateByName(c.state.AudioDevices, name)
					audioIcon := "mic"

					for _, IOconfig := range c.uiConfig.OutputConfiguration {
						if config.Name != IOconfig.Name {
							continue
						}

						audioIcon = IOconfig.Icon
					}

					ad := AudioDevice{
						ID: ID(config.ID),
						IconPair: IconPair{
							Name: config.DisplayName,
							Icon: Icon{audioIcon},
						},
					}

					if state.Volume != nil {
						ad.Level = *state.Volume
					}

					if state.Muted != nil {
						ad.Muted = *state.Muted
					}

					if !ad.Muted {
						ag.Muted = false
					}

					ag.AudioDevices = append(ag.AudioDevices, ad)
				}

				cg.AudioGroups = append(cg.AudioGroups, ag)
			}
		} else {
			// create the displaysAG
			if len(preset.AudioDevices) >= 1 {
				ag := AudioGroup{
					ID:    "displaysAG",
					Name:  "Display Volume Mixing",
					Muted: true,
				}

				for _, name := range preset.AudioDevices {
					config := GetDeviceConfigByName(c.room.Devices, name)
					state := GetAudioDeviceStateByName(c.state.AudioDevices, name)

					ad := AudioDevice{
						ID: ID(config.ID),
						IconPair: IconPair{
							Name: config.DisplayName,
							Icon: Icon{"tv"},
						},
					}

					if state.Volume != nil {
						ad.Level = *state.Volume
					}

					if state.Muted != nil {
						ad.Muted = *state.Muted
					}

					if !ad.Muted {
						ag.Muted = false
					}

					ag.AudioDevices = append(ag.AudioDevices, ad)
				}

				cg.AudioGroups = append(cg.AudioGroups, ag)
			}

			// create the micsAG
			if len(preset.IndependentAudioDevices) >= 1 {

				ag := AudioGroup{
					ID:    "micsAG",
					Name:  "Microphones",
					Muted: true,
				}

				for _, name := range preset.IndependentAudioDevices {
					config := GetDeviceConfigByName(c.room.Devices, name)
					state := GetAudioDeviceStateByName(c.state.AudioDevices, name)

					ad := AudioDevice{
						ID: ID(config.ID),
						IconPair: IconPair{
							Name: config.DisplayName,
							Icon: Icon{"mic"},
						},
					}

					if state.Volume != nil {
						ad.Level = *state.Volume
					}

					if state.Muted != nil {
						ad.Muted = *state.Muted
					}

					if !ad.Muted {
						ag.Muted = false
					}

					ag.AudioDevices = append(ag.AudioDevices, ad)
				}

				cg.AudioGroups = append(cg.AudioGroups, ag)
			}

		}

		room.ControlGroups[string(cg.ID)] = cg
		// TODO PresentGroups
	}

	return room
}