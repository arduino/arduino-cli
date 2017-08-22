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

func executeWithArgs(args ...string) {
	if args != nil {
		cmd.InitFlags()
		cmd.InitCommands()
		cmd.ArduinoCmd.SetArgs(args)
	}
	err := cmd.ArduinoCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func TestArduinoCmd(t *testing.T) {
	tempFile := createTempRedirect(t)
	defer cleanTempRedirect(tempFile)
	want := []string{
		`{"error":"Invalid Call : should show Help, but it is available only in TEXT mode"}`,
	}

	// arduino --format json
	// arduino --format=json
	executeWithArgs("--format", "json")

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
	executeWithArgs("lib", "search", "you")
	// arduino lib search youtu --format json
	// arduino lib search youtu --format=json
	executeWithArgs("lib", "search", "youtu", "--format", "json")

	checkOutput(t, want, tempFile)
}

func TestLibDownload(t *testing.T) {
	tempFile := createTempRedirect(t)
	defer cleanTempRedirect(tempFile)

	// getting the paths to create the want path of the want object.
	stagingFolder, err := common.GetDownloadCacheFolder("libraries")
	if err != nil {
		t.Error("Cannot get cache folder")
	}

	// getting what I want...
	var have, want output.LibProcessResults
	jsonObj := fmt.Sprintf(`{"libraries":[{"name":"invalidLibrary","error":"Library not found"},{"name":"YoutubeApi","status":"Downloaded","path":"%s/YoutubeApi-1.0.0.zip"},{"name":"YouMadeIt","error":"Version Not Found"}]}`,
		stagingFolder)
	t.Log(jsonObj)
	err = json.Unmarshal([]byte(jsonObj), &want)
	if err != nil {
		t.Error("JSON marshalling error. TestLibDownload want. " + err.Error())
	}

	// arduino lib download YoutubeApi --format json
	executeWithArgs("lib", "download", "YoutubeApi", "invalidLibrary", "YouMadeIt@invalidVersion", "--format", "json")

	//resetting the file to allow the full read (it has been written by executeWithArgs)
	_, err = tempFile.Seek(0, 0)
	if err != nil {
		t.Error("Cannot set file for read mode")
	}

	d, _ := ioutil.ReadAll(tempFile)
	err = json.Unmarshal(d, &have)
	if err != nil {
		t.Error("JSON marshalling error. TestLibDownload have ", jsonObj)
	}

	//checking if it is what I want...
	if len(have.Libraries) != len(want.Libraries) {
		t.Error("Output not matching, different line number from command")
	}

	//since the order of the libraries is random I have to scan the whole array everytime.
	for _, itemHave := range have.Libraries {
		ok := false
		for _, itemWant := range want.Libraries {
			t.Log(itemHave, " -------- ", itemWant)
			if itemHave.String() == itemWant.String() {
				ok = true
				break
			}
		}
		if !ok {
			t.Errorf("Got %s not found", itemHave)
		}
	}
}

func TestCoreDownload(t *testing.T) {
	tempFile := createTempRedirect(t)
	defer cleanTempRedirect(tempFile)

	// getting the paths to create the want path of the want object.
	stagingFolder, err := common.GetDownloadCacheFolder("packages")
	if err != nil {
		t.Error("Cannot get cache folder")
	}

	// getting what I want...
	var have, want output.CoreProcessResults
	jsonObj := strings.Replace(`{"cores":[{"name":"unparsablearg","error":"Invalid item (not PACKAGER:CORE[=VERSION])"},{"name":"sam","error":"Version notexistingversion Not Found"},{"name":"sam","error":"Version 1.0.0 Not Found"},{"name":"samd","status":"Downloaded","path":"%s/samd-1.6.15.tar.bz2"}],"tools":[{"name":"arduinoOTA","status":"Downloaded","path":"%s/arduinoOTA-1.2.0-linux_amd64.tar.bz2"},{"name":"openocd","status":"Downloaded","path":"%s/openocd-0.9.0-arduino6-static-x86_64-linux-gnu.tar.bz2"},{"name":"CMSIS-Atmel","status":"Downloaded","path":"%s/CMSIS-Atmel-1.1.0.tar.bz2"},{"name":"CMSIS","status":"Downloaded","path":"%s/CMSIS-4.5.0.tar.bz2"},{"name":"arm-none-eabi-gcc","status":"Downloaded","path":"%s/gcc-arm-none-eabi-4.8.3-2014q1-linux64.tar.gz"},{"name":"bossac","status":"Downloaded","path":"%s/bossac-1.7.0-x86_64-linux-gnu.tar.gz"}]}`,
		"%s", stagingFolder, 8)

	t.Log(jsonObj)
	err = json.Unmarshal([]byte(jsonObj), &want)
	if err != nil {
		t.Error("JSON marshalling error. TestCoreDownload want. " + err.Error())
	}

	// arduino lib download YoutubeApi --format json
	executeWithArgs("core", "download", "arduino:samd", "unparsablearg", "arduino:sam=notexistingversion", "arduino:sam=1.0.0", "--format", "json")

	//resetting the file to allow the full read (it has been written by executeWithArgs)
	_, err = tempFile.Seek(0, 0)
	if err != nil {
		t.Error("Cannot set file for read mode")
	}

	d, _ := ioutil.ReadAll(tempFile)
	err = json.Unmarshal(d, &have)
	if err != nil {
		t.Error("JSON marshalling error ", d)
	}

	//checking if it is what I want...
	if len(have.Cores) != len(want.Cores) || len(have.Tools) != len(want.Tools) {
		t.Error("Output not matching, different line number from command", len(have.Cores), len(want.Cores), len(have.Tools), len(want.Tools))
	}

	for _, itemHave := range have.Cores {
		ok := false
		for _, itemWant := range want.Cores {
			//t.Log(itemHave, " -------- ", itemWant)
			if itemHave.String() == itemWant.String() {
				ok = true
				break
			}
		}
		if !ok {
			t.Errorf(`Got "%s" not found`, itemHave)
		}
	}

	for _, itemHave := range have.Tools {
		ok := false
		for _, itemWant := range want.Tools {
			//t.Log(itemHave, " -------- ", itemWant)
			if itemHave.String() == itemWant.String() {
				ok = true
				break
			}
		}
		if !ok {
			t.Errorf(`Got "%s" not found`, itemHave)
		}
	}
}

func checkOutput(t *testing.T, want []string, tempFile *os.File) {
	_, err := tempFile.Seek(0, 0)
	if err != nil {
		t.Error("Cannot set file for read mode")
	}

	d, _ := ioutil.ReadAll(tempFile)
	have := strings.Split(strings.TrimSpace(string(d)), "\n")
	if len(have) != len(want) {
		t.Error("Output not matching, different line number from command")
	}

	for i := range have {
		if have[i] != want[i] {
			fmt.Fprintln(stdOut, have)
			fmt.Fprintln(stdOut, want)
			t.Errorf(`Expected "%s", but had "%s"`, want[i], have[i])
		}
	}
}
