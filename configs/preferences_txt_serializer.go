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
	"strings"

	"github.com/arduino/go-paths-helper"

	"github.com/arduino/go-properties-orderedmap"
	"github.com/sirupsen/logrus"
)

// IsBundledInDesktopIDE returns true if the CLI is bundled with the Arduino IDE.
func (config *Configuration) IsBundledInDesktopIDE() bool {
	if config.IDEBundledCheckResult != nil {
		return *config.IDEBundledCheckResult
	}

	res := false
	config.IDEBundledCheckResult = &res

	logrus.Info("Checking if CLI is Bundled into the IDE")
	executable, err := os.Executable()
	if err != nil {
		logrus.WithError(err).Warn("Cannot get executable path")
		return false
	}
	executablePath := paths.New(executable)
	if err := executablePath.FollowSymLink(); err != nil {
		logrus.WithError(err).Warn("Cannot get executable path")
		return false
	}
	ideDir := executablePath.Parent()
	logrus.Info("Candidate IDE Directory: ", ideDir)

	tests := []string{
		"tools-builder",
		"examples/01.Basics/Blink",
	}
	for _, test := range tests {
		if !ideDir.Join(test).Exist() {
			return false
		}
	}

	portable := "portable"
	if ideDir.Join(portable).Exist() {
		logrus.Info("IDE is portable")
		config.IsPortable = true
	}

	config.ArduinoIDEDirectory = ideDir
	res = true
	return true
}

// LoadFromDesktopIDEPreferences loads the config from the Desktop IDE preferences.txt file
func (config *Configuration) LoadFromDesktopIDEPreferences() error {
	logrus.Info("Unserializing from IDE preferences")
	if config.IsPortable {
		config.DataDir = config.ArduinoIDEDirectory.Join("portable")
		config.SketchbookDir = config.ArduinoIDEDirectory.Join("portable").Join("sketchbook")
	}
	preferenceTxtPath := config.DataDir.Join("preferences.txt")
	props, err := properties.LoadFromPath(preferenceTxtPath)
	if err != nil {
		logrus.WithError(err).Warn("Error during unserialize from IDE preferences")
		return err
	}
	err = config.proxyConfigsFromIDEPrefs(props)
	if err != nil {
		logrus.WithError(err).Warn("Error loading proxy settings from IDE preferences")
	}
	if dir, has := props.GetOk("sketchbook.path"); has {
		config.SketchbookDir = paths.New(dir)
	}
	if URLs, has := props.GetOk("boardsmanager.additional.urls"); has {
		for _, URL := range strings.Split(URLs, ",") {
			if newURL, err := url.Parse(URL); err == nil {
				config.BoardManagerAdditionalUrls = append(config.BoardManagerAdditionalUrls, newURL)
			}
		}
	}
	return nil
}

func (config *Configuration) proxyConfigsFromIDEPrefs(props *properties.Map) error {
	proxy := props.SubTree("proxy")
	switch proxy.Get("type") {
	case "auto":
		// Automatic proxy
		break
	case "manual":
		// Manual proxy configuration
		manualConfig := proxy.SubTree("manual")
		hostname, exists := manualConfig.GetOk("hostname")
		if !exists {
			return errors.New("proxy hostname not found in preferences.txt")
		}
		username := manualConfig.Get("username")
		password := manualConfig.Get("password")

		config.ProxyType = "manual"
		config.ProxyHostname = hostname
		config.ProxyUsername = username
		config.ProxyPassword = password
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
func (config *Configuration) IDEBundledLibrariesDir() *paths.Path {
	if config.IsBundledInDesktopIDE() {
		libDir := config.ArduinoIDEDirectory.Join("libraries")
		if libDir.IsDir() {
			return libDir
		}
	}
	return nil
}
