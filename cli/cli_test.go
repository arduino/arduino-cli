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

package cli

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"bou.ke/monkey"
	paths "github.com/arduino/go-paths-helper"
	"github.com/stretchr/testify/require"
	semver "go.bug.st/relaxed-semver"
)

var (
	// Redirecting stdOut so we can analyze output line by
	// line and check with what we want.
	stdOut = os.Stdout

	currDownloadDir   string
	currDataDir       string
	currSketchbookDir string
)

type stdOutRedirect struct {
	tempFile *os.File
}

func (grabber *stdOutRedirect) Open() {
	tempFile, err := ioutil.TempFile(os.TempDir(), "test")
	if err != nil {
		panic("Opening temp output file")
	}
	os.Stdout = tempFile
	grabber.tempFile = tempFile
}

func (grabber *stdOutRedirect) GetOutput() []byte {
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

func (grabber *stdOutRedirect) Close() {
	grabber.tempFile.Close()
	err := os.Remove(grabber.tempFile.Name())
	if err != nil {
		panic("Removing temp output file")
	}
	os.Stdout = stdOut
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
	os.Setenv("ARDUINO_DATA_DIR", currDataDir)
	currDownloadDir = tmpDirOrDie()
	os.Setenv("ARDUINO_DOWNLOADS_DIR", currDownloadDir)
	currSketchbookDir = filepath.Join("testdata", "sketchbook_with_custom_hardware")
	os.Setenv("ARDUINO_SKETCHBOOK_DIR", currSketchbookDir)

	// Run
	res := m.Run()

	// TearDown
	os.RemoveAll(currDataDir)
	os.Unsetenv("ARDUINO_DATA_DIR")
	currDataDir = ""
	os.RemoveAll(currDownloadDir)
	os.Unsetenv("ARDUINO_DOWNLOADS_DIR")
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

	// This closure is here because we won't that the defer are executed after the end of the "executeWithArgs" method
	func() {
		// Create an empty config for the CLI test in the same dir of the test
		conf := paths.New("arduino-cli.yaml")
		if conf.Exist() {
			panic("config file must not exist already")
		}

		if err := conf.WriteFile([]byte("board_manager:\n  additional_urls:\n")); err != nil {
			panic(err)
		}

		defer func() {
			if err := conf.Remove(); err != nil {
				panic(err)
			}
		}()

		redirect := &stdOutRedirect{}
		redirect.Open()
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
		ArduinoCli.Execute()
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
	fmt.Println("Latest AVR core version:", latest)
	return latest.String()
}

// END -- Utility functions

func TestUploadIntegration(t *testing.T) {
	exitCode, _ := executeWithArgs("core", "update-index")
	require.Zero(t, exitCode)

	exitCode, _ = executeWithArgs("core", "install", "arduino:avr")
	require.Zero(t, exitCode)

	// -i flag
	exitCode, d := executeWithArgs("upload", "-i", filepath.Join(currSketchbookDir, "test.hex"), "-b", "test:avr:testboard", "-p", "/dev/ttyACM0")
	require.Zero(t, exitCode)
	require.Contains(t, string(d), "QUIET")
	require.Contains(t, string(d), "NOVERIFY")
	require.Contains(t, string(d), "testdata/sketchbook_with_custom_hardware/test.hex")

	// -i flag with implicit extension
	exitCode, d = executeWithArgs("upload", "-i", filepath.Join(currSketchbookDir, "test"), "-b", "test:avr:testboard", "-p", "/dev/ttyACM0")
	require.Zero(t, exitCode)
	require.Contains(t, string(d), "QUIET")
	require.Contains(t, string(d), "NOVERIFY")
	require.Contains(t, string(d), "testdata/sketchbook_with_custom_hardware/test.hex")

	// -i with absolute path
	fullPath, err := filepath.Abs(filepath.Join(currSketchbookDir, "test.hex"))
	require.NoError(t, err)
	exitCode, d = executeWithArgs("upload", "-i", fullPath, "-b", "test:avr:testboard", "-p", "/dev/ttyACM0")
	require.Zero(t, exitCode)
	require.Contains(t, string(d), "QUIET")
	require.Contains(t, string(d), "NOVERIFY")
	require.Contains(t, string(d), "testdata/sketchbook_with_custom_hardware/test.hex")

	// -v verbose
	exitCode, d = executeWithArgs("upload", "-v", "-t", "-i", filepath.Join(currSketchbookDir, "test.hex"), "-b", "test:avr:testboard", "-p", "/dev/ttyACM0")
	require.Zero(t, exitCode)
	require.Contains(t, string(d), "VERBOSE")
	require.Contains(t, string(d), "VERIFY")
	require.Contains(t, string(d), "testdata/sketchbook_with_custom_hardware/test.hex")

	// -t verify
	exitCode, d = executeWithArgs("upload", "-i", filepath.Join(currSketchbookDir, "test.hex"), "-b", "test:avr:testboard", "-p", "/dev/ttyACM0")
	require.Zero(t, exitCode)
	require.Contains(t, string(d), "QUIET")
	require.Contains(t, string(d), "NOVERIFY")
	require.Contains(t, string(d), "testdata/sketchbook_with_custom_hardware/test.hex")

	// -v -t verbose verify
	exitCode, d = executeWithArgs("upload", "-v", "-t", "-i", filepath.Join(currSketchbookDir, "test.hex"), "-b", "test:avr:testboard", "-p", "/dev/ttyACM0")
	require.Zero(t, exitCode)
	require.Contains(t, string(d), "VERBOSE")
	require.Contains(t, string(d), "VERIFY")
	require.Contains(t, string(d), "testdata/sketchbook_with_custom_hardware/test.hex")

	// non-existent file
	exitCode, _ = executeWithArgs("upload", "-i", filepath.Join(currSketchbookDir, "test123.hex"), "-b", "test:avr:testboard", "-p", "/dev/ttyACM0")
	require.NotZero(t, exitCode)

	// sketch
	exitCode, d = executeWithArgs("upload", filepath.Join(currSketchbookDir, "TestSketch"), "-b", "test:avr:testboard", "-p", "/dev/ttyACM0")
	require.Zero(t, exitCode)
	require.Contains(t, string(d), "QUIET")
	require.Contains(t, string(d), "NOVERIFY")
	require.Contains(t, string(d), "testdata/sketchbook_with_custom_hardware/TestSketch/TestSketch.test.avr.testboard.hex")

	// sketch without build
	exitCode, _ = executeWithArgs("upload", filepath.Join(currSketchbookDir, "TestSketch2"), "-b", "test:avr:testboard", "-p", "/dev/ttyACM0")
	require.NotZero(t, exitCode)

	// platform without 'recipe.output.tmp_file' property
	exitCode, _ = executeWithArgs("upload", "-i", filepath.Join(currSketchbookDir, "test.hex"), "-b", "test2:avr:testboard", "-p", "/dev/ttyACM0")
	require.NotZero(t, exitCode)
}

func TestLibSearchIntegration(t *testing.T) {
	exitCode, output := executeWithArgs("lib", "search", "audiozer", "--format", "json")
	require.Zero(t, exitCode)
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

	exitCode, output = executeWithArgs("lib", "search", "audiozero", "--names")
	require.Zero(t, exitCode, "process exit code")
	require.Equal(t, "AudioZero\n", string(output))

	exitCode, output = executeWithArgs("lib", "search", "audiozer", "--names")
	require.Zero(t, exitCode, "process exit code")
	require.Equal(t, "AudioZero\n", string(output))

	exitCode, output = executeWithArgs("lib", "search", "audiozerooooo", "--names")
	require.Zero(t, exitCode, "process exit code")
	require.Equal(t, "", string(output))
}

func TestLibUserIntegration(t *testing.T) {
	// source of test custom libs
	libDir := filepath.Join("testdata", "libs")
	// install destination
	libInstallDir := filepath.Join(currSketchbookDir, "libraries")
	require.NoError(t, os.MkdirAll(libInstallDir, os.FileMode(0755)))
	defer os.RemoveAll(libInstallDir)

	installLib := func(lib string) {
		cmd := exec.Command("cp", "-r", "-a", filepath.Join(libDir, lib), libInstallDir)
		require.NoError(t, cmd.Run(), cmd.Args)
	}

	// List libraries (valid libs)
	installLib("MyLib")
	exitCode, d := executeWithArgs("lib", "list")
	require.Zero(t, exitCode)
	require.Contains(t, string(d), "MyLib")
	require.Contains(t, string(d), "1.0.5")

	// List libraries (pre-1.5 format)
	installLib("MyLibPre15")
	exitCode, d = executeWithArgs("lib", "list")
	require.Zero(t, exitCode)
	require.Contains(t, string(d), "MyLibPre15")

	// List libraries (invalid version lib)
	installLib("MyLibWithWrongVersion")
	exitCode, d = executeWithArgs("lib", "list")
	require.Zero(t, exitCode)
	require.Contains(t, string(d), "MyLibWithWrongVersion")
}

func TestLibDownloadAndInstallIntegration(t *testing.T) {
	exitCode, _ := executeWithArgs("core", "update-index")
	require.Zero(t, exitCode)

	// Download inexistent
	exitCode, d := executeWithArgs("lib", "download", "inexistentLibrary", "--format", "json")
	require.NotZero(t, exitCode)
	require.Contains(t, string(d), "library inexistentLibrary not found")

	exitCode, d = executeWithArgs("lib", "download", "inexistentLibrary")
	require.NotZero(t, exitCode)
	require.Contains(t, string(d), "library inexistentLibrary not found")

	// Download latest
	exitCode, d = executeWithArgs("lib", "download", "Audio")
	require.Zero(t, exitCode)
	require.Contains(t, string(d), "Audio@")
	require.Contains(t, string(d), "downloaded")

	// Download non existent version
	exitCode, d = executeWithArgs("lib", "download", "Audio@1.2.3-nonexistent")
	require.NotZero(t, exitCode)
	require.Contains(t, string(d), "not found")

	// Install latest
	exitCode, d = executeWithArgs("lib", "install", "Audio")
	require.Zero(t, exitCode)
	require.Contains(t, string(d), "Audio@")
	require.Contains(t, string(d), "Installed")

	exitCode, d = executeWithArgs("lib", "list")
	require.Zero(t, exitCode)
	require.Contains(t, string(d), "Audio")

	// Already installed
	exitCode, d = executeWithArgs("lib", "install", "Audio")
	require.NotZero(t, exitCode)
	require.Contains(t, string(d), "Audio@")
	require.Contains(t, string(d), "already installed")

	// Install another version
	exitCode, d = executeWithArgs("lib", "install", "Audio@1.0.4")
	require.Zero(t, exitCode)
	require.Contains(t, string(d), "Audio@1.0.4")
	require.Contains(t, string(d), "Installed")
	exitCode, d = executeWithArgs("lib", "list")
	require.Zero(t, exitCode)
	require.Contains(t, string(d), "Audio")
	require.Contains(t, string(d), "1.0.4")

	// List updatable
	exitCode, d = executeWithArgs("lib", "list", "--updatable")
	require.Zero(t, exitCode)
	require.Contains(t, string(d), "Audio")
	require.Contains(t, string(d), "1.0.4")
	require.Contains(t, string(d), "1.0.5")

	// Uninstall version not installed
	exitCode, d = executeWithArgs("lib", "uninstall", "Audio@1.0.3")
	require.NotZero(t, exitCode)
	require.Contains(t, string(d), "Audio@1.0.3")
	require.Contains(t, string(d), "not installed")

	// Upgrade libraries
	exitCode, d = executeWithArgs("lib", "upgrade")
	require.Zero(t, exitCode)
	require.Contains(t, string(d), "Installed")
	require.Contains(t, string(d), "Audio")
	require.Contains(t, string(d), "1.0.5")

	// Uninstall (without version)
	exitCode, d = executeWithArgs("lib", "uninstall", "Audio")
	require.Zero(t, exitCode)
	require.Contains(t, string(d), "Uninstalling")
	require.Contains(t, string(d), "Audio")
	require.Contains(t, string(d), "1.0.5")
	exitCode, d = executeWithArgs("lib", "list")
	require.Zero(t, exitCode)
	require.NotContains(t, string(d), "Audio")

	// Uninstall (with version)
	exitCode, d = executeWithArgs("lib", "install", "Audio@1.0.4")
	require.Zero(t, exitCode)
	require.Contains(t, string(d), "Audio@1.0.4")
	require.Contains(t, string(d), "Installed")
	exitCode, d = executeWithArgs("lib", "list")
	require.Zero(t, exitCode)
	require.Contains(t, string(d), "Audio")
	require.Contains(t, string(d), "1.0.4")
	exitCode, d = executeWithArgs("lib", "uninstall", "Audio@1.0.4")
	require.Zero(t, exitCode)
	require.Contains(t, string(d), "Uninstalling")
	require.Contains(t, string(d), "Audio")
	require.Contains(t, string(d), "1.0.4")
	exitCode, d = executeWithArgs("lib", "list")
	require.Zero(t, exitCode)
	require.NotContains(t, string(d), "Audio")
}

func TestSketchCommandsIntegration(t *testing.T) {
	exitCode, _ := executeWithArgs("sketch", "new", "Test")
	require.Zero(t, exitCode)
}

func TestCompileCommandsIntegration(t *testing.T) {
	// Set staging dir to a temporary dir
	tmp := tmpDirOrDie()
	defer os.RemoveAll(tmp)

	// override SetUp dirs
	os.Setenv("ARDUINO_SKETCHBOOK_DIR", tmp)
	currSketchbookDir = tmp

	exitCode, _ := executeWithArgs("core", "update-index")
	require.Zero(t, exitCode)

	// Download latest AVR
	exitCode, _ = executeWithArgs("core", "install", "arduino:avr")
	require.Zero(t, exitCode)

	// Create a test sketch
	exitCode, d := executeWithArgs("sketch", "new", "Test1")
	require.Zero(t, exitCode)
	require.Contains(t, string(d), "Sketch created")

	// Build sketch without FQBN
	test1 := filepath.Join(currSketchbookDir, "Test1")
	exitCode, d = executeWithArgs("compile", test1)
	require.NotZero(t, exitCode)
	require.Contains(t, string(d), "no FQBN provided")

	// Build sketch for arduino:avr:uno
	exitCode, d = executeWithArgs("compile", "-b", "arduino:avr:uno", test1)
	require.Zero(t, exitCode)
	require.Contains(t, string(d), "Sketch uses")
	require.True(t, paths.New(test1).Join("Test1.arduino.avr.uno.hex").Exist())

	// Build sketch for arduino:avr:nano (without options)
	exitCode, d = executeWithArgs("compile", "-b", "arduino:avr:nano", test1)
	require.Zero(t, exitCode)
	require.Contains(t, string(d), "Sketch uses")
	require.True(t, paths.New(test1).Join("Test1.arduino.avr.nano.hex").Exist())

	// Build sketch with --output path
	pwd, err := os.Getwd()
	require.NoError(t, err)
	defer func() { require.NoError(t, os.Chdir(pwd)) }()
	require.NoError(t, os.Chdir(tmp))

	exitCode, d = executeWithArgs("compile", "-b", "arduino:avr:nano", "-o", "test", test1)
	require.Zero(t, exitCode)
	require.Contains(t, string(d), "Sketch uses")
	require.True(t, paths.New("test.hex").Exist())

	exitCode, d = executeWithArgs("compile", "-b", "arduino:avr:nano", "-o", "test2.hex", test1)
	require.Zero(t, exitCode)
	require.Contains(t, string(d), "Sketch uses")
	require.True(t, paths.New("test2.hex").Exist())
	require.NoError(t, paths.New(tmp, "anothertest").MkdirAll())

	exitCode, d = executeWithArgs("compile", "-b", "arduino:avr:nano", "-o", "anothertest/test", test1)
	require.Zero(t, exitCode)
	require.Contains(t, string(d), "Sketch uses")
	require.True(t, paths.New("anothertest", "test.hex").Exist())

	exitCode, d = executeWithArgs("compile", "-b", "arduino:avr:nano", "-o", tmp+"/anothertest/test2", test1)
	require.Zero(t, exitCode)
	require.Contains(t, string(d), "Sketch uses")
	require.True(t, paths.New("anothertest", "test2.hex").Exist())
}

func TestInvalidCoreURLIntegration(t *testing.T) {
	// override SetUp dirs
	tmp := tmpDirOrDie()
	defer os.RemoveAll(tmp)
	os.Setenv("ARDUINO_SKETCHBOOK_DIR", tmp)
	currSketchbookDir = tmp

	configFile := filepath.Join(currDataDir, "arduino-cli.yaml")
	err := ioutil.WriteFile(configFile, []byte(`
board_manager:
  additional_urls:
    - http://www.invalid-domain-asjkdakdhadjkh.com/package_example_index.json
`), os.FileMode(0644))
	require.NoError(t, err, "writing dummy config "+configFile)

	err = ioutil.WriteFile(filepath.Join(currDataDir, "package_index.json"), []byte(`{ "packages": [] }`), os.FileMode(0644))
	require.NoError(t, err, "Writing empty json index file")

	err = ioutil.WriteFile(filepath.Join(currDataDir, "package_example_index.json"), []byte(`{ "packages": [] }`), os.FileMode(0644))
	require.NoError(t, err, "Writing empty json index file")

	err = ioutil.WriteFile(filepath.Join(currDataDir, "library_index.json"), []byte(`{ "libraries": [] }`), os.FileMode(0644))
	require.NoError(t, err, "Writing empty json index file")

	// Dump config with cmd-line specific file
	exitCode, d := executeWithArgs("--config-file", configFile, "config", "dump")
	require.Zero(t, exitCode)
	require.Contains(t, string(d), "- http://www.invalid-domain-asjkdakdhadjkh.com/package_example_index.json")

	// Update inexistent index
	exitCode, _ = executeWithArgs("--config-file", configFile, "core", "update-index")
	require.NotZero(t, exitCode)
}

func TestCoreCommandsIntegration(t *testing.T) {
	// override SetUp dirs
	tmp := tmpDirOrDie()
	defer os.RemoveAll(tmp)
	os.Setenv("ARDUINO_SKETCHBOOK_DIR", tmp)
	currSketchbookDir = tmp

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
