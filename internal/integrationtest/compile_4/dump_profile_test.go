// This file is part of arduino-cli.
//
// Copyright 2023 ARDUINO SA (http://www.arduino.cc/)
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
	"strings"
	"testing"

	"github.com/arduino/arduino-cli/internal/integrationtest"
	"github.com/arduino/go-paths-helper"
	"github.com/stretchr/testify/require"
)

func TestDumpProfileClean(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	t.Cleanup(env.CleanUp)

	// Install Arduino AVR Boards
	_, _, err := cli.Run("core", "install", "arduino:avr@1.8.6")
	require.NoError(t, err)

	validSketchPath, err := paths.New("testdata", "ValidSketch").Abs()
	require.NoError(t, err)
	invalidSketchPath, err := paths.New("testdata", "InvalidSketch").Abs()
	require.NoError(t, err)

	validProfile := `profiles:
  uno:
    fqbn: arduino:avr:uno
    platforms:
      - platform: arduino:avr (1.8.6)`
	t.Run("NoVerbose", func(t *testing.T) {
		stdout, stderr, err := cli.Run("compile", "-b", "arduino:avr:uno", validSketchPath.String(), "--dump-profile")
		require.NoError(t, err)
		require.Empty(t, stderr)
		profile := strings.TrimSpace(string(stdout))
		require.Equal(t, validProfile, profile)
	})
	t.Run("Verbose", func(t *testing.T) {
		stdout, stderr, err := cli.Run("compile", "-b", "arduino:avr:uno", validSketchPath.String(), "--dump-profile", "--verbose")
		require.NoError(t, err)
		require.Empty(t, stderr)
		profile := strings.TrimSpace(string(stdout))
		require.Equal(t, validProfile, profile)
	})
	t.Run("ErrorNoVerbose", func(t *testing.T) {
		stdout, stderr, err := cli.Run("compile", "-b", "arduino:avr:uno", invalidSketchPath.String(), "--dump-profile")
		require.Error(t, err)
		require.NotEmpty(t, stderr)
		require.NotContains(t, string(stdout), validProfile)
	})
	t.Run("ErrorVerbose", func(t *testing.T) {
		stdout, stderr, err := cli.Run("compile", "-b", "arduino:avr:uno", invalidSketchPath.String(), "--dump-profile", "--verbose")
		require.Error(t, err)
		require.NotEmpty(t, stderr)
		require.NotContains(t, string(stdout), validProfile)
	})
}
