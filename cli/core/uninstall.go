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

package core

import (
	"context"
	"os"

	"github.com/arduino/arduino-cli/cli"
	"github.com/arduino/arduino-cli/commands/core"
	"github.com/arduino/arduino-cli/common/formatter"
	"github.com/arduino/arduino-cli/output"
	rpc "github.com/arduino/arduino-cli/rpc/commands"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func initUninstallCommand() *cobra.Command {
	return &cobra.Command{
		Use:     "uninstall PACKAGER:ARCH ...",
		Short:   "Uninstalls one or more cores and corresponding tool dependencies if no more used.",
		Long:    "Uninstalls one or more cores and corresponding tool dependencies if no more used.",
		Example: "  " + cli.VersionInfo.Application + " core uninstall arduino:samd\n",
		Args:    cobra.MinimumNArgs(1),
		Run:     runUninstallCommand,
	}
}

func runUninstallCommand(cmd *cobra.Command, args []string) {
	instance := cli.CreateInstance()
	logrus.Info("Executing `arduino core uninstall`")

	platformsRefs := parsePlatformReferenceArgs(args)

	for _, platformRef := range platformsRefs {
		if platformRef.Version != "" {
			formatter.PrintErrorMessage("Invalid parameter " + platformRef.String() + ": version not allowed")
			os.Exit(cli.ErrBadArgument)
		}
	}
	for _, platformRef := range platformsRefs {
		_, err := core.PlatformUninstall(context.Background(), &rpc.PlatformUninstallReq{
			Instance:        instance,
			PlatformPackage: platformRef.Package,
			Architecture:    platformRef.Architecture,
		}, output.NewTaskProgressCB())
		if err != nil {
			formatter.PrintError(err, "Error during uninstall")
			os.Exit(cli.ErrGeneric)
		}
	}
}
