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
	"strings"

	"github.com/arduino/arduino-cli/cli/arguments"
	"github.com/arduino/arduino-cli/cli/errorcodes"
	"github.com/arduino/arduino-cli/cli/feedback"
	"github.com/arduino/arduino-cli/cli/instance"
	"github.com/arduino/arduino-cli/commands/monitor"
	"github.com/arduino/arduino-cli/configuration"
	"github.com/arduino/arduino-cli/i18n"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	"github.com/arduino/arduino-cli/table"
	"github.com/fatih/color"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	portArgs arguments.Port
	describe bool
	configs  []string
	quiet    bool
	fqbn     arguments.Fqbn
	tr       = i18n.Tr
)

// NewCommand created a new `monitor` command
func NewCommand() *cobra.Command {
	monitorCommand := &cobra.Command{
		Use:   "monitor",
		Short: tr("Open a communication port with a board."),
		Long:  tr("Open a communication port with a board."),
		Example: "" +
			"  " + os.Args[0] + " monitor -p /dev/ttyACM0\n" +
			"  " + os.Args[0] + " monitor -p /dev/ttyACM0 --describe",
		Run: runMonitorCmd,
	}
	portArgs.AddToCommand(monitorCommand)
	monitorCommand.Flags().BoolVar(&describe, "describe", false, tr("Show all the settings of the communication port."))
	monitorCommand.Flags().StringSliceVarP(&configs, "config", "c", []string{}, tr("Configuration of the port."))
	monitorCommand.Flags().BoolVarP(&quiet, "quiet", "q", false, tr("Run in silent mode, show only monitor input and output."))
	fqbn.AddToCommand(monitorCommand)
	monitorCommand.MarkFlagRequired("port")
	return monitorCommand
}

func runMonitorCmd(cmd *cobra.Command, args []string) {
	instance.Init()
	logrus.Info("Executing `arduino-cli monitor`")

	if !configuration.HasConsole {
		quiet = true
	}

	inst := instance.Get()
	portAddress, portProtocol, err := portArgs.GetPortAddressAndProtocol(inst.ToRPC(), nil)
	if err != nil {
		feedback.Error(err)
		os.Exit(errorcodes.ErrGeneric)
	}

	enumerateResp, err := monitor.EnumerateMonitorPortSettings(context.Background(), &rpc.EnumerateMonitorPortSettingsRequest{
		Instance:     inst.ToRPC(),
		PortProtocol: portProtocol,
		Fqbn:         fqbn.String(),
	})
	if err != nil {
		feedback.Error(tr("Error getting port settings details: %s", err))
		os.Exit(errorcodes.ErrGeneric)
	}
	if describe {
		feedback.PrintResult(&detailsResult{Settings: enumerateResp.Settings})
		return
	}

	tty, err := newStdInOutTerminal()
	if err != nil {
		feedback.Error(err)
		os.Exit(errorcodes.ErrGeneric)
	}
	defer tty.Close()

	configuration := &rpc.MonitorPortConfiguration{}
	if len(configs) > 0 {
		for _, config := range configs {
			split := strings.SplitN(config, "=", 2)
			k := ""
			v := config
			if len(split) == 2 {
				k = split[0]
				v = split[1]
			}

			var setting *rpc.MonitorPortSettingDescriptor
			for _, s := range enumerateResp.GetSettings() {
				if k == "" {
					if contains(s.EnumValues, v) {
						setting = s
						break
					}
				} else {
					if strings.EqualFold(s.SettingId, k) {
						if !contains(s.EnumValues, v) {
							feedback.Error(tr("invalid port configuration value for %s: %s", k, v))
							os.Exit(errorcodes.ErrBadArgument)
						}
						setting = s
						break
					}
				}
			}
			if setting == nil {
				feedback.Error(tr("invalid port configuration: %s", config))
				os.Exit(errorcodes.ErrBadArgument)
			}
			configuration.Settings = append(configuration.Settings, &rpc.MonitorPortSetting{
				SettingId: setting.SettingId,
				Value:     v,
			})
			if !quiet {
				feedback.Print(tr("Monitor port settings:"))
				feedback.Print(fmt.Sprintf("%s=%s", setting.SettingId, v))
			}
		}
	}
	portProxy, _, err := monitor.Monitor(context.Background(), &rpc.MonitorRequest{
		Instance:          inst.ToRPC(),
		Port:              &rpc.Port{Address: portAddress, Protocol: portProtocol},
		Fqbn:              fqbn.String(),
		PortConfiguration: configuration,
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

	if !quiet {
		feedback.Print(tr("Connected to %s! Press CTRL-C to exit.", portAddress))
	}

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
	t.SetHeader(tr("ID"), tr("Setting"), tr("Default"), tr("Values"))
	sort.Slice(r.Settings, func(i, j int) bool {
		return r.Settings[i].Label < r.Settings[j].Label
	})
	for _, setting := range r.Settings {
		values := strings.Join(setting.EnumValues, ", ")
		t.AddRow(setting.SettingId, setting.Label, table.NewCell(setting.Value, green), values)
	}
	return t.Render()
}

func contains(s []string, searchterm string) bool {
	for _, item := range s {
		if strings.EqualFold(item, searchterm) {
			return true
		}
	}
	return false
}
