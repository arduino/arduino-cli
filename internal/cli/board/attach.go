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

	"github.com/arduino/arduino-cli/internal/cli/arguments"
	"github.com/arduino/arduino-cli/internal/cli/feedback"
	"github.com/spf13/cobra"
)

func initAttachCommand() *cobra.Command {
	var port arguments.Port
	attachCommand := &cobra.Command{
		Use:   fmt.Sprintf("attach [-p <%s>] [-b <%s>] [%s]", tr("port"), tr("FQBN"), tr("sketchPath")),
		Short: tr("Attaches a sketch to a board."),
		Long:  tr("Sets the default values for port and FQBN. If no port or FQBN are specified, the current default port and FQBN are displayed."),
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
	sk := arguments.MustNewSketch(sketchPath)

	var currentPort *boardAttachPortResult
	if currentAddress, currentProtocol := sk.GetDefaultPortAddressAndProtocol(); currentAddress != "" {
		currentPort = &boardAttachPortResult{
			Address:  currentAddress,
			Protocol: currentProtocol,
		}
	}
	current := &boardAttachResult{
		Port: currentPort,
		Fqbn: sk.GetDefaultFQBN(),
	}
	address, protocol, _ := port.GetPortAddressAndProtocol(nil, sk)
	if address != "" {
		if err := sk.SetDefaultPort(address, protocol); err != nil {
			feedback.Fatal(fmt.Sprintf("%s: %s", tr("Error saving sketch metadata"), err), feedback.ErrGeneric)
		}
		current.Port = &boardAttachPortResult{
			Address:  address,
			Protocol: protocol,
		}
	}
	if fqbn != "" {
		if err := sk.SetDefaultFQBN(fqbn); err != nil {
			feedback.Fatal(fmt.Sprintf("%s: %s", tr("Error saving sketch metadata"), err), feedback.ErrGeneric)
		}
		current.Fqbn = fqbn
	}

	feedback.PrintResult(current)
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
	Fqbn string                 `json:"fqbn,omitempty"`
	Port *boardAttachPortResult `json:"port,omitempty"`
}

func (b *boardAttachResult) Data() interface{} {
	return b
}

func (b *boardAttachResult) String() string {
	if b.Port == nil && b.Fqbn == "" {
		return tr("No default port or FQBN set")
	}
	res := fmt.Sprintf("%s: %s\n", tr("Default port set to"), b.Port)
	res += fmt.Sprintf("%s: %s\n", tr("Default FQBN set to"), b.Fqbn)
	return res
}
