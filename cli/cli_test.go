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

package cli_test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"bou.ke/monkey"
	"github.com/arduino/arduino-cli/cli"
	paths "github.com/arduino/go-paths-helper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	semver "go.bug.st/relaxed-semver"
)

// Redirecting stdOut so we can analyze output line by
// line and check with what we want.
var stdOut = os.Stdout

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
func executeWithArgs(t *testing.T, args ...string) (int, []byte) {
	var output []byte
	var exitCode int
	fmt.Printf("RUNNING: %s\n", args)

	// This closure is here because we won't that the defer are executed after the end of the "executeWithArgs" method
	func() {
		// Create an empty config for the CLI test
		conf := paths.New("arduino-cli.yaml")
		require.False(t, conf.Exist())
		err := conf.WriteFile([]byte("board_manager:\n  additional_urls:\n"))
		require.NoError(t, err)
		defer func() {
			require.NoError(t, conf.Remove())
		}()

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
		cli.ArduinoCli.SetArgs(args)
		cli.ArduinoCli.Execute()
	}()

	return exitCode, output
}

var currDownloadDir *paths.Path

func useSharedDownloadDir(t *testing.T) func() {
	tmp := paths.TempDir().Join("arduino-cli-test-staging")
	err := tmp.MkdirAll()
	require.NoError(t, err, "making shared staging dir")
	os.Setenv("ARDUINO_DOWNLOADS_DIR", tmp.String())
	currDownloadDir = tmp
	fmt.Printf("ARDUINO_DOWNLOADS_DIR = %s\n", os.Getenv("ARDUINO_DOWNLOADS_DIR"))

	return func() {
		os.Unsetenv("ARDUINO_DOWNLOADS_DIR")
		currDownloadDir = nil
		fmt.Printf("ARDUINO_DOWNLOADS_DIR = %s\n", os.Getenv("ARDUINO_DOWNLOADS_DIR"))
	}
}

var currDataDir *paths.Path

func makeTempDataDir(t *testing.T) func() {
	tmp, err := paths.MkTempDir("", "test")
	require.NoError(t, err, "making temporary data dir")
	os.Setenv("ARDUINO_DATA_DIR", tmp.String())
	currDataDir = tmp
	fmt.Printf("ARDUINO_DATA_DIR = %s\n", os.Getenv("ARDUINO_DATA_DIR"))

	err = tmp.RemoveAll() // To test if the data dir is automatically created
	require.NoError(t, err)

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
	err = tmp.RemoveAll() // To test if the sketchbook dir is automatically created
	require.NoError(t, err)

	fmt.Printf("ARDUINO_SKETCHBOOK_DIR = %s\n", os.Getenv("ARDUINO_SKETCHBOOK_DIR"))
	return func() {
		os.Unsetenv("ARDUINO_SKETCHBOOK_DIR")
		currSketchbookDir = nil
		tmp.RemoveAll()
		fmt.Printf("ARDUINO_SKETCHBOOK_DIR = %s\n", os.Getenv("ARDUINO_SKETCHBOOK_DIR"))
	}
}

func setSketchbookDir(t *testing.T, tmp *paths.Path) func() {
	os.Setenv("ARDUINO_SKETCHBOOK_DIR", tmp.String())
	currSketchbookDir = tmp

	fmt.Printf("ARDUINO_SKETCHBOOK_DIR = %s\n", os.Getenv("ARDUINO_SKETCHBOOK_DIR"))
	return func() {
		os.Unsetenv("ARDUINO_SKETCHBOOK_DIR")
		currSketchbookDir = nil
		fmt.Printf("ARDUINO_SKETCHBOOK_DIR = %s\n", os.Getenv("ARDUINO_SKETCHBOOK_DIR"))
	}
}

// END -- Utility functions

func TestUploadCommands(t *testing.T) {
	defer makeTempDataDir(t)()
	defer useSharedDownloadDir(t)()
	defer setSketchbookDir(t, paths.New("testdata", "sketchbook_with_custom_hardware"))()

	updateCoreIndex(t)

	exitCode, _ := executeWithArgs(t, "core", "install", "arduino:avr")
	require.Zero(t, exitCode, "exit code")

	// -i flag
	exitCode, d := executeWithArgs(t, "upload", "-i", currSketchbookDir.Join("test.hex").String(), "-b", "test:avr:testboard", "-p", "/dev/ttyACM0")
	require.Zero(t, exitCode, "exit code")
	require.Contains(t, string(d), "QUIET")
	require.Contains(t, string(d), "NOVERIFY")
	require.Contains(t, string(d), "testdata/sketchbook_with_custom_hardware/test.hex")

	// -i flag with implicit extension
	exitCode, d = executeWithArgs(t, "upload", "-i", currSketchbookDir.Join("test").String(), "-b", "test:avr:testboard", "-p", "/dev/ttyACM0")
	require.Zero(t, exitCode, "exit code")
	require.Contains(t, string(d), "QUIET")
	require.Contains(t, string(d), "NOVERIFY")
	require.Contains(t, string(d), "testdata/sketchbook_with_custom_hardware/test.hex")

	// -i with absolute path
	fullPath, err := currSketchbookDir.Join("test.hex").Abs()
	require.NoError(t, err, "absolute path of test.hex")
	exitCode, d = executeWithArgs(t, "upload", "-i", fullPath.String(), "-b", "test:avr:testboard", "-p", "/dev/ttyACM0")
	require.Zero(t, exitCode, "exit code")
	require.Contains(t, string(d), "QUIET")
	require.Contains(t, string(d), "NOVERIFY")
	require.Contains(t, string(d), "testdata/sketchbook_with_custom_hardware/test.hex")

	// -v verbose
	exitCode, d = executeWithArgs(t, "upload", "-v", "-i", currSketchbookDir.Join("test.hex").String(), "-b", "test:avr:testboard", "-p", "/dev/ttyACM0")
	require.Zero(t, exitCode, "exit code")
	require.Contains(t, string(d), "VERBOSE")
	require.Contains(t, string(d), "NOVERIFY")
	require.Contains(t, string(d), "testdata/sketchbook_with_custom_hardware/test.hex")

	// -t verify
	exitCode, d = executeWithArgs(t, "upload", "-t", "-i", currSketchbookDir.Join("test.hex").String(), "-b", "test:avr:testboard", "-p", "/dev/ttyACM0")
	require.Zero(t, exitCode, "exit code")
	require.Contains(t, string(d), "QUIET")
	require.Contains(t, string(d), "VERIFY")
	require.Contains(t, string(d), "testdata/sketchbook_with_custom_hardware/test.hex")

	// -v -t verbose verify
	exitCode, d = executeWithArgs(t, "upload", "-v", "-t", "-i", currSketchbookDir.Join("test.hex").String(), "-b", "test:avr:testboard", "-p", "/dev/ttyACM0")
	require.Zero(t, exitCode, "exit code")
	require.Contains(t, string(d), "VERBOSE")
	require.Contains(t, string(d), "VERIFY")
	require.Contains(t, string(d), "testdata/sketchbook_with_custom_hardware/test.hex")

	// non-existent file
	exitCode, _ = executeWithArgs(t, "upload", "-i", currSketchbookDir.Join("test123.hex").String(), "-b", "test:avr:testboard", "-p", "/dev/ttyACM0")
	require.NotZero(t, exitCode, "exit code")

	// sketch
	exitCode, d = executeWithArgs(t, "upload", currSketchbookDir.Join("TestSketch").String(), "-b", "test:avr:testboard", "-p", "/dev/ttyACM0")
	require.Zero(t, exitCode, "exit code")
	require.Contains(t, string(d), "QUIET")
	require.Contains(t, string(d), "NOVERIFY")
	require.Contains(t, string(d), "testdata/sketchbook_with_custom_hardware/TestSketch/TestSketch.test.avr.testboard.hex")

	// sketch without build
	exitCode, _ = executeWithArgs(t, "upload", currSketchbookDir.Join("TestSketch2").String(), "-b", "test:avr:testboard", "-p", "/dev/ttyACM0")
	require.NotZero(t, exitCode, "exit code")

	// platform without 'recipe.output.tmp_file' property
	exitCode, _ = executeWithArgs(t, "upload", "-i", currSketchbookDir.Join("test.hex").String(), "-b", "test2:avr:testboard", "-p", "/dev/ttyACM0")
	require.NotZero(t, exitCode, "exit code")
}

func TestLibSearch(t *testing.T) {
	defer makeTempDataDir(t)()
	defer makeTempSketchbookDir(t)()
	defer useSharedDownloadDir(t)()

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
	defer useSharedDownloadDir(t)()
	libDir := currSketchbookDir.Join("libraries")
	err := libDir.MkdirAll()
	require.NoError(t, err, "creating 'sketchbook/libraries' dir")

	installLib := func(lib string) {
		libPath := paths.New("testdata", "libs", lib)
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

func TestSketchCommands(t *testing.T) {
	defer makeTempDataDir(t)()
	defer makeTempSketchbookDir(t)()
	defer useSharedDownloadDir(t)()

	exitCode, _ := executeWithArgs(t, "sketch", "new", "Test")
	require.Zero(t, exitCode, "exit code")
}

func TestLibDownloadAndInstall(t *testing.T) {
	defer makeTempDataDir(t)()
	defer makeTempSketchbookDir(t)()
	defer useSharedDownloadDir(t)()

	exitCode, _ := executeWithArgs(t, "core", "update-index")
	require.Zero(t, exitCode, "exit code")

	// Download inexistent
	exitCode, d := executeWithArgs(t, "lib", "download", "inexistentLibrary", "--format", "json")
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

func TestCompileCommands(t *testing.T) {
	defer makeTempDataDir(t)()
	defer makeTempSketchbookDir(t)()
	defer useSharedDownloadDir(t)()

	// Set staging dir to a temporary dir
	tmp, err := ioutil.TempDir(os.TempDir(), "test")
	require.NoError(t, err, "making temporary staging dir")
	defer os.RemoveAll(tmp)

	updateCoreIndex(t)

	// Download latest AVR
	exitCode, _ := executeWithArgs(t, "core", "install", "arduino:avr")
	require.Zero(t, exitCode, "exit code")

	// Create a test sketch
	exitCode, d := executeWithArgs(t, "sketch", "new", "Test1")
	require.Zero(t, exitCode, "exit code")
	require.Contains(t, string(d), "Sketch created")

	// Build sketch without FQBN
	test1 := currSketchbookDir.Join("Test1").String()
	exitCode, d = executeWithArgs(t, "compile", test1)
	require.NotZero(t, exitCode, "exit code")
	require.Contains(t, string(d), "no FQBN provided")

	// Build sketch for arduino:avr:uno
	exitCode, d = executeWithArgs(t, "compile", "-b", "arduino:avr:uno", test1)
	require.Zero(t, exitCode, "exit code")
	require.Contains(t, string(d), "Sketch uses")
	require.True(t, paths.New(test1).Join("Test1.arduino.avr.uno.hex").Exist())

	// Build sketch for arduino:avr:nano (without options)
	exitCode, d = executeWithArgs(t, "compile", "-b", "arduino:avr:nano", test1)
	require.Zero(t, exitCode, "exit code")
	require.Contains(t, string(d), "Sketch uses")
	require.True(t, paths.New(test1).Join("Test1.arduino.avr.nano.hex").Exist())

	// Build sketch with --output path
	{
		pwd, err := os.Getwd()
		require.NoError(t, err)
		defer func() { require.NoError(t, os.Chdir(pwd)) }()
		require.NoError(t, os.Chdir(tmp))

		exitCode, d = executeWithArgs(t, "compile", "-b", "arduino:avr:nano", "-o", "test", test1)
		require.Zero(t, exitCode, "exit code")
		require.Contains(t, string(d), "Sketch uses")
		require.True(t, paths.New("test.hex").Exist())

		exitCode, d = executeWithArgs(t, "compile", "-b", "arduino:avr:nano", "-o", "test2.hex", test1)
		require.Zero(t, exitCode, "exit code")
		require.Contains(t, string(d), "Sketch uses")
		require.True(t, paths.New("test2.hex").Exist())
		require.NoError(t, paths.New(tmp, "anothertest").MkdirAll())

		exitCode, d = executeWithArgs(t, "compile", "-b", "arduino:avr:nano", "-o", "anothertest/test", test1)
		require.Zero(t, exitCode, "exit code")
		require.Contains(t, string(d), "Sketch uses")
		require.True(t, paths.New("anothertest", "test.hex").Exist())

		exitCode, d = executeWithArgs(t, "compile", "-b", "arduino:avr:nano", "-o", tmp+"/anothertest/test2", test1)
		require.Zero(t, exitCode, "exit code")
		require.Contains(t, string(d), "Sketch uses")
		require.True(t, paths.New("anothertest", "test2.hex").Exist())
	}
}

func TestInvalidCoreURL(t *testing.T) {
	defer makeTempDataDir(t)()
	defer makeTempSketchbookDir(t)()
	defer useSharedDownloadDir(t)()

	tmp, err := paths.MkTempDir("", "")
	require.NoError(t, err, "making temporary dir")
	defer tmp.RemoveAll()

	configFile := tmp.Join("arduino-cli.yaml")
	err = configFile.WriteFile([]byte(`
board_manager:
  additional_urls:
    - http://www.invalid-domain-asjkdakdhadjkh.com/package_example_index.json
`))
	require.NoError(t, err, "writing dummy config "+configFile.String())

	require.NoError(t, currDataDir.MkdirAll())
	err = currDataDir.Join("package_index.json").WriteFile([]byte(`{ "packages": [] }`))
	require.NoError(t, err, "Writing empty json index file")
	err = currDataDir.Join("package_example_index.json").WriteFile([]byte(`{ "packages": [] }`))
	require.NoError(t, err, "Writing empty json index file")
	err = currDataDir.Join("library_index.json").WriteFile([]byte(`{ "libraries": [] }`))
	require.NoError(t, err, "Writing empty json index file")

	// Empty cores list
	exitCode, _ := executeWithArgs(t, "--config-file", configFile.String(), "core", "list")
	require.Zero(t, exitCode, "exit code")

	// Dump config with cmd-line specific file
	exitCode, d := executeWithArgs(t, "--config-file", configFile.String(), "config", "dump")
	require.Zero(t, exitCode, "exit code")
	require.Contains(t, string(d), "- http://www.invalid-domain-asjkdakdhadjkh.com/package_example_index.json")

	// Update inexistent index
	exitCode, _ = executeWithArgs(t, "--config-file", configFile.String(), "core", "update-index")
	require.NotZero(t, exitCode, "exit code")

	// Empty cores list
	exitCode, d = executeWithArgs(t, "--config-file", configFile.String(), "core", "list")
	require.Zero(t, exitCode, "exit code")
	require.Empty(t, strings.TrimSpace(string(d)))
}

func TestCoreCommands(t *testing.T) {
	defer makeTempDataDir(t)()
	defer makeTempSketchbookDir(t)()
	defer useSharedDownloadDir(t)()

	updateCoreIndex(t)
	AVR := "arduino:avr@" + detectLatestAVRCore(t)

	// Download a specific core version
	exitCode, d := executeWithArgs(t, "core", "download", "arduino:avr@1.6.16")
	require.Zero(t, exitCode, "exit code")
	require.Regexp(t, "arduino:avr-gcc@4.9.2-atmel3.5.3-arduino2 (already )?downloaded", string(d))
	require.Regexp(t, "arduino:avrdude@6.3.0-arduino8 (already )?downloaded", string(d))
	require.Regexp(t, "arduino:arduinoOTA@1.0.0 (already )?downloaded", string(d))
	require.Regexp(t, "arduino:avr@1.6.16 (already )?downloaded", string(d))

	// Download latest
	exitCode, d = executeWithArgs(t, "core", "download", "arduino:avr")
	require.Zero(t, exitCode, "exit code")
	require.Regexp(t, AVR+" (already )?downloaded", string(d))

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

	exitCode, d = executeWithArgs(t, "core", "list", "--format", "json")
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

	exitCode, d = executeWithArgs(t, "core", "list", "--updatable", "--format", "json")
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

	// Uninstall arduino:avr
	exitCode, d = executeWithArgs(t, "core", "uninstall", "arduino:avr")
	require.Zero(t, exitCode, "exit code")
	require.Contains(t, string(d), AVR+" uninstalled")

	// Empty cores list
	exitCode, d = executeWithArgs(t, "core", "list")
	require.Zero(t, exitCode, "exit code")
	require.Empty(t, strings.TrimSpace(string(d)))
}
