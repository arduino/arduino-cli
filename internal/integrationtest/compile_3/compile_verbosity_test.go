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
	"github.com/arduino/go-paths-helper"
	"github.com/stretchr/testify/require"
)

func TestCompileVerbosity(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	_, _, err := cli.Run("core", "update-index")
	require.NoError(t, err)
	_, _, err = cli.Run("core", "install", "arduino:avr")
	require.NoError(t, err)

	goodSketch, err := paths.New("testdata", "bare_minimum").Abs()
	require.NoError(t, err)
	badSketch, err := paths.New("testdata", "blink_with_error_directive").Abs()
	require.NoError(t, err)

	hasSketchSize := func(t *testing.T, out []byte) {
		require.Contains(t, string(out), "Sketch uses")
	}
	noSketchSize := func(t *testing.T, out []byte) {
		require.NotContains(t, string(out), "Sketch uses")
	}
	hasRecapTable := func(t *testing.T, out []byte) {
		require.Contains(t, string(out), "Used platform")
	}
	noRecapTable := func(t *testing.T, out []byte) {
		require.NotContains(t, string(out), "Used platform")
	}

	t.Run("DefaultVerbosity/SuccessfulBuild", func(t *testing.T) {
		stdout, stderr, err := cli.Run("compile", "--fqbn", "arduino:avr:uno", goodSketch.String())
		require.NoError(t, err)
		hasSketchSize(t, stdout)
		noRecapTable(t, stdout)
		require.Empty(t, stderr)
	})

	t.Run("DefaultVerbosity/BuildWithErrors", func(t *testing.T) {
		stdout, stderr, err := cli.Run("compile", "--fqbn", "arduino:avr:uno", badSketch.String())
		require.Error(t, err)
		hasRecapTable(t, stdout)
		require.NotEmpty(t, stderr)
	})

	t.Run("VerboseVerbosity/SuccessfulBuild", func(t *testing.T) {
		stdout, stderr, err := cli.Run("compile", "--fqbn", "arduino:avr:uno", "-v", goodSketch.String())
		require.NoError(t, err)
		hasSketchSize(t, stdout)
		hasRecapTable(t, stdout)
		require.Empty(t, stderr)
	})

	t.Run("VerboseVerbosity/BuildWithErrors", func(t *testing.T) {
		stdout, stderr, err := cli.Run("compile", "--fqbn", "arduino:avr:uno", "-v", badSketch.String())
		require.Error(t, err)
		hasRecapTable(t, stdout)
		require.NotEmpty(t, stderr)
	})

	t.Run("QuietVerbosity/SuccessfulBuild", func(t *testing.T) {
		stdout, stderr, err := cli.Run("compile", "--fqbn", "arduino:avr:uno", "-q", goodSketch.String())
		require.NoError(t, err)
		noSketchSize(t, stdout)
		noRecapTable(t, stdout)
		require.Empty(t, stdout) // Empty output
		require.Empty(t, stderr)
	})

	t.Run("QuietVerbosity/BuildWithErrors", func(t *testing.T) {
		stdout, stderr, err := cli.Run("compile", "--fqbn", "arduino:avr:uno", "-q", badSketch.String())
		require.Error(t, err)
		noRecapTable(t, stdout)
		require.NotEmpty(t, stderr)
	})

	t.Run("ConflictingVerbosityOptions", func(t *testing.T) {
		_, _, err := cli.Run("compile", "--fqbn", "arduino:avr:uno", "-v", "-q", goodSketch.String())
		require.Error(t, err)
	})
}
