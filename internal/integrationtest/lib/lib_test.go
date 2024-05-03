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
	"fmt"
	"io"
	"net/http"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/arduino/arduino-cli/internal/integrationtest"
	"github.com/arduino/go-paths-helper"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
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
	stdOut, _, err := cli.Run("lib", "list", "--json")
	require.NoError(t, err)
	requirejson.Contains(t, stdOut, `{"installed_libraries": [ { "library":{ "name":"Servo", "version": "1.1.6" } } ]}`)

	_, _, err = cli.Run("lib", "upgrade", "Servo")
	require.NoError(t, err)
	stdOut, _, err = cli.Run("lib", "list", "--json")
	require.NoError(t, err)
	jsonOut := requirejson.Parse(t, stdOut)
	jsonOut.MustNotContain(`{"installed_libraries": [{"library":{ "name":"Servo", "version": "1.1.6" }}]}`)
	servoVersion := jsonOut.Query(`.installed_libraries.[].library | select(.name=="Servo") | .version`).String()

	// Upgrade of already up-to-date library
	_, _, err = cli.Run("lib", "upgrade", "Servo")
	require.NoError(t, err)
	stdOut, _, err = cli.Run("lib", "list", "--json")
	require.NoError(t, err)
	requirejson.Query(t, stdOut, `.installed_libraries.[].library | select(.name=="Servo") | .version`, servoVersion)
}

func TestLibCommandsUsingNameInsteadOfDirName(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	_, _, err := cli.Run("lib", "install", "Robot Motor")
	require.NoError(t, err)

	jsonOut, _, err := cli.Run("lib", "examples", "Robot Motor", "--json")
	require.NoError(t, err)
	requirejson.Len(t, jsonOut, 1, "Library 'Robot Motor' not matched in lib examples command.")

	jsonOut, _, err = cli.Run("lib", "list", "Robot Motor", "--json")
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
	jsonOut, _, err := cli.Run("lib", "list", "--json")
	require.NoError(t, err)
	// Count how many libraries with the name "Arduino SigFox for MKRFox1200" are installed
	requirejson.Parse(t, jsonOut).
		Query(`[.installed_libraries.[].library.name | select(. == "Arduino SigFox for MKRFox1200")]`).
		LengthMustEqualTo(1, "Found multiple installations of Arduino SigFox for MKRFox1200'")

	// Check that 'lib upgrade' didn't create a double install
	// https://github.com/arduino/arduino-cli/issues/1870
	_, _, err = cli.Run("lib", "uninstall", "Arduino SigFox for MKRFox1200")
	require.NoError(t, err)
	_, _, err = cli.RunWithCustomEnv(cliEnv, "lib", "install", "--git-url", "https://github.com/arduino-libraries/SigFox#1.0.3")
	require.NoError(t, err)
	_, _, err = cli.Run("lib", "upgrade", "Arduino SigFox for MKRFox1200")
	require.NoError(t, err)
	jsonOut, _, err = cli.Run("lib", "list", "--json")
	require.NoError(t, err)
	// Count how many libraries with the name "Arduino SigFox for MKRFox1200" are installed
	requirejson.Parse(t, jsonOut).
		Query(`[.installed_libraries.[].library.name | select(. == "Arduino SigFox for MKRFox1200")]`).
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
	jsonOut, _, err := cli.Run("lib", "list", "--json")
	require.NoError(t, err)
	requirejson.Query(t, jsonOut, `.installed_libraries | length`, `2`, "Duplicate library install is not detected by the CLI")

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

	jsonOut, _, err := cli.Run("lib", "list", "--json")
	require.NoError(t, err)
	// Count how many libraries with the name "Arduino SigFox for MKRFox1200" are installed
	requirejson.Parse(t, jsonOut).
		Query(`[.installed_libraries.[].library.name | select(. == "Arduino SigFox for MKRFox1200")]`).
		LengthMustEqualTo(1, "Found multiple installations of Arduino SigFox for MKRFox1200'")

	// Try to make a double install by upgrade
	_, _, err = cli.Run("lib", "upgrade")
	require.NoError(t, err)

	// Check if double install happened
	jsonOut, _, err = cli.Run("lib", "list", "--json")
	require.NoError(t, err)
	requirejson.Parse(t, jsonOut).
		Query(`[.installed_libraries.[].library.name | select(. == "Arduino SigFox for MKRFox1200")]`).
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
	jsonOut, _, err = cli.Run("lib", "list", "--json")
	require.NoError(t, err)
	requirejson.Parse(t, jsonOut).
		Query(`[.installed_libraries.[].library.name | select(. == "Arduino SigFox for MKRFox1200")]`).
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

	stdOut, _, err = cli.Run("lib", "deps", "Arduino_ConnectionHandler@0.6.6", "--json")
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
	stdOut, _, err := cli.Run("lib", "deps", "Arduino_ConnectionHandler@0.3.3", "--json")
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
	stdOut, _, err = cli.Run("lib", "deps", "Arduino_ConnectionHandler", "--json")
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
	require.Contains(t, string(stdout), "No libraries installed.")
	stdout, stderr, err = cli.Run("lib", "list", "--json")
	require.NoError(t, err)
	require.Empty(t, stderr)
	requirejson.Query(t, stdout, `.installed_libraries | length`, `0`)

	// Install something we can list at a version older than latest
	_, _, err = cli.Run("lib", "install", "ArduinoJson@6.11.0")
	require.NoError(t, err)

	// Look at the plain text output
	stdout, stderr, err = cli.Run("lib", "list")
	require.NoError(t, err)
	require.Empty(t, stderr)
	lines := strings.Split(strings.TrimSpace(string(stdout)), "\n")
	require.Len(t, lines, 2)
	lines[1] = strings.Join(strings.Fields(lines[1]), " ")
	toks := strings.SplitN(lines[1], " ", 5)
	// Verifies the expected number of field
	require.Len(t, toks, 5)
	// be sure line contain the current version AND the available version
	require.NotEmpty(t, toks[1])
	require.NotEmpty(t, toks[2])
	// Verifies library sentence
	require.Contains(t, toks[4], "An efficient and elegant JSON library...")

	// Look at the JSON output
	stdout, stderr, err = cli.Run("lib", "list", "--json")
	require.NoError(t, err)
	require.Empty(t, stderr)
	requirejson.Query(t, stdout, `.installed_libraries | length`, `1`)
	// be sure data contains the available version
	requirejson.Query(t, stdout, `.installed_libraries.[0] | .release | .version != ""`, "true")

	// Install something we can list without provides_includes field given in library.properties
	_, _, err = cli.Run("lib", "install", "Arduino_APDS9960@1.0.3")
	require.NoError(t, err)
	// Look at the JSON output
	stdout, stderr, err = cli.Run("lib", "list", "Arduino_APDS9960", "--json")
	require.NoError(t, err)
	require.Empty(t, stderr)
	requirejson.Query(t, stdout, `.installed_libraries | length`, `1`)
	// be sure data contains the correct provides_includes field
	requirejson.Query(t, stdout, ".installed_libraries.[0] | .library | .provides_includes | .[0]", `"Arduino_APDS9960.h"`)
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
	require.Contains(t, string(stderr), "Error listing libraries: Unknown FQBN: platform arduino:samd is not installed")

	_, _, err = cli.Run("lib", "install", "AllThingsTalk LoRaWAN SDK")
	require.NoError(t, err)

	// Verifies lib list command keeps failing
	_, stderr, err = cli.Run("lib", "list", "-b", "arduino:samd:mkr1000")
	require.Error(t, err)
	require.Contains(t, string(stderr), "Error listing libraries: Unknown FQBN: platform arduino:samd is not installed")

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
	_, _, err = cli.Run("core", "install", "arduino:avr@1.8.6")
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
	// Check if output contains bundled libraries
	require.Contains(t, string(stdout), "ArduinoJson")
	require.Contains(t, string(stdout), "EEPROM")
	require.Contains(t, string(stdout), "HID")
	lines := strings.Split(strings.TrimSpace(string(stdout)), "\n")
	require.Len(t, lines, 7)

	// Verifies library is compatible
	lines[1] = strings.Join(strings.Fields(lines[1]), " ")
	toks := strings.SplitN(lines[1], " ", 5)
	require.Len(t, toks, 5)
	require.Equal(t, "ArduinoJson", toks[0])

	// Look at the JSON output
	stdout, stderr, err = cli.Run("lib", "list", "-b", "arduino:avr:uno", "--json")
	require.NoError(t, err)
	require.Empty(t, stderr)
	requirejson.Query(t, stdout, `.installed_libraries | length`, `6`)

	// Verifies library is compatible
	requirejson.Query(t, stdout, `.installed_libraries | sort_by(.library | .name) | .[0] | .library | .name`, `"ArduinoJson"`)
	requirejson.Query(t, stdout, `.installed_libraries | sort_by(.library | .name) | .[0] | .library | .compatible_with | ."arduino:avr:uno"`, `true`)

	// Verifies bundled libs are shown if -b flag is used
	requirejson.Parse(t, stdout).Query(`.installed_libraries.[] | .library | select(.container_platform=="arduino:avr@1.8.6")`).MustNotBeEmpty()
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
	stdout, stderr, err := cli.Run("lib", "list", "--all", "--fqbn", "arduino:avr:uno", "--json")
	require.NoError(t, err)
	require.Empty(t, stderr)

	requirejson.Query(t, stdout, `.installed_libraries | length`, `6`)

	requirejson.Query(t, stdout, "[.installed_libraries.[] | .library | { (.name): .provides_includes }] | add",
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

	// Test installing a library with a "relaxed" version
	// https://github.com/arduino/arduino-cli/issues/1727
	_, _, err = cli.Run("lib", "install", "ILI9341_t3@1.0")
	require.NoError(t, err)
	stdout, _, err := cli.Run("lib", "list", "--json")
	require.NoError(t, err)
	requirejson.Parse(t, stdout).Query(`.installed_libraries.[] | select(.library.name == "ILI9341_t3") | .library.version`).MustEqual(`"1.0"`)
	_, _, err = cli.Run("lib", "install", "ILI9341_t3@1")
	require.NoError(t, err)
	_, _, err = cli.Run("lib", "install", "ILI9341_t3@1.0.0")
	require.NoError(t, err)
}

func TestInstallLibraryWithDependencies(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	_, _, err := cli.Run("update")
	require.NoError(t, err)

	// Verifies libraries are not installed
	stdout, _, err := cli.Run("lib", "list", "--json")
	require.NoError(t, err)
	requirejson.Query(t, stdout, `.installed_libraries | length`, `0`)

	// Install library
	_, _, err = cli.Run("lib", "install", "MD_Parola@3.5.5")
	require.NoError(t, err)

	// Verifies library's dependencies are correctly installed
	stdout, _, err = cli.Run("lib", "list", "--json")
	require.NoError(t, err)
	requirejson.Query(t, stdout, `[.installed_libraries.[] | .library | .name ] | sort`, `["MD_MAX72XX","MD_Parola"]`)

	// Try upgrading with --no-overwrite (should fail) and without --no-overwrite (should succeed)
	_, _, err = cli.Run("lib", "install", "MD_Parola@3.6.1", "--no-overwrite")
	require.Error(t, err)
	_, _, err = cli.Run("lib", "install", "MD_Parola@3.6.1")
	require.NoError(t, err)

	// Test --no-overwrite with transitive dependencies
	_, _, err = cli.Run("lib", "install", "SD")
	require.NoError(t, err)
	_, _, err = cli.Run("lib", "install", "Arduino_Builtin", "--no-overwrite")
	require.NoError(t, err)
	_, _, err = cli.Run("lib", "install", "SD@1.2.3")
	require.NoError(t, err)
	// This time it should accept the installation with the currently installed SD 1.2.3
	out, _, err := cli.Run("lib", "install", "Arduino_Builtin", "--no-overwrite")
	require.NoError(t, err)
	require.Contains(t, string(out), "Already installed SD@1.2.3")
}

func TestInstallNoDeps(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	_, _, err := cli.Run("update")
	require.NoError(t, err)

	// Verifies libraries are not installed
	stdout, _, err := cli.Run("lib", "list", "--json")
	require.NoError(t, err)
	requirejson.Query(t, stdout, `.installed_libraries | length`, `0`)

	// Install library skipping dependencies installation
	_, _, err = cli.Run("lib", "install", "MD_Parola@3.5.5", "--no-deps")
	require.NoError(t, err)

	// Verifies library's dependencies are not installed
	stdout, _, err = cli.Run("lib", "list", "--json")
	require.NoError(t, err)
	requirejson.Query(t, stdout, `.installed_libraries.[] | .library | .name`, `"MD_Parola"`)
}

func TestInstallWithGitUrl(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	// Initialize configs to enable --git-url flag
	envVar := cli.GetDefaultEnv()
	envVar["ARDUINO_ENABLE_UNSAFE_LIBRARY_INSTALL"] = "true"
	_, _, err := cli.RunWithCustomEnv(envVar, "config", "init", "--dest-dir", ".")
	require.NoError(t, err)

	libInstallDir := cli.SketchbookDir().Join("libraries", "WiFi101")
	// Verifies library is not already installed
	require.NoDirExists(t, libInstallDir.String())

	gitUrl := "https://github.com/arduino-libraries/WiFi101.git"

	// Test git-url library install
	stdout, _, err := cli.Run("lib", "install", "--git-url", gitUrl, "--config-file", "arduino-cli.yaml")
	require.NoError(t, err)
	require.Contains(t, string(stdout), "--git-url and --zip-path flags allow installing untrusted files, use it at your own risk.")

	// Verifies library is installed in expected path
	require.DirExists(t, libInstallDir.String())

	// Reinstall library
	_, _, err = cli.Run("lib", "install", "--git-url", gitUrl, "--config-file", "arduino-cli.yaml")
	require.NoError(t, err)

	// Verifies library remains installed
	require.DirExists(t, libInstallDir.String())
}

func TestInstallWithGitUrlFragmentAsBranch(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	// Initialize configs to enable --git-url flag
	envVar := cli.GetDefaultEnv()
	envVar["ARDUINO_ENABLE_UNSAFE_LIBRARY_INSTALL"] = "true"
	_, _, err := cli.RunWithCustomEnv(envVar, "config", "init", "--dest-dir", ".")
	require.NoError(t, err)

	libInstallDir := cli.SketchbookDir().Join("libraries", "WiFi101")
	// Verifies library is not already installed
	require.NoDirExists(t, libInstallDir.String())

	gitUrl := "https://github.com/arduino-libraries/WiFi101.git"

	// Test that a bad ref fails
	_, _, err = cli.Run("lib", "install", "--git-url", gitUrl+"#x-ref-does-not-exist", "--config-file", "arduino-cli.yaml")
	require.Error(t, err)

	// Verifies library is installed in expected path
	_, _, err = cli.Run("lib", "install", "--git-url", gitUrl+"#0.16.0", "--config-file", "arduino-cli.yaml")
	require.NoError(t, err)
	require.DirExists(t, libInstallDir.String())

	// Reinstall library at an existing ref
	_, _, err = cli.Run("lib", "install", "--git-url", gitUrl+"#master", "--config-file", "arduino-cli.yaml")
	require.NoError(t, err)

	// Verifies library remains installed
	require.DirExists(t, libInstallDir.String())
}

func TestUpdateIndex(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	stdout, _, err := cli.Run("lib", "update-index")
	require.NoError(t, err)
	require.Contains(t, string(stdout), "Downloading index: library_index.tar.bz2 downloaded")
}

func TestUninstall(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	libs := []string{"Arduino_BQ24195", "WiFiNINA"}
	_, _, err := cli.Run("lib", "install", libs[0], libs[1])
	require.NoError(t, err)

	_, _, err = cli.Run("lib", "uninstall", libs[0], libs[1])
	require.NoError(t, err)
}

func TestUninstallSpaces(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	key := "LiquidCrystal I2C"
	_, _, err := cli.Run("lib", "install", key)
	require.NoError(t, err)
	_, _, err = cli.Run("lib", "uninstall", key)
	require.NoError(t, err)
	stdout, _, err := cli.Run("lib", "list", "--json")
	require.NoError(t, err)
	requirejson.Query(t, stdout, `.installed_libraries | length`, `0`)
}

func TestLibOpsCaseInsensitive(t *testing.T) {
	/*This test is supposed to (un)install the following library,
	  As you can see the name is all caps:

	  Name: "PCM"
	    Author: David Mellis <d.mellis@bcmi-labs.cc>, Michael Smith <michael@hurts.ca>
	    Maintainer: David Mellis <d.mellis@bcmi-labs.cc>
	    Sentence: Playback of short audio samples.
	    Paragraph: These samples are encoded directly in the Arduino sketch as an array of numbers.
	    Website: http://highlowtech.org/?p=1963
	    Category: Signal Input/Output
	    Architecture: avr
	    Types: Contributed
	    Versions: [1.0.0]*/

	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	key := "pcm"
	_, _, err := cli.Run("lib", "install", key)
	require.NoError(t, err)
	_, _, err = cli.Run("lib", "uninstall", key)
	require.NoError(t, err)
	stdout, _, err := cli.Run("lib", "list", "--json")
	require.NoError(t, err)
	requirejson.Query(t, stdout, `.installed_libraries | length`, `0`)
}

func TestSearch(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	stdout, _, err := cli.Run("lib", "search", "--names")
	require.NoError(t, err)
	lines := strings.Split(strings.TrimSpace(string(stdout)), "\n")
	var libs []string
	for i, v := range lines {
		lines[i] = strings.TrimSpace(v)
		if strings.Contains(v, "Name:") {
			libs = append(libs, strings.Trim(strings.SplitN(v, " ", 2)[1], "\""))
		}
	}

	expected := []string{"WiFi101", "WiFi101OTA", "Firebase Arduino based on WiFi101", "WiFi101_Generic"}
	require.Subset(t, libs, expected)

	stdout, _, err = cli.Run("lib", "search", "--names", "--json")
	require.NoError(t, err)
	requirejson.Query(t, stdout, ".libraries | length", fmt.Sprint(len(libs)))

	runSearch := func(args string, expectedLibs []string) {
		stdout, _, err = cli.Run("lib", "search", "--names", "--json", args)
		require.NoError(t, err)
		libraries := requirejson.Parse(t, stdout).Query("[ .libraries | .[] | .name ]").String()
		for _, l := range expectedLibs {
			require.Contains(t, libraries, l)
		}
	}
	runSearch("Arduino_MKRIoTCarrier", []string{"Arduino_MKRIoTCarrier"})
	runSearch("Arduino mkr iot carrier", []string{"Arduino_MKRIoTCarrier"})
	runSearch("mkr iot carrier", []string{"Arduino_MKRIoTCarrier"})
	runSearch("mkriotcarrier", []string{"Arduino_MKRIoTCarrier"})
	runSearch("dht", []string{"DHT sensor library", "DHT sensor library for ESPx", "DHT12", "SimpleDHT", "TinyDHT sensor library", "SDHT"})
	runSearch("dht11", []string{"DHT sensor library", "DHT sensor library for ESPx", "SimpleDHT", "SDHT"})
	runSearch("dht12", []string{"DHT12", "DHT12 sensor library", "SDHT"})
	runSearch("dht22", []string{"DHT sensor library", "DHT sensor library for ESPx", "SimpleDHT", "SDHT"})
	runSearch("dht sensor", []string{"DHT sensor library", "DHT sensor library for ESPx", "SimpleDHT", "SDHT"})
	runSearch("sensor dht", []string{})
	runSearch("arduino json", []string{"ArduinoJson", "Arduino_JSON"})
	runSearch("arduinojson", []string{"ArduinoJson"})
	runSearch("json", []string{"ArduinoJson", "Arduino_JSON"})
}

func TestQualifiedSearch(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	runSearch := func(args string, expectedLibs []string) {
		stdout, _, err := cli.Run("lib", "search", "--names", "--json", args)
		require.NoError(t, err)
		libraries := requirejson.Parse(t, stdout).Query("[ .libraries | .[] | .name ]").String()
		for _, l := range expectedLibs {
			require.Contains(t, libraries, l)
		}
	}
	runSearch("name:MKRIoTCarrier", []string{"Arduino_MKRIoTCarrier"})
	runSearch("name=Arduino_MKRIoTCarrier", []string{"Arduino_MKRIoTCarrier"})
	// Embedded space in double-quoted string
	runSearch("name=\"dht sensor library\"", []string{"DHT sensor library"})
	// No closing double-quote
	runSearch("name=\"dht sensor library", []string{"DHT sensor library"})
	runSearch("name:\"sensor dht\"", []string{})
	// Literal double-quote
	runSearch("sentence:\\\"", []string{"RTCtime"})
}

func TestSearchParagraph(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	// Search for a string that's only present in the `paragraph` field
	// within the index file.
	_, _, err := cli.Run("lib", "update-index")
	require.NoError(t, err)
	stdout, _, err := cli.Run("lib", "search", "A simple and efficient JSON library", "--names", "--json")
	require.NoError(t, err)
	requirejson.Contains(t, stdout, `{
		"libraries": [
			{
				"name": "ArduinoJson"
			}
		]
	}`)
}

func TestLibListWithUpdatableFlag(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	// Init the environment explicitly
	_, _, err := cli.Run("lib", "update-index")
	require.NoError(t, err)

	// No libraries to update
	stdout, stderr, err := cli.Run("lib", "list", "--updatable")
	require.NoError(t, err)
	require.Empty(t, stderr)
	require.Contains(t, string(stdout), "No libraries update is available.")
	// No library to update in json
	stdout, stderr, err = cli.Run("lib", "list", "--updatable", "--json")
	require.NoError(t, err)
	require.Empty(t, stderr)
	requirejson.Query(t, stdout, `.installed_libraries | length`, `0`)

	// Install outdated library
	_, _, err = cli.Run("lib", "install", "ArduinoJson@6.11.0")
	require.NoError(t, err)
	// Install latest version of library
	_, _, err = cli.Run("lib", "install", "WiFi101")
	require.NoError(t, err)

	stdout, stderr, err = cli.Run("lib", "list", "--updatable")
	require.NoError(t, err)
	require.Empty(t, stderr)
	var lines [][]string
	for _, v := range strings.Split(strings.TrimSpace(string(stdout)), "\n") {
		v = strings.Join(strings.Fields(v), " ")
		lines = append(lines, strings.SplitN(v, " ", 5))
	}
	require.Len(t, lines, 2)
	require.Subset(t, lines[0], []string{"Name", "Installed", "Available", "Location", "Description"})
	require.Equal(t, "ArduinoJson", lines[1][0])
	require.Equal(t, "6.11.0", lines[1][1])
	// Verifies available version is not equal to installed one and not empty
	require.NotEqual(t, "6.11.0", lines[1][2])
	require.NotEmpty(t, lines[1][2])
	require.Equal(t, "An efficient and elegant JSON library...", lines[1][4])

	// Look at the JSON output
	stdout, stderr, err = cli.Run("lib", "list", "--updatable", "--json")
	require.NoError(t, err)
	require.Empty(t, stderr)
	requirejson.Query(t, stdout, `.installed_libraries | length`, `1`)
	// be sure data contains the available version
	requirejson.Query(t, stdout, `.installed_libraries.[0] | .library | .version`, `"6.11.0"`)
	requirejson.Query(t, stdout, `.installed_libraries.[0] | .release | .version != "6.11.0"`, `true`)
	requirejson.Query(t, stdout, `.installed_libraries.[0] | .release | .version != ""`, `true`)
}

func TestInstallWithGitUrlFromCurrentDirectory(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	_, _, err := cli.Run("update")
	require.NoError(t, err)

	envVar := cli.GetDefaultEnv()
	envVar["ARDUINO_ENABLE_UNSAFE_LIBRARY_INSTALL"] = "true"

	libInstallDir := cli.SketchbookDir().Join("libraries", "WiFi101")
	// Verifies library is not installed
	require.NoDirExists(t, libInstallDir.String())

	// Clone repository locally
	gitUrl := "https://github.com/arduino-libraries/WiFi101.git"
	repoDir := cli.SketchbookDir().Join("WiFi101")
	_, err = git.PlainClone(repoDir.String(), false, &git.CloneOptions{
		URL: gitUrl,
	})
	require.NoError(t, err)

	cli.SetWorkingDir(repoDir)
	_, _, err = cli.RunWithCustomEnv(envVar, "lib", "install", "--git-url", ".")
	require.NoError(t, err)

	// Verifies library is installed to correct folder
	require.DirExists(t, libInstallDir.String())
}

func TestInstallWithGitLocalUrl(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	_, _, err := cli.Run("update")
	require.NoError(t, err)

	envVar := cli.GetDefaultEnv()
	envVar["ARDUINO_ENABLE_UNSAFE_LIBRARY_INSTALL"] = "true"

	libInstallDir := cli.SketchbookDir().Join("libraries", "WiFi101")
	// Verifies library is not installed
	require.NoDirExists(t, libInstallDir.String())

	// Clone repository locally
	gitUrl := "https://github.com/arduino-libraries/WiFi101.git"
	repoDir := cli.SketchbookDir().Join("WiFi101")
	_, err = git.PlainClone(repoDir.String(), false, &git.CloneOptions{
		URL: gitUrl,
	})
	require.NoError(t, err)

	_, _, err = cli.RunWithCustomEnv(envVar, "lib", "install", "--git-url", repoDir.String())
	require.NoError(t, err)

	// Verifies library is installed
	require.DirExists(t, libInstallDir.String())
}

func TestInstallWithGitUrlRelativePath(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	_, _, err := cli.Run("update")
	require.NoError(t, err)

	envVar := cli.GetDefaultEnv()
	envVar["ARDUINO_ENABLE_UNSAFE_LIBRARY_INSTALL"] = "true"

	libInstallDir := cli.SketchbookDir().Join("libraries", "WiFi101")
	// Verifies library is not installed
	require.NoDirExists(t, libInstallDir.String())

	// Clone repository locally
	gitUrl := "https://github.com/arduino-libraries/WiFi101.git"
	repoDir := cli.SketchbookDir().Join("WiFi101")
	_, err = git.PlainClone(repoDir.String(), false, &git.CloneOptions{
		URL: gitUrl,
	})
	require.NoError(t, err)

	cli.SetWorkingDir(cli.SketchbookDir())
	_, _, err = cli.RunWithCustomEnv(envVar, "lib", "install", "--git-url", "./WiFi101")
	require.NoError(t, err)

	// Verifies library is installed
	require.DirExists(t, libInstallDir.String())
}

func TestInstallWithGitUrlDoesNotCreateGitRepo(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	_, _, err := cli.Run("update")
	require.NoError(t, err)

	envVar := cli.GetDefaultEnv()
	envVar["ARDUINO_ENABLE_UNSAFE_LIBRARY_INSTALL"] = "true"

	libInstallDir := cli.SketchbookDir().Join("libraries", "WiFi101")
	// Verifies library is not installed
	require.NoDirExists(t, libInstallDir.String())

	// Clone repository locally
	gitUrl := "https://github.com/arduino-libraries/WiFi101.git"
	repoDir := cli.SketchbookDir().Join("WiFi101")
	_, err = git.PlainClone(repoDir.String(), false, &git.CloneOptions{
		URL: gitUrl,
	})
	require.NoError(t, err)

	_, _, err = cli.RunWithCustomEnv(envVar, "lib", "install", "--git-url", repoDir.String())
	require.NoError(t, err)

	// Verifies installed library is not a git repository
	require.NoDirExists(t, libInstallDir.Join(".git").String())
}

func TestInstallWithGitUrlMultipleLibraries(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	_, _, err := cli.Run("update")
	require.NoError(t, err)

	envVar := cli.GetDefaultEnv()
	envVar["ARDUINO_ENABLE_UNSAFE_LIBRARY_INSTALL"] = "true"

	wifiInstallDir := cli.SketchbookDir().Join("libraries", "WiFi101")
	bleInstallDir := cli.SketchbookDir().Join("libraries", "ArduinoBLE")
	// Verifies library are not installed
	require.NoDirExists(t, wifiInstallDir.String())
	require.NoDirExists(t, bleInstallDir.String())

	wifiUrl := "https://github.com/arduino-libraries/WiFi101.git"
	bleUrl := "https://github.com/arduino-libraries/ArduinoBLE.git"

	_, _, err = cli.RunWithCustomEnv(envVar, "lib", "install", "--git-url", wifiUrl, bleUrl)
	require.NoError(t, err)

	// Verifies library are installed
	require.DirExists(t, wifiInstallDir.String())
	require.DirExists(t, bleInstallDir.String())
}

func TestLibExamples(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	_, _, err := cli.Run("update")
	require.NoError(t, err)

	_, _, err = cli.Run("lib", "install", "Arduino_JSON@0.1.0")
	require.NoError(t, err)

	stdout, _, err := cli.Run("lib", "examples", "Arduino_JSON", "--json")
	require.NoError(t, err)
	requirejson.Query(t, stdout, `.examples | length`, `1`)
	examples := requirejson.Parse(t, stdout).Query(".examples.[0] | .examples").String()
	examples = strings.ReplaceAll(examples, "\\\\", "\\")
	require.Contains(t, examples, cli.SketchbookDir().Join("libraries", "Arduino_JSON", "examples", "JSONArray").String())
	require.Contains(t, examples, cli.SketchbookDir().Join("libraries", "Arduino_JSON", "examples", "JSONKitchenSink").String())
	require.Contains(t, examples, cli.SketchbookDir().Join("libraries", "Arduino_JSON", "examples", "JSONObject").String())
}

func TestLibExamplesWithPdeFile(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	_, _, err := cli.Run("update")
	require.NoError(t, err)

	_, _, err = cli.Run("lib", "install", "Encoder@1.4.1")
	require.NoError(t, err)

	stdout, _, err := cli.Run("lib", "examples", "Encoder", "--json")
	require.NoError(t, err)
	requirejson.Query(t, stdout, `.examples | length`, `1`)
	examples := requirejson.Parse(t, stdout).Query(".examples.[0] | .examples").String()
	examples = strings.ReplaceAll(examples, "\\\\", "\\")
	require.Contains(t, examples, cli.SketchbookDir().Join("libraries", "Encoder", "examples", "Basic").String())
	require.Contains(t, examples, cli.SketchbookDir().Join("libraries", "Encoder", "examples", "NoInterrupts").String())
	require.Contains(t, examples, cli.SketchbookDir().Join("libraries", "Encoder", "examples", "SpeedTest").String())
	require.Contains(t, examples, cli.SketchbookDir().Join("libraries", "Encoder", "examples", "TwoKnobs").String())
}

func TestLibExamplesWithCaseMismatch(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	_, _, err := cli.Run("update")
	require.NoError(t, err)

	_, _, err = cli.Run("lib", "install", "WiFiManager@2.0.3-alpha")
	require.NoError(t, err)

	stdout, _, err := cli.Run("lib", "examples", "WiFiManager", "--json")
	require.NoError(t, err)
	requirejson.Query(t, stdout, `.examples | length`, `1`)
	requirejson.Query(t, stdout, ".examples.[0] | .examples | length", "14")

	examples := requirejson.Parse(t, stdout).Query(".examples.[0] | .examples").String()
	examples = strings.ReplaceAll(examples, "\\\\", "\\")
	examplesPath := cli.SketchbookDir().Join("libraries", "WiFiManager", "examples")
	// Verifies sketches with correct casing are listed
	require.Contains(t, examples, examplesPath.Join("Advanced").String())
	require.Contains(t, examples, examplesPath.Join("AutoConnect", "AutoConnectWithFeedbackLED").String())
	require.Contains(t, examples, examplesPath.Join("AutoConnect", "AutoConnectWithFSParameters").String())
	require.Contains(t, examples, examplesPath.Join("AutoConnect", "AutoConnectWithFSParametersAndCustomIP").String())
	require.Contains(t, examples, examplesPath.Join("Basic").String())
	require.Contains(t, examples, examplesPath.Join("DEV", "OnDemandConfigPortal").String())
	require.Contains(t, examples, examplesPath.Join("NonBlocking", "AutoConnectNonBlocking").String())
	require.Contains(t, examples, examplesPath.Join("NonBlocking", "AutoConnectNonBlockingwParams").String())
	require.Contains(t, examples, examplesPath.Join("Old_examples", "AutoConnectWithFeedback").String())
	require.Contains(t, examples, examplesPath.Join("Old_examples", "AutoConnectWithReset").String())
	require.Contains(t, examples, examplesPath.Join("Old_examples", "AutoConnectWithStaticIP").String())
	require.Contains(t, examples, examplesPath.Join("Old_examples", "AutoConnectWithTimeout").String())
	require.Contains(t, examples, examplesPath.Join("OnDemand", "OnDemandConfigPortal").String())
	require.Contains(t, examples, examplesPath.Join("ParamsChildClass").String())
	// Verifies sketches with wrong casing are not returned
	require.NotContains(t, examples, examplesPath.Join("NonBlocking", "OnDemandNonBlocking").String())
	require.NotContains(t, examples, examplesPath.Join("OnDemand", "OnDemandWebPortal").String())
}

func TestLibCommandsWithLibraryHavingInvalidVersion(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	_, _, err := cli.Run("update")
	require.NoError(t, err)

	// Install a library
	_, _, err = cli.Run("lib", "install", "WiFi101@0.16.1")
	require.NoError(t, err)

	// Verifies library is correctly returned
	stdout, _, err := cli.Run("lib", "list", "--json")
	require.NoError(t, err)
	requirejson.Query(t, stdout, `.installed_libraries | length`, `1`)
	requirejson.Query(t, stdout, ".installed_libraries.[0] | .library | .version", `"0.16.1"`)

	// Changes the version of the currently installed library so that it's
	// invalid
	libPath := cli.SketchbookDir().Join("libraries", "WiFi101", "library.properties")
	require.NoError(t, libPath.WriteFile([]byte("name=WiFi101\nversion=1.0001")))

	// Verifies version is now empty
	stdout, _, err = cli.Run("lib", "list", "--json")
	require.NoError(t, err)
	requirejson.Query(t, stdout, `.installed_libraries | length`, `1`)
	requirejson.Query(t, stdout, ".installed_libraries.[0] | .library | .version", "null")

	// Upgrade library
	_, _, err = cli.Run("lib", "upgrade", "WiFi101")
	require.NoError(t, err)

	// Verifies library has been updated
	stdout, _, err = cli.Run("lib", "list", "--json")
	require.NoError(t, err)
	requirejson.Query(t, stdout, `.installed_libraries | length`, `1`)
	requirejson.Query(t, stdout, ".installed_libraries.[0] | .library | .version != \"\"", "true")
}

func TestInstallZipLibWithMacosMetadata(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	// Initialize configs to enable --zip-path flag
	envVar := cli.GetDefaultEnv()
	envVar["ARDUINO_ENABLE_UNSAFE_LIBRARY_INSTALL"] = "true"
	_, _, err := cli.RunWithCustomEnv(envVar, "config", "init", "--dest-dir", ".")
	require.NoError(t, err)

	libInstallDir := cli.SketchbookDir().Join("libraries", "fake-lib")
	// Verifies library is not already installed
	require.NoDirExists(t, libInstallDir.String())

	zipPath, err := paths.New("..", "testdata", "fake-lib.zip").Abs()
	require.NoError(t, err)
	// Test zip-path install
	stdout, _, err := cli.Run("lib", "install", "--zip-path", zipPath.String(), "--config-file", "arduino-cli.yaml")
	require.NoError(t, err)
	require.Contains(t, string(stdout), "--git-url and --zip-path flags allow installing untrusted files, use it at your own risk.")

	// Verifies library is installed in expected path
	require.DirExists(t, libInstallDir.String())
	require.FileExists(t, libInstallDir.Join("library.properties").String())
	require.FileExists(t, libInstallDir.Join("src", "fake-lib.h").String())

	// Reinstall library
	_, _, err = cli.Run("lib", "install",
		"--zip-path", zipPath.String(),
		"--config-file", "arduino-cli.yaml")
	require.NoError(t, err)

	// Verifies library remains installed
	require.DirExists(t, libInstallDir.String())
	require.FileExists(t, libInstallDir.Join("library.properties").String())
	require.FileExists(t, libInstallDir.Join("src", "fake-lib.h").String())
}

func TestInstallZipInvalidLibrary(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	// Initialize configs to enable --zip-path flag
	envVar := cli.GetDefaultEnv()
	envVar["ARDUINO_ENABLE_UNSAFE_LIBRARY_INSTALL"] = "true"
	_, _, err := cli.RunWithCustomEnv(envVar, "config", "init", "--dest-dir", ".")
	require.NoError(t, err)

	libInstallDir := cli.SketchbookDir().Join("libraries", "lib-without-header")
	// Verifies library is not already installed
	require.NoDirExists(t, libInstallDir.String())

	zipPath, err := paths.New("..", "testdata", "lib-without-header.zip").Abs()
	require.NoError(t, err)
	// Test zip-path install
	_, stderr, err := cli.Run("lib", "install", "--zip-path", zipPath.String(), "--config-file", "arduino-cli.yaml")
	require.Error(t, err)
	require.Contains(t, string(stderr), "library not valid")

	libInstallDir = cli.SketchbookDir().Join("libraries", "lib-without-properties")
	// Verifies library is not already installed
	require.NoDirExists(t, libInstallDir.String())

	zipPath, err = paths.New("..", "testdata", "lib-without-properties.zip").Abs()
	require.NoError(t, err)
	// Test zip-path install
	_, stderr, err = cli.Run("lib", "install", "--zip-path", zipPath.String(), "--config-file", "arduino-cli.yaml")
	require.Error(t, err)
	require.Contains(t, string(stderr), "library not valid")
}

func TestInstallGitInvalidLibrary(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	// Initialize configs to enable --zip-path flag
	envVar := cli.GetDefaultEnv()
	envVar["ARDUINO_ENABLE_UNSAFE_LIBRARY_INSTALL"] = "true"
	_, _, err := cli.RunWithCustomEnv(envVar, "config", "init", "--dest-dir", ".")
	require.NoError(t, err)

	// Create fake library repository
	repoDir := cli.SketchbookDir().Join("lib-without-header")
	repo, err := git.PlainInit(repoDir.String(), false)
	require.NoError(t, err)
	libProperties := repoDir.Join("library.properties")
	f, err := libProperties.Create()
	require.NoError(t, err)
	require.NoError(t, f.Close())
	tree, err := repo.Worktree()
	require.NoError(t, err)
	_, err = tree.Add("library.properties")
	require.NoError(t, err)
	_, err = tree.Commit("First commit", &git.CommitOptions{
		All: false, Author: &object.Signature{Name: "a", Email: "b", When: time.Now()}, Committer: nil, Parents: nil, SignKey: nil})
	require.NoError(t, err)

	libInstallDir := cli.SketchbookDir().Join("libraries", "lib-without-header")
	// Verifies library is not already installed
	require.NoDirExists(t, libInstallDir.String())

	_, stderr, err := cli.RunWithCustomEnv(envVar, "lib", "install", "--git-url", repoDir.String(), "--config-file", "arduino-cli.yaml")
	require.Error(t, err)
	require.Contains(t, string(stderr), "library not valid")
	require.NoDirExists(t, libInstallDir.String())

	// Create another fake library repository
	repoDir = cli.SketchbookDir().Join("lib-without-properties")
	repo, err = git.PlainInit(repoDir.String(), false)
	require.NoError(t, err)
	libHeader := repoDir.Join("src", "lib-without-properties.h")
	require.NoError(t, libHeader.Parent().MkdirAll())
	f, err = libHeader.Create()
	require.NoError(t, err)
	require.NoError(t, f.Close())
	tree, err = repo.Worktree()
	require.NoError(t, err)
	_, err = tree.Add("src/lib-without-properties.h")
	require.NoError(t, err)
	_, err = tree.Commit("First commit", &git.CommitOptions{
		All: false, Author: &object.Signature{Name: "a", Email: "b", When: time.Now()}, Committer: nil, Parents: nil, SignKey: nil})
	require.NoError(t, err)

	libInstallDir = cli.SketchbookDir().Join("libraries", "lib-without-properties")
	// Verifies library is not already installed
	require.NoDirExists(t, libInstallDir.String())

	_, stderr, err = cli.RunWithCustomEnv(envVar, "lib", "install", "--git-url", repoDir.String(), "--config-file", "arduino-cli.yaml")
	require.Error(t, err)
	require.Contains(t, string(stderr), "library not valid")
	require.NoDirExists(t, libInstallDir.String())
}

func TestUpgradeDoesNotTryToUpgradeBundledCoreLibrariesInSketchbook(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	testPlatformName := "platform_with_bundled_library"
	platformInstallDir := cli.SketchbookDir().Join("hardware", "arduino-beta-dev", testPlatformName)
	require.NoError(t, platformInstallDir.Parent().MkdirAll())

	// Install platform in Sketchbook hardware dir
	require.NoError(t, paths.New("..", "testdata", testPlatformName).CopyDirTo(platformInstallDir))

	_, _, err := cli.Run("update")
	require.NoError(t, err)

	// Install latest version of library identical to one
	// bundled with test platform
	_, _, err = cli.Run("lib", "install", "USBHost")
	require.NoError(t, err)

	stdout, _, err := cli.Run("lib", "list", "--all", "--json")
	require.NoError(t, err)
	requirejson.Query(t, stdout, `.installed_libraries | length`, `2`)
	// Verify both libraries have the same name
	requirejson.Query(t, stdout, ".installed_libraries.[0] | .library | .name", `"USBHost"`)
	requirejson.Query(t, stdout, ".installed_libraries.[1] | .library | .name", `"USBHost"`)

	stdout, _, err = cli.Run("lib", "upgrade")
	require.NoError(t, err)
	// Empty output means nothing has been updated as expected
	require.Empty(t, stdout)
}

func TestUpgradeDoesNotTryToUpgradeBundledCoreLibraries(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	testPlatformName := "platform_with_bundled_library"
	platformInstallDir := cli.DataDir().Join("packages", "arduino", "hardware", "arch", "4.2.0")
	require.NoError(t, platformInstallDir.Parent().MkdirAll())

	// Install platform in Sketchbook hardware dir
	require.NoError(t, paths.New("..", "testdata", testPlatformName).CopyDirTo(platformInstallDir))

	_, _, err := cli.Run("update")
	require.NoError(t, err)

	// Install latest version of library identical to one
	// bundled with test platform
	_, _, err = cli.Run("lib", "install", "USBHost")
	require.NoError(t, err)

	stdout, _, err := cli.Run("lib", "list", "--all", "--json")
	require.NoError(t, err)
	requirejson.Query(t, stdout, `.installed_libraries | length`, `2`)
	// Verify both libraries have the same name
	requirejson.Query(t, stdout, ".installed_libraries.[0] | .library | .name", `"USBHost"`)
	requirejson.Query(t, stdout, ".installed_libraries.[1] | .library | .name", `"USBHost"`)

	stdout, _, err = cli.Run("lib", "upgrade")
	require.NoError(t, err)
	// Empty output means nothing has been updated as expected
	require.Empty(t, stdout)
}

func downloadLib(t *testing.T, url string, zipPath *paths.Path) {
	response, err := http.Get(url)
	require.NoError(t, err)
	require.Equal(t, response.StatusCode, 200)
	zip, err := zipPath.Create()
	require.NoError(t, err)
	_, err = io.Copy(zip, response.Body)
	require.NoError(t, err)
	require.NoError(t, response.Body.Close())
	require.NoError(t, zip.Close())
}

func TestInstallGitUrlAndZipPathFlagsVisibility(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	// Verifies installation fail because flags are not found
	gitUrl := "https://github.com/arduino-libraries/WiFi101.git"
	_, stderr, err := cli.Run("lib", "install", "--git-url", gitUrl)
	require.Error(t, err)
	require.Contains(t, string(stderr), "--git-url and --zip-path are disabled by default, for more information see:")

	// Download library
	url := "https://github.com/arduino-libraries/AudioZero/archive/refs/tags/1.1.1.zip"
	zipPath := cli.DownloadDir().Join("libraries", "AudioZero.zip")
	require.NoError(t, zipPath.Parent().MkdirAll())
	downloadLib(t, url, zipPath)

	_, stderr, err = cli.Run("lib", "install", "--zip-path", zipPath.String())
	require.Error(t, err)
	require.Contains(t, string(stderr), "--git-url and --zip-path are disabled by default, for more information see:")

	envVar := cli.GetDefaultEnv()
	envVar["ARDUINO_ENABLE_UNSAFE_LIBRARY_INSTALL"] = "true"
	// Verifies installation is successful when flags are enabled with env var
	stdout, _, err := cli.RunWithCustomEnv(envVar, "lib", "install", "--git-url", gitUrl)
	require.NoError(t, err)
	require.Contains(t, string(stdout), "--git-url and --zip-path flags allow installing untrusted files, use it at your own risk.")

	stdout, _, err = cli.RunWithCustomEnv(envVar, "lib", "install", "--zip-path", zipPath.String())
	require.NoError(t, err)
	require.Contains(t, string(stdout), "--git-url and --zip-path flags allow installing untrusted files, use it at your own risk.")

	// Uninstall libraries to install them again
	_, _, err = cli.Run("lib", "uninstall", "WiFi101", "AudioZero")
	require.NoError(t, err)

	// Verifies installation is successful when flags are enabled with settings file
	_, _, err = cli.RunWithCustomEnv(envVar, "config", "init", "--dest-dir", ".")
	require.NoError(t, err)

	stdout, _, err = cli.Run("lib", "install", "--git-url", gitUrl, "--config-file", "arduino-cli.yaml")
	require.NoError(t, err)
	require.Contains(t, string(stdout), "--git-url and --zip-path flags allow installing untrusted files, use it at your own risk.")

	stdout, _, err = cli.Run("lib", "install", "--zip-path", zipPath.String(), "--config-file", "arduino-cli.yaml")
	require.NoError(t, err)
	require.Contains(t, string(stdout), "--git-url and --zip-path flags allow installing untrusted files, use it at your own risk.")
}

func TestInstallWithZipPath(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	// Initialize configs to enable --zip-path flag
	envVar := cli.GetDefaultEnv()
	envVar["ARDUINO_ENABLE_UNSAFE_LIBRARY_INSTALL"] = "true"
	_, _, err := cli.RunWithCustomEnv(envVar, "config", "init", "--dest-dir", ".")
	require.NoError(t, err)

	// Download a specific lib version
	// Download library
	url := "https://github.com/arduino-libraries/AudioZero/archive/refs/tags/1.1.1.zip"
	zipPath := cli.DownloadDir().Join("libraries", "AudioZero.zip")
	require.NoError(t, zipPath.Parent().MkdirAll())
	downloadLib(t, url, zipPath)

	libInstallDir := cli.SketchbookDir().Join("libraries", "AudioZero")
	// Verifies library is not already installed
	require.NoDirExists(t, libInstallDir.String())

	// Test zip-path install
	stdout, _, err := cli.Run("lib", "install", "--zip-path", zipPath.String(), "--config-file", "arduino-cli.yaml")
	require.NoError(t, err)
	require.Contains(t, string(stdout), "--git-url and --zip-path flags allow installing untrusted files, use it at your own risk.")

	// Verifies library is installed in expected path
	require.DirExists(t, libInstallDir.String())
	files, err := libInstallDir.ReadDirRecursive()
	require.NoError(t, err)
	require.Contains(t, files, libInstallDir.Join("examples", "SimpleAudioPlayerZero", "SimpleAudioPlayerZero.ino"))
	require.Contains(t, files, libInstallDir.Join("src", "AudioZero.h"))
	require.Contains(t, files, libInstallDir.Join("src", "AudioZero.cpp"))
	require.Contains(t, files, libInstallDir.Join("keywords.txt"))
	require.Contains(t, files, libInstallDir.Join("library.properties"))
	require.Contains(t, files, libInstallDir.Join("README.adoc"))

	// Reinstall library
	_, _, err = cli.Run("lib", "install", "--zip-path", zipPath.String(), "--config-file", "arduino-cli.yaml")
	require.NoError(t, err)

	// Verifies library remains installed
	require.DirExists(t, libInstallDir.String())
	files, err = libInstallDir.ReadDirRecursive()
	require.NoError(t, err)
	require.Contains(t, files, libInstallDir.Join("examples", "SimpleAudioPlayerZero", "SimpleAudioPlayerZero.ino"))
	require.Contains(t, files, libInstallDir.Join("src", "AudioZero.h"))
	require.Contains(t, files, libInstallDir.Join("src", "AudioZero.cpp"))
	require.Contains(t, files, libInstallDir.Join("keywords.txt"))
	require.Contains(t, files, libInstallDir.Join("library.properties"))
	require.Contains(t, files, libInstallDir.Join("README.adoc"))
}

func TestInstallWithZipPathMultipleLibraries(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	_, _, err := cli.Run("update")
	require.NoError(t, err)

	envVar := cli.GetDefaultEnv()
	envVar["ARDUINO_ENABLE_UNSAFE_LIBRARY_INSTALL"] = "true"

	// Downloads zips to be installed later
	wifiZipPath := cli.DownloadDir().Join("libraries", "WiFi101-0.16.1.zip")
	bleZipPath := cli.DownloadDir().Join("libraries", "ArduinoBLE-1.1.3.zip")
	downloadLib(t, "https://github.com/arduino-libraries/WiFi101/archive/refs/tags/0.16.1.zip", wifiZipPath)
	downloadLib(t, "https://github.com/arduino-libraries/ArduinoBLE/archive/refs/tags/1.1.3.zip", bleZipPath)

	wifiInstallDir := cli.SketchbookDir().Join("libraries", "WiFi101")
	bleInstallDir := cli.SketchbookDir().Join("libraries", "ArduinoBLE")
	// Verifies libraries are not installed
	require.NoDirExists(t, wifiInstallDir.String())
	require.NoDirExists(t, bleInstallDir.String())

	_, _, err = cli.RunWithCustomEnv(envVar, "lib", "install", "--zip-path", wifiZipPath.String(), bleZipPath.String())
	require.NoError(t, err)

	// Verifies libraries are installed
	require.DirExists(t, wifiInstallDir.String())
	require.DirExists(t, bleInstallDir.String())
}

func TestInstallWithGitUrlLocalFileUri(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Using a file uri as git url doesn't work on Windows, " +
			"this must be removed when this issue is fixed: https://github.com/go-git/go-git/issues/247")
	}

	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	envVar := cli.GetDefaultEnv()
	envVar["ARDUINO_ENABLE_UNSAFE_LIBRARY_INSTALL"] = "true"

	libInstallDir := cli.SketchbookDir().Join("libraries", "WiFi101")
	// Verifies library is not installed
	require.NoDirExists(t, libInstallDir.String())

	// Clone repository locally
	gitUrl := "https://github.com/arduino-libraries/WiFi101.git"
	repoDir := cli.SketchbookDir().Join("WiFi101")
	_, err := git.PlainClone(repoDir.String(), false, &git.CloneOptions{
		URL: gitUrl,
	})
	require.NoError(t, err)

	_, _, err = cli.RunWithCustomEnv(envVar, "lib", "install", "--git-url", "file://"+repoDir.String())
	require.NoError(t, err)

	// Verifies library is installed
	require.DirExists(t, libInstallDir.String())
}

func TestLibQueryParameters(t *testing.T) {
	// This test should not use shared download directory because it needs to download the libraries from scratch
	env := integrationtest.NewEnvironment(t)
	cli := integrationtest.NewArduinoCliWithinEnvironment(env, &integrationtest.ArduinoCLIConfig{
		ArduinoCLIPath: integrationtest.FindArduinoCLIPath(t),
	})
	defer env.CleanUp()

	// Updates index for cores and libraries
	_, _, err := cli.Run("core", "update-index")
	require.NoError(t, err)
	_, _, err = cli.Run("lib", "update-index")
	require.NoError(t, err)

	// Check query=install when a library is installed
	stdout, _, err := cli.Run("lib", "install", "USBHost@1.0.0", "-v", "--log-level", "debug", "--log-format", "json")
	require.NoError(t, err)
	require.Contains(t, string(stdout), `"url":"https://downloads.arduino.cc/libraries/github.com/arduino-libraries/USBHost-1.0.0.zip?query=install"`)

	// Check query=upgrade when a library is upgraded
	stdout, _, err = cli.Run("lib", "upgrade", "USBHost", "-v", "--log-level", "debug", "--log-format", "json")
	require.NoError(t, err)
	require.Contains(t, string(stdout), `"url":"https://downloads.arduino.cc/libraries/github.com/arduino-libraries/USBHost-1.0.5.zip?query=upgrade"`)

	// Check query=depends when a library dependency is installed
	stdout, _, err = cli.Run("lib", "deps", "MD_Parola@3.5.5", "--json")
	require.NoError(t, err)
	// determine the version installed as dependency
	MDMAX72XXversion := strings.Trim(requirejson.Parse(t, stdout).Query(`.dependencies[0].version_required`).String(), `"`)

	stdout, _, err = cli.Run("lib", "install", "MD_Parola@3.5.5", "-v", "--log-level", "debug", "--log-format", "json")
	require.NoError(t, err)
	require.Contains(t, string(stdout), `"url":"https://downloads.arduino.cc/libraries/github.com/MajicDesigns/MD_MAX72XX-`+MDMAX72XXversion+`.zip?query=depends"`)

	// Check query=download when a library is downloaded
	stdout, _, err = cli.Run("lib", "download", "WiFi101@0.16.1", "-v", "--log-level", "debug", "--log-format", "json")
	require.NoError(t, err)
	require.Contains(t, string(stdout), `"url":"https://downloads.arduino.cc/libraries/github.com/arduino-libraries/WiFi101-0.16.1.zip?query=download"`)

	// Check query=install-builtin when a library dependency is installed in builtin-directory
	cliEnv := cli.GetDefaultEnv()
	cliEnv["ARDUINO_DIRECTORIES_BUILTIN_LIBRARIES"] = cli.DataDir().Join("libraries").String()
	stdout, _, err = cli.RunWithCustomEnv(cliEnv, "lib", "install", "Firmata@2.5.3", "--install-in-builtin-dir", "-v", "--log-level", "debug", "--log-format", "json")
	require.NoError(t, err)
	require.Contains(t, string(stdout), `"url":"https://downloads.arduino.cc/libraries/github.com/firmata/Firmata-2.5.3.zip?query=install-builtin`)

	// Check query=update-builtin when a library dependency is updated in builtin-directory
	stdout, _, err = cli.RunWithCustomEnv(cliEnv, "lib", "install", "Firmata@2.5.9", "--install-in-builtin-dir", "-v", "--log-level", "debug", "--log-format", "json")
	require.NoError(t, err)
	require.Contains(t, string(stdout), `"url":"https://downloads.arduino.cc/libraries/github.com/firmata/Firmata-2.5.9.zip?query=upgrade-builtin"`)
}

func TestLibBundlesWhenLibWithTheSameNameIsInstalledGlobally(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	// See: https://github.com/arduino/arduino-cli/issues/1566
	_, _, err := cli.Run("core", "install", "arduino:samd@1.8.13")
	require.NoError(t, err)
	{
		stdout, _, err := cli.Run("lib", "list", "--all", "--fqbn", "arduino:samd:mkrzero", "USBHost", "--json")
		require.NoError(t, err)
		j := requirejson.Parse(t, stdout)
		j.Query(`.installed_libraries.[0].library.name`).MustEqual(`"USBHost"`)
		j.Query(`.installed_libraries.[0].library.compatible_with."arduino:samd:mkrzero"`).MustEqual(`true`)
	}
	_, _, err = cli.Run("lib", "install", "USBHost@1.0.5")
	require.NoError(t, err)
	{
		// Check that the architecture-specific library is still listed
		stdout, _, err := cli.Run("lib", "list", "--all", "--fqbn", "arduino:samd:mkrzero", "USBHost", "--json")
		require.NoError(t, err)
		j := requirejson.Parse(t, stdout)
		j.Query(`.installed_libraries.[0].library.name`).MustEqual(`"USBHost"`)
		j.Query(`.installed_libraries.[0].library.compatible_with."arduino:samd:mkrzero"`).MustEqual(`true`)
	}

	// See: https://github.com/arduino/arduino-cli/issues/1656
	{
		_, _, err = cli.Run("core", "update-index", "--additional-urls", "https://arduino.esp8266.com/stable/package_esp8266com_index.json")
		require.NoError(t, err)
		_, _, err = cli.Run("core", "install", "--additional-urls", "https://arduino.esp8266.com/stable/package_esp8266com_index.json", "esp8266:esp8266@3.0.2")
		require.NoError(t, err)
		_, _, err = cli.Run("lib", "install", "ArduinoOTA@1.0.7")
		require.NoError(t, err)
		stdout, _, err := cli.Run("lib", "examples", "--fqbn", "esp8266:esp8266:generic", "ArduinoOTA", "--json")
		require.NoError(t, err)
		requirejson.Parse(t, stdout).Query(`.examples.[].library.examples[0]`).MustContain(`"BasicOTA"`)
		requirejson.Parse(t, stdout).Query(`.examples.[].library.examples[1]`).MustContain(`"OTALeds"`)
	}
}

func TestLibListDoesNotIncludeEmptyLibraries(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	_, _, err := cli.Run("lib", "update-index")
	require.NoError(t, err)

	// Create a library that does not have neither a library.properties nor a ".h" file
	emptyLib := cli.SketchbookDir().Join("libraries", "empty")
	require.NoError(t, emptyLib.MkdirAll())

	// Check that the output of lib list and lib examples is empty
	stdout, _, err := cli.Run("lib", "list", "--json")
	require.NoError(t, err)
	requirejson.Query(t, stdout, `.installed_libraries | length`, `0`)

	stdout, _, err = cli.Run("lib", "examples", "--json")
	require.NoError(t, err)
	requirejson.Query(t, stdout, `.examples | length`, `0`)

	// Create a library with a header
	libWithHeader := cli.SketchbookDir().Join("libraries", "libWithHeader")
	require.NoError(t, libWithHeader.MkdirAll())
	f, err := libWithHeader.Join("libWithHeader.h").Create()
	require.NoError(t, err)
	require.NoError(t, f.Close())

	// Check that the output of lib list and lib examples contains the library
	stdout, _, err = cli.Run("lib", "list", "--json")
	require.NoError(t, err)
	requirejson.Query(t, stdout, `.installed_libraries | length`, `1`)

	stdout, _, err = cli.Run("lib", "examples", "--json")
	require.NoError(t, err)
	requirejson.Query(t, stdout, `.examples | length`, `1`)
}

func TestDependencyResolver(t *testing.T) {
	// See: https://github.com/arduino/arduino-cli/issues/2135

	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	done := make(chan bool)
	go func() {
		_, _, err := cli.Run("lib", "install", "NTPClient_Generic")
		close(done)
		require.Error(t, err)
	}()

	select {
	case <-done:
	case <-time.After(time.Second * 10):
		require.FailNow(t, "The install command didn't complete in the allocated time")
	}
}

func TestDependencyResolverNoOverwrite(t *testing.T) {
	// https://github.com/arduino/arduino-cli/issues/1799
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	_, _, err := cli.Run("lib", "install", "Bounce2@2.53.0")
	require.NoError(t, err)

	out, _, err := cli.Run("lib", "deps", "EncoderTool@2.2.0", "--json")
	require.NoError(t, err)
	outjson := requirejson.Parse(t, out)
	outjson.MustContain(`{
		"dependencies": [
			{
				"name": "Bounce2",
				"version_installed": "2.53"
			},
			{
				"name": "EncoderTool",
				"version_required": "2.2.0"
			}
		]
	}`)
	require.NotEqual(t, outjson.Query("dependencies[0].version_required").String(), `"2.53.0"`)
	require.NotEqual(t, outjson.Query("dependencies[0].version_required").String(), `"2.53"`)

	out, _, err = cli.Run("lib", "deps", "EncoderTool@2.2.0", "--no-overwrite", "--json")
	require.NoError(t, err)
	outjson = requirejson.Parse(t, out)
	outjson.MustContain(`{
		"dependencies": [
			{
				"name": "Bounce2",
				"version_required": "2.53",
				"version_installed": "2.53"
			},
			{
				"name": "EncoderTool",
				"version_required": "2.2.0"
			}
		]
	}`)

	_, _, err = cli.Run("lib", "install", "EncoderTool@2.2.0", "--no-overwrite")
	require.NoError(t, err)
}
