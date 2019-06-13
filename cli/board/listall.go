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
	"context"
	"fmt"
	"os"
	"sort"

	"github.com/arduino/arduino-cli/cli"
	"github.com/arduino/arduino-cli/commands/board"
	"github.com/arduino/arduino-cli/common/formatter"
	"github.com/arduino/arduino-cli/output"
	rpc "github.com/arduino/arduino-cli/rpc/commands"
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
			"  " + cli.VersionInfo.Application + " board listall\n" +
			"  " + cli.VersionInfo.Application + " board listall zero",
		Args: cobra.ArbitraryArgs,
		Run:  runListAllCommand,
	}
	return listAllCommand
}

// runListAllCommand list all installed boards
func runListAllCommand(cmd *cobra.Command, args []string) {
	instance := cli.CreateInstance()

	list, err := board.ListAll(context.Background(), &rpc.BoardListAllReq{
		Instance:   instance,
		SearchArgs: args,
	})
	if err != nil {
		formatter.PrintError(err, "Error listing boards")
		os.Exit(cli.ErrGeneric)
	}
	if cli.OutputJSONOrElse(list) {
		outputBoardListAll(list)
	}
}

func outputBoardListAll(list *rpc.BoardListAllResp) {
	sort.Slice(list.Boards, func(i, j int) bool {
		return list.Boards[i].GetName() < list.Boards[j].GetName()
	})

	table := output.NewTable()
	table.SetHeader("Board Name", "FQBN")
	for _, item := range list.GetBoards() {
		table.AddRow(item.GetName(), item.GetFQBN())
	}
	fmt.Print(table.Render())
}
