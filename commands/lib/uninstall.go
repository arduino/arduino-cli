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
	semver "go.bug.st/relaxed-semver"
)

func LibraryUninstall(ctx context.Context, req *rpc.LibraryUninstallReq) (*rpc.LibraryUninstallResp, error) {
	logrus.Info("Executing `arduino lib uninstall`")

	lm := commands.GetLibraryManager(req)
	var version *semver.Version
	if v, err := semver.Parse(req.GetVersion()); err == nil {
		version = v
	} else {
		return nil, fmt.Errorf("invalid version: %s", err)
	}
	ref := &librariesindex.Reference{Name: req.GetName(), Version: version}
	lib := lm.FindByReference(ref)
	if lib == nil {
		return nil, fmt.Errorf("library not installed: %s", ref.String())
	} else {
		//	formatter.Print("Uninstalling " + lib.String()) tramite il CBTAsk
		lm.Uninstall(lib)
	}

	return &rpc.LibraryUninstallResp{}, nil
}
