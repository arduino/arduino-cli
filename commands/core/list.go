/*
 * This file is part of arduino-cli.
 *
 * arduino-cli is free software; you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation; either version 2 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program; if not, write to the Free Software
 * Foundation, Inc., 51 Franklin St, Fifth Floor, Boston, MA  02110-1301  USA
 *
 * As a special exception, you may use this file as part of a free software
 * library without restriction.  Specifically, if other files instantiate
 * templates or use macros or inline functions from this file, or you compile
 * this file and link it with other files to produce an executable, this
 * file does not by itself cause the resulting executable to be covered by
 * the GNU General Public License.  This exception does not however
 * invalidate any other reasons why the executable file might be covered by
 * the GNU General Public License.
 *
 * Copyright 2017 ARDUINO AG (http://www.arduino.cc/)
 */

package core

import (
	"github.com/bcmi-labs/arduino-cli/commands"
	"github.com/bcmi-labs/arduino-cli/common/formatter"
	"github.com/bcmi-labs/arduino-cli/common/formatter/output"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func initListCommand() *cobra.Command {
	listCommand := &cobra.Command{
		Use:   "list",
		Short: "Shows the list of installed cores.",
		Long: "Shows the list of installed cores.\n" +
			"With -v tag (up to 2 times) can provide more verbose output.",
		Example: "arduino core list -v # for a medium verbosity level.",
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
