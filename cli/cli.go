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
	"io/ioutil"
	"os"

	"github.com/arduino/arduino-cli/cli/board"
	"github.com/arduino/arduino-cli/cli/compile"
	"github.com/arduino/arduino-cli/cli/config"
	"github.com/arduino/arduino-cli/cli/core"
	"github.com/arduino/arduino-cli/cli/daemon"
	"github.com/arduino/arduino-cli/cli/errorcodes"
	"github.com/arduino/arduino-cli/cli/generatedocs"
	"github.com/arduino/arduino-cli/cli/globals"
	"github.com/arduino/arduino-cli/cli/lib"
	"github.com/arduino/arduino-cli/cli/sketch"
	"github.com/arduino/arduino-cli/cli/upload"
	"github.com/arduino/arduino-cli/cli/version"
	"github.com/arduino/arduino-cli/common/formatter"
	"github.com/mattn/go-colorable"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh/terminal"
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

	// ErrLogrus represents the logrus instance, which has the role to
	// log all non info messages.
	ErrLogrus = logrus.New()

	outputFormat string
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

	cmd.PersistentFlags().BoolVar(&globals.Debug, "debug", false, "Enables debug output (super verbose, used to debug the CLI).")
	cmd.PersistentFlags().StringVar(&outputFormat, "format", "text", "The output format, can be [text|json].")
	cmd.PersistentFlags().StringVar(&globals.YAMLConfigFile, "config-file", "", "The custom config file (if not specified the default will be used).")
}

func preRun(cmd *cobra.Command, args []string) {
	// Reset logrus if debug flag changed.
	if !globals.Debug {
		// Discard logrus output if no debug.
		logrus.SetOutput(ioutil.Discard)
	} else {
		// Else print on stderr.

		// Workaround to get colored output on windows
		if terminal.IsTerminal(int(os.Stdout.Fd())) {
			logrus.SetFormatter(&logrus.TextFormatter{ForceColors: true})
		}
		logrus.SetOutput(colorable.NewColorableStdout())
		ErrLogrus.Out = colorable.NewColorableStderr()
		formatter.SetLogger(ErrLogrus)
	}
	globals.InitConfigs()

	logrus.Info(globals.VersionInfo.Application + "-" + globals.VersionInfo.VersionString)
	logrus.Info("Starting root command preparation (`arduino`)")
	switch outputFormat {
	case "text":
		formatter.SetFormatter("text")
		globals.OutputJSON = false
	case "json":
		formatter.SetFormatter("json")
		globals.OutputJSON = true
	default:
		formatter.PrintErrorMessage("Invalid output format: " + outputFormat)
		os.Exit(errorcodes.ErrBadCall)
	}

	logrus.Info("Formatter set")
	if !formatter.IsCurrentFormat("text") {
		cmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
			logrus.Warn("Calling help on JSON format")
			formatter.PrintErrorMessage("Invalid Call : should show Help, but it is available only in TEXT mode.")
			os.Exit(errorcodes.ErrBadCall)
		})
	}
}
