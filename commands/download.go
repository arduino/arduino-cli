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
	"net/http"
	"net/url"
	"time"

	"github.com/arduino/arduino-cli/httpclient"
	rpc "github.com/arduino/arduino-cli/rpc/commands"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"go.bug.st/downloader"
)

// GetDownloaderConfig returns the downloader configuration based on
// current settings.
func GetDownloaderConfig() (*downloader.Config, error) {
	res := &downloader.Config{
		RequestHeaders: http.Header{
			"User-Agent": []string{httpclient.UserAgent()},
		},
	}
	if viper.IsSet("network.proxy") {
		proxy := viper.GetString("network.proxy")
		if _, err := url.Parse(proxy); err != nil {
			return nil, errors.New("Invalid network.proxy '" + proxy + "': " + err.Error())
		}
		res.ProxyURL = proxy
		logrus.Infof("Using proxy %s", proxy)
	}
	return res, nil
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
	downloadCB(&rpc.DownloadProgress{Completed: true})
	return nil
}
