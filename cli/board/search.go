// This file is part of arduino-cli.
//
// Copyright 2020 ARDUINO SA (http://www.arduino.cc/)
//
// This software is released under the GNU General Public License version 3,
// which covers the main part of arduino-cli.
// The terms of this license can be found at:
// https://www.gnu.org/licenses/gpl-3.0.en.html
//
// You can be released from the requirements of the above licenses by purchasing
// a commercial license. Buying such a license is mandatory if you want to
// modify or otherwise use the software for commercial activities involving the
// Arduino software without disclosing the source code of your own applications.
// To purchase a commercial license, send an email to license@arduino.cc.

package board

import (
	"context"
	"os"
	"sort"
	"strings"

	"github.com/arduino/arduino-cli/cli/errorcodes"
	"github.com/arduino/arduino-cli/cli/feedback"
	"github.com/arduino/arduino-cli/cli/instance"
	"github.com/arduino/arduino-cli/commands/board"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	"github.com/arduino/arduino-cli/table"
	"github.com/spf13/cobra"
)

func initSearchCommand() *cobra.Command {
	var searchCommand = &cobra.Command{
		Use:   "search [boardname]",
		Short: "List all known boards and their corresponding FQBN.",
		Long: "" +
			"List all boards that have the support platform installed. You can search\n" +
			"for a specific board if you specify the board name",
		Example: "" +
			"  " + os.Args[0] + " board search\n" +
			"  " + os.Args[0] + " board search zero",
		Args: cobra.ArbitraryArgs,
		Run:  runSearchCommand,
	}
	searchCommand.Flags().BoolVarP(&searchFlags.showHiddenBoard, "show-hidden", "a", false, "Show also boards marked as 'hidden' in the platform")
	return searchCommand
}

var searchFlags struct {
	showHiddenBoard bool
}

func runSearchCommand(cmd *cobra.Command, args []string) {
	inst := instance.CreateAndInit()

	res, err := board.Search(context.Background(), &rpc.BoardSearchRequest{
		Instance:            inst,
		SearchArgs:          strings.Join(args, " "),
		IncludeHiddenBoards: searchFlags.showHiddenBoard,
	})
	if err != nil {
		feedback.Errorf("Error searching boards: %v", err)
		os.Exit(errorcodes.ErrGeneric)
	}

	feedback.PrintResult(searchResults{res.Boards})
}

// output from this command requires special formatting so we create a dedicated
// feedback.Result implementation
type searchResults struct {
	boards []*rpc.BoardListItem
}

func (r searchResults) Data() interface{} {
	return r.boards
}

func (r searchResults) String() string {
	sort.Slice(r.boards, func(i, j int) bool {
		return r.boards[i].GetName() < r.boards[j].GetName()
	})

	t := table.New()
	t.SetHeader("Board Name", "FQBN", "Platform ID", "")
	for _, item := range r.boards {
		hidden := ""
		if item.IsHidden {
			hidden = "(hidden)"
		}
		t.AddRow(item.GetName(), item.GetFqbn(), item.Platform.Id, hidden)
	}
	return t.Render()
}
