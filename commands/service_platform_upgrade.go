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

	"github.com/arduino/arduino-cli/commands/internal/instances"
	"github.com/arduino/arduino-cli/internal/arduino/cores"
	"github.com/arduino/arduino-cli/internal/arduino/cores/packagemanager"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
)

// PlatformUpgradeStreamResponseToCallbackFunction returns a gRPC stream to be used in PlatformUpgrade that sends
// all responses to the callback function.
func PlatformUpgradeStreamResponseToCallbackFunction(ctx context.Context, downloadCB rpc.DownloadProgressCB, taskCB rpc.TaskProgressCB) (rpc.ArduinoCoreService_PlatformUpgradeServer, func() *rpc.PlatformUpgradeResponse_Result) {
	var resp *rpc.PlatformUpgradeResponse_Result
	return streamResponseToCallback(ctx, func(r *rpc.PlatformUpgradeResponse) error {
			if r.GetProgress() != nil {
				downloadCB(r.GetProgress())
			}
			if r.GetTaskProgress() != nil {
				taskCB(r.GetTaskProgress())
			}
			if r.GetResult() != nil {
				resp = r.GetResult()
			}
			return nil
		}), func() *rpc.PlatformUpgradeResponse_Result {
			return resp
		}
}

// PlatformUpgrade upgrades a platform package
func (s *arduinoCoreServerImpl) PlatformUpgrade(req *rpc.PlatformUpgradeRequest, stream rpc.ArduinoCoreService_PlatformUpgradeServer) error {
	syncSend := NewSynchronizedSend(stream.Send)
	ctx := stream.Context()
	downloadCB := func(p *rpc.DownloadProgress) {
		syncSend.Send(&rpc.PlatformUpgradeResponse{
			Message: &rpc.PlatformUpgradeResponse_Progress{
				Progress: p,
			},
		})
	}
	taskCB := func(p *rpc.TaskProgress) {
		syncSend.Send(&rpc.PlatformUpgradeResponse{
			Message: &rpc.PlatformUpgradeResponse_TaskProgress{
				TaskProgress: p,
			},
		})
	}

	upgrade := func() (*cores.PlatformRelease, error) {
		pme, release, err := instances.GetPackageManagerExplorer(req.GetInstance())
		if err != nil {
			return nil, err
		}
		defer release()

		// Extract all PlatformReference to platforms that have updates
		ref := &packagemanager.PlatformReference{
			Package:              req.GetPlatformPackage(),
			PlatformArchitecture: req.GetArchitecture(),
		}
		platform, err := pme.DownloadAndInstallPlatformUpgrades(ctx, ref, downloadCB, taskCB, req.GetSkipPostInstall(), req.GetSkipPreUninstall())
		if err != nil {
			return platform, err
		}

		return platform, nil
	}

	platformRelease, err := upgrade()
	if platformRelease != nil {
		syncSend.Send(&rpc.PlatformUpgradeResponse{
			Message: &rpc.PlatformUpgradeResponse_Result_{
				Result: &rpc.PlatformUpgradeResponse_Result{
					Platform: &rpc.Platform{
						Metadata: platformToRPCPlatformMetadata(platformRelease.Platform),
						Release:  platformReleaseToRPC(platformRelease),
					},
				},
			},
		})
	}
	if err != nil {
		return err
	}

	if err := s.Init(&rpc.InitRequest{Instance: req.GetInstance()}, InitStreamResponseToCallbackFunction(ctx, nil)); err != nil {
		return err
	}

	return nil
}
