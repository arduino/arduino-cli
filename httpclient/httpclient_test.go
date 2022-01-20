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

package httpclient

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUserAgentHeader(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, r.Header.Get("User-Agent"))
	}))
	defer ts.Close()

	Init(&Config{
		UserAgent: "test-user-agent",
	})
	client := Get()

	request, err := http.NewRequest("GET", ts.URL, nil)
	require.NoError(t, err)

	response, err := client.Do(request)
	require.NoError(t, err)

	b, err := ioutil.ReadAll(response.Body)
	require.NoError(t, err)

	require.Equal(t, "test-user-agent", string(b))
}

func TestProxy(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer ts.Close()

	proxyURL, err := url.Parse(ts.URL)
	require.NoError(t, err)

	Init(&Config{
		Proxy: proxyURL,
	})
	client := Get()

	request, err := http.NewRequest("GET", "http://arduino.cc", nil)
	require.NoError(t, err)

	response, err := client.Do(request)
	require.NoError(t, err)
	require.Equal(t, http.StatusNoContent, response.StatusCode)
}
