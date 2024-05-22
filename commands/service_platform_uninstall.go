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

// PlatformUninstallStreamResponseToCallbackFunction returns a gRPC stream to be used in PlatformUninstall that sends
// all responses to the callback function.
func PlatformUninstallStreamResponseToCallbackFunction(ctx context.Context, taskCB rpc.TaskProgressCB) rpc.ArduinoCoreService_PlatformUninstallServer {
	return streamResponseToCallback(ctx, func(r *rpc.PlatformUninstallResponse) error {
		if r.GetTaskProgress() != nil {
			taskCB(r.GetTaskProgress())
		}
		return nil
	})
}

// PlatformUninstall uninstalls a platform package
func (s *arduinoCoreServerImpl) PlatformUninstall(req *rpc.PlatformUninstallRequest, stream rpc.ArduinoCoreService_PlatformUninstallServer) error {
	syncSend := NewSynchronizedSend(stream.Send)
	ctx := stream.Context()
	taskCB := func(p *rpc.TaskProgress) {
		syncSend.Send(&rpc.PlatformUninstallResponse{
			Message: &rpc.PlatformUninstallResponse_TaskProgress{
				TaskProgress: p,
			},
		})
	}
	if err := platformUninstall(ctx, req, taskCB); err != nil {
		return err
	}
	if err := s.Init(&rpc.InitRequest{Instance: req.GetInstance()}, InitStreamResponseToCallbackFunction(ctx, nil)); err != nil {
		return err
	}
	return syncSend.Send(&rpc.PlatformUninstallResponse{
		Message: &rpc.PlatformUninstallResponse_Result_{
			Result: &rpc.PlatformUninstallResponse_Result{},
		},
	})
}

// platformUninstall is the implementation of platform unistaller
func platformUninstall(_ context.Context, req *rpc.PlatformUninstallRequest, taskCB rpc.TaskProgressCB) error {
	pme, release, err := instances.GetPackageManagerExplorer(req.GetInstance())
	if err != nil {
		return &cmderrors.InvalidInstanceError{}
	}
	defer release()

	ref := &packagemanager.PlatformReference{
		Package:              req.GetPlatformPackage(),
		PlatformArchitecture: req.GetArchitecture(),
	}
	if ref.PlatformVersion == nil {
		platform := pme.FindPlatform(ref)
		if platform == nil {
			return &cmderrors.PlatformNotFoundError{Platform: ref.String()}
		}
		platformRelease := pme.GetInstalledPlatformRelease(platform)
		if platformRelease == nil {
			return &cmderrors.PlatformNotFoundError{Platform: ref.String()}
		}
		ref.PlatformVersion = platformRelease.Version
	}

	platform, tools, err := pme.FindPlatformReleaseDependencies(ref)
	if err != nil {
		return &cmderrors.NotFoundError{Message: i18n.Tr("Can't find dependencies for platform %s", ref), Cause: err}
	}

	// TODO: pass context
	if err := pme.UninstallPlatform(platform, taskCB, req.GetSkipPreUninstall()); err != nil {
		return err
	}

	for _, tool := range tools {
		if !pme.IsToolRequired(tool) {
			taskCB(&rpc.TaskProgress{Name: i18n.Tr("Uninstalling %s, tool is no more required", tool)})
			pme.UninstallTool(tool, taskCB, req.GetSkipPreUninstall())
		}
	}

	return nil
}
