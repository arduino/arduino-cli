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
	"testing"

	"github.com/arduino/arduino-cli/configuration"
	"github.com/arduino/arduino-cli/internal/cli/instance"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	"github.com/arduino/go-paths-helper"
	"github.com/stretchr/testify/require"
)

func TestPlatformSearch(t *testing.T) {
	dataDir := paths.TempDir().Join("test", "data_dir")
	downloadDir := paths.TempDir().Join("test", "staging")
	t.Setenv("ARDUINO_DATA_DIR", dataDir.String())
	t.Setenv("ARDUINO_DOWNLOADS_DIR", downloadDir.String())
	dataDir.MkdirAll()
	downloadDir.MkdirAll()
	defer paths.TempDir().Join("test").RemoveAll()
	err := paths.New("testdata").Join("package_index.json").CopyTo(dataDir.Join("package_index.json"))
	require.Nil(t, err)

	configuration.Settings = configuration.Init(paths.TempDir().Join("test", "arduino-cli.yaml").String())

	inst := instance.CreateAndInit()
	require.NotNil(t, inst)

	t.Run("SearchAllVersions", func(t *testing.T) {
		res, stat := PlatformSearch(&rpc.PlatformSearchRequest{
			Instance:   inst,
			SearchArgs: "retrokit",
		})
		require.Nil(t, stat)
		require.NotNil(t, res)

		require.Len(t, res.GetSearchOutput(), 1)
		require.Contains(t, res.GetSearchOutput(), &rpc.PlatformSummary{
			Metadata: &rpc.PlatformMetadata{
				Id:         "Retrokits-RK002:arm",
				Maintainer: "Retrokits (www.retrokits.com)",
				Website:    "https://www.retrokits.com",
				Email:      "info@retrokits.com",
				Indexed:    true,
			},
			Releases: map[string]*rpc.PlatformRelease{
				"1.0.5": {
					Name:       "RK002",
					Type:       []string{"Contributed"},
					Installed:  false,
					Version:    "1.0.5",
					Boards:     []*rpc.Board{{Name: "RK002"}},
					Help:       &rpc.HelpResources{Online: "https://www.retrokits.com/rk002/arduino"},
					Compatible: false,
				},
				"1.0.6": {
					Name:       "RK002",
					Type:       []string{"Contributed"},
					Installed:  false,
					Version:    "1.0.6",
					Boards:     []*rpc.Board{{Name: "RK002"}},
					Help:       &rpc.HelpResources{Online: "https://www.retrokits.com/rk002/arduino"},
					Compatible: false,
				},
			},
			InstalledVersion: "",
		})
	})

	t.Run("SearchThePackageMaintainer", func(t *testing.T) {
		res, stat := PlatformSearch(&rpc.PlatformSearchRequest{
			Instance:   inst,
			SearchArgs: "Retrokits (www.retrokits.com)",
		})
		require.Nil(t, stat)
		require.NotNil(t, res)
		require.Len(t, res.GetSearchOutput(), 1)
		require.Contains(t, res.GetSearchOutput(), &rpc.PlatformSummary{
			Metadata: &rpc.PlatformMetadata{
				Id:         "Retrokits-RK002:arm",
				Maintainer: "Retrokits (www.retrokits.com)",
				Website:    "https://www.retrokits.com",
				Email:      "info@retrokits.com",
				Indexed:    true,
			},
			Releases: map[string]*rpc.PlatformRelease{
				"1.0.5": {
					Name:       "RK002",
					Type:       []string{"Contributed"},
					Installed:  false,
					Version:    "1.0.5",
					Boards:     []*rpc.Board{{Name: "RK002"}},
					Help:       &rpc.HelpResources{Online: "https://www.retrokits.com/rk002/arduino"},
					Compatible: false,
				},
				"1.0.6": {
					Name:       "RK002",
					Type:       []string{"Contributed"},
					Installed:  false,
					Version:    "1.0.6",
					Boards:     []*rpc.Board{{Name: "RK002"}},
					Help:       &rpc.HelpResources{Online: "https://www.retrokits.com/rk002/arduino"},
					Compatible: false,
				},
			},
			InstalledVersion: "",
		})
	})

	t.Run("SearchPackageName", func(t *testing.T) {
		res, stat := PlatformSearch(&rpc.PlatformSearchRequest{
			Instance:   inst,
			SearchArgs: "Retrokits-RK002",
		})
		require.Nil(t, stat)
		require.NotNil(t, res)
		require.Len(t, res.GetSearchOutput(), 1)
		require.Contains(t, res.GetSearchOutput(), &rpc.PlatformSummary{
			Metadata: &rpc.PlatformMetadata{
				Id:         "Retrokits-RK002:arm",
				Maintainer: "Retrokits (www.retrokits.com)",
				Website:    "https://www.retrokits.com",
				Email:      "info@retrokits.com",
				Indexed:    true,
			},
			Releases: map[string]*rpc.PlatformRelease{
				"1.0.5": {
					Name:       "RK002",
					Type:       []string{"Contributed"},
					Installed:  false,
					Version:    "1.0.5",
					Boards:     []*rpc.Board{{Name: "RK002"}},
					Help:       &rpc.HelpResources{Online: "https://www.retrokits.com/rk002/arduino"},
					Compatible: false,
				},
				"1.0.6": {
					Name:       "RK002",
					Type:       []string{"Contributed"},
					Installed:  false,
					Version:    "1.0.6",
					Boards:     []*rpc.Board{{Name: "RK002"}},
					Help:       &rpc.HelpResources{Online: "https://www.retrokits.com/rk002/arduino"},
					Compatible: false,
				},
			},
			InstalledVersion: "",
		})
	})

	t.Run("SearchPlatformName", func(t *testing.T) {
		res, stat := PlatformSearch(&rpc.PlatformSearchRequest{
			Instance:   inst,
			SearchArgs: "rk002",
		})
		require.Nil(t, stat)
		require.NotNil(t, res)
		require.Len(t, res.GetSearchOutput(), 1)
		require.Contains(t, res.GetSearchOutput(), &rpc.PlatformSummary{
			Metadata: &rpc.PlatformMetadata{
				Id:         "Retrokits-RK002:arm",
				Maintainer: "Retrokits (www.retrokits.com)",
				Website:    "https://www.retrokits.com",
				Email:      "info@retrokits.com",
				Indexed:    true,
			},
			Releases: map[string]*rpc.PlatformRelease{
				"1.0.5": {
					Name:       "RK002",
					Type:       []string{"Contributed"},
					Installed:  false,
					Version:    "1.0.5",
					Boards:     []*rpc.Board{{Name: "RK002"}},
					Help:       &rpc.HelpResources{Online: "https://www.retrokits.com/rk002/arduino"},
					Compatible: false,
				},
				"1.0.6": {
					Name:       "RK002",
					Type:       []string{"Contributed"},
					Installed:  false,
					Version:    "1.0.6",
					Boards:     []*rpc.Board{{Name: "RK002"}},
					Help:       &rpc.HelpResources{Online: "https://www.retrokits.com/rk002/arduino"},
					Compatible: false,
				},
			},
			InstalledVersion: "",
		})
	})

	t.Run("SearchBoardName", func(t *testing.T) {
		res, stat := PlatformSearch(&rpc.PlatformSearchRequest{
			Instance:   inst,
			SearchArgs: "Yún",
		})
		require.Nil(t, stat)
		require.NotNil(t, res)
		require.Len(t, res.GetSearchOutput(), 1)
		require.Contains(t, res.GetSearchOutput(), &rpc.PlatformSummary{
			Metadata: &rpc.PlatformMetadata{
				Id:         "arduino:avr",
				Maintainer: "Arduino",
				Website:    "https://www.arduino.cc/",
				Email:      "packages@arduino.cc",
				Indexed:    true,
			},
			Releases: map[string]*rpc.PlatformRelease{
				"1.8.3": {
					Name:      "Arduino AVR Boards",
					Type:      []string{"Arduino"},
					Installed: false,
					Version:   "1.8.3",
					Boards: []*rpc.Board{
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
					Help:       &rpc.HelpResources{Online: "http://www.arduino.cc/en/Reference/HomePage"},
					Compatible: false,
				},
			},
			InstalledVersion: "",
		})
	})

	t.Run("SearchBoardName2", func(t *testing.T) {
		res, stat := PlatformSearch(&rpc.PlatformSearchRequest{
			Instance:   inst,
			SearchArgs: "yun",
		})
		require.Nil(t, stat)
		require.NotNil(t, res)
		require.Len(t, res.GetSearchOutput(), 1)
		require.Contains(t, res.GetSearchOutput(), &rpc.PlatformSummary{
			Metadata: &rpc.PlatformMetadata{
				Id:         "arduino:avr",
				Indexed:    true,
				Maintainer: "Arduino",
				Website:    "https://www.arduino.cc/",
				Email:      "packages@arduino.cc",
			},
			Releases: map[string]*rpc.PlatformRelease{
				"1.8.3": {
					Name:      "Arduino AVR Boards",
					Type:      []string{"Arduino"},
					Installed: false,
					Version:   "1.8.3",
					Boards: []*rpc.Board{
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
					Help:       &rpc.HelpResources{Online: "http://www.arduino.cc/en/Reference/HomePage"},
					Compatible: false,
				},
			},
			InstalledVersion: "",
		})
	})
}

func TestPlatformSearchSorting(t *testing.T) {
	dataDir := paths.TempDir().Join("test", "data_dir")
	downloadDir := paths.TempDir().Join("test", "staging")
	t.Setenv("ARDUINO_DATA_DIR", dataDir.String())
	t.Setenv("ARDUINO_DOWNLOADS_DIR", downloadDir.String())
	dataDir.MkdirAll()
	downloadDir.MkdirAll()
	defer paths.TempDir().Join("test").RemoveAll()
	err := paths.New("testdata").Join("package_index.json").CopyTo(dataDir.Join("package_index.json"))
	require.Nil(t, err)

	configuration.Settings = configuration.Init(paths.TempDir().Join("test", "arduino-cli.yaml").String())

	inst := instance.CreateAndInit()
	require.NotNil(t, inst)

	res, stat := PlatformSearch(&rpc.PlatformSearchRequest{
		Instance:   inst,
		SearchArgs: "",
	})
	require.Nil(t, stat)
	require.NotNil(t, res)

	require.Len(t, res.GetSearchOutput(), 3)
	require.Equal(t, res.GetSearchOutput()[0].GetSortedReleases()[0].GetName(), "Arduino AVR Boards")
	require.Equal(t, res.GetSearchOutput()[0].GetMetadata().GetDeprecated(), false)
	require.Equal(t, res.GetSearchOutput()[1].GetSortedReleases()[0].GetName(), "RK002")
	require.Equal(t, res.GetSearchOutput()[1].GetMetadata().GetDeprecated(), false)
	require.Equal(t, res.GetSearchOutput()[2].GetLatestRelease().GetName(), "Platform")
	require.Equal(t, res.GetSearchOutput()[2].GetMetadata().GetDeprecated(), true)
}
