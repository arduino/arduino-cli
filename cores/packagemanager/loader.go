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
)

// LoadHardwareFromDirectories read all plaforms from the configured paths
func (pm *packageManager) LoadHardware() error {
	dirs, err := configs.HardwareDirectories()
	if err != nil {
		return fmt.Errorf("getting hardware folder: %s", err)
	}
	return pm.LoadHardwareFromDirectories(dirs)
}

// LoadHardwareFromDirectories read all plaforms from a set of directories
func (pm *packageManager) LoadHardwareFromDirectories(hardwarePaths []string) error {
	for _, path := range hardwarePaths {
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
			// First exclude all "tools" folders
			packager := packagerPathInfo.Name()
			if packager == "tools" {
				continue
			}

			// Follow symlinks
			packagerPath := filepath.Join(path, packager)
			packagerPath, err := filepath.EvalSymlinks(packagerPath)
			if err != nil {
				return fmt.Errorf("following possible symlink %s: %s", path, err)
			}

			// There are two possible package folder structures:
			// - PACKAGER/ARCHITECTURE/...
			// - PACKAGER/hardware/ARCHITECTURE/VERSION/...
			// if we found the latter we just move into "hardware" folder and continue
			hardwareSubdirPath := filepath.Join(packagerPath, "hardware")
			if info, err := os.Stat(hardwareSubdirPath); err == nil && info.IsDir() {
				// move down into the "hardware" folder
				packagerPath = hardwareSubdirPath
			} else if info, err := os.Stat(packagerPath); err == nil && info.IsDir() {
				// we are already at the correct level
			} else {
				// error: do nothing
				continue
			}

			targetPackage := pm.packages.GetOrCreatePackage(packager)
			if err := targetPackage.Load(packagerPath); err != nil {
				return fmt.Errorf("loading package %s: %s", packager, err)
			}
		}
	}

	return nil
}
