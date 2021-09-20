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

package monitor

import (
	"context"
	"os"

	"github.com/arduino/arduino-cli/cli/arguments"
	"github.com/arduino/arduino-cli/cli/errorcodes"
	"github.com/arduino/arduino-cli/cli/feedback"
	"github.com/arduino/arduino-cli/cli/instance"
	"github.com/arduino/arduino-cli/commands/monitor"
	"github.com/arduino/arduino-cli/i18n"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	"github.com/arduino/arduino-cli/table"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var tr = i18n.Tr

var portArgs arguments.Port
var describe bool

// NewCommand created a new `monitor` command
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "monitor",
		Short: tr("Open a communication port with a board."),
		Long:  tr("Open a communication port with a board."),
		Example: "" +
			"  " + os.Args[0] + " monitor -p /dev/ttyACM0\n" +
			"  " + os.Args[0] + " monitor -p /dev/ttyACM0 --describe",
		Run: runMonitorCmd,
	}
	portArgs.AddToCommand(cmd)
	cmd.Flags().BoolVar(&describe, "describe", false, tr("Show all the settings of the communication port."))
	cmd.MarkFlagRequired("port")
	return cmd
}

func runMonitorCmd(cmd *cobra.Command, args []string) {
	instance := instance.CreateAndInit()

	port, err := portArgs.GetPort(instance, nil)
	if err != nil {
		if err != nil {
			feedback.Error(err)
			os.Exit(errorcodes.ErrGeneric)
		}
	}

	res, err := monitor.EnumerateMonitorPortSettings(context.Background(), &rpc.EnumerateMonitorPortSettingsRequest{
		Instance: instance,
		Port:     port.ToRPC(),
		Fqbn:     "",
	})
	if err != nil {
		feedback.Error(tr("Error getting port settings details: %s"), err)
		os.Exit(errorcodes.ErrGeneric)
	}
	feedback.PrintResult(&detailsResult{Settings: res.Settings})
}

type detailsResult struct {
	Settings []*rpc.MonitorPortSettingDescriptor `json:"settings"`
}

func (r *detailsResult) Data() interface{} {
	return r
}

func (r *detailsResult) String() string {
	t := table.New()
	green := color.New(color.FgGreen)
	t.SetHeader(tr("ID"), tr("Setting"), tr("Type"), "", tr("Values"))
	for _, setting := range r.Settings {
		for i, v := range setting.EnumValues {
			selected := table.NewCell("", nil)
			value := table.NewCell(v, nil)
			if v == setting.Value {
				selected = table.NewCell("✔", green)
				value = table.NewCell(v, green)
			}
			if i == 0 {
				t.AddRow(setting.SettingId, setting.Label, setting.Type, selected, value)
			} else {
				t.AddRow("", "", "", selected, value)
			}
		}
	}
	return t.Render()
}
