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
	"github.com/arduino/arduino-cli/commands"
	"github.com/arduino/arduino-cli/common/formatter"
	"github.com/arduino/arduino-cli/common/formatter/output"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func initListCommand() *cobra.Command {
	listCommand := &cobra.Command{
		Use:   "list",
		Short: "Shows the list of installed cores.",
		Long: "Shows the list of installed cores.\n" +
			"With -v tag (up to 2 times) can provide more verbose output.",
		Example: "  " + commands.AppName + "core list -v # for a medium verbosity level.",
		Args:    cobra.NoArgs,
		Run:     runListCommand,
	}
	return listCommand
}

func runListCommand(cmd *cobra.Command, args []string) {
	logrus.Info("Executing `arduino core list`")

	pm := commands.InitPackageManager()

	res := output.InstalledPlatformReleases{}
	for _, targetPackage := range pm.GetPackages().Packages {
		for _, platform := range targetPackage.Platforms {
			if platformRelease := platform.GetInstalled(); platformRelease != nil {
				res = append(res, platformRelease)
			}
		}
	}

	if len(res) > 0 {
		formatter.Print(res)
	}
}
