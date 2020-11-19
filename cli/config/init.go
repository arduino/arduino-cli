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

	"github.com/arduino/arduino-cli/cli/errorcodes"
	"github.com/arduino/arduino-cli/cli/feedback"
	"github.com/arduino/arduino-cli/configuration"
	"github.com/arduino/go-paths-helper"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	destDir   string
	destFile  string
	overwrite bool
)

const defaultFileName = "arduino-cli.yaml"

func initInitCommand() *cobra.Command {
	initCommand := &cobra.Command{
		Use:   "init",
		Short: "Writes current configuration to a configuration file.",
		Long:  "Creates or updates the configuration file in the data directory or custom directory with the current configuration settings.",
		Example: "" +
			"  # Writes current configuration to the configuration file in the data directory.\n" +
			"  " + os.Args[0] + " config init\n" +
			"  " + os.Args[0] + " config init --dest-dir /home/user/MyDirectory\n" +
			"  " + os.Args[0] + " config init --dest-file /home/user/MyDirectory/my_settings.yaml",
		Args: cobra.NoArgs,
		Run:  runInitCommand,
	}
	initCommand.Flags().StringVar(&destDir, "dest-dir", "", "Sets where to save the configuration file.")
	initCommand.Flags().StringVar(&destFile, "dest-file", "", "Sets where to save the configuration file.")
	initCommand.Flags().BoolVar(&overwrite, "overwrite", false, "Overwrite existing config file.")
	return initCommand
}

func runInitCommand(cmd *cobra.Command, args []string) {
	if destFile != "" && destDir != "" {
		feedback.Errorf("Can't use both --dest-file and --dest-dir flags at the same time.")
		os.Exit(errorcodes.ErrGeneric)
	}

	var configFileAbsPath *paths.Path
	var absPath *paths.Path
	var err error

	switch {
	case destFile != "":
		configFileAbsPath, err = paths.New(destFile).Abs()
		if err != nil {
			feedback.Errorf("Cannot find absolute path: %v", err)
			os.Exit(errorcodes.ErrGeneric)
		}

		absPath = configFileAbsPath.Parent()
	case destDir == "":
		destDir = configuration.Settings.GetString("directories.Data")
		fallthrough
	default:
		absPath, err = paths.New(destDir).Abs()
		if err != nil {
			feedback.Errorf("Cannot find absolute path: %v", err)
			os.Exit(errorcodes.ErrGeneric)
		}
		configFileAbsPath = absPath.Join(defaultFileName)
	}

	if !overwrite && configFileAbsPath.Exist() {
		feedback.Error("Config file already exists, use --overwrite to discard the existing one.")
		os.Exit(errorcodes.ErrGeneric)
	}

	logrus.Infof("Writing config file to: %s", absPath)

	if err := absPath.MkdirAll(); err != nil {
		feedback.Errorf("Cannot create config file directory: %v", err)
		os.Exit(errorcodes.ErrGeneric)
	}

	newSettings := viper.New()
	configuration.SetDefaults(newSettings)
	configuration.BindFlags(cmd, newSettings)

	if err := newSettings.WriteConfigAs(configFileAbsPath.String()); err != nil {
		feedback.Errorf("Cannot create config file: %v", err)
		os.Exit(errorcodes.ErrGeneric)
	}

	msg := "Config file written to: " + configFileAbsPath.String()
	logrus.Info(msg)
	feedback.Print(msg)
}
