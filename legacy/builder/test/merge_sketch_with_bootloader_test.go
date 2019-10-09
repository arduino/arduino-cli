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
	"path/filepath"
	"strings"
	"testing"

	"github.com/arduino/arduino-cli/legacy/builder"
	"github.com/arduino/arduino-cli/legacy/builder/constants"
	"github.com/arduino/arduino-cli/legacy/builder/types"
	paths "github.com/arduino/go-paths-helper"
	"github.com/stretchr/testify/require"
)

func TestMergeSketchWithBootloader(t *testing.T) {
	DownloadCoresAndToolsAndLibraries(t)

	ctx := &types.Context{
		HardwareDirs:         paths.NewPathList(filepath.Join("..", "hardware"), "hardware", "downloaded_hardware"),
		BuiltInToolsDirs:     paths.NewPathList("downloaded_tools"),
		BuiltInLibrariesDirs: paths.NewPathList("downloaded_libraries"),
		OtherLibrariesDirs:   paths.NewPathList("libraries"),
		SketchLocation:       paths.New("sketch1", "sketch.ino"),
		FQBN:                 parseFQBN(t, "arduino:avr:uno"),
		ArduinoAPIVersion:    "10600",
	}

	buildPath := SetupBuildPath(t, ctx)
	defer buildPath.RemoveAll()

	err := buildPath.Join("sketch").MkdirAll()
	NoError(t, err)

	fakeSketchHex := "row 1\n" +
		"row 2\n"
	err = buildPath.Join("sketch", "sketch.ino.hex").WriteFile([]byte(fakeSketchHex))
	NoError(t, err)

	commands := []types.Command{
		&builder.ContainerSetupHardwareToolsLibsSketchAndProps{},
		&builder.MergeSketchWithBootloader{},
	}

	for _, command := range commands {
		err := command.Run(ctx)
		NoError(t, err)
	}

	bytes, err := buildPath.Join("sketch", "sketch.ino.with_bootloader.hex").ReadFile()
	NoError(t, err)
	mergedSketchHex := string(bytes)

	require.True(t, strings.HasPrefix(mergedSketchHex, "row 1\n:107E0000112484B714BE81FFF0D085E080938100F7\n"))
	require.True(t, strings.HasSuffix(mergedSketchHex, ":0400000300007E007B\n:00000001FF\n"))
}

func TestMergeSketchWithBootloaderSketchInBuildPath(t *testing.T) {
	DownloadCoresAndToolsAndLibraries(t)

	ctx := &types.Context{
		HardwareDirs:         paths.NewPathList(filepath.Join("..", "hardware"), "hardware", "downloaded_hardware"),
		BuiltInToolsDirs:     paths.NewPathList("downloaded_tools"),
		BuiltInLibrariesDirs: paths.NewPathList("downloaded_libraries"),
		OtherLibrariesDirs:   paths.NewPathList("libraries"),
		SketchLocation:       paths.New("sketch1", "sketch.ino"),
		FQBN:                 parseFQBN(t, "arduino:avr:uno"),
		ArduinoAPIVersion:    "10600",
	}

	buildPath := SetupBuildPath(t, ctx)
	defer buildPath.RemoveAll()

	err := buildPath.Join("sketch").MkdirAll()
	NoError(t, err)

	fakeSketchHex := "row 1\n" +
		"row 2\n"
	err = buildPath.Join("sketch.ino.hex").WriteFile([]byte(fakeSketchHex))
	NoError(t, err)

	commands := []types.Command{
		&builder.ContainerSetupHardwareToolsLibsSketchAndProps{},
		&builder.MergeSketchWithBootloader{},
	}

	for _, command := range commands {
		err := command.Run(ctx)
		NoError(t, err)
	}

	bytes, err := buildPath.Join("sketch.ino.with_bootloader.hex").ReadFile()
	NoError(t, err)
	mergedSketchHex := string(bytes)

	require.True(t, strings.HasPrefix(mergedSketchHex, "row 1\n:107E0000112484B714BE81FFF0D085E080938100F7\n"))
	require.True(t, strings.HasSuffix(mergedSketchHex, ":0400000300007E007B\n:00000001FF\n"))
}

func TestMergeSketchWithBootloaderWhenNoBootloaderAvailable(t *testing.T) {
	DownloadCoresAndToolsAndLibraries(t)

	ctx := &types.Context{
		HardwareDirs:         paths.NewPathList(filepath.Join("..", "hardware"), "hardware", "downloaded_hardware"),
		BuiltInToolsDirs:     paths.NewPathList("downloaded_tools"),
		BuiltInLibrariesDirs: paths.NewPathList("downloaded_libraries"),
		OtherLibrariesDirs:   paths.NewPathList("libraries"),
		SketchLocation:       paths.New("sketch1", "sketch.ino"),
		FQBN:                 parseFQBN(t, "arduino:avr:uno"),
		ArduinoAPIVersion:    "10600",
	}

	buildPath := SetupBuildPath(t, ctx)
	defer buildPath.RemoveAll()

	commands := []types.Command{
		&builder.ContainerSetupHardwareToolsLibsSketchAndProps{},
	}

	for _, command := range commands {
		err := command.Run(ctx)
		NoError(t, err)
	}

	buildProperties := ctx.BuildProperties
	buildProperties.Remove(constants.BUILD_PROPERTIES_BOOTLOADER_NOBLINK)
	buildProperties.Remove(constants.BUILD_PROPERTIES_BOOTLOADER_FILE)

	command := &builder.MergeSketchWithBootloader{}
	err := command.Run(ctx)
	NoError(t, err)

	exist, err := buildPath.Join("sketch.ino.with_bootloader.hex").ExistCheck()
	require.NoError(t, err)
	require.False(t, exist)
}

func TestMergeSketchWithBootloaderPathIsParameterized(t *testing.T) {
	DownloadCoresAndToolsAndLibraries(t)

	ctx := &types.Context{
		HardwareDirs:         paths.NewPathList(filepath.Join("..", "hardware"), "hardware", "downloaded_hardware", "user_hardware"),
		BuiltInToolsDirs:     paths.NewPathList("downloaded_tools"),
		BuiltInLibrariesDirs: paths.NewPathList("downloaded_libraries"),
		OtherLibrariesDirs:   paths.NewPathList("libraries"),
		SketchLocation:       paths.New("sketch1", "sketch.ino"),
		FQBN:                 parseFQBN(t, "my_avr_platform:avr:mymega:cpu=atmega2560"),
		ArduinoAPIVersion:    "10600",
	}

	buildPath := SetupBuildPath(t, ctx)
	defer buildPath.RemoveAll()

	err := buildPath.Join("sketch").MkdirAll()
	NoError(t, err)

	fakeSketchHex := "row 1\n" +
		"row 2\n"
	err = buildPath.Join("sketch", "sketch.ino.hex").WriteFile([]byte(fakeSketchHex))
	NoError(t, err)

	commands := []types.Command{
		&builder.ContainerSetupHardwareToolsLibsSketchAndProps{},
		&builder.MergeSketchWithBootloader{},
	}

	for _, command := range commands {
		err := command.Run(ctx)
		NoError(t, err)
	}

	bytes, err := buildPath.Join("sketch", "sketch.ino.with_bootloader.hex").ReadFile()
	NoError(t, err)
	mergedSketchHex := string(bytes)

	require.True(t, strings.HasPrefix(mergedSketchHex, "row 1\n:020000023000CC"))
	require.True(t, strings.HasSuffix(mergedSketchHex, ":040000033000E000E9\n:00000001FF\n"))
}
