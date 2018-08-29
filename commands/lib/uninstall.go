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
	"os"

	"github.com/arduino/arduino-cli/arduino/libraries/librariesindex"
	"github.com/arduino/arduino-cli/commands"
	"github.com/arduino/arduino-cli/common/formatter"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// These const are unused
//const (
//	versionAll    string = "all"
//	versionLatest string = "latest"
//)

func initUninstallCommand() *cobra.Command {
	uninstallCommand := &cobra.Command{
		Use:     "uninstall LIBRARY_NAME(S)",
		Short:   "Uninstalls one or more libraries.",
		Long:    "Uninstalls one or more libraries.",
		Example: "  " + commands.AppName + " lib uninstall AudioZero",
		Args:    cobra.MinimumNArgs(1),
		Run:     runUninstallCommand,
	}
	return uninstallCommand
}

func runUninstallCommand(cmd *cobra.Command, args []string) {
	logrus.Info("Executing `arduino lib uninstall`")

	lm := commands.InitLibraryManager(nil)
	libRefs, err := librariesindex.ParseArgs(args)
	if err != nil {
		formatter.PrintError(err, "Arguments error")
		os.Exit(commands.ErrBadArgument)
	}
	for _, libRef := range libRefs {
		lib := lm.FindByReference(libRef)
		if lib == nil {
			formatter.PrintErrorMessage("Library not installed: " + libRef.String())
			os.Exit(commands.ErrGeneric)
		} else {
			formatter.Print("Uninstalling " + lib.String())
			lm.Uninstall(lib)
		}
	}

	logrus.Info("Done")
}
