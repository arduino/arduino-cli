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

package board

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sort"

	"github.com/arduino/arduino-cli/commands"
	"github.com/arduino/arduino-cli/commands/cmderrors"
	"github.com/arduino/arduino-cli/internal/arduino/cores"
	"github.com/arduino/arduino-cli/internal/cli/arguments"
	"github.com/arduino/arduino-cli/internal/cli/feedback"
	"github.com/arduino/arduino-cli/internal/cli/feedback/result"
	"github.com/arduino/arduino-cli/internal/cli/feedback/table"
	"github.com/arduino/arduino-cli/internal/cli/instance"
	"github.com/arduino/arduino-cli/internal/i18n"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func initListCommand(srv rpc.ArduinoCoreServiceServer) *cobra.Command {
	var timeoutArg arguments.DiscoveryTimeout
	var watch bool
	var fqbn arguments.Fqbn
	listCommand := &cobra.Command{
		Use:     "list",
		Short:   i18n.Tr("List connected boards."),
		Long:    i18n.Tr("Detects and displays a list of boards connected to the current computer."),
		Example: "  " + os.Args[0] + " board list --discovery-timeout 10s",
		Args:    cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			runListCommand(cmd.Context(), srv, watch, timeoutArg.Get().Milliseconds(), fqbn.String())
		},
	}

	timeoutArg.AddToCommand(listCommand)
	fqbn.AddToCommand(listCommand, srv)
	listCommand.Flags().BoolVarP(&watch, "watch", "w", false, i18n.Tr("Command keeps running and prints list of connected boards whenever there is a change."))
	return listCommand
}

// runListCommand detects and lists the connected arduino boards
func runListCommand(ctx context.Context, srv rpc.ArduinoCoreServiceServer, watch bool, timeout int64, fqbn string) {
	inst := instance.CreateAndInit(ctx, srv)

	logrus.Info("Executing `arduino-cli board list`")

	if watch {
		watchList(ctx, inst, srv)
		return
	}

	list, err := srv.BoardList(ctx, &rpc.BoardListRequest{
		Instance: inst,
		Timeout:  timeout,
		Fqbn:     fqbn,
	})
	ports := list.GetPorts()
	discoveryErrors := list.GetWarnings()
	var invalidFQBNErr *cmderrors.InvalidFQBNError
	if errors.As(err, &invalidFQBNErr) {
		feedback.Fatal(err.Error(), feedback.ErrBadArgument)
	}
	if err != nil {
		feedback.Warning(i18n.Tr("Error detecting boards: %v", err))
	}
	for _, err := range discoveryErrors {
		feedback.Warning(i18n.Tr("Error starting discovery: %v", err))
	}

	feedback.PrintResult(listResult{result.NewDetectedPorts(ports)})
}

func watchList(ctx context.Context, inst *rpc.Instance, srv rpc.ArduinoCoreServiceServer) {
	stream, eventsChan := commands.BoardListWatchProxyToChan(ctx)
	err := srv.BoardListWatch(&rpc.BoardListWatchRequest{Instance: inst}, stream)
	if err != nil {
		feedback.Fatal(i18n.Tr("Error detecting boards: %v", err), feedback.ErrNetwork)
	}

	// This is done to avoid printing the header each time a new event is received
	if feedback.GetFormat() == feedback.Text {
		t := table.New()
		t.SetHeader(i18n.Tr("Port"), i18n.Tr("Type"), i18n.Tr("Event"), i18n.Tr("Board Name"), i18n.Tr("FQBN"), i18n.Tr("Core"))
		feedback.Print(t.Render())
	}

	for event := range eventsChan {
		if res := result.NewBoardListWatchResponse(event); res != nil {
			feedback.PrintResult(watchEventResult{
				Type:   res.EventType,
				Boards: res.Port.MatchingBoards,
				Port:   res.Port.Port,
				Error:  res.Error,
			})
		}
	}
}

// output from this command requires special formatting, let's create a dedicated
// feedback.Result implementation
type listResult struct {
	Ports []*result.DetectedPort `json:"detected_ports"`
}

func (dr listResult) Data() interface{} {
	return dr
}

func (dr listResult) String() string {
	if len(dr.Ports) == 0 {
		return i18n.Tr("No boards found.")
	}

	sort.Slice(dr.Ports, func(i, j int) bool {
		x, y := dr.Ports[i].Port, dr.Ports[j].Port
		return x.Protocol < y.Protocol ||
			(x.Protocol == y.Protocol && x.Address < y.Address)
	})

	t := table.New()
	t.SetHeader(i18n.Tr("Port"), i18n.Tr("Protocol"), i18n.Tr("Type"), i18n.Tr("Board Name"), i18n.Tr("FQBN"), i18n.Tr("Core"))
	for _, detectedPort := range dr.Ports {
		port := detectedPort.Port
		protocol := port.Protocol
		address := port.Address
		if port.Protocol == "serial" {
			address = port.Address
		}
		protocolLabel := port.ProtocolLabel
		if boards := detectedPort.MatchingBoards; len(boards) > 0 {
			sort.Slice(boards, func(i, j int) bool {
				x, y := boards[i], boards[j]
				return x.Name < y.Name || (x.Name == y.Name && x.Fqbn < y.Fqbn)
			})
			for _, b := range boards {
				board := b.Name

				// to improve the user experience, show on a dedicated column
				// the name of the core supporting the board detected
				var coreName = ""
				fqbn, err := cores.ParseFQBN(b.Fqbn)
				if err == nil {
					coreName = fmt.Sprintf("%s:%s", fqbn.Package, fqbn.PlatformArch)
				}

				t.AddRow(address, protocol, protocolLabel, board, fqbn, coreName)

				// reset address and protocol, we only show them on the first row
				address = ""
				protocol = ""
			}
		} else {
			board := i18n.Tr("Unknown")
			fqbn := ""
			coreName := ""
			t.AddRow(address, protocol, protocolLabel, board, fqbn, coreName)
		}
	}
	return t.Render()
}

type watchEventResult struct {
	Type   string                  `json:"eventType"`
	Boards []*result.BoardListItem `json:"matching_boards,omitempty"`
	Port   *result.Port            `json:"port,omitempty"`
	Error  string                  `json:"error,omitempty"`
}

func (dr watchEventResult) Data() interface{} {
	return dr
}

func (dr watchEventResult) String() string {
	t := table.New()

	event := map[string]string{
		"add":    i18n.Tr("Connected"),
		"remove": i18n.Tr("Disconnected"),
	}[dr.Type]

	address := fmt.Sprintf("%s://%s", dr.Port.Protocol, dr.Port.Address)
	if dr.Port.Protocol == "serial" || dr.Port.Protocol == "" {
		address = dr.Port.Address
	}
	protocol := dr.Port.ProtocolLabel
	if boards := dr.Boards; len(boards) > 0 {
		sort.Slice(boards, func(i, j int) bool {
			x, y := boards[i], boards[j]
			return x.Name < y.Name || (x.Name == y.Name && x.Fqbn < y.Fqbn)
		})
		for _, b := range boards {
			board := b.Name

			// to improve the user experience, show on a dedicated column
			// the name of the core supporting the board detected
			var coreName = ""
			fqbn, err := cores.ParseFQBN(b.Fqbn)
			if err == nil {
				coreName = fmt.Sprintf("%s:%s", fqbn.Package, fqbn.PlatformArch)
			}

			t.AddRow(address, protocol, event, board, fqbn, coreName)

			// reset address and protocol, we only show them on the first row
			address = ""
			protocol = ""
		}
	} else {
		board := ""
		fqbn := ""
		coreName := ""
		t.AddRow(address, protocol, event, board, fqbn, coreName)
	}
	return t.Render()
}
