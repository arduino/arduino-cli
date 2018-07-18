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

package libraries

import (
	"sort"
)

// List is a list of Libraries
type List []*Library

// Contains check if a lib is contained in the list
func (list *List) Contains(lib *Library) bool {
	for _, l := range *list {
		if l == lib {
			return true
		}
	}
	return false
}

// Add appends all libraries passed as parameter in the list
func (list *List) Add(libs ...*Library) {
	for _, lib := range libs {
		*list = append(*list, lib)
	}
}

// FindByName returns the first library in the list that match
// the specified name or nil if not found
func (list *List) FindByName(name string) *Library {
	for _, lib := range *list {
		if lib.Name == name {
			return lib
		}
	}
	return nil
}

// SortByArchitecturePriority sorts the libraries in descending order using
// the Arduino lib priority ordering (the first has the higher priority)
func (list *List) SortByArchitecturePriority(arch string) {
	sort.Slice(*list, func(i, j int) bool {
		a, b := (*list)[i], (*list)[j]
		return a.PriorityForArchitecture(arch) > b.PriorityForArchitecture(arch)
	})
}

/*
// HasHigherPriority returns true if library x has higher priority compared to library
// y for the given header and architecture.
func HasHigherPriority(libX, libY *Library, header string, arch string) bool {
	//return computePriority(libX, header, arch) > computePriority(libY, header, arch)
	header = strings.TrimSuffix(header, filepath.Ext(header))

	simplify := func(name string) string {
		name = utils.SanitizeName(name)
		name = strings.ToLower(name)
		return name
	}
	header = simplify(header)
	nameX := simplify(libX.Name)
	nameY := simplify(libY.Name)

	compareLocations := func() bool {
		// XXX: priority inversion case.
		if libX.Location < libY.Location {
			return true
		}
		return false
	}

	checks := []func(name, header string) bool{
		func(name, header string) bool { return name == header },
		func(name, header string) bool { return name == header+"-master" },
		strings.HasPrefix,
		strings.HasSuffix,
		strings.Contains,
	}
	// Run all checks to sort priorities based on library name
	// If both library match the same name check, then fallback to
	// compare locations
	for _, check := range checks {
		x := check(nameX, header)
		y := check(nameY, header)
		if x && y {
			return compareLocations()
		}
		if x {
			return true
		}
		if y {
			return false
		}
	}

	return compareLocations()
}
*/
