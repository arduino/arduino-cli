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

package lib

import (
	"context"
	"fmt"

	"github.com/arduino/arduino-cli/arduino/libraries/librariesindex"
	"github.com/arduino/arduino-cli/arduino/libraries/librariesmanager"
	"github.com/arduino/arduino-cli/commands"
	rpc "github.com/arduino/arduino-cli/rpc/commands"
	"github.com/sirupsen/logrus"
)

// LibraryDownload FIXMEDOC
func LibraryDownload(ctx context.Context, req *rpc.LibraryDownloadReq, downloadCB commands.DownloadProgressCB) (*rpc.LibraryDownloadResp, error) {
	logrus.Info("Executing `arduino lib download`")

	lm := commands.GetLibraryManager(req.GetInstance().GetId())

	logrus.Info("Preparing download")

	lib, err := findLibraryIndexRelease(lm, req)
	if err != nil {
		return nil, fmt.Errorf("looking for library: %s", err)
	}

	if err := downloadLibrary(lm, lib, downloadCB, func(*rpc.TaskProgress) {}); err != nil {
		return nil, err
	}

	return &rpc.LibraryDownloadResp{}, nil
}

func downloadLibrary(lm *librariesmanager.LibrariesManager, libRelease *librariesindex.Release,
	downloadCB commands.DownloadProgressCB, taskCB commands.TaskProgressCB) error {

	taskCB(&rpc.TaskProgress{Name: "Downloading " + libRelease.String()})
	config, err := commands.GetDownloaderConfig()
	if err != nil {
		return err
	}
	if d, err := libRelease.Resource.Download(lm.DownloadsDir, config); err != nil {
		return err
	} else if err := commands.Download(d, libRelease.String(), downloadCB); err != nil {
		return err
	}
	taskCB(&rpc.TaskProgress{Completed: true})

	return nil
}
