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
	"os"

	"github.com/arduino/arduino-cli/cli"
	"github.com/arduino/arduino-cli/commands/board"
	"github.com/arduino/arduino-cli/common/formatter"
	rpc "github.com/arduino/arduino-cli/rpc/commands"
	"github.com/spf13/cobra"
)

func initAttachCommand() *cobra.Command {
	attachCommand := &cobra.Command{
		Use:   "attach <port>|<FQBN> [sketchPath]",
		Short: "Attaches a sketch to a board.",
		Long:  "Attaches a sketch to a board.",
		Example: "  " + cli.VersionInfo.Application + " board attach serial:///dev/tty/ACM0\n" +
			"  " + cli.VersionInfo.Application + " board attach serial:///dev/tty/ACM0 HelloWorld\n" +
			"  " + cli.VersionInfo.Application + " board attach arduino:samd:mkr1000",
		Args: cobra.RangeArgs(1, 2),
		Run:  runAttachCommand,
	}
	attachCommand.Flags().StringVar(&attachFlags.searchTimeout, "timeout", "5s",
		"The timeout of the search of connected devices, try to high it if your board is not found (e.g. to 10s).")
	return attachCommand
}

var attachFlags struct {
	searchTimeout string // Expressed in a parsable duration, is the timeout for the list and attach commands.
}

func runAttachCommand(cmd *cobra.Command, args []string) {
	instance := cli.CreateInstance()
	var path string
	if len(args) > 0 {
		path = args[1]
	}
	_, err := board.Attach(context.Background(), &rpc.BoardAttachReq{
		Instance:      instance,
		BoardUri:      args[0],
		SketchPath:    path,
		SearchTimeout: attachFlags.searchTimeout,
	}, cli.OutputTaskProgress())
	if err != nil {
		formatter.PrintError(err, "attach board error")
		os.Exit(cli.ErrGeneric)
	}
}
