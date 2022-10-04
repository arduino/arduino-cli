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
	"strings"

	"github.com/arduino/arduino-cli/arduino"
	"github.com/arduino/arduino-cli/arduino/utils"
	"github.com/arduino/arduino-cli/commands"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
)

// ListAll FIXMEDOC
func ListAll(ctx context.Context, req *rpc.BoardListAllRequest) (*rpc.BoardListAllResponse, error) {
	pme, release := commands.GetPackageManagerExplorer(req)
	if pme == nil {
		return nil, &arduino.InvalidInstanceError{}
	}
	defer release()

	searchArgs := strings.Join(req.GetSearchArgs(), " ")

	list := &rpc.BoardListAllResponse{Boards: []*rpc.BoardListItem{}}
	for _, targetPackage := range pme.GetPackages() {
		for _, platform := range targetPackage.Platforms {
			installedPlatformRelease := pme.GetInstalledPlatformRelease(platform)
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
				Id:                platform.String(),
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

			for _, board := range installedPlatformRelease.GetBoards() {
				if !req.GetIncludeHiddenBoards() && board.IsHidden() {
					continue
				}

				toTest := append(toTest, board.Name())
				toTest = append(toTest, board.FQBN())
				if !utils.MatchAny(searchArgs, toTest) {
					continue
				}

				list.Boards = append(list.Boards, &rpc.BoardListItem{
					Name:     board.Name(),
					Fqbn:     board.FQBN(),
					IsHidden: board.IsHidden(),
					Platform: rpcPlatform,
				})
			}
		}
	}

	return list, nil
}
