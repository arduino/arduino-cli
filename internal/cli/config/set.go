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
	"strconv"

	"github.com/arduino/arduino-cli/configuration"
	"github.com/arduino/arduino-cli/internal/cli/feedback"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func initSetCommand() *cobra.Command {
	setCommand := &cobra.Command{
		Use:   "set",
		Short: tr("Sets a setting value."),
		Long:  tr("Sets a setting value."),
		Example: "" +
			"  " + os.Args[0] + " config set logging.level trace\n" +
			"  " + os.Args[0] + " config set logging.file my-log.txt\n" +
			"  " + os.Args[0] + " config set sketch.always_export_binaries true\n" +
			"  " + os.Args[0] + " config set board_manager.additional_urls https://example.com/package_example_index.json https://another-url.com/package_another_index.json",
		Args: cobra.MinimumNArgs(2),
		Run:  runSetCommand,
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return configuration.Settings.AllKeys(), cobra.ShellCompDirectiveDefault
		},
	}
	return setCommand
}

func runSetCommand(cmd *cobra.Command, args []string) {
	logrus.Info("Executing `arduino-cli config set`")
	key := args[0]
	kind := validateKey(key)

	if kind != reflect.Slice && len(args) > 2 {
		feedback.Fatal(tr("Can't set multiple values in key %v", key), feedback.ErrGeneric)
	}

	var value interface{}
	switch kind {
	case reflect.Slice:
		value = uniquify(args[1:])
	case reflect.String:
		value = args[1]
	case reflect.Bool:
		var err error
		value, err = strconv.ParseBool(args[1])
		if err != nil {
			feedback.Fatal(tr("error parsing value: %v", err), feedback.ErrGeneric)
		}
	}

	configuration.Settings.Set(key, value)

	if err := configuration.Settings.WriteConfig(); err != nil {
		feedback.Fatal(tr("Writing config file: %v", err), feedback.ErrGeneric)
	}
}
