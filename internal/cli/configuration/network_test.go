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

package configuration_test

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/arduino/arduino-cli/internal/cli/configuration"
	"github.com/stretchr/testify/require"
)

func TestUserAgentHeader(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, r.Header.Get("User-Agent"))
	}))
	defer ts.Close()

	settings := configuration.NewSettings()
	require.NoError(t, settings.Set("network.user_agent_ext", "test-user-agent"))
	client, err := settings.NewHttpClient()
	require.NoError(t, err)

	request, err := http.NewRequest("GET", ts.URL, nil)
	require.NoError(t, err)

	response, err := client.Do(request)
	require.NoError(t, err)

	b, err := io.ReadAll(response.Body)
	require.NoError(t, err)

	fmt.Println("RESPONSE:", string(b))
	require.Contains(t, string(b), "test-user-agent")
}

func TestProxy(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer ts.Close()

	settings := configuration.NewSettings()
	settings.Set("network.proxy", ts.URL)
	client, err := settings.NewHttpClient()
	require.NoError(t, err)

	request, err := http.NewRequest("GET", "http://arduino.cc", nil)
	require.NoError(t, err)

	response, err := client.Do(request)
	require.NoError(t, err)
	require.Equal(t, http.StatusNoContent, response.StatusCode)
}
