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

package builder

import (
	"github.com/arduino/arduino-cli/arduino/sketch"
	"github.com/arduino/go-paths-helper"
)

// nolint
const (
	BuildPropertiesArchiveFile          = "archive_file"
	BuildPropertiesArchiveFilePath      = "archive_file_path"
	BuildPropertiesObjectFile           = "object_file"
	RecipeARPattern                     = "recipe.ar.pattern"
	BuildPropertiesIncludes             = "includes"
	BuildPropertiesCompilerWarningFlags = "compiler.warning_flags"
	Space                               = " "
)

// Builder is a Sketch builder.
type Builder struct {
	sketch *sketch.Sketch

	// core related
	coreBuildCachePath *paths.Path
}

// NewBuilder creates a sketch Builder.
func NewBuilder(sk *sketch.Sketch, coreBuildCachePath *paths.Path) *Builder {
	return &Builder{
		sketch:             sk,
		coreBuildCachePath: coreBuildCachePath,
	}
}
