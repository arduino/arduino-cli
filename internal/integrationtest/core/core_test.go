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
	"go.bug.st/testifyjson/requirejson"
)

func TestCorrectHandlingOfPlatformVersionProperty(t *testing.T) {
	// See: https://github.com/arduino/arduino-cli/issues/1823
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	// Copy test platform
	testPlatform := paths.New("testdata", "issue_1823", "DxCore-dev")
	require.NoError(t, testPlatform.CopyDirTo(cli.SketchbookDir().Join("hardware", "DxCore-dev")))

	// Trigger problematic call
	out, _, err := cli.Run("core", "list", "--format", "json")
	require.NoError(t, err)
	requirejson.Contains(t, out, `[{"id":"DxCore-dev:megaavr","installed":"1.4.10","name":"DxCore"}]`)
}

func TestCoreSearch(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

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

	checkPlatformIsInJSONOutput := func(stdout []byte, id, version string) {
		jqquery := fmt.Sprintf(`[{id:"%s", latest:"%s"}]`, id, version)
		requirejson.Contains(t, out, jqquery, "platform %s@%s is missing from the output", id, version)
	}

	// Search all Retrokit platforms
	out, _, err = cli.Run("core", "search", "retrokit", "--all", "--additional-urls="+url.String(), "--format", "json")
	require.NoError(t, err)
	checkPlatformIsInJSONOutput(out, "Retrokits-RK002:arm", "1.0.5")
	checkPlatformIsInJSONOutput(out, "Retrokits-RK002:arm", "1.0.6")

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

	// Check search with case, accents and spaces
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

func TestCoreSearchNoArgs(t *testing.T) {
	// This tests `core search` with and without additional URLs in case no args
	// are passed (i.e. all results are shown).

	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	// Set up an http server to serve our custom index file
	testIndex := paths.New("..", "testdata", "test_index.json")
	url := env.HTTPServeFile(8000, testIndex)

	// update custom index and install test core (installed cores affect `core search`)
	_, _, err := cli.Run("core", "update-index", "--additional-urls="+url.String())
	require.NoError(t, err)
	_, _, err = cli.Run("core", "install", "test:x86", "--additional-urls="+url.String())
	require.NoError(t, err)

	// list all with no additional urls, ensure the test core won't show up
	stdout, _, err := cli.Run("core", "search")
	require.NoError(t, err)
	var lines [][]string
	for _, v := range strings.Split(strings.TrimSpace(string(stdout)), "\n") {
		lines = append(lines, strings.Fields(strings.TrimSpace(v)))
	}
	// The header is printed on the first lines
	require.Equal(t, []string{"test:x86", "2.0.0", "test_core"}, lines[19])
	// We use black to format and flake8 to lint .py files but they disagree on certain
	// things like this one, thus we ignore this specific flake8 rule and stand by black
	// opinion.
	// We ignore this specific case because ignoring it globally would probably cause more
	// issue. For more info about the rule see: https://www.flake8rules.com/rules/E203.html
	numPlatforms := len(lines) - 1 // noqa: E203

	// same thing in JSON format, also check the number of platforms found is the same
	stdout, _, err = cli.Run("core", "search", "--format", "json")
	require.NoError(t, err)
	requirejson.Query(t, stdout, " .[] | select(.name == \"test_core\") | . != \"\"", "true")
	requirejson.Query(t, stdout, "length", fmt.Sprint(numPlatforms))

	// list all with additional urls, check the test core is there
	stdout, _, err = cli.Run("core", "search", "--additional-urls="+url.String())
	require.NoError(t, err)
	lines = nil
	for _, v := range strings.Split(strings.TrimSpace(string(stdout)), "\n") {
		lines = append(lines, strings.Fields(strings.TrimSpace(v)))
	}
	// The header is printed on the first lines
	require.Equal(t, []string{"test:x86", "2.0.0", "test_core"}, lines[20])
	// We use black to format and flake8 to lint .py files but they disagree on certain
	// things like this one, thus we ignore this specific flake8 rule and stand by black
	// opinion.
	// We ignore this specific case because ignoring it globally would probably cause more
	// issue. For more info about the rule see: https://www.flake8rules.com/rules/E203.html
	numPlatforms = len(lines) - 1 // noqa: E203

	// same thing in JSON format, also check the number of platforms found is the same
	stdout, _, err = cli.Run("core", "search", "--format", "json", "--additional-urls="+url.String())
	require.NoError(t, err)
	requirejson.Query(t, stdout, " .[] | select(.name == \"test_core\") | . != \"\"", "true")
	requirejson.Query(t, stdout, "length", fmt.Sprint(numPlatforms))
}

func TestCoreUpdateIndexUrlNotFound(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	_, _, err := cli.Run("core", "update-index")
	require.NoError(t, err)

	// Brings up a local server to fake a failure
	url := env.HTTPServeFileError(8000, paths.New("test_index.json"), 404)

	stdout, stderr, err := cli.Run("core", "update-index", "--additional-urls="+url.String())
	require.Error(t, err)
	require.Contains(t, string(stdout), "Downloading index: test_index.json Server responded with: 404 Not Found")
	require.Contains(t, string(stderr), "Some indexes could not be updated.")
}

func TestCoreUpdateIndexInternalServerError(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	_, _, err := cli.Run("core", "update-index")
	require.NoError(t, err)

	// Brings up a local server to fake a failure
	url := env.HTTPServeFileError(8000, paths.New("test_index.json"), 500)

	stdout, _, err := cli.Run("core", "update-index", "--additional-urls="+url.String())
	require.Error(t, err)
	require.Contains(t, string(stdout), "Downloading index: test_index.json Server responded with: 500 Internal Server Error")
}

func TestCoreInstallWithoutUpdateIndex(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	// Missing "core update-index"
	// Download samd core pinned to 1.8.6
	stdout, _, err := cli.Run("core", "install", "arduino:samd@1.8.6")
	require.NoError(t, err)
	require.Contains(t, string(stdout), "Downloading index: package_index.tar.bz2 downloaded")
}

func TestCoreDownload(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	_, _, err := cli.Run("core", "update-index")
	require.NoError(t, err)

	// Download a specific core version
	_, _, err = cli.Run("core", "download", "arduino:avr@1.6.16")
	require.NoError(t, err)
	require.FileExists(t, cli.DownloadDir().Join("packages", "avr-1.6.16.tar.bz2").String())

	// Wrong core version
	_, _, err = cli.Run("core", "download", "arduino:avr@69.42.0")
	require.Error(t, err)

	// Wrong core
	_, _, err = cli.Run("core", "download", "bananas:avr")
	require.Error(t, err)

	// Wrong casing
	_, _, err = cli.Run("core", "download", "Arduino:Samd@1.8.12")
	require.NoError(t, err)
	require.FileExists(t, cli.DownloadDir().Join("packages", "core-ArduinoCore-samd-1.8.12.tar.bz2").String())
}

func TestCoreInstall(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	_, _, err := cli.Run("core", "update-index")
	require.NoError(t, err)

	// Install a specific core version
	_, _, err = cli.Run("core", "install", "arduino:avr@1.6.16")
	require.NoError(t, err)
	stdout, _, err := cli.Run("core", "list", "--format", "json")
	require.NoError(t, err)
	requirejson.Query(t, stdout, ".[] | select(.id == \"arduino:avr\") | .installed==\"1.6.16\"", "true")

	// Replace it with the same with --no-overwrite (should NOT fail)
	_, _, err = cli.Run("core", "install", "arduino:avr@1.6.16", "--no-overwrite")
	require.NoError(t, err)

	// Replace it with a more recent one with --no-overwrite (should fail)
	_, _, err = cli.Run("core", "install", "arduino:avr@1.6.17", "--no-overwrite")
	require.Error(t, err)

	// Replace it with a more recent one without --no-overwrite (should succeed)
	_, _, err = cli.Run("core", "install", "arduino:avr@1.6.17")
	require.NoError(t, err)
	stdout, _, err = cli.Run("core", "list", "--format", "json")
	require.NoError(t, err)
	requirejson.Query(t, stdout, ".[] | select(.id == \"arduino:avr\") | .installed==\"1.6.17\"", "true")

	// Confirm core is listed as "updatable"
	stdout, _, err = cli.Run("core", "list", "--updatable", "--format", "json")
	require.NoError(t, err)
	requirejson.Query(t, stdout, ".[] | select(.id == \"arduino:avr\") | .installed==\"1.6.17\"", "true")

	// Upgrade the core to latest version
	_, _, err = cli.Run("core", "upgrade", "arduino:avr")
	require.NoError(t, err)
	stdout, _, err = cli.Run("core", "list", "--format", "json")
	require.NoError(t, err)
	requirejson.Query(t, stdout, ".[] | select(.id == \"arduino:avr\") | .installed==\"1.6.17\"", "false")
	// double check the code isn't updatable anymore
	stdout, _, err = cli.Run("core", "list", "--updatable", "--format", "json")
	require.NoError(t, err)
	requirejson.Empty(t, stdout)
}

func TestCoreUninstall(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	_, _, err := cli.Run("core", "update-index")
	require.NoError(t, err)
	_, _, err = cli.Run("core", "install", "arduino:avr")
	require.NoError(t, err)
	stdout, _, err := cli.Run("core", "list", "--format", "json")
	require.NoError(t, err)
	requirejson.Query(t, stdout, ".[] | select(.id == \"arduino:avr\") | . != \"\"", "true")
	_, _, err = cli.Run("core", "uninstall", "arduino:avr")
	require.NoError(t, err)
	stdout, _, err = cli.Run("core", "list", "--format", "json")
	require.NoError(t, err)
	requirejson.Empty(t, stdout)
}

func TestCoreUninstallToolDependencyRemoval(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	// These platforms both have a dependency on the arduino:avr-gcc@7.3.0-atmel3.6.1-arduino5 tool
	// arduino:avr@1.8.2 has a dependency on arduino:avrdude@6.3.0-arduino17
	_, _, err := cli.Run("core", "install", "arduino:avr@1.8.2")
	require.NoError(t, err)
	// arduino:megaavr@1.8.4 has a dependency on arduino:avrdude@6.3.0-arduino16
	_, _, err = cli.Run("core", "install", "arduino:megaavr@1.8.4")
	require.NoError(t, err)
	_, _, err = cli.Run("core", "uninstall", "arduino:avr")
	require.NoError(t, err)

	arduinoToolsPath := cli.DataDir().Join("packages", "arduino", "tools")

	avrGccBinariesPath := arduinoToolsPath.Join("avr-gcc", "7.3.0-atmel3.6.1-arduino5", "bin")
	// The tool arduino:avr-gcc@7.3.0-atmel3.6.1-arduino5 that is a dep of another installed platform should remain
	require.True(t, avrGccBinariesPath.Join("avr-gcc").Exist() || avrGccBinariesPath.Join("avr-gcc.exe").Exist())

	avrDudeBinariesPath := arduinoToolsPath.Join("avrdude", "6.3.0-arduino17", "bin")
	// The tool arduino:avrdude@6.3.0-arduino17 that is only a dep of arduino:avr should have been removed
	require.False(t, avrDudeBinariesPath.Join("avrdude").Exist() || avrDudeBinariesPath.Join("avrdude.exe").Exist())
}

func TestCoreZipslip(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	url := "https://raw.githubusercontent.com/arduino/arduino-cli/master/test/testdata/test_index.json"
	_, _, err := cli.Run("core", "update-index", "--additional-urls="+url)
	require.NoError(t, err)

	// Install a core and check if malicious content has been extracted.
	_, _, err = cli.Run("core", "install", "zipslip:x86", "--additional-urls="+url)
	require.Error(t, err)
	require.NoFileExists(t, paths.TempDir().Join("evil.txt").String())
}
