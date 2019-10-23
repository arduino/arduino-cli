/*
 * This file is part of Arduino Builder.
 *
 * Arduino Builder is free software; you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation; either version 2 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program; if not, write to the Free Software
 * Foundation, Inc., 51 Franklin St, Fifth Floor, Boston, MA  02110-1301  USA
 *
 * As a special exception, you may use this file as part of a free software
 * library without restriction.  Specifically, if other files instantiate
 * templates or use macros or inline functions from this file, or you compile
 * this file and link it with other files to produce an executable, this
 * file does not by itself cause the resulting executable to be covered by
 * the GNU General Public License.  This exception does not however
 * invalidate any other reasons why the executable file might be covered by
 * the GNU General Public License.
 *
 * Copyright 2015 Arduino LLC (http://www.arduino.cc/)
 */

package test

import (
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/arduino/go-paths-helper"

	"github.com/arduino/arduino-cli/legacy/builder"
	"github.com/arduino/arduino-cli/legacy/builder/builder_utils"
	"github.com/arduino/arduino-cli/legacy/builder/constants"
	"github.com/arduino/arduino-cli/legacy/builder/types"
	"github.com/stretchr/testify/require"
)

func prepareBuilderTestContext(t *testing.T, sketchPath *paths.Path, fqbn string) *types.Context {
	return &types.Context{
		SketchLocation:       sketchPath,
		FQBN:                 parseFQBN(t, fqbn),
		HardwareDirs:         paths.NewPathList(filepath.Join("..", "hardware"), "hardware", "downloaded_hardware"),
		BuiltInToolsDirs:     paths.NewPathList("downloaded_tools"),
		BuiltInLibrariesDirs: paths.NewPathList("downloaded_libraries"),
		OtherLibrariesDirs:   paths.NewPathList("libraries"),
		ArduinoAPIVersion:    "10600",
		Verbose:              false,
	}
}

func TestBuilderEmptySketch(t *testing.T) {
	DownloadCoresAndToolsAndLibraries(t)

	ctx := prepareBuilderTestContext(t, paths.New("sketch1", "sketch.ino"), "arduino:avr:uno")
	ctx.DebugLevel = 10

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
	exist, err = buildPath.Join(constants.FOLDER_SKETCH, "sketch.ino.cpp.o").ExistCheck()
	NoError(t, err)
	require.True(t, exist)
	exist, err = buildPath.Join("sketch.ino.elf").ExistCheck()
	NoError(t, err)
	require.True(t, exist)
	exist, err = buildPath.Join("sketch.ino.hex").ExistCheck()
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

	ctx := prepareBuilderTestContext(t, paths.New("sketch_no_functions", "main.ino"), "RedBearLab:avr:blend")
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

	ctx := prepareBuilderTestContext(t, paths.New("sketch_with_backup_files", "sketch.ino"), "arduino:avr:uno")
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

	ctx := prepareBuilderTestContext(t, paths.New("sketch_with_old_lib", "sketch.ino"), "arduino:avr:uno")

	buildPath := SetupBuildPath(t, ctx)
	defer buildPath.RemoveAll()

	// Run builder
	command := builder.Builder{}
	err := command.Run(ctx)
	NoError(t, err)
}

func TestBuilderSketchWithSubfolders(t *testing.T) {
	DownloadCoresAndToolsAndLibraries(t)

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

	ctx := prepareBuilderTestContext(t, paths.New("sketch1", "sketch.ino"), "arduino:avr:uno")

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

	ctx := prepareBuilderTestContext(t, paths.New("sketch1", "sketch.ino"), "arduino:avr:uno")

	SetupBuildPath(t, ctx)
	defer ctx.BuildPath.RemoveAll()
	SetupBuildCachePath(t, ctx)
	defer ctx.BuildCachePath.RemoveAll()

	// Run build
	bldr := builder.Builder{}
	err := bldr.Run(ctx)
	NoError(t, err)

	// Pick timestamp of cached core
	coreFolder := paths.New("downloaded_hardware", "arduino", "avr")
	coreFileName := builder_utils.GetCachedCoreArchiveFileName(ctx.FQBN.String(), coreFolder)
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
