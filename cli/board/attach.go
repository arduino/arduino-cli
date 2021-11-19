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

	"github.com/arduino/arduino-cli/cli/arguments"
	"github.com/arduino/arduino-cli/cli/errorcodes"
	"github.com/arduino/arduino-cli/cli/feedback"
	"github.com/arduino/arduino-cli/cli/instance"
	"github.com/arduino/arduino-cli/cli/output"
	"github.com/arduino/arduino-cli/commands/board"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	"github.com/spf13/cobra"
)

var searchTimeout string // Expressed in a parsable duration, is the timeout for the list and attach commands.

func initAttachCommand() *cobra.Command {
	attachCommand := &cobra.Command{
		Use:   fmt.Sprintf("attach <%s>|<%s> [%s]", tr("port"), tr("FQBN"), tr("sketchPath")),
		Short: tr("Attaches a sketch to a board."),
		Long:  tr("Attaches a sketch to a board."),
		Example: "  " + os.Args[0] + " board attach serial:///dev/ttyACM0\n" +
			"  " + os.Args[0] + " board attach serial:///dev/ttyACM0 HelloWorld\n" +
			"  " + os.Args[0] + " board attach arduino:samd:mkr1000",
		Args: cobra.RangeArgs(1, 2),
		Run:  runAttachCommand,
	}
	attachCommand.Flags().StringVar(&searchTimeout, "timeout", "5s",
		tr("The connected devices search timeout, raise it if your board doesn't show up (e.g. to %s).", "10s"))
	return attachCommand
}

func runAttachCommand(cmd *cobra.Command, args []string) {
	instance := instance.CreateAndInit()

	path := ""
	if len(args) > 1 {
		path = args[1]
	}
	sketchPath := arguments.InitSketchPath(path)

	if _, err := board.Attach(context.Background(), &rpc.BoardAttachRequest{
		Instance:      instance,
		BoardUri:      args[0],
		SketchPath:    sketchPath.String(),
		SearchTimeout: searchTimeout,
	}, output.TaskProgress()); err != nil {
		feedback.Errorf(tr("Attach board error: %v"), err)
		os.Exit(errorcodes.ErrGeneric)
	}
}
