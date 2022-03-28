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
	"github.com/arduino/arduino-cli/arduino"
	"github.com/arduino/arduino-cli/arduino/cores"
	"github.com/arduino/arduino-cli/arduino/cores/packagemanager"
	"github.com/arduino/arduino-cli/arduino/httpclient"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
)

// DownloadToolRelease downloads a ToolRelease
func DownloadToolRelease(pm *packagemanager.PackageManager, toolRelease *cores.ToolRelease, downloadCB DownloadProgressCB) error {
	config, err := httpclient.GetDownloaderConfig()
	if err != nil {
		return err
	}
	return pm.DownloadToolRelease(toolRelease, config, toolRelease.String(), downloadCB.FromRPC())
}

// InstallToolRelease installs a ToolRelease
func InstallToolRelease(pm *packagemanager.PackageManager, toolRelease *cores.ToolRelease, taskCB TaskProgressCB) error {
	log := pm.Log.WithField("Tool", toolRelease)

	if toolRelease.IsInstalled() {
		log.Warn("Tool already installed")
		taskCB(&rpc.TaskProgress{Name: tr("Tool %s already installed", toolRelease), Completed: true})
		return nil
	}

	log.Info("Installing tool")
	taskCB(&rpc.TaskProgress{Name: tr("Installing %s", toolRelease)})
	err := pm.InstallTool(toolRelease)
	if err != nil {
		log.WithError(err).Warn("Cannot install tool")
		return &arduino.FailedInstallError{Message: tr("Cannot install tool %s", toolRelease), Cause: err}
	}
	log.Info("Tool installed")
	taskCB(&rpc.TaskProgress{Message: tr("%s installed", toolRelease), Completed: true})

	return nil
}
