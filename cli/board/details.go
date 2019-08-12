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

	"github.com/arduino/arduino-cli/cli/errorcodes"
	"github.com/arduino/arduino-cli/cli/instance"
	"github.com/arduino/arduino-cli/cli/output"
	"github.com/arduino/arduino-cli/commands/board"
	"github.com/arduino/arduino-cli/common/formatter"
	rpc "github.com/arduino/arduino-cli/rpc/commands"
	"github.com/cheynewallace/tabby"
	"github.com/spf13/cobra"
)

var detailsCommand = &cobra.Command{
	Use:     "details <FQBN>",
	Short:   "Print details about a board.",
	Long:    "Show information about a board, in particular if the board has options to be specified in the FQBN.",
	Example: "  " + os.Args[0] + " board details arduino:avr:nano",
	Args:    cobra.ExactArgs(1),
	Run:     runDetailsCommand,
}

func runDetailsCommand(cmd *cobra.Command, args []string) {
	res, err := board.Details(context.Background(), &rpc.BoardDetailsReq{
		Instance: instance.CreateInstance(),
		Fqbn:     args[0],
	})

	if err != nil {
		formatter.PrintError(err, "Error getting board details")
		os.Exit(errorcodes.ErrGeneric)
	}
	if output.JSONOrElse(res) {
		outputDetailsResp(res)
	}
}

func outputDetailsResp(details *rpc.BoardDetailsResp) {
	// Table is 5 columns wide:
	// |               |                             | |                       |
	// Board name:     Arduino Nano
	//
	// Required tools: arduino:avr-gcc                 5.4.0-atmel3.6.1-arduino2
	//                 arduino:avrdude                 6.3.0-arduino14
	//                 arduino:arduinoOTA              1.2.1
	//
	// Option:         Processor                       cpu
	//                 ATmega328P                    * cpu=atmega328
	//                 ATmega328P (Old Bootloader)     cpu=atmega328old
	//                 ATmega168                       cpu=atmega168
	table := tabby.New()

	table.AddLine("Board name:", details.Name, "")

	for i, tool := range details.RequiredTools {
		if i == 0 {
			table.AddLine("", "", "", "") // get some space from above
			table.AddLine("Required tools:", tool.Packager+":"+tool.Name, "", tool.Version)
			continue
		}
		table.AddLine("", tool.Packager+":"+tool.Name, "", tool.Version)
	}

	for _, option := range details.ConfigOptions {
		table.AddLine("", "", "", "") // get some space from above
		table.AddLine("Option:", option.OptionLabel, "", option.Option)
		for _, value := range option.Values {
			checked := ""
			if value.Selected {
				checked = "✔"
			}
			table.AddLine("", value.ValueLabel, checked, option.Option+"="+value.Value)
		}
	}
	table.Print()
}
