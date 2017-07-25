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
)

// InstalledStuff represents an output set of tools or cores to format output from `core lib list` command.
type InstalledStuff struct {
	Name     string   `json:"name,required"`
	Versions []string `json:"versions,required"`
}

// InstalledPackage represents a single package to format output from `core lib list` command.
type InstalledPackage struct {
	Name           string           `json:"package,required"`
	InstalledCores []InstalledStuff `json:"cores,required"`
	InstalledTools []InstalledStuff `json:"tools,required"`
}

// InstalledPackageList represents an output structure to format output from `core lib list` command.
type InstalledPackageList struct {
	InstalledPackages []InstalledPackage `json:"packages,required"`
}

func (is InstalledStuff) String() string {
	return fmt.Sprintln("  Name:", is.Name) +
		fmt.Sprintln("  Versions:", is.Versions)
}

func (ip InstalledPackage) String() string {
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

func (icl InstalledPackageList) String() string {
	ret := ""
	for _, pkg := range icl.InstalledPackages {
		ret += fmt.Sprintln(pkg)
	}
	return strings.TrimSpace(ret)
}
