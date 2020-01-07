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

package core

import (
	"errors"
	"regexp"
	"strings"

	"github.com/arduino/arduino-cli/arduino/cores"
	"github.com/arduino/arduino-cli/commands"
	rpc "github.com/arduino/arduino-cli/rpc/commands"
)

func match(line, searchArgs string) bool {
	return strings.Contains(strings.ToLower(line), searchArgs)
}

// PlatformSearch FIXMEDOC
func PlatformSearch(instanceID int32, searchArgs string, allVersions bool) (*rpc.PlatformSearchResp, error) {
	pm := commands.GetPackageManager(instanceID)
	if pm == nil {
		return nil, errors.New("invalid instance")
	}

	res := []*cores.PlatformRelease{}
	if isUsb, _ := regexp.MatchString("[0-9a-f]{4}:[0-9a-f]{4}", searchArgs); isUsb {
		vid, pid := searchArgs[:4], searchArgs[5:]
		res = pm.FindPlatformReleaseProvidingBoardsWithVidPid(vid, pid)
	} else {
		for _, targetPackage := range pm.Packages {
			for _, platform := range targetPackage.Platforms {
				platformRelease := platform.GetLatestRelease()
				if platformRelease == nil {
					continue
				}

				// platform has a release, check if it matches the search arguments
				if match(platform.Name, searchArgs) || match(platform.Architecture, searchArgs) {
					if allVersions {
						res = append(res, platform.GetAllReleases()...)
					} else {
						res = append(res, platformRelease)
					}
				} else {
					// if we didn't find a match in the platform data, search for
					// a match in the boards manifest
					for _, board := range platformRelease.BoardsManifest {
						if match(board.Name, searchArgs) {
							if allVersions {
								res = append(res, platform.GetAllReleases()...)
							} else {
								res = append(res, platformRelease)
							}
							break
						}
					}
				}
			}
		}
	}

	out := make([]*rpc.Platform, len(res))
	for i, platformRelease := range res {
		out[i] = PlatformReleaseToRPC(platformRelease)
	}
	return &rpc.PlatformSearchResp{SearchOutput: out}, nil
}
