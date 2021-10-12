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

package output

import (
	"fmt"

	"github.com/arduino/arduino-cli/commands"
	"github.com/arduino/arduino-cli/i18n"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	"github.com/cmaglie/pb"
)

var (
	// OutputFormat can be "text" or "json"
	OutputFormat string
	tr           = i18n.Tr
)

// ProgressBar returns a DownloadProgressCB that prints a progress bar.
// If JSON output format has been selected, the callback outputs nothing.
func ProgressBar() commands.DownloadProgressCB {
	if OutputFormat != "json" {
		return NewDownloadProgressBarCB()
	}
	return func(curr *rpc.DownloadProgress) {
		// XXX: Output progress in JSON?
	}
}

// TaskProgress returns a TaskProgressCB that prints the task progress.
// If JSON output format has been selected, the callback outputs nothing.
func TaskProgress() commands.TaskProgressCB {
	if OutputFormat != "json" {
		return NewTaskProgressCB()
	}
	return func(curr *rpc.TaskProgress) {
		// XXX: Output progress in JSON?
	}
}

// NewDownloadProgressBarCB creates a progress bar callback that outputs a progress
// bar on the terminal
func NewDownloadProgressBarCB() func(*rpc.DownloadProgress) {
	var bar *pb.ProgressBar
	var prefix string
	return func(curr *rpc.DownloadProgress) {
		// fmt.Printf(">>> %v\n", curr)
		if filename := curr.GetFile(); filename != "" {
			if curr.GetCompleted() {
				fmt.Println(tr("%s already downloaded", filename))
				return
			}
			prefix = filename
			bar = pb.StartNew(int(curr.GetTotalSize()))
			bar.Prefix(prefix)
			bar.SetUnits(pb.U_BYTES)
		}
		if curr.GetDownloaded() != 0 {
			bar.Set(int(curr.GetDownloaded()))
		}
		if curr.GetCompleted() {
			bar.FinishPrintOver(tr("%s downloaded", prefix))
		}
	}
}

// NewNullDownloadProgressCB returns a progress bar callback that outputs nothing.
func NewNullDownloadProgressCB() func(*rpc.DownloadProgress) {
	return func(*rpc.DownloadProgress) {}
}

// NewTaskProgressCB returns a commands.TaskProgressCB progress listener
// that outputs to terminal
func NewTaskProgressCB() func(curr *rpc.TaskProgress) {
	var name string
	return func(curr *rpc.TaskProgress) {
		// fmt.Printf(">>> %v\n", curr)
		msg := curr.GetMessage()
		if curr.GetName() != "" {
			name = curr.GetName()
			if msg == "" {
				msg = name
			}
		}
		if msg != "" {
			fmt.Print(msg)
			if curr.GetCompleted() {
				fmt.Println()
			} else {
				fmt.Println("...")
			}
		}
	}
}

// NewNullTaskProgressCB returns a progress bar callback that outputs nothing.
func NewNullTaskProgressCB() func(curr *rpc.TaskProgress) {
	return func(curr *rpc.TaskProgress) {}
}
