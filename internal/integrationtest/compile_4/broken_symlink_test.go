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
	"testing"

	"github.com/arduino/arduino-cli/internal/integrationtest"
	"github.com/arduino/go-paths-helper"
	"github.com/stretchr/testify/require"
)

func TestCompileWithBrokenSymLinks(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	t.Cleanup(env.CleanUp)

	// Install Arduino AVR Boards
	_, _, err := cli.Run("core", "install", "arduino:avr@1.8.6")
	require.NoError(t, err)

	t.Run("NonSketchFileBroken", func(t *testing.T) {
		sketch, err := paths.New("testdata", "ValidSketchWithBrokenSymlink").Abs()
		require.NoError(t, err)
		_, _, err = cli.Run("compile", "-b", "arduino:avr:uno", sketch.String())
		require.NoError(t, err)
	})

	t.Run("SketchFileBroken", func(t *testing.T) {
		sketch, err := paths.New("testdata", "ValidSketchWithBrokenSketchFileSymlink").Abs()
		require.NoError(t, err)
		_, _, err = cli.Run("compile", "-b", "arduino:avr:uno", sketch.String())
		require.Error(t, err)
	})

	t.Run("NonInoSketchFileBroken", func(t *testing.T) {
		sketch, err := paths.New("testdata", "ValidSketchWithNonInoBrokenSketchFileSymlink").Abs()
		require.NoError(t, err)
		_, _, err = cli.Run("compile", "-b", "arduino:avr:uno", sketch.String())
		require.Error(t, err)
	})
}
