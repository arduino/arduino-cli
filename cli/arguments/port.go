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
	"fmt"
	"net/url"
	"os"
	"time"

	"github.com/arduino/arduino-cli/arduino"
	"github.com/arduino/arduino-cli/arduino/discovery"
	"github.com/arduino/arduino-cli/arduino/sketch"
	"github.com/arduino/arduino-cli/cli/errorcodes"
	"github.com/arduino/arduino-cli/cli/feedback"
	"github.com/arduino/arduino-cli/commands"
	"github.com/arduino/arduino-cli/commands/board"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	"github.com/pkg/errors"
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
func (p *Port) AddToCommand(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&p.address, "port", "p", "", tr("Upload port address, e.g.: COM3 or /dev/ttyACM2"))
	cmd.RegisterFlagCompletionFunc("port", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return GetConnectedBoards(), cobra.ShellCompDirectiveDefault
	})
	cmd.Flags().StringVarP(&p.protocol, "protocol", "l", "", tr("Upload port protocol, e.g: serial"))
	cmd.RegisterFlagCompletionFunc("protocol", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return GetInstalledProtocols(), cobra.ShellCompDirectiveDefault
	})
	p.timeout.AddToCommand(cmd)
}

// GetPortAddressAndProtocol returns only the port address and the port protocol
// without any other port metadata obtained from the discoveries. This method allows
// to bypass the discoveries unless the protocol is not specified: in this
// case the discoveries are needed to autodetect the protocol.
func (p *Port) GetPortAddressAndProtocol(instance *rpc.Instance, sk *sketch.Sketch) (string, string, error) {
	if p.protocol != "" {
		return p.address, p.protocol, nil
	}
	port, err := p.GetPort(instance, sk)
	if err != nil {
		return "", "", err
	}
	return port.Address, port.Protocol, nil
}

// GetPort returns the Port obtained by parsing command line arguments.
// The extra metadata for the ports is obtained using the pluggable discoveries.
func (p *Port) GetPort(instance *rpc.Instance, sk *sketch.Sketch) (*discovery.Port, error) {
	// TODO: REMOVE sketch.Sketch from here
	// TODO: REMOVE discovery from here (use board.List instead)

	address := p.address
	protocol := p.protocol

	if address == "" && sk != nil && sk.Metadata != nil {
		deviceURI, err := url.Parse(sk.Metadata.CPU.Port)
		if err != nil {
			return nil, errors.Errorf("invalid Device URL format: %s", err)
		}
		if deviceURI.Scheme == "serial" {
			address = deviceURI.Host + deviceURI.Path
		}
	}
	if address == "" {
		// If no address is provided we assume the user is trying to upload
		// to a board that supports a tool that automatically detects
		// the attached board without specifying explictly a port.
		// Tools that work this way must be specified using the property
		// "BOARD_ID.upload.tool.default" in the platform's boards.txt.
		return &discovery.Port{
			Protocol: "default",
		}, nil
	}
	logrus.WithField("port", address).Tracef("Upload port")

	pm := commands.GetPackageManager(instance.Id)
	if pm == nil {
		return nil, errors.New("invalid instance")
	}
	dm := pm.DiscoveryManager()
	watcher, err := dm.Watch()
	if err != nil {
		return nil, err
	}
	defer watcher.Close()

	deadline := time.After(p.timeout.Get())
	for {
		select {
		case portEvent := <-watcher.Feed():
			if portEvent.Type != "add" {
				continue
			}
			port := portEvent.Port
			if (protocol == "" || protocol == port.Protocol) && address == port.Address {
				return port, nil
			}

		case <-deadline:
			// No matching port found
			if protocol == "" {
				return &discovery.Port{
					Address:  address,
					Protocol: "serial",
				}, nil
			}
			return nil, fmt.Errorf(tr("port not found: %[1]s %[2]s"), address, protocol)
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
func (p *Port) DetectFQBN(inst *rpc.Instance) (string, *rpc.Port) {
	detectedPorts, _, err := board.List(&rpc.BoardListRequest{
		Instance: inst,
		Timeout:  p.timeout.Get().Milliseconds(),
	})
	if err != nil {
		feedback.Errorf(tr("Error during FQBN detection: %v", err))
		os.Exit(errorcodes.ErrGeneric)
	}
	for _, detectedPort := range detectedPorts {
		port := detectedPort.GetPort()
		if p.address != port.GetAddress() {
			continue
		}
		if p.protocol != "" && p.protocol != port.GetProtocol() {
			continue
		}
		if len(detectedPort.MatchingBoards) > 1 {
			feedback.Error(&arduino.MultipleBoardsDetectedError{Port: port})
			os.Exit(errorcodes.ErrBadArgument)
		}
		if len(detectedPort.MatchingBoards) == 0 {
			feedback.Error(&arduino.NoBoardsDetectedError{Port: port})
			os.Exit(errorcodes.ErrBadArgument)
		}
		return detectedPort.MatchingBoards[0].Fqbn, port
	}
	return "", nil
}
