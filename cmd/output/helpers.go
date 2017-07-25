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
 * Copyright 2017 BCMI LABS SA (http://www.arduino.cc/)
 */

package output

import (
	"fmt"
	"strings"

	"github.com/bcmi-labs/arduino-cli/cmd/formatter"
)

// LibResultsFromMap returns a LibProcessResults struct from a specified map of results.
func LibResultsFromMap(resMap map[string]interface{}) LibProcessResults {
	results := LibProcessResults{
		Libraries: make([]libProcessResult, len(resMap)),
	}
	i := 0
	for libName, libResult := range resMap {
		_, isError := libResult.(error)
		if isError {
			results.Libraries[i] = libProcessResult{
				LibraryName: libName,
				Status:      "",
				Error:       fmt.Sprint(libResult),
			}
		} else {
			results.Libraries[i] = libProcessResult{
				LibraryName: libName,
				Status:      fmt.Sprint(libResult),
				Error:       "",
			}
		}
		i++
	}
	return results
}

type installedStuff struct {
	Name     string   `json:"name,required"`
	Versions []string `json:"versions,required"`
}

func (is installedStuff) String() string {
	return fmt.Sprintln("  Name:", is.Name) +
		fmt.Sprintln("  Versions:", is.Versions)
}

type installedPackage struct {
	Name           string           `json:"package,required"`
	InstalledCores []installedStuff `json:"cores,required"`
	InstalledTools []installedStuff `json:"tools,required"`
}

func (ip installedPackage) String() string {
	ret := ""
	thereAreCores := len(ip.InstalledCores) > 0
	thereAreTools := len(ip.InstalledTools) > 0
	if thereAreCores || thereAreTools {
		ret += fmt.Sprintln("Package", ip.Name)
	}
	if thereAreCores {
		ret += fmt.Sprintln("Cores:")
		for _, core := range ip.InstalledCores {
			ret += fmt.Sprintln(core)
		}
	}
	if thereAreTools {
		ret += fmt.Sprintln("Tools:")
		for _, tool := range ip.InstalledTools {
			ret += fmt.Sprintln(tool)
		}
	}
	return strings.TrimSpace(ret)
}

type installedPackageList struct {
	InstalledPackages []installedPackage `json:"packages,required"`
}

func (icl installedPackageList) String() string {
	ret := ""
	for _, pkg := range icl.InstalledPackages {
		ret += fmt.Sprintln(pkg)
	}
	return strings.TrimSpace(ret)
}

// InstalledCoresToolsFromMaps pretty prints a list of ALREADY INSTALLED cores and tools.
func InstalledCoresToolsFromMaps(coreMap map[string]map[string][]string, toolMap map[string]map[string][]string) {
	packages := installedPackageList{
		InstalledPackages: make([]installedPackage, len(coreMap)),
	}

	var i, j int
	i = 0
	for pkg, coresData := range coreMap {
		j = 0
		packages.InstalledPackages[i] = installedPackage{
			Name:           pkg,
			InstalledCores: make([]installedStuff, len(coresData)),
		}
		for core, versions := range coresData {
			packages.InstalledPackages[i].InstalledCores[j] = installedStuff{
				Name:     core,
				Versions: versions,
			}
		}
		j++
		i++
	}
	for i := range packages.InstalledPackages {
		toolData := toolMap[packages.InstalledPackages[i].Name]
		packages.InstalledPackages[i].InstalledTools = make([]installedStuff, len(toolData))
		j = 0
		for tool, versions := range toolData {
			packages.InstalledPackages[i].InstalledTools[j] = installedStuff{
				Name:     tool,
				Versions: versions,
			}
			j++
		}
	}
	formatter.Print(packages)
}
