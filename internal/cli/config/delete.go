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

	"github.com/arduino/arduino-cli/internal/cli/configuration"
	"github.com/arduino/arduino-cli/internal/cli/feedback"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func initDeleteCommand(srv rpc.ArduinoCoreServiceServer, defaultSettings *configuration.Settings) *cobra.Command {
	configFile := defaultSettings.ConfigFileUsed()
	deleteCommand := &cobra.Command{
		Use:   "delete",
		Short: tr("Deletes a settings key and all its sub keys."),
		Long:  tr("Deletes a settings key and all its sub keys."),
		Example: "" +
			"  " + os.Args[0] + " config delete board_manager\n" +
			"  " + os.Args[0] + " config delete board_manager.additional_urls",
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			runDeleteCommand(srv, args, configFile)
		},
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return defaultSettings.AllKeys(), cobra.ShellCompDirectiveDefault
		},
	}
	return deleteCommand
}

func runDeleteCommand(srv rpc.ArduinoCoreServiceServer, args []string, configFile string) {
	logrus.Info("Executing `arduino-cli config delete`")
	ctx := context.Background()

	toDelete := args[0]
	_, err := srv.SettingsDelete(ctx, &rpc.SettingsDeleteRequest{Key: toDelete})
	if err != nil {
		feedback.Fatal(tr("Cannot delete the key %[1]s: %[2]v", toDelete, err), feedback.ErrGeneric)
	}
	_, err = srv.SettingsWrite(ctx, &rpc.SettingsWriteRequest{FilePath: configFile})
	if err != nil {
		feedback.Fatal(tr("Cannot write the file %[1]s: %[2]v", configFile, err), feedback.ErrGeneric)
	}
}
