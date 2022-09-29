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

package compile_test

import (
	"strings"
	"testing"

	"github.com/arduino/arduino-cli/internal/integrationtest"
	"github.com/stretchr/testify/require"
)

func TestCompileSketchWithPdeExtension(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	// Init the environment explicitly
	_, _, err := cli.Run("update")
	require.NoError(t, err)

	// Install core to compile
	_, _, err = cli.Run("core", "install", "arduino:avr@1.8.3")
	require.NoError(t, err)

	sketchName := "CompilePdeSketch"
	sketchPath := cli.SketchbookDir().Join(sketchName)
	fqbn := "arduino:avr:uno"

	// Create a test sketch
	_, _, err = cli.Run("sketch", "new", sketchPath.String())
	require.NoError(t, err)

	sketchFileIno := sketchPath.Join(sketchName + ".ino")
	sketchFilePde := sketchPath.Join(sketchName + ".pde")
	err = sketchFileIno.Rename(sketchFilePde)
	require.NoError(t, err)

	// Build sketch from folder
	_, stderr, err := cli.Run("compile", "--clean", "-b", fqbn, sketchPath.String())
	require.NoError(t, err)
	require.Contains(t, string(stderr), "Sketches with .pde extension are deprecated, please rename the following files to .ino:")
	require.Contains(t, string(stderr), sketchFilePde.String())

	// Build sketch from file
	_, stderr, err = cli.Run("compile", "--clean", "-b", fqbn, sketchFilePde.String())
	require.NoError(t, err)
	require.Contains(t, string(stderr), "Sketches with .pde extension are deprecated, please rename the following files to .ino:")
	require.Contains(t, string(stderr), sketchFilePde.String())
}

func TestCompileSketchWithMultipleMainFiles(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	// Init the environment explicitly
	_, _, err := cli.Run("update")
	require.NoError(t, err)

	// Install core to compile
	_, _, err = cli.Run("core", "install", "arduino:avr@1.8.3")
	require.NoError(t, err)

	sketchName := "CompileSketchMultipleMainFiles"
	sketchPath := cli.SketchbookDir().Join(sketchName)
	fqbn := "arduino:avr:uno"

	// Create a test sketch
	_, _, err = cli.Run("sketch", "new", sketchPath.String())
	require.NoError(t, err)

	// Copy .ino sketch file to .pde
	sketchFileIno := sketchPath.Join(sketchName + ".ino")
	sketchFilePde := sketchPath.Join(sketchName + ".pde")
	err = sketchFileIno.CopyTo(sketchFilePde)
	require.NoError(t, err)

	// Build sketch from folder
	_, stderr, err := cli.Run("compile", "--clean", "-b", fqbn, sketchPath.String())
	require.Error(t, err)
	require.Contains(t, string(stderr), "Error opening sketch: multiple main sketch files found")

	// Build sketch from .ino file
	_, stderr, err = cli.Run("compile", "--clean", "-b", fqbn, sketchFileIno.String())
	require.Error(t, err)
	require.Contains(t, string(stderr), "Error opening sketch: multiple main sketch files found")

	// Build sketch from .pde file
	_, stderr, err = cli.Run("compile", "--clean", "-b", fqbn, sketchFilePde.String())
	require.Error(t, err)
	require.Contains(t, string(stderr), "Error opening sketch: multiple main sketch files found")
}

func TestCompileSketchCaseMismatchFails(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	// Init the environment explicitly
	_, _, err := cli.Run("update")
	require.NoError(t, err)

	// Install core to compile
	_, _, err = cli.Run("core", "install", "arduino:avr@1.8.3")
	require.NoError(t, err)

	sketchName := "CompileSketchCaseMismatch"
	sketchPath := cli.SketchbookDir().Join(sketchName)
	fqbn := "arduino:avr:uno"

	_, _, err = cli.Run("sketch", "new", sketchPath.String())
	require.NoError(t, err)

	// Rename main .ino file so casing is different from sketch name
	sketchFile := sketchPath.Join(sketchName + ".ino")
	sketchMainFile := sketchPath.Join(strings.ToLower(sketchName) + ".ino")
	err = sketchFile.Rename(sketchMainFile)
	require.NoError(t, err)

	// Verifies compilation fails when:
	// * Compiling with sketch path
	_, stderr, err := cli.Run("compile", "--clean", "-b", fqbn, sketchPath.String())
	require.Error(t, err)
	require.Contains(t, string(stderr), "Error opening sketch:")
	// * Compiling with sketch main file
	_, stderr, err = cli.Run("compile", "--clean", "-b", fqbn, sketchMainFile.String())
	require.Error(t, err)
	require.Contains(t, string(stderr), "Error opening sketch:")
	// * Compiling in sketch path
	cli.SetWorkingDir(sketchPath)
	_, stderr, err = cli.Run("compile", "--clean", "-b", fqbn)
	require.Error(t, err)
	require.Contains(t, string(stderr), "Error opening sketch:")
}
