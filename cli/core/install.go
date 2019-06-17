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

func initInstallCommand() *cobra.Command {
	installCommand := &cobra.Command{
		Use:   "install PACKAGER:ARCH[@VERSION] ...",
		Short: "Installs one or more cores and corresponding tool dependencies.",
		Long:  "Installs one or more cores and corresponding tool dependencies.",
		Example: "  # download the latest version of arduino SAMD core.\n" +
			"  " + cli.VersionInfo.Application + " core install arduino:samd\n\n" +
			"  # download a specific version (in this case 1.6.9).\n" +
			"  " + cli.VersionInfo.Application + " core install arduino:samd@1.6.9",
		Args: cobra.MinimumNArgs(1),
		Run:  runInstallCommand,
	}
	return installCommand
}

func runInstallCommand(cmd *cobra.Command, args []string) {
	instance := cli.CreateInstance()
	logrus.Info("Executing `arduino core install`")

	platformsRefs := parsePlatformReferenceArgs(args)

	for _, platformRef := range platformsRefs {
		plattformInstallReq := &rpc.PlatformInstallReq{
			Instance:        instance,
			PlatformPackage: platformRef.Package,
			Architecture:    platformRef.Architecture,
			Version:         platformRef.Version,
		}
		_, err := core.PlatformInstall(context.Background(), plattformInstallReq, cli.OutputProgressBar(),
			cli.OutputTaskProgress(), cli.HTTPClientHeader)
		if err != nil {
			formatter.PrintError(err, "Error during install")
			os.Exit(cli.ErrGeneric)
		}
	}
}
