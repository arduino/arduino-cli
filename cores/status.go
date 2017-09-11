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
	"strings"

	"github.com/bcmi-labs/arduino-cli/common/releases"

	"github.com/bcmi-labs/arduino-cli/cmd/output"
	"github.com/pmylund/sortutil"
)

// StatusContext represents the Context of the Cores and Tools in the system.
type StatusContext struct {
	Packages map[string]*Package
}

// CoreDependency is a representation of a parsed core dependency (single ToolRelease).
type CoreDependency struct {
	ToolName string       `json:"tool,required"`
	Release  *ToolRelease `json:"release,required"`
}

func (cd CoreDependency) String() string {
	return strings.TrimSpace(fmt.Sprintln(cd.ToolName, " v.", cd.Release.Version))
}

// Add adds a package to a context from an indexPackage.
//
// NOTE: If the package is already in the context, it is overwritten!
func (sc *StatusContext) Add(indexPackage *indexPackage) {
	sc.Packages[indexPackage.Name] = indexPackage.extractPackage()
}

// Names returns the array containing the name of the packages.
func (sc StatusContext) Names() []string {
	res := make([]string, len(sc.Packages))
	i := 0
	for n := range sc.Packages {
		res[i] = n
		i++
	}
	sortutil.CiAsc(res)
	return res
}

func (tdep ToolDependency) extractTool(sc StatusContext) (*Tool, error) {
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

func (tdep ToolDependency) extractRelease(sc StatusContext) (*ToolRelease, error) {
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

// CreateStatusContext creates a status context from index data.
func (index Index) CreateStatusContext() StatusContext {
	// Start with an empty status context
	packages := StatusContext{
		Packages: make(map[string]*Package, len(index.Packages)),
	}
	for _, packageManager := range index.Packages {
		packages.Add(packageManager)
	}
	return packages
}

// GetDeps returns the deps of a specified release of a core.
func (sc StatusContext) GetDeps(release *Release) ([]CoreDependency, error) {
	ret := make([]CoreDependency, 0, 5)
	if release == nil {
		return nil, errors.New("release cannot be nil")
	}
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
		ret = append(ret, CoreDependency{
			ToolName: dep.ToolName,
			Release:  toolRelease,
		})
	}
	return ret, nil
}

// Process takes a set of ID tuples and returns
// a set of items to download and a set of outputs for non
// existing cores.
func (sc StatusContext) Process(items []CoreIDTuple) ([]DownloadItem, []DownloadItem, []output.ProcessResult) {
	itemC := len(items)
	retCores := make([]DownloadItem, 0, itemC)
	retTools := make([]DownloadItem, 0, itemC)
	fails := make([]output.ProcessResult, 0, itemC)

	// value is not used, this map is only to check if an item is inside (set implementation)
	// see https://stackoverflow.com/questions/34018908/golang-why-dont-we-have-a-set-datastructure
	presenceMap := make(map[string]bool, itemC)

	for _, item := range items {
		if item.Package == "invalid-arg" {
			fails = append(fails, output.ProcessResult{
				ItemName: item.CoreName,
				Error:    "Invalid item (not PACKAGER:CORE[=VERSION])",
			})
			continue
		}
		pkg, exists := sc.Packages[item.Package]
		if !exists {
			fails = append(fails, output.ProcessResult{
				ItemName: item.CoreName,
				Error:    fmt.Sprintf("Package %s not found", item.Package),
			})
			continue
		}
		core, exists := pkg.Cores[item.CoreName]
		if !exists {
			fails = append(fails, output.ProcessResult{
				ItemName: item.CoreName,
				Error:    "Core not found",
			})
			continue
		}

		_, exists = presenceMap[item.CoreName]
		if exists { //skip
			continue
		}

		release := core.GetVersion(item.CoreVersion)
		if release == nil {
			fails = append(fails, output.ProcessResult{
				ItemName: item.CoreName,
				Error:    fmt.Sprintf("Version %s Not Found", item.CoreVersion),
			})
			continue
		}

		// replaces "latest" with latest version too
		deps, err := sc.GetDeps(release)
		if err != nil {
			fails = append(fails, output.ProcessResult{
				ItemName: item.CoreName,
				Error:    fmt.Sprintf("Cannot get tool dependencies of %s core: %s", core.Name, err.Error()),
			})
			continue
		}

		retCores = append(retCores, DownloadItem{
			Package: pkg.Name,
			DownloadItem: releases.DownloadItem{
				Name:    core.Architecture,
				Release: release,
			},
		})

		presenceMap[core.Name] = true
		for _, tool := range deps {
			_, exists = presenceMap[tool.ToolName]
			if exists { //skip
				continue
			}

			presenceMap[tool.ToolName] = true
			retTools = append(retTools, DownloadItem{
				Package: pkg.Name,
				DownloadItem: releases.DownloadItem{
					Name:    tool.ToolName,
					Release: tool.Release,
				},
			})
		}
	}
	return retCores, retTools, fails
}
