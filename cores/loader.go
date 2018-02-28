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

package cores

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	properties "github.com/arduino/go-properties-map"
)

// LoadHardwareFromFolders read all plaforms from a folder
func (packages *Packages) LoadHardwareFromFolders(folders []string) error {
	for _, folder := range folders {
		folder, err := filepath.Abs(folder)
		if err != nil {
			return fmt.Errorf("find abs path: %s", err)
		}
		// TODO: IS THIS CHECK NEEDED? can we ignore and let it fail at next ReadDir?
		if stat, err := os.Stat(folder); err != nil {
			return fmt.Errorf("reading %s stat info: %s", folder, err)
		} else if !stat.IsDir() {
			return fmt.Errorf("%s is not a folder", folder)
		}

		// TODO: IS THIS REALLY NEEDED? this is used only to get ctags properties AFAIK
		platformTxtPath := filepath.Join(folder, "platform.txt")
		if props, err := properties.SafeLoad(platformTxtPath); err == nil {
			packages.Properties.Merge(props)
		} else {
			return fmt.Errorf("reading %s: %s", platformTxtPath, err)
		}

		// Scan subfolders.
		files, err := ioutil.ReadDir(folder)
		if err != nil {
			return fmt.Errorf("reading %s directory: %s", folder, err)
		}
		for _, subfolder := range files {
			// First exclude all "tools" folders
			packager := subfolder.Name()
			if packager == "tools" {
				continue
			}

			if err := packages.LoadPackage(packager, filepath.Join(folder, packager)); err != nil {
				return err
			}
		}
	}

	return nil
}

// LoadPackage load the package identified by packager in the specified path
func (packages *Packages) LoadPackage(packager string, path string) error {
	// Follow symlinks
	path, err := filepath.EvalSymlinks(path)
	if err != nil {
		return fmt.Errorf("following possible symlink %s: %s", path, err)
	}

	// There are two possible package folder structures:
	// - PACKAGER/ARCHITECTURE/...
	// - PACKAGER/hardware/ARCHITECTURE/VERSION/...
	// if we found the latter we just move into "hardware" folder and continue
	var packagePath string
	hardwareSubdirPath := filepath.Join(path, "hardware")
	if info, err := os.Stat(hardwareSubdirPath); err == nil && info.IsDir() {
		packagePath = hardwareSubdirPath
	} else if info, err := os.Stat(path); err == nil && info.IsDir() {
		packagePath = path
	} else {
		// do nothing
		return nil
	}

	targetPackage := packages.getOrCreatePackage(packager)
	if err := targetPackage.load(packagePath); err != nil {
		return fmt.Errorf("loading package %s: %s", packager, err)
	}
	return nil
}

func (packages *Packages) getOrCreatePackage(packager string) *Package {
	if targetPackage, ok := packages.Packages[packager]; ok {
		return targetPackage
	}
	targetPackage := &Package{
		Name:      packager,
		Platforms: map[string]*Platform{},
		Packages:  packages,
		//Properties: properties.Map{},
	}
	packages.Packages[packager] = targetPackage
	return targetPackage
}

func (targetPackage *Package) load(folder string) error {
	// packagePlatformTxt, err := properties.SafeLoad(filepath.Join(folder, constants.FILE_PLATFORM_TXT))
	// if err != nil {
	// 	return err
	// }
	// targetPackage.Properties.Merge(packagePlatformTxt)

	files, err := ioutil.ReadDir(folder)
	if err != nil {
		return fmt.Errorf("reading directory %s: %s", folder, err)
	}

	for _, file := range files {
		architecure := file.Name()
		platformPath := filepath.Join(folder, architecure)
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
			// There is a boards.txt here, this is an unversioned Platform

			platform := targetPackage.getOrCreatePlatform(architecure)
			release := platform.getOrCreateRelease("")
			if err := release.load(platformPath); err != nil {
				return fmt.Errorf("loading platform release: %s", err)
			}

		} else if os.IsNotExist(err) {
			// There are no boards.txt here, let's fetch version folders

			platform := targetPackage.getOrCreatePlatform(architecure)
			versionDirs, err := ioutil.ReadDir(platformPath)
			if err != nil {
				return fmt.Errorf("reading dir %s: %s", platformPath, err)
			}
			for _, versionDir := range versionDirs {
				if !versionDir.IsDir() {
					continue
				}
				version := versionDir.Name()
				release := platform.getOrCreateRelease(version)
				platformWithVersionPath := filepath.Join(platformPath, version)

				if err := release.load(platformWithVersionPath); err != nil {
					return fmt.Errorf("loading platform release %s: %s", version, err)
				}
			}
		} else {
			return fmt.Errorf("looking for boards.txt in %s: %s", possibleBoardTxtPath, err)
		}
	}

	return nil
}

func (targetPackage *Package) getOrCreatePlatform(architecure string) *Platform {
	if platform, ok := targetPackage.Platforms[architecure]; ok {
		return platform
	}
	targetPlatform := &Platform{
		Architecture: architecure,
		Releases:     map[string]*PlatformRelease{},
		Package:      targetPackage,
	}
	targetPackage.Platforms[architecure] = targetPlatform
	return targetPlatform
}

func (platform *Platform) getOrCreateRelease(version string) *PlatformRelease {
	if release, ok := platform.Releases[version]; ok {
		return release
	}
	release := &PlatformRelease{
		Boards:      map[string]*Board{},
		Properties:  properties.Map{},
		Programmers: map[string]properties.Map{},
		Platform:    platform,
	}
	platform.Releases[version] = release
	return release
}

func (platform *PlatformRelease) load(folder string) error {
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

	if err := platform.loadBoards(); err != nil {
		return err
	}

	return nil
}

func (platform *PlatformRelease) loadBoards() error {
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
		board := platform.getOrCreateBoard(boardID)
		board.Properties.Merge(boardProperties)
		platform.Boards[boardID] = board
	}

	return nil
}

func (platform *PlatformRelease) getOrCreateBoard(boardID string) *Board {
	if board, ok := platform.Boards[boardID]; ok {
		return board
	}
	return &Board{
		BoardId:         boardID,
		Properties:      properties.Map{},
		PlatformRelease: platform,
	}
}
