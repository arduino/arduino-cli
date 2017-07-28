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

//Package represents a package in the system.
type Package struct {
	Name       string           // Name of the package.
	Maintainer string           // Name of the maintainer.
	WebsiteURL string           // Website of maintainer.
	Email      string           // Email of maintainer.
	Cores      map[string]*Core // The cores in the system.
	Tools      map[string]*Tool // The tools in the system.
}

// addCore adds a core to the context.
func (pm *Package) addCore(indexCore *indexCoreRelease) {
	name := indexCore.Name
	if pm.Cores[name] == nil {
		pm.Cores[name] = indexCore.extractCore()
	} else {
		release := indexCore.extractRelease()
		core := pm.Cores[name]
		core.Releases[release.Version] = release
	}
}

// addTool adds a tool to the context.
func (pm *Package) addTool(indexTool *indexToolRelease) {
	name := indexTool.Name
	if pm.Tools[name] == nil {
		pm.Tools[name] = indexTool.extractTool()
	} else {
		pm.Tools[name].Releases[indexTool.Version] = indexTool.extractRelease()
	}
}
