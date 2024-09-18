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

package commands

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/arduino/arduino-cli/internal/arduino/cores"
	"github.com/arduino/arduino-cli/internal/arduino/cores/packagemanager"
	"github.com/arduino/arduino-cli/internal/arduino/sketch"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	paths "github.com/arduino/go-paths-helper"
	properties "github.com/arduino/go-properties-orderedmap"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"go.bug.st/downloader/v2"
)

func TestDetectSketchNameFromBuildPath(t *testing.T) {
	sk1, err1 := detectSketchNameFromBuildPath(paths.New("testdata/upload/build_path_1"))
	require.NoError(t, err1)
	require.Equal(t, "sketch.ino", sk1)

	sk2, err2 := detectSketchNameFromBuildPath(paths.New("testdata/upload/build_path_2"))
	require.NoError(t, err2)
	require.Equal(t, "Blink.ino", sk2)

	sk3, err3 := detectSketchNameFromBuildPath(paths.New("testdata/upload/build_path_3"))
	require.Error(t, err3)
	require.Equal(t, "", sk3)

	sk4, err4 := detectSketchNameFromBuildPath(paths.New("testdata/upload/build_path_4"))
	require.Error(t, err4)
	require.Equal(t, "", sk4)

	sk5, err5 := detectSketchNameFromBuildPath(paths.New("testdata/upload/build_path_invalid"))
	require.Error(t, err5)
	require.Equal(t, "", sk5)
}

func TestDetermineBuildPathAndSketchName(t *testing.T) {
	type test struct {
		importFile    string
		importDir     string
		sketch        *sketch.Sketch
		fqbn          *cores.FQBN
		resBuildPath  string
		resSketchName string
	}

	blonk, err := sketch.New(paths.New("testdata/upload/Blonk"))
	require.NoError(t, err)

	fqbn, err := cores.ParseFQBN("arduino:samd:mkr1000")
	require.NoError(t, err)

	srv := NewArduinoCoreServer().(*arduinoCoreServerImpl)

	tests := []test{
		// 00: error: no data passed in
		{"", "", nil, nil, "<nil>", ""},
		// 01: use importFile to detect build.path and project_name
		{"testdata/upload/build_path_2/Blink.ino.hex", "", nil, nil, "testdata/upload/build_path_2", "Blink.ino"},
		// 02: use importPath as build.path and project_name
		{"", "testdata/upload/build_path_2", nil, nil, "testdata/upload/build_path_2", "Blink.ino"},
		// 03: error: used both importPath and importFile
		{"testdata/upload/build_path_2/Blink.ino.hex", "testdata/upload/build_path_2", nil, nil, "<nil>", ""},
		// 04: only sketch without FQBN
		{"", "", blonk, nil, srv.getDefaultSketchBuildPath(blonk, nil).String(), "Blonk.ino"},
		// 05: use importFile to detect build.path and project_name, sketch is ignored.
		{"testdata/upload/build_path_2/Blink.ino.hex", "", blonk, nil, "testdata/upload/build_path_2", "Blink.ino"},
		// 06: use importPath as build.path and Blink as project name, ignore the sketch Blonk
		{"", "testdata/upload/build_path_2", blonk, nil, "testdata/upload/build_path_2", "Blink.ino"},
		// 07: error: used both importPath and importFile
		{"testdata/upload/build_path_2/Blink.ino.hex", "testdata/upload/build_path_2", blonk, nil, "<nil>", ""},
		// 08: error: no data passed in
		{"", "", nil, fqbn, "<nil>", ""},
		// 09: use importFile to detect build.path and project_name, fqbn ignored
		{"testdata/upload/build_path_2/Blink.ino.hex", "", nil, fqbn, "testdata/upload/build_path_2", "Blink.ino"},
		// 10: use importPath as build.path and project_name, fqbn ignored
		{"", "testdata/upload/build_path_2", nil, fqbn, "testdata/upload/build_path_2", "Blink.ino"},
		// 11: error: used both importPath and importFile
		{"testdata/upload/build_path_2/Blink.ino.hex", "testdata/upload/build_path_2", nil, fqbn, "<nil>", ""},
		// 12: use sketch to determine project name and sketch+fqbn to determine build path
		{"", "", blonk, fqbn, srv.getDefaultSketchBuildPath(blonk, nil).String(), "Blonk.ino"},
		// 13: use importFile to detect build.path and project_name, sketch+fqbn is ignored.
		{"testdata/upload/build_path_2/Blink.ino.hex", "", blonk, fqbn, "testdata/upload/build_path_2", "Blink.ino"},
		// 14: use importPath as build.path and Blink as project name, ignore the sketch Blonk, ignore fqbn
		{"", "testdata/upload/build_path_2", blonk, fqbn, "testdata/upload/build_path_2", "Blink.ino"},
		// 15: error: used both importPath and importFile
		{"testdata/upload/build_path_2/Blink.ino.hex", "testdata/upload/build_path_2", blonk, fqbn, "<nil>", ""},
		// 16: importPath containing multiple firmwares, but one has the same name as the containing folder
		{"", "testdata/upload/firmware", nil, fqbn, "testdata/upload/firmware", "firmware.ino"},
		// 17: importFile among multiple firmwares
		{"testdata/upload/firmware/another_firmware.ino.bin", "", nil, fqbn, "testdata/upload/firmware", "another_firmware.ino"},
	}
	for i, test := range tests {
		t.Run(fmt.Sprintf("SubTest%02d", i), func(t *testing.T) {
			buildPath, sketchName, err := srv.determineBuildPathAndSketchName(test.importFile, test.importDir, test.sketch)
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

func TestUploadPropertiesComposition(t *testing.T) {
	pmb := packagemanager.NewBuilder(nil, nil, nil, nil, nil, "test", downloader.GetDefaultConfig())
	errs := pmb.LoadHardwareFromDirectory(paths.New("testdata", "upload", "hardware"))
	require.Len(t, errs, 0)
	buildPath1 := paths.New("testdata", "upload", "build_path_1")
	logrus.SetLevel(logrus.TraceLevel)
	type test struct {
		importDir       *paths.Path
		fqbn            string
		port            string
		protocol        string
		programmer      string
		burnBootloader  bool
		expectedOutput  string
		expectedOutput2 string
	}

	cwdPath, err := paths.Getwd()
	require.NoError(t, err)
	cwd := strings.ReplaceAll(cwdPath.String(), "\\", "/")

	tests := []test{
		// 0: classic upload, requires port
		{buildPath1, "alice:avr:board1", "port", "serial", "", false, "conf-board1 conf-general conf-upload $$VERBOSE-VERIFY$$ protocol port -bspeed testdata/upload/build_path_1/sketch.ino.hex\n", ""},
		{buildPath1, "alice:avr:board1", "", "", "", false, "FAIL", ""},
		// 2: classic upload, no port
		{buildPath1, "alice:avr:board2", "port", "serial", "", false, "conf-board1 conf-general conf-upload $$VERBOSE-VERIFY$$ protocol -bspeed testdata/upload/build_path_1/sketch.ino.hex\n", ""},
		{buildPath1, "alice:avr:board2", "", "", "", false, "conf-board1 conf-general conf-upload $$VERBOSE-VERIFY$$ protocol -bspeed testdata/upload/build_path_1/sketch.ino.hex\n", ""},

		// 4: upload with programmer, requires port
		{buildPath1, "alice:avr:board1", "port", "serial", "progr1", false, "conf-board1 conf-general conf-program $$VERBOSE-VERIFY$$ progprotocol port -bspeed testdata/upload/build_path_1/sketch.ino.hex\n", ""},
		{buildPath1, "alice:avr:board1", "", "", "progr1", false, "FAIL", ""},
		// 6: upload with programmer, no port
		{buildPath1, "alice:avr:board1", "port", "serial", "progr2", false, "conf-board1 conf-general conf-program $$VERBOSE-VERIFY$$ prog2protocol -bspeed testdata/upload/build_path_1/sketch.ino.hex\n", ""},
		{buildPath1, "alice:avr:board1", "", "", "progr2", false, "conf-board1 conf-general conf-program $$VERBOSE-VERIFY$$ prog2protocol -bspeed testdata/upload/build_path_1/sketch.ino.hex\n", ""},
		// 8: upload with programmer, require port through extra params
		{buildPath1, "alice:avr:board1", "port", "serial", "progr3", false, "conf-board1 conf-general conf-program $$VERBOSE-VERIFY$$ prog3protocol port -bspeed testdata/upload/build_path_1/sketch.ino.hex\n", ""},
		{buildPath1, "alice:avr:board1", "", "", "progr3", false, "FAIL", ""},

		// 10: burn bootloader, require port
		{buildPath1, "alice:avr:board1", "port", "serial", "", true, "FAIL", ""}, // requires programmer
		{buildPath1, "alice:avr:board1", "port", "serial", "progr1", true,
			"ERASE conf-board1 conf-general conf-erase $$VERBOSE-VERIFY$$ genprog1protocol port -bspeed\n",
			"BURN conf-board1 conf-general conf-bootloader $$VERBOSE-VERIFY$$ genprog1protocol port -bspeed -F0xFF " + cwd + "/testdata/upload/hardware/alice/avr/bootloaders/niceboot/niceboot.hex\n"},

		// 12: burn bootloader, preferences override from programmers.txt
		{buildPath1, "alice:avr:board1", "port", "serial", "progr4", true,
			"ERASE conf-board1 conf-two-general conf-two-erase $$VERBOSE-VERIFY$$ prog4protocol-bootloader port -bspeed\n",
			"BURN conf-board1 conf-two-general conf-two-bootloader $$VERBOSE-VERIFY$$ prog4protocol-bootloader port -bspeed -F0xFF " + cwd + "/testdata/upload/hardware/alice/avr/bootloaders/niceboot/niceboot.hex\n"},
	}

	pm := pmb.Build()
	pme, release := pm.NewExplorer()
	defer release()

	srv := NewArduinoCoreServer().(*arduinoCoreServerImpl)
	testRunner := func(t *testing.T, test test, verboseVerify bool) {
		outStream := &bytes.Buffer{}
		errStream := &bytes.Buffer{}
		_, err := srv.runProgramAction(
			context.Background(),
			pme,
			nil,                     // sketch
			"",                      // importFile
			test.importDir.String(), // importDir
			test.fqbn,               // FQBN
			&rpc.Port{Address: test.port, Protocol: test.protocol},
			test.programmer,     // programmer
			verboseVerify,       // verbose
			verboseVerify,       // verify
			test.burnBootloader, // burnBootloader
			outStream,
			errStream,
			false,
			map[string]string{},
			nil,
		)
		verboseVerifyOutput := "verbose verify"
		if !verboseVerify {
			verboseVerifyOutput = "quiet noverify"
		}
		if test.expectedOutput == "FAIL" {
			require.NotNil(t, err)
		} else {
			require.Nil(t, err)
			outFiltered := strings.ReplaceAll(outStream.String(), "\r", "")
			outFiltered = strings.ReplaceAll(outFiltered, "\\", "/")
			require.Contains(t, outFiltered, strings.ReplaceAll(test.expectedOutput, "$$VERBOSE-VERIFY$$", verboseVerifyOutput))
			require.Contains(t, outFiltered, strings.ReplaceAll(test.expectedOutput2, "$$VERBOSE-VERIFY$$", verboseVerifyOutput))
		}
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("SubTest%02d", i), func(t *testing.T) {
			testRunner(t, test, false)
		})
		t.Run(fmt.Sprintf("SubTest%02d-WithVerifyAndVerbose", i), func(t *testing.T) {
			testRunner(t, test, true)
		})
	}
}

func TestGetToolId(t *testing.T) {
	props, err := properties.LoadFromBytes([]byte(`
bootloader.tool=avrdude
bootloader.tool.serial=avrdude
upload.tool=bossac
upload.tool.serial=bossac
upload.tool.network=arduino_ota`))
	require.NoError(t, err)

	toolID, err := getToolID(props, "upload", "serial")
	require.NoError(t, err)
	require.Equal(t, "bossac", toolID)

	toolID, err = getToolID(props, "upload", "network")
	require.NoError(t, err)
	require.Equal(t, "arduino_ota", toolID)

	toolID, err = getToolID(props, "bootloader", "serial")
	require.NoError(t, err)
	require.Equal(t, "avrdude", toolID)

	toolID, err = getToolID(props, "bootloader", "network")
	require.EqualError(t, err, "Property 'bootloader.tool.network' is undefined")
	require.Equal(t, "", toolID)

	props, err = properties.LoadFromBytes([]byte(`
	bootloader.tool.default=avrdude
	upload.tool.default=bossac`))
	require.NoError(t, err)

	toolID, err = getToolID(props, "upload", "serial")
	require.NoError(t, err)
	require.Equal(t, "bossac", toolID)

	toolID, err = getToolID(props, "upload", "network")
	require.NoError(t, err)
	require.Equal(t, "bossac", toolID)

	toolID, err = getToolID(props, "bootloader", "serial")
	require.NoError(t, err)
	require.Equal(t, "avrdude", toolID)

	toolID, err = getToolID(props, "bootloader", "network")
	require.NoError(t, err)
	require.Equal(t, "avrdude", toolID)
}

func TestGetUserFields(t *testing.T) {
	platformRelease := &cores.PlatformRelease{}

	props, err := properties.LoadFromBytes([]byte(`
tools.avrdude.upload.field.username=Username
tools.avrdude.upload.field.password=Password
tools.avrdude.upload.field.password.secret=true
tools.arduino_ota.upload.field.username=Username
tools.arduino_ota.upload.field.password=Password
tools.arduino_ota.upload.field.password.secret=true`))
	require.NoError(t, err)

	platformRelease.Properties = props

	userFields := getUserFields("avrdude", platformRelease)
	require.Len(t, userFields, 2)
	require.Equal(t, userFields[0].GetToolId(), "avrdude")
	require.Equal(t, userFields[0].GetName(), "username")
	require.Equal(t, userFields[0].GetLabel(), "Username")
	require.False(t, userFields[0].GetSecret())
	require.Equal(t, userFields[1].GetToolId(), "avrdude")
	require.Equal(t, userFields[1].GetName(), "password")
	require.Equal(t, userFields[1].GetLabel(), "Password")
	require.True(t, userFields[1].GetSecret())

	props, err = properties.LoadFromBytes([]byte(`
tools.arduino_ota.upload.field.password=Password
tools.arduino_ota.upload.field.password.secret=THIS_IS_NOT_A_BOOLEAN`))
	require.NoError(t, err)
	platformRelease.Properties = props

	props, err = properties.LoadFromBytes([]byte(`
tools.arduino_ota.upload.field.some_field=This is a really long label that ideally must never be set by any platform
`))
	require.NoError(t, err)
	platformRelease.Properties = props
	userFields = getUserFields("arduino_ota", platformRelease)
	require.Len(t, userFields, 1)
	require.Equal(t, userFields[0].GetToolId(), "arduino_ota")
	require.Equal(t, userFields[0].GetName(), "some_field")
	require.Equal(t, userFields[0].GetLabel(), "This is a really long label that ideally must nev…")
	require.False(t, userFields[0].GetSecret())
}

func TestOverrideProtocolProperties(t *testing.T) {
	props, err := properties.LoadFromBytes([]byte(`
	upload.speed=256
	upload.serial.speed=57600
	upload.network.speed=19200
	upload.unrelated_property=ok`))
	require.NoError(t, err)

	res := overrideProtocolProperties("upload", "serial", props)
	require.Equal(t, res.Get("upload.speed"), "57600")
	require.Equal(t, res.Get("upload.serial.speed"), "57600")
	require.Equal(t, res.Get("upload.network.speed"), "19200")
	require.Equal(t, res.Get("upload.unrelated_property"), "ok")

	res = overrideProtocolProperties("upload", "network", props)
	require.Equal(t, res.Get("upload.speed"), "19200")
	require.Equal(t, res.Get("upload.serial.speed"), "57600")
	require.Equal(t, res.Get("upload.network.speed"), "19200")
	require.Equal(t, res.Get("upload.unrelated_property"), "ok")

	res = overrideProtocolProperties("upload", "some_other_protocol", props)
	require.Equal(t, res.Get("upload.speed"), "256")
	require.Equal(t, res.Get("upload.serial.speed"), "57600")
	require.Equal(t, res.Get("upload.network.speed"), "19200")
	require.Equal(t, res.Get("upload.unrelated_property"), "ok")

	res = overrideProtocolProperties("bootloader", "serial", props)
	require.Equal(t, res.Get("upload.speed"), "256")
	require.Equal(t, res.Get("upload.serial.speed"), "57600")
	require.Equal(t, res.Get("upload.network.speed"), "19200")
	require.Equal(t, res.Get("upload.unrelated_property"), "ok")

}
