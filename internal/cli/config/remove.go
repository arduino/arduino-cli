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
	"slices"

	"github.com/arduino/arduino-cli/internal/cli/feedback"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func initRemoveCommand(srv rpc.ArduinoCoreServiceServer) *cobra.Command {
	removeCommand := &cobra.Command{
		Use:   "remove",
		Short: tr("Removes one or more values from a setting."),
		Long:  tr("Removes one or more values from a setting."),
		Example: "" +
			"  " + os.Args[0] + " config remove board_manager.additional_urls https://example.com/package_example_index.json\n" +
			"  " + os.Args[0] + " config remove board_manager.additional_urls https://example.com/package_example_index.json https://another-url.com/package_another_index.json\n",
		Args: cobra.MinimumNArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			ctx := cmd.Context()
			runRemoveCommand(ctx, srv, args)
		},
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			ctx := cmd.Context()
			return getAllArraySettingsKeys(ctx, srv), cobra.ShellCompDirectiveDefault
		},
	}
	return removeCommand
}

func runRemoveCommand(ctx context.Context, srv rpc.ArduinoCoreServiceServer, args []string) {
	logrus.Info("Executing `arduino-cli config remove`")
	key := args[0]

	if !slices.Contains(getAllArraySettingsKeys(ctx, srv), key) {
		msg := tr("The key '%[1]v' is not a list of items, can't remove from it.\nMaybe use '%[2]s'?", key, "config delete")
		feedback.Fatal(msg, feedback.ErrGeneric)
	}

	var currentValues []string
	if resp, err := srv.SettingsGetValue(ctx, &rpc.SettingsGetValueRequest{Key: key}); err != nil {
		feedback.Fatal(tr("Cannot get the configuration key %[1]s: %[2]v", key, err), feedback.ErrGeneric)
	} else if err := json.Unmarshal([]byte(resp.GetEncodedValue()), &currentValues); err != nil {
		feedback.Fatal(tr("Cannot get the configuration key %[1]s: %[2]v", key, err), feedback.ErrGeneric)
	}

	for _, arg := range args[1:] {
		currentValues = slices.DeleteFunc(currentValues, func(in string) bool { return in == arg })
	}

	if newValuesJSON, err := json.Marshal(currentValues); err != nil {
		feedback.Fatal(tr("Cannot remove the configuration key %[1]s: %[2]v", key, err), feedback.ErrGeneric)
	} else if _, err := srv.SettingsSetValue(ctx, &rpc.SettingsSetValueRequest{Key: key, EncodedValue: string(newValuesJSON)}); err != nil {
		feedback.Fatal(tr("Cannot remove the configuration key %[1]s: %[2]v", key, err), feedback.ErrGeneric)
	}

	saveConfiguration(ctx, srv)
}
