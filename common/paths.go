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
 * Copyright 2017 ARDUINO AG (http://www.arduino.cc/)
 */

package common

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"runtime"

	"github.com/bcmi-labs/arduino-cli/task"
)

var (
	// SketchbookFolder represents the current root of the sketchbooks tree (defaulted to `$HOME/Arduino`).
	SketchbookFolder string

	// ArduinoDataFolder represents the current root of the arduino tree (defaulted to `$HOME/.arduino15` on linux).
	ArduinoDataFolder string

	// ArduinoIDEFolder represents the current folder where the Arduino IDE relies, not used if CLI is not bundled with the IDE.
	ArduinoIDEFolder string
)

// GetFolder gets a folder on a path, and creates it if createIfMissing == true and not found.
func GetFolder(folder string, label string, createIfMissing bool) (string, error) {
	_, err := os.Stat(folder)
	if os.IsNotExist(err) && createIfMissing {
		err = os.MkdirAll(folder, 0755)
		if err != nil {
			return "", err
		}
	} else if err != nil {
		msgFormat := "Cannot get %s folder, it does not exist"
		return "", fmt.Errorf(msgFormat, label)
	}
	return folder, nil
}

// GetDefaultArduinoFolder returns the default data folder for Arduino platform
func GetDefaultArduinoFolder() (string, error) {
	if ArduinoDataFolder == "" {
		usr, err := user.Current()
		if err != nil {
			return "", err
		}
		ArduinoDataFolder = usr.HomeDir

		switch runtime.GOOS {
		case "linux":
			ArduinoDataFolder = filepath.Join(ArduinoDataFolder, ".arduino15")
			break
		case "darwin":
			ArduinoDataFolder = filepath.Join(ArduinoDataFolder, "Library", "arduino15")
			break
		case "windows":
			return "", fmt.Errorf("Windows temporarily unsupported")
		default:
			return "", fmt.Errorf("Unsupported OS: %s", runtime.GOOS)
		}
	}
	return GetFolder(ArduinoDataFolder, "default arduino", true)
}

// GetDefaultArduinoHomeFolder gets the home directory for arduino CLI.
func GetDefaultArduinoHomeFolder() (string, error) {
	if SketchbookFolder == "" {
		usr, err := user.Current()
		if err != nil {
			return "", err
		}
		// TODO: before doing this, check IDE's preferences.txt for different sketchbook path
		SketchbookFolder = filepath.Join(usr.HomeDir, "Arduino")
	}
	return GetFolder(SketchbookFolder, "Arduino home", true)
}

// GetDefaultFolder returns the default folder with specified name and label.
func GetDefaultFolder(baseFolderFunc func() (string, error), folderName string, folderLabel string, createIfMissing bool) (string, error) {
	baseFolder, err := baseFolderFunc()
	if err != nil {
		return "", err
	}
	destFolder := filepath.Join(baseFolder, folderName)
	return GetFolder(destFolder, folderLabel, createIfMissing)
}

// GetDefaultLibFolder gets the default folder of downloaded libraries.
func GetDefaultLibFolder() (string, error) {
	return GetDefaultFolder(GetDefaultArduinoHomeFolder, "libraries", "libraries", true)
}

// GetDefaultPkgFolder gets the default folder of downloaded packages.
func GetDefaultPkgFolder() (string, error) {
	return GetDefaultFolder(GetDefaultArduinoFolder, "packages", "packages", true)
}

// GetDefaultCoresFolder gets the default folder of downloaded cores.
func GetDefaultCoresFolder(packageName string) (string, error) {
	// TODO: wrong, this is not the correct location of the cores (in Java IDE)
	return GetDefaultFolder(GetDefaultPkgFolder, filepath.Join(packageName, "hardware"), "cores", true)
}

// GetDefaultToolsFolder gets the default folder of downloaded packages.
func GetDefaultToolsFolder(packageName string) (string, error) {
	return GetDefaultFolder(GetDefaultPkgFolder, filepath.Join(packageName, "tools"), "tools", true)
}

// GetDownloadCacheFolder gets a generic cache folder for downloads.
func GetDownloadCacheFolder(item string) (string, error) {
	return GetDefaultFolder(GetDefaultArduinoFolder, filepath.Join("staging", item), "cache", true)
}

// ExecUpdateIndex is a generic procedure to update an index file.
func ExecUpdateIndex(updateTask task.Wrapper, verbosity int) {
	updateTask.Execute(verbosity)
}

// IndexPath returns the path of the specified index file.
func IndexPath(fileName string) (string, error) {
	baseFolder, err := GetDefaultArduinoFolder()
	if err != nil {
		return "", err
	}
	return filepath.Join(baseFolder, fileName), nil
}
