// This file is part of arduino-cli.
//
// Copyright 2024 ARDUINO SA (http://www.arduino.cc/)
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

package compile

import (
	"testing"

	"github.com/arduino/arduino-cli/internal/integrationtest"
	"github.com/arduino/go-paths-helper"
	"github.com/stretchr/testify/require"
)

func TestBuildCacheLibWithNonASCIIChars(t *testing.T) {
	// See: https://github.com/arduino/arduino-cli/issues/2671

	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	t.Cleanup(env.CleanUp)

	tmpUserDir, err := paths.MkTempDir("", "Håkan")
	require.NoError(t, err)
	t.Cleanup(func() { tmpUserDir.RemoveAll() })
	customEnv := cli.GetDefaultEnv()
	customEnv["ARDUINO_DIRECTORIES_USER"] = tmpUserDir.String()

	// Install Arduino AVR Boards and Servo lib
	_, _, err = cli.Run("core", "install", "arduino:avr@1.8.6")
	require.NoError(t, err)
	_, _, err = cli.RunWithCustomEnv(customEnv, "lib", "install", "Servo")
	require.NoError(t, err)

	// Make a temp sketch
	sketchDir := tmpUserDir.Join("ServoSketch")
	sketchFile := sketchDir.Join("ServoSketch.ino")
	require.NoError(t, sketchDir.Mkdir())
	require.NoError(t, sketchFile.WriteFile(
		[]byte("#include <Servo.h>\nvoid setup() {}\nvoid loop() {}\n"),
	))

	// Compile sketch
	_, _, err = cli.RunWithCustomEnv(customEnv, "compile", "-b", "arduino:avr:uno", sketchFile.String())
	require.NoError(t, err)

	// Compile sketch again
	out, _, err := cli.RunWithCustomEnv(customEnv, "compile", "-b", "arduino:avr:uno", "-v", sketchFile.String())
	require.NoError(t, err)
	require.Contains(t, string(out), "Compiling library \"Servo\"\nUsing previously compiled file")
}
