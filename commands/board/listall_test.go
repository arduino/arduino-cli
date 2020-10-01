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

package board

import (
	"context"
	"os"
	"testing"

	"github.com/arduino/arduino-cli/cli/instance"
	"github.com/arduino/arduino-cli/cli/output"
	"github.com/arduino/arduino-cli/commands/core"
	"github.com/arduino/arduino-cli/configuration"
	rpc "github.com/arduino/arduino-cli/rpc/commands"
	"github.com/arduino/go-paths-helper"
	"github.com/stretchr/testify/require"
)

func TestListAll(t *testing.T) {
	dataDir := paths.TempDir().Join("test", "data_dir")
	downloadDir := paths.TempDir().Join("test", "staging")
	os.Setenv("ARDUINO_DATA_DIR", dataDir.String())
	os.Setenv("ARDUINO_DOWNLOADS_DIR", downloadDir.String())
	dataDir.MkdirAll()
	downloadDir.MkdirAll()
	defer paths.TempDir().Join("test").RemoveAll()
	err := paths.New("testdata").Join("package_index.json").CopyTo(dataDir.Join("package_index.json"))
	require.Nil(t, err)

	configuration.Init(paths.TempDir().Join("test", "arduino-cli.yaml").String())

	inst, err := instance.CreateInstance()
	require.Nil(t, err)

	_, err = core.PlatformInstall(context.Background(),
		&rpc.PlatformInstallReq{
			Instance:        inst,
			PlatformPackage: "arduino",
			Architecture:    "avr",
			Version:         "1.8.3",
			SkipPostInstall: true,
		}, output.NewDownloadProgressBarCB(), output.NewTaskProgressCB())
	require.Nil(t, err)

	res, err := ListAll(context.Background(), &rpc.BoardListAllReq{
		Instance:            inst,
		SearchArgs:          []string{},
		IncludeHiddenBoards: true,
	})
	require.Nil(t, err)
	require.NotNil(t, res)
	require.Equal(t, 26, len(res.Boards))
	require.Contains(t, res.Boards, &rpc.BoardListItem{
		Name:     "Arduino Yún",
		FQBN:     "arduino:avr:yun",
		IsHidden: false,
	})
	require.Contains(t, res.Boards, &rpc.BoardListItem{
		Name:     "Arduino Uno",
		FQBN:     "arduino:avr:uno",
		IsHidden: false,
	})
	require.Contains(t, res.Boards, &rpc.BoardListItem{
		Name:     "Arduino Duemilanove or Diecimila",
		FQBN:     "arduino:avr:diecimila",
		IsHidden: false,
	})
	require.Contains(t, res.Boards, &rpc.BoardListItem{
		Name:     "Arduino Nano",
		FQBN:     "arduino:avr:nano",
		IsHidden: false,
	})
	require.Contains(t, res.Boards, &rpc.BoardListItem{
		Name:     "Arduino Mega or Mega 2560",
		FQBN:     "arduino:avr:mega",
		IsHidden: false,
	})
	require.Contains(t, res.Boards, &rpc.BoardListItem{
		Name:     "Arduino Mega ADK",
		FQBN:     "arduino:avr:megaADK",
		IsHidden: false,
	})
	require.Contains(t, res.Boards, &rpc.BoardListItem{
		Name:     "Arduino Leonardo",
		FQBN:     "arduino:avr:leonardo",
		IsHidden: false,
	})
	require.Contains(t, res.Boards, &rpc.BoardListItem{
		Name:     "Arduino Leonardo ETH",
		FQBN:     "arduino:avr:leonardoeth",
		IsHidden: false,
	})
	require.Contains(t, res.Boards, &rpc.BoardListItem{
		Name:     "Arduino Micro",
		FQBN:     "arduino:avr:micro",
		IsHidden: false,
	})
	require.Contains(t, res.Boards, &rpc.BoardListItem{
		Name:     "Arduino Esplora",
		FQBN:     "arduino:avr:esplora",
		IsHidden: false,
	})
	require.Contains(t, res.Boards, &rpc.BoardListItem{
		Name:     "Arduino Mini",
		FQBN:     "arduino:avr:mini",
		IsHidden: false,
	})
	require.Contains(t, res.Boards, &rpc.BoardListItem{
		Name:     "Arduino Ethernet",
		FQBN:     "arduino:avr:ethernet",
		IsHidden: false,
	})
	require.Contains(t, res.Boards, &rpc.BoardListItem{
		Name:     "Arduino Fio",
		FQBN:     "arduino:avr:fio",
		IsHidden: false,
	})
	require.Contains(t, res.Boards, &rpc.BoardListItem{
		Name:     "Arduino BT",
		FQBN:     "arduino:avr:bt",
		IsHidden: false,
	})
	require.Contains(t, res.Boards, &rpc.BoardListItem{
		Name:     "LilyPad Arduino USB",
		FQBN:     "arduino:avr:LilyPadUSB",
		IsHidden: false,
	})
	require.Contains(t, res.Boards, &rpc.BoardListItem{
		Name:     "LilyPad Arduino",
		FQBN:     "arduino:avr:lilypad",
		IsHidden: false,
	})
	require.Contains(t, res.Boards, &rpc.BoardListItem{
		Name:     "Arduino Pro or Pro Mini",
		FQBN:     "arduino:avr:pro",
		IsHidden: false,
	})
	require.Contains(t, res.Boards, &rpc.BoardListItem{
		Name:     "Arduino NG or older",
		FQBN:     "arduino:avr:atmegang",
		IsHidden: false,
	})
	require.Contains(t, res.Boards, &rpc.BoardListItem{
		Name:     "Arduino Robot Control",
		FQBN:     "arduino:avr:robotControl",
		IsHidden: false,
	})
	require.Contains(t, res.Boards, &rpc.BoardListItem{
		Name:     "Arduino Robot Motor",
		FQBN:     "arduino:avr:robotMotor",
		IsHidden: false,
	})
	require.Contains(t, res.Boards, &rpc.BoardListItem{
		Name:     "Arduino Gemma",
		FQBN:     "arduino:avr:gemma",
		IsHidden: false,
	})
	require.Contains(t, res.Boards, &rpc.BoardListItem{
		Name:     "Adafruit Circuit Playground",
		FQBN:     "arduino:avr:circuitplay32u4cat",
		IsHidden: false,
	})
	require.Contains(t, res.Boards, &rpc.BoardListItem{
		Name:     "Arduino Yún Mini",
		FQBN:     "arduino:avr:yunmini",
		IsHidden: false,
	})
	require.Contains(t, res.Boards, &rpc.BoardListItem{
		Name:     "Arduino Industrial 101",
		FQBN:     "arduino:avr:chiwawa",
		IsHidden: false,
	})
	require.Contains(t, res.Boards, &rpc.BoardListItem{
		Name:     "Linino One",
		FQBN:     "arduino:avr:one",
		IsHidden: false,
	})
	require.Contains(t, res.Boards, &rpc.BoardListItem{
		Name:     "Arduino Uno WiFi",
		FQBN:     "arduino:avr:unowifi",
		IsHidden: false,
	})

}
