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
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/arduino/arduino-cli/commands/updatecheck"
	"github.com/arduino/arduino-cli/internal/cli/board"
	"github.com/arduino/arduino-cli/internal/cli/burnbootloader"
	"github.com/arduino/arduino-cli/internal/cli/cache"
	"github.com/arduino/arduino-cli/internal/cli/compile"
	"github.com/arduino/arduino-cli/internal/cli/completion"
	"github.com/arduino/arduino-cli/internal/cli/config"
	"github.com/arduino/arduino-cli/internal/cli/configuration"
	"github.com/arduino/arduino-cli/internal/cli/core"
	"github.com/arduino/arduino-cli/internal/cli/daemon"
	"github.com/arduino/arduino-cli/internal/cli/debug"
	"github.com/arduino/arduino-cli/internal/cli/feedback"
	"github.com/arduino/arduino-cli/internal/cli/generatedocs"
	"github.com/arduino/arduino-cli/internal/cli/lib"
	"github.com/arduino/arduino-cli/internal/cli/monitor"
	"github.com/arduino/arduino-cli/internal/cli/outdated"
	"github.com/arduino/arduino-cli/internal/cli/sketch"
	"github.com/arduino/arduino-cli/internal/cli/update"
	"github.com/arduino/arduino-cli/internal/cli/updater"
	"github.com/arduino/arduino-cli/internal/cli/upgrade"
	"github.com/arduino/arduino-cli/internal/cli/upload"
	"github.com/arduino/arduino-cli/internal/cli/version"
	"github.com/arduino/arduino-cli/internal/i18n"
	"github.com/arduino/arduino-cli/internal/inventory"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	versioninfo "github.com/arduino/arduino-cli/version"
	"github.com/fatih/color"
	"github.com/mattn/go-colorable"
	"github.com/rifflock/lfshook"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	semver "go.bug.st/relaxed-semver"
)

var (
	verbose      bool
	jsonOutput   bool
	outputFormat string
	configFile   string
)

// NewCommand creates a new ArduinoCli command root
func NewCommand(srv rpc.ArduinoCoreServiceServer) *cobra.Command {
	cobra.AddTemplateFunc("tr", i18n.Tr)

	var updaterMessageChan chan *semver.Version

	// ArduinoCli is the root command
	cmd := &cobra.Command{
		Use:     "arduino-cli",
		Short:   tr("Arduino CLI."),
		Long:    tr("Arduino Command Line Interface (arduino-cli)."),
		Example: fmt.Sprintf("  %s <%s> [%s...]", os.Args[0], tr("command"), tr("flags")),
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			if jsonOutput {
				outputFormat = "json"
			}

			preRun(cmd, args)

			if cmd.Name() != "version" {
				updaterMessageChan = make(chan *semver.Version)
				go func() {
					res, err := updatecheck.CheckForArduinoCLIUpdates(context.Background(), &rpc.CheckForArduinoCLIUpdatesRequest{})
					if err != nil {
						logrus.Warnf("Error checking for updates: %v", err)
						updaterMessageChan <- nil
						return
					}
					if v := res.GetNewestVersion(); v == "" {
						updaterMessageChan <- nil
					} else if latest, err := semver.Parse(v); err != nil {
						logrus.Warnf("Error parsing version: %v", err)
					} else {
						logrus.Infof("New version available: %s", v)
						updaterMessageChan <- latest
					}
				}()
			}
		},
		PersistentPostRun: func(cmd *cobra.Command, args []string) {
			if updaterMessageChan != nil {
				if latestVersion := <-updaterMessageChan; latestVersion != nil {
					// Notify the user a new version is available
					updater.NotifyNewVersionIsAvailable(latestVersion.String())
				}
			}
		},
	}

	cmd.SetUsageTemplate(getUsageTemplate())

	cmd.AddCommand(board.NewCommand(srv))
	cmd.AddCommand(cache.NewCommand())
	cmd.AddCommand(compile.NewCommand(srv))
	cmd.AddCommand(completion.NewCommand())
	cmd.AddCommand(config.NewCommand())
	cmd.AddCommand(core.NewCommand())
	cmd.AddCommand(daemon.NewCommand())
	cmd.AddCommand(generatedocs.NewCommand())
	cmd.AddCommand(lib.NewCommand(srv))
	cmd.AddCommand(monitor.NewCommand(srv))
	cmd.AddCommand(outdated.NewCommand())
	cmd.AddCommand(sketch.NewCommand())
	cmd.AddCommand(update.NewCommand())
	cmd.AddCommand(upgrade.NewCommand())
	cmd.AddCommand(upload.NewCommand(srv))
	cmd.AddCommand(debug.NewCommand(srv))
	cmd.AddCommand(burnbootloader.NewCommand(srv))
	cmd.AddCommand(version.NewCommand())
	cmd.AddCommand(feedback.NewCommand())
	cmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, tr("Print the logs on the standard output."))
	cmd.Flag("verbose").Hidden = true
	cmd.PersistentFlags().BoolVar(&verbose, "log", false, tr("Print the logs on the standard output."))
	validLogLevels := []string{"trace", "debug", "info", "warn", "error", "fatal", "panic"}
	cmd.PersistentFlags().String("log-level", "", tr("Messages with this level and above will be logged. Valid levels are: %s", strings.Join(validLogLevels, ", ")))
	cmd.RegisterFlagCompletionFunc("log-level", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return validLogLevels, cobra.ShellCompDirectiveDefault
	})
	cmd.PersistentFlags().String("log-file", "", tr("Path to the file where logs will be written."))
	validLogFormats := []string{"text", "json"}
	cmd.PersistentFlags().String("log-format", "", tr("The output format for the logs, can be: %s", strings.Join(validLogFormats, ", ")))
	cmd.RegisterFlagCompletionFunc("log-format", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return validLogFormats, cobra.ShellCompDirectiveDefault
	})
	validOutputFormats := []string{"text", "json", "jsonmini"}
	cmd.PersistentFlags().StringVar(&outputFormat, "format", "text", tr("The command output format, can be: %s", strings.Join(validOutputFormats, ", ")))
	cmd.RegisterFlagCompletionFunc("format", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return validOutputFormats, cobra.ShellCompDirectiveDefault
	})
	cmd.Flag("format").Hidden = true
	cmd.PersistentFlags().BoolVar(&jsonOutput, "json", false, tr("Print the output in JSON format."))
	cmd.PersistentFlags().StringVar(&configFile, "config-file", "", tr("The custom config file (if not specified the default will be used)."))
	cmd.PersistentFlags().StringSlice("additional-urls", []string{}, tr("Comma-separated list of additional URLs for the Boards Manager."))
	cmd.PersistentFlags().Bool("no-color", false, "Disable colored output.")
	configuration.BindFlags(cmd, configuration.Settings)

	return cmd
}

// convert the string passed to the `--log-level` option to the corresponding
// logrus formal level.
func toLogLevel(s string) (t logrus.Level, found bool) {
	t, found = map[string]logrus.Level{
		"trace": logrus.TraceLevel,
		"debug": logrus.DebugLevel,
		"info":  logrus.InfoLevel,
		"warn":  logrus.WarnLevel,
		"error": logrus.ErrorLevel,
		"fatal": logrus.FatalLevel,
		"panic": logrus.PanicLevel,
	}[s]

	return
}

func preRun(cmd *cobra.Command, args []string) {
	configFile := configuration.Settings.ConfigFileUsed()

	// initialize inventory
	err := inventory.Init(configuration.DataDir(configuration.Settings).String())
	if err != nil {
		feedback.Fatal(fmt.Sprintf("Error: %v", err), feedback.ErrInitializingInventory)
	}

	// https://no-color.org/
	color.NoColor = configuration.Settings.GetBool("output.no_color") || os.Getenv("NO_COLOR") != ""

	// Set default feedback output to colorable
	feedback.SetOut(colorable.NewColorableStdout())
	feedback.SetErr(colorable.NewColorableStderr())

	//
	// Prepare logging
	//

	// decide whether we should log to stdout
	if verbose {
		// if we print on stdout, do it in full colors
		logrus.SetOutput(colorable.NewColorableStdout())
		logrus.SetFormatter(&logrus.TextFormatter{
			ForceColors:   true,
			DisableColors: color.NoColor,
		})
	} else {
		logrus.SetOutput(io.Discard)
	}

	// set the Logger format
	logFormat := strings.ToLower(configuration.Settings.GetString("logging.format"))
	if logFormat == "json" {
		logrus.SetFormatter(&logrus.JSONFormatter{})
	}

	// should we log to file?
	logFile := configuration.Settings.GetString("logging.file")
	if logFile != "" {
		file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			feedback.Fatal(tr("Unable to open file for logging: %s", logFile), feedback.ErrGeneric)
		}

		// we use a hook so we don't get color codes in the log file
		if logFormat == "json" {
			logrus.AddHook(lfshook.NewHook(file, &logrus.JSONFormatter{}))
		} else {
			logrus.AddHook(lfshook.NewHook(file, &logrus.TextFormatter{}))
		}
	}

	// configure logging filter
	if lvl, found := toLogLevel(configuration.Settings.GetString("logging.level")); !found {
		feedback.Fatal(tr("Invalid option for --log-level: %s", configuration.Settings.GetString("logging.level")), feedback.ErrBadArgument)
	} else {
		logrus.SetLevel(lvl)
	}

	//
	// Prepare the Feedback system
	//

	// check the right output format was passed
	format, found := feedback.ParseOutputFormat(outputFormat)
	if !found {
		feedback.Fatal(tr("Invalid output format: %s", outputFormat), feedback.ErrBadArgument)
	}

	// use the output format to configure the Feedback
	feedback.SetFormat(format)

	//
	// Print some status info and check command is consistent
	//

	if configFile != "" {
		logrus.Infof("Using config file: %s", configFile)
	} else {
		logrus.Info("Config file not found, using default values")
	}

	logrus.Info(versioninfo.VersionInfo.Application + " version " + versioninfo.VersionInfo.VersionString)

	if outputFormat != "text" {
		cmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
			logrus.Warn("Calling help on JSON format")
			feedback.Fatal(tr("Invalid Call : should show Help, but it is available only in TEXT mode."), feedback.ErrBadArgument)
		})
	}
}
