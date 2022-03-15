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
	"github.com/arduino/arduino-cli/arduino/resources"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
)

// DownloadProgressCB is a callback to get updates on download progress
type DownloadProgressCB func(curr *rpc.DownloadProgress)

// FromRPC converts the gRPC DownloadProgessCB in a resources.DownloadProgressCB
func (rpcCB DownloadProgressCB) FromRPC() resources.DownloadProgressCB {
	return func(cb *resources.DownloadProgress) {
		rpcCB(&rpc.DownloadProgress{
			Url:        cb.URL,
			File:       cb.File,
			TotalSize:  cb.TotalSize,
			Downloaded: cb.Downloaded,
			Completed:  cb.Completed,
		})
	}
}

// TaskProgressCB is a callback to receive progress messages
type TaskProgressCB func(msg *rpc.TaskProgress)
