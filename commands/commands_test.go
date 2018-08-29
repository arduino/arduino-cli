/*
 * This file is part of arduino-cli.
 *
 * Copyright 2018 ARDUINO SA (http://www.arduino.cc/)
 *
 * This software is released under the GNU General Public License version 3,
 * which covers the main part of arduino-cli.
 * The terms of this license can be found at:
 * https://www.gnu.org/licenses/gpl-3.0.en.html
 *
 * You can be released from the requirements of the above licenses by purchasing
 * a commercial license. Buying such a license is mandatory if you want to modify or
 * otherwise use the software for commercial activities involving the Arduino
 * software without disclosing the source code of your own applications. To purchase
 * a commercial license, send an email to license@arduino.cc.
 */

package commands_test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/arduino/arduino-cli/commands/root"
	"github.com/arduino/go-paths-helper"
	"github.com/bouk/monkey"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.bug.st/relaxed-semver"
)

// Redirecting stdOut so we can analyze output line by
// line and check with what we want.
var stdOut = os.Stdout // *os.File

type stdOutRedirect struct {
	tempFile *os.File
	t        *testing.T
}

func (grabber *stdOutRedirect) Open(t *testing.T) {
	tempFile, err := ioutil.TempFile(os.TempDir(), "test")
	require.NoError(t, err, "Opening temp output file")
	os.Stdout = tempFile
	grabber.tempFile = tempFile
	grabber.t = t
}

func (grabber *stdOutRedirect) GetOutput() []byte {
	_, err := grabber.tempFile.Seek(0, 0)
	require.NoError(grabber.t, err, "Rewinding temp output file")
	output, err := ioutil.ReadAll(grabber.tempFile)
	require.NoError(grabber.t, err, "Reading temp output file")
	return output
}

func (grabber *stdOutRedirect) Close() {
	grabber.tempFile.Close()
	err := os.Remove(grabber.tempFile.Name())
	assert.NoError(grabber.t, err, "Removing temp output file")
	os.Stdout = stdOut
}

// executeWithArgs executes the Cobra Command with the given arguments
// and intercepts any errors (even `os.Exit()` ones), returning the exit code
func executeWithArgs(t *testing.T, args ...string) (exitCode int, output []byte) {
	fmt.Printf("RUNNING: %s\n", args)

	redirect := &stdOutRedirect{}
	redirect.Open(t)
	defer func() {
		output = redirect.GetOutput()
		redirect.Close()
		fmt.Print(string(output))
		fmt.Println()
	}()

	// Mock the os.Exit function, so that we can use the
	// error result for the test and prevent the test from exiting
	fakeExitFired := false
	fakeExit := func(code int) {
		exitCode = code
		fakeExitFired = true

		// use panic to exit and jump to deferred recover
		panic(fmt.Errorf("os.Exit(%d)", code))
	}
	patch := monkey.Patch(os.Exit, fakeExit)
	defer patch.Unpatch()
	defer func() {
		if fakeExitFired {
			recover()
		}
	}()

	// Execute the CLI command, in this process
	cmd := root.Init()
	cmd.SetArgs(args)
	cmd.Execute()

	return exitCode, output
}

var currDataDir *paths.Path

func makeTempDataDir(t *testing.T) func() {
	tmp, err := paths.MkTempDir("", "test")
	require.NoError(t, err, "making temporary staging dir")
	os.Setenv("ARDUINO_DATA_DIR", tmp.String())
	currDataDir = tmp
	fmt.Printf("ARDUINO_DATA_DIR = %s\n", os.Getenv("ARDUINO_DATA_DIR"))
	return func() {
		os.Unsetenv("ARDUINO_DATA_DIR")
		currDataDir = nil
		tmp.RemoveAll()
		fmt.Printf("ARDUINO_DATA_DIR = %s\n", os.Getenv("ARDUINO_DATA_DIR"))
	}
}

var currSketchbookDir *paths.Path

func makeTempSketchbookDir(t *testing.T) func() {
	tmp, err := paths.MkTempDir("", "test")
	require.NoError(t, err, "making temporary staging dir")
	os.Setenv("ARDUINO_SKETCHBOOK_DIR", tmp.String())
	currSketchbookDir = tmp
	fmt.Printf("ARDUINO_SKETCHBOOK_DIR = %s\n", os.Getenv("ARDUINO_SKETCHBOOK_DIR"))
	return func() {
		os.Unsetenv("ARDUINO_SKETCHBOOK_DIR")
		currSketchbookDir = nil
		tmp.RemoveAll()
		fmt.Printf("ARDUINO_SKETCHBOOK_DIR = %s\n", os.Getenv("ARDUINO_SKETCHBOOK_DIR"))
	}
}

// END -- Utility functions

func TestLibSearch(t *testing.T) {
	defer makeTempDataDir(t)()
	defer makeTempSketchbookDir(t)()

	exitCode, output := executeWithArgs(t, "lib", "search", "audiozer", "--format", "json")
	require.Zero(t, exitCode, "process exit code")
	var res struct {
		Libraries []struct {
			Name string
		}
	}
	err := json.Unmarshal(output, &res)
	require.NoError(t, err, "decoding json output")
	require.NotNil(t, res.Libraries)
	require.Len(t, res.Libraries, 1)
	require.Equal(t, res.Libraries[0].Name, "AudioZero")

	exitCode, output = executeWithArgs(t, "lib", "search", "audiozero", "--names")
	require.Zero(t, exitCode, "process exit code")
	require.Equal(t, "AudioZero\n", string(output))

	exitCode, output = executeWithArgs(t, "lib", "search", "audiozer", "--names")
	require.Zero(t, exitCode, "process exit code")
	require.Equal(t, "AudioZero\n", string(output))

	exitCode, output = executeWithArgs(t, "lib", "search", "audiozerooooo", "--names")
	require.Zero(t, exitCode, "process exit code")
	require.Equal(t, "", string(output))
}

func TestUserLibs(t *testing.T) {
	defer makeTempDataDir(t)()
	defer makeTempSketchbookDir(t)()
	libDir := currSketchbookDir.Join("libraries")
	err := libDir.MkdirAll()
	require.NoError(t, err, "creating 'sketchbook/libraries' dir")

	installLib := func(lib string) {
		libPath := paths.New("testdata/" + lib)
		fmt.Printf("COPYING: %s in %s\n", libPath, libDir)
		err = libPath.CopyDirTo(libDir.Join(lib))
		require.NoError(t, err, "copying "+lib+" in sketchbook")
	}

	// List libraries (valid libs)
	installLib("MyLib")
	exitCode, d := executeWithArgs(t, "lib", "list")
	require.Zero(t, exitCode, "exit code")
	require.Contains(t, string(d), "MyLib")
	require.Contains(t, string(d), "1.0.5")

	// List libraries (pre-1.5 format)
	installLib("MyLibPre15")
	exitCode, d = executeWithArgs(t, "lib", "list")
	require.Zero(t, exitCode, "exit code")
	require.Contains(t, string(d), "MyLibPre15")

	// List libraries (invalid version lib)
	installLib("MyLibWithWrongVersion")
	exitCode, d = executeWithArgs(t, "lib", "list")
	require.Zero(t, exitCode, "exit code")
	require.Contains(t, string(d), "MyLibWithWrongVersion")
}

func TestLibDownloadAndInstall(t *testing.T) {
	defer makeTempDataDir(t)()
	defer makeTempSketchbookDir(t)()
	var d []byte
	var exitCode int

	exitCode, _ = executeWithArgs(t, "core", "update-index")
	require.Zero(t, exitCode, "exit code")

	// Download inexistent
	exitCode, d = executeWithArgs(t, "lib", "download", "inexistentLibrary", "--format", "json")
	require.NotZero(t, exitCode, "exit code")
	require.Contains(t, string(d), "library inexistentLibrary not found")

	exitCode, d = executeWithArgs(t, "lib", "download", "inexistentLibrary")
	require.NotZero(t, exitCode, "exit code")
	require.Contains(t, string(d), "library inexistentLibrary not found")

	// Download latest
	exitCode, d = executeWithArgs(t, "lib", "download", "Audio")
	require.Zero(t, exitCode, "exit code")
	require.Contains(t, string(d), "Audio@")
	require.Contains(t, string(d), "downloaded")

	// Download non existent version
	exitCode, d = executeWithArgs(t, "lib", "download", "Audio@1.2.3-nonexistent")
	require.NotZero(t, exitCode, "exit code")
	require.Contains(t, string(d), "not found")

	// Install latest
	exitCode, d = executeWithArgs(t, "lib", "install", "Audio")
	require.Zero(t, exitCode, "exit code")
	require.Contains(t, string(d), "Audio@")
	require.Contains(t, string(d), "Installed")

	exitCode, d = executeWithArgs(t, "lib", "list")
	require.Zero(t, exitCode, "exit code")
	require.Contains(t, string(d), "Audio")

	// Already installed
	exitCode, d = executeWithArgs(t, "lib", "install", "Audio")
	require.NotZero(t, exitCode, "exit code")
	require.Contains(t, string(d), "Audio@")
	require.Contains(t, string(d), "already installed")

	// Install another version
	exitCode, d = executeWithArgs(t, "lib", "install", "Audio@1.0.4")
	require.Zero(t, exitCode, "exit code")
	require.Contains(t, string(d), "Audio@1.0.4")
	require.Contains(t, string(d), "Installed")
	exitCode, d = executeWithArgs(t, "lib", "list")
	require.Zero(t, exitCode, "exit code")
	require.Contains(t, string(d), "Audio")
	require.Contains(t, string(d), "1.0.4")

	// List updatable
	exitCode, d = executeWithArgs(t, "lib", "list", "--updatable")
	require.Zero(t, exitCode, "exit code")
	require.Contains(t, string(d), "Audio")
	require.Contains(t, string(d), "1.0.4")
	require.Contains(t, string(d), "1.0.5")

	// Uninstall version not installed
	exitCode, d = executeWithArgs(t, "lib", "uninstall", "Audio@1.0.3")
	require.NotZero(t, exitCode, "exit code")
	require.Contains(t, string(d), "Audio@1.0.3")
	require.Contains(t, string(d), "not installed")

	// Upgrade libraries
	exitCode, d = executeWithArgs(t, "lib", "upgrade")
	require.Zero(t, exitCode, "exit code")
	require.Contains(t, string(d), "Installed")
	require.Contains(t, string(d), "Audio")
	require.Contains(t, string(d), "1.0.5")

	// Uninstall (without version)
	exitCode, d = executeWithArgs(t, "lib", "uninstall", "Audio")
	require.Zero(t, exitCode, "exit code")
	require.Contains(t, string(d), "Uninstalling")
	require.Contains(t, string(d), "Audio")
	require.Contains(t, string(d), "1.0.5")
	exitCode, d = executeWithArgs(t, "lib", "list")
	require.Zero(t, exitCode, "exit code")
	require.NotContains(t, string(d), "Audio")

	// Uninstall (with version)
	exitCode, d = executeWithArgs(t, "lib", "install", "Audio@1.0.4")
	require.Zero(t, exitCode, "exit code")
	require.Contains(t, string(d), "Audio@1.0.4")
	require.Contains(t, string(d), "Installed")
	exitCode, d = executeWithArgs(t, "lib", "list")
	require.Zero(t, exitCode, "exit code")
	require.Contains(t, string(d), "Audio")
	require.Contains(t, string(d), "1.0.4")
	exitCode, d = executeWithArgs(t, "lib", "uninstall", "Audio@1.0.4")
	require.Zero(t, exitCode, "exit code")
	require.Contains(t, string(d), "Uninstalling")
	require.Contains(t, string(d), "Audio")
	require.Contains(t, string(d), "1.0.4")
	exitCode, d = executeWithArgs(t, "lib", "list")
	require.Zero(t, exitCode, "exit code")
	require.NotContains(t, string(d), "Audio")
}

func updateCoreIndex(t *testing.T) {
	// run a "core update-index" to download the package_index.json
	exitCode, _ := executeWithArgs(t, "core", "update-index")
	require.Equal(t, 0, exitCode, "exit code")
}

func detectLatestAVRCore(t *testing.T) string {
	jsonFile := currDataDir.Join("package_index.json")
	type index struct {
		Packages []struct {
			Name      string
			Platforms []struct {
				Architecture string
				Version      string
			}
		}
	}
	var jsonIndex index
	jsonData, err := jsonFile.ReadFile()
	require.NoError(t, err, "reading package_index.json")
	err = json.Unmarshal(jsonData, &jsonIndex)
	require.NoError(t, err, "parsing package_index.json")
	latest := semver.MustParse("0.0.1")
	for _, p := range jsonIndex.Packages {
		if p.Name == "arduino" {
			for _, pl := range p.Platforms {
				ver, err := semver.Parse(pl.Version)
				require.NoError(t, err, "version parsing")
				if pl.Architecture == "avr" && ver.GreaterThan(latest) {
					latest = ver
				}
			}
			break
		}
	}
	require.NotEmpty(t, latest, "latest avr core version")
	fmt.Println("Latest AVR core version:", latest)
	return latest.String()
}

func TestCoreCommands(t *testing.T) {
	defer makeTempDataDir(t)()
	defer makeTempSketchbookDir(t)()

	// Set staging dir to a temporary dir
	tmp, err := ioutil.TempDir(os.TempDir(), "test")
	require.NoError(t, err, "making temporary staging dir")
	defer os.RemoveAll(tmp)

	updateCoreIndex(t)
	AVR := "arduino:avr@" + detectLatestAVRCore(t)

	// Download a specific core version
	exitCode, d := executeWithArgs(t, "core", "download", "arduino:avr@1.6.16")
	require.Zero(t, exitCode, "exit code")
	require.Contains(t, string(d), "arduino:avr-gcc@4.9.2-atmel3.5.3-arduino2 downloaded")
	require.Contains(t, string(d), "arduino:avrdude@6.3.0-arduino8 downloaded")
	require.Contains(t, string(d), "arduino:arduinoOTA@1.0.0 downloaded")
	require.Contains(t, string(d), "arduino:avr@1.6.16 downloaded")

	// Download latest
	exitCode, d = executeWithArgs(t, "core", "download", "arduino:avr")
	require.Zero(t, exitCode, "exit code")
	require.Contains(t, string(d), AVR+" downloaded")

	// Wrong downloads
	exitCode, d = executeWithArgs(t, "core", "download", "arduino:samd@1.2.3-notexisting")
	require.NotZero(t, exitCode, "exit code")
	require.Contains(t, string(d), "required version 1.2.3-notexisting not found for platform arduino:samd")

	exitCode, d = executeWithArgs(t, "core", "download", "arduino:notexistent")
	require.NotZero(t, exitCode, "exit code")
	require.Contains(t, string(d), "not found")

	exitCode, d = executeWithArgs(t, "core", "download", "wrongparameter")
	require.NotZero(t, exitCode, "exit code")
	require.Contains(t, string(d), "invalid item")

	// Empty cores list
	exitCode, d = executeWithArgs(t, "core", "list")
	require.Zero(t, exitCode, "exit code")
	require.Empty(t, strings.TrimSpace(string(d)))

	// Install avr 1.6.16
	exitCode, d = executeWithArgs(t, "core", "install", "arduino:avr@1.6.16")
	require.Zero(t, exitCode, "exit code")
	require.Contains(t, string(d), "arduino:avr@1.6.16 installed")

	exitCode, d = executeWithArgs(t, "core", "list")
	require.Zero(t, exitCode, "exit code")
	require.Contains(t, string(d), "arduino:avr")
	require.Contains(t, string(d), "1.6.16")

	// Replace avr with 1.6.17
	exitCode, d = executeWithArgs(t, "core", "install", "arduino:avr@1.6.17")
	require.Zero(t, exitCode, "exit code")
	require.Contains(t, string(d), "Updating arduino:avr@1.6.16 with arduino:avr@1.6.17")
	require.Contains(t, string(d), "arduino:avr@1.6.17 installed")

	exitCode, d = executeWithArgs(t, "core", "list")
	require.Zero(t, exitCode, "exit code")
	require.Contains(t, string(d), "arduino:avr")
	require.Contains(t, string(d), "1.6.17")

	// List updatable cores
	exitCode, d = executeWithArgs(t, "core", "list", "--updatable")
	require.Zero(t, exitCode, "exit code")
	require.Contains(t, string(d), "arduino:avr")

	// Upgrade platform
	exitCode, d = executeWithArgs(t, "core", "upgrade", "arduino:avr@1.6.18")
	require.NotZero(t, exitCode, "exit code")
	require.Contains(t, string(d), "Invalid item arduino:avr@1.6.18")

	exitCode, d = executeWithArgs(t, "core", "upgrade", "other:avr")
	require.NotZero(t, exitCode, "exit code")
	require.Contains(t, string(d), "other:avr not found")

	exitCode, d = executeWithArgs(t, "core", "upgrade", "arduino:samd")
	require.NotZero(t, exitCode, "exit code")
	require.Contains(t, string(d), "arduino:samd is not installed")

	exitCode, d = executeWithArgs(t, "core", "upgrade", "arduino:avr")
	require.Zero(t, exitCode, "exit code")
	require.Contains(t, string(d), "Updating arduino:avr@1.6.17 with "+AVR)

	// List updatable cores
	exitCode, d = executeWithArgs(t, "core", "list", "--updatable")
	require.Zero(t, exitCode, "exit code")
	require.NotContains(t, string(d), "arduino:avr")

	exitCode, d = executeWithArgs(t, "core", "list")
	require.Zero(t, exitCode, "exit code")
	require.Contains(t, string(d), "arduino:avr")

	// Build sketch for arduino:avr:uno
	exitCode, d = executeWithArgs(t, "sketch", "new", "Test1")
	require.Zero(t, exitCode, "exit code")
	require.Contains(t, string(d), "Sketch created")

	exitCode, d = executeWithArgs(t, "compile", "-b", "arduino:avr:uno", currSketchbookDir.Join("Test1").String())
	require.Zero(t, exitCode, "exit code")
	require.Contains(t, string(d), "Sketch uses")

	// Uninstall arduino:avr
	exitCode, d = executeWithArgs(t, "core", "uninstall", "arduino:avr")
	require.Zero(t, exitCode, "exit code")
	require.Contains(t, string(d), AVR+" uninstalled")

	// Empty cores list
	exitCode, d = executeWithArgs(t, "core", "list")
	require.Zero(t, exitCode, "exit code")
	require.Empty(t, strings.TrimSpace(string(d)))
}
