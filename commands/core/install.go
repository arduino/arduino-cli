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

package core

import (
	"context"
	"fmt"

	"github.com/arduino/arduino-cli/arduino/cores"
	"github.com/arduino/arduino-cli/arduino/cores/packagemanager"
	"github.com/arduino/arduino-cli/commands"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	"github.com/pkg/errors"
)

// PlatformInstall FIXMEDOC
func PlatformInstall(ctx context.Context, req *rpc.PlatformInstallRequest,
	downloadCB commands.DownloadProgressCB, taskCB commands.TaskProgressCB) (*rpc.PlatformInstallResponse, error) {

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

	err = installPlatform(pm, platform, tools, downloadCB, taskCB, req.GetSkipPostInstall())
	if err != nil {
		return nil, err
	}

	_, status := commands.Init(&rpc.InitRequest{Instance: &rpc.Instance{Id: req.Instance.Id}})
	if status != nil {
		return nil, status.Err()
	}

	return &rpc.PlatformInstallResponse{}, nil
}

func installPlatform(pm *packagemanager.PackageManager,
	platformRelease *cores.PlatformRelease, requiredTools []*cores.ToolRelease,
	downloadCB commands.DownloadProgressCB, taskCB commands.TaskProgressCB,
	skipPostInstall bool) error {
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
		if err := downloadTool(pm, tool, downloadCB); err != nil {
			return err
		}
	}
	err := downloadPlatform(pm, platformRelease, downloadCB)
	if err != nil {
		return err
	}
	taskCB(&rpc.TaskProgress{Completed: true})

	// Install tools first
	for _, tool := range toolsToInstall {
		err := commands.InstallToolRelease(pm, tool, taskCB)
		if err != nil {
			return err
		}
	}

	installed := pm.GetInstalledPlatformRelease(platformRelease.Platform)
	installedTools := []*cores.ToolRelease{}
	if installed == nil {
		// No version of this platform is installed
		log.Info("Installing platform")
		taskCB(&rpc.TaskProgress{Name: "Installing " + platformRelease.String()})
	} else {
		// A platform with a different version is already installed
		log.Info("Upgrading platform " + installed.String())
		taskCB(&rpc.TaskProgress{Name: "Upgrading " + installed.String() + " with " + platformRelease.String()})
		platformRef := &packagemanager.PlatformReference{
			Package:              platformRelease.Platform.Package.Name,
			PlatformArchitecture: platformRelease.Platform.Architecture,
			PlatformVersion:      installed.Version,
		}

		// Get a list of tools used by the currently installed platform version.
		// This must be done so tools used by the currently installed version are
		// removed if not used also by the newly installed version.
		var err error
		_, installedTools, err = pm.FindPlatformReleaseDependencies(platformRef)
		if err != nil {
			return fmt.Errorf("can't find dependencies for platform %s: %w", platformRef, err)
		}
	}

	// Install
	err = pm.InstallPlatform(platformRelease)
	if err != nil {
		log.WithError(err).Error("Cannot install platform")
		return err
	}

	// If upgrading remove previous release
	if installed != nil {
		errUn := pm.UninstallPlatform(installed)

		// In case of error try to rollback
		if errUn != nil {
			log.WithError(errUn).Error("Error upgrading platform.")
			taskCB(&rpc.TaskProgress{Message: "Error upgrading platform: " + err.Error()})

			// Rollback
			if err := pm.UninstallPlatform(platformRelease); err != nil {
				log.WithError(err).Error("Error rolling-back changes.")
				taskCB(&rpc.TaskProgress{Message: "Error rolling-back changes: " + err.Error()})
			}

			return fmt.Errorf("upgrading platform: %s", errUn)
		}

		// Uninstall unused tools
		for _, tool := range installedTools {
			if !pm.IsToolRequired(tool) {
				uninstallToolRelease(pm, tool, taskCB)
			}
		}

	}

	// Perform post install
	if !skipPostInstall {
		log.Info("Running post_install script")
		taskCB(&rpc.TaskProgress{Message: "Configuring platform"})
		if err := pm.RunPostInstallScript(platformRelease); err != nil {
			taskCB(&rpc.TaskProgress{Message: fmt.Sprintf("WARNING: cannot run post install: %s", err)})
		}
	} else {
		log.Info("Skipping platform configuration (post_install run).")
		taskCB(&rpc.TaskProgress{Message: "Skipping platform configuration"})
	}

	log.Info("Platform installed")
	taskCB(&rpc.TaskProgress{Message: platformRelease.String() + " installed", Completed: true})
	return nil
}
