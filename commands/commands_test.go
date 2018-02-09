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
	"strings"
	"testing"

	"github.com/bcmi-labs/arduino-cli/commands/root"
	"github.com/bcmi-labs/arduino-cli/common"
	"github.com/bcmi-labs/arduino-cli/common/formatter/output"
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

func cleanTempRedirect(t *testing.T, tempFile *os.File) {
	tempFile.Close()
	err := os.Remove(tempFile.Name())
	assert.NoError(t, err, "Removing temp file")
	os.Stdout = stdOut
}

func executeWithArgs(t *testing.T, args ...string) {
	// Init only once.
	if !root.Command.HasFlags() {
		root.Init()
	}
	if args != nil {
		root.Command.SetArgs(args)
	}

	err := root.Command.Execute()
	fmt.Fprintln(stdOut, err)
	require.NoError(t, err, "Error executing command")
}

func TestLibSearch(t *testing.T) {
	tempFile := createTempRedirect(t)
	defer cleanTempRedirect(t, tempFile)
	want := []string{
		//`"YouMadeIt"`,
		//`"YoutubeApi"`,
		`{"libraries":["YoutubeApi"]}`,
	}

	// arduino lib search you
	//executeWithArgs(t, "lib", "search", "you") //test not working on drone, but working locally
	// arduino lib search youtu --format json
	// arduino lib search youtu --format=json
	executeWithArgs(t, "lib", "search", "youtu", "--format", "json", "--names")

	checkOutput(t, want, tempFile)
}

func TestLibDownload(t *testing.T) {
	tempFile := createTempRedirect(t)
	defer cleanTempRedirect(t, tempFile)

	// getting the paths to create the want path of the want object.
	stagingFolder, err := common.DownloadCacheFolder("libraries").Get()
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

	// read output
	_, err = tempFile.Seek(0, 0)
	require.NoError(t, err, "Rewinding output file")
	d, err := ioutil.ReadAll(tempFile)
	require.NoError(t, err, "Reading output file")

	var have output.LibProcessResults
	err = json.Unmarshal(d, &have)
	require.NoError(t, err, "Unmarshaling json output")

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

func TestCoreDownload(t *testing.T) {
	tempFile := createTempRedirect(t)
	defer cleanTempRedirect(t, tempFile)

	// getting the paths to create the want path of the want object.
	stagingFolder, err := common.DownloadCacheFolder("packages").Get()
	require.NoError(t, err, "Getting cache folder")

	// desired output
	want := output.CoreProcessResults{
		Cores: []output.ProcessResult{
			{ItemName: "unparsablearg", Error: "Invalid item (not PACKAGER:CORE[=VERSION])"},
			{ItemName: "sam", Error: "Version notexistingversion Not Found"},
			{ItemName: "sam", Error: "Version 1.0.0 Not Found"},
			{ItemName: "samd", Status: "Downloaded", Path: stagingFolder + "/samd-1.6.16.tar.bz2"},
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

	// core download arduino:samd=1.6.16 unparsablearg arduino:sam=notexistingversion arduino:sam=1.0.0 --format json
	executeWithArgs(t, "core", "download", "arduino:samd=1.6.16", "unparsablearg", "arduino:sam=notexistingversion", "arduino:sam=1.0.0", "--format", "json")

	// read output
	_, err = tempFile.Seek(0, 0)
	require.NoError(t, err, "Rewinding output file")
	d, err := ioutil.ReadAll(tempFile)
	require.NoError(t, err, "Reading output file")

	var have output.CoreProcessResults
	err = json.Unmarshal(d, &have)
	require.NoError(t, err, "Unmarshaling json output")
	t.Log("HAVE: \n", have)
	t.Log("WANT: \n", want)

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
