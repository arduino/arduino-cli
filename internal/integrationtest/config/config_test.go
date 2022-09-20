// This file is part of arduino-cli.
//
// Copyright 2022 ARDUINO SA (http://www.arduino.cc/)
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

package config_test

import (
	"testing"

	"github.com/arduino/arduino-cli/internal/integrationtest"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

func TestInit(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	stdout, stderr, err := cli.Run("config", "init")
	require.Empty(t, stderr)
	require.NoError(t, err)
	require.Contains(t, string(stdout), cli.DataDir().String())
}

func TestInitWithExistingCustomConfig(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	stdout, _, err := cli.Run("config", "init", "--additional-urls", "https://example.com")
	require.NoError(t, err)
	require.Contains(t, string(stdout), cli.DataDir().String())

	configFile, err := cli.DataDir().Join("arduino-cli.yaml").ReadFile()
	require.NoError(t, err)
	config := make(map[string]map[string]interface{})
	err = yaml.Unmarshal(configFile, config)
	require.NoError(t, err)
	require.Equal(t, config["board_manager"]["additional_urls"].([]interface{})[0].(string), "https://example.com")
	require.Equal(t, config["daemon"]["port"].(string), "50051")
	require.Equal(t, config["directories"]["data"].(string), cli.DataDir().String())
	require.Equal(t, config["directories"]["downloads"].(string), env.SharedDownloadsDir().String())
	require.Equal(t, config["directories"]["user"].(string), cli.SketchbookDir().String())
	require.Empty(t, config["logging"]["file"])
	require.Equal(t, config["logging"]["format"].(string), "text")
	require.Equal(t, config["logging"]["level"].(string), "info")
	require.Equal(t, config["metrics"]["addr"].(string), ":9090")
	require.True(t, config["metrics"]["enabled"].(bool))

	configFilePath := cli.WorkingDir().Join("config", "test", "config.yaml")
	require.NoFileExists(t, configFilePath.String())
	stdout, _, err = cli.Run("config", "init", "--dest-file", configFilePath.String())
	require.NoError(t, err)
	require.Contains(t, string(stdout), configFilePath.String())

	configFile, err = configFilePath.ReadFile()
	require.NoError(t, err)
	err = yaml.Unmarshal(configFile, config)
	require.NoError(t, err)
	require.Empty(t, config["board_manager"]["additional_urls"])
	require.Equal(t, config["daemon"]["port"].(string), "50051")
	require.Equal(t, config["directories"]["data"].(string), cli.DataDir().String())
	require.Equal(t, config["directories"]["downloads"].(string), env.SharedDownloadsDir().String())
	require.Equal(t, config["directories"]["user"].(string), cli.SketchbookDir().String())
	require.Empty(t, config["logging"]["file"])
	require.Equal(t, config["logging"]["format"].(string), "text")
	require.Equal(t, config["logging"]["level"].(string), "info")
	require.Equal(t, config["metrics"]["addr"].(string), ":9090")
	require.True(t, config["metrics"]["enabled"].(bool))
}

func TestInitOverwriteExistingCustomFile(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	stdout, _, err := cli.Run("config", "init", "--additional-urls", "https://example.com")
	require.NoError(t, err)
	require.Contains(t, string(stdout), cli.DataDir().String())

	configFile, err := cli.DataDir().Join("arduino-cli.yaml").ReadFile()
	require.NoError(t, err)
	config := make(map[string]map[string]interface{})
	err = yaml.Unmarshal(configFile, config)
	require.NoError(t, err)
	require.Equal(t, config["board_manager"]["additional_urls"].([]interface{})[0].(string), "https://example.com")
	require.Equal(t, config["daemon"]["port"].(string), "50051")
	require.Equal(t, config["directories"]["data"].(string), cli.DataDir().String())
	require.Equal(t, config["directories"]["downloads"].(string), env.SharedDownloadsDir().String())
	require.Equal(t, config["directories"]["user"].(string), cli.SketchbookDir().String())
	require.Empty(t, config["logging"]["file"])
	require.Equal(t, config["logging"]["format"].(string), "text")
	require.Equal(t, config["logging"]["level"].(string), "info")
	require.Equal(t, config["metrics"]["addr"].(string), ":9090")
	require.True(t, config["metrics"]["enabled"].(bool))

	stdout, _, err = cli.Run("config", "init", "--overwrite")
	require.NoError(t, err)
	require.Contains(t, string(stdout), cli.DataDir().String())

	configFile, err = cli.DataDir().Join("arduino-cli.yaml").ReadFile()
	require.NoError(t, err)
	err = yaml.Unmarshal(configFile, config)
	require.NoError(t, err)
	require.Empty(t, config["board_manager"]["additional_urls"])
	require.Equal(t, config["daemon"]["port"].(string), "50051")
	require.Equal(t, config["directories"]["data"].(string), cli.DataDir().String())
	require.Equal(t, config["directories"]["downloads"].(string), env.SharedDownloadsDir().String())
	require.Equal(t, config["directories"]["user"].(string), cli.SketchbookDir().String())
	require.Empty(t, config["logging"]["file"])
	require.Equal(t, config["logging"]["format"].(string), "text")
	require.Equal(t, config["logging"]["level"].(string), "info")
	require.Equal(t, config["metrics"]["addr"].(string), ":9090")
	require.True(t, config["metrics"]["enabled"].(bool))
}
