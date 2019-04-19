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

package lib

import (
	"context"
	"fmt"

	"github.com/arduino/arduino-cli/arduino/libraries/librariesindex"
	"github.com/arduino/arduino-cli/arduino/libraries/librariesmanager"
	"github.com/arduino/arduino-cli/commands"
	"github.com/arduino/arduino-cli/rpc"
	"github.com/sirupsen/logrus"
	semver "go.bug.st/relaxed-semver"
)

func LibraryDownload(ctx context.Context, req *rpc.LibraryDownloadReq, downloadCB commands.DownloadProgressCB) (*rpc.LibraryDownloadResp, error) {
	logrus.Info("Executing `arduino lib download`")

	lm := commands.GetLibraryManager(req)

	logrus.Info("Preparing download")

	var version *semver.Version
	if req.GetVersion() != "" {
		if v, err := semver.Parse(req.GetVersion()); err == nil {
			version = v
		} else {
			return nil, fmt.Errorf("invalid version: %s", err)
		}
	}

	ref := &librariesindex.Reference{Name: req.GetName(), Version: version}
	lib := lm.Index.FindRelease(ref)
	if lib == nil {
		return nil, fmt.Errorf("library %s not found", ref.String())
	}

	if err := downloadLibrary(lm, lib, downloadCB); err != nil {
		return nil, err
	}

	return &rpc.LibraryDownloadResp{}, nil
}

func downloadLibrary(lm *librariesmanager.LibrariesManager, libRelease *librariesindex.Release, downloadCB commands.DownloadProgressCB) error {
	d, err := libRelease.Resource.Download(lm.DownloadsDir)
	if err != nil {
		return fmt.Errorf("download error: %s", err)
	}
	return commands.Download(d, libRelease.String(), downloadCB)
}
