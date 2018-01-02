/*
 * This file is part of arduino-cli.
 *
 * arduino-cli is free software; you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation; either version 2 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program; if not, write to the Free Software
 * Foundation, Inc., 51 Franklin St, Fifth Floor, Boston, MA  02110-1301  USA
 *
 * As a special exception, you may use this file as part of a free software
 * library without restriction.  Specifically, if other files instantiate
 * templates or use macros or inline functions from this file, or you compile
 * this file and link it with other files to produce an executable, this
 * file does not by itself cause the resulting executable to be covered by
 * the GNU General Public License.  This exception does not however
 * invalidate any other reasons why the executable file might be covered by
 * the GNU General Public License.
 *
 * Copyright 2017 ARDUINO AG (http://www.arduino.cc/)
 */

package config

import (
	"os"

	"github.com/bcmi-labs/arduino-cli/commands"
	"github.com/bcmi-labs/arduino-cli/common/formatter"
	"github.com/bcmi-labs/arduino-cli/configs"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	command.AddCommand(initCommand)
	initCommand.Flags().BoolVar(&initFlags._default, "default", false, "If omitted, ask questions to the user about setting configuration properties, otherwise use default configuration.")
	initCommand.Flags().StringVar(&initFlags.location, "save-as", configs.FileLocation, "Sets where to save the configuration file [default is ./.cli-config.yml].")
}

var initFlags struct {
	_default bool   // If false, ask questions to the user about setting configuration properties, otherwise use default configuration.
	location string // The custom location of the file to create.
}

var initCommand = &cobra.Command{
	Use:   "init",
	Short: "Initializes a new config file into the default location.",
	Long:  "Initializes a new config file into the default location ($EXE_DIR/cli-config.yml).",
	Example: "" +
		"arduino config init           # Creates a config file by asking questions to the user into the default location.\n" +
		"arduino config init --default # Creates a config file with default configuration into default location.",
	Args: cobra.NoArgs,
	Run:  runInitCommand,
}

func runInitCommand(cmd *cobra.Command, args []string) {
	logrus.Info("Executing `arduino config init`")

	var conf configs.Configs

	if !initFlags._default {
		if !formatter.IsCurrentFormat("text") {
			formatter.PrintErrorMessage("The interactive mode is supported only in text mode.")
			os.Exit(commands.ErrBadCall)
		}
		logrus.Info("Asking questions to the user to populate configuration")
		conf = configsFromQuestions()
	} else {
		logrus.Info("Setting default configuration")
		conf = configs.Default()
	}
	err := conf.Serialize(initFlags.location)
	if err != nil {
		formatter.PrintError(err, "Cannot create config file.")
		os.Exit(commands.ErrGeneric)
	} else {
		formatter.PrintResult("Config file PATH: " + initFlags.location)
	}
	logrus.Info("Done")
}

// ConfigsFromQuestions asks some questions to the user to properly initialize configs.
// It does not have much sense to use it in JSON formatting, though.
func configsFromQuestions() configs.Configs {
	ret := configs.Default()
	// Set of questions here.
	return ret
}
