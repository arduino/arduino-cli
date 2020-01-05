// This file is part of arduino-cli.
//
// Copyright 2020 ARDUINO SA (http://www.arduino.cc/)
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

// These tests are mocked and won't work on OSX
// +build !darwin

package cli

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/spf13/viper"

	"github.com/arduino/arduino-cli/cli/errorcodes"
	"github.com/arduino/arduino-cli/cli/feedback"

	"bou.ke/monkey"
	paths "github.com/arduino/go-paths-helper"
	"github.com/stretchr/testify/require"
	semver "go.bug.st/relaxed-semver"
)

var (
	// Redirecting stdOut so we can analyze output line by
	// line and check with what we want.
	stdOut = os.Stdout
	stdErr = os.Stderr

	currDownloadDir string
	currDataDir     string
	currUserDir     string
)

type outputRedirect struct {
	tempFile *os.File
}

func (grabber *outputRedirect) Open() {
	tempFile, err := ioutil.TempFile(os.TempDir(), "test")
	if err != nil {
		panic("Opening temp output file")
	}
	os.Stdout = tempFile
	os.Stderr = tempFile
	grabber.tempFile = tempFile
}

func (grabber *outputRedirect) GetOutput() []byte {
	_, err := grabber.tempFile.Seek(0, 0)
	if err != nil {
		panic("Rewinding temp output file")
	}

	output, err := ioutil.ReadAll(grabber.tempFile)
	if err != nil {
		panic("Reading temp output file")
	}

	return output
}

func (grabber *outputRedirect) Close() {
	grabber.tempFile.Close()
	err := os.Remove(grabber.tempFile.Name())
	if err != nil {
		panic("Removing temp output file")
	}
	os.Stdout = stdOut
	os.Stderr = stdErr
}

func TestMain(m *testing.M) {
	// all these tests perform actual operations, don't run in short mode
	flag.Parse()
	if testing.Short() {
		fmt.Println("skip integration tests")
		os.Exit(0)
	}

	// SetUp
	currDataDir = tmpDirOrDie()
	os.MkdirAll(filepath.Join(currDataDir, "packages"), 0755)
	os.Setenv("ARDUINO_DIRECTORIES_DATA", currDataDir)
	currDownloadDir = tmpDirOrDie()
	os.Setenv("ARDUINO_DIRECTORIES_DOWNLOADS", currDownloadDir)
	currUserDir = filepath.Join("testdata", "custom_hardware")
	// use ARDUINO_SKETCHBOOK_DIR instead of ARDUINO_DIRECTORIES_USER to
	// ensure the backward compat code is working
	os.Setenv("ARDUINO_SKETCHBOOK_DIR", currUserDir)

	// Run
	res := m.Run()

	// TearDown
	os.RemoveAll(currDataDir)
	os.Unsetenv("ARDUINO_DIRECTORIES_DATA")
	currDataDir = ""
	os.RemoveAll(currDownloadDir)
	os.Unsetenv("ARDUINO_DIRECTORIES_DOWNLOADS")
	currDownloadDir = ""
	os.Unsetenv("ARDUINO_SKETCHBOOK_DIR")

	os.Exit(res)
}

func tmpDirOrDie() string {
	dir, err := ioutil.TempDir(os.TempDir(), "cli_test")
	if err != nil {
		panic(fmt.Sprintf("error creating tmp dir: %v", err))
	}
	return dir
}

// executeWithArgs executes the Cobra Command with the given arguments
// and intercepts any errors (even `os.Exit()` ones), returning the exit code
func executeWithArgs(args ...string) (int, []byte) {
	var output []byte
	var exitCode int
	fmt.Printf("RUNNING: %s\n", args)
	viper.Reset()

	// This closure is here because we won't that the defer are executed after the end of the "executeWithArgs" method
	func() {
		redirect := &outputRedirect{}
		redirect.Open()
		// re-init feedback so it'll write to our grabber
		feedback.SetDefaultFeedback(feedback.New(os.Stdout, os.Stdout, feedback.Text))
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

		// Execute the CLI command, start fresh every time
		ArduinoCli.ResetCommands()
		ArduinoCli.ResetFlags()
		createCliCommandTree(ArduinoCli)
		ArduinoCli.SetArgs(args)
		if err := ArduinoCli.Execute(); err != nil {
			exitCode = errorcodes.ErrGeneric
		}
	}()

	return exitCode, output
}

func detectLatestAVRCore(t *testing.T) string {
	jsonFile := filepath.Join(currDataDir, "package_index.json")
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
	jsonData, err := ioutil.ReadFile(jsonFile)
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
	return latest.String()
}

// END -- Utility functions

func TestUploadIntegration(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("This test runs only on Linux")
	}

	exitCode, _ := executeWithArgs("core", "update-index")
	require.Zero(t, exitCode)

	exitCode, _ = executeWithArgs("core", "install", "arduino:avr")
	require.Zero(t, exitCode)

	// -i flag
	exitCode, d := executeWithArgs("upload", "-i", filepath.Join(currUserDir, "test.hex"), "-b", "test:avr:testboard", "-p", "/dev/ttyACM0")
	require.Zero(t, exitCode)
	require.Contains(t, string(d), "QUIET")
	require.Contains(t, string(d), "NOVERIFY")
	require.Contains(t, string(d), "testdata/custom_hardware/test.hex")

	// -i flag with implicit extension
	exitCode, d = executeWithArgs("upload", "-i", filepath.Join(currUserDir, "test"), "-b", "test:avr:testboard", "-p", "/dev/ttyACM0")
	require.Zero(t, exitCode)
	require.Contains(t, string(d), "QUIET")
	require.Contains(t, string(d), "NOVERIFY")
	require.Contains(t, string(d), "testdata/custom_hardware/test.hex")

	// -i with absolute path
	fullPath, err := filepath.Abs(filepath.Join(currUserDir, "test.hex"))
	require.NoError(t, err)
	exitCode, d = executeWithArgs("upload", "-i", fullPath, "-b", "test:avr:testboard", "-p", "/dev/ttyACM0")
	require.Zero(t, exitCode)
	require.Contains(t, string(d), "QUIET")
	require.Contains(t, string(d), "NOVERIFY")
	require.Contains(t, string(d), "testdata/custom_hardware/test.hex")

	// -v verbose
	exitCode, d = executeWithArgs("upload", "-v", "-t", "-i", filepath.Join(currUserDir, "test.hex"), "-b", "test:avr:testboard", "-p", "/dev/ttyACM0")
	require.Zero(t, exitCode)
	require.Contains(t, string(d), "VERBOSE")
	require.Contains(t, string(d), "VERIFY")
	require.Contains(t, string(d), "testdata/custom_hardware/test.hex")

	// -t verify
	exitCode, d = executeWithArgs("upload", "-i", filepath.Join(currUserDir, "test.hex"), "-b", "test:avr:testboard", "-p", "/dev/ttyACM0")
	require.Zero(t, exitCode)
	require.Contains(t, string(d), "QUIET")
	require.Contains(t, string(d), "NOVERIFY")
	require.Contains(t, string(d), "testdata/custom_hardware/test.hex")

	// -v -t verbose verify
	exitCode, d = executeWithArgs("upload", "-v", "-t", "-i", filepath.Join(currUserDir, "test.hex"), "-b", "test:avr:testboard", "-p", "/dev/ttyACM0")
	require.Zero(t, exitCode)
	require.Contains(t, string(d), "VERBOSE")
	require.Contains(t, string(d), "VERIFY")
	require.Contains(t, string(d), "testdata/custom_hardware/test.hex")

	// non-existent file
	exitCode, _ = executeWithArgs("upload", "-i", filepath.Join(currUserDir, "test123.hex"), "-b", "test:avr:testboard", "-p", "/dev/ttyACM0")
	require.NotZero(t, exitCode)

	// sketch
	exitCode, d = executeWithArgs("upload", filepath.Join(currUserDir, "TestSketch"), "-b", "test:avr:testboard", "-p", "/dev/ttyACM0")
	require.Zero(t, exitCode)
	require.Contains(t, string(d), "QUIET")
	require.Contains(t, string(d), "NOVERIFY")
	require.Contains(t, string(d), "testdata/custom_hardware/TestSketch/TestSketch.test.avr.testboard.hex")

	// sketch without build
	exitCode, _ = executeWithArgs("upload", filepath.Join(currUserDir, "TestSketch2"), "-b", "test:avr:testboard", "-p", "/dev/ttyACM0")
	require.NotZero(t, exitCode)

	// platform without 'recipe.output.tmp_file' property
	exitCode, _ = executeWithArgs("upload", "-i", filepath.Join(currUserDir, "test.hex"), "-b", "test2:avr:testboard", "-p", "/dev/ttyACM0")
	require.NotZero(t, exitCode)
}

func TestCompileCommandsIntegration(t *testing.T) {
	// Set staging dir to a temporary dir
	tmp := tmpDirOrDie()
	defer os.RemoveAll(tmp)

	exitCode, _ := executeWithArgs("core", "update-index")
	require.Zero(t, exitCode)

	// Download latest AVR
	exitCode, _ = executeWithArgs("core", "install", "arduino:avr")
	require.Zero(t, exitCode)

	// Create a test sketch
	sketchPath := filepath.Join(tmp, "Test1")
	exitCode, d := executeWithArgs("sketch", "new", sketchPath)
	require.Zero(t, exitCode)

	// Build sketch without FQBN
	exitCode, d = executeWithArgs("compile", sketchPath)
	require.NotZero(t, exitCode)
	require.Contains(t, string(d), "Error: no FQBN provided. Set --fqbn flag or attach board to sketch")

	// Build sketch for arduino:avr:uno
	exitCode, d = executeWithArgs("compile", "-b", "arduino:avr:uno", sketchPath)
	require.Zero(t, exitCode)
	require.Contains(t, string(d), "Sketch uses")
	require.True(t, paths.New(sketchPath).Join("Test1.arduino.avr.uno.hex").Exist())

	// Build sketch for arduino:avr:nano (without options)
	exitCode, d = executeWithArgs("compile", "-b", "arduino:avr:nano", sketchPath)
	require.Zero(t, exitCode)
	require.Contains(t, string(d), "Sketch uses")
	require.True(t, paths.New(sketchPath).Join("Test1.arduino.avr.nano.hex").Exist())

	// Build sketch with --output path
	pwd, err := os.Getwd()
	require.NoError(t, err)
	defer func() { require.NoError(t, os.Chdir(pwd)) }()
	require.NoError(t, os.Chdir(tmp))

	exitCode, d = executeWithArgs("compile", "-b", "arduino:avr:nano", "-o", "test", sketchPath)
	require.Zero(t, exitCode)
	require.Contains(t, string(d), "Sketch uses")
	require.True(t, paths.New("test.hex").Exist())

	exitCode, d = executeWithArgs("compile", "-b", "arduino:avr:nano", "-o", "test2.hex", sketchPath)
	require.Zero(t, exitCode)
	require.Contains(t, string(d), "Sketch uses")
	require.True(t, paths.New("test2.hex").Exist())
	require.NoError(t, paths.New(tmp, "anothertest").MkdirAll())

	exitCode, d = executeWithArgs("compile", "-b", "arduino:avr:nano", "-o", "anothertest/test", sketchPath)
	require.Zero(t, exitCode)
	require.Contains(t, string(d), "Sketch uses")
	require.True(t, paths.New("anothertest", "test.hex").Exist())

	exitCode, d = executeWithArgs("compile", "-b", "arduino:avr:nano", "-o", tmp+"/anothertest/test2", sketchPath)
	require.Zero(t, exitCode)
	require.Contains(t, string(d), "Sketch uses")
	require.True(t, paths.New("anothertest", "test2.hex").Exist())
}

func TestInvalidCoreURLIntegration(t *testing.T) {
	configFile := filepath.Join("testdata", t.Name())

	// Dump config with cmd-line specific file
	exitCode, d := executeWithArgs("--config-file", configFile, "config", "dump")
	require.Zero(t, exitCode)
	require.Contains(t, string(d), "- http://www.invalid-domain-asjkdakdhadjkh.com/package_example_index.json")

	// Update inexistent index
	exitCode, _ = executeWithArgs("--config-file", configFile, "core", "update-index")
	require.NotZero(t, exitCode)
}

func Test3rdPartyCoreIntegration(t *testing.T) {
	configFile := filepath.Join("testdata", t.Name())

	// Update index and install esp32:esp32
	exitCode, _ := executeWithArgs("--config-file", configFile, "core", "update-index")
	require.Zero(t, exitCode)
	exitCode, d := executeWithArgs("--config-file", configFile, "core", "install", "esp32:esp32")
	require.Zero(t, exitCode)
	require.Contains(t, string(d), "installed")

	// Build a simple sketch and check if all artifacts are copied
	tmp := tmpDirOrDie()
	defer os.RemoveAll(tmp)
	tmpSketch := paths.New(tmp).Join("Blink")
	err := paths.New("testdata/Blink").CopyDirTo(tmpSketch)
	require.NoError(t, err, "copying test sketch into temp dir")
	exitCode, d = executeWithArgs("--config-file", configFile, "compile", "-b", "esp32:esp32:esp32", tmpSketch.String())
	require.Zero(t, exitCode)
	require.True(t, tmpSketch.Join("Blink.esp32.esp32.esp32.bin").Exist())
	require.True(t, tmpSketch.Join("Blink.esp32.esp32.esp32.elf").Exist())
	require.True(t, tmpSketch.Join("Blink.esp32.esp32.esp32.partitions.bin").Exist()) // https://github.com/arduino/arduino-cli/issues/163
}

func TestCoreCommandsIntegration(t *testing.T) {
	exitCode, _ := executeWithArgs("core", "update-index")
	require.Zero(t, exitCode)

	AVR := "arduino:avr@" + detectLatestAVRCore(t)

	// Download a specific core version
	exitCode, d := executeWithArgs("core", "download", "arduino:avr@1.6.16")
	require.Zero(t, exitCode)
	require.Regexp(t, "arduino:avr-gcc@4.9.2-atmel3.5.3-arduino2 (already )?downloaded", string(d))
	require.Regexp(t, "arduino:avrdude@6.3.0-arduino8 (already )?downloaded", string(d))
	require.Regexp(t, "arduino:arduinoOTA@1.0.0 (already )?downloaded", string(d))
	require.Regexp(t, "arduino:avr@1.6.16 (already )?downloaded", string(d))

	// Download latest
	exitCode, d = executeWithArgs("core", "download", "arduino:avr")
	require.Zero(t, exitCode)
	require.Regexp(t, AVR+" (already )?downloaded", string(d))

	// Wrong downloads
	exitCode, d = executeWithArgs("core", "download", "arduino:samd@1.2.3-notexisting")
	require.NotZero(t, exitCode)
	require.Contains(t, string(d), "required version 1.2.3-notexisting not found for platform arduino:samd")

	exitCode, d = executeWithArgs("core", "download", "arduino:notexistent")
	require.NotZero(t, exitCode)
	require.Contains(t, string(d), "not found")

	exitCode, d = executeWithArgs("core", "download", "wrongparameter")
	require.NotZero(t, exitCode)
	require.Contains(t, string(d), "invalid item")

	// Install avr 1.6.16
	exitCode, d = executeWithArgs("core", "install", "arduino:avr@1.6.16")
	require.Zero(t, exitCode)
	require.Contains(t, string(d), "arduino:avr@1.6.16 installed")

	exitCode, d = executeWithArgs("core", "list")
	require.Zero(t, exitCode)
	require.Contains(t, string(d), "arduino:avr")
	require.Contains(t, string(d), "1.6.16")

	exitCode, d = executeWithArgs("core", "list", "--format", "json")
	require.Zero(t, exitCode)
	require.Contains(t, string(d), "arduino:avr")
	require.Contains(t, string(d), "1.6.16")

	// Replace avr with 1.6.17
	exitCode, d = executeWithArgs("core", "install", "arduino:avr@1.6.17")
	require.Zero(t, exitCode)
	require.Contains(t, string(d), "Updating arduino:avr@1.6.16 with arduino:avr@1.6.17")
	require.Contains(t, string(d), "arduino:avr@1.6.17 installed")

	exitCode, d = executeWithArgs("core", "list")
	require.Zero(t, exitCode)
	require.Contains(t, string(d), "arduino:avr")
	require.Contains(t, string(d), "1.6.17")

	// List updatable cores
	exitCode, d = executeWithArgs("core", "list", "--updatable")
	require.Zero(t, exitCode)
	require.Contains(t, string(d), "arduino:avr")

	exitCode, d = executeWithArgs("core", "list", "--updatable", "--format", "json")
	require.Zero(t, exitCode)
	require.Contains(t, string(d), "arduino:avr")

	// Upgrade platform
	exitCode, d = executeWithArgs("core", "upgrade", "arduino:avr@1.6.18")
	require.NotZero(t, exitCode)
	require.Contains(t, string(d), "Invalid item arduino:avr@1.6.18")

	exitCode, d = executeWithArgs("core", "upgrade", "other:avr")
	require.NotZero(t, exitCode)
	require.Contains(t, string(d), "other:avr not found")

	exitCode, d = executeWithArgs("core", "upgrade", "arduino:samd")
	require.NotZero(t, exitCode)
	require.Contains(t, string(d), "arduino:samd is not installed")

	exitCode, d = executeWithArgs("core", "upgrade", "arduino:avr")
	require.Zero(t, exitCode)
	require.Contains(t, string(d), "Updating arduino:avr@1.6.17 with "+AVR)

	// List updatable cores
	exitCode, d = executeWithArgs("core", "list", "--updatable")
	require.Zero(t, exitCode)
	require.NotContains(t, string(d), "arduino:avr")

	exitCode, d = executeWithArgs("core", "list")
	require.Zero(t, exitCode)
	require.Contains(t, string(d), "arduino:avr")

	// Uninstall arduino:avr
	exitCode, d = executeWithArgs("core", "uninstall", "arduino:avr")
	require.Zero(t, exitCode)
	require.Contains(t, string(d), AVR+" uninstalled")
}

func TestSearchConfigTreeNotFound(t *testing.T) {
	tmp := tmpDirOrDie()
	require.Empty(t, searchConfigTree(tmp))
}

func TestSearchConfigTreeSameFolder(t *testing.T) {
	tmp := tmpDirOrDie()
	defer os.RemoveAll(tmp)
	_, err := os.Create(filepath.Join(tmp, "arduino-cli.yaml"))
	require.Nil(t, err)
	require.Equal(t, searchConfigTree(tmp), tmp)
}

func TestSearchConfigTreeInParent(t *testing.T) {
	tmp := tmpDirOrDie()
	defer os.RemoveAll(tmp)
	target := filepath.Join(tmp, "foo", "bar")
	err := os.MkdirAll(target, os.ModePerm)
	require.Nil(t, err)
	_, err = os.Create(filepath.Join(tmp, "arduino-cli.yaml"))
	require.Nil(t, err)
	require.Equal(t, searchConfigTree(target), tmp)
}

var result string

func BenchmarkSearchConfigTree(b *testing.B) {
	tmp := tmpDirOrDie()
	defer os.RemoveAll(tmp)
	target := filepath.Join(tmp, "foo", "bar", "baz")
	os.MkdirAll(target, os.ModePerm)

	var s string
	for n := 0; n < b.N; n++ {
		s = searchConfigTree(target)
	}
	result = s
}
