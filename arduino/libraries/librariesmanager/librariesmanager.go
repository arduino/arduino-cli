// This file is part of arduino-cli.
//
// Copyright 2020 ARDUINO SA (http://www.arduino.cc/)
//
// This software is released under the GNU General Public License version 3,
// which covers the main part of arduino-cli.
// The terms of this license can be found at:
// https://www.gnu.org/licenses/gpl-3.0.en.html
//
// You can be released from the requirements of the above licenses by purchasing
// a commercial license. Buying such a license is mandatory if you want to
// modify or otherwise use the software for commercial activities involving the
// Arduino software without disclosing the source code of your own applications.
// To purchase a commercial license, send an email to license@arduino.cc.

package librariesmanager

import (
	"fmt"
	"os"

	"github.com/arduino/arduino-cli/arduino/cores"
	"github.com/arduino/arduino-cli/arduino/libraries"
	"github.com/arduino/arduino-cli/arduino/libraries/librariesindex"
	"github.com/arduino/arduino-cli/arduino/utils"
	paths "github.com/arduino/go-paths-helper"
	"github.com/pmylund/sortutil"
	"github.com/sirupsen/logrus"
	semver "go.bug.st/relaxed-semver"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// LibrariesManager keeps the current status of the libraries in the system
// (the list of libraries, revisions, installed paths, etc.)
type LibrariesManager struct {
	LibrariesDir []*LibrariesDir
	Libraries    map[string]*LibraryAlternatives `json:"libraries"`

	Index              *librariesindex.Index
	IndexFile          *paths.Path
	IndexFileSignature *paths.Path
	DownloadsDir       *paths.Path
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
func (alts *LibraryAlternatives) FindVersion(version *semver.Version) *libraries.Library {
	for _, lib := range alts.Alternatives {
		if lib.Version.Equal(version) {
			return lib
		}
	}
	return nil
}

// Names returns an array with all the names of the installed libraries.
func (lm LibrariesManager) Names() []string {
	res := make([]string, len(lm.Libraries))
	i := 0
	for n := range lm.Libraries {
		res[i] = n
		i++
	}
	sortutil.CiAsc(res)
	return res
}

// NewLibraryManager creates a new library manager
func NewLibraryManager(indexDir *paths.Path, downloadsDir *paths.Path) *LibrariesManager {
	var indexFile, indexFileSignature *paths.Path
	if indexDir != nil {
		indexFile = indexDir.Join("library_index.json")
		indexFileSignature = indexDir.Join("library_index.json.sig")
	}
	return &LibrariesManager{
		Libraries:          map[string]*LibraryAlternatives{},
		IndexFile:          indexFile,
		IndexFileSignature: indexFileSignature,
		DownloadsDir:       downloadsDir,
		Index:              librariesindex.EmptyIndex,
	}
}

// LoadIndex reads a library_index.json from a file and returns
// the corresponding Index structure.
func (lm *LibrariesManager) LoadIndex() error {
	index, err := librariesindex.LoadIndex(lm.IndexFile)
	if err != nil {
		lm.Index = librariesindex.EmptyIndex
		return err
	}
	lm.Index = index
	return nil
}

// AddLibrariesDir adds path to the list of directories
// to scan when searching for libraries. If a path is already
// in the list it is ignored.
func (lm *LibrariesManager) AddLibrariesDir(path *paths.Path, location libraries.LibraryLocation) {
	for _, dir := range lm.LibrariesDir {
		if dir.Path.EquivalentTo(path) {
			return
		}
	}
	logrus.WithField("dir", path).WithField("location", location.String()).Info("Adding libraries dir")
	lm.LibrariesDir = append(lm.LibrariesDir, &LibrariesDir{
		Path:     path,
		Location: location,
	})
}

// AddPlatformReleaseLibrariesDir add the libraries directory in the
// specified PlatformRelease to the list of directories to scan when
// searching for libraries.
func (lm *LibrariesManager) AddPlatformReleaseLibrariesDir(plaftormRelease *cores.PlatformRelease, location libraries.LibraryLocation) {
	path := plaftormRelease.GetLibrariesDir()
	if path == nil {
		return
	}
	for _, dir := range lm.LibrariesDir {
		if dir.Path.EquivalentTo(path) {
			return
		}
	}
	logrus.WithField("dir", path).WithField("location", location.String()).Info("Adding libraries dir")
	lm.LibrariesDir = append(lm.LibrariesDir, &LibrariesDir{
		Path:            path,
		Location:        location,
		PlatformRelease: plaftormRelease,
	})
}

// RescanLibraries reload all installed libraries in the system.
func (lm *LibrariesManager) RescanLibraries() []*status.Status {
	statuses := []*status.Status{}
	for _, dir := range lm.LibrariesDir {
		if errs := lm.LoadLibrariesFromDir(dir); len(errs) > 0 {
			statuses = append(statuses, errs...)
		}
	}
	return statuses
}

func (lm *LibrariesManager) getUserLibrariesDir() *paths.Path {
	for _, dir := range lm.LibrariesDir {
		if dir.Location == libraries.User {
			return dir.Path
		}
	}
	return nil
}

// LoadLibrariesFromDir loads all libraries in the given directory. Returns
// nil if the directory doesn't exists.
func (lm *LibrariesManager) LoadLibrariesFromDir(librariesDir *LibrariesDir) []*status.Status {
	statuses := []*status.Status{}
	subDirs, err := librariesDir.Path.ReadDir()
	if os.IsNotExist(err) {
		return statuses
	}
	if err != nil {
		s := status.Newf(codes.FailedPrecondition, "reading dir %s: %s", librariesDir.Path, err)
		return append(statuses, s)
	}
	subDirs.FilterDirs()
	subDirs.FilterOutHiddenFiles()

	for _, subDir := range subDirs {
		library, err := libraries.Load(subDir, librariesDir.Location)
		if err != nil {
			s := status.Newf(codes.Internal, "loading library from %s: %s", subDir, err)
			statuses = append(statuses, s)
			continue
		}
		library.ContainerPlatform = librariesDir.PlatformRelease
		alternatives, ok := lm.Libraries[library.Name]
		if !ok {
			alternatives = &LibraryAlternatives{}
			lm.Libraries[library.Name] = alternatives
		}
		alternatives.Add(library)
	}

	return statuses
}

// LoadLibraryFromDir loads one single library from the libRootDir.
// libRootDir must point to the root of a valid library.
// An error is returned if the path doesn't exist or loading of the library fails.
func (lm *LibrariesManager) LoadLibraryFromDir(libRootDir *paths.Path, location libraries.LibraryLocation) error {
	if libRootDir.NotExist() {
		return fmt.Errorf("library path does not exist: %s", libRootDir)
	}

	library, err := libraries.Load(libRootDir, location)
	if err != nil {
		return fmt.Errorf("loading library from %s: %s", libRootDir, err)
	}

	alternatives, ok := lm.Libraries[library.Name]
	if !ok {
		alternatives = &LibraryAlternatives{}
		lm.Libraries[library.Name] = alternatives
	}
	alternatives.Add(library)

	return nil
}

// FindByReference return the installed library matching the Reference
// name and version or, if the version is nil, the library installed
// in the User folder.
func (lm *LibrariesManager) FindByReference(libRef *librariesindex.Reference) *libraries.Library {
	saneName := utils.SanitizeName(libRef.Name)
	alternatives, have := lm.Libraries[saneName]
	if !have {
		return nil
	}
	// TODO: Move "search into user" into another method...
	if libRef.Version == nil {
		for _, candidate := range alternatives.Alternatives {
			if candidate.Location == libraries.User {
				return candidate
			}
		}
		return nil
	}
	return alternatives.FindVersion(libRef.Version)
}
