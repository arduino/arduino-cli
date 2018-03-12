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
	"errors"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/bcmi-labs/arduino-cli/pathutils"

	"github.com/arduino/go-properties-map"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

var arduinoIDEDirectory *string

// IsBundledInDesktopIDE returns true if the CLI is bundled with the Arduino IDE.
func IsBundledInDesktopIDE() bool {
	if arduinoIDEDirectory != nil {
		return *arduinoIDEDirectory != ""
	}
	empty := ""
	arduinoIDEDirectory = &empty

	logrus.Info("Checking if CLI is Bundled into the IDE")
	executable, err := os.Executable()
	if err != nil {
		logrus.WithError(err).Warn("Cannot get executable path")
		return false
	}
	executable, err = filepath.EvalSymlinks(executable)
	if err != nil {
		logrus.WithError(err).Warn("Cannot get executable path (symlinks error)")
		return false
	}
	ideDir := filepath.Dir(executable)
	logrus.Info("Candidate IDE Directory: ", ideDir)

	tests := []string{"tools-builder", "Examples/01.Basics/Blink"}
	for _, test := range tests {
		filePath := filepath.Join(ideDir, test)
		_, err := os.Stat(filePath)
		if !os.IsNotExist(err) {
			arduinoIDEDirectory = &ideDir
			break
		}
	}

	return *arduinoIDEDirectory != ""
}

// LoadFromDesktopIDEPreferences loads the config from the Desktop IDE preferences.txt file
func LoadFromDesktopIDEPreferences() error {
	logrus.Info("Unserializing from IDE preferences")
	dataDir, err := ArduinoDataFolder.Get()
	if err != nil {
		logrus.WithError(err).Warn("Error looking for IDE preferences")
		return err
	}
	props, err := properties.Load(filepath.Join(dataDir, "preferences.txt"))
	if err != nil {
		logrus.WithError(err).Warn("Error during unserialize from IDE preferences")
		return err
	}
	err = proxyConfigsFromIDEPrefs(props)
	if err != nil {
		logrus.WithError(err).Warn("Error during unserialize from IDE preferences")
		return err
	}
	if dir, has := props["sketchbook.path"]; has {
		SketchbookFolder = pathutils.NewConstPath(dir)
		ArduinoHomeFolder = SketchbookFolder
	}
	if URLs, has := props["boardsmanager.additional.urls"]; has {
		for _, URL := range strings.Split(URLs, ",") {
			if newURL, err := url.Parse(URL); err == nil {
				BoardManagerAdditionalUrls = append(BoardManagerAdditionalUrls, newURL)
			}
		}
	}
	return nil
}

func proxyConfigsFromIDEPrefs(props properties.Map) error {
	proxy := props.SubTree("proxy")
	switch proxy["type"] {
	case "auto":
		// Automatic proxy
		viper.Set("proxy.type", "auto")
		break
	case "manual":
		// Manual proxy configuration
		manualConfig := proxy.SubTree("manual")
		hostname, exists := manualConfig["hostname"]
		if !exists {
			return errors.New("Proxy hostname not found in preferences.txt")
		}
		username := manualConfig["username"]
		password := manualConfig["password"]

		ProxyType = "manual"
		ProxyHostname = hostname
		ProxyUsername = username
		ProxyPassword = password
		break
	case "none":
		// No proxy
		break
	default:
		return errors.New("Unsupported proxy config")
	}
	return nil
}
