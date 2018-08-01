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

package lib

import (
	"github.com/bcmi-labs/arduino-cli/arduino/libraries/librariesindex"
	"github.com/bcmi-labs/arduino-cli/commands"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func initUpgradeCommand() *cobra.Command {
	listCommand := &cobra.Command{
		Use:   "upgrade",
		Short: "Upgrades installed libraries.",
		Long: "This command ungrades all installed libraries to the latest available version." +
			"To upgrade a single library use the 'install' command.",
		Example: "arduino lib upgrade",
		Args:    cobra.NoArgs,
		Run:     runUpgradeCommand,
	}
	return listCommand
}

func runUpgradeCommand(cmd *cobra.Command, args []string) {
	lm := commands.InitLibraryManager(nil)
	list := listLibraries(lm, true)
	libReleases := []*librariesindex.Release{}
	for _, upgradeDesc := range list.Libraries {
		libReleases = append(libReleases, upgradeDesc.Available)
	}

	downloadLibraries(lm, libReleases)
	installLibraries(lm, libReleases)
	logrus.Info("Done")
}
