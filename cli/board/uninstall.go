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

	"github.com/arduino/arduino-cli/cli/errorcodes"
	"github.com/arduino/arduino-cli/cli/feedback"
	"github.com/arduino/arduino-cli/cli/instance"
	"github.com/arduino/arduino-cli/cli/output"
	"github.com/arduino/arduino-cli/commands/board"
	rpc "github.com/arduino/arduino-cli/rpc/commands"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func initUninstallCommand() *cobra.Command {
	installCommand := &cobra.Command{
		Use:   "uninstall <FQBN | BOARD_ALIAS>",
		Short: "Uninstalls SDK for a board.",
		Long: "Uninstalls the platform and tools to support build and upload for the specified board.\n" +
			"The board may be specified via FQBN or board alias.",
		Example: "  # uninstall the SDK for the Arduino Zero.\n" +
			"  " + os.Args[0] + " board uninstall zero\n\n" +
			"  # uninstall the SDK for the Arduino UNO using the FQBN as parameter.\n" +
			"  " + os.Args[0] + " board uninstall arduino:avr:uno",
		Args: cobra.ExactArgs(1),
		Run:  runUninstallCommand,
	}
	return installCommand
}

func runUninstallCommand(cmd *cobra.Command, args []string) {
	inst, err := instance.CreateInstance()
	if err != nil {
		feedback.Errorf("Error uninstalling: %v", err)
		os.Exit(errorcodes.ErrGeneric)
	}

	boardArg := args[1]
	logrus.WithField("board", boardArg).Info("Executing `board uninstall`")

	boardUninstallReq := &rpc.BoardUninstallReq{
		Instance: inst,
		Board:    boardArg,
	}
	_, err = board.Uninstall(context.Background(), boardUninstallReq, output.TaskProgress())
	if err != nil {
		feedback.Errorf("Error uninstalling board: %v", err)
		os.Exit(errorcodes.ErrGeneric)
	}
}
