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
	"github.com/arduino/arduino-cli/cli/globals"
	"github.com/arduino/arduino-cli/cli/instance"
	"github.com/arduino/arduino-cli/cli/output"
	"github.com/arduino/arduino-cli/commands/board"
	rpc "github.com/arduino/arduino-cli/rpc/commands"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func initInstallCommand() *cobra.Command {
	installCommand := &cobra.Command{
		Use:   "install <FQBN | BOARD_ALIAS>",
		Short: "Installs SDK for a board.",
		Long: "Installs the platform and tools to support build and upload for the specified board.\n" +
			"The board may be specified via FQBN or board alias.",
		Example: "  # install the SDK for the Arduino Zero.\n" +
			"  " + os.Args[0] + " board install zero\n\n" +
			"  # install the SDK for the Arduino UNO using the FQBN as parameter.\n" +
			"  " + os.Args[0] + " board install arduino:avr:uno",
		Args: cobra.ExactArgs(1),
		Run:  runInstallCommand,
	}
	return installCommand
}

func runInstallCommand(cmd *cobra.Command, args []string) {
	inst, err := instance.CreateInstance()
	if err != nil {
		feedback.Errorf("Error installing: %v", err)
		os.Exit(errorcodes.ErrGeneric)
	}

	boardArg := args[1]
	logrus.WithField("board", boardArg).Info("Executing `board install`")

	boardInstallReq := &rpc.BoardInstallReq{
		Instance: inst,
		Board:    boardArg,
	}
	_, err = board.Install(context.Background(), boardInstallReq, output.ProgressBar(),
		output.TaskProgress(), globals.NewHTTPClientHeader())
	if err != nil {
		feedback.Errorf("Error installing board: %v", err)
		os.Exit(errorcodes.ErrGeneric)
	}
}
