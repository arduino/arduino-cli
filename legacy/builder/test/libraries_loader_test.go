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
	"sort"
	"testing"

	"github.com/arduino/arduino-cli/arduino/libraries"
	"github.com/arduino/arduino-cli/legacy/builder"
	"github.com/arduino/arduino-cli/legacy/builder/constants"
	"github.com/arduino/arduino-cli/legacy/builder/types"
	paths "github.com/arduino/go-paths-helper"
	"github.com/stretchr/testify/require"
)

func extractLibraries(ctx *types.Context) []*libraries.Library {
	res := []*libraries.Library{}
	for _, lib := range ctx.LibrariesManager.Libraries {
		for _, libAlternative := range lib.Alternatives {
			res = append(res, libAlternative)
		}
	}
	return res
}
func TestLoadLibrariesAVR(t *testing.T) {
	DownloadCoresAndToolsAndLibraries(t)

	ctx := &types.Context{
		HardwareDirs:         paths.NewPathList(filepath.Join("..", "hardware"), "hardware", "downloaded_hardware"),
		BuiltInLibrariesDirs: paths.NewPathList("downloaded_libraries"),
		OtherLibrariesDirs:   paths.NewPathList("libraries"),
		FQBN:                 parseFQBN(t, "arduino:avr:leonardo"),
	}

	commands := []types.Command{
		&builder.AddAdditionalEntriesToContext{},
		&builder.HardwareLoader{},
		&builder.TargetBoardResolver{},
		&builder.LibrariesLoader{},
	}

	for _, command := range commands {
		err := command.Run(ctx)
		NoError(t, err)
	}

	librariesFolders := ctx.LibrariesManager.LibrariesDir
	require.Equal(t, 3, len(librariesFolders))
	require.True(t, Abs(t, paths.New("downloaded_libraries")).EquivalentTo(librariesFolders[0].Path))
	require.True(t, Abs(t, paths.New("downloaded_hardware", "arduino", "avr", "libraries")).EquivalentTo(librariesFolders[1].Path))
	require.True(t, Abs(t, paths.New("libraries")).EquivalentTo(librariesFolders[2].Path))

	libs := extractLibraries(ctx)
	require.Equal(t, 24, len(libs))

	sort.Sort(ByLibraryName(libs))

	idx := 0

	require.Equal(t, "ANewLibrary-master", libs[idx].Name)

	idx++
	require.Equal(t, "Adafruit_PN532", libs[idx].Name)
	require.True(t, Abs(t, paths.New("downloaded_libraries/Adafruit_PN532")).EquivalentTo(libs[idx].InstallDir))
	require.True(t, Abs(t, paths.New("downloaded_libraries/Adafruit_PN532")).EquivalentTo(libs[idx].SourceDir))
	require.Equal(t, 1, len(libs[idx].Architectures))
	require.Equal(t, constants.LIBRARY_ALL_ARCHS, libs[idx].Architectures[0])
	require.False(t, libs[idx].IsLegacy)

	idx++
	require.Equal(t, "Audio", libs[idx].Name)

	idx++
	require.Equal(t, "Balanduino", libs[idx].Name)
	require.True(t, libs[idx].IsLegacy)

	idx++
	bridgeLib := libs[idx]
	require.Equal(t, "Bridge", bridgeLib.Name)
	require.True(t, Abs(t, paths.New("downloaded_libraries/Bridge")).EquivalentTo(bridgeLib.InstallDir))
	require.True(t, Abs(t, paths.New("downloaded_libraries/Bridge/src")).EquivalentTo(bridgeLib.SourceDir))
	require.Equal(t, 1, len(bridgeLib.Architectures))
	require.Equal(t, constants.LIBRARY_ALL_ARCHS, bridgeLib.Architectures[0])
	require.Equal(t, "Arduino", bridgeLib.Author)
	require.Equal(t, "Arduino <info@arduino.cc>", bridgeLib.Maintainer)

	idx++
	require.Equal(t, "CapacitiveSensor", libs[idx].Name)
	idx++
	require.Equal(t, "EEPROM", libs[idx].Name)
	idx++
	require.Equal(t, "Ethernet", libs[idx].Name)
	idx++
	require.Equal(t, "FakeAudio", libs[idx].Name)
	idx++
	require.Equal(t, "FastLED", libs[idx].Name)
	idx++
	require.Equal(t, "HID", libs[idx].Name)
	idx++
	require.Equal(t, "IRremote", libs[idx].Name)
	idx++
	require.Equal(t, "Robot_IR_Remote", libs[idx].Name)
	idx++
	require.Equal(t, "SPI", libs[idx].Name)
	idx++
	require.Equal(t, "SPI", libs[idx].Name)
	idx++
	require.Equal(t, "ShouldNotRecurseWithOldLibs", libs[idx].Name)
	idx++
	require.Equal(t, "SoftwareSerial", libs[idx].Name)
	idx++
	require.Equal(t, "USBHost", libs[idx].Name)
	idx++
	require.Equal(t, "Wire", libs[idx].Name)

	libs = ctx.LibrariesResolver.AlternativesFor("Audio.h")
	require.Len(t, libs, 2)
	sort.Sort(ByLibraryName(libs))
	require.Equal(t, "Audio", libs[0].Name)
	require.Equal(t, "FakeAudio", libs[1].Name)

	libs = ctx.LibrariesResolver.AlternativesFor("FakeAudio.h")
	require.Len(t, libs, 1)
	require.Equal(t, "FakeAudio", libs[0].Name)

	libs = ctx.LibrariesResolver.AlternativesFor("Adafruit_PN532.h")
	require.Len(t, libs, 1)
	require.Equal(t, "Adafruit_PN532", libs[0].Name)

	libs = ctx.LibrariesResolver.AlternativesFor("IRremote.h")
	require.Len(t, libs, 1)
	require.Equal(t, "IRremote", libs[0].Name)
}

func TestLoadLibrariesSAM(t *testing.T) {
	DownloadCoresAndToolsAndLibraries(t)

	ctx := &types.Context{
		HardwareDirs:         paths.NewPathList(filepath.Join("..", "hardware"), "hardware", "downloaded_hardware"),
		BuiltInLibrariesDirs: paths.NewPathList("downloaded_libraries"),
		OtherLibrariesDirs:   paths.NewPathList("libraries"),
		FQBN:                 parseFQBN(t, "arduino:sam:arduino_due_x_dbg"),
	}

	commands := []types.Command{
		&builder.AddAdditionalEntriesToContext{},
		&builder.HardwareLoader{},
		&builder.TargetBoardResolver{},
		&builder.LibrariesLoader{},
	}

	for _, command := range commands {
		err := command.Run(ctx)
		NoError(t, err)
	}

	librariesFolders := ctx.LibrariesManager.LibrariesDir
	require.Equal(t, 3, len(librariesFolders))
	require.True(t, Abs(t, paths.New("downloaded_libraries")).EquivalentTo(librariesFolders[0].Path))
	require.True(t, Abs(t, paths.New("downloaded_hardware", "arduino", "sam", "libraries")).EquivalentTo(librariesFolders[1].Path))
	require.True(t, Abs(t, paths.New("libraries")).EquivalentTo(librariesFolders[2].Path))

	libraries := extractLibraries(ctx)
	require.Equal(t, 22, len(libraries))

	sort.Sort(ByLibraryName(libraries))

	idx := 0
	require.Equal(t, "ANewLibrary-master", libraries[idx].Name)
	idx++
	require.Equal(t, "Adafruit_PN532", libraries[idx].Name)
	idx++
	require.Equal(t, "Audio", libraries[idx].Name)
	idx++
	require.Equal(t, "Balanduino", libraries[idx].Name)
	idx++
	require.Equal(t, "Bridge", libraries[idx].Name)
	idx++
	require.Equal(t, "CapacitiveSensor", libraries[idx].Name)
	idx++
	require.Equal(t, "Ethernet", libraries[idx].Name)
	idx++
	require.Equal(t, "FakeAudio", libraries[idx].Name)
	idx++
	require.Equal(t, "FastLED", libraries[idx].Name)
	idx++
	require.Equal(t, "HID", libraries[idx].Name)
	idx++
	require.Equal(t, "IRremote", libraries[idx].Name)
	idx++
	require.Equal(t, "Robot_IR_Remote", libraries[idx].Name)
	idx++
	require.Equal(t, "SPI", libraries[idx].Name)
	idx++
	require.Equal(t, "SPI", libraries[idx].Name)
	idx++
	require.Equal(t, "ShouldNotRecurseWithOldLibs", libraries[idx].Name)
	idx++
	require.Equal(t, "USBHost", libraries[idx].Name)
	idx++
	require.Equal(t, "Wire", libraries[idx].Name)

	libs := ctx.LibrariesResolver.AlternativesFor("Audio.h")
	require.Len(t, libs, 2)
	sort.Sort(ByLibraryName(libs))
	require.Equal(t, "Audio", libs[0].Name)
	require.Equal(t, "FakeAudio", libs[1].Name)

	libs = ctx.LibrariesResolver.AlternativesFor("FakeAudio.h")
	require.Len(t, libs, 1)
	require.Equal(t, "FakeAudio", libs[0].Name)

	libs = ctx.LibrariesResolver.AlternativesFor("IRremote.h")
	require.Len(t, libs, 1)
	require.Equal(t, "IRremote", libs[0].Name)
}

func TestLoadLibrariesAVRNoDuplicateLibrariesFolders(t *testing.T) {
	DownloadCoresAndToolsAndLibraries(t)

	ctx := &types.Context{
		HardwareDirs:         paths.NewPathList(filepath.Join("..", "hardware"), "hardware", "downloaded_hardware"),
		BuiltInLibrariesDirs: paths.NewPathList("downloaded_libraries"),
		OtherLibrariesDirs:   paths.NewPathList("libraries", filepath.Join("downloaded_hardware", "arduino", "avr", "libraries")),
		FQBN:                 parseFQBN(t, "arduino:avr:leonardo"),
	}

	commands := []types.Command{
		&builder.AddAdditionalEntriesToContext{},
		&builder.HardwareLoader{},
		&builder.TargetBoardResolver{},
		&builder.LibrariesLoader{},
	}

	for _, command := range commands {
		err := command.Run(ctx)
		NoError(t, err)
	}

	librariesFolders := ctx.LibrariesManager.LibrariesDir
	require.Equal(t, 3, len(librariesFolders))
	require.True(t, Abs(t, paths.New("downloaded_libraries")).EquivalentTo(librariesFolders[0].Path))
	require.True(t, Abs(t, paths.New("downloaded_hardware", "arduino", "avr", "libraries")).EquivalentTo(librariesFolders[1].Path))
	require.True(t, Abs(t, paths.New("libraries")).EquivalentTo(librariesFolders[2].Path))
}

func TestLoadLibrariesMyAVRPlatform(t *testing.T) {
	DownloadCoresAndToolsAndLibraries(t)

	ctx := &types.Context{
		HardwareDirs:         paths.NewPathList(filepath.Join("..", "hardware"), "hardware", "user_hardware", "downloaded_hardware"),
		BuiltInLibrariesDirs: paths.NewPathList("downloaded_libraries"),
		OtherLibrariesDirs:   paths.NewPathList("libraries", filepath.Join("downloaded_hardware", "arduino", "avr", "libraries")),
		FQBN:                 parseFQBN(t, "my_avr_platform:avr:custom_yun"),
	}

	commands := []types.Command{
		&builder.AddAdditionalEntriesToContext{},
		&builder.HardwareLoader{},
		&builder.TargetBoardResolver{},
		&builder.LibrariesLoader{},
	}

	for _, command := range commands {
		err := command.Run(ctx)
		NoError(t, err)
	}

	librariesFolders := ctx.LibrariesManager.LibrariesDir
	require.Equal(t, 4, len(librariesFolders))
	require.True(t, Abs(t, paths.New("downloaded_libraries")).EquivalentTo(librariesFolders[0].Path))
	require.True(t, Abs(t, paths.New("downloaded_hardware", "arduino", "avr", "libraries")).EquivalentTo(librariesFolders[1].Path))
	require.True(t, Abs(t, paths.New("user_hardware", "my_avr_platform", "avr", "libraries")).EquivalentTo(librariesFolders[2].Path))
	require.True(t, Abs(t, paths.New("libraries")).EquivalentTo(librariesFolders[3].Path))
}
