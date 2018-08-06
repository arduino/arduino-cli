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

package librariesindex

import (
	"sort"

	"github.com/bcmi-labs/arduino-cli/arduino/libraries"
	"github.com/bcmi-labs/arduino-cli/arduino/resources"
	"go.bug.st/relaxed-semver"
)

// Index represents the list of libraries available for download
type Index struct {
	Libraries map[string]*Library
}

// Library is a library available for download
type Library struct {
	Name     string
	Releases map[string]*Release
	Latest   *Release `json:"-"`
	Index    *Index   `json:"-"`
}

// Release is a release of a library available for download
type Release struct {
	Author        string
	Version       *semver.Version
	Maintainer    string
	Sentence      string
	Paragraph     string
	Website       string
	Category      string
	Architectures []string
	Types         []string
	Resource      *resources.DownloadResource

	Library *Library `json:"-"`
}

func (r *Release) String() string {
	return r.Library.Name + "@" + r.Version.String()
}

// FindRelease search a library Release in the index. Returns nil if the
// release is not found
func (idx *Index) FindRelease(ref *Reference) *Release {
	if library, exists := idx.Libraries[ref.Name]; exists {
		if ref.Version == nil {
			return library.Latest
		}
		return library.Releases[ref.Version.String()]
	}
	return nil
}

// FindIndexedLibrary search an indexed library that matches the provided
// installed library or nil if not found
func (idx *Index) FindIndexedLibrary(lib *libraries.Library) *Library {
	return idx.Libraries[lib.Name]
}

// FindLibraryUpdate check if an installed library may be updated using
// one of the indexed libraries. This function returns the Release to install
// to update the library if found, otherwise nil is returned.
func (idx *Index) FindLibraryUpdate(lib *libraries.Library) *Release {
	indexLib := idx.FindIndexedLibrary(lib)
	if indexLib == nil {
		return nil
	}
	if indexLib.Latest.Version.GreaterThan(lib.Version) {
		return indexLib.Latest
	}
	return nil
}

// Versions returns an array of all versions available of the library
func (library *Library) Versions() []*semver.Version {
	res := []*semver.Version{}
	for version := range library.Releases {
		v, err := semver.Parse(version)
		if err == nil {
			res = append(res, v)
		}
	}
	sort.Sort(semver.List(res))
	return res
}
