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

package packagemanager

import (
	"fmt"

	"github.com/arduino/arduino-cli/arduino/cores"
)

// InstallPlatform installs a specific release of a platform.
func (pm *PackageManager) InstallPlatform(platformRelease *cores.PlatformRelease) error {
	destDir := pm.PackagesDir.Join(
		platformRelease.Platform.Package.Name,
		"hardware",
		platformRelease.Platform.Architecture,
		platformRelease.Version.String())
	return platformRelease.Resource.Install(pm.DownloadDir, pm.TempDir, destDir)
}

// IsManagedPlatformRelease returns true if the PlatforRelease is managed by the PackageManager
func (pm *PackageManager) IsManagedPlatformRelease(platformRelease *cores.PlatformRelease) bool {
	if pm.PackagesDir == nil {
		return false
	}
	installDir := platformRelease.InstallDir.Clone()
	if installDir.FollowSymLink() != nil {
		return false
	}
	packagesDir := pm.PackagesDir.Clone()
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
func (pm *PackageManager) UninstallPlatform(platformRelease *cores.PlatformRelease) error {
	if platformRelease.InstallDir == nil {
		return fmt.Errorf("platform not installed")
	}

	// Safety measure
	if !pm.IsManagedPlatformRelease(platformRelease) {
		return fmt.Errorf("%s is not managed by package manager", platformRelease)
	}

	if err := platformRelease.InstallDir.RemoveAll(); err != nil {
		return fmt.Errorf("removing platform files: %s", err)
	}
	platformRelease.InstallDir = nil
	return nil
}

// InstallTool installs a specific release of a tool.
func (pm *PackageManager) InstallTool(toolRelease *cores.ToolRelease) error {
	toolResource := toolRelease.GetCompatibleFlavour()
	if toolResource == nil {
		return fmt.Errorf("no compatible version of %s tools found for the current os", toolRelease.Tool.Name)
	}
	destDir := pm.PackagesDir.Join(
		toolRelease.Tool.Package.Name,
		"tools",
		toolRelease.Tool.Name,
		toolRelease.Version.String())
	return toolResource.Install(pm.DownloadDir, pm.TempDir, destDir)
}

// IsManagedToolRelease returns true if the ToolRelease is managed by the PackageManager
func (pm *PackageManager) IsManagedToolRelease(toolRelease *cores.ToolRelease) bool {
	if pm.PackagesDir == nil {
		return false
	}
	installDir := toolRelease.InstallDir.Clone()
	if installDir.FollowSymLink() != nil {
		return false
	}
	packagesDir := pm.PackagesDir.Clone()
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
func (pm *PackageManager) UninstallTool(toolRelease *cores.ToolRelease) error {
	if toolRelease.InstallDir == nil {
		return fmt.Errorf("tool not installed")
	}

	// Safety measure
	if !pm.IsManagedToolRelease(toolRelease) {
		return fmt.Errorf("tool %s is not managed by package manager", toolRelease)
	}

	if err := toolRelease.InstallDir.RemoveAll(); err != nil {
		return fmt.Errorf("removing tool files: %s", err)
	}
	toolRelease.InstallDir = nil
	return nil
}

// IsToolRequired returns true if any of the installed platforms requires the toolRelease
// passed as parameter
func (pm *PackageManager) IsToolRequired(toolRelease *cores.ToolRelease) bool {
	// Search in all installed platforms
	for _, targetPackage := range pm.Packages {
		for _, platform := range targetPackage.Platforms {
			if platformRelease := pm.GetInstalledPlatformRelease(platform); platformRelease != nil {
				if platformRelease.RequiresToolRelease(toolRelease) {
					return true
				}
			}
		}
	}
	return false
}
