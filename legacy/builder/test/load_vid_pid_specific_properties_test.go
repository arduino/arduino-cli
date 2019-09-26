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

func TestLoadVIDPIDSpecificPropertiesWhenNoVIDPIDAreProvided(t *testing.T) {
	DownloadCoresAndToolsAndLibraries(t)

	ctx := &types.Context{
		HardwareDirs:      paths.NewPathList(filepath.Join("..", "hardware"), "hardware", "downloaded_hardware"),
		BuiltInToolsDirs:  paths.NewPathList("downloaded_tools", "./tools_builtin"),
		SketchLocation:    paths.New("sketch1", "sketch.ino"),
		FQBN:              parseFQBN(t, "arduino:avr:micro"),
		ArduinoAPIVersion: "10600",
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

	require.Equal(t, "0x0037", buildProperties.Get("pid.0"))
	require.Equal(t, "\"Genuino Micro\"", buildProperties.Get("vid.4.build.usb_product"))
	require.Equal(t, "0x8037", buildProperties.Get("build.pid"))
}

func TestLoadVIDPIDSpecificProperties(t *testing.T) {
	DownloadCoresAndToolsAndLibraries(t)

	ctx := &types.Context{
		HardwareDirs:      paths.NewPathList(filepath.Join("..", "hardware"), "hardware", "downloaded_hardware"),
		BuiltInToolsDirs:  paths.NewPathList("downloaded_tools", "./tools_builtin"),
		SketchLocation:    paths.New("sketch1", "sketch.ino"),
		FQBN:              parseFQBN(t, "arduino:avr:micro"),
		ArduinoAPIVersion: "10600",
	}

	buildPath := SetupBuildPath(t, ctx)
	defer buildPath.RemoveAll()

	ctx.USBVidPid = "0x2341_0x0237"

	commands := []types.Command{
		&builder.ContainerSetupHardwareToolsLibsSketchAndProps{},
	}

	for _, command := range commands {
		err := command.Run(ctx)
		NoError(t, err)
	}

	buildProperties := ctx.BuildProperties

	require.Equal(t, "0x0037", buildProperties.Get("pid.0"))
	require.Equal(t, "\"Genuino Micro\"", buildProperties.Get("vid.4.build.usb_product"))
	require.Equal(t, "0x2341", buildProperties.Get("build.vid"))
	require.Equal(t, "0x8237", buildProperties.Get("build.pid"))
}
