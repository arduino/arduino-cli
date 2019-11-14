// This file is part of arduino-cli.
//
// Copyright 2019 ARDUINO SA (http://www.arduino.cc/)
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
	"path/filepath"

	"github.com/arduino/arduino-cli/configuration"

	"github.com/arduino/arduino-cli/cli/errorcodes"
	"github.com/arduino/arduino-cli/cli/feedback"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func initInitCommand() *cobra.Command {
	initCommand := &cobra.Command{
		Use:   "init",
		Short: "Initializes a new configuration file into the default location.",
		Long:  "Initializes a new configuration file into the default location ($EXE_DIR/cli-config.yml).",
		Example: "" +
			"  # Creates a default configuration file into the default location.\n" +
			"  " + os.Args[0] + " config init",
		Args: cobra.NoArgs,
		Run:  runInitCommand,
	}
	initCommand.Flags().StringVar(&initFlags.location, "save-as", "",
		"Sets where to save the configuration file [default is ./arduino-cli.yaml].")
	return initCommand
}

var initFlags struct {
	location string // The custom location of the file to create.
}

func runInitCommand(cmd *cobra.Command, args []string) {
	logrus.Info("Executing `arduino config init`")

	configFile := filepath.Join(configuration.GetDefaultArduinoDataDir(), "arduino-cli.yaml")
	err := viper.WriteConfigAs(configFile)
	if err != nil {
		feedback.Errorf("Cannot create config file: %v", err)
		os.Exit(errorcodes.ErrGeneric)
	}

	feedback.Print("Config file written: " + configFile)
	logrus.Info("Done")
}
