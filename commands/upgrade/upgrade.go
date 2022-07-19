// This file is part of arduino-cli.
//
// Copyright 2022 ARDUINO SA (http://www.arduino.cc/)
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

package upgrade

import (
	"context"
	"strings"

	"github.com/arduino/arduino-cli/commands/core"
	"github.com/arduino/arduino-cli/commands/lib"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
)

// Upgrade downloads and installs outdated Cores and Libraries
func Upgrade(ctx context.Context, req *rpc.UpgradeRequest, downloadCB rpc.DownloadProgressCB, taskCB rpc.TaskProgressCB) error {
	libraryListResponse, err := lib.LibraryList(ctx, &rpc.LibraryListRequest{
		Instance:  req.GetInstance(),
		Updatable: true,
	})
	if err != nil {
		return err
	}

	getPlatformsResp, err := core.GetPlatforms(&rpc.PlatformListRequest{
		Instance:      req.GetInstance(),
		UpdatableOnly: true,
	})
	if err != nil {
		return err
	}

	for _, libToUpgrade := range libraryListResponse.GetInstalledLibraries() {
		err := lib.LibraryInstall(ctx, &rpc.LibraryInstallRequest{
			Instance: req.GetInstance(),
			Name:     libToUpgrade.GetLibrary().GetName(),
		}, downloadCB, taskCB)
		if err != nil {
			return err
		}
	}

	for _, platformToUpgrade := range getPlatformsResp {
		split := strings.Split(platformToUpgrade.GetId(), ":")
		_, err := core.PlatformUpgrade(ctx, &rpc.PlatformUpgradeRequest{
			Instance:        req.GetInstance(),
			PlatformPackage: split[0],
			Architecture:    split[1],
			SkipPostInstall: req.GetSkipPostInstall(),
		}, downloadCB, taskCB)
		if err != nil {
			return err
		}
	}

	return nil
}
