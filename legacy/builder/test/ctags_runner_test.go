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
	"strings"
	"testing"

	"github.com/arduino/arduino-cli/legacy/builder"
	"github.com/arduino/arduino-cli/legacy/builder/types"
	paths "github.com/arduino/go-paths-helper"
	"github.com/stretchr/testify/require"
)

func TestCTagsRunner(t *testing.T) {
	DownloadCoresAndToolsAndLibraries(t)

	sketchLocation := Abs(t, paths.New("downloaded_libraries", "Bridge", "examples", "Bridge", "Bridge.ino"))

	ctx := &types.Context{
		HardwareDirs:         paths.NewPathList(filepath.Join("..", "hardware"), "hardware", "downloaded_hardware"),
		BuiltInToolsDirs:     paths.NewPathList("downloaded_tools"),
		BuiltInLibrariesDirs: paths.NewPathList("downloaded_libraries"),
		OtherLibrariesDirs:   paths.NewPathList("libraries"),
		SketchLocation:       sketchLocation,
		FQBN:                 parseFQBN(t, "arduino:avr:leonardo"),
		ArduinoAPIVersion:    "10600",
		Verbose:              true,
	}

	buildPath := SetupBuildPath(t, ctx)
	defer buildPath.RemoveAll()

	commands := []types.Command{

		&builder.ContainerSetupHardwareToolsLibsSketchAndProps{},

		&builder.ContainerMergeCopySketchFiles{},

		&builder.ContainerFindIncludes{},

		&builder.PrintUsedLibrariesIfVerbose{},
		&builder.WarnAboutArchIncompatibleLibraries{},
		&builder.CTagsTargetFileSaver{Source: &ctx.Source, TargetFileName: "ctags_target.cpp"},
		&builder.CTagsRunner{},
	}

	for _, command := range commands {
		err := command.Run(ctx)
		NoError(t, err)
	}

	quotedSketchLocation := strings.Replace(sketchLocation.String(), "\\", "\\\\", -1)
	expectedOutput := "server	" + quotedSketchLocation + "	/^BridgeServer server;$/;\"	kind:variable	line:31\n" +
		"setup	" + quotedSketchLocation + "	/^void setup() {$/;\"	kind:function	line:33	signature:()	returntype:void\n" +
		"loop	" + quotedSketchLocation + "	/^void loop() {$/;\"	kind:function	line:46	signature:()	returntype:void\n" +
		"process	" + quotedSketchLocation + "	/^void process(BridgeClient client) {$/;\"	kind:function	line:62	signature:(BridgeClient client)	returntype:void\n" +
		"digitalCommand	" + quotedSketchLocation + "	/^void digitalCommand(BridgeClient client) {$/;\"	kind:function	line:82	signature:(BridgeClient client)	returntype:void\n" +
		"analogCommand	" + quotedSketchLocation + "	/^void analogCommand(BridgeClient client) {$/;\"	kind:function	line:109	signature:(BridgeClient client)	returntype:void\n" +
		"modeCommand	" + quotedSketchLocation + "	/^void modeCommand(BridgeClient client) {$/;\"	kind:function	line:149	signature:(BridgeClient client)	returntype:void\n"

	require.Equal(t, expectedOutput, strings.Replace(ctx.CTagsOutput, "\r\n", "\n", -1))
}

func TestCTagsRunnerSketchWithClass(t *testing.T) {
	DownloadCoresAndToolsAndLibraries(t)

	sketchLocation := Abs(t, paths.New("sketch_with_class", "sketch.ino"))

	ctx := &types.Context{
		HardwareDirs:         paths.NewPathList(filepath.Join("..", "hardware"), "hardware", "downloaded_hardware"),
		BuiltInToolsDirs:     paths.NewPathList("downloaded_tools"),
		BuiltInLibrariesDirs: paths.NewPathList("downloaded_libraries"),
		OtherLibrariesDirs:   paths.NewPathList("libraries"),
		SketchLocation:       sketchLocation,
		FQBN:                 parseFQBN(t, "arduino:avr:leonardo"),
		ArduinoAPIVersion:    "10600",
		Verbose:              true,
	}

	buildPath := SetupBuildPath(t, ctx)
	defer buildPath.RemoveAll()

	commands := []types.Command{

		&builder.ContainerSetupHardwareToolsLibsSketchAndProps{},

		&builder.ContainerMergeCopySketchFiles{},

		&builder.ContainerFindIncludes{},

		&builder.PrintUsedLibrariesIfVerbose{},
		&builder.WarnAboutArchIncompatibleLibraries{},
		&builder.CTagsTargetFileSaver{Source: &ctx.Source, TargetFileName: "ctags_target.cpp"},
		&builder.CTagsRunner{},
	}

	for _, command := range commands {
		err := command.Run(ctx)
		NoError(t, err)
	}

	quotedSketchLocation := strings.Replace(sketchLocation.String(), "\\", "\\\\", -1)
	expectedOutput := "set_values\t" + quotedSketchLocation + "\t/^    void set_values (int,int);$/;\"\tkind:prototype\tline:4\tclass:Rectangle\tsignature:(int,int)\treturntype:void\n" +
		"area\t" + quotedSketchLocation + "\t/^    int area() {return width*height;}$/;\"\tkind:function\tline:5\tclass:Rectangle\tsignature:()\treturntype:int\n" +
		"set_values\t" + quotedSketchLocation + "\t/^void Rectangle::set_values (int x, int y) {$/;\"\tkind:function\tline:8\tclass:Rectangle\tsignature:(int x, int y)\treturntype:void\n" +
		"setup\t" + quotedSketchLocation + "\t/^void setup() {$/;\"\tkind:function\tline:13\tsignature:()\treturntype:void\n" +
		"loop\t" + quotedSketchLocation + "\t/^void loop() {$/;\"\tkind:function\tline:17\tsignature:()\treturntype:void\n"

	require.Equal(t, expectedOutput, strings.Replace(ctx.CTagsOutput, "\r\n", "\n", -1))
}

func TestCTagsRunnerSketchWithTypename(t *testing.T) {
	DownloadCoresAndToolsAndLibraries(t)

	sketchLocation := Abs(t, paths.New("sketch_with_typename", "sketch.ino"))

	ctx := &types.Context{
		HardwareDirs:         paths.NewPathList(filepath.Join("..", "hardware"), "hardware", "downloaded_hardware"),
		BuiltInToolsDirs:     paths.NewPathList("downloaded_tools"),
		BuiltInLibrariesDirs: paths.NewPathList("downloaded_libraries"),
		OtherLibrariesDirs:   paths.NewPathList("libraries"),
		SketchLocation:       sketchLocation,
		FQBN:                 parseFQBN(t, "arduino:avr:leonardo"),
		ArduinoAPIVersion:    "10600",
		Verbose:              true,
	}

	buildPath := SetupBuildPath(t, ctx)
	defer buildPath.RemoveAll()

	commands := []types.Command{

		&builder.ContainerSetupHardwareToolsLibsSketchAndProps{},

		&builder.ContainerMergeCopySketchFiles{},

		&builder.ContainerFindIncludes{},

		&builder.PrintUsedLibrariesIfVerbose{},
		&builder.WarnAboutArchIncompatibleLibraries{},
		&builder.CTagsTargetFileSaver{Source: &ctx.Source, TargetFileName: "ctags_target.cpp"},
		&builder.CTagsRunner{},
	}

	for _, command := range commands {
		err := command.Run(ctx)
		NoError(t, err)
	}

	quotedSketchLocation := strings.Replace(sketchLocation.String(), "\\", "\\\\", -1)
	expectedOutput := "Foo\t" + quotedSketchLocation + "\t/^  struct Foo{$/;\"\tkind:struct\tline:2\n" +
		"setup\t" + quotedSketchLocation + "\t/^void setup() {$/;\"\tkind:function\tline:6\tsignature:()\treturntype:void\n" +
		"loop\t" + quotedSketchLocation + "\t/^void loop() {}$/;\"\tkind:function\tline:10\tsignature:()\treturntype:void\n" +
		"func\t" + quotedSketchLocation + "\t/^typename Foo<char>::Bar func(){$/;\"\tkind:function\tline:12\tsignature:()\treturntype:Foo::Bar\n"

	require.Equal(t, expectedOutput, strings.Replace(ctx.CTagsOutput, "\r\n", "\n", -1))
}

func TestCTagsRunnerSketchWithNamespace(t *testing.T) {
	DownloadCoresAndToolsAndLibraries(t)

	sketchLocation := Abs(t, paths.New("sketch_with_namespace", "sketch.ino"))

	ctx := &types.Context{
		HardwareDirs:         paths.NewPathList(filepath.Join("..", "hardware"), "hardware", "downloaded_hardware"),
		BuiltInToolsDirs:     paths.NewPathList("downloaded_tools"),
		BuiltInLibrariesDirs: paths.NewPathList("downloaded_libraries"),
		OtherLibrariesDirs:   paths.NewPathList("libraries"),
		SketchLocation:       sketchLocation,
		FQBN:                 parseFQBN(t, "arduino:avr:leonardo"),
		ArduinoAPIVersion:    "10600",
		Verbose:              true,
	}

	buildPath := SetupBuildPath(t, ctx)
	defer buildPath.RemoveAll()

	commands := []types.Command{

		&builder.ContainerSetupHardwareToolsLibsSketchAndProps{},

		&builder.ContainerMergeCopySketchFiles{},

		&builder.ContainerFindIncludes{},

		&builder.PrintUsedLibrariesIfVerbose{},
		&builder.WarnAboutArchIncompatibleLibraries{},
		&builder.CTagsTargetFileSaver{Source: &ctx.Source, TargetFileName: "ctags_target.cpp"},
		&builder.CTagsRunner{},
	}

	for _, command := range commands {
		err := command.Run(ctx)
		NoError(t, err)
	}

	quotedSketchLocation := strings.Replace(sketchLocation.String(), "\\", "\\\\", -1)
	expectedOutput := "value\t" + quotedSketchLocation + "\t/^\tint value() {$/;\"\tkind:function\tline:2\tnamespace:Test\tsignature:()\treturntype:int\n" +
		"setup\t" + quotedSketchLocation + "\t/^void setup() {}$/;\"\tkind:function\tline:7\tsignature:()\treturntype:void\n" +
		"loop\t" + quotedSketchLocation + "\t/^void loop() {}$/;\"\tkind:function\tline:8\tsignature:()\treturntype:void\n"

	require.Equal(t, expectedOutput, strings.Replace(ctx.CTagsOutput, "\r\n", "\n", -1))
}

func TestCTagsRunnerSketchWithTemplates(t *testing.T) {
	DownloadCoresAndToolsAndLibraries(t)

	sketchLocation := Abs(t, paths.New("sketch_with_templates_and_shift", "template_and_shift.cpp"))

	ctx := &types.Context{
		HardwareDirs:         paths.NewPathList(filepath.Join("..", "hardware"), "hardware", "downloaded_hardware"),
		BuiltInToolsDirs:     paths.NewPathList("downloaded_tools"),
		BuiltInLibrariesDirs: paths.NewPathList("downloaded_libraries"),
		OtherLibrariesDirs:   paths.NewPathList("libraries"),
		SketchLocation:       sketchLocation,
		FQBN:                 parseFQBN(t, "arduino:avr:leonardo"),
		ArduinoAPIVersion:    "10600",
		Verbose:              true,
	}

	buildPath := SetupBuildPath(t, ctx)
	defer buildPath.RemoveAll()

	commands := []types.Command{

		&builder.ContainerSetupHardwareToolsLibsSketchAndProps{},

		&builder.ContainerMergeCopySketchFiles{},

		&builder.ContainerFindIncludes{},

		&builder.PrintUsedLibrariesIfVerbose{},
		&builder.WarnAboutArchIncompatibleLibraries{},
		&builder.CTagsTargetFileSaver{Source: &ctx.Source, TargetFileName: "ctags_target.cpp"},
		&builder.CTagsRunner{},
	}

	for _, command := range commands {
		err := command.Run(ctx)
		NoError(t, err)
	}

	quotedSketchLocation := strings.Replace(sketchLocation.String(), "\\", "\\\\", -1)
	expectedOutput := "printGyro\t" + quotedSketchLocation + "\t/^void printGyro()$/;\"\tkind:function\tline:10\tsignature:()\treturntype:void\n" +
		"bVar\t" + quotedSketchLocation + "\t/^c< 8 > bVar;$/;\"\tkind:variable\tline:15\n" +
		"aVar\t" + quotedSketchLocation + "\t/^c< 1<<8 > aVar;$/;\"\tkind:variable\tline:16\n" +
		"func\t" + quotedSketchLocation + "\t/^template<int X> func( c< 1<<X> & aParam) {$/;\"\tkind:function\tline:18\tsignature:( c< 1<<X> & aParam)\treturntype:template\n"

	require.Equal(t, expectedOutput, strings.Replace(ctx.CTagsOutput, "\r\n", "\n", -1))
}
