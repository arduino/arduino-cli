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

func TestCompileSketchWithCxxOrCc(t *testing.T) {
	// See: https://github.com/arduino/arduino-cli/issues/1149

	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	// Run update-index with our test index
	_, _, err := cli.Run("core", "install", "arduino:avr@1.8.5")
	require.NoError(t, err)

	// Prepare sketchbook and sketch
	sketch, err := paths.New("testdata", "sketch_with_cxx_cc").Abs()
	require.NoError(t, err)

	_, _, err = cli.Run("compile", "-v", "-b", "arduino:avr:uno", sketch.String())
	require.NoError(t, err)
}
