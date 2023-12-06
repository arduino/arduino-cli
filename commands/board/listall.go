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

	"github.com/arduino/arduino-cli/commands"
	"github.com/arduino/arduino-cli/commands/cmderrors"
	"github.com/arduino/arduino-cli/commands/internal/instances"
	"github.com/arduino/arduino-cli/internal/arduino/cores"
	"github.com/arduino/arduino-cli/internal/arduino/utils"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
)

// ListAll FIXMEDOC
func ListAll(ctx context.Context, req *rpc.BoardListAllRequest) (*rpc.BoardListAllResponse, error) {
	pme, release := instances.GetPackageManagerExplorer(req.GetInstance())
	if pme == nil {
		return nil, &cmderrors.InvalidInstanceError{}
	}
	defer release()

	searchArgs := strings.Join(req.GetSearchArgs(), " ")

	list := &rpc.BoardListAllResponse{Boards: []*rpc.BoardListItem{}}
	for _, targetPackage := range toSortedPackageArray(pme.GetPackages()) {
		for _, platform := range toSortedPlatformArray(targetPackage.Platforms) {
			installedPlatformRelease := pme.GetInstalledPlatformRelease(platform)
			// We only want to list boards for installed platforms
			if installedPlatformRelease == nil {
				continue
			}

			rpcPlatform := &rpc.Platform{
				Metadata: commands.PlatformToRPCPlatformMetadata(platform),
				Release:  commands.PlatformReleaseToRPC(installedPlatformRelease),
			}

			toTest := []string{
				platform.String(),
				installedPlatformRelease.Name,
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

				list.Boards = append(list.GetBoards(), &rpc.BoardListItem{
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

// TODO use a generic function instead of the two below once go >1.18 is adopted.
//		Without generics we either have to create multiple functions for different map types
//		or resort to type assertions on the caller side

// toSortedPackageArray takes a packages map and returns its values as array
// ordered by the map keys alphabetically
func toSortedPackageArray(sourceMap cores.Packages) []*cores.Package {
	keys := []string{}
	for key := range sourceMap {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	sortedValues := make([]*cores.Package, len(keys))
	for i, key := range keys {
		sortedValues[i] = sourceMap[key]
	}
	return sortedValues
}

// toSortedPlatformArray takes a packages map and returns its values as array
// ordered by the map keys alphabetically
func toSortedPlatformArray(sourceMap map[string]*cores.Platform) []*cores.Platform {
	keys := []string{}
	for key := range sourceMap {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	sortedValues := make([]*cores.Platform, len(keys))
	for i, key := range keys {
		sortedValues[i] = sourceMap[key]
	}
	return sortedValues
}
