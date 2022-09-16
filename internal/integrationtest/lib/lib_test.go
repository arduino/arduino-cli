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

package lib_test

import (
	"testing"

	"github.com/arduino/arduino-cli/internal/integrationtest"
	"github.com/stretchr/testify/require"
	"go.bug.st/testifyjson/requirejson"
)

func TestLibUpgradeCommand(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	// Updates index for cores and libraries
	_, _, err := cli.Run("core", "update-index")
	require.NoError(t, err)
	_, _, err = cli.Run("lib", "update-index")
	require.NoError(t, err)

	// Install core (this will help to check interaction with platform bundled libraries)
	_, _, err = cli.Run("core", "install", "arduino:avr@1.6.3")
	require.NoError(t, err)

	// Test upgrade of not-installed library
	_, stdErr, err := cli.Run("lib", "upgrade", "Servo")
	require.Error(t, err)
	require.Contains(t, string(stdErr), "Library 'Servo' not found")

	// Test upgrade of installed library
	_, _, err = cli.Run("lib", "install", "Servo@1.1.6")
	require.NoError(t, err)
	stdOut, _, err := cli.Run("lib", "list", "--format", "json")
	require.NoError(t, err)
	requirejson.Contains(t, stdOut, `[ { "library":{ "name":"Servo", "version": "1.1.6" } } ]`)

	_, _, err = cli.Run("lib", "upgrade", "Servo")
	require.NoError(t, err)
	stdOut, _, err = cli.Run("lib", "list", "--format", "json")
	require.NoError(t, err)
	jsonOut := requirejson.Parse(t, stdOut)
	jsonOut.MustNotContain(`[ { "library":{ "name":"Servo", "version": "1.1.6" } } ]`)
	servoVersion := jsonOut.Query(`.[].library | select(.name=="Servo") | .version`).String()

	// Upgrade of already up-to-date library
	_, _, err = cli.Run("lib", "upgrade", "Servo")
	require.NoError(t, err)
	stdOut, _, err = cli.Run("lib", "list", "--format", "json")
	require.NoError(t, err)
	requirejson.Query(t, stdOut, `.[].library | select(.name=="Servo") | .version`, servoVersion)
}

func TestLibInstallMultipleSameLibrary(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()
	cliEnv := cli.GetDefaultEnv()
	cliEnv["ARDUINO_LIBRARY_ENABLE_UNSAFE_INSTALL"] = "true"

	// Check that 'lib install' didn't create a double install
	// https://github.com/arduino/arduino-cli/issues/1870
	_, _, err := cli.RunWithCustomEnv(cliEnv, "lib", "install", "--git-url", "https://github.com/arduino-libraries/SigFox#1.0.3")
	require.NoError(t, err)
	_, _, err = cli.Run("lib", "install", "Arduino SigFox for MKRFox1200")
	require.NoError(t, err)
	jsonOut, _, err := cli.Run("lib", "list", "--format", "json")
	require.NoError(t, err)
	requirejson.Len(t, jsonOut, 1, "A duplicate library install has been detected")

	// Check that 'lib upgrade' didn't create a double install
	// https://github.com/arduino/arduino-cli/issues/1870
	_, _, err = cli.Run("lib", "uninstall", "Arduino SigFox for MKRFox1200")
	require.NoError(t, err)
	_, _, err = cli.RunWithCustomEnv(cliEnv, "lib", "install", "--git-url", "https://github.com/arduino-libraries/SigFox#1.0.3")
	require.NoError(t, err)
	_, _, err = cli.Run("lib", "upgrade", "Arduino SigFox for MKRFox1200")
	require.NoError(t, err)
	jsonOut, _, err = cli.Run("lib", "list", "--format", "json")
	require.NoError(t, err)
	requirejson.Len(t, jsonOut, 1, "A duplicate library install has been detected")
}

func TestDuplicateLibInstallDetection(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()
	cliEnv := cli.GetDefaultEnv()
	cliEnv["ARDUINO_LIBRARY_ENABLE_UNSAFE_INSTALL"] = "true"

	// Make a double install in the sketchbook/user directory
	_, _, err := cli.Run("lib", "install", "ArduinoOTA@1.0.0")
	require.NoError(t, err)
	otaLibPath := cli.SketchbookDir().Join("libraries", "ArduinoOTA")
	err = otaLibPath.CopyDirTo(otaLibPath.Parent().Join("CopyOfArduinoOTA"))
	require.NoError(t, err)
	jsonOut, _, err := cli.Run("lib", "list", "--format", "json")
	require.NoError(t, err)
	requirejson.Len(t, jsonOut, 2, "Duplicate library install is not detected by the CLI")

	_, _, err = cli.Run("lib", "install", "ArduinoOTA")
	require.Error(t, err)
	_, _, err = cli.Run("lib", "upgrade", "ArduinoOTA")
	require.Error(t, err)
	_, _, err = cli.Run("lib", "uninstall", "ArduinoOTA")
	require.Error(t, err)
}
