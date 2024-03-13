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
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
)

// LibraryUpgradeAll upgrades all the available libraries
func LibraryUpgradeAll(req *rpc.LibraryUpgradeAllRequest, downloadCB rpc.DownloadProgressCB, taskCB rpc.TaskProgressCB) error {
	li, err := instances.GetLibrariesIndex(req.GetInstance())
	if err != nil {
		return err
	}

	lme, release, err := instances.GetLibraryManagerExplorer(req.GetInstance())
	if err != nil {
		return err
	}
	libsToUpgrade := listLibraries(lme, li, true, false)
	release()

	if err := upgrade(req.GetInstance(), libsToUpgrade, downloadCB, taskCB); err != nil {
		return err
	}

	if err := Init(&rpc.InitRequest{Instance: req.GetInstance()}, nil); err != nil {
		return err
	}

	return nil
}

// LibraryUpgrade upgrades a library
func LibraryUpgrade(ctx context.Context, req *rpc.LibraryUpgradeRequest, downloadCB rpc.DownloadProgressCB, taskCB rpc.TaskProgressCB) error {
	li, err := instances.GetLibrariesIndex(req.GetInstance())
	if err != nil {
		return err
	}

	lme, release, err := instances.GetLibraryManagerExplorer(req.GetInstance())
	if err != nil {
		return err
	}
	libs := listLibraries(lme, li, false, false)
	release()

	// Get the library to upgrade
	name := req.GetName()
	lib := filterByName(libs, name)
	if lib == nil {
		// library not installed...
		return &cmderrors.LibraryNotFoundError{Library: name}
	}
	if lib.Available == nil {
		taskCB(&rpc.TaskProgress{Message: tr("Library %s is already at the latest version", name), Completed: true})
		return nil
	}

	// Install update
	return upgrade(req.GetInstance(), []*installedLib{lib}, downloadCB, taskCB)
}

func upgrade(instance *rpc.Instance, libs []*installedLib, downloadCB rpc.DownloadProgressCB, taskCB rpc.TaskProgressCB) error {
	for _, lib := range libs {
		libInstallReq := &rpc.LibraryInstallRequest{
			Instance:    instance,
			Name:        lib.Library.Name,
			Version:     "",
			NoDeps:      false,
			NoOverwrite: false,
		}
		err := LibraryInstall(context.Background(), libInstallReq, downloadCB, taskCB)
		if err != nil {
			return err
		}
	}

	return nil
}

func filterByName(libs []*installedLib, name string) *installedLib {
	for _, lib := range libs {
		if lib.Library.Name == name {
			return lib
		}
	}
	return nil
}
