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

package config

import (
	"context"
	"os"
	"strings"

	f "github.com/arduino/arduino-cli/internal/algorithms"
	"github.com/arduino/arduino-cli/internal/cli/feedback"
	"github.com/arduino/arduino-cli/internal/i18n"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	"github.com/arduino/go-paths-helper"
	"github.com/spf13/cobra"
)

var tr = i18n.Tr

// NewCommand created a new `config` command
func NewCommand(srv rpc.ArduinoCoreServiceServer, settings *rpc.Configuration) *cobra.Command {
	configCommand := &cobra.Command{
		Use:     "config",
		Short:   tr("Arduino configuration commands."),
		Example: "  " + os.Args[0] + " config init",
	}

	configCommand.AddCommand(initAddCommand(srv))
	configCommand.AddCommand(initDeleteCommand(srv))
	configCommand.AddCommand(initDumpCommand(srv))
	configCommand.AddCommand(initGetCommand(srv))
	configCommand.AddCommand(initInitCommand(srv))
	configCommand.AddCommand(initRemoveCommand(srv))
	configCommand.AddCommand(initSetCommand(srv))

	return configCommand
}

func getAllSettingsKeys(ctx context.Context, srv rpc.ArduinoCoreServiceServer) []string {
	res, _ := srv.SettingsEnumerate(ctx, &rpc.SettingsEnumerateRequest{})
	allKeys := f.Map(res.GetEntries(), (*rpc.SettingsEnumerateResponse_Entry).GetKey)
	return allKeys
}

func getAllArraySettingsKeys(ctx context.Context, srv rpc.ArduinoCoreServiceServer) []string {
	res, _ := srv.SettingsEnumerate(ctx, &rpc.SettingsEnumerateRequest{})
	arrayEntries := f.Filter(res.GetEntries(), func(e *rpc.SettingsEnumerateResponse_Entry) bool {
		return strings.HasPrefix(e.GetType(), "[]")
	})
	arrayKeys := f.Map(arrayEntries, (*rpc.SettingsEnumerateResponse_Entry).GetKey)
	return arrayKeys
}

func saveConfiguration(ctx context.Context, srv rpc.ArduinoCoreServiceServer) {
	var outConfig []byte
	if res, err := srv.ConfigurationSave(ctx, &rpc.ConfigurationSaveRequest{SettingsFormat: "yaml"}); err != nil {
		feedback.Fatal(tr("Error writing to file: %v", err), feedback.ErrGeneric)
	} else {
		outConfig = []byte(res.GetEncodedSettings())
	}

	configFile := ctx.Value("config_file").(string)
	if err := paths.New(configFile).WriteFile(outConfig); err != nil {
		feedback.Fatal(tr("Error writing to file: %v", err), feedback.ErrGeneric)
	}
}
