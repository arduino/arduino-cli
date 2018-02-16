/*
 * This file is part of arduino-cli.
 *
 * arduino-cli is free software; you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation; either version 2 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program; if not, write to the Free Software
 * Foundation, Inc., 51 Franklin St, Fifth Floor, Boston, MA  02110-1301  USA
 *
 * As a special exception, you may use this file as part of a free software
 * library without restriction.  Specifically, if other files instantiate
 * templates or use macros or inline functions from this file, or you compile
 * this file and link it with other files to produce an executable, this
 * file does not by itself cause the resulting executable to be covered by
 * the GNU General Public License.  This exception does not however
 * invalidate any other reasons why the executable file might be covered by
 * the GNU General Public License.
 *
 * Copyright 2017 ARDUINO AG (http://www.arduino.cc/)
 */

package root

import (
	"io/ioutil"
	"os"
	"strings"

	"github.com/bcmi-labs/arduino-cli/commands"
	"github.com/bcmi-labs/arduino-cli/commands/board"
	"github.com/bcmi-labs/arduino-cli/commands/compile"
	"github.com/bcmi-labs/arduino-cli/commands/config"
	"github.com/bcmi-labs/arduino-cli/commands/core"
	"github.com/bcmi-labs/arduino-cli/commands/generatedocs"
	"github.com/bcmi-labs/arduino-cli/commands/lib"
	"github.com/bcmi-labs/arduino-cli/commands/login"
	"github.com/bcmi-labs/arduino-cli/commands/logout"
	"github.com/bcmi-labs/arduino-cli/commands/sketch"
	"github.com/bcmi-labs/arduino-cli/commands/validate"
	"github.com/bcmi-labs/arduino-cli/commands/version"
	"github.com/bcmi-labs/arduino-cli/common/formatter"
	"github.com/bcmi-labs/arduino-cli/configs"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	bashAutoCompletionFunction = `
	__arduino_autocomplete()
	{
		case $(last_command) in
			arduino)
				opts="board compile config core generate-docs help lib login sketch version"
				;;
			arduino_board)
				opts="attach list"
				;;
			arduino_config)
				opts="init"
				;;
			arduino_core)
				opts="download install list search uninstall --update-index"
				;;
			arduino_help)
				opts="board compile config core generate-docs lib login sketch version"
				;;
			arduino_lib)
				opts="download install list search uninstall --update-index"
				;;
			arduino_sketch)
				opts="sync"
				;;
		esac
		if [[ ${cur} == " *" ]] ; then
			COMPREPLY=( $(compgen -W "${opts}" -- ${cur}) )
			return 0
		fi
		return 1
	}`
)

var testing = false

// Init prepares the command.
func Init() {
	Command.PersistentFlags().BoolVar(&commands.GlobalFlags.Debug, "debug", false, "Enables debug output (super verbose, used to debug the CLI).")
	Command.PersistentFlags().StringVar(&commands.GlobalFlags.Format, "format", "text", "The output format, can be [text|json].")
	Command.PersistentFlags().StringVar(&configs.ConfigFilePath, "config-file", configs.ConfigFilePath, "The custom config file (if not specified ./.cli-config.yml will be used).")
	board.Init(Command)
	compile.Init(Command)
	config.Init(Command)
	core.Init(Command)
	generatedocs.Init(Command)
	lib.Init(Command)
	login.Init(Command)
	logout.Init(Command)
	sketch.Init(Command)
	validate.Init(Command)
	version.Init(Command)
}

// Command represents the base command when called without any subcommands.
var Command = &cobra.Command{
	Use:                    "arduino",
	Short:                  "Arduino CLI.",
	Long:                   "Arduino Create Command Line Interface (arduino-cli).",
	Example:                "arduino generate-docs # To generate the docs and autocompletion for the whole CLI.",
	BashCompletionFunction: bashAutoCompletionFunction,
	PersistentPreRun:       preRun,
}

func preRun(cmd *cobra.Command, args []string) {
	// Reset logrus if debug flag changed.
	if !commands.GlobalFlags.Debug { // Discard logrus output if no debug.
		logrus.SetOutput(ioutil.Discard) // For standard logger.
	} else { // Else print on stderr.
		commands.ErrLogrus.Out = os.Stderr
		formatter.SetLogger(commands.ErrLogrus)
	}
	initConfigs()

	logrus.Info("Starting root command preparation (`arduino`)")
	if !formatter.IsSupported(commands.GlobalFlags.Format) {
		logrus.WithField("inserted format", commands.GlobalFlags.Format).Warn("Unsupported format, using text as default")
		commands.GlobalFlags.Format = "text"
	}
	formatter.SetFormatter(commands.GlobalFlags.Format)
	logrus.Info("Formatter set")
	if !formatter.IsCurrentFormat("text") {
		cmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
			logrus.Warn("Calling help on JSON format")
			formatter.PrintErrorMessage("Invalid Call : should show Help, but it is available only in TEXT mode.")
			os.Exit(commands.ErrBadCall)
		})
	}

	if !testing {
		logrus.Info("Initializing viper configuration")
		cobra.OnInitialize(initViper)
	}
}

// initConfigs initializes the configuration from the specified file.
func initConfigs() {
	logrus.Info("Initiating configuration")
	err := configs.Unserialize(configs.ConfigFilePath)
	if err != nil {
		logrus.WithError(err).Warn("Did not manage to get config file, using default configuration")
	}
	if configs.Bundled() {
		logrus.Info("CLI is bundled into the IDE")
		err := configs.UnserializeFromIDEPreferences()
		if err != nil {
			logrus.WithError(err).Warn("Did not manage to get config file of IDE, using default configuration")
		}
	} else {
		logrus.Info("CLI is not bundled into the IDE")
	}
	configs.LoadFromEnv()
	logrus.Info("Configuration set")
}

func initViper() {
	logrus.Info("Initiating viper config")

	defHome, err := configs.ArduinoHomeFolder.Get()
	if err != nil {
		commands.ErrLogrus.WithError(err).Warn("Cannot get default Arduino Home")
	}
	defArduinoData, err := configs.ArduinoDataFolder.Get()
	if err != nil {
		logrus.WithError(err).Warn("Cannot get default Arduino folder")
	}

	viper.SetConfigName(".cli-config")
	viper.AddConfigPath(".")
	viper.SetConfigType("yaml")

	logrus.Info("Reading configuration for viper")
	err = viper.ReadInConfig()
	if err != nil {
		formatter.PrintError(err, "Cannot read configuration file in any of the default folders.")
		os.Exit(commands.ErrNoConfigFile)
	}

	logrus.Info("Setting defaults")
	viper.SetDefault("paths.sketchbook", defHome)
	viper.SetDefault("paths.arduino_data", defArduinoData)
	viper.SetDefault("proxy.type", "auto")
	viper.SetDefault("proxy.hostname", "")
	viper.SetDefault("proxy.username", "")
	viper.SetDefault("proxy.password", "")

	viper.AutomaticEnv()

	logrus.Info("Setting proxy")
	if viper.GetString("proxy.type") == "manual" {
		hostname := viper.GetString("proxy.hostname")
		if hostname == "" {
			commands.ErrLogrus.Error("With manual proxy configuration, hostname is required.")
			formatter.PrintErrorMessage("With manual proxy configuration, hostname is required.")
			os.Exit(commands.ErrCoreConfig)
		}

		if strings.HasPrefix(hostname, "http") {
			os.Setenv("HTTP_PROXY", hostname)
		}
		if strings.HasPrefix(hostname, "https") {
			os.Setenv("HTTPS_PROXY", hostname)
		}

		username := viper.GetString("proxy.username")
		if username != "" { // Put username and pass somewhere.

		}
	}
	logrus.Info("Done viper configuration loading")
}

// TestInit creates an initialization for tests.
func TestInit() {
	initConfigs()

	cobra.OnInitialize(func() {
		viper.SetConfigFile("./test-config.yml")
	})

	testing = true
}

// IgnoreConfigs is used in tests to ignore the config file.
func IgnoreConfigs() {
	logrus.Info("Ignoring configurations and using always default ones")
}
