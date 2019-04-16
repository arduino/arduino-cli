//

package commands

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/url"
	"path"
	"time"

	"github.com/arduino/arduino-cli/arduino/cores/packageindex"

	"github.com/arduino/arduino-cli/arduino/cores/packagemanager"
	"github.com/arduino/arduino-cli/arduino/libraries"
	"github.com/arduino/arduino-cli/arduino/libraries/librariesmanager"
	"github.com/arduino/arduino-cli/configs"
	"github.com/arduino/arduino-cli/rpc"
	paths "github.com/arduino/go-paths-helper"
	"github.com/sirupsen/logrus"
	"go.bug.st/downloader"
)

// this map contains all the running Arduino Core Services instances
// referenced by an int32 handle
var instances = map[int32]*CoreInstance{}
var instancesCount int32 = 1

// CoreInstance is an instance of the Arduino Core Services. The user can
// instantiate as many as needed by providing a different configuration
// for each one.
type CoreInstance struct {
	config     *configs.Configuration
	pm         *packagemanager.PackageManager
	lm         *librariesmanager.LibrariesManager
	getLibOnly bool
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

	config, err := configs.NewConfiguration()
	if err != nil {
		return nil, fmt.Errorf("getting default config values: %s", err)
	}
	config.DataDir = paths.New(inConfig.DataDir)
	config.SketchbookDir = paths.New(inConfig.SketchbookDir)
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
	pm, lm, err := createInstance(ctx, config, req.GetLibraryManagerOnly())
	if err != nil {
		return nil, fmt.Errorf("Impossible create instance")
	}
	instance := &CoreInstance{
		config:     config,
		pm:         pm,
		lm:         lm,
		getLibOnly: req.GetLibraryManagerOnly()}
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
func UpdateLibrariesIndex(ctx context.Context, lm *librariesmanager.LibrariesManager, progressCallback func(*rpc.DownloadProgress)) {
	logrus.Info("Updating libraries index")
	d, err := lm.UpdateIndex()
	if err != nil {
		//formatter.PrintError(err, "Error downloading librarires index")
		//os.Exit(ErrNetwork)
	}
	//formatter.DownloadProgressBar(d, "Updating index: library_index.json")
	//d.Run()
	progressCallback(&rpc.DownloadProgress{
		File:      "library_index.json",
		Url:       d.URL,
		TotalSize: d.Size(),
	})
	d.RunAndPoll(func(downloaded int64) {
		progressCallback(&rpc.DownloadProgress{Downloaded: downloaded})
	}, 250*time.Millisecond)

	if d.Error() != nil {
		//formatter.PrintError(d.Error(), "Error downloading librarires index")
		//os.Exit(ErrNetwork)
	}
}

func UpdateIndex(ctx context.Context, req *rpc.UpdateIndexReq, downloadCB DownloadProgressCB) (*rpc.UpdateIndexResp, error) {
	id := req.Instance.Id
	coreInstance, ok := instances[id]
	if !ok {
		return nil, fmt.Errorf("invalid handle")
	}

	indexpath := coreInstance.config.IndexesDir()
	for _, URL := range coreInstance.config.BoardManagerAdditionalUrls {
		logrus.WithField("url", URL).Print("Updating index")

		tmpFile, err := ioutil.TempFile("", "")
		if err != nil {
			return nil, fmt.Errorf("Error creating temp file for download", err)

		}
		if err := tmpFile.Close(); err != nil {
			return nil, fmt.Errorf("Error creating temp file for download", err)
		}
		tmp := paths.New(tmpFile.Name())
		defer tmp.Remove()

		d, err := downloader.Download(tmp.String(), URL.String())
		if err != nil {
			return nil, fmt.Errorf("Error downloading index "+URL.String(), err)
		}
		coreIndexPath := indexpath.Join(path.Base(URL.Path))
		Download(d, "Updating index: "+coreIndexPath.Base(), downloadCB)
		if d.Error() != nil {
			return nil, fmt.Errorf("Error downloading index "+URL.String(), d.Error())
		}

		if _, err := packageindex.LoadIndex(tmp); err != nil {
			return nil, fmt.Errorf("Invalid package index in "+URL.String(), err)
		}

		if err := indexpath.MkdirAll(); err != nil {
			return nil, fmt.Errorf("Can't create data directory "+indexpath.String(), err)
		}

		if err := tmp.CopyTo(coreIndexPath); err != nil {
			return nil, fmt.Errorf("Error saving downloaded index "+URL.String(), err)
		}
	}
	Rescan(ctx, &rpc.RescanReq{Instance: req.Instance})
	return &rpc.UpdateIndexResp{}, nil
}

func Rescan(ctx context.Context, req *rpc.RescanReq) (*rpc.RescanResp, error) {
	id := req.Instance.Id
	coreInstance, ok := instances[id]
	if !ok {
		return nil, fmt.Errorf("invalid handle")
	}

	pm, lm, err := createInstance(ctx, coreInstance.config, coreInstance.getLibOnly)
	if err != nil {
		return nil, fmt.Errorf("rescanning: %s", err)
	}
	coreInstance.pm = pm
	coreInstance.lm = lm
	return &rpc.RescanResp{}, nil
}

func createInstance(ctx context.Context, config *configs.Configuration, getLibOnly bool) (*packagemanager.PackageManager, *librariesmanager.LibrariesManager, error) {
	var pm *packagemanager.PackageManager
	if !getLibOnly {
		pm = packagemanager.NewPackageManager(
			config.IndexesDir(),
			config.PackagesDir(),
			config.DownloadsDir(),
			config.DataDir.Join("tmp"))

		for _, URL := range config.BoardManagerAdditionalUrls {
			if err := pm.LoadPackageIndex(URL); err != nil {
				return nil, nil, fmt.Errorf("loading "+URL.String()+" package index: %s", err)
			}
		}

		if err := pm.LoadHardware(config); err != nil {
			return nil, nil, fmt.Errorf("loading hardware packages: %s", err)
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
		UpdateLibrariesIndex(ctx, lm, func(curr *rpc.DownloadProgress) {
			fmt.Printf(">> %+v\n", curr)
		})
		if err := lm.LoadIndex(); err != nil {
			return nil, nil, fmt.Errorf("loading libraries index: %s", err)
		}
	}

	// Scan for libraries
	if err := lm.RescanLibraries(); err != nil {
		return nil, nil, fmt.Errorf("libraries rescan: %s", err)
	}
	return pm, lm, nil
}

func Download(d *downloader.Downloader, label string, downloadCB DownloadProgressCB) error {
	if d == nil {
		// This signal means that the file is already downloaded
		downloadCB(&rpc.DownloadProgress{
			File:      label,
			Completed: true,
		})
		return nil
	}
	downloadCB(&rpc.DownloadProgress{
		File:      label,
		Url:       d.URL,
		TotalSize: d.Size(),
	})
	d.RunAndPoll(func(downloaded int64) {
		downloadCB(&rpc.DownloadProgress{Downloaded: downloaded})
	}, 250*time.Millisecond)
	if d.Error() != nil {
		return d.Error()
	}
	downloadCB(&rpc.DownloadProgress{Completed: true})
	return nil
}
