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

	"github.com/arduino/arduino-cli/internal/integrationtest"
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
