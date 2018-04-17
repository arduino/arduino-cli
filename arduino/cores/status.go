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
 * Copyright 2017 ARDUINO AG (http://www.arduino.cc/)
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
		return nil, errors.New("Package not found")
	}
	tool, exists := pkg.Tools[tdep.ToolName]
	if !exists {
		return nil, errors.New("Tool not found")
	}
	return tool, nil
}

func (tdep ToolDependency) extractRelease(sc Packages) (*ToolRelease, error) {
	tool, err := tdep.extractTool(sc)
	if err != nil {
		return nil, err
	}
	release, exists := tool.Releases[tdep.ToolVersion]
	if !exists {
		return nil, errors.New("Release Not Found")
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
			return nil, fmt.Errorf("Package %s not found", dep.ToolPackager)
		}
		tool, exists := pkg.Tools[dep.ToolName]
		if !exists {
			return nil, fmt.Errorf("Tool %s not found", dep.ToolName)
		}
		toolRelease, exists := tool.Releases[dep.ToolVersion]
		if !exists {
			return nil, fmt.Errorf("Tool version %s not found", dep.ToolVersion)
		}
		ret = append(ret, toolRelease)
	}
	return ret, nil
}
