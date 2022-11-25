// This file is part of arduino-cli.
//
// Copyright 2022 ARDUINO SA (http://www.arduino.cc/)
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

package upload_test

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/arduino/arduino-cli/internal/integrationtest"
	"github.com/arduino/go-paths-helper"
	"github.com/stretchr/testify/require"
	"go.bug.st/testifyjson/requirejson"
)

type board struct {
	address      string
	fqbn         string
	pack         string
	architecture string
	id           string
	core         string
}

func detectedBoards(t *testing.T, cli *integrationtest.ArduinoCLI) []board {
	// This fixture provides a list of all the boards attached to the host.
	// This fixture will parse the JSON output of `arduino-cli board list --format json`
	//to extract all the connected boards data.

	// :returns a list `Board` objects.

	var boards []board
	stdout, _, err := cli.Run("board", "list", "--format", "json")
	require.NoError(t, err)
	len, err := strconv.Atoi(requirejson.Parse(t, stdout).Query(".[] | .matching_boards | length").String())
	require.NoError(t, err)
	for i := 0; i < len; i++ {
		fqbn := strings.Trim(requirejson.Parse(t, stdout).Query(".[] | .matching_boards | .["+fmt.Sprint(i)+"] | .fqbn").String(), "\"")
		boards = append(boards, board{
			address:      strings.Trim(requirejson.Parse(t, stdout).Query(".[] | .port | .address").String(), "\""),
			fqbn:         fqbn,
			pack:         strings.Split(fqbn, ":")[0],
			architecture: strings.Split(fqbn, ":")[1],
			id:           strings.Split(fqbn, ":")[2],
			core:         strings.Split(fqbn, ":")[0] + ":" + strings.Split(fqbn, ":")[1],
		})
	}
	return boards
}

func waitForBoard(t *testing.T, cli *integrationtest.ArduinoCLI) {
	timeEnd := time.Now().Unix() + 10
	for time.Now().Unix() < timeEnd {
		stdout, _, err := cli.Run("board", "list", "--format", "json")
		require.NoError(t, err)
		len, err := strconv.Atoi(requirejson.Parse(t, stdout).Query("length").String())
		require.NoError(t, err)
		numBoards := 0
		for i := 0; i < len; i++ {
			numBoards, err = strconv.Atoi(requirejson.Parse(t, stdout).Query(".[] | .matching_boards | length").String())
			require.NoError(t, err)
			if numBoards > 0 {
				break
			}
		}
		if numBoards > 0 {
			break
		}
	}
}

func TestUpload(t *testing.T) {
	if os.Getenv("CI") != "" {
		t.Skip("VMs have no serial ports")
	}

	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	// Init the environment explicitly
	_, _, err := cli.Run("core", "update-index")
	require.NoError(t, err)

	for _, board := range detectedBoards(t, cli) {
		// Download platform
		_, _, err = cli.Run("core", "install", board.core)
		require.NoError(t, err)
		// Create a sketch
		sketchName := "TestUploadSketch" + board.id
		sketchPath := cli.SketchbookDir().Join(sketchName)
		fqbn := board.fqbn
		address := board.address
		_, _, err = cli.Run("sketch", "new", sketchPath.String())
		require.NoError(t, err)
		// Build sketch
		_, _, err = cli.Run("compile", "-b", fqbn, sketchPath.String())
		require.NoError(t, err)

		// Verifies binaries are not exported
		require.NoFileExists(t, sketchPath.Join("build").String())

		// Upload without port must fail
		_, _, err = cli.Run("upload", "-b", fqbn, sketchPath.String())
		require.Error(t, err)

		// Upload
		_, _, err = cli.Run("upload", "-b", fqbn, "-p", address, sketchPath.String())
		require.NoError(t, err)
	}
}

func TestUploadWithInputDirFlag(t *testing.T) {
	if os.Getenv("CI") != "" {
		t.Skip("VMs have no serial ports")
	}

	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	// Init the environment explicitly
	_, _, err := cli.Run("core", "update-index")
	require.NoError(t, err)

	for _, board := range detectedBoards(t, cli) {
		// Download board platform
		_, _, err = cli.Run("core", "install", board.core)
		require.NoError(t, err)

		// Create a sketch
		sketchName := "TestUploadSketch" + board.id
		sketchPath := cli.SketchbookDir().Join(sketchName)
		fqbn := board.fqbn
		address := board.address
		_, _, err = cli.Run("sketch", "new", sketchPath.String())
		require.NoError(t, err)

		// Build sketch and export binaries to custom directory
		outputDir := cli.SketchbookDir().Join("test_dir", sketchName, "build")
		_, _, err = cli.Run("compile", "-b", fqbn, sketchPath.String(), "--output-dir", outputDir.String())
		require.NoError(t, err)

		// Upload with --input-dir flag
		_, _, err = cli.Run("upload", "-b", fqbn, "-p", address, "--input-dir", outputDir.String())
		require.NoError(t, err)
	}
}

func TestUploadWithInputFileFlag(t *testing.T) {
	if os.Getenv("CI") != "" {
		t.Skip("VMs have no serial ports")
	}

	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	// Init the environment explicitly
	_, _, err := cli.Run("core", "update-index")
	require.NoError(t, err)

	for _, board := range detectedBoards(t, cli) {
		// Download board platform
		_, _, err = cli.Run("core", "install", board.core)
		require.NoError(t, err)

		// Create a sketch
		sketchName := "TestUploadSketch" + board.id
		sketchPath := cli.SketchbookDir().Join(sketchName)
		fqbn := board.fqbn
		address := board.address
		_, _, err = cli.Run("sketch", "new", sketchPath.String())
		require.NoError(t, err)

		// Build sketch and export binaries to custom directory
		outputDir := cli.SketchbookDir().Join("test_dir", sketchName, "build")
		_, _, err = cli.Run("compile", "-b", fqbn, sketchPath.String(), "--output-dir", outputDir.String())
		require.NoError(t, err)

		// We don't need a specific file when using the --input-file flag to upload since
		// it's just used to calculate the directory, so it's enough to get a random file
		// that's inside that directory
		inputFile := outputDir.Join(sketchName + ".ino.bin")
		// Upload using --input-file
		_, _, err = cli.Run("upload", "-b", fqbn, "-p", address, "--input-file", inputFile.String())
		require.NoError(t, err)
	}
}

func TestCompileAndUploadCombo(t *testing.T) {
	if os.Getenv("CI") != "" {
		t.Skip("VMs have no serial ports")
	}

	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	// Init the environment explicitly
	_, _, err := cli.Run("core", "update-index")
	require.NoError(t, err)

	// Create a test sketch
	sketchName := "CompileAndUploadIntegrationTest"
	sketchPath := cli.SketchbookDir().Join(sketchName)
	sketchMainFile := sketchPath.Join(sketchName + ".ino")
	stdout, _, err := cli.Run("sketch", "new", sketchPath.String())
	require.NoError(t, err)
	require.Contains(t, string(stdout), "Sketch created in: "+sketchPath.String())

	// Build sketch for each detected board
	for _, board := range detectedBoards(t, cli) {
		logFileName := strings.ReplaceAll(board.fqbn, ":", "-") + "-compile.log"
		logFilePath := cli.SketchbookDir().Join(logFileName)

		_, _, err = cli.Run("core", "install", board.core)
		require.NoError(t, err)

		runTest := func(s string) {
			waitForBoard(t, cli)
			_, _, err := cli.Run("compile", "-b", board.fqbn, "--upload", "-p", board.address, s,
				"--log-format", "json", "--log-file", logFilePath.String(), "--log-level", "trace")
			require.NoError(t, err)
			logJson, err := logFilePath.ReadFile()
			require.NoError(t, err)

			// check from the logs if the bin file were uploaded on the current board
			logJson = []byte("[" + strings.ReplaceAll(strings.TrimSuffix(string(logJson), "\n"), "\n", ",") + "]")
			traces := requirejson.Parse(t, logJson).Query("[ .[] | select(.level==\"trace\") | .msg ]").String()
			traces = strings.ReplaceAll(traces, "\\\\", "\\")
			require.Contains(t, traces, "Compile "+sketchPath.String()+" for "+board.fqbn+" started")
			require.Contains(t, traces, "Compile "+sketchName+" for "+board.fqbn+" successful")
			require.Contains(t, traces, "Upload "+sketchPath.String()+" on "+board.fqbn+" started")
			require.Contains(t, traces, "Upload successful")
		}

		runTest(sketchPath.String())
		runTest(sketchMainFile.String())
	}
}

func TestCompileAndUploadComboWithCustomBuildPath(t *testing.T) {
	if os.Getenv("CI") != "" {
		t.Skip("VMs have no serial ports")
	}

	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	// Init the environment explicitly
	_, _, err := cli.Run("core", "update-index")
	require.NoError(t, err)

	// Create a test sketch
	sketchName := "CompileAndUploadCustomBuildPathIntegrationTest"
	sketchPath := cli.SketchbookDir().Join(sketchName)
	_, _, err = cli.Run("sketch", "new", sketchPath.String())
	require.NoError(t, err)

	// Build sketch for each detected board
	for _, board := range detectedBoards(t, cli) {
		fqbnNormalized := strings.ReplaceAll(board.fqbn, ":", "-")
		logFileName := fqbnNormalized + "-compile.log"
		logFilePath := cli.SketchbookDir().Join(logFileName)

		_, _, err = cli.Run("core", "install", board.core)
		require.NoError(t, err)

		waitForBoard(t, cli)

		buildPath := cli.SketchbookDir().Join("test_dir", fqbnNormalized, "build_dir")
		_, _, err := cli.Run("compile", "-b", board.fqbn, "--upload", "-p", board.address, "--build-path", buildPath.String(),
			sketchPath.String(), "--log-format", "json", "--log-file", logFilePath.String(), "--log-level", "trace")
		require.NoError(t, err)
		logJson, err := logFilePath.ReadFile()
		require.NoError(t, err)

		// check from the logs if the bin file were uploaded on the current board
		logJson = []byte("[" + strings.ReplaceAll(strings.TrimSuffix(string(logJson), "\n"), "\n", ",") + "]")
		traces := requirejson.Parse(t, logJson).Query("[ .[] | select(.level==\"trace\") | .msg ]").String()
		traces = strings.ReplaceAll(traces, "\\\\", "\\")
		require.Contains(t, traces, "Compile "+sketchPath.String()+" for "+board.fqbn+" started")
		require.Contains(t, traces, "Compile "+sketchName+" for "+board.fqbn+" successful")
		require.Contains(t, traces, "Upload "+sketchPath.String()+" on "+board.fqbn+" started")
		require.Contains(t, traces, "Upload successful")
	}
}

func TestCompileAndUploadComboSketchWithPdeExtension(t *testing.T) {
	if os.Getenv("CI") != "" {
		t.Skip("VMs have no serial ports")
	}

	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	_, _, err := cli.Run("update")
	require.NoError(t, err)

	sketchName := "CompileAndUploadPdeSketch"
	sketchPath := cli.SketchbookDir().Join(sketchName)

	// Create a test sketch
	_, _, err = cli.Run("sketch", "new", sketchPath.String())
	require.NoError(t, err)

	// Renames sketch file to pde
	sketchFile := sketchPath.Join(sketchName + ".pde")
	require.NoError(t, sketchPath.Join(sketchName+".ino").Rename(sketchFile))

	for _, board := range detectedBoards(t, cli) {
		// Install core
		_, _, err = cli.Run("core", "install", board.core)
		require.NoError(t, err)

		// Build sketch and upload from folder
		waitForBoard(t, cli)
		_, stderr, err := cli.Run("compile", "--clean", "-b", board.fqbn, "-u", "-p", board.address, sketchPath.String())
		require.NoError(t, err)
		require.Contains(t, string(stderr), "Sketches with .pde extension are deprecated, please rename the following files to .ino")
		require.Contains(t, string(stderr), sketchFile.String())

		// Build sketch and upload from file
		waitForBoard(t, cli)
		_, stderr, err = cli.Run("compile", "--clean", "-b", board.fqbn, "-u", "-p", board.address, sketchFile.String())
		require.NoError(t, err)
		require.Contains(t, string(stderr), "Sketches with .pde extension are deprecated, please rename the following files to .ino")
		require.Contains(t, string(stderr), sketchFile.String())
	}
}

func TestUploadSketchWithPdeExtension(t *testing.T) {
	if os.Getenv("CI") != "" {
		t.Skip("VMs have no serial ports")
	}

	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	_, _, err := cli.Run("update")
	require.NoError(t, err)

	sketchName := "UploadPdeSketch"
	sketchPath := cli.SketchbookDir().Join(sketchName)

	// Create a test sketch
	_, _, err = cli.Run("sketch", "new", sketchPath.String())
	require.NoError(t, err)

	// Renames sketch file to pde
	sketchFile := sketchPath.Join(sketchName + ".pde")
	require.NoError(t, sketchPath.Join(sketchName+".ino").Rename(sketchFile))

	for _, board := range detectedBoards(t, cli) {
		// Install core
		_, _, err = cli.Run("core", "install", board.core)
		require.NoError(t, err)

		// Compile sketch first
		stdout, _, err := cli.Run("compile", "--clean", "-b", board.fqbn, sketchPath.String(), "--format", "json")
		require.NoError(t, err)
		buildDir := requirejson.Parse(t, stdout).Query(".builder_result | .build_path").String()
		buildDir = strings.Trim(strings.ReplaceAll(buildDir, "\\\\", "\\"), "\"")

		// Upload from sketch folder
		waitForBoard(t, cli)
		_, _, err = cli.Run("upload", "-b", board.fqbn, "-p", board.address, sketchPath.String())
		require.NoError(t, err)

		// Upload from sketch file
		waitForBoard(t, cli)
		_, _, err = cli.Run("upload", "-b", board.fqbn, "-p", board.address, sketchFile.String())
		require.NoError(t, err)

		waitForBoard(t, cli)
		_, stderr, err := cli.Run("upload", "-b", board.fqbn, "-p", board.address, "--input-dir", buildDir)
		require.NoError(t, err)
		require.Contains(t, string(stderr), "Sketches with .pde extension are deprecated, please rename the following files to .ino:")

		// Upload from binary file
		waitForBoard(t, cli)
		// We don't need a specific file when using the --input-file flag to upload since
		// it's just used to calculate the directory, so it's enough to get a random file
		// that's inside that directory
		binaryFile := paths.New(buildDir, sketchName+".pde.bin")
		_, stderr, err = cli.Run("upload", "-b", board.fqbn, "-p", board.address, "--input-file", binaryFile.String())
		require.NoError(t, err)
		require.Contains(t, string(stderr), "Sketches with .pde extension are deprecated, please rename the following files to .ino:")
	}
}

func TestUploadWithInputDirContainingMultipleBinaries(t *testing.T) {
	if os.Getenv("CI") != "" {
		t.Skip("VMs have no serial ports")
	}

	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	// This tests verifies the behaviour outlined in this issue:
	// https://github.com/arduino/arduino-cli/issues/765#issuecomment-699678646
	_, _, err := cli.Run("update")
	require.NoError(t, err)

	// Create two different sketches
	sketchOneName := "UploadMultipleBinariesSketchOne"
	sketchOnePath := cli.SketchbookDir().Join(sketchOneName)
	_, _, err = cli.Run("sketch", "new", sketchOnePath.String())
	require.NoError(t, err)

	sketchTwoName := "UploadMultipleBinariesSketchTwo"
	sketchTwoPath := cli.SketchbookDir().Join(sketchTwoName)
	_, _, err = cli.Run("sketch", "new", sketchTwoPath.String())
	require.NoError(t, err)

	for _, board := range detectedBoards(t, cli) {
		// Install core
		_, _, err = cli.Run("core", "install", board.core)
		require.NoError(t, err)

		// Compile both sketches and copy binaries in the same build directory
		binariesDir := cli.SketchbookDir().Join("build", "BuiltBinaries")
		_, _, err = cli.Run("compile", "--clean", "-b", board.fqbn, sketchOnePath.String(), "--build-path", binariesDir.String())
		require.NoError(t, err)
		stdout, _, err := cli.Run("compile", "--clean", "-b", board.fqbn, sketchTwoPath.String(), "--format", "json")
		require.NoError(t, err)
		buildDirTwo := requirejson.Parse(t, stdout).Query(".builder_result | .build_path").String()
		buildDirTwo = strings.Trim(strings.ReplaceAll(buildDirTwo, "\\\\", "\\"), "\"")
		require.NoError(t, paths.New(buildDirTwo).Join(sketchTwoName+".ino.bin").CopyTo(binariesDir.Join(sketchTwoName+".ino.bin")))

		waitForBoard(t, cli)
		// Verifies upload fails because multiple binaries are found
		_, stderr, err := cli.Run("upload", "-b", board.fqbn, "-p", board.address, "--input-dir", binariesDir.String())
		require.Error(t, err)
		require.Contains(t, string(stderr), "Error during Upload: ")
		require.Contains(t, string(stderr), "Error finding build artifacts: ")
		require.Contains(t, string(stderr), "autodetect build artifact: ")
		require.Contains(t, string(stderr), "multiple build artifacts found:")

		// Copy binaries to folder with same name of a sketch
		binariesDirSketch := cli.SketchbookDir().Join("build", "UploadMultipleBinariesSketchOne")
		require.NoError(t, binariesDir.CopyDirTo(binariesDirSketch))

		waitForBoard(t, cli)
		// Verifies upload is successful using the binaries with the same name of the containing folder
		_, _, err = cli.Run("upload", "-b", board.fqbn, "-p", board.address, "--input-dir", binariesDirSketch.String())
		require.NoError(t, err)
	}
}
