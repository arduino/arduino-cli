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
	"errors"
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/arduino/arduino-cli/internal/arduino/cores"
	"github.com/arduino/arduino-cli/internal/arduino/libraries"
	"github.com/arduino/arduino-cli/internal/arduino/libraries/librariesindex"
	"github.com/arduino/arduino-cli/internal/i18n"
	paths "github.com/arduino/go-paths-helper"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// LibrariesManager keeps the current status of the libraries in the system
// (the list of libraries, revisions, installed paths, etc.)
type LibrariesManager struct {
	LibrariesDir []*LibrariesDir
	Libraries    map[string]libraries.List `json:"libraries"`

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

var tr = i18n.Tr

// Names returns an array with all the names of the installed libraries.
func (lm LibrariesManager) Names() []string {
	res := make([]string, len(lm.Libraries))
	i := 0
	for n := range lm.Libraries {
		res[i] = n
		i++
	}
	slices.SortFunc(res, func(a, b string) int {
		if strings.ToLower(a) < strings.ToLower(b) {
			return -1
		}
		return 1
	})
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
		Libraries:          map[string]libraries.List{},
		IndexFile:          indexFile,
		IndexFileSignature: indexFileSignature,
		DownloadsDir:       downloadsDir,
		Index:              librariesindex.EmptyIndex,
	}
}

// LoadIndex reads a library_index.json from a file and returns
// the corresponding Index structure.
func (lm *LibrariesManager) LoadIndex() error {
	logrus.WithField("index", lm.IndexFile).Info("Loading libraries index file")
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
	lm.clearLibraries()
	statuses := []*status.Status{}
	for _, dir := range lm.LibrariesDir {
		if errs := lm.LoadLibrariesFromDir(dir); len(errs) > 0 {
			statuses = append(statuses, errs...)
		}
	}
	return statuses
}

func (lm *LibrariesManager) getLibrariesDir(installLocation libraries.LibraryLocation) (*paths.Path, error) {
	for _, dir := range lm.LibrariesDir {
		if dir.Location == installLocation {
			return dir.Path, nil
		}
	}
	switch installLocation {
	case libraries.User:
		return nil, errors.New(tr("user directory not set"))
	case libraries.IDEBuiltIn:
		return nil, errors.New(tr("built-in libraries directory not set"))
	default:
		return nil, fmt.Errorf("libraries directory not set: %s", installLocation.String())
	}
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
		s := status.Newf(codes.FailedPrecondition, tr("reading dir %[1]s: %[2]s"), librariesDir.Path, err)
		return append(statuses, s)
	}
	subDirs.FilterDirs()
	subDirs.FilterOutHiddenFiles()

	for _, subDir := range subDirs {
		library, err := libraries.Load(subDir, librariesDir.Location)
		if err != nil {
			s := status.Newf(codes.Internal, tr("loading library from %[1]s: %[2]s"), subDir, err)
			statuses = append(statuses, s)
			continue
		}
		library.ContainerPlatform = librariesDir.PlatformRelease
		alternatives := lm.Libraries[library.Name]
		alternatives.Add(library)
		lm.Libraries[library.Name] = alternatives
	}

	return statuses
}

// LoadLibraryFromDir loads one single library from the libRootDir.
// libRootDir must point to the root of a valid library.
// An error is returned if the path doesn't exist or loading of the library fails.
func (lm *LibrariesManager) LoadLibraryFromDir(libRootDir *paths.Path, location libraries.LibraryLocation) error {
	if libRootDir.NotExist() {
		return fmt.Errorf(tr("library path does not exist: %s"), libRootDir)
	}

	library, err := libraries.Load(libRootDir, location)
	if err != nil {
		return fmt.Errorf(tr("loading library from %[1]s: %[2]s"), libRootDir, err)
	}

	alternatives := lm.Libraries[library.Name]
	alternatives.Add(library)
	lm.Libraries[library.Name] = alternatives
	return nil
}

// FindByReference return the installed libraries matching the Reference
// name and version or, if the version is nil, the libraries installed
// in the installLocation.
func (lm *LibrariesManager) FindByReference(libRef *librariesindex.Reference, installLocation libraries.LibraryLocation) libraries.List {
	alternatives := lm.Libraries[libRef.Name]
	if alternatives == nil {
		return nil
	}
	return alternatives.FilterByVersionAndInstallLocation(libRef.Version, installLocation)
}

// FindAllInstalled returns all the installed libraries
func (lm *LibrariesManager) FindAllInstalled() libraries.List {
	var res libraries.List
	for _, libAlternatives := range lm.Libraries {
		for _, libRelease := range libAlternatives {
			if libRelease.InstallDir == nil {
				continue
			}
			res.Add(libRelease)
		}
	}
	return res
}

func (lm *LibrariesManager) clearLibraries() {
	for k := range lm.Libraries {
		delete(lm.Libraries, k)
	}
}
