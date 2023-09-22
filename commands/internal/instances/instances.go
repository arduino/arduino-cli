package instances

import (
	"sync"

	"github.com/arduino/arduino-cli/arduino"
	"github.com/arduino/arduino-cli/arduino/cores/packagemanager"
	"github.com/arduino/arduino-cli/arduino/libraries/librariesmanager"
	"github.com/arduino/arduino-cli/configuration"
	"github.com/arduino/arduino-cli/i18n"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	"github.com/arduino/arduino-cli/version"
)

var tr = i18n.Tr

// coreInstance is an instance of the Arduino Core Services. The user can
// instantiate as many as needed by providing a different configuration
// for each one.
type coreInstance struct {
	pm *packagemanager.PackageManager
	lm *librariesmanager.LibrariesManager
}

// coreInstancesContainer has methods to add an remove instances atomically.
type coreInstancesContainer struct {
	instances      map[int32]*coreInstance
	instancesCount int32
	instancesMux   sync.Mutex
}

// instances contains all the running Arduino Core Services instances
var instances = &coreInstancesContainer{
	instances:      map[int32]*coreInstance{},
	instancesCount: 1,
}

// GetInstance returns a CoreInstance for the given ID, or nil if ID
// doesn't exist
func (c *coreInstancesContainer) GetInstance(id int32) *coreInstance {
	c.instancesMux.Lock()
	defer c.instancesMux.Unlock()
	return c.instances[id]
}

// AddAndAssignID saves the CoreInstance and assigns a unique ID to
// retrieve it later
func (c *coreInstancesContainer) AddAndAssignID(i *coreInstance) int32 {
	c.instancesMux.Lock()
	defer c.instancesMux.Unlock()
	id := c.instancesCount
	c.instances[id] = i
	c.instancesCount++
	return id
}

// RemoveID removes the CoreInstance referenced by id. Returns true
// if the operation is successful, or false if the CoreInstance does
// not exist
func (c *coreInstancesContainer) RemoveID(id int32) bool {
	c.instancesMux.Lock()
	defer c.instancesMux.Unlock()
	if _, ok := c.instances[id]; !ok {
		return false
	}
	delete(c.instances, id)
	return true
}

// GetPackageManager returns a PackageManager. If the package manager is not found
// (because the instance is invalid or has been destroyed), nil is returned.
// Deprecated: use GetPackageManagerExplorer instead.
func GetPackageManager(instance *rpc.Instance) *packagemanager.PackageManager {
	i := instances.GetInstance(instance.GetId())
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
func GetLibraryManager(req *rpc.Instance) *librariesmanager.LibrariesManager {
	i := instances.GetInstance(req.GetId())
	if i == nil {
		return nil
	}
	return i.lm
}

// SetLibraryManager sets the library manager for the given instance.
func SetLibraryManager(inst *rpc.Instance, lm *librariesmanager.LibrariesManager) bool {
	coreInstance := instances.GetInstance(inst.GetId())
	if coreInstance == nil {
		return false
	}
	coreInstance.lm = lm
	return true
}

// Create a new *rpc.Instance ready to be initialized, supporting directories are also created.
func Create(extraUserAgent ...string) (*rpc.Instance, error) {
	instance := &coreInstance{}

	// Setup downloads directory
	downloadsDir := configuration.DownloadsDir(configuration.Settings)
	if downloadsDir.NotExist() {
		err := downloadsDir.MkdirAll()
		if err != nil {
			return nil, &arduino.PermissionDeniedError{Message: tr("Failed to create downloads directory"), Cause: err}
		}
	}

	// Setup data directory
	dataDir := configuration.DataDir(configuration.Settings)
	packagesDir := configuration.PackagesDir(configuration.Settings)
	if packagesDir.NotExist() {
		err := packagesDir.MkdirAll()
		if err != nil {
			return nil, &arduino.PermissionDeniedError{Message: tr("Failed to create data directory"), Cause: err}
		}
	}

	// Create package manager
	userAgent := "arduino-cli/" + version.VersionInfo.VersionString
	for _, ua := range extraUserAgent {
		userAgent += " " + ua
	}
	instance.pm = packagemanager.NewBuilder(
		dataDir,
		configuration.PackagesDir(configuration.Settings),
		downloadsDir,
		dataDir.Join("tmp"),
		userAgent,
	).Build()
	instance.lm = librariesmanager.NewLibraryManager(
		dataDir,
		downloadsDir,
	)

	// Save instance
	instanceID := instances.AddAndAssignID(instance)
	return &rpc.Instance{Id: instanceID}, nil
}

// IsValid returns true if the given instance is valid.
func IsValid(inst *rpc.Instance) bool {
	if inst == nil {
		return false
	}
	return instances.GetInstance(inst.GetId()) != nil
}

// Delete removes an instance.
func Delete(inst *rpc.Instance) bool {
	if inst == nil {
		return false
	}
	return instances.RemoveID(inst.GetId())
}
