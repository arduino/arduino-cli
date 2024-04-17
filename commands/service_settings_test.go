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
	"testing"

	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	"github.com/arduino/go-paths-helper"
	"github.com/stretchr/testify/require"
)

func loadConfig(t *testing.T, srv rpc.ArduinoCoreServiceServer, confPath *paths.Path) {
	confPath.ToAbs()
	conf, err := confPath.ReadFile()
	require.NoError(t, err)
	_, err = srv.ConfigurationOpen(context.Background(), &rpc.ConfigurationOpenRequest{
		EncodedSettings: string(conf),
		SettingsFormat:  "yaml",
	})
	require.NoError(t, err)
}

func TestGetAll(t *testing.T) {
	srv := NewArduinoCoreServer()
	loadConfig(t, srv, paths.New("testdata", "arduino-cli.yml"))
	resp, err := srv.ConfigurationGet(context.Background(), &rpc.ConfigurationGetRequest{})
	require.Nil(t, err)

	defaultUserDir, err := srv.SettingsGetValue(context.Background(), &rpc.SettingsGetValueRequest{Key: "directories.user"})
	require.NoError(t, err)

	content, err := json.Marshal(resp.GetConfiguration())
	require.Nil(t, err)
	require.JSONEq(t, `{
		"board_manager": {
			"additional_urls": [ "http://foobar.com", "http://example.com" ]
		},
		"build_cache": {
			"compilations_before_purge": 10,
			"ttl_secs": 2592000
		},
		"directories": {
			"builtin": {},
			"data": "/home/massi/.arduino15",
			"downloads": "/home/massi/.arduino15/staging",
			"user": `+defaultUserDir.GetEncodedValue()+`
		},
		"library": {},
		"locale": "en",
		"logging": {
			"format": "text",
			"level": "info"
		},
		"daemon":{
			"port":"50051"
		},
		"network":{
			"proxy":"123"
		},
		"output": {},
		"sketch": {},
		"updater": {}
	}`, string(content))
}

func TestMerge(t *testing.T) {
	srv := NewArduinoCoreServer()
	loadConfig(t, srv, paths.New("testdata", "arduino-cli.yml"))

	ctx := context.Background()

	get := func(key string) string {
		resp, err := srv.SettingsGetValue(ctx, &rpc.SettingsGetValueRequest{Key: key})
		if err != nil {
			return "<error>"
		}
		return resp.GetEncodedValue()
	}

	// Verify defaults
	require.Equal(t, `"50051"`, get("daemon.port"))
	require.Equal(t, "<error>", get("foo"))
	require.Equal(t, "false", get("sketch.always_export_binaries"))

	bulkSettings := `{"foo": "bar", "daemon":{"port":"420"}, "sketch": {"always_export_binaries": "true"}}`
	_, err := srv.ConfigurationOpen(ctx, &rpc.ConfigurationOpenRequest{EncodedSettings: bulkSettings, SettingsFormat: "json"})
	require.Error(t, err)

	bulkSettings = `{"daemon":{"port":"420"}, "sketch": {"always_export_binaries": "true"}}`
	_, err = srv.ConfigurationOpen(ctx, &rpc.ConfigurationOpenRequest{EncodedSettings: bulkSettings, SettingsFormat: "json"})
	require.Error(t, err)

	bulkSettings = `{"daemon":{"port":"420"}, "sketch": {"always_export_binaries": true}}`
	_, err = srv.ConfigurationOpen(ctx, &rpc.ConfigurationOpenRequest{EncodedSettings: bulkSettings, SettingsFormat: "json"})
	require.NoError(t, err)

	require.Equal(t, `"420"`, get("daemon.port"))
	require.Equal(t, `<error>`, get("foo"))
	require.Equal(t, "true", get("sketch.always_export_binaries"))

	bulkSettings = `{"daemon": {}, "sketch": {"always_export_binaries": false}}`
	_, err = srv.ConfigurationOpen(ctx, &rpc.ConfigurationOpenRequest{EncodedSettings: bulkSettings, SettingsFormat: "json"})
	require.NoError(t, err)

	require.Equal(t, `"50051"`, get("daemon.port"))
	require.Equal(t, "<error>", get("foo"))
	require.Equal(t, "false", get("sketch.always_export_binaries"))

	_, err = srv.SettingsSetValue(ctx, &rpc.SettingsSetValueRequest{Key: "daemon.port", EncodedValue: ""})
	require.NoError(t, err)

	require.Equal(t, `"50051"`, get("daemon.port"))
	// Verifies other values are not changed
	require.Equal(t, "<error>", get("foo"))
	require.Equal(t, "false", get("sketch.always_export_binaries"))

}

func TestGetValue(t *testing.T) {
	srv := NewArduinoCoreServer()
	loadConfig(t, srv, paths.New("testdata", "arduino-cli.yml"))

	key := &rpc.SettingsGetValueRequest{Key: "daemon"}
	resp, err := srv.SettingsGetValue(context.Background(), key)
	require.NoError(t, err)
	require.Equal(t, `{"port":"50051"}`, resp.GetEncodedValue())

	key = &rpc.SettingsGetValueRequest{Key: "daemon.port"}
	resp, err = srv.SettingsGetValue(context.Background(), key)
	require.NoError(t, err)
	require.Equal(t, `"50051"`, resp.GetEncodedValue())
}

func TestGetOfSettedValue(t *testing.T) {
	srv := NewArduinoCoreServer()
	loadConfig(t, srv, paths.New("testdata", "arduino-cli.yml"))

	// Verifies value is not set (try with a key without a default, like "directories.builtin.libraries")
	ctx := context.Background()
	res, err := srv.SettingsGetValue(ctx, &rpc.SettingsGetValueRequest{Key: "directories.builtin.libraries"})
	require.Nil(t, res)
	require.Error(t, err, "Error getting settings value")

	// Set value
	_, err = srv.SettingsSetValue(ctx, &rpc.SettingsSetValueRequest{
		Key:          "directories.builtin.libraries",
		EncodedValue: `"bar"`})
	require.NoError(t, err)

	// Verifies value is correctly returned
	res, err = srv.SettingsGetValue(ctx, &rpc.SettingsGetValueRequest{Key: "directories.builtin.libraries"})
	require.NoError(t, err)
	require.Equal(t, `"bar"`, res.GetEncodedValue())
}

func TestGetValueNotFound(t *testing.T) {
	srv := NewArduinoCoreServer()
	loadConfig(t, srv, paths.New("testdata", "arduino-cli.yml"))

	key := &rpc.SettingsGetValueRequest{Key: "DOESNTEXIST"}
	_, err := srv.SettingsGetValue(context.Background(), key)
	require.Error(t, err)
}

func TestWrite(t *testing.T) {
	srv := NewArduinoCoreServer()
	loadConfig(t, srv, paths.New("testdata", "arduino-cli.yml"))

	// Writes some settings
	val := &rpc.SettingsSetValueRequest{
		Key:          "directories.builtin.libraries",
		EncodedValue: `"bar"`,
	}
	_, err := srv.SettingsSetValue(context.Background(), val)
	require.NoError(t, err)

	resp, err := srv.ConfigurationSave(context.Background(), &rpc.ConfigurationSaveRequest{
		SettingsFormat: "yaml",
	})
	require.NoError(t, err)

	// Verify encoded content
	require.YAMLEq(t, `
board_manager:
  additional_urls:
    - http://foobar.com
    - http://example.com

daemon:
  port: "50051"

directories:
  data: /home/massi/.arduino15
  downloads: /home/massi/.arduino15/staging
  builtin:
    libraries: bar

logging:
  file: ""
  format: text
  level: info

network:
  proxy: "123"
`, resp.GetEncodedSettings())
}

func TestDelete(t *testing.T) {
	srv := NewArduinoCoreServer()
	loadConfig(t, srv, paths.New("testdata", "arduino-cli.yml"))

	_, err := srv.SettingsGetValue(context.Background(), &rpc.SettingsGetValueRequest{Key: "network"})
	require.NoError(t, err)

	_, err = srv.SettingsSetValue(context.Background(), &rpc.SettingsSetValueRequest{Key: "network", EncodedValue: ""})
	require.NoError(t, err)

	_, err = srv.SettingsGetValue(context.Background(), &rpc.SettingsGetValueRequest{Key: "network"})
	require.Error(t, err)
}
