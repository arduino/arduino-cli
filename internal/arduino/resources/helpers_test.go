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

package resources

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/arduino/arduino-cli/internal/arduino/httpclient"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	"github.com/arduino/go-paths-helper"
	"github.com/stretchr/testify/require"
	"go.bug.st/downloader/v2"
)

type EchoHandler struct{}

// EchoHandler echos back the request as a response if used as http handler
func (h *EchoHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	request.Write(writer)
}

func TestDownloadApplyUserAgentHeaderUsingConfig(t *testing.T) {
	goldUserAgentValue := "arduino-cli/0.0.0-test.preview (amd64; linux; go1.12.4) Commit:deadbeef/Build:2019-06-12 11:11:11.111"
	goldUserAgentString := "User-Agent: " + goldUserAgentValue

	tmp, err := paths.MkTempDir("", "")
	require.NoError(t, err)
	defer tmp.RemoveAll()

	// startup echo server
	srv := httptest.NewServer(&EchoHandler{})
	defer srv.Close()

	r := &DownloadResource{
		ArchiveFileName: "echo.txt",
		CachePath:       "cache",
		URL:             srv.URL,
	}

	httpClient := httpclient.NewWithConfig(&httpclient.Config{UserAgent: goldUserAgentValue})

	err = r.Download(tmp, &downloader.Config{HttpClient: *httpClient}, "", func(progress *rpc.DownloadProgress) {}, "")
	require.NoError(t, err)

	// leverage the download helper to download the echo for the request made by the downloader itself
	//
	// expect something like:
	//    GET /echo HTTP/1.1
	//    Host: 127.0.0.1:64999
	//    User-Agent: arduino-cli/0.0.0-test.preview  (amd64; linux; go1.12.4) Commit:deadbeef/Build:2019-06-12 11:11:11.111
	//    Accept-Encoding: gzip

	b, err := os.ReadFile(tmp.String() + "/cache/echo.txt") // just pass the file name
	require.NoError(t, err)

	requestLines := strings.Split(string(b), "\r\n")
	userAgentHeaderString := ""
	for _, line := range requestLines {
		if strings.Contains(line, "User-Agent: ") {
			userAgentHeaderString = line
		}
	}
	require.Equal(t, goldUserAgentString, userAgentHeaderString)

}
