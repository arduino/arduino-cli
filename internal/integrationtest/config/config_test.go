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
	"path/filepath"
	"testing"

	"github.com/arduino/arduino-cli/internal/integrationtest"
	"github.com/arduino/go-paths-helper"
	"github.com/stretchr/testify/require"
	"go.bug.st/testifyjson/requirejson"
	"gopkg.in/yaml.v3"
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
	require.Equal(t, config["directories"]["downloads"].(string), cli.DownloadDir().String())
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
	require.Equal(t, config["directories"]["downloads"].(string), cli.DownloadDir().String())
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
	require.Equal(t, config["directories"]["downloads"].(string), cli.DownloadDir().String())
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
	require.Equal(t, config["directories"]["downloads"].(string), cli.DownloadDir().String())
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
	require.Contains(t, string(stderr), "Can't use the following flags together: --dest-file, --dest-dir")
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

	stdout, _, err := cli.Run("config", "dump", "--config-file", configFile.String(), "--json")
	require.NoError(t, err)
	requirejson.Query(t, stdout, ".config | .board_manager | .additional_urls", "[]")

	stdout, _, err = cli.Run("config", "init", "--additional-urls", "https://example.com")
	require.NoError(t, err)
	configFile = cli.DataDir().Join("arduino-cli.yaml")
	require.Contains(t, string(stdout), configFile.String())
	require.FileExists(t, configFile.String())

	stdout, _, err = cli.Run("config", "dump", "--json")
	require.NoError(t, err)
	requirejson.Query(t, stdout, ".config | .board_manager | .additional_urls", "[\"https://example.com\"]")
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

	stdout, _, err := cli.Run("config", "dump", "--config-file", configFile.String(), "--json")
	require.NoError(t, err)
	requirejson.Query(t, stdout, ".config | .board_manager | .additional_urls", "[\"https://example.com\"]")

	stdout, _, err = cli.Run(
		"config",
		"dump",
		"--config-file",
		configFile.String(),
		"--additional-urls=https://another-url.com",
		"--json",
	)
	require.NoError(t, err)
	requirejson.Query(t, stdout, ".config | .board_manager | .additional_urls", "[\"https://another-url.com\"]")
}

func TestAddRemoveSetDeleteOnUnexistingKey(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	// Create a config file
	_, _, err := cli.Run("config", "init", "--dest-dir", ".")
	require.NoError(t, err)

	_, stderr, err := cli.Run("config", "add", "some.key", "some_value", "--config-file", "arduino-cli.yaml")
	require.Error(t, err)
	require.Contains(t, string(stderr), "Settings key doesn't exist")

	_, stderr, err = cli.Run("config", "remove", "some.key", "some_value", "--config-file", "arduino-cli.yaml")
	require.Error(t, err)
	require.Contains(t, string(stderr), "Settings key doesn't exist")

	_, stderr, err = cli.Run("config", "set", "some.key", "some_value", "--config-file", "arduino-cli.yaml")
	require.Error(t, err)
	require.Contains(t, string(stderr), "Settings key doesn't exist")

	_, stderr, err = cli.Run("config", "delete", "some.key", "--config-file", "arduino-cli.yaml")
	require.Error(t, err)
	require.Contains(t, string(stderr), "Cannot delete the key some.key: key not found in settings\n")
}

func TestAddSingleArgument(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	// Create a config file
	_, _, err := cli.Run("config", "init", "--dest-dir", ".")
	require.NoError(t, err)

	// Verifies no additional urls are present
	stdout, _, err := cli.Run("config", "dump", "--json", "--config-file", "arduino-cli.yaml")
	require.NoError(t, err)
	requirejson.Query(t, stdout, ".config | .board_manager | .additional_urls", "[]")

	// Adds one URL
	url := "https://example.com"
	_, _, err = cli.Run("config", "add", "board_manager.additional_urls", url, "--config-file", "arduino-cli.yaml")
	require.NoError(t, err)

	// Verifies URL has been saved
	stdout, _, err = cli.Run("config", "dump", "--json", "--config-file", "arduino-cli.yaml")
	require.NoError(t, err)
	requirejson.Query(t, stdout, ".config | .board_manager | .additional_urls", "[\"https://example.com\"]")

	// Adds the same URL (should not error)
	_, _, err = cli.Run("config", "add",
		"board_manager.additional_urls", url,
		"--config-file", "arduino-cli.yaml")
	require.NoError(t, err)

	// Verifies a second copy has NOT been added
	stdout, _, err = cli.Run("config", "dump", "--json", "--config-file", "arduino-cli.yaml")
	require.NoError(t, err)
	requirejson.Query(t, stdout, ".config | .board_manager | .additional_urls", "[\"https://example.com\"]")
}

func TestAddMultipleArguments(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	// Create a config file
	_, _, err := cli.Run("config", "init", "--dest-dir", ".")
	require.NoError(t, err)

	// Verifies no additional urls are present
	stdout, _, err := cli.Run("config", "dump", "--json", "--config-file", "arduino-cli.yaml")
	require.NoError(t, err)
	requirejson.Query(t, stdout, ".config | .board_manager | .additional_urls", "[]")

	// Adds multiple URLs at the same time
	urls := [3]string{
		"https://example.com/package_example_index.json",
		"https://example.com/yet_another_package_example_index.json",
	}
	_, _, err = cli.Run("config", "add",
		"board_manager.additional_urls", urls[0], urls[1],
		"--config-file", "arduino-cli.yaml")
	require.NoError(t, err)

	// Verifies URL has been saved
	stdout, _, err = cli.Run("config", "dump", "--json", "--config-file", "arduino-cli.yaml")
	require.NoError(t, err)
	requirejson.Query(t, stdout, ".config | .board_manager | .additional_urls | length", "2")
	requirejson.Contains(t, stdout, `
	{
		"config": {
			"board_manager": {
				"additional_urls": [
					"https://example.com/package_example_index.json",
					"https://example.com/yet_another_package_example_index.json"
				]
			}
		}
	}`)

	// Adds both the same URLs a second time
	_, _, err = cli.Run("config", "add", "board_manager.additional_urls", urls[0], urls[1], "--config-file", "arduino-cli.yaml")
	require.NoError(t, err)

	// Verifies no change in result array
	stdout, _, err = cli.Run("config", "dump", "--json", "--config-file", "arduino-cli.yaml")
	require.NoError(t, err)
	requirejson.Query(t, stdout, ".config | .board_manager | .additional_urls | length", "2")
	requirejson.Contains(t, stdout, `
	{
		"config": {
			"board_manager": {
				"additional_urls": [
					"https://example.com/package_example_index.json",
					"https://example.com/yet_another_package_example_index.json"
				]
			}
		}
	}`)

	// Adds multiple URLs ... the middle one is the only new URL
	urls = [3]string{
		"https://example.com/package_example_index.json",
		"https://example.com/a_third_package_example_index.json",
		"https://example.com/yet_another_package_example_index.json",
	}
	_, _, err = cli.Run("config", "add", "board_manager.additional_urls", urls[0], urls[1], urls[2], "--config-file", "arduino-cli.yaml")
	require.NoError(t, err)

	// Verifies URL has been saved
	stdout, _, err = cli.Run("config", "dump", "--json", "--config-file", "arduino-cli.yaml")
	require.NoError(t, err)
	requirejson.Query(t, stdout, ".config | .board_manager | .additional_urls | length", "3")
	requirejson.Contains(t, stdout, `
	{
		"config": {
			"board_manager": {
				"additional_urls": [
					"https://example.com/package_example_index.json",
					"https://example.com/a_third_package_example_index.json",
					"https://example.com/yet_another_package_example_index.json"
				]
			}
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
	stdout, _, err := cli.Run("config", "dump", "--json", "--config-file", "arduino-cli.yaml")
	require.NoError(t, err)
	requirejson.Query(t, stdout, ".config | .daemon | .port", "\"50051\"")

	// Tries and fails to add a new item
	_, stderr, err := cli.Run("config", "add", "daemon.port", "50000", "--config-file", "arduino-cli.yaml")
	require.Error(t, err)
	require.Contains(t, string(stderr), "The key 'daemon.port' is not a list of items, can't add to it.\nMaybe use 'config set'?")

	// Verifies value is not changed
	stdout, _, err = cli.Run("config", "dump", "--json", "--config-file", "arduino-cli.yaml")
	require.NoError(t, err)
	requirejson.Query(t, stdout, ".config | .daemon | .port", "\"50051\"")
}

func TestRemoveSingleArgument(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	// Create a config file
	_, _, err := cli.Run("config", "init", "--dest-dir", ".")
	require.NoError(t, err)

	// Adds URLs
	urls := [2]string{
		"https://example.com/package_example_index.json",
		"https://example.com/yet_another_package_example_index.json",
	}
	_, _, err = cli.Run("config", "add", "board_manager.additional_urls", urls[0], urls[1], "--config-file", "arduino-cli.yaml")
	require.NoError(t, err)

	// Verifies default state
	stdout, _, err := cli.Run("config", "dump", "--json", "--config-file", "arduino-cli.yaml")
	require.NoError(t, err)
	requirejson.Query(t, stdout, ".config | .board_manager | .additional_urls | length", "2")
	requirejson.Contains(t, stdout, `
	{
		"config": {
			"board_manager": {
				"additional_urls": [
					"https://example.com/package_example_index.json",
					"https://example.com/yet_another_package_example_index.json"
				]
			}
		}
	}`)

	// Remove first URL
	_, _, err = cli.Run("config", "remove", "board_manager.additional_urls", urls[0], "--config-file", "arduino-cli.yaml")
	require.NoError(t, err)

	// Verifies URLs has been removed
	stdout, _, err = cli.Run("config", "dump", "--json", "--config-file", "arduino-cli.yaml")
	require.NoError(t, err)
	requirejson.Query(t, stdout, ".config | .board_manager | .additional_urls", "[\"https://example.com/yet_another_package_example_index.json\"]")
}

func TestRemoveMultipleArguments(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	// Create a config file
	_, _, err := cli.Run("config", "init", "--dest-dir", ".")
	require.NoError(t, err)

	// Adds URLs
	urls := [2]string{
		"https://example.com/package_example_index.json",
		"https://example.com/yet_another_package_example_index.json",
	}
	_, _, err = cli.Run("config", "add", "board_manager.additional_urls", urls[0], urls[1], "--config-file", "arduino-cli.yaml")
	require.NoError(t, err)

	// Verifies default state
	stdout, _, err := cli.Run("config", "dump", "--json", "--config-file", "arduino-cli.yaml")
	require.NoError(t, err)
	requirejson.Query(t, stdout, ".config | .board_manager | .additional_urls | length", "2")
	requirejson.Contains(t, stdout, `
	{
		"config": {
			"board_manager": {
				"additional_urls": [
					"https://example.com/package_example_index.json",
					"https://example.com/yet_another_package_example_index.json"
				]
			}
		}
	}`)

	// Remove all URLs
	_, _, err = cli.Run("config", "remove", "board_manager.additional_urls", urls[0], urls[1], "--config-file", "arduino-cli.yaml")
	require.NoError(t, err)

	// Verifies all URLs have been removed
	stdout, _, err = cli.Run("config", "dump", "--json", "--config-file", "arduino-cli.yaml")
	require.NoError(t, err)
	requirejson.Query(t, stdout, ".config | .board_manager | .additional_urls", "[]")
}

func TestRemoveOnUnsupportedKey(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	// Create a config file
	_, _, err := cli.Run("config", "init", "--dest-dir", ".")
	require.NoError(t, err)

	// Verifies default value
	stdout, _, err := cli.Run("config", "dump", "--json", "--config-file", "arduino-cli.yaml")
	require.NoError(t, err)
	requirejson.Query(t, stdout, ".config | .daemon | .port", "\"50051\"")

	// Tries and fails to remove an item
	_, stderr, err := cli.Run("config", "remove", "daemon.port", "50051", "--config-file", "arduino-cli.yaml")
	require.Error(t, err)
	require.Contains(t, string(stderr), "The key 'daemon.port' is not a list of items, can't remove from it.\nMaybe use 'config delete'?")

	// Verifies value is not changed
	stdout, _, err = cli.Run("config", "dump", "--json", "--config-file", "arduino-cli.yaml")
	require.NoError(t, err)
	requirejson.Query(t, stdout, ".config | .daemon | .port", "\"50051\"")
}

func TestSetSliceWithSingleArgument(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	// Create a config file
	_, _, err := cli.Run("config", "init", "--dest-dir", ".")
	require.NoError(t, err)

	// Verifies default state
	stdout, _, err := cli.Run("config", "dump", "--json", "--config-file", "arduino-cli.yaml")
	require.NoError(t, err)
	requirejson.Query(t, stdout, ".config | .board_manager | .additional_urls", "[]")

	// Set an URL in the list
	url := "https://example.com/package_example_index.json"
	_, _, err = cli.Run("config", "set", "board_manager.additional_urls", url, "--config-file", "arduino-cli.yaml")
	require.NoError(t, err)

	// Verifies value is changed
	stdout, _, err = cli.Run("config", "dump", "--json", "--config-file", "arduino-cli.yaml")
	require.NoError(t, err)
	requirejson.Query(t, stdout, ".config | .board_manager | .additional_urls", "[\"https://example.com/package_example_index.json\"]")

	// Set an URL in the list
	url = "https://example.com/yet_another_package_example_index.json"
	_, _, err = cli.Run("config", "set", "board_manager.additional_urls", url, "--config-file", "arduino-cli.yaml")
	require.NoError(t, err)

	// Verifies value is changed
	stdout, _, err = cli.Run("config", "dump", "--json", "--config-file", "arduino-cli.yaml")
	require.NoError(t, err)
	requirejson.Query(t, stdout, ".config | .board_manager | .additional_urls", "[\"https://example.com/yet_another_package_example_index.json\"]")
}

func TestSetSliceWithMultipleArguments(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	// Create a config file
	_, _, err := cli.Run("config", "init", "--dest-dir", ".")
	require.NoError(t, err)

	// Verifies default state
	stdout, _, err := cli.Run("config", "dump", "--json", "--config-file", "arduino-cli.yaml")
	require.NoError(t, err)
	requirejson.Query(t, stdout, ".config | .board_manager | .additional_urls", "[]")

	// Set some URLs in the list
	urls := [7]string{
		"https://example.com/first_package_index.json",
		"https://example.com/second_package_index.json",
	}
	_, _, err = cli.Run("config", "set", "board_manager.additional_urls", urls[0], urls[1], "--config-file", "arduino-cli.yaml")
	require.NoError(t, err)

	// Verifies value is changed
	stdout, _, err = cli.Run("config", "dump", "--json", "--config-file", "arduino-cli.yaml")
	require.NoError(t, err)
	requirejson.Query(t, stdout, ".config | .board_manager | .additional_urls | length", "2")
	requirejson.Contains(t, stdout, `
	{
		"config": {
			"board_manager": {
				"additional_urls": [
					"https://example.com/first_package_index.json",
					"https://example.com/second_package_index.json"
				]
			}
		}
	}`)

	// Set some URLs in the list
	urls = [7]string{
		"https://example.com/third_package_index.json",
		"https://example.com/fourth_package_index.json",
	}
	_, _, err = cli.Run("config", "set", "board_manager.additional_urls", urls[0], urls[1], "--config-file", "arduino-cli.yaml")
	require.NoError(t, err)

	// Verifies previous value is overwritten
	stdout, _, err = cli.Run("config", "dump", "--json", "--config-file", "arduino-cli.yaml")
	require.NoError(t, err)
	requirejson.Query(t, stdout, ".config | .board_manager | .additional_urls | length", "2")
	requirejson.Contains(t, stdout, `
	{
		"config": {
			"board_manager": {
				"additional_urls": [
					"https://example.com/third_package_index.json",
					"https://example.com/fourth_package_index.json"
				]
			}
		}
	}`)

	// Sets a third set of 7 URLs (with only 4 unique values)
	urls = [7]string{
		"https://example.com/first_package_index.json",
		"https://example.com/second_package_index.json",
		"https://example.com/first_package_index.json",
		"https://example.com/fifth_package_index.json",
		"https://example.com/second_package_index.json",
		"https://example.com/sixth_package_index.json",
		"https://example.com/first_package_index.json",
	}
	_, _, err = cli.Run("config",
		"set",
		"board_manager.additional_urls", urls[0], urls[1], urls[2], urls[3], urls[4], urls[5], urls[6],
		"--config-file", "arduino-cli.yaml")
	require.NoError(t, err)

	// Verifies all unique values exist in config
	stdout, _, err = cli.Run("config", "dump", "--json", "--config-file", "arduino-cli.yaml")
	require.NoError(t, err)
	requirejson.Query(t, stdout, ".config | .board_manager | .additional_urls | length", "4")
	requirejson.Contains(t, stdout, `
	{
		"config": {
			"board_manager": {
				"additional_urls": [
					"https://example.com/first_package_index.json",
					"https://example.com/second_package_index.json",
					"https://example.com/fifth_package_index.json",
					"https://example.com/sixth_package_index.json"
				]
			}
		}
	}`)
}

func TestSetStringWithSingleArgument(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	// Create a config file
	_, _, err := cli.Run("config", "init", "--dest-dir", ".")
	require.NoError(t, err)

	// Verifies default state
	stdout, _, err := cli.Run("config", "dump", "--json", "--config-file", "arduino-cli.yaml")
	require.NoError(t, err)
	requirejson.Query(t, stdout, ".config | .logging | .level", "\"info\"")

	// Changes value
	_, _, err = cli.Run("config", "set", "logging.level", "trace", "--config-file", "arduino-cli.yaml")
	require.NoError(t, err)

	// Verifies value is changed
	stdout, _, err = cli.Run("config", "dump", "--json", "--config-file", "arduino-cli.yaml")
	require.NoError(t, err)
	requirejson.Query(t, stdout, ".config | .logging | .level", "\"trace\"")
}

func TestSetStringWithMultipleArguments(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	// Create a config file
	_, _, err := cli.Run("config", "init", "--dest-dir", ".")
	require.NoError(t, err)

	// Verifies default state
	stdout, _, err := cli.Run("config", "dump", "--json", "--config-file", "arduino-cli.yaml")
	require.NoError(t, err)
	requirejson.Query(t, stdout, ".config | .logging | .level", "\"info\"")

	// Tries to change value
	_, stderr, err := cli.Run("config", "set", "logging.level", "trace", "debug")
	require.Error(t, err)
	require.Contains(t, string(stderr), "Can't set multiple values in key logging.level")
}

func TestSetBoolWithSingleArgument(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	// Create a config file
	_, _, err := cli.Run("config", "init", "--dest-dir", ".")
	require.NoError(t, err)

	// Verifies default state
	stdout, _, err := cli.Run("config", "dump", "--json", "--config-file", "arduino-cli.yaml")
	require.NoError(t, err)
	requirejson.Query(t, stdout, ".config | .library | .enable_unsafe_install", "false")

	// Changes value
	_, _, err = cli.Run("config", "set", "library.enable_unsafe_install", "true", "--config-file", "arduino-cli.yaml")
	require.NoError(t, err)

	// Verifies value is changed
	stdout, _, err = cli.Run("config", "dump", "--json", "--config-file", "arduino-cli.yaml")
	require.NoError(t, err)
	requirejson.Query(t, stdout, ".config | .library | .enable_unsafe_install", "true")
}

func TestSetBoolWithMultipleArguments(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	// Create a config file
	_, _, err := cli.Run("config", "init", "--dest-dir", ".")
	require.NoError(t, err)

	// Verifies default state
	stdout, _, err := cli.Run("config", "dump", "--json", "--config-file", "arduino-cli.yaml")
	require.NoError(t, err)
	requirejson.Query(t, stdout, ".config | .library | .enable_unsafe_install", "false")

	// Changes value
	_, stderr, err := cli.Run("config", "set", "library.enable_unsafe_install", "true", "foo", "--config-file", "arduino-cli.yaml")
	require.Error(t, err)
	require.Contains(t, string(stderr), "Can't set multiple values in key library.enable_unsafe_install")
}

func TestDelete(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	// Create a config file
	_, _, err := cli.Run("config", "init", "--dest-dir", ".")
	require.NoError(t, err)

	// Verifies default state
	stdout, _, err := cli.Run("config", "dump", "--json", "--config-file", "arduino-cli.yaml")
	require.NoError(t, err)
	requirejson.Query(t, stdout, ".config | .library | .enable_unsafe_install", "false")

	// Delete config key
	_, _, err = cli.Run("config", "delete", "library.enable_unsafe_install", "--config-file", "arduino-cli.yaml")
	require.NoError(t, err)

	// Verifies value is not found, we read directly from file instead of using
	// the dump command since that would still print the deleted value if it has
	// a default
	configFile := cli.WorkingDir().Join("arduino-cli.yaml")
	configLines, err := configFile.ReadFileAsLines()
	require.NoError(t, err)
	require.NotContains(t, configLines, "enable_unsafe_install")

	// Verifies default state
	stdout, _, err = cli.Run("config", "dump", "--json", "--config-file", "arduino-cli.yaml")
	require.NoError(t, err)
	requirejson.Query(t, stdout, ".config | .board_manager | .additional_urls", "[]")

	// Delete config key and sub keys
	_, _, err = cli.Run("config", "delete", "board_manager", "--config-file", "arduino-cli.yaml")
	require.NoError(t, err)

	// Verifies value is not found, we read directly from file instead of using
	// the dump command since that would still print the deleted value if it has
	// a default
	configFile = cli.WorkingDir().Join("arduino-cli.yaml")
	configLines, err = configFile.ReadFileAsLines()
	require.NoError(t, err)
	require.NotContains(t, configLines, "additional_urls")
	require.NotContains(t, configLines, "board_manager")
}

func TestGet(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	// Create a config file
	_, _, err := cli.Run("config", "init", "--dest-dir", ".")
	require.NoError(t, err)

	// Verifies default state
	stdout, _, err := cli.Run("config", "dump", "--json", "--config-file", "arduino-cli.yaml")
	require.NoError(t, err)
	requirejson.Query(t, stdout, ".config | .daemon | .port", `"50051"`)

	// Get simple key value
	stdout, _, err = cli.Run("config", "get", "daemon.port", "--json", "--config-file", "arduino-cli.yaml")
	require.NoError(t, err)
	requirejson.Contains(t, stdout, `"50051"`)

	// Get structured key value
	stdout, _, err = cli.Run("config", "get", "daemon", "--json", "--config-file", "arduino-cli.yaml")
	require.NoError(t, err)
	requirejson.Contains(t, stdout, `{"port":"50051"}`)

	// Get undefined key
	_, stderr, err := cli.Run("config", "get", "foo", "--json", "--config-file", "arduino-cli.yaml")
	require.Error(t, err)
	requirejson.Contains(t, stderr, `{"error":"Cannot get the configuration key foo: key not found in settings"}`)
}

func TestInitializationOrderOfConfigThroughFlagAndEnv(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	tmp := t.TempDir()
	cliConfig, envConfig := paths.New(filepath.Join(tmp, "cli.yaml")), paths.New(filepath.Join(tmp, "env.yaml"))
	cliConfig.WriteFile([]byte(`cli-test: "test"`))
	envConfig.WriteFile([]byte(`env-test: "test"`))

	// No flag nor env specified.
	stdout, _, err := cli.Run("config", "dump", "--json")
	require.NoError(t, err)
	requirejson.NotEmpty(t, stdout)

	// Flag specified
	stdout, _, err = cli.Run("config", "dump", "--config-file", cliConfig.String(), "--json")
	require.NoError(t, err)
	requirejson.Contains(t, stdout, `{"config":{ "cli-test": "test" }}`)

	// Env specified
	customEnv := map[string]string{"ARDUINO_CONFIG_FILE": envConfig.String()}
	stdout, _, err = cli.RunWithCustomEnv(customEnv, "config", "dump", "--json")
	require.NoError(t, err)
	requirejson.Contains(t, stdout, `{"config":{ "env-test": "test" }}`)

	// Flag and env specified, flag takes precedence
	stdout, _, err = cli.RunWithCustomEnv(customEnv, "config", "dump", "--config-file", cliConfig.String(), "--json")
	require.NoError(t, err)
	requirejson.Contains(t, stdout, `{"config":{ "cli-test": "test" }}`)
}
