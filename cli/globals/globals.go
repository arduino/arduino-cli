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

	"github.com/arduino/arduino-cli/configs"
	"github.com/arduino/arduino-cli/version"
)

var (
	// Debug determines whether to dump debug output to stderr or not
	Debug bool
	// HTTPClientHeader is the object that will be propagated to configure the clients inside the downloaders
	HTTPClientHeader = getHTTPClientHeader()
	// VersionInfo contains all info injected during build
	VersionInfo = version.NewInfo(filepath.Base(os.Args[0]))
	// Config FIXMEDOC
	Config *configs.Configuration
	// YAMLConfigFile contains the path to the config file
	YAMLConfigFile string
	// AdditionalUrls contains the list of additional urls the boards manager can use
	AdditionalUrls []string
	// LogLevel is temporarily exported because the compile command will
	// forward this information to the underlying legacy package
	LogLevel string
)

func getHTTPClientHeader() http.Header {
	userAgentValue := fmt.Sprintf("%s/%s (%s; %s; %s) Commit:%s", VersionInfo.Application,
		VersionInfo.VersionString, runtime.GOARCH, runtime.GOOS, runtime.Version(), VersionInfo.Commit)
	downloaderHeaders := http.Header{"User-Agent": []string{userAgentValue}}
	return downloaderHeaders
}
