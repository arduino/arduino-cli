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

package output

import (
	"fmt"
	"sort"

	"github.com/arduino/arduino-cli/arduino/libraries"
	"github.com/arduino/arduino-cli/arduino/libraries/librariesindex"
	"github.com/gosuri/uitable"
	semver "go.bug.st/relaxed-semver"
)

// InstalledLibraries is a list of installed libraries
type InstalledLibraries struct {
	Libraries []*InstalledLibary `json:"libraries"`
}

// InstalledLibary is an installed library
type InstalledLibary struct {
	Library   *libraries.Library      `json:"library"`
	Available *librariesindex.Release `omitempy,json:"available"`
}

func (il InstalledLibraries) Len() int { return len(il.Libraries) }
func (il InstalledLibraries) Swap(i, j int) {
	il.Libraries[i], il.Libraries[j] = il.Libraries[j], il.Libraries[i]
}
func (il InstalledLibraries) Less(i, j int) bool {
	return il.Libraries[i].Library.String() < il.Libraries[j].Library.String()
}

func (il InstalledLibraries) String() string {
	table := uitable.New()
	table.MaxColWidth = 100
	table.Wrap = true

	hasUpdates := false
	for _, libMeta := range il.Libraries {
		if libMeta.Available != nil {
			hasUpdates = true
		}
	}

	if hasUpdates {
		table.AddRow("Name", "Installed", "Available", "Location")
	} else {
		table.AddRow("Name", "Installed", "Location")
	}
	sort.Sort(il)
	lastName := ""
	for _, libMeta := range il.Libraries {
		lib := libMeta.Library
		name := lib.Name
		if name == lastName {
			name = ` "`
		} else {
			lastName = name
		}

		location := lib.Location.String()
		if lib.ContainerPlatform != nil {
			location = lib.ContainerPlatform.String()
		}
		if hasUpdates {
			var available *semver.Version
			if libMeta.Available != nil {
				available = libMeta.Available.Version
			}
			table.AddRow(name, lib.Version, available, location)
		} else {
			table.AddRow(name, lib.Version, location)
		}
	}
	return fmt.Sprintln(table)
}
