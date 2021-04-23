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

package core

import (
	"os"
	"testing"

	"github.com/arduino/arduino-cli/cli/instance"
	"github.com/arduino/arduino-cli/configuration"
	"github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	"github.com/arduino/go-paths-helper"
	"github.com/stretchr/testify/require"
)

func TestPlatformSearch(t *testing.T) {

	dataDir := paths.TempDir().Join("test", "data_dir")
	downloadDir := paths.TempDir().Join("test", "staging")
	os.Setenv("ARDUINO_DATA_DIR", dataDir.String())
	os.Setenv("ARDUINO_DOWNLOADS_DIR", downloadDir.String())
	dataDir.MkdirAll()
	downloadDir.MkdirAll()
	defer paths.TempDir().Join("test").RemoveAll()
	err := paths.New("testdata").Join("package_index.json").CopyTo(dataDir.Join("package_index.json"))
	require.Nil(t, err)

	configuration.Settings = configuration.Init(paths.TempDir().Join("test", "arduino-cli.yaml").String())

	inst := instance.CreateAndInit()
	require.NotNil(t, inst)

	res, err := PlatformSearch(&rpc.PlatformSearchRequest{
		Instance:    inst,
		SearchArgs:  "retrokit",
		AllVersions: true,
	})
	require.Nil(t, err)
	require.NotNil(t, res)

	require.Len(t, res.SearchOutput, 2)
	require.Contains(t, res.SearchOutput, &commands.Platform{
		Id:         "Retrokits-RK002:arm",
		Installed:  "",
		Latest:     "1.0.5",
		Name:       "RK002",
		Maintainer: "Retrokits (www.retrokits.com)",
		Website:    "https://www.retrokits.com",
		Email:      "info@retrokits.com",
		Boards:     []*commands.Board{{Name: "RK002"}},
	})
	require.Contains(t, res.SearchOutput, &commands.Platform{
		Id:         "Retrokits-RK002:arm",
		Installed:  "",
		Latest:     "1.0.6",
		Name:       "RK002",
		Maintainer: "Retrokits (www.retrokits.com)",
		Website:    "https://www.retrokits.com",
		Email:      "info@retrokits.com",
		Boards:     []*commands.Board{{Name: "RK002"}},
	})

	res, err = PlatformSearch(&rpc.PlatformSearchRequest{
		Instance:    inst,
		SearchArgs:  "retrokit",
		AllVersions: false,
	})
	require.Nil(t, err)
	require.NotNil(t, res)
	require.Len(t, res.SearchOutput, 1)
	require.Contains(t, res.SearchOutput, &commands.Platform{
		Id:         "Retrokits-RK002:arm",
		Installed:  "",
		Latest:     "1.0.6",
		Name:       "RK002",
		Maintainer: "Retrokits (www.retrokits.com)",
		Website:    "https://www.retrokits.com",
		Email:      "info@retrokits.com",
		Boards:     []*commands.Board{{Name: "RK002"}},
	})

	// Search the Package Maintainer
	res, err = PlatformSearch(&rpc.PlatformSearchRequest{
		Instance:    inst,
		SearchArgs:  "Retrokits (www.retrokits.com)",
		AllVersions: true,
	})
	require.Nil(t, err)
	require.NotNil(t, res)
	require.Len(t, res.SearchOutput, 2)
	require.Contains(t, res.SearchOutput, &commands.Platform{
		Id:         "Retrokits-RK002:arm",
		Installed:  "",
		Latest:     "1.0.5",
		Name:       "RK002",
		Maintainer: "Retrokits (www.retrokits.com)",
		Website:    "https://www.retrokits.com",
		Email:      "info@retrokits.com",
		Boards:     []*commands.Board{{Name: "RK002"}},
	})
	require.Contains(t, res.SearchOutput, &commands.Platform{
		Id:         "Retrokits-RK002:arm",
		Installed:  "",
		Latest:     "1.0.6",
		Name:       "RK002",
		Maintainer: "Retrokits (www.retrokits.com)",
		Website:    "https://www.retrokits.com",
		Email:      "info@retrokits.com",
		Boards:     []*commands.Board{{Name: "RK002"}},
	})

	// Search using the Package name
	res, err = PlatformSearch(&rpc.PlatformSearchRequest{
		Instance:    inst,
		SearchArgs:  "Retrokits-RK002",
		AllVersions: true,
	})
	require.Nil(t, err)
	require.NotNil(t, res)
	require.Len(t, res.SearchOutput, 2)
	require.Contains(t, res.SearchOutput, &commands.Platform{
		Id:         "Retrokits-RK002:arm",
		Installed:  "",
		Latest:     "1.0.5",
		Name:       "RK002",
		Maintainer: "Retrokits (www.retrokits.com)",
		Website:    "https://www.retrokits.com",
		Email:      "info@retrokits.com",
		Boards:     []*commands.Board{{Name: "RK002"}},
	})
	require.Contains(t, res.SearchOutput, &commands.Platform{
		Id:         "Retrokits-RK002:arm",
		Installed:  "",
		Latest:     "1.0.6",
		Name:       "RK002",
		Maintainer: "Retrokits (www.retrokits.com)",
		Website:    "https://www.retrokits.com",
		Email:      "info@retrokits.com",
		Boards:     []*commands.Board{{Name: "RK002"}},
	})

	// Search using the Platform name
	res, err = PlatformSearch(&rpc.PlatformSearchRequest{
		Instance:    inst,
		SearchArgs:  "rk002",
		AllVersions: true,
	})
	require.Nil(t, err)
	require.NotNil(t, res)
	require.Len(t, res.SearchOutput, 2)
	require.Contains(t, res.SearchOutput, &commands.Platform{
		Id:         "Retrokits-RK002:arm",
		Installed:  "",
		Latest:     "1.0.5",
		Name:       "RK002",
		Maintainer: "Retrokits (www.retrokits.com)",
		Website:    "https://www.retrokits.com",
		Email:      "info@retrokits.com",
		Boards:     []*commands.Board{{Name: "RK002"}},
	})
	require.Contains(t, res.SearchOutput, &commands.Platform{
		Id:         "Retrokits-RK002:arm",
		Installed:  "",
		Latest:     "1.0.6",
		Name:       "RK002",
		Maintainer: "Retrokits (www.retrokits.com)",
		Website:    "https://www.retrokits.com",
		Email:      "info@retrokits.com",
		Boards:     []*commands.Board{{Name: "RK002"}},
	})

	// Search using a board name
	res, err = PlatformSearch(&rpc.PlatformSearchRequest{
		Instance:    inst,
		SearchArgs:  "Yún",
		AllVersions: true,
	})
	require.Nil(t, err)
	require.NotNil(t, res)
	require.Len(t, res.SearchOutput, 1)
	require.Contains(t, res.SearchOutput, &commands.Platform{
		Id:         "arduino:avr",
		Installed:  "",
		Latest:     "1.8.3",
		Name:       "Arduino AVR Boards",
		Maintainer: "Arduino",
		Website:    "https://www.arduino.cc/",
		Email:      "packages@arduino.cc",
		Boards: []*commands.Board{
			{Name: "Arduino Yún"},
			{Name: "Arduino Uno"},
			{Name: "Arduino Uno WiFi"},
			{Name: "Arduino Diecimila"},
			{Name: "Arduino Nano"},
			{Name: "Arduino Mega"},
			{Name: "Arduino MegaADK"},
			{Name: "Arduino Leonardo"},
			{Name: "Arduino Leonardo Ethernet"},
			{Name: "Arduino Micro"},
			{Name: "Arduino Esplora"},
			{Name: "Arduino Mini"},
			{Name: "Arduino Ethernet"},
			{Name: "Arduino Fio"},
			{Name: "Arduino BT"},
			{Name: "Arduino LilyPadUSB"},
			{Name: "Arduino Lilypad"},
			{Name: "Arduino Pro"},
			{Name: "Arduino ATMegaNG"},
			{Name: "Arduino Robot Control"},
			{Name: "Arduino Robot Motor"},
			{Name: "Arduino Gemma"},
			{Name: "Adafruit Circuit Playground"},
			{Name: "Arduino Yún Mini"},
			{Name: "Arduino Industrial 101"},
			{Name: "Linino One"},
		},
	})

	res, err = PlatformSearch(&rpc.PlatformSearchRequest{
		Instance:    inst,
		SearchArgs:  "yun",
		AllVersions: true,
	})
	require.Nil(t, err)
	require.NotNil(t, res)
	require.Len(t, res.SearchOutput, 1)
	require.Contains(t, res.SearchOutput, &commands.Platform{
		Id:         "arduino:avr",
		Installed:  "",
		Latest:     "1.8.3",
		Name:       "Arduino AVR Boards",
		Maintainer: "Arduino",
		Website:    "https://www.arduino.cc/",
		Email:      "packages@arduino.cc",
		Boards: []*commands.Board{
			{Name: "Arduino Yún"},
			{Name: "Arduino Uno"},
			{Name: "Arduino Uno WiFi"},
			{Name: "Arduino Diecimila"},
			{Name: "Arduino Nano"},
			{Name: "Arduino Mega"},
			{Name: "Arduino MegaADK"},
			{Name: "Arduino Leonardo"},
			{Name: "Arduino Leonardo Ethernet"},
			{Name: "Arduino Micro"},
			{Name: "Arduino Esplora"},
			{Name: "Arduino Mini"},
			{Name: "Arduino Ethernet"},
			{Name: "Arduino Fio"},
			{Name: "Arduino BT"},
			{Name: "Arduino LilyPadUSB"},
			{Name: "Arduino Lilypad"},
			{Name: "Arduino Pro"},
			{Name: "Arduino ATMegaNG"},
			{Name: "Arduino Robot Control"},
			{Name: "Arduino Robot Motor"},
			{Name: "Arduino Gemma"},
			{Name: "Adafruit Circuit Playground"},
			{Name: "Arduino Yún Mini"},
			{Name: "Arduino Industrial 101"},
			{Name: "Linino One"},
		},
	})
}

func TestPlatformSearchSorting(t *testing.T) {
	dataDir := paths.TempDir().Join("test", "data_dir")
	downloadDir := paths.TempDir().Join("test", "staging")
	os.Setenv("ARDUINO_DATA_DIR", dataDir.String())
	os.Setenv("ARDUINO_DOWNLOADS_DIR", downloadDir.String())
	dataDir.MkdirAll()
	downloadDir.MkdirAll()
	defer paths.TempDir().Join("test").RemoveAll()
	err := paths.New("testdata").Join("package_index.json").CopyTo(dataDir.Join("package_index.json"))
	require.Nil(t, err)

	configuration.Settings = configuration.Init(paths.TempDir().Join("test", "arduino-cli.yaml").String())

	inst, err := instance.CreateInstance()
	require.Nil(t, err)
	require.NotNil(t, inst)

	res, err := PlatformSearch(&rpc.PlatformSearchRequest{
		Instance:    inst,
		SearchArgs:  "",
		AllVersions: false,
	})
	require.Nil(t, err)
	require.NotNil(t, res)

	require.Len(t, res.SearchOutput, 3)
	require.Equal(t, res.SearchOutput[0].Name, "Arduino AVR Boards")
	require.Equal(t, res.SearchOutput[0].Deprecated, false)
	require.Equal(t, res.SearchOutput[1].Name, "RK002")
	require.Equal(t, res.SearchOutput[1].Deprecated, false)
	require.Equal(t, res.SearchOutput[2].Name, "Platform")
	require.Equal(t, res.SearchOutput[2].Deprecated, true)

}
