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

	"github.com/arduino/arduino-cli/arduino/discovery"
	"github.com/arduino/arduino-cli/arduino/sketch"
	"github.com/arduino/arduino-cli/commands"
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
}

// AddToCommand adds the flags used to set port and protocol to the specified Command
func (p *Port) AddToCommand(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&p.address, "port", "p", "", "Upload port address, e.g.: COM3 or /dev/ttyACM2")
	cmd.Flags().StringVarP(&p.protocol, "protocol", "l", "", "Upload port protocol, e.g: serial")
}

// GetPort returns the Port obtained by parsing command line arguments.
// The extra metadata for the ports is obtained using the pluggable discoveries.
func (p *Port) GetPort(instance *rpc.Instance, sk *sketch.Sketch) (*discovery.Port, error) {
	address := p.address
	protocol := p.protocol
	if address != "" && protocol != "" {
		return &discovery.Port{
			Address:  address,
			Protocol: protocol,
		}, nil
	}

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
		return nil, nil
	}
	logrus.WithField("port", address).Tracef("Upload port")

	pm := commands.GetPackageManager(instance.Id)
	if pm == nil {
		return nil, errors.New("invalid instance")
	}

	if err := pm.DiscoveryManager().RunAll(); err != nil {
		return nil, err
	}
	if err := pm.DiscoveryManager().StartAll(); err != nil {
		return nil, err
	}

	defer func() {
		// Quit all discoveries at the end.
		err := pm.DiscoveryManager().QuitAll()
		if err != nil {
			logrus.Errorf("quitting discoveries when getting port metadata: %s", err)
		}
	}()

	ports := pm.DiscoveryManager().List()

	matchingPorts := []*discovery.Port{}
	for _, port := range ports {
		if address == port.Address {
			matchingPorts = append(matchingPorts, port)
			if len(matchingPorts) > 1 {
				// Too many matching ports found, can't handle this case.
				// This must never happen.
				return nil, fmt.Errorf("multiple ports found matching address %s", address)
			}
		}
	}

	if len(matchingPorts) == 1 {
		// Only one matching port found, use it
		return matchingPorts[0], nil
	}

	// In case no matching port is found assume the address refers to a serial port
	return &discovery.Port{
		Address:  address,
		Protocol: "serial",
	}, nil
}
