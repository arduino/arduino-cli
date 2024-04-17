// This file is part of arduino-cli.
//
// Copyright 2024 ARDUINO SA (http://www.arduino.cc/)
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

package configuration

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"runtime"

	"github.com/arduino/arduino-cli/commands/cmderrors"
	"github.com/arduino/arduino-cli/version"
	"go.bug.st/downloader/v2"
)

// UserAgent returns the user agent (mainly used by HTTP clients)
func (settings *Settings) UserAgent() string {
	subComponent := ""
	if settings != nil {
		subComponent = settings.GetString("network.user_agent_ext")
	}
	if subComponent != "" {
		subComponent = " " + subComponent
	}

	extendedUA := os.Getenv("ARDUINO_CLI_USER_AGENT_EXTENSION")
	if extendedUA != "" {
		extendedUA = " " + extendedUA
	}

	return fmt.Sprintf("%s/%s%s (%s; %s; %s) Commit:%s%s",
		version.VersionInfo.Application,
		version.VersionInfo.VersionString,
		subComponent,
		runtime.GOARCH, runtime.GOOS, runtime.Version(),
		version.VersionInfo.Commit,
		extendedUA)
}

// ExtraUserAgent returns the extended user-agent section provided via configuration settings
func (settings *Settings) ExtraUserAgent() string {
	return settings.GetString("network.user_agent_ext")
}

// NetworkProxy returns the proxy configuration (mainly used by HTTP clients)
func (settings *Settings) NetworkProxy() (*url.URL, error) {
	if proxyConfig, ok, _ := settings.GetStringOk("network.proxy"); !ok {
		return nil, nil
	} else if proxy, err := url.Parse(proxyConfig); err != nil {
		return nil, fmt.Errorf(tr("Invalid network.proxy '%[1]s': %[2]s"), proxyConfig, err)
	} else {
		return proxy, nil
	}
}

// NewHttpClient returns a new http client for use in the arduino-cli
func (settings *Settings) NewHttpClient() (*http.Client, error) {
	proxy, err := settings.NetworkProxy()
	if err != nil {
		return nil, err
	}
	return &http.Client{
		Transport: &httpClientRoundTripper{
			transport: &http.Transport{
				Proxy: http.ProxyURL(proxy),
			},
			userAgent: settings.UserAgent(),
		},
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

// DownloaderConfig returns the downloader configuration based on current settings.
func (settings *Settings) DownloaderConfig() (downloader.Config, error) {
	httpClient, err := settings.NewHttpClient()
	if err != nil {
		return downloader.Config{}, &cmderrors.InvalidArgumentError{
			Message: tr("Could not connect via HTTP"),
			Cause:   err}
	}
	return downloader.Config{
		HttpClient: *httpClient,
	}, nil
}
