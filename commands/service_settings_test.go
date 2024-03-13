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

package commands

import (
	"context"
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/arduino/arduino-cli/internal/cli/configuration"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	"github.com/arduino/go-paths-helper"
	"github.com/stretchr/testify/require"
)

var svc = NewArduinoCoreServer("")

func init() {
	configuration.Settings = configuration.Init(filepath.Join("testdata", "arduino-cli.yaml"))
}

func reset() {
	configuration.Settings = configuration.Init(filepath.Join("testdata", "arduino-cli.yaml"))
}

func TestGetAll(t *testing.T) {
	resp, err := svc.SettingsGetAll(context.Background(), &rpc.SettingsGetAllRequest{})
	require.Nil(t, err)

	content, err := json.Marshal(configuration.Settings.AllSettings())
	require.Nil(t, err)

	require.Equal(t, string(content), resp.GetJsonData())
}

func TestMerge(t *testing.T) {
	// Verify defaults
	require.Equal(t, "50051", configuration.Settings.GetString("daemon.port"))
	require.Equal(t, "", configuration.Settings.GetString("foo"))
	require.Equal(t, false, configuration.Settings.GetBool("sketch.always_export_binaries"))

	bulkSettings := `{"foo": "bar", "daemon":{"port":"420"}, "sketch": {"always_export_binaries": "true"}}`
	res, err := svc.SettingsMerge(context.Background(), &rpc.SettingsMergeRequest{JsonData: bulkSettings})
	require.NotNil(t, res)
	require.NoError(t, err)

	require.Equal(t, "420", configuration.Settings.GetString("daemon.port"))
	require.Equal(t, "bar", configuration.Settings.GetString("foo"))
	require.Equal(t, true, configuration.Settings.GetBool("sketch.always_export_binaries"))

	bulkSettings = `{"foo":"", "daemon": {}, "sketch": {"always_export_binaries": "false"}}`
	res, err = svc.SettingsMerge(context.Background(), &rpc.SettingsMergeRequest{JsonData: bulkSettings})
	require.NotNil(t, res)
	require.NoError(t, err)

	require.Equal(t, "50051", configuration.Settings.GetString("daemon.port"))
	require.Equal(t, "", configuration.Settings.GetString("foo"))
	require.Equal(t, false, configuration.Settings.GetBool("sketch.always_export_binaries"))

	bulkSettings = `{"daemon": {"port":""}}`
	res, err = svc.SettingsMerge(context.Background(), &rpc.SettingsMergeRequest{JsonData: bulkSettings})
	require.NotNil(t, res)
	require.NoError(t, err)

	require.Equal(t, "", configuration.Settings.GetString("daemon.port"))
	// Verifies other values are not changed
	require.Equal(t, "", configuration.Settings.GetString("foo"))
	require.Equal(t, false, configuration.Settings.GetBool("sketch.always_export_binaries"))

	bulkSettings = `{"network": {}}`
	res, err = svc.SettingsMerge(context.Background(), &rpc.SettingsMergeRequest{JsonData: bulkSettings})
	require.NotNil(t, res)
	require.NoError(t, err)

	require.Equal(t, "", configuration.Settings.GetString("proxy"))

	reset()
}

func TestGetValue(t *testing.T) {
	key := &rpc.SettingsGetValueRequest{Key: "daemon"}
	resp, err := svc.SettingsGetValue(context.Background(), key)
	require.NoError(t, err)
	require.Equal(t, `{"port":"50051"}`, resp.GetJsonData())

	key = &rpc.SettingsGetValueRequest{Key: "daemon.port"}
	resp, err = svc.SettingsGetValue(context.Background(), key)
	require.NoError(t, err)
	require.Equal(t, `"50051"`, resp.GetJsonData())
}

func TestGetMergedValue(t *testing.T) {
	// Verifies value is not set
	key := &rpc.SettingsGetValueRequest{Key: "foo"}
	res, err := svc.SettingsGetValue(context.Background(), key)
	require.Nil(t, res)
	require.Error(t, err, "Error getting settings value")

	// Merge value
	bulkSettings := `{"foo": "bar"}`
	_, err = svc.SettingsMerge(context.Background(), &rpc.SettingsMergeRequest{JsonData: bulkSettings})
	require.NoError(t, err)

	// Verifies value is correctly returned
	key = &rpc.SettingsGetValueRequest{Key: "foo"}
	res, err = svc.SettingsGetValue(context.Background(), key)
	require.NoError(t, err)
	require.Equal(t, `"bar"`, res.GetJsonData())

	reset()
}

func TestGetValueNotFound(t *testing.T) {
	key := &rpc.SettingsGetValueRequest{Key: "DOESNTEXIST"}
	_, err := svc.SettingsGetValue(context.Background(), key)
	require.Error(t, err)
}

func TestSetValue(t *testing.T) {
	val := &rpc.SettingsSetValueRequest{
		Key:      "foo",
		JsonData: `"bar"`,
	}
	_, err := svc.SettingsSetValue(context.Background(), val)
	require.Nil(t, err)
	require.Equal(t, "bar", configuration.Settings.GetString("foo"))
}

func TestWrite(t *testing.T) {
	// Writes some settings
	val := &rpc.SettingsSetValueRequest{
		Key:      "foo",
		JsonData: `"bar"`,
	}
	_, err := svc.SettingsSetValue(context.Background(), val)
	require.NoError(t, err)

	tempDir := paths.TempDir()
	testFolder, err := tempDir.MkTempDir("testdata")
	require.NoError(t, err)
	defer testFolder.RemoveAll()

	// Verifies config files doesn't exist
	configFile := testFolder.Join("arduino-cli.yml")
	require.True(t, configFile.NotExist())

	_, err = svc.SettingsWrite(context.Background(), &rpc.SettingsWriteRequest{
		FilePath: configFile.String(),
	})
	require.NoError(t, err)

	// Verifies config file is created.
	// We don't verify the content since we expect config library, Viper, to work
	require.True(t, configFile.Exist())
}

func TestDelete(t *testing.T) {
	_, err := svc.SettingsDelete(context.Background(), &rpc.SettingsDeleteRequest{
		Key: "doesnotexist",
	})
	require.Error(t, err)

	_, err = svc.SettingsDelete(context.Background(), &rpc.SettingsDeleteRequest{
		Key: "network",
	})
	require.NoError(t, err)

	_, err = svc.SettingsGetValue(context.Background(), &rpc.SettingsGetValueRequest{Key: "network"})
	require.Error(t, err)
}
