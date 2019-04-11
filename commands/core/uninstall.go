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

	"github.com/arduino/arduino-cli/arduino/cores"
	"github.com/arduino/arduino-cli/arduino/cores/packagemanager"
	"github.com/arduino/arduino-cli/commands"
	"github.com/arduino/arduino-cli/common/formatter"
	"github.com/arduino/arduino-cli/rpc"
	semver "go.bug.st/relaxed-semver"
)

func PlatformUninstall(ctx context.Context, req *rpc.PlatformUninstallReq) (*rpc.PlatformUninstallResp, error) {
	// If no version is specified consider the installed
	version, err := semver.Parse(req.Version)
	if err != nil {
		return nil, fmt.Errorf("invalid version: %s", err)
	}
	pm := commands.GetPackageManager(req)
	if pm == nil {
		return nil, errors.New("invalid instance")
	}
	ref := &packagemanager.PlatformReference{
		Package:              req.PlatformPackage,
		PlatformArchitecture: req.Architecture,
		PlatformVersion:      version}
	if ref.PlatformVersion == nil {
		platform := pm.FindPlatform(ref)
		if platform == nil {
			formatter.PrintErrorMessage("Platform not found " + ref.String())
			return nil, fmt.Errorf("platform not found: %s", ref.String())

		}
		platformRelease := pm.GetInstalledPlatformRelease(platform)
		if platformRelease == nil {
			formatter.PrintErrorMessage("Platform not installed " + ref.String())
			return nil, fmt.Errorf("platform not installed: %s", ref.String())

		}
		ref.PlatformVersion = platformRelease.Version
	}

	platform, tools, err := pm.FindPlatformReleaseDependencies(ref)
	if err != nil {
		formatter.PrintError(err, "Could not determine platform dependencies")
		return nil, fmt.Errorf("finding platform dependencies: %s", err)
	}

	err = uninstallPlatformRelease(pm, platform)
	if err != nil {
		formatter.PrintError(err, "Error uninstalling "+platform.String())
		return nil, err
	}

	for _, tool := range tools {
		if !pm.IsToolRequired(tool) {
			fmt.Printf("Tool %s is no more required\n", tool)
			uninstallToolRelease(pm, tool)
		}
	}

	_, err = commands.Rescan(ctx, &rpc.RescanReq{Instance: req.Instance})
	if err != nil {
		return nil, err
	}
	return &rpc.PlatformUninstallResp{}, nil
}

func uninstallPlatformRelease(pm *packagemanager.PackageManager, platformRelease *cores.PlatformRelease) error {
	log := pm.Log.WithField("platform", platformRelease)

	log.Info("Uninstalling platform")
	formatter.Print("Uninstalling " + platformRelease.String() + "...")

	if err := pm.UninstallPlatform(platformRelease); err != nil {
		log.WithError(err).Error("Error uninstalling")
		formatter.PrintError(err, "Error uninstalling "+platformRelease.String())
		return fmt.Errorf("Error uninstalling "+platformRelease.String(), err)
	}

	log.Info("Platform uninstalled")
	formatter.Print(platformRelease.String() + " uninstalled")
	return nil
}

func uninstallToolRelease(pm *packagemanager.PackageManager, toolRelease *cores.ToolRelease) error {
	log := pm.Log.WithField("Tool", toolRelease)

	log.Info("Uninstalling tool")
	formatter.Print("Uninstalling " + toolRelease.String() + "...")

	if err := pm.UninstallTool(toolRelease); err != nil {
		log.WithError(err).Error("Error uninstalling")
		formatter.PrintError(err, "Error uninstalling "+toolRelease.String())
		return fmt.Errorf("Error uninstalling "+toolRelease.String(), err)
	}

	log.Info("Tool uninstalled")
	formatter.Print(toolRelease.String() + " uninstalled")
	return nil
}
