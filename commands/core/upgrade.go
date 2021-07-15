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

package core

import (
	"context"

	"github.com/arduino/arduino-cli/arduino/cores/packagemanager"
	"github.com/arduino/arduino-cli/commands"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// PlatformUpgrade FIXMEDOC
func PlatformUpgrade(ctx context.Context, req *rpc.PlatformUpgradeRequest,
	downloadCB commands.DownloadProgressCB, taskCB commands.TaskProgressCB) (*rpc.PlatformUpgradeResponse, *status.Status) {

	pm := commands.GetPackageManager(req.GetInstance().GetId())
	if pm == nil {
		return nil, status.New(codes.InvalidArgument, "invalid instance")
	}

	// Extract all PlatformReference to platforms that have updates
	ref := &packagemanager.PlatformReference{
		Package:              req.PlatformPackage,
		PlatformArchitecture: req.Architecture,
	}
	if err := upgradePlatform(pm, ref, downloadCB, taskCB, req.GetSkipPostInstall()); err != nil {
		return nil, err
	}

	status := commands.Init(&rpc.InitRequest{Instance: req.Instance}, nil)
	if status != nil {
		return nil, status
	}

	return &rpc.PlatformUpgradeResponse{}, nil
}

func upgradePlatform(pm *packagemanager.PackageManager, platformRef *packagemanager.PlatformReference,
	downloadCB commands.DownloadProgressCB, taskCB commands.TaskProgressCB,
	skipPostInstall bool) *status.Status {
	if platformRef.PlatformVersion != nil {
		return status.New(codes.InvalidArgument, "upgrade doesn't accept parameters with version")
	}

	// Search the latest version for all specified platforms
	platform := pm.FindPlatform(platformRef)
	if platform == nil {
		return status.Newf(codes.InvalidArgument, "platform %s not found", platformRef)
	}
	installed := pm.GetInstalledPlatformRelease(platform)
	if installed == nil {
		return status.Newf(codes.InvalidArgument, "platform %s is not installed", platformRef)
	}
	latest := platform.GetLatestRelease()
	if !latest.Version.GreaterThan(installed.Version) {
		status, e := status.New(codes.AlreadyExists, "platform already at latest version").WithDetails(&rpc.AlreadyAtLatestVersionError{})
		if e != nil { // should never happen
			panic(e)
		}
		return status
	}
	platformRef.PlatformVersion = latest.Version

	platformRelease, tools, err := pm.FindPlatformReleaseDependencies(platformRef)
	if err != nil {
		return status.Newf(codes.FailedPrecondition, "platform %s is not installed", platformRef)
	}
	err = installPlatform(pm, platformRelease, tools, downloadCB, taskCB, skipPostInstall)
	if err != nil {
		return status.Convert(err)
	}

	return nil
}
