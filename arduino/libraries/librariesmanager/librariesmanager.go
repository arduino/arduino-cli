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
	"os"

	paths "github.com/arduino/go-paths-helper"
	"github.com/bcmi-labs/arduino-cli/arduino/cores"
	"github.com/bcmi-labs/arduino-cli/arduino/libraries/librariesindex"
	"github.com/sirupsen/logrus"

	"github.com/bcmi-labs/arduino-cli/arduino/libraries"
	"github.com/pmylund/sortutil"
)

// LibrariesManager keeps the current status of the libraries in the system
// (the list of libraries, revisions, installed paths, etc.)
type LibrariesManager struct {
	LibrariesDir []*LibrariesDir
	Libraries    map[string]*LibraryAlternatives `json:"libraries"`

	Index        *librariesindex.Index
	IndexFile    *paths.Path
	DownloadsDir *paths.Path
}

// LibrariesDir is a directory containing libraries
type LibrariesDir struct {
	Path            *paths.Path
	Location        libraries.LibraryLocation
	PlatformRelease *cores.PlatformRelease
}

// LibraryAlternatives is a list of different versions of the same library
// installed in the system
type LibraryAlternatives struct {
	Alternatives libraries.List
}

// Add adds a library to the alternatives
func (alts *LibraryAlternatives) Add(library *libraries.Library) {
	if len(alts.Alternatives) > 0 && alts.Alternatives[0].Name != library.Name {
		panic(fmt.Sprintf("the library name is different from the set (%s != %s)", alts.Alternatives[0].Name, library.Name))
	}
	alts.Alternatives = append(alts.Alternatives, library)
}

// Remove removes the library from the alternatives
func (alts *LibraryAlternatives) Remove(library *libraries.Library) {
	for i, lib := range alts.Alternatives {
		if lib == library {
			alts.Alternatives = append(alts.Alternatives[:i], alts.Alternatives[i+1:]...)
			return
		}
	}
}

// FindVersion returns the library mathching the provided version or nil if not found
func (alts *LibraryAlternatives) FindVersion(version string) *libraries.Library {
	for _, lib := range alts.Alternatives {
		if lib.Version == version {
			return lib
		}
	}
	return nil
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
func NewLibraryManager(indexDir *paths.Path, downloadsDir *paths.Path) *LibrariesManager {
	var indexFile *paths.Path
	if indexDir != nil {
		indexFile = indexDir.Join("library_index.json")
	}
	return &LibrariesManager{
		Libraries:    map[string]*LibraryAlternatives{},
		IndexFile:    indexFile,
		DownloadsDir: downloadsDir,
	}
}

// LoadIndex reads a library_index.json from a file and returns
// the corresponding Index structure.
func (sc *LibrariesManager) LoadIndex() error {
	index, err := librariesindex.LoadIndex(sc.IndexFile)
	sc.Index = index
	return err
}

// AddLibrariesDir adds path to the list of directories
// to scan when searching for libraries. If a path is already
// in the list it is ignored.
func (sc *LibrariesManager) AddLibrariesDir(path *paths.Path, location libraries.LibraryLocation) {
	for _, dir := range sc.LibrariesDir {
		if dir.Path.EquivalentTo(path) {
			return
		}
	}
	logrus.WithField("dir", path).WithField("location", location.String()).Info("Adding libraries dir")
	sc.LibrariesDir = append(sc.LibrariesDir, &LibrariesDir{
		Path:     path,
		Location: location,
	})
}

// AddPlatformReleaseLibrariesDir add the libraries folder in the
// specified PlatformRelease to the list of directories to scan when
// searching for libraries.
func (sc *LibrariesManager) AddPlatformReleaseLibrariesDir(plaftormRelease *cores.PlatformRelease, location libraries.LibraryLocation) {
	path := plaftormRelease.GetLibrariesDir()
	if path == nil {
		return
	}
	for _, dir := range sc.LibrariesDir {
		if dir.Path.EquivalentTo(path) {
			return
		}
	}
	logrus.WithField("dir", path).WithField("location", location.String()).Info("Adding libraries dir")
	sc.LibrariesDir = append(sc.LibrariesDir, &LibrariesDir{
		Path:            path,
		Location:        location,
		PlatformRelease: plaftormRelease,
	})
}

// RescanLibraries reload all installed libraries in the system.
func (sc *LibrariesManager) RescanLibraries() error {
	for _, dir := range sc.LibrariesDir {
		if err := sc.LoadLibrariesFromDir(dir); err != nil {
			return fmt.Errorf("loading libs from %s: %s", dir.Path, err)
		}
	}
	return nil
}

func (sc *LibrariesManager) getSketchbookLibrariesDir() *paths.Path {
	for _, dir := range sc.LibrariesDir {
		if dir.Location == libraries.Sketchbook {
			return dir.Path
		}
	}
	return nil
}

// LoadLibrariesFromDir loads all libraries in the given directory. Returns
// nil if the directory doesn't exists.
func (sc *LibrariesManager) LoadLibrariesFromDir(librariesDir *LibrariesDir) error {
	subFolders, err := librariesDir.Path.ReadDir()
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("reading dir %s: %s", librariesDir.Path, err)
	}
	subFolders.FilterDirs()
	subFolders.FilterOutHiddenFiles()

	for _, subFolder := range subFolders {
		library, err := libraries.Load(subFolder, librariesDir.Location)
		if err != nil {
			return fmt.Errorf("loading library from %s: %s", subFolder, err)
		}
		library.ContainerPlatform = librariesDir.PlatformRelease
		alternatives, ok := sc.Libraries[library.Name]
		if !ok {
			alternatives = &LibraryAlternatives{}
			sc.Libraries[library.Name] = alternatives
		}
		alternatives.Add(library)
	}
	return nil
}

// FindByReference return the installed library matching the Reference
// name and version or if the version is the empty string the library
// installed in the sketchbook.
func (sc *LibrariesManager) FindByReference(libRef *librariesindex.Reference) *libraries.Library {
	alternatives, have := sc.Libraries[libRef.Name]
	if !have {
		return nil
	}
	if libRef.Version == "" {
		for _, candidate := range alternatives.Alternatives {
			if candidate.Location == libraries.Sketchbook {
				return candidate
			}
		}
		return nil
	}
	return alternatives.FindVersion(libRef.Version)
}
