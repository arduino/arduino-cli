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

package librariesindex

import (
	"sort"

	"github.com/arduino/arduino-cli/arduino/libraries"
	"github.com/arduino/arduino-cli/arduino/resources"
	semver "go.bug.st/relaxed-semver"
)

// Index represents the list of libraries available for download
type Index struct {
	Libraries map[string]*Library
}

// EmptyIndex is an empty library index
var EmptyIndex = &Index{Libraries: map[string]*Library{}}

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
// release is not found. If the version is not specified returns the latest
// version available.
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
