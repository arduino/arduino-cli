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

package resources

import (
	"fmt"
	"os"

	"github.com/arduino/arduino-cli/arduino/httpclient"
	paths "github.com/arduino/go-paths-helper"
	"go.bug.st/downloader/v2"
)

// Download performs a download loop using the provided downloader.Downloader.
// Messages are passed back to the DownloadProgressCB using label as text for the File field.
func (r *DownloadResource) Download(downloadDir *paths.Path, config *downloader.Config, label string, downloadCB httpclient.DownloadProgressCB) error {
	path, err := r.ArchivePath(downloadDir)
	if err != nil {
		return fmt.Errorf(tr("getting archive path: %s"), err)
	}

	if _, err := path.Stat(); os.IsNotExist(err) {
		// normal download
	} else if err == nil {
		// check local file integrity
		ok, err := r.TestLocalArchiveIntegrity(downloadDir)
		if err != nil || !ok {
			if err := path.Remove(); err != nil {
				return fmt.Errorf(tr("removing corrupted archive file: %s"), err)
			}
		} else {
			// File is cached, nothing to do here

			// This signal means that the file is already downloaded
			downloadCB(&httpclient.DownloadProgress{
				File:      label,
				Completed: true,
			})
			return nil
		}
	} else {
		return fmt.Errorf(tr("getting archive file info: %s"), err)
	}
	return httpclient.DownloadFile(path, r.URL, label, downloadCB, config)
}
