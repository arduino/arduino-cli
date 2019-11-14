/*
 * This file is part of arduino-
 *
 * Copyright 2018 ARDUINO SA (http://www.arduino.cc/)
 *
 * This software is released under the GNU General Public License version 3,
 * which covers the main part of arduino-
 * The terms of this license can be found at:
 * https://www.gnu.org/licenses/gpl-3.0.en.html
 *
 * You can be released from the requirements of the above licenses by purchasing
 * a commercial license. Buying such a license is mandatory if you want to modify or
 * otherwise use the software for commercial activities involving the Arduino
 * software without disclosing the source code of your own applications. To purchase
 * a commercial license, send an email to license@arduino.cc.
 */

package cli

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/arduino/arduino-cli/cli/board"
	"github.com/arduino/arduino-cli/cli/compile"
	"github.com/arduino/arduino-cli/cli/config"
	"github.com/arduino/arduino-cli/cli/core"
	"github.com/arduino/arduino-cli/cli/daemon"
	"github.com/arduino/arduino-cli/cli/errorcodes"
	"github.com/arduino/arduino-cli/cli/feedback"
	"github.com/arduino/arduino-cli/cli/generatedocs"
	"github.com/arduino/arduino-cli/cli/globals"
	"github.com/arduino/arduino-cli/cli/lib"
	"github.com/arduino/arduino-cli/cli/output"
	"github.com/arduino/arduino-cli/cli/sketch"
	"github.com/arduino/arduino-cli/cli/upload"
	"github.com/arduino/arduino-cli/cli/version"
	"github.com/arduino/arduino-cli/configuration"
	"github.com/mattn/go-colorable"
	"github.com/rifflock/lfshook"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	// ArduinoCli is the root command
	ArduinoCli = &cobra.Command{
		Use:              "arduino-cli",
		Short:            "Arduino CLI.",
		Long:             "Arduino Command Line Interface (arduino-cli).",
		Example:          "  " + os.Args[0] + " <command> [flags...]",
		PersistentPreRun: preRun,
	}

	verbose      bool
	logFile      string
	logFormat    string
	outputFormat string
)

const (
	defaultLogLevel = "info"
)

// Init the cobra root command
func init() {
	createCliCommandTree(ArduinoCli)
}

// this is here only for testing
func createCliCommandTree(cmd *cobra.Command) {
	cmd.AddCommand(board.NewCommand())
	cmd.AddCommand(compile.NewCommand())
	cmd.AddCommand(config.NewCommand())
	cmd.AddCommand(core.NewCommand())
	cmd.AddCommand(daemon.NewCommand())
	cmd.AddCommand(generatedocs.NewCommand())
	cmd.AddCommand(lib.NewCommand())
	cmd.AddCommand(sketch.NewCommand())
	cmd.AddCommand(upload.NewCommand())
	cmd.AddCommand(version.NewCommand())

	cmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Print the logs on the standard output.")
	cmd.PersistentFlags().StringVar(&globals.LogLevel, "log-level", defaultLogLevel, "Messages with this level and above will be logged.")
	cmd.PersistentFlags().StringVar(&logFile, "log-file", "", "Path to the file where logs will be written.")
	cmd.PersistentFlags().StringVar(&logFormat, "log-format", "text", "The output format for the logs, can be [text|json].")
	cmd.PersistentFlags().StringVar(&outputFormat, "format", "text", "The output format, can be [text|json].")
	cmd.PersistentFlags().StringVar(&globals.YAMLConfigFile, "config-file", "", "The custom config file (if not specified the default will be used).")
	cmd.PersistentFlags().StringSliceVar(&globals.AdditionalUrls, "additional-urls", []string{}, "Additional URLs for the board manager.")
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

func parseFormatString(arg string) (feedback.OutputFormat, bool) {
	f, found := map[string]feedback.OutputFormat{
		"json": feedback.JSON,
		"text": feedback.Text,
	}[arg]

	return f, found
}

func preRun(cmd *cobra.Command, args []string) {
	// normalize the format strings
	outputFormat = strings.ToLower(outputFormat)
	// configure the output package
	output.OutputFormat = outputFormat
	logFormat = strings.ToLower(logFormat)

	// should we log to file?
	if logFile != "" {
		file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			fmt.Printf("Unable to open file for logging: %s", logFile)
			os.Exit(errorcodes.ErrBadCall)
		}

		// we use a hook so we don't get color codes in the log file
		if logFormat == "json" {
			logrus.AddHook(lfshook.NewHook(file, &logrus.JSONFormatter{}))
		} else {
			logrus.AddHook(lfshook.NewHook(file, &logrus.TextFormatter{}))
		}
	}

	// should we log to stdout?
	if verbose {
		logrus.SetOutput(colorable.NewColorableStdout())
		logrus.SetFormatter(&logrus.TextFormatter{
			ForceColors: true,
		})
	} else {
		// Discard logrus output if no writer was set
		logrus.SetOutput(ioutil.Discard)
	}

	// configure logging filter
	if lvl, found := toLogLevel(globals.LogLevel); !found {
		fmt.Printf("Invalid option for --log-level: %s", globals.LogLevel)
		os.Exit(errorcodes.ErrBadArgument)
	} else {
		logrus.SetLevel(lvl)
	}

	// set the Logger format
	if logFormat == "json" {
		logrus.SetFormatter(&logrus.JSONFormatter{})
	}

	// check the right output format was passed
	format, found := parseFormatString(outputFormat)
	if !found {
		feedback.Error("Invalid output format: " + outputFormat)
		os.Exit(errorcodes.ErrBadCall)
	}

	// use the output format to configure the Feedback
	feedback.SetFormat(format)

	configuration.Init()

	logrus.Info(globals.VersionInfo.Application + "-" + globals.VersionInfo.VersionString)
	logrus.Info("Starting root command preparation (`arduino`)")

	logrus.Info("Formatter set")
	if outputFormat != "text" {
		cmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
			logrus.Warn("Calling help on JSON format")
			feedback.Error("Invalid Call : should show Help, but it is available only in TEXT mode.")
			os.Exit(errorcodes.ErrBadCall)
		})
	}
}
