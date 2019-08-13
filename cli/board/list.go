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
	"os"
	"sort"
	"time"

	"github.com/arduino/arduino-cli/cli/errorcodes"
	"github.com/arduino/arduino-cli/cli/feedback"
	"github.com/arduino/arduino-cli/cli/instance"
	"github.com/arduino/arduino-cli/cli/output"
	"github.com/arduino/arduino-cli/commands/board"
	rpc "github.com/arduino/arduino-cli/rpc/commands"
	"github.com/cheynewallace/tabby"
	"github.com/spf13/cobra"
)

func initListCommand() *cobra.Command {
	listCommand := &cobra.Command{
		Use:     "list",
		Short:   "List connected boards.",
		Long:    "Detects and displays a list of connected boards to the current computer.",
		Example: "  " + os.Args[0] + " board list --timeout 10s",
		Args:    cobra.NoArgs,
		Run:     runListCommand,
	}

	listCommand.Flags().StringVar(&listFlags.timeout, "timeout", "0s",
		"The timeout of the search of connected devices, try to increase it if your board is not found (e.g. to 10s).")
	return listCommand
}

var listFlags struct {
	timeout string // Expressed in a parsable duration, is the timeout for the list and attach commands.
}

// runListCommand detects and lists the connected arduino boards
func runListCommand(cmd *cobra.Command, args []string) {
	if timeout, err := time.ParseDuration(listFlags.timeout); err != nil {
		feedback.Errorf("Invalid timeout: %v", err)
		os.Exit(errorcodes.ErrBadArgument)
	} else {
		time.Sleep(timeout)
	}

	ports, err := board.List(instance.CreateInstance().GetId())
	if err != nil {
		feedback.Errorf("Error detecting boards: %v", err)
		os.Exit(errorcodes.ErrNetwork)
	}

	if output.JSONOrElse(ports) {
		outputListResp(ports)
	}
}

func outputListResp(ports []*rpc.DetectedPort) {
	if len(ports) == 0 {
		feedback.Print("No boards found.")
		return
	}
	sort.Slice(ports, func(i, j int) bool {
		x, y := ports[i], ports[j]
		return x.GetProtocol() < y.GetProtocol() ||
			(x.GetProtocol() == y.GetProtocol() && x.GetAddress() < y.GetAddress())
	})

	table := tabby.New()
	table.AddHeader("Port", "Type", "Board Name", "FQBN")
	for _, port := range ports {
		address := port.GetProtocol() + "://" + port.GetAddress()
		if port.GetProtocol() == "serial" {
			address = port.GetAddress()
		}
		protocol := port.GetProtocolLabel()
		if boards := port.GetBoards(); len(boards) > 0 {
			sort.Slice(boards, func(i, j int) bool {
				x, y := boards[i], boards[j]
				return x.GetName() < y.GetName() || (x.GetName() == y.GetName() && x.GetFQBN() < y.GetFQBN())
			})
			for _, b := range boards {
				board := b.GetName()
				fqbn := b.GetFQBN()
				table.AddLine(address, protocol, board, fqbn)
				// show address and protocol only on the first row
				address = ""
				protocol = ""
			}
		} else {
			board := "Unknown"
			fqbn := ""
			table.AddLine(address, protocol, board, fqbn)
		}
	}
	table.Print()
}
