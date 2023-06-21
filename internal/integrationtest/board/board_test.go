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

package board_test

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/arduino/arduino-cli/internal/integrationtest"
	"github.com/stretchr/testify/require"
	semver "go.bug.st/relaxed-semver"
	"go.bug.st/testifyjson/requirejson"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
)

func TestCorrectBoardListOrdering(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	// install two cores, boards must be ordered by package name and platform name
	_, _, err := cli.Run("core", "install", "arduino:sam")
	require.NoError(t, err)
	_, _, err = cli.Run("core", "install", "arduino:avr")
	require.NoError(t, err)
	jsonOut, _, err := cli.Run("board", "listall", "--format", "json")
	require.NoError(t, err)
	requirejson.Query(t, jsonOut, "[.boards[] | .fqbn]", `[
		"arduino:avr:yun",
		"arduino:avr:uno",
		"arduino:avr:unomini",
		"arduino:avr:diecimila",
		"arduino:avr:nano",
		"arduino:avr:mega",
		"arduino:avr:megaADK",
		"arduino:avr:leonardo",
		"arduino:avr:leonardoeth",
		"arduino:avr:micro",
		"arduino:avr:esplora",
		"arduino:avr:mini",
		"arduino:avr:ethernet",
		"arduino:avr:fio",
		"arduino:avr:bt",
		"arduino:avr:LilyPadUSB",
		"arduino:avr:lilypad",
		"arduino:avr:pro",
		"arduino:avr:atmegang",
		"arduino:avr:robotControl",
		"arduino:avr:robotMotor",
		"arduino:avr:gemma",
		"arduino:avr:circuitplay32u4cat",
		"arduino:avr:yunmini",
		"arduino:avr:chiwawa",
		"arduino:avr:one",
		"arduino:avr:unowifi",
		"arduino:sam:arduino_due_x_dbg",
		"arduino:sam:arduino_due_x"
	]`)
}

func TestBoardList(t *testing.T) {
	if os.Getenv("CI") != "" {
		t.Skip("VMs have no serial ports")
	}

	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	_, _, err := cli.Run("core", "update-index")
	require.NoError(t, err)
	stdout, _, err := cli.Run("board", "list", "--format", "json")
	require.NoError(t, err)
	// check is a valid json and contains a list of ports
	requirejson.Parse(t, stdout).
		Query(`[ .[].port | select(.protocol == null or .protocol_label == null) ]`).
		MustBeEmpty()
}

func TestBoardListWithFqbnFilter(t *testing.T) {
	if os.Getenv("CI") != "" {
		t.Skip("VMs have no serial ports")
	}

	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	_, _, err := cli.Run("core", "update-index")
	require.NoError(t, err)
	stdout, _, err := cli.Run("board", "list", "-b", "foo:bar:baz", "--format", "json")
	require.NoError(t, err)
	// this is a bit of a passpartout test, it actually filters the "bluetooth boards" locally
	// but it would succeed even if the filtering wasn't working properly
	// TODO: find a way to simulate connected boards or create a unit test which
	// mocks or initializes multiple components
	requirejson.Parse(t, stdout).
		MustBeEmpty()
}

func TestBoardListWithFqbnFilterInvalid(t *testing.T) {
	if os.Getenv("CI") != "" {
		t.Skip("VMs have no serial ports")
	}

	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	_, stderr, err := cli.Run("board", "list", "-b", "yadayada", "--format", "json")
	require.Error(t, err)
	requirejson.Query(t, stderr, ".error", `"Invalid FQBN: not an FQBN: yadayada"`)
}

func TestBoardListWithInvalidDiscovery(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	_, _, err := cli.Run("core", "update-index")
	require.NoError(t, err)
	_, _, err = cli.Run("board", "list")
	require.NoError(t, err)

	// check that the CLI does not crash if an invalid discovery is installed
	// (for example if the installation fails midway).
	// https://github.com/arduino/arduino-cli/issues/1669
	toolDir := cli.DataDir().Join("packages", "builtin", "tools", "serial-discovery")
	dirsToEmpty, err := toolDir.ReadDir()
	require.NoError(t, err)
	require.Len(t, dirsToEmpty, 1)
	require.NoError(t, dirsToEmpty[0].RemoveAll())
	require.NoError(t, dirsToEmpty[0].MkdirAll())

	_, stderr, err := cli.Run("board", "list")
	require.NoError(t, err)
	require.Contains(t, string(stderr), "builtin:serial-discovery")
}

func TestBoardListall(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	_, _, err := cli.Run("update")
	require.NoError(t, err)
	_, _, err = cli.Run("core", "install", "arduino:avr@1.8.3")
	require.NoError(t, err)

	stdout, _, err := cli.Run("board", "listall", "--format", "json")
	require.NoError(t, err)

	requirejson.Query(t, stdout, ".boards | length", "26")

	requirejson.Contains(t, stdout, `{
		"boards":[
			{
				"name": "Arduino Yún",
				"fqbn": "arduino:avr:yun",
				"platform": {
					"id": "arduino:avr",
					"installed": "1.8.3",
					"name": "Arduino AVR Boards"
				}
			},
			{
				"name": "Arduino Uno",
				"fqbn": "arduino:avr:uno",
				"platform": {
					"id": "arduino:avr",
					"installed": "1.8.3",
					"name": "Arduino AVR Boards"
				}
			}
		]
	}`)

	// Check if the boards' "latest" value is not empty
	requirejson.Parse(t, stdout).
		Query(`[ .boards | .[] | .platform | select(.latest == "") ]`).
		MustBeEmpty()
}

func TestBoardListallWithManuallyInstalledPlatform(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	_, _, err := cli.Run("update")
	require.NoError(t, err)

	// Manually installs a core in sketchbooks hardware folder
	gitUrl := "https://github.com/arduino/ArduinoCore-samd.git"
	repoDir := cli.SketchbookDir().Join("hardware", "arduino-beta-development", "samd")
	_, err = git.PlainClone(repoDir.String(), false, &git.CloneOptions{
		URL:           gitUrl,
		ReferenceName: plumbing.NewTagReferenceName("1.8.11"),
	})
	require.NoError(t, err)

	stdout, _, err := cli.Run("board", "listall", "--format", "json")
	require.NoError(t, err)

	requirejson.Query(t, stdout, ".boards | length", "17")

	requirejson.Contains(t, stdout, `{
		"boards": [
			{
				"name": "Arduino MKR1000",
				"fqbn": "arduino-beta-development:samd:mkr1000",
				"platform": {
				  "id": "arduino-beta-development:samd",
				  "installed": "1.8.11",
				  "latest": "1.8.11",
				  "name": "Arduino SAMD (32-bits ARM Cortex-M0+) Boards",
				}
			},
			{
				"name": "Arduino NANO 33 IoT",
      			"fqbn": "arduino-beta-development:samd:nano_33_iot",
      			"platform": {
        			"id": "arduino-beta-development:samd",
        			"installed": "1.8.11",
        			"latest": "1.8.11",
        			"name": "Arduino SAMD (32-bits ARM Cortex-M0+) Boards"
				}
			}
		]
	}`)
}

func TestBoardDetails(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	_, _, err := cli.Run("core", "update-index")
	require.NoError(t, err)
	// Download samd core pinned to 1.8.6
	_, _, err = cli.Run("core", "install", "arduino:samd@1.8.6")
	require.NoError(t, err)

	// Test board listall with and without showing hidden elements
	stdout, _, err := cli.Run("board", "listall", "MIPS", "--format", "json")
	require.NoError(t, err)
	require.Equal(t, string(stdout), "{}\n")

	stdout, _, err = cli.Run("board", "listall", "MIPS", "-a", "--format", "json")
	require.NoError(t, err)
	requirejson.Contains(t, stdout, `{
		"boards": [
			{
				"name": "Arduino Tian (MIPS Console port)"
			}
		]
	}`)

	stdout, _, err = cli.Run("board", "details", "-b", "arduino:samd:nano_33_iot", "--format", "json")
	require.NoError(t, err)

	requirejson.Contains(t, stdout, `{
		"fqbn": "arduino:samd:nano_33_iot",
		"name": "Arduino NANO 33 IoT",
		"version": "1.8.6",
		"properties_id": "nano_33_iot",
		"official": true,
		"package": {
	  		"maintainer": "Arduino",
	  		"url": "https://downloads.arduino.cc/packages/package_index.tar.bz2",
	  		"website_url": "http://www.arduino.cc/",
	 		 "email": "packages@arduino.cc",
	  		"name": "arduino",
	  		"help": {
				"online": "http://www.arduino.cc/en/Reference/HomePage"
	  		}
		},
		"platform": {
	  		"architecture": "samd",
	  		"category": "Arduino",
	  		"url": "http://downloads.arduino.cc/cores/samd-1.8.6.tar.bz2",
			"archive_filename": "samd-1.8.6.tar.bz2",
			"checksum": "SHA-256:68a4fffa6fe6aa7886aab2e69dff7d3f94c02935bbbeb42de37f692d7daf823b",
			"size": 2980953,
			"name": "Arduino SAMD Boards (32-bits ARM Cortex-M0+)"
		},
		"identification_properties": [
			{
				"properties": {
				"vid": "0x2341",
				"pid": "0x8057"
				}
			},
			{
				"properties": {
				"vid": "0x2341",
				"pid": "0x0057"
				}
			}
		],
		"programmers": [
	  		{
				"platform": "Arduino SAMD Boards (32-bits ARM Cortex-M0+)",
				"id": "edbg",
				"name": "Atmel EDBG"
			},
			{
				"platform": "Arduino SAMD Boards (32-bits ARM Cortex-M0+)",
				"id": "atmel_ice",
				"name": "Atmel-ICE"
			},
			{
				"platform": "Arduino SAMD Boards (32-bits ARM Cortex-M0+)",
				"id": "sam_ice",
				"name": "Atmel SAM-ICE"
			}
		]
	}`)

	// Download samd core pinned to 1.8.8
	_, _, err = cli.Run("core", "install", "arduino:samd@1.8.8")
	require.NoError(t, err)

	stdout, _, err = cli.Run("board", "details", "-b", "arduino:samd:nano_33_iot", "--format", "json")
	require.NoError(t, err)
	requirejson.Contains(t, stdout, `{"debugging_supported": true}`)
}

func TestBoardDetailsNoFlags(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	_, _, err := cli.Run("core", "update-index")
	require.NoError(t, err)
	// Download samd core pinned to 1.8.6
	_, _, err = cli.Run("core", "install", "arduino:samd@1.8.6")
	require.NoError(t, err)
	stdout, stderr, err := cli.Run("board", "details")
	require.Error(t, err)
	require.Contains(t, string(stderr), "Error: required flag(s) \"fqbn\" not set")
	require.Empty(t, stdout)
}

func TestBoardDetailsListProgrammersWithoutFlag(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	_, _, err := cli.Run("core", "update-index")
	require.NoError(t, err)
	// Download samd core pinned to 1.8.6
	_, _, err = cli.Run("core", "install", "arduino:samd@1.8.6")
	require.NoError(t, err)
	stdout, _, err := cli.Run("board", "details", "-b", "arduino:samd:nano_33_iot")
	require.NoError(t, err)
	split := strings.Split(string(stdout), "\n")
	lines := make([][]string, len(split))
	for i, l := range split {
		lines[i] = strings.Fields(l)
	}
	require.Contains(t, lines, []string{"Programmers:", "ID", "Name"})
	require.Contains(t, lines, []string{"edbg", "Atmel", "EDBG"})
	require.Contains(t, lines, []string{"atmel_ice", "Atmel-ICE"})
	require.Contains(t, lines, []string{"sam_ice", "Atmel", "SAM-ICE"})
}

func TestBoardDetailsListProgrammersFlag(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	_, _, err := cli.Run("core", "update-index")
	require.NoError(t, err)
	// Download samd core pinned to 1.8.6
	_, _, err = cli.Run("core", "install", "arduino:samd@1.8.6")
	require.NoError(t, err)
	stdout, _, err := cli.Run("board", "details", "-b", "arduino:samd:nano_33_iot", "--list-programmers")
	require.NoError(t, err)
	lines := strings.Split(string(stdout), "\n")
	for i, l := range lines {
		lines[i] = strings.TrimSpace(l)
	}
	require.Contains(t, lines, "Id        Programmer name")
	require.Contains(t, lines, "edbg      Atmel EDBG")
	require.Contains(t, lines, "atmel_ice Atmel-ICE")
	require.Contains(t, lines, "sam_ice   Atmel SAM-ICE")
}

func TestBoardSearch(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	_, _, err := cli.Run("update")
	require.NoError(t, err)

	stdout, _, err := cli.Run("board", "search", "--format", "json")
	require.NoError(t, err)
	// Verifies boards are returned
	requirejson.NotEmpty(t, stdout)
	// Verifies no board has FQBN set since no platform is installed
	requirejson.Query(t, stdout, "[ .[] | select(.fqbn) ] | length", "0")
	requirejson.Contains(t, stdout, `[
		{"name": "Arduino UNO"},
		{"name": "Arduino Yún"},
		{"name": "Arduino Zero"},
		{"name": "Arduino Nano 33 BLE"},
		{"name": "Arduino Portenta H7"}
	]`)

	// Search in non installed boards
	stdout, _, err = cli.Run("board", "search", "--format", "json", "nano", "33")
	require.NoError(t, err)
	// Verifies boards are returned
	requirejson.NotEmpty(t, stdout)
	// Verifies no board has FQBN set since no platform is installed
	requirejson.Query(t, stdout, "[ .[] | select(.fqbn) ] | length", "0")
	requirejson.Contains(t, stdout, `[
		{"name": "Arduino Nano 33 BLE"},
		{"name": "Arduino Nano 33 IoT"}
	]`)

	// Install a platform from index
	_, _, err = cli.Run("core", "install", "arduino:avr@1.8.3")
	require.NoError(t, err)

	stdout, _, err = cli.Run("board", "search", "--format", "json")
	require.NoError(t, err)
	requirejson.NotEmpty(t, stdout)
	// Verifies some FQBNs are now returned after installing a platform
	requirejson.Query(t, stdout, "[ .[] | select(.fqbn) ] | length", "26")
	requirejson.Contains(t, stdout, `[
		{
			"name": "Arduino Yún",
    		"fqbn": "arduino:avr:yun"
		},
		{
			"name": "Arduino Uno",
    		"fqbn": "arduino:avr:uno"
		}
	]`)

	stdout, _, err = cli.Run("board", "search", "--format", "json", "arduino", "yun")
	require.NoError(t, err)
	requirejson.NotEmpty(t, stdout)
	requirejson.Contains(t, stdout, `[
		{
			"name": "Arduino Yún",
			"fqbn": "arduino:avr:yun"
		}
	]`)

	// Manually installs a core in sketchbooks hardware folder
	gitUrl := "https://github.com/arduino/ArduinoCore-samd.git"
	repoDir := cli.SketchbookDir().Join("hardware", "arduino-beta-development", "samd")
	_, err = git.PlainClone(repoDir.String(), false, &git.CloneOptions{
		URL:           gitUrl,
		ReferenceName: plumbing.NewTagReferenceName("1.8.11"),
	})
	require.NoError(t, err)

	stdout, _, err = cli.Run("board", "search", "--format", "json")
	require.NoError(t, err)
	requirejson.NotEmpty(t, stdout)
	// Verifies some FQBNs are now returned after installing a platform
	requirejson.Query(t, stdout, "[ .[] | select(.fqbn) ] | length", "43")
	requirejson.Contains(t, stdout, `[
		{
			"name": "Arduino Uno",
    		"fqbn": "arduino:avr:uno"
		},
		{
			"name": "Arduino Yún",
    		"fqbn": "arduino:avr:yun"
		},
		{
			"name": "Arduino MKR WiFi 1010",
    		"fqbn": "arduino-beta-development:samd:mkrwifi1010"
		},
		{
			"name": "Arduino MKR1000",
    		"fqbn": "arduino-beta-development:samd:mkr1000"
		},
		{
			"name": "Arduino MKRZERO",
    		"fqbn": "arduino-beta-development:samd:mkrzero"
		},
		{
			"name": "Arduino NANO 33 IoT",
    		"fqbn": "arduino-beta-development:samd:nano_33_iot"
		},
		{
			"fqbn": "arduino-beta-development:samd:arduino_zero_native"
		}
	]`)

	stdout, _, err = cli.Run("board", "search", "--format", "json", "mkr1000")
	require.NoError(t, err)
	requirejson.NotEmpty(t, stdout)
	// Verifies some FQBNs are now returned after installing a platform
	requirejson.Contains(t, stdout, `[
		{
			"name": "Arduino MKR1000",
    		"fqbn": "arduino-beta-development:samd:mkr1000"
		}
	]`)
}

func TestBoardAttach(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	sketchName := "BoardAttach"
	sketchPath := cli.SketchbookDir().Join(sketchName)
	sketchProjectFlie := sketchPath.Join("sketch.yaml")

	// Create a test sketch
	_, _, err := cli.Run("sketch", "new", sketchPath.String())
	require.NoError(t, err)

	{
		stdout, _, err := cli.Run("board", "attach", "-b", "arduino:avr:uno", sketchPath.String(), "--format", "json")
		require.NoError(t, err)
		requirejson.Query(t, stdout, ".fqbn", `"arduino:avr:uno"`)

		yamlData, err := sketchProjectFlie.ReadFile()
		require.NoError(t, err)
		require.Contains(t, string(yamlData), "default_fqbn: arduino:avr:uno")
		require.NotContains(t, string(yamlData), "default_port:")
		require.NotContains(t, string(yamlData), "default_protocol:")
	}
	{
		stdout, _, err := cli.Run("board", "attach", "-p", "/dev/ttyACM0", "-l", "serial", sketchPath.String(), "--format", "json")
		require.NoError(t, err)
		requirejson.Query(t, stdout, ".fqbn", `"arduino:avr:uno"`)
		requirejson.Query(t, stdout, ".port.address", `"/dev/ttyACM0"`)
		requirejson.Query(t, stdout, ".port.protocol", `"serial"`)

		yamlData, err := sketchProjectFlie.ReadFile()
		require.NoError(t, err)
		require.Contains(t, string(yamlData), "default_fqbn: arduino:avr:uno")
		require.Contains(t, string(yamlData), "default_port: /dev/ttyACM0")
		require.Contains(t, string(yamlData), "default_protocol: serial")
	}
	{
		stdout, _, err := cli.Run("board", "attach", "-p", "/dev/ttyACM0", sketchPath.String(), "--format", "json")
		require.NoError(t, err)
		requirejson.Query(t, stdout, ".fqbn", `"arduino:avr:uno"`)
		requirejson.Query(t, stdout, ".port.address", `"/dev/ttyACM0"`)
		requirejson.Query(t, stdout, ".port.protocol", `null`)

		yamlData, err := sketchProjectFlie.ReadFile()
		require.NoError(t, err)
		require.Contains(t, string(yamlData), "default_fqbn: arduino:avr:uno")
		require.Contains(t, string(yamlData), "default_port: /dev/ttyACM0")
		require.NotContains(t, string(yamlData), "default_protocol:")
	}
}

func TestBoardSearchWithOutdatedCore(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	_, _, err := cli.Run("update")
	require.NoError(t, err)

	// Install an old core version
	_, _, err = cli.Run("core", "install", "arduino:samd@1.8.6")
	require.NoError(t, err)

	stdout, _, err := cli.Run("board", "search", "arduino:samd:mkrwifi1010", "--format", "json")
	require.NoError(t, err)
	requirejson.Len(t, stdout, 1)
	var data []map[string]interface{}
	err = json.Unmarshal(stdout, &data)
	require.NoError(t, err)
	board := data[0]
	require.Equal(t, board["name"], "Arduino MKR WiFi 1010")
	require.Equal(t, board["fqbn"], "arduino:samd:mkrwifi1010")
	samdCore := board["platform"].(map[string]interface{})
	require.Equal(t, samdCore["id"], "arduino:samd")
	installedVersion, err := semver.Parse(samdCore["installed"].(string))
	require.NoError(t, err)
	latestVersion, err := semver.Parse(samdCore["latest"].(string))
	require.NoError(t, err)
	// Installed version must be older than latest
	require.True(t, installedVersion.LessThan(latestVersion))
}

func TestBoardListWithFailedBuiltinInstallation(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	_, _, err := cli.Run("core", "update-index")
	require.NoError(t, err)

	// board list to install builtin tools
	_, _, err = cli.Run("board", "list")
	require.NoError(t, err)

	// remove files from serial-discovery directory to simulate a failed installation
	serialDiscovery, err := cli.DataDir().Join("packages", "builtin", "tools", "serial-discovery").ReadDir()
	require.NoError(t, err)
	require.NoError(t, serialDiscovery[0].Join("serial-discovery.exe").Remove())

	// board list should install serial-discovery again
	stdout, stderr, err := cli.Run("board", "list")
	require.NoError(t, err)
	require.Empty(t, stderr)
	require.Contains(t, string(stdout), "Downloading missing tool builtin:serial-discovery")
}
