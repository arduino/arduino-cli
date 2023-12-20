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
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/arduino/arduino-cli/arduino/builder/cpp"
	"github.com/arduino/arduino-cli/internal/integrationtest"
	"github.com/arduino/go-paths-helper"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
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
		{"WithCachePurgeNeeded", compileWithCachePurgeNeeded},
		{"OutputFlagDefaultPath", compileOutputFlagDefaultPath},
		{"WithSketchWithSymlinkSelfloop", compileWithSketchWithSymlinkSelfloop},
		{"BlacklistedSketchname", compileBlacklistedSketchname},
		{"WithBuildPropertiesFlag", compileWithBuildPropertiesFlag},
		{"WithBuildPropertyContainingQuotes", compileWithBuildPropertyContainingQuotes},
		{"WithMultipleBuildPropertyFlags", compileWithMultipleBuildPropertyFlags},
		{"WithOutputDirFlag", compileWithOutputDirFlag},
		{"WithExportBinariesFlag", compileWithExportBinariesFlag},
		{"WithExportBinariesEnvVar", compileWithExportBinariesEnvVar},
		{"WithExportBinariesConfig", compileWithExportBinariesConfig},
		{"WithInvalidUrl", compileWithInvalidUrl},
		{"WithPdeExtension", compileWithPdeExtension},
		{"WithMultipleMainFiles", compileWithMultipleMainFiles},
		{"CaseMismatchFails", compileCaseMismatchFails},
		{"OnlyCompilationDatabaseFlag", compileOnlyCompilationDatabaseFlag},
		{"UsingPlatformLocalTxt", compileUsingPlatformLocalTxt},
		{"UsingBoardsLocalTxt", compileUsingBoardsLocalTxt},
		{"WithInvalidBuildOptionJson", compileWithInvalidBuildOptionJson},
		{"WithRelativeBuildPath", compileWithRelativeBuildPath},
		{"WithFakeSecureBootCore", compileWithFakeSecureBootCore},
		{"PreprocessFlagDoNotMessUpWithOutput", preprocessFlagDoNotMessUpWithOutput},
		{"WithCustomBuildPath", buildWithCustomBuildPath},
		{"WithCustomBuildPathAndOUtputDirFlag", buildWithCustomBuildPathAndOUtputDirFlag},
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
	compileWithSimpleSketchCustomEnv(t, env, cli, cli.GetDefaultEnv())
}

func compileWithCachePurgeNeeded(t *testing.T, env *integrationtest.Environment, cli *integrationtest.ArduinoCLI) {
	// create directories that must be purged
	baseDir := paths.TempDir().Join("arduino", "sketches")

	// purge case: last used file too old
	oldDir1 := baseDir.Join("test_old_sketch_1")
	require.NoError(t, oldDir1.MkdirAll())
	require.NoError(t, oldDir1.Join(".last-used").WriteFile([]byte{}))
	require.NoError(t, oldDir1.Join(".last-used").Chtimes(time.Now(), time.Unix(0, 0)))
	// no purge case: last used file not existing
	missingFileDir := baseDir.Join("test_sketch_2")
	require.NoError(t, missingFileDir.MkdirAll())

	defer oldDir1.RemoveAll()
	defer missingFileDir.RemoveAll()

	customEnv := cli.GetDefaultEnv()
	customEnv["ARDUINO_BUILD_CACHE_COMPILATIONS_BEFORE_PURGE"] = "1"
	compileWithSimpleSketchCustomEnv(t, env, cli, customEnv)

	// check that purge has been run
	require.NoFileExists(t, oldDir1.String())
	require.DirExists(t, missingFileDir.String())
}

func compileWithSimpleSketchCustomEnv(t *testing.T, env *integrationtest.Environment, cli *integrationtest.ArduinoCLI, customEnv map[string]string) {
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
	stdout, _, err = cli.RunWithCustomEnv(customEnv, "compile", "-b", fqbn, sketchPath.String(), "--format", "json")
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
	buildDir := paths.TempDir().Join("arduino", "sketches", sketchPathMd5)
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
		require.Contains(t, string(stderr), "Can't open sketch:")
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
		require.Contains(t, string(stderr), "Can't open sketch:")
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
	buildDir := paths.TempDir().Join("arduino", "sketches", sketchPathMd5)
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
	stdout, _, err := cli.Run("config", "dump", "--format", "json", "--config-file", "arduino-cli.yaml")
	require.NoError(t, err)
	requirejson.Contains(t, stdout, `
		{
			"sketch": {
			"always_export_binaries": "true"
	  		}
		}`)

	// Test compilation with export binaries env var set
	_, _, err = cli.Run("compile", "-b", fqbn, "--config-file", "arduino-cli.yaml", sketchPath.String())
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

	_, stderr, err := cli.Run("compile", "-b", fqbn, "--config-file", "arduino-cli.yaml", sketchPath.String())
	require.NoError(t, err)
	require.Contains(t, string(stderr), "Error initializing instance: Loading index file: loading json index file")
	expectedIndexfile := cli.DataDir().Join("package_example_index.json")
	require.Contains(t, string(stderr), "loading json index file "+expectedIndexfile.String()+": open "+expectedIndexfile.String()+":")
}

func compileWithPdeExtension(t *testing.T, env *integrationtest.Environment, cli *integrationtest.ArduinoCLI) {
	sketchName := "CompilePdeSketch"
	sketchPath := cli.SketchbookDir().Join(sketchName)
	defer sketchPath.RemoveAll()
	fqbn := "arduino:avr:uno"

	// Create a test sketch
	_, _, err := cli.Run("sketch", "new", sketchPath.String())
	require.NoError(t, err)

	// Renames sketch file to pde
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

func compileWithMultipleMainFiles(t *testing.T, env *integrationtest.Environment, cli *integrationtest.ArduinoCLI) {
	sketchName := "CompileSketchMultipleMainFiles"
	sketchPath := cli.SketchbookDir().Join(sketchName)
	defer sketchPath.RemoveAll()
	fqbn := "arduino:avr:uno"

	// Create a test sketch
	_, _, err := cli.Run("sketch", "new", sketchPath.String())
	require.NoError(t, err)

	// Copy .ino sketch file to .pde
	sketchFileIno := sketchPath.Join(sketchName + ".ino")
	sketchFilePde := sketchPath.Join(sketchName + ".pde")
	err = sketchFileIno.CopyTo(sketchFilePde)
	require.NoError(t, err)

	// Build sketch from folder
	_, stderr, err := cli.Run("compile", "--clean", "-b", fqbn, sketchPath.String())
	require.Error(t, err)
	require.Contains(t, string(stderr), "Can't open sketch: multiple main sketch files found")

	// Build sketch from .ino file
	_, stderr, err = cli.Run("compile", "--clean", "-b", fqbn, sketchFileIno.String())
	require.Error(t, err)
	require.Contains(t, string(stderr), "Can't open sketch: multiple main sketch files found")

	// Build sketch from .pde file
	_, stderr, err = cli.Run("compile", "--clean", "-b", fqbn, sketchFilePde.String())
	require.Error(t, err)
	require.Contains(t, string(stderr), "Can't open sketch: multiple main sketch files found")
}

func compileCaseMismatchFails(t *testing.T, env *integrationtest.Environment, cli *integrationtest.ArduinoCLI) {
	sketchName := "CompileSketchCaseMismatch"
	sketchPath := cli.SketchbookDir().Join(sketchName)
	defer sketchPath.RemoveAll()
	fqbn := "arduino:avr:uno"

	_, _, err := cli.Run("sketch", "new", sketchPath.String())
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
	require.Contains(t, string(stderr), "Can't open sketch:")
	// * Compiling with sketch main file
	_, stderr, err = cli.Run("compile", "--clean", "-b", fqbn, sketchMainFile.String())
	require.Error(t, err)
	require.Contains(t, string(stderr), "Can't open sketch:")
	// * Compiling in sketch path
	cli.SetWorkingDir(sketchPath)
	defer cli.SetWorkingDir(env.RootDir())
	_, stderr, err = cli.Run("compile", "--clean", "-b", fqbn)
	require.Error(t, err)
	require.Contains(t, string(stderr), "Can't open sketch:")
}

func compileOnlyCompilationDatabaseFlag(t *testing.T, env *integrationtest.Environment, cli *integrationtest.ArduinoCLI) {
	sketchName := "CompileSketchOnlyCompilationDatabaseFlag"
	sketchPath := cli.SketchbookDir().Join(sketchName)
	defer sketchPath.RemoveAll()
	fqbn := "arduino:avr:uno"

	_, _, err := cli.Run("sketch", "new", sketchPath.String())
	require.NoError(t, err)

	// Verifies no binaries exist
	buildPath := sketchPath.Join("build")
	require.NoDirExists(t, buildPath.String())

	// Compile with both --export-binaries and --only-compilation-database flags
	_, _, err = cli.Run("compile", "--export-binaries", "--only-compilation-database", "--clean", "-b", fqbn, sketchPath.String())
	require.NoError(t, err)

	// Verifies no binaries are exported
	require.NoDirExists(t, buildPath.String())

	// Verifies no binaries exist
	buildPath = cli.SketchbookDir().Join("export-dir")
	require.NoDirExists(t, buildPath.String())

	// Compile by setting the --output-dir flag and --only-compilation-database flags
	_, _, err = cli.Run("compile", "--output-dir", buildPath.String(), "--only-compilation-database", "--clean", "-b", fqbn, sketchPath.String())
	require.NoError(t, err)

	// Verifies no binaries are exported
	require.NoDirExists(t, buildPath.String())
}

func compileUsingPlatformLocalTxt(t *testing.T, env *integrationtest.Environment, cli *integrationtest.ArduinoCLI) {
	sketchName := "CompileSketchUsingPlatformLocalTxt"
	sketchPath := cli.SketchbookDir().Join(sketchName)
	defer sketchPath.RemoveAll()
	fqbn := "arduino:avr:uno"

	_, _, err := cli.Run("sketch", "new", sketchPath.String())
	require.NoError(t, err)

	// Verifies compilation works without issues
	_, _, err = cli.Run("compile", "--clean", "-b", fqbn, sketchPath.String())
	require.NoError(t, err)

	// Overrides default platform compiler with an unexisting one
	platformLocalTxt := cli.DataDir().Join("packages", "arduino", "hardware", "avr", "1.8.5", "platform.local.txt")
	err = platformLocalTxt.WriteFile([]byte("compiler.c.cmd=my-compiler-that-does-not-exist"))
	require.NoError(t, err)
	// Remove the file at the end of the test to avoid disrupting following tests
	defer platformLocalTxt.Remove()

	// Verifies compilation now fails because compiler is not found
	_, stderr, err := cli.Run("compile", "--clean", "-b", fqbn, sketchPath.String())
	require.Error(t, err)
	require.Contains(t, string(stderr), "my-compiler-that-does-not-exist")
}

func compileUsingBoardsLocalTxt(t *testing.T, env *integrationtest.Environment, cli *integrationtest.ArduinoCLI) {
	sketchName := "CompileSketchUsingBoardsLocalTxt"
	sketchPath := cli.SketchbookDir().Join(sketchName)
	defer sketchPath.RemoveAll()
	// Usa a made up board
	fqbn := "arduino:avr:nessuno"

	_, _, err := cli.Run("sketch", "new", sketchPath.String())
	require.NoError(t, err)

	// Verifies compilation fails because board doesn't exist
	_, stderr, err := cli.Run("compile", "--clean", "-b", fqbn, sketchPath.String())
	require.Error(t, err)
	require.Contains(t, string(stderr), "Error during build: Invalid FQBN: board arduino:avr:nessuno not found")

	// Use custom boards.local.txt with made arduino:avr:nessuno board
	boardsLocalTxt := cli.DataDir().Join("packages", "arduino", "hardware", "avr", "1.8.5", "boards.local.txt")
	err = paths.New("..", "testdata", "boards.local.txt").CopyTo(boardsLocalTxt)
	require.NoError(t, err)
	// Remove the file at the end of the test to avoid disrupting following tests
	defer boardsLocalTxt.Remove()

	_, _, err = cli.Run("compile", "--clean", "-b", fqbn, sketchPath.String())
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

func TestCompileWithCustomLibraries(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	// Creates config with additional URL to install necessary core
	url := "http://arduino.esp8266.com/stable/package_esp8266com_index.json"
	_, _, err := cli.Run("config", "init", "--dest-dir", ".", "--additional-urls", url)
	require.NoError(t, err)

	// Init the environment explicitly
	_, _, err = cli.Run("update", "--config-file", "arduino-cli.yaml")
	require.NoError(t, err)

	_, _, err = cli.Run("core", "install", "esp8266:esp8266", "--config-file", "arduino-cli.yaml")
	require.NoError(t, err)

	sketchName := "sketch_with_multiple_custom_libraries"
	sketchPath := cli.CopySketch(sketchName)
	fqbn := "esp8266:esp8266:nodemcu:xtal=80,vt=heap,eesz=4M1M,wipe=none,baud=115200"

	firstLib := sketchPath.Join("libraries1")
	secondLib := sketchPath.Join("libraries2")
	_, _, err = cli.Run("compile", "--libraries",
		firstLib.String(),
		"--libraries", secondLib.String(),
		"-b", fqbn,
		"--config-file", "arduino-cli.yaml",
		sketchPath.String())
	require.NoError(t, err)
}

func TestCompileWithArchivesAndLongPaths(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	// Creates config with additional URL to install necessary core
	url := "http://arduino.esp8266.com/stable/package_esp8266com_index.json"
	_, _, err := cli.Run("config", "init", "--dest-dir", ".", "--additional-urls", url)
	require.NoError(t, err)

	// Init the environment explicitly
	_, _, err = cli.Run("update", "--config-file", "arduino-cli.yaml")
	require.NoError(t, err)

	// Install core to compile
	_, _, err = cli.Run("core", "install", "esp8266:esp8266@2.7.4", "--config-file", "arduino-cli.yaml")
	require.NoError(t, err)

	// Install test library
	_, _, err = cli.Run("lib", "install", "ArduinoIoTCloud", "--config-file", "arduino-cli.yaml")
	require.NoError(t, err)

	stdout, _, err := cli.Run("lib", "examples", "ArduinoIoTCloud", "--format", "json", "--config-file", "arduino-cli.yaml")
	require.NoError(t, err)
	var libOutput []map[string]interface{}
	err = json.Unmarshal(stdout, &libOutput)
	require.NoError(t, err)
	sketchPath := paths.New(libOutput[0]["library"].(map[string]interface{})["install_dir"].(string))
	sketchPath = sketchPath.Join("examples", "ArduinoIoTCloud-Advanced")

	t.Run("Compile", func(t *testing.T) {
		_, _, err = cli.Run("compile", "-b", "esp8266:esp8266:huzzah", sketchPath.String(), "--config-file", "arduino-cli.yaml")
		require.NoError(t, err)
	})

	t.Run("CheckCachingOfFolderArchives", func(t *testing.T) {
		// Run compile again and check if the archive is re-used (cached)
		out, _, err := cli.Run("compile", "-b", "esp8266:esp8266:huzzah", sketchPath.String(), "--config-file", "arduino-cli.yaml", "-v")
		require.NoError(t, err)
		require.True(t, regexp.MustCompile(`(?m)^Using previously compiled file:.*libraries.ArduinoIoTCloud.objs\.a$`).Match(out))
	})
}

func TestCompileWithPrecompileLibrary(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	_, _, err := cli.Run("update")
	require.NoError(t, err)

	_, _, err = cli.Run("core", "install", "arduino:samd@1.8.11")
	require.NoError(t, err)
	fqbn := "arduino:samd:mkrzero"

	// Install precompiled library
	// For more information see:
	// https://arduino.github.io/arduino-cli/latest/library-specification/#precompiled-binaries
	_, _, err = cli.Run("lib", "install", "BSEC Software Library@1.5.1474")
	require.NoError(t, err)
	sketchFolder := cli.SketchbookDir().Join("libraries", "BSEC_Software_Library", "examples", "basic")

	// Compile and verify dependencies detection for fully precompiled library is not skipped
	stdout, _, err := cli.Run("compile", "-b", fqbn, sketchFolder.String(), "-v")
	require.NoError(t, err)
	require.NotContains(t, string(stdout), "Skipping dependencies detection for precompiled library BSEC Software Library")
}

func TestCompileWithFullyPrecompiledLibrary(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	_, _, err := cli.Run("update")
	require.NoError(t, err)

	_, _, err = cli.Run("core", "install", "arduino:mbed@1.3.1")
	require.NoError(t, err)
	fqbn := "arduino:mbed:nano33ble"

	// Create settings with library unsafe install set to true
	envVar := cli.GetDefaultEnv()
	envVar["ARDUINO_LIBRARY_ENABLE_UNSAFE_INSTALL"] = "true"
	_, _, err = cli.RunWithCustomEnv(envVar, "config", "init", "--dest-dir", ".")
	require.NoError(t, err)

	// Install fully precompiled library
	// For more information see:
	// https://arduino.github.io/arduino-cli/latest/library-specification/#precompiled-binaries
	wd, err := paths.Getwd()
	require.NoError(t, err)
	_, _, err = cli.Run("lib", "install",
		"--zip-path", wd.Parent().Join("testdata", "Arduino_TensorFlowLite-2.1.0-ALPHA-precompiled.zip").String(),
		"--config-file", "arduino-cli.yaml",
	)
	require.NoError(t, err)
	sketchFolder := cli.SketchbookDir().Join("libraries", "Arduino_TensorFlowLite", "examples", "hello_world")

	// Install example dependency
	_, _, err = cli.Run("lib", "install", "Arduino_LSM9DS1", "--config-file", "arduino-cli.yaml")
	require.NoError(t, err)

	// Compile and verify dependencies detection for fully precompiled library is skipped
	stdout, _, err := cli.Run("compile", "-b", fqbn, "--config-file", "arduino-cli.yaml", sketchFolder.String(), "-v")
	require.NoError(t, err)
	require.Contains(t, string(stdout), "Skipping dependencies detection for precompiled library Arduino_TensorFlowLite")
}

func TestCompileManuallyInstalledPlatform(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	_, _, err := cli.Run("update")
	require.NoError(t, err)

	sketchName := "CompileSketchManuallyInstalledPlatformUsingPlatformLocalTxt"
	sketchPath := cli.SketchbookDir().Join(sketchName)
	fqbn := "arduino-beta-development:avr:uno"
	_, _, err = cli.Run("sketch", "new", sketchPath.String())
	require.NoError(t, err)

	// Manually installs a core in sketchbooks hardware folder
	gitUrl := "https://github.com/arduino/ArduinoCore-avr.git"
	repoDir := cli.SketchbookDir().Join("hardware", "arduino-beta-development", "avr")
	_, err = git.PlainClone(repoDir.String(), false, &git.CloneOptions{
		URL:           gitUrl,
		ReferenceName: plumbing.NewTagReferenceName("1.8.3"),
	})
	require.NoError(t, err)

	// Installs also the same core via CLI so all the necessary tools are installed
	_, _, err = cli.Run("core", "install", "arduino:avr@1.8.3")
	require.NoError(t, err)

	// Verifies compilation works without issues
	_, _, err = cli.Run("compile", "--clean", "-b", fqbn, sketchPath.String())
	require.NoError(t, err)
}

func TestCompileManuallyInstalledPlatformUsingPlatformLocalTxt(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	_, _, err := cli.Run("update")
	require.NoError(t, err)

	sketchName := "CompileSketchManuallyInstalledPlatformUsingPlatformLocalTxt"
	sketchPath := cli.SketchbookDir().Join(sketchName)
	fqbn := "arduino-beta-development:avr:uno"
	_, _, err = cli.Run("sketch", "new", sketchPath.String())
	require.NoError(t, err)

	// Manually installs a core in sketchbooks hardware folder
	gitUrl := "https://github.com/arduino/ArduinoCore-avr.git"
	repoDir := cli.SketchbookDir().Join("hardware", "arduino-beta-development", "avr")
	_, err = git.PlainClone(repoDir.String(), false, &git.CloneOptions{
		URL:           gitUrl,
		ReferenceName: plumbing.NewTagReferenceName("1.8.3"),
	})
	require.NoError(t, err)

	// Installs also the same core via CLI so all the necessary tools are installed
	_, _, err = cli.Run("core", "install", "arduino:avr@1.8.3")
	require.NoError(t, err)

	// Verifies compilation works without issues
	_, _, err = cli.Run("compile", "--clean", "-b", fqbn, sketchPath.String())
	require.NoError(t, err)

	// Overrides default platform compiler with an unexisting one
	platformLocalTxt := repoDir.Join("platform.local.txt")
	platformLocalTxt.WriteFile([]byte("compiler.c.cmd=my-compiler-that-does-not-exist"))

	// Verifies compilation now fails because compiler is not found
	_, stderr, err := cli.Run("compile", "--clean", "-b", fqbn, sketchPath.String())
	require.Error(t, err)
	require.Contains(t, string(stderr), "my-compiler-that-does-not-exist")
}

func compileWithInvalidBuildOptionJson(t *testing.T, env *integrationtest.Environment, cli *integrationtest.ArduinoCLI) {
	sketchName := "CompileInvalidBuildOptionsJson"
	sketchPath := cli.SketchbookDir().Join(sketchName)
	defer sketchPath.RemoveAll()
	fqbn := "arduino:avr:uno"

	// Create a test sketch
	_, _, err := cli.Run("sketch", "new", sketchPath.String())
	require.NoError(t, err)

	// Get the build directory
	md5 := md5.Sum(([]byte(sketchPath.String())))
	sketchPathMd5 := strings.ToUpper(hex.EncodeToString(md5[:]))
	require.NotEmpty(t, sketchPathMd5)
	buildDir := paths.TempDir().Join("arduino", "sketches", sketchPathMd5)

	_, _, err = cli.Run("compile", "-b", fqbn, sketchPath.String(), "--verbose")
	require.NoError(t, err)

	// Breaks the build.options.json file
	buildOptionsJson := buildDir.Join("build.options.json")
	err = buildOptionsJson.WriteFile([]byte("invalid json"))
	require.NoError(t, err)

	_, _, err = cli.Run("compile", "-b", fqbn, sketchPath.String(), "--verbose")
	require.NoError(t, err)
}

func compileWithRelativeBuildPath(t *testing.T, env *integrationtest.Environment, cli *integrationtest.ArduinoCLI) {
	sketchName := "sketch_simple"
	sketchPath := cli.CopySketch(sketchName)
	defer sketchPath.RemoveAll()
	fqbn := "arduino:avr:uno"

	buildPath := paths.New("..").Join("build_path")
	newWorkingDir := cli.SketchbookDir().Join("working_dir")
	err := newWorkingDir.Mkdir()
	require.NoError(t, err)
	defer newWorkingDir.RemoveAll()
	cli.SetWorkingDir(newWorkingDir)
	_, _, err = cli.Run("compile", "-b", fqbn, "--build-path", buildPath.String(), sketchPath.String(), "-v")
	require.NoError(t, err)
	cli.SetWorkingDir(env.RootDir())

	absoluteBuildPath := cli.SketchbookDir().Join("build_path")
	builtFiles, err := absoluteBuildPath.ReadDir()
	require.NoError(t, err)

	expectedFiles := []string{
		sketchName + ".ino.eep",
		sketchName + ".ino.elf",
		sketchName + ".ino.hex",
		sketchName + ".ino.with_bootloader.bin",
		sketchName + ".ino.with_bootloader.hex",
		"build.options.json",
		"compile_commands.json",
		"core",
		"includes.cache",
		"libraries",
		"sketch",
	}

	foundFiles := []string{}
	for _, builtFile := range builtFiles {
		if sliceIncludes(expectedFiles, builtFile.Base()) {
			foundFiles = append(foundFiles, builtFile.Base())
		}
	}
	sort.Strings(expectedFiles)
	sort.Strings(foundFiles)
	require.Equal(t, expectedFiles, foundFiles)
}

// TODO: remove this when a generic library is introduced
func sliceIncludes[T comparable](slice []T, target T) bool {
	for _, e := range slice {
		if e == target {
			return true
		}
	}
	return false
}

func compileWithFakeSecureBootCore(t *testing.T, env *integrationtest.Environment, cli *integrationtest.ArduinoCLI) {
	sketchName := "SketchSimple"
	sketchPath := cli.SketchbookDir().Join(sketchName)
	defer sketchPath.RemoveAll()
	fqbn := "arduino:avr:uno"

	_, _, err := cli.Run("sketch", "new", sketchPath.String())
	require.NoError(t, err)

	// Verifies compilation works
	_, _, err = cli.Run("compile", "--clean", "-b", fqbn, sketchPath.String())
	require.NoError(t, err)

	// Overrides default platform adding secure_boot support using platform.local.txt
	avrPlatformPath := cli.DataDir().Join("packages", "arduino", "hardware", "avr", "1.8.5", "platform.local.txt")
	defer avrPlatformPath.Remove()
	testPlatformName := "platform_with_secure_boot"
	err = paths.New("..", "testdata", testPlatformName, "platform.local.txt").CopyTo(avrPlatformPath)
	require.NoError(t, err)

	// Overrides default board adding secure boot support using board.local.txt
	avrBoardPath := cli.DataDir().Join("packages", "arduino", "hardware", "avr", "1.8.5", "boards.local.txt")
	defer avrBoardPath.Remove()
	err = paths.New("..", "testdata", testPlatformName, "boards.local.txt").CopyTo(avrBoardPath)
	require.NoError(t, err)

	// Verifies compilation works with secure boot disabled
	stdout, _, err := cli.Run("compile", "--clean", "-b", fqbn+":security=none", sketchPath.String(), "-v")
	require.NoError(t, err)
	require.Contains(t, string(stdout), "echo exit")

	// Verifies compilation works with secure boot enabled
	stdout, _, err = cli.Run("compile", "--clean", "-b", fqbn+":security=sien", sketchPath.String(), "-v")
	require.NoError(t, err)
	require.Contains(t, string(stdout), "Default_Keys/default-signing-key.pem")
	require.Contains(t, string(stdout), "Default_Keys/default-encrypt-key.pem")

	// Verifies compilation does not work with secure boot enabled and using only one flag
	_, stderr, err := cli.Run(
		"compile",
		"--clean",
		"-b",
		fqbn+":security=sien",
		sketchPath.String(),
		"--keys-keychain",
		cli.SketchbookDir().String(),
		"-v",
	)
	require.Error(t, err)
	require.Contains(t, string(stderr), "Flag --sign-key is mandatory when used in conjunction with: --keys-keychain, --sign-key, --encrypt-key")

	// Verifies compilation works with secure boot enabled and when overriding the sign key and encryption key used
	keysDir := cli.SketchbookDir().Join("keys_dir")
	err = keysDir.Mkdir()
	require.NoError(t, err)
	signKeyPath := keysDir.Join("my-sign-key.pem")
	err = signKeyPath.WriteFile([]byte{})
	require.NoError(t, err)
	encryptKeyPath := cli.SketchbookDir().Join("my-encrypt-key.pem")
	err = encryptKeyPath.WriteFile([]byte{})
	require.NoError(t, err)
	stdout, _, err = cli.Run(
		"compile",
		"--clean",
		"-b",
		fqbn+":security=sien",
		sketchPath.String(),
		"--keys-keychain",
		keysDir.String(),
		"--sign-key",
		"my-sign-key.pem",
		"--encrypt-key",
		"my-encrypt-key.pem",
		"-v",
	)
	require.NoError(t, err)
	require.Contains(t, string(stdout), "my-sign-key.pem")
	require.Contains(t, string(stdout), "my-encrypt-key.pem")
}

func preprocessFlagDoNotMessUpWithOutput(t *testing.T, env *integrationtest.Environment, cli *integrationtest.ArduinoCLI) {
	// https://github.com/arduino/arduino-cli/issues/2150

	// go test -v ./internal/integrationtest/compile_1 --run=TestCompile$/PreprocessFlagDoNotMessUpWithOutput

	sketchPath := cli.SketchbookDir().Join("SketchSimple")
	defer sketchPath.RemoveAll()
	fqbn := "arduino:avr:uno"
	_, _, err := cli.Run("sketch", "new", sketchPath.String())
	require.NoError(t, err)

	expected := `#include <Arduino.h>
#line 1 %SKETCH_PATH%

#line 2 %SKETCH_PATH%
void setup();
#line 5 %SKETCH_PATH%
void loop();
#line 2 %SKETCH_PATH%
void setup() {
}

void loop() {
}

`
	expected = strings.ReplaceAll(expected, "%SKETCH_PATH%", cpp.QuoteString(sketchPath.Join("SketchSimple.ino").String()))

	jsonOut, _, err := cli.Run("compile", "-b", fqbn, "--preprocess", sketchPath.String(), "--format", "json")
	require.NoError(t, err)
	var ex struct {
		CompilerOut string `json:"compiler_out"`
	}
	require.NoError(t, json.Unmarshal(jsonOut, &ex))
	require.Equal(t, expected, ex.CompilerOut)

	output, _, err := cli.Run("compile", "-b", fqbn, "--preprocess", sketchPath.String())
	require.NoError(t, err)
	require.Equal(t, expected, string(output))
}

func buildWithCustomBuildPath(t *testing.T, env *integrationtest.Environment, cli *integrationtest.ArduinoCLI) {
	sketchName := "bare_minimum"
	sketchPath := cli.CopySketch(sketchName)
	defer sketchPath.RemoveAll()

	t.Run("OutsideSketch", func(t *testing.T) {
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
		buildDir := paths.TempDir().Join("arduino", "sketches", sketchPathMd5)
		require.NoFileExists(t, buildDir.Join(sketchName+".ino.eep").String())
		require.NoFileExists(t, buildDir.Join(sketchName+".ino.elf").String())
		require.NoFileExists(t, buildDir.Join(sketchName+".ino.hex").String())
		require.NoFileExists(t, buildDir.Join(sketchName+".ino.with_bootloader.bin").String())
		require.NoFileExists(t, buildDir.Join(sketchName+".ino.with_bootloader.hex").String())
	})

	t.Run("InsideSketch", func(t *testing.T) {
		buildPath := sketchPath.Join("build")

		// Run build
		_, _, err := cli.Run("compile", "-b", "arduino:avr:uno", "--build-path", buildPath.String(), sketchPath.String())
		require.NoError(t, err)
		// Run build twice, to verify the build still works when the build directory is present at the start
		_, _, err = cli.Run("compile", "-b", "arduino:avr:uno", "--build-path", buildPath.String(), sketchPath.String())
		require.NoError(t, err)

		// Run again a couple of times with a different build path, to verify that old build
		// path is not copied back in the sketch build recursively.
		// https://github.com/arduino/arduino-cli/issues/2266
		secondBuildPath := sketchPath.Join("build2")
		_, _, err = cli.Run("compile", "-b", "arduino:avr:uno", "--build-path", secondBuildPath.String(), sketchPath.String())
		require.NoError(t, err)
		_, _, err = cli.Run("compile", "-b", "arduino:avr:uno", "--build-path", buildPath.String(), sketchPath.String())
		require.NoError(t, err)
		_, _, err = cli.Run("compile", "-b", "arduino:avr:uno", "--build-path", secondBuildPath.String(), sketchPath.String())
		require.NoError(t, err)
		_, _, err = cli.Run("compile", "-b", "arduino:avr:uno", "--build-path", buildPath.String(), sketchPath.String())
		require.NoError(t, err)

		// Print build path content for debugging purposes
		bp, _ := buildPath.ReadDirRecursive()
		fmt.Println("Build path content:")
		for _, file := range bp {
			fmt.Println("> ", file.String())
		}

		require.False(t, buildPath.Join("sketch", "build2", "sketch").Exist())
	})

	t.Run("SameAsSektch", func(t *testing.T) {
		// Run build
		_, _, err := cli.Run("compile", "-b", "arduino:avr:uno", "--build-path", sketchPath.String(), sketchPath.String())
		require.Error(t, err)
	})
}

func buildWithCustomBuildPathAndOUtputDirFlag(t *testing.T, env *integrationtest.Environment, cli *integrationtest.ArduinoCLI) {
	fqbn := "arduino:avr:uno"
	sketchName := "bare_minimum"
	sketchPath := cli.SketchbookDir().Join(sketchName)
	defer sketchPath.RemoveAll()

	// Create a test sketch
	_, _, err := cli.Run("sketch", "new", sketchPath.String())
	require.NoError(t, err)

	buildPath := cli.DataDir().Join("test_dir", "build_dir")
	outputDirPath := buildPath
	_, _, err = cli.Run("compile", "-b", fqbn, sketchPath.String(), "--build-path", buildPath.String(), "--output-dir", outputDirPath.String())
	require.NoError(t, err)

	// Verifies that output binaries are not empty.
	require.DirExists(t, buildPath.String())
	files := []*paths.Path{
		buildPath.Join(sketchName + ".ino.eep"),
		buildPath.Join(sketchName + ".ino.elf"),
		buildPath.Join(sketchName + ".ino.hex"),
		buildPath.Join(sketchName + ".ino.with_bootloader.bin"),
		buildPath.Join(sketchName + ".ino.with_bootloader.hex"),
	}
	for _, file := range files {
		content, err := file.ReadFile()
		require.NoError(t, err)
		require.NotEmpty(t, content)
	}
}
