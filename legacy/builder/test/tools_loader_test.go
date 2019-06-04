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
	"sort"
	"testing"

	"github.com/arduino/arduino-cli/arduino/cores"
	"github.com/arduino/arduino-cli/legacy/builder"
	"github.com/arduino/arduino-cli/legacy/builder/types"
	paths "github.com/arduino/go-paths-helper"
	"github.com/stretchr/testify/require"
)

type ByToolIDAndVersion []*cores.ToolRelease

func (s ByToolIDAndVersion) Len() int {
	return len(s)
}
func (s ByToolIDAndVersion) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s ByToolIDAndVersion) Less(i, j int) bool {
	if s[i].Tool.Name != s[j].Tool.Name {
		return s[i].Tool.Name < s[j].Tool.Name
	}
	if !s[i].Version.Equal(s[j].Version) {
		return s[i].Version.LessThan(s[j].Version)
	}
	return s[i].InstallDir.String() < s[j].InstallDir.String()
}

func requireEquivalentPaths(t *testing.T, actual string, expected ...string) {
	if len(expected) == 1 {
		actualAbs, err := paths.New(actual).Abs()
		require.NoError(t, err)
		expectedAbs, err := paths.New(expected[0]).Abs()
		require.NoError(t, err)
		require.Equal(t, expectedAbs.String(), actualAbs.String())
	} else {
		actualAbs, err := paths.New(actual).Abs()
		require.NoError(t, err)
		expectedAbs := paths.NewPathList(expected...)
		require.NoError(t, expectedAbs.ToAbs())
		require.Contains(t, expectedAbs.AsStrings(), actualAbs.String())
	}
}

func TestLoadTools(t *testing.T) {
	DownloadCoresAndToolsAndLibraries(t)

	ctx := &types.Context{
		BuiltInToolsDirs: paths.NewPathList("downloaded_tools", "tools_builtin"),
	}

	NoError(t, (&builder.HardwareLoader{}).Run(ctx))
	NoError(t, (&builder.ToolsLoader{}).Run(ctx))

	tools := ctx.AllTools
	require.Equal(t, 9, len(tools))

	sort.Sort(ByToolIDAndVersion(tools))

	idx := 0
	require.Equal(t, ":arduino-preprocessor@0.1.5", tools[idx].String())
	requireEquivalentPaths(t, tools[idx].InstallDir.String(), "downloaded_tools/arduino-preprocessor/0.1.5")
	idx++
	require.Equal(t, ":arm-none-eabi-gcc@4.8.3-2014q1", tools[idx].String())
	requireEquivalentPaths(t, tools[idx].InstallDir.String(), "downloaded_tools/arm-none-eabi-gcc/4.8.3-2014q1")
	idx++
	require.Equal(t, ":avr-gcc@4.8.1-arduino5", tools[idx].String())
	requireEquivalentPaths(t, tools[idx].InstallDir.String(), "downloaded_tools/avr-gcc/4.8.1-arduino5")
	idx++
	require.Equal(t, "arduino:avr-gcc@4.8.1-arduino5", tools[idx].String())
	requireEquivalentPaths(t, tools[idx].InstallDir.String(), "tools_builtin/avr")
	idx++
	require.Equal(t, ":avrdude@6.0.1-arduino5", tools[idx].String())
	requireEquivalentPaths(t, tools[idx].InstallDir.String(), "downloaded_tools/avrdude/6.0.1-arduino5")
	idx++
	require.Equal(t, "arduino:avrdude@6.0.1-arduino5", tools[idx].String())
	requireEquivalentPaths(t, tools[idx].InstallDir.String(), "tools_builtin/avr")
	idx++
	require.Equal(t, ":bossac@1.5-arduino", tools[idx].String())
	requireEquivalentPaths(t, tools[idx].InstallDir.String(), "downloaded_tools/bossac/1.5-arduino")
	idx++
	require.Equal(t, ":bossac@1.6.1-arduino", tools[idx].String())
	requireEquivalentPaths(t, tools[idx].InstallDir.String(), "downloaded_tools/bossac/1.6.1-arduino")
	idx++
	require.Equal(t, ":ctags@5.8-arduino11", tools[idx].String())
	requireEquivalentPaths(t, tools[idx].InstallDir.String(), "downloaded_tools/ctags/5.8-arduino11")
}

func TestLoadToolsWithBoardManagerFolderStructure(t *testing.T) {
	DownloadCoresAndToolsAndLibraries(t)
	ctx := &types.Context{
		HardwareDirs: paths.NewPathList("downloaded_board_manager_stuff"),
	}

	NoError(t, (&builder.HardwareLoader{}).Run(ctx))
	NoError(t, (&builder.ToolsLoader{}).Run(ctx))

	tools := ctx.AllTools
	require.Equal(t, 3, len(tools))

	sort.Sort(ByToolIDAndVersion(tools))

	idx := 0
	require.Equal(t, "arduino:CMSIS@4.0.0-atmel", tools[idx].String())
	requireEquivalentPaths(t, tools[idx].InstallDir.String(), "downloaded_board_manager_stuff/arduino/tools/CMSIS/4.0.0-atmel")
	idx++
	require.Equal(t, "RFduino:arm-none-eabi-gcc@4.8.3-2014q1", tools[idx].String())
	requireEquivalentPaths(t, tools[idx].InstallDir.String(), "downloaded_board_manager_stuff/RFduino/tools/arm-none-eabi-gcc/4.8.3-2014q1")
	idx++
	require.Equal(t, "arduino:openocd@0.9.0-arduino", tools[idx].String())
	requireEquivalentPaths(t, tools[idx].InstallDir.String(), "downloaded_board_manager_stuff/arduino/tools/openocd/0.9.0-arduino")
}

func TestLoadLotsOfTools(t *testing.T) {
	DownloadCoresAndToolsAndLibraries(t)

	ctx := &types.Context{
		HardwareDirs:     paths.NewPathList("downloaded_board_manager_stuff"),
		BuiltInToolsDirs: paths.NewPathList("downloaded_tools", "tools_builtin"),
	}

	NoError(t, (&builder.HardwareLoader{}).Run(ctx))
	NoError(t, (&builder.ToolsLoader{}).Run(ctx))

	tools := ctx.AllTools
	require.Equal(t, 12, len(tools))

	sort.Sort(ByToolIDAndVersion(tools))

	idx := 0
	require.Equal(t, "arduino:CMSIS@4.0.0-atmel", tools[idx].String())
	requireEquivalentPaths(t, tools[idx].InstallDir.String(), "downloaded_board_manager_stuff/arduino/tools/CMSIS/4.0.0-atmel")
	idx++
	require.Equal(t, ":arduino-preprocessor@0.1.5", tools[idx].String())
	requireEquivalentPaths(t, tools[idx].InstallDir.String(), "downloaded_tools/arduino-preprocessor/0.1.5")
	idx++
	require.Equal(t, "RFduino:arm-none-eabi-gcc@4.8.3-2014q1", tools[idx].String())
	requireEquivalentPaths(t, tools[idx].InstallDir.String(), "downloaded_board_manager_stuff/RFduino/tools/arm-none-eabi-gcc/4.8.3-2014q1")
	idx++
	require.Equal(t, ":arm-none-eabi-gcc@4.8.3-2014q1", tools[idx].String())
	requireEquivalentPaths(t, tools[idx].InstallDir.String(), "downloaded_tools/arm-none-eabi-gcc/4.8.3-2014q1")
	idx++
	require.Equal(t, ":avr-gcc@4.8.1-arduino5", tools[idx].String())
	requireEquivalentPaths(t, tools[idx].InstallDir.String(), "downloaded_tools/avr-gcc/4.8.1-arduino5")
	idx++
	require.Equal(t, "arduino:avr-gcc@4.8.1-arduino5", tools[idx].String())
	requireEquivalentPaths(t, tools[idx].InstallDir.String(), "tools_builtin/avr")
	idx++
	require.Equal(t, ":avrdude@6.0.1-arduino5", tools[idx].String())
	requireEquivalentPaths(t, tools[idx].InstallDir.String(), "downloaded_tools/avrdude/6.0.1-arduino5")
	idx++
	require.Equal(t, "arduino:avrdude@6.0.1-arduino5", tools[idx].String())
	requireEquivalentPaths(t, tools[idx].InstallDir.String(), "tools_builtin/avr")
	idx++
	require.Equal(t, ":bossac@1.5-arduino", tools[idx].String())
	requireEquivalentPaths(t, tools[idx].InstallDir.String(), "downloaded_tools/bossac/1.5-arduino")
	idx++
	require.Equal(t, ":bossac@1.6.1-arduino", tools[idx].String())
	requireEquivalentPaths(t, tools[idx].InstallDir.String(), tools[idx].InstallDir.String(), "downloaded_tools/bossac/1.6.1-arduino")
	idx++
	require.Equal(t, ":ctags@5.8-arduino11", tools[idx].String())
	requireEquivalentPaths(t, tools[idx].InstallDir.String(), "downloaded_tools/ctags/5.8-arduino11")
	idx++
	require.Equal(t, "arduino:openocd@0.9.0-arduino", tools[idx].String())
	requireEquivalentPaths(t, tools[idx].InstallDir.String(), "downloaded_board_manager_stuff/arduino/tools/openocd/0.9.0-arduino")
}
