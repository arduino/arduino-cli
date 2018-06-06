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
	"github.com/bcmi-labs/arduino-cli/commands/upload"
	"github.com/bcmi-labs/arduino-cli/commands/validate"
	"github.com/bcmi-labs/arduino-cli/commands/version"
	"github.com/bcmi-labs/arduino-cli/common/formatter"
	"github.com/bcmi-labs/arduino-cli/configs"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// Init prepares the cobra root command.
func Init() *cobra.Command {
	command := &cobra.Command{
		Use:              "arduino-cli",
		Short:            "Arduino CLI.",
		Long:             "Arduino Command Line Interface (arduino-cli).",
		Example:          "arduino <command> [flags...]",
		PersistentPreRun: preRun,
	}
	command.PersistentFlags().BoolVar(&commands.GlobalFlags.Debug, "debug", false, "Enables debug output (super verbose, used to debug the CLI).")
	command.PersistentFlags().StringVar(&commands.GlobalFlags.Format, "format", "text", "The output format, can be [text|json].")
	command.PersistentFlags().StringVar(&configs.ConfigFilePath, "config-file", configs.ConfigFilePath, "The custom config file (if not specified ./.cli-config.yml will be used).")
	command.AddCommand(board.InitCommand())
	command.AddCommand(compile.InitCommand())
	command.AddCommand(config.InitCommand())
	command.AddCommand(core.InitCommand())
	command.AddCommand(generatedocs.InitCommand())
	command.AddCommand(lib.InitCommand())
	command.AddCommand(login.InitCommand())
	command.AddCommand(logout.InitCommand())
	command.AddCommand(sketch.InitCommand())
	command.AddCommand(upload.InitCommand())
	command.AddCommand(validate.InitCommand())
	command.AddCommand(version.InitCommand())
	return command
}

func preRun(cmd *cobra.Command, args []string) {
	// Reset logrus if debug flag changed.
	if !commands.GlobalFlags.Debug {
		// Discard logrus output if no debug.
		logrus.SetOutput(ioutil.Discard)
	} else {
		// Else print on stderr.
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
}

// initConfigs initializes the configuration from the specified file.
func initConfigs() {
	logrus.Info("Initiating configuration")
	err := configs.LoadFromYAML(configs.ConfigFilePath)
	if err != nil {
		logrus.WithError(err).Warn("Did not manage to get config file, using default configuration")
	}
	if configs.IsBundledInDesktopIDE() {
		logrus.Info("CLI is bundled into the IDE")
		err := configs.LoadFromDesktopIDEPreferences()
		if err != nil {
			logrus.WithError(err).Warn("Did not manage to get config file of IDE, using default configuration")
		}
	} else {
		logrus.Info("CLI is not bundled into the IDE")
	}
	configs.LoadFromEnv()
	logrus.Info("Configuration set")
}
