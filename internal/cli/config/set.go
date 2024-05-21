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
	"encoding/json"
	"os"

	f "github.com/arduino/arduino-cli/internal/algorithms"
	"github.com/arduino/arduino-cli/internal/cli/feedback"
	"github.com/arduino/arduino-cli/internal/i18n"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func initSetCommand(srv rpc.ArduinoCoreServiceServer) *cobra.Command {
	setCommand := &cobra.Command{
		Use:   "set",
		Short: i18n.Tr("Sets a setting value."),
		Long:  i18n.Tr("Sets a setting value."),
		Example: "" +
			"  " + os.Args[0] + " config set logging.level trace\n" +
			"  " + os.Args[0] + " config set logging.file my-log.txt\n" +
			"  " + os.Args[0] + " config set sketch.always_export_binaries true\n" +
			"  " + os.Args[0] + " config set board_manager.additional_urls https://example.com/package_example_index.json https://another-url.com/package_another_index.json",
		Args: cobra.MinimumNArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			runSetCommand(cmd.Context(), srv, args)
		},
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			ctx := cmd.Context()
			return getAllSettingsKeys(ctx, srv), cobra.ShellCompDirectiveDefault
		},
	}
	return setCommand
}

func runSetCommand(ctx context.Context, srv rpc.ArduinoCoreServiceServer, args []string) {
	logrus.Info("Executing `arduino-cli config set`")

	req := &rpc.SettingsSetValueRequest{
		Key: args[0],
	}
	if len(args) == 2 {
		// Single value
		req.EncodedValue = args[1]
		req.ValueFormat = "cli"
	} else {
		// Uniq Array
		jsonValues, err := json.Marshal(f.Uniq(args[1:]))
		if err != nil {
			feedback.Fatal(i18n.Tr("Error setting value: %v", err), feedback.ErrGeneric)
		}
		req.EncodedValue = string(jsonValues)
		req.ValueFormat = "json"
	}

	if _, err := srv.SettingsSetValue(ctx, req); err != nil {
		feedback.Fatal(i18n.Tr("Error setting value: %v", err), feedback.ErrGeneric)
	}

	saveConfiguration(ctx, srv)
}
