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
	"context"
	"encoding/json"
	"testing"

	"github.com/spf13/viper"

	"github.com/arduino/arduino-cli/configuration"
	rpc "github.com/arduino/arduino-cli/rpc/settings"
	"github.com/stretchr/testify/require"
)

var svc = SettingsService{}

func init() {
	configuration.Init("testdata")
}

func reset() {
	viper.Reset()
	configuration.Init("testdata")
}

func TestGetAll(t *testing.T) {
	resp, err := svc.GetAll(context.Background(), &rpc.GetAllRequest{})
	require.Nil(t, err)

	content, err := json.Marshal(viper.AllSettings())
	require.Nil(t, err)

	require.Equal(t, string(content), resp.GetJsonData())
}

func TestMerge(t *testing.T) {
	bulkSettings := `{"foo": "bar", "daemon":{"port":"420"}}`
	_, err := svc.Merge(context.Background(), &rpc.RawData{JsonData: bulkSettings})
	require.Nil(t, err)

	require.Equal(t, "420", viper.GetString("daemon.port"))
	require.Equal(t, "bar", viper.GetString("foo"))

	reset()
}

func TestGetValue(t *testing.T) {
	key := &rpc.GetValueRequest{Key: "daemon"}
	resp, err := svc.GetValue(context.Background(), key)
	require.Nil(t, err)
	require.Equal(t, `{"port":"50051"}`, resp.GetJsonData())
}

func TestGetValueNotFound(t *testing.T) {
	key := &rpc.GetValueRequest{Key: "DOESNTEXIST"}
	_, err := svc.GetValue(context.Background(), key)
	require.NotNil(t, err)
	require.Equal(t, `key not found in settings`, err.Error())
}

func TestSetValue(t *testing.T) {
	val := &rpc.Value{
		Key:      "foo",
		JsonData: `"bar"`,
	}
	_, err := svc.SetValue(context.Background(), val)
	require.Nil(t, err)
	require.Equal(t, "bar", viper.GetString("foo"))
}
