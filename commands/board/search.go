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
	"sort"
	"strings"

	"github.com/arduino/arduino-cli/arduino/utils"
	"github.com/arduino/arduino-cli/commands"
	rpc "github.com/arduino/arduino-cli/rpc/commands"
)

// Search returns all boards that match the search arg.
// Boards are searched in all platforms, including those in the index that are not yet
// installed. Note that platforms that are not installed don't include boards' FQBNs.
// If no search argument is used all boards are returned.
func Search(ctx context.Context, req *rpc.BoardSearchReq) (*rpc.BoardSearchResp, error) {
	pm := commands.GetPackageManager(req.GetInstance().GetId())
	if pm == nil {
		return nil, errors.New("invalid instance")
	}

	searchArgs := strings.Split(strings.Trim(req.SearchArgs, " "), " ")

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

	res := &rpc.BoardSearchResp{Boards: []*rpc.BoardListItem{}}
	for _, targetPackage := range pm.Packages {
		for _, platform := range targetPackage.Platforms {
			latestPlatformRelease := platform.GetLatestRelease()
			if latestPlatformRelease == nil {
				continue
			}
			installedVersion := ""
			if installedPlatformRelease := pm.GetInstalledPlatformRelease(platform); installedPlatformRelease != nil {
				installedVersion = installedPlatformRelease.Version.String()
			}

			rpcPlatform := &rpc.Platform{
				ID:                platform.String(),
				Installed:         installedVersion,
				Latest:            latestPlatformRelease.Version.String(),
				Name:              platform.Name,
				Maintainer:        platform.Package.Maintainer,
				Website:           platform.Package.WebsiteURL,
				Email:             platform.Package.Email,
				ManuallyInstalled: platform.ManuallyInstalled,
			}

			// Platforms that are not installed don't have a list of boards
			// generated from their boards.txt file so we need two different
			// ways of reading board data.
			// The only boards information for platforms that are not installed
			// is that found in the index, usually that's only a board name.
			if len(latestPlatformRelease.Boards) != 0 {
				for _, board := range latestPlatformRelease.Boards {
					if !req.GetIncludeHiddenBoards() && board.IsHidden() {
						continue
					}

					toTest := append(strings.Split(board.Name(), " "), board.Name(), board.FQBN())
					if ok, err := match(toTest); err != nil {
						return nil, err
					} else if !ok {
						continue
					}

					res.Boards = append(res.Boards, &rpc.BoardListItem{
						Name:     board.Name(),
						FQBN:     board.FQBN(),
						IsHidden: board.IsHidden(),
						Platform: rpcPlatform,
					})
				}
			} else {
				for _, board := range latestPlatformRelease.BoardsManifest {
					toTest := append(strings.Split(board.Name, " "), board.Name)
					if ok, err := match(toTest); err != nil {
						return nil, err
					} else if !ok {
						continue
					}

					res.Boards = append(res.Boards, &rpc.BoardListItem{
						Name:     strings.Trim(board.Name, " \n"),
						Platform: rpcPlatform,
					})
				}
			}
		}
	}

	sort.Slice(res.Boards, func(i, j int) bool {
		return res.Boards[i].Name < res.Boards[j].Name
	})
	return res, nil
}
