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
	"go.bug.st/testifyjson/requirejson"
)

func TestCompileCommandsJSONGeneration(t *testing.T) {
	// See: https://github.com/arduino/arduino-cli/issues/2401

	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	// Run update-index with our test index
	_, _, err := cli.Run("core", "install", "arduino:avr@1.8.5")
	require.NoError(t, err)

	// Create a test sketch
	out, _, err := cli.Run("sketch", "new", "Test", "--format", "json")
	require.NoError(t, err)
	var s struct {
		Path string `json:"sketch_path"`
	}
	require.NoError(t, json.Unmarshal(out, &s))
	sketchPath := paths.New(s.Path)
	buildPath := sketchPath.Join("build")

	{
		// Normal build
		_, _, err = cli.Run(
			"compile",
			"-b", "arduino:avr:uno",
			"--build-path", buildPath.String(),
			sketchPath.String())
		require.NoError(t, err)

		compileCommandsPath := buildPath.Join("compile_commands.json")
		require.True(t, compileCommandsPath.Exist())
		compileCommandJson, err := compileCommandsPath.ReadFile()
		require.NoError(t, err)
		compileCommands := requirejson.Parse(t, compileCommandJson)
		// Check that the variant include path is present, one of the arguments must be
		// something like:
		// "-I/home/user/.arduino15/packages/arduino/hardware/avr/1.8.6/variants/standard"
		compileCommands.Query(`[ .[0].arguments[] | contains("standard") ] | any`).MustEqual(`true`)
	}

	{
		// Build with skip-library-check
		_, _, err = cli.Run(
			"compile",
			"-b", "arduino:avr:uno",
			"--only-compilation-database",
			"--skip-libraries-discovery",
			"--build-path", buildPath.String(),
			sketchPath.String())
		require.NoError(t, err)

		compileCommandsPath := buildPath.Join("compile_commands.json")
		require.True(t, compileCommandsPath.Exist())
		compileCommandJson, err := compileCommandsPath.ReadFile()
		require.NoError(t, err)
		compileCommands := requirejson.Parse(t, compileCommandJson)
		// Check that the variant include path is present, one of the arguments must be
		// something like:
		// "-I/home/user/.arduino15/packages/arduino/hardware/avr/1.8.6/variants/standard"
		compileCommands.Query(`[ .[0].arguments[] | contains("standard") ] | any`).MustEqual(`true`)
	}
}
