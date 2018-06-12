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

// GetBuildProperties returns the build properties and the build
// platform for the Board with the configuration passed as parameter.
func (b *Board) GetBuildProperties(configs properties.Map) (properties.Map, error) {
	// Start with board's base properties
	buildProperties := b.Properties.Clone()

	// Add all sub-configurations one by one
	menu := b.Properties.SubTree("menu")
	for option, value := range configs {
		if option == "" {
			return nil, fmt.Errorf("invalid empty option found")
		}

		optionMenu := menu.SubTree(option)
		if len(optionMenu) == 0 {
			return nil, fmt.Errorf("invalid option '%s'", option)
		}
		if _, ok := optionMenu[value]; !ok {
			return nil, fmt.Errorf("invalid value '%s' for option '%s'", value, option)
		}

		optionsConf := optionMenu.SubTree(value)
		buildProperties.Merge(optionsConf)
	}

	return buildProperties, nil
}

// GeneratePropertiesForConfiguration returns the board properties for a particular
// configuration. The parameter is the latest part of the FQBN, for example if
// the full FQBN is "arduino:avr:mega:cpu=atmega2560" the config part must be
// "cpu=atmega2560".
// FIXME: deprecated, use GetBuildProperties instead
func (b *Board) GeneratePropertiesForConfiguration(config string) (properties.Map, error) {
	fqbn, err := ParseFQBN(b.String() + ":" + config)
	if err != nil {
		return nil, fmt.Errorf("parsing fqbn: %s", err)
	}
	return b.GetBuildProperties(fqbn.Configs)
}
