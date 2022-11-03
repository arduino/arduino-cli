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
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/arduino/arduino-cli/internal/integrationtest"
	"github.com/arduino/go-paths-helper"
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
	// Count how many libraries with the name "Arduino SigFox for MKRFox1200" are installed
	requirejson.Parse(t, jsonOut).
		Query(`[.[].library.name | select(. == "Arduino SigFox for MKRFox1200")]`).
		LengthMustEqualTo(1, "Found multiple installations of Arduino SigFox for MKRFox1200'")

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
	// Count how many libraries with the name "Arduino SigFox for MKRFox1200" are installed
	requirejson.Parse(t, jsonOut).
		Query(`[.[].library.name | select(. == "Arduino SigFox for MKRFox1200")]`).
		LengthMustEqualTo(1, "Found multiple installations of Arduino SigFox for MKRFox1200'")
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

func TestDuplicateLibInstallFromGitDetection(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()
	cliEnv := cli.GetDefaultEnv()
	cliEnv["ARDUINO_LIBRARY_ENABLE_UNSAFE_INSTALL"] = "true"

	// Make a double install in the sketchbook/user directory
	_, _, err := cli.Run("lib", "install", "Arduino SigFox for MKRFox1200")
	require.NoError(t, err)

	_, _, err = cli.RunWithCustomEnv(cliEnv, "lib", "install", "--git-url", "https://github.com/arduino-libraries/SigFox#1.0.3")
	require.NoError(t, err)

	jsonOut, _, err := cli.Run("lib", "list", "--format", "json")
	require.NoError(t, err)
	// Count how many libraries with the name "Arduino SigFox for MKRFox1200" are installed
	requirejson.Parse(t, jsonOut).
		Query(`[.[].library.name | select(. == "Arduino SigFox for MKRFox1200")]`).
		LengthMustEqualTo(1, "Found multiple installations of Arduino SigFox for MKRFox1200'")

	// Try to make a double install by upgrade
	_, _, err = cli.Run("lib", "upgrade")
	require.NoError(t, err)

	// Check if double install happened
	jsonOut, _, err = cli.Run("lib", "list", "--format", "json")
	require.NoError(t, err)
	requirejson.Parse(t, jsonOut).
		Query(`[.[].library.name | select(. == "Arduino SigFox for MKRFox1200")]`).
		LengthMustEqualTo(1, "Found multiple installations of Arduino SigFox for MKRFox1200'")

	// Try to make a double install by zip-installing
	tmp, err := paths.MkTempDir("", "")
	require.NoError(t, err)
	defer tmp.RemoveAll()
	tmpZip := tmp.Join("SigFox.zip")
	defer tmpZip.Remove()

	f, err := tmpZip.Create()
	require.NoError(t, err)
	resp, err := http.Get("https://github.com/arduino-libraries/SigFox/archive/refs/tags/1.0.3.zip")
	require.NoError(t, err)
	_, err = io.Copy(f, resp.Body)
	require.NoError(t, err)
	require.NoError(t, f.Close())

	_, _, err = cli.RunWithCustomEnv(cliEnv, "lib", "install", "--zip-path", tmpZip.String())
	require.NoError(t, err)

	// Check if double install happened
	jsonOut, _, err = cli.Run("lib", "list", "--format", "json")
	require.NoError(t, err)
	requirejson.Parse(t, jsonOut).
		Query(`[.[].library.name | select(. == "Arduino SigFox for MKRFox1200")]`).
		LengthMustEqualTo(1, "Found multiple installations of Arduino SigFox for MKRFox1200'")
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
	_, _, err = cli.Run("lib", "install", "Arduino_DebugUtils")
	require.NoError(t, err)
	_, _, err = cli.Run("lib", "install", "MKRGSM")
	require.NoError(t, err)
	_, _, err = cli.Run("lib", "install", "MKRNB")
	require.NoError(t, err)
	_, _, err = cli.Run("lib", "install", "WiFiNINA")
	require.NoError(t, err)

	stdOut, _, err := cli.Run("lib", "deps", "Arduino_ConnectionHandler@0.6.6", "--no-color")
	require.NoError(t, err)
	lines := strings.Split(strings.TrimSpace(string(stdOut)), "\n")
	require.Len(t, lines, 7)
	require.Regexp(t, `^✓ Arduino_DebugUtils \d+\.\d+\.\d+ is already installed\.$`, lines[0])
	require.Regexp(t, `^✓ MKRGSM \d+\.\d+\.\d+ is already installed\.$`, lines[1])
	require.Regexp(t, `^✓ MKRNB \d+\.\d+\.\d+ is already installed\.$`, lines[2])
	require.Regexp(t, `^✓ WiFiNINA \d+\.\d+\.\d+ is already installed\.$`, lines[3])
	require.Regexp(t, `^✕ Arduino_ConnectionHandler \d+\.\d+\.\d+ must be installed\.$`, lines[4])
	require.Regexp(t, `^✕ MKRWAN \d+\.\d+\.\d+ must be installed\.$`, lines[5])
	require.Regexp(t, `^✕ WiFi101 \d+\.\d+\.\d+ must be installed\.$`, lines[6])

	stdOut, _, err = cli.Run("lib", "deps", "Arduino_ConnectionHandler@0.6.6", "--format", "json")
	require.NoError(t, err)

	var jsonDeps struct {
		Dependencies []struct {
			Name             string `json:"name"`
			VersionRequired  string `json:"version_required"`
			VersionInstalled string `json:"version_installed"`
		} `json:"dependencies"`
	}
	err = json.Unmarshal(stdOut, &jsonDeps)
	require.NoError(t, err)

	require.Equal(t, "Arduino_ConnectionHandler", jsonDeps.Dependencies[0].Name)
	require.Empty(t, jsonDeps.Dependencies[0].VersionInstalled)
	require.Equal(t, "Arduino_DebugUtils", jsonDeps.Dependencies[1].Name)
	require.Equal(t, jsonDeps.Dependencies[1].VersionInstalled, jsonDeps.Dependencies[1].VersionRequired)
	require.Equal(t, "MKRGSM", jsonDeps.Dependencies[2].Name)
	require.Equal(t, jsonDeps.Dependencies[2].VersionInstalled, jsonDeps.Dependencies[2].VersionRequired)
	require.Equal(t, "MKRNB", jsonDeps.Dependencies[3].Name)
	require.Equal(t, jsonDeps.Dependencies[3].VersionInstalled, jsonDeps.Dependencies[3].VersionRequired)
	require.Equal(t, "MKRWAN", jsonDeps.Dependencies[4].Name)
	require.Empty(t, jsonDeps.Dependencies[4].VersionInstalled)
	require.Equal(t, "WiFi101", jsonDeps.Dependencies[5].Name)
	require.Empty(t, jsonDeps.Dependencies[5].VersionInstalled)
	require.Equal(t, "WiFiNINA", jsonDeps.Dependencies[6].Name)
	require.Equal(t, jsonDeps.Dependencies[6].VersionInstalled, jsonDeps.Dependencies[6].VersionRequired)
}

func TestUpgradeLibraryWithDependencies(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	// Updates index for cores and libraries
	_, _, err := cli.Run("core", "update-index")
	require.NoError(t, err)
	_, _, err = cli.Run("lib", "update-index")
	require.NoError(t, err)

	// Install library
	_, _, err = cli.Run("lib", "install", "Arduino_ConnectionHandler@0.3.3")
	require.NoError(t, err)
	stdOut, _, err := cli.Run("lib", "deps", "Arduino_ConnectionHandler@0.3.3", "--format", "json")
	require.NoError(t, err)

	var jsonDeps struct {
		Dependencies []struct {
			Name             string `json:"name"`
			VersionRequired  string `json:"version_required"`
			VersionInstalled string `json:"version_installed"`
		} `json:"dependencies"`
	}
	err = json.Unmarshal(stdOut, &jsonDeps)
	require.NoError(t, err)

	require.Len(t, jsonDeps.Dependencies, 6)
	require.Equal(t, "Arduino_ConnectionHandler", jsonDeps.Dependencies[0].Name)
	require.Equal(t, "Arduino_DebugUtils", jsonDeps.Dependencies[1].Name)
	require.Equal(t, "MKRGSM", jsonDeps.Dependencies[2].Name)
	require.Equal(t, "MKRNB", jsonDeps.Dependencies[3].Name)
	require.Equal(t, "WiFi101", jsonDeps.Dependencies[4].Name)
	require.Equal(t, "WiFiNINA", jsonDeps.Dependencies[5].Name)

	// Test lib upgrade also install new dependencies of already installed library
	_, _, err = cli.Run("lib", "upgrade", "Arduino_ConnectionHandler")
	require.NoError(t, err)
	stdOut, _, err = cli.Run("lib", "deps", "Arduino_ConnectionHandler", "--format", "json")
	require.NoError(t, err)

	jsonOut := requirejson.Parse(t, stdOut)
	dependency := jsonOut.Query(`.dependencies[] | select(.name=="MKRWAN")`)
	require.Equal(t, dependency.Query(".version_required"), dependency.Query(".version_installed"))
}

func TestList(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	// Init the environment explicitly
	_, _, err := cli.Run("core", "update-index")
	require.NoError(t, err)

	// When output is empty, nothing is printed out, no matter the output format
	stdout, stderr, err := cli.Run("lib", "list")
	require.NoError(t, err)
	require.Empty(t, stderr)
	require.Contains(t, strings.TrimSpace(string(stdout)), "No libraries installed.")
	stdout, stderr, err = cli.Run("lib", "list", "--format", "json")
	require.NoError(t, err)
	require.Empty(t, stderr)
	requirejson.Empty(t, stdout)

	// Install something we can list at a version older than latest
	_, _, err = cli.Run("lib", "install", "ArduinoJson@6.11.0")
	require.NoError(t, err)

	// Look at the plain text output
	stdout, stderr, err = cli.Run("lib", "list")
	require.NoError(t, err)
	require.Empty(t, stderr)
	lines := strings.Split(strings.TrimSpace(string(stdout)), "\n")
	require.Equal(t, 2, len(lines))
	lines[1] = strings.Join(strings.Fields(lines[1]), " ")
	toks := strings.SplitN(lines[1], " ", 5)
	// Verifies the expected number of field
	require.Equal(t, 5, len(toks))
	// be sure line contain the current version AND the available version
	require.NotEmpty(t, toks[1])
	require.NotEmpty(t, toks[2])
	// Verifies library sentence
	require.Contains(t, toks[4], "An efficient and elegant JSON library...")

	// Look at the JSON output
	stdout, stderr, err = cli.Run("lib", "list", "--format", "json")
	require.NoError(t, err)
	require.Empty(t, stderr)
	requirejson.Len(t, stdout, 1)
	// be sure data contains the available version
	requirejson.Query(t, stdout, ".[0] | .release | .version != \"\"", "true")

	// Install something we can list without provides_includes field given in library.properties
	_, _, err = cli.Run("lib", "install", "Arduino_APDS9960@1.0.3")
	require.NoError(t, err)
	// Look at the JSON output
	stdout, stderr, err = cli.Run("lib", "list", "Arduino_APDS9960", "--format", "json")
	require.NoError(t, err)
	require.Empty(t, stderr)
	requirejson.Len(t, stdout, 1)
	// be sure data contains the correct provides_includes field
	requirejson.Query(t, stdout, ".[0] | .library | .provides_includes | .[0]", "\"Arduino_APDS9960.h\"")
}

func TestListExitCode(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	// Init the environment explicitly
	_, _, err := cli.Run("core", "update-index")
	require.NoError(t, err)

	_, _, err = cli.Run("core", "list")
	require.NoError(t, err)

	// Verifies lib list doesn't fail when platform is not specified
	_, stderr, err := cli.Run("lib", "list")
	require.NoError(t, err)
	require.Empty(t, stderr)

	// Verify lib list command fails because specified platform is not installed
	_, stderr, err = cli.Run("lib", "list", "-b", "arduino:samd:mkr1000")
	require.Error(t, err)
	require.Contains(t, string(stderr), "Error listing Libraries: Unknown FQBN: platform arduino:samd is not installed")

	_, _, err = cli.Run("lib", "install", "AllThingsTalk LoRaWAN SDK")
	require.NoError(t, err)

	// Verifies lib list command keeps failing
	_, stderr, err = cli.Run("lib", "list", "-b", "arduino:samd:mkr1000")
	require.Error(t, err)
	require.Contains(t, string(stderr), "Error listing Libraries: Unknown FQBN: platform arduino:samd is not installed")

	_, _, err = cli.Run("core", "install", "arduino:samd")
	require.NoError(t, err)

	// Verifies lib list command now works since platform has been installed
	_, stderr, err = cli.Run("lib", "list", "-b", "arduino:samd:mkr1000")
	require.NoError(t, err)
	require.Empty(t, stderr)
}

func TestListWithFqbn(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	// Init the environment explicitly
	_, _, err := cli.Run("core", "update-index")
	require.NoError(t, err)

	// Install core
	_, _, err = cli.Run("core", "install", "arduino:avr")
	require.NoError(t, err)

	// Look at the plain text output
	_, _, err = cli.Run("lib", "install", "ArduinoJson")
	require.NoError(t, err)
	_, _, err = cli.Run("lib", "install", "wm8978-esp32")
	require.NoError(t, err)

	// Look at the plain text output
	stdout, stderr, err := cli.Run("lib", "list", "-b", "arduino:avr:uno")
	require.NoError(t, err)
	require.Empty(t, stderr)
	lines := strings.Split(strings.TrimSpace(string(stdout)), "\n")
	require.Len(t, lines, 2)

	// Verifies library is compatible
	lines[1] = strings.Join(strings.Fields(lines[1]), " ")
	toks := strings.SplitN(lines[1], " ", 5)
	require.Len(t, toks, 5)
	require.Equal(t, "ArduinoJson", toks[0])

	// Look at the JSON output
	stdout, stderr, err = cli.Run("lib", "list", "-b", "arduino:avr:uno", "--format", "json")
	require.NoError(t, err)
	require.Empty(t, stderr)
	requirejson.Len(t, stdout, 1)

	// Verifies library is compatible
	requirejson.Query(t, stdout, ".[0] | .library | .name", "\"ArduinoJson\"")
	requirejson.Query(t, stdout, ".[0] | .library | .compatible_with | .\"arduino:avr:uno\"", "true")
}

func TestListProvidesIncludesFallback(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	// Verifies "provides_includes" field is returned even if libraries don't declare
	// the "includes" property in their "library.properties" file
	_, _, err := cli.Run("update")
	require.NoError(t, err)

	// Install core
	_, _, err = cli.Run("core", "install", "arduino:avr@1.8.3")
	require.NoError(t, err)
	_, _, err = cli.Run("lib", "install", "ArduinoJson@6.17.2")
	require.NoError(t, err)

	// List all libraries, even the ones installed with the above core
	stdout, stderr, err := cli.Run("lib", "list", "--all", "--fqbn", "arduino:avr:uno", "--format", "json")
	require.NoError(t, err)
	require.Empty(t, stderr)

	requirejson.Len(t, stdout, 6)

	requirejson.Query(t, stdout, "[.[] | .library | { (.name): .provides_includes }] | add",
		`{
			"SPI": [
		  		"SPI.h"
			],
			"SoftwareSerial": [
		  		"SoftwareSerial.h"
			],
			"Wire": [
		  		"Wire.h"
			],
			"ArduinoJson": [
		  		"ArduinoJson.h",
		  		"ArduinoJson.hpp"
			],
			"EEPROM": [
		  		"EEPROM.h"
			],
			"HID": [
		  		"HID.h"
			]
	  	}`)
}

func TestLibDownload(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	// Download a specific lib version
	_, _, err := cli.Run("lib", "download", "AudioZero@1.0.0")
	require.NoError(t, err)
	require.FileExists(t, cli.DownloadDir().Join("libraries", "AudioZero-1.0.0.zip").String())

	// Wrong lib version
	_, _, err = cli.Run("lib", "download", "AudioZero@69.42.0")
	require.Error(t, err)

	// Wrong lib
	_, _, err = cli.Run("lib", "download", "AudioZ")
	require.Error(t, err)
}

func TestInstall(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	libs := []string{"Arduino_BQ24195", "CMMC MQTT Connector", "WiFiNINA"}
	// Should be safe to run install multiple times
	_, _, err := cli.Run("lib", "install", libs[0], libs[1], libs[2])
	require.NoError(t, err)
	_, _, err = cli.Run("lib", "install", libs[0], libs[1], libs[2])
	require.NoError(t, err)

	// Test failing-install of library with wrong dependency
	// (https://github.com/arduino/arduino-cli/issues/534)
	_, stderr, err := cli.Run("lib", "install", "MD_Parola@3.2.0")
	require.Error(t, err)
	require.Contains(t, string(stderr), "No valid dependencies solution found: dependency 'MD_MAX72xx' is not available")
}
