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
	"time"

	"github.com/arduino/arduino-cli/arduino"
	"github.com/arduino/arduino-cli/httpclient"
	paths "github.com/arduino/go-paths-helper"
	"go.bug.st/downloader/v2"
)

// Download performs a download loop using the provided downloader.Downloader.
// Messages are passed back to the DownloadProgressCB using label as text for the File field.
func (r *DownloadResource) Download(downloadDir *paths.Path, config *downloader.Config, label string, downloadCB DownloadProgressCB) error {
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
			downloadCB(&DownloadProgress{
				File:      label,
				Completed: true,
			})
			return nil
		}
	} else {
		return fmt.Errorf(tr("getting archive file info: %s"), err)
	}
	return DownloadFile(path, r.URL, label, downloadCB, config)
}

func DownloadFile(path *paths.Path, URL string, label string, downloadCB DownloadProgressCB, config *downloader.Config, options ...downloader.DownloadOptions) error {
	d, err := downloader.DownloadWithConfig(path.String(), URL, *config, options...)
	if err != nil {
		return err
	}
	downloadCB(&DownloadProgress{
		File:      label,
		URL:       d.URL,
		TotalSize: d.Size(),
	})

	err = d.RunAndPoll(func(downloaded int64) {
		downloadCB(&DownloadProgress{Downloaded: downloaded})
	}, 250*time.Millisecond)
	if err != nil {
		return err
	}

	// The URL is not reachable for some reason
	if d.Resp.StatusCode >= 400 && d.Resp.StatusCode <= 599 {
		return &arduino.FailedDownloadError{Message: tr("Server responded with: %s", d.Resp.Status)}
	}

	downloadCB(&DownloadProgress{Completed: true})
	return nil
}

// GetDownloaderConfig returns the downloader configuration based on
// current settings.
func GetDownloaderConfig() (*downloader.Config, error) {
	httpClient, err := httpclient.New()
	if err != nil {
		return nil, &arduino.InvalidArgumentError{Message: tr("Could not connect via HTTP"), Cause: err}
	}
	return &downloader.Config{
		HttpClient: *httpClient,
	}, nil
}

// DownloadProgress is a report of the download progress, not all fields may be
// filled and multiple reports may be sent during a download.
type DownloadProgress struct {
	// URL of the download.
	URL string
	// The file being downloaded.
	File string
	// TotalSize is the total size of the file being downloaded.
	TotalSize int64
	// Downloaded is the size of the downloaded portion of the file.
	Downloaded int64
	// Completed reports whether the download is complete.
	Completed bool
}

// DownloadProgressCB is a callback function to report download progress
type DownloadProgressCB func(progress *DownloadProgress)
