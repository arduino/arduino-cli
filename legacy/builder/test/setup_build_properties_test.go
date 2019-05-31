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

func TestSetupBuildProperties(t *testing.T) {
	DownloadCoresAndToolsAndLibraries(t)

	ctx := &types.Context{
		HardwareDirs:      paths.NewPathList(filepath.Join("..", "hardware"), "hardware", "downloaded_hardware", "user_hardware"),
		BuiltInToolsDirs:  paths.NewPathList("downloaded_tools", "tools_builtin"),
		SketchLocation:    paths.New("sketch1", "sketch.ino"),
		FQBN:              parseFQBN(t, "arduino:avr:uno"),
		ArduinoAPIVersion: "10600",
	}

	buildPath := SetupBuildPath(t, ctx)
	defer buildPath.RemoveAll()

	commands := []types.Command{
		&builder.AddAdditionalEntriesToContext{},
		&builder.HardwareLoader{},
		&builder.TargetBoardResolver{},
		&builder.ToolsLoader{},
		&builder.SketchLoader{},
		&builder.SetupBuildProperties{},
	}

	for _, command := range commands {
		err := command.Run(ctx)
		NoError(t, err)
	}

	buildProperties := ctx.BuildProperties

	require.Equal(t, "ARDUINO", buildProperties.Get("software"))

	require.Equal(t, "uno", buildProperties.Get("_id"))
	require.Equal(t, "Arduino/Genuino Uno", buildProperties.Get("name"))
	require.Equal(t, "0x2341", buildProperties.Get("vid.0"))
	require.Equal(t, "\"{compiler.path}{compiler.c.cmd}\" {compiler.c.flags} -mmcu={build.mcu} -DF_CPU={build.f_cpu} -DARDUINO={runtime.ide.version} -DARDUINO_{build.board} -DARDUINO_ARCH_{build.arch} {compiler.c.extra_flags} {build.extra_flags} {includes} \"{source_file}\" -o \"{object_file}\"", buildProperties.Get("recipe.c.o.pattern"))
	require.Equal(t, "{path}/etc/avrdude.conf", buildProperties.Get("tools.avrdude.config.path"))

	requireEquivalentPaths(t, buildProperties.Get("runtime.platform.path"), "downloaded_hardware/arduino/avr")
	requireEquivalentPaths(t, buildProperties.Get("runtime.hardware.path"), "downloaded_hardware/arduino")
	require.Equal(t, "10600", buildProperties.Get("runtime.ide.version"))
	require.NotEmpty(t, buildProperties.Get("runtime.os"))

	requireEquivalentPaths(t, buildProperties.Get("runtime.tools.arm-none-eabi-gcc.path"), "downloaded_tools/arm-none-eabi-gcc/4.8.3-2014q1")
	requireEquivalentPaths(t, buildProperties.Get("runtime.tools.arm-none-eabi-gcc-4.8.3-2014q1.path"), "downloaded_tools/arm-none-eabi-gcc/4.8.3-2014q1")

	requireEquivalentPaths(t, buildProperties.Get("runtime.tools.avrdude-6.0.1-arduino5.path"), "tools_builtin/avr", "downloaded_tools/avrdude/6.0.1-arduino5")
	requireEquivalentPaths(t, buildProperties.Get("runtime.tools.avrdude.path"), "tools_builtin/avr", "downloaded_tools/avrdude/6.0.1-arduino5")

	bossacPath := buildProperties.Get("runtime.tools.bossac.path")
	bossac161Path := buildProperties.Get("runtime.tools.bossac-1.6.1-arduino.path")
	bossac15Path := buildProperties.Get("runtime.tools.bossac-1.5-arduino.path")
	requireEquivalentPaths(t, bossac161Path, "downloaded_tools/bossac/1.6.1-arduino")
	requireEquivalentPaths(t, bossac15Path, "downloaded_tools/bossac/1.5-arduino")
	requireEquivalentPaths(t, bossacPath, bossac161Path, bossac15Path)

	requireEquivalentPaths(t, buildProperties.Get("runtime.tools.avr-gcc.path"), "downloaded_tools/avr-gcc/4.8.1-arduino5", "tools_builtin/avr")

	requireEquivalentPaths(t, buildProperties.Get("build.source.path"), "sketch1")

	require.True(t, buildProperties.ContainsKey("extra.time.utc"))
	require.True(t, buildProperties.ContainsKey("extra.time.local"))
	require.True(t, buildProperties.ContainsKey("extra.time.zone"))
	require.True(t, buildProperties.ContainsKey("extra.time.dst"))
}

func TestSetupBuildPropertiesWithSomeCustomOverrides(t *testing.T) {
	DownloadCoresAndToolsAndLibraries(t)

	ctx := &types.Context{
		HardwareDirs:      paths.NewPathList(filepath.Join("..", "hardware"), "hardware", "downloaded_hardware"),
		BuiltInToolsDirs:  paths.NewPathList("downloaded_tools", "tools_builtin"),
		SketchLocation:    paths.New("sketch1", "sketch.ino"),
		FQBN:              parseFQBN(t, "arduino:avr:uno"),
		ArduinoAPIVersion: "10600",

		CustomBuildProperties: []string{"name=fake name", "tools.avrdude.config.path=non existent path with space and a ="},
	}

	buildPath := SetupBuildPath(t, ctx)
	defer buildPath.RemoveAll()

	commands := []types.Command{
		&builder.AddAdditionalEntriesToContext{},
		&builder.HardwareLoader{},
		&builder.TargetBoardResolver{},
		&builder.ToolsLoader{},
		&builder.SketchLoader{},
		&builder.SetupBuildProperties{},
		&builder.SetCustomBuildProperties{},
	}

	for _, command := range commands {
		err := command.Run(ctx)
		NoError(t, err)
	}

	buildProperties := ctx.BuildProperties

	require.Equal(t, "ARDUINO", buildProperties.Get("software"))

	require.Equal(t, "uno", buildProperties.Get("_id"))
	require.Equal(t, "fake name", buildProperties.Get("name"))
	require.Equal(t, "\"{compiler.path}{compiler.c.cmd}\" {compiler.c.flags} -mmcu={build.mcu} -DF_CPU={build.f_cpu} -DARDUINO={runtime.ide.version} -DARDUINO_{build.board} -DARDUINO_ARCH_{build.arch} {compiler.c.extra_flags} {build.extra_flags} {includes} \"{source_file}\" -o \"{object_file}\"", buildProperties.Get("recipe.c.o.pattern"))
	require.Equal(t, "non existent path with space and a =", buildProperties.Get("tools.avrdude.config.path"))
}

func TestSetupBuildPropertiesUserHardware(t *testing.T) {
	DownloadCoresAndToolsAndLibraries(t)

	ctx := &types.Context{
		HardwareDirs:      paths.NewPathList(filepath.Join("..", "hardware"), "hardware", "downloaded_hardware", "user_hardware"),
		BuiltInToolsDirs:  paths.NewPathList("downloaded_tools", "tools_builtin"),
		SketchLocation:    paths.New("sketch1", "sketch.ino"),
		FQBN:              parseFQBN(t, "my_avr_platform:avr:custom_yun"),
		ArduinoAPIVersion: "10600",
	}

	buildPath := SetupBuildPath(t, ctx)
	defer buildPath.RemoveAll()

	commands := []types.Command{
		&builder.AddAdditionalEntriesToContext{},
		&builder.HardwareLoader{},
		&builder.TargetBoardResolver{},
		&builder.ToolsLoader{},
		&builder.SketchLoader{},
		&builder.SetupBuildProperties{},
	}

	for _, command := range commands {
		err := command.Run(ctx)
		NoError(t, err)
	}

	buildProperties := ctx.BuildProperties

	require.Equal(t, "ARDUINO", buildProperties.Get("software"))

	require.Equal(t, "custom_yun", buildProperties.Get("_id"))
	require.Equal(t, "caterina/Caterina-custom_yun.hex", buildProperties.Get("bootloader.file"))
	requireEquivalentPaths(t, buildProperties.Get("runtime.platform.path"), filepath.Join("user_hardware", "my_avr_platform", "avr"))
	requireEquivalentPaths(t, buildProperties.Get("runtime.hardware.path"), filepath.Join("user_hardware", "my_avr_platform"))
}

func TestSetupBuildPropertiesWithMissingPropsFromParentPlatformTxtFiles(t *testing.T) {
	DownloadCoresAndToolsAndLibraries(t)

	ctx := &types.Context{
		HardwareDirs:      paths.NewPathList(filepath.Join("..", "hardware"), "hardware", "downloaded_hardware", "user_hardware"),
		BuiltInToolsDirs:  paths.NewPathList("downloaded_tools", "tools_builtin"),
		SketchLocation:    paths.New("sketch1", "sketch.ino"),
		FQBN:              parseFQBN(t, "my_avr_platform:avr:custom_yun"),
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

	require.Equal(t, "ARDUINO", buildProperties.Get("software"))

	require.Equal(t, "custom_yun", buildProperties.Get("_id"))
	require.Equal(t, "Arduino Yún", buildProperties.Get("name"))
	require.Equal(t, "0x2341", buildProperties.Get("vid.0"))
	require.Equal(t, "\"{compiler.path}{compiler.c.cmd}\" {compiler.c.flags} -mmcu={build.mcu} -DF_CPU={build.f_cpu} -DARDUINO={runtime.ide.version} -DARDUINO_{build.board} -DARDUINO_ARCH_{build.arch} {compiler.c.extra_flags} {build.extra_flags} {includes} \"{source_file}\" -o \"{object_file}\"", buildProperties.Get("recipe.c.o.pattern"))
	require.Equal(t, "{path}/etc/avrdude.conf", buildProperties.Get("tools.avrdude.config.path"))

	requireEquivalentPaths(t, buildProperties.Get("runtime.platform.path"), "user_hardware/my_avr_platform/avr")
	requireEquivalentPaths(t, buildProperties.Get("runtime.hardware.path"), "user_hardware/my_avr_platform")
	require.Equal(t, "10600", buildProperties.Get("runtime.ide.version"))
	require.NotEmpty(t, buildProperties.Get("runtime.os"))

	requireEquivalentPaths(t, buildProperties.Get("runtime.tools.arm-none-eabi-gcc.path"), "downloaded_tools/arm-none-eabi-gcc/4.8.3-2014q1")
	requireEquivalentPaths(t, buildProperties.Get("runtime.tools.arm-none-eabi-gcc-4.8.3-2014q1.path"), "downloaded_tools/arm-none-eabi-gcc/4.8.3-2014q1")
	requireEquivalentPaths(t, buildProperties.Get("runtime.tools.bossac-1.6.1-arduino.path"), "downloaded_tools/bossac/1.6.1-arduino")
	requireEquivalentPaths(t, buildProperties.Get("runtime.tools.bossac-1.5-arduino.path"), "downloaded_tools/bossac/1.5-arduino")

	requireEquivalentPaths(t, buildProperties.Get("runtime.tools.bossac.path"), "downloaded_tools/bossac/1.6.1-arduino", "downloaded_tools/bossac/1.5-arduino")
	requireEquivalentPaths(t, buildProperties.Get("runtime.tools.avrdude.path"), "downloaded_tools/avrdude/6.0.1-arduino5", "tools_builtin/avr")

	requireEquivalentPaths(t, buildProperties.Get("runtime.tools.avrdude-6.0.1-arduino5.path"), "downloaded_tools/avrdude/6.0.1-arduino5", "tools_builtin/avr")

	requireEquivalentPaths(t, buildProperties.Get("runtime.tools.avr-gcc.path"), "downloaded_tools/avr-gcc/4.8.1-arduino5", "tools_builtin/avr")
	requireEquivalentPaths(t, buildProperties.Get("runtime.tools.avr-gcc-4.8.1-arduino5.path"), "downloaded_tools/avr-gcc/4.8.1-arduino5", "tools_builtin/avr")

	requireEquivalentPaths(t, buildProperties.Get("build.source.path"), "sketch1")

	require.True(t, buildProperties.ContainsKey("extra.time.utc"))
	require.True(t, buildProperties.ContainsKey("extra.time.local"))
	require.True(t, buildProperties.ContainsKey("extra.time.zone"))
	require.True(t, buildProperties.ContainsKey("extra.time.dst"))
}
