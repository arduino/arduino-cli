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
	"path/filepath"
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
	"github.com/arduino/arduino-cli/configuration"
	"github.com/arduino/arduino-cli/i18n"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	"github.com/arduino/arduino-cli/version"
	paths "github.com/arduino/go-paths-helper"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var tr = i18n.Tr

// CoreInstance is an instance of the Arduino Core Services. The user can
// instantiate as many as needed by providing a different configuration
// for each one.
type CoreInstance struct {
	pm *packagemanager.PackageManager
	lm *librariesmanager.LibrariesManager
}

// coreInstancesContainer has methods to add an remove instances atomically.
type coreInstancesContainer struct {
	instances      map[int32]*CoreInstance
	instancesCount int32
	instancesMux   sync.Mutex
}

// instances contains all the running Arduino Core Services instances
var instances = &coreInstancesContainer{
	instances:      map[int32]*CoreInstance{},
	instancesCount: 1,
}

// GetInstance returns a CoreInstance for the given ID, or nil if ID
// doesn't exist
func (c *coreInstancesContainer) GetInstance(id int32) *CoreInstance {
	c.instancesMux.Lock()
	defer c.instancesMux.Unlock()
	return c.instances[id]
}

// AddAndAssignID saves the CoreInstance and assigns a unique ID to
// retrieve it later
func (c *coreInstancesContainer) AddAndAssignID(i *CoreInstance) int32 {
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
func GetPackageManager(instance rpc.InstanceCommand) *packagemanager.PackageManager {
	i := instances.GetInstance(instance.GetInstance().GetId())
	if i == nil {
		return nil
	}
	return i.pm
}

// GetPackageManagerExplorer returns a new package manager Explorer. The
// explorer holds a read lock on the underlying PackageManager and it should
// be released by calling the returned "release" function.
func GetPackageManagerExplorer(req rpc.InstanceCommand) (explorer *packagemanager.Explorer, release func()) {
	pm := GetPackageManager(req)
	if pm == nil {
		return nil, nil
	}
	return pm.NewExplorer()
}

// GetLibraryManager returns the library manager for the given instance.
func GetLibraryManager(req rpc.InstanceCommand) *librariesmanager.LibrariesManager {
	i := instances.GetInstance(req.GetInstance().GetId())
	if i == nil {
		return nil
	}
	return i.lm
}

func installTool(pm *packagemanager.PackageManager, tool *cores.ToolRelease, downloadCB rpc.DownloadProgressCB, taskCB rpc.TaskProgressCB) error {
	pme, release := pm.NewExplorer()
	defer release()
	taskCB(&rpc.TaskProgress{Name: tr("Downloading missing tool %s", tool)})
	if err := pme.DownloadToolRelease(tool, nil, downloadCB); err != nil {
		return fmt.Errorf(tr("downloading %[1]s tool: %[2]s"), tool, err)
	}
	taskCB(&rpc.TaskProgress{Completed: true})
	if err := pme.InstallTool(tool, taskCB, true); err != nil {
		return fmt.Errorf(tr("installing %[1]s tool: %[2]s"), tool, err)
	}
	return nil
}

// Create a new CoreInstance ready to be initialized, supporting directories are also created.
func Create(req *rpc.CreateRequest, extraUserAgent ...string) (*rpc.CreateResponse, error) {
	instance := &CoreInstance{}

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

	return &rpc.CreateResponse{
		Instance: &rpc.Instance{Id: instanceID},
	}, nil
}

type rpcStatusConverter interface {
	ToRPCStatus() *status.Status
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
	instance := instances.GetInstance(reqInst.GetId())
	if instance == nil {
		return &arduino.InvalidInstanceError{}
	}

	// Setup callback functions
	if responseCallback == nil {
		responseCallback = func(r *rpc.InitResponse) {}
	}

	responseError := func(e rpcStatusConverter) {
		responseCallback(&rpc.InitResponse{
			Message: &rpc.InitResponse_Error{
				Error: e.ToRPCStatus().Proto(),
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

	{
		// We need to rebuild the PackageManager currently in use by this instance
		// in case this is not the first Init on this instances, that might happen
		// after reinitializing an instance after installing or uninstalling a core.
		// If this is not done the information of the uninstall core is kept in memory,
		// even if it should not.
		pmb, commitPackageManager := instance.pm.NewBuilder()

		loadBuiltinTools := func() []error {
			builtinPackage := pmb.GetOrCreatePackage("builtin")
			return pmb.LoadToolsFromPackageDir(builtinPackage, pmb.PackagesDir.Join("builtin", "tools"))
		}

		// Load Platforms
		if profile == nil {
			for _, err := range pmb.LoadHardware() {
				responseError(&arduino.InitFailedError{
					Code:   codes.Internal,
					Cause:  err,
					Reason: rpc.FailedInstanceInitReason_FAILED_INSTANCE_INIT_REASON_TOOL_LOAD_ERROR,
				})
			}
		} else {
			// Load platforms from profile
			errs := pmb.LoadHardwareForProfile(
				profile, true, downloadCallback, taskCallback,
			)
			for _, err := range errs {
				responseError(&arduino.PlatformLoadingError{Cause: err})
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
				responseError(&arduino.InitFailedError{
					Code:   codes.InvalidArgument,
					Cause:  fmt.Errorf(tr("Invalid additional URL: %v", err)),
					Reason: rpc.FailedInstanceInitReason_FAILED_INSTANCE_INIT_REASON_INVALID_INDEX_URL,
				})
				continue
			}

			var loadFunc func(*url.URL) error = pmb.LoadPackageIndex
			if URL.Scheme == "file" {
				loadFunc = func(u *url.URL) error {
					_, err := pmb.LoadPackageIndexFromFile(paths.New(URL.Path))
					return err
				}
			}

			if err := loadFunc(URL); err != nil {
				responseError(&arduino.InitFailedError{
					Code:   codes.FailedPrecondition,
					Cause:  fmt.Errorf(tr("Loading index file: %v", err)),
					Reason: rpc.FailedInstanceInitReason_FAILED_INSTANCE_INIT_REASON_INDEX_LOAD_ERROR,
				})
			}
		}

		// We load hardware before verifying builtin tools are installed
		// otherwise we wouldn't find them and reinstall them each time
		// and they would never get reloaded.

		builtinToolsToInstall := []*cores.ToolRelease{}
		for name, tool := range pmb.GetOrCreatePackage("builtin").Tools {
			latest := tool.LatestRelease()
			if latest == nil {
				responseError(&arduino.InitFailedError{
					Code:   codes.Internal,
					Cause:  fmt.Errorf(tr("can't find latest release of tool %s", name)),
					Reason: rpc.FailedInstanceInitReason_FAILED_INSTANCE_INIT_REASON_TOOL_LOAD_ERROR,
				})
			} else if !latest.IsInstalled() {
				builtinToolsToInstall = append(builtinToolsToInstall, latest)
			}
		}

		// Install builtin tools if necessary
		if len(builtinToolsToInstall) > 0 {
			for _, toolRelease := range builtinToolsToInstall {
				if err := installTool(pmb.Build(), toolRelease, downloadCallback, taskCallback); err != nil {
					responseError(&arduino.InitFailedError{
						Code:   codes.Internal,
						Cause:  err,
						Reason: rpc.FailedInstanceInitReason_FAILED_INSTANCE_INIT_REASON_TOOL_LOAD_ERROR,
					})
				}
			}

			// We installed at least one builtin tool after loading hardware
			// so we must reload again otherwise we would never found them.
			for _, err := range loadBuiltinTools() {
				responseError(&arduino.InitFailedError{
					Code:   codes.Internal,
					Cause:  err,
					Reason: rpc.FailedInstanceInitReason_FAILED_INSTANCE_INIT_REASON_TOOL_LOAD_ERROR,
				})
			}
		}

		commitPackageManager()
	}

	pme, release := instance.pm.NewExplorer()
	defer release()

	for _, err := range pme.LoadDiscoveries() {
		responseError(&arduino.PlatformLoadingError{Cause: err})
	}

	// Create library manager and add libraries directories
	lm := librariesmanager.NewLibraryManager(
		pme.IndexDir,
		pme.DownloadDir,
	)
	instance.lm = lm

	// Load libraries
	for _, pack := range pme.GetPackages() {
		for _, platform := range pack.Platforms {
			if platformRelease := pme.GetInstalledPlatformRelease(platform); platformRelease != nil {
				lm.AddPlatformReleaseLibrariesDir(platformRelease, libraries.PlatformBuiltIn)
			}
		}
	}

	if err := lm.LoadIndex(); err != nil {
		responseError(&arduino.InitFailedError{
			Code:   codes.FailedPrecondition,
			Cause:  fmt.Errorf(tr("Loading index file: %v", err)),
			Reason: rpc.FailedInstanceInitReason_FAILED_INSTANCE_INIT_REASON_INDEX_LOAD_ERROR,
		})
	}

	if profile == nil {
		// Add directories of libraries bundled with IDE
		if bundledLibsDir := configuration.IDEBuiltinLibrariesDir(configuration.Settings); bundledLibsDir != nil {
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
					responseError(&arduino.LibraryNotFoundError{Library: libraryRef.Library})
					continue
				}
				if err := libRelease.Resource.Download(lm.DownloadsDir, nil, libRelease.String(), downloadCallback, ""); err != nil {
					taskCallback(&rpc.TaskProgress{Name: tr("Error downloading library %s", libraryRef)})
					responseError(&arduino.FailedLibraryInstallError{Cause: err})
					continue
				}
				taskCallback(&rpc.TaskProgress{Completed: true})

				// Install library
				taskCallback(&rpc.TaskProgress{Name: tr("Installing library %s", libraryRef)})
				if err := libRelease.Resource.Install(lm.DownloadsDir, libRoot, libDir); err != nil {
					taskCallback(&rpc.TaskProgress{Name: tr("Error installing library %s", libraryRef)})
					responseError(&arduino.FailedLibraryInstallError{Cause: err})
					continue
				}
				taskCallback(&rpc.TaskProgress{Completed: true})
			}

			lm.AddLibrariesDir(libRoot, libraries.User)
		}
	}

	for _, err := range lm.RescanLibraries() {
		responseError(&arduino.InitFailedError{
			Code:   codes.FailedPrecondition,
			Cause:  fmt.Errorf(tr("Loading libraries: %v"), err),
			Reason: rpc.FailedInstanceInitReason_FAILED_INSTANCE_INIT_REASON_LIBRARY_LOAD_ERROR,
		})
	}

	// Refreshes the locale used, this will change the
	// language of the CLI if the locale is different
	// after started.
	i18n.Init(configuration.Settings.GetString("locale"))

	return nil
}

// Destroy FIXMEDOC
func Destroy(ctx context.Context, req *rpc.DestroyRequest) (*rpc.DestroyResponse, error) {
	if ok := instances.RemoveID(req.GetInstance().GetId()); !ok {
		return nil, &arduino.InvalidInstanceError{}
	}
	return &rpc.DestroyResponse{}, nil
}

// UpdateLibrariesIndex updates the library_index.json
func UpdateLibrariesIndex(ctx context.Context, req *rpc.UpdateLibrariesIndexRequest, downloadCB rpc.DownloadProgressCB) error {
	logrus.Info("Updating libraries index")
	lm := GetLibraryManager(req)
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
		URL: librariesmanager.LibraryIndexWithSignatureArchiveURL,
	}
	if err := indexResource.Download(lm.IndexFile.Parent(), downloadCB); err != nil {
		return err
	}

	return nil
}

// UpdateIndex FIXMEDOC
func UpdateIndex(ctx context.Context, req *rpc.UpdateIndexRequest, downloadCB rpc.DownloadProgressCB) error {
	if instances.GetInstance(req.GetInstance().GetId()) == nil {
		return &arduino.InvalidInstanceError{}
	}

	indexpath := configuration.DataDir(configuration.Settings)

	urls := []string{globals.DefaultIndexURL}
	if !req.GetIgnoreCustomPackageIndexes() {
		urls = append(urls, configuration.Settings.GetStringSlice("board_manager.additional_urls")...)
	}

	failed := false
	for _, u := range urls {
		URL, err := utils.URLParse(u)
		if err != nil {
			logrus.Warnf("unable to parse additional URL: %s", u)
			msg := fmt.Sprintf("%s: %v", tr("Unable to parse URL"), err)
			downloadCB.Start(u, tr("Downloading index: %s", u))
			downloadCB.End(false, msg)
			failed = true
			continue
		}

		logrus.WithField("url", URL).Print("Updating index")

		if URL.Scheme == "file" {
			downloadCB.Start(u, tr("Downloading index: %s", filepath.Base(URL.Path)))
			path := paths.New(URL.Path)
			if _, err := packageindex.LoadIndexNoSign(path); err != nil {
				msg := fmt.Sprintf("%s: %v", tr("Invalid package index in %s", path), err)
				downloadCB.End(false, msg)
				failed = true
			} else {
				downloadCB.End(true, "")
			}
			continue
		}

		indexResource := resources.IndexResource{URL: URL}
		if strings.HasSuffix(URL.Host, "arduino.cc") && strings.HasSuffix(URL.Path, ".json") {
			indexResource.SignatureURL, _ = url.Parse(u) // should not fail because we already parsed it
			indexResource.SignatureURL.Path += ".sig"
		}
		if err := indexResource.Download(indexpath, downloadCB); err != nil {
			failed = true
		}
	}

	if failed {
		return &arduino.FailedDownloadError{Message: tr("Some indexes could not be updated.")}
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
