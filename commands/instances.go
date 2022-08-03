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
	"net/url"
	"os"
	"strings"
	"sync"

	"github.com/arduino/arduino-cli/arduino"
	"github.com/arduino/arduino-cli/arduino/cores"
	"github.com/arduino/arduino-cli/arduino/cores/packageindex"
	"github.com/arduino/arduino-cli/arduino/cores/packagemanager"
	"github.com/arduino/arduino-cli/arduino/globals"
	"github.com/arduino/arduino-cli/arduino/libraries"
	"github.com/arduino/arduino-cli/arduino/libraries/librariesindex"
	"github.com/arduino/arduino-cli/arduino/libraries/librariesmanager"
	"github.com/arduino/arduino-cli/arduino/resources"
	"github.com/arduino/arduino-cli/arduino/sketch"
	"github.com/arduino/arduino-cli/arduino/utils"
	cliglobals "github.com/arduino/arduino-cli/cli/globals"
	"github.com/arduino/arduino-cli/configuration"
	"github.com/arduino/arduino-cli/i18n"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	paths "github.com/arduino/go-paths-helper"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var tr = i18n.Tr

// this map contains all the running Arduino Core Services instances
// referenced by an int32 handle
var instances = map[int32]*CoreInstance{}
var instancesCount int32 = 1
var instancesMux sync.Mutex

// CoreInstance is an instance of the Arduino Core Services. The user can
// instantiate as many as needed by providing a different configuration
// for each one.
type CoreInstance struct {
	PackageManager *packagemanager.PackageManager
	lm             *librariesmanager.LibrariesManager
}

// InstanceContainer FIXMEDOC
type InstanceContainer interface {
	GetInstance() *rpc.Instance
}

// GetInstance returns a CoreInstance for the given ID, or nil if ID
// doesn't exist
func GetInstance(id int32) *CoreInstance {
	instancesMux.Lock()
	defer instancesMux.Unlock()
	return instances[id]
}

// GetPackageManager returns a PackageManager for the given ID, or nil if
// ID doesn't exist
func GetPackageManager(id int32) *packagemanager.PackageManager {
	i := GetInstance(id)
	if i == nil {
		return nil
	}
	return i.PackageManager
}

// GetLibraryManager returns the library manager for the given instance ID
func GetLibraryManager(id int32) *librariesmanager.LibrariesManager {
	i := GetInstance(id)
	if i == nil {
		return nil
	}
	return i.lm
}

func (instance *CoreInstance) installToolIfMissing(tool *cores.ToolRelease, downloadCB rpc.DownloadProgressCB, taskCB rpc.TaskProgressCB) (bool, error) {
	if tool.IsInstalled() {
		return false, nil
	}
	taskCB(&rpc.TaskProgress{Name: tr("Downloading missing tool %s", tool)})
	if err := instance.PackageManager.DownloadToolRelease(tool, nil, downloadCB); err != nil {
		return false, fmt.Errorf(tr("downloading %[1]s tool: %[2]s"), tool, err)
	}
	taskCB(&rpc.TaskProgress{Completed: true})
	if err := instance.PackageManager.InstallTool(tool, taskCB); err != nil {
		return false, fmt.Errorf(tr("installing %[1]s tool: %[2]s"), tool, err)
	}
	return true, nil
}

// Create a new CoreInstance ready to be initialized, supporting directories are also created.
func Create(req *rpc.CreateRequest, extraUserAgent ...string) (*rpc.CreateResponse, error) {
	instance := &CoreInstance{}

	// Setup downloads directory
	downloadsDir := paths.New(configuration.Settings.GetString("directories.Downloads"))
	if downloadsDir.NotExist() {
		err := downloadsDir.MkdirAll()
		if err != nil {
			return nil, &arduino.PermissionDeniedError{Message: tr("Failed to create downloads directory"), Cause: err}
		}
	}

	// Setup data directory
	dataDir := paths.New(configuration.Settings.GetString("directories.Data"))
	packagesDir := configuration.PackagesDir(configuration.Settings)
	if packagesDir.NotExist() {
		err := packagesDir.MkdirAll()
		if err != nil {
			return nil, &arduino.PermissionDeniedError{Message: tr("Failed to create data directory"), Cause: err}
		}
	}

	// Create package manager
	userAgent := "arduino-cli/" + cliglobals.VersionInfo.VersionString
	for _, ua := range extraUserAgent {
		userAgent += " " + ua
	}
	instance.PackageManager = packagemanager.NewPackageManager(
		dataDir,
		configuration.PackagesDir(configuration.Settings),
		downloadsDir,
		dataDir.Join("tmp"),
		userAgent,
	)
	instance.lm = librariesmanager.NewLibraryManager(
		dataDir,
		downloadsDir,
	)

	// Save instance
	instancesMux.Lock()
	instanceID := instancesCount
	instances[instanceID] = instance
	instancesCount++
	instancesMux.Unlock()

	return &rpc.CreateResponse{
		Instance: &rpc.Instance{Id: instanceID},
	}, nil
}

// Init loads installed libraries and Platforms in CoreInstance with specified ID,
// a gRPC status error is returned if the CoreInstance doesn't exist.
// All responses are sent through responseCallback, can be nil to ignore all responses.
// Failures don't stop the loading process, in case of loading failure the Platform or library
// is simply skipped and an error gRPC status is sent to responseCallback.
func Init(req *rpc.InitRequest, responseCallback func(r *rpc.InitResponse)) error {
	if responseCallback == nil {
		responseCallback = func(r *rpc.InitResponse) {}
	}
	reqInst := req.GetInstance()
	if reqInst == nil {
		return &arduino.InvalidInstanceError{}
	}
	instance := GetInstance(reqInst.GetId())
	if instance == nil {
		return &arduino.InvalidInstanceError{}
	}

	// Setup callback functions
	if responseCallback == nil {
		responseCallback = func(r *rpc.InitResponse) {}
	}
	responseError := func(st *status.Status) {
		responseCallback(&rpc.InitResponse{
			Message: &rpc.InitResponse_Error{
				Error: st.Proto(),
			},
		})
	}
	taskCallback := func(msg *rpc.TaskProgress) {
		responseCallback(&rpc.InitResponse{
			Message: &rpc.InitResponse_InitProgress{
				InitProgress: &rpc.InitResponse_Progress{
					TaskProgress: msg,
				},
			},
		})
	}
	downloadCallback := func(msg *rpc.DownloadProgress) {
		responseCallback(&rpc.InitResponse{
			Message: &rpc.InitResponse_InitProgress{
				InitProgress: &rpc.InitResponse_Progress{
					DownloadProgress: msg,
				},
			},
		})
	}

	// We need to clear the PackageManager currently in use by this instance
	// in case this is not the first Init on this instances, that might happen
	// after reinitializing an instance after installing or uninstalling a core.
	// If this is not done the information of the uninstall core is kept in memory,
	// even if it should not.
	pm := instance.PackageManager
	pm.Clear()

	// Try to extract profile if specified
	var profile *sketch.Profile
	if req.GetProfile() != "" {
		sk, err := sketch.New(paths.New(req.GetSketchPath()))
		if err != nil {
			return &arduino.InvalidArgumentError{Cause: err}
		}
		profile = sk.GetProfile(req.GetProfile())
		if profile == nil {
			return &arduino.UnknownProfileError{Profile: req.GetProfile()}
		}
		responseCallback(&rpc.InitResponse{
			Message: &rpc.InitResponse_Profile{
				Profile: &rpc.Profile{
					Name: req.GetProfile(),
					Fqbn: profile.FQBN,
					// TODO: Other profile infos may be provided here...
				},
			},
		})
	}

	loadBuiltinTools := func() []error {
		builtinPackage := pm.Packages.GetOrCreatePackage("builtin")
		return pm.LoadToolsFromPackageDir(builtinPackage, pm.PackagesDir.Join("builtin", "tools"))
	}

	// Load Platforms
	if profile == nil {
		for _, err := range pm.LoadHardware() {
			s := &arduino.PlatformLoadingError{Cause: err}
			responseError(s.ToRPCStatus())
		}
	} else {
		// Load platforms from profile
		errs := pm.LoadHardwareForProfile(
			profile, true, downloadCallback, taskCallback,
		)
		for _, err := range errs {
			s := &arduino.PlatformLoadingError{Cause: err}
			responseError(s.ToRPCStatus())
		}

		// Load "builtin" tools
		_ = loadBuiltinTools()
	}

	// Load packages index
	urls := []string{globals.DefaultIndexURL}
	if profile == nil {
		urls = append(urls, configuration.Settings.GetStringSlice("board_manager.additional_urls")...)
	}
	for _, u := range urls {
		URL, err := utils.URLParse(u)
		if err != nil {
			s := status.Newf(codes.InvalidArgument, tr("Invalid additional URL: %v"), err)
			responseError(s)
			continue
		}

		if URL.Scheme == "file" {
			_, err := pm.LoadPackageIndexFromFile(paths.New(URL.Path))
			if err != nil {
				s := status.Newf(codes.FailedPrecondition, tr("Loading index file: %v"), err)
				responseError(s)
			}
			continue
		}

		if err := pm.LoadPackageIndex(URL); err != nil {
			s := status.Newf(codes.FailedPrecondition, tr("Loading index file: %v"), err)
			responseError(s)
		}
	}

	// We load hardware before verifying builtin tools are installed
	// otherwise we wouldn't find them and reinstall them each time
	// and they would never get reloaded.

	builtinToolReleases := []*cores.ToolRelease{}
	for name, tool := range pm.Packages.GetOrCreatePackage("builtin").Tools {
		latestRelease := tool.LatestRelease()
		if latestRelease == nil {
			s := status.Newf(codes.Internal, tr("can't find latest release of tool %s", name))
			responseError(s)
			continue
		}
		builtinToolReleases = append(builtinToolReleases, latestRelease)
	}

	// Install tools if necessary
	toolsHaveBeenInstalled := false
	for _, toolRelease := range builtinToolReleases {
		installed, err := instance.installToolIfMissing(toolRelease, downloadCallback, taskCallback)
		if err != nil {
			s := status.Newf(codes.Internal, err.Error())
			responseError(s)
			continue
		}
		toolsHaveBeenInstalled = toolsHaveBeenInstalled || installed
	}

	if toolsHaveBeenInstalled {
		// We installed at least one new tool after loading hardware
		// so we must reload again otherwise we would never found them.
		for _, err := range loadBuiltinTools() {
			s := &arduino.PlatformLoadingError{Cause: err}
			responseError(s.ToRPCStatus())
		}
	}

	for _, err := range pm.LoadDiscoveries() {
		s := &arduino.PlatformLoadingError{Cause: err}
		responseError(s.ToRPCStatus())
	}

	// Create library manager and add libraries directories
	lm := librariesmanager.NewLibraryManager(
		pm.IndexDir,
		pm.DownloadDir,
	)
	instance.lm = lm

	// Load libraries
	for _, pack := range pm.Packages {
		for _, platform := range pack.Platforms {
			if platformRelease := pm.GetInstalledPlatformRelease(platform); platformRelease != nil {
				lm.AddPlatformReleaseLibrariesDir(platformRelease, libraries.PlatformBuiltIn)
			}
		}
	}

	if err := lm.LoadIndex(); err != nil {
		s := status.Newf(codes.FailedPrecondition, tr("Loading index file: %v"), err)
		responseError(s)
	}

	if profile == nil {
		// Add directories of libraries bundled with IDE
		if bundledLibsDir := configuration.IDEBundledLibrariesDir(configuration.Settings); bundledLibsDir != nil {
			lm.AddLibrariesDir(bundledLibsDir, libraries.IDEBuiltIn)
		}

		// Add libraries directory from config file
		lm.AddLibrariesDir(configuration.LibrariesDir(configuration.Settings), libraries.User)
	} else {
		// Load libraries required for profile
		for _, libraryRef := range profile.Libraries {
			uid := libraryRef.InternalUniqueIdentifier()
			libRoot := configuration.ProfilesCacheDir(configuration.Settings).Join(uid)
			libDir := libRoot.Join(libraryRef.Library)

			if !libDir.IsDir() {
				// Download library
				taskCallback(&rpc.TaskProgress{Name: tr("Downloading library %s", libraryRef)})
				libRelease := lm.Index.FindRelease(&librariesindex.Reference{
					Name:    libraryRef.Library,
					Version: libraryRef.Version,
				})
				if libRelease == nil {
					taskCallback(&rpc.TaskProgress{Name: tr("Library %s not found", libraryRef)})
					err := &arduino.LibraryNotFoundError{Library: libraryRef.Library}
					responseError(err.ToRPCStatus())
					continue
				}
				if err := libRelease.Resource.Download(lm.DownloadsDir, nil, libRelease.String(), downloadCallback); err != nil {
					taskCallback(&rpc.TaskProgress{Name: tr("Error downloading library %s", libraryRef)})
					e := &arduino.FailedLibraryInstallError{Cause: err}
					responseError(e.ToRPCStatus())
					continue
				}
				taskCallback(&rpc.TaskProgress{Completed: true})

				// Install library
				taskCallback(&rpc.TaskProgress{Name: tr("Installing library %s", libraryRef)})
				if err := libRelease.Resource.Install(lm.DownloadsDir, libRoot, libDir); err != nil {
					taskCallback(&rpc.TaskProgress{Name: tr("Error installing library %s", libraryRef)})
					e := &arduino.FailedLibraryInstallError{Cause: err}
					responseError(e.ToRPCStatus())
					continue
				}
				taskCallback(&rpc.TaskProgress{Completed: true})
			}

			lm.AddLibrariesDir(libRoot, libraries.User)
		}
	}

	for _, err := range lm.RescanLibraries() {
		s := status.Newf(codes.FailedPrecondition, tr("Loading libraries: %v"), err)
		responseError(s)
	}

	// Refreshes the locale used, this will change the
	// language of the CLI if the locale is different
	// after started.
	i18n.Init(configuration.Settings.GetString("locale"))

	return nil
}

// Destroy FIXMEDOC
func Destroy(ctx context.Context, req *rpc.DestroyRequest) (*rpc.DestroyResponse, error) {
	id := req.GetInstance().GetId()

	instancesMux.Lock()
	defer instancesMux.Unlock()
	if _, ok := instances[id]; !ok {
		return nil, &arduino.InvalidInstanceError{}
	}
	delete(instances, id)
	return &rpc.DestroyResponse{}, nil
}

// UpdateLibrariesIndex updates the library_index.json
func UpdateLibrariesIndex(ctx context.Context, req *rpc.UpdateLibrariesIndexRequest, downloadCB rpc.DownloadProgressCB) error {
	logrus.Info("Updating libraries index")
	lm := GetLibraryManager(req.GetInstance().GetId())
	if lm == nil {
		return &arduino.InvalidInstanceError{}
	}

	if err := lm.IndexFile.Parent().MkdirAll(); err != nil {
		return &arduino.PermissionDeniedError{Message: tr("Could not create index directory"), Cause: err}
	}

	// Create a temp dir to stage all downloads
	tmp, err := paths.MkTempDir("", "library_index_download")
	if err != nil {
		return &arduino.TempDirCreationFailedError{Cause: err}
	}
	defer tmp.RemoveAll()

	indexResource := resources.IndexResource{
		URL:          librariesmanager.LibraryIndexGZURL,
		SignatureURL: librariesmanager.LibraryIndexSignature,
	}
	if err := indexResource.Download(lm.IndexFile.Parent(), downloadCB); err != nil {
		return err
	}

	return nil
}

// UpdateIndex FIXMEDOC
func UpdateIndex(ctx context.Context, req *rpc.UpdateIndexRequest, downloadCB rpc.DownloadProgressCB) (*rpc.UpdateIndexResponse, error) {
	if GetInstance(req.GetInstance().GetId()) == nil {
		return nil, &arduino.InvalidInstanceError{}
	}

	indexpath := paths.New(configuration.Settings.GetString("directories.Data"))

	urls := []string{globals.DefaultIndexURL}
	urls = append(urls, configuration.Settings.GetStringSlice("board_manager.additional_urls")...)
	for _, u := range urls {
		logrus.Info("URL: ", u)
		URL, err := utils.URLParse(u)
		if err != nil {
			logrus.Warnf("unable to parse additional URL: %s", u)
			continue
		}

		logrus.WithField("url", URL).Print("Updating index")

		if URL.Scheme == "file" {
			path := paths.New(URL.Path)
			if _, err := packageindex.LoadIndexNoSign(path); err != nil {
				return nil, &arduino.InvalidArgumentError{Message: tr("Invalid package index in %s", path), Cause: err}
			}

			fi, _ := os.Stat(path.String())
			downloadCB(&rpc.DownloadProgress{
				File:      tr("Downloading index: %s", path.Base()),
				TotalSize: fi.Size(),
			})
			downloadCB(&rpc.DownloadProgress{Completed: true})
			continue
		}

		indexResource := resources.IndexResource{
			URL: URL,
		}
		if strings.HasSuffix(URL.Host, "arduino.cc") && strings.HasSuffix(URL.Path, ".json") {
			indexResource.SignatureURL, _ = url.Parse(u) // should not fail because we already parsed it
			indexResource.SignatureURL.Path += ".sig"
		}
		if err := indexResource.Download(indexpath, downloadCB); err != nil {
			return nil, err
		}
	}

	return &rpc.UpdateIndexResponse{}, nil
}

// UpdateCoreLibrariesIndex updates both Cores and Libraries indexes
func UpdateCoreLibrariesIndex(ctx context.Context, req *rpc.UpdateCoreLibrariesIndexRequest, downloadCB rpc.DownloadProgressCB) error {
	_, err := UpdateIndex(ctx, &rpc.UpdateIndexRequest{
		Instance: req.Instance,
	}, downloadCB)
	if err != nil {
		return err
	}

	err = UpdateLibrariesIndex(ctx, &rpc.UpdateLibrariesIndexRequest{
		Instance: req.Instance,
	}, downloadCB)
	if err != nil {
		return err
	}

	return nil
}

// LoadSketch collects and returns all files composing a sketch
func LoadSketch(ctx context.Context, req *rpc.LoadSketchRequest) (*rpc.LoadSketchResponse, error) {
	// TODO: This should be a ToRpc function for the Sketch struct
	sk, err := sketch.New(paths.New(req.SketchPath))
	if err != nil {
		return nil, &arduino.CantOpenSketchError{Cause: err}
	}

	otherSketchFiles := make([]string, sk.OtherSketchFiles.Len())
	for i, file := range sk.OtherSketchFiles {
		otherSketchFiles[i] = file.String()
	}

	additionalFiles := make([]string, sk.AdditionalFiles.Len())
	for i, file := range sk.AdditionalFiles {
		additionalFiles[i] = file.String()
	}

	rootFolderFiles := make([]string, sk.RootFolderFiles.Len())
	for i, file := range sk.RootFolderFiles {
		rootFolderFiles[i] = file.String()
	}

	return &rpc.LoadSketchResponse{
		MainFile:         sk.MainFile.String(),
		LocationPath:     sk.FullPath.String(),
		OtherSketchFiles: otherSketchFiles,
		AdditionalFiles:  additionalFiles,
		RootFolderFiles:  rootFolderFiles,
	}, nil
}
