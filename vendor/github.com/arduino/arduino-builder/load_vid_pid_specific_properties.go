/*
 * This file is part of Arduino Builder.
 *
 * Arduino Builder is free software; you can redistribute it and/or modify
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
 * Copyright 2015 Arduino LLC (http://www.arduino.cc/)
 */

package builder

import (
	"strconv"
	"strings"

	"github.com/arduino/arduino-builder/constants"
	"github.com/arduino/arduino-builder/i18n"
	"github.com/arduino/arduino-builder/types"
	"github.com/arduino/go-properties-orderedmap"
)

type LoadVIDPIDSpecificProperties struct{}

func (s *LoadVIDPIDSpecificProperties) Run(ctx *types.Context) error {
	if ctx.USBVidPid == "" {
		return nil
	}

	vidPid := ctx.USBVidPid
	vidPid = strings.ToLower(vidPid)
	vidPidParts := strings.Split(vidPid, "_")
	vid := vidPidParts[0]
	pid := vidPidParts[1]

	buildProperties := ctx.BuildProperties
	VIDPIDIndex, err := findVIDPIDIndex(buildProperties, vid, pid)
	if err != nil {
		return i18n.WrapError(err)
	}
	if VIDPIDIndex < 0 {
		return nil
	}

	vidPidSpecificProperties := buildProperties.SubTree(constants.BUILD_PROPERTIES_VID).SubTree(strconv.Itoa(VIDPIDIndex))
	buildProperties.Merge(vidPidSpecificProperties)

	return nil
}

func findVIDPIDIndex(buildProperties *properties.Map, vid, pid string) (int, error) {
	for key, value := range buildProperties.SubTree(constants.BUILD_PROPERTIES_VID).AsMap() {
		if !strings.Contains(key, ".") {
			if vid == strings.ToLower(value) && pid == strings.ToLower(buildProperties.Get(constants.BUILD_PROPERTIES_PID+"."+key)) {
				return strconv.Atoi(key)
			}
		}
	}

	return -1, nil
}
