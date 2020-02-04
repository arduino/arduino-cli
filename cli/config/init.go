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
	"path/filepath"

	"github.com/arduino/arduino-cli/cli/errorcodes"
	"github.com/arduino/arduino-cli/cli/feedback"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var destDir string

const defaultFileName = "arduino-cli.yaml"

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
	initCommand.Flags().StringVar(&destDir, "dest-dir", "", "Sets where to save the configuration file.")
	return initCommand
}

func runInitCommand(cmd *cobra.Command, args []string) {
	if destDir == "" {
		destDir = viper.GetString("directories.Data")
	}
	logrus.Infof("Writing config file to: %s", destDir)

	if err := os.MkdirAll(destDir, os.FileMode(0755)); err != nil {
		feedback.Errorf("Cannot create config file directory: %v", err)
		os.Exit(errorcodes.ErrGeneric)
	}

	configFile := filepath.Join(destDir, defaultFileName)
	if err := viper.WriteConfigAs(configFile); err != nil {
		feedback.Errorf("Cannot create config file: %v", err)
		os.Exit(errorcodes.ErrGeneric)
	}

	msg := "Config file written to: " + configFile
	logrus.Info(msg)
	feedback.Print(msg)
}
