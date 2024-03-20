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
	"reflect"

	"github.com/arduino/arduino-cli/internal/cli/configuration"
	"github.com/arduino/arduino-cli/internal/cli/feedback"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func initRemoveCommand(defaultSettings *configuration.Settings) *cobra.Command {
	removeCommand := &cobra.Command{
		Use:   "remove",
		Short: tr("Removes one or more values from a setting."),
		Long:  tr("Removes one or more values from a setting."),
		Example: "" +
			"  " + os.Args[0] + " config remove board_manager.additional_urls https://example.com/package_example_index.json\n" +
			"  " + os.Args[0] + " config remove board_manager.additional_urls https://example.com/package_example_index.json https://another-url.com/package_another_index.json\n",
		Args: cobra.MinimumNArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			runRemoveCommand(args, defaultSettings)
		},
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return GetSlicesConfigurationKeys(defaultSettings), cobra.ShellCompDirectiveDefault
		},
	}
	return removeCommand
}

func runRemoveCommand(args []string, defaultSettings *configuration.Settings) {
	logrus.Info("Executing `arduino-cli config remove`")
	key := args[0]
	kind := validateKey(key)

	if kind != reflect.Slice {
		msg := tr("The key '%[1]v' is not a list of items, can't remove from it.\nMaybe use '%[2]s'?", key, "config delete")
		feedback.Fatal(msg, feedback.ErrGeneric)
	}

	mappedValues := map[string]bool{}
	for _, v := range defaultSettings.GetStringSlice(key) {
		mappedValues[v] = true
	}
	for _, arg := range args[1:] {
		delete(mappedValues, arg)
	}
	values := []string{}
	for k := range mappedValues {
		values = append(values, k)
	}
	defaultSettings.Set(key, values)

	if err := defaultSettings.WriteConfig(); err != nil {
		feedback.Fatal(tr("Can't write config file: %v", err), feedback.ErrGeneric)
	}
}
