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
	"fmt"
	"os"
	"sort"

	"github.com/arduino/arduino-cli/arduino/cores"
	"github.com/arduino/arduino-cli/commands/board"
	"github.com/arduino/arduino-cli/internal/cli/arguments"
	"github.com/arduino/arduino-cli/internal/cli/feedback"
	"github.com/arduino/arduino-cli/internal/cli/instance"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	"github.com/arduino/arduino-cli/table"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	timeoutArg arguments.DiscoveryTimeout
	watch      bool
)

func initListCommand() *cobra.Command {
	listCommand := &cobra.Command{
		Use:     "list",
		Short:   tr("List connected boards."),
		Long:    tr("Detects and displays a list of boards connected to the current computer."),
		Example: "  " + os.Args[0] + " board list --discovery-timeout 10s",
		Args:    cobra.NoArgs,
		Run:     runListCommand,
	}

	timeoutArg.AddToCommand(listCommand)
	listCommand.Flags().BoolVarP(&watch, "watch", "w", false, tr("Command keeps running and prints list of connected boards whenever there is a change."))

	return listCommand
}

// runListCommand detects and lists the connected arduino boards
func runListCommand(cmd *cobra.Command, args []string) {
	inst := instance.CreateAndInit()

	logrus.Info("Executing `arduino-cli board list`")

	if watch {
		watchList(cmd, inst)
		return
	}

	ports, discvoeryErrors, err := board.List(&rpc.BoardListRequest{
		Instance: inst,
		Timeout:  timeoutArg.Get().Milliseconds(),
	})
	if err != nil {
		feedback.Warning(tr("Error detecting boards: %v", err))
	}
	for _, err := range discvoeryErrors {
		feedback.Warning(tr("Error starting discovery: %v", err))
	}
	feedback.PrintResult(result{ports})
}

func watchList(cmd *cobra.Command, inst *rpc.Instance) {
	eventsChan, closeCB, err := board.Watch(&rpc.BoardListWatchRequest{Instance: inst})
	if err != nil {
		feedback.Fatal(tr("Error detecting boards: %v", err), feedback.ErrNetwork)
	}
	defer closeCB()

	// This is done to avoid printing the header each time a new event is received
	if feedback.GetFormat() == feedback.Text {
		t := table.New()
		t.SetHeader(tr("Port"), tr("Type"), tr("Event"), tr("Board Name"), tr("FQBN"), tr("Core"))
		feedback.Print(t.Render())
	}

	for event := range eventsChan {
		feedback.PrintResult(watchEvent{
			Type:          event.EventType,
			Label:         event.Port.Port.Label,
			Address:       event.Port.Port.Address,
			Protocol:      event.Port.Port.Protocol,
			ProtocolLabel: event.Port.Port.ProtocolLabel,
			Properties:    event.Port.Port.Properties,
			Boards:        event.Port.MatchingBoards,
			Error:         event.Error,
		})
	}
}

// output from this command requires special formatting, let's create a dedicated
// feedback.Result implementation
type result struct {
	ports []*rpc.DetectedPort
}

func (dr result) Data() interface{} {
	return dr.ports
}

func (dr result) String() string {
	if len(dr.ports) == 0 {
		return tr("No boards found.")
	}

	sort.Slice(dr.ports, func(i, j int) bool {
		x, y := dr.ports[i].Port, dr.ports[j].Port
		return x.GetProtocol() < y.GetProtocol() ||
			(x.GetProtocol() == y.GetProtocol() && x.GetAddress() < y.GetAddress())
	})

	t := table.New()
	t.SetHeader(tr("Port"), tr("Protocol"), tr("Type"), tr("Board Name"), tr("FQBN"), tr("Core"))
	for _, detectedPort := range dr.ports {
		port := detectedPort.Port
		protocol := port.GetProtocol()
		address := port.GetAddress()
		if port.GetProtocol() == "serial" {
			address = port.GetAddress()
		}
		protocolLabel := port.GetProtocolLabel()
		if boards := detectedPort.GetMatchingBoards(); len(boards) > 0 {
			sort.Slice(boards, func(i, j int) bool {
				x, y := boards[i], boards[j]
				return x.GetName() < y.GetName() || (x.GetName() == y.GetName() && x.GetFqbn() < y.GetFqbn())
			})
			for _, b := range boards {
				board := b.GetName()

				// to improve the user experience, show on a dedicated column
				// the name of the core supporting the board detected
				var coreName = ""
				fqbn, err := cores.ParseFQBN(b.GetFqbn())
				if err == nil {
					coreName = fmt.Sprintf("%s:%s", fqbn.Package, fqbn.PlatformArch)
				}

				t.AddRow(address, protocol, protocolLabel, board, fqbn, coreName)

				// reset address and protocol, we only show them on the first row
				address = ""
				protocol = ""
			}
		} else {
			board := tr("Unknown")
			fqbn := ""
			coreName := ""
			t.AddRow(address, protocol, protocolLabel, board, fqbn, coreName)
		}
	}
	return t.Render()
}

type watchEvent struct {
	Type          string               `json:"type"`
	Address       string               `json:"address,omitempty"`
	Label         string               `json:"label,omitempty"`
	Protocol      string               `json:"protocol,omitempty"`
	ProtocolLabel string               `json:"protocol_label,omitempty"`
	Properties    map[string]string    `json:"properties"`
	Boards        []*rpc.BoardListItem `json:"boards,omitempty"`
	Error         string               `json:"error,omitempty"`
}

func (dr watchEvent) Data() interface{} {
	return dr
}

func (dr watchEvent) String() string {
	t := table.New()

	event := map[string]string{
		"add":    tr("Connected"),
		"remove": tr("Disconnected"),
	}[dr.Type]

	address := fmt.Sprintf("%s://%s", dr.Protocol, dr.Address)
	if dr.Protocol == "serial" || dr.Protocol == "" {
		address = dr.Address
	}
	protocol := dr.ProtocolLabel
	if boards := dr.Boards; len(boards) > 0 {
		sort.Slice(boards, func(i, j int) bool {
			x, y := boards[i], boards[j]
			return x.GetName() < y.GetName() || (x.GetName() == y.GetName() && x.GetFqbn() < y.GetFqbn())
		})
		for _, b := range boards {
			board := b.GetName()

			// to improve the user experience, show on a dedicated column
			// the name of the core supporting the board detected
			var coreName = ""
			fqbn, err := cores.ParseFQBN(b.GetFqbn())
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
