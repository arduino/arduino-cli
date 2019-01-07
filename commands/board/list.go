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
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/arduino/arduino-cli/output"

	"github.com/arduino/arduino-cli/arduino/discovery"
	"github.com/arduino/arduino-cli/commands"
	"github.com/arduino/arduino-cli/commands/core"
	"github.com/arduino/arduino-cli/common/formatter"
	"github.com/spf13/cobra"
)

func initListCommand() *cobra.Command {
	listCommand := &cobra.Command{
		Use:     "list",
		Short:   "List connected boards.",
		Long:    "Detects and displays a list of connected boards to the current computer.",
		Example: "  " + commands.AppName + " board list",
		Args:    cobra.NoArgs,
		Run:     runListCommand,
	}

	listCommand.Flags().StringVar(&listFlags.timeout, "timeout", "1s",
		"The timeout of the search of connected devices, try to increase it if your board is not found (e.g. to 10s).")
	return listCommand
}

var listFlags struct {
	timeout string // Expressed in a parsable duration, is the timeout for the list and attach commands.
}

// runListCommand detects and lists the connected arduino boards
func runListCommand(cmd *cobra.Command, args []string) {
	pm := commands.InitPackageManager()

	timeout, err := time.ParseDuration(listFlags.timeout)
	if err != nil {
		formatter.PrintError(err, "Invalid timeout.")
		os.Exit(commands.ErrBadArgument)
	}


	discoveries := discovery.ExtractDiscoveriesFromPlatforms(pm)

	res := []*detectedPort{}
	for discName, disc := range discoveries {
		disc.Timeout = timeout
		disc.Start()
		defer disc.Close()

		ports, err := disc.List()
		if err != nil {
			fmt.Printf("Error getting port list from discovery %s: %s\n", discName, err)
			continue
		}
		for _, port := range ports {
			b := detectedBoards{}
			for _, board := range pm.IdentifyBoard(port.IdentificationPrefs) {
				b = append(b, &detectedBoard{
					Name: board.Name(),
					FQBN: board.FQBN(),
				})
			}
			p := &detectedPort{
				Address:       port.Address,
				Protocol:      port.Protocol,
				ProtocolLabel: port.ProtocolLabel,
				Boards:        b,
			}
			res = append(res, p)
		}
	}
	output.Emit(&detectedPorts{
		Ports: res,
	})
}

type detectedPorts struct {
	Ports []*detectedPort `json:"ports"`
}

type detectedPort struct {
	Address       string         `json:"address"`
	Protocol      string         `json:"protocol"`
	ProtocolLabel string         `json:"protocol_label"`
	Boards        detectedBoards `json:"boards"`
}

type detectedBoards []*detectedBoard

type detectedBoard struct {
	Name string `json:"name"`
	FQBN string `json:"fqbn"`
}

func (b detectedBoards) Less(i, j int) bool {
	x := b[i]
	y := b[j]
	if x.Name < y.Name {
		return true
	}
	return x.FQBN < y.FQBN
}

func (p detectedPorts) Less(i, j int) bool {
	x := p.Ports[i]
	y := p.Ports[j]
	if x.Protocol < y.Protocol {
		return true
	}
	if x.Address < y.Address {
		return true
	}
	return false
}

func (p detectedPorts) EmitJSON() string {
	d, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		formatter.PrintError(err, "Error encoding json")
		os.Exit(commands.ErrGeneric)
	}
	return string(d)
}

func (p detectedPorts) EmitTerminal() string {
	sort.Slice(p.Ports, p.Less)
	table := output.NewTable()
	table.SetHeader("Port", "Type", "Board Name", "FQBN")
	for _, port := range p.Ports {
		address := port.Protocol + "://" + port.Address
		if port.Protocol == "serial" {
			address = port.Address
		}
		protocol := port.ProtocolLabel
		if len(port.Boards) > 0 {
			sort.Slice(port.Boards, port.Boards.Less)
			for _, b := range port.Boards {
				board := b.Name
				fqbn := b.FQBN
				table.AddRow(address, protocol, board, fqbn)
				// show address and protocol only on the first row
				address = ""
				protocol = ""
			}
		} else {
			board := "Unknown"
			fqbn := ""
			table.AddRow(address, protocol, board, fqbn)
		}
	}
	return table.Render()
}
