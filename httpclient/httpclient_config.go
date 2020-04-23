package httpclient

import (
	"errors"
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

// DefaultConfig returns the default http client config
func DefaultConfig() (*Config, error) {
	var proxy *url.URL
	var err error
	if viper.IsSet("network.proxy") {
		proxyConfig := viper.GetString("network.proxy")
		if proxy, err = url.Parse(proxyConfig); err != nil {
			return nil, errors.New("Invalid network.proxy '" + proxyConfig + "': " + err.Error())
		}
	}

	return &Config{
		UserAgent: UserAgent(),
		Proxy:     proxy,
	}, nil
}

// UserAgent returns the user agent for the cli http client
func UserAgent() string {
	subComponent := viper.GetString("network.user_agent_ext")
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
