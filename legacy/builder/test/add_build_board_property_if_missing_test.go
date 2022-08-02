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
	"testing"

	"github.com/arduino/arduino-cli/arduino/cores"
	"github.com/arduino/arduino-cli/legacy/builder"
	"github.com/arduino/arduino-cli/legacy/builder/types"
	"github.com/arduino/go-paths-helper"
	"github.com/stretchr/testify/require"
)

func parseFQBN(t *testing.T, fqbnIn string) *cores.FQBN {
	fqbn, err := cores.ParseFQBN(fqbnIn)
	require.NoError(t, err)
	return fqbn
}

func TestAddBuildBoardPropertyIfMissing(t *testing.T) {
	DownloadCoresAndToolsAndLibraries(t)

	ctx := &types.Context{
		HardwareDirs: paths.NewPathList(filepath.Join("..", "hardware"), "downloaded_hardware", "user_hardware"),
		FQBN:         parseFQBN(t, "my_avr_platform:avr:mymega"),
		Verbose:      true,
	}

	commands := []types.Command{
		&builder.HardwareLoader{},
		&builder.TargetBoardResolver{},
	}

	for _, command := range commands {
		err := command.Run(ctx)
		NoError(t, err)
	}

	targetPackage := ctx.TargetPackage
	require.Equal(t, "my_avr_platform", targetPackage.Name)
	targetPlatform := ctx.TargetPlatform
	require.NotNil(t, targetPlatform)
	require.NotNil(t, targetPlatform.Platform)
	require.Equal(t, "avr", targetPlatform.Platform.Architecture)
	targetBoard := ctx.TargetBoard
	require.Equal(t, "mymega", targetBoard.BoardID)
	targetBoardBuildProperties := ctx.TargetBoardBuildProperties
	require.Equal(t, "atmega2560", targetBoardBuildProperties.Get("build.mcu"))
	require.Equal(t, "AVR_MYMEGA", targetBoardBuildProperties.Get("build.board"))
}

func TestAddBuildBoardPropertyIfMissingNotMissing(t *testing.T) {
	DownloadCoresAndToolsAndLibraries(t)

	ctx := &types.Context{
		HardwareDirs: paths.NewPathList(filepath.Join("..", "hardware"), "downloaded_hardware", "user_hardware"),
		FQBN:         parseFQBN(t, "my_avr_platform:avr:mymega:cpu=atmega1280"),
		Verbose:      true,
	}

	commands := []types.Command{
		&builder.HardwareLoader{},
		&builder.TargetBoardResolver{},
	}

	for _, command := range commands {
		err := command.Run(ctx)
		NoError(t, err)
	}

	targetPackage := ctx.TargetPackage
	require.Equal(t, "my_avr_platform", targetPackage.Name)
	targetPlatform := ctx.TargetPlatform
	require.Equal(t, "avr", targetPlatform.Platform.Architecture)
	targetBoard := ctx.TargetBoard
	require.Equal(t, "mymega", targetBoard.BoardID)
	targetBoardBuildProperties := ctx.TargetBoardBuildProperties
	require.Equal(t, "atmega1280", targetBoardBuildProperties.Get("build.mcu"))
	require.Equal(t, "MYMEGA1280", targetBoardBuildProperties.Get("build.board"))
}
