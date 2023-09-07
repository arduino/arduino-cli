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
	"runtime"
	"testing"

	"github.com/arduino/arduino-cli/legacy/builder/types"
	paths "github.com/arduino/go-paths-helper"
	"github.com/stretchr/testify/require"
)

func TestLoadHardwareWithBoardManagerFolderStructure(t *testing.T) {
	ctx := &types.Context{
		HardwareDirs: paths.NewPathList("downloaded_board_manager_stuff"),
	}
	ctx = prepareBuilderTestContext(t, ctx, nil, "", skipLibraries)
	defer cleanUpBuilderTestContext(t, ctx)

	packages := ctx.PackageManager.GetPackages()
	require.Equal(t, 3, len(packages))
	require.NotNil(t, packages["arduino"])
	require.Equal(t, 1, len(packages["arduino"].Platforms))
	require.NotNil(t, packages["RedBearLab"])
	require.Equal(t, 1, len(packages["RedBearLab"].Platforms))
	require.NotNil(t, packages["RFduino"])
	require.Equal(t, 0, len(packages["RFduino"].Platforms))

	samdPlatform := packages["arduino"].Platforms["samd"].Releases["1.6.5"]
	require.Equal(t, 3, len(samdPlatform.Boards))

	require.Equal(t, "arduino_zero_edbg", samdPlatform.Boards["arduino_zero_edbg"].BoardID)
	require.Equal(t, "arduino_zero_edbg", samdPlatform.Boards["arduino_zero_edbg"].Properties.Get("_id"))

	require.Equal(t, "arduino_zero", samdPlatform.Boards["arduino_zero_native"].Properties.Get("build.variant"))
	require.Equal(t, "-D__SAMD21G18A__ {build.usb_flags}", samdPlatform.Boards["arduino_zero_native"].Properties.Get("build.extra_flags"))

	require.Equal(t, "Arduino SAMD (32-bits ARM Cortex-M0+) Boards", samdPlatform.Properties.Get("name"))
	require.Equal(t, "-d3", samdPlatform.Properties.Get("tools.openocd.erase.params.verbose"))

	require.Equal(t, 3, len(samdPlatform.Programmers))

	require.Equal(t, "Atmel EDBG", samdPlatform.Programmers["edbg"].Name)
	require.Equal(t, "openocd", samdPlatform.Programmers["edbg"].Properties.Get("program.tool"))

	avrRedBearPlatform := packages["RedBearLab"].Platforms["avr"].Releases["1.0.0"]
	require.Equal(t, 3, len(avrRedBearPlatform.Boards))

	require.Equal(t, "blend", avrRedBearPlatform.Boards["blend"].BoardID)
	require.Equal(t, "blend", avrRedBearPlatform.Boards["blend"].Properties.Get("_id"))
	require.Equal(t, "arduino:arduino", avrRedBearPlatform.Boards["blend"].Properties.Get("build.core"))
}

func TestLoadLotsOfHardware(t *testing.T) {
	ctx := &types.Context{
		HardwareDirs: paths.NewPathList("downloaded_board_manager_stuff", "downloaded_hardware", filepath.Join("..", "hardware"), "user_hardware"),
	}
	ctx = prepareBuilderTestContext(t, ctx, nil, "", skipLibraries)
	defer cleanUpBuilderTestContext(t, ctx)

	packages := ctx.PackageManager.GetPackages()

	if runtime.GOOS == "windows" {
		//a package is a symlink, and windows does not support them
		require.Equal(t, 4, len(packages))
	} else {
		require.Equal(t, 5, len(packages))
	}

	require.NotNil(t, packages["arduino"])
	require.NotNil(t, packages["my_avr_platform"])

	require.Equal(t, 3, len(packages["arduino"].Platforms))
	require.Equal(t, 20, len(packages["arduino"].Platforms["avr"].Releases["1.6.10"].Boards))
	require.Equal(t, 2, len(packages["arduino"].Platforms["sam"].Releases["1.6.7"].Boards))
	require.Equal(t, 3, len(packages["arduino"].Platforms["samd"].Releases["1.6.5"].Boards))

	require.Equal(t, 1, len(packages["my_avr_platform"].Platforms))
	require.Equal(t, 2, len(packages["my_avr_platform"].Platforms["avr"].Releases["9.9.9"].Boards))

	if runtime.GOOS != "windows" {
		require.Equal(t, 1, len(packages["my_symlinked_avr_platform"].Platforms))
		require.Equal(t, 2, len(packages["my_symlinked_avr_platform"].Platforms["avr"].Releases["9.9.9"].Boards))
	}
}
