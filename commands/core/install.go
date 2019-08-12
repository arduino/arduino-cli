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
	"net/http"

	"github.com/arduino/arduino-cli/arduino/cores"
	"github.com/arduino/arduino-cli/arduino/cores/packagemanager"
	"github.com/arduino/arduino-cli/commands"
	rpc "github.com/arduino/arduino-cli/rpc/commands"
)

// PlatformInstall FIXMEDOC
func PlatformInstall(ctx context.Context, req *rpc.PlatformInstallReq,
	downloadCB commands.DownloadProgressCB, taskCB commands.TaskProgressCB, downloaderHeaders http.Header) (*rpc.PlatformInstallResp, error) {

	pm := commands.GetPackageManager(req.GetInstance().GetId())
	if pm == nil {
		return nil, errors.New("invalid instance")
	}

	version, err := commands.ParseVersion(req)
	if err != nil {
		return nil, fmt.Errorf("invalid version: %s", err)
	}

	platform, tools, err := pm.FindPlatformReleaseDependencies(&packagemanager.PlatformReference{
		Package:              req.PlatformPackage,
		PlatformArchitecture: req.Architecture,
		PlatformVersion:      version,
	})
	if err != nil {
		return nil, fmt.Errorf("finding platform dependencies: %s", err)
	}

	err = installPlatform(pm, platform, tools, downloadCB, taskCB, downloaderHeaders)
	if err != nil {
		return nil, err
	}

	_, err = commands.Rescan(req.GetInstance().GetId())
	if err != nil {
		return nil, err
	}

	return &rpc.PlatformInstallResp{}, nil
}

func installPlatform(pm *packagemanager.PackageManager,
	platformRelease *cores.PlatformRelease, requiredTools []*cores.ToolRelease,
	downloadCB commands.DownloadProgressCB, taskCB commands.TaskProgressCB, downloaderHeaders http.Header) error {
	log := pm.Log.WithField("platform", platformRelease)

	// Prerequisite checks before install
	if platformRelease.IsInstalled() {
		log.Warn("Platform already installed")
		taskCB(&rpc.TaskProgress{Name: "Platform " + platformRelease.String() + " already installed", Completed: true})
		return nil
	}
	toolsToInstall := []*cores.ToolRelease{}
	for _, tool := range requiredTools {
		if tool.IsInstalled() {
			log.WithField("tool", tool).Warn("Tool already installed")
			taskCB(&rpc.TaskProgress{Name: "Tool " + tool.String() + " already installed", Completed: true})
		} else {
			toolsToInstall = append(toolsToInstall, tool)
		}
	}

	// Package download
	taskCB(&rpc.TaskProgress{Name: "Downloading packages"})
	for _, tool := range toolsToInstall {
		downloadTool(pm, tool, downloadCB, downloaderHeaders)
	}
	downloadPlatform(pm, platformRelease, downloadCB, downloaderHeaders)
	taskCB(&rpc.TaskProgress{Completed: true})

	// Install tools first
	for _, tool := range toolsToInstall {
		err := commands.InstallToolRelease(pm, tool, taskCB)
		if err != nil {
			// TODO: handle error
		}
	}

	// Are we installing or upgrading?
	platform := platformRelease.Platform
	installed := pm.GetInstalledPlatformRelease(platform)
	if installed == nil {
		log.Info("Installing platform")
		taskCB(&rpc.TaskProgress{Name: "Installing " + platformRelease.String()})
	} else {
		log.Info("Updating platform " + installed.String())
		taskCB(&rpc.TaskProgress{Name: "Updating " + installed.String() + " with " + platformRelease.String()})
	}

	// Install
	err := pm.InstallPlatform(platformRelease)
	if err != nil {
		log.WithError(err).Error("Cannot install platform")
		return err
	}

	// If upgrading remove previous release
	if installed != nil {
		errUn := pm.UninstallPlatform(installed)

		// In case of error try to rollback
		if errUn != nil {
			log.WithError(errUn).Error("Error updating platform.")
			taskCB(&rpc.TaskProgress{Message: "Error updating platform: " + err.Error()})

			// Rollback
			if err := pm.UninstallPlatform(platformRelease); err != nil {
				log.WithError(err).Error("Error rolling-back changes.")
				taskCB(&rpc.TaskProgress{Message: "Error rolling-back changes: " + err.Error()})
			}

			return fmt.Errorf("updating platform: %s", errUn)
		}
	}

	log.Info("Platform installed")
	taskCB(&rpc.TaskProgress{Message: platformRelease.String() + " installed", Completed: true})
	return nil
}
