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

package httpclient

import (
	"context"
	"time"

	"github.com/arduino/arduino-cli/commands/cmderrors"
	"github.com/arduino/arduino-cli/internal/i18n"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	"github.com/arduino/go-paths-helper"
	"github.com/sirupsen/logrus"
	"go.bug.st/downloader/v2"
)

var tr = i18n.Tr

// DownloadFile downloads a file from a URL into the specified path. An optional config and options may be passed (or nil to use the defaults).
// A DownloadProgressCB callback function must be passed to monitor download progress.
// If a not empty queryParameter is passed, it is appended to the URL for analysis purposes.
func DownloadFile(ctx context.Context, path *paths.Path, URL string, queryParameter string, label string, downloadCB rpc.DownloadProgressCB, config downloader.Config, options ...downloader.DownloadOptions) (returnedError error) {
	if queryParameter != "" {
		URL = URL + "?query=" + queryParameter
	}
	logrus.WithField("url", URL).Info("Starting download")
	downloadCB.Start(URL, label)
	defer func() {
		if returnedError == nil {
			downloadCB.End(true, "")
		} else {
			downloadCB.End(false, returnedError.Error())
		}
	}()

	d, err := downloader.DownloadWithConfigAndContext(ctx, path.String(), URL, config, options...)
	if err != nil {
		return err
	}

	err = d.RunAndPoll(func(downloaded int64) {
		downloadCB.Update(downloaded, d.Size())
	}, 250*time.Millisecond)
	if err != nil {
		return err
	}

	// The URL is not reachable for some reason
	if d.Resp.StatusCode >= 400 && d.Resp.StatusCode <= 599 {
		msg := tr("Server responded with: %s", d.Resp.Status)
		return &cmderrors.FailedDownloadError{Message: msg}
	}

	return nil
}
