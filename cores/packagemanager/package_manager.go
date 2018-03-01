package packagemanager

import (
	"sync"

	"github.com/bcmi-labs/arduino-cli/cores/packageindex"
	"fmt"
	"github.com/bcmi-labs/arduino-cli/cores"
	"github.com/bcmi-labs/arduino-cli/configs"
	"os"
	"path/filepath"
	"github.com/juju/errors"
)

var packageManagerInstance *packageManager
var once sync.Once

// PackageManager defines the superior oracle which understands all about
// Arduino Packages, how to parse them, download, and so on.
//
// The manager also keeps track of the status of the Packages (their Platform Releases, actually)
// installed in the system.
type packageManager struct {
	/* FIXME

		What's the typical flow?

		You start by downloading AN PackageIndex (remember: multiple repositories are allowed??)
	*/

	packages *cores.Packages
}

// PackageManager return the instance of the PackageManager
// yeah, that's a singleton by the way...
func PackageManager() *packageManager {
	once.Do(func() {
		// TODO: why not just use the Go pkg init()?
		packageManagerInstance = &packageManager{}
		// TODO: perhaps directly use the loading from PackagesIndex file?
		packageManagerInstance.packages = cores.NewPackages()
	})

	fmt.Println("==> Cool! You're using my new singleton PackageManager.")

	return packageManagerInstance
}

// FIXME add an handler to be invoked on each verbose operation, in order to let commands display results through the formatter
// as for the progress bars during download

// FIXME this is currently hard-coded with the default PackageIndex and won't merge or check existing Packages!!
func (pm *packageManager) AddDefaultPackageIndex() (*packageManager, error) {
	var index packageindex.Index
	err := packageindex.LoadIndex(&index)
	if err != nil {
		//TODO: The original version would automatically fix corrupted index?
		/*status, err := prettyPrints.CorruptedCoreIndexFix(index)
		if err != nil {
			return pm, err
		}
		pm.packages = &status
		return pm, nil*/

		return pm, errors.Annotate(err, fmt.Sprintf("failed to load the package index, is probably corrupted" ))
	}

	// TODO: if this really is a singleton, a lock is needed :(
	pm.packages = index.CreateStatusContext()
	return pm, nil
}

// TODO: implement the generic version (with merge)
/*func (pm *packageManager) AddPackageIndex() *packageManager {

}*/

// DownloadPackagesFile downloads the core packages index file from Arduino repository.
func (pm *packageManager) DownloadPackagesFile() error {
	return packageindex.DownloadPackagesFile()
}

func (pm *packageManager) Package(name string) *packageActions {
	//TODO: perhaps these 2 structure should be merged? cores.Packages vs pkgmgr??
	var err error
	thePackage := pm.packages.Packages[name]
	if thePackage == nil {
		err = fmt.Errorf("package '%s' not found", name)
	}
	return &packageActions{
		aPackage:     thePackage,
		forwardError: err,
	}
}

type packageActions struct {
	aPackage     *cores.Package
	forwardError error
}

func (pa *packageActions) Tool(name string) *toolActions {
	var tool *cores.Tool
	err := pa.forwardError
	if err == nil {
		tool = pa.aPackage.Tools[name]

		if tool == nil {
			err = fmt.Errorf("tool '%s' not found in package '%s'", name, pa.aPackage.Name)
		}
	}
	return &toolActions{
		tool:         tool,
		forwardError: err,
	}
}

type toolActions struct{
	tool         *cores.Tool
	forwardError error
}

func (ta *toolActions) Get() (*cores.Tool, error) {
	err := ta.forwardError
	if err == nil {
		return ta.tool, nil
	}
	return nil, err
}

func (ta *toolActions) IsInstalled() (bool, error) {
	err := ta.forwardError
	if err == nil {
		location, err := configs.ToolsFolder(ta.tool.Package.Name).Get()
		if err != nil {
			return false, err
		}
		_, err = os.Stat(filepath.Join(location, ta.tool.Name))
		if !os.IsNotExist(err) {
			return true, nil
		}
		return false, nil
	}
	return false, err
}