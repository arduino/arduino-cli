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
	"strings"
	"testing"
	"slices"

	"github.com/arduino/arduino-cli/internal/integrationtest"
	"github.com/arduino/go-paths-helper"
	"github.com/stretchr/testify/require"
)

func TestCompileCoreCacheGeneration(t *testing.T) {
	// See:
	// https://forum.arduino.cc/t/teensy-compile-sketch-fails-clear-out-temp-build-completes/1110104/6
	// https://forum.pjrc.com/threads/72572-Teensyduino-1-59-Beta-2?p=324071&viewfull=1#post324071
	// https://github.com/arduino/arduino-ide/issues/1990

	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	// Run update-index with our test index
	_, _, err := cli.Run("core", "install", "arduino:avr@1.8.5")
	require.NoError(t, err)

	// Prepare sketch
	sketch, err := paths.New("testdata", "bare_minimum").Abs()
	require.NoError(t, err)

	// Perform two compile that should result in different cached cores
	_, _, err = cli.Run("compile", "-b", "arduino:avr:mega", sketch.String())
	require.NoError(t, err)
	_, _, err = cli.Run("compile", "-b", "arduino:avr:mega:cpu=atmega1280", sketch.String())
	require.NoError(t, err)

	// Perform the same compile again and track the cached cores
	extractCachedCoreFromStdout := func(stdout []byte) string {
		prefix := "Using precompiled core: "
		lines := strings.Split(string(stdout), "\n")
		i := slices.IndexFunc(lines, func(l string) bool {
			return strings.Contains(l, prefix)
		})
		require.NotEqual(t, -1, i, "Could not find precompiled core in output")
		return strings.TrimPrefix(lines[i], prefix)
	}
	stdout, _, err := cli.Run("compile", "-b", "arduino:avr:mega", sketch.String(), "-v")
	require.NoError(t, err)
	core1 := extractCachedCoreFromStdout(stdout)
	stdout, _, err = cli.Run("compile", "-b", "arduino:avr:mega:cpu=atmega1280", sketch.String(), "-v")
	require.NoError(t, err)
	core2 := extractCachedCoreFromStdout(stdout)
	require.NotEqual(t, core1, core2, "Precompile core must be different!")
}
