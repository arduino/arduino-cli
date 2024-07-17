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

package main_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/arduino/arduino-cli/internal/integrationtest"
	"github.com/stretchr/testify/require"
	semver "go.bug.st/relaxed-semver"
	"go.bug.st/testifyjson/requirejson"
)

func TestHelp(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	// Run help and check the output message
	stdout, stderr, err := cli.Run("help")
	require.NoError(t, err)
	require.Empty(t, stderr)
	require.Contains(t, string(stdout), "Usage")
}

func TestVersion(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	// Run version and check the output message
	stdout, stderr, err := cli.Run("version")
	require.NoError(t, err)
	require.Contains(t, string(stdout), "Version:")
	require.Contains(t, string(stdout), "Commit:")
	require.Empty(t, string(stderr))

	// Checks if "version --json" has a json as an output
	stdout, _, err = cli.Run("version", "--json")
	require.NoError(t, err)
	var jsonMap map[string]string
	err = json.Unmarshal(stdout, &jsonMap)
	require.NoError(t, err)

	// Checks if Application's value is arduino-cli
	require.Equal(t, jsonMap["Application"], "arduino-cli")

	// Checks if VersionString's value is 1.0.0-snapshot, nightly or a valid semantic versioning
	switch version := jsonMap["VersionString"]; version {
	case "1.0.0-snapshot":
		require.Contains(t, version, "1.0.0-snapshot")
	case "nightly":
		require.Contains(t, version, "nightly")
	default:
		_, err = semver.Parse(version)
		require.NoError(t, err)
	}

	// Checks if Commit's value is not empty
	require.NotEmpty(t, jsonMap["Commit"])
}

func TestLogOptions(t *testing.T) {
	// Using version as a test command
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	// No logs
	stdout, _, err := cli.Run("version")
	require.NoError(t, err)
	trimOut := strings.TrimSpace(string(stdout))
	outLines := strings.Split(trimOut, "\n")
	require.Len(t, outLines, 1)

	// Plain text logs on stdout
	stdout, _, err = cli.Run("version", "-v")
	require.NoError(t, err)
	trimOut = strings.TrimSpace(string(stdout))
	outLines = strings.Split(trimOut, "\n")
	require.Greater(t, len(outLines), 1)
	require.True(t, strings.HasPrefix(outLines[0], "\x1b[36mINFO\x1b[0m")) // account for the colors

	// Plain text logs on file
	logFile := cli.DataDir().Join("log.txt")
	_, _, err = cli.Run("version", "--log-file", logFile.String())
	require.NoError(t, err)
	lines, _ := logFile.ReadFileAsLines()
	require.True(t, strings.HasPrefix(lines[0], "time=\""))

	// json on stdout
	stdout, _, err = cli.Run("version", "-v", "--log-format", "JSON")
	require.NoError(t, err)
	trimOut = strings.TrimSpace(string(stdout))
	outLines = strings.Split(trimOut, "\n")
	requirejson.Contains(t, []byte(outLines[0]), `{ "level" }`)

	// Check if log.json contains readable json in each line
	var v interface{}
	logFileJson := cli.DataDir().Join("log.json")
	_, _, err = cli.Run("version", "--log-format", "JSON", "--log-file", logFileJson.String())
	require.NoError(t, err)
	fileContent, err := logFileJson.ReadFileAsLines()
	require.NoError(t, err)
	for _, line := range fileContent {
		// exclude empty lines since they are not valid json
		if line == "" {
			continue
		}
		err = json.Unmarshal([]byte(line), &v)
		require.NoError(t, err)
	}
}

func TestInventoryCreation(t *testing.T) {
	// Using version as a test command
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	// no logs
	stdout, _, err := cli.Run("version")
	require.NoError(t, err)
	line := strings.TrimSpace(string(stdout))
	outLines := strings.Split(line, "\n")
	require.Len(t, outLines, 1)

	// parse inventory file
	inventoryFile := cli.DataDir().Join("inventory.yaml")
	stream, err := inventoryFile.ReadFile()
	require.NoError(t, err)
	require.True(t, strings.Contains(string(stream), "installation"))
}
