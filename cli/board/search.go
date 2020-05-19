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
	rpc "github.com/arduino/arduino-cli/rpc/commands"
	"github.com/arduino/arduino-cli/table"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func initSearchCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "search <KEYWORDS>...",
		Short: "Search for a board.",
		Long: "List all boards that have name or description or other properties matching \n" +
			"the provided KEYWORDS.",
		Example: "  # search info about Arduino Zero.\n" +
			"  " + os.Args[0] + " board search zero",
		Args: cobra.MinimumNArgs(1),
		Run:  runSearchCommand,
	}
	return cmd
}

func runSearchCommand(cmd *cobra.Command, args []string) {
	inst, err := instance.CreateInstance()
	if err != nil {
		feedback.Errorf("Error searching: %v", err)
		os.Exit(errorcodes.ErrGeneric)
	}

	query := strings.Join(args[1:], " ")
	logrus.WithField("query", query).Info("Executing `board search`")

	boardSearchReq := &rpc.BoardSearchReq{
		Instance: inst,
		Query:    query,
	}
	if res, err := board.Search(context.Background(), boardSearchReq); err != nil {
		feedback.Errorf("Error searching for board: %v", err)
		os.Exit(errorcodes.ErrGeneric)
	} else {
		feedback.PrintResult(searchResult{res.GetBoards()})
	}
}

// output from this command requires special formatting, let's create a dedicated
// feedback.Result implementation
type searchResult struct {
	boards []*rpc.SearchedBoard
}

func (res searchResult) Data() interface{} {
	return res.boards
}

func (res searchResult) String() string {
	if len(res.boards) == 0 {
		return "No boards found."
	}

	sort.Slice(res.boards, func(i, j int) bool {
		x, y := res.boards[i], res.boards[j]
		return x.GetName() < y.GetName()
	})

	t := table.New()
	t.SetHeader("Board", "Board Name", "FQBN", "Platform")
	for _, board := range res.boards {
		platform := "Arduino"
		if board.GetExternalPlatformUrl() != "" {
			platform = "3rd party"
		}
		t.AddRow(board.GetAlias(), board.GetName(), board.GetFqbn(), platform)
	}
	return t.Render()
}
