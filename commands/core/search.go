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
	"regexp"
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

	search := strings.ToLower(strings.Join(args, " "))
	formatter.Print("Searching for platforms matching '" + search + "'")
	formatter.Print("")

	res := output.PlatformReleases{}
	if isUsb, _ := regexp.MatchString("[0-9a-f]{4}:[0-9a-f]{4}", search); isUsb {
		vid, pid := search[:4], search[5:]
		res = pm.FindPlatformReleaseProvidingBoardsWithVidPid(vid, pid)
	} else {
		match := func(line string) bool {
			return strings.Contains(strings.ToLower(line), search)
		}
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
				for _, board := range platformRelease.BoardsManifest {
					if match(board.Name) {
						res = append(res, platformRelease)
						break
					}
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
