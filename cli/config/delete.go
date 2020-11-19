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
	"os"
	"strings"

	"github.com/arduino/arduino-cli/cli/errorcodes"
	"github.com/arduino/arduino-cli/cli/feedback"
	"github.com/arduino/arduino-cli/configuration"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func initDeleteCommand() *cobra.Command {
	addCommand := &cobra.Command{
		Use:   "delete",
		Short: "Deletes a settings key and all its sub keys.",
		Long:  "Deletes a settings key and all its sub keys.",
		Example: "" +
			"  " + os.Args[0] + " config delete board_manager\n" +
			"  " + os.Args[0] + " config delete board_manager.additional_urls",
		Args: cobra.ExactArgs(1),
		Run:  runDeleteCommand,
	}
	return addCommand
}

func runDeleteCommand(cmd *cobra.Command, args []string) {
	toDelete := args[0]

	keys := []string{}
	exists := false
	for _, v := range configuration.Settings.AllKeys() {
		if !strings.HasPrefix(v, toDelete) {
			keys = append(keys, v)
			continue
		}
		exists = true
	}

	if !exists {
		feedback.Errorf("Settings key doesn't exist")
		os.Exit(errorcodes.ErrGeneric)
	}

	updatedSettings := viper.New()
	for _, k := range keys {
		updatedSettings.Set(k, configuration.Settings.Get(k))
	}

	if err := updatedSettings.WriteConfigAs(configuration.Settings.ConfigFileUsed()); err != nil {
		feedback.Errorf("Can't write config file: %v", err)
		os.Exit(errorcodes.ErrGeneric)
	}
}
