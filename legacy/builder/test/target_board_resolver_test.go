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

	"github.com/arduino/arduino-cli/legacy/builder"
	"github.com/arduino/arduino-cli/legacy/builder/types"
	paths "github.com/arduino/go-paths-helper"
	"github.com/stretchr/testify/require"
)

func TestTargetBoardResolverUno(t *testing.T) {
	ctx := &types.Context{
		HardwareDirs: paths.NewPathList(filepath.Join("..", "hardware"), "hardware", "downloaded_hardware"),
		FQBN:         parseFQBN(t, "arduino:avr:uno"),
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
	require.Equal(t, "arduino", targetPackage.Name)
	targetPlatform := ctx.TargetPlatform
	require.Equal(t, "avr", targetPlatform.Platform.Architecture)
	targetBoard := ctx.TargetBoard
	require.Equal(t, "uno", targetBoard.BoardID)
	require.Equal(t, "atmega328p", targetBoard.Properties.Get("build.mcu"))
}

func TestTargetBoardResolverDue(t *testing.T) {
	ctx := &types.Context{
		HardwareDirs: paths.NewPathList(filepath.Join("..", "hardware"), "hardware", "downloaded_hardware"),
		FQBN:         parseFQBN(t, "arduino:sam:arduino_due_x"),
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
	require.Equal(t, "arduino", targetPackage.Name)
	targetPlatform := ctx.TargetPlatform
	require.Equal(t, "sam", targetPlatform.Platform.Architecture)
	targetBoard := ctx.TargetBoard
	require.Equal(t, "arduino_due_x", targetBoard.BoardID)
	require.Equal(t, "cortex-m3", targetBoard.Properties.Get("build.mcu"))
}

func TestTargetBoardResolverMega1280(t *testing.T) {
	ctx := &types.Context{
		HardwareDirs: paths.NewPathList(filepath.Join("..", "hardware"), "hardware", "downloaded_hardware"),
		FQBN:         parseFQBN(t, "arduino:avr:mega:cpu=atmega1280"),
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
	require.Equal(t, "arduino", targetPackage.Name)
	targetPlatform := ctx.TargetPlatform
	require.Equal(t, "avr", targetPlatform.Platform.Architecture)
	targetBoard := ctx.TargetBoard
	require.Equal(t, "mega", targetBoard.BoardID)
	require.Equal(t, "atmega1280", targetBoard.Properties.Get("build.mcu"))
	require.Equal(t, "AVR_MEGA", targetBoard.Properties.Get("build.board"))
}

func TestTargetBoardResolverMega2560(t *testing.T) {
	ctx := &types.Context{
		HardwareDirs: paths.NewPathList(filepath.Join("..", "hardware"), "hardware", "downloaded_hardware"),
		FQBN:         parseFQBN(t, "arduino:avr:mega:cpu=atmega2560"),
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
	require.Equal(t, "arduino", targetPackage.Name)
	targetPlatform := ctx.TargetPlatform
	require.Equal(t, "avr", targetPlatform.Platform.Architecture)
	targetBoard := ctx.TargetBoard
	require.Equal(t, "mega", targetBoard.BoardID)
	require.Equal(t, "atmega2560", targetBoard.Properties.Get("build.mcu"))
	require.Equal(t, "AVR_MEGA2560", targetBoard.Properties.Get("build.board"))
}

func TestTargetBoardResolverCustomYun(t *testing.T) {
	ctx := &types.Context{
		HardwareDirs: paths.NewPathList(filepath.Join("..", "hardware"), "hardware", "downloaded_hardware", "user_hardware"),
		FQBN:         parseFQBN(t, "my_avr_platform:avr:custom_yun"),
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
	require.Equal(t, "custom_yun", targetBoard.BoardID)
	require.Equal(t, "atmega32u4", targetBoard.Properties.Get("build.mcu"))
	require.Equal(t, "AVR_YUN", targetBoard.Properties.Get("build.board"))
}

func TestTargetBoardResolverCustomCore(t *testing.T) {
	ctx := &types.Context{
		HardwareDirs: paths.NewPathList(filepath.Join("..", "hardware"), "hardware", "downloaded_hardware", "user_hardware"),
		FQBN:         parseFQBN(t, "watterott:avr:attiny841:core=spencekonde,info=info"),
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
	require.Equal(t, "watterott", targetPackage.Name)
	targetPlatform := ctx.TargetPlatform
	require.Equal(t, "avr", targetPlatform.Platform.Architecture)
	targetBoard := ctx.TargetBoard
	require.Equal(t, "attiny841", targetBoard.BoardID)
	require.Equal(t, "tiny841", ctx.BuildCore)
	require.Equal(t, "tiny14", targetBoard.Properties.Get("build.variant"))
}
