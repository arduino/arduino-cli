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
 * Copyright 2017 BCMI LABS SA (http://www.arduino.cc/)
 */

package cmd_test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/bcmi-labs/arduino-cli/cmd"
	"github.com/bcmi-labs/arduino-cli/cmd/output"
	"github.com/bcmi-labs/arduino-cli/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

/*
NOTE: the use of func init() for test is discouraged, please create public InitFunctions and call them,
	  or use (untested) cmd.PersistentPreRun or cmd.PreRun to reinitialize the flags and the commands every time.
*/

// Redirecting stdOut so we can analyze output line by
// line and check with what we want.
var stdOut *os.File = os.Stdout

func createTempRedirect(t *testing.T) *os.File {
	tempFile, err := ioutil.TempFile(os.TempDir(), "test")
	require.NoError(t, err, "Opening temp file")
	os.Stdout = tempFile
	return tempFile
}

func cleanTempRedirect(tempFile *os.File) {
	tempFile.Close()
	os.Remove(tempFile.Name())
	os.Stdout = stdOut
}

func executeWithArgs(t *testing.T, args ...string) {
	if args != nil {
		cmd.InitFlags()
		cmd.InitCommands()
		cmd.ArduinoCmd.SetArgs(args)
	}
	err := cmd.ArduinoCmd.Execute()
	require.NoError(t, err, "Error executing command")
}

func TestArduinoCmd(t *testing.T) {
	tempFile := createTempRedirect(t)
	defer cleanTempRedirect(tempFile)
	want := []string{
		`{"error":"Invalid Call : should show Help, but it is available only in TEXT mode"}`,
	}

	// arduino --format json
	// arduino --format=json
	executeWithArgs(t, "--format", "json")

	checkOutput(t, want, tempFile)
}

func TestLibSearch(t *testing.T) {
	tempFile := createTempRedirect(t)
	defer cleanTempRedirect(tempFile)
	want := []string{
		`"YouMadeIt"`,
		`"YoutubeApi"`,
		`{"libraries":["YoutubeApi"]}`,
	}

	// arduino lib search you
	executeWithArgs(t, "lib", "search", "you")
	// arduino lib search youtu --format json
	// arduino lib search youtu --format=json
	executeWithArgs(t, "lib", "search", "youtu", "--format", "json")

	checkOutput(t, want, tempFile)
}

func TestLibDownload(t *testing.T) {
	tempFile := createTempRedirect(t)
	defer cleanTempRedirect(tempFile)

	// getting the paths to create the want path of the want object.
	stagingFolder, err := common.GetDownloadCacheFolder("libraries")
	require.NoError(t, err, "Getting cache folder")

	// desired output
	want := output.LibProcessResults{
		Libraries: []output.ProcessResult{
			{ItemName: "invalidLibrary", Error: "Library not found"},
			{ItemName: "YoutubeApi", Status: "Downloaded", Path: stagingFolder + "/YoutubeApi-1.0.0.zip"},
			{ItemName: "YouMadeIt", Error: "Version Not Found"},
		},
	}

	// lib download YoutubeApi invalidLibrary YouMadeIt@invalidVersion --format json
	executeWithArgs(t, "lib", "download", "YoutubeApi", "invalidLibrary", "YouMadeIt@invalidVersion", "--format", "json")

	// resetting the file to allow the full read (it has been written by executeWithArgs)
	_, err = tempFile.Seek(0, 0)
	require.NoError(t, err, "Rewinding output file")

	d, err := ioutil.ReadAll(tempFile)
	require.NoError(t, err, "Reading output file")

	var have output.LibProcessResults
	err = json.Unmarshal(d, &have)
	require.NoError(t, err, "Unmarshaling json output")

	// checking if it is what I want...
	assert.Equal(t, len(want.Libraries), len(have.Libraries), "Number of libraries in the output")

	// since the order of the libraries is random I have to scan the whole array everytime.
	pop := func(lib *output.ProcessResult) bool {
		for idx, h := range have.Libraries {
			if lib.String() == h.String() {
				// XXX: Consider changing the Libraries field to an array of pointers
				//have.Libraries[idx] = nil
				have.Libraries[idx] = output.ProcessResult{ItemName: ""}
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

func TestCoreDownload(t *testing.T) {
	tempFile := createTempRedirect(t)
	defer cleanTempRedirect(tempFile)

	// getting the paths to create the want path of the want object.
	stagingFolder, err := common.GetDownloadCacheFolder("packages")
	require.NoError(t, err, "Getting cache folder")

	// desired output
	want := output.CoreProcessResults{
		Cores: []output.ProcessResult{
			{ItemName: "unparsablearg", Error: "Invalid item (not PACKAGER:CORE[=VERSION])"},
			{ItemName: "sam", Error: "Version notexistingversion Not Found"},
			{ItemName: "sam", Error: "Version 1.0.0 Not Found"},
			{ItemName: "samd", Status: "Downloaded", Path: stagingFolder + "/samd-1.6.15.tar.bz2"},
		},
		Tools: []output.ProcessResult{
			{ItemName: "arduinoOTA", Status: "Downloaded", Path: stagingFolder + "/arduinoOTA-1.2.0-linux_amd64.tar.bz2"},
			{ItemName: "openocd", Status: "Downloaded", Path: stagingFolder + "/openocd-0.9.0-arduino6-static-x86_64-linux-gnu.tar.bz2"},
			{ItemName: "CMSIS-Atmel", Status: "Downloaded", Path: stagingFolder + "/CMSIS-Atmel-1.1.0.tar.bz2"},
			{ItemName: "CMSIS", Status: "Downloaded", Path: stagingFolder + "/CMSIS-4.5.0.tar.bz2"},
			{ItemName: "arm-none-eabi-gcc", Status: "Downloaded", Path: stagingFolder + "/gcc-arm-none-eabi-4.8.3-2014q1-linux64.tar.gz"},
			{ItemName: "bossac", Status: "Downloaded", Path: stagingFolder + "/bossac-1.7.0-x86_64-linux-gnu.tar.gz"},
		},
	}

	// core download arduino:samd unparsablearg arduino:sam=notexistingversion arduino:sam=1.0.0 --format json
	executeWithArgs(t, "core", "download", "arduino:samd", "unparsablearg", "arduino:sam=notexistingversion", "arduino:sam=1.0.0", "--format", "json")

	//resetting the file to allow the full read (it has been written by executeWithArgs)
	_, err = tempFile.Seek(0, 0)
	require.NoError(t, err, "Rewinding output file")
	d, err := ioutil.ReadAll(tempFile)
	require.NoError(t, err, "Reading output file")

	var have output.CoreProcessResults
	err = json.Unmarshal(d, &have)
	require.NoError(t, err, "Unmarshaling json output")

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
		assert.True(t, pop(&w), "Expected core '%s' is missing from output", w)
	}
	for _, h := range have.Cores {
		assert.Empty(t, h.String(), "Unexpected core '%s' is inside output", h)
	}

	assert.Equal(t, len(want.Tools), len(have.Tools), "Number of tools in the output")

	pop = func(tool *output.ProcessResult) bool {
		for idx, h := range have.Tools {
			if tool.String() == h.String() {
				// XXX: Consider changing the Tools field to an array of pointers
				//have.Tools[idx] = nil
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
