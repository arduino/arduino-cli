/*
 * This file is part of arduino-cli.
 *
 * Copyright 2018 ARDUINO AG (http://www.arduino.cc/)
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
 */

package librariesmanager

import (
	"github.com/bcmi-labs/arduino-cli/arduino/libraries/librariesindex"

	"github.com/bcmi-labs/arduino-cli/arduino/libraries"
	"github.com/pmylund/sortutil"
)

// StatusContext keeps the current status of the libraries in the system
// (the list of libraries, revisions, installed paths, etc.)
type StatusContext struct {
	Libraries map[string]*libraries.Library `json:"libraries"`
	Index     *librariesindex.Index
}

// // AddLibrary adds an indexRelease to the status context
// func (sc *StatusContext) AddLibrary(indexLib *indexRelease) {
// 	name := indexLib.Name
// 	if sc.Libraries[name] == nil {
// 		sc.Libraries[name] = indexLib.extractLibrary()
// 	} else {
// 		release := indexLib.extractRelease()
// 		lib := sc.Libraries[name]
// 		lib.Releases[fmt.Sprint(release.Version)] = release
// 		release.Library = lib
// 	}
// }

// Names returns an array with all the names of the registered libraries.
func (sc StatusContext) Names() []string {
	res := make([]string, len(sc.Libraries))
	i := 0
	for n := range sc.Libraries {
		res[i] = n
		i++
	}
	sortutil.CiAsc(res)
	return res
}

// NewLibraryManager creates a new library manager
func NewLibraryManager() *StatusContext {
	return &StatusContext{
		Libraries: map[string]*libraries.Library{},
	}
}

// LoadIndex reads a library_index.json from a file and returns
// the corresponding Index structure.
func (sc *StatusContext) LoadIndex() error {
	index, err := librariesindex.LoadIndex(IndexPath())
	sc.Index = index
	return err
}
