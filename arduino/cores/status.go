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

package cores

import (
	"errors"
	"fmt"

	"github.com/pmylund/sortutil"
)

// Packages represents a set of Packages
type Packages struct {
	Packages map[string]*Package // Maps packager name to Package
}

// NewPackages creates a new Packages object
func NewPackages() *Packages {
	return &Packages{
		Packages: map[string]*Package{},
	}
}

// Package represents a package in the system.
type Package struct {
	Name       string               // Name of the package.
	Maintainer string               // Name of the maintainer.
	WebsiteURL string               // Website of maintainer.
	Email      string               // Email of maintainer.
	Platforms  map[string]*Platform // The platforms in the system.
	Tools      map[string]*Tool     // The tools in the system.
	Packages   *Packages            `json:"-"`
}

// GetOrCreatePackage returns the specified Package or create an empty one
// filling all the cross-references
func (packages *Packages) GetOrCreatePackage(packager string) *Package {
	if targetPackage, ok := packages.Packages[packager]; ok {
		return targetPackage
	}
	targetPackage := &Package{
		Name:      packager,
		Platforms: map[string]*Platform{},
		Tools:     map[string]*Tool{},
		Packages:  packages,
		//Properties: properties.Map{},
	}
	packages.Packages[packager] = targetPackage
	return targetPackage
}

// Names returns the array containing the name of the packages.
func (packages *Packages) Names() []string {
	res := make([]string, len(packages.Packages))
	i := 0
	for n := range packages.Packages {
		res[i] = n
		i++
	}
	sortutil.CiAsc(res)
	return res
}

// GetOrCreatePlatform returns the Platform object with the specified architecture
// or creates a new one if not found
func (targetPackage *Package) GetOrCreatePlatform(architecure string) *Platform {
	if platform, ok := targetPackage.Platforms[architecure]; ok {
		return platform
	}
	targetPlatform := &Platform{
		Architecture: architecure,
		Releases:     map[string]*PlatformRelease{},
		Package:      targetPackage,
	}
	targetPackage.Platforms[architecure] = targetPlatform
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

func (tdep ToolDependency) extractTool(sc Packages) (*Tool, error) {
	pkg, exists := sc.Packages[tdep.ToolPackager]
	if !exists {
		return nil, errors.New("package not found")
	}
	tool, exists := pkg.Tools[tdep.ToolName]
	if !exists {
		return nil, errors.New("tool not found")
	}
	return tool, nil
}

func (tdep ToolDependency) extractRelease(sc Packages) (*ToolRelease, error) {
	tool, err := tdep.extractTool(sc)
	if err != nil {
		return nil, err
	}
	release, exists := tool.Releases[tdep.ToolVersion.String()]
	if !exists {
		return nil, errors.New("release Not Found")
	}
	return release, nil
}

// GetDepsOfPlatformRelease returns the deps of a specified release of a core.
func (packages *Packages) GetDepsOfPlatformRelease(release *PlatformRelease) ([]*ToolRelease, error) {
	if release == nil {
		return nil, errors.New("release cannot be nil")
	}
	ret := []*ToolRelease{}
	for _, dep := range release.Dependencies {
		pkg, exists := packages.Packages[dep.ToolPackager]
		if !exists {
			return nil, fmt.Errorf("package %s not found", dep.ToolPackager)
		}
		tool, exists := pkg.Tools[dep.ToolName]
		if !exists {
			return nil, fmt.Errorf("tool %s not found", dep.ToolName)
		}
		toolRelease, exists := tool.Releases[dep.ToolVersion.String()]
		if !exists {
			return nil, fmt.Errorf("tool version %s not found", dep.ToolVersion)
		}
		ret = append(ret, toolRelease)
	}
	return ret, nil
}
