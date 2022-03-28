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
	"net/http"
	"net/url"
	"time"

	"github.com/arduino/arduino-cli/arduino"
	"github.com/arduino/arduino-cli/configuration"
	"github.com/arduino/arduino-cli/i18n"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	"github.com/arduino/go-paths-helper"
	"go.bug.st/downloader/v2"
)

var tr = i18n.Tr

// DownloadFile downloads a file from a URL into the specified path. An optional config and options may be passed (or nil to use the defaults).
// A DownloadProgressCB callback function must be passed to monitor download progress.
func DownloadFile(path *paths.Path, URL string, label string, downloadCB rpc.DownloadProgressCB, config *downloader.Config, options ...downloader.DownloadOptions) error {
	if config == nil {
		c, err := GetDownloaderConfig()
		if err != nil {
			return err
		}
		config = c
	}

	d, err := downloader.DownloadWithConfig(path.String(), URL, *config, options...)
	if err != nil {
		return err
	}
	downloadCB(&rpc.DownloadProgress{
		File:      label,
		Url:       d.URL,
		TotalSize: d.Size(),
	})

	err = d.RunAndPoll(func(downloaded int64) {
		downloadCB(&rpc.DownloadProgress{Downloaded: downloaded})
	}, 250*time.Millisecond)
	if err != nil {
		return err
	}

	// The URL is not reachable for some reason
	if d.Resp.StatusCode >= 400 && d.Resp.StatusCode <= 599 {
		return &arduino.FailedDownloadError{Message: tr("Server responded with: %s", d.Resp.Status)}
	}

	downloadCB(&rpc.DownloadProgress{Completed: true})
	return nil
}

// Config is the configuration of the http client
type Config struct {
	UserAgent string
	Proxy     *url.URL
}

// New returns a default http client for use in the arduino-cli
func New() (*http.Client, error) {
	userAgent := configuration.UserAgent(configuration.Settings)
	proxy, err := configuration.NetworkProxy(configuration.Settings)
	if err != nil {
		return nil, err
	}
	return NewWithConfig(&Config{UserAgent: userAgent, Proxy: proxy}), nil
}

// NewWithConfig creates a http client for use in the arduino-cli, with a given configuration
func NewWithConfig(config *Config) *http.Client {
	return &http.Client{
		Transport: &httpClientRoundTripper{
			transport: &http.Transport{
				Proxy: http.ProxyURL(config.Proxy),
			},
			userAgent: config.UserAgent,
		},
	}
}

// GetDownloaderConfig returns the downloader configuration based on current settings.
func GetDownloaderConfig() (*downloader.Config, error) {
	httpClient, err := New()
	if err != nil {
		return nil, &arduino.InvalidArgumentError{Message: tr("Could not connect via HTTP"), Cause: err}
	}
	return &downloader.Config{
		HttpClient: *httpClient,
	}, nil
}

type httpClientRoundTripper struct {
	transport http.RoundTripper
	userAgent string
}

func (h *httpClientRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Add("User-Agent", h.userAgent)
	return h.transport.RoundTrip(req)
}
