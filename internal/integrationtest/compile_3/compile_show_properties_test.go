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
	"github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	"github.com/arduino/go-properties-orderedmap"
	"github.com/stretchr/testify/require"
)

type cliCompileResponse struct {
	BuilderResult *commands.CompileResponse `json:"builder_result"`
}

func TestCompileShowProperties(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	_, _, err := cli.Run("core", "update-index")
	require.NoError(t, err)
	_, _, err = cli.Run("core", "install", "arduino:avr")
	require.NoError(t, err)

	bareMinimum := cli.CopySketch("bare_minimum")

	// Test --show-properties output is clean
	// properties are not expanded
	stdout, stderr, err := cli.Run("compile", "--fqbn", "arduino:avr:uno", "-v", "--show-properties", bareMinimum.String())
	require.NoError(t, err)
	props, err := properties.LoadFromBytes(stdout)
	require.NoError(t, err, "Output must be a clean property list")
	require.Empty(t, stderr)
	require.True(t, props.ContainsKey("archive_file_path"))
	require.Contains(t, props.Get("archive_file_path"), "{build.path}")

	// Test --show-properties --format JSON output is clean
	// properties are not expanded
	stdout, stderr, err = cli.Run("compile", "--fqbn", "arduino:avr:uno", "-v", "--show-properties", "--format", "json", bareMinimum.String())
	require.NoError(t, err)
	require.Empty(t, stderr)
	props, err = properties.LoadFromSlice(
		requireCompileResponseJson(t, stdout).BuilderResult.GetBuildProperties())
	require.NoError(t, err)
	require.True(t, props.ContainsKey("archive_file_path"))
	require.Contains(t, props.Get("archive_file_path"), "{build.path}")

	// Test --show-properties output is clean, with a wrong FQBN
	stdout, stderr, err = cli.Run("compile", "--fqbn", "arduino:avr:unoa", "-v", "--show-properties", bareMinimum.String())
	require.Error(t, err)
	_, err = properties.LoadFromBytes(stdout)
	require.NoError(t, err, "Output must be a clean property list")
	require.NotEmpty(t, stderr)

	// Test --show-properties --format JSON output is clean, with a wrong FQBN
	stdout, stderr, err = cli.Run("compile", "--fqbn", "arduino:avr:unoa", "-v", "--show-properties", "--format", "json", bareMinimum.String())
	require.Error(t, err)
	require.Empty(t, stderr)
	requireCompileResponseJson(t, stdout)

	// Test --show-properties=unexpanded output is clean
	// properties are not expanded
	stdout, stderr, err = cli.Run("compile", "--fqbn", "arduino:avr:uno", "-v", "--show-properties=unexpanded", bareMinimum.String())
	require.NoError(t, err)
	props, err = properties.LoadFromBytes(stdout)
	require.NoError(t, err, "Output must be a clean property list")
	require.Empty(t, stderr)
	require.True(t, props.ContainsKey("archive_file_path"))
	require.Contains(t, props.Get("archive_file_path"), "{build.path}")

	// Test --show-properties=unexpanded output is clean
	// properties are not expanded
	stdout, stderr, err = cli.Run("compile", "--fqbn", "arduino:avr:uno", "-v", "--show-properties=unexpanded", "--format", "json", bareMinimum.String())
	require.NoError(t, err)
	require.Empty(t, stderr)
	props, err = properties.LoadFromSlice(
		requireCompileResponseJson(t, stdout).BuilderResult.GetBuildProperties())
	require.NoError(t, err)
	require.True(t, props.ContainsKey("archive_file_path"))
	require.Contains(t, props.Get("archive_file_path"), "{build.path}")

	// Test --show-properties=expanded output is clean
	// properties are expanded
	stdout, stderr, err = cli.Run("compile", "--fqbn", "arduino:avr:uno", "-v", "--show-properties=expanded", bareMinimum.String())
	require.NoError(t, err)
	props, err = properties.LoadFromBytes(stdout)
	require.NoError(t, err, "Output must be a clean property list")
	require.Empty(t, stderr)
	require.True(t, props.ContainsKey("archive_file_path"))
	require.NotContains(t, props.Get("archive_file_path"), "{build.path}")

	// Test --show-properties=expanded --format JSON output is clean
	// properties are expanded
	stdout, stderr, err = cli.Run("compile", "--fqbn", "arduino:avr:uno", "-v", "--show-properties=expanded", "--format", "json", bareMinimum.String())
	require.NoError(t, err)
	require.Empty(t, stderr)
	props, err = properties.LoadFromSlice(
		requireCompileResponseJson(t, stdout).BuilderResult.GetBuildProperties())
	require.NoError(t, err)
	require.True(t, props.ContainsKey("archive_file_path"))
	require.NotContains(t, props.Get("archive_file_path"), "{build.path}")
}

func requireCompileResponseJson(t *testing.T, stdout []byte) *cliCompileResponse {
	var compileResponse cliCompileResponse
	require.NoError(t, json.Unmarshal(stdout, &compileResponse))
	return &compileResponse
}
