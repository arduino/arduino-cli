/*
 * This file is part of arduino-cli.
 *
 * Copyright 2018 ARDUINO SA (http://www.arduino.cc/)
 *
 * This software is released under the GNU General Public License version 3,
 * which covers the main part of arduino-cli.
 * The terms of this license can be found at:
 * https://www.gnu.org/licenses/gpl-3.0.en.html
 *
 * You can be released from the requirements of the above licenses by purchasing
 * a commercial license. Buying such a license is mandatory if you want to modify or
 * otherwise use the software for commercial activities involving the Arduino
 * software without disclosing the source code of your own applications. To purchase
 * a commercial license, send an email to license@arduino.cc.
 */

package config

import (
	"os"

	"github.com/arduino/arduino-cli/commands"
	"github.com/arduino/arduino-cli/common/formatter"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func initInitCommand() *cobra.Command {
	initCommand := &cobra.Command{
		Use:   "init",
		Short: "Initializes a new config file into the default location.",
		Long:  "Initializes a new config file into the default location ($EXE_DIR/cli-config.yml).",
		Example: "" +
			"  # Creates a config file by asking questions to the user into the default location.\n" +
			"  " + commands.AppName + " config init\n\n" +
			"  # Creates a config file with default configuration into default location.\n" +
			"  " + commands.AppName + " config init --default\n",
		Args: cobra.NoArgs,
		Run:  runInitCommand,
	}
	initCommand.Flags().BoolVar(&initFlags._default,
		"default",
		false,
		"If omitted, ask questions to the user about setting configuration properties, otherwise use default configuration.")
	initCommand.Flags().StringVar(&initFlags.location,
		"save-as",
		"",
		"Sets where to save the configuration file [default is ./.cli-config.yml].")
	return initCommand
}

var initFlags struct {
	_default bool   // If false, ask questions to the user about setting configuration properties, otherwise use default configuration.
	location string // The custom location of the file to create.
}

func runInitCommand(cmd *cobra.Command, args []string) {
	logrus.Info("Executing `arduino config init`")

	if !initFlags._default {
		if !formatter.IsCurrentFormat("text") {
			formatter.PrintErrorMessage("The interactive mode is supported only in text mode.")
			os.Exit(commands.ErrBadCall)
		}
	}

	filepath := initFlags.location
	if filepath == "" {
		filepath = commands.Config.ConfigFile.String()
	}
	err := commands.Config.SaveToYAML(filepath)
	if err != nil {
		formatter.PrintError(err, "Cannot create config file.")
		os.Exit(commands.ErrGeneric)
	}
	formatter.PrintResult("Config file PATH: " + filepath)
	logrus.Info("Done")
}
