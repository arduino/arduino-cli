//
// This file is part of arduino-cli.
//
// Copyright 2018 ARDUINO SA (http://www.arduino.cc/)
//
// This software is released under the GNU General Public License version 3,
// which covers the main part of arduino-cli.
// The terms of this license can be found at:
// https://www.gnu.org/licenses/gpl-3.0.en.html
//
// You can be released from the requirements of the above licenses by purchasing
// a commercial license. Buying such a license is mandatory if you want to modify or
// otherwise use the software for commercial activities involving the Arduino
// software without disclosing the source code of your own applications. To purchase
// a commercial license, send an email to license@arduino.cc.
//

package output

import (
	"github.com/arduino/arduino-cli/cli/globals"
	"github.com/arduino/arduino-cli/commands"
	rpc "github.com/arduino/arduino-cli/rpc/commands"
)

// ProgressBar returns a DownloadProgressCB that prints a progress bar.
// If JSON output format has been selected, the callback outputs nothing.
func ProgressBar() commands.DownloadProgressCB {
	if globals.OutputFormat != "json" {
		return NewDownloadProgressBarCB()
	}
	return func(curr *rpc.DownloadProgress) {
		// XXX: Output progress in JSON?
	}
}

// TaskProgress returns a TaskProgressCB that prints the task progress.
// If JSON output format has been selected, the callback outputs nothing.
func TaskProgress() commands.TaskProgressCB {
	if globals.OutputFormat != "json" {
		return NewTaskProgressCB()
	}
	return func(curr *rpc.TaskProgress) {
		// XXX: Output progress in JSON?
	}
}
