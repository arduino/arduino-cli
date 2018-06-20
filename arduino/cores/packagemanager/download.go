/*
 * This file is part of arduino-cli.
 *
 * arduino-cli is free software; you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation; either version 2 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program; if not, write to the Free Software
 * Foundation, Inc., 51 Franklin St, Fifth Floor, Boston, MA  02110-1301  USA
 *
 * As a special exception, you may use this file as part of a free software
 * library without restriction.  Specifically, if other files instantiate
 * templates or use macros or inline functions from this file, or you compile
 * this file and link it with other files to produce an executable, this
 * file does not by itself cause the resulting executable to be covered by
 * the GNU General Public License.  This exception does not however
 * invalidate any other reasons why the executable file might be covered by
 * the GNU General Public License.
 *
 * Copyright 2017-2018 ARDUINO AG (http://www.arduino.cc/)
 */

package packagemanager

import (
	"fmt"
	"os"

	"github.com/cavaliercoder/grab"

	"github.com/bcmi-labs/arduino-cli/arduino/cores"
	"github.com/bcmi-labs/arduino-cli/common/formatter/output"
)

// PlatformReference represents a tuple to identify a Platform
type PlatformReference struct {
	Package              string // The package where this Platform belongs to.
	PlatformArchitecture string
	PlatformVersion      string
}

func (platform *PlatformReference) String() string {
	return platform.Package + ":" + platform.PlatformArchitecture + "@" + platform.PlatformVersion
}

// FindPlatform returns the PlatformRelease matching the PlatformReference or nil if not found
func (pm *PackageManager) FindPlatform(ref *PlatformReference) *cores.PlatformRelease {
	targetPackage, ok := pm.GetPackages().Packages[ref.Package]
	if !ok {
		return nil
	}
	platform, ok := targetPackage.Platforms[ref.PlatformArchitecture]
	if !ok {
		return nil
	}
	platformRelease, ok := platform.Releases[ref.PlatformVersion]
	if !ok {
		return nil
	}
	return platformRelease
}

// FindItemsToDownload takes a set of PlatformReference and returns a set of items to download and
// a set of outputs for non existing platforms.
func (pm *PackageManager) FindItemsToDownload(items []PlatformReference) (
	[]*cores.PlatformRelease, []*cores.ToolRelease, error) {

	retPlatforms := []*cores.PlatformRelease{}
	retTools := []*cores.ToolRelease{}
	added := map[string]bool{}

	for _, item := range items {
		targetPackage, exists := pm.packages.Packages[item.Package]
		if !exists {
			return nil, nil, fmt.Errorf("package %s not found", item.Package)
		}
		platform, exists := targetPackage.Platforms[item.PlatformArchitecture]
		if !exists {
			return nil, nil, fmt.Errorf("platform %s not found in package %s", item.PlatformArchitecture, targetPackage.String())
		}

		if added[platform.String()] {
			continue
		}
		added[platform.String()] = true

		var release *cores.PlatformRelease
		if item.PlatformVersion != "" {
			release = platform.GetRelease(item.PlatformVersion)
			if release == nil {
				return nil, nil, fmt.Errorf("required version %s not found for platform %s", item.PlatformVersion, platform.String())
			}
		} else {
			release = platform.GetLatestRelease()
			if release == nil {
				return nil, nil, fmt.Errorf("platform %s has no available releases", platform.String())
			}
		}
		retPlatforms = append(retPlatforms, release)

		// replaces "latest" with latest version too
		toolDeps, err := pm.packages.GetDepsOfPlatformRelease(release)
		if err != nil {
			return nil, nil, fmt.Errorf("getting tool dependencies for platform %s: %s", release.String(), err)
		}
		for _, tool := range toolDeps {
			if added[tool.String()] {
				continue
			}
			added[tool.String()] = true
			retTools = append(retTools, tool)
		}
	}
	return retPlatforms, retTools, nil
}

// DownloadToolRelease downloads a ToolRelease. If the tool is already downloaded a nil Response
// is returned.
func (pm *PackageManager) DownloadToolRelease(tool *cores.ToolRelease) (*grab.Response, error) {
	resource := tool.GetCompatibleFlavour()
	if resource == nil {
		return nil, fmt.Errorf("tool not available for your OS")
	}
	return resource.Download()
}

// DownloadPlatformRelease downloads a PlatformRelease. If the platform is already downloaded a
// nil Response is returned.
func (pm *PackageManager) DownloadPlatformRelease(platform *cores.PlatformRelease) (*grab.Response, error) {
	return platform.Resource.Download()
}

// FIXME: Make more generic and decouple the error print logic (that list should not exists;
// rather a failure @ the first package)

func (pm *PackageManager) InstallToolReleases(toolReleasesToDownload []*cores.ToolRelease,
	result *output.CoreProcessResults) error {

	for _, item := range toolReleasesToDownload {
		pm.Log.WithField("Package", item.Tool.Package.Name).
			WithField("Name", item.Tool.Name).
			WithField("Version", item.Version).
			Info("Installing tool")

		err := cores.InstallTool(item)
		var processResult output.ProcessResult
		if err != nil {
			if os.IsExist(err) {
				pm.Log.WithError(err).Warnf("Cannot install tool `%s`, it is already installed", item.Tool.Name)
				processResult = output.ProcessResult{
					Status: "Already Installed",
				}
			} else {
				pm.Log.WithError(err).Warnf("Cannot install tool `%s`", item.Tool.Name)
				processResult = output.ProcessResult{
					Error: err.Error(),
				}
			}
		} else {
			pm.Log.Info("Adding installed tool to final result")
			processResult = output.ProcessResult{
				Status: "Installed",
			}
		}
		name := item.String()
		processResult.ItemName = name
		result.Tools[name] = processResult
	}
	return nil
}

func (pm *PackageManager) InstallPlatformReleases(platformReleasesToDownload []*cores.PlatformRelease,
	outputResults *output.CoreProcessResults) error {

	for _, item := range platformReleasesToDownload {
		pm.Log.WithField("Package", item.Platform.Package.Name).
			WithField("Name", item.Platform.Name).
			WithField("Version", item.Version).
			Info("Installing core")

		err := cores.InstallPlatform(item)
		var result output.ProcessResult
		if err != nil {
			if os.IsExist(err) {
				pm.Log.WithError(err).Warnf("Cannot install core `%s`, it is already installed", item.Platform.Name)
				result = output.ProcessResult{
					Status: "Already Installed",
				}
			} else {
				pm.Log.WithError(err).Warnf("Cannot install core `%s`", item.Platform.Name)
				result = output.ProcessResult{
					Error: err.Error(),
				}
			}
		} else {
			pm.Log.Info("Adding installed core to final result")

			result = output.ProcessResult{
				Status: "Installed",
			}
		}
		name := item.String()
		result.ItemName = name
		outputResults.Cores[name] = result
	}
	return nil
}
