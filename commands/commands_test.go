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
	"strconv"
	"strings"
	"testing"

	"github.com/bcmi-labs/arduino-cli/commands"
	"github.com/bcmi-labs/arduino-cli/commands/root"
	"github.com/bcmi-labs/arduino-cli/common/formatter/output"
	"github.com/bcmi-labs/arduino-cli/configs"
	"github.com/bouk/monkey"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Utility functions

// Redirecting stdOut so we can analyze output line by
// line and check with what we want.
var stdOut *os.File = os.Stdout

func createTempRedirect(t *testing.T) *os.File {
	tempFile, err := ioutil.TempFile(os.TempDir(), "test")
	require.NoError(t, err, "Opening temp file")
	os.Stdout = tempFile

	return tempFile
}

func cleanTempRedirect(t *testing.T, tempFile *os.File) {
	tempFile.Close()
	err := os.Remove(tempFile.Name())
	assert.NoError(t, err, "Removing temp file")
	os.Stdout = stdOut
}

// executeWithArgsNoError is a commodity function, which does the same as executeWithArgs,
// while failing the test unless no error has occurred
func executeWithArgsNoError(t *testing.T, args ...string) {
	err := executeWithArgs(t, args...)
	require.NoError(t, err, "Expected no error executing command")
}

// executeWithArgsError is a commodity function, which does the same as executeWithArgs,
// while failing the test if no error has occurred
func executeWithArgsError(t *testing.T, args ...string) error {
	err := executeWithArgs(t, args...)
	require.Error(t, err, "Expected an error executing command")
	return err
}

// executeWithArgs executes the Cobra Command with the given arguments
// and intercepts any errors (even `os.Exit()` ones), returning the resulting error
// and also logging it for debugging purpose
func executeWithArgs(t *testing.T, args ...string) error {
	err := executeWithArgsInternal(t, args...)
	t.Logf("Running: %s", args)
	if err != nil {
		exitCode, conversionError := strconv.Atoi(err.Error())
		if conversionError != nil {
			t.Logf("Executing command with args '%s', resulted in error: %s", args, err)
		} else {
			t.Logf("Executing command with args '%s', resulted in exit code: %d", args, exitCode)
		}
	}
	return err
}

// Please use executeWithArgs instead
func executeWithArgsInternal(t *testing.T, args ...string) (err error) {
	// Init only once.
	if !root.Command.HasFlags() {
		root.Init(true)
	}
	if args != nil {
		root.Command.SetArgs(args)
	}

	// Mock the os.Exit function, so that we can use the
	// error result for the test and prevent the test from exiting
	fakeExit := func(exitCode int) {
		panic(exitCode)
	}
	patch := monkey.Patch(os.Exit, fakeExit)
	defer patch.Unpatch()
	defer func() {
		if exitCode := recover(); exitCode != nil {
			err = fmt.Errorf("%d", exitCode)
		}
	}()

	// Execute the CLI command, in this process
	root.Command.Execute()

	return err
}

// END -- Utility functions

func TestLibSearchSuccessful(t *testing.T) {
	tempFile := createTempRedirect(t)
	defer cleanTempRedirect(t, tempFile)
	want := []string{
		//`"YouMadeIt"`,
		//`"YoutubeApi"`,
		`{"libraries":["YoutubeApi"]}`,
	}

	// arduino lib search you
	//executeWithArgsNoError(t, "lib", "search", "you") //test not working on drone, but working locally
	// arduino lib search youtu --format json
	// arduino lib search youtu --format=json
	executeWithArgsNoError(t, "lib", "search", "youtu", "--format", "json", "--names")

	checkOutput(t, want, tempFile)
}

func TestLibDownloadSuccessful(t *testing.T) {
	tempFile := createTempRedirect(t)
	defer cleanTempRedirect(t, tempFile)

	// getting the paths to create the want path of the want object.
	stagingFolder, err := configs.DownloadCacheFolder("libraries").Get()
	require.NoError(t, err, "Getting cache folder")

	// desired output
	want := output.LibProcessResults{
		Libraries: map[string]output.ProcessResult{
			"invalidLibrary":           {ItemName: "invalidLibrary", Error: "Library not found"},
			"YoutubeApi":               {ItemName: "YoutubeApi", Status: "Downloaded", Path: stagingFolder + "/YoutubeApi-1.1.0.zip"},
			"YouMadeIt@invalidVersion": {ItemName: "YouMadeIt", Error: "Version Not Found"},
		},
	}

	// lib download YoutubeApi invalidLibrary YouMadeIt@invalidVersion --format json
	librariesArgs := []string{}
	for libraryKey, _ := range want.Libraries {
		librariesArgs = append(librariesArgs, libraryKey)
	}

	executeWithArgsNoError(t, append(append([]string{"lib", "download"}, librariesArgs...), "--format", "json")...)

	// read output
	_, err = tempFile.Seek(0, 0)
	require.NoError(t, err, "Rewinding output file")
	d, err := ioutil.ReadAll(tempFile)
	require.NoError(t, err, "Reading output file")

	var have output.LibProcessResults
	err = json.Unmarshal(d, &have)
	require.NoError(t, err, "Unmarshaling json output")
	require.NotNil(t, have.Libraries, "Unmarshaling json output: have '%s'", d)

	// checking output

	assert.Equal(t, len(want.Libraries), len(have.Libraries), "Number of libraries in the output")

	pop := func(lib *output.ProcessResult) bool {
		for idx, h := range have.Libraries {
			if lib.String() == h.String() {
				// XXX: Consider changing the Libraries field to an array of pointers
				//have.Libraries[idx] = nil
				have.Libraries[idx] = output.ProcessResult{ItemName: ""} // Mark library as matched
				return true
			}
		}
		return false
	}

	for _, w := range want.Libraries {
		assert.True(t, pop(&w), "Expected library '%s' is missing from output", w)
	}
	for _, h := range have.Libraries {
		assert.Empty(t, h.String(), "Unexpected library '%s' is inside output", h)
	}
}

func TestCoreDownloadSuccessful(t *testing.T) {
	// getting the paths to create the want path of the want object.
	stagingFolder, err := configs.DownloadCacheFolder("packages").Get()
	require.NoError(t, err, "Getting cache folder")

	// desired output
	want := output.CoreProcessResults{
		Cores: map[string]output.ProcessResult{
			"arduino:samd=1.6.16":            {ItemName: "arduino:samd@1.6.16", Status: "Downloaded", Path: stagingFolder + "/samd-1.6.16.tar.bz2"},
			"arduino:sam=notexistingversion": {ItemName: "sam", Error: "Version notexistingversion Not Found"},
			"arduino:sam=1.0.0":              {ItemName: "sam", Error: "Version 1.0.0 Not Found"},
		},
		Tools: map[string]output.ProcessResult{
			"arduinoOTA":        {ItemName: "arduino:arduinoOTA@1.2.0", Status: "Downloaded", Path: stagingFolder + "/arduinoOTA-1.2.0-linux_amd64.tar.bz2"},
			"openocd":           {ItemName: "arduino:openocd@0.9.0-arduino6-static", Status: "Downloaded", Path: stagingFolder + "/openocd-0.9.0-arduino6-static-x86_64-linux-gnu.tar.bz2"},
			"CMSIS-Atmel":       {ItemName: "arduino:CMSIS-Atmel@1.1.0", Status: "Downloaded", Path: stagingFolder + "/CMSIS-Atmel-1.1.0.tar.bz2"},
			"CMSIS":             {ItemName: "arduino:CMSIS@4.5.0", Status: "Downloaded", Path: stagingFolder + "/CMSIS-4.5.0.tar.bz2"},
			"arm-none-eabi-gcc": {ItemName: "arduino:arm-none-eabi-gcc@4.8.3-2014q1", Status: "Downloaded", Path: stagingFolder + "/gcc-arm-none-eabi-4.8.3-2014q1-linux64.tar.gz"},
			"bossac":            {ItemName: "arduino:bossac@1.7.0", Status: "Downloaded", Path: stagingFolder + "/bossac-1.7.0-x86_64-linux-gnu.tar.gz"},
		},
	}

	testCoreDownload(t, want, func(err error, stdOut []byte) {
		if err != nil {
			t.Log("COMMAND OUTPUT:\n", string(stdOut))
		}
		require.NoError(t, err, "Expected no error executing command")

		var have output.CoreProcessResults
		err = json.Unmarshal(stdOut, &have)
		require.NoError(t, err, "Unmarshaling json output")
		require.NotNil(t, have.Cores, "Unmarshaling json output: have '%s'", stdOut)

		// checking output

		assert.Equal(t, len(want.Cores), len(have.Cores), "Number of cores in the output")

		pop := func(core *output.ProcessResult) bool {
			for idx, h := range have.Cores {
				if core.String() == h.String() {
					// XXX: Consider changing the Cores field to an array of pointers
					//have.Cores[idx] = nil
					have.Cores[idx] = output.ProcessResult{ItemName: ""} // Mark core as matched
					return true
				}
			}
			return false
		}
		for _, w := range want.Cores {
			popR := pop(&w)
			t.Log(w)
			t.Log(popR)
			assert.True(t, popR, "Expected core '%s' is missing from output", w)
		}
		for _, h := range have.Cores {
			assert.Empty(t, h.ItemName, "Unexpected core '%s' is inside output", h)
		}

		assert.Equal(t, len(want.Tools), len(have.Tools), "Number of tools in the output")

		pop = func(tool *output.ProcessResult) bool {
			for idx, h := range have.Tools {
				if tool.String() == h.String() {
					// XXX: Consider changing the Tools field to an array of pointers
					// have.Tools[idx] = nil
					have.Tools[idx] = output.ProcessResult{ItemName: ""} // Mark tool as matched
					return true
				}
			}
			return false
		}

		for _, w := range want.Tools {
			assert.True(t, pop(&w), "Expected tool '%s' is missing from output", w)
		}
		for _, h := range have.Tools {
			assert.Empty(t, h.String(), "Unexpected tool '%s' is inside output", h)
		}
	})
}

func TestCoreDownloadBadArgument(t *testing.T) {
	// desired output
	want := output.CoreProcessResults{
		Cores: map[string]output.ProcessResult{
			"unparsablearg": {ItemName: "unparsablearg", Error: "Invalid item (not PACKAGER:CORE[=VERSION])"},
		},
		Tools: map[string]output.ProcessResult{},
	}

	testCoreDownload(t, want, func(err error, stdOut []byte) {
		require.EqualError(t, err, strconv.Itoa(commands.ErrBadArgument),
			fmt.Sprintf("Expected an '%s' error (exit code '%d') executing command",
				"commands.ErrBadArgument",
				commands.ErrBadArgument))
	})
}

func testCoreDownload(t *testing.T, want output.CoreProcessResults, handleResults func(err error, stdOut []byte)) {
	// run a "core update-index" to download the package_index.json
	err := executeWithArgs(t, "core", "update-index")
	require.NoError(t, err, "running 'core update-index'")

	// start output capture
	tempFile := createTempRedirect(t)
	defer cleanTempRedirect(t, tempFile)

	// core download arduino:samd=1.6.16 unparsablearg arduino:sam=notexistingversion arduino:sam=1.0.0 --format json
	coresArgs := []string{"core", "download"}
	for coreKey, _ := range want.Cores {
		coresArgs = append(coresArgs, coreKey)
	}
	coresArgs = append(coresArgs, "--format", "json")
	err = executeWithArgs(t, coresArgs...)

	// read output
	var stdOut []byte
	if err == nil {
		_, err = tempFile.Seek(0, 0)
		require.NoError(t, err, "Rewinding output file")
		stdOut, err = ioutil.ReadAll(tempFile)
		require.NoError(t, err, "Reading output file")
	}
	handleResults(err, stdOut)
}

func checkOutput(t *testing.T, want []string, tempFile *os.File) {
	_, err := tempFile.Seek(0, 0)
	require.NoError(t, err, "Rewinding output file")
	d, err := ioutil.ReadAll(tempFile)
	require.NoError(t, err, "Reading output file")

	have := strings.Split(strings.TrimSpace(string(d)), "\n")
	assert.Equal(t, len(want), len(have), "Number of lines in the output")

	for i := range have {
		assert.Equal(t, want[i], have[i], "Content of line %d", i)
	}
}
