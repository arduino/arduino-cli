//
// This file is part of arduino-cli.
//
// Copyright 2018 ARDUINO SA (http://www.arduino.cc/)
//
// This software is released under the GNU General Public License version 3,
// which covers the main part of arduino-cli.
// The terms of this license can be found at:
// https://www.gnu.org/licenses/gpl-3.0.en.html
//
// You can be released from the requirements of the above licenses by purchasing
// a commercial license. Buying such a license is mandatory if you want to modify or
// otherwise use the software for commercial activities involving the Arduino
// software without disclosing the source code of your own applications. To purchase
// a commercial license, send an email to license@arduino.cc.
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
	"github.com/arduino/arduino-cli/arduino/discovery"
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
	config      *configs.Configuration
	pm          *packagemanager.PackageManager
	lm          *librariesmanager.LibrariesManager
	getLibOnly  bool
	discoveries []*discovery.Discovery
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

func GetDiscoveries(req InstanceContainer) []*discovery.Discovery {
	i, ok := instances[req.GetInstance().GetId()]
	if !ok {
		return nil
	}
	return i.discoveries
}

func (instance *CoreInstance) startDiscoveries(downloadCB DownloadProgressCB, taskCB TaskProgressCB) error {
	discoveriesToStop := instance.discoveries
	discoveriesToStart := discovery.ExtractDiscoveriesFromPlatforms(instance.pm)

	instance.discoveries = []*discovery.Discovery{}
	for _, disc := range discoveriesToStart {
		sharedDisc, err := StartSharedDiscovery(disc)
		if err != nil {
			return fmt.Errorf("starting discovery: %s", err)
		}
		instance.discoveries = append(instance.discoveries, sharedDisc)
	}
	for _, disc := range discoveriesToStop {
		StopSharedDiscovery(disc)
	}
	return nil
}

func Init(ctx context.Context, req *rpc.InitReq, downloadCB DownloadProgressCB, taskCB TaskProgressCB) (*rpc.InitResp, error) {
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
	for _, rawurl := range inConfig.BoardManagerAdditionalUrls {
		if u, err := url.Parse(rawurl); err == nil {
			config.BoardManagerAdditionalUrls = append(config.BoardManagerAdditionalUrls, u)
		} else {
			return nil, fmt.Errorf("parsing url %s: %s", rawurl, err)
		}
	}

	pm, lm, reqPltIndex, reqLibIndex, err := createInstance(ctx, config, req.GetLibraryManagerOnly())
	if err != nil {
		return nil, fmt.Errorf("cannot initialize package manager: %s", err)
	}
	instance := &CoreInstance{
		config:     config,
		pm:         pm,
		lm:         lm,
		getLibOnly: req.GetLibraryManagerOnly()}
	handle := instancesCount
	instancesCount++
	instances[handle] = instance

	if err := instance.startDiscoveries(downloadCB, taskCB); err != nil {
		// TODO: handle discovery errors
		fmt.Println(err)
	}
	return &rpc.InitResp{
		Instance:             &rpc.Instance{Id: handle},
		PlatformsIndexErrors: reqPltIndex,
		LibrariesIndexError:  reqLibIndex,
	}, nil
}

func Destroy(ctx context.Context, req *rpc.DestroyReq) (*rpc.DestroyResp, error) {
	id := req.GetInstance().GetId()
	if _, ok := instances[id]; !ok {
		return nil, fmt.Errorf("invalid handle")
	}

	for _, disc := range GetDiscoveries(req) {
		StopSharedDiscovery(disc)
	}

	delete(instances, id)
	return &rpc.DestroyResp{}, nil
}

// UpdateLibrariesIndex updates the library_index.json
func UpdateLibrariesIndex(ctx context.Context, req *rpc.UpdateLibrariesIndexReq, downloadCB func(*rpc.DownloadProgress)) error {
	logrus.Info("Updating libraries index")
	lm := GetLibraryManager(req)
	if lm == nil {
		return fmt.Errorf("invalid handle")
	}
	d, err := lm.UpdateIndex()
	if err != nil {
		return err
	}
	Download(d, "Updating index: library_index.json", downloadCB)
	if d.Error() != nil {
		return d.Error()
	}
	if _, err := Rescan(ctx, &rpc.RescanReq{Instance: req.Instance}); err != nil {
		return fmt.Errorf("rescanning filesystem: %s", err)
	}
	return nil
}

func UpdateIndex(ctx context.Context, req *rpc.UpdateIndexReq, downloadCB DownloadProgressCB) (*rpc.UpdateIndexResp, error) {
	id := req.GetInstance().GetId()
	coreInstance, ok := instances[id]
	if !ok {
		return nil, fmt.Errorf("invalid handle")
	}

	indexpath := coreInstance.config.IndexesDir()
	for _, URL := range coreInstance.config.BoardManagerAdditionalUrls {
		logrus.WithField("url", URL).Print("Updating index")

		tmpFile, err := ioutil.TempFile("", "")
		if err != nil {
			return nil, fmt.Errorf("creating temp file for download: %s", err)
		}
		if err := tmpFile.Close(); err != nil {
			return nil, fmt.Errorf("creating temp file for download: %s", err)
		}
		tmp := paths.New(tmpFile.Name())
		defer tmp.Remove()

		d, err := downloader.Download(tmp.String(), URL.String())
		if err != nil {
			return nil, fmt.Errorf("downloading index %s: %s", URL, err)
		}
		coreIndexPath := indexpath.Join(path.Base(URL.Path))
		Download(d, "Updating index: "+coreIndexPath.Base(), downloadCB)
		if d.Error() != nil {
			return nil, fmt.Errorf("downloading index %s: %s", URL, d.Error())
		}

		if _, err := packageindex.LoadIndex(tmp); err != nil {
			return nil, fmt.Errorf("invalid package index in %s: %s", URL, err)
		}

		if err := indexpath.MkdirAll(); err != nil {
			return nil, fmt.Errorf("can't create data directory %s: %s", indexpath, err)
		}

		if err := tmp.CopyTo(coreIndexPath); err != nil {
			return nil, fmt.Errorf("saving downloaded index %s: %s", URL, err)
		}
	}
	if _, err := Rescan(ctx, &rpc.RescanReq{Instance: req.Instance}); err != nil {
		return nil, fmt.Errorf("rescanning filesystem: %s", err)
	}
	return &rpc.UpdateIndexResp{}, nil
}

func Rescan(ctx context.Context, req *rpc.RescanReq) (*rpc.RescanResp, error) {
	id := req.GetInstance().GetId()
	coreInstance, ok := instances[id]
	if !ok {
		return nil, fmt.Errorf("invalid handle")
	}

	pm, lm, reqPltIndex, reqLibIndex, err := createInstance(ctx, coreInstance.config, coreInstance.getLibOnly)
	if err != nil {
		return nil, fmt.Errorf("rescanning filesystem: %s", err)
	}
	coreInstance.pm = pm
	coreInstance.lm = lm

	coreInstance.startDiscoveries(nil, nil)

	return &rpc.RescanResp{
		PlatformsIndexErrors: reqPltIndex,
		LibrariesIndexError:  reqLibIndex,
	}, nil
}

func createInstance(ctx context.Context, config *configs.Configuration, getLibOnly bool) (*packagemanager.PackageManager, *librariesmanager.LibrariesManager, []string, string, error) {
	var pm *packagemanager.PackageManager
	platformIndexErrors := []string{}
	if !getLibOnly {
		pm = packagemanager.NewPackageManager(
			config.IndexesDir(),
			config.PackagesDir(),
			config.DownloadsDir(),
			config.DataDir.Join("tmp"))

		for _, URL := range config.BoardManagerAdditionalUrls {
			if err := pm.LoadPackageIndex(URL); err != nil {
				platformIndexErrors = append(platformIndexErrors, err.Error())
			}
		}

		if err := pm.LoadHardware(config); err != nil {
			return nil, nil, nil, "", fmt.Errorf("loading hardware packages: %s", err)
		}
	}
	if len(platformIndexErrors) == 0 {
		platformIndexErrors = nil
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
	librariesIndexError := ""
	if err := lm.LoadIndex(); err != nil {
		librariesIndexError = err.Error()
	}

	// Scan for libraries
	if err := lm.RescanLibraries(); err != nil {
		return nil, nil, nil, "", fmt.Errorf("libraries rescan: %s", err)
	}
	return pm, lm, platformIndexErrors, librariesIndexError, nil
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
