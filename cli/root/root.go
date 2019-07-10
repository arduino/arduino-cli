/*
 * This file is part of arduino-cli.
 *
 * Copyright 2018 ARDUINO SA (http://www.arduino.cc/)
 *
 * This software is released under the GNU General Public License version 3,
 * which covers the main part of arduino-cli.
 * The terms of this license can be found at:
 * https://www.gnu.org/licenses/gpl-3.0.en.html
 *
 * You can be released from the requirements of the above licenses by purchasing
 * a commercial license. Buying such a license is mandatory if you want to modify or
 * otherwise use the software for commercial activities involving the Arduino
 * software without disclosing the source code of your own applications. To purchase
 * a commercial license, send an email to license@arduino.cc.
 */

package root

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/arduino/arduino-cli/cli"
	"github.com/arduino/arduino-cli/cli/board"
	"github.com/arduino/arduino-cli/cli/compile"
	"github.com/arduino/arduino-cli/cli/config"
	"github.com/arduino/arduino-cli/cli/core"
	"github.com/arduino/arduino-cli/cli/daemon"
	"github.com/arduino/arduino-cli/cli/generatedocs"
	"github.com/arduino/arduino-cli/cli/lib"
	"github.com/arduino/arduino-cli/cli/sketch"
	"github.com/arduino/arduino-cli/cli/upload"
	"github.com/arduino/arduino-cli/cli/version"
	"github.com/arduino/arduino-cli/common/formatter"
	"github.com/arduino/arduino-cli/configs"
	"github.com/arduino/go-paths-helper"
	"github.com/mattn/go-colorable"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh/terminal"
)

// Init prepares the cobra root command.
func Init() *cobra.Command {
	command := &cobra.Command{
		Use:              "arduino-cli",
		Short:            "Arduino CLI.",
		Long:             "Arduino Command Line Interface (arduino-cli).",
		Example:          "  " + cli.VersionInfo.Application + " <command> [flags...]",
		PersistentPreRun: preRun,
	}
	command.PersistentFlags().BoolVar(&cli.GlobalFlags.Debug, "debug", false, "Enables debug output (super verbose, used to debug the CLI).")
	command.PersistentFlags().StringVar(&outputFormat, "format", "text", "The output format, can be [text|json].")
	command.PersistentFlags().StringVar(&yamlConfigFile, "config-file", "", "The custom config file (if not specified the default will be used).")
	command.AddCommand(board.InitCommand())
	command.AddCommand(compile.InitCommand())
	command.AddCommand(config.InitCommand())
	command.AddCommand(core.InitCommand())
	command.AddCommand(daemon.InitCommand())
	command.AddCommand(generatedocs.InitCommand())
	command.AddCommand(lib.InitCommand())
	command.AddCommand(sketch.InitCommand())
	command.AddCommand(upload.InitCommand())
	command.AddCommand(version.InitCommand())
	return command
}

var outputFormat string
var yamlConfigFile string

func preRun(cmd *cobra.Command, args []string) {
	// Reset logrus if debug flag changed.
	if !cli.GlobalFlags.Debug {
		// Discard logrus output if no debug.
		logrus.SetOutput(ioutil.Discard)
	} else {
		// Else print on stderr.

		// Workaround to get colored output on windows
		if terminal.IsTerminal(int(os.Stdout.Fd())) {
			logrus.SetFormatter(&logrus.TextFormatter{ForceColors: true})
		}
		logrus.SetOutput(colorable.NewColorableStdout())
		cli.ErrLogrus.Out = colorable.NewColorableStderr()
		formatter.SetLogger(cli.ErrLogrus)
	}
	initConfigs()

	logrus.Info(cli.VersionInfo.Application + "-" + cli.VersionInfo.VersionString)
	logrus.Info("Starting root command preparation (`arduino`)")
	switch outputFormat {
	case "text":
		formatter.SetFormatter("text")
		cli.GlobalFlags.OutputJSON = false
	case "json":
		formatter.SetFormatter("json")
		cli.GlobalFlags.OutputJSON = true
	default:
		formatter.PrintErrorMessage("Invalid output format: " + outputFormat)
		os.Exit(cli.ErrBadCall)
	}

	logrus.Info("Formatter set")
	if !formatter.IsCurrentFormat("text") {
		cmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
			logrus.Warn("Calling help on JSON format")
			formatter.PrintErrorMessage("Invalid Call : should show Help, but it is available only in TEXT mode.")
			os.Exit(cli.ErrBadCall)
		})
	}
}

// initConfigs initializes the configuration from the specified file.
func initConfigs() {
	// Start with default configuration
	if conf, err := configs.NewConfiguration(); err != nil {
		logrus.WithError(err).Error("Error creating default configuration")
		formatter.PrintError(err, "Error creating default configuration")
		os.Exit(cli.ErrGeneric)
	} else {
		cli.Config = conf
	}

	// Read configuration from global config file
	logrus.Info("Checking for config file in: " + cli.Config.ConfigFile.String())
	if cli.Config.ConfigFile.Exist() {
		readConfigFrom(cli.Config.ConfigFile)
	}

	if cli.Config.IsBundledInDesktopIDE() {
		logrus.Info("CLI is bundled into the IDE")
		err := cli.Config.LoadFromDesktopIDEPreferences()
		if err != nil {
			logrus.WithError(err).Warn("Did not manage to get config file of IDE, using default configuration")
		}
	} else {
		logrus.Info("CLI is not bundled into the IDE")
	}

	// Read configuration from parent folders (project config)
	if pwd, err := paths.Getwd(); err != nil {
		logrus.WithError(err).Warn("Did not manage to find current path")
		if path := paths.New("arduino-cli.yaml"); path.Exist() {
			readConfigFrom(path)
		}
	} else {
		cli.Config.Navigate(pwd)
	}

	// Read configuration from old configuration file if found, but output a warning.
	if old := paths.New(".cli-config.yml"); old.Exist() {
		logrus.Errorf("Old configuration file detected: %s.", old)
		logrus.Info("The name of this file has been changed to `arduino-cli.yaml`, please rename the file fix it.")
		formatter.PrintError(
			fmt.Errorf("WARNING: Old configuration file detected: %s", old),
			"The name of this file has been changed to `arduino-cli.yaml`, in a future release we will not support"+
				"the old name `.cli-config.yml` anymore. Please rename the file to `arduino-cli.yaml` to silence this warning.")
		readConfigFrom(old)
	}

	// Read configuration from environment vars
	cli.Config.LoadFromEnv()

	// Read configuration from user specified file
	if yamlConfigFile != "" {
		cli.Config.ConfigFile = paths.New(yamlConfigFile)
		readConfigFrom(cli.Config.ConfigFile)
	}

	logrus.Info("Configuration set")
}

func readConfigFrom(path *paths.Path) {
	logrus.Infof("Reading configuration from %s", path)
	if err := cli.Config.LoadFromYAML(path); err != nil {
		logrus.WithError(err).Warnf("Could not read configuration from %s", path)
	}
}
