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
	"errors"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path"

	"github.com/arduino/arduino-cli/arduino/builder"
	"github.com/arduino/arduino-cli/arduino/cores"
	"github.com/arduino/arduino-cli/arduino/cores/packageindex"
	"github.com/arduino/arduino-cli/arduino/cores/packagemanager"
	"github.com/arduino/arduino-cli/arduino/libraries"
	"github.com/arduino/arduino-cli/arduino/libraries/librariesindex"
	"github.com/arduino/arduino-cli/arduino/libraries/librariesmanager"
	"github.com/arduino/arduino-cli/arduino/security"
	"github.com/arduino/arduino-cli/arduino/utils"
	"github.com/arduino/arduino-cli/cli/globals"
	"github.com/arduino/arduino-cli/configuration"
	rpc "github.com/arduino/arduino-cli/rpc/commands"
	paths "github.com/arduino/go-paths-helper"
	"github.com/sirupsen/logrus"
	"go.bug.st/downloader/v2"
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

func (instance *CoreInstance) installToolIfMissing(tool *cores.ToolRelease, downloadCB DownloadProgressCB, taskCB TaskProgressCB) (bool, error) {
	if tool.IsInstalled() {
		return false, nil
	}
	taskCB(&rpc.TaskProgress{Name: "Downloading missing tool " + tool.String()})
	if err := DownloadToolRelease(instance.PackageManager, tool, downloadCB); err != nil {
		return false, fmt.Errorf("downloading %s tool: %s", tool, err)
	}
	taskCB(&rpc.TaskProgress{Completed: true})
	if err := InstallToolRelease(instance.PackageManager, tool, taskCB); err != nil {
		return false, fmt.Errorf("installing %s tool: %s", tool, err)
	}
	return true, nil
}

func (instance *CoreInstance) checkForBuiltinTools(downloadCB DownloadProgressCB, taskCB TaskProgressCB) error {
	// Check for ctags tool
	ctags, _ := getBuiltinCtagsTool(instance.PackageManager)
	ctagsInstalled, err := instance.installToolIfMissing(ctags, downloadCB, taskCB)
	if err != nil {
		return err
	}

	// Check for builtin serial-discovery tool
	serialDiscoveryTool, _ := getBuiltinSerialDiscoveryTool(instance.PackageManager)
	serialDiscoveryInstalled, err := instance.installToolIfMissing(serialDiscoveryTool, downloadCB, taskCB)
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
func Init(ctx context.Context, req *rpc.InitReq, downloadCB DownloadProgressCB, taskCB TaskProgressCB) (*rpc.InitResp, error) {
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

	if err := instance.checkForBuiltinTools(downloadCB, taskCB); err != nil {
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
	config, err := GetDownloaderConfig()
	if err != nil {
		return err
	}
	d, err := lm.UpdateIndex(config)
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
				return nil, fmt.Errorf("invalid package index in %s: %s", path, err)
			}

			fi, _ := os.Stat(path.String())
			downloadCB(&rpc.DownloadProgress{
				File:      "Updating index: " + path.Base(),
				TotalSize: fi.Size(),
			})
			downloadCB(&rpc.DownloadProgress{Completed: true})
			continue
		}

		var tmp *paths.Path
		if tmpFile, err := ioutil.TempFile("", ""); err != nil {
			return nil, fmt.Errorf("creating temp file for index download: %s", err)
		} else if err := tmpFile.Close(); err != nil {
			return nil, fmt.Errorf("creating temp file for index download: %s", err)
		} else {
			tmp = paths.New(tmpFile.Name())
		}
		defer tmp.Remove()

		config, err := GetDownloaderConfig()
		if err != nil {
			return nil, fmt.Errorf("downloading index %s: %s", URL, err)
		}
		d, err := downloader.DownloadWithConfig(tmp.String(), URL.String(), *config)
		if err != nil {
			return nil, fmt.Errorf("downloading index %s: %s", URL, err)
		}
		coreIndexPath := indexpath.Join(path.Base(URL.Path))
		err = Download(d, "Updating index: "+coreIndexPath.Base(), downloadCB)
		if err != nil {
			return nil, fmt.Errorf("downloading index %s: %s", URL, err)
		}

		// Check for signature
		var tmpSig *paths.Path
		var coreIndexSigPath *paths.Path
		if URL.Hostname() == "downloads.arduino.cc" {
			URLSig, err := url.Parse(URL.String())
			if err != nil {
				return nil, fmt.Errorf("parsing url for index signature check: %s", err)
			}
			URLSig.Path += ".sig"

			if t, err := ioutil.TempFile("", ""); err != nil {
				return nil, fmt.Errorf("creating temp file for index signature download: %s", err)
			} else if err := t.Close(); err != nil {
				return nil, fmt.Errorf("creating temp file for index signature download: %s", err)
			} else {
				tmpSig = paths.New(t.Name())
			}
			defer tmpSig.Remove()

			d, err := downloader.DownloadWithConfig(tmpSig.String(), URLSig.String(), *config)
			if err != nil {
				return nil, fmt.Errorf("downloading index signature %s: %s", URLSig, err)
			}

			coreIndexSigPath = indexpath.Join(path.Base(URLSig.Path))
			Download(d, "Updating index: "+coreIndexSigPath.Base(), downloadCB)
			if d.Error() != nil {
				return nil, fmt.Errorf("downloading index signature %s: %s", URL, d.Error())
			}

			valid, _, err := security.VerifyArduinoDetachedSignature(tmp, tmpSig)
			if err != nil {
				return nil, fmt.Errorf("signature verification error: %s", err)
			}
			if !valid {
				return nil, fmt.Errorf("index has an invalid signature")
			}
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
		if tmpSig != nil {
			if err := tmpSig.CopyTo(coreIndexSigPath); err != nil {
				return nil, fmt.Errorf("saving downloaded index signature: %s", err)
			}
		}
	}
	if _, err := Rescan(id); err != nil {
		return nil, fmt.Errorf("rescanning filesystem: %s", err)
	}
	return &rpc.UpdateIndexResp{}, nil
}

// UpdateCoreLibrariesIndex updates both Cores and Libraries indexes
func UpdateCoreLibrariesIndex(ctx context.Context, req *rpc.UpdateCoreLibrariesIndexReq, downloadCB DownloadProgressCB) error {
	_, err := UpdateIndex(ctx, &rpc.UpdateIndexReq{
		Instance: req.Instance,
	}, downloadCB)
	if err != nil {
		return err
	}

	err = UpdateLibrariesIndex(ctx, &rpc.UpdateLibrariesIndexReq{
		Instance: req.Instance,
	}, downloadCB)
	if err != nil {
		return err
	}

	return nil
}

// Outdated returns a list struct containing both Core and Libraries that can be updated
func Outdated(ctx context.Context, req *rpc.OutdatedReq) (*rpc.OutdatedResp, error) {
	id := req.GetInstance().GetId()

	libraryManager := GetLibraryManager(id)
	if libraryManager == nil {
		return nil, errors.New("invalid instance")
	}

	outdatedLibraries := []*rpc.InstalledLibrary{}
	for _, libAlternatives := range libraryManager.Libraries {
		for _, library := range libAlternatives.Alternatives {
			if library.Location != libraries.User {
				continue
			}
			available := libraryManager.Index.FindLibraryUpdate(library)
			if available == nil {
				continue
			}

			outdatedLibraries = append(outdatedLibraries, &rpc.InstalledLibrary{
				Library: getOutputLibrary(library),
				Release: getOutputRelease(available),
			})
		}
	}

	packageManager := GetPackageManager(id)
	if packageManager == nil {
		return nil, errors.New("invalid instance")
	}

	outdatedPlatforms := []*rpc.Platform{}
	for _, targetPackage := range packageManager.Packages {
		for _, installed := range targetPackage.Platforms {
			if installedRelease := packageManager.GetInstalledPlatformRelease(installed); installedRelease != nil {
				latest := installed.GetLatestRelease()
				if latest == nil || latest == installedRelease {
					continue
				}
				rpcPlatform := PlatformReleaseToRPC(latest)
				rpcPlatform.Installed = installedRelease.Version.String()

				outdatedPlatforms = append(
					outdatedPlatforms,
					rpcPlatform,
				)
			}
		}
	}

	return &rpc.OutdatedResp{
		OutdatedLibrary:  outdatedLibraries,
		OutdatedPlatform: outdatedPlatforms,
	}, nil
}

func getOutputLibrary(lib *libraries.Library) *rpc.Library {
	insdir := ""
	if lib.InstallDir != nil {
		insdir = lib.InstallDir.String()
	}
	srcdir := ""
	if lib.SourceDir != nil {
		srcdir = lib.SourceDir.String()
	}
	utldir := ""
	if lib.UtilityDir != nil {
		utldir = lib.UtilityDir.String()
	}
	cntplat := ""
	if lib.ContainerPlatform != nil {
		cntplat = lib.ContainerPlatform.String()
	}

	return &rpc.Library{
		Name:              lib.Name,
		Author:            lib.Author,
		Maintainer:        lib.Maintainer,
		Sentence:          lib.Sentence,
		Paragraph:         lib.Paragraph,
		Website:           lib.Website,
		Category:          lib.Category,
		Architectures:     lib.Architectures,
		Types:             lib.Types,
		InstallDir:        insdir,
		SourceDir:         srcdir,
		UtilityDir:        utldir,
		Location:          lib.Location.ToRPCLibraryLocation(),
		ContainerPlatform: cntplat,
		Layout:            lib.Layout.ToRPCLibraryLayout(),
		RealName:          lib.RealName,
		DotALinkage:       lib.DotALinkage,
		Precompiled:       lib.Precompiled,
		LdFlags:           lib.LDflags,
		IsLegacy:          lib.IsLegacy,
		Version:           lib.Version.String(),
		License:           lib.License,
	}
}

func getOutputRelease(lib *librariesindex.Release) *rpc.LibraryRelease {
	if lib != nil {
		return &rpc.LibraryRelease{
			Author:        lib.Author,
			Version:       lib.Version.String(),
			Maintainer:    lib.Maintainer,
			Sentence:      lib.Sentence,
			Paragraph:     lib.Paragraph,
			Website:       lib.Website,
			Category:      lib.Category,
			Architectures: lib.Architectures,
			Types:         lib.Types,
		}
	}
	return &rpc.LibraryRelease{}
}

// Upgrade downloads and installs outdated Cores and Libraries
func Upgrade(ctx context.Context, req *rpc.UpgradeReq, downloadCB DownloadProgressCB, taskCB TaskProgressCB) error {
	downloaderConfig, err := GetDownloaderConfig()
	if err != nil {
		return err
	}

	lm := GetLibraryManager(req.Instance.Id)
	if lm == nil {
		return fmt.Errorf("invalid handle")
	}

	for _, libAlternatives := range lm.Libraries {
		for _, library := range libAlternatives.Alternatives {
			if library.Location != libraries.User {
				continue
			}
			available := lm.Index.FindLibraryUpdate(library)
			if available == nil {
				continue
			}

			// Downloads latest library release
			taskCB(&rpc.TaskProgress{Name: "Downloading " + available.String()})
			if d, err := available.Resource.Download(lm.DownloadsDir, downloaderConfig); err != nil {
				return err
			} else if err := Download(d, available.String(), downloadCB); err != nil {
				return err
			}

			// Installs downloaded library
			taskCB(&rpc.TaskProgress{Name: "Installing " + available.String()})
			libPath, libReplaced, err := lm.InstallPrerequisiteCheck(available)
			if err == librariesmanager.ErrAlreadyInstalled {
				taskCB(&rpc.TaskProgress{Message: "Already installed " + available.String(), Completed: true})
				continue
			} else if err != nil {
				return fmt.Errorf("checking lib install prerequisites: %s", err)
			}

			if libReplaced != nil {
				taskCB(&rpc.TaskProgress{Message: fmt.Sprintf("Replacing %s with %s", libReplaced, available)})
			}

			if err := lm.Install(available, libPath); err != nil {
				return err
			}

			taskCB(&rpc.TaskProgress{Message: "Installed " + available.String(), Completed: true})
		}
	}

	pm := GetPackageManager(req.Instance.Id)
	if pm == nil {
		return fmt.Errorf("invalid handle")
	}

	for _, targetPackage := range pm.Packages {
		for _, installed := range targetPackage.Platforms {
			if installedRelease := pm.GetInstalledPlatformRelease(installed); installedRelease != nil {
				latest := installed.GetLatestRelease()
				if latest == nil || latest == installedRelease {
					continue
				}

				ref := &packagemanager.PlatformReference{
					Package:              installedRelease.Platform.Package.Name,
					PlatformArchitecture: installedRelease.Platform.Architecture,
					PlatformVersion:      installedRelease.Version,
				}
				// Get list of installed tools needed by the currently installed version
				_, installedTools, err := pm.FindPlatformReleaseDependencies(ref)
				if err != nil {
					return err
				}

				ref = &packagemanager.PlatformReference{
					Package:              latest.Platform.Package.Name,
					PlatformArchitecture: latest.Platform.Architecture,
					PlatformVersion:      latest.Version,
				}

				taskCB(&rpc.TaskProgress{Name: "Downloading " + latest.String()})
				_, tools, err := pm.FindPlatformReleaseDependencies(ref)
				if err != nil {
					return fmt.Errorf("platform %s is not installed", ref)
				}

				toolsToInstall := []*cores.ToolRelease{}
				for _, tool := range tools {
					if tool.IsInstalled() {
						logrus.WithField("tool", tool).Warn("Tool already installed")
						taskCB(&rpc.TaskProgress{Name: "Tool " + tool.String() + " already installed", Completed: true})
					} else {
						toolsToInstall = append(toolsToInstall, tool)
					}
				}

				// Downloads platform tools
				for _, tool := range toolsToInstall {
					if err := DownloadToolRelease(pm, tool, downloadCB); err != nil {
						taskCB(&rpc.TaskProgress{Message: "Error downloading tool " + tool.String()})
						return err
					}
				}

				// Downloads platform
				if d, err := pm.DownloadPlatformRelease(latest, downloaderConfig); err != nil {
					return err
				} else if err := Download(d, latest.String(), downloadCB); err != nil {
					return err
				}

				logrus.Info("Updating platform " + installed.String())
				taskCB(&rpc.TaskProgress{Name: "Updating " + latest.String()})

				// Installs tools
				for _, tool := range toolsToInstall {
					if err := InstallToolRelease(pm, tool, taskCB); err != nil {
						taskCB(&rpc.TaskProgress{Message: "Error installing tool " + tool.String()})
						return err
					}
				}

				// Installs platform
				err = pm.InstallPlatform(latest)
				if err != nil {
					logrus.WithError(err).Error("Cannot install platform")
					taskCB(&rpc.TaskProgress{Message: "Error installing " + latest.String()})
					return err
				}

				// Uninstall previously installed release
				err = pm.UninstallPlatform(installedRelease)

				// In case uninstall fails tries to rollback
				if err != nil {
					logrus.WithError(err).Error("Error updating platform.")
					taskCB(&rpc.TaskProgress{Message: "Error upgrading platform: " + err.Error()})

					// Rollback
					if err := pm.UninstallPlatform(latest); err != nil {
						logrus.WithError(err).Error("Error rolling-back changes.")
						taskCB(&rpc.TaskProgress{Message: "Error rolling-back changes: " + err.Error()})
						return err
					}
				}

				// Uninstall unused tools
				for _, toolRelease := range installedTools {
					if !pm.IsToolRequired(toolRelease) {
						log := pm.Log.WithField("Tool", toolRelease)

						log.Info("Uninstalling tool")
						taskCB(&rpc.TaskProgress{Name: "Uninstalling " + toolRelease.String() + ", tool is no more required"})

						if err := pm.UninstallTool(toolRelease); err != nil {
							log.WithError(err).Error("Error uninstalling")
							return err
						}

						log.Info("Tool uninstalled")
						taskCB(&rpc.TaskProgress{Message: toolRelease.String() + " uninstalled", Completed: true})
					}
				}

				// Perform post install
				if !req.SkipPostInstall {
					logrus.Info("Running post_install script")
					taskCB(&rpc.TaskProgress{Message: "Configuring platform"})
					if err := pm.RunPostInstallScript(latest); err != nil {
						taskCB(&rpc.TaskProgress{Message: fmt.Sprintf("WARNING: cannot run post install: %s", err)})
					}
				} else {
					logrus.Info("Skipping platform configuration (post_install run).")
					taskCB(&rpc.TaskProgress{Message: "Skipping platform configuration"})
				}
			}
		}
	}

	return nil
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
	downloadsDir := paths.New(configuration.Settings.GetString("directories.Downloads"))
	if downloadsDir.NotExist() {
		err := downloadsDir.MkdirAll()
		if err != nil {
			return nil, err
		}
	}

	// setup data directory
	dataDir := paths.New(configuration.Settings.GetString("directories.Data"))
	packagesDir := configuration.PackagesDir(configuration.Settings)
	if packagesDir.NotExist() {
		err := packagesDir.MkdirAll()
		if err != nil {
			return nil, err
		}
	}

	if !getLibOnly {
		res.Pm = packagemanager.NewPackageManager(dataDir, configuration.PackagesDir(configuration.Settings),
			downloadsDir, dataDir.Join("tmp"))

		urls := []string{globals.DefaultIndexURL}
		urls = append(urls, configuration.Settings.GetStringSlice("board_manager.additional_urls")...)
		for _, u := range urls {
			URL, err := utils.URLParse(u)
			if err != nil {
				logrus.Warnf("Unable to parse index URL: %s, skip...", u)
				continue
			}

			if URL.Scheme == "file" {
				path := paths.New(URL.Path)
				if err != nil {
					return nil, fmt.Errorf("can't get absolute path of %v: %w", path, err)
				}

				_, err = res.Pm.LoadPackageIndexFromFile(path)
				if err != nil {
					res.PlatformIndexErrors = append(res.PlatformIndexErrors, err.Error())
				}
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
	if bundledLibsDir := configuration.IDEBundledLibrariesDir(configuration.Settings); bundledLibsDir != nil {
		res.Lm.AddLibrariesDir(bundledLibsDir, libraries.IDEBuiltIn)
	}

	// Add user libraries dir
	libDir := configuration.LibrariesDir(configuration.Settings)
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

// LoadSketch collects and returns all files composing a sketch
func LoadSketch(ctx context.Context, req *rpc.LoadSketchReq) (*rpc.LoadSketchResp, error) {
	sketch, err := builder.SketchLoad(req.SketchPath, "")
	if err != nil {
		return nil, fmt.Errorf("Error loading sketch %v: %v", req.SketchPath, err)
	}

	otherSketchFiles := make([]string, len(sketch.OtherSketchFiles))
	for i, file := range sketch.OtherSketchFiles {
		otherSketchFiles[i] = file.Path
	}

	additionalFiles := make([]string, len(sketch.AdditionalFiles))
	for i, file := range sketch.AdditionalFiles {
		additionalFiles[i] = file.Path
	}

	rootFolderFiles := make([]string, len(sketch.RootFolderFiles))
	for i, file := range sketch.RootFolderFiles {
		rootFolderFiles[i] = file.Path
	}

	return &rpc.LoadSketchResp{
		MainFile:         sketch.MainFile.Path,
		LocationPath:     sketch.LocationPath,
		OtherSketchFiles: otherSketchFiles,
		AdditionalFiles:  additionalFiles,
		RootFolderFiles:  rootFolderFiles,
	}, nil
}
