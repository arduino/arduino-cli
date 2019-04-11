/*
 * This file is part of arduino-cli.
 *
 * Copyright 2018 ARDUINO SA (http://www.arduino.cc/)
 *
 * This software is released under the GNU General Public License version 3,
 * which covers the main part of arduino-cli.
 * The terms of this license can be found at:
 * https://www.gnu.org/licenses/gpl-3.0.en.html
 *
 * You can be released from the requirements of the above licenses by purchasing
 * a commercial license. Buying such a license is mandatory if you want to modify or
 * otherwise use the software for commercial activities involving the Arduino
 * software without disclosing the source code of your own applications. To purchase
 * a commercial license, send an email to license@arduino.cc.
 */

package core

import (
	"context"
	"errors"
	"fmt"

	"github.com/arduino/arduino-cli/arduino/cores/packagemanager"
	"github.com/arduino/arduino-cli/commands"
	"github.com/arduino/arduino-cli/common/formatter"
	"github.com/arduino/arduino-cli/rpc"
	semver "go.bug.st/relaxed-semver"
)

func PlatformUpgrade(ctx context.Context, req *rpc.PlatformUpgradeReq,
	progress commands.ProgressCB, taskCB commands.TaskProgressCB) (*rpc.PlatformUpgradeResp, error) {
	// Extract all PlatformReference to platforms that have updates
	var version *semver.Version
	if req.Version != "" {
		if v, err := semver.Parse(req.Version); err == nil {
			version = v
		} else {
			return nil, fmt.Errorf("invalid version: %s", err)
		}
	}
	pm := commands.GetPackageManager(req)
	if pm == nil {
		return nil, errors.New("invalid instance")
	}

	ref := &packagemanager.PlatformReference{
		Package:              req.PlatformPackage,
		PlatformArchitecture: req.Architecture,
		PlatformVersion:      version}
	err := upgradePlatform(pm, ref, progress, taskCB)
	if err != nil {
		return nil, err
	}

	_, err = commands.Rescan(ctx, &rpc.RescanReq{Instance: req.Instance})
	if err != nil {
		return nil, err
	}

	return &rpc.PlatformUpgradeResp{}, nil
}

func upgradePlatform(pm *packagemanager.PackageManager, platformRef *packagemanager.PlatformReference,
	progress commands.ProgressCB, taskCB commands.TaskProgressCB) error {
	if platformRef.PlatformVersion != nil {
		formatter.PrintErrorMessage("Invalid item " + platformRef.String() + ", upgrade doesn't accept parameters with version")
		return fmt.Errorf("Invalid item " + platformRef.String() + ", upgrade doesn't accept parameters with version")
	}

	// Search the latest version for all specified platforms
	toInstallRefs := []*packagemanager.PlatformReference{}
	platform := pm.FindPlatform(platformRef)
	if platform == nil {
		formatter.PrintErrorMessage("Platform " + platformRef.String() + " not found")
		return fmt.Errorf("Platform " + platformRef.String() + " not found")
	}
	installed := pm.GetInstalledPlatformRelease(platform)
	if installed == nil {
		formatter.PrintErrorMessage("Platform " + platformRef.String() + " is not installed")
		return fmt.Errorf("Platform " + platformRef.String() + " is not installed")
	}
	latest := platform.GetLatestRelease()
	if !latest.Version.GreaterThan(installed.Version) {
		formatter.PrintResult("Platform " + platformRef.String() + " is already at the latest version.")
		return fmt.Errorf("Platform " + platformRef.String() + " is already at the latest version.")
	}
	platformRef.PlatformVersion = latest.Version
	toInstallRefs = append(toInstallRefs, platformRef)

	for _, platformRef := range toInstallRefs {
		platform, tools, err := pm.FindPlatformReleaseDependencies(platformRef)
		if err != nil {
			return fmt.Errorf("Platform " + platformRef.String() + " is not installed")
		}
		err = installPlatform(pm, platform, tools, progress, taskCB)
		if err != nil {
			return err
		}
	}
	return nil
}
