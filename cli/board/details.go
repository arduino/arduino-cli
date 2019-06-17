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
	"fmt"
	"os"

	"github.com/arduino/arduino-cli/cli"
	"github.com/arduino/arduino-cli/commands/board"
	"github.com/arduino/arduino-cli/common/formatter"
	"github.com/arduino/arduino-cli/output"
	rpc "github.com/arduino/arduino-cli/rpc/commands"
	"github.com/spf13/cobra"
)

func initDetailsCommand() *cobra.Command {
	detailsCommand := &cobra.Command{
		Use:     "details <FQBN>",
		Short:   "Print details about a board.",
		Long:    "Show information about a board, in particular if the board has options to be specified in the FQBN.",
		Example: "  " + cli.VersionInfo.Application + " board details arduino:avr:nano",
		Args:    cobra.ExactArgs(1),
		Run:     runDetailsCommand,
	}
	return detailsCommand
}

func runDetailsCommand(cmd *cobra.Command, args []string) {
	instance := cli.CreateInstance()

	res, err := board.Details(context.Background(), &rpc.BoardDetailsReq{
		Instance: instance,
		Fqbn:     args[0],
	})

	if err != nil {
		formatter.PrintError(err, "Error getting board details")
		os.Exit(cli.ErrGeneric)
	}
	if cli.OutputJSONOrElse(res) {
		outputDetailsResp(res)
	}
}

func outputDetailsResp(details *rpc.BoardDetailsResp) {
	table := output.NewTable()
	table.SetColumnWidthMode(1, output.Average)
	table.AddRow("Board name:", details.Name)
	for i, tool := range details.RequiredTools {
		head := ""
		if i == 0 {
			table.AddRow()
			head = "Required tools:"
		}
		table.AddRow(head, tool.Packager+":"+tool.Name, "", tool.Version)
	}
	for _, option := range details.ConfigOptions {
		table.AddRow()
		table.AddRow("Option:",
			option.OptionLabel,
			"", option.Option)
		for _, value := range option.Values {
			if value.Selected {
				table.AddRow("",
					output.Green(value.ValueLabel),
					output.Green("✔"), output.Green(option.Option+"="+value.Value))
			} else {
				table.AddRow("",
					value.ValueLabel,
					"", option.Option+"="+value.Value)
			}
		}
	}
	fmt.Print(table.Render())
}
