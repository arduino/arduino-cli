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
	"sort"
	"strings"

	"github.com/arduino/arduino-cli/arduino"
	"github.com/arduino/arduino-cli/arduino/utils"
	"github.com/arduino/arduino-cli/commands"
	"github.com/arduino/arduino-cli/commands/internal/instances"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
)

// Search returns all boards that match the search arg.
// Boards are searched in all platforms, including those in the index that are not yet
// installed. Note that platforms that are not installed don't include boards' FQBNs.
// If no search argument is used all boards are returned.
func Search(ctx context.Context, req *rpc.BoardSearchRequest) (*rpc.BoardSearchResponse, error) {
	pme, release := instances.GetPackageManagerExplorer(req.GetInstance())
	if pme == nil {
		return nil, &arduino.InvalidInstanceError{}
	}
	defer release()

	foundBoards := []*rpc.BoardListItem{}
	for _, targetPackage := range pme.GetPackages() {
		for _, platform := range targetPackage.Platforms {
			latestPlatformRelease := platform.GetLatestCompatibleRelease()
			installedPlatformRelease := pme.GetInstalledPlatformRelease(platform)

			if latestPlatformRelease == nil && installedPlatformRelease == nil {
				continue
			}

			// Platforms that are not installed don't have a list of boards
			// generated from their boards.txt file so we need two different
			// ways of reading board data.
			// The only boards information for platforms that are not installed
			// is that found in the index, usually that's only a board name.
			if installedPlatformRelease != nil {
				for _, board := range installedPlatformRelease.Boards {
					if !req.GetIncludeHiddenBoards() && board.IsHidden() {
						continue
					}

					toTest := append(strings.Split(board.Name(), " "), board.Name(), board.FQBN())
					if !utils.MatchAny(req.GetSearchArgs(), toTest) {
						continue
					}

					foundBoards = append(foundBoards, &rpc.BoardListItem{
						Name:     board.Name(),
						Fqbn:     board.FQBN(),
						IsHidden: board.IsHidden(),
						Platform: &rpc.Platform{
							Metadata: commands.PlatformToRPCPlatformMetadata(platform),
							Release:  commands.PlatformReleaseToRPC(installedPlatformRelease),
						},
					})
				}
			} else if latestPlatformRelease != nil {
				for _, board := range latestPlatformRelease.BoardsManifest {
					toTest := append(strings.Split(board.Name, " "), board.Name)
					if !utils.MatchAny(req.GetSearchArgs(), toTest) {
						continue
					}

					foundBoards = append(foundBoards, &rpc.BoardListItem{
						Name: strings.Trim(board.Name, " \n"),
						Platform: &rpc.Platform{
							Metadata: commands.PlatformToRPCPlatformMetadata(platform),
							Release:  commands.PlatformReleaseToRPC(latestPlatformRelease),
						},
					})
				}
			}
		}
	}

	sort.Slice(foundBoards, func(i, j int) bool {
		if foundBoards[i].GetName() != foundBoards[j].GetName() {
			return foundBoards[i].GetName() < foundBoards[j].GetName()
		}
		return foundBoards[i].GetPlatform().GetMetadata().GetId() < foundBoards[j].GetPlatform().GetMetadata().GetId()
	})

	return &rpc.BoardSearchResponse{Boards: foundBoards}, nil
}
