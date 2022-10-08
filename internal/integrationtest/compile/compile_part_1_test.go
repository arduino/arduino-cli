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
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/arduino/arduino-cli/internal/integrationtest"
	"github.com/arduino/go-paths-helper"
	"github.com/stretchr/testify/require"
)

func TestCompile(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	// Init the environment explicitly
	_, _, err := cli.Run("core", "update-index")
	require.NoError(t, err)

	// Install Arduino AVR Boards
	_, _, err = cli.Run("core", "install", "arduino:avr@1.8.5")
	require.NoError(t, err)

	t.Run("WithoutFqbn", func(t *testing.T) { compileWithoutFqbn(t, env, cli) })
	t.Run("ErrorMessage", func(t *testing.T) { compileErrorMessage(t, env, cli) })
	t.Run("WithSimpleSketch", func(t *testing.T) { compileWithSimpleSketch(t, env, cli) })
	t.Run("OutputFlagDefaultPath", func(t *testing.T) { compileOutputFlagDefaultPath(t, env, cli) })
	t.Run("WithSketchWithSymlinkSelfloop", func(t *testing.T) { compileWithSketchWithSymlinkSelfloop(t, env, cli) })
	t.Run("BlacklistedSketchname", func(t *testing.T) { compileBlacklistedSketchname(t, env, cli) })
	t.Run("WithBuildPropertiesFlag", func(t *testing.T) { compileWithBuildPropertiesFlag(t, env, cli) })
	t.Run("WithBuildPropertyContainingQuotes", func(t *testing.T) { compileWithBuildPropertyContainingQuotes(t, env, cli) })
	t.Run("WithMultipleBuildPropertyFlags", func(t *testing.T) { compileWithMultipleBuildPropertyFlags(t, env, cli) })
	t.Run("WithOutputDirFlag", func(t *testing.T) { compileWithOutputDirFlag(t, env, cli) })
	t.Run("WithExportBinariesFlag", func(t *testing.T) { compileWithExportBinariesFlag(t, env, cli) })
	t.Run("WithCustomBuildPath", func(t *testing.T) { compileWithCustomBuildPath(t, env, cli) })
	t.Run("WithExportBinariesEnvVar", func(t *testing.T) { compileWithExportBinariesEnvVar(t, env, cli) })
	t.Run("WithExportBinariesConfig", func(t *testing.T) { compileWithExportBinariesConfig(t, env, cli) })
	t.Run("WithInvalidUrl", func(t *testing.T) { compileWithInvalidUrl(t, env, cli) })
}

func compileWithoutFqbn(t *testing.T, env *integrationtest.Environment, cli *integrationtest.ArduinoCLI) {
	// Build sketch without FQBN
	_, _, err := cli.Run("compile")
	require.Error(t, err)
}

func compileErrorMessage(t *testing.T, env *integrationtest.Environment, cli *integrationtest.ArduinoCLI) {
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

func compileWithSimpleSketch(t *testing.T, env *integrationtest.Environment, cli *integrationtest.ArduinoCLI) {
	sketchName := "CompileIntegrationTest"
	sketchPath := cli.SketchbookDir().Join(sketchName)
	defer sketchPath.RemoveAll()
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

func compileOutputFlagDefaultPath(t *testing.T, env *integrationtest.Environment, cli *integrationtest.ArduinoCLI) {
	// Create a test sketch
	sketchPath := cli.SketchbookDir().Join("test_output_flag_default_path")
	defer sketchPath.RemoveAll()
	fqbn := "arduino:avr:uno"
	_, _, err := cli.Run("sketch", "new", sketchPath.String())
	require.NoError(t, err)

	// Test the --output-dir flag defaulting to current working dir
	target := cli.WorkingDir().Join("test")
	_, _, err = cli.Run("compile", "-b", fqbn, sketchPath.String(), "--output-dir", "test")
	require.NoError(t, err)
	require.DirExists(t, target.String())
}

func compileWithSketchWithSymlinkSelfloop(t *testing.T, env *integrationtest.Environment, cli *integrationtest.ArduinoCLI) {
	{
		sketchName := "CompileIntegrationTestSymlinkSelfLoop"
		sketchPath := cli.SketchbookDir().Join(sketchName)
		defer sketchPath.RemoveAll()
		fqbn := "arduino:avr:uno"

		// Create a test sketch
		stdout, _, err := cli.Run("sketch", "new", sketchPath.String())
		require.NoError(t, err)
		require.Contains(t, string(stdout), "Sketch created in: "+sketchPath.String())

		// create a symlink that loops on himself
		loopFilePath := sketchPath.Join("loop")
		err = os.Symlink(loopFilePath.String(), loopFilePath.String())
		require.NoError(t, err)

		// Build sketch for arduino:avr:uno
		_, stderr, err := cli.Run("compile", "-b", fqbn, sketchPath.String())
		// The assertion is a bit relaxed in this case because win behaves differently from macOs and linux
		// returning a different error detailed message
		require.Contains(t, string(stderr), "Error opening sketch:")
		require.Error(t, err)
	}
	{
		sketchName := "CompileIntegrationTestSymlinkDirLoop"
		sketchPath := cli.SketchbookDir().Join(sketchName)
		defer sketchPath.RemoveAll()
		fqbn := "arduino:avr:uno"

		// Create a test sketch
		stdout, _, err := cli.Run("sketch", "new", sketchPath.String())
		require.NoError(t, err)
		require.Contains(t, string(stdout), "Sketch created in: "+sketchPath.String())

		// create a symlink that loops on the upper level
		loopDirPath := sketchPath.Join("loop_dir")
		err = loopDirPath.Mkdir()
		require.NoError(t, err)
		loopDirSymlinkPath := loopDirPath.Join("loop_dir_symlink")
		err = os.Symlink(loopDirPath.String(), loopDirSymlinkPath.String())
		require.NoError(t, err)

		// Build sketch for arduino:avr:uno
		_, stderr, err := cli.Run("compile", "-b", fqbn, sketchPath.String())
		// The assertion is a bit relaxed in this case because win behaves differently from macOs and linux
		// returning a different error detailed message
		require.Contains(t, string(stderr), "Error opening sketch:")
		require.Error(t, err)
	}
}

func compileBlacklistedSketchname(t *testing.T, env *integrationtest.Environment, cli *integrationtest.ArduinoCLI) {
	// Compile should ignore folders named `RCS`, `.git` and the likes, but
	// it should be ok for a sketch to be named like RCS.ino
	sketchName := "RCS"
	sketchPath := cli.SketchbookDir().Join(sketchName)
	defer sketchPath.RemoveAll()
	fqbn := "arduino:avr:uno"

	// Create a test sketch
	stdout, _, err := cli.Run("sketch", "new", sketchPath.String())
	require.NoError(t, err)
	require.Contains(t, string(stdout), "Sketch created in: "+sketchPath.String())

	// Build sketch for arduino:avr:uno
	_, _, err = cli.Run("compile", "-b", fqbn, sketchPath.String())
	require.NoError(t, err)
}

func TestCompileWithoutPrecompiledLibraries(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	// Init the environment explicitly
	url := "https://adafruit.github.io/arduino-board-index/package_adafruit_index.json"
	_, _, err := cli.Run("core", "update-index", "--additional-urls="+url)
	require.NoError(t, err)
	_, _, err = cli.Run("core", "install", "arduino:mbed@1.3.1", "--additional-urls="+url)
	require.NoError(t, err)

	// // Precompiled version of Arduino_TensorflowLite
	//	_, _, err = cli.Run("lib", "install", "Arduino_LSM9DS1")
	//	require.NoError(t, err)
	//	_, _, err = cli.Run("lib", "install", "Arduino_TensorflowLite@2.1.1-ALPHA-precompiled")
	//	require.NoError(t, err)

	//	sketchPath := cli.SketchbookDir().Join("libraries", "Arduino_TensorFlowLite", "examples", "hello_world")
	//	_, _, err = cli.Run("compile", "-b", "arduino:mbed:nano33ble", sketchPath.String())
	//	require.NoError(t, err)

	_, _, err = cli.Run("core", "install", "arduino:samd@1.8.7", "--additional-urls="+url)
	require.NoError(t, err)
	//	_, _, err = cli.Run("core", "install", "adafruit:samd@1.6.4", "--additional-urls="+url)
	//	require.NoError(t, err)
	//	// should work on adafruit too after https://github.com/arduino/arduino-cli/pull/1134
	//	_, _, err = cli.Run("compile", "-b", "adafruit:samd:adafruit_feather_m4", sketchPath.String())
	//	require.NoError(t, err)

	//	// Non-precompiled version of Arduino_TensorflowLite
	//	_, _, err = cli.Run("lib", "install", "Arduino_TensorflowLite@2.1.0-ALPHA")
	//	require.NoError(t, err)
	//	_, _, err = cli.Run("compile", "-b", "arduino:mbed:nano33ble", sketchPath.String())
	//	require.NoError(t, err)
	//	_, _, err = cli.Run("compile", "-b", "adafruit:samd:adafruit_feather_m4", sketchPath.String())
	//	require.NoError(t, err)

	// Bosch sensor library
	_, _, err = cli.Run("lib", "install", "BSEC Software Library@1.5.1474")
	require.NoError(t, err)
	sketchPath := cli.SketchbookDir().Join("libraries", "BSEC_Software_Library", "examples", "basic")
	_, _, err = cli.Run("compile", "-b", "arduino:samd:mkr1000", sketchPath.String())
	require.NoError(t, err)
	_, _, err = cli.Run("compile", "-b", "arduino:mbed:nano33ble", sketchPath.String())
	require.NoError(t, err)

	// USBBlaster library
	_, _, err = cli.Run("lib", "install", "USBBlaster@1.0.0")
	require.NoError(t, err)
	sketchPath = cli.SketchbookDir().Join("libraries", "USBBlaster", "examples", "USB_Blaster")
	_, _, err = cli.Run("compile", "-b", "arduino:samd:mkrvidor4000", sketchPath.String())
	require.NoError(t, err)
}

func compileWithBuildPropertiesFlag(t *testing.T, env *integrationtest.Environment, cli *integrationtest.ArduinoCLI) {
	{
		sketchName := "sketch_with_single_string_define"
		sketchPath := cli.CopySketch(sketchName)
		defer sketchPath.RemoveAll()
		fqbn := "arduino:avr:uno"

		// Compile using a build property with quotes
		_, stderr, err := cli.Run("compile", "-b", fqbn, "--build-properties=\"build.extra_flags=\"-DMY_DEFINE=\"hello world\"\"", sketchPath.String(), "--verbose", "--clean")
		require.Error(t, err)
		require.NotContains(t, string(stderr), "Flag --build-properties has been deprecated, please use --build-property instead.")

		// Try again with quotes
		_, stderr, err = cli.Run("compile", "-b", fqbn, "--build-properties=\"build.extra_flags=-DMY_DEFINE=\"hello\"\"", sketchPath.String(), "--verbose", "--clean")
		require.Error(t, err)
		require.NotContains(t, string(stderr), "Flag --build-properties has been deprecated, please use --build-property instead.")
	}
	{
		sketchName := "sketch_with_single_int_define"
		sketchPath := cli.CopySketch(sketchName)
		defer sketchPath.RemoveAll()
		fqbn := "arduino:avr:uno"

		// Try without quotes
		stdout, stderr, err := cli.Run("compile", "-b", fqbn, "--build-properties=\"build.extra_flags=-DMY_DEFINE=1\"", sketchPath.String(), "--verbose", "--clean")
		require.NoError(t, err)
		require.Contains(t, string(stderr), "Flag --build-properties has been deprecated, please use --build-property instead.")
		require.Contains(t, string(stdout), "-DMY_DEFINE=1")
	}
	{
		sketchName := "sketch_with_multiple_int_defines"
		sketchPath := cli.CopySketch(sketchName)
		defer sketchPath.RemoveAll()
		fqbn := "arduino:avr:uno"

		stdout, stderr, err := cli.Run("compile", "-b", fqbn,
			"--build-properties", "build.extra_flags=-DFIRST_PIN=1,compiler.cpp.extra_flags=-DSECOND_PIN=2",
			sketchPath.String(), "--verbose", "--clean")
		require.NoError(t, err)
		require.Contains(t, string(stderr), "Flag --build-properties has been deprecated, please use --build-property instead.")
		require.Contains(t, string(stdout), "-DFIRST_PIN=1")
		require.Contains(t, string(stdout), "-DSECOND_PIN=2")
	}
}

func compileWithBuildPropertyContainingQuotes(t *testing.T, env *integrationtest.Environment, cli *integrationtest.ArduinoCLI) {
	sketchName := "sketch_with_single_string_define"
	sketchPath := cli.CopySketch(sketchName)
	defer sketchPath.RemoveAll()
	fqbn := "arduino:avr:uno"

	// Compile using a build property with quotes
	stdout, _, err := cli.Run("compile",
		"-b",
		fqbn,
		`--build-property=build.extra_flags="-DMY_DEFINE="hello world""`,
		sketchPath.String(),
		"--verbose")
	require.NoError(t, err)
	require.Contains(t, string(stdout), `-DMY_DEFINE=\"hello world\"`)
}

func compileWithMultipleBuildPropertyFlags(t *testing.T, env *integrationtest.Environment, cli *integrationtest.ArduinoCLI) {
	sketchName := "sketch_with_multiple_defines"
	sketchPath := cli.CopySketch(sketchName)
	defer sketchPath.RemoveAll()
	fqbn := "arduino:avr:uno"

	// Compile using multiple build properties separated by a space
	_, _, err := cli.Run("compile", "-b", fqbn,
		"--build-property=compiler.cpp.extra_flags=\"-DPIN=2 -DSSID=\"This is a String\"\"",
		sketchPath.String(), "--verbose", "--clean",
	)
	require.Error(t, err)

	// Compile using multiple build properties separated by a space and properly quoted
	stdout, _, err := cli.Run(
		"compile", "-b", fqbn,
		"--build-property=compiler.cpp.extra_flags=-DPIN=2 \"-DSSID=\"This is a String\"\"",
		sketchPath.String(), "--verbose", "--clean",
	)
	require.NoError(t, err)
	require.Contains(t, string(stdout), "-DPIN=2 \"-DSSID=\\\"This is a String\\\"\"")

	// Tries compilation using multiple build properties separated by a comma
	_, _, err = cli.Run(
		"compile", "-b", fqbn,
		"--build-property=compiler.cpp.extra_flags=\"-DPIN=2,-DSSID=\"This is a String\"\"",
		sketchPath.String(), "--verbose", "--clean",
	)
	require.Error(t, err)

	stdout, _, err = cli.Run(
		"compile", "-b", fqbn,
		"--build-property=compiler.cpp.extra_flags=\"-DPIN=2\"",
		"--build-property=compiler.cpp.extra_flags=\"-DSSID=\"This is a String\"\"",
		sketchPath.String(), "--verbose", "--clean",
	)
	require.Error(t, err)
	require.NotContains(t, string(stdout), "-DPIN=2")
	require.Contains(t, string(stdout), "-DSSID=\\\"This is a String\\\"")

	stdout, _, err = cli.Run(
		"compile", "-b", fqbn,
		"--build-property=compiler.cpp.extra_flags=\"-DPIN=2\"",
		"--build-property=build.extra_flags=\"-DSSID=\"hello world\"\"",
		sketchPath.String(), "--verbose", "--clean",
	)
	require.NoError(t, err)
	require.Contains(t, string(stdout), "-DPIN=2")
	require.Contains(t, string(stdout), "-DSSID=\\\"hello world\\\"")
}
