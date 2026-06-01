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
	"fmt"
	"os"

	"github.com/arduino/arduino-cli/commands"
	"github.com/arduino/arduino-cli/internal/cli/arguments"
	"github.com/arduino/arduino-cli/internal/cli/feedback"
	"github.com/arduino/arduino-cli/internal/cli/instance"
	"github.com/arduino/arduino-cli/internal/i18n"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func initUninstallCommand(srv rpc.ArduinoCoreServiceServer) *cobra.Command {
	uninstallCommand := &cobra.Command{
		Use:     fmt.Sprintf("uninstall %s...", i18n.Tr("LIBRARY_NAME")),
		Short:   i18n.Tr("Uninstalls one or more libraries."),
		Long:    i18n.Tr("Uninstalls one or more libraries."),
		Example: "  " + os.Args[0] + " lib uninstall AudioZero",
		Args:    cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			runUninstallCommand(cmd.Context(), srv, args)
		},
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return arguments.GetUninstallableLibraries(cmd.Context(), srv), cobra.ShellCompDirectiveDefault
		},
	}
	return uninstallCommand
}

func runUninstallCommand(ctx context.Context, srv rpc.ArduinoCoreServiceServer, args []string) {
	logrus.Info("Executing `arduino-cli lib uninstall`")
	instance := instance.CreateAndInit(ctx, srv)

	refs, err := ParseLibraryReferenceArgsAndAdjustCase(ctx, srv, instance, args)
	if err != nil {
		feedback.Fatal(i18n.Tr("Invalid argument passed: %v", err), feedback.ErrBadArgument)
	}

	for _, library := range refs {
		req := &rpc.LibraryUninstallRequest{
			Instance: instance,
			Name:     library.Name,
			Version:  library.Version,
		}
		stream := commands.LibraryUninstallStreamResponseToCallbackFunction(ctx, feedback.TaskProgress())
		if err := srv.LibraryUninstall(req, stream); err != nil {
			feedback.Fatal(i18n.Tr("Error uninstalling %[1]s: %[2]v", library, err), feedback.ErrGeneric)
		}
	}

	logrus.Info("Done")
}
