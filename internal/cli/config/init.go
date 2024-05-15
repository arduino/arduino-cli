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
	"encoding/json"
	"os"
	"strings"

	"github.com/arduino/arduino-cli/commands"
	"github.com/arduino/arduino-cli/internal/cli/arguments"
	"github.com/arduino/arduino-cli/internal/cli/feedback"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	"github.com/arduino/go-paths-helper"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
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
		Short: tr("Writes current configuration to a configuration file."),
		Long:  tr("Creates or updates the configuration file in the data directory or custom directory with the current configuration settings."),
		Example: "" +
			"  # " + tr("Writes current configuration to the configuration file in the data directory.") + "\n" +
			"  " + os.Args[0] + " config init\n" +
			"  " + os.Args[0] + " config init --dest-dir /home/user/MyDirectory\n" +
			"  " + os.Args[0] + " config init --dest-file /home/user/MyDirectory/my_settings.yaml",
		Args: cobra.NoArgs,
		PreRun: func(cmd *cobra.Command, args []string) {
			arguments.CheckFlagsConflicts(cmd, "dest-file", "dest-dir")
		},
		Run: func(cmd *cobra.Command, args []string) {
			runInitCommand(cmd.Context(), cmd)
		},
	}
	initCommand.Flags().StringVar(&destDir, "dest-dir", "", tr("Sets where to save the configuration file."))
	initCommand.Flags().StringVar(&destFile, "dest-file", "", tr("Sets where to save the configuration file."))
	initCommand.Flags().BoolVar(&overwrite, "overwrite", false, tr("Overwrite existing config file."))
	return initCommand
}

func runInitCommand(ctx context.Context, cmd *cobra.Command) {
	logrus.Info("Executing `arduino-cli config init`")

	var configFileAbsPath *paths.Path
	var configFileDir *paths.Path
	var err error

	switch {
	case destFile != "":
		configFileAbsPath, err = paths.New(destFile).Abs()
		if err != nil {
			feedback.Fatal(tr("Cannot find absolute path: %v", err), feedback.ErrGeneric)
		}
		configFileDir = configFileAbsPath.Parent()

	case destDir != "":
		configFileDir, err = paths.New(destDir).Abs()
		if err != nil {
			feedback.Fatal(tr("Cannot find absolute path: %v", err), feedback.ErrGeneric)
		}
		configFileAbsPath = configFileDir.Join(defaultFileName)

	default:
		configFileAbsPath = paths.New(GetConfigFile(ctx))
		configFileDir = configFileAbsPath.Parent()
	}

	if !overwrite && configFileAbsPath.Exist() {
		feedback.Fatal(tr("Config file already exists, use --overwrite to discard the existing one."), feedback.ErrGeneric)
	}

	logrus.Infof("Writing config file to: %s", configFileDir)

	if err := configFileDir.MkdirAll(); err != nil {
		feedback.Fatal(tr("Cannot create config file directory: %v", err), feedback.ErrGeneric)
	}

	tmpSrv := commands.NewArduinoCoreServer()

	if _, err := tmpSrv.ConfigurationOpen(ctx, &rpc.ConfigurationOpenRequest{SettingsFormat: "yaml", EncodedSettings: ""}); err != nil {
		feedback.Fatal(tr("Error creating configuration: %v", err), feedback.ErrGeneric)
	}

	// Ensure to always output an empty array for additional urls
	if _, err := tmpSrv.SettingsSetValue(ctx, &rpc.SettingsSetValueRequest{
		Key: "board_manager.additional_urls", EncodedValue: "[]",
	}); err != nil {
		feedback.Fatal(tr("Error creating configuration: %v", err), feedback.ErrGeneric)
	}

	ApplyGlobalFlagsToConfiguration(ctx, cmd, tmpSrv)

	resp, err := tmpSrv.ConfigurationSave(ctx, &rpc.ConfigurationSaveRequest{SettingsFormat: "yaml"})
	if err != nil {
		feedback.Fatal(tr("Error creating configuration: %v", err), feedback.ErrGeneric)
	}

	if err := configFileAbsPath.WriteFile([]byte(resp.GetEncodedSettings())); err != nil {
		feedback.Fatal(tr("Cannot create config file: %v", err), feedback.ErrGeneric)
	}

	feedback.PrintResult(initResult{ConfigFileAbsPath: configFileAbsPath})
}

// ApplyGlobalFlagsToConfiguration overrides server settings with the flags from the command line
func ApplyGlobalFlagsToConfiguration(ctx context.Context, cmd *cobra.Command, srv rpc.ArduinoCoreServiceServer) {
	set := func(k string, v any) {
		if jsonValue, err := json.Marshal(v); err != nil {
			feedback.Fatal(tr("Error creating configuration: %v", err), feedback.ErrGeneric)
		} else if _, err := srv.SettingsSetValue(ctx, &rpc.SettingsSetValueRequest{
			Key: k, EncodedValue: string(jsonValue),
		}); err != nil {
			feedback.Fatal(tr("Error creating configuration: %v", err), feedback.ErrGeneric)
		}

	}

	if f := cmd.Flags().Lookup("log-level"); f.Changed {
		logLevel, _ := cmd.Flags().GetString("log-level")
		set("logging.level", logLevel)
	}
	if f := cmd.Flags().Lookup("log-file"); f.Changed {
		logFile, _ := cmd.Flags().GetString("log-file")
		set("logging.file", logFile)
	}
	if f := cmd.Flags().Lookup("no-color"); f.Changed {
		noColor, _ := cmd.Flags().GetBool("no-color")
		set("output.no_color", noColor)
	}
	if f := cmd.Flags().Lookup("additional-urls"); f.Changed {
		urls, _ := cmd.Flags().GetStringSlice("additional-urls")
		for _, url := range urls {
			if strings.Contains(url, ",") {
				feedback.Fatal(
					tr("Urls cannot contain commas. Separate multiple urls exported as env var with a space:\n%s", url),
					feedback.ErrBadArgument)
			}
		}
		set("board_manager.additional_urls", urls)
	}
}

// output from this command requires special formatting, let's create a dedicated
// feedback.Result implementation
type initResult struct {
	ConfigFileAbsPath *paths.Path `json:"config_path"`
}

func (dr initResult) Data() interface{} {
	return dr
}

func (dr initResult) String() string {
	msg := tr("Config file written to: %s", dr.ConfigFileAbsPath.String())
	logrus.Info(msg)
	return msg
}
