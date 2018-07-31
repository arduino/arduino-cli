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
	"github.com/bcmi-labs/arduino-cli/arduino/cores/packagemanager"
	"github.com/bcmi-labs/arduino-cli/arduino/libraries/librariesindex"
	"github.com/bcmi-labs/arduino-cli/commands"
	"github.com/bcmi-labs/arduino-cli/common/formatter"
	"github.com/bcmi-labs/arduino-cli/common/formatter/output"
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
	listCommand.Flags().BoolVar(&listFlags.all, "all", false, "Include built-in libraries (from platforms and IDE) in listing.")
	listCommand.Flags().BoolVar(&listFlags.updatable, "updatable", false, "List updatable libraries.")
	return listCommand
}

var listFlags struct {
	all       bool
	updatable bool
}

func runListCommand(cmd *cobra.Command, args []string) {
	var pm *packagemanager.PackageManager
	if listFlags.all {
		pm = commands.InitPackageManager()
	}
	lm := commands.InitLibraryManager(pm)

	res := output.InstalledLibraries{}
	for _, libAlternatives := range lm.Libraries {
		for _, lib := range libAlternatives.Alternatives {
			var available *librariesindex.Release
			if listFlags.updatable {
				available = lm.Index.FindLibraryUpdate(lib)
				if available == nil {
					continue
				}
			}
			res.Libraries = append(res.Libraries, &output.InstalledLibary{
				Library:   lib,
				Available: available,
			})
		}
	}
	logrus.Info("Listing")

	if len(res.Libraries) > 0 {
		formatter.Print(res)
	}
	logrus.Info("Done")
}
