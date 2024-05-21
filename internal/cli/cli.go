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

	"github.com/arduino/arduino-cli/internal/cli/board"
	"github.com/arduino/arduino-cli/internal/cli/burnbootloader"
	"github.com/arduino/arduino-cli/internal/cli/cache"
	"github.com/arduino/arduino-cli/internal/cli/compile"
	"github.com/arduino/arduino-cli/internal/cli/completion"
	"github.com/arduino/arduino-cli/internal/cli/config"
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

// NewCommand creates a new ArduinoCli command root
func NewCommand(srv rpc.ArduinoCoreServiceServer) *cobra.Command {
	cobra.AddTemplateFunc("tr", i18n.Tr)

	var updaterMessageChan chan *semver.Version

	var (
		verbose        bool
		noColor        bool
		logLevel       string
		logFile        string
		logFormat      string
		jsonOutput     bool
		outputFormat   string
		additionalUrls []string
	)

	resp, err := srv.ConfigurationGet(context.Background(), &rpc.ConfigurationGetRequest{})
	if err != nil {
		panic("Error creating configuration: " + err.Error())
	}
	settings := resp.GetConfiguration()

	defaultLogFile := settings.GetLogging().GetFile()
	defaultLogFormat := settings.GetLogging().GetFormat()
	defaultLogLevel := settings.GetLogging().GetLevel()
	defaultAdditionalURLs := settings.GetBoardManager().GetAdditionalUrls()
	defaultOutputNoColor := settings.GetOutput().GetNoColor()

	cmd := &cobra.Command{
		Use:     "arduino-cli",
		Short:   i18n.Tr("Arduino CLI."),
		Long:    i18n.Tr("Arduino Command Line Interface (arduino-cli)."),
		Example: fmt.Sprintf("  %s <%s> [%s...]", os.Args[0], i18n.Tr("command"), i18n.Tr("flags")),
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			ctx := cmd.Context()

			config.ApplyGlobalFlagsToConfiguration(ctx, cmd, srv)

			if jsonOutput {
				outputFormat = "json"
			}
			if outputFormat != "text" {
				cmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
					feedback.Fatal(i18n.Tr("Should show help message, but it is available only in TEXT mode."), feedback.ErrBadArgument)
				})
			}

			preRun(verbose, outputFormat, logLevel, logFile, logFormat, noColor, settings)

			// Log the configuration file used
			if configFile := config.GetConfigFile(ctx); configFile != "" {
				logrus.Infof("Using config file: %s", configFile)
			} else {
				logrus.Info("Config file not found, using default values")
			}

			if cmd.Name() != "version" {
				updaterMessageChan = make(chan *semver.Version)
				go func() {
					res, err := srv.CheckForArduinoCLIUpdates(ctx, &rpc.CheckForArduinoCLIUpdatesRequest{})
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
	cmd.AddCommand(cache.NewCommand(srv))
	cmd.AddCommand(compile.NewCommand(srv, settings))
	cmd.AddCommand(completion.NewCommand())
	cmd.AddCommand(config.NewCommand(srv, settings))
	cmd.AddCommand(core.NewCommand(srv))
	cmd.AddCommand(daemon.NewCommand(srv, settings))
	cmd.AddCommand(generatedocs.NewCommand())
	cmd.AddCommand(lib.NewCommand(srv, settings))
	cmd.AddCommand(monitor.NewCommand(srv))
	cmd.AddCommand(outdated.NewCommand(srv))
	cmd.AddCommand(sketch.NewCommand(srv))
	cmd.AddCommand(update.NewCommand(srv))
	cmd.AddCommand(upgrade.NewCommand(srv))
	cmd.AddCommand(upload.NewCommand(srv))
	cmd.AddCommand(debug.NewCommand(srv))
	cmd.AddCommand(burnbootloader.NewCommand(srv))
	cmd.AddCommand(version.NewCommand(srv))
	cmd.AddCommand(feedback.NewCommand())

	cmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, i18n.Tr("Print the logs on the standard output."))
	cmd.Flag("verbose").Hidden = true
	cmd.PersistentFlags().BoolVar(&verbose, "log", false, i18n.Tr("Print the logs on the standard output."))
	validLogLevels := []string{"trace", "debug", "info", "warn", "error", "fatal", "panic"}
	cmd.PersistentFlags().StringVar(&logLevel, "log-level", defaultLogLevel, i18n.Tr("Messages with this level and above will be logged. Valid levels are: %s", strings.Join(validLogLevels, ", ")))
	cmd.RegisterFlagCompletionFunc("log-level", cobra.FixedCompletions(validLogLevels, cobra.ShellCompDirectiveDefault))
	cmd.PersistentFlags().StringVar(&logFile, "log-file", defaultLogFile, i18n.Tr("Path to the file where logs will be written."))
	validLogFormats := []string{"text", "json"}
	cmd.PersistentFlags().StringVar(&logFormat, "log-format", defaultLogFormat, i18n.Tr("The output format for the logs, can be: %s", strings.Join(validLogFormats, ", ")))
	cmd.RegisterFlagCompletionFunc("log-format", cobra.FixedCompletions(validLogFormats, cobra.ShellCompDirectiveDefault))
	validOutputFormats := []string{"text", "json", "jsonmini"}
	cmd.PersistentFlags().StringVar(&outputFormat, "format", "text", i18n.Tr("The command output format, can be: %s", strings.Join(validOutputFormats, ", ")))
	cmd.RegisterFlagCompletionFunc("format", cobra.FixedCompletions(validOutputFormats, cobra.ShellCompDirectiveDefault))
	cmd.Flag("format").Hidden = true
	cmd.PersistentFlags().BoolVar(&jsonOutput, "json", false, i18n.Tr("Print the output in JSON format."))
	cmd.PersistentFlags().StringSliceVar(&additionalUrls, "additional-urls", defaultAdditionalURLs, i18n.Tr("Comma-separated list of additional URLs for the Boards Manager."))
	cmd.PersistentFlags().BoolVar(&noColor, "no-color", defaultOutputNoColor, "Disable colored output.")

	// We are not using cobra to parse this flag, because we manually parse it in main.go.
	// Just leaving it here so cobra will not complain about it.
	cmd.PersistentFlags().String("config-file", "", i18n.Tr("The custom config file (if not specified the default will be used)."))
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

func preRun(verbose bool, outputFormat string, logLevel, logFile, logFormat string, noColor bool, settings *rpc.Configuration) {
	//
	// Prepare the Feedback system
	//

	// Set default feedback output to colorable
	color.NoColor = noColor || os.Getenv("NO_COLOR") != "" // https://no-color.org/
	feedback.SetOut(colorable.NewColorableStdout())
	feedback.SetErr(colorable.NewColorableStderr())

	// use the output format to configure the Feedback
	format, ok := feedback.ParseOutputFormat(outputFormat)
	if !ok {
		feedback.Fatal(i18n.Tr("Invalid output format: %s", outputFormat), feedback.ErrBadArgument)
	}
	feedback.SetFormat(format)

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
	logFormat = strings.ToLower(logFormat)
	if logFormat == "json" {
		logrus.SetFormatter(&logrus.JSONFormatter{})
	}

	// should we log to file?
	if logFile != "" {
		file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			feedback.Fatal(i18n.Tr("Unable to open file for logging: %s", logFile), feedback.ErrGeneric)
		}

		// we use a hook so we don't get color codes in the log file
		if logFormat == "json" {
			logrus.AddHook(lfshook.NewHook(file, &logrus.JSONFormatter{}))
		} else {
			logrus.AddHook(lfshook.NewHook(file, &logrus.TextFormatter{}))
		}
	}

	// configure logging filter
	if logrusLevel, found := toLogLevel(logLevel); !found {
		feedback.Fatal(i18n.Tr("Invalid logging level: %s", logLevel), feedback.ErrBadArgument)
	} else {
		logrus.SetLevel(logrusLevel)
	}

	// Print some status info and check command is consistent
	logrus.Info(versioninfo.VersionInfo.Application + " version " + versioninfo.VersionInfo.VersionString)

	//
	// Initialize inventory
	//
	err := inventory.Init(settings.GetDirectories().GetData())
	if err != nil {
		feedback.Fatal(fmt.Sprintf("Error: %v", err), feedback.ErrInitializingInventory)
	}
}
