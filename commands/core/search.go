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
	"github.com/arduino/arduino-cli/arduino/utils"
	"github.com/arduino/arduino-cli/commands"
	rpc "github.com/arduino/arduino-cli/rpc/commands"
)

// maximumSearchDistance is the maximum Levenshtein distance accepted when using fuzzy search.
// This value is completely arbitrary and picked randomly.
const maximumSearchDistance = 20

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

		searchArgs := strings.Split(searchArgs, " ")

		match := func(toTest []string) (bool, error) {
			if len(searchArgs) == 0 {
				return true, nil
			}

			for _, t := range toTest {
				matches, err := utils.Match(t, searchArgs)
				if err != nil {
					return false, err
				}
				if matches {
					return matches, nil
				}
			}
			return false, nil
		}

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

				// Gather all strings that can be used for searching
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

				// Search
				if ok, err := match(toTest); err != nil {
					return nil, err
				} else if !ok {
					continue
				}

				if allVersions {
					res = append(res, platform.GetAllReleases()...)
				} else {
					res = append(res, platformRelease)
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
