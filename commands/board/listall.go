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

package board

import (
	"sort"
	"strings"

	"github.com/arduino/arduino-cli/commands"
	"github.com/arduino/arduino-cli/common/formatter"
	"github.com/arduino/arduino-cli/common/formatter/output"
	"github.com/spf13/cobra"
)

func initListAllCommand() *cobra.Command {
	listAllCommand := &cobra.Command{
		Use:   "listall [boardname]",
		Short: "List all known boards and their corresponding FQBN.",
		Long: "" +
			"List all boards that have the support platform installed. You can search\n" +
			"for a specific board if you specify the board name",
		Example: "" +
			"  " + commands.AppName + " board listall\n" +
			"  " + commands.AppName + " board listall zero",
		Args: cobra.ArbitraryArgs,
		Run:  runListAllCommand,
	}
	return listAllCommand
}

// runListAllCommand list all installed boards
func runListAllCommand(cmd *cobra.Command, args []string) {
	pm := commands.InitPackageManager()

	match := func(name string) bool {
		name = strings.ToLower(name)
		for _, term := range args {
			if !strings.Contains(name, strings.ToLower(term)) {
				return false
			}
		}
		return true
	}

	list := &output.BoardList{}
	for _, targetPackage := range pm.GetPackages().Packages {
		for _, platform := range targetPackage.Platforms {
			platformRelease := pm.GetInstalledPlatformRelease(platform)
			if platformRelease == nil {
				continue
			}
			for _, board := range platformRelease.Boards {
				if !match(board.Name()) {
					continue
				}
				list.Boards = append(list.Boards, &output.BoardListItem{
					Name: board.Name(),
					Fqbn: board.FQBN(),
				})
			}
		}
	}
	sort.Sort(list)
	formatter.Print(list)
}
