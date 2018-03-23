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

	"github.com/bcmi-labs/arduino-cli/configs"
	"github.com/bcmi-labs/arduino-cli/sketches"
	sk "github.com/bcmi-labs/arduino-modules/sketches"

	"github.com/bcmi-labs/arduino-cli/common/formatter"
	"github.com/bcmi-labs/arduino-cli/common/releases"
	"github.com/bcmi-labs/arduino-cli/cores/packagemanager"
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
	// FIXME: Parse all 3rd party indexes
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
	return pm
}

func InitSketch(sketchPath string) (*sk.Sketch, error) {
	if sketchPath != "" {
		return sketches.NewSketchFromPath(sketchPath)
	}

	wd, err := os.Getwd()
	logrus.Infof("Reading sketch from dir: %s", wd)
	if err != nil {
		return nil, fmt.Errorf("getting current directory: %s", err)
	}
	return sketches.NewSketchFromPath(wd)
}

// CLIPackageManagerEventHandler defines an event handler which outputs the PackageManager events
// in the CLI format
type CLIPackageManagerEventHandler struct{}

// Implement packagemanager.EventHandler interface
func (cliEH *CLIPackageManagerEventHandler) OnDownloadingSomething() releases.ParallelDownloadProgressHandler {
	return GenerateDownloadProgressFormatter()
}

// END -- Implement packagemanager.EventHandler interface

// FIXME: Move away? Where should the display logic reside; in the formatter?
func GenerateDownloadProgressFormatter() releases.ParallelDownloadProgressHandler {
	if formatter.IsCurrentFormat("text") {
		return &ProgressBarFormatter{}
	}
	return nil
}
