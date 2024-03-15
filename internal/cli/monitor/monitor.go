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

	"github.com/arduino/arduino-cli/commands"
	"github.com/arduino/arduino-cli/internal/cli/arguments"
	"github.com/arduino/arduino-cli/internal/cli/feedback"
	"github.com/arduino/arduino-cli/internal/cli/feedback/result"
	"github.com/arduino/arduino-cli/internal/cli/feedback/table"
	"github.com/arduino/arduino-cli/internal/cli/instance"
	"github.com/arduino/arduino-cli/internal/i18n"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	"github.com/fatih/color"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"go.bug.st/cleanup"
)

var tr = i18n.Tr

// NewCommand created a new `monitor` command
func NewCommand(srv rpc.ArduinoCoreServiceServer) *cobra.Command {
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
			runMonitorCmd(srv, &portArgs, &fqbnArg, &profileArg, sketchPath, configs, describe, timestamp, quiet, raw)
		},
	}
	portArgs.AddToCommand(monitorCommand, srv)
	profileArg.AddToCommand(monitorCommand)
	monitorCommand.Flags().BoolVar(&raw, "raw", false, tr("Set terminal in raw mode (unbuffered)."))
	monitorCommand.Flags().BoolVar(&describe, "describe", false, tr("Show all the settings of the communication port."))
	monitorCommand.Flags().StringSliceVarP(&configs, "config", "c", []string{}, tr("Configure communication port settings. The format is <ID>=<value>[,<ID>=<value>]..."))
	monitorCommand.Flags().BoolVarP(&quiet, "quiet", "q", false, tr("Run in silent mode, show only monitor input and output."))
	monitorCommand.Flags().BoolVar(&timestamp, "timestamp", false, tr("Timestamp each incoming line."))
	fqbnArg.AddToCommand(monitorCommand, srv)
	return monitorCommand
}

func runMonitorCmd(
	srv rpc.ArduinoCoreServiceServer,
	portArgs *arguments.Port, fqbnArg *arguments.Fqbn, profileArg *arguments.Profile, sketchPathArg string,
	configs []string, describe, timestamp, quiet, raw bool,
) {
	logrus.Info("Executing `arduino-cli monitor`")
	ctx := context.Background()

	if !feedback.HasConsole() {
		quiet = true
	}

	var (
		inst                         *rpc.Instance
		profile                      *rpc.SketchProfile
		fqbn                         string
		defaultPort, defaultProtocol string
	)

	// Flags takes maximum precedence over sketch.yaml
	// If {--port --fqbn --profile} are set we ignore the profile.
	// If both {--port --profile} are set we read the fqbn in the following order: profile -> default_fqbn -> discovery
	// If only --port is set we read the fqbn in the following order: default_fqbn -> discovery
	// If only --fqbn is set we read the port in the following order: default_port
	sketchPath := arguments.InitSketchPath(sketchPathArg)
	sketch, err := commands.LoadSketch(context.Background(), &rpc.LoadSketchRequest{SketchPath: sketchPath.String()})
	if err != nil && !portArgs.IsPortFlagSet() {
		feedback.Fatal(
			tr("Error getting default port from `sketch.yaml`. Check if you're in the correct sketch folder or provide the --port flag: %s", err),
			feedback.ErrGeneric,
		)
	}
	if sketch != nil {
		defaultPort, defaultProtocol = sketch.GetDefaultPort(), sketch.GetDefaultProtocol()
	}
	if fqbnArg.String() == "" {
		if profileArg.Get() == "" {
			inst, profile = instance.CreateAndInitWithProfile(ctx, srv, sketch.GetDefaultProfile().GetName(), sketchPath)
		} else {
			inst, profile = instance.CreateAndInitWithProfile(ctx, srv, profileArg.Get(), sketchPath)
		}
	}
	if inst == nil {
		inst = instance.CreateAndInit(ctx, srv)
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
		fqbn, _ = portArgs.DetectFQBN(inst, srv)
	}

	portAddress, portProtocol, err := portArgs.GetPortAddressAndProtocol(inst, srv, defaultPort, defaultProtocol)
	if err != nil {
		feedback.FatalError(err, feedback.ErrGeneric)
	}

	enumerateResp, err := commands.EnumerateMonitorPortSettings(context.Background(), &rpc.EnumerateMonitorPortSettingsRequest{
		Instance:     inst,
		PortProtocol: portProtocol,
		Fqbn:         fqbn,
	})
	if err != nil {
		feedback.Fatal(tr("Error getting port settings details: %s", err), feedback.ErrGeneric)
	}
	if describe {
		settings := make([]*result.MonitorPortSettingDescriptor, len(enumerateResp.GetSettings()))
		for i, v := range enumerateResp.GetSettings() {
			settings[i] = result.NewMonitorPortSettingDescriptor(v)
		}
		feedback.PrintResult(&detailsResult{Settings: settings})
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
					if contains(s.GetEnumValues(), v) {
						setting = s
						break
					}
				} else {
					if strings.EqualFold(s.GetSettingId(), k) {
						if !contains(s.GetEnumValues(), v) {
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
			configuration.Settings = append(configuration.GetSettings(), &rpc.MonitorPortSetting{
				SettingId: setting.GetSettingId(),
				Value:     v,
			})
			if !quiet {
				feedback.Print(tr("Monitor port settings:"))
				feedback.Print(fmt.Sprintf("%s=%s", setting.GetSettingId(), v))
			}
		}
	}

	ttyIn, ttyOut, err := feedback.InteractiveStreams()
	if err != nil {
		feedback.FatalError(err, feedback.ErrGeneric)
	}

	if timestamp {
		ttyOut = newTimeStampWriter(ttyOut)
	}

	ctx, cancel := cleanup.InterruptableContext(ctx)
	if raw {
		if feedback.IsInteractive() {
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

	monitorServer, portProxy := commands.MonitorServerToReadWriteCloser(ctx, &rpc.MonitorPortOpenRequest{
		Instance:          inst,
		Port:              &rpc.Port{Address: portAddress, Protocol: portProtocol},
		Fqbn:              fqbn,
		PortConfiguration: configuration,
	})
	go func() {
		if !quiet {
			feedback.Print(tr("Connecting to %s. Press CTRL-C to exit.", portAddress))
		}
		if err := srv.Monitor(monitorServer); err != nil {
			feedback.FatalError(err, feedback.ErrGeneric)
		}
		portProxy.Close()
		cancel()
	}()
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
	Settings []*result.MonitorPortSettingDescriptor `json:"settings"`
}

func (r *detailsResult) Data() interface{} {
	return r
}

func (r *detailsResult) String() string {
	if len(r.Settings) == 0 {
		return ""
	}
	t := table.New()
	t.SetHeader(tr("ID"), tr("Setting"), tr("Default"), tr("Values"))

	green := color.New(color.FgGreen)
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
