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
	rpc "github.com/arduino/arduino-cli/rpc/commands"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func initUpgradeCommand() *cobra.Command {
	upgradeCommand := &cobra.Command{
		Use:   "upgrade [PACKAGER:ARCH] ...",
		Short: "Upgrades one or all installed platforms to the latest version.",
		Long:  "Upgrades one or all installed platforms to the latest version.",
		Example: "" +
			"  # upgrade everything to the latest version\n" +
			"  " + cli.VersionInfo.Application + " core upgrade\n\n" +
			"  # upgrade arduino:samd to the latest version\n" +
			"  " + cli.VersionInfo.Application + " core upgrade arduino:samd",
		Run: runUpgradeCommand,
	}
	return upgradeCommand
}

func runUpgradeCommand(cmd *cobra.Command, args []string) {
	instance := cli.CreateInstance()
	logrus.Info("Executing `arduino core upgrade`")

	platformsRefs := parsePlatformReferenceArgs(args)
	for i, platformRef := range platformsRefs {
		if platformRef.Version != "" {
			formatter.PrintErrorMessage(("Invalid item " + args[i]))
			os.Exit(cli.ErrBadArgument)
		}
	}
	for _, platformRef := range platformsRefs {
		_, err := core.PlatformUpgrade(context.Background(), &rpc.PlatformUpgradeReq{
			Instance:        instance,
			PlatformPackage: platformRef.Package,
			Architecture:    platformRef.Architecture,
		}, cli.OutputProgressBar(), cli.OutputTaskProgress(),
			cli.HTTPClientHeader)
		if err != nil {
			formatter.PrintError(err, "Error during upgrade")
			os.Exit(cli.ErrGeneric)
		}
	}
}
