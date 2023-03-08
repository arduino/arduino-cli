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

package packagemanager

import (
	"encoding/json"
	"fmt"
	"runtime"

	"github.com/arduino/arduino-cli/arduino"
	"github.com/arduino/arduino-cli/arduino/cores"
	"github.com/arduino/arduino-cli/arduino/cores/packageindex"
	"github.com/arduino/arduino-cli/executils"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	"github.com/arduino/go-paths-helper"
	"github.com/pkg/errors"
)

// DownloadAndInstallPlatformUpgrades runs a full installation process to upgrade the given platform.
// This method takes care of downloading missing archives, upgrading platforms and tools, and
// removing the previously installed platform/tools that are no longer needed after the upgrade.
func (pme *Explorer) DownloadAndInstallPlatformUpgrades(
	platformRef *PlatformReference,
	downloadCB rpc.DownloadProgressCB,
	taskCB rpc.TaskProgressCB,
	skipPostInstall bool,
) error {
	if platformRef.PlatformVersion != nil {
		return &arduino.InvalidArgumentError{Message: tr("Upgrade doesn't accept parameters with version")}
	}

	// Search the latest version for all specified platforms
	platform := pme.FindPlatform(platformRef)
	if platform == nil {
		return &arduino.PlatformNotFoundError{Platform: platformRef.String()}
	}
	installed := pme.GetInstalledPlatformRelease(platform)
	if installed == nil {
		return &arduino.PlatformNotFoundError{Platform: platformRef.String()}
	}
	latest := platform.GetLatestRelease()
	if !latest.Version.GreaterThan(installed.Version) {
		return &arduino.PlatformAlreadyAtTheLatestVersionError{}
	}
	platformRef.PlatformVersion = latest.Version

	platformRelease, tools, err := pme.FindPlatformReleaseDependencies(platformRef)
	if err != nil {
		return &arduino.PlatformNotFoundError{Platform: platformRef.String()}
	}
	if err := pme.DownloadAndInstallPlatformAndTools(platformRelease, tools, downloadCB, taskCB, skipPostInstall); err != nil {
		return err
	}

	return nil
}

// DownloadAndInstallPlatformAndTools runs a full installation process for the given platform and tools.
// This method takes care of downloading missing archives, installing/upgrading platforms and tools, and
// removing the previously installed platform/tools that are no longer needed after the upgrade.
func (pme *Explorer) DownloadAndInstallPlatformAndTools(
	platformRelease *cores.PlatformRelease, requiredTools []*cores.ToolRelease,
	downloadCB rpc.DownloadProgressCB, taskCB rpc.TaskProgressCB,
	skipPostInstall bool) error {
	log := pme.log.WithField("platform", platformRelease)

	// Prerequisite checks before install
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
		if err := pme.DownloadToolRelease(tool, nil, downloadCB); err != nil {
			return err
		}
	}
	if err := pme.DownloadPlatformRelease(platformRelease, nil, downloadCB); err != nil {
		return err
	}
	taskCB(&rpc.TaskProgress{Completed: true})

	// Install tools first
	for _, tool := range toolsToInstall {
		if err := pme.InstallTool(tool, taskCB, skipPostInstall); err != nil {
			return err
		}
	}

	installed := pme.GetInstalledPlatformRelease(platformRelease.Platform)
	installedTools := []*cores.ToolRelease{}
	if installed == nil {
		// No version of this platform is installed
		log.Info("Installing platform")
		taskCB(&rpc.TaskProgress{Name: tr("Installing platform %s", platformRelease)})
	} else {
		// A platform with a different version is already installed
		log.Info("Replacing platform " + installed.String())
		taskCB(&rpc.TaskProgress{Name: tr("Replacing platform %[1]s with %[2]s", installed, platformRelease)})
		platformRef := &PlatformReference{
			Package:              platformRelease.Platform.Package.Name,
			PlatformArchitecture: platformRelease.Platform.Architecture,
			PlatformVersion:      installed.Version,
		}

		// Get a list of tools used by the currently installed platform version.
		// This must be done so tools used by the currently installed version are
		// removed if not used also by the newly installed version.
		var err error
		_, installedTools, err = pme.FindPlatformReleaseDependencies(platformRef)
		if err != nil {
			return &arduino.NotFoundError{Message: tr("Can't find dependencies for platform %s", platformRef), Cause: err}
		}
	}

	// Install
	if err := pme.InstallPlatform(platformRelease); err != nil {
		log.WithError(err).Error("Cannot install platform")
		return &arduino.FailedInstallError{Message: tr("Cannot install platform"), Cause: err}
	}

	// If upgrading remove previous release
	if installed != nil {
		uninstallErr := pme.UninstallPlatform(installed, taskCB)

		// In case of error try to rollback
		if uninstallErr != nil {
			log.WithError(uninstallErr).Error("Error upgrading platform.")
			taskCB(&rpc.TaskProgress{Message: tr("Error upgrading platform: %s", uninstallErr)})

			// Rollback
			if err := pme.UninstallPlatform(platformRelease, taskCB); err != nil {
				log.WithError(err).Error("Error rolling-back changes.")
				taskCB(&rpc.TaskProgress{Message: tr("Error rolling-back changes: %s", err)})
			}

			return &arduino.FailedInstallError{Message: tr("Cannot upgrade platform"), Cause: uninstallErr}
		}

		// Uninstall unused tools
		for _, tool := range installedTools {
			taskCB(&rpc.TaskProgress{Name: tr("Uninstalling %s, tool is no more required", tool)})
			if !pme.IsToolRequired(tool) {
				pme.UninstallTool(tool, taskCB)
			}
		}

	}

	// Perform post install
	if !skipPostInstall {
		log.Info("Running post_install script")
		taskCB(&rpc.TaskProgress{Message: tr("Configuring platform.")})
		if !platformRelease.IsInstalled() {
			return errors.New(tr("platform not installed"))
		}
		if err := pme.RunPostInstallScript(platformRelease.InstallDir); err != nil {
			taskCB(&rpc.TaskProgress{Message: tr("WARNING cannot configure platform: %s", err)})
		}
	} else {
		log.Info("Skipping platform configuration.")
		taskCB(&rpc.TaskProgress{Message: tr("Skipping platform configuration.")})
	}

	log.Info("Platform installed")
	taskCB(&rpc.TaskProgress{Message: tr("Platform %s installed", platformRelease), Completed: true})
	return nil
}

// InstallPlatform installs a specific release of a platform.
func (pme *Explorer) InstallPlatform(platformRelease *cores.PlatformRelease) error {
	destDir := pme.PackagesDir.Join(
		platformRelease.Platform.Package.Name,
		"hardware",
		platformRelease.Platform.Architecture,
		platformRelease.Version.String())
	return pme.InstallPlatformInDirectory(platformRelease, destDir)
}

// InstallPlatformInDirectory installs a specific release of a platform in a specific directory.
func (pme *Explorer) InstallPlatformInDirectory(platformRelease *cores.PlatformRelease, destDir *paths.Path) error {
	if err := platformRelease.Resource.Install(pme.DownloadDir, pme.tempDir, destDir); err != nil {
		return errors.Errorf(tr("installing platform %[1]s: %[2]s"), platformRelease, err)
	}
	if d, err := destDir.Abs(); err == nil {
		platformRelease.InstallDir = d
	} else {
		return err
	}
	if err := pme.cacheInstalledJSON(platformRelease); err != nil {
		return errors.Errorf(tr("creating installed.json in %[1]s: %[2]s"), platformRelease.InstallDir, err)
	}
	return nil
}

func (pme *Explorer) cacheInstalledJSON(platformRelease *cores.PlatformRelease) error {
	index := packageindex.IndexFromPlatformRelease(platformRelease)
	platformJSON, err := json.MarshalIndent(index, "", "  ")
	if err != nil {
		return err
	}
	installedJSON := platformRelease.InstallDir.Join("installed.json")
	installedJSON.WriteFile(platformJSON)
	return nil
}

// RunPostInstallScript runs the post_install.sh (or post_install.bat) script for the
// specified platformRelease or toolRelease.
func (pme *Explorer) RunPostInstallScript(installDir *paths.Path) error {
	postInstallFilename := "post_install.sh"
	if runtime.GOOS == "windows" {
		postInstallFilename = "post_install.bat"
	}
	postInstall := installDir.Join(postInstallFilename)
	if postInstall.Exist() && postInstall.IsNotDir() {
		cmd, err := executils.NewProcessFromPath(pme.GetEnvVarsForSpawnedProcess(), postInstall)
		if err != nil {
			return err
		}
		cmd.SetDirFromPath(installDir)
		if err := cmd.Run(); err != nil {
			return err
		}
	}
	return nil
}

// IsManagedPlatformRelease returns true if the PlatforRelease is managed by the PackageManager
func (pme *Explorer) IsManagedPlatformRelease(platformRelease *cores.PlatformRelease) bool {
	if pme.PackagesDir == nil {
		return false
	}
	installDir := platformRelease.InstallDir.Clone()
	if installDir.FollowSymLink() != nil {
		return false
	}
	packagesDir := pme.PackagesDir.Clone()
	if packagesDir.FollowSymLink() != nil {
		return false
	}
	managed, err := installDir.IsInsideDir(packagesDir)
	if err != nil {
		return false
	}
	return managed
}

// UninstallPlatform remove a PlatformRelease.
func (pme *Explorer) UninstallPlatform(platformRelease *cores.PlatformRelease, taskCB rpc.TaskProgressCB) error {
	log := pme.log.WithField("platform", platformRelease)

	log.Info("Uninstalling platform")
	taskCB(&rpc.TaskProgress{Name: tr("Uninstalling %s", platformRelease)})

	if platformRelease.InstallDir == nil {
		err := fmt.Errorf(tr("platform not installed"))
		log.WithError(err).Error("Error uninstalling")
		return &arduino.FailedUninstallError{Message: err.Error()}
	}

	// Safety measure
	if !pme.IsManagedPlatformRelease(platformRelease) {
		err := fmt.Errorf(tr("%s is not managed by package manager"), platformRelease)
		log.WithError(err).Error("Error uninstalling")
		return &arduino.FailedUninstallError{Message: err.Error()}
	}

	if err := platformRelease.InstallDir.RemoveAll(); err != nil {
		err = fmt.Errorf(tr("removing platform files: %s"), err)
		log.WithError(err).Error("Error uninstalling")
		return &arduino.FailedUninstallError{Message: err.Error()}
	}

	platformRelease.InstallDir = nil

	log.Info("Platform uninstalled")
	taskCB(&rpc.TaskProgress{Message: tr("Platform %s uninstalled", platformRelease), Completed: true})
	return nil
}

// InstallTool installs a specific release of a tool.
func (pme *Explorer) InstallTool(toolRelease *cores.ToolRelease, taskCB rpc.TaskProgressCB, skipPostInstall bool) error {
	log := pme.log.WithField("Tool", toolRelease)

	if toolRelease.IsInstalled() {
		log.Warn("Tool already installed")
		taskCB(&rpc.TaskProgress{Name: tr("Tool %s already installed", toolRelease), Completed: true})
		return nil
	}

	log.Info("Installing tool")
	taskCB(&rpc.TaskProgress{Name: tr("Installing %s", toolRelease)})

	toolResource := toolRelease.GetCompatibleFlavour()
	if toolResource == nil {
		return fmt.Errorf(tr("no compatible version of %[1]s tools found for the current os, try contacting %[2]s"),
			toolRelease.Tool.Name, toolRelease.Tool.Package.Email)
	}
	destDir := pme.PackagesDir.Join(
		toolRelease.Tool.Package.Name,
		"tools",
		toolRelease.Tool.Name,
		toolRelease.Version.String())
	err := toolResource.Install(pme.DownloadDir, pme.tempDir, destDir)
	if err != nil {
		log.WithError(err).Warn("Cannot install tool")
		return &arduino.FailedInstallError{Message: tr("Cannot install tool %s", toolRelease), Cause: err}
	}
	if d, err := destDir.Abs(); err == nil {
		toolRelease.InstallDir = d
	} else {
		return err
	}
	// Perform post install
	if !skipPostInstall {
		log.Info("Running tool post_install script")
		taskCB(&rpc.TaskProgress{Message: tr("Configuring tool.")})
		if err := pme.RunPostInstallScript(toolRelease.InstallDir); err != nil {
			taskCB(&rpc.TaskProgress{Message: tr("WARNING cannot configure tool: %s", err)})
		}
	} else {
		log.Info("Skipping tool configuration.")
		taskCB(&rpc.TaskProgress{Message: tr("Skipping tool configuration.")})
	}
	log.Info("Tool installed")
	taskCB(&rpc.TaskProgress{Message: tr("%s installed", toolRelease), Completed: true})

	return nil
}

// IsManagedToolRelease returns true if the ToolRelease is managed by the PackageManager
func (pme *Explorer) IsManagedToolRelease(toolRelease *cores.ToolRelease) bool {
	if pme.PackagesDir == nil {
		return false
	}
	installDir := toolRelease.InstallDir.Clone()
	if installDir.FollowSymLink() != nil {
		return false
	}
	packagesDir := pme.PackagesDir.Clone()
	if packagesDir.FollowSymLink() != nil {
		return false
	}
	managed, err := installDir.IsInsideDir(packagesDir)
	if err != nil {
		return false
	}
	return managed
}

// UninstallTool remove a ToolRelease.
func (pme *Explorer) UninstallTool(toolRelease *cores.ToolRelease, taskCB rpc.TaskProgressCB) error {
	log := pme.log.WithField("Tool", toolRelease)
	log.Info("Uninstalling tool")

	if toolRelease.InstallDir == nil {
		return fmt.Errorf(tr("tool not installed"))
	}

	// Safety measure
	if !pme.IsManagedToolRelease(toolRelease) {
		err := &arduino.FailedUninstallError{Message: tr("tool %s is not managed by package manager", toolRelease)}
		log.WithError(err).Error("Error uninstalling")
		return err
	}

	if err := toolRelease.InstallDir.RemoveAll(); err != nil {
		err = &arduino.FailedUninstallError{Message: err.Error()}
		log.WithError(err).Error("Error uninstalling")
		return err
	}

	toolRelease.InstallDir = nil

	log.Info("Tool uninstalled")
	taskCB(&rpc.TaskProgress{Message: tr("Tool %s uninstalled", toolRelease), Completed: true})
	return nil
}

// IsToolRequired returns true if any of the installed platforms requires the toolRelease
// passed as parameter
func (pme *Explorer) IsToolRequired(toolRelease *cores.ToolRelease) bool {
	// Search in all installed platforms
	for _, targetPackage := range pme.packages {
		for _, platform := range targetPackage.Platforms {
			if platformRelease := pme.GetInstalledPlatformRelease(platform); platformRelease != nil {
				if platformRelease.RequiresToolRelease(toolRelease) {
					return true
				}
			}
		}
	}
	return false
}
