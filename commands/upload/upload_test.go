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

package upload

import (
	"fmt"
	"testing"

	"github.com/arduino/arduino-cli/arduino/cores"
	"github.com/arduino/arduino-cli/arduino/sketches"
	paths "github.com/arduino/go-paths-helper"
	"github.com/stretchr/testify/require"
)

func TestDetectSketchNameFromBuildPath(t *testing.T) {
	sk1, err1 := detectSketchNameFromBuildPath(paths.New("testdata/build_path_1"))
	require.NoError(t, err1)
	require.Equal(t, "sketch.ino", sk1)

	sk2, err2 := detectSketchNameFromBuildPath(paths.New("testdata/build_path_2"))
	require.NoError(t, err2)
	require.Equal(t, "Blink.ino", sk2)

	sk3, err3 := detectSketchNameFromBuildPath(paths.New("testdata/build_path_3"))
	require.Error(t, err3)
	require.Equal(t, "", sk3)

	sk4, err4 := detectSketchNameFromBuildPath(paths.New("testdata/build_path_4"))
	require.Error(t, err4)
	require.Equal(t, "", sk4)

	sk5, err5 := detectSketchNameFromBuildPath(paths.New("testdata/build_path_invalid"))
	require.Error(t, err5)
	require.Equal(t, "", sk5)
}

func TestDetermineBuildPathAndSketchName(t *testing.T) {
	type test struct {
		importFile    string
		importDir     string
		sketch        *sketches.Sketch
		fqbn          *cores.FQBN
		resBuildPath  string
		resSketchName string
	}

	blonk, err := sketches.NewSketchFromPath(paths.New("testdata/Blonk"))
	require.NoError(t, err)

	fqbn, err := cores.ParseFQBN("arduino:samd:mkr1000")
	require.NoError(t, err)

	tests := []test{
		// 00: error: no data passed in
		{"", "", nil, nil, "<nil>", ""},
		// 01: use importFile to detect build.path and project_name
		{"testdata/build_path_2/Blink.ino.hex", "", nil, nil, "testdata/build_path_2", "Blink.ino"},
		// 02: use importPath as build.path and project_name
		{"", "testdata/build_path_2", nil, nil, "testdata/build_path_2", "Blink.ino"},
		// 03: error: used both importPath and importFile
		{"testdata/build_path_2/Blink.ino.hex", "testdata/build_path_2", nil, nil, "<nil>", ""},
		// 04: error: only sketch without FQBN
		{"", "", blonk, nil, "<nil>", ""},
		// 05: use importFile to detect build.path and project_name, sketch is ignored.
		{"testdata/build_path_2/Blink.ino.hex", "", blonk, nil, "testdata/build_path_2", "Blink.ino"},
		// 06: use importPath as build.path and Blink as project name, ignore the sketch Blonk
		{"", "testdata/build_path_2", blonk, nil, "testdata/build_path_2", "Blink.ino"},
		// 07: error: used both importPath and importFile
		{"testdata/build_path_2/Blink.ino.hex", "testdata/build_path_2", blonk, nil, "<nil>", ""},

		// 08: error: no data passed in
		{"", "", nil, fqbn, "<nil>", ""},
		// 09: use importFile to detect build.path and project_name, fqbn ignored
		{"testdata/build_path_2/Blink.ino.hex", "", nil, fqbn, "testdata/build_path_2", "Blink.ino"},
		// 10: use importPath as build.path and project_name, fqbn ignored
		{"", "testdata/build_path_2", nil, fqbn, "testdata/build_path_2", "Blink.ino"},
		// 11: error: used both importPath and importFile
		{"testdata/build_path_2/Blink.ino.hex", "testdata/build_path_2", nil, fqbn, "<nil>", ""},
		// 12: use sketch to determine project name and sketch+fqbn to determine build path
		{"", "", blonk, fqbn, "testdata/Blonk/build/arduino.samd.mkr1000", "Blonk.ino"},
		// 13: use importFile to detect build.path and project_name, sketch+fqbn is ignored.
		{"testdata/build_path_2/Blink.ino.hex", "", blonk, fqbn, "testdata/build_path_2", "Blink.ino"},
		// 14: use importPath as build.path and Blink as project name, ignore the sketch Blonk, ignore fqbn
		{"", "testdata/build_path_2", blonk, fqbn, "testdata/build_path_2", "Blink.ino"},
		// 15: error: used both importPath and importFile
		{"testdata/build_path_2/Blink.ino.hex", "testdata/build_path_2", blonk, fqbn, "<nil>", ""},

		// 16: importPath containing multiple firmwares, but one has the same name as the containing folder
		{"", "testdata/firmware", nil, fqbn, "testdata/firmware", "firmware.ino"},
		// 17: importFile among multiple firmwares
		{"testdata/firmware/another_firmware.ino.bin", "", nil, fqbn, "testdata/firmware", "another_firmware.ino"},
	}
	for i, test := range tests {
		t.Run(fmt.Sprintf("SubTest%02d", i), func(t *testing.T) {
			buildPath, sketchName, err := determineBuildPathAndSketchName(test.importFile, test.importDir, test.sketch, test.fqbn)
			if test.resBuildPath == "<nil>" {
				require.Error(t, err)
				require.Nil(t, buildPath)
			} else {
				require.NoError(t, err)
				resBuildPath := paths.New(test.resBuildPath)
				require.NoError(t, resBuildPath.ToAbs())
				require.NotNil(t, buildPath)
				require.NoError(t, buildPath.ToAbs())
				require.Equal(t, resBuildPath.String(), buildPath.String())
			}
			require.Equal(t, test.resSketchName, sketchName)
		})
	}
}
