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

	"github.com/arduino/go-properties-map"
	"github.com/pmylund/sortutil"
)

// Packages represents a set of Packages
type Packages struct {
	Packages map[string]*Package // Maps packager name to Package

	Properties properties.Map `json:"-"` // TODO: used to add "ctags" properties, TO BE REMOVED
}

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
	Packages   *Packages `json:"-"`
}

// Names returns the array containing the name of the packages.
func (sc Packages) Names() []string {
	res := make([]string, len(sc.Packages))
	i := 0
	for n := range sc.Packages {
		res[i] = n
		i++
	}
	sortutil.CiAsc(res)
	return res
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
func (sc Packages) GetDepsOfPlatformRelease(release *PlatformRelease) ([]*ToolRelease, error) {
	if release == nil {
		return nil, errors.New("release cannot be nil")
	}
	ret := []*ToolRelease{}
	for _, dep := range release.Dependencies {
		pkg, exists := sc.Packages[dep.ToolPackager]
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
