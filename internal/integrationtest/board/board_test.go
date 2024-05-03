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
	"runtime"
	"strings"
	"testing"

	"github.com/arduino/arduino-cli/internal/integrationtest"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/stretchr/testify/require"
	"go.bug.st/testifyjson/requirejson"
)

func TestCorrectBoardListOrdering(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	// install two cores, boards must be ordered by package name and platform name
	_, _, err := cli.Run("core", "install", "arduino:sam")
	require.NoError(t, err)
	_, _, err = cli.Run("core", "install", "arduino:avr")
	require.NoError(t, err)
	jsonOut, _, err := cli.Run("board", "listall", "--json")
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
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	_, _, err := cli.Run("core", "update-index")
	require.NoError(t, err)

	cli.InstallMockedSerialDiscovery(t)

	stdout, _, err := cli.Run("board", "list", "--json")
	require.NoError(t, err)
	// check is a valid json and contains a list of ports
	requirejson.Parse(t, stdout).
		Query(`[ .detected_ports | .[].port | select(.protocol == null or .protocol_label == null) ]`).
		MustBeEmpty()
}

func TestBoardListMock(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	_, _, err := cli.Run("core", "update-index")
	require.NoError(t, err)

	cli.InstallMockedSerialDiscovery(t)

	stdout, _, err := cli.Run("board", "list", "--json")
	require.NoError(t, err)

	// check is a valid json and contains a list of ports
	requirejson.Contains(t, stdout, `{
		  "detected_ports": [
			{
			  "matching_boards": [
				{
				  "name": "Arduino Yún",
				  "fqbn": "arduino:avr:yun"
				}
			  ],
			  "port": {
				"address": "/dev/ttyCIAO",
				"label": "Mocked Serial port",
				"protocol": "serial",
				"protocol_label": "Serial",
				"properties": {
				  "pid": "0x0041",
				  "serial": "123456",
				  "vid": "0x2341"
				},
				"hardware_id": "123456"
			  }
			}
		  ]
	}`)
}

func TestBoardListWithFqbnFilter(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	_, _, err := cli.Run("core", "update-index")
	require.NoError(t, err)

	cli.InstallMockedSerialDiscovery(t)

	stdout, _, err := cli.Run("board", "list", "-b", "foo:bar:baz", "--json")
	require.NoError(t, err)
	requirejson.Query(t, stdout, `.detected_ports | length`, `0`)
}

func TestBoardListWithFqbnFilterInvalid(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	_, _, err := cli.Run("core", "update-index")
	require.NoError(t, err)

	cli.InstallMockedSerialDiscovery(t)

	_, stderr, err := cli.Run("board", "list", "-b", "yadayada", "--json")
	require.Error(t, err)
	requirejson.Query(t, stderr, ".error", `"Invalid FQBN: not an FQBN: yadayada"`)
}

func TestBoardListall(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	_, _, err := cli.Run("update")
	require.NoError(t, err)
	_, _, err = cli.Run("core", "install", "arduino:avr@1.8.3")
	require.NoError(t, err)

	stdout, _, err := cli.Run("board", "listall", "--json")
	require.NoError(t, err)

	requirejson.Query(t, stdout, ".boards | length", "26")

	requirejson.Contains(t, stdout, `{
		"boards":[
			{
				"name": "Arduino Yún",
				"fqbn": "arduino:avr:yun",
				"platform": {
					"metadata": {
						"id": "arduino:avr"
					},
					"release": {
						"name": "Arduino AVR Boards",
						"version": "1.8.3",
						"installed": true
					}
				}
			},
			{
				"name": "Arduino Uno",
				"fqbn": "arduino:avr:uno",
				"platform": {
					"metadata": {
						"id": "arduino:avr"
					},
					"release": {
						"name": "Arduino AVR Boards",
						"version": "1.8.3",
						"installed": true
					}
				}
			}
		]
	}`)

	// Check if the boards' "version" value is not empty
	requirejson.Parse(t, stdout).
		Query(`[ .boards | .[] | .platform | select(.version == "") ]`).
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

	stdout, _, err := cli.Run("board", "listall", "--json")
	require.NoError(t, err)

	requirejson.Query(t, stdout, ".boards | length", "17")

	requirejson.Contains(t, stdout, `{
		"boards": [
			{
				"name": "Arduino MKR1000",
				"fqbn": "arduino-beta-development:samd:mkr1000",
				"platform": {
					"metadata": {
					  "id": "arduino-beta-development:samd",
					},
					"release": {
						"installed": true,
						"version": "1.8.11",
						"name": "Arduino SAMD (32-bits ARM Cortex-M0+) Boards"
					},
				}
			},
			{
				"name": "Arduino NANO 33 IoT",
      			"fqbn": "arduino-beta-development:samd:nano_33_iot",
      			"platform": {
					"metadata": {
					  "id": "arduino-beta-development:samd",
					},
					"release": {
						"installed": true,
						"version": "1.8.11",
						"name": "Arduino SAMD (32-bits ARM Cortex-M0+) Boards"
					},
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
	// Download samd core pinned to 1.8.13
	_, _, err = cli.Run("core", "install", "arduino:samd@1.8.13")
	require.NoError(t, err)

	// Test board listall with and without showing hidden elements
	stdout, _, err := cli.Run("board", "listall", "MIPS", "--json")
	require.NoError(t, err)
	require.Equal(t, string(stdout), "{}\n")

	stdout, _, err = cli.Run("board", "listall", "MIPS", "-a", "--json")
	require.NoError(t, err)
	requirejson.Contains(t, stdout, `{
		"boards": [
			{
				"name": "Arduino Tian (MIPS Console port)"
			}
		]
	}`)

	stdout, _, err = cli.Run("board", "details", "-b", "arduino:samd:nano_33_iot", "--json")
	require.NoError(t, err)

	requirejson.Contains(t, stdout, `{
		  "fqbn": "arduino:samd:nano_33_iot",
		  "name": "Arduino NANO 33 IoT",
		  "version": "1.8.13",
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
			"url": "http://downloads.arduino.cc/cores/core-ArduinoCore-samd-1.8.13.tar.bz2",
			"archive_filename": "core-ArduinoCore-samd-1.8.13.tar.bz2",
			"checksum": "SHA-256:47d44c80a5fd4ea224eb64fd676169e896caa6856f338d78feb4a12d42b4ea67",
			"size": 3074191,
			"name": "Arduino SAMD Boards (32-bits ARM Cortex-M0+)"
		  },
		  "programmers": [
			{
			  "platform": "Arduino SAMD Boards (32-bits ARM Cortex-M0+)",
			  "id": "jlink",
			  "name": "Segger J-Link"
			},
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
		  ],
		  "identification_properties": [
			{
			  "properties": {
				"pid": "0x8057",
				"vid": "0x2341"
			  }
			},
			{
			  "properties": {
				"pid": "0x0057",
				"vid": "0x2341"
			  }
			},
			{
			  "properties": {
				"board": "nano_33_iot"
			  }
			}
		  ]
		}`)
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

	stdout, _, err := cli.Run("board", "search", "--json")
	require.NoError(t, err)
	// Verifies boards are returned
	requirejson.NotEmpty(t, stdout)
	// Verifies no board has FQBN set since no platform is installed
	requirejson.Query(t, stdout, "[ .boards[] | select(.fqbn) ] | length", "0")
	requirejson.Contains(t, stdout, `{
		"boards": [
				{"name": "Arduino UNO"},
				{"name": "Arduino Yún"},
				{"name": "Arduino Zero"},
				{"name": "Arduino Nano 33 BLE"},
				{"name": "Arduino Portenta H7"}
		]}`)

	// Search in non installed boards
	stdout, _, err = cli.Run("board", "search", "--json", "nano", "33")
	require.NoError(t, err)
	// Verifies boards are returned
	requirejson.NotEmpty(t, stdout)
	// Verifies no board has FQBN set since no platform is installed
	requirejson.Query(t, stdout, "[ .boards[] | select(.fqbn) ] | length", "0")
	requirejson.Contains(t, stdout, `{
		"boards": [
			{"name": "Arduino Nano 33 BLE"},
			{"name": "Arduino Nano 33 IoT"}
		]}`)

	// Install a platform from index
	_, _, err = cli.Run("core", "install", "arduino:avr@1.8.3")
	require.NoError(t, err)

	stdout, _, err = cli.Run("board", "search", "--json")
	require.NoError(t, err)
	requirejson.NotEmpty(t, stdout)
	// Verifies some FQBNs are now returned after installing a platform
	requirejson.Query(t, stdout, "[ .boards[] | select(.fqbn) ] | length", "26")
	requirejson.Contains(t, stdout, `{
		"boards": [
			{
				"name": "Arduino Yún",
				"fqbn": "arduino:avr:yun"
			},
			{
				"name": "Arduino Uno",
				"fqbn": "arduino:avr:uno"
			}
		]}`)

	stdout, _, err = cli.Run("board", "search", "--json", "arduino", "yun")
	require.NoError(t, err)
	requirejson.NotEmpty(t, stdout)
	requirejson.Contains(t, stdout, `{
		"boards": [
			{
				"name": "Arduino Yún",
				"fqbn": "arduino:avr:yun"
			}
		]}`)

	// Manually installs a core in sketchbooks hardware folder
	gitUrl := "https://github.com/arduino/ArduinoCore-samd.git"
	repoDir := cli.SketchbookDir().Join("hardware", "arduino-beta-development", "samd")
	_, err = git.PlainClone(repoDir.String(), false, &git.CloneOptions{
		URL:           gitUrl,
		ReferenceName: plumbing.NewTagReferenceName("1.8.11"),
	})
	require.NoError(t, err)

	stdout, _, err = cli.Run("board", "search", "--json")
	require.NoError(t, err)
	requirejson.NotEmpty(t, stdout)
	// Verifies some FQBNs are now returned after installing a platform
	requirejson.Query(t, stdout, "[ .boards[] | select(.fqbn) ] | length", "43")
	requirejson.Contains(t, stdout, `{
		"boards":
		[
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
		]}`)

	stdout, _, err = cli.Run("board", "search", "--json", "mkr1000")
	require.NoError(t, err)
	requirejson.NotEmpty(t, stdout)
	// Verifies some FQBNs are now returned after installing a platform
	requirejson.Contains(t, stdout, `{
		"boards": [
			{
				"name": "Arduino MKR1000",
				"fqbn": "arduino-beta-development:samd:mkr1000"
			}
		]}`)
}

func TestBoardAttach(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	sketchName := "BoardAttach"
	sketchPath := cli.SketchbookDir().Join(sketchName)
	sketchProjectFile := sketchPath.Join("sketch.yaml")

	// Create a test sketch
	_, _, err := cli.Run("sketch", "new", sketchPath.String())
	require.NoError(t, err)

	{
		stdout, _, err := cli.Run("board", "attach", "-b", "arduino:avr:uno", sketchPath.String(), "--json")
		require.NoError(t, err)
		requirejson.Query(t, stdout, ".fqbn", `"arduino:avr:uno"`)

		yamlData, err := sketchProjectFile.ReadFile()
		require.NoError(t, err)
		require.Contains(t, string(yamlData), "default_fqbn: arduino:avr:uno")
		require.NotContains(t, string(yamlData), "default_port:")
		require.NotContains(t, string(yamlData), "default_protocol:")
	}
	{
		stdout, _, err := cli.Run("board", "attach", "-p", "/dev/ttyACM0", "-l", "serial", sketchPath.String(), "--json")
		require.NoError(t, err)
		requirejson.Query(t, stdout, ".fqbn", `"arduino:avr:uno"`)
		requirejson.Query(t, stdout, ".port.address", `"/dev/ttyACM0"`)
		requirejson.Query(t, stdout, ".port.protocol", `"serial"`)

		yamlData, err := sketchProjectFile.ReadFile()
		require.NoError(t, err)
		require.Contains(t, string(yamlData), "default_fqbn: arduino:avr:uno")
		require.Contains(t, string(yamlData), "default_port: /dev/ttyACM0")
		require.Contains(t, string(yamlData), "default_protocol: serial")
	}
	{
		stdout, _, err := cli.Run("board", "attach", "-p", "/dev/ttyACM0", sketchPath.String(), "--json")
		require.NoError(t, err)
		requirejson.Query(t, stdout, ".fqbn", `"arduino:avr:uno"`)
		requirejson.Query(t, stdout, ".port.address", `"/dev/ttyACM0"`)
		requirejson.Query(t, stdout, ".port.protocol", `null`)

		yamlData, err := sketchProjectFile.ReadFile()
		require.NoError(t, err)
		require.Contains(t, string(yamlData), "default_fqbn: arduino:avr:uno")
		require.Contains(t, string(yamlData), "default_port: /dev/ttyACM0")
		require.NotContains(t, string(yamlData), "default_protocol:")
	}
	{
		stdout, _, err := cli.Run("board", "attach", "-b", "arduino:samd:mkr1000", "-P", "atmel_ice", sketchPath.String(), "--json")
		require.NoError(t, err)
		requirejson.Query(t, stdout, ".fqbn", `"arduino:samd:mkr1000"`)
		requirejson.Query(t, stdout, ".programmer", `"atmel_ice"`)
		requirejson.Query(t, stdout, ".port.address", `"/dev/ttyACM0"`)
		requirejson.Query(t, stdout, ".port.protocol", `null`)

		yamlData, err := sketchProjectFile.ReadFile()
		require.NoError(t, err)
		require.Contains(t, string(yamlData), "default_fqbn: arduino:samd:mkr1000")
		require.Contains(t, string(yamlData), "default_programmer: atmel_ice")
		require.Contains(t, string(yamlData), "default_port: /dev/ttyACM0")
		require.NotContains(t, string(yamlData), "default_protocol:")
	}
}

func TestBoardListWithFailedBuiltinInstallation(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	_, _, err := cli.Run("core", "update-index")
	require.NoError(t, err)

	// board list to install builtin tools
	_, _, err = cli.Run("board", "list")
	require.NoError(t, err)

	ext := ""
	if runtime.GOOS == "windows" {
		ext = ".exe"
	}
	// remove files from serial-discovery directory to simulate a failed installation
	serialDiscovery, err := cli.DataDir().Join("packages", "builtin", "tools", "serial-discovery").ReadDir()
	require.NoError(t, err)
	require.NoError(t, serialDiscovery[0].Join("serial-discovery"+ext).Remove())

	// board list should install serial-discovery again
	stdout, stderr, err := cli.Run("board", "list")
	require.NoError(t, err)
	require.Empty(t, stderr)
	require.Contains(t, string(stdout), "Downloading missing tool builtin:serial-discovery")
}

func TestCLIStartupWithCorruptedInventory(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	_, _, err := cli.Run("core", "update-index")
	require.NoError(t, err)

	f, err := cli.DataDir().Join("inventory.yaml").Append()
	require.NoError(t, err)
	_, err = f.WriteString(`data: '[{"name":"WCH;32?'","fqbn":"esp32:esp32:esp32s3camlcd"}]'`)
	require.NoError(t, err)
	require.NoError(t, f.Close())

	// the CLI should not be able to load inventory and report it to the logs
	_, stderr, err := cli.Run("core", "update-index", "-v")
	require.NoError(t, err)
	require.Contains(t, string(stderr), "Error loading inventory store")
}
