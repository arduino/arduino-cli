//

package commands

import (
	"context"
	"fmt"
	"net/url"

	"github.com/arduino/arduino-cli/arduino/cores/packagemanager"
	"github.com/arduino/arduino-cli/arduino/libraries"
	"github.com/arduino/arduino-cli/arduino/libraries/librariesmanager"
	"github.com/arduino/arduino-cli/configs"
	"github.com/arduino/arduino-cli/rpc"
	paths "github.com/arduino/go-paths-helper"
)

// this map contains all the running Arduino Core Services instances
// referenced by an int32 handle
var instances = map[int32]*CoreInstance{}
var instancesCount int32 = 1

// CoreInstance is an instance of the Arduino Core Services. The user can
// instantiate as many as needed by providing a different configuration
// for each one.
type CoreInstance struct {
	config *configs.Configuration
	pm     *packagemanager.PackageManager
	lm     *librariesmanager.LibrariesManager
}

type InstanceContainer interface {
	GetInstance() *rpc.Instance
}

func GetPackageManager(req InstanceContainer) *packagemanager.PackageManager {
	i, ok := instances[req.GetInstance().GetId()]
	if !ok {
		return nil
	}
	return i.pm
}

func GetLibraryManager(req InstanceContainer) *librariesmanager.LibrariesManager {
	i, ok := instances[req.GetInstance().GetId()]
	if !ok {
		return nil
	}
	return i.lm
}

func Init(ctx context.Context, req *rpc.InitReq) (*rpc.InitResp, error) {
	inConfig := req.GetConfiguration()
	if inConfig == nil {
		return nil, fmt.Errorf("invalid request")
	}

	config := &configs.Configuration{
		DataDir:       paths.New(inConfig.DataDir),
		SketchbookDir: paths.New(inConfig.SketchbookDir),
	}
	if inConfig.DownloadsDir != "" {
		config.ArduinoDownloadsDir = paths.New(inConfig.DownloadsDir)
	}
	urls := []*url.URL{}
	for _, rawurl := range inConfig.BoardManagerAdditionalUrls {
		if u, err := url.Parse(rawurl); err == nil {
			urls = append(urls, u)
		} else {
			return nil, fmt.Errorf("parsing url %s: %s", rawurl, err)
		}
	}

	var pm *packagemanager.PackageManager
	if !req.GetLibraryManagerOnly() {
		pm = packagemanager.NewPackageManager(
			config.IndexesDir(),
			config.PackagesDir(),
			config.DownloadsDir(),
			config.DataDir.Join("tmp"))

		for _, URL := range config.BoardManagerAdditionalUrls {
			if err := pm.LoadPackageIndex(URL); err != nil {
				return nil, fmt.Errorf("loading "+URL.String()+" package index: %s", err)
			}
		}

		if err := pm.LoadHardware(config); err != nil {
			return nil, fmt.Errorf("loading hardware packages: %s", err)
		}
	}

	// Initialize library manager
	// --------------------------
	lm := librariesmanager.NewLibraryManager(
		config.IndexesDir(),
		config.DownloadsDir())

	// Add IDE builtin libraries dir
	if bundledLibsDir := config.IDEBundledLibrariesDir(); bundledLibsDir != nil {
		lm.AddLibrariesDir(bundledLibsDir, libraries.IDEBuiltIn)
	}

	// Add sketchbook libraries dir
	lm.AddLibrariesDir(config.LibrariesDir(), libraries.Sketchbook)

	// Add libraries dirs from installed platforms
	if pm != nil {
		for _, targetPackage := range pm.GetPackages().Packages {
			for _, platform := range targetPackage.Platforms {
				if platformRelease := pm.GetInstalledPlatformRelease(platform); platformRelease != nil {
					lm.AddPlatformReleaseLibrariesDir(platformRelease, libraries.PlatformBuiltIn)
				}
			}
		}
	}

	// Load index and auto-update it if needed
	if err := lm.LoadIndex(); err != nil {
		UpdateLibrariesIndex(ctx, lm)
		if err := lm.LoadIndex(); err != nil {
			return nil, fmt.Errorf("loading libraries index: %s", err)
		}
	}

	// Scan for libraries
	if err := lm.RescanLibraries(); err != nil {
		return nil, fmt.Errorf("libraries rescan: %s", err)
	}

	instance := &CoreInstance{
		config: config,
		pm:     pm,
		lm:     lm}
	handle := instancesCount
	instancesCount++
	instances[handle] = instance

	return &rpc.InitResp{
		Instance: &rpc.Instance{Id: handle},
	}, nil
}

func Destroy(ctx context.Context, req *rpc.DestroyReq) (*rpc.DestroyResp, error) {
	id := req.Instance.Id
	_, ok := instances[id]
	if !ok {
		return nil, fmt.Errorf("invalid handle")
	}
	delete(instances, id)
	return &rpc.DestroyResp{}, nil
}

// UpdateLibrariesIndex updates the library_index.json
func UpdateLibrariesIndex(ctx context.Context, lm *librariesmanager.LibrariesManager) {
	//logrus.Info("Updating libraries index")
	d, err := lm.UpdateIndex()
	if err != nil {
		//formatter.PrintError(err, "Error downloading librarires index")
		//os.Exit(ErrNetwork)
	}
	//formatter.DownloadProgressBar(d, "Updating index: library_index.json")
	d.Run()
	if d.Error() != nil {
		//formatter.PrintError(d.Error(), "Error downloading librarires index")
		//os.Exit(ErrNetwork)
	}
}
