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

	"github.com/arduino/arduino-cli/arduino/builder"
	"github.com/arduino/arduino-cli/arduino/sketch"
	paths "github.com/arduino/go-paths-helper"
	"github.com/arduino/go-properties-orderedmap"
	"github.com/stretchr/testify/require"
)

func TestCheckIfBuildOptionsChanged(t *testing.T) {
	hardwareDirs := paths.NewPathList("hardware")
	builtInToolsDirs := paths.NewPathList("tools")
	builtInLibrariesDirs := paths.New("built-in libraries")
	otherLibrariesDirs := paths.NewPathList("libraries")
	fqbn := parseFQBN(t, "my:nice:fqbn")

	buildPath := SetupBuildPath(t)
	defer buildPath.RemoveAll()

	buildProperties := properties.NewFromHashmap(map[string]string{"compiler.optimization_flags": "-Os"})
	buildOptionsManager := builder.NewBuildOptionsManager(
		hardwareDirs, builtInToolsDirs, otherLibrariesDirs,
		builtInLibrariesDirs, buildPath, &sketch.Sketch{FullPath: paths.New("sketchLocation")}, []string{"custom=prop"},
		fqbn.String(), false, 
		buildProperties.Get("compiler.optimization_flags"),
		buildProperties.GetPath("runtime.platform.path"),
		buildProperties.GetPath("build.core.path"),
		nil,
	)

	err := buildOptionsManager.WipeBuildPath()
	require.NoError(t, err)

	exist, err := buildPath.Join("build.options.json").ExistCheck()
	require.NoError(t, err)
	require.True(t, exist)

	bytes, err := buildPath.Join("build.options.json").ReadFile()
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

//func TestWipeoutBuildPathIfBuildOptionsChanged(t *testing.T) {
//	buildPath := SetupBuildPath(t)
//	defer buildPath.RemoveAll()
//
//	buildOptionsJsonPrevious := "{ \"old\":\"old\" }"
//	buildOptionsJson := "{ \"new\":\"new\" }"
//
//	buildPath.Join("should_be_deleted.txt").Truncate()
//
//	_, err := builder.WipeoutBuildPathIfBuildOptionsChanged(
//		false,
//		buildPath,
//		buildOptionsJson,
//		buildOptionsJsonPrevious,
//		nil,
//	)
//	require.NoError(t, err)
//
//	exist, err := buildPath.ExistCheck()
//	require.NoError(t, err)
//	require.True(t, exist)
//
//	files, err := buildPath.ReadDir()
//	require.NoError(t, err)
//	require.Equal(t, 0, len(files))
//
//	exist, err = buildPath.Join("should_be_deleted.txt").ExistCheck()
//	require.NoError(t, err)
//	require.False(t, exist)
//}
//
//func TestWipeoutBuildPathIfBuildOptionsChangedNoPreviousBuildOptions(t *testing.T) {
//	buildPath := SetupBuildPath(t)
//	defer buildPath.RemoveAll()
//
//	buildOptionsJson := "{ \"new\":\"new\" }"
//
//	require.NoError(t, buildPath.Join("should_not_be_deleted.txt").Truncate())
//
//	_, err := builder.WipeoutBuildPathIfBuildOptionsChanged(
//		false,
//		buildPath,
//		buildOptionsJson,
//		"",
//		nil,
//	)
//	require.NoError(t, err)
//
//	exist, err := buildPath.ExistCheck()
//	require.NoError(t, err)
//	require.True(t, exist)
//
//	files, err := buildPath.ReadDir()
//	require.NoError(t, err)
//	require.Equal(t, 1, len(files))
//
//	exist, err = buildPath.Join("should_not_be_deleted.txt").ExistCheck()
//	require.NoError(t, err)
//	require.True(t, exist)
//}
