/*
 * This file is part of arduino-cli.
 *
 * arduino-cli is free software; you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation; either version 2 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program; if not, write to the Free Software
 * Foundation, Inc., 51 Franklin St, Fifth Floor, Boston, MA  02110-1301  USA
 *
 * As a special exception, you may use this file as part of a free software
 * library without restriction.  Specifically, if other files instantiate
 * templates or use macros or inline functions from this file, or you compile
 * this file and link it with other files to produce an executable, this
 * file does not by itself cause the resulting executable to be covered by
 * the GNU General Public License.  This exception does not however
 * invalidate any other reasons why the executable file might be covered by
 * the GNU General Public License.
 *
 * Copyright 2017-2018 ARDUINO AG (http://www.arduino.cc/)
 */

package cores

import (
	"fmt"
	"strings"

	"github.com/arduino/go-properties-map"
)

// Board represents a board loaded from an installed platform
type Board struct {
	BoardId         string
	Properties      properties.Map   `json:"-"`
	PlatformRelease *PlatformRelease `json:"-"`
}

// HasUsbID returns true if the board match the usb vid and pid parameters
func (b *Board) HasUsbID(reqVid, reqPid string) bool {
	vids := b.Properties.SubTree("vid")
	pids := b.Properties.SubTree("pid")
	for id, vid := range vids {
		if pid, ok := pids[id]; ok {
			if strings.EqualFold(vid, reqVid) && strings.EqualFold(pid, reqPid) {
				return true
			}
		}
	}
	return false
}

// Name returns the board name as defined in boards.txt properties
func (b *Board) Name() string {
	return b.Properties["name"]
}

// FQBN return the Fully-Qualified-Board-Name for the default configuration of this board
func (b *Board) FQBN() string {
	platform := b.PlatformRelease.Platform
	return platform.Package.Name + ":" + platform.Architecture + ":" + b.BoardId
}

func (b *Board) String() string {
	return b.FQBN()
}

// GeneratePropertiesForConfiguration returns the board properties for a particular
// configuration. The parameter is the latest part of the FQBN, for example if
// the full FQBN is "arduino:avr:mega:cpu=atmega2560" the config part must be
// "cpu=atmega2560".
func (b *Board) GeneratePropertiesForConfiguration(config string) (properties.Map, error) {
	menu := b.Properties.SubTree("menu")
	if menu == nil || len(menu) == 0 {
		return nil, fmt.Errorf("there are no options available")
	}
	res := b.Properties.Clone()

	options := strings.Split(config, ",")
	for _, option := range options {
		parts := strings.Split(option, "=")
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid option '%s'", option)
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		if key == "" || value == "" {
			return nil, fmt.Errorf("invalid option '%s'", option)
		}

		optionMenu := menu.SubTree(key)
		if optionMenu == nil || len(optionMenu) == 0 {
			return nil, fmt.Errorf("invalid option '%s'", option)
		}

		if _, ok := optionMenu[value]; !ok {
			return nil, fmt.Errorf("invalid option value '%s' for '%s'", value, option)
		}
		optionsConf := optionMenu.SubTree(value)
		if optionsConf == nil || len(optionsConf) == 0 {
			continue
		}

		res.Merge(optionsConf)
	}

	return res, nil
}
