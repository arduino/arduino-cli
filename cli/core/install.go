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

	"github.com/arduino/arduino-cli/cli/arguments"
	"github.com/arduino/arduino-cli/cli/errorcodes"
	"github.com/arduino/arduino-cli/cli/feedback"
	"github.com/arduino/arduino-cli/cli/instance"
	"github.com/arduino/arduino-cli/cli/output"
	"github.com/arduino/arduino-cli/commands/core"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	postInstallFlags arguments.PostInstallFlags
)

func initInstallCommand() *cobra.Command {
	installCommand := &cobra.Command{
		Use:   fmt.Sprintf("install %s:%s[@%s]...", tr("PACKAGER"), tr("ARCH"), tr("VERSION")),
		Short: tr("Installs one or more cores and corresponding tool dependencies."),
		Long:  tr("Installs one or more cores and corresponding tool dependencies."),
		Example: "  # " + tr("download the latest version of Arduino SAMD core.") + "\n" +
			"  " + os.Args[0] + " core install arduino:samd\n\n" +
			"  # " + tr("download a specific version (in this case 1.6.9).") + "\n" +
			"  " + os.Args[0] + " core install arduino:samd@1.6.9",
		Args: cobra.MinimumNArgs(1),
		PreRun: func(cmd *cobra.Command, args []string) {
			arguments.CheckFlagsConflicts(cmd, "run-post-install", "skip-post-install")
		},
		Run: runInstallCommand,
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return arguments.GetInstallableCores(), cobra.ShellCompDirectiveDefault
		},
	}
	postInstallFlags.AddToCommand(installCommand)
	return installCommand
}

func runInstallCommand(cmd *cobra.Command, args []string) {
	instance.Init()
	inst := instance.Get()
	logrus.Info("Executing `arduino-cli core install`")

	platformsRefs, err := arguments.ParseReferences(args)
	if err != nil {
		feedback.Errorf(tr("Invalid argument passed: %v"), err)
		os.Exit(errorcodes.ErrBadArgument)
	}

	for _, platformRef := range platformsRefs {
		platformInstallRequest := &rpc.PlatformInstallRequest{
			Instance:        inst.ToRPC(),
			PlatformPackage: platformRef.PackageName,
			Architecture:    platformRef.Architecture,
			Version:         platformRef.Version,
			SkipPostInstall: postInstallFlags.DetectSkipPostInstallValue(),
		}
		_, err := core.PlatformInstall(context.Background(), platformInstallRequest, output.ProgressBar(), output.TaskProgress())
		if err != nil {
			feedback.Errorf(tr("Error during install: %v"), err)
			os.Exit(errorcodes.ErrGeneric)
		}
	}
}
