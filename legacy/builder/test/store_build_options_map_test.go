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
	"github.com/arduino/arduino-cli/legacy/builder/constants"
	"github.com/arduino/arduino-cli/legacy/builder/types"
	paths "github.com/arduino/go-paths-helper"
	"github.com/arduino/go-properties-orderedmap"
	"github.com/stretchr/testify/require"
)

func TestStoreBuildOptionsMap(t *testing.T) {
	ctx := &types.Context{
		HardwareDirs:          paths.NewPathList("hardware"),
		BuiltInToolsDirs:      paths.NewPathList("tools"),
		BuiltInLibrariesDirs:  paths.New("built-in libraries"),
		OtherLibrariesDirs:    paths.NewPathList("libraries"),
		Sketch:                &sketch.Sketch{FullPath: paths.New("sketchLocation")},
		FQBN:                  parseFQBN(t, "my:nice:fqbn"),
		CustomBuildProperties: []string{"custom=prop"},
		Verbose:               true,
		BuildProperties:       properties.NewFromHashmap(map[string]string{"compiler.optimization_flags": "-Os"}),
	}

	buildPath := SetupBuildPath(t, ctx)
	defer buildPath.RemoveAll()

	buildPropertiesJSON, err := builder.CreateBuildOptionsMap(
		ctx.HardwareDirs, ctx.BuiltInToolsDirs, ctx.OtherLibrariesDirs,
		ctx.BuiltInLibrariesDirs, ctx.Sketch, ctx.CustomBuildProperties,
		ctx.FQBN.String(), ctx.BuildProperties.Get("compiler.optimization_flags"),
	)
	require.NoError(t, err)
	ctx.BuildOptionsJson = buildPropertiesJSON

	err = builder.StoreBuildOptionsMap(ctx.BuildPath, ctx.BuildOptionsJson)
	require.NoError(t, err)

	exist, err := buildPath.Join(constants.BUILD_OPTIONS_FILE).ExistCheck()
	require.NoError(t, err)
	require.True(t, exist)

	bytes, err := buildPath.Join(constants.BUILD_OPTIONS_FILE).ReadFile()
	require.NoError(t, err)

	require.Equal(t, `{
  "additionalFiles": "",
  "builtInLibrariesFolders": "built-in libraries",
  "builtInToolsFolders": "tools",
  "compiler.optimization_flags": "-Os",
  "customBuildProperties": "custom=prop",
  "fqbn": "my:nice:fqbn",
  "hardwareFolders": "hardware",
  "otherLibrariesFolders": "libraries",
  "sketchLocation": "sketchLocation"
}`, string(bytes))
}
