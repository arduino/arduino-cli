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

	"github.com/arduino/arduino-cli/configuration"
)

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

type httpClientRoundTripper struct {
	transport http.RoundTripper
	userAgent string
}

func (h *httpClientRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Add("User-Agent", h.userAgent)
	return h.transport.RoundTrip(req)
}
