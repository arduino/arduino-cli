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
	"regexp"
	"sort"
	"strings"

	"github.com/arduino/arduino-cli/arduino"
	"github.com/arduino/arduino-cli/arduino/cores"
	"github.com/arduino/arduino-cli/arduino/utils"
	"github.com/arduino/arduino-cli/commands"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
)

// PlatformSearch FIXMEDOC
func PlatformSearch(req *rpc.PlatformSearchRequest) (*rpc.PlatformSearchResponse, error) {
	pme, release := commands.GetPackageManagerExplorer(req)
	if pme == nil {
		return nil, &arduino.InvalidInstanceError{}
	}
	defer release()

	res := []*cores.PlatformRelease{}
	if isUsb, _ := regexp.MatchString("[0-9a-f]{4}:[0-9a-f]{4}", req.SearchArgs); isUsb {
		vid, pid := req.SearchArgs[:4], req.SearchArgs[5:]
		res = pme.FindPlatformReleaseProvidingBoardsWithVidPid(vid, pid)
	} else {
		searchArgs := utils.SearchTermsFromQueryString(req.SearchArgs)
		allVersions := req.AllVersions
		for _, targetPackage := range pme.GetPackages() {
			for _, platform := range targetPackage.Platforms {
				// Users can install platforms manually in the Sketchbook hardware folder,
				// the core search command must operate only on platforms installed through
				// the PlatformManager, thus we skip the manually installed ones.
				if platform == nil || platform.Name == "" || platform.ManuallyInstalled {
					continue
				}

				// Discard platforms with no releases
				latestRelease := platform.GetLatestRelease()
				if latestRelease == nil {
					continue
				}

				// Gather all strings that can be used for searching
				toTest := platform.String() + " " +
					platform.Name + " " +
					platform.Architecture + " " +
					targetPackage.Name + " " +
					targetPackage.Maintainer + " " +
					targetPackage.WebsiteURL
				for _, board := range latestRelease.BoardsManifest {
					toTest += board.Name + " "
				}

				// Search
				if !utils.Match(toTest, searchArgs) {
					continue
				}

				if allVersions {
					res = append(res, platform.GetAllReleases()...)
				} else {
					res = append(res, latestRelease)
				}
			}
		}
	}

	out := make([]*rpc.Platform, len(res))
	for i, platformRelease := range res {
		out[i] = commands.PlatformReleaseToRPC(platformRelease)
	}
	// Sort result alphabetically and put deprecated platforms at the bottom
	sort.Slice(
		out, func(i, j int) bool {
			return strings.ToLower(out[i].Name) < strings.ToLower(out[j].Name)
		})
	sort.SliceStable(
		out, func(i, j int) bool {
			if !out[i].Deprecated && out[j].Deprecated {
				return true
			}
			return false
		})
	return &rpc.PlatformSearchResponse{SearchOutput: out}, nil
}
