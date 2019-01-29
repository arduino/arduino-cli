/*
 * This file is part of arduino-cli.
 *
 * Copyright 2018 ARDUINO SA (http://www.arduino.cc/)
 *
 * This software is released under the GNU General Public License version 3,
 * which covers the main part of arduino-cli.
 * The terms of this license can be found at:
 * https://www.gnu.org/licenses/gpl-3.0.en.html
 *
 * You can be released from the requirements of the above licenses by purchasing
 * a commercial license. Buying such a license is mandatory if you want to modify or
 * otherwise use the software for commercial activities involving the Arduino
 * software without disclosing the source code of your own applications. To purchase
 * a commercial license, send an email to license@arduino.cc.
 */

package configs

import (
	"fmt"
	"os/user"
	"runtime"

	paths "github.com/arduino/go-paths-helper"
	win32 "github.com/arduino/go-win32-utils"
	homedir "github.com/mitchellh/go-homedir"
)

// getDefaultConfigFilePath returns the default path for arduino-cli.yaml
func getDefaultConfigFilePath() *paths.Path {
	usr, err := user.Current()
	if err != nil {
		panic(fmt.Errorf("retrieving user home dir: %s", err))
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
			panic(err)
		}
		arduinoDataDir = paths.New(localAppDataPath).Join("Arduino15")
	default:
		panic(fmt.Errorf("unsupported OS: %s", runtime.GOOS))
	}

	return arduinoDataDir.Join("arduino-cli.yaml")
}

func getDefaultArduinoDataDir() (*paths.Path, error) {
	homedir, err := homedir.Dir()
	if err != nil {
		return nil, fmt.Errorf("retrieving user home dir: %s", err)
	}
	arduinoDataDir := paths.New(homedir)

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
	homedir, err := homedir.Dir()
	if err != nil {
		return nil, fmt.Errorf("retrieving home dir: %s", err)
	}

	switch runtime.GOOS {
	case "linux":
		return paths.New(homedir).Join("Arduino"), nil
	case "darwin":
		return paths.New(homedir).Join("Documents", "Arduino"), nil
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
