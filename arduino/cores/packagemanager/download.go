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
	"os"

	"github.com/bcmi-labs/arduino-cli/arduino/cores"
	"github.com/bcmi-labs/arduino-cli/common/formatter/output"
	"github.com/cavaliercoder/grab"
	"go.bug.st/relaxed-semver"
)

// PlatformReference represents a tuple to identify a Platform
type PlatformReference struct {
	Package              string // The package where this Platform belongs to.
	PlatformArchitecture string
	PlatformVersion      *semver.Version
}

func (platform *PlatformReference) String() string {
	return platform.Package + ":" + platform.PlatformArchitecture + "@" + platform.PlatformVersion.String()
}

// FindPlatform returns the Platform matching the PlatformReference or nil if not found.
// The PlatformVersion field of the reference is ignored.
func (pm *PackageManager) FindPlatform(ref *PlatformReference) *cores.Platform {
	targetPackage, ok := pm.GetPackages().Packages[ref.Package]
	if !ok {
		return nil
	}
	platform, ok := targetPackage.Platforms[ref.PlatformArchitecture]
	if !ok {
		return nil
	}
	return platform
}

// FindPlatformRelease returns the PlatformRelease matching the PlatformReference or nil if not found
func (pm *PackageManager) FindPlatformRelease(ref *PlatformReference) *cores.PlatformRelease {
	platform := pm.FindPlatform(ref)
	if platform == nil {
		return nil
	}
	platformRelease, ok := platform.Releases[ref.PlatformVersion.String()]
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
		if item.PlatformVersion != nil {
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
	return resource.Download(pm.DownloadDir)
}

// DownloadPlatformRelease downloads a PlatformRelease. If the platform is already downloaded a
// nil Response is returned.
func (pm *PackageManager) DownloadPlatformRelease(platform *cores.PlatformRelease) (*grab.Response, error) {
	return platform.Resource.Download(pm.DownloadDir)
}

// FIXME: Make more generic and decouple the error print logic (that list should not exists;
// rather a failure @ the first package)

func (pm *PackageManager) InstallToolReleases(toolReleases []*cores.ToolRelease,
	result *output.CoreProcessResults) error {

	for _, toolRelease := range toolReleases {
		pm.Log.WithField("Package", toolRelease.Tool.Package.Name).
			WithField("Name", toolRelease.Tool.Name).
			WithField("Version", toolRelease.Version).
			Info("Installing tool")

		err := pm.InstallTool(toolRelease)
		var processResult output.ProcessResult
		if err != nil {
			if os.IsExist(err) {
				pm.Log.WithError(err).Warnf("Cannot install tool `%s`, it is already installed", toolRelease.Tool.Name)
				processResult = output.ProcessResult{
					Status: "Already Installed",
				}
			} else {
				pm.Log.WithError(err).Warnf("Cannot install tool `%s`", toolRelease.Tool.Name)
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
		name := toolRelease.String()
		processResult.ItemName = name
		result.Tools[name] = processResult
	}
	return nil
}

func (pm *PackageManager) InstallPlatformReleases(platformReleases []*cores.PlatformRelease,
	outputResults *output.CoreProcessResults) error {

	for _, platformRelease := range platformReleases {
		pm.Log.WithField("Package", platformRelease.Platform.Package.Name).
			WithField("Name", platformRelease.Platform.Name).
			WithField("Version", platformRelease.Version).
			Info("Installing platform")

		err := pm.InstallPlatform(platformRelease)
		var result output.ProcessResult
		if err != nil {
			if os.IsExist(err) {
				pm.Log.WithError(err).Warnf("Cannot install platform `%s`, it is already installed", platformRelease.Platform.Name)
				result = output.ProcessResult{
					Status: "Already Installed",
				}
			} else {
				pm.Log.WithError(err).Warnf("Cannot install platform `%s`", platformRelease.Platform.Name)
				result = output.ProcessResult{
					Error: err.Error(),
				}
			}
		} else {
			pm.Log.Info("Adding installed platform to final result")

			result = output.ProcessResult{
				Status: "Installed",
			}
		}
		name := platformRelease.String()
		result.ItemName = name
		outputResults.Cores[name] = result
	}
	return nil
}
