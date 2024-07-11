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

package arguments

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/arduino/arduino-cli/commands"
	"github.com/arduino/arduino-cli/commands/cmderrors"
	f "github.com/arduino/arduino-cli/internal/algorithms"
	"github.com/arduino/arduino-cli/internal/i18n"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// Port contains the port arguments result.
// This is useful so all flags used by commands that need
// this information are consistent with each other.
type Port struct {
	address  string
	protocol string
	timeout  DiscoveryTimeout
}

// AddToCommand adds the flags used to set port and protocol to the specified Command
func (p *Port) AddToCommand(cmd *cobra.Command, srv rpc.ArduinoCoreServiceServer) {
	cmd.Flags().StringVarP(&p.address, "port", "p", "", i18n.Tr("Upload port address, e.g.: COM3 or /dev/ttyACM2"))
	cmd.RegisterFlagCompletionFunc("port", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return f.Map(GetAvailablePorts(cmd.Context(), srv), (*rpc.Port).GetAddress), cobra.ShellCompDirectiveDefault
	})
	cmd.Flags().StringVarP(&p.protocol, "protocol", "l", "", i18n.Tr("Upload port protocol, e.g: serial"))
	cmd.RegisterFlagCompletionFunc("protocol", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return f.Map(GetAvailablePorts(cmd.Context(), srv), (*rpc.Port).GetProtocol), cobra.ShellCompDirectiveDefault
	})
	p.timeout.AddToCommand(cmd)
}

// GetPortAddressAndProtocol returns only the port address and the port protocol
// without any other port metadata obtained from the discoveries.
// This method allows will bypass the discoveries if:
// - a nil instance is passed: in this case the plain port and protocol arguments are returned (even if empty)
// - a protocol is specified: in this case the discoveries are not needed to autodetect the protocol.
func (p *Port) GetPortAddressAndProtocol(ctx context.Context, instance *rpc.Instance, srv rpc.ArduinoCoreServiceServer, defaultAddress, defaultProtocol string) (string, string, error) {
	if p.protocol != "" || instance == nil {
		return p.address, p.protocol, nil
	}

	port, err := p.GetPort(ctx, instance, srv, defaultAddress, defaultProtocol)
	if err != nil {
		return "", "", err
	}
	return port.GetAddress(), port.GetProtocol(), nil
}

// GetPort returns the Port obtained by parsing command line arguments.
// The extra metadata for the ports is obtained using the pluggable discoveries.
func (p *Port) GetPort(ctx context.Context, instance *rpc.Instance, srv rpc.ArduinoCoreServiceServer, defaultAddress, defaultProtocol string) (*rpc.Port, error) {
	address := p.address
	protocol := p.protocol
	if address == "" && (defaultAddress != "" || defaultProtocol != "") {
		address, protocol = defaultAddress, defaultProtocol
	}
	if address == "" {
		// If no address is provided we assume the user is trying to upload
		// to a board that supports a tool that automatically detects
		// the attached board without specifying explicitly a port.
		// Tools that work this way must be specified using the property
		// "BOARD_ID.upload.tool.default" in the platform's boards.txt.
		return &rpc.Port{
			Protocol: "default",
		}, nil
	}
	logrus.WithField("port", address).Tracef("Upload port")

	ctx, cancel := context.WithTimeout(ctx, p.timeout.Get())
	defer cancel()

	stream, watcher := commands.BoardListWatchProxyToChan(ctx)
	go func() {
		_ = srv.BoardListWatch(&rpc.BoardListWatchRequest{Instance: instance}, stream)
	}()

	for {
		select {
		case portEvent := <-watcher:
			if portEvent.GetEventType() != "add" {
				continue
			}
			port := portEvent.GetPort().GetPort()
			if (protocol == "" || protocol == port.GetProtocol()) && address == port.GetAddress() {
				return port, nil
			}

		case <-ctx.Done():
			// No matching port found
			if protocol == "" {
				return &rpc.Port{
					Address:  address,
					Protocol: "serial",
				}, nil
			}
			return nil, errors.New(i18n.Tr("port not found: %[1]s %[2]s", address, protocol))
		}
	}
}

// GetSearchTimeout returns the timeout
func (p *Port) GetSearchTimeout() time.Duration {
	return p.timeout.Get()
}

// DetectFQBN tries to identify the board connected to the port and returns the
// discovered Port object together with the FQBN. If the port does not match
// exactly 1 board,
func (p *Port) DetectFQBN(ctx context.Context, inst *rpc.Instance, srv rpc.ArduinoCoreServiceServer) (string, *rpc.Port, error) {
	detectedPorts, err := srv.BoardList(ctx, &rpc.BoardListRequest{
		Instance: inst,
		Timeout:  p.timeout.Get().Milliseconds(),
	})
	if err != nil {
		return "", nil, fmt.Errorf("%s: %w", i18n.Tr("Error during board detection"), err)
	}
	for _, detectedPort := range detectedPorts.GetPorts() {
		port := detectedPort.GetPort()
		if p.address != port.GetAddress() {
			continue
		}
		if p.protocol != "" && p.protocol != port.GetProtocol() {
			continue
		}
		if len(detectedPort.GetMatchingBoards()) > 1 {
			return "", nil, &cmderrors.MultipleBoardsDetectedError{Port: port}
		}
		if len(detectedPort.GetMatchingBoards()) == 0 {
			return "", nil, &cmderrors.NoBoardsDetectedError{Port: port}
		}
		return detectedPort.GetMatchingBoards()[0].GetFqbn(), port, nil
	}
	return "", nil, &cmderrors.NoBoardsDetectedError{Port: &rpc.Port{Address: p.address, Protocol: p.protocol}}
}

// IsPortFlagSet returns true if the port address is provided
func (p *Port) IsPortFlagSet() bool {
	return p.address != ""
}
