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

package board

import (
	"context"
	"errors"
	"strings"

	"github.com/arduino/arduino-cli/commands"
	rpc "github.com/arduino/arduino-cli/rpc/commands"
	"github.com/lithammer/fuzzysearch/fuzzy"
)

// maximumSearchDistance is the maximum Levenshtein distance accepted when using fuzzy search.
// This value is completely arbitrary and picked randomly.
const maximumSearchDistance = 20

// ListAll FIXMEDOC
func ListAll(ctx context.Context, req *rpc.BoardListAllReq) (*rpc.BoardListAllResp, error) {
	pm := commands.GetPackageManager(req.GetInstance().GetId())
	if pm == nil {
		return nil, errors.New("invalid instance")
	}

	searchArgs := strings.Join(req.SearchArgs, " ")

	match := func(toTest []string) bool {
		if len(searchArgs) == 0 {
			return true
		}
		for _, rank := range fuzzy.RankFindNormalizedFold(searchArgs, toTest) {
			if rank.Distance < maximumSearchDistance {
				return true
			}
		}
		return false
	}

	list := &rpc.BoardListAllResp{Boards: []*rpc.BoardListItem{}}
	for _, targetPackage := range pm.Packages {
		for _, platform := range targetPackage.Platforms {
			installedPlatformRelease := pm.GetInstalledPlatformRelease(platform)
			// We only want to list boards for installed platforms
			if installedPlatformRelease == nil {
				continue
			}

			installedVersion := installedPlatformRelease.Version.String()

			latestVersion := ""
			if latestPlatformRelease := platform.GetLatestRelease(); latestPlatformRelease != nil {
				latestVersion = latestPlatformRelease.Version.String()
			}

			rpcPlatform := &rpc.Platform{
				ID:                platform.String(),
				Installed:         installedVersion,
				Latest:            latestVersion,
				Name:              platform.Name,
				Maintainer:        platform.Package.Maintainer,
				Website:           platform.Package.WebsiteURL,
				Email:             platform.Package.Email,
				ManuallyInstalled: platform.ManuallyInstalled,
			}

			toTest := []string{
				platform.String(),
				platform.Name,
				platform.Architecture,
				targetPackage.Name,
				targetPackage.Maintainer,
			}

			for _, board := range installedPlatformRelease.Boards {
				if !req.GetIncludeHiddenBoards() && board.IsHidden() {
					continue
				}

				toTest := toTest
				toTest = append(toTest, strings.Split(board.Name(), " ")...)
				toTest = append(toTest, board.FQBN())
				if !match(toTest) {
					continue
				}

				list.Boards = append(list.Boards, &rpc.BoardListItem{
					Name:     board.Name(),
					FQBN:     board.FQBN(),
					IsHidden: board.IsHidden(),
					Platform: rpcPlatform,
				})
			}
		}
	}

	return list, nil
}
