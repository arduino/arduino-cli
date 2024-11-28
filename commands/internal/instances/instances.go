// This file is part of arduino-cli.
//
// Copyright 2024 ARDUINO SA (http://www.arduino.cc/)
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

package instances

import (
	"sync"

	"github.com/arduino/arduino-cli/commands/cmderrors"
	"github.com/arduino/arduino-cli/internal/arduino/cores/packagemanager"
	"github.com/arduino/arduino-cli/internal/arduino/libraries/librariesindex"
	"github.com/arduino/arduino-cli/internal/arduino/libraries/librariesmanager"
	"github.com/arduino/arduino-cli/internal/version"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	"github.com/arduino/go-paths-helper"
	"go.bug.st/downloader/v2"
)

// coreInstance is an instance of the Arduino Core Services. The user can
// instantiate as many as needed by providing a different configuration
// for each one.
type coreInstance struct {
	pm *packagemanager.PackageManager
	lm *librariesmanager.LibrariesManager
	li *librariesindex.Index
}

// instances contains all the running Arduino Core Services instances
var instances = map[int32]*coreInstance{}
var instancesCount int32 = 1
var instancesMux sync.Mutex

// GetPackageManager returns a PackageManager. If the package manager is not found
// (because the instance is invalid or has been destroyed), nil is returned.
// Deprecated: use GetPackageManagerExplorer instead.
func GetPackageManager(inst *rpc.Instance) (*packagemanager.PackageManager, error) {
	instancesMux.Lock()
	i := instances[inst.GetId()]
	instancesMux.Unlock()
	if i == nil {
		return nil, &cmderrors.InvalidInstanceError{}
	}
	return i.pm, nil
}

// GetPackageManagerExplorer returns a new package manager Explorer. The
// explorer holds a read lock on the underlying PackageManager and it should
// be released by calling the returned "release" function.
func GetPackageManagerExplorer(req *rpc.Instance) (explorer *packagemanager.Explorer, release func(), _err error) {
	pm, err := GetPackageManager(req)
	if err != nil {
		return nil, nil, err
	}
	pme, release := pm.NewExplorer()
	return pme, release, nil
}

// GetLibraryManager returns the library manager for the given instance.
func GetLibraryManager(inst *rpc.Instance) (*librariesmanager.LibrariesManager, error) {
	instancesMux.Lock()
	i := instances[inst.GetId()]
	instancesMux.Unlock()
	if i == nil {
		return nil, &cmderrors.InvalidInstanceError{}
	}
	return i.lm, nil
}

// GetLibraryManagerExplorer returns the library manager Explorer for the given instance.
func GetLibraryManagerExplorer(inst *rpc.Instance) (*librariesmanager.Explorer, func(), error) {
	lm, err := GetLibraryManager(inst)
	if err != nil {
		return nil, nil, err
	}
	lmi, release := lm.NewExplorer()
	return lmi, release, nil
}

// GetLibraryManagerInstaller returns the library manager Installer for the given instance.
func GetLibraryManagerInstaller(inst *rpc.Instance) (*librariesmanager.Installer, func(), error) {
	lm, err := GetLibraryManager(inst)
	if err != nil {
		return nil, nil, err
	}
	lmi, release := lm.NewInstaller()
	return lmi, release, nil
}

// GetLibrariesIndex returns the library index for the given instance.
func GetLibrariesIndex(inst *rpc.Instance) (*librariesindex.Index, error) {
	instancesMux.Lock()
	defer instancesMux.Unlock()
	i := instances[inst.GetId()]
	if i == nil {
		return nil, &cmderrors.InvalidInstanceError{}
	}
	return i.li, nil
}

// SetLibrariesIndex sets the library index for the given instance.
func SetLibrariesIndex(inst *rpc.Instance, li *librariesindex.Index) error {
	instancesMux.Lock()
	defer instancesMux.Unlock()
	i := instances[inst.GetId()]
	if i == nil {
		return &cmderrors.InvalidInstanceError{}
	}
	i.li = li
	return nil
}

// SetLibraryManager sets the library manager for the given instance.
func SetLibraryManager(inst *rpc.Instance, lm *librariesmanager.LibrariesManager) bool {
	instancesMux.Lock()
	i := instances[inst.GetId()]
	instancesMux.Unlock()
	if i == nil {
		return false
	}
	i.lm = lm
	return true
}

// Create a new *rpc.Instance ready to be initialized
func Create(dataDir, packagesDir, userPackagesDir, downloadsDir *paths.Path, extraUserAgent string, downloaderConfig downloader.Config) (*rpc.Instance, error) {
	// Create package manager
	userAgent := "arduino-cli/" + version.VersionInfo.VersionString
	if extraUserAgent != "" {
		userAgent += " " + extraUserAgent
	}
	tempDir := dataDir.Join("tmp")

	pm := packagemanager.NewBuilder(dataDir, packagesDir, userPackagesDir, downloadsDir, tempDir, userAgent, downloaderConfig).Build()
	lm, _ := librariesmanager.NewBuilder().Build()

	instance := &coreInstance{
		pm: pm,
		lm: lm,
		li: librariesindex.EmptyIndex,
	}

	// Save instance
	instancesMux.Lock()
	id := instancesCount
	instances[id] = instance
	instancesCount++
	instancesMux.Unlock()

	return &rpc.Instance{Id: id}, nil
}

// IsValid returns true if the given instance is valid.
func IsValid(inst *rpc.Instance) bool {
	instancesMux.Lock()
	i := instances[inst.GetId()]
	instancesMux.Unlock()
	return i != nil
}

// Delete removes an instance.
func Delete(inst *rpc.Instance) bool {
	instancesMux.Lock()
	defer instancesMux.Unlock()
	if _, ok := instances[inst.GetId()]; !ok {
		return false
	}
	delete(instances, inst.GetId())
	return true
}
