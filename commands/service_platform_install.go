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
	"fmt"

	"github.com/arduino/arduino-cli/commands/cmderrors"
	"github.com/arduino/arduino-cli/commands/internal/instances"
	"github.com/arduino/arduino-cli/internal/arduino/cores/packagemanager"
	"github.com/arduino/arduino-cli/internal/arduino/resources"
	"github.com/arduino/arduino-cli/internal/i18n"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
)

// UpdateIndexStreamResponseToCallbackFunction returns a gRPC stream to be used in PlatformInstall that sends
// all responses to the callback function.
func PlatformInstallStreamResponseToCallbackFunction(ctx context.Context, downloadCB rpc.DownloadProgressCB, taskCB rpc.TaskProgressCB) rpc.ArduinoCoreService_PlatformInstallServer {
	return streamResponseToCallback(ctx, func(r *rpc.PlatformInstallResponse) error {
		if r.GetProgress() != nil {
			downloadCB(r.GetProgress())
		}
		if r.GetTaskProgress() != nil {
			taskCB(r.GetTaskProgress())
		}
		return nil
	})
}

// PlatformInstall installs a platform package
func (s *arduinoCoreServerImpl) PlatformInstall(req *rpc.PlatformInstallRequest, stream rpc.ArduinoCoreService_PlatformInstallServer) error {
	ctx := stream.Context()
	syncSend := NewSynchronizedSend(stream.Send)
	taskCB := func(p *rpc.TaskProgress) {
		syncSend.Send(&rpc.PlatformInstallResponse{
			Message: &rpc.PlatformInstallResponse_TaskProgress{
				TaskProgress: p,
			},
		})
	}
	downloadCB := func(p *rpc.DownloadProgress) {
		syncSend.Send(&rpc.PlatformInstallResponse{
			Message: &rpc.PlatformInstallResponse_Progress{
				Progress: p,
			},
		})
	}

	install := func() error {
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
		fmt.Println(ref)
		platformRelease, tools, err := pme.FindPlatformReleaseDependencies(ref)
		if err != nil {
			return &cmderrors.PlatformNotFoundError{Platform: ref.String(), Cause: err}
		}

		// Prerequisite checks before install
		if platformRelease.IsInstalled() {
			taskCB(&rpc.TaskProgress{Name: i18n.Tr("Platform %s already installed", platformRelease), Completed: true})
			return nil
		}

		if req.GetNoOverwrite() {
			if installed := pme.GetInstalledPlatformRelease(platformRelease.Platform); installed != nil {
				return fmt.Errorf("%s: %s",
					i18n.Tr("Platform %s already installed", installed),
					i18n.Tr("could not overwrite"))
			}
		}

		checks := resources.IntegrityCheckFull
		if s.settings.BoardManagerEnableUnsafeInstall() {
			checks = resources.IntegrityCheckNone
		}
		if err := pme.DownloadAndInstallPlatformAndTools(ctx, platformRelease, tools, downloadCB, taskCB, req.GetSkipPostInstall(), req.GetSkipPreUninstall(), checks); err != nil {
			return err
		}

		return nil
	}

	if err := install(); err != nil {
		return err
	}

	err := s.Init(
		&rpc.InitRequest{Instance: req.GetInstance()},
		InitStreamResponseToCallbackFunction(ctx, nil),
	)
	if err != nil {
		return err
	}

	return syncSend.Send(&rpc.PlatformInstallResponse{
		Message: &rpc.PlatformInstallResponse_Result_{
			Result: &rpc.PlatformInstallResponse_Result{},
		},
	})
}
