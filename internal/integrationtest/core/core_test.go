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

package core_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/arduino/arduino-cli/internal/integrationtest"
	"github.com/arduino/go-paths-helper"
	"github.com/stretchr/testify/require"
	"go.bug.st/testsuite"
	"go.bug.st/testsuite/requirejson"
)

func TestCoreSearch(t *testing.T) {
	env := testsuite.NewEnvironment(t)
	defer env.CleanUp()

	cli := integrationtest.NewArduinoCliWithinEnvironment(env, &integrationtest.ArduinoCLIConfig{
		ArduinoCLIPath:         paths.New("..", "..", "..", "arduino-cli"),
		UseSharedStagingFolder: true,
	})

	// Set up an http server to serve our custom index file
	test_index := paths.New("..", "testdata", "test_index.json")
	url := env.HTTPServeFile(8000, test_index)

	// Run update-index with our test index
	_, _, err := cli.Run("core", "update-index", "--additional-urls="+url.String())
	require.NoError(t, err)

	// Search a specific core
	out, _, err := cli.Run("core", "search", "avr")
	require.NoError(t, err)
	require.Greater(t, len(strings.Split(string(out), "\n")), 2)

	out, _, err = cli.Run("core", "search", "avr", "--format", "json")
	require.NoError(t, err)
	requirejson.NotEmpty(t, out)

	// additional URL
	out, _, err = cli.Run("core", "search", "test_core", "--format", "json", "--additional-urls="+url.String())
	require.NoError(t, err)
	requirejson.Len(t, out, 1)

	// show all versions
	out, _, err = cli.Run("core", "search", "test_core", "--all", "--format", "json", "--additional-urls="+url.String())
	require.NoError(t, err)
	requirejson.Len(t, out, 2)
	// requirejson.Len(t, out, 3) // Test failure

	checkPlatformIsInJSONOutput := func(stdout []byte, id, version string) {
		jqquery := fmt.Sprintf(`[{id:"%s", latest:"%s"}]`, id, version)
		requirejson.Contains(t, out, jqquery, "platform %s@%s is missing from the output", id, version)
	}

	// Search all Retrokit platforms
	out, _, err = cli.Run("core", "search", "retrokit", "--all", "--additional-urls="+url.String(), "--format", "json")
	require.NoError(t, err)
	checkPlatformIsInJSONOutput(out, "Retrokits-RK002:arm", "1.0.5")
	checkPlatformIsInJSONOutput(out, "Retrokits-RK002:arm", "1.0.6")
	//checkPlatformIsInJSONOutput(out, "Retrokits-RK002:arm", "1.0.9") // Test failure

	// Search using Retrokit Package Maintainer
	out, _, err = cli.Run("core", "search", "Retrokits-RK002", "--all", "--additional-urls="+url.String(), "--format", "json")
	require.NoError(t, err)
	checkPlatformIsInJSONOutput(out, "Retrokits-RK002:arm", "1.0.5")
	checkPlatformIsInJSONOutput(out, "Retrokits-RK002:arm", "1.0.6")

	// Search using the Retrokit Platform name
	out, _, err = cli.Run("core", "search", "rk002", "--all", "--additional-urls="+url.String(), "--format", "json")
	require.NoError(t, err)
	checkPlatformIsInJSONOutput(out, "Retrokits-RK002:arm", "1.0.5")
	checkPlatformIsInJSONOutput(out, "Retrokits-RK002:arm", "1.0.6")

	// Search using board names
	out, _, err = cli.Run("core", "search", "myboard", "--all", "--additional-urls="+url.String(), "--format", "json")
	require.NoError(t, err)
	checkPlatformIsInJSONOutput(out, "Package:x86", "1.2.3")

	// Check search with case, accents and spaces
	runSearch := func(searchArgs string, expectedIDs ...string) {
		args := []string{"core", "search", "--format", "json"}
		args = append(args, strings.Split(searchArgs, " ")...)
		out, _, err := cli.Run(args...)
		require.NoError(t, err)

		for _, id := range expectedIDs {
			jqquery := fmt.Sprintf(`[{id:"%s"}]`, id)
			requirejson.Contains(t, out, jqquery, "platform %s is missing from the output", id)
		}
	}

	runSearch("mkr1000", "arduino:samd")
	runSearch("mkr 1000", "arduino:samd")

	runSearch("yún", "arduino:avr")
	runSearch("yùn", "arduino:avr")
	runSearch("yun", "arduino:avr")

	runSearch("nano 33", "arduino:samd", "arduino:mbed_nano")
	runSearch("nano ble", "arduino:mbed_nano")
	runSearch("ble", "arduino:mbed_nano")
	runSearch("ble nano", "arduino:mbed_nano")
	runSearch("nano", "arduino:avr", "arduino:megaavr", "arduino:samd", "arduino:mbed_nano")
}
