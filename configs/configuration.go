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

// Package configs contains all CLI configurations handling.
package configs

import (
	"fmt"

	"github.com/arduino/go-paths-helper"
)

// Configuration contains a running configuration
type Configuration struct {
	// ConfigFilePath represents the default location of the config file (same directory as executable).
	ConfigFile *paths.Path

	// DataDir represents the current root of the arduino tree (defaulted to `$HOME/.arduino15` on linux).
	DataDir *paths.Path

	// SketchbookDir represents the current root of the sketchbooks tree (defaulted to `$HOME/Arduino`).
	SketchbookDir *paths.Path
}

// NewConfiguration returns a new Configuration with the default values
func NewConfiguration() (*Configuration, error) {
	dataDir, err := getDefaultArduinoDataDir()
	if err != nil {
		return nil, fmt.Errorf("getting default arduino data dir: %s", err)
	}
	sketchbookDir, err := getDefaultSketchbookDir()
	if err != nil {
		return nil, fmt.Errorf("getting default sketchbook dir: %s", err)
	}

	return &Configuration{
		ConfigFile:    getDefaultConfigFilePath(),
		DataDir:       dataDir,
		SketchbookDir: sketchbookDir,
	}, nil
}

// LibrariesDir returns the directory for installed libraries.
func (config *Configuration) LibrariesDir() *paths.Path {
	return config.SketchbookDir.Join("libraries")
}

// PackagesDir return the directory for installed packages.
func (config *Configuration) PackagesDir() *paths.Path {
	return config.DataDir.Join("packages")
}

// DownloadsDir returns the directory for archive downloads.
func (config *Configuration) DownloadsDir() *paths.Path {
	return config.DataDir.Join("staging")
}

// IndexesDir returns the directory for the indexes
func (config *Configuration) IndexesDir() *paths.Path {
	return config.DataDir
}
