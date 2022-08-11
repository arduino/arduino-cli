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
	"testing"

	"github.com/arduino/arduino-cli/internal/integrationtest"
	"github.com/arduino/go-paths-helper"
	"github.com/stretchr/testify/require"
	semver "go.bug.st/relaxed-semver"
	"go.bug.st/testsuite"
)

func TestHelp(t *testing.T) {
	env := testsuite.NewEnvironment(t)
	defer env.CleanUp()

	cli := integrationtest.NewArduinoCliWithinEnvironment(env, &integrationtest.ArduinoCLIConfig{
		ArduinoCLIPath:         paths.New("..", "..", "..", "arduino-cli"),
		UseSharedStagingFolder: true,
	})

	// Run help and check the output message
	stdout, stderr, err := cli.Run("help")
	require.NoError(t, err)
	require.Empty(t, stderr)
	require.Contains(t, string(stdout), "Usage")
}

func TestVersion(t *testing.T) {
	env := testsuite.NewEnvironment(t)
	defer env.CleanUp()

	cli := integrationtest.NewArduinoCliWithinEnvironment(env, &integrationtest.ArduinoCLIConfig{
		ArduinoCLIPath:         paths.New("..", "..", "..", "arduino-cli"),
		UseSharedStagingFolder: true,
	})

	// Run version and check the output message
	stdout, stderr, err := cli.Run("version")
	require.NoError(t, err)
	require.Contains(t, string(stdout), "Version:")
	require.Contains(t, string(stdout), "Commit:")
	require.Empty(t, stderr)

	// Checks if "version --format json" has a json as an output
	stdout, _, err = cli.Run("version", "--format", "json")
	require.NoError(t, err)
	var jsonMap map[string]string
	err = json.Unmarshal(stdout, &jsonMap)
	require.NoError(t, err)

	// Checks if Application's value is arduino-cli
	require.Equal(t, jsonMap["Application"], "arduino-cli")

	// Checks if VersionString's value is git-snapshot, nightly or a valid semantic versioning
	switch version := jsonMap["VersionString"]; version {
	case "git-snapshot":
		require.Contains(t, version, "git-snapshot")
	case "nigthly":
		require.Contains(t, version, "nightly")
	default:
		_, err = semver.Parse(version)
		require.NoError(t, err)
	}

	// Checks if Commit's value is not empty
	require.NotEmpty(t, jsonMap["Commit"])
}
