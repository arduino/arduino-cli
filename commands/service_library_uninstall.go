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
	"github.com/arduino/arduino-cli/internal/arduino/libraries"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	"github.com/arduino/go-paths-helper"
)

// LibraryUninstall FIXMEDOC
func LibraryUninstall(ctx context.Context, req *rpc.LibraryUninstallRequest, taskCB rpc.TaskProgressCB) error {
	lm, err := instances.GetLibraryManager(req.GetInstance())
	if err != nil {
		return err
	}

	version, err := ParseVersion(req.GetVersion())
	if err != nil {
		return err
	}
	lmi, release := lm.NewInstaller()
	defer release()

	libs := lmi.FindByReference(req.GetName(), version, libraries.User)
	if len(libs) == 0 {
		taskCB(&rpc.TaskProgress{Message: tr("Library %s is not installed", req.GetName()), Completed: true})
		return nil
	}

	if len(libs) == 1 {
		taskCB(&rpc.TaskProgress{Name: tr("Uninstalling %s", libs)})
		lmi.Uninstall(libs[0])
		taskCB(&rpc.TaskProgress{Completed: true})
		return nil
	}

	libsDir := paths.NewPathList()
	for _, lib := range libs {
		libsDir.Add(lib.InstallDir)
	}
	return &cmderrors.MultipleLibraryInstallDetected{
		LibName: libs[0].Name,
		LibsDir: libsDir,
		Message: tr("Automatic library uninstall can't be performed in this case, please manually remove them."),
	}
}
