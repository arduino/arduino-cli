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
	"strings"
	"testing"

	"github.com/arduino/arduino-cli/internal/integrationtest"
	"github.com/arduino/go-paths-helper"
	"github.com/stretchr/testify/require"
)

func TestCompileWithOutputDirFlag(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	// Init the environment explicitly
	_, _, err := cli.Run("core", "update-index")
	require.NoError(t, err)

	// Download latest AVR
	_, _, err = cli.Run("core", "install", "arduino:avr")
	require.NoError(t, err)

	sketchName := "CompileWithOutputDir"
	sketchPath := cli.SketchbookDir().Join(sketchName)
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

func TestCompileWithExportBinariesFlag(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	// Init the environment explicitly
	_, _, err := cli.Run("core", "update-index")
	require.NoError(t, err)

	// Download latest AVR
	_, _, err = cli.Run("core", "install", "arduino:avr")
	require.NoError(t, err)

	sketchName := "CompileWithExportBinariesFlag"
	sketchPath := cli.SketchbookDir().Join(sketchName)
	fqbn := "arduino:avr:uno"

	// Create a test sketch
	_, _, err = cli.Run("sketch", "new", sketchPath.String())
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

func TestCompileWithCustomBuildPath(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	// Init the environment explicitly
	_, _, err := cli.Run("core", "update-index")
	require.NoError(t, err)

	// Download latest AVR
	_, _, err = cli.Run("core", "install", "arduino:avr")
	require.NoError(t, err)

	sketchName := "CompileWithBuildPath"
	sketchPath := cli.SketchbookDir().Join(sketchName)
	fqbn := "arduino:avr:uno"

	// Create a test sketch
	_, _, err = cli.Run("sketch", "new", sketchPath.String())
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

func TestCompileWithExportBinariesEnvVar(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	// Init the environment explicitly
	_, _, err := cli.Run("core", "update-index")
	require.NoError(t, err)

	// Download latest AVR
	_, _, err = cli.Run("core", "install", "arduino:avr")
	require.NoError(t, err)

	sketchName := "CompileWithExportBinariesEnvVar"
	sketchPath := cli.SketchbookDir().Join(sketchName)
	fqbn := "arduino:avr:uno"

	// Create a test sketch
	_, _, err = cli.Run("sketch", "new", sketchPath.String())
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

func TestCompileWithInvalidUrl(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	// Init the environment explicitly
	_, _, err := cli.Run("core", "update-index")
	require.NoError(t, err)

	// Download latest AVR
	_, _, err = cli.Run("core", "install", "arduino:avr")
	require.NoError(t, err)

	sketchName := "CompileWithInvalidURL"
	sketchPath := cli.SketchbookDir().Join(sketchName)
	fqbn := "arduino:avr:uno"

	// Create a test sketch
	_, _, err = cli.Run("sketch", "new", sketchPath.String())
	require.NoError(t, err)

	_, _, err = cli.Run("config", "init", "--dest-dir", ".", "--additional-urls", "https://example.com/package_example_index.json")
	require.NoError(t, err)

	_, stderr, err := cli.Run("compile", "-b", fqbn, sketchPath.String())
	require.NoError(t, err)
	require.Contains(t, string(stderr), "Error initializing instance: Loading index file: loading json index file")
	expectedIndexfile := cli.DataDir().Join("package_example_index.json")
	require.Contains(t, string(stderr), "loading json index file "+expectedIndexfile.String()+": open "+expectedIndexfile.String()+":")
}
