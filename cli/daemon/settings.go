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

package daemon

import (
	"fmt"

	"github.com/arduino/go-paths-helper"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// settings defines all the configuration that can be set
// for the daemon command.
// The mapstructure tag must be identical to the name
// of the flag as defined in the cobra.Command used to load
// the settings.
type settings struct {
	IP               string   `mapstructure:"ip"`
	Port             string   `mapstructure:"port"`
	Daemonize        bool     `mapstructure:"daemonize"`
	Debug            bool     `mapstructure:"debug"`
	DebugFilter      []string `mapstructure:"debug-filter"`
	Verbose          bool     `mapstructure:"verbose"`
	OutputFormat     string   `mapstructure:"format"`
	NoColor          bool     `mapstructure:"no-color"`
	NetworkProxy     string   `mapstructure:"network-proxy"`
	NetworkUserAgent string   `mapstructure:"network-user-agent"`
	MetricsEnabled   bool     `mapstructure:"metrics-enabled"`
	MetricsAddress   string   `mapstructure:"metrics-address"`
	LogLevel         string   `mapstructure:"log-level"`
	LogFile          string   `mapstructure:"log-file"`
	LogFormat        string   `mapstructure:"log-format"`
}

// load returns a settings struct populated with all the configurations
// read from configFile if it exists, flags override file values.
// If the config file doesn't exist only uses flag values.
// Returns error if it fails to read the file.
func load(cmd *cobra.Command, configFile string) (s *settings, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic: %s", r)
		}
	}()

	v := viper.New()
	v.SetConfigFile(configFile)
	v.BindPFlags(cmd.Flags())

	// Try to read config file only if it exists.
	// We fallback to flags default
	if configFile != "" && paths.New(configFile).Exist() {
		if err := v.ReadInConfig(); err != nil {
			return nil, err
		}
	}

	s = &settings{}
	if err := v.Unmarshal(s); err != nil {
		return nil, err
	}

	return s, nil
}
