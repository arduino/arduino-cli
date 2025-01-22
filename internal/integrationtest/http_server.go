// This file is part of arduino-cli.
//
// Copyright 2022 ARDUINO SA (http://www.arduino.cc/)
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

package integrationtest

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/arduino/go-paths-helper"
	"github.com/stretchr/testify/require"
)

// HTTPServeFile spawn an http server that serve a single file. The server
// is started on the given port. The URL to the file and a cleanup function are returned.
func (env *Environment) HTTPServeFile(port uint16, path *paths.Path, isDaemon bool) *url.URL {
	t := env.T()
	mux := http.NewServeMux()
	mux.HandleFunc("/"+path.Base(), func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, path.String())
		if isDaemon {
			// Test that the user-agent contains metadata from the context when the CLI is in daemon mode
			userAgent := r.Header.Get("User-Agent")
			require.Contains(t, userAgent, "arduino-cli/git-snapshot")
			require.Contains(t, userAgent, "cli-test/0.0.0")
			require.Contains(t, userAgent, "grpc-go")
		}
	})
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
	}

	fileURL, err := url.Parse(fmt.Sprintf("http://127.0.0.1:%d/%s", port, path.Base()))
	require.NoError(t, err)

	go func() {
		err := server.ListenAndServe()
		require.Equal(t, err, http.ErrServerClosed)
	}()

	env.RegisterCleanUpCallback(func() {
		server.Close()
	})

	return fileURL
}

// HTTPServeFileError spawns an http server that serves a single file and responds
// with the given error code.
func (env *Environment) HTTPServeFileError(port uint16, path *paths.Path, code int) *url.URL {
	mux := http.NewServeMux()
	mux.HandleFunc("/"+path.Base(), func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(code)
	})
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
	}

	t := env.T()
	fileURL, err := url.Parse(fmt.Sprintf("http://127.0.0.1:%d/%s", port, path.Base()))
	require.NoError(t, err)

	go func() {
		err := server.ListenAndServe()
		require.Equal(t, err, http.ErrServerClosed)
	}()

	env.RegisterCleanUpCallback(func() {
		server.Close()
	})

	return fileURL
}
