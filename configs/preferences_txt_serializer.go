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
	"errors"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/arduino/go-paths-helper"

	"github.com/arduino/go-properties-map"
	"github.com/sirupsen/logrus"
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
func (config *Configuration) LoadFromDesktopIDEPreferences() error {
	logrus.Info("Unserializing from IDE preferences")
	preferenceTxtPath := config.DataDir.Join("preferences.txt")
	props, err := properties.LoadFromPath(preferenceTxtPath)
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
		config.SketchbookDir = paths.New(dir)
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
		break
	case "manual":
		// Manual proxy configuration
		manualConfig := proxy.SubTree("manual")
		hostname, exists := manualConfig["hostname"]
		if !exists {
			return errors.New("proxy hostname not found in preferences.txt")
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
		return errors.New("unsupported proxy config")
	}
	return nil
}

// IDEBundledLibrariesDir returns the libraries directory bundled in
// the Arduino IDE. If there is no Arduino IDE or the directory doesn't
// exists then nil is returned
func IDEBundledLibrariesDir() *paths.Path {
	if IsBundledInDesktopIDE() {
		libDir := paths.New(*arduinoIDEDirectory, "libraries")
		if isDir, _ := libDir.IsDir(); isDir {
			return libDir
		}
	}
	return nil
}
