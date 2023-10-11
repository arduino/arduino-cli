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
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"os"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/arduino/arduino-cli/internal/integrationtest"
	"github.com/arduino/go-paths-helper"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/stretchr/testify/require"
	semver "go.bug.st/relaxed-semver"
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
	requirejson.Contains(t, out, `[{"id":"DxCore-dev:megaavr","installed_version":"1.4.10","name":"DxCore"}]`)
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

	_, _, err = cli.Run("core", "install", "arduino:avr@1.8.6")
	require.NoError(t, err)
	out, _, err = cli.Run("core", "search", "avr", "--format", "json")
	require.NoError(t, err)
	requirejson.NotEmpty(t, out)
	// Verify that "installed_version" is set
	requirejson.Contains(t, out, `[{installed_version: "1.8.6"}]`)

	// additional URL
	out, _, err = cli.Run("core", "search", "test_core", "--format", "json", "--additional-urls="+url.String())
	require.NoError(t, err)
	requirejson.Len(t, out, 1)

	// show all versions
	out, _, err = cli.Run("core", "search", "test_core", "--all", "--format", "json", "--additional-urls="+url.String())
	require.NoError(t, err)
	requirejson.Query(t, out, `.[].releases | length`, "3")

	checkPlatformIsInJSONOutput := func(stdout []byte, id, version string) {
		jqquery := fmt.Sprintf(`[{id:"%s", releases:{"%s":{}}}]`, id, version)
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
	_, _, err = cli.Run("core", "install", "test:x86@2.0.0", "--additional-urls="+url.String())
	require.NoError(t, err)

	// list all with no additional urls, ensure the test core won't show up
	stdout, _, err := cli.Run("core", "search")
	require.NoError(t, err)
	var lines [][]string
	for _, v := range strings.Split(strings.TrimSpace(string(stdout)), "\n") {
		lines = append(lines, strings.Fields(strings.TrimSpace(v)))
	}
	// Check the presence of test:x86@2.0.0
	require.Contains(t, lines, []string{"test:x86", "2.0.0", "test_core"})
	numPlatforms := len(lines) - 1

	// same thing in JSON format, also check the number of platforms found is the same
	stdout, _, err = cli.Run("core", "search", "--format", "json")
	require.NoError(t, err)
	requirejson.Contains(t, stdout, `[{"id": "test:x86", "releases": { "2.0.0": {"name":"test_core"}}}]`)
	requirejson.Query(t, stdout, "length", fmt.Sprint(numPlatforms))

	// list all with additional urls, check the test core is there
	stdout, _, err = cli.Run("core", "search", "--additional-urls="+url.String())
	require.NoError(t, err)
	lines = nil
	for _, v := range strings.Split(strings.TrimSpace(string(stdout)), "\n") {
		lines = append(lines, strings.Fields(strings.TrimSpace(v)))
	}
	// Check the presence of test:x86@3.0.0
	require.Contains(t, lines, []string{"test:x86", "3.0.0", "test_core"})
	numPlatforms = len(lines) - 1

	// same thing in JSON format, also check the number of platforms found is the same
	stdout, _, err = cli.Run("core", "search", "--format", "json", "--additional-urls="+url.String())
	require.NoError(t, err)
	requirejson.Contains(t, stdout, `[{"id": "test:x86", "releases": { "3.0.0": {"name":"test_core"}}}]`)
	requirejson.Query(t, stdout, `[.[].releases | length] | add`, fmt.Sprint(numPlatforms))
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

func TestCoreInstallEsp32(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	// update index
	url := "https://raw.githubusercontent.com/espressif/arduino-esp32/gh-pages/package_esp32_index.json"
	_, _, err := cli.Run("core", "update-index", "--additional-urls="+url)
	require.NoError(t, err)
	// install 3rd-party core
	_, _, err = cli.Run("core", "install", "esp32:esp32@2.0.0", "--additional-urls="+url)
	require.NoError(t, err)
	// create a sketch and compile to double check the core was successfully installed
	sketchName := "test_core_install_esp32"
	sketchPath := cli.SketchbookDir().Join(sketchName)
	_, _, err = cli.Run("sketch", "new", sketchPath.String())
	require.NoError(t, err)
	_, _, err = cli.Run("compile", "-b", "esp32:esp32:esp32", sketchPath.String())
	require.NoError(t, err)
	// prevent regressions for https://github.com/arduino/arduino-cli/issues/163
	md5 := md5.Sum(([]byte(sketchPath.String())))
	sketchPathMd5 := strings.ToUpper(hex.EncodeToString(md5[:]))
	require.NotEmpty(t, sketchPathMd5)
	buildDir := paths.TempDir().Join("arduino", "sketches", sketchPathMd5)
	require.FileExists(t, buildDir.Join(sketchName+".ino.partitions.bin").String())
}

func TestCoreDownload(t *testing.T) {
	env := integrationtest.NewEnvironment(t)
	defer env.CleanUp()

	cli := integrationtest.NewArduinoCliWithinEnvironment(env, &integrationtest.ArduinoCLIConfig{
		ArduinoCLIPath:         integrationtest.FindArduinoCLIPath(t),
		UseSharedStagingFolder: false,
	})

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
	requirejson.Query(t, stdout, `.[] | select(.id == "arduino:avr") | .installed_version`, `"1.6.16"`)

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
	requirejson.Query(t, stdout, `.[] | select(.id == "arduino:avr") | .installed_version`, `"1.6.17"`)

	// Confirm core is listed as "updatable"
	stdout, _, err = cli.Run("core", "list", "--updatable", "--format", "json")
	require.NoError(t, err)
	jsonout := requirejson.Parse(t, stdout)
	q := jsonout.Query(`.[] | select(.id == "arduino:avr")`)
	q.Query(".installed_version").MustEqual(`"1.6.17"`)
	latest := q.Query(".latest_version")

	// Upgrade the core to latest version
	_, _, err = cli.Run("core", "upgrade", "arduino:avr")
	require.NoError(t, err)
	stdout, _, err = cli.Run("core", "list", "--format", "json")
	require.NoError(t, err)
	requirejson.Query(t, stdout, `.[] | select(.id == "arduino:avr") | .installed_version`, latest.String())

	// double check the core isn't updatable anymore
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
	requirejson.Contains(t, stdout, `[ { "id": "arduino:avr" } ]`)
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

	url := "https://raw.githubusercontent.com/arduino/arduino-cli/master/internal/integrationtest/testdata/test_index.json"
	_, _, err := cli.Run("core", "update-index", "--additional-urls="+url)
	require.NoError(t, err)

	// Install a core and check if malicious content has been extracted.
	_, _, err = cli.Run("core", "install", "zipslip:x86", "--additional-urls="+url)
	require.Error(t, err)
	require.NoFileExists(t, "/tmp/evil.txt")
}

func TestCoreBrokenInstall(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	url := "https://raw.githubusercontent.com/arduino/arduino-cli/master/internal/integrationtest/testdata/test_index.json"
	_, _, err := cli.Run("core", "update-index", "--additional-urls="+url)
	require.NoError(t, err)
	_, _, err = cli.Run("core", "install", "brokenchecksum:x86", "--additional-urls="+url)
	require.Error(t, err)
}

func TestCoreUpdateWithLocalUrl(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	wd, _ := paths.Getwd()
	testIndex := wd.Parent().Join("testdata", "test_index.json").String()
	if runtime.GOOS == "windows" {
		testIndex = "/" + strings.ReplaceAll(testIndex, "\\", "/")
	}

	stdout, _, err := cli.Run("core", "update-index", "--additional-urls=file://"+testIndex)
	require.NoError(t, err)
	require.Contains(t, string(stdout), "Downloading index: test_index.json downloaded")
}

func TestCoreSearchManuallyInstalledCoresNotPrinted(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	_, _, err := cli.Run("core", "update-index")
	require.NoError(t, err)

	// Verifies only cores in board manager are shown
	stdout, _, err := cli.Run("core", "search", "--format", "json")
	require.NoError(t, err)
	requirejson.NotEmpty(t, stdout)
	oldJson := stdout

	// Manually installs a core in sketchbooks hardware folder
	gitUrl := "https://github.com/arduino/ArduinoCore-avr.git"
	repoDir := cli.SketchbookDir().Join("hardware", "arduino-beta-development", "avr")
	_, err = git.PlainClone(repoDir.String(), false, &git.CloneOptions{
		URL:           gitUrl,
		ReferenceName: plumbing.NewTagReferenceName("1.8.3"),
	})
	require.NoError(t, err)

	// Verifies manually installed core is not shown
	stdout, _, err = cli.Run("core", "search", "--format", "json")
	require.NoError(t, err)
	requirejson.NotContains(t, stdout, `[{"id": "arduino-beta-development:avr"}]`)
	require.Equal(t, oldJson, stdout)
}

func TestCoreListAllManuallyInstalledCore(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	_, _, err := cli.Run("core", "update-index")
	require.NoError(t, err)

	// Verifies only cores in board manager are shown
	stdout, _, err := cli.Run("core", "list", "--all", "--format", "json")
	require.NoError(t, err)
	requirejson.NotEmpty(t, stdout)
	length, err := strconv.Atoi(requirejson.Parse(t, stdout).Query("length").String())
	require.NoError(t, err)

	// Manually installs a core in sketchbooks hardware folder
	gitUrl := "https://github.com/arduino/ArduinoCore-avr.git"
	repoDir := cli.SketchbookDir().Join("hardware", "arduino-beta-development", "avr")
	_, err = git.PlainClone(repoDir.String(), false, &git.CloneOptions{
		URL:           gitUrl,
		ReferenceName: plumbing.NewTagReferenceName("1.8.3"),
	})
	require.NoError(t, err)

	// Verifies manually installed core is shown
	stdout, _, err = cli.Run("core", "list", "--all", "--format", "json")
	require.NoError(t, err)
	requirejson.Len(t, stdout, length+1)
	requirejson.Contains(t, stdout, `[
		{
			"id": "arduino-beta-development:avr",
			"latest_version": "1.8.3",
			"releases": {
				"1.8.3": {
					"name": "Arduino AVR Boards"
				}
			}
		}
	]`)
}

func TestCoreListUpdatableAllFlags(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	_, _, err := cli.Run("core", "update-index")
	require.NoError(t, err)

	// Verifies only cores in board manager are shown
	stdout, _, err := cli.Run("core", "list", "--all", "--updatable", "--format", "json")
	require.NoError(t, err)
	requirejson.NotEmpty(t, stdout)
	length, err := strconv.Atoi(requirejson.Parse(t, stdout).Query("length").String())
	require.NoError(t, err)

	// Manually installs a core in sketchbooks hardware folder
	gitUrl := "https://github.com/arduino/ArduinoCore-avr.git"
	repoDir := cli.SketchbookDir().Join("hardware", "arduino-beta-development", "avr")
	_, err = git.PlainClone(repoDir.String(), false, &git.CloneOptions{
		URL:           gitUrl,
		ReferenceName: plumbing.NewTagReferenceName("1.8.3"),
	})
	require.NoError(t, err)

	// Verifies using both --updatable and --all flags --all takes precedence
	stdout, _, err = cli.Run("core", "list", "--all", "--updatable", "--format", "json")
	require.NoError(t, err)
	requirejson.Len(t, stdout, length+1)
	requirejson.Contains(t, stdout, `[
		{
			"id": "arduino-beta-development:avr",
			"latest_version": "1.8.3",
			"releases": {
				"1.8.3": {
					"name": "Arduino AVR Boards"
				}
			}
		}
	]`)
}

func TestCoreUpgradeRemovesUnusedTools(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	_, _, err := cli.Run("core", "update-index")
	require.NoError(t, err)

	// Installs a core
	_, _, err = cli.Run("core", "install", "arduino:avr@1.8.2")
	require.NoError(t, err)

	// Verifies expected tool is installed
	toolPath := cli.DataDir().Join("packages", "arduino", "tools", "avr-gcc", "7.3.0-atmel3.6.1-arduino5")
	require.DirExists(t, toolPath.String())

	// Upgrades core
	_, _, err = cli.Run("core", "upgrade", "arduino:avr")
	require.NoError(t, err)

	// Verifies tool is uninstalled since it's not used by newer core version
	require.NoDirExists(t, toolPath.String())
}

func TestCoreInstallRemovesUnusedTools(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	_, _, err := cli.Run("core", "update-index")
	require.NoError(t, err)

	// Installs a core
	_, _, err = cli.Run("core", "install", "arduino:avr@1.8.2")
	require.NoError(t, err)

	// Verifies expected tool is installed
	toolPath := cli.DataDir().Join("packages", "arduino", "tools", "avr-gcc", "7.3.0-atmel3.6.1-arduino5")
	require.DirExists(t, toolPath.String())

	// Installs newer version of already installed core
	_, _, err = cli.Run("core", "install", "arduino:avr@1.8.3")
	require.NoError(t, err)

	// Verifies tool is uninstalled since it's not used by newer core version
	require.NoDirExists(t, toolPath.String())
}

func TestCoreListWithInstalledJson(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	_, _, err := cli.Run("update")
	require.NoError(t, err)

	// Install core
	url := "https://adafruit.github.io/arduino-board-index/package_adafruit_index.json"
	_, _, err = cli.Run("core", "update-index", "--additional-urls="+url)
	require.NoError(t, err)
	_, _, err = cli.Run("core", "install", "adafruit:avr@1.4.13", "--additional-urls="+url)
	require.NoError(t, err)

	// Verifies installed core is correctly found and name is set
	stdout, _, err := cli.Run("core", "list", "--format", "json")
	require.NoError(t, err)
	requirejson.Len(t, stdout, 1)
	requirejson.Contains(t, stdout, `[
		{
			"id": "adafruit:avr",
			"releases": {
				"1.4.13": {
					"name": "Adafruit AVR Boards"
				}
			}
		}
	]`)

	// Deletes installed.json file, this file stores information about the core,
	// that is used mostly when removing package indexes and their cores are still installed;
	// this way we don't lose much information about it.
	// It might happen that the user has old cores installed before the addition of
	// the installed.json file so we need to handle those cases.
	installedJson := cli.DataDir().Join("packages", "adafruit", "hardware", "avr", "1.4.13", "installed.json")
	require.NoError(t, installedJson.Remove())

	// Verifies installed core is still found and name is set
	stdout, _, err = cli.Run("core", "list", "--format", "json")
	require.NoError(t, err)
	requirejson.Len(t, stdout, 1)
	// Name for this core changes since if there's installed.json file we read it from
	// platform.txt, turns out that this core has different names used in different files
	// thus the change.
	requirejson.Contains(t, stdout, `[
		{
			"id": "adafruit:avr",
			"releases": {
				"1.4.13": {
					"name": "Adafruit Boards"
				}
			}
		}
	]`)
}

func TestCoreSearchUpdateIndexDelay(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	// Verifies index update is not run
	stdout, _, err := cli.Run("core", "search")
	require.NoError(t, err)
	require.Contains(t, string(stdout), "Downloading index")

	// Change edit time of package index file
	indexFile := cli.DataDir().Join("package_index.json")
	date := time.Now().Local().Add(time.Hour * (-25))
	require.NoError(t, os.Chtimes(indexFile.String(), date, date))

	// Verifies index update is run
	stdout, _, err = cli.Run("core", "search")
	require.NoError(t, err)
	require.Contains(t, string(stdout), "Downloading index")

	// Verifies index update is not run again
	stdout, _, err = cli.Run("core", "search")
	require.NoError(t, err)
	require.NotContains(t, string(stdout), "Downloading index")
}

func TestCoreSearchSortedResults(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	// Set up the server to serve our custom index file
	testIndex := paths.New("..", "testdata", "test_index.json")
	url := env.HTTPServeFile(8000, testIndex)

	// update custom index
	_, _, err := cli.Run("core", "update-index", "--additional-urls="+url.String())
	require.NoError(t, err)

	// This is done only to avoid index update output when calling core search
	// since it automatically updates them if they're outdated and it makes it
	// harder to parse the list of cores
	_, _, err = cli.Run("core", "search")
	require.NoError(t, err)

	// list all with additional url specified
	stdout, _, err := cli.Run("core", "search", "--additional-urls="+url.String())
	require.NoError(t, err)

	out := strings.Split(strings.TrimSpace(string(stdout)), "\n")
	var lines, deprecated, notDeprecated [][]string
	for i, v := range out {
		if i > 0 {
			v = strings.Join(strings.Fields(v), " ")
			lines = append(lines, strings.SplitN(v, " ", 3))
		}
	}
	for _, v := range lines {
		if strings.HasPrefix(v[2], "[DEPRECATED]") {
			deprecated = append(deprecated, v)
		} else {
			notDeprecated = append(notDeprecated, v)
		}
	}

	// verify that results are already sorted correctly
	require.True(t, sort.SliceIsSorted(deprecated, func(i, j int) bool {
		return strings.ToLower(deprecated[i][2]) < strings.ToLower(deprecated[j][2])
	}))
	require.True(t, sort.SliceIsSorted(notDeprecated, func(i, j int) bool {
		return strings.ToLower(notDeprecated[i][2]) < strings.ToLower(notDeprecated[j][2])
	}))

	// verify that deprecated platforms are the last ones
	require.Equal(t, lines, append(notDeprecated, deprecated...))

	// test same behaviour with json output
	stdout, _, err = cli.Run("core", "search", "--additional-urls="+url.String(), "--format=json")
	require.NoError(t, err)

	// verify that results are already sorted correctly
	sortedDeprecated := requirejson.Parse(t, stdout).Query(
		"[ .[] | select(.deprecated == true) | .name |=ascii_downcase | .name ] | sort").String()
	notSortedDeprecated := requirejson.Parse(t, stdout).Query(
		"[.[] | select(.deprecated == true) | .name |=ascii_downcase | .name]").String()
	require.Equal(t, sortedDeprecated, notSortedDeprecated)

	sortedNotDeprecated := requirejson.Parse(t, stdout).Query(
		"[ .[] | select(.deprecated != true) | .name |=ascii_downcase | .name ] | sort").String()
	notSortedNotDeprecated := requirejson.Parse(t, stdout).Query(
		"[.[] | select(.deprecated != true) | .name |=ascii_downcase | .name]").String()
	require.Equal(t, sortedNotDeprecated, notSortedNotDeprecated)

	// verify that deprecated platforms are the last ones
	platform := requirejson.Parse(t, stdout).Query(
		"[.[] | .name |=ascii_downcase | .name]").String()
	require.Equal(t, platform, strings.TrimRight(notSortedNotDeprecated, "]")+","+strings.TrimLeft(notSortedDeprecated, "["))
}

func TestCoreListSortedResults(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	// Set up the server to serve our custom index file
	testIndex := paths.New("..", "testdata", "test_index.json")
	url := env.HTTPServeFile(8000, testIndex)

	// update custom index
	_, _, err := cli.Run("core", "update-index", "--additional-urls="+url.String())
	require.NoError(t, err)

	// install some core for testing
	_, _, err = cli.Run("core", "install", "test:x86@2.0.0", "Retrokits-RK002:arm", "Package:x86", "--additional-urls="+url.String())
	require.NoError(t, err)

	// list all with additional url specified
	stdout, _, err := cli.Run("core", "list", "--additional-urls="+url.String())
	require.NoError(t, err)

	out := strings.Split(strings.TrimSpace(string(stdout)), "\n")
	var lines, deprecated, notDeprecated [][]string
	for i, v := range out {
		if i > 0 {
			v = strings.Join(strings.Fields(v), " ")
			lines = append(lines, strings.SplitN(v, " ", 4))
		}
	}
	require.Len(t, lines, 3)
	for _, v := range lines {
		if strings.HasPrefix(v[3], "[DEPRECATED]") {
			deprecated = append(deprecated, v)
		} else {
			notDeprecated = append(notDeprecated, v)
		}
	}

	// verify that results are already sorted correctly
	require.True(t, sort.SliceIsSorted(deprecated, func(i, j int) bool {
		return strings.ToLower(deprecated[i][3]) < strings.ToLower(deprecated[j][3])
	}))
	require.True(t, sort.SliceIsSorted(notDeprecated, func(i, j int) bool {
		return strings.ToLower(notDeprecated[i][3]) < strings.ToLower(notDeprecated[j][3])
	}))

	// verify that deprecated platforms are the last ones
	require.Equal(t, lines, append(notDeprecated, deprecated...))

	// test same behaviour with json output
	stdout, _, err = cli.Run("core", "list", "--additional-urls="+url.String(), "--format=json")
	require.NoError(t, err)
	requirejson.Len(t, stdout, 3)

	// verify that results are already sorted correctly
	sortedDeprecated := requirejson.Parse(t, stdout).Query(
		"[ .[] | select(.deprecated == true) | .name |=ascii_downcase | .name ] | sort").String()
	notSortedDeprecated := requirejson.Parse(t, stdout).Query(
		"[.[] | select(.deprecated == true) | .name |=ascii_downcase | .name]").String()
	require.Equal(t, sortedDeprecated, notSortedDeprecated)

	sortedNotDeprecated := requirejson.Parse(t, stdout).Query(
		"[ .[] | select(.deprecated != true) | .name |=ascii_downcase | .name ] | sort").String()
	notSortedNotDeprecated := requirejson.Parse(t, stdout).Query(
		"[.[] | select(.deprecated != true) | .name |=ascii_downcase | .name]").String()
	require.Equal(t, sortedNotDeprecated, notSortedNotDeprecated)

	// verify that deprecated platforms are the last ones
	platform := requirejson.Parse(t, stdout).Query(
		"[.[] | .name |=ascii_downcase | .name]").String()
	require.Equal(t, platform, strings.TrimRight(notSortedNotDeprecated, "]")+","+strings.TrimLeft(notSortedDeprecated, "["))
}

func TestCoreListDeprecatedPlatformWithInstalledJson(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	// Set up the server to serve our custom index file
	testIndex := paths.New("..", "testdata", "test_index.json")
	url := env.HTTPServeFile(8000, testIndex)

	// update custom index
	_, _, err := cli.Run("core", "update-index", "--additional-urls="+url.String())
	require.NoError(t, err)

	// install some core for testing
	_, _, err = cli.Run("core", "install", "Package:x86", "--additional-urls="+url.String())
	require.NoError(t, err)

	installedJsonFile := cli.DataDir().Join("packages", "Package", "hardware", "x86", "1.2.3", "installed.json")
	installedJsonData, err := installedJsonFile.ReadFile()
	require.NoError(t, err)

	installedJson := requirejson.Parse(t, installedJsonData)
	updatedInstalledJsonData := installedJson.Query(`del( .packages[0].platforms[0].deprecated )`).String()
	require.NoError(t, installedJsonFile.WriteFile([]byte(updatedInstalledJsonData)))

	// test same behaviour with json output
	stdout, _, err := cli.Run("core", "list", "--additional-urls="+url.String(), "--format=json")
	require.NoError(t, err)

	requirejson.Len(t, stdout, 1)
	requirejson.Query(t, stdout, ".[] | .deprecated", "true")
}

func TestCoreListPlatformWithoutPlatformTxt(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	_, _, err := cli.Run("update")
	require.NoError(t, err)

	stdout, _, err := cli.Run("core", "list", "--format", "json")
	require.NoError(t, err)
	requirejson.Len(t, stdout, 0)

	// Simulates creation of a new core in the sketchbook hardware folder
	// without a platforms.txt
	testBoardsTxt := paths.New("..", "testdata", "boards.local.txt")
	boardsTxt := cli.SketchbookDir().Join("hardware", "some-packager", "some-arch", "boards.txt")
	require.NoError(t, boardsTxt.Parent().MkdirAll())
	require.NoError(t, testBoardsTxt.CopyTo(boardsTxt))

	// Verifies no core is installed
	stdout, _, err = cli.Run("core", "list", "--format", "json")
	require.NoError(t, err)
	requirejson.Len(t, stdout, 1)
	requirejson.Query(t, stdout, ".[] | .id", "\"some-packager:some-arch\"")
	requirejson.Query(t, stdout, ".[] | .name", "\"some-packager-some-arch\"")
}

func TestCoreDownloadMultiplePlatforms(t *testing.T) {
	if runtime.GOOS == "windows" || runtime.GOOS == "darwin" {
		t.Skip("macOS by default is case insensitive https://github.com/actions/virtual-environments/issues/865 ",
			"Windows too is case insensitive",
			"https://stackoverflow.com/questions/7199039/file-paths-in-windows-environment-not-case-sensitive")
	}
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	_, _, err := cli.Run("update")
	require.NoError(t, err)

	// Verifies no core is installed
	stdout, _, err := cli.Run("core", "list", "--format", "json")
	require.NoError(t, err)
	requirejson.Len(t, stdout, 0)

	// Simulates creation of two new cores in the sketchbook hardware folder
	wd, _ := paths.Getwd()
	testBoardsTxt := wd.Parent().Join("testdata", "boards.local.txt")
	boardsTxt := cli.DataDir().Join("packages", "PACKAGER", "hardware", "ARCH", "1.0.0", "boards.txt")
	require.NoError(t, boardsTxt.Parent().MkdirAll())
	require.NoError(t, testBoardsTxt.CopyTo(boardsTxt))

	boardsTxt1 := cli.DataDir().Join("packages", "packager", "hardware", "arch", "1.0.0", "boards.txt")
	require.NoError(t, boardsTxt1.Parent().MkdirAll())
	require.NoError(t, testBoardsTxt.CopyTo(boardsTxt1))

	// Verifies the two cores are detected
	stdout, _, err = cli.Run("core", "list", "--format", "json")
	require.NoError(t, err)
	requirejson.Len(t, stdout, 2)

	// Try to do an operation on the fake cores.
	// The cli should not allow it since optimizing the casing results in finding two cores
	_, stderr, err := cli.Run("core", "upgrade", "Packager:Arch")
	require.Error(t, err)
	require.Contains(t, string(stderr), "Invalid argument passed: Found 2 platforms matching")
}

func TestCoreWithMissingCustomBoardOptionsIsLoaded(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	// Install platform in Sketchbook hardware dir
	testPlatformName := "platform_with_missing_custom_board_options"
	platformInstallDir := cli.SketchbookDir().Join("hardware", "arduino-beta-dev")
	require.NoError(t, platformInstallDir.MkdirAll())
	require.NoError(t, paths.New("..", "testdata", testPlatformName).CopyDirTo(platformInstallDir.Join(testPlatformName)))

	_, _, err := cli.Run("update")
	require.NoError(t, err)

	stdout, _, err := cli.Run("core", "list", "--format", "json")
	require.NoError(t, err)
	requirejson.Len(t, stdout, 1)
	// Verifies platform is loaded except excluding board with missing options
	requirejson.Contains(t, stdout, `[
		{
			"id": "arduino-beta-dev:platform_with_missing_custom_board_options"
		}
	]`)
	requirejson.Query(t, stdout, ".[] | select(.id == \"arduino-beta-dev:platform_with_missing_custom_board_options\") | .boards | length", "2")
	// Verify board with malformed options is not loaded
	// while other board is loaded
	requirejson.Contains(t, stdout, `[
		{
			"id": "arduino-beta-dev:platform_with_missing_custom_board_options",
			"boards": [
				{
					"fqbn": "arduino-beta-dev:platform_with_missing_custom_board_options:nessuno"
				},
				{
					"fqbn": "arduino-beta-dev:platform_with_missing_custom_board_options:altra"
				}
			]
		}
	]`)
}

func TestCoreListOutdatedCore(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	_, _, err := cli.Run("update")
	require.NoError(t, err)

	// Install an old core version
	_, _, err = cli.Run("core", "install", "arduino:samd@1.8.6")
	require.NoError(t, err)

	stdout, _, err := cli.Run("core", "list", "--format", "json")
	require.NoError(t, err)
	requirejson.Len(t, stdout, 1)
	requirejson.Query(t, stdout, ".[0] | .installed_version", "\"1.8.6\"")
	installedVersion, err := semver.Parse(strings.Trim(requirejson.Parse(t, stdout).Query(".[0] | .installed_version").String(), "\""))
	require.NoError(t, err)
	latestVersion, err := semver.Parse(strings.Trim(requirejson.Parse(t, stdout).Query(".[0] | .latest_version").String(), "\""))
	require.NoError(t, err)
	// Installed version must be older than latest
	require.True(t, installedVersion.LessThan(latestVersion))
}

func TestCoreLoadingPackageManager(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	// Create empty architecture folder (this condition is normally produced by `core uninstall`)
	require.NoError(t, cli.DataDir().Join("packages", "foovendor", "hardware", "fooarch").MkdirAll())

	_, _, err := cli.Run("core", "list", "--all", "--format", "json")
	require.NoError(t, err) // this should not make the cli crash
}

func TestCoreIndexWithoutChecksum(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	_, _, err := cli.Run("config", "init", "--dest-dir", ".")
	require.NoError(t, err)
	url := "https://raw.githubusercontent.com/keyboardio/ArduinoCore-GD32-Keyboardio/ae5938af2f485910729e7d27aa233032a1cb4734/package_gd32_index.json" // noqa: E501
	_, _, err = cli.Run("config", "add", "board_manager.additional_urls", url, "--config-file", "arduino-cli.yaml")
	require.NoError(t, err)

	_, _, err = cli.Run("core", "update-index", "--config-file", "arduino-cli.yaml")
	require.NoError(t, err)
	_, _, err = cli.Run("core", "list", "--all", "--config-file", "arduino-cli.yaml")
	require.NoError(t, err) // this should not make the cli crash
}

func TestCoreInstallCreatesInstalledJson(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	_, _, err := cli.Run("core", "update-index")
	require.NoError(t, err)
	_, _, err = cli.Run("core", "install", "arduino:avr@1.6.23")
	require.NoError(t, err)

	installedJsonFile := cli.DataDir().Join("packages", "arduino", "hardware", "avr", "1.6.23", "installed.json")
	installedJson, err := installedJsonFile.ReadFile()
	require.NoError(t, err)
	installed := requirejson.Parse(t, installedJson, "Parsing installed.json")
	packages := installed.Query(".packages")
	packages.LengthMustEqualTo(1)
	arduinoPackage := packages.Query(".[0]")
	arduinoPackage.Query(".name").MustEqual(`"arduino"`)
	platforms := arduinoPackage.Query(".platforms")
	platforms.LengthMustEqualTo(1)
	avr := platforms.Query(".[0]")
	avr.Query(".name").MustEqual(`"Arduino AVR Boards"`)
	avr.Query(".architecture").MustEqual(`"avr"`)
	tools := arduinoPackage.Query(".tools")
	tools.MustContain(`[
		{ "name": "CMSIS-Atmel" },
		{ "name": "espflash" },
		{ "name": "avrdude" },
		{ "name": "CMSIS" },
		{ "name": "avr-gcc" }
	]`)
}

func TestCoreInstallRunsToolPostInstallScript(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	url := env.HTTPServeFile(8080, paths.New("testdata", "package_with_postinstall_index.json"))
	env.HTTPServeFile(8081, paths.New("testdata", "core_with_postinst.zip"))

	_, _, err := cli.Run("core", "update-index", "--additional-urls", url.String())
	require.NoError(t, err)

	// Checks that the post_install script is correctly skipped on the CI
	stdout, _, err := cli.Run("core", "install", "Test:x86", "--additional-urls", url.String())
	require.NoError(t, err)
	require.Contains(t, string(stdout), "Skipping tool configuration.")
	require.Contains(t, string(stdout), "Skipping platform configuration.")
}

func TestCoreBrokenDependency(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	// Set up an http server to serve our custom index file
	test_index := paths.New("..", "testdata", "test_index.json")
	url := env.HTTPServeFile(8000, test_index)

	// Run update-index with our test index
	_, _, err := cli.Run("core", "update-index", "--additional-urls="+url.String())
	require.NoError(t, err)

	// Check that the download fails and the correct message is displayed
	_, stderr, err := cli.Run("core", "install", "test:x86@3.0.0", "--additional-urls="+url.String())
	require.Error(t, err)
	require.Contains(t, string(stderr), "try contacting test@example.com")
}

func TestCoreUpgradeWarningWithPackageInstalledButNotIndexed(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	url := env.HTTPServeFile(8000, paths.New("..", "testdata", "test_index.json")).String()

	t.Run("missing additional-urls", func(t *testing.T) {
		// update index
		_, _, err := cli.Run("core", "update-index", "--additional-urls="+url)
		require.NoError(t, err)
		// install 3rd-party core outdated version
		_, _, err = cli.Run("core", "install", "test:x86@1.0.0", "--additional-urls="+url)
		require.NoError(t, err)
		// upgrade without index fires a warning
		jsonStdout, _, _ := cli.Run("core", "upgrade", "test:x86", "--format", "json")
		requirejson.Query(t, jsonStdout, ".warnings[]", `"missing package index for test:x86, future updates cannot be guaranteed"`)
	})

	// removing installed.json
	installedJson := cli.DataDir().Join("packages", "test", "hardware", "x86", "1.0.0", "installed.json")
	require.NoError(t, os.Remove(installedJson.String()))

	t.Run("missing both installed.json and additional-urls", func(t *testing.T) {
		jsonStdout, _, _ := cli.Run("core", "upgrade", "test:x86", "--format", "json")
		requirejson.Query(t, jsonStdout, ".warnings[]", `"missing package index for test:x86, future updates cannot be guaranteed"`)
	})
}

func TestCoreListWhenNoPlatformAreInstalled(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	stdout, _, err := cli.Run("core", "list", "--format", "json")
	require.NoError(t, err)
	requirejson.Empty(t, stdout)

	stdout, _, err = cli.Run("core", "list")
	require.NoError(t, err)
	require.Equal(t, "No platforms installed.\n", string(stdout))
}
