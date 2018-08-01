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

package configs

import (
	"fmt"
	"os"
	"os/user"
	"runtime"

	"github.com/arduino/go-paths-helper"

	"github.com/arduino/go-win32-utils"
)

// getDefaultConfigFilePath returns the default path for .cli-config.yml,
// this is the directory where the arduino-cli executable resides.
func getDefaultConfigFilePath() *paths.Path {
	executablePath, err := os.Executable()
	if err != nil {
		executablePath = "."
	}
	return paths.New(executablePath).Parent().Join(".cli-config.yml")
}

func getDefaultArduinoDataDir() (*paths.Path, error) {
	usr, err := user.Current()
	if err != nil {
		return nil, fmt.Errorf("retrieving user home dir: %s", err)
	}
	arduinoDataDir := paths.New(usr.HomeDir)

	switch runtime.GOOS {
	case "linux":
		arduinoDataDir = arduinoDataDir.Join(".arduino15")
	case "darwin":
		arduinoDataDir = arduinoDataDir.Join("Library", "arduino15")
	case "windows":
		localAppDataPath, err := win32.GetLocalAppDataFolder()
		if err != nil {
			return nil, fmt.Errorf("getting LocalAppData path: %s", err)
		}
		arduinoDataDir = paths.New(localAppDataPath).Join("Arduino15")
	default:
		return nil, fmt.Errorf("unsupported OS: %s", runtime.GOOS)
	}
	return arduinoDataDir, nil
}

func getDefaultSketchbookDir() (*paths.Path, error) {
	usr, err := user.Current()
	if err != nil {
		return nil, fmt.Errorf("retrieving home dir: %s", err)
	}

	switch runtime.GOOS {
	case "linux":
		return paths.New(usr.HomeDir).Join("Arduino"), nil
	case "darwin":
		return paths.New(usr.HomeDir).Join("Documents", "Arduino"), nil
	case "windows":
		documentsPath, err := win32.GetDocumentsFolder()
		if err != nil {
			return nil, fmt.Errorf("getting Documents path: %s", err)
		}
		return paths.New(documentsPath).Join("Arduino"), nil
	default:
		return nil, fmt.Errorf("unsupported OS: %s", runtime.GOOS)
	}
}
