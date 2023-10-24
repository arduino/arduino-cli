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
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/arduino/arduino-cli/commands/monitor"
	sk "github.com/arduino/arduino-cli/commands/sketch"
	"github.com/arduino/arduino-cli/configuration"
	"github.com/arduino/arduino-cli/i18n"
	"github.com/arduino/arduino-cli/internal/cli/arguments"
	"github.com/arduino/arduino-cli/internal/cli/feedback"
	"github.com/arduino/arduino-cli/internal/cli/instance"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	"github.com/arduino/arduino-cli/table"
	"github.com/fatih/color"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"go.bug.st/cleanup"
)

var tr = i18n.Tr

// NewCommand created a new `monitor` command
func NewCommand() *cobra.Command {
	var (
		portArgs   arguments.Port
		fqbnArg    arguments.Fqbn
		profileArg arguments.Profile
		raw        bool
		describe   bool
		configs    []string
		quiet      bool
		timestamp  bool
	)
	monitorCommand := &cobra.Command{
		Use:   "monitor",
		Short: tr("Open a communication port with a board."),
		Long:  tr("Open a communication port with a board."),
		Example: "" +
			"  " + os.Args[0] + " monitor -p /dev/ttyACM0\n" +
			"  " + os.Args[0] + " monitor -p /dev/ttyACM0 --describe",
		Run: func(cmd *cobra.Command, args []string) {
			sketchPath := ""
			if len(args) > 0 {
				sketchPath = args[0]
			}
			runMonitorCmd(&portArgs, &fqbnArg, &profileArg, sketchPath, configs, describe, timestamp, quiet, raw)
		},
	}
	portArgs.AddToCommand(monitorCommand)
	profileArg.AddToCommand(monitorCommand)
	monitorCommand.Flags().BoolVar(&raw, "raw", false, tr("Set terminal in raw mode (unbuffered)."))
	monitorCommand.Flags().BoolVar(&describe, "describe", false, tr("Show all the settings of the communication port."))
	monitorCommand.Flags().StringSliceVarP(&configs, "config", "c", []string{}, tr("Configure communication port settings. The format is <ID>=<value>[,<ID>=<value>]..."))
	monitorCommand.Flags().BoolVarP(&quiet, "quiet", "q", false, tr("Run in silent mode, show only monitor input and output."))
	monitorCommand.Flags().BoolVar(&timestamp, "timestamp", false, tr("Timestamp each incoming line."))
	fqbnArg.AddToCommand(monitorCommand)
	return monitorCommand
}

func runMonitorCmd(
	portArgs *arguments.Port, fqbnArg *arguments.Fqbn, profileArg *arguments.Profile, sketchPathArg string,
	configs []string, describe, timestamp, quiet, raw bool,
) {
	logrus.Info("Executing `arduino-cli monitor`")

	if !configuration.HasConsole {
		quiet = true
	}

	var (
		inst                         *rpc.Instance
		profile                      *rpc.Profile
		fqbn                         string
		defaultPort, defaultProtocol string
	)

	sketchPath := arguments.InitSketchPath(sketchPathArg)
	sketch, err := sk.LoadSketch(context.Background(), &rpc.LoadSketchRequest{SketchPath: sketchPath.String()})
	if err != nil && !portArgs.IsPortFlagSet() {
		feedback.Fatal(
			tr("Error getting default port from `sketch.yaml`. Check if you're in the correct sketch folder or provide the --port flag: %s", err),
			feedback.ErrGeneric,
		)
	}
	if sketch != nil {
		defaultPort, defaultProtocol = sketch.GetDefaultPort(), sketch.GetDefaultProtocol()
		if profileArg.Get() == "" {
			inst, profile = instance.CreateAndInitWithProfile(sketch.GetDefaultProfile().GetName(), sketchPath)
		} else {
			inst, profile = instance.CreateAndInitWithProfile(profileArg.Get(), sketchPath)
		}
	}
	if inst == nil {
		inst = instance.CreateAndInit()
	}
	// Priority on how to retrieve the fqbn
	// 1. from flag
	// 2. from profile
	// 3. from default_fqbn specified in the sketch.yaml
	// 4. try to detect from the port
	switch {
	case fqbnArg.String() != "":
		fqbn = fqbnArg.String()
	case profile.GetFqbn() != "":
		fqbn = profile.GetFqbn()
	case sketch.GetDefaultFqbn() != "":
		fqbn = sketch.GetDefaultFqbn()
	default:
		fqbn, _ = portArgs.DetectFQBN(inst)
	}

	portAddress, portProtocol, err := portArgs.GetPortAddressAndProtocol(inst, defaultPort, defaultProtocol)
	if err != nil {
		feedback.FatalError(err, feedback.ErrGeneric)
	}

	enumerateResp, err := monitor.EnumerateMonitorPortSettings(context.Background(), &rpc.EnumerateMonitorPortSettingsRequest{
		Instance:     inst,
		PortProtocol: portProtocol,
		Fqbn:         fqbn,
	})
	if err != nil {
		feedback.Fatal(tr("Error getting port settings details: %s", err), feedback.ErrGeneric)
	}
	if describe {
		feedback.PrintResult(&detailsResult{Settings: enumerateResp.Settings})
		return
	}

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
							feedback.Fatal(tr("invalid port configuration value for %s: %s", k, v), feedback.ErrBadArgument)
						}
						setting = s
						break
					}
				}
			}
			if setting == nil {
				feedback.Fatal(tr("invalid port configuration: %s", config), feedback.ErrBadArgument)
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
		Instance:          inst,
		Port:              &rpc.Port{Address: portAddress, Protocol: portProtocol},
		Fqbn:              fqbn,
		PortConfiguration: configuration,
	})
	if err != nil {
		feedback.FatalError(err, feedback.ErrGeneric)
	}
	defer portProxy.Close()

	if !quiet {
		feedback.Print(tr("Connected to %s! Press CTRL-C to exit.", portAddress))
	}

	ttyIn, ttyOut, err := feedback.InteractiveStreams()
	if err != nil {
		feedback.FatalError(err, feedback.ErrGeneric)
	}

	if timestamp {
		ttyOut = newTimeStampWriter(ttyOut)
	}

	ctx, cancel := cleanup.InterruptableContext(context.Background())
	if raw {
		if feedback.IsTerminal() {
			if err := feedback.SetRawModeStdin(); err != nil {
				feedback.Warning(tr("Error setting raw mode: %s", err.Error()))
			}
			defer feedback.RestoreModeStdin()
		}

		// In RAW mode CTRL-C is not converted into an Interrupt by
		// the terminal, we must intercept ASCII 3 (CTRL-C) on our own...
		ctrlCDetector := &charDetectorWriter{
			callback:     cancel,
			detectedChar: 3, // CTRL-C
		}
		ttyIn = io.TeeReader(ttyIn, ctrlCDetector)
	}

	go func() {
		_, err := io.Copy(ttyOut, portProxy)
		if err != nil && !errors.Is(err, io.EOF) {
			if !quiet {
				feedback.Print(tr("Port closed: %v", err))
			}
		}
		cancel()
	}()
	go func() {
		_, err := io.Copy(portProxy, ttyIn)
		if err != nil && !errors.Is(err, io.EOF) {
			if !quiet {
				feedback.Print(tr("Port closed: %v", err))
			}
		}
		cancel()
	}()

	// Wait for port closed
	<-ctx.Done()
}

type charDetectorWriter struct {
	callback     func()
	detectedChar byte
}

func (cd *charDetectorWriter) Write(buf []byte) (int, error) {
	if bytes.IndexByte(buf, cd.detectedChar) != -1 {
		cd.callback()
	}
	return len(buf), nil
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

type timeStampWriter struct {
	writer            io.Writer
	sendTimeStampNext bool
}

func newTimeStampWriter(writer io.Writer) *timeStampWriter {
	return &timeStampWriter{
		writer:            writer,
		sendTimeStampNext: true,
	}
}

func (t *timeStampWriter) Write(p []byte) (int, error) {
	written := 0
	for _, b := range p {
		if t.sendTimeStampNext {
			_, err := t.writer.Write([]byte(time.Now().Format("[2006-01-02 15:04:05] ")))
			if err != nil {
				return written, err
			}
			t.sendTimeStampNext = false
		}
		n, err := t.writer.Write([]byte{b})
		written += n
		if err != nil {
			return written, err
		}
		t.sendTimeStampNext = b == '\n'
	}
	return written, nil
}
