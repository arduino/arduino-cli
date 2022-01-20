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
	"fmt"
	"net/url"
	"runtime"

	"github.com/arduino/arduino-cli/cli/globals"
	"github.com/spf13/viper"
)

// Config is the configuration of the http client
type Config struct {
	UserAgent string
	Proxy     *url.URL
}

// ConfigFromSettings returns httpConfigs using specified settings
func ConfigFromSettings(settings *viper.Viper) (*Config, error) {
	var proxy *url.URL
	var err error
	if settings.IsSet("network.proxy") {
		proxyConfig := settings.GetString("network.proxy")
		if proxyConfig == "" {
			// empty configuration
			// this workaround must be here until viper can UnSet properties:
			// https://github.com/spf13/viper/pull/519
		} else if proxy, err = url.Parse(proxyConfig); err != nil {
			return nil, fmt.Errorf(tr("Invalid network.proxy '%[1]s': %[2]s"), proxyConfig, err)
		}
	}

	return &Config{
		UserAgent: UserAgent(settings.GetString("network.user_agent_ext")),
		Proxy:     proxy,
	}, nil
}

// UserAgent returns the user agent for the cli http client.
// If a non-empty subComponent is set it will be append to the main component.
func UserAgent(subComponent string) string {
	// subComponent := settings.GetString("network.user_agent_ext")
	if subComponent != "" {
		subComponent = " " + subComponent
	}

	return fmt.Sprintf("%s/%s%s (%s; %s; %s) Commit:%s",
		globals.VersionInfo.Application,
		globals.VersionInfo.VersionString,
		subComponent,
		runtime.GOARCH, runtime.GOOS, runtime.Version(),
		globals.VersionInfo.Commit)
}
