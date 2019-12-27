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

package instance

import (
	"context"
	"os"

	"github.com/arduino/arduino-cli/cli/errorcodes"
	"github.com/arduino/arduino-cli/cli/feedback"
	"github.com/arduino/arduino-cli/cli/globals"
	"github.com/arduino/arduino-cli/cli/output"
	"github.com/arduino/arduino-cli/commands"
	rpc "github.com/arduino/arduino-cli/rpc/commands"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// CreateInstanceIgnorePlatformIndexErrors creates and return an instance of the
// Arduino Core Engine, but won't stop on platforms index loading errors.
func CreateInstanceIgnorePlatformIndexErrors() *rpc.Instance {
	i, _ := getInitResponse()
	return i.GetInstance()
}

// CreateInstance creates and return an instance of the Arduino Core engine
func CreateInstance() *rpc.Instance {
	resp := initInstance()
	if resp.GetPlatformsIndexErrors() != nil {
		for _, err := range resp.GetPlatformsIndexErrors() {
			feedback.Errorf("Error loading index: %v", err)
		}
		feedback.Errorf("Launch '%s core update-index' to fix or download indexes.", os.Args[0])
		os.Exit(errorcodes.ErrGeneric)
	}
	return resp.GetInstance()
}

func initInstance() *rpc.InitResp {
	logrus.Info("Initializing package manager")
	req := packageManagerInitReq()

	resp, err := commands.Init(context.Background(), req, output.ProgressBar(), output.TaskProgress(), globals.NewHTTPClientHeader())
	if err != nil {
		feedback.Errorf("Error initializing package manager: %v", err)
		os.Exit(errorcodes.ErrGeneric)
	}
	if resp.GetLibrariesIndexError() != "" {
		commands.UpdateLibrariesIndex(context.Background(),
			&rpc.UpdateLibrariesIndexReq{Instance: resp.GetInstance()}, output.ProgressBar())
		rescResp, err := commands.Rescan(resp.GetInstance().GetId())
		if rescResp.GetLibrariesIndexError() != "" {
			feedback.Errorf("Error loading library index: %v", rescResp.GetLibrariesIndexError())
			os.Exit(errorcodes.ErrGeneric)
		}
		if err != nil {
			feedback.Errorf("Error loading library index: %v", err)
			os.Exit(errorcodes.ErrGeneric)
		}
		resp.LibrariesIndexError = rescResp.LibrariesIndexError
		resp.PlatformsIndexErrors = rescResp.PlatformsIndexErrors
	}
	return resp
}

func packageManagerInitReq() *rpc.InitReq {
	urls := []string{globals.DefaultIndexURL}

	for _, URL := range viper.GetStringSlice("board_manager.additional_urls") {
		urls = append(urls, URL)
	}

	conf := &rpc.Configuration{}
	conf.DataDir = viper.GetString("directories.Data")
	conf.DownloadsDir = viper.GetString("directories.Downloads")
	conf.BoardManagerAdditionalUrls = urls
	conf.SketchbookDir = viper.GetString("directories.User")

	return &rpc.InitReq{Configuration: conf}
}
