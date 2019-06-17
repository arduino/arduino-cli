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

package cli

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/arduino/arduino-cli/commands"
	"github.com/arduino/arduino-cli/common/formatter"
	"github.com/arduino/arduino-cli/output"
	rpc "github.com/arduino/arduino-cli/rpc/commands"
)

// OutputJSONOrElse outputs the JSON encoding of v if the JSON output format has been
// selected by the user and returns false. Otherwise no output is produced and the
// function returns true.
func OutputJSONOrElse(v interface{}) bool {
	if !GlobalFlags.OutputJSON {
		return true
	}
	d, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		formatter.PrintError(err, "Error during JSON encoding of the output")
		os.Exit(ErrGeneric)
	}
	fmt.Print(string(d))
	return false
}

// OutputProgressBar returns a DownloadProgressCB that prints a progress bar.
// If JSON output format has been selected, the callback outputs nothing.
func OutputProgressBar() commands.DownloadProgressCB {
	if !GlobalFlags.OutputJSON {
		return output.NewDownloadProgressBarCB()
	}
	return func(curr *rpc.DownloadProgress) {
		// XXX: Output progress in JSON?
	}
}

// OutputTaskProgress returns a TaskProgressCB that prints the task progress.
// If JSON output format has been selected, the callback outputs nothing.
func OutputTaskProgress() commands.TaskProgressCB {
	if !GlobalFlags.OutputJSON {
		return output.NewTaskProgressCB()
	}
	return func(curr *rpc.TaskProgress) {
		// XXX: Output progress in JSON?
	}
}
