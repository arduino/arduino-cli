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

package test

import (
	"testing"

	"github.com/arduino/arduino-cli/arduino/sketch"
	"github.com/arduino/arduino-cli/legacy/builder"
	"github.com/arduino/arduino-cli/legacy/builder/types"
	"github.com/arduino/go-paths-helper"
	"github.com/stretchr/testify/require"
)

func TestCreateBuildOptionsMap(t *testing.T) {
	ctx := &types.Context{
		HardwareDirs:       paths.NewPathList("hardware", "hardware2"),
		BuiltInToolsDirs:   paths.NewPathList("tools"),
		OtherLibrariesDirs: paths.NewPathList("libraries"),
		Sketch:             &sketch.Sketch{FullPath: paths.New("sketchLocation")},
		FQBN:               parseFQBN(t, "my:nice:fqbn"),
		Verbose:            true,
		BuildPath:          paths.New("buildPath"),
		OptimizationFlags:  "-Os",
	}

	create := builder.CreateBuildOptionsMap{}
	err := create.Run(ctx)
	NoError(t, err)

	require.Equal(t, `{
  "additionalFiles": "",
  "builtInToolsFolders": "tools",
  "compiler.optimization_flags": "-Os",
  "customBuildProperties": "",
  "fqbn": "my:nice:fqbn",
  "hardwareFolders": "hardware,hardware2",
  "otherLibrariesFolders": "libraries",
  "sketchLocation": "sketchLocation"
}`, ctx.BuildOptionsJson)
}
