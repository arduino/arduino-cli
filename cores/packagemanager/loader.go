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
	"io/ioutil"
	"os"
	"path/filepath"

	properties "github.com/arduino/go-properties-map"
	"github.com/bcmi-labs/arduino-cli/configs"
	"github.com/bcmi-labs/arduino-cli/cores"
	"github.com/sirupsen/logrus"
)

// LoadHardware read all plaforms from the configured paths
func (pm *PackageManager) LoadHardware() error {
	dirs, err := configs.HardwareDirectories()
	if err != nil {
		return fmt.Errorf("getting hardware folder: %s", err)
	}
	if err := pm.LoadHardwareFromDirectories(dirs); err != nil {
		return err
	}
	dirs, err = configs.BundleToolsDirectories()
	if err != nil {
		return fmt.Errorf("getting hardware folder: %s", err)
	}
	return pm.LoadToolsFromBundleDirectories(dirs)
}

// LoadHardwareFromDirectories load plaforms from a set of directories
func (pm *PackageManager) LoadHardwareFromDirectories(hardwarePaths []string) error {
	for _, path := range hardwarePaths {
		if err := pm.LoadHardwareFromDirectory(path); err != nil {
			return fmt.Errorf("loading hardware from %s: %s", path, err)
		}
	}
	return nil
}

// LoadHardwareFromDirectory read a plaform from the path passed as parameter
func (pm *PackageManager) LoadHardwareFromDirectory(path string) error {
	logrus.Infof("Loading hardware from: %s", path)
	path, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("find abs path: %s", err)
	}

	// TODO: IS THIS CHECK NEEDED? can we ignore and let it fail at next ReadDir?
	if stat, err := os.Stat(path); err != nil {
		return fmt.Errorf("reading %s stat info: %s", path, err)
	} else if !stat.IsDir() {
		return fmt.Errorf("%s is not a folder", path)
	}

	// TODO: IS THIS REALLY NEEDED? this is used only to get ctags properties AFAIK
	platformTxtPath := filepath.Join(path, "platform.txt")
	if props, err := properties.SafeLoad(platformTxtPath); err == nil {
		pm.packages.Properties.Merge(props)
	} else {
		return fmt.Errorf("reading %s: %s", platformTxtPath, err)
	}

	// Scan subfolders.
	files, err := ioutil.ReadDir(path)
	if err != nil {
		return fmt.Errorf("reading %s directory: %s", path, err)
	}
	for _, packagerPathInfo := range files {
		packager := packagerPathInfo.Name()

		// First exclude all "tools" folders
		if packager == "tools" {
			logrus.Debugf("Excluding folder: %s", filepath.Join(path, packager))
			continue
		}

		// Follow symlinks
		packagerPath := filepath.Join(path, packager)
		packagerPath, err := filepath.EvalSymlinks(packagerPath) // ex: .arduino15/packages/arduino/
		if err != nil {
			return fmt.Errorf("following possible symlink %s: %s", path, err)
		}

		// There are two possible package folder structures:
		// - PACKAGER/ARCHITECTURE-1/boards.txt...                   (ex: arduino/avr/...)
		//   PACKAGER/ARCHITECTURE-2/boards.txt...                   (ex: arduino/sam/...)
		//   PACKAGER/ARCHITECTURE-3/boards.txt...                   (ex: arduino/samd/...)
		// or
		// - PACKAGER/hardware/ARCHITECTURE-1/VERSION/boards.txt...  (ex: arduino/hardware/avr/1.6.15/...)
		//   PACKAGER/hardware/ARCHITECTURE-2/VERSION/boards.txt...  (ex: arduino/hardware/sam/1.6.6/...)
		//   PACKAGER/hardware/ARCHITECTURE-3/VERSION/boards.txt...  (ex: arduino/hardware/samd/1.6.12/...)
		//   PACKAGER/tools/...                                      (ex: arduino/tools/...)
		// in the latter case we just move into "hardware" folder and continue
		architectureParentPath := ""
		hardwareSubdirPath := filepath.Join(packagerPath, "hardware") // ex: .arduino15/packages/arduino/hardware
		if info, err := os.Stat(hardwareSubdirPath); err == nil && info.IsDir() {
			// we found the "hardware" folder move down into that
			architectureParentPath = hardwareSubdirPath // ex: .arduino15/packages/arduino/
		} else if info, err := os.Stat(packagerPath); err == nil && info.IsDir() {
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

		// Check if we have tools to load, the folder structure is as follows:
		// - PACKAGER/tools/TOOL-NAME/TOOL-VERSION/... (ex: arduino/tools/bossac/1.7.0/...)
		toolsSubdirPath := filepath.Join(packagerPath, "tools")
		if info, err := os.Stat(toolsSubdirPath); err == nil && info.IsDir() {
			logrus.Debugf("Checking existence of 'tools' path: %s", toolsSubdirPath)
			if err := pm.loadToolsFromPackage(targetPackage, toolsSubdirPath); err != nil {
				return fmt.Errorf("loading tools from %s: %s", toolsSubdirPath, err)
			}
		}
	}

	return nil
}

// loadPlatforms load plaftorms from the specified directory assuming that they belongs
// to the targetPackage object passed as parameter.
func (pm *PackageManager) loadPlatforms(targetPackage *cores.Package, packageFolder string) error {
	logrus.Infof("Loading package %s from: %s", targetPackage.Name, packageFolder)

	// packagePlatformTxt, err := properties.SafeLoad(filepath.Join(folder, constants.FILE_PLATFORM_TXT))
	// if err != nil {
	// 	return err
	// }
	// targetPackage.Properties.Merge(packagePlatformTxt)

	files, err := ioutil.ReadDir(packageFolder)
	if err != nil {
		return fmt.Errorf("reading directory %s: %s", packageFolder, err)
	}

	for _, file := range files {
		architecure := file.Name()
		platformPath := filepath.Join(packageFolder, architecure)
		if architecure == "tools" ||
			architecure == "platform.txt" { // TODO: Check if this "platform.txt" condition should be here....
			continue
		}

		// There are two possible platform folder structures:
		// - ARCHITECTURE/boards.txt
		// - ARCHITECTURE/VERSION/boards.txt
		// We identify them by checking where is the bords.txt file
		possibleBoardTxtPath := filepath.Join(platformPath, "boards.txt")
		if _, err := os.Stat(possibleBoardTxtPath); err == nil {
			// case: ARCHITECTURE/boards.txt
			// this is an unversioned Platform

			platform := targetPackage.GetOrCreatePlatform(architecure)
			release := platform.GetOrCreateRelease("")
			if err := pm.loadPlatformRelease(release, platformPath); err != nil {
				return fmt.Errorf("loading platform release: %s", err)
			}
			logrus.WithField("platform", release).Debugf("Loaded platform")

		} else if os.IsNotExist(err) {
			// case: ARCHITECTURE/VERSION/boards.txt
			// let's dive into VERSION folders

			platform := targetPackage.GetOrCreatePlatform(architecure)
			versionDirs, err := ioutil.ReadDir(platformPath)
			if err != nil {
				return fmt.Errorf("reading dir %s: %s", platformPath, err)
			}
			for _, versionDir := range versionDirs {
				if !versionDir.IsDir() {
					continue
				}
				version := versionDir.Name()
				release := platform.GetOrCreateRelease(version)
				platformWithVersionPath := filepath.Join(platformPath, version)

				if err := pm.loadPlatformRelease(release, platformWithVersionPath); err != nil {
					return fmt.Errorf("loading platform release %s: %s", version, err)
				}
				logrus.WithField("platform", release).Debugf("Loaded platform")
			}
		} else {
			return fmt.Errorf("looking for boards.txt in %s: %s", possibleBoardTxtPath, err)
		}
	}

	return nil
}

func (pm *PackageManager) loadPlatformRelease(platform *cores.PlatformRelease, folder string) error {
	if _, err := os.Stat(filepath.Join(folder, "boards.txt")); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("opening boards.txt: %s", err)
	} else if os.IsNotExist(err) {
		return fmt.Errorf("invalid platform directory %s: boards.txt not found", folder)
	}
	platform.Folder = folder

	// Some useful paths
	platformTxtPath := filepath.Join(folder, "platform.txt")
	platformTxtLocalPath := filepath.Join(folder, "platform.local.txt")
	programmersTxtPath := filepath.Join(folder, "programmers.txt")

	// Create platform properties
	platform.Properties = platform.Properties.Clone() // TODO: why CLONE?
	if p, err := properties.SafeLoad(platformTxtPath); err == nil {
		platform.Properties.Merge(p)
	} else {
		return fmt.Errorf("loading %s: %s", platformTxtPath, err)
	}
	if p, err := properties.SafeLoad(platformTxtLocalPath); err == nil {
		platform.Properties.Merge(p)
	} else {
		return fmt.Errorf("loading %s: %s", platformTxtLocalPath, err)
	}

	// Create programmers properties
	if programmersProperties, err := properties.SafeLoad(programmersTxtPath); err == nil {
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
	if platform.Folder == "" {
		return fmt.Errorf("platform not installed")
	}
	boardsTxtPath := filepath.Join(platform.Folder, "boards.txt")
	boardsLocalTxtPath := filepath.Join(platform.Folder, "boards.local.txt")

	boardsProperties, err := properties.Load(boardsTxtPath)
	if err != nil {
		return err
	}
	if localProperties, err := properties.SafeLoad(boardsLocalTxtPath); err == nil {
		boardsProperties.Merge(localProperties)
	} else {
		return err
	}

	propertiesByBoard := boardsProperties.FirstLevelOf()
	delete(propertiesByBoard, "menu") // TODO: check this one

	for boardID, boardProperties := range propertiesByBoard {
		boardProperties["_id"] = boardID // TODO: What is that for??
		board := platform.GetOrCreateBoard(boardID)
		board.Properties.Merge(boardProperties)
	}

	return nil
}

func (pm *PackageManager) loadToolsFromPackage(targetPackage *cores.Package, toolsPath string) error {
	logrus.Infof("Loading tools from dir: %s", toolsPath)

	toolsInfo, err := ioutil.ReadDir(toolsPath)
	if err != nil {
		return fmt.Errorf("reading directory %s: %s", toolsPath, err)
	}
	for _, toolInfo := range toolsInfo {
		if !toolInfo.IsDir() {
			continue
		}

		name := toolInfo.Name()
		tool := targetPackage.GetOrCreateTool(name)
		toolPath := filepath.Join(toolsPath, name)
		if err = loadToolReleasesFromTool(tool, toolPath); err != nil {
			return fmt.Errorf("loading tool release in %s: %s", toolPath, err)
		}
	}
	return nil
}

func loadToolReleasesFromTool(tool *cores.Tool, toolPath string) error {
	toolVersions, err := ioutil.ReadDir(toolPath)
	if err != nil {
		return err
	}
	for _, versionInfo := range toolVersions {
		version := versionInfo.Name()
		if toolReleasePath, err := filepath.Abs(filepath.Join(toolPath, version)); err == nil {
			release := tool.GetOrCreateRelease(version)
			release.Folder = toolReleasePath
			logrus.WithField("tool", release).Debugf("Loaded tool")
		} else {
			return err
		}
	}

	return nil
}

func (pm *PackageManager) LoadToolsFromBundleDirectories(dirs []string) error {
	for _, dir := range dirs {
		if err := pm.LoadToolsFromBundleDirectory(dir); err != nil {
			return fmt.Errorf("loading bundled tools from %s: %s", dir, err)
		}
	}
	return nil
}

func (pm *PackageManager) LoadToolsFromBundleDirectory(toolsPath string) error {
	logrus.Infof("Loading tools from bundle dir: %s", toolsPath)

	// We scan toolsPath content to find a "builtin_tools_versions.txt", if such file exists
	// then the all the tools are available in the same directory, mixed together, and their
	// name and version are written in the "builtin_tools_versions.txt" file.
	// If no "builtin_tools_versions.txt" is found, then the folder structure is the classic
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
	if err := filepath.Walk(toolsPath, findBuiltInToolsVersionsTxt); err != nil {
		return fmt.Errorf("searching for builtin_tools_versions.txt in %s: %s", toolsPath, err)
	}

	if builtinToolsVersionsTxtPath != "" {
		// If builtin_tools_versions.txt is found create tools based on the info
		// contained in that file
		logrus.Debugf("Found builtin_tools_versions.txt")
		toolPath, err := filepath.Abs(filepath.Dir(builtinToolsVersionsTxtPath))
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
				release := tool.GetOrCreateRelease(toolVersion)
				release.Folder = toolPath
				logrus.WithField("tool", release).Debugf("Loaded tool")
			}
		}
	} else {
		// otherwise load the tools inside the unnamed package
		unnamedPackage := pm.packages.GetOrCreatePackage("")
		pm.loadToolsFromPackage(unnamedPackage, toolsPath)
	}
	return nil
}
