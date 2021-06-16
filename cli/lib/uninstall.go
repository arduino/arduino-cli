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

package lib

import (
	"context"
	"os"

	"github.com/arduino/arduino-cli/cli/errorcodes"
	"github.com/arduino/arduino-cli/cli/feedback"
	"github.com/arduino/arduino-cli/cli/instance"
	"github.com/arduino/arduino-cli/cli/output"
	"github.com/arduino/arduino-cli/commands/lib"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func initUninstallCommand() *cobra.Command {
	uninstallCommand := &cobra.Command{
		Use:     "uninstall LIBRARY_NAME(S)",
		Short:   "Uninstalls one or more libraries.",
		Long:    "Uninstalls one or more libraries.",
		Example: "  " + os.Args[0] + " lib uninstall AudioZero",
		Args:    cobra.MinimumNArgs(1),
		Run:     runUninstallCommand,
	}
	return uninstallCommand
}

func runUninstallCommand(cmd *cobra.Command, args []string) {
	logrus.Info("Executing `arduino lib uninstall`")

	instance := instance.CreateAndInit()
	refs, err := ParseLibraryReferenceArgsAndAdjustCase(instance, args)
	if err != nil {
		feedback.Errorf("Invalid argument passed: %v", err)
		os.Exit(errorcodes.ErrBadArgument)
	}

	for _, library := range refs {
		err := lib.LibraryUninstall(context.Background(), &rpc.LibraryUninstallRequest{
			Instance: instance,
			Name:     library.Name,
			Version:  library.Version,
		}, output.TaskProgress())
		if err != nil {
			feedback.Errorf("Error uninstalling %s: %v", library, err)
			os.Exit(errorcodes.ErrGeneric)
		}
	}

	logrus.Info("Done")
}
