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
	"github.com/arduino/arduino-cli/internal/arduino/cores"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
)

// platformToRPCPlatformMetadata converts our internal structure to the RPC structure.
func platformToRPCPlatformMetadata(platform *cores.Platform) *rpc.PlatformMetadata {
	return &rpc.PlatformMetadata{
		Id:                platform.String(),
		Maintainer:        platform.Package.Maintainer,
		Website:           platform.Package.WebsiteURL,
		Email:             platform.Package.Email,
		ManuallyInstalled: platform.ManuallyInstalled,
		Deprecated:        platform.Deprecated,
		Indexed:           platform.Indexed,
	}
}

// platformReleaseToRPC converts our internal structure to the RPC structure.
// Note: this function does not touch the "Installed" field of rpc.Platform as it's not always clear that the
// platformRelease we're currently converting is actually installed.
func platformReleaseToRPC(platformRelease *cores.PlatformRelease) *rpc.PlatformRelease {
	// If the boards are not installed yet, the `platformRelease.Boards` will be a zero length slice.
	// In such case, we have to use the `platformRelease.BoardsManifest` instead.
	// So that we can retrieve the name of the boards at least.
	var boards []*rpc.Board
	if len(platformRelease.Boards) > 0 {
		boards = make([]*rpc.Board, len(platformRelease.Boards))
		i := 0
		for _, b := range platformRelease.Boards {
			boards[i] = &rpc.Board{
				Name: b.Name(),
				Fqbn: b.FQBN(),
			}
			i++
		}
	} else {
		boards = make([]*rpc.Board, len(platformRelease.BoardsManifest))
		i := 0
		for _, m := range platformRelease.BoardsManifest {
			boards[i] = &rpc.Board{
				Name: m.Name,
				// FQBN is not available. Boards have to be installed first (-> `boards.txt`).
			}
			i++
		}
	}

	// This field make sense only if the platformRelease is installed otherwise is an "undefined behaviour"
	missingMetadata := platformRelease.IsInstalled() && !platformRelease.HasMetadata()
	return &rpc.PlatformRelease{
		Name:            platformRelease.Name,
		Help:            &rpc.HelpResources{Online: platformRelease.Platform.Package.Help.Online},
		Boards:          boards,
		Version:         platformRelease.Version.String(),
		Installed:       platformRelease.IsInstalled(),
		MissingMetadata: missingMetadata,
		Types:           []string{platformRelease.Category},
		Deprecated:      platformRelease.Deprecated,
		Compatible:      platformRelease.IsCompatible(),
	}
}
