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

package lib

import (
	"os"

	paths "github.com/arduino/go-paths-helper"
	"github.com/bcmi-labs/arduino-cli/arduino/libraries"
	"github.com/bcmi-labs/arduino-cli/commands"
	"github.com/bcmi-labs/arduino-cli/common/formatter"
	"github.com/bcmi-labs/arduino-cli/common/formatter/output"
	"github.com/bcmi-labs/arduino-cli/configs"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func initListCommand() *cobra.Command {
	listCommand := &cobra.Command{
		Use:   "list",
		Short: "Shows a list of all installed libraries.",
		Long: "Shows a list of all installed libraries.\n" +
			"Can be used with -v (or --verbose) flag (up to 2 times) to have longer output.",
		Example: "" +
			"arduino lib list    # to show all installed library names.\n" +
			"arduino lib list -v # to show more details.",
		Args: cobra.NoArgs,
		Run:  runListCommand,
	}
	return listCommand
}

func runListCommand(cmd *cobra.Command, args []string) {
	libHome, err := configs.LibrariesFolder.Get()
	if err != nil {
		formatter.PrintError(err, "Cannot get libraries folder.")
		os.Exit(commands.ErrCoreConfig)
	}

	libs, err := libraries.LoadLibrariesFromDir(paths.New(libHome))
	if err != nil {
		formatter.PrintError(err, "Error loading libraries.")
		os.Exit(commands.ErrCoreConfig)
	}

	res := output.InstalledLibraries{}
	res.Libraries = append(res.Libraries, libs...)
	logrus.Info("Listing")

	if len(libs) < 1 {
		formatter.PrintErrorMessage("No library installed.")
	} else {
		formatter.Print(res)
	}
	logrus.Info("Done")
}
