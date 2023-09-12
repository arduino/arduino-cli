// This file is part of arduino-cli.
//
// Copyright 2020 ARDUINO SA (http://www.arduino.cc/)
//
// This software is released under the GNU General Public License version 3,
// which covers the main part of arduino-cli.
// The terms of this license can be found at:
// https://www.gnu.org/licenses/gpl-3.0.en.html
//
// You can be released from the requirements of the above licenses by purchasing
// a commercial license. Buying such a license is mandatory if you want to
// modify or otherwise use the software for commercial activities involving the
// Arduino software without disclosing the source code of your own applications.
// To purchase a commercial license, send an email to license@arduino.cc.

package types

import (
	"github.com/arduino/arduino-cli/arduino/builder"
	"github.com/arduino/arduino-cli/arduino/builder/compilation"
	"github.com/arduino/arduino-cli/arduino/builder/detector"
	"github.com/arduino/arduino-cli/arduino/builder/logger"
	"github.com/arduino/arduino-cli/arduino/builder/progress"
	"github.com/arduino/arduino-cli/arduino/builder/sizer"
	"github.com/arduino/arduino-cli/arduino/cores"
	"github.com/arduino/arduino-cli/arduino/cores/packagemanager"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	paths "github.com/arduino/go-paths-helper"
)

// Context structure
type Context struct {
	Builder                 *builder.Builder
	SketchLibrariesDetector *detector.SketchLibrariesDetector
	BuilderLogger           *logger.BuilderLogger

	// Build options
	HardwareDirs         paths.PathList
	BuiltInToolsDirs     paths.PathList
	BuiltInLibrariesDirs *paths.Path
	OtherLibrariesDirs   paths.PathList

	PackageManager *packagemanager.Explorer
	TargetPlatform *cores.PlatformRelease
	ActualPlatform *cores.PlatformRelease

	CoreArchiveFilePath  *paths.Path
	CoreObjectsFiles     paths.PathList
	LibrariesObjectFiles paths.PathList
	SketchObjectFiles    paths.PathList

	// C++ Parsing
	LineOffset int

	// Dry run, only create progress map
	Progress progress.Struct
	// Send progress events to this callback
	ProgressCB rpc.TaskProgressCB

	// Sizer results
	ExecutableSectionsSize sizer.ExecutablesFileSections

	// Compilation Database to build/update
	CompilationDatabase *compilation.Database
	// Set to true to skip build and produce only Compilation Database
	OnlyUpdateCompilationDatabase bool
}

func (ctx *Context) PushProgress() {
	if ctx.ProgressCB != nil {
		ctx.ProgressCB(&rpc.TaskProgress{
			Percent:   ctx.Progress.Progress,
			Completed: ctx.Progress.Progress >= 100.0,
		})
	}
}
