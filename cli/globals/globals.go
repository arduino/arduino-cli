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
// a commercial license. Buying such a license is mandatory if you want to modify or
// otherwise use the software for commercial activities involving the Arduino
// software without disclosing the source code of your own applications. To purchase
// a commercial license, send an email to license@arduino.cc.

package globals

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"runtime"

	"github.com/arduino/arduino-cli/version"
)

var (
	// VersionInfo contains all info injected during build
	VersionInfo = version.NewInfo(filepath.Base(os.Args[0]))
	// DefaultIndexURL is the default index url
	DefaultIndexURL = "https://downloads.arduino.cc/packages/package_index.json"
)

// NewHTTPClientHeader returns the http.Header object that must be used by the clients inside the downloaders
func NewHTTPClientHeader() http.Header {
	userAgentValue := fmt.Sprintf("%s/%s (%s; %s; %s) Commit:%s", VersionInfo.Application,
		VersionInfo.VersionString, runtime.GOARCH, runtime.GOOS, runtime.Version(), VersionInfo.Commit)
	return http.Header{"User-Agent": []string{userAgentValue}}
}
