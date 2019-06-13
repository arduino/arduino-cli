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

// LibraryUninstall FIXMEDOC
func LibraryUninstall(ctx context.Context, req *rpc.LibraryUninstallReq, taskCB commands.TaskProgressCB) error {
	lm := commands.GetLibraryManager(req)

	lib, err := findLibrary(lm, req)
	if err != nil {
		return fmt.Errorf("looking for library: %s", err)
	}

	taskCB(&rpc.TaskProgress{Name: "Uninstalling " + lib.String()})
	lm.Uninstall(lib)
	taskCB(&rpc.TaskProgress{Completed: true})

	return nil
}
