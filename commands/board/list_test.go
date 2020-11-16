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
	"testing"

	"github.com/arduino/arduino-cli/commands"
	"github.com/arduino/arduino-cli/configuration"
	"github.com/arduino/go-properties-orderedmap"
	"github.com/stretchr/testify/require"
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
	require.Equal(t, "arduino:samd:mkr1000", res[0].FQBN)
	require.Equal(t, "0xf420", res[0].VID)
	require.Equal(t, "0XF069", res[0].PID)

	// wrong vid (too long), wrong pid (not an hex value)
	res, err = apiByVidPid("0xfffff", "0xDEFG")
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
	port := &commands.BoardPort{
		IdentificationPrefs: properties.NewMap(),
	}
	items, err := identifyViaCloudAPI(port)
	require.Equal(t, err, ErrNotFound)
	require.Empty(t, items)
}
