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
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/arduino/arduino-cli/arduino/cores/packagemanager"
	"github.com/arduino/arduino-cli/arduino/discovery"
	"github.com/arduino/arduino-cli/configuration"
	"github.com/arduino/go-paths-helper"
	"github.com/arduino/go-properties-orderedmap"
	"github.com/stretchr/testify/require"
	semver "go.bug.st/relaxed-semver"
)

func init() {
	configuration.Settings = configuration.Init("")
}

func TestGetByVidPid(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, `
{
	"architecture": "samd",
	"fqbn": "arduino:samd:mkr1000",
	"href": "/v3/boards/arduino:samd:mkr1000",
	"id": "mkr1000",
	"name": "Arduino/Genuino MKR1000",
	"package": "arduino",
	"plan": "create-free"
}
		`)
	}))
	defer ts.Close()

	vidPidURL = ts.URL
	res, err := apiByVidPid("0xf420", "0XF069")
	require.Nil(t, err)
	require.Len(t, res, 1)
	require.Equal(t, "Arduino/Genuino MKR1000", res[0].Name)
	require.Equal(t, "arduino:samd:mkr1000", res[0].Fqbn)

	// wrong vid (too long), wrong pid (not an hex value)
	_, err = apiByVidPid("0xfffff", "0xDEFG")
	require.NotNil(t, err)
}

func TestGetByVidPidNotFound(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	vidPidURL = ts.URL
	res, err := apiByVidPid("0x0420", "0x0069")
	require.NotNil(t, err)
	require.Equal(t, "board not found", err.Error())
	require.Len(t, res, 0)
}

func TestGetByVidPid5xx(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("500 - Ooooops!"))
	}))
	defer ts.Close()

	vidPidURL = ts.URL
	res, err := apiByVidPid("0x0420", "0x0069")
	require.NotNil(t, err)
	require.Equal(t, "the server responded with status 500 Internal Server Error", err.Error())
	require.Len(t, res, 0)
}

func TestGetByVidPidMalformedResponse(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "{}")
	}))
	defer ts.Close()

	vidPidURL = ts.URL
	res, err := apiByVidPid("0x0420", "0x0069")
	require.NotNil(t, err)
	require.Equal(t, "wrong format in server response", err.Error())
	require.Len(t, res, 0)
}

func TestBoardDetectionViaAPIWithNonUSBPort(t *testing.T) {
	port := &discovery.Port{
		Properties: properties.NewMap(),
	}
	items, err := identifyViaCloudAPI(port)
	require.ErrorIs(t, err, ErrNotFound)
	require.Empty(t, items)
}

func TestBoardIdentifySorting(t *testing.T) {
	dataDir := paths.TempDir().Join("test", "data_dir")
	os.Setenv("ARDUINO_DATA_DIR", dataDir.String())
	dataDir.MkdirAll()
	defer paths.TempDir().Join("test").RemoveAll()

	// We don't really care about the paths in this case
	pmb := packagemanager.NewBuilder(dataDir, dataDir, dataDir, dataDir, "test")

	// Create some boards with identical VID:PID combination
	pack := pmb.Packages.GetOrCreatePackage("packager")
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
	pack = pmb.Packages.GetOrCreatePackage("arduino")
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

	res, err := identify(pmb.Build(), &discovery.Port{Properties: idPrefs})
	require.NoError(t, err)
	require.NotNil(t, res)
	require.Len(t, res, 4)

	// Verify expected sorting
	require.Equal(t, res[0].Fqbn, "arduino:avr:assurdo")
	require.Equal(t, res[1].Fqbn, "arduino:avr:nessuno")
	require.Equal(t, res[2].Fqbn, "packager:platform:boardA")
	require.Equal(t, res[3].Fqbn, "packager:platform:boardB")
}
