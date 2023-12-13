package instances

import (
	"sync"

	"github.com/arduino/arduino-cli/internal/arduino/cores/packagemanager"
	"github.com/arduino/arduino-cli/internal/arduino/libraries/librariesmanager"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	"github.com/arduino/arduino-cli/version"
	"github.com/arduino/go-paths-helper"
)

// coreInstance is an instance of the Arduino Core Services. The user can
// instantiate as many as needed by providing a different configuration
// for each one.
type coreInstance struct {
	pm *packagemanager.PackageManager
	lm *librariesmanager.LibrariesManager
}

// instances contains all the running Arduino Core Services instances
var instances = map[int32]*coreInstance{}
var instancesCount int32 = 1
var instancesMux sync.Mutex

// GetPackageManager returns a PackageManager. If the package manager is not found
// (because the instance is invalid or has been destroyed), nil is returned.
// Deprecated: use GetPackageManagerExplorer instead.
func GetPackageManager(inst *rpc.Instance) *packagemanager.PackageManager {
	instancesMux.Lock()
	i := instances[inst.GetId()]
	instancesMux.Unlock()
	if i == nil {
		return nil
	}
	return i.pm
}

// GetPackageManagerExplorer returns a new package manager Explorer. The
// explorer holds a read lock on the underlying PackageManager and it should
// be released by calling the returned "release" function.
func GetPackageManagerExplorer(req *rpc.Instance) (explorer *packagemanager.Explorer, release func()) {
	pm := GetPackageManager(req)
	if pm == nil {
		return nil, nil
	}
	return pm.NewExplorer()
}

// GetLibraryManager returns the library manager for the given instance.
func GetLibraryManager(inst *rpc.Instance) *librariesmanager.LibrariesManager {
	instancesMux.Lock()
	i := instances[inst.GetId()]
	instancesMux.Unlock()
	if i == nil {
		return nil
	}
	return i.lm
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
func Create(dataDir, packagesDir, downloadsDir *paths.Path, extraUserAgent ...string) (*rpc.Instance, error) {
	instance := &coreInstance{}

	// Create package manager
	userAgent := "arduino-cli/" + version.VersionInfo.VersionString
	for _, ua := range extraUserAgent {
		userAgent += " " + ua
	}
	tempDir := dataDir.Join("tmp")
	instance.pm = packagemanager.NewBuilder(dataDir, packagesDir, downloadsDir, tempDir, userAgent).Build()
	instance.lm = librariesmanager.NewLibraryManager(dataDir, downloadsDir)

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
