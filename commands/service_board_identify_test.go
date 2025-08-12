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

package commands

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/arduino/arduino-cli/internal/arduino/cores/packagemanager"
	"github.com/arduino/arduino-cli/internal/cli/configuration"
	"github.com/arduino/go-paths-helper"
	"github.com/arduino/go-properties-orderedmap"
	"github.com/stretchr/testify/require"
	"go.bug.st/downloader/v2"
	semver "go.bug.st/relaxed-semver"
)

func TestGetByVidPid(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, `
{
  "items": [
    {
      "fqbn": "arduino:avr:uno",
      "vendor": "arduino",
      "architecture": "avr",
      "board_id": "uno",
      "name": "Arduino Uno"
    }
  ]
}
		`)
	}))
	defer ts.Close()

	vidPidURL = ts.URL
	settings := configuration.NewSettings()
	res, err := apiByVidPid(context.Background(), "0x2341", "0x0043", settings)
	require.Nil(t, err)
	require.Len(t, res, 1)
	require.Equal(t, "Arduino Uno", res[0].GetName())
	require.Equal(t, "arduino:avr:uno", res[0].GetFqbn())

	// wrong vid (too long), wrong pid (not an hex value)

	_, err = apiByVidPid(context.Background(), "0xfffff", "0xDEFG", settings)
	require.NotNil(t, err)
}

func TestGetByVidPidNotFound(t *testing.T) {
	settings := configuration.NewSettings()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	vidPidURL = ts.URL
	res, err := apiByVidPid(context.Background(), "0x0420", "0x0069", settings)
	require.NoError(t, err)
	require.Empty(t, res)
}

func TestGetByVidPid5xx(t *testing.T) {
	settings := configuration.NewSettings()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("500 - Ooooops!"))
	}))
	defer ts.Close()

	vidPidURL = ts.URL
	res, err := apiByVidPid(context.Background(), "0x0420", "0x0069", settings)
	require.NotNil(t, err)
	require.Equal(t, "the server responded with status 500 Internal Server Error", err.Error())
	require.Len(t, res, 0)
}

func TestGetByVidPidMalformedResponse(t *testing.T) {
	settings := configuration.NewSettings()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, `{"items":[{}]}`)
	}))
	defer ts.Close()

	vidPidURL = ts.URL
	res, err := apiByVidPid(context.Background(), "0x0420", "0x0069", settings)
	require.NotNil(t, err)
	require.Equal(t, "wrong format in server response", err.Error())
	require.Len(t, res, 0)
}

func TestBoardDetectionViaAPIWithNonUSBPort(t *testing.T) {
	settings := configuration.NewSettings()
	items, err := identifyViaCloudAPI(context.Background(), properties.NewMap(), settings)
	require.NoError(t, err)
	require.Empty(t, items)
}

func TestBoardIdentifySorting(t *testing.T) {
	dataDir := paths.TempDir().Join("test", "data_dir")
	t.Setenv("ARDUINO_DATA_DIR", dataDir.String())
	dataDir.MkdirAll()
	defer paths.TempDir().Join("test").RemoveAll()

	// We don't really care about the paths in this case
	pmb := packagemanager.NewBuilder(dataDir, dataDir, nil, dataDir, dataDir, "test", downloader.GetDefaultConfig())

	// Create some boards with identical VID:PID combination
	pack := pmb.GetOrCreatePackage("packager")
	pack.Maintainer = "NotArduino"
	platform := pack.GetOrCreatePlatform("platform")
	platformRelease := platform.GetOrCreateRelease(semver.MustParse("0.0.0"))
	platformRelease.InstallDir = dataDir
	board := platformRelease.GetOrCreateBoard("boardA")
	board.Properties.Set("upload_port.vid", "0x0000")
	board.Properties.Set("upload_port.pid", "0x0000")
	board = platformRelease.GetOrCreateBoard("boardB")
	board.Properties.Set("upload_port.vid", "0x0000")
	board.Properties.Set("upload_port.pid", "0x0000")

	// Create some Arduino boards with same VID:PID combination as boards created previously
	pack = pmb.GetOrCreatePackage("arduino")
	pack.Maintainer = "Arduino"
	platform = pack.GetOrCreatePlatform("avr")
	platformRelease = platform.GetOrCreateRelease(semver.MustParse("0.0.0"))
	platformRelease.InstallDir = dataDir
	board = platformRelease.GetOrCreateBoard("nessuno")
	board.Properties.Set("upload_port.vid", "0x0000")
	board.Properties.Set("upload_port.pid", "0x0000")
	board = platformRelease.GetOrCreateBoard("assurdo")
	board.Properties.Set("upload_port.vid", "0x0000")
	board.Properties.Set("upload_port.pid", "0x0000")

	idPrefs := properties.NewMap()
	idPrefs.Set("vid", "0x0000")
	idPrefs.Set("pid", "0x0000")

	pm := pmb.Build()
	pme, release := pm.NewExplorer()
	defer release()

	settings := configuration.NewSettings()
	res, err := identify(context.Background(), pme, idPrefs, settings, true)
	require.NoError(t, err)
	require.NotNil(t, res)
	require.Len(t, res, 4)

	// Verify expected sorting
	require.Equal(t, res[0].GetFqbn(), "arduino:avr:assurdo")
	require.Equal(t, res[1].GetFqbn(), "arduino:avr:nessuno")
	require.Equal(t, res[2].GetFqbn(), "packager:platform:boardA")
	require.Equal(t, res[3].GetFqbn(), "packager:platform:boardB")
}
