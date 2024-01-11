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
	"sync"

	"github.com/arduino/arduino-cli/internal/arduino/cores"
	"github.com/arduino/arduino-cli/internal/arduino/libraries"
	"github.com/arduino/arduino-cli/internal/i18n"
	paths "github.com/arduino/go-paths-helper"
	"github.com/sirupsen/logrus"
	semver "go.bug.st/relaxed-semver"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Builder is used to create a new LibrariesManager. The builder has
// methods to load and parse libraries and to actually build the
// LibraryManager. Once the LibrariesManager is built, it cannot be
// altered anymore.
type Builder struct {
	*LibrariesManager
}

// Explorer is used to query the library manager about the installed libraries.
type Explorer struct {
	*LibrariesManager
}

// Installer is used to rescan installed libraries after install/uninstall.
type Installer struct {
	*LibrariesManager
}

// LibrariesManager keeps the current status of the libraries in the system
// (the list of libraries, revisions, installed paths, etc.)
type LibrariesManager struct {
	librariesLock sync.RWMutex
	librariesDir  []*LibrariesDir
	libraries     map[string]libraries.List
}

// LibrariesDir is a directory containing libraries
type LibrariesDir struct {
	Path            *paths.Path
	Location        libraries.LibraryLocation
	PlatformRelease *cores.PlatformRelease
	IsSingleLibrary bool // true if Path points directly to a library instad of a dir of libraries
	scanned         bool
}

var tr = i18n.Tr

// Names returns an array with all the names of the installed libraries.
func (lm *Explorer) Names() []string {
	res := make([]string, len(lm.libraries))
	i := 0
	for n := range lm.libraries {
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

// NewBuilder creates a new library manager builder.
func NewBuilder() *Builder {
	return &Builder{
		&LibrariesManager{
			libraries: map[string]libraries.List{},
		},
	}
}

// Clone creates a Builder starting with a copy of the same configuration
// of this LibrariesManager. At the moment of the Build() only the added
// libraries directories will be scanned, keeping the existing directories
// "cached" to optimize scan. If you need to do a full rescan you must use
// the RescanLibraries method of the Installer.
func (lm *LibrariesManager) Clone() *Builder {
	lmb := NewBuilder()
	lmb.librariesDir = append(lmb.librariesDir, lm.librariesDir...)
	for libName, libAlternatives := range lm.libraries {
		// TODO: Maybe we should deep clone libAlternatives...
		lmb.libraries[libName] = append(lmb.libraries[libName], libAlternatives...)
	}
	return lmb
}

// NewExplorer returns a new Explorer. The returned function must be called
// to release the lock on the LibrariesManager.
func (lm *LibrariesManager) NewExplorer() (*Explorer, func()) {
	lm.librariesLock.RLock()
	return &Explorer{lm}, lm.librariesLock.RUnlock
}

// NewInstaller returns a new Installer. The returned function must be called
// to release the lock on the LibrariesManager.
func (lm *LibrariesManager) NewInstaller() (*Installer, func()) {
	lm.librariesLock.Lock()
	return &Installer{lm}, lm.librariesLock.Unlock
}

// Build builds a new LibrariesManager.
func (lmb *Builder) Build() (*LibrariesManager, []*status.Status) {
	var statuses []*status.Status
	res := &LibrariesManager{}
	for _, dir := range lmb.librariesDir {
		if !dir.scanned {
			if errs := lmb.loadLibrariesFromDir(dir); len(errs) > 0 {
				statuses = append(statuses, errs...)
			}
		}
	}
	lmb.BuildIntoExistingLibrariesManager(res)
	return res, statuses
}

// BuildIntoExistingLibrariesManager will overwrite the given LibrariesManager instead
// of building a new one.
func (lmb *Builder) BuildIntoExistingLibrariesManager(old *LibrariesManager) {
	old.librariesLock.Lock()
	old.librariesDir = lmb.librariesDir
	old.libraries = lmb.libraries
	old.librariesLock.Unlock()
}

// AddLibrariesDir adds path to the list of directories
// to scan when searching for libraries. If a path is already
// in the list it is ignored.
func (lmb *Builder) AddLibrariesDir(libDir LibrariesDir) {
	if libDir.Path == nil {
		return
	}
	for _, dir := range lmb.librariesDir {
		if dir.Path.EquivalentTo(libDir.Path) {
			return
		}
	}
	logrus.WithField("dir", libDir.Path).
		WithField("location", libDir.Location.String()).
		WithField("isSingleLibrary", libDir.IsSingleLibrary).
		Info("Adding libraries dir")
	lmb.librariesDir = append(lmb.librariesDir, &libDir)
}

// RescanLibraries reload all installed libraries in the system.
func (lmi *Installer) RescanLibraries() []*status.Status {
	lmi.libraries = map[string]libraries.List{}
	statuses := []*status.Status{}
	for _, dir := range lmi.librariesDir {
		if errs := lmi.loadLibrariesFromDir(dir); len(errs) > 0 {
			statuses = append(statuses, errs...)
		}
	}
	return statuses
}

func (lm *LibrariesManager) getLibrariesDir(installLocation libraries.LibraryLocation) (*paths.Path, error) {
	for _, dir := range lm.librariesDir {
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

// loadLibrariesFromDir loads all libraries in the given directory. Returns
// nil if the directory doesn't exists.
func (lm *LibrariesManager) loadLibrariesFromDir(librariesDir *LibrariesDir) []*status.Status {
	statuses := []*status.Status{}

	librariesDir.scanned = true

	var libDirs paths.PathList
	if librariesDir.IsSingleLibrary {
		libDirs.Add(librariesDir.Path)
	} else {
		d, err := librariesDir.Path.ReadDir()
		if os.IsNotExist(err) {
			return statuses
		}
		if err != nil {
			s := status.Newf(codes.FailedPrecondition, tr("reading dir %[1]s: %[2]s"), librariesDir.Path, err)
			return append(statuses, s)
		}
		d.FilterDirs()
		d.FilterOutHiddenFiles()
		libDirs = d
	}

	for _, libDir := range libDirs {
		library, err := libraries.Load(libDir, librariesDir.Location)
		if err != nil {
			s := status.Newf(codes.Internal, tr("loading library from %[1]s: %[2]s"), libDir, err)
			statuses = append(statuses, s)
			continue
		}
		library.ContainerPlatform = librariesDir.PlatformRelease
		alternatives := lm.libraries[library.Name]
		alternatives.Add(library)
		lm.libraries[library.Name] = alternatives
	}

	return statuses
}

// FindByReference return the installed libraries matching the Reference
// name and version or, if the version is nil, the libraries installed
// in the installLocation.
func (lmi *Installer) FindByReference(name string, version *semver.Version, installLocation libraries.LibraryLocation) libraries.List {
	alternatives := lmi.libraries[name]
	if alternatives == nil {
		return nil
	}
	return alternatives.FilterByVersionAndInstallLocation(version, installLocation)
}

// FindAllInstalled returns all the installed libraries
func (lm *LibrariesManager) FindAllInstalled() libraries.List {
	var res libraries.List
	for _, libAlternatives := range lm.libraries {
		for _, libRelease := range libAlternatives {
			// TODO: is this check redundant?
			if libRelease.InstallDir == nil {
				continue
			}
			res.Add(libRelease)
		}
	}
	return res
}
