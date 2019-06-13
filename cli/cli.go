/*
 * This file is part of arduino-cli.
 *
 * Copyright 2018 ARDUINO SA (http://www.arduino.cc/)
 *
 * This software is released under the GNU General Public License version 3,
 * which covers the main part of arduino-cli.
 * The terms of this license can be found at:
 * https://www.gnu.org/licenses/gpl-3.0.en.html
 *
 * You can be released from the requirements of the above licenses by purchasing
 * a commercial license. Buying such a license is mandatory if you want to modify or
 * otherwise use the software for commercial activities involving the Arduino
 * software without disclosing the source code of your own applications. To purchase
 * a commercial license, send an email to license@arduino.cc.
 */

package cli

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"runtime"

	"github.com/arduino/arduino-cli/arduino/cores/packagemanager"
	"github.com/arduino/arduino-cli/arduino/libraries/librariesmanager"
	"github.com/arduino/arduino-cli/commands"
	"github.com/arduino/arduino-cli/common/formatter"
	"github.com/arduino/arduino-cli/configs"
	"github.com/arduino/arduino-cli/rpc"
	"github.com/arduino/arduino-cli/version"
	"github.com/arduino/go-paths-helper"
	"github.com/sirupsen/logrus"
)

// Error codes to be used for os.Exit().
const (
	_          = iota // 0 is not a valid exit error code
	ErrGeneric        // 1 is the reserved "catchall" code in Unix
	_                 // 2 is reserved in Unix
	ErrNoConfigFile
	ErrBadCall
	ErrNetwork
	// ErrCoreConfig represents an error in the cli core config, for example some basic
	// files shipped with the installation are missing, or cannot create or get basic
	// directories vital for the CLI to work.
	ErrCoreConfig
	ErrBadArgument
)

// appName is the command line name of the Arduino CLI executable on the user system (users may change it)
var appName = filepath.Base(os.Args[0])

// VersionInfo contains all info injected during build
var VersionInfo = version.NewInfo(appName)

var HTTPClientHeader = getHTTPClientHeader()

// ErrLogrus represents the logrus instance, which has the role to
// log all non info messages.
var ErrLogrus = logrus.New()

// GlobalFlags represents flags available in all the program.
var GlobalFlags struct {
	Debug      bool // If true, dump debug output to stderr.
	OutputJSON bool // true output in JSON, false output as Text
}

var Config *configs.Configuration

func packageManagerInitReq() *rpc.InitReq {
	urls := []string{}
	for _, URL := range Config.BoardManagerAdditionalUrls {
		urls = append(urls, URL.String())
	}

	conf := &rpc.Configuration{}
	conf.DataDir = Config.DataDir.String()
	conf.DownloadsDir = Config.DownloadsDir().String()
	conf.BoardManagerAdditionalUrls = urls
	if Config.SketchbookDir != nil {
		conf.SketchbookDir = Config.SketchbookDir.String()
	}

	return &rpc.InitReq{Configuration: conf}
}

func getHTTPClientHeader() http.Header {
	userAgentValue := fmt.Sprintf("%s/%s (%s; %s; %s) Commit:%s/Build:%s", VersionInfo.Application,
		VersionInfo.VersionString, runtime.GOARCH, runtime.GOOS, runtime.Version(), VersionInfo.Commit, VersionInfo.BuildDate)
	downloaderHeaders := http.Header{"User-Agent": []string{userAgentValue}}
	return downloaderHeaders
}

func InitInstance() *rpc.InitResp {
	logrus.Info("Initializing package manager")
	req := packageManagerInitReq()

	resp, err := commands.Init(context.Background(), req, OutputProgressBar(), OutputTaskProgress(), HTTPClientHeader)
	if err != nil {
		formatter.PrintError(err, "Error initializing package manager")
		os.Exit(ErrGeneric)
	}
	if resp.GetLibrariesIndexError() != "" {
		commands.UpdateLibrariesIndex(context.Background(),
			&rpc.UpdateLibrariesIndexReq{Instance: resp.GetInstance()}, OutputProgressBar())
		rescResp, err := commands.Rescan(context.Background(), &rpc.RescanReq{Instance: resp.GetInstance()})
		if rescResp.GetLibrariesIndexError() != "" {
			formatter.PrintErrorMessage("Error loading library index: " + rescResp.GetLibrariesIndexError())
			os.Exit(ErrGeneric)
		}
		if err != nil {
			formatter.PrintError(err, "Error loading library index")
			os.Exit(ErrGeneric)
		}
		resp.LibrariesIndexError = rescResp.LibrariesIndexError
		resp.PlatformsIndexErrors = rescResp.PlatformsIndexErrors
	}
	return resp
}

// CreateInstance creates and return an instance of the Arduino Core engine
func CreateInstance() *rpc.Instance {
	resp := InitInstance()
	if resp.GetPlatformsIndexErrors() != nil {
		for _, err := range resp.GetPlatformsIndexErrors() {
			formatter.PrintError(errors.New(err), "Error loading index")
		}
		formatter.PrintErrorMessage("Launch '" + VersionInfo.Application + " core update-index' to fix or download indexes.")
		os.Exit(ErrGeneric)
	}
	return resp.GetInstance()
}

// CreateInstaceIgnorePlatformIndexErrors creates and return an instance of the
// Arduino Core Engine, but won't stop on platforms index loading errors.
func CreateInstaceIgnorePlatformIndexErrors() *rpc.Instance {
	return InitInstance().GetInstance()
}

// InitPackageAndLibraryManager initializes the PackageManager and the
// LibaryManager with the default configuration. (DEPRECATED)
func InitPackageAndLibraryManager() (*packagemanager.PackageManager, *librariesmanager.LibrariesManager) {
	resp := InitInstance()
	return commands.GetPackageManager(resp), commands.GetLibraryManager(resp)
}

// InitSketchPath returns sketchPath if specified or the current working
// directory if sketchPath is nil.
func InitSketchPath(sketchPath *paths.Path) *paths.Path {
	if sketchPath != nil {
		return sketchPath
	}

	wd, err := paths.Getwd()
	if err != nil {
		formatter.PrintError(err, "Couldn't get current working directory")
		os.Exit(ErrGeneric)
	}
	logrus.Infof("Reading sketch from dir: %s", wd)
	return wd
}
