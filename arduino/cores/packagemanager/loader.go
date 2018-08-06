/*
 * This file is part of arduino-cli.
 *
 * arduino-cli is free software; you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation; either version 2 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program; if not, write to the Free Software
 * Foundation, Inc., 51 Franklin St, Fifth Floor, Boston, MA  02110-1301  USA
 *
 * As a special exception, you may use this file as part of a free software
 * library without restriction.  Specifically, if other files instantiate
 * templates or use macros or inline functions from this file, or you compile
 * this file and link it with other files to produce an executable, this
 * file does not by itself cause the resulting executable to be covered by
 * the GNU General Public License.  This exception does not however
 * invalidate any other reasons why the executable file might be covered by
 * the GNU General Public License.
 *
 * Copyright 2017-2018 ARDUINO AG (http://www.arduino.cc/)
 */

package packagemanager

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/arduino/go-paths-helper"
	properties "github.com/arduino/go-properties-map"
	"github.com/bcmi-labs/arduino-cli/arduino/cores"
	"github.com/bcmi-labs/arduino-cli/configs"
	"go.bug.st/relaxed-semver"
)

// LoadHardware read all plaforms from the configured paths
func (pm *PackageManager) LoadHardware(config *configs.Configuration) error {
	dirs, err := config.HardwareDirectories()
	if err != nil {
		return fmt.Errorf("getting hardware directory: %s", err)
	}
	if err := pm.LoadHardwareFromDirectories(dirs); err != nil {
		return err
	}
	dirs, err = configs.BundleToolsDirectories()
	if err != nil {
		return fmt.Errorf("getting hardware directory: %s", err)
	}
	return pm.LoadToolsFromBundleDirectories(dirs)
}

// LoadHardwareFromDirectories load plaforms from a set of directories
func (pm *PackageManager) LoadHardwareFromDirectories(hardwarePaths paths.PathList) error {
	for _, path := range hardwarePaths {
		if err := pm.LoadHardwareFromDirectory(path); err != nil {
			return fmt.Errorf("loading hardware from %s: %s", path, err)
		}
	}
	return nil
}

// LoadHardwareFromDirectory read a plaform from the path passed as parameter
func (pm *PackageManager) LoadHardwareFromDirectory(path *paths.Path) error {
	pm.Log.Infof("Loading hardware from: %s", path)
	if err := path.ToAbs(); err != nil {
		return fmt.Errorf("find abs path: %s", err)
	}

	// TODO: IS THIS CHECK NEEDED? can we ignore and let it fail at next ReadDir?
	if isDir, err := path.IsDir(); err != nil {
		return fmt.Errorf("reading %s stat info: %s", path, err)
	} else if !isDir {
		return fmt.Errorf("%s is not a directory", path)
	}

	// Scan subdirs
	files, err := path.ReadDir()
	if err != nil {
		return fmt.Errorf("reading %s directory: %s", path, err)
	}
	files.FilterOutHiddenFiles()
	for _, packagerPath := range files {
		packager := packagerPath.Base()

		// First exclude all "tools" directory
		if packager == "tools" {
			pm.Log.Infof("Excluding directory: %s", packagerPath)
			continue
		}

		// Follow symlinks
		err := packagerPath.FollowSymLink() // ex: .arduino15/packages/arduino/
		if err != nil {
			return fmt.Errorf("following possible symlink %s: %s", path, err)
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
		if isDir, _ := hardwareSubdirPath.IsDir(); isDir {
			// we found the "hardware" directory move down into that
			architectureParentPath = hardwareSubdirPath // ex: .arduino15/packages/arduino/
		} else if isDir, _ := packagerPath.IsDir(); isDir {
			// we are already at the correct level
			architectureParentPath = packagerPath
		} else {
			// error: do nothing
			continue
		}

		targetPackage := pm.packages.GetOrCreatePackage(packager)
		if err := pm.loadPlatforms(targetPackage, architectureParentPath); err != nil {
			return fmt.Errorf("loading package %s: %s", packager, err)
		}

		// Check if we have tools to load, the directory structure is as follows:
		// - PACKAGER/tools/TOOL-NAME/TOOL-VERSION/... (ex: arduino/tools/bossac/1.7.0/...)
		toolsSubdirPath := packagerPath.Join("tools")
		if isDir, _ := toolsSubdirPath.IsDir(); isDir {
			pm.Log.Infof("Checking existence of 'tools' path: %s", toolsSubdirPath)
			if err := pm.loadToolsFromPackage(targetPackage, toolsSubdirPath); err != nil {
				return fmt.Errorf("loading tools from %s: %s", toolsSubdirPath, err)
			}
		}
	}

	return nil
}

// loadPlatforms load plaftorms from the specified directory assuming that they belongs
// to the targetPackage object passed as parameter.
func (pm *PackageManager) loadPlatforms(targetPackage *cores.Package, packageDir *paths.Path) error {
	pm.Log.Infof("Loading package %s from: %s", targetPackage.Name, packageDir)

	files, err := packageDir.ReadDir()
	if err != nil {
		return fmt.Errorf("reading directory %s: %s", packageDir, err)
	}

	for _, file := range files {
		architecure := file.Base()
		if strings.HasPrefix(architecure, ".") || architecure == "tools" ||
			architecure == "platform.txt" { // TODO: Check if this "platform.txt" condition should be here....
			continue
		}
		platformPath := packageDir.Join(architecure)
		if isDir, _ := platformPath.IsDir(); !isDir {
			continue
		}

		// There are two possible platform directory structures:
		// - ARCHITECTURE/boards.txt
		// - ARCHITECTURE/VERSION/boards.txt
		// We identify them by checking where is the bords.txt file
		possibleBoardTxtPath := platformPath.Join("boards.txt")
		if exist, err := possibleBoardTxtPath.Exist(); err != nil {

			return fmt.Errorf("looking for boards.txt in %s: %s", possibleBoardTxtPath, err)

		} else if exist {

			// case: ARCHITECTURE/boards.txt
			// this is an unversioned Platform

			// FIXME: this check is duplicated, find a better way to handle this
			if exist, err := platformPath.Join("boards.txt").Exist(); err != nil {
				return fmt.Errorf("opening boards.txt: %s", err)
			} else if !exist {
				continue
			}

			platform := targetPackage.GetOrCreatePlatform(architecure)
			release, err := platform.GetOrCreateRelease(nil)
			if err != nil {
				return fmt.Errorf("loading platform release: %s", err)
			}
			if err := pm.loadPlatformRelease(release, platformPath); err != nil {
				return fmt.Errorf("loading platform release: %s", err)
			}
			pm.Log.WithField("platform", release).Infof("Loaded platform")

		} else /* !exist */ {

			// case: ARCHITECTURE/VERSION/boards.txt
			// let's dive into VERSION directories

			platform := targetPackage.GetOrCreatePlatform(architecure)
			versionDirs, err := platformPath.ReadDir()
			if err != nil {
				return fmt.Errorf("reading dir %s: %s", platformPath, err)
			}
			versionDirs.FilterDirs()
			versionDirs.FilterOutHiddenFiles()
			for _, versionDir := range versionDirs {
				if exist, err := versionDir.Join("boards.txt").Exist(); err != nil {
					return fmt.Errorf("opening boards.txt: %s", err)
				} else if !exist {
					continue
				}

				version, err := semver.Parse(versionDir.Base())
				if err != nil {
					return fmt.Errorf("invalid version dir %s: %s", versionDir, err)
				}
				release, err := platform.GetOrCreateRelease(version)
				if err != nil {
					return fmt.Errorf("loading platform release %s: %s", versionDir, err)
				}
				if err := pm.loadPlatformRelease(release, versionDir); err != nil {
					return fmt.Errorf("loading platform release %s: %s", versionDir, err)
				}
				pm.Log.WithField("platform", release).Infof("Loaded platform")
			}
		}
	}

	return nil
}

func (pm *PackageManager) loadPlatformRelease(platform *cores.PlatformRelease, path *paths.Path) error {
	platform.InstallDir = path

	// Some useful paths
	platformTxtPath := path.Join("platform.txt")
	platformTxtLocalPath := path.Join("platform.local.txt")
	programmersTxtPath := path.Join("programmers.txt")

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

	// Create programmers properties
	if programmersProperties, err := properties.SafeLoad(programmersTxtPath.String()); err == nil {
		platform.Programmers = properties.MergeMapsOfProperties(
			map[string]properties.Map{},
			platform.Programmers, // TODO: Very weird, why not an empty one?
			programmersProperties.FirstLevelOf())
	} else {
		return err
	}

	if err := pm.loadBoards(platform); err != nil {
		return err
	}

	return nil
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
		boardProperties["_id"] = boardID // TODO: What is that for??
		board := platform.GetOrCreateBoard(boardID)
		board.Properties.Merge(boardProperties)
	}

	return nil
}

func (pm *PackageManager) loadToolsFromPackage(targetPackage *cores.Package, toolsPath *paths.Path) error {
	pm.Log.Infof("Loading tools from dir: %s", toolsPath)

	toolsPaths, err := toolsPath.ReadDir()
	if err != nil {
		return fmt.Errorf("reading directory %s: %s", toolsPath, err)
	}
	toolsPaths.FilterDirs()
	toolsPaths.FilterOutHiddenFiles()
	for _, toolPath := range toolsPaths {
		name := toolPath.Base()
		tool := targetPackage.GetOrCreateTool(name)
		if err = pm.loadToolReleasesFromTool(tool, toolPath); err != nil {
			return fmt.Errorf("loading tool release in %s: %s", toolPath, err)
		}
	}
	return nil
}

func (pm *PackageManager) loadToolReleasesFromTool(tool *cores.Tool, toolPath *paths.Path) error {
	toolVersions, err := toolPath.ReadDir()
	if err != nil {
		return err
	}
	toolVersions.FilterDirs()
	toolVersions.FilterOutHiddenFiles()
	for _, versionPath := range toolVersions {
		version, err := semver.Parse(versionPath.Base())
		if err != nil {
			return fmt.Errorf("invalid tool version path %s: %s", versionPath, err)
		}
		if toolReleasePath, err := versionPath.Abs(); err == nil {
			release := tool.GetOrCreateRelease(version)
			release.InstallDir = toolReleasePath
			pm.Log.WithField("tool", release).Infof("Loaded tool")
		} else {
			return err
		}
	}

	return nil
}

func (pm *PackageManager) LoadToolsFromBundleDirectories(dirs paths.PathList) error {
	for _, dir := range dirs {
		if err := pm.LoadToolsFromBundleDirectory(dir); err != nil {
			return fmt.Errorf("loading bundled tools from %s: %s", dir, err)
		}
	}
	return nil
}

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
			targetPackage := pm.packages.GetOrCreatePackage(packager)

			for toolName, toolVersion := range toolsData {
				tool := targetPackage.GetOrCreateTool(toolName)
				version, err := semver.Parse(toolVersion)
				if err != nil {
					return fmt.Errorf("invalid tool version in %s: %s", builtinToolsVersionsTxtPath, err)
				}
				release := tool.GetOrCreateRelease(version)
				release.InstallDir = toolPath
				pm.Log.WithField("tool", release).Infof("Loaded tool")
			}
		}
	} else {
		// otherwise load the tools inside the unnamed package
		unnamedPackage := pm.packages.GetOrCreatePackage("")
		pm.loadToolsFromPackage(unnamedPackage, toolsPath)
	}
	return nil
}
