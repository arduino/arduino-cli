// This file is part of arduino-cli.
//
// Copyright 2022 ARDUINO SA (http://www.arduino.cc/)
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

package compile_test

import (
	"encoding/json"
	"testing"

	"github.com/arduino/arduino-cli/internal/integrationtest"
	"github.com/arduino/go-paths-helper"
	"github.com/stretchr/testify/require"
)

func TestCompileWithoutPrecompiledLibraries(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	// Init the environment explicitly
	url := "https://adafruit.github.io/arduino-board-index/package_adafruit_index.json"
	_, _, err := cli.Run("core", "update-index", "--additional-urls="+url)
	require.NoError(t, err)
	_, _, err = cli.Run("core", "install", "arduino:mbed@1.3.1", "--additional-urls="+url)
	require.NoError(t, err)

	// // Precompiled version of Arduino_TensorflowLite
	//	_, _, err = cli.Run("lib", "install", "Arduino_LSM9DS1")
	//	require.NoError(t, err)
	//	_, _, err = cli.Run("lib", "install", "Arduino_TensorflowLite@2.1.1-ALPHA-precompiled")
	//	require.NoError(t, err)

	//	sketchPath := cli.SketchbookDir().Join("libraries", "Arduino_TensorFlowLite", "examples", "hello_world")
	//	_, _, err = cli.Run("compile", "-b", "arduino:mbed:nano33ble", sketchPath.String())
	//	require.NoError(t, err)

	_, _, err = cli.Run("core", "install", "arduino:samd@1.8.7", "--additional-urls="+url)
	require.NoError(t, err)
	//	_, _, err = cli.Run("core", "install", "adafruit:samd@1.6.4", "--additional-urls="+url)
	//	require.NoError(t, err)
	//	// should work on adafruit too after https://github.com/arduino/arduino-cli/pull/1134
	//	_, _, err = cli.Run("compile", "-b", "adafruit:samd:adafruit_feather_m4", sketchPath.String())
	//	require.NoError(t, err)

	//	// Non-precompiled version of Arduino_TensorflowLite
	//	_, _, err = cli.Run("lib", "install", "Arduino_TensorflowLite@2.1.0-ALPHA")
	//	require.NoError(t, err)
	//	_, _, err = cli.Run("compile", "-b", "arduino:mbed:nano33ble", sketchPath.String())
	//	require.NoError(t, err)
	//	_, _, err = cli.Run("compile", "-b", "adafruit:samd:adafruit_feather_m4", sketchPath.String())
	//	require.NoError(t, err)

	// Bosch sensor library
	_, _, err = cli.Run("lib", "install", "BSEC Software Library@1.5.1474")
	require.NoError(t, err)
	sketchPath := cli.SketchbookDir().Join("libraries", "BSEC_Software_Library", "examples", "basic")
	_, _, err = cli.Run("compile", "-b", "arduino:samd:mkr1000", sketchPath.String())
	require.NoError(t, err)
	_, _, err = cli.Run("compile", "-b", "arduino:mbed:nano33ble", sketchPath.String())
	require.NoError(t, err)

	// USBBlaster library
	_, _, err = cli.Run("lib", "install", "USBBlaster@1.0.0")
	require.NoError(t, err)
	sketchPath = cli.SketchbookDir().Join("libraries", "USBBlaster", "examples", "USB_Blaster")
	_, _, err = cli.Run("compile", "-b", "arduino:samd:mkrvidor4000", sketchPath.String())
	require.NoError(t, err)
}

func TestCompileWithCustomLibraries(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	// Creates config with additional URL to install necessary core
	url := "http://arduino.esp8266.com/stable/package_esp8266com_index.json"
	_, _, err := cli.Run("config", "init", "--dest-dir", ".", "--additional-urls", url)
	require.NoError(t, err)

	// Init the environment explicitly
	_, _, err = cli.Run("update")
	require.NoError(t, err)

	_, _, err = cli.Run("core", "install", "esp8266:esp8266")
	require.NoError(t, err)

	sketchName := "sketch_with_multiple_custom_libraries"
	sketchPath := cli.CopySketch(sketchName)
	fqbn := "esp8266:esp8266:nodemcu:xtal=80,vt=heap,eesz=4M1M,wipe=none,baud=115200"

	firstLib := sketchPath.Join("libraries1")
	secondLib := sketchPath.Join("libraries2")
	_, _, err = cli.Run("compile", "--libraries", firstLib.String(), "--libraries", secondLib.String(), "-b", fqbn, sketchPath.String())
	require.NoError(t, err)
}

func TestCompileWithArchivesAndLongPaths(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	// Creates config with additional URL to install necessary core
	url := "http://arduino.esp8266.com/stable/package_esp8266com_index.json"
	_, _, err := cli.Run("config", "init", "--dest-dir", ".", "--additional-urls", url)
	require.NoError(t, err)

	// Init the environment explicitly
	_, _, err = cli.Run("update")
	require.NoError(t, err)

	// Install core to compile
	_, _, err = cli.Run("core", "install", "esp8266:esp8266@2.7.4")
	require.NoError(t, err)

	// Install test library
	_, _, err = cli.Run("lib", "install", "ArduinoIoTCloud")
	require.NoError(t, err)

	stdout, _, err := cli.Run("lib", "examples", "ArduinoIoTCloud", "--format", "json")
	require.NoError(t, err)
	var libOutput []map[string]interface{}
	err = json.Unmarshal(stdout, &libOutput)
	require.NoError(t, err)
	sketchPath := paths.New(libOutput[0]["library"].(map[string]interface{})["install_dir"].(string))
	sketchPath = sketchPath.Join("examples", "ArduinoIoTCloud-Advanced")

	_, _, err = cli.Run("compile", "-b", "esp8266:esp8266:huzzah", sketchPath.String())
	require.NoError(t, err)
}

func TestCompileWithPrecompileLibrary(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	_, _, err := cli.Run("update")
	require.NoError(t, err)

	_, _, err = cli.Run("core", "install", "arduino:samd@1.8.11")
	require.NoError(t, err)
	fqbn := "arduino:samd:mkrzero"

	// Install precompiled library
	// For more information see:
	// https://arduino.github.io/arduino-cli/latest/library-specification/#precompiled-binaries
	_, _, err = cli.Run("lib", "install", "BSEC Software Library@1.5.1474")
	require.NoError(t, err)
	sketchFolder := cli.SketchbookDir().Join("libraries", "BSEC_Software_Library", "examples", "basic")

	// Compile and verify dependencies detection for fully precompiled library is not skipped
	stdout, _, err := cli.Run("compile", "-b", fqbn, sketchFolder.String(), "-v")
	require.NoError(t, err)
	require.NotContains(t, string(stdout), "Skipping dependencies detection for precompiled library BSEC Software Library")
}
