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

	"github.com/arduino/arduino-cli/commands"
	"github.com/arduino/arduino-cli/internal/cli/arguments"
	"github.com/arduino/arduino-cli/internal/cli/feedback"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	"github.com/spf13/cobra"
)

func initAttachCommand() *cobra.Command {
	var port arguments.Port
	var fqbn arguments.Fqbn
	var programmer arguments.Programmer
	attachCommand := &cobra.Command{
		Use:   fmt.Sprintf("attach [-p <%s>] [-b <%s>] [-P <%s>] [%s]", tr("port"), tr("FQBN"), tr("programmer"), tr("sketchPath")),
		Short: tr("Attaches a sketch to a board."),
		Long:  tr("Sets the default values for port and FQBN. If no port, FQBN or programmer are specified, the current default port, FQBN and programmer are displayed."),
		Example: "  " + os.Args[0] + " board attach -p /dev/ttyACM0\n" +
			"  " + os.Args[0] + " board attach -p /dev/ttyACM0 HelloWorld\n" +
			"  " + os.Args[0] + " board attach -b arduino:samd:mkr1000" +
			"  " + os.Args[0] + " board attach -P atmel_ice",
		Args: cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			sketchPath := ""
			if len(args) > 0 {
				sketchPath = args[0]
			}
			runAttachCommand(sketchPath, &port, fqbn.String(), &programmer)
		},
	}
	fqbn.AddToCommand(attachCommand)
	port.AddToCommand(attachCommand)
	programmer.AddToCommand(attachCommand)

	return attachCommand
}

func runAttachCommand(path string, port *arguments.Port, fqbn string, programmer *arguments.Programmer) {
	sketchPath := arguments.InitSketchPath(path)

	portAddress, portProtocol, _ := port.GetPortAddressAndProtocol(nil, "", "")
	newDefaults, err := commands.SetSketchDefaults(context.Background(), &rpc.SetSketchDefaultsRequest{
		SketchPath:          sketchPath.String(),
		DefaultFqbn:         fqbn,
		DefaultProgrammer:   programmer.GetProgrammer(),
		DefaultPortAddress:  portAddress,
		DefaultPortProtocol: portProtocol,
	})
	if err != nil {
		feedback.FatalError(err, feedback.ErrGeneric)
	}

	res := &boardAttachResult{
		Fqbn:       newDefaults.GetDefaultFqbn(),
		Programmer: newDefaults.GetDefaultProgrammer(),
	}
	if newDefaults.GetDefaultPortAddress() != "" {
		res.Port = &boardAttachPortResult{
			Address:  newDefaults.GetDefaultPortAddress(),
			Protocol: newDefaults.GetDefaultPortProtocol(),
		}
	}
	feedback.PrintResult(res)
}

type boardAttachPortResult struct {
	Address  string `json:"address,omitempty"`
	Protocol string `json:"protocol,omitempty"`
}

func (b *boardAttachPortResult) String() string {
	port := b.Address
	if b.Protocol != "" {
		port += " (" + b.Protocol + ")"
	}
	return port
}

type boardAttachResult struct {
	Fqbn       string                 `json:"fqbn,omitempty"`
	Programmer string                 `json:"programmer,omitempty"`
	Port       *boardAttachPortResult `json:"port,omitempty"`
}

func (b *boardAttachResult) Data() interface{} {
	return b
}

func (b *boardAttachResult) String() string {
	if b.Port == nil && b.Fqbn == "" && b.Programmer == "" {
		return tr("No default port, FQBN or programmer set")
	}
	res := fmt.Sprintf("%s: %s\n", tr("Default port set to"), b.Port)
	res += fmt.Sprintf("%s: %s\n", tr("Default FQBN set to"), b.Fqbn)
	res += fmt.Sprintf("%s: %s\n", tr("Default programmer set to"), b.Programmer)
	return res
}
