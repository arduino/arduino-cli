/*
 * This file is part of arduino-cli.
 *
 * Copyright 2018 ARDUINO SA (http://www.arduino.cc/)
 *
 * This software is released under the GNU General Public License version 3,
 * which covers the main part of arduino-cli.
 * The terms of this license can be found at:
 * https://www.gnu.org/licenses/gpl-3.0.en.html
 *
 * You can be released from the requirements of the above licenses by purchasing
 * a commercial license. Buying such a license is mandatory if you want to modify or
 * otherwise use the software for commercial activities involving the Arduino
 * software without disclosing the source code of your own applications. To purchase
 * a commercial license, send an email to license@arduino.cc.
 */

package lib

import (
	"context"
	"fmt"

	"github.com/arduino/arduino-cli/arduino/libraries/librariesindex"
	"github.com/arduino/arduino-cli/commands"
	"github.com/arduino/arduino-cli/rpc"
	"github.com/sirupsen/logrus"
)

func LibraryInstall(ctx context.Context, req *rpc.LibraryInstallReq,
	downloadCB commands.DownloadProgressCB, taskCB commands.TaskProgressCB) error {

	lm := commands.GetLibraryManager(req)

	version, err := commands.ParseVersion(req)
	if err != nil {
		return fmt.Errorf("invalid version: %s", err)
	}

	ref := &librariesindex.Reference{Name: req.GetName(), Version: version}
	libRelease := lm.Index.FindRelease(ref)
	if libRelease == nil {
		return fmt.Errorf("library not found: %s", ref.String())
	}

	taskCB(&rpc.TaskProgress{Name: "Downloading " + libRelease.String()})
	if err := downloadLibrary(lm, libRelease, downloadCB); err != nil {
		return err
	}
	taskCB(&rpc.TaskProgress{Completed: true})

	taskCB(&rpc.TaskProgress{Name: "Installing " + libRelease.String()})
	logrus.WithField("library", libRelease).Info("Installing library")
	libPath, libReplaced, err := lm.InstallPrerequisiteCheck(libRelease)
	if err != nil {
		return fmt.Errorf("checking lib install prerequisites: %s", err)
	}
	if libReplaced != nil {
		taskCB(&rpc.TaskProgress{Message: fmt.Sprintf("Replacing %s with %s", libReplaced, libRelease)})
	}
	if err := lm.Install(libRelease, libPath); err != nil {
		return err
	}
	taskCB(&rpc.TaskProgress{Completed: true})

	if _, err := commands.Rescan(ctx, &rpc.RescanReq{Instance: req.Instance}); err != nil {
		return fmt.Errorf("rescanning libraries: %s", err)
	}
	return nil
}
