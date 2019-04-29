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

	"github.com/arduino/arduino-cli/commands"
	"github.com/arduino/arduino-cli/rpc"
)

func LibraryUpgradeAll(ctx context.Context, req *rpc.LibraryUpgradeAllReq, downloadCB commands.DownloadProgressCB, taskCB commands.TaskProgressCB) error {
	lm := commands.GetLibraryManager(req)

	// Obtain the list of upgradable libraries
	list := ListLibraries(lm, true)

	for _, upgradeDesc := range list.Libraries {
		if err := downloadLibrary(lm, upgradeDesc.Available, downloadCB, taskCB); err != nil {
			return err
		}
	}
	for _, upgradeDesc := range list.Libraries {
		installLibrary(lm, upgradeDesc.Available, taskCB)
	}

	if _, err := commands.Rescan(ctx, &rpc.RescanReq{Instance: req.Instance}); err != nil {
		return fmt.Errorf("rescanning libraries: %s", err)
	}
	return nil
}
