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

package commands

import (
	"context"

	"github.com/arduino/arduino-cli/commands/cmderrors"
	"github.com/arduino/arduino-cli/commands/internal/instances"
	"github.com/arduino/arduino-cli/internal/arduino/httpclient"
	"github.com/arduino/arduino-cli/internal/arduino/libraries/librariesindex"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	"github.com/arduino/go-paths-helper"
	"github.com/sirupsen/logrus"
)

// LibraryDownload executes the download of the library.
// A DownloadProgressCB callback function must be passed to monitor download progress.
func LibraryDownload(ctx context.Context, req *rpc.LibraryDownloadRequest, downloadCB rpc.DownloadProgressCB) (*rpc.LibraryDownloadResponse, error) {
	logrus.Info("Executing `arduino-cli lib download`")

	var downloadsDir *paths.Path
	if pme, release, err := instances.GetPackageManagerExplorer(req.GetInstance()); err != nil {
		return nil, err
	} else {
		downloadsDir = pme.DownloadDir
		release()
	}

	li, err := instances.GetLibrariesIndex(req.GetInstance())
	if err != nil {
		return nil, err
	}

	logrus.Info("Preparing download")

	version, err := ParseVersion(req.GetVersion())
	if err != nil {
		return nil, err
	}

	lib, err := li.FindRelease(req.GetName(), version)
	if err != nil {
		return nil, err
	}

	if err := downloadLibrary(downloadsDir, lib, downloadCB, func(*rpc.TaskProgress) {}, "download"); err != nil {
		return nil, err
	}

	return &rpc.LibraryDownloadResponse{}, nil
}

func downloadLibrary(downloadsDir *paths.Path, libRelease *librariesindex.Release,
	downloadCB rpc.DownloadProgressCB, taskCB rpc.TaskProgressCB, queryParameter string) error {

	taskCB(&rpc.TaskProgress{Name: tr("Downloading %s", libRelease)})
	config, err := httpclient.GetDownloaderConfig()
	if err != nil {
		return &cmderrors.FailedDownloadError{Message: tr("Can't download library"), Cause: err}
	}
	if err := libRelease.Resource.Download(downloadsDir, config, libRelease.String(), downloadCB, queryParameter); err != nil {
		return &cmderrors.FailedDownloadError{Message: tr("Can't download library"), Cause: err}
	}
	taskCB(&rpc.TaskProgress{Completed: true})

	return nil
}
