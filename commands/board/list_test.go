// This file is part of arduino-cli.
//
// Copyright 2019 ARDUINO SA (http://www.arduino.cc/)
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

	"github.com/stretchr/testify/require"
)

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

	res, err := apiByVidPid(ts.URL)
	require.Nil(t, err)
	require.Len(t, res, 1)
	require.Equal(t, "Arduino/Genuino MKR1000", res[0].Name)
	require.Equal(t, "arduino:samd:mkr1000", res[0].FQBN)

	// wrong url
	res, err = apiByVidPid("http://0.0.0.0")
	require.NotNil(t, err)
}

func TestGetByVidPidNotFound(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	res, err := apiByVidPid(ts.URL)
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

	res, err := apiByVidPid(ts.URL)
	require.NotNil(t, err)
	require.Equal(t, "the server responded with status 500 Internal Server Error", err.Error())
	require.Len(t, res, 0)
}

func TestGetByVidPidMalformedResponse(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "{}")
	}))
	defer ts.Close()

	res, err := apiByVidPid(ts.URL)
	require.NotNil(t, err)
	require.Equal(t, "wrong format in server response", err.Error())
	require.Len(t, res, 0)
}
