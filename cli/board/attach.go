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
	"fmt"
	"os"

	"github.com/arduino/arduino-cli/cli/arguments"
	"github.com/arduino/arduino-cli/cli/errorcodes"
	"github.com/arduino/arduino-cli/cli/feedback"
	"github.com/spf13/cobra"
)

func initAttachCommand() *cobra.Command {
	var port arguments.Port
	attachCommand := &cobra.Command{
		Use:   fmt.Sprintf("attach [-p <%s>] [-b <%s>] [%s]", tr("port"), tr("FQBN"), tr("sketchPath")),
		Short: tr("Attaches a sketch to a board."),
		Long:  tr("Attaches a sketch to a board."),
		Example: "  " + os.Args[0] + " board attach -p /dev/ttyACM0\n" +
			"  " + os.Args[0] + " board attach -p /dev/ttyACM0 HelloWorld\n" +
			"  " + os.Args[0] + " board attach -b arduino:samd:mkr1000",
		Args: cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			sketchPath := ""
			if len(args) > 0 {
				sketchPath = args[0]
			}
			runAttachCommand(sketchPath, &port, fqbn.String())
		},
	}
	fqbn.AddToCommand(attachCommand)
	port.AddToCommand(attachCommand)

	return attachCommand
}

func runAttachCommand(path string, port *arguments.Port, fqbn string) {
	sketchPath := arguments.InitSketchPath(path)
	sk := arguments.NewSketch(sketchPath)

	address, protocol, _ := port.GetPortAddressAndProtocol(nil, sk)
	if address != "" {
		if err := sk.SetDefaultPort(address, protocol); err != nil {
			feedback.Errorf("%s: %s", tr("Error saving sketch metadata"), err)
			os.Exit(errorcodes.ErrGeneric)
		}
		msg := fmt.Sprintf("%s: %s", tr("Default port set to"), address)
		if protocol != "" {
			msg += " (" + protocol + ")"
		}
		feedback.Print(msg)
	}
	if fqbn != "" {
		if err := sk.SetDefaultFQBN(fqbn); err != nil {
			feedback.Errorf("%s: %s", tr("Error saving sketch metadata"), err)
			os.Exit(errorcodes.ErrGeneric)
		}
		msg := fmt.Sprintf("%s: %s", tr("Default FQBN set to"), fqbn)
		feedback.Print(msg)
	}
}
