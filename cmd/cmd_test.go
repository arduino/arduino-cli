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
)

/*
This test file will always fail if all tests are executed at the same time
this is a go test error with cobra which relies on init() function
there is an open issue about that : https://github.com/bcmi-labs/arduino-cli/issues/58
For now test all test functions separately.
*/

var stdOut *os.File

func init() {
	stdOut = os.Stdout
}

func createTempRedirect() *os.File {
	tempFile, err := ioutil.TempFile(os.TempDir(), "test")
	if err != nil {
		fmt.Fprint(stdOut, err)
	}
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
	tempFile := createTempRedirect()
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
	tempFile := createTempRedirect()
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
	tempFile := createTempRedirect()
	defer cleanTempRedirect(tempFile)

	// getting the paths to create the want path of the want object.
	stagingFolder, err := common.GetDownloadCacheFolder("libraries")
	if err != nil {
		t.Error("Cannot get cache folder")
	}

	// getting what I want...
	var have, want output.LibProcessResults
	err = json.Unmarshal([]byte(fmt.Sprintf(`{"libraries":[{"name":"invalidLibrary","error":"Library not found"},{"name":"YoutubeApi","status":"Downloaded","path":"%s/YoutubeApi-1.0.0.zip"},{"name":"YouMadeIt","error":"Version Not Found"}]}`,
		stagingFolder)), &want)
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
		t.Error("JSON marshalling error. TestLibDownload have")
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
