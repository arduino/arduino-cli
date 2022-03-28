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

package cores

import (
	"errors"
	"fmt"

	"github.com/pmylund/sortutil"
)

// Packages represents a set of Packages
type Packages map[string]*Package // Maps packager name to Package

// NewPackages creates a new Packages object
func NewPackages() Packages {
	return map[string]*Package{}
}

// PackageHelp contains info on how to reach maintainers for help
type PackageHelp struct {
	Online string `json:"online,omitempty"`
}

// Package represents a package in the system.
type Package struct {
	Name       string               // Name of the package.
	Maintainer string               // Name of the maintainer.
	WebsiteURL string               // Website of maintainer.
	URL        string               // origin URL for package index json file.
	Email      string               // Email of maintainer.
	Platforms  map[string]*Platform // The platforms in the system.
	Tools      map[string]*Tool     // The tools in the system.
	Help       PackageHelp          `json:"-"`
	Packages   Packages             `json:"-"`
}

// GetOrCreatePackage returns the specified Package or create an empty one
// filling all the cross-references
func (packages Packages) GetOrCreatePackage(packager string) *Package {
	if targetPackage, ok := packages[packager]; ok {
		return targetPackage
	}
	targetPackage := &Package{
		Name:      packager,
		Platforms: map[string]*Platform{},
		Tools:     map[string]*Tool{},
		Packages:  packages,
	}
	packages[packager] = targetPackage
	return targetPackage
}

// Names returns the array containing the name of the packages.
func (packages Packages) Names() []string {
	res := make([]string, len(packages))
	i := 0
	for n := range packages {
		res[i] = n
		i++
	}
	sortutil.CiAsc(res)
	return res
}

// GetPlatformReleaseToolDependencies returns the tool releases needed by the specified PlatformRelease
func (packages Packages) GetPlatformReleaseToolDependencies(release *PlatformRelease) ([]*ToolRelease, error) {
	if release == nil {
		return nil, errors.New(tr("release cannot be nil"))
	}
	ret := []*ToolRelease{}
	for _, dep := range release.ToolDependencies {
		pkg, exists := packages[dep.ToolPackager]
		if !exists {
			return nil, fmt.Errorf(tr("package %s not found"), dep.ToolPackager)
		}
		tool, exists := pkg.Tools[dep.ToolName]
		if !exists {
			return nil, fmt.Errorf(tr("tool %s not found"), dep.ToolName)
		}
		toolRelease, exists := tool.Releases[dep.ToolVersion.String()]
		if !exists {
			return nil, fmt.Errorf(tr("tool version %s not found"), dep.ToolVersion)
		}
		ret = append(ret, toolRelease)
	}
	return ret, nil
}

// GetPlatformReleaseDiscoveryDependencies returns the discovery releases needed by the specified PlatformRelease
func (packages Packages) GetPlatformReleaseDiscoveryDependencies(release *PlatformRelease) ([]*ToolRelease, error) {
	if release == nil {
		return nil, fmt.Errorf(tr("release cannot be nil"))
	}

	res := []*ToolRelease{}
	for _, discovery := range release.DiscoveryDependencies {
		pkg, exists := packages[discovery.Packager]
		if !exists {
			return nil, fmt.Errorf(tr("package %s not found"), discovery.Packager)
		}
		tool, exists := pkg.Tools[discovery.Name]
		if !exists {
			return nil, fmt.Errorf(tr("tool %s not found"), discovery.Name)
		}

		// We always want to use the latest available release for discoveries
		latestRelease := tool.LatestRelease()
		if latestRelease == nil {
			return nil, fmt.Errorf(tr("can't find latest release of %s"), discovery.Name)
		}
		res = append(res, latestRelease)
	}
	return res, nil
}

// GetPlatformReleaseMonitorDependencies returns the monitor releases needed by the specified PlatformRelease
func (packages Packages) GetPlatformReleaseMonitorDependencies(release *PlatformRelease) ([]*ToolRelease, error) {
	if release == nil {
		return nil, fmt.Errorf(tr("release cannot be nil"))
	}

	res := []*ToolRelease{}
	for _, monitor := range release.MonitorDependencies {
		pkg, exists := packages[monitor.Packager]
		if !exists {
			return nil, fmt.Errorf(tr("package %s not found"), monitor.Packager)
		}
		tool, exists := pkg.Tools[monitor.Name]
		if !exists {
			return nil, fmt.Errorf(tr("tool %s not found"), monitor.Name)
		}

		// We always want to use the latest available release for monitors
		latestRelease := tool.LatestRelease()
		if latestRelease == nil {
			return nil, fmt.Errorf(tr("can't find latest release of %s"), monitor.Name)
		}
		res = append(res, latestRelease)
	}
	return res, nil
}

// GetOrCreatePlatform returns the Platform object with the specified architecture
// or creates a new one if not found
func (targetPackage *Package) GetOrCreatePlatform(architecture string) *Platform {
	if platform, ok := targetPackage.Platforms[architecture]; ok {
		return platform
	}
	targetPlatform := &Platform{
		Architecture: architecture,
		Releases:     map[string]*PlatformRelease{},
		Package:      targetPackage,
	}
	targetPackage.Platforms[architecture] = targetPlatform
	return targetPlatform
}

// GetOrCreateTool returns the Tool object with the specified name
// or creates a new one if not found
func (targetPackage *Package) GetOrCreateTool(name string) *Tool {
	if tool, ok := targetPackage.Tools[name]; ok {
		return tool
	}
	tool := &Tool{
		Name:     name,
		Package:  targetPackage,
		Releases: map[string]*ToolRelease{},
	}
	targetPackage.Tools[name] = tool
	return tool
}

func (targetPackage *Package) String() string {
	return targetPackage.Name
}
