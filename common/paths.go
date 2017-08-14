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
 * Copyright 2017 BCMI LABS SA (http://www.arduino.cc/)
 */

package common

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/bcmi-labs/arduino-cli/cmd/formatter"
	"github.com/bcmi-labs/arduino-cli/task"
)

// RootDirPath represents the current root of the arduino tree.
var RootDirPath = ""

// GetFolder gets a folder on a path, and creates it if createIfMissing == true and not found.
func GetFolder(folder string, label string, createIfMissing bool) (string, error) {
	_, err := os.Stat(folder)
	if os.IsNotExist(err) && createIfMissing {
		formatter.Print(fmt.Sprintf("Cannot find default %s folder, attemping to create it ...", label))
		err = os.MkdirAll(folder, 0755)
		if err != nil {
			formatter.Print("ERROR")
			formatter.PrintErrorMessage(fmt.Sprintf("Folder %s missing and cannot create it", label))
			return "", err
		}
		formatter.Print("OK")
	} else if err != nil {
		msgFormat := "Cannot get %s folder, it does not exist"
		formatter.PrintErrorMessage(fmt.Sprintf(msgFormat, label))
		return "", fmt.Errorf(msgFormat, label)
	}
	return folder, nil
}

// GetDefaultArduinoFolder returns the default data folder for Arduino platform
func GetDefaultArduinoFolder() (string, error) {
	var folder string
	switch runtime.GOOS {
	case "linux":
		folder = filepath.Join(RootDirPath, ".arduino15")
	case "darwin":
		folder = filepath.Join(RootDirPath, "Library", "arduino15")
	default:
		return folder, fmt.Errorf("Unsupported OS: %s", runtime.GOOS)
	}
	return GetFolder(folder, "default arduino", true)
}

// GetDefaultArduinoHomeFolder gets the home directory for arduino CLI.
func GetDefaultArduinoHomeFolder() (string, error) {
	return GetFolder(filepath.Join(RootDirPath, "Arduino"), "Arduino home", true)
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
