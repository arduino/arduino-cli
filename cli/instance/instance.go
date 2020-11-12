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

package instance

import (
	"context"
	"fmt"
	"strings"

	"github.com/arduino/arduino-cli/cli/output"
	"github.com/arduino/arduino-cli/commands"
	"github.com/arduino/arduino-cli/configuration"
	rpc "github.com/arduino/arduino-cli/rpc/commands"
	"github.com/arduino/go-paths-helper"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// CreateInstanceIgnorePlatformIndexErrors creates and return an instance of the
// Arduino Core Engine, but won't stop on platforms index loading errors.
func CreateInstanceIgnorePlatformIndexErrors() *rpc.Instance {
	i, _ := getInitResponse()
	return i.GetInstance()
}

// CreateInstance creates and return an instance of the Arduino Core engine
func CreateInstance() (*rpc.Instance, error) {
	resp, err := getInitResponse()
	if err != nil {
		return nil, err
	}

	return resp.GetInstance(), checkPlatformErrors(resp)
}

func getInitResponse() (*rpc.InitResp, error) {
	// invoke Init()
	resp, err := commands.Init(context.Background(), &rpc.InitReq{}, output.ProgressBar(), output.TaskProgress())

	// Init() failed
	if err != nil {
		return nil, errors.Wrap(err, "creating instance")
	}

	// Gets the data directory to verify if library_index.json and package_index.json exist
	dataDir := paths.New(configuration.Settings.GetString("directories.data"))

	// The library_index.json file doesn't exists, that means the CLI is run for the first time
	// so we proceed with the first update that downloads the file
	libraryIndex := dataDir.Join("library_index.json")
	if libraryIndex.NotExist() {
		logrus.Warnf("There were errors loading the library index, trying again...")

		// update all indexes
		err := commands.UpdateLibrariesIndex(context.Background(),
			&rpc.UpdateLibrariesIndexReq{Instance: resp.GetInstance()}, output.ProgressBar())
		if err != nil {
			return nil, errors.Wrap(err, "updating the library index")
		}

		// rescan libraries
		rescanResp, err := commands.Rescan(resp.GetInstance().GetId())
		if err != nil {
			return nil, errors.Wrap(err, "during rescan")
		}

		// errors persist
		if rescanResp.GetLibrariesIndexError() != "" {
			return nil, errors.New("still errors after rescan: " + rescanResp.GetLibrariesIndexError())
		}

		// succeeded, copy over PlatformsIndexErrors in case errors occurred
		// during rescan
		resp.LibrariesIndexError = ""
		resp.PlatformsIndexErrors = rescanResp.PlatformsIndexErrors
	}

	// The package_index.json file doesn't exists, that means the CLI is run for the first time,
	// similarly to the library update we download that file and all the other package indexes
	// from additional_urls
	packageIndex := dataDir.Join("package_index.json")
	if packageIndex.NotExist() {
		// update platform index
		_, err := commands.UpdateIndex(context.Background(),
			&rpc.UpdateIndexReq{Instance: resp.GetInstance()}, output.ProgressBar())
		if err != nil {
			return nil, errors.Wrap(err, "updating the core index")
		}

		// rescan
		rescanResp, err := commands.Rescan(resp.GetInstance().GetId())
		if err != nil {
			return nil, errors.Wrap(err, "during rescan")
		}

		// errors persist
		if rescanResp.GetPlatformsIndexErrors() != nil {
			for _, err := range rescanResp.GetPlatformsIndexErrors() {
				logrus.Errorf("Still errors after rescan: %v", err)
			}
		}

		// succeeded, copy over PlatformsIndexErrors in case errors occurred
		// during rescan
		resp.PlatformsIndexErrors = rescanResp.PlatformsIndexErrors
	}

	return resp, nil
}

func checkPlatformErrors(resp *rpc.InitResp) error {
	// Init() and/or rescan succeeded, but there were errors loading platform indexes
	if resp.GetPlatformsIndexErrors() != nil {
		// log each error
		for _, err := range resp.GetPlatformsIndexErrors() {
			logrus.Errorf("Error loading platform index: %v", err)
		}
		return fmt.Errorf("error loading platform index: \n%v", strings.Join(resp.GetPlatformsIndexErrors(), "\n"))
	}

	return nil
}
