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
	"fmt"
	"reflect"

	"github.com/arduino/arduino-cli/commands/cmderrors"
	f "github.com/arduino/arduino-cli/internal/algorithms"
	"github.com/arduino/arduino-cli/internal/cli/configuration"
	"github.com/arduino/arduino-cli/internal/go-configmap"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	"google.golang.org/protobuf/proto"
	"gopkg.in/yaml.v3"
)

// SettingsGetValue returns a settings value given its key. If the key is not present
// an error will be returned, so that we distinguish empty settings from missing
// ones.
func (s *arduinoCoreServerImpl) ConfigurationGet(ctx context.Context, req *rpc.ConfigurationGetRequest) (*rpc.ConfigurationGetResponse, error) {
	conf := &rpc.Configuration{
		Directories: &rpc.Configuration_Directories{
			Builtin:   &rpc.Configuration_Directories_Builtin{},
			Data:      s.settings.DataDir().String(),
			Downloads: s.settings.DownloadsDir().String(),
			User:      s.settings.UserDir().String(),
		},
		Network: &rpc.Configuration_Network{},
		Sketch: &rpc.Configuration_Sketch{
			AlwaysExportBinaries: s.settings.SketchAlwaysExportBinaries(),
		},
		BuildCache: &rpc.Configuration_BuildCache{
			CompilationsBeforePurge: uint64(s.settings.GetCompilationsBeforeBuildCachePurge()),
			TtlSecs:                 uint64(s.settings.GetBuildCacheTTL().Seconds()),
		},
		BoardManager: &rpc.Configuration_BoardManager{
			AdditionalUrls: s.settings.BoardManagerAdditionalUrls(),
		},
		Daemon: &rpc.Configuration_Daemon{
			Port: s.settings.DaemonPort(),
		},
		Output: &rpc.Configuration_Output{
			NoColor: s.settings.NoColor(),
		},
		Logging: &rpc.Configuration_Logging{
			Level:  s.settings.LoggingLevel(),
			Format: s.settings.LoggingFormat(),
		},
		Library: &rpc.Configuration_Library{
			EnableUnsafeInstall: s.settings.LibraryEnableUnsafeInstall(),
		},
		Updater: &rpc.Configuration_Updater{
			EnableNotification: s.settings.UpdaterEnableNotification(),
		},
	}

	if builtinLibs := s.settings.IDEBuiltinLibrariesDir(); builtinLibs != nil {
		conf.Directories.Builtin.Libraries = proto.String(builtinLibs.String())
	}

	if ua := s.settings.ExtraUserAgent(); ua != "" {
		conf.Network.ExtraUserAgent = &ua
	}
	if proxy, err := s.settings.NetworkProxy(); err == nil && proxy != nil {
		conf.Network.Proxy = proto.String(proxy.String())
	}

	if logFile := s.settings.LoggingFile(); logFile != nil {
		file := logFile.String()
		conf.Logging.File = &file
	}

	if locale := s.settings.Locale(); locale != "" {
		conf.Locale = &locale
	}

	return &rpc.ConfigurationGetResponse{Configuration: conf}, nil
}

func (s *arduinoCoreServerImpl) SettingsSetValue(ctx context.Context, req *rpc.SettingsSetValueRequest) (*rpc.SettingsSetValueResponse, error) {
	// Determine the existence and the kind of the value
	key := req.GetKey()

	// Extract the value from the request
	encodedValue := []byte(req.GetEncodedValue())
	if len(encodedValue) == 0 {
		// If the value is empty, unset the key
		s.settings.Delete(key)
		return &rpc.SettingsSetValueResponse{}, nil
	}

	var newValue any
	switch req.GetValueFormat() {
	case "", "json":
		if err := json.Unmarshal(encodedValue, &newValue); err != nil {
			return nil, &cmderrors.InvalidArgumentError{Message: fmt.Sprintf("invalid value: %v", err)}
		}
	case "yaml":
		if err := yaml.Unmarshal(encodedValue, &newValue); err != nil {
			return nil, &cmderrors.InvalidArgumentError{Message: fmt.Sprintf("invalid value: %v", err)}
		}
	case "cli":
		err := s.settings.SetFromCLIArgs(key, req.GetEncodedValue())
		if err != nil {
			return nil, err
		}
		return &rpc.SettingsSetValueResponse{}, nil
	default:
		return nil, &cmderrors.InvalidArgumentError{Message: fmt.Sprintf("unsupported value format: %s", req.ValueFormat)}
	}

	// If the value is "null", unset the key
	if reflect.TypeOf(newValue) == reflect.TypeOf(nil) {
		s.settings.Delete(key)
		return &rpc.SettingsSetValueResponse{}, nil
	}

	// Set the value
	if err := s.settings.Set(key, newValue); err != nil {
		return nil, err
	}

	return &rpc.SettingsSetValueResponse{}, nil
}

func (s *arduinoCoreServerImpl) SettingsGetValue(ctx context.Context, req *rpc.SettingsGetValueRequest) (*rpc.SettingsGetValueResponse, error) {
	key := req.GetKey()
	value, ok := s.settings.GetOk(key)
	if !ok {
		value, ok = s.settings.Defaults.GetOk(key)
	}
	if !ok {
		return nil, &cmderrors.InvalidArgumentError{Message: fmt.Sprintf("key %s not found", key)}
	}

	switch req.GetValueFormat() {
	case "", "json":
		valueJson, err := json.Marshal(value)
		if err != nil {
			return nil, fmt.Errorf("error marshalling value: %v", err)
		}
		return &rpc.SettingsGetValueResponse{EncodedValue: string(valueJson)}, nil
	case "yaml":
		valueYaml, err := yaml.Marshal(value)
		if err != nil {
			return nil, fmt.Errorf("error marshalling value: %v", err)
		}
		return &rpc.SettingsGetValueResponse{EncodedValue: string(valueYaml)}, nil
	default:
		return nil, &cmderrors.InvalidArgumentError{Message: fmt.Sprintf("unsupported value format: %s", req.ValueFormat)}
	}
}

// ConfigurationSave encodes the current configuration in the specified format
func (s *arduinoCoreServerImpl) ConfigurationSave(ctx context.Context, req *rpc.ConfigurationSaveRequest) (*rpc.ConfigurationSaveResponse, error) {
	switch req.GetSettingsFormat() {
	case "yaml":
		data, err := yaml.Marshal(s.settings)
		if err != nil {
			return nil, fmt.Errorf("error marshalling settings: %v", err)
		}
		return &rpc.ConfigurationSaveResponse{EncodedSettings: string(data)}, nil
	case "json":
		data, err := json.MarshalIndent(s.settings, "", "  ")
		if err != nil {
			return nil, fmt.Errorf("error marshalling settings: %v", err)
		}
		return &rpc.ConfigurationSaveResponse{EncodedSettings: string(data)}, nil
	default:
		return nil, &cmderrors.InvalidArgumentError{Message: fmt.Sprintf("unsupported format: %s", req.GetSettingsFormat())}
	}
}

// SettingsReadFromFile read settings from a YAML file and replace the settings currently stored in memory.
func (s *arduinoCoreServerImpl) ConfigurationOpen(ctx context.Context, req *rpc.ConfigurationOpenRequest) (*rpc.ConfigurationOpenResponse, error) {
	warnings := []string{}

	switch req.GetSettingsFormat() {
	case "yaml":
		err := yaml.Unmarshal([]byte(req.GetEncodedSettings()), s.settings)
		if errs, ok := err.(*configmap.UnmarshalErrors); ok {
			warnings = f.Map(errs.WrappedErrors(), (error).Error)
		} else if err != nil {
			return nil, fmt.Errorf("error unmarshalling settings: %v", err)
		}
	case "json":
		err := json.Unmarshal([]byte(req.GetEncodedSettings()), s.settings)
		if errs, ok := err.(*configmap.UnmarshalErrors); ok {
			warnings = f.Map(errs.WrappedErrors(), (error).Error)
		} else if err != nil {
			return nil, fmt.Errorf("error unmarshalling settings: %v", err)
		}
	default:
		return nil, &cmderrors.InvalidArgumentError{Message: fmt.Sprintf("unsupported format: %s", req.GetSettingsFormat())}
	}

	configuration.InjectEnvVars(s.settings)
	return &rpc.ConfigurationOpenResponse{Warnings: warnings}, nil
}

// SettingsEnumerate returns the list of all the settings keys.
func (s *arduinoCoreServerImpl) SettingsEnumerate(ctx context.Context, req *rpc.SettingsEnumerateRequest) (*rpc.SettingsEnumerateResponse, error) {
	var entries []*rpc.SettingsEnumerateResponse_Entry
	for k, t := range s.settings.Defaults.Schema() {
		entries = append(entries, &rpc.SettingsEnumerateResponse_Entry{
			Key:  k,
			Type: t.String(),
		})
	}
	return &rpc.SettingsEnumerateResponse{
		Entries: entries,
	}, nil
}
