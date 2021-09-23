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
	"errors"
	"fmt"
	"io"
	"os"
	"sort"

	pluggableMonitor "github.com/arduino/arduino-cli/arduino/monitor"
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

	if describe {
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
		return
	}

	tty, err := newNullTerminal()
	if err != nil {
		feedback.Error(err)
		os.Exit(errorcodes.ErrGeneric)
	}
	defer tty.Close()

	portProxy, descriptor, err := monitor.Monitor(context.Background(), &rpc.MonitorRequest{
		Instance: instance,
		Port:     port.ToRPC(),
		Fqbn:     "",
	})
	if err != nil {
		feedback.Error(err)
		os.Exit(errorcodes.ErrGeneric)
	}
	defer portProxy.Close()

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		_, err := io.Copy(tty, portProxy)
		if err != nil && !errors.Is(err, io.EOF) {
			feedback.Error(tr("Port closed:"), err)
		}
		cancel()
	}()
	go func() {
		_, err := io.Copy(portProxy, tty)
		if err != nil && !errors.Is(err, io.EOF) {
			feedback.Error(tr("Port closed:"), err)
		}
		cancel()
	}()

	params := descriptor.ConfigurationParameters
	paramsKeys := sort.StringSlice{}
	for k := range params {
		paramsKeys = append(paramsKeys, k)
	}
	sort.Sort(paramsKeys)
	state := 0
	var setting *pluggableMonitor.PortParameterDescriptor
	settingKey := ""
	feedback.Print(tr("Connected to %s! Press CTRL-C to exit.", port.String()))

	// TODO: This is a work in progress...
	tty.AddEscapeCallback(func(r rune) bool {
		switch state {
		case 0:
			fmt.Println()
			fmt.Println("Commands available:")
			fmt.Println("  CTRL+C - Quit monitor")
			for i, key := range paramsKeys {
				fmt.Printf("  %c - Change %s\n", 'a'+i, params[key].Label)
			}
			fmt.Println("ESC - back to terminal...")
			fmt.Print("> ")
			state = 1

		case 1:
			if r >= 'a' && r <= rune('a'+len(paramsKeys)) {
				settingKey = paramsKeys[int(r-'a')]
				setting = params[settingKey]
				fmt.Printf("Chaging option %s, please select:\n", setting.Label)
				for i, v := range setting.Values {
					fmt.Printf("  %c - Select %s\n", 'a'+i, v)
				}
				fmt.Print("> ")
				state = 2
				return true
			}
			switch r {
			case 27:
				fmt.Println("ESC")
				state = 0
				return false
			default:
				fmt.Println("Invalid command... back to terminal")
				state = 0
				return false
			}

		case 2:
			if r >= 'a' && r <= rune('a'+len(setting.Values)) {
				settingValue := setting.Values[int(r-'a')]
				fmt.Printf("Selected %s <= %s\n", setting.Label, settingValue)
				if err := portProxy.Config(settingKey, settingValue); err != nil {
					fmt.Println("Error setting configuration:", err)
				}
			} else {
				fmt.Println("Invalid command... back to terminal")
			}
			state = 0
			return false
		}
		return true
	})

	// Wait for port closed
	<-ctx.Done()
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
