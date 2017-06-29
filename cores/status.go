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

package cores

import "github.com/pmylund/sortutil"

// StatusContext represents the Context of the Cores and Tools in the system.
type StatusContext struct {
	Packages map[string]*Package
}

//Package represents a package in the system.
type Package struct {
	Name       string
	Maintainer string
	WebsiteURL string
	Email      string
	Cores      map[string]*Core // The cores in the system.
	// Tools map[string]*Tool // The tools in the system.
}

// AddPackage adds a package to a context from an indexPackage.
//
// NOTE: If the package is already in the context, it is overwritten!
func (sc *StatusContext) AddPackage(indexPackage *indexPackage) {
	sc.Packages[indexPackage.Name] = indexPackage.extractPackage()
}

// AddCore adds a core to the context.
func (pm *Package) AddCore(indexCore *indexCoreRelease) {
	name := indexCore.Name
	if pm.Cores[name] == nil {
		pm.Cores[name] = indexCore.extractCore()
	} else {
		release := indexCore.extractRelease()
		core := pm.Cores[name]
		core.Releases[release.Version] = release
	}
}

// CoreNames returns an array with all the names of the registered cores.
func (pm *Package) CoreNames() []string {
	res := make([]string, len(pm.Cores))
	i := 0
	for n := range pm.Cores {
		res[i] = n
		i++
	}
	sortutil.CiAsc(res)
	return res
}

// CreateStatusContextFromIndex creates a status context from index data.
func CreateStatusContextFromIndex(index *Index) (*StatusContext, error) {
	// Start with an empty status context
	packages := StatusContext{
		Packages: make(map[string]*Package, len(index.Packages)),
	}
	for _, packageManager := range index.Packages {
		packages.AddPackage(packageManager)
	}
	return &packages, nil
}
