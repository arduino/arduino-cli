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
 * Copyright 2018 ARDUINO AG (http://www.arduino.cc/)
 */

package core

import (
	"os"
	"strings"

	"github.com/bcmi-labs/arduino-cli/commands"
	"github.com/bcmi-labs/arduino-cli/common/formatter"
	"github.com/bcmi-labs/arduino-cli/common/formatter/output"
	"github.com/spf13/cobra"
)

func initSearchCommand() *cobra.Command {
	searchCommand := &cobra.Command{
		Use:     "search <keywords...>",
		Short:   "Search for a core in the package index.",
		Long:    "Search for a core in the package index using the specified keywords.",
		Example: "arduino core search MKRZero -v",
		Args:    cobra.MinimumNArgs(1),
		Run:     runSearchCommand,
	}
	return searchCommand
}

func runSearchCommand(cmd *cobra.Command, args []string) {
	pm := commands.InitPackageManager()
	if err := pm.LoadHardware(); err != nil {
		formatter.PrintError(err, "Error loading hardware packages")
		os.Exit(commands.ErrCoreConfig)
	}

	search := strings.ToLower(strings.Join(args, " "))
	formatter.Print("Searching for platforms matching '" + search + "'")
	formatter.Print("")

	match := func(line string) bool {
		return strings.Contains(strings.ToLower(line), search)
	}
	res := output.PlatformReleases{}
	for _, targetPackage := range pm.GetPackages().Packages {
		for _, platform := range targetPackage.Platforms {
			platformRelease := platform.GetLatestRelease()
			if platformRelease == nil {
				continue
			}
			if match(platform.Name) || match(platform.Architecture) {
				res = append(res, platformRelease)
				continue
			}
			for _, boards := range platformRelease.BoardsManifest {
				if match(boards.Name) {
					res = append(res, platformRelease)
					break
				}
			}
		}
	}

	if len(res) == 0 {
		formatter.Print("No platforms matching your search")
	} else {
		formatter.Print(res)
	}
}
