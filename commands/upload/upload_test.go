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
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/arduino/arduino-cli/arduino/cores"
	"github.com/arduino/arduino-cli/arduino/cores/packagemanager"
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
		hasError      bool
	}

	blonk, err := sketches.NewSketchFromPath(paths.New("testdata/Blonk"))
	require.NoError(t, err)

	fqbn, err := cores.ParseFQBN("arduino:samd:mkr1000")
	require.NoError(t, err)

	tests := []test{
		// 00: error: no data passed in
		{"", "", nil, nil, "<nil>", "", true},
		// 01: use importFile to detect build.path and project_name
		{"testdata/build_path_2/Blink.ino.hex", "", nil, nil, "testdata/build_path_2", "Blink.ino", false},
		// 02: use importPath as build.path and project_name
		{"", "testdata/build_path_2", nil, nil, "testdata/build_path_2", "Blink.ino", false},
		// 03: error: used both importPath and importFile
		{"testdata/build_path_2/Blink.ino.hex", "testdata/build_path_2", nil, nil, "<nil>", "", true},
		// 04: error: only sketch without FQBN
		{"", "", blonk, nil, "<nil>", "", true},
		// 05: use importFile to detect build.path and project_name, sketch is ignored.
		{"testdata/build_path_2/Blink.ino.hex", "", blonk, nil, "testdata/build_path_2", "Blink.ino", false},
		// 06: use importPath as build.path and Blink as project name, ignore the sketch Blonk
		{"", "testdata/build_path_2", blonk, nil, "testdata/build_path_2", "Blink.ino", false},
		// 07: error: used both importPath and importFile
		{"testdata/build_path_2/Blink.ino.hex", "testdata/build_path_2", blonk, nil, "<nil>", "", true},

		// 08: error: no data passed in
		{"", "", nil, fqbn, "<nil>", "", true},
		// 09: use importFile to detect build.path and project_name, fqbn ignored
		{"testdata/build_path_2/Blink.ino.hex", "", nil, fqbn, "testdata/build_path_2", "Blink.ino", false},
		// 10: use importPath as build.path and project_name, fqbn ignored
		{"", "testdata/build_path_2", nil, fqbn, "testdata/build_path_2", "Blink.ino", false},
		// 11: error: used both importPath and importFile
		{"testdata/build_path_2/Blink.ino.hex", "testdata/build_path_2", nil, fqbn, "<nil>", "", true},
		// 12: use sketch to determine project name and sketch+fqbn to determine build path
		{"", "", blonk, fqbn, "testdata/Blonk/build/arduino.samd.mkr1000", "Blonk.ino", false},
		// 13: use importFile to detect build.path and project_name, sketch+fqbn is ignored.
		{"testdata/build_path_2/Blink.ino.hex", "", blonk, fqbn, "testdata/build_path_2", "Blink.ino", false},
		// 14: use importPath as build.path and Blink as project name, ignore the sketch Blonk, ignore fqbn
		{"", "testdata/build_path_2", blonk, fqbn, "testdata/build_path_2", "Blink.ino", false},
		// 15: error: used both importPath and importFile
		{"testdata/build_path_2/Blink.ino.hex", "testdata/build_path_2", blonk, fqbn, "<nil>", "", true},
	}
	for i, test := range tests {
		t.Run(fmt.Sprintf("SubTest%02d", i), func(t *testing.T) {
			buildPath, sketchName, err := determineBuildPathAndSketchName(test.importFile, test.importDir, test.sketch, test.fqbn)
			if test.hasError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			if test.resBuildPath == "<nil>" {
				require.Nil(t, buildPath)
			} else {
				resBuildPath := paths.New(test.resBuildPath)
				require.NoError(t, resBuildPath.ToAbs())
				require.NoError(t, buildPath.ToAbs())
				require.Equal(t, resBuildPath.String(), buildPath.String())
			}
			require.Equal(t, test.resSketchName, sketchName)
		})
	}
}

func TestUploadPropertiesComposition(t *testing.T) {
	pm := packagemanager.NewPackageManager(nil, nil, nil, nil)
	err := pm.LoadHardwareFromDirectory(paths.New("testdata", "hardware"))
	require.NoError(t, err)
	buildPath1 := paths.New("testdata", "build_path_1")

	type test struct {
		importDir      *paths.Path
		fqbn           string
		port           string
		programmer     string
		burnBootloader bool
		expected       string
	}

	tests := []test{
		// classic upload, requires port
		{buildPath1, "alice:avr:board1", "port", "", false, "conf-board1 conf-general conf-upload $$VERBOSE-VERIFY$$ protocol port -bspeed testdata/build_path_1/sketch.ino.hex"},
		{buildPath1, "alice:avr:board1", "", "", false, "FAIL"},
		// classic upload, no port
		{buildPath1, "alice:avr:board2", "port", "", false, "conf-board1 conf-general conf-upload $$VERBOSE-VERIFY$$ protocol -bspeed testdata/build_path_1/sketch.ino.hex"},
		{buildPath1, "alice:avr:board2", "", "", false, "conf-board1 conf-general conf-upload $$VERBOSE-VERIFY$$ protocol -bspeed testdata/build_path_1/sketch.ino.hex"},

		// upload with programmer, requires port
		{buildPath1, "alice:avr:board1", "port", "progr1", false, "conf-board1 conf-general conf-program $$VERBOSE-VERIFY$$ progprotocol port -bspeed testdata/build_path_1/sketch.ino.hex"},
		{buildPath1, "alice:avr:board1", "", "progr1", false, "FAIL"},
		// upload with programmer, no port
		{buildPath1, "alice:avr:board1", "port", "progr2", false, "conf-board1 conf-general conf-program $$VERBOSE-VERIFY$$ prog2protocol -bspeed testdata/build_path_1/sketch.ino.hex"},
		{buildPath1, "alice:avr:board1", "", "progr2", false, "conf-board1 conf-general conf-program $$VERBOSE-VERIFY$$ prog2protocol -bspeed testdata/build_path_1/sketch.ino.hex"},
		// upload with programmer, require port through extra params
		{buildPath1, "alice:avr:board1", "port", "progr3", false, "conf-board1 conf-general conf-program $$VERBOSE-VERIFY$$ prog3protocol port -bspeed testdata/build_path_1/sketch.ino.hex"},
		{buildPath1, "alice:avr:board1", "", "progr3", false, "FAIL"},
	}
	for i, test := range tests {
		t.Run(fmt.Sprintf("SubTest%02d", i), func(t *testing.T) {
			outStream := &bytes.Buffer{}
			errStream := &bytes.Buffer{}
			err := runProgramAction(
				pm,
				nil,                     // sketch
				"",                      // importFile
				test.importDir.String(), // importDir
				test.fqbn,               // FQBN
				test.port,               // port
				test.programmer,         // programmer
				false,                   // verbose
				false,                   // verify
				test.burnBootloader,     // burnBootloader
				outStream,
				errStream,
			)
			if test.expected == "FAIL" {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				out := strings.TrimSpace(outStream.String())
				require.Equal(t, strings.ReplaceAll(test.expected, "$$VERBOSE-VERIFY$$", "quiet noverify"), out)
			}
		})
		t.Run(fmt.Sprintf("SubTest%02d-WithVerifyAndVerbose", i), func(t *testing.T) {
			outStream := &bytes.Buffer{}
			errStream := &bytes.Buffer{}
			err := runProgramAction(
				pm,
				nil,                     // sketch
				"",                      // importFile
				test.importDir.String(), // importDir
				test.fqbn,               // FQBN
				test.port,               // port
				test.programmer,         // programmer
				true,                    // verbose
				true,                    // verify
				test.burnBootloader,     // burnBootloader
				outStream,
				errStream,
			)
			if test.expected == "FAIL" {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				out := strings.Split(outStream.String(), "\n")
				// With verbose enabled, the upload will output 3 lines:
				// - the command line that the cli is going to run
				// - the output of the command
				// - an empty line
				// we are interested in the second line
				require.Len(t, out, 3)
				require.Equal(t, strings.ReplaceAll(test.expected, "$$VERBOSE-VERIFY$$", "verbose verify"), out[1])
			}
		})
	}
}
