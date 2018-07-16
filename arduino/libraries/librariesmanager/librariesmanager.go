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
	"fmt"

	paths "github.com/arduino/go-paths-helper"
	"github.com/bcmi-labs/arduino-cli/arduino/libraries/librariesindex"

	"github.com/bcmi-labs/arduino-cli/arduino/libraries"
	"github.com/pmylund/sortutil"
)

// LibrariesManager keeps the current status of the libraries in the system
// (the list of libraries, revisions, installed paths, etc.)
type LibrariesManager struct {
	Libraries map[string]*LibraryAlternatives `json:"libraries"`
	Index     *librariesindex.Index
}

// LibraryAlternatives is a list of different versions of the same library
// installed in the system
type LibraryAlternatives struct {
	Alternatives []*libraries.Library
}

// Add adds a library to the alternatives
func (alts *LibraryAlternatives) Add(library *libraries.Library) {
	if len(alts.Alternatives) > 0 && alts.Alternatives[0].Name != library.Name {
		panic(fmt.Sprintf("the library name is different from the set (%s != %s)", alts.Alternatives[0].Name, library.Name))
	}
	alts.Alternatives = append(alts.Alternatives, library)
}

// Select returns the library with the highest priority between the alternatives
func (alts *LibraryAlternatives) Select() *libraries.Library {
	// TODO
	return alts.Alternatives[len(alts.Alternatives)-1]
}

// Names returns an array with all the names of the installed libraries.
func (sc LibrariesManager) Names() []string {
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
func NewLibraryManager() *LibrariesManager {
	return &LibrariesManager{
		Libraries: map[string]*LibraryAlternatives{},
	}
}

// LoadIndex reads a library_index.json from a file and returns
// the corresponding Index structure.
func (sc *LibrariesManager) LoadIndex() error {
	index, err := librariesindex.LoadIndex(IndexPath())
	sc.Index = index
	return err
}

// LoadLibrariesFromDir loads all libraries in the given folder
func (sc *LibrariesManager) LoadLibrariesFromDir(librariesDir *LibrariesDir) error {
	subFolders, err := librariesDir.ReadDir()
	if err != nil {
		return fmt.Errorf("reading dir %s: %s", librariesDir, err)
	}
	subFolders.FilterDirs()
	subFolders.FilterOutHiddenFiles()

	for _, subFolder := range subFolders {
		library, err := libraries.Load(subFolder)
		if err != nil {
			return fmt.Errorf("loading library from %s: %s", subFolder, err)
		}
		alternatives, ok := sc.Libraries[library.Name]
		if !ok {
			alternatives = &LibraryAlternatives{}
			sc.Libraries[library.Name] = alternatives
		}
		alternatives.Add(library)
	}
	return nil
}
