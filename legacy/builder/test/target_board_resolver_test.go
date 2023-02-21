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

	"github.com/arduino/arduino-cli/legacy/builder"
	"github.com/arduino/arduino-cli/legacy/builder/types"
	paths "github.com/arduino/go-paths-helper"
	"github.com/stretchr/testify/require"
)

func TestTargetBoardResolverUno(t *testing.T) {
	ctx := &types.Context{
		HardwareDirs: paths.NewPathList(filepath.Join("..", "hardware"), "downloaded_hardware"),
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
	targetBoardBuildProperties := ctx.TargetBoardBuildProperties
	require.Equal(t, "atmega328p", targetBoardBuildProperties.Get("build.mcu"))
}

func TestTargetBoardResolverDue(t *testing.T) {
	ctx := &types.Context{
		HardwareDirs: paths.NewPathList(filepath.Join("..", "hardware"), "downloaded_hardware"),
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
	targetBoardBuildProperties := ctx.TargetBoardBuildProperties
	require.Equal(t, "cortex-m3", targetBoardBuildProperties.Get("build.mcu"))
}

func TestTargetBoardResolverMega1280(t *testing.T) {
	ctx := &types.Context{
		HardwareDirs: paths.NewPathList(filepath.Join("..", "hardware"), "downloaded_hardware"),
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
	targetBoardBuildProperties := ctx.TargetBoardBuildProperties
	require.Equal(t, "atmega1280", targetBoardBuildProperties.Get("build.mcu"))
	require.Equal(t, "AVR_MEGA", targetBoardBuildProperties.Get("build.board"))
}

func TestTargetBoardResolverMega2560(t *testing.T) {
	ctx := &types.Context{
		HardwareDirs: paths.NewPathList(filepath.Join("..", "hardware"), "downloaded_hardware"),
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
	targetBoardBuildProperties := ctx.TargetBoardBuildProperties
	require.Equal(t, "atmega2560", targetBoardBuildProperties.Get("build.mcu"))
	require.Equal(t, "AVR_MEGA2560", targetBoardBuildProperties.Get("build.board"))
}

func TestTargetBoardResolverCustomYun(t *testing.T) {
	ctx := &types.Context{
		HardwareDirs: paths.NewPathList(filepath.Join("..", "hardware"), "downloaded_hardware", "user_hardware"),
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
	targetBoardBuildProperties := ctx.TargetBoardBuildProperties
	require.Equal(t, "atmega32u4", targetBoardBuildProperties.Get("build.mcu"))
	require.Equal(t, "AVR_YUN", targetBoardBuildProperties.Get("build.board"))
}

func TestTargetBoardResolverCustomCore(t *testing.T) {
	ctx := &types.Context{
		HardwareDirs: paths.NewPathList("hardware"),
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
	require.Equal(t, "tiny841", ctx.TargetBoardBuildProperties.Get("build.core"))
	targetBoardBuildProperties := ctx.TargetBoardBuildProperties
	require.Equal(t, "tiny14", targetBoardBuildProperties.Get("build.variant"))
}
