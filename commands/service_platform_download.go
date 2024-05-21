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
	"github.com/arduino/arduino-cli/internal/arduino/cores/packagemanager"
	"github.com/arduino/arduino-cli/internal/i18n"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
)

var tr = i18n.Tr

// PlatformDownloadStreamResponseToCallbackFunction returns a gRPC stream to be used in PlatformDownload that sends
// all responses to the callback function.
func PlatformDownloadStreamResponseToCallbackFunction(ctx context.Context, downloadCB rpc.DownloadProgressCB) rpc.ArduinoCoreService_PlatformDownloadServer {
	return streamResponseToCallback(ctx, func(r *rpc.PlatformDownloadResponse) error {
		if r.GetProgress() != nil {
			downloadCB(r.GetProgress())
		}
		return nil
	})
}

// PlatformDownload downloads a platform package
func (s *arduinoCoreServerImpl) PlatformDownload(req *rpc.PlatformDownloadRequest, stream rpc.ArduinoCoreService_PlatformDownloadServer) error {
	ctx := stream.Context()
	syncSend := NewSynchronizedSend(stream.Send)

	pme, release, err := instances.GetPackageManagerExplorer(req.GetInstance())
	if err != nil {
		return err
	}
	defer release()

	version, err := parseVersion(req.GetVersion())
	if err != nil {
		return &cmderrors.InvalidVersionError{Cause: err}
	}

	ref := &packagemanager.PlatformReference{
		Package:              req.GetPlatformPackage(),
		PlatformArchitecture: req.GetArchitecture(),
		PlatformVersion:      version,
	}
	platform, tools, err := pme.FindPlatformReleaseDependencies(ref)
	if err != nil {
		return &cmderrors.PlatformNotFoundError{Platform: ref.String(), Cause: err}
	}

	downloadCB := func(p *rpc.DownloadProgress) {
		syncSend.Send(&rpc.PlatformDownloadResponse{
			Message: &rpc.PlatformDownloadResponse_Progress{
				Progress: p,
			},
		})
	}

	if err := pme.DownloadPlatformRelease(ctx, platform, downloadCB); err != nil {
		return err
	}

	for _, tool := range tools {
		if err := pme.DownloadToolRelease(ctx, tool, downloadCB); err != nil {
			return err
		}
	}

	return syncSend.Send(&rpc.PlatformDownloadResponse{Message: &rpc.PlatformDownloadResponse_Result_{
		Result: &rpc.PlatformDownloadResponse_Result{},
	}})
}
