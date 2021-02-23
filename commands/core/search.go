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
	"github.com/lithammer/fuzzysearch/fuzzy"
)

// PlatformSearch FIXMEDOC
func PlatformSearch(req *rpc.PlatformSearchReq) (*rpc.PlatformSearchResp, error) {
	searchArgs := strings.Trim(req.SearchArgs, " ")
	allVersions := req.AllVersions
	pm := commands.GetPackageManager(req.Instance.Id)
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
				// discard invalid platforms
				// Users can install platforms manually in the Sketchbook hardware folder,
				// the core search command must operate only on platforms installed through
				// the PlatformManager, thus we skip the manually installed ones.
				if platform == nil || platform.Name == "" || platform.ManuallyInstalled {
					continue
				}

				// discard invalid releases
				platformRelease := platform.GetLatestRelease()
				if platformRelease == nil {
					continue
				}

				if searchArgs == "" {
					if allVersions {
						res = append(res, platform.GetAllReleases()...)
					} else {
						res = append(res, platformRelease)
					}
					continue
				}

				words := strings.Split(searchArgs, " ")
				toTest := []string{
					platform.String(),
					platform.Name,
					platform.Architecture,
					targetPackage.Name,
					targetPackage.Maintainer,
					targetPackage.WebsiteURL,
				}
				for _, board := range platformRelease.BoardsManifest {
					toTest = append(toTest, board.Name)
				}
				for _, word := range words {
					if len(fuzzy.FindNormalizedFold(word, toTest)) > 0 {
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

	out := make([]*rpc.Platform, len(res))
	for i, platformRelease := range res {
		out[i] = commands.PlatformReleaseToRPC(platformRelease)
	}
	return &rpc.PlatformSearchResp{SearchOutput: out}, nil
}
