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

package configuration

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestInit(t *testing.T) {
	settings := NewSettings()

	require.Equal(t, "info", settings.Defaults.GetString("logging.level"))
	require.Equal(t, "text", settings.Defaults.GetString("logging.format"))

	require.Empty(t, settings.Defaults.GetStringSlice("board_manager.additional_urls"))

	require.NotEmpty(t, settings.Defaults.GetString("directories.data"))
	require.Empty(t, settings.Defaults.GetString("directories.downloads"))
	require.NotEmpty(t, settings.DownloadsDir().String())
	require.NotEmpty(t, settings.Defaults.GetString("directories.user"))

	require.Equal(t, "50051", settings.Defaults.GetString("daemon.port"))

	require.Equal(t, true, settings.Defaults.GetBool("metrics.enabled"))
	require.Equal(t, ":9090", settings.Defaults.GetString("metrics.addr"))
}

func TestFindConfigFile(t *testing.T) {
	defaultConfigFile := filepath.Join(getDefaultArduinoDataDir(), "arduino-cli.yaml")
	configFile := FindConfigFlagsInArgsOrFallbackOnEnv([]string{"--config-file"})
	require.Equal(t, defaultConfigFile, configFile)

	configFile = FindConfigFlagsInArgsOrFallbackOnEnv([]string{"--config-file", "some/path/to/config"})
	require.Equal(t, "some/path/to/config", configFile)

	configFile = FindConfigFlagsInArgsOrFallbackOnEnv([]string{"--config-file", "some/path/to/config/arduino-cli.yaml"})
	require.Equal(t, "some/path/to/config/arduino-cli.yaml", configFile)

	configFile = FindConfigFlagsInArgsOrFallbackOnEnv([]string{})
	require.Equal(t, defaultConfigFile, configFile)

	t.Setenv("ARDUINO_CONFIG_FILE", "some/path/to/config")
	configFile = FindConfigFlagsInArgsOrFallbackOnEnv([]string{})
	require.Equal(t, "some/path/to/config", configFile)

	// when both env and flag are specified flag takes precedence
	configFile = FindConfigFlagsInArgsOrFallbackOnEnv([]string{"--config-file", "flag/path"})
	require.Equal(t, "flag/path", configFile)
}
