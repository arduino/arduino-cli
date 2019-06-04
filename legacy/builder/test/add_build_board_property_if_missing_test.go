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
		HardwareDirs: paths.NewPathList(filepath.Join("..", "hardware"), "hardware", "downloaded_hardware", "user_hardware"),
		FQBN:         parseFQBN(t, "my_avr_platform:avr:mymega"),
		Verbose:      true,
	}

	commands := []types.Command{
		&builder.HardwareLoader{},
		&builder.TargetBoardResolver{},
		&builder.AddBuildBoardPropertyIfMissing{},
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
	require.Equal(t, "atmega2560", targetBoard.Properties.Get("build.mcu"))
	require.Equal(t, "AVR_MYMEGA2560", targetBoard.Properties.Get("build.board"))
}

func TestAddBuildBoardPropertyIfMissingNotMissing(t *testing.T) {
	DownloadCoresAndToolsAndLibraries(t)

	ctx := &types.Context{
		HardwareDirs: paths.NewPathList(filepath.Join("..", "hardware"), "hardware", "downloaded_hardware", "user_hardware"),
		FQBN:         parseFQBN(t, "my_avr_platform:avr:mymega:cpu=atmega1280"),
		Verbose:      true,
	}

	commands := []types.Command{
		&builder.HardwareLoader{},
		&builder.TargetBoardResolver{},
		&builder.AddBuildBoardPropertyIfMissing{},
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
	require.Equal(t, "atmega1280", targetBoard.Properties.Get("build.mcu"))
	require.Equal(t, "AVR_MYMEGA", targetBoard.Properties.Get("build.board"))
}
