// This file is part of arduino-cli.
//
// Copyright 2020 ARDUINO SA (http://www.arduino.cc/)
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

package test

import (
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/arduino/arduino-cli/arduino/sketch"
	"github.com/arduino/arduino-cli/legacy/builder"
	"github.com/arduino/arduino-cli/legacy/builder/constants"
	"github.com/arduino/arduino-cli/legacy/builder/phases"
	"github.com/arduino/arduino-cli/legacy/builder/types"
	"github.com/arduino/go-paths-helper"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func prepareBuilderTestContext(t *testing.T, sketchPath *paths.Path, fqbn string) *types.Context {
	sk, err := sketch.New(sketchPath)
	require.NoError(t, err)
	return &types.Context{
		Sketch:               sk,
		FQBN:                 parseFQBN(t, fqbn),
		HardwareDirs:         paths.NewPathList(filepath.Join("..", "hardware"), "downloaded_hardware"),
		BuiltInToolsDirs:     paths.NewPathList("downloaded_tools"),
		BuiltInLibrariesDirs: paths.New("downloaded_libraries"),
		OtherLibrariesDirs:   paths.NewPathList("libraries"),
		ArduinoAPIVersion:    "10600",
		Verbose:              false,
	}
}

func TestBuilderEmptySketch(t *testing.T) {
	DownloadCoresAndToolsAndLibraries(t)

	ctx := prepareBuilderTestContext(t, paths.New("sketch1", "sketch1.ino"), "arduino:avr:uno")

	buildPath := SetupBuildPath(t, ctx)
	defer buildPath.RemoveAll()

	// Run builder
	command := builder.Builder{}
	err := command.Run(ctx)
	NoError(t, err)

	exist, err := buildPath.Join(constants.FOLDER_CORE, "HardwareSerial.cpp.o").ExistCheck()
	NoError(t, err)
	require.True(t, exist)
	exist, err = buildPath.Join(constants.FOLDER_PREPROC, constants.FILE_CTAGS_TARGET_FOR_GCC_MINUS_E).ExistCheck()
	NoError(t, err)
	require.True(t, exist)
	exist, err = buildPath.Join(constants.FOLDER_SKETCH, "sketch1.ino.cpp.o").ExistCheck()
	NoError(t, err)
	require.True(t, exist)
	exist, err = buildPath.Join("sketch1.ino.elf").ExistCheck()
	NoError(t, err)
	require.True(t, exist)
	exist, err = buildPath.Join("sketch1.ino.hex").ExistCheck()
	NoError(t, err)
	require.True(t, exist)
}

func TestBuilderBridge(t *testing.T) {
	DownloadCoresAndToolsAndLibraries(t)

	ctx := prepareBuilderTestContext(t, paths.New("downloaded_libraries", "Bridge", "examples", "Bridge", "Bridge.ino"), "arduino:avr:leonardo")

	buildPath := SetupBuildPath(t, ctx)
	defer buildPath.RemoveAll()

	// Run builder
	command := builder.Builder{}
	err := command.Run(ctx)
	NoError(t, err)

	exist, err := buildPath.Join(constants.FOLDER_CORE, "HardwareSerial.cpp.o").ExistCheck()
	NoError(t, err)
	require.True(t, exist)
	exist, err = buildPath.Join(constants.FOLDER_PREPROC, constants.FILE_CTAGS_TARGET_FOR_GCC_MINUS_E).ExistCheck()
	NoError(t, err)
	require.True(t, exist)
	exist, err = buildPath.Join(constants.FOLDER_SKETCH, "Bridge.ino.cpp.o").ExistCheck()
	NoError(t, err)
	require.True(t, exist)
	exist, err = buildPath.Join("Bridge.ino.elf").ExistCheck()
	NoError(t, err)
	require.True(t, exist)
	exist, err = buildPath.Join("Bridge.ino.hex").ExistCheck()
	NoError(t, err)
	require.True(t, exist)
	exist, err = buildPath.Join("libraries", "Bridge", "Mailbox.cpp.o").ExistCheck()
	NoError(t, err)
	require.True(t, exist)
}

func TestBuilderSketchWithConfig(t *testing.T) {
	DownloadCoresAndToolsAndLibraries(t)

	ctx := prepareBuilderTestContext(t, paths.New("sketch_with_config", "sketch_with_config.ino"), "arduino:avr:leonardo")

	buildPath := SetupBuildPath(t, ctx)
	defer buildPath.RemoveAll()

	// Run builder
	command := builder.Builder{}
	err := command.Run(ctx)
	NoError(t, err)

	exist, err := buildPath.Join(constants.FOLDER_CORE, "HardwareSerial.cpp.o").ExistCheck()
	NoError(t, err)
	require.True(t, exist)
	exist, err = buildPath.Join(constants.FOLDER_PREPROC, constants.FILE_CTAGS_TARGET_FOR_GCC_MINUS_E).ExistCheck()
	NoError(t, err)
	require.True(t, exist)
	exist, err = buildPath.Join(constants.FOLDER_SKETCH, "sketch_with_config.ino.cpp.o").ExistCheck()
	NoError(t, err)
	require.True(t, exist)
	exist, err = buildPath.Join("sketch_with_config.ino.elf").ExistCheck()
	NoError(t, err)
	require.True(t, exist)
	exist, err = buildPath.Join("sketch_with_config.ino.hex").ExistCheck()
	NoError(t, err)
	require.True(t, exist)
	exist, err = buildPath.Join("libraries", "Bridge", "Mailbox.cpp.o").ExistCheck()
	NoError(t, err)
	require.True(t, exist)
}

func TestBuilderBridgeTwice(t *testing.T) {
	DownloadCoresAndToolsAndLibraries(t)

	ctx := prepareBuilderTestContext(t, paths.New("downloaded_libraries", "Bridge", "examples", "Bridge", "Bridge.ino"), "arduino:avr:leonardo")

	buildPath := SetupBuildPath(t, ctx)
	defer buildPath.RemoveAll()

	// Run builder
	command := builder.Builder{}
	err := command.Run(ctx)
	NoError(t, err)

	// Run builder again
	command = builder.Builder{}
	err = command.Run(ctx)
	NoError(t, err)

	exist, err := buildPath.Join(constants.FOLDER_CORE, "HardwareSerial.cpp.o").ExistCheck()
	NoError(t, err)
	require.True(t, exist)
	exist, err = buildPath.Join(constants.FOLDER_PREPROC, constants.FILE_CTAGS_TARGET_FOR_GCC_MINUS_E).ExistCheck()
	NoError(t, err)
	require.True(t, exist)
	exist, err = buildPath.Join(constants.FOLDER_SKETCH, "Bridge.ino.cpp.o").ExistCheck()
	NoError(t, err)
	require.True(t, exist)
	exist, err = buildPath.Join("Bridge.ino.elf").ExistCheck()
	NoError(t, err)
	require.True(t, exist)
	exist, err = buildPath.Join("Bridge.ino.hex").ExistCheck()
	NoError(t, err)
	require.True(t, exist)
	exist, err = buildPath.Join("libraries", "Bridge", "Mailbox.cpp.o").ExistCheck()
	NoError(t, err)
	require.True(t, exist)
}

func TestBuilderBridgeSAM(t *testing.T) {
	DownloadCoresAndToolsAndLibraries(t)

	ctx := prepareBuilderTestContext(t, paths.New("downloaded_libraries", "Bridge", "examples", "Bridge", "Bridge.ino"), "arduino:sam:arduino_due_x_dbg")
	ctx.WarningsLevel = "all"

	buildPath := SetupBuildPath(t, ctx)
	defer buildPath.RemoveAll()

	// Run builder
	command := builder.Builder{}
	err := command.Run(ctx)
	NoError(t, err)

	exist, err := buildPath.Join(constants.FOLDER_CORE, "syscalls_sam3.c.o").ExistCheck()
	NoError(t, err)
	require.True(t, exist)
	exist, err = buildPath.Join(constants.FOLDER_CORE, "USB", "PluggableUSB.cpp.o").ExistCheck()
	NoError(t, err)
	require.True(t, exist)
	exist, err = buildPath.Join(constants.FOLDER_CORE, "avr", "dtostrf.c.d").ExistCheck()
	NoError(t, err)
	require.True(t, exist)
	exist, err = buildPath.Join(constants.FOLDER_PREPROC, constants.FILE_CTAGS_TARGET_FOR_GCC_MINUS_E).ExistCheck()
	NoError(t, err)
	require.True(t, exist)
	exist, err = buildPath.Join(constants.FOLDER_SKETCH, "Bridge.ino.cpp.o").ExistCheck()
	NoError(t, err)
	require.True(t, exist)
	exist, err = buildPath.Join("Bridge.ino.elf").ExistCheck()
	NoError(t, err)
	require.True(t, exist)
	exist, err = buildPath.Join("Bridge.ino.bin").ExistCheck()
	NoError(t, err)
	require.True(t, exist)
	exist, err = buildPath.Join("libraries", "Bridge", "Mailbox.cpp.o").ExistCheck()
	NoError(t, err)
	require.True(t, exist)

	cmd := exec.Command(filepath.Join("downloaded_tools", "arm-none-eabi-gcc", "4.8.3-2014q1", "bin", "arm-none-eabi-objdump"), "-f", buildPath.Join(constants.FOLDER_CORE, "core.a").String())
	bytes, err := cmd.CombinedOutput()
	NoError(t, err)
	require.NotContains(t, string(bytes), "variant.cpp.o")
}

func TestBuilderBridgeRedBearLab(t *testing.T) {
	DownloadCoresAndToolsAndLibraries(t)

	ctx := prepareBuilderTestContext(t, paths.New("downloaded_libraries", "Bridge", "examples", "Bridge", "Bridge.ino"), "RedBearLab:avr:blend")
	ctx.HardwareDirs = append(ctx.HardwareDirs, paths.New("downloaded_board_manager_stuff"))
	ctx.BuiltInToolsDirs = append(ctx.BuiltInToolsDirs, paths.New("downloaded_board_manager_stuff"))

	buildPath := SetupBuildPath(t, ctx)
	defer buildPath.RemoveAll()

	// Run builder
	command := builder.Builder{}
	err := command.Run(ctx)
	NoError(t, err)

	exist, err := buildPath.Join(constants.FOLDER_CORE, "HardwareSerial.cpp.o").ExistCheck()
	NoError(t, err)
	require.True(t, exist)
	exist, err = buildPath.Join(constants.FOLDER_PREPROC, constants.FILE_CTAGS_TARGET_FOR_GCC_MINUS_E).ExistCheck()
	NoError(t, err)
	require.True(t, exist)
	exist, err = buildPath.Join(constants.FOLDER_SKETCH, "Bridge.ino.cpp.o").ExistCheck()
	NoError(t, err)
	require.True(t, exist)
	exist, err = buildPath.Join("Bridge.ino.elf").ExistCheck()
	NoError(t, err)
	require.True(t, exist)
	exist, err = buildPath.Join("Bridge.ino.hex").ExistCheck()
	NoError(t, err)
	require.True(t, exist)
	exist, err = buildPath.Join("libraries", "Bridge", "Mailbox.cpp.o").ExistCheck()
	NoError(t, err)
	require.True(t, exist)
}

func TestBuilderSketchNoFunctions(t *testing.T) {
	DownloadCoresAndToolsAndLibraries(t)

	ctx := prepareBuilderTestContext(t, paths.New("sketch_no_functions", "sketch_no_functions.ino"), "RedBearLab:avr:blend")
	ctx.HardwareDirs = append(ctx.HardwareDirs, paths.New("downloaded_board_manager_stuff"))
	ctx.BuiltInToolsDirs = append(ctx.BuiltInToolsDirs, paths.New("downloaded_board_manager_stuff"))

	buildPath := SetupBuildPath(t, ctx)
	defer buildPath.RemoveAll()

	// Run builder
	command := builder.Builder{}
	err := command.Run(ctx)
	require.Error(t, err)
}

func TestBuilderSketchWithBackup(t *testing.T) {
	DownloadCoresAndToolsAndLibraries(t)

	ctx := prepareBuilderTestContext(t, paths.New("sketch_with_backup_files", "sketch_with_backup_files.ino"), "arduino:avr:uno")
	ctx.HardwareDirs = append(ctx.HardwareDirs, paths.New("downloaded_board_manager_stuff"))
	ctx.BuiltInToolsDirs = append(ctx.BuiltInToolsDirs, paths.New("downloaded_board_manager_stuff"))

	buildPath := SetupBuildPath(t, ctx)
	defer buildPath.RemoveAll()

	// Run builder
	command := builder.Builder{}
	err := command.Run(ctx)
	NoError(t, err)
}

func TestBuilderSketchWithOldLib(t *testing.T) {
	DownloadCoresAndToolsAndLibraries(t)

	ctx := prepareBuilderTestContext(t, paths.New("sketch_with_old_lib", "sketch_with_old_lib.ino"), "arduino:avr:uno")

	buildPath := SetupBuildPath(t, ctx)
	defer buildPath.RemoveAll()

	// Run builder
	command := builder.Builder{}
	err := command.Run(ctx)
	NoError(t, err)
}

func TestBuilderSketchWithSubfolders(t *testing.T) {
	DownloadCoresAndToolsAndLibraries(t)

	logrus.SetLevel(logrus.DebugLevel)
	ctx := prepareBuilderTestContext(t, paths.New("sketch_with_subfolders", "sketch_with_subfolders.ino"), "arduino:avr:uno")

	buildPath := SetupBuildPath(t, ctx)
	defer buildPath.RemoveAll()

	// Run builder
	command := builder.Builder{}
	err := command.Run(ctx)
	NoError(t, err)
}

func TestBuilderSketchBuildPathContainsUnusedPreviouslyCompiledLibrary(t *testing.T) {
	DownloadCoresAndToolsAndLibraries(t)

	ctx := prepareBuilderTestContext(t, paths.New("downloaded_libraries", "Bridge", "examples", "Bridge", "Bridge.ino"), "arduino:avr:leonardo")

	buildPath := SetupBuildPath(t, ctx)
	defer buildPath.RemoveAll()

	NoError(t, buildPath.Join("libraries", "SPI").MkdirAll())

	// Run builder
	command := builder.Builder{}
	err := command.Run(ctx)
	NoError(t, err)

	exist, err := buildPath.Join("libraries", "SPI").ExistCheck()
	NoError(t, err)
	require.False(t, exist)
	exist, err = buildPath.Join("libraries", "Bridge").ExistCheck()
	NoError(t, err)
	require.True(t, exist)
}

func TestBuilderWithBuildPathInSketchDir(t *testing.T) {
	DownloadCoresAndToolsAndLibraries(t)

	ctx := prepareBuilderTestContext(t, paths.New("sketch1", "sketch1.ino"), "arduino:avr:uno")

	var err error
	ctx.BuildPath, err = paths.New("sketch1", "build").Abs()
	NoError(t, err)
	defer ctx.BuildPath.RemoveAll()

	// Run build
	command := builder.Builder{}
	err = command.Run(ctx)
	NoError(t, err)

	// Run build twice, to verify the build still works when the
	// build directory is present at the start
	err = command.Run(ctx)
	NoError(t, err)
}

func TestBuilderCacheCoreAFile(t *testing.T) {
	DownloadCoresAndToolsAndLibraries(t)

	ctx := prepareBuilderTestContext(t, paths.New("sketch1", "sketch1.ino"), "arduino:avr:uno")

	SetupBuildPath(t, ctx)
	defer ctx.BuildPath.RemoveAll()
	SetupBuildCachePath(t, ctx)
	defer ctx.CoreBuildCachePath.RemoveAll()

	// Run build
	bldr := builder.Builder{}
	err := bldr.Run(ctx)
	NoError(t, err)

	// Pick timestamp of cached core
	coreFolder := paths.New("downloaded_hardware", "arduino", "avr")
	coreFileName := phases.GetCachedCoreArchiveFileName(ctx.FQBN.String(), ctx.OptimizationFlags, coreFolder)
	cachedCoreFile := ctx.CoreBuildCachePath.Join(coreFileName)
	coreStatBefore, err := cachedCoreFile.Stat()
	require.NoError(t, err)

	// Run build again, to verify that the builder skips rebuilding core.a
	err = bldr.Run(ctx)
	NoError(t, err)

	coreStatAfterRebuild, err := cachedCoreFile.Stat()
	require.NoError(t, err)
	require.Equal(t, coreStatBefore.ModTime(), coreStatAfterRebuild.ModTime())

	// Touch a file of the core and check if the builder invalidate the cache
	time.Sleep(time.Second)
	now := time.Now().Local()
	err = coreFolder.Join("cores", "arduino", "Arduino.h").Chtimes(now, now)
	require.NoError(t, err)

	// Run build again, to verify that the builder rebuilds core.a
	err = bldr.Run(ctx)
	NoError(t, err)

	coreStatAfterTouch, err := cachedCoreFile.Stat()
	require.NoError(t, err)
	require.NotEqual(t, coreStatBefore.ModTime(), coreStatAfterTouch.ModTime())
}
