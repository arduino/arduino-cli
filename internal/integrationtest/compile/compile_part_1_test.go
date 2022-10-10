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
	"go.bug.st/testifyjson/requirejson"
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

	integrationtest.CLISubtests{
		{"WithoutFqbn", compileWithoutFqbn},
		{"ErrorMessage", compileErrorMessage},
		{"WithSimpleSketch", compileWithSimpleSketch},
		{"OutputFlagDefaultPath", compileOutputFlagDefaultPath},
		{"WithSketchWithSymlinkSelfloop", compileWithSketchWithSymlinkSelfloop},
		{"BlacklistedSketchname", compileBlacklistedSketchname},
		{"WithBuildPropertiesFlag", compileWithBuildPropertiesFlag},
		{"WithBuildPropertyContainingQuotes", compileWithBuildPropertyContainingQuotes},
		{"WithMultipleBuildPropertyFlags", compileWithMultipleBuildPropertyFlags},
		{"WithOutputDirFlag", compileWithOutputDirFlag},
		{"WithExportBinariesFlag", compileWithExportBinariesFlag},
		{"WithCustomBuildPath", compileWithCustomBuildPath},
		{"WithExportBinariesEnvVar", compileWithExportBinariesEnvVar},
		{"WithExportBinariesConfig", compileWithExportBinariesConfig},
		{"WithInvalidUrl", compileWithInvalidUrl},
	}.Run(t, env, cli)
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

func compileWithOutputDirFlag(t *testing.T, env *integrationtest.Environment, cli *integrationtest.ArduinoCLI) {
	sketchName := "CompileWithOutputDir"
	sketchPath := cli.SketchbookDir().Join(sketchName)
	defer sketchPath.RemoveAll()
	fqbn := "arduino:avr:uno"

	// Create a test sketch
	stdout, _, err := cli.Run("sketch", "new", sketchPath.String())
	require.NoError(t, err)
	require.Contains(t, string(stdout), "Sketch created in: "+sketchPath.String())

	// Test the --output-dir flag with absolute path
	outputDir := cli.SketchbookDir().Join("test_dir", "output_dir")
	_, _, err = cli.Run("compile", "-b", fqbn, sketchPath.String(), "--output-dir", outputDir.String())
	require.NoError(t, err)

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

	// Verifies binaries are exported when --output-dir flag is specified
	require.DirExists(t, outputDir.String())
	require.FileExists(t, outputDir.Join(sketchName+".ino.eep").String())
	require.FileExists(t, outputDir.Join(sketchName+".ino.elf").String())
	require.FileExists(t, outputDir.Join(sketchName+".ino.hex").String())
	require.FileExists(t, outputDir.Join(sketchName+".ino.with_bootloader.bin").String())
	require.FileExists(t, outputDir.Join(sketchName+".ino.with_bootloader.hex").String())
}

func compileWithExportBinariesFlag(t *testing.T, env *integrationtest.Environment, cli *integrationtest.ArduinoCLI) {
	sketchName := "CompileWithExportBinariesFlag"
	sketchPath := cli.SketchbookDir().Join(sketchName)
	defer sketchPath.RemoveAll()
	fqbn := "arduino:avr:uno"

	// Create a test sketch
	_, _, err := cli.Run("sketch", "new", sketchPath.String())
	require.NoError(t, err)

	// Test the --output-dir flag with absolute path
	_, _, err = cli.Run("compile", "-b", fqbn, sketchPath.String(), "--export-binaries")
	require.NoError(t, err)
	require.DirExists(t, sketchPath.Join("build").String())

	// Verifies binaries are exported when --export-binaries flag is set
	fqbn = strings.ReplaceAll(fqbn, ":", ".")
	require.FileExists(t, sketchPath.Join("build", fqbn, sketchName+".ino.eep").String())
	require.FileExists(t, sketchPath.Join("build", fqbn, sketchName+".ino.elf").String())
	require.FileExists(t, sketchPath.Join("build", fqbn, sketchName+".ino.hex").String())
	require.FileExists(t, sketchPath.Join("build", fqbn, sketchName+".ino.with_bootloader.bin").String())
	require.FileExists(t, sketchPath.Join("build", fqbn, sketchName+".ino.with_bootloader.hex").String())
}

func compileWithCustomBuildPath(t *testing.T, env *integrationtest.Environment, cli *integrationtest.ArduinoCLI) {
	sketchName := "CompileWithBuildPath"
	sketchPath := cli.SketchbookDir().Join(sketchName)
	defer sketchPath.RemoveAll()
	fqbn := "arduino:avr:uno"

	// Create a test sketch
	_, _, err := cli.Run("sketch", "new", sketchPath.String())
	require.NoError(t, err)

	// Test the --build-path flag with absolute path
	buildPath := cli.DataDir().Join("test_dir", "build_dir")
	_, _, err = cli.Run("compile", "-b", fqbn, sketchPath.String(), "--build-path", buildPath.String())
	require.NoError(t, err)

	// Verifies expected binaries have been built to build_path
	require.DirExists(t, buildPath.String())
	require.FileExists(t, buildPath.Join(sketchName+".ino.eep").String())
	require.FileExists(t, buildPath.Join(sketchName+".ino.elf").String())
	require.FileExists(t, buildPath.Join(sketchName+".ino.hex").String())
	require.FileExists(t, buildPath.Join(sketchName+".ino.with_bootloader.bin").String())
	require.FileExists(t, buildPath.Join(sketchName+".ino.with_bootloader.hex").String())

	// Verifies there are no binaries in temp directory
	md5 := md5.Sum(([]byte(sketchPath.String())))
	sketchPathMd5 := strings.ToUpper(hex.EncodeToString(md5[:]))
	require.NotEmpty(t, sketchPathMd5)
	buildDir := paths.TempDir().Join("arduino-sketch-" + sketchPathMd5)
	require.NoFileExists(t, buildDir.Join(sketchName+".ino.eep").String())
	require.NoFileExists(t, buildDir.Join(sketchName+".ino.elf").String())
	require.NoFileExists(t, buildDir.Join(sketchName+".ino.hex").String())
	require.NoFileExists(t, buildDir.Join(sketchName+".ino.with_bootloader.bin").String())
	require.NoFileExists(t, buildDir.Join(sketchName+".ino.with_bootloader.hex").String())
}

func compileWithExportBinariesEnvVar(t *testing.T, env *integrationtest.Environment, cli *integrationtest.ArduinoCLI) {
	sketchName := "CompileWithExportBinariesEnvVar"
	sketchPath := cli.SketchbookDir().Join(sketchName)
	defer sketchPath.RemoveAll()
	fqbn := "arduino:avr:uno"

	// Create a test sketch
	_, _, err := cli.Run("sketch", "new", sketchPath.String())
	require.NoError(t, err)

	envVar := cli.GetDefaultEnv()
	envVar["ARDUINO_SKETCH_ALWAYS_EXPORT_BINARIES"] = "true"

	// Test compilation with export binaries env var set
	_, _, err = cli.RunWithCustomEnv(envVar, "compile", "-b", fqbn, sketchPath.String())
	require.NoError(t, err)
	require.DirExists(t, sketchPath.Join("build").String())

	// Verifies binaries are exported when export binaries env var is set
	fqbn = strings.ReplaceAll(fqbn, ":", ".")
	require.FileExists(t, sketchPath.Join("build", fqbn, sketchName+".ino.eep").String())
	require.FileExists(t, sketchPath.Join("build", fqbn, sketchName+".ino.elf").String())
	require.FileExists(t, sketchPath.Join("build", fqbn, sketchName+".ino.hex").String())
	require.FileExists(t, sketchPath.Join("build", fqbn, sketchName+".ino.with_bootloader.bin").String())
	require.FileExists(t, sketchPath.Join("build", fqbn, sketchName+".ino.with_bootloader.hex").String())
}

func compileWithExportBinariesConfig(t *testing.T, env *integrationtest.Environment, cli *integrationtest.ArduinoCLI) {
	sketchName := "CompileWithExportBinariesEnvVar"
	sketchPath := cli.SketchbookDir().Join(sketchName)
	defer sketchPath.RemoveAll()
	fqbn := "arduino:avr:uno"

	// Create a test sketch
	_, _, err := cli.Run("sketch", "new", sketchPath.String())
	require.NoError(t, err)

	// Create settings with export binaries set to true
	envVar := cli.GetDefaultEnv()
	envVar["ARDUINO_SKETCH_ALWAYS_EXPORT_BINARIES"] = "true"
	_, _, err = cli.RunWithCustomEnv(envVar, "config", "init", "--dest-dir", ".")
	require.NoError(t, err)
	defer cli.WorkingDir().Join("arduino-cli.yaml").Remove()

	// Test if arduino-cli config file written in the previous run has the `always_export_binaries` flag set.
	stdout, _, err := cli.Run("config", "dump", "--format", "json")
	require.NoError(t, err)
	requirejson.Contains(t, stdout, `
		{
			"sketch": {
			"always_export_binaries": "true"
	  		}
		}`)

	// Test compilation with export binaries env var set
	_, _, err = cli.Run("compile", "-b", fqbn, sketchPath.String())
	require.NoError(t, err)
	require.DirExists(t, sketchPath.Join("build").String())

	// Verifies binaries are exported when export binaries env var is set
	fqbn = strings.ReplaceAll(fqbn, ":", ".")
	require.FileExists(t, sketchPath.Join("build", fqbn, sketchName+".ino.eep").String())
	require.FileExists(t, sketchPath.Join("build", fqbn, sketchName+".ino.elf").String())
	require.FileExists(t, sketchPath.Join("build", fqbn, sketchName+".ino.hex").String())
	require.FileExists(t, sketchPath.Join("build", fqbn, sketchName+".ino.with_bootloader.bin").String())
	require.FileExists(t, sketchPath.Join("build", fqbn, sketchName+".ino.with_bootloader.hex").String())
}

func compileWithInvalidUrl(t *testing.T, env *integrationtest.Environment, cli *integrationtest.ArduinoCLI) {
	sketchName := "CompileWithInvalidURL"
	sketchPath := cli.SketchbookDir().Join(sketchName)
	defer sketchPath.RemoveAll()
	fqbn := "arduino:avr:uno"

	// Create a test sketch
	_, _, err := cli.Run("sketch", "new", sketchPath.String())
	require.NoError(t, err)

	_, _, err = cli.Run("config", "init", "--dest-dir", ".", "--additional-urls", "https://example.com/package_example_index.json")
	require.NoError(t, err)
	defer cli.WorkingDir().Join("arduino-cli.yaml").Remove()

	_, stderr, err := cli.Run("compile", "-b", fqbn, sketchPath.String())
	require.NoError(t, err)
	require.Contains(t, string(stderr), "Error initializing instance: Loading index file: loading json index file")
	expectedIndexfile := cli.DataDir().Join("package_example_index.json")
	require.Contains(t, string(stderr), "loading json index file "+expectedIndexfile.String()+": open "+expectedIndexfile.String()+":")
}
