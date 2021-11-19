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
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	port arguments.Port
)

func initAttachCommand() *cobra.Command {
	attachCommand := &cobra.Command{
		Use:   fmt.Sprintf("attach -p <%s>|-b <%s> [%s]", tr("port"), tr("FQBN"), tr("sketchPath")),
		Short: tr("Attaches a sketch to a board."),
		Long:  tr("Attaches a sketch to a board."),
		Example: "  " + os.Args[0] + " board attach -p /dev/ttyACM0\n" +
			"  " + os.Args[0] + " board attach -p /dev/ttyACM0 HelloWorld\n" +
			"  " + os.Args[0] + " board attach -b arduino:samd:mkr1000",
		Args: cobra.MaximumNArgs(1),
		Run:  runAttachCommand,
	}
	fqbn.AddToCommand(attachCommand)
	port.AddToCommand(attachCommand)

	return attachCommand
}

func runAttachCommand(cmd *cobra.Command, args []string) {
	instance := instance.CreateAndInit()

	logrus.Info("Executing `arduino-cli board attach`")

	path := ""
	if len(args) > 0 {
		path = args[0]
	}
	sketchPath := arguments.InitSketchPath(path)

	// ugly hack to allow user to specify fqbn and port as flags (consistency)
	// a more meaningful fix would be to fix board.Attach
	var boardURI string
	discoveryPort, _ := port.GetPort(instance, nil)
	if fqbn.String() != "" {
		boardURI = fqbn.String()
	} else if discoveryPort != nil {
		boardURI = discoveryPort.Address
	}
	if _, err := board.Attach(context.Background(), &rpc.BoardAttachRequest{
		Instance:      instance,
		BoardUri:      boardURI,
		SketchPath:    sketchPath.String(),
		SearchTimeout: port.GetSearchTimeout().String(),
	}, output.TaskProgress()); err != nil {
		feedback.Errorf(tr("Attach board error: %v"), err)
		os.Exit(errorcodes.ErrGeneric)
	}
}
