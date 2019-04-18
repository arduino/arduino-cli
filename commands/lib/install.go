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
	"github.com/arduino/arduino-cli/arduino/libraries/librariesmanager"
	"github.com/arduino/arduino-cli/commands"
	"github.com/arduino/arduino-cli/rpc"
	"github.com/sirupsen/logrus"
	semver "go.bug.st/relaxed-semver"
)

func LibraryInstall(ctx context.Context, req *rpc.LibraryInstallReq, downloadCB commands.DownloadProgressCB) (*rpc.LibraryInstallResp, error) {

	lm := commands.GetLibraryManager(req)
	var version *semver.Version
	if v, err := semver.Parse(req.GetVersion()); err == nil {
		version = v
	} else {
		return nil, fmt.Errorf("invalid version: %s", err)
	}
	ref := &librariesindex.Reference{Name: req.GetName(), Version: version}
	library := lm.Index.FindRelease(ref)
	if library == nil {
		return nil, fmt.Errorf("library not found: %s", ref.String())
	}
	err := downloadLibrary(lm, library, downloadCB)
	if err != nil {
		return nil, err
	}
	err = installLibraries(lm, library)
	if err != nil {
		return nil, err
	}

	_, err = commands.Rescan(ctx, &rpc.RescanReq{Instance: req.Instance})
	if err != nil {
		return nil, err
	}
	return &rpc.LibraryInstallResp{}, nil
}

func installLibraries(lm *librariesmanager.LibrariesManager, libRelease *librariesindex.Release) error {

	logrus.WithField("library", libRelease).Info("Installing library")

	if _, err := lm.Install(libRelease); err != nil {
		//logrus.WithError(err).Warn("Error installing library ", libRelease)
		//formatter.PrintError(err, "Error installing library: "+libRelease.String())
		return fmt.Errorf("library not installed: %s", err)
	}

	//formatter.Print("Installed " + libRelease.String())
	return nil
}
