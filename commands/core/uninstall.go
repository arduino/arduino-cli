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

	"github.com/arduino/arduino-cli/arduino"
	"github.com/arduino/arduino-cli/arduino/cores/packagemanager"
	"github.com/arduino/arduino-cli/commands"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
)

// PlatformUninstall FIXMEDOC
func PlatformUninstall(ctx context.Context, req *rpc.PlatformUninstallRequest, taskCB rpc.TaskProgressCB) (*rpc.PlatformUninstallResponse, error) {
	pm := commands.GetPackageManager(req.GetInstance().GetId())
	if pm == nil {
		return nil, &arduino.InvalidInstanceError{}
	}

	ref := &packagemanager.PlatformReference{
		Package:              req.PlatformPackage,
		PlatformArchitecture: req.Architecture,
	}
	if ref.PlatformVersion == nil {
		platform := pm.FindPlatform(ref)
		if platform == nil {
			return nil, &arduino.PlatformNotFoundError{Platform: ref.String()}
		}
		platformRelease := pm.GetInstalledPlatformRelease(platform)
		if platformRelease == nil {
			return nil, &arduino.PlatformNotFoundError{Platform: ref.String()}
		}
		ref.PlatformVersion = platformRelease.Version
	}

	platform, tools, err := pm.FindPlatformReleaseDependencies(ref)
	if err != nil {
		return nil, &arduino.NotFoundError{Message: tr("Can't find dependencies for platform %s", ref), Cause: err}
	}

	if err := pm.UninstallPlatform(platform, taskCB); err != nil {
		return nil, err
	}

	for _, tool := range tools {
		if !pm.IsToolRequired(tool) {
			taskCB(&rpc.TaskProgress{Name: tr("Uninstalling %s, tool is no more required", tool)})
			pm.UninstallTool(tool, taskCB)
		}
	}

	if err := commands.Init(&rpc.InitRequest{Instance: req.Instance}, nil); err != nil {
		return nil, err
	}

	return &rpc.PlatformUninstallResponse{}, nil
}
