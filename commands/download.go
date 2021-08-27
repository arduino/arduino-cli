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
	"time"

	"github.com/arduino/arduino-cli/httpclient"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	"go.bug.st/downloader/v2"
)

// GetDownloaderConfig returns the downloader configuration based on
// current settings.
func GetDownloaderConfig() (*downloader.Config, error) {
	httpClient, err := httpclient.New()
	if err != nil {
		return nil, &InvalidArgumentError{Message: tr("Could not connect via HTTP"), Cause: err}
	}
	return &downloader.Config{
		HttpClient: *httpClient,
	}, nil
}

// Download performs a download loop using the provided downloader.Downloader.
// Messages are passed back to the DownloadProgressCB using label as text for the File field.
func Download(d *downloader.Downloader, label string, downloadCB DownloadProgressCB) error {
	if d == nil {
		// This signal means that the file is already downloaded
		downloadCB(&rpc.DownloadProgress{
			File:      label,
			Completed: true,
		})
		return nil
	}
	downloadCB(&rpc.DownloadProgress{
		File:      label,
		Url:       d.URL,
		TotalSize: d.Size(),
	})
	d.RunAndPoll(func(downloaded int64) {
		downloadCB(&rpc.DownloadProgress{Downloaded: downloaded})
	}, 250*time.Millisecond)
	if d.Error() != nil {
		return d.Error()
	}
	// The URL is not reachable for some reason
	if d.Resp.StatusCode >= 400 && d.Resp.StatusCode <= 599 {
		return &FailedDownloadError{Message: tr("Server responded with: %s", d.Resp.Status)}
	}
	downloadCB(&rpc.DownloadProgress{Completed: true})
	return nil
}
