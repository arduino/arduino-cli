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
