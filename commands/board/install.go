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

package board

import (
	"context"

	"github.com/arduino/arduino-cli/commands"
	"github.com/arduino/arduino-cli/commands/core"
	rpc "github.com/arduino/arduino-cli/rpc/commands"
	"github.com/pkg/errors"
)

// Install FIXMEDOC
func Install(ctx context.Context, req *rpc.BoardInstallReq,
	downloadCB commands.DownloadProgressCB, taskCB commands.TaskProgressCB) (r *rpc.BoardInstallResp, e error) {

	pm := commands.GetPackageManager(req.GetInstance().GetId())
	if pm == nil {
		return nil, errors.New("invalid instance")
	}

	fqbn, _, _, err := pm.FindBoard(req.GetBoard(), nil)
	if err != nil {
		return nil, errors.Errorf("board '%s' not found: %s", req.GetBoard(), err)
	}

	platformInstallReq := &rpc.PlatformInstallReq{
		Instance:        req.GetInstance(),
		PlatformPackage: fqbn.Package,
		Architecture:    fqbn.PlatformArch,
	}
	if _, err := core.PlatformInstall(ctx, platformInstallReq, downloadCB, taskCB); err != nil {
		return nil, errors.WithMessage(err, "installing board platforms")
	}
	return &rpc.BoardInstallResp{}, nil
}
