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

package debug_test

import (
	"testing"

	"github.com/arduino/arduino-cli/internal/integrationtest"
	"github.com/arduino/go-paths-helper"
	"github.com/stretchr/testify/require"
	"go.bug.st/testifyjson/requirejson"
)

func TestDebug(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	// Init the environment explicitly
	_, _, err := cli.Run("core", "update-index")
	require.NoError(t, err)

	// Install cores
	_, _, err = cli.Run("core", "install", "arduino:avr")
	require.NoError(t, err)
	_, _, err = cli.Run("core", "install", "arduino:samd")
	require.NoError(t, err)

	// Install custom core
	customHw, err := paths.New("testdata", "hardware").Abs()
	require.NoError(t, err)
	err = customHw.CopyDirTo(cli.SketchbookDir().Join("hardware"))
	require.NoError(t, err)

	integrationtest.CLISubtests{
		{"Start", testDebuggerStarts},
		{"WithPdeSketchStarts", testDebuggerWithPdeSketchStarts},
		{"DebugInformation", testAllDebugInformation},
		{"DebugCheck", testDebugCheck},
	}.Run(t, env, cli)
}

func testDebuggerStarts(t *testing.T, env *integrationtest.Environment, cli *integrationtest.ArduinoCLI) {
	// Create sketch for testing
	sketchName := "DebuggerStartTest"
	sketchPath := cli.DataDir().Join(sketchName)
	defer sketchPath.RemoveAll()
	fqbn := "arduino:samd:mkr1000"

	_, _, err := cli.Run("sketch", "new", sketchPath.String())
	require.NoError(t, err)

	// Build sketch
	_, _, err = cli.Run("compile", "-b", fqbn, sketchPath.String())
	require.NoError(t, err)

	programmer := "atmel_ice"
	// Starts debugger
	_, _, err = cli.Run("debug", "-b", fqbn, "-P", programmer, sketchPath.String(), "--info")
	require.NoError(t, err)
}

func testDebuggerWithPdeSketchStarts(t *testing.T, env *integrationtest.Environment, cli *integrationtest.ArduinoCLI) {
	sketchName := "DebuggerPdeSketchStartTest"
	sketchPath := cli.DataDir().Join(sketchName)
	defer sketchPath.RemoveAll()
	fqbn := "arduino:samd:mkr1000"

	_, _, err := cli.Run("sketch", "new", sketchPath.String())
	require.NoError(t, err)

	// Looks for sketch file .ino
	pathDir, err := sketchPath.ReadDir()
	require.NoError(t, err)
	fileIno := pathDir[0]

	// Renames sketch file to pde
	filePde := sketchPath.Join(sketchName + ".pde")
	err = fileIno.Rename(filePde)
	require.NoError(t, err)

	// Build sketch
	_, _, err = cli.Run("compile", "-b", fqbn, filePde.String())
	require.NoError(t, err)

	programmer := "atmel_ice"
	// Starts debugger
	_, _, err = cli.Run("debug", "-b", fqbn, "-P", programmer, filePde.String(), "--info")
	require.NoError(t, err)
}

func testAllDebugInformation(t *testing.T, env *integrationtest.Environment, cli *integrationtest.ArduinoCLI) {
	// Create sketch for testing
	sketchPath := cli.DataDir().Join("DebuggerStartTest")
	defer sketchPath.RemoveAll()
	_, _, err := cli.Run("sketch", "new", sketchPath.String())
	require.NoError(t, err)

	// Build sketch
	_, _, err = cli.Run("compile", "-b", "my:samd:my", sketchPath.String(), "--json")
	require.NoError(t, err)

	{
		// Starts debugger
		jsonDebugOut, _, err := cli.Run("debug", "-b", "my:samd:my", "-P", "atmel_ice", sketchPath.String(), "--info", "--json")
		require.NoError(t, err)
		debugOut := requirejson.Parse(t, jsonDebugOut)
		debugOut.MustContain(`
		{
			"toolchain": "gcc",
			"toolchain_path": "gcc-path",
			"toolchain_prefix": "gcc-prefix",
			"server": "openocd",
			"server_path": "openocd-path",
			"server_configuration": {
				"path": "openocd-path",
				"scripts_dir": "openocd-scripts-dir",
				"scripts": [
					"first",
					"second",
					"third",
					"fourth"
				]
			},
			"svd_file": "svd-file",
			"custom_configs": {
				"cortex-debug": {
					"aBoolean": true,
					"aStringBoolean": "true",
					"aStringNumber": "10",
					"aNumber": 10,
					"anotherNumber": 10.2,
					"anObject": {
						"boolean": true,
						"key": "value"
					},
					"anotherObject": {
						"boolean": true,
						"key": "value"
					},
					"anotherStringParamer": "hellooo",
					"overrideRestartCommands": [
						"monitor reset halt",
						"monitor gdb_sync",
						"thb setup",
						"c"
					],
					"postAttachCommands": [
						"set remote hardware-watchpoint-limit 2",
						"monitor reset halt",
						"monitor gdb_sync",
						"thb setup",
						"c"
					]
				},
			},
			"programmer": "atmel_ice"
		}`)
	}

	// Starts debugger with another programmer
	{
		jsonDebugOut, _, err := cli.Run("debug", "-b", "my:samd:my", "-P", "my_cold_ice", sketchPath.String(), "--info", "--json")
		require.NoError(t, err)
		debugOut := requirejson.Parse(t, jsonDebugOut)
		debugOut.MustContain(`
		{
			"toolchain": "gcc",
			"toolchain_path": "gcc-path",
			"toolchain_prefix": "gcc-prefix",
			"server": "openocd",
			"server_path": "openocd-path",
			"server_configuration": {
				"path": "openocd-path",
				"scripts_dir": "openocd-scripts-dir",
				"scripts": [
					"first",
					"second",
					"cold_ice_script",
					"fourth"
				]
			},
			"svd_file": "svd-file",
			"custom_configs": {
				"cortex-debug": {
					"aBoolean": true,
					"aStringBoolean": "true",
					"aStringNumber": "10",
					"aNumber": 10,
					"anotherNumber": 10.2,
					"anObject": {
						"boolean": true,
						"key": "value"
					},
					"anotherObject": {
						"boolean": true,
						"key": "value"
					},
					"anotherStringParamer": "hellooo",
					"overrideRestartCommands": [
						"monitor reset halt",
						"monitor gdb_sync",
						"thb setup",
						"c"
					],
					"postAttachCommands": [
						"set remote hardware-watchpoint-limit 2",
						"monitor reset halt",
						"monitor gdb_sync",
						"thb setup",
						"c"
					]
				},
			},
			"programmer": "my_cold_ice"
		}`)

		{
			// Starts debugger with an old-style openocd script definition
			jsonDebugOut, _, err := cli.Run("debug", "-b", "my:samd:my2", "-P", "atmel_ice", sketchPath.String(), "--info", "--json")
			require.NoError(t, err)
			debugOut := requirejson.Parse(t, jsonDebugOut)
			debugOut.MustContain(`
			{
				"toolchain": "gcc",
				"toolchain_path": "gcc-path",
				"toolchain_prefix": "gcc-prefix",
				"server": "openocd",
				"server_path": "openocd-path",
				"server_configuration": {
					"path": "openocd-path",
					"scripts_dir": "openocd-scripts-dir",
					"scripts": [
						"single-script"
					]
				},
				"svd_file": "svd-file",
				"programmer": "atmel_ice"
			}`)
		}

		{
			// Starts debugger with mixed old and new-style openocd script definition
			jsonDebugOut, _, err := cli.Run("debug", "-b", "my:samd:my2", "-P", "my_cold_ice", sketchPath.String(), "--info", "--json")
			require.NoError(t, err)
			debugOut := requirejson.Parse(t, jsonDebugOut)
			debugOut.MustContain(`
			{
				"toolchain": "gcc",
				"toolchain_path": "gcc-path",
				"toolchain_prefix": "gcc-prefix",
				"server": "openocd",
				"server_path": "openocd-path",
				"server_configuration": {
					"path": "openocd-path",
					"scripts_dir": "openocd-scripts-dir",
					"scripts": [
						"cold_ice_script"
					]
				},
				"svd_file": "svd-file",
				"programmer": "my_cold_ice"
			}`)
		}

		{
			// Mixing programmer and additional_config
			jsonDebugOut, _, err := cli.Run("debug", "-b", "my:samd:my3", "-P", "my_cold_ice", sketchPath.String(), "--info", "--json")
			require.NoError(t, err)
			debugOut := requirejson.Parse(t, jsonDebugOut)
			debugOut.MustContain(`
			{
				"toolchain": "gcc",
				"toolchain_path": "gcc-path",
				"toolchain_prefix": "gcc-prefix",
				"server": "openocd",
				"server_path": "openocd-path",
				"server_configuration": {
					"path": "openocd-path",
					"scripts_dir": "openocd-scripts-dir",
					"scripts": [
						"cold_ice_script"
					]
				},
				"custom_configs": {
					"cortex-debug": {
						"test1": "true"
					}
				},
				"svd_file": "test1.svd",
				"programmer": "my_cold_ice"
			}`)
		}

		{
			// Mixing programmer and additional_config selected by another variable
			jsonDebugOut, _, err := cli.Run("debug", "-b", "my:samd:my4", "-P", "my_cold_ice", sketchPath.String(), "--info", "--json")
			require.NoError(t, err)
			debugOut := requirejson.Parse(t, jsonDebugOut)
			debugOut.MustContain(`
			{
				"toolchain": "gcc",
				"toolchain_path": "gcc-path",
				"toolchain_prefix": "gcc-prefix",
				"server": "openocd",
				"server_path": "openocd-path",
				"server_configuration": {
					"path": "openocd-path",
					"scripts_dir": "openocd-scripts-dir",
					"scripts": [
						"cold_ice_script"
					]
				},
				"custom_configs": {
					"cortex-debug": {
						"test2": "true"
					}
				},
				"svd_file": "test2.svd",
				"programmer": "my_cold_ice"
			}`)
		}
	}
}

func testDebugCheck(t *testing.T, env *integrationtest.Environment, cli *integrationtest.ArduinoCLI) {
	out, _, err := cli.Run("debug", "check", "-b", "arduino:samd:mkr1000", "-P", "atmel_ice")
	require.NoError(t, err)
	require.Contains(t, string(out), "The given board/programmer configuration supports debugging.")

	out, _, err = cli.Run("debug", "check", "-b", "arduino:samd:mkr1000", "-P", "atmel_ice", "--json")
	require.NoError(t, err)
	requirejson.Query(t, out, `.debugging_supported`, `true`)

	out, _, err = cli.Run("debug", "check", "-b", "arduino:avr:uno", "-P", "atmel_ice")
	require.NoError(t, err)
	require.Contains(t, string(out), "The given board/programmer configuration does NOT support debugging.")

	out, _, err = cli.Run("debug", "check", "-b", "arduino:avr:uno", "-P", "atmel_ice", "--json")
	require.NoError(t, err)
	requirejson.Query(t, out, `.debugging_supported`, `false`)

	// Test minimum FQBN compute
	out, _, err = cli.Run("debug", "check", "-b", "my:samd:my5", "-P", "atmel_ice", "--json")
	require.NoError(t, err)
	requirejson.Contains(t, out, `{ "debugging_supported" : false }`)

	out, _, err = cli.Run("debug", "check", "-b", "my:samd:my5:dbg=on", "-P", "atmel_ice", "--json")
	require.NoError(t, err)
	requirejson.Contains(t, out, `{ "debugging_supported" : true, "debug_fqbn" : "my:samd:my5:dbg=on" }`)

	out, _, err = cli.Run("debug", "check", "-b", "my:samd:my5:dbg=on,cpu=150m", "-P", "atmel_ice", "--json")
	require.NoError(t, err)
	requirejson.Contains(t, out, `{ "debugging_supported" : true, "debug_fqbn" : "my:samd:my5:dbg=on" }`)

	out, _, err = cli.Run("debug", "check", "-b", "my:samd:my6", "-P", "atmel_ice", "--json")
	require.NoError(t, err)
	requirejson.Contains(t, out, `{ "debugging_supported" : true, "debug_fqbn" : "my:samd:my6" }`)

	out, _, err = cli.Run("debug", "check", "-b", "my:samd:my6:dbg=on", "-P", "atmel_ice", "--json")
	require.NoError(t, err)
	requirejson.Contains(t, out, `{ "debugging_supported" : true, "debug_fqbn" : "my:samd:my6" }`)
}

func TestDebugProfile(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	// Init the environment explicitly
	_, _, err := cli.Run("core", "update-index")
	require.NoError(t, err)

	sketchPath := cli.CopySketch("sketch_with_profile")

	// Compile the sketch using the profile
	_, _, err = cli.Run("compile", "--profile", "samd", sketchPath.String())
	require.NoError(t, err)

	// Debug using the profile
	_, _, err = cli.Run("debug", "--profile", "samd", sketchPath.String(), "--info")
	require.NoError(t, err)
}
