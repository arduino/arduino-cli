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
	"go.bug.st/testifyjson/requirejson"
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

func TestInitDestAbsolutePath(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	dest := cli.WorkingDir().Join("config", "test")
	expectedConfigFile := dest.Join("arduino-cli.yaml")
	require.NoFileExists(t, expectedConfigFile.String())
	stdout, _, err := cli.Run("config", "init", "--dest-dir", dest.String())
	require.NoError(t, err)
	require.Contains(t, string(stdout), expectedConfigFile.String())
	require.FileExists(t, expectedConfigFile.String())
}

func TestInistDestRelativePath(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	dest := cli.WorkingDir().Join("config", "test")
	expectedConfigFile := dest.Join("arduino-cli.yaml")
	require.NoFileExists(t, expectedConfigFile.String())
	stdout, _, err := cli.Run("config", "init", "--dest-dir", "config/test")
	require.NoError(t, err)
	require.Contains(t, string(stdout), expectedConfigFile.String())
	require.FileExists(t, expectedConfigFile.String())
}

func TestInitDestFlagWithOverwriteFlag(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	dest := cli.WorkingDir().Join("config", "test")
	expectedConfigFile := dest.Join("arduino-cli.yaml")
	require.NoFileExists(t, expectedConfigFile.String())

	_, _, err := cli.Run("config", "init", "--dest-dir", dest.String())
	require.NoError(t, err)
	require.FileExists(t, expectedConfigFile.String())

	_, stderr, err := cli.Run("config", "init", "--dest-dir", dest.String())
	require.Error(t, err)
	require.Contains(t, string(stderr), "Config file already exists, use --overwrite to discard the existing one.")

	stdout, _, err := cli.Run("config", "init", "--dest-dir", dest.String(), "--overwrite")
	require.NoError(t, err)
	require.Contains(t, string(stdout), expectedConfigFile.String())
}

func TestInitDestAndConfigFileFlags(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	_, stderr, err := cli.Run("config", "init", "--dest-file", "some_other_path", "--dest-dir", "some_path")
	require.Error(t, err)
	require.Contains(t, string(stderr), "Can't use --dest-file and --dest-dir flags at the same time.")
}

func TestInitConfigFileFlagAbsolutePath(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	configFile := cli.WorkingDir().Join("config", "test", "config.yaml")
	require.NoFileExists(t, configFile.String())

	stdout, _, err := cli.Run("config", "init", "--dest-file", configFile.String())
	require.NoError(t, err)
	require.Contains(t, string(stdout), configFile.String())
	require.FileExists(t, configFile.String())
}

func TestInitConfigFileFlagRelativePath(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	configFile := cli.WorkingDir().Join("config.yaml")
	require.NoFileExists(t, configFile.String())

	stdout, _, err := cli.Run("config", "init", "--dest-file", "config.yaml")
	require.NoError(t, err)
	require.Contains(t, string(stdout), configFile.String())
	require.FileExists(t, configFile.String())
}

func TestInitConfigFileFlagWithOverwriteFlag(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	configFile := cli.WorkingDir().Join("config", "test", "config.yaml")
	require.NoFileExists(t, configFile.String())

	_, _, err := cli.Run("config", "init", "--dest-file", configFile.String())
	require.NoError(t, err)
	require.FileExists(t, configFile.String())

	_, stderr, err := cli.Run("config", "init", "--dest-file", configFile.String())
	require.Error(t, err)
	require.Contains(t, string(stderr), "Config file already exists, use --overwrite to discard the existing one.")

	stdout, _, err := cli.Run("config", "init", "--dest-file", configFile.String(), "--overwrite")
	require.NoError(t, err)
	require.Contains(t, string(stdout), configFile.String())
}

func TestDump(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	// Create a config file first
	configFile := cli.WorkingDir().Join("config", "test", "config.yaml")
	require.NoFileExists(t, configFile.String())
	_, _, err := cli.Run("config", "init", "--dest-file", configFile.String())
	require.NoError(t, err)
	require.FileExists(t, configFile.String())

	stdout, _, err := cli.Run("config", "dump", "--config-file", configFile.String(), "--format", "json")
	require.NoError(t, err)
	requirejson.Query(t, stdout, ".board_manager | .additional_urls", "[]")

	stdout, _, err = cli.Run("config", "init", "--additional-urls", "https://example.com")
	require.NoError(t, err)
	configFile = cli.DataDir().Join("arduino-cli.yaml")
	require.Contains(t, string(stdout), configFile.String())
	require.FileExists(t, configFile.String())

	stdout, _, err = cli.Run("config", "dump", "--format", "json")
	require.NoError(t, err)
	requirejson.Query(t, stdout, ".board_manager | .additional_urls", "[\"https://example.com\"]")
}

func TestDumpWithConfigFileFlag(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	// Create a config file first
	configFile := cli.WorkingDir().Join("config", "test", "config.yaml")
	require.NoFileExists(t, configFile.String())
	_, _, err := cli.Run("config", "init", "--dest-file", configFile.String(), "--additional-urls=https://example.com")
	require.NoError(t, err)
	require.FileExists(t, configFile.String())

	stdout, _, err := cli.Run("config", "dump", "--config-file", configFile.String(), "--format", "json")
	require.NoError(t, err)
	requirejson.Query(t, stdout, ".board_manager | .additional_urls", "[\"https://example.com\"]")

	stdout, _, err = cli.Run(
		"config",
		"dump",
		"--config-file",
		configFile.String(),
		"--additional-urls=https://another-url.com",
		"--format",
		"json",
	)
	require.NoError(t, err)
	requirejson.Query(t, stdout, ".board_manager | .additional_urls", "[\"https://another-url.com\"]")
}

func TestAddRemoveSetDeleteOnUnexistingKey(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	// Create a config file
	_, _, err := cli.Run("config", "init", "--dest-dir", ".")
	require.NoError(t, err)

	_, stderr, err := cli.Run("config", "add", "some.key", "some_value")
	require.Error(t, err)
	require.Contains(t, string(stderr), "Settings key doesn't exist")

	_, stderr, err = cli.Run("config", "remove", "some.key", "some_value")
	require.Error(t, err)
	require.Contains(t, string(stderr), "Settings key doesn't exist")

	_, stderr, err = cli.Run("config", "set", "some.key", "some_value")
	require.Error(t, err)
	require.Contains(t, string(stderr), "Settings key doesn't exist")

	_, stderr, err = cli.Run("config", "delete", "some.key")
	require.Error(t, err)
	require.Contains(t, string(stderr), "Settings key doesn't exist")
}

func TestAddSingleArgument(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	// Create a config file
	_, _, err := cli.Run("config", "init", "--dest-dir", ".")
	require.NoError(t, err)

	// Verifies no additional urls are present
	stdout, _, err := cli.Run("config", "dump", "--format", "json")
	require.NoError(t, err)
	requirejson.Query(t, stdout, ".board_manager | .additional_urls", "[]")

	// Adds one URL
	url := "https://example.com"
	_, _, err = cli.Run("config", "add", "board_manager.additional_urls", url)
	require.NoError(t, err)

	// Verifies URL has been saved
	stdout, _, err = cli.Run("config", "dump", "--format", "json")
	require.NoError(t, err)
	requirejson.Query(t, stdout, ".board_manager | .additional_urls", "[\"https://example.com\"]")

	// Adds the same URL (should not error)
	_, _, err = cli.Run("config", "add", "board_manager.additional_urls", url)
	require.NoError(t, err)

	// Verifies a second copy has NOT been added
	stdout, _, err = cli.Run("config", "dump", "--format", "json")
	require.NoError(t, err)
	requirejson.Query(t, stdout, ".board_manager | .additional_urls", "[\"https://example.com\"]")
}

func TestAddMultipleArguments(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	// Create a config file
	_, _, err := cli.Run("config", "init", "--dest-dir", ".")
	require.NoError(t, err)

	// Verifies no additional urls are present
	stdout, _, err := cli.Run("config", "dump", "--format", "json")
	require.NoError(t, err)
	requirejson.Query(t, stdout, ".board_manager | .additional_urls", "[]")

	// Adds multiple URLs at the same time
	urls := [3]string{
		"https://example.com/package_example_index.json",
		"https://example.com/yet_another_package_example_index.json",
	}
	_, _, err = cli.Run("config", "add", "board_manager.additional_urls", urls[0], urls[1])
	require.NoError(t, err)

	// Verifies URL has been saved
	stdout, _, err = cli.Run("config", "dump", "--format", "json")
	require.NoError(t, err)
	requirejson.Query(t, stdout, ".board_manager | .additional_urls | length", "2")
	requirejson.Contains(t, stdout, `
	{
		"board_manager": {
			"additional_urls": [
				"https://example.com/package_example_index.json",
				"https://example.com/yet_another_package_example_index.json"
			]
		}
	}`)

	// Adds both the same URLs a second time
	_, _, err = cli.Run("config", "add", "board_manager.additional_urls", urls[0], urls[1])
	require.NoError(t, err)

	// Verifies no change in result array
	stdout, _, err = cli.Run("config", "dump", "--format", "json")
	require.NoError(t, err)
	requirejson.Query(t, stdout, ".board_manager | .additional_urls | length", "2")
	requirejson.Contains(t, stdout, `
	{
		"board_manager": {
			"additional_urls": [
				"https://example.com/package_example_index.json",
				"https://example.com/yet_another_package_example_index.json"
			]
		}
	}`)

	// Adds multiple URLs ... the middle one is the only new URL
	urls = [3]string{
		"https://example.com/package_example_index.json",
		"https://example.com/a_third_package_example_index.json",
		"https://example.com/yet_another_package_example_index.json",
	}
	_, _, err = cli.Run("config", "add", "board_manager.additional_urls", urls[0], urls[1], urls[2])
	require.NoError(t, err)

	// Verifies URL has been saved
	stdout, _, err = cli.Run("config", "dump", "--format", "json")
	require.NoError(t, err)
	requirejson.Query(t, stdout, ".board_manager | .additional_urls | length", "3")
	requirejson.Contains(t, stdout, `
	{
		"board_manager": {
			"additional_urls": [
				"https://example.com/package_example_index.json",
				"https://example.com/a_third_package_example_index.json",
				"https://example.com/yet_another_package_example_index.json"
			]
		}
	}`)
}

func TestAddOnUnsupportedKey(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	// Create a config file
	_, _, err := cli.Run("config", "init", "--dest-dir", ".")
	require.NoError(t, err)

	// Verifies default value
	stdout, _, err := cli.Run("config", "dump", "--format", "json")
	require.NoError(t, err)
	requirejson.Query(t, stdout, ".daemon | .port", "\"50051\"")

	// Tries and fails to add a new item
	_, stderr, err := cli.Run("config", "add", "daemon.port", "50000")
	require.Error(t, err)
	require.Contains(t, string(stderr), "The key 'daemon.port' is not a list of items, can't add to it.\nMaybe use 'config set'?")

	// Verifies value is not changed
	stdout, _, err = cli.Run("config", "dump", "--format", "json")
	require.NoError(t, err)
	requirejson.Query(t, stdout, ".daemon | .port", "\"50051\"")
}
