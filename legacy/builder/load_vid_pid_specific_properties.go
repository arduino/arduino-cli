// This file is part of arduino-cli.
//
// Copyright 2020 ARDUINO SA (http://www.arduino.cc/)
//
// This software is released under the GNU General Public License version 3,
// which covers the main part of arduino-cli.
// The terms of this license can be found at:
// https://www.gnu.org/licenses/gpl-3.0.en.html
//
// You can be released from the requirements of the above licenses by purchasing
// a commercial license. Buying such a license is mandatory if you want to
// modify or otherwise use the software for commercial activities involving the
// Arduino software without disclosing the source code of your own applications.
// To purchase a commercial license, send an email to license@arduino.cc.

package builder

import (
	"strconv"
	"strings"

	"github.com/arduino/arduino-cli/legacy/builder/constants"
	"github.com/arduino/arduino-cli/legacy/builder/types"
	"github.com/arduino/go-properties-orderedmap"
	"github.com/pkg/errors"
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
		return errors.WithStack(err)
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
