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
	"os"

	"github.com/arduino/arduino-cli/output"

	"github.com/arduino/arduino-cli/commands"
	"github.com/arduino/arduino-cli/common/formatter"
	"github.com/spf13/cobra"
)

func initDetailsCommand() *cobra.Command {
	detailsCommand := &cobra.Command{
		Use:     "details <FQBN>",
		Short:   "Print details about a board.",
		Long:    "Show information about a board, in particular if the board has options to be specified in the FQBN.",
		Example: "  " + commands.AppName + " board details arduino:avr:nano",
		Args:    cobra.ExactArgs(1),
		Run:     runDetailsCommand,
	}
	return detailsCommand
}

func runDetailsCommand(cmd *cobra.Command, args []string) {
	pm := commands.InitPackageManager()
	fqbnIn := args[0]
	board, err := pm.FindBoardWithFQBN(fqbnIn)
	if err != nil {
		formatter.PrintError(err, "Error loading board data")
		os.Exit(commands.ErrBadArgument)
	}

	details := &boardDetails{}
	details.Name = board.Name()
	details.ConfigOptions = []*boardConfigOption{}
	options := board.GetConfigOptions()
	for _, option := range options.Keys() {
		configOption := &boardConfigOption{}
		configOption.Option = option
		configOption.OptionLabel = options.Get(option)
		values := board.GetConfigOptionValues(option)
		for i, value := range values.Keys() {
			configValue := &boardConfigValue{}
			if i == 0 {
				t := true
				configValue.Default = &t
			}
			configValue.Value = value
			configValue.ValueLabel = values.Get(value)
			configOption.Values = append(configOption.Values, configValue)
		}

		details.ConfigOptions = append(details.ConfigOptions, configOption)
	}

	output.Emit(details)
}

type boardDetails struct {
	Name          string
	ConfigOptions []*boardConfigOption
}

type boardConfigOption struct {
	Option      string
	OptionLabel string
	Values      []*boardConfigValue
}

type boardConfigValue struct {
	Value      string
	ValueLabel string
	Default    *bool `json:",omitempty"`
}

func (details *boardDetails) EmitJSON() string {
	d, err := json.MarshalIndent(details, "", "  ")
	if err != nil {
		formatter.PrintError(err, "Error encoding json")
		os.Exit(commands.ErrGeneric)
	}
	return string(d)
}

func (details *boardDetails) EmitTerminal() string {
	table := output.NewTable()
	table.AddRow("Board name:", output.Red(details.Name))
	table.SetColumnWidthMode(1, output.Average)
	for _, option := range details.ConfigOptions {
		table.AddRow("Option:", option.OptionLabel)
		for _, value := range option.Values {
			table.AddRow("", value.ValueLabel, option.Option+"="+value.Value)
		}
	}
	return table.Render()
}
