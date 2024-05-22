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

package core

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
	var preUninstallFlags arguments.PrePostScriptsFlags
	uninstallCommand := &cobra.Command{
		Use:     fmt.Sprintf("uninstall %s:%s ...", i18n.Tr("PACKAGER"), i18n.Tr("ARCH")),
		Short:   i18n.Tr("Uninstalls one or more cores and corresponding tool dependencies if no longer used."),
		Long:    i18n.Tr("Uninstalls one or more cores and corresponding tool dependencies if no longer used."),
		Example: "  " + os.Args[0] + " core uninstall arduino:samd\n",
		Args:    cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			runUninstallCommand(cmd.Context(), srv, args, preUninstallFlags)
		},
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return arguments.GetUninstallableCores(cmd.Context(), srv), cobra.ShellCompDirectiveDefault
		},
	}
	preUninstallFlags.AddToCommand(uninstallCommand)
	return uninstallCommand
}

func runUninstallCommand(ctx context.Context, srv rpc.ArduinoCoreServiceServer, args []string, preUninstallFlags arguments.PrePostScriptsFlags) {
	logrus.Info("Executing `arduino-cli core uninstall`")
	inst := instance.CreateAndInit(ctx, srv)

	platformsRefs, err := arguments.ParseReferences(ctx, srv, args)
	if err != nil {
		feedback.Fatal(i18n.Tr("Invalid argument passed: %v", err), feedback.ErrBadArgument)
	}

	for _, platformRef := range platformsRefs {
		if platformRef.Version != "" {
			feedback.Fatal(i18n.Tr("Invalid parameter %s: version not allowed", platformRef), feedback.ErrBadArgument)
		}
	}
	for _, platformRef := range platformsRefs {
		req := &rpc.PlatformUninstallRequest{
			Instance:         inst,
			PlatformPackage:  platformRef.PackageName,
			Architecture:     platformRef.Architecture,
			SkipPreUninstall: preUninstallFlags.DetectSkipPreUninstallValue(),
		}
		stream := commands.PlatformUninstallStreamResponseToCallbackFunction(ctx, feedback.NewTaskProgressCB())
		if err := srv.PlatformUninstall(req, stream); err != nil {
			feedback.Fatal(i18n.Tr("Error during uninstall: %v", err), feedback.ErrGeneric)
		}
	}
}
