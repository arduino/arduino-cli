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

package packagemanager

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/arduino/arduino-cli/arduino/cores"
	"github.com/arduino/arduino-cli/configuration"
	"github.com/arduino/go-paths-helper"
	properties "github.com/arduino/go-properties-orderedmap"
	semver "go.bug.st/relaxed-semver"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// LoadHardware read all plaforms from the configured paths
func (pm *PackageManager) LoadHardware() []*status.Status {
	statuses := []*status.Status{}
	dirs := configuration.HardwareDirectories(configuration.Settings)
	if errs := pm.LoadHardwareFromDirectories(dirs); len(errs) > 0 {
		statuses = append(statuses, errs...)
	}

	dirs = configuration.BundleToolsDirectories(configuration.Settings)
	if errs := pm.LoadToolsFromBundleDirectories(dirs); len(errs) > 0 {
		statuses = append(statuses, errs...)
	}
	return statuses
}

// LoadHardwareFromDirectories load plaforms from a set of directories
func (pm *PackageManager) LoadHardwareFromDirectories(hardwarePaths paths.PathList) []*status.Status {
	statuses := []*status.Status{}
	for _, path := range hardwarePaths {
		if errs := pm.LoadHardwareFromDirectory(path); len(errs) > 0 {
			statuses = append(statuses, errs...)
		}
	}
	return statuses
}

// LoadHardwareFromDirectory read a plaform from the path passed as parameter
func (pm *PackageManager) LoadHardwareFromDirectory(path *paths.Path) []*status.Status {
	pm.Log.Infof("Loading hardware from: %s", path)
	statuses := []*status.Status{}
	if err := path.ToAbs(); err != nil {
		s := status.Newf(codes.FailedPrecondition, "find abs path: %s", err)
		return append(statuses, s)
	}

	if path.IsNotDir() {
		s := status.Newf(codes.FailedPrecondition, "%s is not a directory", path)
		return append(statuses, s)
	}

	// Scan subdirs
	packagersPaths, err := path.ReadDir()
	if err != nil {
		s := status.Newf(codes.FailedPrecondition, "reading %s directory: %s", path, err)
		return append(statuses, s)
	}
	packagersPaths.FilterOutHiddenFiles()
	packagersPaths.FilterDirs()

	// Load custom platform properties if available
	// "Global" platform.txt used to overwrite all installed platforms.
	// For more info: https://arduino.github.io/arduino-cli/latest/platform-specification/#global-platformtxt
	if globalPlatformTxt := path.Join("platform.txt"); globalPlatformTxt.Exist() {
		pm.Log.Infof("Loading custom platform properties: %s", globalPlatformTxt)
		if p, err := properties.LoadFromPath(globalPlatformTxt); err != nil {
			pm.Log.WithError(err).Errorf("Error loading properties.")
		} else {
			pm.CustomGlobalProperties.Merge(p)
		}
	}

	for _, packagerPath := range packagersPaths {
		packager := packagerPath.Base()

		// Skip tools, they're not packages and don't contain Platforms
		if packager == "tools" {
			pm.Log.Infof("Excluding directory: %s", packagerPath)
			continue
		}

		// Follow symlinks
		err := packagerPath.FollowSymLink() // ex: .arduino15/packages/arduino/
		if err != nil {
			s := status.Newf(codes.Internal, "following possible symlink %s: %s", path, err)
			statuses = append(statuses, s)
			continue
		}

		// There are two possible package directory structures:
		// - PACKAGER/ARCHITECTURE-1/boards.txt...                   (ex: arduino/avr/...)
		//   PACKAGER/ARCHITECTURE-2/boards.txt...                   (ex: arduino/sam/...)
		//   PACKAGER/ARCHITECTURE-3/boards.txt...                   (ex: arduino/samd/...)
		// or
		// - PACKAGER/hardware/ARCHITECTURE-1/VERSION/boards.txt...  (ex: arduino/hardware/avr/1.6.15/...)
		//   PACKAGER/hardware/ARCHITECTURE-2/VERSION/boards.txt...  (ex: arduino/hardware/sam/1.6.6/...)
		//   PACKAGER/hardware/ARCHITECTURE-3/VERSION/boards.txt...  (ex: arduino/hardware/samd/1.6.12/...)
		//   PACKAGER/tools/...                                      (ex: arduino/tools/...)
		// in the latter case we just move into "hardware" directory and continue
		var architectureParentPath *paths.Path
		hardwareSubdirPath := packagerPath.Join("hardware") // ex: .arduino15/packages/arduino/hardware
		if hardwareSubdirPath.IsDir() {
			// we found the "hardware" directory move down into that
			architectureParentPath = hardwareSubdirPath // ex: .arduino15/packages/arduino/
		} else {
			// we are already at the correct level
			architectureParentPath = packagerPath
		}

		targetPackage := pm.Packages.GetOrCreatePackage(packager)
		if errs := pm.loadPlatforms(targetPackage, architectureParentPath); len(errs) > 0 {
			statuses = append(statuses, errs...)
		}

		// Check if we have tools to load, the directory structure is as follows:
		// - PACKAGER/tools/TOOL-NAME/TOOL-VERSION/... (ex: arduino/tools/bossac/1.7.0/...)
		toolsSubdirPath := packagerPath.Join("tools")
		if toolsSubdirPath.IsDir() {
			pm.Log.Infof("Checking existence of 'tools' path: %s", toolsSubdirPath)
			if errs := pm.loadToolsFromPackage(targetPackage, toolsSubdirPath); len(errs) > 0 {
				statuses = append(statuses, errs...)
			}
		}
	}

	return statuses
}

// loadPlatforms load plaftorms from the specified directory assuming that they belongs
// to the targetPackage object passed as parameter.
// A list of gRPC Status error is returned for each Platform failed to load.
func (pm *PackageManager) loadPlatforms(targetPackage *cores.Package, packageDir *paths.Path) []*status.Status {
	pm.Log.Infof("Loading package %s from: %s", targetPackage.Name, packageDir)

	statuses := []*status.Status{}

	platformsDirs, err := packageDir.ReadDir()
	if err != nil {
		s := status.Newf(codes.FailedPrecondition, "reading directory %s: %s", packageDir, err)
		return append(statuses, s)
	}

	// A platform can only be inside a directory, thus we skip everything else.
	platformsDirs.FilterDirs()
	// Filter out directories like .git and similar things
	platformsDirs.FilterOutPrefix(".")
	for _, platformPath := range platformsDirs {
		// Tools are not a platform
		if platformPath.Base() == "tools" {
			continue
		}
		if err := pm.loadPlatform(targetPackage, platformPath); err != nil {
			statuses = append(statuses, err)
		}
	}

	return statuses
}

// loadPlatform loads a single platform and all its installed releases given a platformPath.
// platformPath must be a directory.
// Returns a gRPC Status error in case of failures.
func (pm *PackageManager) loadPlatform(targetPackage *cores.Package, platformPath *paths.Path) *status.Status {
	// This is not a platform
	if platformPath.IsNotDir() {
		return status.Newf(codes.NotFound, "path is not a platform directory: %s", platformPath)
	}

	architecture := platformPath.Base()

	// There are two possible platform directory structures:
	// - ARCHITECTURE/boards.txt
	// - ARCHITECTURE/VERSION/boards.txt
	// We identify them by checking where is the bords.txt file
	possibleBoardTxtPath := platformPath.Join("boards.txt")
	if exist, err := possibleBoardTxtPath.ExistCheck(); err != nil {
		return status.Newf(codes.FailedPrecondition, "looking for boards.txt in %s: %s", possibleBoardTxtPath, err)
	} else if exist {
		// case: ARCHITECTURE/boards.txt

		platformTxtPath := platformPath.Join("platform.txt")
		platformProperties, err := properties.SafeLoad(platformTxtPath.String())
		if err != nil {
			return status.Newf(codes.FailedPrecondition, "loading platform.txt: %v", err)
		}

		version := semver.MustParse(platformProperties.Get("version"))

		// Check if package_bundled_index.json exists.
		// This is used indirectly by the Java IDE since it's necessary for the arduino-builder
		// to find cores bundled with that version of the IDE.
		// TODO: This piece of logic MUST be removed as soon as the Java IDE stops using the arduino-builder.
		isIDEBundled := false
		packageBundledIndexPath := platformPath.Parent().Parent().Join("package_index_bundled.json")
		if packageBundledIndexPath.Exist() {
			// particular case: ARCHITECTURE/boards.txt with package_bundled_index.json

			// this is an unversioned Platform with a package_index_bundled.json that
			// gives information about the version and tools needed

			// Parse the bundled index and merge to the general index
			index, err := pm.LoadPackageIndexFromFile(packageBundledIndexPath)
			if err != nil {
				return status.Newf(codes.FailedPrecondition, "parsing IDE bundled index: %s", err)
			}

			// Now export the bundled index in a temporary core.Packages to retrieve the bundled package version
			tmp := cores.NewPackages()
			index.MergeIntoPackages(tmp)
			if tmpPackage := tmp.GetOrCreatePackage(targetPackage.Name); tmpPackage == nil {
				pm.Log.Warnf("Can't determine bundle platform version for %s", targetPackage.Name)
			} else if tmpPlatform := tmpPackage.GetOrCreatePlatform(architecture); tmpPlatform == nil {
				pm.Log.Warnf("Can't determine bundle platform version for %s:%s", targetPackage.Name, architecture)
			} else if tmpPlatformRelease := tmpPlatform.GetLatestRelease(); tmpPlatformRelease == nil {
				pm.Log.Warnf("Can't determine bundle platform version for %s:%s, no valid release found", targetPackage.Name, architecture)
			} else {
				version = tmpPlatformRelease.Version
			}

			isIDEBundled = true
		}

		platform := targetPackage.GetOrCreatePlatform(architecture)
		if !isIDEBundled {
			platform.ManuallyInstalled = true
		}
		release := platform.GetOrCreateRelease(version)
		release.IsIDEBundled = isIDEBundled
		if isIDEBundled {
			pm.Log.Infof("Package is built-in")
		}
		if err := pm.loadPlatformRelease(release, platformPath); err != nil {
			return status.Newf(codes.FailedPrecondition, "loading platform release: %s", err)

		}
		pm.Log.WithField("platform", release).Infof("Loaded platform")

	} else {
		// case: ARCHITECTURE/VERSION/boards.txt
		// let's dive into VERSION directories

		platform := targetPackage.GetOrCreatePlatform(architecture)
		versionDirs, err := platformPath.ReadDir()
		if err != nil {
			return status.Newf(codes.FailedPrecondition, "reading dir %s: %s", platformPath, err)
		}
		versionDirs.FilterDirs()
		versionDirs.FilterOutHiddenFiles()
		for _, versionDir := range versionDirs {
			if exist, err := versionDir.Join("boards.txt").ExistCheck(); err != nil {
				return status.Newf(codes.FailedPrecondition, "opening boards.txt: %s", err)
			} else if !exist {
				continue
			}

			version, err := semver.Parse(versionDir.Base())
			if err != nil {
				return status.Newf(codes.FailedPrecondition, "invalid version dir %s: %s", versionDir, err)
			}
			release := platform.GetOrCreateRelease(version)
			if err := pm.loadPlatformRelease(release, versionDir); err != nil {
				return status.Newf(codes.FailedPrecondition, "loading platform release: %s", err)
			}
			pm.Log.WithField("platform", release).Infof("Loaded platform")
		}
	}

	return nil
}

func (pm *PackageManager) loadPlatformRelease(platform *cores.PlatformRelease, path *paths.Path) error {
	platform.InstallDir = path

	// Some useful paths
	installedJSONPath := path.Join("installed.json")
	platformTxtPath := path.Join("platform.txt")
	platformTxtLocalPath := path.Join("platform.local.txt")
	programmersTxtPath := path.Join("programmers.txt")

	// If the installed.json file is found load it, this is done to handle the
	// case in which the platform's index and its url have been deleted locally,
	// if we don't load it some information about the platform is lost
	if installedJSONPath.Exist() {
		if _, err := pm.LoadPackageIndexFromFile(installedJSONPath); err != nil {
			return fmt.Errorf("loading %s: %s", installedJSONPath, err)
		}
	}

	// Create platform properties
	platform.Properties = platform.Properties.Clone() // TODO: why CLONE?
	if p, err := properties.SafeLoad(platformTxtPath.String()); err == nil {
		platform.Properties.Merge(p)
	} else {
		return fmt.Errorf("loading %s: %s", platformTxtPath, err)
	}
	if p, err := properties.SafeLoad(platformTxtLocalPath.String()); err == nil {
		platform.Properties.Merge(p)
	} else {
		return fmt.Errorf("loading %s: %s", platformTxtLocalPath, err)
	}

	if platform.Platform.Name == "" {
		platform.Platform.Name = platform.Properties.Get("name")
	}

	// Create programmers properties
	if programmersProperties, err := properties.SafeLoad(programmersTxtPath.String()); err == nil {
		for programmerID, programmerProperties := range programmersProperties.FirstLevelOf() {
			platform.Programmers[programmerID] = pm.loadProgrammer(programmerProperties)
			platform.Programmers[programmerID].PlatformRelease = platform
		}
	} else {
		return err
	}

	if err := pm.loadBoards(platform); err != nil {
		return err
	}

	return nil
}

func (pm *PackageManager) loadProgrammer(programmerProperties *properties.Map) *cores.Programmer {
	return &cores.Programmer{
		Name:       programmerProperties.Get("name"),
		Properties: programmerProperties,
	}
}

func (pm *PackageManager) loadBoards(platform *cores.PlatformRelease) error {
	if platform.InstallDir == nil {
		return fmt.Errorf("platform not installed")
	}

	boardsTxtPath := platform.InstallDir.Join("boards.txt")
	boardsProperties, err := properties.LoadFromPath(boardsTxtPath)
	if err != nil {
		return err
	}

	boardsLocalTxtPath := platform.InstallDir.Join("boards.local.txt")
	if localProperties, err := properties.SafeLoadFromPath(boardsLocalTxtPath); err == nil {
		boardsProperties.Merge(localProperties)
	} else {
		return err
	}

	propertiesByBoard := boardsProperties.FirstLevelOf()

	platform.Menus = propertiesByBoard["menu"]

	delete(propertiesByBoard, "menu") // TODO: check this one

	for boardID, boardProperties := range propertiesByBoard {
		boardProperties.Set("_id", boardID) // TODO: What is that for??
		board := platform.GetOrCreateBoard(boardID)
		board.Properties.Merge(boardProperties)
	}

	return nil
}

func (pm *PackageManager) loadToolsFromPackage(targetPackage *cores.Package, toolsPath *paths.Path) []*status.Status {
	pm.Log.Infof("Loading tools from dir: %s", toolsPath)

	statuses := []*status.Status{}
	toolsPaths, err := toolsPath.ReadDir()
	if err != nil {
		s := status.Newf(codes.FailedPrecondition, "reading directory %s: %s", toolsPath, err)
		return append(statuses, s)
	}
	toolsPaths.FilterDirs()
	toolsPaths.FilterOutHiddenFiles()
	for _, toolPath := range toolsPaths {
		name := toolPath.Base()
		tool := targetPackage.GetOrCreateTool(name)
		if err = pm.loadToolReleasesFromTool(tool, toolPath); err != nil {
			s := status.Newf(codes.FailedPrecondition, "loading tool release in %s: %s", toolPath, err)
			statuses = append(statuses, s)
		}
	}
	return statuses
}

func (pm *PackageManager) loadToolReleasesFromTool(tool *cores.Tool, toolPath *paths.Path) error {
	toolVersions, err := toolPath.ReadDir()
	if err != nil {
		return err
	}
	toolVersions.FilterDirs()
	toolVersions.FilterOutHiddenFiles()
	for _, versionPath := range toolVersions {
		if toolReleasePath, err := versionPath.Abs(); err == nil {
			version := semver.ParseRelaxed(versionPath.Base())
			release := tool.GetOrCreateRelease(version)
			release.InstallDir = toolReleasePath
			pm.Log.WithField("tool", release).Infof("Loaded tool")
		} else {
			return err
		}
	}

	return nil
}

// LoadToolsFromBundleDirectories FIXMEDOC
func (pm *PackageManager) LoadToolsFromBundleDirectories(dirs paths.PathList) []*status.Status {
	statuses := []*status.Status{}
	for _, dir := range dirs {
		if err := pm.LoadToolsFromBundleDirectory(dir); err != nil {
			statuses = append(statuses, status.Newf(codes.FailedPrecondition, "loading bundled tools from %s: %s", dir, err))
		}
	}
	return statuses
}

// LoadToolsFromBundleDirectory FIXMEDOC
func (pm *PackageManager) LoadToolsFromBundleDirectory(toolsPath *paths.Path) error {
	pm.Log.Infof("Loading tools from bundle dir: %s", toolsPath)

	// We scan toolsPath content to find a "builtin_tools_versions.txt", if such file exists
	// then the all the tools are available in the same directory, mixed together, and their
	// name and version are written in the "builtin_tools_versions.txt" file.
	// If no "builtin_tools_versions.txt" is found, then the directory structure is the classic
	// TOOLSPATH/TOOL-NAME/TOOL-VERSION and it will be parsed as such and associated to an
	// "unnamed" packager.

	// TODO: get rid of "builtin_tools_versions.txt"

	// Search for builtin_tools_versions.txt
	builtinToolsVersionsTxtPath := ""
	findBuiltInToolsVersionsTxt := func(currentPath string, info os.FileInfo, err error) error {
		if err != nil {
			// Ignore errors
			return nil
		}
		if builtinToolsVersionsTxtPath != "" {
			return filepath.SkipDir
		}
		if info.Name() == "builtin_tools_versions.txt" {
			builtinToolsVersionsTxtPath = currentPath
			return filepath.SkipDir
		}
		return nil
	}
	if err := filepath.Walk(toolsPath.String(), findBuiltInToolsVersionsTxt); err != nil {
		return fmt.Errorf("searching for builtin_tools_versions.txt in %s: %s", toolsPath, err)
	}

	if builtinToolsVersionsTxtPath != "" {
		// If builtin_tools_versions.txt is found create tools based on the info
		// contained in that file
		pm.Log.Infof("Found builtin_tools_versions.txt")
		toolPath, err := paths.New(builtinToolsVersionsTxtPath).Parent().Abs()
		if err != nil {
			return fmt.Errorf("getting parent dir of %s: %s", builtinToolsVersionsTxtPath, err)
		}

		all, err := properties.Load(builtinToolsVersionsTxtPath)
		if err != nil {
			return fmt.Errorf("reading %s: %s", builtinToolsVersionsTxtPath, err)
		}

		for packager, toolsData := range all.FirstLevelOf() {
			targetPackage := pm.Packages.GetOrCreatePackage(packager)

			for toolName, toolVersion := range toolsData.AsMap() {
				tool := targetPackage.GetOrCreateTool(toolName)
				version := semver.ParseRelaxed(toolVersion)
				release := tool.GetOrCreateRelease(version)
				release.InstallDir = toolPath
				pm.Log.WithField("tool", release).Infof("Loaded tool")
			}
		}
	} else {
		// otherwise load the tools inside the unnamed package
		unnamedPackage := pm.Packages.GetOrCreatePackage("")
		pm.loadToolsFromPackage(unnamedPackage, toolsPath)
	}
	return nil
}
