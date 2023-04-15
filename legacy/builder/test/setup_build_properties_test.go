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

func TestSetupBuildProperties(t *testing.T) {
	ctx := &types.Context{
		HardwareDirs:     paths.NewPathList(filepath.Join("..", "hardware"), "downloaded_hardware", "user_hardware"),
		BuiltInToolsDirs: paths.NewPathList("downloaded_tools", "tools_builtin"),
	}
	ctx = prepareBuilderTestContext(t, ctx, paths.New("sketch1", "sketch1.ino"), "arduino:avr:uno")
	buildPath := SetupBuildPath(t, ctx)
	defer buildPath.RemoveAll()

	commands := []types.Command{
		&builder.AddAdditionalEntriesToContext{},
		&builder.HardwareLoader{},
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
	require.Equal(t, "10607", buildProperties.Get("runtime.ide.version"))
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
	ctx := &types.Context{
		HardwareDirs:          paths.NewPathList(filepath.Join("..", "hardware"), "downloaded_hardware"),
		BuiltInToolsDirs:      paths.NewPathList("downloaded_tools", "tools_builtin"),
		CustomBuildProperties: []string{"name=fake name", "tools.avrdude.config.path=non existent path with space and a ="},
	}
	ctx = prepareBuilderTestContext(t, ctx, paths.New("sketch1", "sketch1.ino"), "arduino:avr:uno")

	buildPath := SetupBuildPath(t, ctx)
	defer buildPath.RemoveAll()

	commands := []types.Command{
		&builder.AddAdditionalEntriesToContext{},
		&builder.HardwareLoader{},
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
	ctx := &types.Context{
		HardwareDirs:     paths.NewPathList(filepath.Join("..", "hardware"), "downloaded_hardware", "user_hardware"),
		BuiltInToolsDirs: paths.NewPathList("downloaded_tools", "tools_builtin"),
	}
	ctx = prepareBuilderTestContext(t, ctx, paths.New("sketch1", "sketch1.ino"), "my_avr_platform:avr:custom_yun")
	buildPath := SetupBuildPath(t, ctx)
	defer buildPath.RemoveAll()

	commands := []types.Command{
		&builder.AddAdditionalEntriesToContext{},
		&builder.HardwareLoader{},
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
	ctx := &types.Context{
		HardwareDirs:     paths.NewPathList(filepath.Join("..", "hardware"), "downloaded_hardware", "user_hardware"),
		BuiltInToolsDirs: paths.NewPathList("downloaded_tools", "tools_builtin"),
	}
	ctx = prepareBuilderTestContext(t, ctx, paths.New("sketch1", "sketch1.ino"), "my_avr_platform:avr:custom_yun")
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
	require.Equal(t, "10607", buildProperties.Get("runtime.ide.version"))
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
