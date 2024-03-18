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
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
)

// LibraryUpgradeAllStreamResponseToCallbackFunction returns a gRPC stream to be used in LibraryUpgradeAll that sends
// all responses to the callback function.
func LibraryUpgradeAllStreamResponseToCallbackFunction(ctx context.Context, downloadCB rpc.DownloadProgressCB, taskCB rpc.TaskProgressCB) rpc.ArduinoCoreService_LibraryUpgradeAllServer {
	return streamResponseToCallback(ctx, func(r *rpc.LibraryUpgradeAllResponse) error {
		if r.GetProgress() != nil {
			downloadCB(r.GetProgress())
		}
		if r.GetTaskProgress() != nil {
			taskCB(r.GetTaskProgress())
		}
		return nil
	})
}

// LibraryUpgradeAll upgrades all the available libraries
func (s *arduinoCoreServerImpl) LibraryUpgradeAll(req *rpc.LibraryUpgradeAllRequest, stream rpc.ArduinoCoreService_LibraryUpgradeAllServer) error {
	ctx := stream.Context()
	syncSend := NewSynchronizedSend(stream.Send)
	downloadCB := func(p *rpc.DownloadProgress) { syncSend.Send(&rpc.LibraryUpgradeAllResponse{Progress: p}) }
	taskCB := func(p *rpc.TaskProgress) { syncSend.Send(&rpc.LibraryUpgradeAllResponse{TaskProgress: p}) }

	li, err := instances.GetLibrariesIndex(req.GetInstance())
	if err != nil {
		return err
	}

	lme, release, err := instances.GetLibraryManagerExplorer(req.GetInstance())
	if err != nil {
		return err
	}
	libsToUpgrade := listLibraries(lme, li, true, false)
	release()

	if err := s.libraryUpgrade(ctx, req.GetInstance(), libsToUpgrade, downloadCB, taskCB); err != nil {
		return err
	}

	err = s.Init(
		&rpc.InitRequest{Instance: req.GetInstance()},
		InitStreamResponseToCallbackFunction(ctx, nil))
	if err != nil {
		return err
	}

	return nil
}

// LibraryUpgradeStreamResponseToCallbackFunction returns a gRPC stream to be used in LibraryUpgrade that sends
// all responses to the callback function.
func LibraryUpgradeStreamResponseToCallbackFunction(ctx context.Context, downloadCB rpc.DownloadProgressCB, taskCB rpc.TaskProgressCB) rpc.ArduinoCoreService_LibraryUpgradeServer {
	return streamResponseToCallback(ctx, func(r *rpc.LibraryUpgradeResponse) error {
		if r.GetProgress() != nil {
			downloadCB(r.GetProgress())
		}
		if r.GetTaskProgress() != nil {
			taskCB(r.GetTaskProgress())
		}
		return nil
	})
}

// LibraryUpgrade upgrades a library
func (s *arduinoCoreServerImpl) LibraryUpgrade(req *rpc.LibraryUpgradeRequest, stream rpc.ArduinoCoreService_LibraryUpgradeServer) error {
	ctx := stream.Context()
	syncSend := NewSynchronizedSend(stream.Send)
	downloadCB := func(p *rpc.DownloadProgress) { syncSend.Send(&rpc.LibraryUpgradeResponse{Progress: p}) }
	taskCB := func(p *rpc.TaskProgress) { syncSend.Send(&rpc.LibraryUpgradeResponse{TaskProgress: p}) }

	li, err := instances.GetLibrariesIndex(req.GetInstance())
	if err != nil {
		return err
	}

	lme, release, err := instances.GetLibraryManagerExplorer(req.GetInstance())
	if err != nil {
		return err
	}
	libs := listLibraries(lme, li, false, false)
	release()

	// Get the library to upgrade
	name := req.GetName()
	lib := filterByName(libs, name)
	if lib == nil {
		// library not installed...
		return &cmderrors.LibraryNotFoundError{Library: name}
	}
	if lib.Available == nil {
		taskCB(&rpc.TaskProgress{Message: tr("Library %s is already at the latest version", name), Completed: true})
		return nil
	}

	// Install update
	return s.libraryUpgrade(ctx, req.GetInstance(), []*installedLib{lib}, downloadCB, taskCB)
}

func (s *arduinoCoreServerImpl) libraryUpgrade(ctx context.Context, instance *rpc.Instance, libs []*installedLib, downloadCB rpc.DownloadProgressCB, taskCB rpc.TaskProgressCB) error {
	for _, lib := range libs {
		libInstallReq := &rpc.LibraryInstallRequest{
			Instance:    instance,
			Name:        lib.Library.Name,
			Version:     "",
			NoDeps:      false,
			NoOverwrite: false,
		}
		stream := LibraryInstallStreamResponseToCallbackFunction(ctx, downloadCB, taskCB)
		if err := s.LibraryInstall(libInstallReq, stream); err != nil {
			return err
		}
	}

	return nil
}

func filterByName(libs []*installedLib, name string) *installedLib {
	for _, lib := range libs {
		if lib.Library.Name == name {
			return lib
		}
	}
	return nil
}
