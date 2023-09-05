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
	"sort"
	"testing"

	bldr "github.com/arduino/arduino-cli/arduino/builder"
	"github.com/arduino/arduino-cli/legacy/builder"
	"github.com/arduino/arduino-cli/legacy/builder/types"
	paths "github.com/arduino/go-paths-helper"
	"github.com/stretchr/testify/require"
)

func TestIncludesToIncludeFoldersDuplicateLibs(t *testing.T) {
	ctx := &types.Context{
		HardwareDirs:         paths.NewPathList(filepath.Join("..", "hardware"), "downloaded_hardware", "user_hardware"),
		BuiltInToolsDirs:     paths.NewPathList("downloaded_tools"),
		BuiltInLibrariesDirs: paths.New("downloaded_libraries"),
		Verbose:              true,
	}
	ctx = prepareBuilderTestContext(t, ctx, paths.New("user_hardware", "my_avr_platform", "avr", "libraries", "SPI", "examples", "BarometricPressureSensor", "BarometricPressureSensor.ino"), "my_avr_platform:avr:custom_yun")
	defer cleanUpBuilderTestContext(t, ctx)

	var _err error
	commands := []types.Command{
		&builder.ContainerSetupHardwareToolsLibsSketchAndProps{},
		types.BareCommand(func(ctx *types.Context) error {
			ctx.LineOffset, _err = bldr.PrepareSketchBuildPath(ctx.Sketch, ctx.SourceOverride, ctx.SketchBuildPath)
			return _err
		}),
		&builder.ContainerFindIncludes{},
	}
	for _, command := range commands {
		err := command.Run(ctx)
		NoError(t, err)
	}

	importedLibraries := ctx.ImportedLibraries
	sort.Sort(ByLibraryName(importedLibraries))
	require.Equal(t, 1, len(importedLibraries))
	require.Equal(t, "SPI", importedLibraries[0].Name)
	requireEquivalentPaths(t, importedLibraries[0].SourceDir.String(), filepath.Join("user_hardware", "my_avr_platform", "avr", "libraries", "SPI"))
}

func TestIncludesToIncludeFoldersDuplicateLibsWithConflictingLibsOutsideOfPlatform(t *testing.T) {
	ctx := &types.Context{
		HardwareDirs:         paths.NewPathList(filepath.Join("..", "hardware"), "downloaded_hardware", "user_hardware"),
		BuiltInToolsDirs:     paths.NewPathList("downloaded_tools"),
		BuiltInLibrariesDirs: paths.New("downloaded_libraries"),
		OtherLibrariesDirs:   paths.NewPathList("libraries"),
		Verbose:              true,
	}
	ctx = prepareBuilderTestContext(t, ctx, paths.New("user_hardware", "my_avr_platform", "avr", "libraries", "SPI", "examples", "BarometricPressureSensor", "BarometricPressureSensor.ino"), "my_avr_platform:avr:custom_yun")
	defer cleanUpBuilderTestContext(t, ctx)

	var _err error
	commands := []types.Command{
		&builder.ContainerSetupHardwareToolsLibsSketchAndProps{},
		types.BareCommand(func(ctx *types.Context) error {
			ctx.LineOffset, _err = bldr.PrepareSketchBuildPath(ctx.Sketch, ctx.SourceOverride, ctx.SketchBuildPath)
			return _err
		}),
		&builder.ContainerFindIncludes{},
	}
	for _, command := range commands {
		err := command.Run(ctx)
		NoError(t, err)
	}

	importedLibraries := ctx.ImportedLibraries
	sort.Sort(ByLibraryName(importedLibraries))
	require.Equal(t, 1, len(importedLibraries))
	require.Equal(t, "SPI", importedLibraries[0].Name)
	requireEquivalentPaths(t, importedLibraries[0].SourceDir.String(), filepath.Join("libraries", "SPI"))
}
