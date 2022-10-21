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
	"strings"
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

func TestLibCommandsUsingNameInsteadOfDirName(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	_, _, err := cli.Run("lib", "install", "Robot Motor")
	require.NoError(t, err)

	jsonOut, _, err := cli.Run("lib", "examples", "Robot Motor", "--format", "json")
	require.NoError(t, err)
	requirejson.Len(t, jsonOut, 1, "Library 'Robot Motor' not matched in lib examples command.")

	jsonOut, _, err = cli.Run("lib", "list", "Robot Motor", "--format", "json")
	require.NoError(t, err)
	requirejson.Len(t, jsonOut, 1, "Library 'Robot Motor' not matched in lib list command.")
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

	_, stdErr, err := cli.Run("lib", "install", "ArduinoOTA")
	require.Error(t, err)
	require.Contains(t, string(stdErr), "The library ArduinoOTA has multiple installations")
	_, stdErr, err = cli.Run("lib", "upgrade", "ArduinoOTA")
	require.Error(t, err)
	require.Contains(t, string(stdErr), "The library ArduinoOTA has multiple installations")
	_, stdErr, err = cli.Run("lib", "uninstall", "ArduinoOTA")
	require.Error(t, err)
	require.Contains(t, string(stdErr), "The library ArduinoOTA has multiple installations")
}

func TestLibDepsOutput(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	// Updates index for cores and libraries
	_, _, err := cli.Run("core", "update-index")
	require.NoError(t, err)
	_, _, err = cli.Run("lib", "update-index")
	require.NoError(t, err)

	// Install some libraries that are dependencies of another library
	_, _, err = cli.Run("lib", "install", "Arduino_DebugUtils@1.3.0")
	require.NoError(t, err)
	_, _, err = cli.Run("lib", "install", "MKRGSM@1.5.0")
	require.NoError(t, err)
	_, _, err = cli.Run("lib", "install", "MKRNB@1.5.1")
	require.NoError(t, err)
	_, _, err = cli.Run("lib", "install", "WiFiNINA@1.8.13")
	require.NoError(t, err)

	stdOut, _, err := cli.Run("lib", "deps", "Arduino_ConnectionHandler@0.6.6", "--no-color")
	require.NoError(t, err)
	lines := strings.Split(strings.TrimSpace(string(stdOut)), "\n")
	require.Len(t, lines, 7)
	require.Equal(t, "✓ Arduino_DebugUtils 1.3.0 is already installed.", lines[0])
	require.Equal(t, "✓ MKRGSM 1.5.0 is already installed.", lines[1])
	require.Equal(t, "✓ MKRNB 1.5.1 is already installed.", lines[2])
	require.Equal(t, "✓ WiFiNINA 1.8.13 is already installed.", lines[3])
	require.Equal(t, "✕ Arduino_ConnectionHandler 0.6.6 must be installed.", lines[4])
	require.Equal(t, "✕ MKRWAN 1.1.0 must be installed.", lines[5])
	require.Equal(t, "✕ WiFi101 0.16.1 must be installed.", lines[6])

	stdOut, _, err = cli.Run("lib", "deps", "Arduino_ConnectionHandler@0.6.6", "--format", "json")
	require.NoError(t, err)
	expectedOutput := `{"dependencies":[
	{"name":"Arduino_ConnectionHandler","version_required":"0.6.6"},
	{"name":"Arduino_DebugUtils","version_required":"1.3.0","version_installed":"1.3.0"},
	{"name":"MKRGSM","version_required":"1.5.0","version_installed":"1.5.0"},
	{"name":"MKRNB","version_required":"1.5.1","version_installed":"1.5.1"},
	{"name":"MKRWAN","version_required":"1.1.0"},
	{"name":"WiFi101","version_required":"0.16.1"},
	{"name":"WiFiNINA","version_required":"1.8.13","version_installed":"1.8.13"}]}`
	require.JSONEq(t, expectedOutput, string(stdOut))
}
