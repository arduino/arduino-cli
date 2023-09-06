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

	"github.com/arduino/arduino-cli/commands/core"
	"github.com/arduino/arduino-cli/internal/cli/arguments"
	"github.com/arduino/arduino-cli/internal/cli/feedback"
	"github.com/arduino/arduino-cli/internal/cli/instance"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func initUninstallCommand() *cobra.Command {
	var preUninstallFlags arguments.PrePostScriptsFlags
	uninstallCommand := &cobra.Command{
		Use:     fmt.Sprintf("uninstall %s:%s ...", tr("PACKAGER"), tr("ARCH")),
		Short:   tr("Uninstalls one or more cores and corresponding tool dependencies if no longer used."),
		Long:    tr("Uninstalls one or more cores and corresponding tool dependencies if no longer used."),
		Example: "  " + os.Args[0] + " core uninstall arduino:samd\n",
		Args:    cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			runUninstallCommand(args, preUninstallFlags)
		},
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return arguments.GetUninstallableCores(), cobra.ShellCompDirectiveDefault
		},
	}
	preUninstallFlags.AddToCommand(uninstallCommand)
	return uninstallCommand
}

func runUninstallCommand(args []string, preUninstallFlags arguments.PrePostScriptsFlags) {
	inst := instance.CreateAndInit()
	logrus.Info("Executing `arduino-cli core uninstall`")

	platformsRefs, err := arguments.ParseReferences(args)
	if err != nil {
		feedback.Fatal(tr("Invalid argument passed: %v", err), feedback.ErrBadArgument)
	}

	for _, platformRef := range platformsRefs {
		if platformRef.Version != "" {
			feedback.Fatal(tr("Invalid parameter %s: version not allowed", platformRef), feedback.ErrBadArgument)
		}
	}
	for _, platformRef := range platformsRefs {
		_, err := core.PlatformUninstall(context.Background(), &rpc.PlatformUninstallRequest{
			Instance:         inst,
			PlatformPackage:  platformRef.PackageName,
			Architecture:     platformRef.Architecture,
			SkipPreUninstall: preUninstallFlags.DetectSkipPreUninstallValue(),
		}, feedback.NewTaskProgressCB())
		if err != nil {
			feedback.Fatal(tr("Error during uninstall: %v", err), feedback.ErrGeneric)
		}
	}
}
