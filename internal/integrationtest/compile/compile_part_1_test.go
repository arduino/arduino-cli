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

package compile_part_1_test

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"strings"
	"testing"

	"github.com/arduino/arduino-cli/internal/integrationtest"
	"github.com/arduino/go-paths-helper"
	"github.com/stretchr/testify/require"
)

func TestCompileWithoutFqbn(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	// Init the environment explicitly
	_, _, err := cli.Run("core", "update-index")
	require.NoError(t, err)

	// Install Arduino AVR Boards
	_, _, err = cli.Run("core", "install", "arduino:avr@1.8.3")
	require.NoError(t, err)

	// Build sketch without FQBN
	_, _, err = cli.Run("compile")
	require.Error(t, err)
}

func TestCompileErrorMessage(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	// Init the environment explicitly
	_, _, err := cli.Run("core", "update-index")
	require.NoError(t, err)

	// Download latest AVR
	_, _, err = cli.Run("core", "install", "arduino:avr")
	require.NoError(t, err)

	// Run a batch of bogus compile in a temp dir to check the error messages
	tmp, err := paths.MkTempDir("", "tmp_dir")
	require.NoError(t, err)
	defer tmp.RemoveAll()
	abcdef := tmp.Join("ABCDEF")
	_, stderr, err := cli.Run("compile", "-b", "arduino:avr:uno", abcdef.String())
	require.Error(t, err)
	require.Contains(t, string(stderr), "no such file or directory:")
	_, stderr, err = cli.Run("compile", "-b", "arduino:avr:uno", abcdef.Join("ABCDEF.ino").String())
	require.Error(t, err)
	require.Contains(t, string(stderr), "no such file or directory:")
	_, stderr, err = cli.Run("compile", "-b", "arduino:avr:uno", abcdef.Join("QWERTY").String())
	require.Error(t, err)
	require.Contains(t, string(stderr), "no such file or directory:")

	err = abcdef.Mkdir()
	require.NoError(t, err)
	_, stderr, err = cli.Run("compile", "-b", "arduino:avr:uno", abcdef.String())
	require.Error(t, err)
	require.Contains(t, string(stderr), "main file missing from sketch:")
	_, stderr, err = cli.Run("compile", "-b", "arduino:avr:uno", abcdef.Join("ABCDEF.ino").String())
	require.Error(t, err)
	require.Contains(t, string(stderr), "no such file or directory:")

	qwertyIno := abcdef.Join("QWERTY.ino")
	f, err := qwertyIno.Create()
	require.NoError(t, err)
	defer f.Close()
	_, stderr, err = cli.Run("compile", "-b", "arduino:avr:uno", qwertyIno.String())
	require.Error(t, err)
	require.Contains(t, string(stderr), "main file missing from sketch:")
}

func TestCompileWithSimpleSketch(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	// Init the environment explicitly
	_, _, err := cli.Run("core", "update-index")
	require.NoError(t, err)

	// Download latest AVR
	_, _, err = cli.Run("core", "install", "arduino:avr")
	require.NoError(t, err)

	sketchName := "CompileIntegrationTest"
	sketchPath := cli.SketchbookDir().Join(sketchName)
	fqbn := "arduino:avr:uno"

	// Create a test sketch
	stdout, _, err := cli.Run("sketch", "new", sketchPath.String())
	require.NoError(t, err)
	require.Contains(t, string(stdout), "Sketch created in: "+sketchPath.String())

	// Build sketch for arduino:avr:uno
	_, _, err = cli.Run("compile", "-b", fqbn, sketchPath.String())
	require.NoError(t, err)

	// Build sketch for arduino:avr:uno with json output
	stdout, _, err = cli.Run("compile", "-b", fqbn, sketchPath.String(), "--format", "json")
	require.NoError(t, err)
	// check is a valid json and contains requested data
	var compileOutput map[string]interface{}
	err = json.Unmarshal(stdout, &compileOutput)
	require.NoError(t, err)
	require.NotEmpty(t, compileOutput["compiler_out"])
	require.Empty(t, compileOutput["compiler_err"])

	// Verifies expected binaries have been built
	md5 := md5.Sum(([]byte(sketchPath.String())))
	sketchPathMd5 := strings.ToUpper(hex.EncodeToString(md5[:]))
	require.NotEmpty(t, sketchPathMd5)
	buildDir := paths.TempDir().Join("arduino-sketch-" + sketchPathMd5)
	require.FileExists(t, buildDir.Join(sketchName+".ino.eep").String())
	require.FileExists(t, buildDir.Join(sketchName+".ino.elf").String())
	require.FileExists(t, buildDir.Join(sketchName+".ino.hex").String())
	require.FileExists(t, buildDir.Join(sketchName+".ino.with_bootloader.bin").String())
	require.FileExists(t, buildDir.Join(sketchName+".ino.with_bootloader.hex").String())

	// Verifies binaries are not exported by default to Sketch folder
	sketchBuildDir := sketchPath.Join("build" + strings.ReplaceAll(fqbn, ":", "."))
	require.NoFileExists(t, sketchBuildDir.Join(sketchName+".ino.eep").String())
	require.NoFileExists(t, sketchBuildDir.Join(sketchName+".ino.elf").String())
	require.NoFileExists(t, sketchBuildDir.Join(sketchName+".ino.hex").String())
	require.NoFileExists(t, sketchBuildDir.Join(sketchName+".ino.with_bootloader.bin").String())
	require.NoFileExists(t, sketchBuildDir.Join(sketchName+".ino.with_bootloader.hex").String())
}
