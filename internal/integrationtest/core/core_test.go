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
	"path/filepath"
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
	out, _, err := cli.Run("core", "list", "--json")
	require.NoError(t, err)
	requirejson.Contains(t, out, `{
		"platforms": [
			{
				"id":"DxCore-dev:megaavr",
				"installed_version":"1.4.10",
				"releases": {
					"1.4.10": {
						"name":"DxCore"
					}
				}
			}
		]
	}`)
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
	out, _, err = cli.Run("core", "search", "avr", "--json")
	require.NoError(t, err)
	requirejson.NotEmpty(t, out)
	// Verify that "installed_version" is set
	requirejson.Contains(t, out, `{"platforms":[{installed_version: "1.8.6"}]}`)

	// additional URL
	out, _, err = cli.Run("core", "search", "test_core", "--json", "--additional-urls="+url.String())
	require.NoError(t, err)
	requirejson.Query(t, out, `.platforms | length`, `1`)

	// show all versions
	out, _, err = cli.Run("core", "search", "test_core", "--all", "--json", "--additional-urls="+url.String())
	require.NoError(t, err)
	requirejson.Query(t, out, `.platforms.[].releases | length`, "3")

	checkPlatformIsInJSONOutput := func(stdout []byte, id, version string) {
		jqquery := fmt.Sprintf(`{"platforms":[{id:"%s", releases:{"%s":{}}}]}`, id, version)
		requirejson.Contains(t, stdout, jqquery, "platform %s@%s is missing from the output", id, version)
	}

	// Search all Retrokit platforms
	out, _, err = cli.Run("core", "search", "retrokit", "--all", "--additional-urls="+url.String(), "--json")
	require.NoError(t, err)
	checkPlatformIsInJSONOutput(out, "Retrokits-RK002:arm", "1.0.5")
	checkPlatformIsInJSONOutput(out, "Retrokits-RK002:arm", "1.0.6")

	// Search using Retrokit Package Maintainer
	out, _, err = cli.Run("core", "search", "Retrokits-RK002", "--all", "--additional-urls="+url.String(), "--json")
	require.NoError(t, err)
	checkPlatformIsInJSONOutput(out, "Retrokits-RK002:arm", "1.0.5")
	checkPlatformIsInJSONOutput(out, "Retrokits-RK002:arm", "1.0.6")

	// Search using the Retrokit Platform name
	out, _, err = cli.Run("core", "search", "rk002", "--all", "--additional-urls="+url.String(), "--json")
	require.NoError(t, err)
	checkPlatformIsInJSONOutput(out, "Retrokits-RK002:arm", "1.0.5")
	checkPlatformIsInJSONOutput(out, "Retrokits-RK002:arm", "1.0.6")

	// Search using board names
	out, _, err = cli.Run("core", "search", "myboard", "--all", "--additional-urls="+url.String(), "--json")
	require.NoError(t, err)
	checkPlatformIsInJSONOutput(out, "Package:x86", "1.2.3")

	runSearch := func(searchArgs string, expectedIDs ...string) {
		args := []string{"core", "search", "--json"}
		args = append(args, strings.Split(searchArgs, " ")...)
		out, _, err := cli.Run(args...)
		require.NoError(t, err)

		for _, id := range expectedIDs {
			jqquery := fmt.Sprintf(`{"platforms":[{id:"%s"}]}`, id)
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

	// Same thing in JSON format, also check the number of platforms found is the same
	stdout, _, err = cli.Run("core", "search", "--json")
	require.NoError(t, err)
	requirejson.Contains(t, stdout, `{"platforms":[{"id": "test:x86", "releases": { "2.0.0": {"name":"test_core"}}}]}`)
	requirejson.Query(t, stdout, ".platforms | length", fmt.Sprint(numPlatforms))

	// list all with additional urls, check the test core is there
	stdout, _, err = cli.Run("core", "search", "--additional-urls="+url.String())
	require.NoError(t, err)
	lines = nil
	for _, v := range strings.Split(strings.TrimSpace(string(stdout)), "\n") {
		lines = append(lines, strings.Fields(strings.TrimSpace(v)))
	}
	// Check the absence of test:x86@3.0.0 because it contains incompatible deps. The latest available should be the 2.0.0
	require.NotContains(t, lines, []string{"test:x86", "3.0.0", "test_core"})
	require.Contains(t, lines, []string{"test:x86", "2.0.0", "test_core"})
	numPlatforms = len(lines) - 1

	// same thing in JSON format, also check the number of platforms found is the same
	stdout, _, err = cli.Run("core", "search", "--json", "--additional-urls="+url.String())
	require.NoError(t, err)
	requirejson.Contains(t, stdout, `{
		"platforms": [
			{
				"id": "test:x86",
				"installed_version": "2.0.0",
				"latest_version": "2.0.0",
				"releases": {
					"1.0.0": {"name":"test_core", "compatible": true},
					"2.0.0": {"name":"test_core", "compatible": true},
					"3.0.0": {"name":"test_core", "compatible": false}
				}
			}
	]}`)
	requirejson.Query(t, stdout, `.platforms | length`, fmt.Sprint(numPlatforms))
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
	stdout, _, err := cli.Run("core", "list", "--json")
	require.NoError(t, err)
	requirejson.Query(t, stdout, `.platforms.[] | select(.id == "arduino:avr") | .installed_version`, `"1.6.16"`)

	// Replace it with the same with --no-overwrite (should NOT fail)
	_, _, err = cli.Run("core", "install", "arduino:avr@1.6.16", "--no-overwrite")
	require.NoError(t, err)

	// Replace it with a more recent one with --no-overwrite (should fail)
	_, _, err = cli.Run("core", "install", "arduino:avr@1.6.17", "--no-overwrite")
	require.Error(t, err)

	// Replace it with a more recent one without --no-overwrite (should succeed)
	_, _, err = cli.Run("core", "install", "arduino:avr@1.6.17")
	require.NoError(t, err)
	stdout, _, err = cli.Run("core", "list", "--json")
	require.NoError(t, err)
	requirejson.Query(t, stdout, `.platforms.[] | select(.id == "arduino:avr") | .installed_version`, `"1.6.17"`)

	// Confirm core is listed as "updatable"
	stdout, _, err = cli.Run("core", "list", "--updatable", "--json")
	require.NoError(t, err)
	jsonout := requirejson.Parse(t, stdout)
	q := jsonout.Query(`.platforms.[] | select(.id == "arduino:avr")`)
	q.Query(".installed_version").MustEqual(`"1.6.17"`)
	latest := q.Query(".latest_version")

	// Upgrade the core to latest version
	_, _, err = cli.Run("core", "upgrade", "arduino:avr")
	require.NoError(t, err)
	stdout, _, err = cli.Run("core", "list", "--json")
	require.NoError(t, err)
	requirejson.Query(t, stdout, `.platforms.[] | select(.id == "arduino:avr") | .installed_version`, latest.String())

	// double check the core isn't updatable anymore
	stdout, _, err = cli.Run("core", "list", "--updatable", "--json")
	require.NoError(t, err)
	requirejson.Query(t, stdout, `.platforms | length`, `0`)
}

func TestCoreUninstall(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	_, _, err := cli.Run("core", "update-index")
	require.NoError(t, err)
	_, _, err = cli.Run("core", "install", "arduino:avr")
	require.NoError(t, err)
	stdout, _, err := cli.Run("core", "list", "--json")
	require.NoError(t, err)
	requirejson.Contains(t, stdout, `{"platforms":[ { "id": "arduino:avr" } ]}`)
	_, _, err = cli.Run("core", "uninstall", "arduino:avr")
	require.NoError(t, err)
	stdout, _, err = cli.Run("core", "list", "--json")
	require.NoError(t, err)
	requirejson.Query(t, stdout, `.platforms | length`, `0`)
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

	stdout, _, err := cli.Run("core", "update-index", "--additional-urls=file://"+testIndex, "--json")
	require.NoError(t, err)
	requirejson.Parse(t, stdout).MustContain(`{"updated_indexes":[{"index_url":"file://` + testIndex + `","status":"skipped"}]}`)
}

func TestCoreSearchManuallyInstalledCoresNotPrinted(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	_, _, err := cli.Run("core", "update-index")
	require.NoError(t, err)

	// Verifies only cores in board manager are shown
	stdout, _, err := cli.Run("core", "search", "--json")
	require.NoError(t, err)
	requirejson.Query(t, stdout, `.platforms | length > 0`, `true`)
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
	stdout, _, err = cli.Run("core", "search", "--json")
	require.NoError(t, err)
	requirejson.NotContains(t, stdout, `{"platforms":[{"id": "arduino-beta-development:avr"}]}`)
	require.Equal(t, oldJson, stdout)
}

func TestCoreListAllManuallyInstalledCore(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	_, _, err := cli.Run("core", "update-index")
	require.NoError(t, err)

	// Verifies only cores in board manager are shown
	stdout, _, err := cli.Run("core", "list", "--all", "--json")
	require.NoError(t, err)
	requirejson.Query(t, stdout, `.platforms | length > 0`, `true`)
	length, err := strconv.Atoi(requirejson.Parse(t, stdout).Query(".platforms | length").String())
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
	stdout, _, err = cli.Run("core", "list", "--all", "--json")
	require.NoError(t, err)
	requirejson.Query(t, stdout, `.platforms | length`, fmt.Sprint(length+1))
	requirejson.Contains(t, stdout, `{"platforms":[
		{
			"id": "arduino-beta-development:avr",
			"latest_version": "1.8.3",
			"releases": {
				"1.8.3": {
					"name": "Arduino AVR Boards"
				}
			}
		}
	]}`)
}

func TestCoreListShowsLatestVersionWhenMultipleReleasesOfAManuallyInstalledCoreArePresent(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	_, _, err := cli.Run("core", "update-index")
	require.NoError(t, err)

	// Verifies only cores in board manager are shown
	stdout, _, err := cli.Run("core", "list", "--all", "--json")
	require.NoError(t, err)
	requirejson.Query(t, stdout, `.platforms | length > 0`, `true`)
	length, err := strconv.Atoi(requirejson.Parse(t, stdout).Query(".platforms | length").String())
	require.NoError(t, err)

	// Manually installs a core in sketchbooks hardware folder
	gitUrl := "https://github.com/arduino/ArduinoCore-avr.git"
	repoDir := cli.SketchbookDir().Join("hardware", "arduino-beta-development", "avr")
	_, err = git.PlainClone(filepath.Join(repoDir.String(), "1.8.3"), false, &git.CloneOptions{
		URL:           gitUrl,
		ReferenceName: plumbing.NewTagReferenceName("1.8.3"),
	})
	require.NoError(t, err)

	tmp := paths.New(t.TempDir(), "1.8.4")
	_, err = git.PlainClone(tmp.String(), false, &git.CloneOptions{
		URL:           gitUrl,
		ReferenceName: plumbing.NewTagReferenceName("1.8.4"),
	})
	require.NoError(t, err)

	err = tmp.Rename(repoDir.Join("1.8.4"))
	require.NoError(t, err)

	// When manually installing 2 releases of the same core, the newest one takes precedence
	stdout, _, err = cli.Run("core", "list", "--all", "--json")
	require.NoError(t, err)
	requirejson.Query(t, stdout, `.platforms | length`, fmt.Sprint(length+1))
	requirejson.Contains(t, stdout, `{"platforms":[
		{
			"id": "arduino-beta-development:avr",
			"latest_version": "1.8.4",
			"installed_version": "1.8.4",
			"releases": {
				"1.8.3": {
					"name": "Arduino AVR Boards"
				},
				"1.8.3": {
					"name": "Arduino AVR Boards"
				}
			}
		}
	]}`)
}

func TestCoreListUpdatableAllFlags(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	_, _, err := cli.Run("core", "update-index")
	require.NoError(t, err)

	// Verifies only cores in board manager are shown
	stdout, _, err := cli.Run("core", "list", "--all", "--updatable", "--json")
	require.NoError(t, err)
	requirejson.Query(t, stdout, `.platforms | length > 0`, `true`)
	length, err := strconv.Atoi(requirejson.Parse(t, stdout).Query(".platforms | length").String())
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
	stdout, _, err = cli.Run("core", "list", "--all", "--updatable", "--json")
	require.NoError(t, err)
	requirejson.Query(t, stdout, `.platforms | length`, fmt.Sprint(length+1))
	requirejson.Contains(t, stdout, `{"platforms":[
		{
			"id": "arduino-beta-development:avr",
			"latest_version": "1.8.3",
			"releases": {
				"1.8.3": {
					"name": "Arduino AVR Boards"
				}
			}
		}
	]}`)
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
	stdout, _, err := cli.Run("core", "list", "--json")
	require.NoError(t, err)
	requirejson.Query(t, stdout, `.platforms | length`, `1`)
	requirejson.Contains(t, stdout, `{"platforms":[
		{
			"id": "adafruit:avr",
			"releases": {
				"1.4.13": {
					"name": "Adafruit AVR Boards"
				}
			}
		}
	]}`)

	// Deletes installed.json file, this file stores information about the core,
	// that is used mostly when removing package indexes and their cores are still installed;
	// this way we don't lose much information about it.
	// It might happen that the user has old cores installed before the addition of
	// the installed.json file so we need to handle those cases.
	installedJson := cli.DataDir().Join("packages", "adafruit", "hardware", "avr", "1.4.13", "installed.json")
	require.NoError(t, installedJson.Remove())

	// Verifies installed core is still found and name is set
	stdout, _, err = cli.Run("core", "list", "--json")
	require.NoError(t, err)
	requirejson.Query(t, stdout, `.platforms | length`, `1`)
	// Name for this core changes since if there's installed.json file we read it from
	// platform.txt, turns out that this core has different names used in different files
	// thus the change.
	requirejson.Contains(t, stdout, `{"platforms":[
		{
			"id": "adafruit:avr",
			"releases": {
				"1.4.13": {
					"name": "Adafruit Boards"
				}
			}
		}
	]}`)
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
			continue
		}
		notDeprecated = append(notDeprecated, v)
	}

	// verify that results are already sorted correctly
	require.True(t, sort.SliceIsSorted(deprecated, func(i, j int) bool {
		return strings.ToLower(deprecated[i][0]) < strings.ToLower(deprecated[j][0])
	}))
	require.True(t, sort.SliceIsSorted(notDeprecated, func(i, j int) bool {
		return strings.ToLower(notDeprecated[i][0]) < strings.ToLower(notDeprecated[j][0])
	}))

	// verify that deprecated platforms are the last ones
	require.Equal(t, lines, append(notDeprecated, deprecated...))

	// test same behaviour with json output
	stdout, _, err = cli.Run("core", "search", "--additional-urls="+url.String(), "--json")
	require.NoError(t, err)

	// verify that results are already sorted correctly
	sortedDeprecated := requirejson.Parse(t, stdout).Query(
		"[ .platforms.[] | select(.deprecated == true) | .id |=ascii_downcase | .id] | sort").String()
	notSortedDeprecated := requirejson.Parse(t, stdout).Query(
		"[ .platforms.[] | select(.deprecated == true) | .id |=ascii_downcase | .id]").String()
	require.Equal(t, sortedDeprecated, notSortedDeprecated)

	sortedNotDeprecated := requirejson.Parse(t, stdout).Query(
		"[ .platforms.[] | select(.deprecated != true) | .id |=ascii_downcase | .id] | sort").String()
	notSortedNotDeprecated := requirejson.Parse(t, stdout).Query(
		"[ .platforms.[] | select(.deprecated != true) | .id |=ascii_downcase | .id]").String()
	require.Equal(t, sortedNotDeprecated, notSortedNotDeprecated)

	// verify that deprecated platforms are the last ones
	platform := requirejson.Parse(t, stdout).Query(
		"[ .platforms.[] | .id |=ascii_downcase | .id]").String()
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
		return strings.ToLower(deprecated[i][0]) < strings.ToLower(deprecated[j][0])
	}))
	require.True(t, sort.SliceIsSorted(notDeprecated, func(i, j int) bool {
		return strings.ToLower(notDeprecated[i][0]) < strings.ToLower(notDeprecated[j][0])
	}))

	// verify that deprecated platforms are the last ones
	require.Equal(t, lines, append(notDeprecated, deprecated...))

	// test same behaviour with json output
	stdout, _, err = cli.Run("core", "list", "--additional-urls="+url.String(), "--json")
	require.NoError(t, err)
	requirejson.Query(t, stdout, `.platforms | length`, `3`)

	// verify that results are already sorted correctly
	sortedDeprecated := requirejson.Parse(t, stdout).Query(
		"[ .platforms.[] | select(.deprecated == true) | .id |=ascii_downcase | .id] | sort").String()
	notSortedDeprecated := requirejson.Parse(t, stdout).Query(
		"[ .platforms.[] | select(.deprecated == true) | .id |=ascii_downcase | .id]").String()
	require.Equal(t, sortedDeprecated, notSortedDeprecated)

	sortedNotDeprecated := requirejson.Parse(t, stdout).Query(
		"[ .platforms.[] | select(.deprecated != true) | .id |=ascii_downcase | .id] | sort").String()
	notSortedNotDeprecated := requirejson.Parse(t, stdout).Query(
		"[ .platforms.[] | select(.deprecated != true) | .id |=ascii_downcase | .id]").String()
	require.Equal(t, sortedNotDeprecated, notSortedNotDeprecated)

	// verify that deprecated platforms are the last ones
	platform := requirejson.Parse(t, stdout).Query("[ .platforms.[] | .id |=ascii_downcase | .id]").String()
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
	stdout, _, err := cli.Run("core", "list", "--additional-urls="+url.String(), "--json")
	require.NoError(t, err)

	requirejson.Query(t, stdout, `.platforms | length`, `1`)
	requirejson.Query(t, stdout, ".platforms.[] | .deprecated", "true")
}

func TestCoreListPlatformWithoutPlatformTxt(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	_, _, err := cli.Run("update")
	require.NoError(t, err)

	stdout, _, err := cli.Run("core", "list", "--json")
	require.NoError(t, err)
	requirejson.Query(t, stdout, `.platforms | length`, `0`)

	// Simulates creation of a new core in the sketchbook hardware folder
	// without a platforms.txt
	testBoardsTxt := paths.New("..", "testdata", "boards.local.txt")
	boardsTxt := cli.SketchbookDir().Join("hardware", "some-packager", "some-arch", "boards.txt")
	require.NoError(t, boardsTxt.Parent().MkdirAll())
	require.NoError(t, testBoardsTxt.CopyTo(boardsTxt))

	// Verifies no core is installed
	stdout, _, err = cli.Run("core", "list", "--json")
	require.NoError(t, err)
	requirejson.Query(t, stdout, `.platforms | length`, `1`)
	requirejson.Query(t, stdout, ".platforms.[] | .id", `"some-packager:some-arch"`)
	requirejson.Query(t, stdout, ".platforms.[] | .releases[.installed_version].name", `"some-packager-some-arch"`)
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
	stdout, _, err := cli.Run("core", "list", "--json")
	require.NoError(t, err)
	requirejson.Query(t, stdout, `.platforms | length`, `0`)

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
	stdout, _, err = cli.Run("core", "list", "--json")
	require.NoError(t, err)
	requirejson.Query(t, stdout, `.platforms | length`, `2`)

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

	stdout, _, err := cli.Run("core", "list", "--json")
	require.NoError(t, err)
	requirejson.Query(t, stdout, `.platforms | length`, `1`)
	// Verifies platform is loaded except excluding board with missing options
	requirejson.Contains(t, stdout, `{"platforms":[{"id": "arduino-beta-dev:platform_with_missing_custom_board_options"}]}`)
	requirejson.Query(t, stdout, ".platforms.[] | select(.id == \"arduino-beta-dev:platform_with_missing_custom_board_options\") | .releases[.installed_version].boards | length", "2")
	// Verify board with malformed options is not loaded
	// while other board is loaded
	requirejson.Contains(t, stdout, `{"platforms":[
		{
			"id": "arduino-beta-dev:platform_with_missing_custom_board_options",
			"releases": {
				"4.2.0": {
					"boards": [
						{
							"fqbn": "arduino-beta-dev:platform_with_missing_custom_board_options:nessuno"
						},
						{
							"fqbn": "arduino-beta-dev:platform_with_missing_custom_board_options:altra"
						}
					]
				}
			}
		}
	]}`)
}

func TestCoreListOutdatedCore(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	_, _, err := cli.Run("update")
	require.NoError(t, err)

	// Install an old core version
	_, _, err = cli.Run("core", "install", "arduino:samd@1.8.6")
	require.NoError(t, err)

	stdout, _, err := cli.Run("core", "list", "--json")
	require.NoError(t, err)
	requirejson.Query(t, stdout, `.platforms | length`, `1`)
	requirejson.Query(t, stdout, ".platforms.[0] | .installed_version", "\"1.8.6\"")
	installedVersion, err := semver.Parse(strings.Trim(requirejson.Parse(t, stdout).Query(".platforms.[0] | .installed_version").String(), "\""))
	require.NoError(t, err)
	latestVersion, err := semver.Parse(strings.Trim(requirejson.Parse(t, stdout).Query(".platforms.[0] | .latest_version").String(), "\""))
	require.NoError(t, err)
	// Installed version must be older than latest
	require.True(t, installedVersion.LessThan(latestVersion))
}

func TestCoreLoadingPackageManager(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	// Create empty architecture folder (this condition is normally produced by `core uninstall`)
	require.NoError(t, cli.DataDir().Join("packages", "foovendor", "hardware", "fooarch").MkdirAll())

	_, _, err := cli.Run("core", "list", "--all", "--json")
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
		jsonStdout, _, _ := cli.Run("core", "upgrade", "test:x86", "--json")
		requirejson.Query(t, jsonStdout, ".warnings[]", `"missing package index for test:x86, future updates cannot be guaranteed"`)
	})

	// removing installed.json
	installedJson := cli.DataDir().Join("packages", "test", "hardware", "x86", "1.0.0", "installed.json")
	require.NoError(t, os.Remove(installedJson.String()))

	t.Run("missing both installed.json and additional-urls", func(t *testing.T) {
		jsonStdout, _, _ := cli.Run("core", "upgrade", "test:x86", "--json")
		requirejson.Query(t, jsonStdout, ".warnings[]", `"missing package index for test:x86, future updates cannot be guaranteed"`)
	})
}

func TestCoreListWhenNoPlatformAreInstalled(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	stdout, _, err := cli.Run("core", "list", "--json")
	require.NoError(t, err)
	requirejson.Query(t, stdout, `.platforms | length`, `0`)

	stdout, _, err = cli.Run("core", "list")
	require.NoError(t, err)
	require.Equal(t, "No platforms installed.\n", string(stdout))
}

func TestCoreHavingIncompatibleDepTools(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	url := env.HTTPServeFile(8000, paths.New("..", "testdata", "test_index.json")).String()
	additionalURLs := "--additional-urls=" + url

	_, _, err := cli.Run("core", "update-index", additionalURLs)
	require.NoError(t, err)

	// the `latest_version` must point to an installable release. In the releases field the latest entry, points to an incompatible version.
	stdout, _, err := cli.Run("core", "list", "--all", "--json", additionalURLs)
	require.NoError(t, err)
	requirejson.Parse(t, stdout).
		Query(`.platforms | .[] | select(.id == "foo_vendor:avr")`).
		MustContain(`{
			"installed_version": "",
			"latest_version": "1.0.1",
			"releases": {
				"1.0.0": {"compatible": true},
				"1.0.1": {"compatible": true},
				"1.0.2": {"compatible": false}
			}
		}`)

	// install latest compatible version
	_, _, err = cli.Run("core", "install", "foo_vendor:avr", additionalURLs)
	require.NoError(t, err)
	stdout, _, err = cli.Run("core", "list", "--all", "--json", additionalURLs)
	require.NoError(t, err)
	requirejson.Parse(t, stdout).
		Query(`.platforms | .[] | select(.id == "foo_vendor:avr")`).
		MustContain(`{
			"latest_version": "1.0.1",
			"installed_version": "1.0.1",
			"releases": {"1.0.1": {"compatible": true}}
		}`)

	// install a specific incompatible version
	_, stderr, err := cli.Run("core", "install", "foo_vendor:avr@1.0.2", additionalURLs)
	require.Error(t, err)
	require.Contains(t, string(stderr), "no versions available for the current OS")

	// install a specific compatible version
	_, _, err = cli.Run("core", "install", "foo_vendor:avr@1.0.0", additionalURLs)
	require.NoError(t, err)
	stdout, _, err = cli.Run("core", "list", "--json", additionalURLs)
	require.NoError(t, err)
	requirejson.Parse(t, stdout).
		Query(`.platforms | .[] | select(.id == "foo_vendor:avr")`).
		MustContain(`{"installed_version": "1.0.0", "releases": {"1.0.0": {"compatible": true}}}`)

	// Lists all updatable cores
	stdout, _, err = cli.Run("core", "list", "--updatable", "--json", additionalURLs)
	require.NoError(t, err)
	requirejson.Parse(t, stdout).
		Query(`.platforms | .[] | select(.id == "foo_vendor:avr")`).
		MustContain(`{"latest_version": "1.0.1", "releases": {"1.0.1": {"compatible": true}}}`)

	// Show outdated cores, must show latest compatible
	stdout, _, err = cli.Run("outdated", "--json", additionalURLs)
	require.NoError(t, err)
	requirejson.Parse(t, stdout).
		Query(`.platforms | .[] | select(.id == "foo_vendor:avr")`).
		MustContain(`{"latest_version": "1.0.1", "releases":{"1.0.1": {"compatible": true}}}`)

	// upgrade to latest compatible (1.0.0 -> 1.0.1)
	_, _, err = cli.Run("core", "upgrade", "foo_vendor:avr", "--json", additionalURLs)
	require.NoError(t, err)
	stdout, _, err = cli.Run("core", "list", "--json", additionalURLs)
	require.NoError(t, err)
	requirejson.Parse(t, stdout).
		Query(`.platforms | .[] | select(.id == "foo_vendor:avr") | .releases[.installed_version]`).
		MustContain(`{"version": "1.0.1", "compatible": true}`)

	// upgrade to latest incompatible not possible (1.0.1 -> 1.0.2)
	_, _, err = cli.Run("core", "upgrade", "foo_vendor:avr", "--json", additionalURLs)
	require.NoError(t, err)
	stdout, _, err = cli.Run("core", "list", "--json", additionalURLs)
	require.NoError(t, err)
	requirejson.Query(t, stdout, `.platforms | .[] | select(.id == "foo_vendor:avr") | .installed_version`, `"1.0.1"`)

	// When no compatible version are found return error
	// When trying to install a platform with no compatible version fails
	_, stderr, err = cli.Run("core", "install", "incompatible_vendor:avr", additionalURLs)
	require.Error(t, err)
	require.Contains(t, string(stderr), "is not available for your OS")

	// Core search
	{
		// core search with and without --all produces the same results.
		stdoutSearchAll, _, err := cli.Run("core", "search", "--all", "--json", additionalURLs)
		require.NoError(t, err)
		stdoutSearch, _, err := cli.Run("core", "search", "--json", additionalURLs)
		require.NoError(t, err)
		require.Equal(t, stdoutSearchAll, stdoutSearch)
		for _, stdout := range [][]byte{stdoutSearchAll, stdoutSearch} {
			requirejson.Parse(t, stdout).
				Query(`.platforms | .[] | select(.id == "foo_vendor:avr")`).
				MustContain(`{
					"latest_version": "1.0.1",
					"releases": {
						"1.0.0": {"compatible": true},
						"1.0.1": {"compatible": true},
						"1.0.2": {"compatible": false}
					}
				}`)
			requirejson.Parse(t, stdout).
				Query(`.platforms | .[] | select(.id == "incompatible_vendor:avr")`).
				MustContain(`{"latest_version": "", "releases": { "1.0.0": {"compatible": false}}}`)
		}
		// In text mode, core search shows `n/a` for core that doesn't have any compatible version
		stdout, _, err := cli.Run("core", "search", additionalURLs)
		require.NoError(t, err)
		var lines [][]string
		for _, v := range strings.Split(strings.TrimSpace(string(stdout)), "\n") {
			lines = append(lines, strings.Fields(strings.TrimSpace(v)))
			if strings.Contains(v, "incompatible_vendor:avr") {
				t.Log(strings.Fields(strings.TrimSpace(v)))
			}
		}
		require.Contains(t, lines, []string{"incompatible_vendor:avr", "n/a", "Incompatible", "Boards"})
	}
}
