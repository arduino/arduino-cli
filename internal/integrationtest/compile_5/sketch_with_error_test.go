// This file is part of arduino-cli.
//
// Copyright 2026 ARDUINO SA (http://www.arduino.cc/)
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

func TestCompileSketchWithError(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	t.Cleanup(env.CleanUp)

	// Install Arduino Zephyr Boards
	_, _, err := cli.Run("core", "install", "arduino:zephyr@0.54.1")
	require.NoError(t, err)

	sketch, err := paths.New("testdata", "SketchWithErrorDirective").Abs()
	require.NoError(t, err)

	// Compile and check error
	_, outerr, err := cli.Run("compile", "-b", "arduino:zephyr:unoq", sketch.String(), "-v")
	require.Error(t, err, "compilation should fail due to error in sketch")
	require.Contains(t, string(outerr), "#error TEST")

	// Compile again, and check error
	_, outerr2, err := cli.Run("compile", "-b", "arduino:zephyr:unoq", sketch.String(), "-v")
	require.Error(t, err, "compilation should fail due to error in sketch")
	require.Contains(t, string(outerr2), "#error TEST")
}
