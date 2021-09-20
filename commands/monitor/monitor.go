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
	"github.com/arduino/arduino-cli/arduino/cores"
	"github.com/arduino/arduino-cli/arduino/cores/packagemanager"
	"github.com/arduino/arduino-cli/commands"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
)

func findMonitorForProtocolAndBoard(pm *packagemanager.PackageManager, port *rpc.Port, fqbn string) (*cores.MonitorDependency, error) {
	if port == nil {
		return nil, &commands.MissingPortError{}
	}
	protocol := port.GetProtocol()
	if protocol == "" {
		return nil, &commands.MissingPortProtocolError{}
	}

	// If a board is specified search the monitor in the board package first
	if fqbn != "" {
		fqbn, err := cores.ParseFQBN(fqbn)
		if err != nil {
			return nil, &commands.InvalidFQBNError{Cause: err}
		}

		_, boardPlatform, _, _, _, err := pm.ResolveFQBN(fqbn)
		if err != nil {
			return nil, &commands.UnknownFQBNError{Cause: err}
		}
		if mon, ok := boardPlatform.Monitors[protocol]; ok {
			return mon, nil
		}
	}

	// Otherwise look in all package for a suitable monitor
	for _, platformRel := range pm.InstalledPlatformReleases() {
		if mon, ok := platformRel.Monitors[protocol]; ok {
			return mon, nil
		}
	}
	return nil, &commands.NoMonitorAvailableForProtocolError{Protocol: protocol}
}
