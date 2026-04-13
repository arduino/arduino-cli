// This file is part of arduino-cli.
//
// Copyright 2026 ARDUINO SA (http://www.arduino.cc/)
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

package core_test

import (
	"fmt"
	"testing"

	"github.com/arduino/arduino-cli/internal/arduino/cores/packageindex"
	"github.com/arduino/arduino-cli/internal/integrationtest"
	"github.com/arduino/go-paths-helper"
	"github.com/stretchr/testify/require"
	"go.bug.st/testifyjson/requirejson"
)

func TestPlatformInstalledJsonMigration(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	url := env.HTTPServeFile(8080, paths.New("testdata", "package_with_lib_deps_index.json"))

	cliEnv := cli.GetDefaultEnv()
	cliEnv["ARDUINO_BOARD_MANAGER_ADDITIONAL_URLS"] = url.String()
	_, _, err := cli.RunWithCustomEnv(cliEnv, "core", "update-index")
	require.NoError(t, err)

	// Checks that the library dependencies are correctly resolved and installed, but not the transitive ones
	_, _, err = cli.RunWithCustomEnv(cliEnv, "core", "install", "Test:samd")
	require.NoError(t, err)

	// Replace installed.json with a version that doesn't have the "libraryDependencies" field, to simulate an
	// old version of the file.
	targetInstalledJson := cli.DataDir().Join("packages", "Test", "hardware", "samd", "1.8.14", "installed.json")
	err = paths.New("testdata", "installed_json", "installed.json").CopyTo(targetInstalledJson)
	require.NoError(t, err)

	// Check that the installed.json is not migrated if the original index is not available
	stdout, _, err := cli.Run("core", "update-index", "--log", "--log-level", "debug")
	require.NoError(t, err)
	require.Contains(t, string(stdout), fmt.Sprintf("Migration to version %d is not possible.", packageindex.Version))

	// Check that the installed.json is migrated in the presence of the original index, and that the
	// "libraryDependencies" field is correctly added to the installed.json file
	stdout, _, err = cli.RunWithCustomEnv(cliEnv, "core", "update-index", "--log", "--log-level", "debug")
	require.NoError(t, err)
	require.Contains(t, string(stdout), fmt.Sprintf("is being updated to version %d", packageindex.Version))
	jsonout := requirejson.ParseFromFile(t, targetInstalledJson.String())
	jsonout.Query(".version").MustEqual(fmt.Sprintf("%d", packageindex.Version))
	jsonout.Query(".packages[0].platforms[0]").MustContain(`{
		"libraryDependencies": [
			{
				"name":"ArduinoBearSSL",
				"version":"1.7.5"
			}
		]
	}`)
}
