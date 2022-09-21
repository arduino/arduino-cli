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

func TestRuntimeToolPropertiesGeneration(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	// Run update-index with our test index
	_, _, err := cli.Run("core", "install", "arduino:avr@1.8.5")
	require.NoError(t, err)

	// Install test data into datadir
	testdata := paths.New("testdata", "platforms_with_conflicting_tools")
	hardwareDir := cli.DataDir().Join("packages")
	err = testdata.Join("alice").CopyDirTo(hardwareDir.Join("alice"))
	require.NoError(t, err)
	err = testdata.Join("bob").CopyDirTo(hardwareDir.Join("bob"))
	require.NoError(t, err)

	sketch, err := paths.New("testdata", "bare_minimum").Abs()
	require.NoError(t, err)

	// Multiple runs must always produce the same result
	for i := 0; i < 3; i++ {
		stdout, _, err := cli.Run("compile", "-b", "alice:avr:alice", "--show-properties", sketch.String())
		require.NoError(t, err)
		// the tools coming from the same packager are selected
		require.Contains(t, string(stdout), "runtime.tools.avr-gcc.path="+hardwareDir.String()+"/alice/tools/avr-gcc/50.0.0")
		require.Contains(t, string(stdout), "runtime.tools.avrdude.path="+hardwareDir.String()+"/alice/tools/avrdude/1.0.0")

		stdout, _, err = cli.Run("compile", "-b", "bob:avr:bob", "--show-properties", sketch.String())
		require.NoError(t, err)
		// the latest version available are selected
		require.Contains(t, string(stdout), "runtime.tools.avr-gcc.path="+hardwareDir.String()+"/alice/tools/avr-gcc/50.0.0")
		require.Contains(t, string(stdout), "runtime.tools.avrdude.path="+hardwareDir.String()+"/arduino/tools/avrdude/6.3.0-arduino17")

		stdout, _, err = cli.Run("compile", "-b", "arduino:avr:uno", "--show-properties", sketch.String())
		require.NoError(t, err)
		// the selected tools are listed as platform dependencies from the index.json
		require.Contains(t, string(stdout), "runtime.tools.avr-gcc.path="+hardwareDir.String()+"/arduino/tools/avr-gcc/7.3.0-atmel3.6.1-arduino7")
		require.Contains(t, string(stdout), "runtime.tools.avrdude.path="+hardwareDir.String()+"/arduino/tools/avrdude/6.3.0-arduino17")
	}
}
