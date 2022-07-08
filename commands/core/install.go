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

	"github.com/arduino/arduino-cli/arduino"
	"github.com/arduino/arduino-cli/arduino/cores"
	"github.com/arduino/arduino-cli/arduino/cores/packagemanager"
	"github.com/arduino/arduino-cli/commands"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
)

// PlatformInstall FIXMEDOC
func PlatformInstall(ctx context.Context, req *rpc.PlatformInstallRequest,
	downloadCB rpc.DownloadProgressCB, taskCB rpc.TaskProgressCB) (*rpc.PlatformInstallResponse, error) {

	pm := commands.GetPackageManager(req.GetInstance().GetId())
	if pm == nil {
		return nil, &arduino.InvalidInstanceError{}
	}

	version, err := commands.ParseVersion(req)
	if err != nil {
		return nil, &arduino.InvalidVersionError{Cause: err}
	}

	ref := &packagemanager.PlatformReference{
		Package:              req.PlatformPackage,
		PlatformArchitecture: req.Architecture,
		PlatformVersion:      version,
	}
	platformRelease, tools, err := pm.FindPlatformReleaseDependencies(ref)
	if err != nil {
		return nil, &arduino.PlatformNotFoundError{Platform: ref.String(), Cause: err}
	}

	didInstall, err := installPlatform(pm, platformRelease, tools, downloadCB, taskCB, req.GetSkipPostInstall())
	if err != nil {
		return nil, err
	}

	if didInstall {
		if err := commands.Init(&rpc.InitRequest{Instance: req.Instance}, nil); err != nil {
			return nil, err
		}
	}

	return &rpc.PlatformInstallResponse{}, nil
}

func installPlatform(pm *packagemanager.PackageManager,
	platformRelease *cores.PlatformRelease, requiredTools []*cores.ToolRelease,
	downloadCB rpc.DownloadProgressCB, taskCB rpc.TaskProgressCB,
	skipPostInstall bool) (bool, error) {
	log := pm.Log.WithField("platform", platformRelease)

	// Prerequisite checks before install
	if platformRelease.IsInstalled() {
		log.Warn("Platform already installed")
		taskCB(&rpc.TaskProgress{Name: tr("Platform %s already installed", platformRelease), Completed: true})
		return false, nil
	}
	toolsToInstall := []*cores.ToolRelease{}
	for _, tool := range requiredTools {
		if tool.IsInstalled() {
			log.WithField("tool", tool).Warn("Tool already installed")
			taskCB(&rpc.TaskProgress{Name: tr("Tool %s already installed", tool), Completed: true})
		} else {
			toolsToInstall = append(toolsToInstall, tool)
		}
	}

	// Package download
	taskCB(&rpc.TaskProgress{Name: tr("Downloading packages")})
	for _, tool := range toolsToInstall {
		if err := downloadTool(pm, tool, downloadCB); err != nil {
			return false, err
		}
	}
	if err := downloadPlatform(pm, platformRelease, downloadCB); err != nil {
		return false, err
	}
	taskCB(&rpc.TaskProgress{Completed: true})

	// Install tools first
	for _, tool := range toolsToInstall {
		if err := commands.InstallToolRelease(pm, tool, taskCB); err != nil {
			return false, err
		}
	}

	installed := pm.GetInstalledPlatformRelease(platformRelease.Platform)
	installedTools := []*cores.ToolRelease{}
	if installed == nil {
		// No version of this platform is installed
		log.Info("Installing platform")
		taskCB(&rpc.TaskProgress{Name: tr("Installing platform %s", platformRelease)})
	} else {
		// A platform with a different version is already installed
		log.Info("Upgrading platform " + installed.String())
		taskCB(&rpc.TaskProgress{Name: tr("Upgrading platform %[1]s with %[2]s", installed, platformRelease)})
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
			return false, &arduino.NotFoundError{Message: tr("Can't find dependencies for platform %s", platformRef), Cause: err}
		}
	}

	// Install
	if err := pm.InstallPlatform(platformRelease); err != nil {
		log.WithError(err).Error("Cannot install platform")
		return false, &arduino.FailedInstallError{Message: tr("Cannot install platform"), Cause: err}
	}

	// If upgrading remove previous release
	if installed != nil {
		uninstallErr := pm.UninstallPlatform(installed)

		// In case of error try to rollback
		if uninstallErr != nil {
			log.WithError(uninstallErr).Error("Error upgrading platform.")
			taskCB(&rpc.TaskProgress{Message: tr("Error upgrading platform: %s", uninstallErr)})

			// Rollback
			if err := pm.UninstallPlatform(platformRelease); err != nil {
				log.WithError(err).Error("Error rolling-back changes.")
				taskCB(&rpc.TaskProgress{Message: tr("Error rolling-back changes: %s", err)})
			}

			return false, &arduino.FailedInstallError{Message: tr("Cannot upgrade platform"), Cause: uninstallErr}
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
		taskCB(&rpc.TaskProgress{Message: tr("Configuring platform.")})
		if err := pm.RunPostInstallScript(platformRelease); err != nil {
			taskCB(&rpc.TaskProgress{Message: tr("WARNING cannot configure platform: %s", err)})
		}
	} else {
		log.Info("Skipping platform configuration.")
		taskCB(&rpc.TaskProgress{Message: tr("Skipping platform configuration.")})
	}

	log.Info("Platform installed")
	taskCB(&rpc.TaskProgress{Message: tr("Platform %s installed", platformRelease), Completed: true})
	return true, nil
}
