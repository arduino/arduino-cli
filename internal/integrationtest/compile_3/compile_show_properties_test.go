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
	"testing"

	"github.com/arduino/arduino-cli/internal/integrationtest"
	"github.com/arduino/go-properties-orderedmap"
	"github.com/stretchr/testify/require"
	"go.bug.st/testifyjson/requirejson"
)

func TestCompileShowProperties(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	_, _, err := cli.Run("core", "update-index")
	require.NoError(t, err)
	_, _, err = cli.Run("core", "install", "arduino:avr")
	require.NoError(t, err)

	bareMinimum := cli.CopySketch("bare_minimum")

	// Test --show-properties output is clean
	stdout, stderr, err := cli.Run("compile", "--fqbn", "arduino:avr:uno", "-v", "--show-properties", bareMinimum.String())
	require.NoError(t, err)
	_, err = properties.LoadFromBytes(stdout)
	require.NoError(t, err, "Output must be a clean property list")
	require.Empty(t, stderr)

	// Test --show-properties --format JSON output is clean
	stdout, stderr, err = cli.Run("compile", "--fqbn", "arduino:avr:uno", "-v", "--show-properties", "--format", "json", bareMinimum.String())
	require.NoError(t, err)
	requirejson.Parse(t, stdout, "Output must be a valid JSON")
	require.Empty(t, stderr)

	// Test --show-properties output is clean, with a wrong FQBN
	stdout, stderr, err = cli.Run("compile", "--fqbn", "arduino:avr:unoa", "-v", "--show-properties", bareMinimum.String())
	require.Error(t, err)
	_, err = properties.LoadFromBytes(stdout)
	require.NoError(t, err, "Output must be a clean property list")
	require.NotEmpty(t, stderr)

	// Test --show-properties --format JSON output is clean, with a wrong FQBN
	stdout, stderr, err = cli.Run("compile", "--fqbn", "arduino:avr:unoa", "-v", "--show-properties", "--format", "json", bareMinimum.String())
	require.Error(t, err)
	requirejson.Parse(t, stdout, "Output must be a valid JSON")
	require.Empty(t, stderr)
}
