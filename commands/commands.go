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

package commands

import (
	"fmt"
	"os"
	"path/filepath"

	paths "github.com/arduino/go-paths-helper"

	"github.com/arduino/arduino-cli/arduino/libraries"
	"github.com/arduino/arduino-cli/arduino/libraries/librariesmanager"
	"github.com/arduino/arduino-cli/arduino/sketches"
	"github.com/arduino/arduino-cli/configs"

	"github.com/arduino/arduino-cli/arduino/cores/packagemanager"
	"github.com/arduino/arduino-cli/common/formatter"
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
	// ErrCoreConfig represents an error in the cli core config, for example some basic
	// files shipped with the installation are missing, or cannot create or get basic
	// directories vital for the CLI to work.
	ErrCoreConfig
	ErrBadArgument
)

// Version is the current CLI version
var Version = "0.3.4-alpha.preview"

// ErrLogrus represents the logrus instance, which has the role to
// log all non info messages.
var ErrLogrus = logrus.New()

// GlobalFlags represents flags available in all the program.
var GlobalFlags struct {
	Debug  bool   // If true, dump debug output to stderr.
	Format string // The Output format (e.g. text, json).
}

// AppName is the command line name of the Arduino CLI executable
var AppName = filepath.Base(os.Args[0])

var Config *configs.Configuration

// InitPackageManagerWithoutBundles initializes the PackageManager
// but ignores bundles and user installed cores
func InitPackageManagerWithoutBundles() *packagemanager.PackageManager {
	logrus.Info("Package manager will scan only managed hardware folder")

	fakeResult := false
	Config.IDEBundledCheckResult = &fakeResult
	Config.SketchbookDir = nil
	return InitPackageManager()
}

// InitPackageManager initializes the PackageManager
// TODO: for the daemon mode, this might be called at startup, but for now only commands needing the PM will call it
func InitPackageManager() *packagemanager.PackageManager {
	logrus.Info("Initializing package manager")

	pm := packagemanager.NewPackageManager(
		Config.IndexesDir(),
		Config.PackagesDir(),
		Config.DownloadsDir(),
		Config.DataDir.Join("tmp"))

	for _, URL := range Config.BoardManagerAdditionalUrls {
		if err := pm.LoadPackageIndex(URL); err != nil {
			formatter.PrintError(err, "Failed to load "+URL.String()+" package index.\n"+
				"Try updating all indexes with `"+AppName+" core update-index`.")
			os.Exit(ErrCoreConfig)
		}
	}

	if err := pm.LoadHardware(Config); err != nil {
		formatter.PrintError(err, "Error loading hardware packages.")
		os.Exit(ErrCoreConfig)
	}

	return pm
}

// InitLibraryManager initializes the LibraryManager using the underlying packagemanager
func InitLibraryManager(pm *packagemanager.PackageManager) *librariesmanager.LibrariesManager {
	logrus.Info("Starting libraries manager")
	lm := librariesmanager.NewLibraryManager(
		Config.IndexesDir(),
		Config.DownloadsDir())

	// Add IDE builtin libraries dir
	if bundledLibsDir := Config.IDEBundledLibrariesDir(); bundledLibsDir != nil {
		lm.AddLibrariesDir(bundledLibsDir, libraries.IDEBuiltIn)
	}

	// Add sketchbook libraries dir
	lm.AddLibrariesDir(Config.LibrariesDir(), libraries.Sketchbook)

	// Add libraries dirs from installed platforms
	if pm != nil {
		for _, targetPackage := range pm.GetPackages().Packages {
			for _, platform := range targetPackage.Platforms {
				if platformRelease := pm.GetInstalledPlatformRelease(platform); platformRelease != nil {
					lm.AddPlatformReleaseLibrariesDir(platformRelease, libraries.PlatformBuiltIn)
				}
			}
		}
	}

	// Auto-update index if needed
	if err := lm.LoadIndex(); err != nil {
		logrus.WithError(err).Warn("Error during libraries index loading, trying to auto-update index")
		UpdateLibrariesIndex(lm)
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
func UpdateLibrariesIndex(lm *librariesmanager.LibrariesManager) {
	logrus.Info("Updating libraries index")
	d, err := lm.UpdateIndex()
	if err != nil {
		formatter.PrintError(err, "Error downloading librarires index")
		os.Exit(ErrNetwork)
	}
	formatter.DownloadProgressBar(d, "Updating index: library_index.json")
	if d.Error() != nil {
		formatter.PrintError(d.Error(), "Error downloading librarires index")
		os.Exit(ErrNetwork)
	}
}

func InitSketch(sketchPath *paths.Path) (*sketches.Sketch, error) {
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
