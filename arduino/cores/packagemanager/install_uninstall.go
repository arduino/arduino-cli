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

// UninstallPlatform remove a PlatformRelease.
func (pm *PackageManager) UninstallPlatform(platformRelease *cores.PlatformRelease) error {
	if platformRelease.InstallDir == nil {
		return fmt.Errorf("platform not installed")
	}

	// Safety measure
	if safe, err := platformRelease.InstallDir.IsInsideDir(pm.PackagesDir); err != nil {
		return fmt.Errorf("checking if plaform is installed in data dir: %s", err)
	} else if !safe {
		return fmt.Errorf("platform is not installed inside data dir")
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

// UninstallTool remove a ToolRelease.
func (pm *PackageManager) UninstallTool(toolRelease *cores.ToolRelease) error {
	if toolRelease.InstallDir == nil {
		return fmt.Errorf("tool not installed")
	}

	// Safety measure
	if safe, err := toolRelease.InstallDir.IsInsideDir(pm.PackagesDir); err != nil {
		return fmt.Errorf("checking if tool is installed in data dir: %s", err)
	} else if !safe {
		return fmt.Errorf("tool is not installed inside data dir")
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
	for _, targetPackage := range pm.packages.Packages {
		for _, platform := range targetPackage.Platforms {
			if platformRelease := platform.GetInstalled(); platformRelease != nil {
				if platformRelease.RequiresToolRelease(toolRelease) {
					return true
				}
			}
		}
	}
	return false
}
