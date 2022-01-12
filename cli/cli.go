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

package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/arduino/arduino-cli/cli/board"
	"github.com/arduino/arduino-cli/cli/burnbootloader"
	"github.com/arduino/arduino-cli/cli/cache"
	"github.com/arduino/arduino-cli/cli/compile"
	"github.com/arduino/arduino-cli/cli/completion"
	"github.com/arduino/arduino-cli/cli/config"
	"github.com/arduino/arduino-cli/cli/core"
	"github.com/arduino/arduino-cli/cli/daemon"
	"github.com/arduino/arduino-cli/cli/debug"
	"github.com/arduino/arduino-cli/cli/errorcodes"
	"github.com/arduino/arduino-cli/cli/feedback"
	"github.com/arduino/arduino-cli/cli/generatedocs"
	"github.com/arduino/arduino-cli/cli/globals"
	"github.com/arduino/arduino-cli/cli/lib"
	"github.com/arduino/arduino-cli/cli/monitor"
	"github.com/arduino/arduino-cli/cli/outdated"
	"github.com/arduino/arduino-cli/cli/sketch"
	"github.com/arduino/arduino-cli/cli/update"
	"github.com/arduino/arduino-cli/cli/updater"
	"github.com/arduino/arduino-cli/cli/upgrade"
	"github.com/arduino/arduino-cli/cli/upload"
	"github.com/arduino/arduino-cli/cli/version"
	"github.com/arduino/arduino-cli/configuration"
	"github.com/arduino/arduino-cli/i18n"
	"github.com/arduino/arduino-cli/inventory"
	"github.com/arduino/arduino-cli/logging"
	"github.com/arduino/arduino-cli/output"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	semver "go.bug.st/relaxed-semver"
)

var (
	configFile         string
	updaterMessageChan chan *semver.Version = make(chan *semver.Version)
)

// NewCommand creates a new ArduinoCli command root
func NewCommand() *cobra.Command {
	cobra.AddTemplateFunc("tr", i18n.Tr)

	// ArduinoCli is the root command
	arduinoCli := &cobra.Command{
		Use:               "arduino-cli",
		Short:             tr("Arduino CLI."),
		Long:              tr("Arduino Command Line Interface (arduino-cli)."),
		Example:           fmt.Sprintf("  %s <%s> [%s...]", os.Args[0], tr("command"), tr("flags")),
		PersistentPreRun:  preRun,
		PersistentPostRun: postRun,
	}

	arduinoCli.SetUsageTemplate(getUsageTemplate())

	createCliCommandTree(arduinoCli)

	return arduinoCli
}

// this is here only for testing
func createCliCommandTree(cmd *cobra.Command) {
	cmd.AddCommand(board.NewCommand())
	cmd.AddCommand(cache.NewCommand())
	cmd.AddCommand(compile.NewCommand())
	cmd.AddCommand(completion.NewCommand())
	cmd.AddCommand(config.NewCommand())
	cmd.AddCommand(core.NewCommand())
	cmd.AddCommand(daemon.NewCommand())
	cmd.AddCommand(generatedocs.NewCommand())
	cmd.AddCommand(lib.NewCommand())
	cmd.AddCommand(monitor.NewCommand())
	cmd.AddCommand(outdated.NewCommand())
	cmd.AddCommand(sketch.NewCommand())
	cmd.AddCommand(update.NewCommand())
	cmd.AddCommand(upgrade.NewCommand())
	cmd.AddCommand(upload.NewCommand())
	cmd.AddCommand(debug.NewCommand())
	cmd.AddCommand(burnbootloader.NewCommand())
	cmd.AddCommand(version.NewCommand())

	validLogLevels := []string{"trace", "debug", "info", "warn", "error", "fatal", "panic"}
	validLogFormats := []string{"text", "json"}
	cmd.PersistentFlags().String("log-level", "info", tr("Messages with this level and above will be logged. Valid levels are: %s", strings.Join(validLogLevels, ", ")))
	cmd.PersistentFlags().String("log-file", "", tr("Path to the file where logs will be written."))
	cmd.PersistentFlags().String("log-format", "text", tr("The output format for the logs, can be: %s", strings.Join(validLogFormats, ", ")))
	cmd.RegisterFlagCompletionFunc("log-level", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return validLogLevels, cobra.ShellCompDirectiveDefault
	})
	cmd.RegisterFlagCompletionFunc("log-format", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return validLogFormats, cobra.ShellCompDirectiveDefault
	})
	validOutputFormats := []string{"text", "json", "jsonmini", "yaml"}
	cmd.PersistentFlags().BoolP("verbose", "v", false, tr("Print the logs on the standard output."))
	cmd.PersistentFlags().String("format", "text", tr("The output format for the logs, can be: %s", strings.Join(validOutputFormats, ", ")))
	cmd.PersistentFlags().Bool("no-color", false, "Disable colored output.")
	cmd.RegisterFlagCompletionFunc("format", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return validOutputFormats, cobra.ShellCompDirectiveDefault
	})
	cmd.PersistentFlags().StringVar(&configFile, "config-file", "", tr("The custom config file (if not specified the default will be used)."))
	cmd.PersistentFlags().StringSlice("additional-urls", []string{}, tr("Comma-separated list of additional URLs for the Boards Manager."))
	configuration.BindFlags(cmd, configuration.Settings)
}

func preRun(cmd *cobra.Command, args []string) {
	if cmd.Name() == "daemon" {
		return
	}

	configFile := configuration.Settings.ConfigFileUsed()

	// initialize inventory
	err := inventory.Init(configuration.Settings.GetString("directories.Data"))
	if err != nil {
		feedback.Errorf("Error: %v", err)
		os.Exit(errorcodes.ErrBadArgument)
	}

	outputFormat, err := cmd.Flags().GetString("format")
	if err != nil {
		feedback.Errorf(tr("Error getting flag value: %s", err))
		os.Exit(errorcodes.ErrBadCall)
	}
	noColor := configuration.Settings.GetBool("output.no_color") || os.Getenv("NO_COLOR") != ""
	output.Setup(outputFormat, noColor)

	updaterMessageChan = make(chan *semver.Version)
	go func() {
		if cmd.Name() == "version" {
			// The version command checks by itself if there's a new available version
			updaterMessageChan <- nil
		}
		// Starts checking for updates
		currentVersion, err := semver.Parse(globals.VersionInfo.VersionString)
		if err != nil {
			updaterMessageChan <- nil
		}
		updaterMessageChan <- updater.CheckForUpdate(currentVersion)
	}()

	// Setups logging if necessary
	verbose, err := cmd.Flags().GetBool("verbose")
	if err != nil {
		feedback.Errorf(tr("Error getting flag value: %s", err))
		os.Exit(errorcodes.ErrBadCall)
	}
	logging.Setup(
		verbose,
		noColor,
		configuration.Settings.GetString("logging.level"),
		configuration.Settings.GetString("logging.file"),
		configuration.Settings.GetString("logging.format"),
	)

	//
	// Print some status info and check command is consistent
	//

	if configFile != "" {
		logrus.Infof("Using config file: %s", configFile)
	} else {
		logrus.Info("Config file not found, using default values")
	}

	logrus.Info(globals.VersionInfo.Application + " version " + globals.VersionInfo.VersionString)

	if outputFormat != "text" {
		cmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
			logrus.Warn("Calling help on JSON format")
			feedback.Error(tr("Invalid Call : should show Help, but it is available only in TEXT mode."))
			os.Exit(errorcodes.ErrBadCall)
		})
	}
}

func postRun(cmd *cobra.Command, args []string) {
	latestVersion := <-updaterMessageChan
	if latestVersion != nil {
		// Notify the user a new version is available
		updater.NotifyNewVersionIsAvailable(latestVersion.String())
	}
}
