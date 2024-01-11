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
	"errors"
	"fmt"
	"strings"

	"github.com/arduino/arduino-cli/internal/cli/configuration"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
)

// SettingsGetAll returns a message with a string field containing all the settings
// currently in use, marshalled in JSON format.
func (s *ArduinoCoreServerImpl) SettingsGetAll(ctx context.Context, req *rpc.SettingsGetAllRequest) (*rpc.SettingsGetAllResponse, error) {
	b, err := json.Marshal(configuration.Settings.AllSettings())
	if err == nil {
		return &rpc.SettingsGetAllResponse{
			JsonData: string(b),
		}, nil
	}

	return nil, err
}

// mapper converts a map of nested maps to a map of scalar values.
// For example:
//
//	{"foo": "bar", "daemon":{"port":"420"}, "sketch": {"always_export_binaries": "true"}}
//
// would convert to:
//
//	{"foo": "bar", "daemon.port":"420", "sketch.always_export_binaries": "true"}
func mapper(toMap map[string]interface{}) map[string]interface{} {
	res := map[string]interface{}{}
	for k, v := range toMap {
		switch data := v.(type) {
		case map[string]interface{}:
			for mK, mV := range mapper(data) {
				// Concatenate keys
				res[fmt.Sprintf("%s.%s", k, mK)] = mV
			}
			// This is done to avoid skipping keys containing empty maps
			if len(data) == 0 {
				res[k] = map[string]interface{}{}
			}
		default:
			res[k] = v
		}
	}
	return res
}

// SettingsMerge applies multiple settings values at once.
func (s *ArduinoCoreServerImpl) SettingsMerge(ctx context.Context, req *rpc.SettingsMergeRequest) (*rpc.SettingsMergeResponse, error) {
	var toMerge map[string]interface{}
	if err := json.Unmarshal([]byte(req.GetJsonData()), &toMerge); err != nil {
		return nil, err
	}

	mapped := mapper(toMerge)

	// Set each value individually.
	// This is done because Viper ignores empty strings or maps when
	// using the MergeConfigMap function.
	updatedSettings := configuration.Init("")
	for k, v := range mapped {
		updatedSettings.Set(k, v)
	}
	configPath := configuration.Settings.ConfigFileUsed()
	updatedSettings.SetConfigFile(configPath)
	configuration.Settings = updatedSettings

	return &rpc.SettingsMergeResponse{}, nil
}

// SettingsGetValue returns a settings value given its key. If the key is not present
// an error will be returned, so that we distinguish empty settings from missing
// ones.
func (s *ArduinoCoreServerImpl) SettingsGetValue(ctx context.Context, req *rpc.SettingsGetValueRequest) (*rpc.SettingsGetValueResponse, error) {
	key := req.GetKey()

	// Check if settings key actually existing, we don't use Viper.InConfig()
	// since that doesn't check for keys formatted like daemon.port or those set
	// with Viper.Set(). This way we check for all existing settings for sure.
	keyExists := false
	for _, k := range configuration.Settings.AllKeys() {
		if k == key || strings.HasPrefix(k, key) {
			keyExists = true
			break
		}
	}
	if !keyExists {
		return nil, errors.New(tr("key not found in settings"))
	}

	b, err := json.Marshal(configuration.Settings.Get(key))
	value := &rpc.SettingsGetValueResponse{}
	if err == nil {
		value.Key = key
		value.JsonData = string(b)
	}

	return value, err
}

// SettingsSetValue updates or set a value for a certain key.
func (s *ArduinoCoreServerImpl) SettingsSetValue(ctx context.Context, val *rpc.SettingsSetValueRequest) (*rpc.SettingsSetValueResponse, error) {
	key := val.GetKey()
	var value interface{}

	err := json.Unmarshal([]byte(val.GetJsonData()), &value)
	if err == nil {
		configuration.Settings.Set(key, value)
	}

	return &rpc.SettingsSetValueResponse{}, err
}

// SettingsWrite to file set in request the settings currently stored in memory.
// We don't have a Read() function, that's not necessary since we only want one config file to be used
// and that's picked up when the CLI is run as daemon, either using the default path or a custom one
// set with the --config-file flag.
func (s *ArduinoCoreServerImpl) SettingsWrite(ctx context.Context, req *rpc.SettingsWriteRequest) (*rpc.SettingsWriteResponse, error) {
	if err := configuration.Settings.WriteConfigAs(req.GetFilePath()); err != nil {
		return nil, err
	}
	return &rpc.SettingsWriteResponse{}, nil
}

// SettingsDelete removes a key from the config file
func (s *ArduinoCoreServerImpl) SettingsDelete(ctx context.Context, req *rpc.SettingsDeleteRequest) (*rpc.SettingsDeleteResponse, error) {
	toDelete := req.GetKey()

	// Check if settings key actually existing, we don't use Viper.InConfig()
	// since that doesn't check for keys formatted like daemon.port or those set
	// with Viper.Set(). This way we check for all existing settings for sure.
	keyExists := false
	keys := []string{}
	for _, k := range configuration.Settings.AllKeys() {
		if !strings.HasPrefix(k, toDelete) {
			keys = append(keys, k)
			continue
		}
		keyExists = true
	}

	if !keyExists {
		return nil, errors.New(tr("key not found in settings"))
	}

	// Override current settings to delete the key
	updatedSettings := configuration.Init("")
	for _, k := range keys {
		updatedSettings.Set(k, configuration.Settings.Get(k))
	}
	configPath := configuration.Settings.ConfigFileUsed()
	updatedSettings.SetConfigFile(configPath)
	configuration.Settings = updatedSettings

	return &rpc.SettingsDeleteResponse{}, nil
}
