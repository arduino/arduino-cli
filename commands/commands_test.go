/*
 * This file is part of arduino-cli.
 *
 * arduino-cli is free software; you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation; either version 2 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program; if not, write to the Free Software
 * Foundation, Inc., 51 Franklin St, Fifth Floor, Boston, MA  02110-1301  USA
 *
 * As a special exception, you may use this file as part of a free software
 * library without restriction.  Specifically, if other files instantiate
 * templates or use macros or inline functions from this file, or you compile
 * this file and link it with other files to produce an executable, this
 * file does not by itself cause the resulting executable to be covered by
 * the GNU General Public License.  This exception does not however
 * invalidate any other reasons why the executable file might be covered by
 * the GNU General Public License.
 *
 * Copyright 2017 ARDUINO AG (http://www.arduino.cc/)
 */

package commands_test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/arduino/go-paths-helper"

	"github.com/bcmi-labs/arduino-cli/commands/root"
	"github.com/bouk/monkey"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Redirecting stdOut so we can analyze output line by
// line and check with what we want.
var stdOut *os.File = os.Stdout

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

	return
}

func makeTempDataDir(t *testing.T) func() {
	tmp, err := paths.MkTempDir("", "test")
	require.NoError(t, err, "making temporary staging dir")
	os.Setenv("ARDUINO_DATA_DIR", tmp.String())
	fmt.Printf("ARDUINO_DATA_DIR = %s\n", os.Getenv("ARDUINO_DATA_DIR"))
	return func() {
		os.Unsetenv("ARDUINO_DATA_DIR")
		tmp.RemoveAll()
		fmt.Printf("ARDUINO_DATA_DIR = %s\n", os.Getenv("ARDUINO_DATA_DIR"))
	}
}

func makeTempSketchbookDir(t *testing.T) func() {
	tmp, err := paths.MkTempDir("", "test")
	require.NoError(t, err, "making temporary staging dir")
	os.Setenv("ARDUINO_SKETCHBOOK_DIR", tmp.String())
	fmt.Printf("ARDUINO_SKETCHBOOK_DIR = %s\n", os.Getenv("ARDUINO_DATA_DIR"))
	return func() {
		os.Unsetenv("ARDUINO_SKETCHBOOK_DIR")
		tmp.RemoveAll()
		fmt.Printf("ARDUINO_SKETCHBOOK_DIR = %s\n", os.Getenv("ARDUINO_DATA_DIR"))
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

func TestLibDownloadAndInstall(t *testing.T) {
	defer makeTempDataDir(t)()
	defer makeTempSketchbookDir(t)()

	exitCode, d := executeWithArgs(t, "core", "update-index")
	require.Zero(t, exitCode, "exit code")

	exitCode, d = executeWithArgs(t, "lib", "download", "inexistentLibrary", "--format", "json")
	require.NotZero(t, exitCode, "exit code")
	require.Contains(t, string(d), "library inexistentLibrary not found")

	exitCode, d = executeWithArgs(t, "lib", "download", "inexistentLibrary")
	require.NotZero(t, exitCode, "exit code")
	require.Contains(t, string(d), "library inexistentLibrary not found")

	exitCode, d = executeWithArgs(t, "lib", "download", "Audio")
	require.Zero(t, exitCode, "exit code")
	require.Contains(t, string(d), "Audio@")
	require.Contains(t, string(d), "downloaded")

	exitCode, d = executeWithArgs(t, "lib", "download", "Audio@1.2.3-nonexistent")
	require.NotZero(t, exitCode, "exit code")
	require.Contains(t, string(d), "not found")

	exitCode, d = executeWithArgs(t, "lib", "install", "Audio")
	require.Zero(t, exitCode, "exit code")
	require.Contains(t, string(d), "Audio@")
	require.Contains(t, string(d), "Installed")

	exitCode, d = executeWithArgs(t, "lib", "list")
	require.Zero(t, exitCode, "exit code")
	require.Contains(t, string(d), "Audio")

	exitCode, d = executeWithArgs(t, "lib", "install", "Audio")
	require.NotZero(t, exitCode, "exit code")
	require.Contains(t, string(d), "Audio@")
	require.Contains(t, string(d), "already installed")

	exitCode, d = executeWithArgs(t, "lib", "install", "Audio@1.0.4")
	require.Zero(t, exitCode, "exit code")
	require.Contains(t, string(d), "Audio@1.0.4")
	require.Contains(t, string(d), "Installed")

	exitCode, d = executeWithArgs(t, "lib", "list")
	require.Zero(t, exitCode, "exit code")
	require.Contains(t, string(d), "Audio")
	require.Contains(t, string(d), "1.0.4")
}

func updateCoreIndex(t *testing.T) {
	// run a "core update-index" to download the package_index.json
	exitCode, _ := executeWithArgs(t, "core", "update-index")
	require.Equal(t, 0, exitCode, "exit code")
}

func TestCoreDownload(t *testing.T) {
	defer makeTempDataDir(t)()
	defer makeTempSketchbookDir(t)()

	// Set staging folder to a temporary folder
	tmp, err := ioutil.TempDir(os.TempDir(), "test")
	require.NoError(t, err, "making temporary staging dir")
	defer os.RemoveAll(tmp)

	updateCoreIndex(t)

	exitCode, d := executeWithArgs(t, "core", "download", "arduino:avr@1.6.16")
	require.Zero(t, exitCode, "exit code")
	require.Contains(t, string(d), "arduino:avr-gcc@4.9.2-atmel3.5.3-arduino2 downloaded")
	require.Contains(t, string(d), "arduino:avrdude@6.3.0-arduino8 downloaded")
	require.Contains(t, string(d), "arduino:arduinoOTA@1.0.0 downloaded")
	require.Contains(t, string(d), "arduino:avr@1.6.16 downloaded")

	exitCode, d = executeWithArgs(t, "core", "download", "arduino:samd@1.2.3-notexisting")
	require.NotZero(t, exitCode, "exit code")
	require.Contains(t, string(d), "required version 1.2.3-notexisting not found for platform arduino:samd")

	exitCode, d = executeWithArgs(t, "core", "download", "arduino:notexistent")
	require.NotZero(t, exitCode, "exit code")
	require.Contains(t, string(d), "not found")

	exitCode, d = executeWithArgs(t, "core", "download", "wrongparameter")
	require.NotZero(t, exitCode, "exit code")
	require.Contains(t, string(d), "invalid item")
}
