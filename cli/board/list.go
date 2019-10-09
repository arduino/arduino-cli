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
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/arduino/arduino-cli/arduino/cores"
	"github.com/arduino/arduino-cli/cli/errorcodes"
	"github.com/arduino/arduino-cli/cli/feedback"
	"github.com/arduino/arduino-cli/cli/instance"
	"github.com/arduino/arduino-cli/commands/board"
	rpc "github.com/arduino/arduino-cli/rpc/commands"
	"github.com/arduino/arduino-cli/table"
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

	feedback.PrintResult(result{ports})
}

// output from this command requires special formatting, let's create a dedicated
// feedback.Result implementation
type result struct {
	ports []*rpc.DetectedPort
}

func (dr result) Data() interface{} {
	return dr.ports
}

func (dr result) String() string {
	if len(dr.ports) == 0 {
		return "No boards found."
	}

	sort.Slice(dr.ports, func(i, j int) bool {
		x, y := dr.ports[i], dr.ports[j]
		return x.GetProtocol() < y.GetProtocol() ||
			(x.GetProtocol() == y.GetProtocol() && x.GetAddress() < y.GetAddress())
	})

	t := table.New()
	t.SetHeader("Port", "Type", "Board Name", "FQBN", "Core")
	for _, port := range dr.ports {
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

				// to improve the user experience, show on a dedicated column
				// the name of the core supporting the board detected
				var coreName = ""
				fqbn, err := cores.ParseFQBN(b.GetFQBN())
				if err == nil {
					coreName = fmt.Sprintf("%s:%s", fqbn.Package, fqbn.PlatformArch)
				}

				t.AddRow(address, protocol, board, fqbn, coreName)

				// reset address and protocol, we only show them on the first row
				address = ""
				protocol = ""
			}
		} else {
			board := "Unknown"
			fqbn := ""
			coreName := ""
			t.AddRow(address, protocol, board, fqbn, coreName)
		}
	}
	return t.Render()
}
