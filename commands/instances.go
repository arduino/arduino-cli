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

package commands

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"time"

	"github.com/arduino/arduino-cli/arduino/cores"
	"github.com/arduino/arduino-cli/arduino/cores/packageindex"
	"github.com/arduino/arduino-cli/arduino/cores/packagemanager"
	"github.com/arduino/arduino-cli/arduino/libraries"
	"github.com/arduino/arduino-cli/arduino/libraries/librariesmanager"
	"github.com/arduino/arduino-cli/cli/globals"
	"github.com/arduino/arduino-cli/configuration"
	rpc "github.com/arduino/arduino-cli/rpc/commands"
	paths "github.com/arduino/go-paths-helper"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
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
	PackageManager *packagemanager.PackageManager
	lm             *librariesmanager.LibrariesManager
	getLibOnly     bool
}

// InstanceContainer FIXMEDOC
type InstanceContainer interface {
	GetInstance() *rpc.Instance
}

type createInstanceResult struct {
	Pm                  *packagemanager.PackageManager
	Lm                  *librariesmanager.LibrariesManager
	PlatformIndexErrors []string
	LibrariesIndexError string
}

// GetInstance returns a CoreInstance for the given ID, or nil if ID
// doesn't exist
func GetInstance(id int32) *CoreInstance {
	return instances[id]
}

// GetPackageManager returns a PackageManager for the given ID, or nil if
// ID doesn't exist
func GetPackageManager(id int32) *packagemanager.PackageManager {
	i, ok := instances[id]
	if !ok {
		return nil
	}
	return i.PackageManager
}

// GetLibraryManager returns the library manager for the given instance ID
func GetLibraryManager(instanceID int32) *librariesmanager.LibrariesManager {
	i, ok := instances[instanceID]
	if !ok {
		return nil
	}
	return i.lm
}

func (instance *CoreInstance) installToolIfMissing(tool *cores.ToolRelease, downloadCB DownloadProgressCB,
	taskCB TaskProgressCB, downloaderHeaders http.Header) (bool, error) {
	if tool.IsInstalled() {
		return false, nil
	}
	taskCB(&rpc.TaskProgress{Name: "Downloading missing tool " + tool.String()})
	if err := DownloadToolRelease(instance.PackageManager, tool, downloadCB, downloaderHeaders); err != nil {
		return false, fmt.Errorf("downloading %s tool: %s", tool, err)
	}
	taskCB(&rpc.TaskProgress{Completed: true})
	if err := InstallToolRelease(instance.PackageManager, tool, taskCB); err != nil {
		return false, fmt.Errorf("installing %s tool: %s", tool, err)
	}
	return true, nil
}

func (instance *CoreInstance) checkForBuiltinTools(downloadCB DownloadProgressCB, taskCB TaskProgressCB,
	downloaderHeaders http.Header) error {
	// Check for ctags tool
	ctags, _ := getBuiltinCtagsTool(instance.PackageManager)
	ctagsInstalled, err := instance.installToolIfMissing(ctags, downloadCB, taskCB, downloaderHeaders)
	if err != nil {
		return err
	}

	// Check for bultin serial-discovery tool
	serialDiscoveryTool, _ := getBuiltinSerialDiscoveryTool(instance.PackageManager)
	serialDiscoveryInstalled, err := instance.installToolIfMissing(serialDiscoveryTool, downloadCB, taskCB, downloaderHeaders)
	if err != nil {
		return err
	}

	if ctagsInstalled || serialDiscoveryInstalled {
		if err := instance.PackageManager.LoadHardware(); err != nil {
			return fmt.Errorf("could not load hardware packages: %s", err)
		}
	}
	return nil
}

// Init FIXMEDOC
func Init(ctx context.Context, req *rpc.InitReq, downloadCB DownloadProgressCB, taskCB TaskProgressCB,
	downloaderHeaders http.Header) (*rpc.InitResp, error) {

	res, err := createInstance(ctx, req.GetLibraryManagerOnly())
	if err != nil {
		return nil, fmt.Errorf("cannot initialize package manager: %s", err)
	}

	instance := &CoreInstance{
		PackageManager: res.Pm,
		lm:             res.Lm,
		getLibOnly:     req.GetLibraryManagerOnly(),
	}
	handle := instancesCount
	instancesCount++
	instances[handle] = instance

	if err := instance.checkForBuiltinTools(downloadCB, taskCB, downloaderHeaders); err != nil {
		fmt.Println(err)
		return nil, err
	}

	return &rpc.InitResp{
		Instance:             &rpc.Instance{Id: handle},
		PlatformsIndexErrors: res.PlatformIndexErrors,
		LibrariesIndexError:  res.LibrariesIndexError,
	}, nil
}

// Destroy FIXMEDOC
func Destroy(ctx context.Context, req *rpc.DestroyReq) (*rpc.DestroyResp, error) {
	id := req.GetInstance().GetId()
	if _, ok := instances[id]; !ok {
		return nil, fmt.Errorf("invalid handle")
	}

	delete(instances, id)
	return &rpc.DestroyResp{}, nil
}

// UpdateLibrariesIndex updates the library_index.json
func UpdateLibrariesIndex(ctx context.Context, req *rpc.UpdateLibrariesIndexReq, downloadCB func(*rpc.DownloadProgress)) error {
	logrus.Info("Updating libraries index")
	lm := GetLibraryManager(req.GetInstance().GetId())
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
	if _, err := Rescan(req.GetInstance().GetId()); err != nil {
		return fmt.Errorf("rescanning filesystem: %s", err)
	}
	return nil
}

// UpdateIndex FIXMEDOC
func UpdateIndex(ctx context.Context, req *rpc.UpdateIndexReq, downloadCB DownloadProgressCB) (*rpc.UpdateIndexResp, error) {
	id := req.GetInstance().GetId()
	_, ok := instances[id]
	if !ok {
		return nil, fmt.Errorf("invalid handle")
	}

	indexpath := paths.New(viper.GetString("directories.Data"))
	urls := []string{globals.DefaultIndexURL}
	urls = append(urls, viper.GetStringSlice("board_manager.additional_urls")...)
	for _, u := range urls {
		URL, err := url.Parse(u)
		if err != nil {
			logrus.Warnf("unable to parse additional URL: %s", u)
			continue
		}

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
	if _, err := Rescan(id); err != nil {
		return nil, fmt.Errorf("rescanning filesystem: %s", err)
	}
	return &rpc.UpdateIndexResp{}, nil
}

// Rescan restart discoveries for the given instance
func Rescan(instanceID int32) (*rpc.RescanResp, error) {
	coreInstance, ok := instances[instanceID]
	if !ok {
		return nil, fmt.Errorf("invalid handle")
	}

	res, err := createInstance(context.Background(), coreInstance.getLibOnly)
	if err != nil {
		return nil, fmt.Errorf("rescanning filesystem: %s", err)
	}
	coreInstance.PackageManager = res.Pm
	coreInstance.lm = res.Lm

	return &rpc.RescanResp{
		PlatformsIndexErrors: res.PlatformIndexErrors,
		LibrariesIndexError:  res.LibrariesIndexError,
	}, nil
}

func createInstance(ctx context.Context, getLibOnly bool) (*createInstanceResult, error) {
	res := &createInstanceResult{}

	// setup downloads directory
	downloadsDir := paths.New(viper.GetString("directories.Downloads"))
	if downloadsDir.NotExist() {
		err := downloadsDir.MkdirAll()
		if err != nil {
			return nil, err
		}
	}

	// setup data directory
	dataDir := paths.New(viper.GetString("directories.Data"))
	packagesDir := configuration.PackagesDir()
	if packagesDir.NotExist() {
		err := packagesDir.MkdirAll()
		if err != nil {
			return nil, err
		}
	}

	if !getLibOnly {
		res.Pm = packagemanager.NewPackageManager(dataDir, configuration.PackagesDir(),
			downloadsDir, dataDir.Join("tmp"))

		urls := []string{globals.DefaultIndexURL}
		urls = append(urls, viper.GetStringSlice("board_manager.additional_urls")...)
		for _, u := range urls {
			URL, err := url.Parse(u)
			if err != nil {
				logrus.Warnf("Unable to parse index URL: %s, skip...", u)
				continue
			}

			if err := res.Pm.LoadPackageIndex(URL); err != nil {
				res.PlatformIndexErrors = append(res.PlatformIndexErrors, err.Error())
			}
		}

		if err := res.Pm.LoadHardware(); err != nil {
			return res, fmt.Errorf("error loading hardware packages: %s", err)
		}
	}

	if len(res.PlatformIndexErrors) == 0 {
		res.PlatformIndexErrors = nil
	}

	// Initialize library manager
	// --------------------------
	res.Lm = librariesmanager.NewLibraryManager(dataDir, downloadsDir)

	// Add IDE builtin libraries dir
	if bundledLibsDir := configuration.IDEBundledLibrariesDir(); bundledLibsDir != nil {
		res.Lm.AddLibrariesDir(bundledLibsDir, libraries.IDEBuiltIn)
	}

	// Add user libraries dir
	libDir := configuration.LibrariesDir()
	res.Lm.AddLibrariesDir(libDir, libraries.User)

	// Add libraries dirs from installed platforms
	if res.Pm != nil {
		for _, targetPackage := range res.Pm.Packages {
			for _, platform := range targetPackage.Platforms {
				if platformRelease := res.Pm.GetInstalledPlatformRelease(platform); platformRelease != nil {
					res.Lm.AddPlatformReleaseLibrariesDir(platformRelease, libraries.PlatformBuiltIn)
				}
			}
		}
	}

	// Load index and auto-update it if needed
	if err := res.Lm.LoadIndex(); err != nil {
		res.LibrariesIndexError = err.Error()
	}

	// Scan for libraries
	if err := res.Lm.RescanLibraries(); err != nil {
		return res, fmt.Errorf("libraries rescan: %s", err)
	}

	return res, nil
}

// Download FIXMEDOC
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
