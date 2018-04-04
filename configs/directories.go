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

// Package configs contains all CLI configurations handling.
package configs

import (
	"fmt"
	"net/url"
	"os"
	"os/user"
	"path"
	"path/filepath"
	"runtime"

	"github.com/arduino/go-win32-utils"
	"github.com/bcmi-labs/arduino-cli/pathutils"
)

// ConfigFilePath represents the default location of the config file (same directory as executable).
var ConfigFilePath = getDefaultConfigFilePath()

// ArduinoDataFolder represents the current root of the arduino tree (defaulted to `$HOME/.arduino15` on linux).
var ArduinoDataFolder = pathutils.NewPath("Arduino Data", getDefaultArduinoDataFolder, true)

// SketchbookFolder represents the current root of the sketchbooks tree (defaulted to `$HOME/Arduino`).
var SketchbookFolder = pathutils.NewPath("Sketchbook", getDefaultSketchbookFolder, true)

// TODO: Remove either ArduinoHomeFolder or SketchbookFolder
var ArduinoHomeFolder = SketchbookFolder

// LibrariesFolder is the default folder of downloaded libraries.
var LibrariesFolder = pathutils.NewSubPath("libraries", ArduinoHomeFolder, "libraries", true)

// PackagesFolder is the default folder of downloaded packages.
var PackagesFolder = pathutils.NewSubPath("packages", ArduinoDataFolder, "packages", true)

// CoresFolder gets the default folder of downloaded cores.
func CoresFolder(packageName string) pathutils.Path {
	// TODO: wrong, this is not the correct location of the cores (in Java IDE)
	return pathutils.NewSubPath("cores", PackagesFolder, filepath.Join(packageName, "hardware"), true)
}

// ToolsFolder gets the default folder of downloaded packages.
func ToolsFolder(packageName string) pathutils.Path {
	return pathutils.NewSubPath("tools", PackagesFolder, filepath.Join(packageName, "tools"), true)
}

// DownloadCacheFolder gets a generic cache folder for downloads.
func DownloadCacheFolder(item string) pathutils.Path {
	return pathutils.NewSubPath("zip archives cache", ArduinoDataFolder, filepath.Join("staging", item), true)
}

// IndexPath returns the path of the specified index file.
func IndexPath(fileName string) pathutils.Path {
	return pathutils.NewSubPath(fileName, ArduinoDataFolder, fileName, false)
}

// IndexPathFromURL returns the path of the index file corresponding to the specified URL
func IndexPathFromURL(URL *url.URL) pathutils.Path {
	filename := path.Base(URL.Path)
	return IndexPath(filename)
}

// getDefaultConfigFilePath returns the default path for .cli-config.yml,
// this is the directory where the arduino-cli executable resides.
func getDefaultConfigFilePath() string {
	fileLocation, err := os.Executable()
	if err != nil {
		fileLocation = "."
	}
	fileLocation = filepath.Dir(fileLocation)
	fileLocation = filepath.Join(fileLocation, ".cli-config.yml")
	return fileLocation
}

func getDefaultArduinoDataFolder() (string, error) {
	usr, err := user.Current()
	if err != nil {
		return "", fmt.Errorf("retrieving home dir: %s", err)
	}
	arduinoDataFolder := usr.HomeDir

	switch runtime.GOOS {
	case "linux":
		arduinoDataFolder = filepath.Join(arduinoDataFolder, ".arduino15")
	case "darwin":
		arduinoDataFolder = filepath.Join(arduinoDataFolder, "Library", "arduino15")
	case "windows":
		localAppDataPath, err := win32.GetLocalAppDataFolder()
		if err != nil {
			return "", fmt.Errorf("getting LocalAppData path: %s", err)
		}
		arduinoDataFolder = filepath.Join(localAppDataPath, "Arduino15")
	default:
		return "", fmt.Errorf("unsupported OS: %s", runtime.GOOS)
	}
	return arduinoDataFolder, nil
}

func getDefaultSketchbookFolder() (string, error) {
	usr, err := user.Current()
	if err != nil {
		return "", fmt.Errorf("retrieving home dir: %s", err)
	}

	switch runtime.GOOS {
	case "linux":
		return filepath.Join(usr.HomeDir, "Arduino"), nil
	case "darwin":
		return filepath.Join(usr.HomeDir, "Documents", "Arduino"), nil
	case "windows":
		documentsPath, err := win32.GetDocumentsFolder()
		if err != nil {
			return "", fmt.Errorf("getting Documents path: %s", err)
		}
		return filepath.Join(documentsPath, "Arduino"), nil
	default:
		return "", fmt.Errorf("unsupported OS: %s", runtime.GOOS)
	}
}
