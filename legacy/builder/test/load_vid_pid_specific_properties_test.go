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

func TestLoadVIDPIDSpecificPropertiesWhenNoVIDPIDAreProvided(t *testing.T) {
	DownloadCoresAndToolsAndLibraries(t)

	ctx := &types.Context{
		HardwareDirs:      paths.NewPathList(filepath.Join("..", "hardware"), "downloaded_hardware"),
		BuiltInToolsDirs:  paths.NewPathList("downloaded_tools", "./tools_builtin"),
		SketchLocation:    paths.New("sketch1", "sketch1.ino"),
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
		HardwareDirs:      paths.NewPathList(filepath.Join("..", "hardware"), "downloaded_hardware"),
		BuiltInToolsDirs:  paths.NewPathList("downloaded_tools", "./tools_builtin"),
		SketchLocation:    paths.New("sketch1", "sketch1.ino"),
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
