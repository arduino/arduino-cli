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

	"github.com/arduino/arduino-cli/arduino/cores"
	"github.com/arduino/arduino-cli/arduino/cores/packagemanager"
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
	ctx := &types.Context{
		HardwareDirs:     paths.NewPathList(filepath.Join("..", "hardware"), "downloaded_hardware"),
		BuiltInToolsDirs: paths.NewPathList("downloaded_tools", "tools_builtin"),
	}
	ctx = prepareBuilderTestContext(t, ctx, nil, "")
	defer cleanUpBuilderTestContext(t, ctx)

	tools := ctx.PackageManager.GetAllInstalledToolsReleases()
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
	ctx := &types.Context{
		HardwareDirs: paths.NewPathList("downloaded_board_manager_stuff"),
	}
	ctx = prepareBuilderTestContext(t, ctx, nil, "")
	defer cleanUpBuilderTestContext(t, ctx)

	tools := ctx.PackageManager.GetAllInstalledToolsReleases()
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
	ctx := &types.Context{
		HardwareDirs:     paths.NewPathList("downloaded_board_manager_stuff"),
		BuiltInToolsDirs: paths.NewPathList("downloaded_tools", "tools_builtin"),
	}
	ctx = prepareBuilderTestContext(t, ctx, nil, "")
	defer cleanUpBuilderTestContext(t, ctx)

	tools := ctx.PackageManager.GetAllInstalledToolsReleases()
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

func TestAllToolsContextIsPopulated(t *testing.T) {
	pmb := packagemanager.NewBuilder(nil, nil, nil, nil, "")
	pmb.LoadHardwareFromDirectories(paths.NewPathList("downloaded_board_manager_stuff"))
	pmb.LoadToolsFromBundleDirectory(paths.New("downloaded_tools", "tools_builtin"))
	pm := pmb.Build()
	pme, release := pm.NewExplorer()
	defer release()

	ctx := &types.Context{
		PackageManager: pme,
	}

	require.NotEmpty(t, ctx.PackageManager.GetAllInstalledToolsReleases())
}
