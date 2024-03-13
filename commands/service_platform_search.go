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

package commands

import (
	"regexp"
	"sort"
	"strings"

	"github.com/arduino/arduino-cli/commands/internal/instances"
	"github.com/arduino/arduino-cli/internal/arduino/cores"
	"github.com/arduino/arduino-cli/internal/arduino/utils"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
)

// PlatformSearch FIXMEDOC
func PlatformSearch(req *rpc.PlatformSearchRequest) (*rpc.PlatformSearchResponse, error) {
	pme, release, err := instances.GetPackageManagerExplorer(req.GetInstance())
	if err != nil {
		return nil, err
	}
	defer release()

	res := []*cores.Platform{}
	if isUsb, _ := regexp.MatchString("[0-9a-f]{4}:[0-9a-f]{4}", req.GetSearchArgs()); isUsb {
		vid, pid := req.GetSearchArgs()[:4], req.GetSearchArgs()[5:]
		res = pme.FindPlatformReleaseProvidingBoardsWithVidPid(vid, pid)
	} else {
		searchArgs := utils.SearchTermsFromQueryString(req.GetSearchArgs())
		for _, targetPackage := range pme.GetPackages() {
			for _, platform := range targetPackage.Platforms {
				if platform == nil {
					continue
				}
				// Users can install platforms manually in the Sketchbook hardware folder,
				// if not explictily requested we skip them.
				if !req.GetManuallyInstalled() && platform.ManuallyInstalled {
					continue
				}

				// Discard platforms with no releases
				latestRelease := platform.GetLatestRelease()
				if latestRelease == nil || latestRelease.Name == "" {
					continue
				}

				// Gather all strings that can be used for searching
				toTest := platform.String() + " " +
					latestRelease.Name + " " +
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

				res = append(res, platform)
			}
		}
	}

	out := []*rpc.PlatformSummary{}
	for _, platform := range res {
		rpcPlatformSummary := &rpc.PlatformSummary{
			Metadata: PlatformToRPCPlatformMetadata(platform),
			Releases: map[string]*rpc.PlatformRelease{},
		}
		if installed := pme.GetInstalledPlatformRelease(platform); installed != nil {
			rpcPlatformSummary.InstalledVersion = installed.Version.String()
		}
		if latestCompatible := platform.GetLatestCompatibleRelease(); latestCompatible != nil {
			rpcPlatformSummary.LatestVersion = latestCompatible.Version.String()
		}
		for _, platformRelease := range platform.GetAllReleases() {
			rpcPlatformRelease := PlatformReleaseToRPC(platformRelease)
			rpcPlatformSummary.Releases[rpcPlatformRelease.GetVersion()] = rpcPlatformRelease
		}
		out = append(out, rpcPlatformSummary)
	}

	// Sort result alphabetically and put deprecated platforms at the bottom
	sort.Slice(out, func(i, j int) bool {
		return strings.ToLower(out[i].GetMetadata().GetId()) < strings.ToLower(out[j].GetMetadata().GetId())
	})
	sort.SliceStable(out, func(i, j int) bool {
		return !out[i].GetMetadata().GetDeprecated() && out[j].GetMetadata().GetDeprecated()
	})
	return &rpc.PlatformSearchResponse{SearchOutput: out}, nil
}
