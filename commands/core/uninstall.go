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

	"github.com/arduino/arduino-cli/commands"
	"github.com/arduino/arduino-cli/commands/internal/instances"
	"github.com/arduino/arduino-cli/internal/arduino"
	"github.com/arduino/arduino-cli/internal/arduino/cores/packagemanager"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
)

// PlatformUninstall FIXMEDOC
func PlatformUninstall(ctx context.Context, req *rpc.PlatformUninstallRequest, taskCB rpc.TaskProgressCB) (*rpc.PlatformUninstallResponse, error) {
	if err := platformUninstall(ctx, req, taskCB); err != nil {
		return nil, err
	}
	if err := commands.Init(&rpc.InitRequest{Instance: req.GetInstance()}, nil); err != nil {
		return nil, err
	}
	return &rpc.PlatformUninstallResponse{}, nil
}

// platformUninstall is the implementation of platform unistaller
func platformUninstall(ctx context.Context, req *rpc.PlatformUninstallRequest, taskCB rpc.TaskProgressCB) error {
	pme, release := instances.GetPackageManagerExplorer(req.GetInstance())
	if pme == nil {
		return &arduino.InvalidInstanceError{}
	}
	defer release()

	ref := &packagemanager.PlatformReference{
		Package:              req.GetPlatformPackage(),
		PlatformArchitecture: req.GetArchitecture(),
	}
	if ref.PlatformVersion == nil {
		platform := pme.FindPlatform(ref)
		if platform == nil {
			return &arduino.PlatformNotFoundError{Platform: ref.String()}
		}
		platformRelease := pme.GetInstalledPlatformRelease(platform)
		if platformRelease == nil {
			return &arduino.PlatformNotFoundError{Platform: ref.String()}
		}
		ref.PlatformVersion = platformRelease.Version
	}

	platform, tools, err := pme.FindPlatformReleaseDependencies(ref)
	if err != nil {
		return &arduino.NotFoundError{Message: tr("Can't find dependencies for platform %s", ref), Cause: err}
	}

	if err := pme.UninstallPlatform(platform, taskCB, req.GetSkipPreUninstall()); err != nil {
		return err
	}

	for _, tool := range tools {
		if !pme.IsToolRequired(tool) {
			taskCB(&rpc.TaskProgress{Name: tr("Uninstalling %s, tool is no more required", tool)})
			pme.UninstallTool(tool, taskCB, req.GetSkipPreUninstall())
		}
	}

	return nil
}
