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
