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
	"github.com/arduino/arduino-cli/internal/arduino/libraries/librariesindex"
	"github.com/arduino/arduino-cli/internal/cli/configuration"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	"github.com/arduino/go-paths-helper"
)

// LibraryDownloadStreamResponseToCallbackFunction returns a gRPC stream to be used in LibraryDownload that sends
// all responses to the callback function.
func LibraryDownloadStreamResponseToCallbackFunction(ctx context.Context, downloadCB rpc.DownloadProgressCB) rpc.ArduinoCoreService_LibraryDownloadServer {
	return streamResponseToCallback(ctx, func(r *rpc.LibraryDownloadResponse) error {
		if r.GetProgress() != nil {
			downloadCB(r.GetProgress())
		}
		return nil
	})
}

// LibraryDownload downloads a library
func (s *arduinoCoreServerImpl) LibraryDownload(req *rpc.LibraryDownloadRequest, stream rpc.ArduinoCoreService_LibraryDownloadServer) error {
	syncSend := NewSynchronizedSend(stream.Send)
	ctx := stream.Context()
	downloadCB := func(p *rpc.DownloadProgress) {
		syncSend.Send(&rpc.LibraryDownloadResponse{
			Message: &rpc.LibraryDownloadResponse_Progress{Progress: p},
		})
	}

	var downloadsDir *paths.Path
	if pme, release, err := instances.GetPackageManagerExplorer(req.GetInstance()); err != nil {
		return err
	} else {
		downloadsDir = pme.DownloadDir
		release()
	}

	li, err := instances.GetLibrariesIndex(req.GetInstance())
	if err != nil {
		return err
	}

	version, err := parseVersion(req.GetVersion())
	if err != nil {
		return err
	}

	lib, err := li.FindRelease(req.GetName(), version)
	if err != nil {
		return err
	}

	if err := downloadLibrary(ctx, downloadsDir, lib, downloadCB, func(*rpc.TaskProgress) {}, "download", s.settings); err != nil {
		return err
	}

	return syncSend.Send(&rpc.LibraryDownloadResponse{
		Message: &rpc.LibraryDownloadResponse_Result_{
			Result: &rpc.LibraryDownloadResponse_Result{},
		},
	})
}

func downloadLibrary(ctx context.Context, downloadsDir *paths.Path, libRelease *librariesindex.Release,
	downloadCB rpc.DownloadProgressCB, taskCB rpc.TaskProgressCB, queryParameter string, settings *configuration.Settings) error {

	taskCB(&rpc.TaskProgress{Name: tr("Downloading %s", libRelease)})
	config, err := settings.DownloaderConfig()
	if err != nil {
		return &cmderrors.FailedDownloadError{Message: tr("Can't download library"), Cause: err}
	}
	if err := libRelease.Resource.Download(ctx, downloadsDir, config, libRelease.String(), downloadCB, queryParameter); err != nil {
		return &cmderrors.FailedDownloadError{Message: tr("Can't download library"), Cause: err}
	}
	taskCB(&rpc.TaskProgress{Completed: true})

	return nil
}
