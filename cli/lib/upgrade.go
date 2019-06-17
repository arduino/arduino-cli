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

	"github.com/arduino/arduino-cli/cli"
	"github.com/arduino/arduino-cli/commands/lib"
	"github.com/arduino/arduino-cli/common/formatter"
	rpc "github.com/arduino/arduino-cli/rpc/commands"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"golang.org/x/net/context"
)

func initUpgradeCommand() *cobra.Command {
	listCommand := &cobra.Command{
		Use:   "upgrade",
		Short: "Upgrades installed libraries.",
		Long: "This command ungrades all installed libraries to the latest available version." +
			"To upgrade a single library use the 'install' command.",
		Example: "  " + cli.VersionInfo.Application + " lib upgrade",
		Args:    cobra.NoArgs,
		Run:     runUpgradeCommand,
	}
	return listCommand
}

func runUpgradeCommand(cmd *cobra.Command, args []string) {
	instance := cli.CreateInstaceIgnorePlatformIndexErrors()

	err := lib.LibraryUpgradeAll(context.Background(), &rpc.LibraryUpgradeAllReq{
		Instance: instance,
	}, cli.OutputProgressBar(), cli.OutputTaskProgress(), cli.HTTPClientHeader)
	if err != nil {
		formatter.PrintError(err, "Error upgrading libraries")
		os.Exit(cli.ErrGeneric)
	}

	logrus.Info("Done")
}
