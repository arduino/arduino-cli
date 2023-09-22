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

// DownloadProgressCB is a callback to get updates on download progress
type DownloadProgressCB func(curr *DownloadProgress)

// Start sends a "start" DownloadProgress message to the callback function
func (d DownloadProgressCB) Start(url, label string) {
	d(&DownloadProgress{
		Message: &DownloadProgress_Start{
			Start: &DownloadProgressStart{
				Url:   url,
				Label: label,
			},
		},
	})
}

// Update sends an "update" DownloadProgress message to the callback function
func (d DownloadProgressCB) Update(downloaded int64, totalSize int64) {
	d(&DownloadProgress{
		Message: &DownloadProgress_Update{
			Update: &DownloadProgressUpdate{
				Downloaded: downloaded,
				TotalSize:  totalSize,
			},
		},
	})
}

// End sends an "end" DownloadProgress message to the callback function
func (d DownloadProgressCB) End(success bool, message string) {
	d(&DownloadProgress{
		Message: &DownloadProgress_End{
			End: &DownloadProgressEnd{
				Success: success,
				Message: message,
			},
		},
	})
}

// TaskProgressCB is a callback to receive progress messages
type TaskProgressCB func(msg *TaskProgress)
