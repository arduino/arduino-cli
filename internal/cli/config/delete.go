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

	"github.com/arduino/arduino-cli/internal/cli/feedback"
	"github.com/arduino/arduino-cli/internal/i18n"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func initDeleteCommand(srv rpc.ArduinoCoreServiceServer) *cobra.Command {
	deleteCommand := &cobra.Command{
		Use:   "delete",
		Short: i18n.Tr("Deletes a settings key and all its sub keys."),
		Long:  i18n.Tr("Deletes a settings key and all its sub keys."),
		Example: "" +
			"  " + os.Args[0] + " config delete board_manager\n" +
			"  " + os.Args[0] + " config delete board_manager.additional_urls",
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			ctx := cmd.Context()
			runDeleteCommand(ctx, srv, args)
		},
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			ctx := cmd.Context()
			return getAllSettingsKeys(ctx, srv), cobra.ShellCompDirectiveDefault
		},
	}
	return deleteCommand
}

func runDeleteCommand(ctx context.Context, srv rpc.ArduinoCoreServiceServer, args []string) {
	logrus.Info("Executing `arduino-cli config delete`")

	key := args[0]
	if _, err := srv.SettingsSetValue(ctx, &rpc.SettingsSetValueRequest{Key: key, EncodedValue: ""}); err != nil {
		feedback.Fatal(i18n.Tr("Cannot delete the key %[1]s: %[2]v", key, err), feedback.ErrGeneric)
	}

	saveConfiguration(ctx, srv)
}
