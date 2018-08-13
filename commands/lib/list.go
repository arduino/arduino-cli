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

package lib

import (
	"github.com/arduino/arduino-cli/arduino/cores/packagemanager"
	"github.com/arduino/arduino-cli/arduino/libraries/librariesindex"
	"github.com/arduino/arduino-cli/arduino/libraries/librariesmanager"
	"github.com/arduino/arduino-cli/commands"
	"github.com/arduino/arduino-cli/common/formatter"
	"github.com/arduino/arduino-cli/common/formatter/output"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func initListCommand() *cobra.Command {
	listCommand := &cobra.Command{
		Use:     "list",
		Short:   "Shows a list of all installed libraries.",
		Long:    "Shows a list of all installed libraries.",
		Example: "  " + commands.AppName + " lib list",
		Args:    cobra.NoArgs,
		Run:     runListCommand,
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
	logrus.Info("Listing")
	var pm *packagemanager.PackageManager
	if listFlags.all {
		pm = commands.InitPackageManager()
	}
	lm := commands.InitLibraryManager(pm)

	res := listLibraries(lm, listFlags.updatable)
	if len(res.Libraries) > 0 {
		formatter.Print(res)
	}
	logrus.Info("Done")
}

func listLibraries(lm *librariesmanager.LibrariesManager, updatable bool) *output.InstalledLibraries {
	res := &output.InstalledLibraries{}
	for _, libAlternatives := range lm.Libraries {
		for _, lib := range libAlternatives.Alternatives {
			var available *librariesindex.Release
			if updatable {
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
	return res
}
