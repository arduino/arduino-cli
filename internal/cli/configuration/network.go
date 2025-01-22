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
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/arduino/arduino-cli/commands/cmderrors"
	"github.com/arduino/arduino-cli/internal/i18n"
	"github.com/arduino/arduino-cli/internal/version"
	"go.bug.st/downloader/v2"
	"google.golang.org/grpc/metadata"
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

// ConnectionTimeout returns the network connection timeout
func (settings *Settings) ConnectionTimeout() time.Duration {
	if timeout, ok, _ := settings.GetDurationOk("network.connection_timeout"); ok {
		return timeout
	}
	return settings.Defaults.GetDuration("network.connection_timeout")
}

// SkipCloudApiForBoardDetection returns whether the cloud API should be ignored for board detection
func (settings *Settings) SkipCloudApiForBoardDetection() bool {
	return settings.GetBool("network.cloud_api.skip_board_detection_calls")
}

// NetworkProxy returns the proxy configuration (mainly used by HTTP clients)
func (settings *Settings) NetworkProxy() (*url.URL, error) {
	if proxyConfig, ok, _ := settings.GetStringOk("network.proxy"); !ok {
		return nil, nil
	} else if proxy, err := url.Parse(proxyConfig); err != nil {
		return nil, errors.New(i18n.Tr("Invalid network.proxy '%[1]s': %[2]s", proxyConfig, err))
	} else {
		return proxy, nil
	}
}

// NewHttpClient returns a new http client for use in the arduino-cli
func (settings *Settings) NewHttpClient(ctx context.Context) (*http.Client, error) {
	proxy, err := settings.NetworkProxy()
	if err != nil {
		return nil, err
	}
	userAgent := settings.UserAgent()
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if extraUserAgent := strings.Join(md.Get("user-agent"), " "); extraUserAgent != "" {
			userAgent += " " + extraUserAgent
		}
	}
	return &http.Client{
		Transport: &httpClientRoundTripper{
			transport: &http.Transport{
				Proxy: http.ProxyURL(proxy),
			},
			userAgent: userAgent,
		},
		Timeout: settings.ConnectionTimeout(),
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
func (settings *Settings) DownloaderConfig(ctx context.Context) (downloader.Config, error) {
	httpClient, err := settings.NewHttpClient(ctx)
	if err != nil {
		return downloader.Config{}, &cmderrors.InvalidArgumentError{
			Message: i18n.Tr("Could not connect via HTTP"),
			Cause:   err}
	}
	return downloader.Config{
		HttpClient: *httpClient,
	}, nil
}
