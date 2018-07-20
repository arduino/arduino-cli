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

package commands

import (
	"fmt"
	"os"

	"github.com/arduino/go-paths-helper"

	"github.com/bcmi-labs/arduino-cli/arduino/libraries"
	"github.com/bcmi-labs/arduino-cli/arduino/libraries/librariesmanager"
	"github.com/bcmi-labs/arduino-cli/arduino/sketches"
	"github.com/bcmi-labs/arduino-cli/configs"
	sk "github.com/bcmi-labs/arduino-modules/sketches"

	"github.com/bcmi-labs/arduino-cli/arduino/cores/packagemanager"
	"github.com/bcmi-labs/arduino-cli/common/formatter"
	"github.com/sirupsen/logrus"
)

// Error codes to be used for os.Exit().
const (
	_          = iota // 0 is not a valid exit error code
	ErrGeneric        // 1 is the reserved "catchall" code in Unix
	_                 // 2 is reserved in Unix
	ErrNoConfigFile
	ErrBadCall
	ErrNetwork
	ErrCoreConfig // Represents an error in the cli core config, for example some basic files shipped with the installation are missing, or cannot create or get basic folder vital for the CLI to work.
	ErrBadArgument

	Version = "0.1.0-alpha.preview"
)

// ErrLogrus represents the logrus instance, which has the role to
// log all non info messages.
var ErrLogrus = logrus.New()

// GlobalFlags represents flags available in all the program.
var GlobalFlags struct {
	Debug  bool   // If true, dump debug output to stderr.
	Format string // The Output format (e.g. text, json).
}

// InitPackageManager initializes the PackageManager
// TODO: for the daemon mode, this might be called at startup, but for now only commands needing the PM will call it
func InitPackageManager() *packagemanager.PackageManager {
	logrus.Info("Loading the default Package index")
	pm := packagemanager.NewPackageManager()

	for _, URL := range configs.BoardManagerAdditionalUrls {
		if err := pm.LoadPackageIndex(URL); err != nil {
			formatter.PrintError(err, "Failed to load "+URL.String()+" package index.\n"+
				"Try updating all indexes with `arduino core update-index`.")
			os.Exit(ErrCoreConfig)
		}
	}

	// TODO: were should we register the event handler? Multiple places?
	if len(pm.GetEventHandlers()) == 0 {
		// During tests this could get registered multiple times,
		// since there is an underlying singleton
		pm.RegisterEventHandler(&CLIPackageManagerEventHandler{})
	}

	if err := pm.LoadHardware(); err != nil {
		formatter.PrintError(err, "Error loading hardware packages.")
		os.Exit(ErrCoreConfig)
	}

	return pm
}

// InitLibraryManager initialize the LibraryManager using the underlying packagemanager
func InitLibraryManager(pm *packagemanager.PackageManager) *librariesmanager.LibrariesManager {
	logrus.Info("Starting libraries manager")
	lm := librariesmanager.NewLibraryManager()

	// Add IDE builtin libraries dir
	if bundledLibsDir := configs.IDEBundledLibrariesDir(); bundledLibsDir != nil {
		lm.AddLibrariesDir(bundledLibsDir, libraries.IDEBuiltIn)
	}

	// Add sketchbook libraries dir
	if libHome, err := configs.LibrariesFolder.Get(); err != nil {
		formatter.PrintError(err, "Cannot get libraries folder.")
		os.Exit(ErrCoreConfig)
	} else {
		lm.AddLibrariesDir(paths.New(libHome), libraries.Sketchbook)
	}

	// Add libraries dirs from installed platforms
	if pm != nil {
		for _, targetPackage := range pm.GetPackages().Packages {
			for _, platform := range targetPackage.Platforms {
				if platformRelease := platform.GetInstalled(); platformRelease != nil {
					lm.AddPlatformReleaseLibrariesDir(platformRelease, libraries.PlatformBuiltIn)
				}
			}
		}
	}

	// Auto-update index if needed
	if err := lm.LoadIndex(); err != nil {
		logrus.WithError(err).Warn("Error during libraries index loading, trying to auto-update index")
		UpdateLibrariesIndex()
	}
	if err := lm.LoadIndex(); err != nil {
		logrus.WithError(err).Error("Error during libraries index loading")
		formatter.PrintError(err, "Error loading libraries index")
		os.Exit(ErrGeneric)
	}

	// Scan for libraries
	if err := lm.RescanLibraries(); err != nil {
		logrus.WithError(err).Error("Error during libraries rescan")
		formatter.PrintError(err, "Error during libraries rescan")
		os.Exit(ErrGeneric)
	}

	return lm
}

// UpdateLibrariesIndex updates the library_index.json
func UpdateLibrariesIndex() {
	logrus.Info("Updating libraries index")
	resp, err := librariesmanager.UpdateIndex()
	if err != nil {
		formatter.PrintError(err, "Error downloading librarires index")
		os.Exit(ErrNetwork)
	}
	formatter.DownloadProgressBar(resp, "Updating index: library_index.json")
	if resp.Err() != nil {
		formatter.PrintError(resp.Err(), "Error downloading librarires index")
		os.Exit(ErrNetwork)
	}
}

func InitSketch(sketchPath *paths.Path) (*sk.Sketch, error) {
	if sketchPath != nil {
		return sketches.NewSketchFromPath(sketchPath)
	}

	wd, err := paths.Getwd()
	if err != nil {
		return nil, fmt.Errorf("getting current directory: %s", err)
	}
	logrus.Infof("Reading sketch from dir: %s", wd)
	return sketches.NewSketchFromPath(wd)
}

// CLIPackageManagerEventHandler defines an event handler which outputs the PackageManager events
// in the CLI format
type CLIPackageManagerEventHandler struct{}

// Implement packagemanager.EventHandler interface

// END -- Implement packagemanager.EventHandler interface
