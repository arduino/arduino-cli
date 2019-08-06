/*
 * This file is part of arduino-cli.
 *
 * Copyright 2018 ARDUINO SA (http://www.arduino.cc/)
 *
 * This software is released under the GNU General Public License version 3,
 * which covers the main part of arduino-cli.
 * The terms of this license can be found at:
 * https://www.gnu.org/licenses/gpl-3.0.en.html
 *
 * You can be released from the requirements of the above licenses by purchasing
 * a commercial license. Buying such a license is mandatory if you want to modify or
 * otherwise use the software for commercial activities involving the Arduino
 * software without disclosing the source code of your own applications. To purchase
 * a commercial license, send an email to license@arduino.cc.
 */

package board

import (
	"github.com/arduino/arduino-cli/commands"
	rpc "github.com/arduino/arduino-cli/rpc/commands"
	"github.com/pkg/errors"
)

// List FIXMEDOC
func List(instanceID int32) (*rpc.BoardListResp, error) {
	pm := commands.GetPackageManager(instanceID)
	if pm == nil {
		return nil, errors.New("invalid instance")
	}

	serialDiscovery, err := commands.NewBuiltinSerialDiscovery(pm)
	if err != nil {
		return nil, errors.Wrap(err, "unable to instance serial-discovery")
	}

	if err := serialDiscovery.Start(); err != nil {
		return nil, errors.Wrap(err, "unable to start serial-discovery")
	}
	defer serialDiscovery.Close()

	resp := &rpc.BoardListResp{Ports: []*rpc.DetectedPort{}}

	ports, err := serialDiscovery.List()
	if err != nil {
		return nil, errors.Wrap(err, "error getting port list from serial-discovery")
	}

	for _, port := range ports {
		b := []*rpc.BoardListItem{}
		for _, board := range pm.IdentifyBoard(port.IdentificationPrefs) {
			b = append(b, &rpc.BoardListItem{
				Name: board.Name(),
				FQBN: board.FQBN(),
			})
		}
		p := &rpc.DetectedPort{
			Address:       port.Address,
			Protocol:      port.Protocol,
			ProtocolLabel: port.ProtocolLabel,
			Boards:        b,
		}
		resp.Ports = append(resp.Ports, p)
	}

	return resp, nil
}
