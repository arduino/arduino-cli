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
	"fmt"

	"github.com/arduino/arduino-cli/arduino/cores"
	"github.com/arduino/arduino-cli/arduino/resources"
	"go.bug.st/downloader/v2"
	semver "go.bug.st/relaxed-semver"
)

// PlatformReference represents a tuple to identify a Platform
type PlatformReference struct {
	Package              string // The package where this Platform belongs to.
	PlatformArchitecture string
	PlatformVersion      *semver.Version
}

func (platform *PlatformReference) String() string {
	res := platform.Package + ":" + platform.PlatformArchitecture
	if platform.PlatformVersion != nil {
		return res + "@" + platform.PlatformVersion.String()
	}
	return res
}

// FindPlatform returns the Platform matching the PlatformReference or nil if not found.
// The PlatformVersion field of the reference is ignored.
func (pm *PackageManager) FindPlatform(ref *PlatformReference) *cores.Platform {
	targetPackage, ok := pm.Packages[ref.Package]
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

// FindPlatformReleaseDependencies takes a PlatformReference and returns a set of items to download and
// a set of outputs for non existing platforms.
func (pm *PackageManager) FindPlatformReleaseDependencies(item *PlatformReference) (*cores.PlatformRelease, []*cores.ToolRelease, error) {
	targetPackage, exists := pm.Packages[item.Package]
	if !exists {
		return nil, nil, fmt.Errorf(tr("package %s not found"), item.Package)
	}
	platform, exists := targetPackage.Platforms[item.PlatformArchitecture]
	if !exists {
		return nil, nil, fmt.Errorf(tr("platform %[1]s not found in package %[2]s"), item.PlatformArchitecture, targetPackage.String())
	}

	var release *cores.PlatformRelease
	if item.PlatformVersion != nil {
		release = platform.FindReleaseWithVersion(item.PlatformVersion)
		if release == nil {
			return nil, nil, fmt.Errorf(tr("required version %[1]s not found for platform %[2]s"), item.PlatformVersion, platform.String())
		}
	} else {
		release = platform.GetLatestRelease()
		if release == nil {
			return nil, nil, fmt.Errorf(tr("platform %s has no available releases"), platform.String())
		}
	}

	// replaces "latest" with latest version too
	toolDeps, err := pm.Packages.GetPlatformReleaseToolDependencies(release)
	if err != nil {
		return nil, nil, fmt.Errorf(tr("getting tool dependencies for platform %[1]s: %[2]s"), release.String(), err)
	}

	// discovery dependencies differ from normal tool since we always want to use the latest
	// available version for the platform package
	discoveryDependencies, err := pm.Packages.GetPlatformReleaseDiscoveryDependencies(release)
	if err != nil {
		return nil, nil, fmt.Errorf(tr("getting discovery dependencies for platform %[1]s: %[2]s"), release.String(), err)
	}
	toolDeps = append(toolDeps, discoveryDependencies...)

	// monitor dependencies differ from normal tool since we always want to use the latest
	// available version for the platform package
	monitorDependencies, err := pm.Packages.GetPlatformReleaseMonitorDependencies(release)
	if err != nil {
		return nil, nil, fmt.Errorf(tr("getting monitor dependencies for platform %[1]s: %[2]s"), release.String(), err)
	}
	toolDeps = append(toolDeps, monitorDependencies...)

	return release, toolDeps, nil
}

// DownloadToolRelease downloads a ToolRelease. If the tool is already downloaded a nil Downloader
// is returned.
func (pm *PackageManager) DownloadToolRelease(tool *cores.ToolRelease, config *downloader.Config, label string, progressCB resources.DownloadProgressCB) error {
	resource := tool.GetCompatibleFlavour()
	if resource == nil {
		return fmt.Errorf(tr("tool not available for your OS"))
	}
	return resource.Download(pm.DownloadDir, config, label, progressCB)
}

// DownloadPlatformRelease downloads a PlatformRelease. If the platform is already downloaded a
// nil Downloader is returned.
func (pm *PackageManager) DownloadPlatformRelease(platform *cores.PlatformRelease, config *downloader.Config, label string, progressCB resources.DownloadProgressCB) error {
	return platform.Resource.Download(pm.DownloadDir, config, label, progressCB)
}
