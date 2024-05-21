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
	"fmt"
	"os"
	"sort"

	"github.com/arduino/arduino-cli/internal/cli/feedback"
	"github.com/arduino/arduino-cli/internal/cli/feedback/result"
	"github.com/arduino/arduino-cli/internal/cli/feedback/table"
	"github.com/arduino/arduino-cli/internal/cli/instance"
	"github.com/arduino/arduino-cli/internal/i18n"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var showHiddenBoard bool

func initListAllCommand(srv rpc.ArduinoCoreServiceServer) *cobra.Command {
	var listAllCommand = &cobra.Command{
		Use:   fmt.Sprintf("listall [%s]", i18n.Tr("boardname")),
		Short: i18n.Tr("List all known boards and their corresponding FQBN."),
		Long: i18n.Tr(`List all boards that have the support platform installed. You can search
for a specific board if you specify the board name`),
		Example: "" +
			"  " + os.Args[0] + " board listall\n" +
			"  " + os.Args[0] + " board listall zero",
		Args: cobra.ArbitraryArgs,
		Run: func(cmd *cobra.Command, args []string) {
			runListAllCommand(cmd.Context(), args, srv)
		},
	}
	listAllCommand.Flags().BoolVarP(&showHiddenBoard, "show-hidden", "a", false, i18n.Tr("Show also boards marked as 'hidden' in the platform"))
	return listAllCommand
}

// runListAllCommand list all installed boards
func runListAllCommand(ctx context.Context, args []string, srv rpc.ArduinoCoreServiceServer) {
	inst := instance.CreateAndInit(ctx, srv)

	logrus.Info("Executing `arduino-cli board listall`")

	list, err := srv.BoardListAll(ctx, &rpc.BoardListAllRequest{
		Instance:            inst,
		SearchArgs:          args,
		IncludeHiddenBoards: showHiddenBoard,
	})
	if err != nil {
		feedback.Fatal(i18n.Tr("Error listing boards: %v", err), feedback.ErrGeneric)
	}

	feedback.PrintResult(resultAll{result.NewBoardListAllResponse(list)})
}

// output from this command requires special formatting, let's create a dedicated
// feedback.Result implementation
type resultAll struct {
	list *result.BoardListAllResponse
}

func (dr resultAll) Data() interface{} {
	return dr.list
}

func (dr resultAll) String() string {
	t := table.New()
	t.SetHeader(i18n.Tr("Board Name"), i18n.Tr("FQBN"), "")

	if dr.list == nil || len(dr.list.Boards) == 0 {
		return t.Render()
	}

	sort.Slice(dr.list.Boards, func(i, j int) bool {
		return dr.list.Boards[i].Name < dr.list.Boards[j].Name
	})

	for _, item := range dr.list.Boards {
		hidden := ""
		if item.IsHidden {
			hidden = i18n.Tr("(hidden)")
		}
		t.AddRow(item.Name, item.Fqbn, hidden)
	}
	return t.Render()
}
