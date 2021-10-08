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
	"io"

	"github.com/arduino/arduino-cli/arduino/cores"
	"github.com/arduino/arduino-cli/arduino/cores/packagemanager"
	pluggableMonitor "github.com/arduino/arduino-cli/arduino/monitor"
	"github.com/arduino/arduino-cli/commands"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
)

// PortProxy is an io.ReadWriteCloser that maps into the monitor port of the board
type PortProxy struct {
	rw               io.ReadWriter
	changeSettingsCB func(setting, value string) error
	closeCB          func() error
}

func (p *PortProxy) Read(buff []byte) (int, error) {
	return p.rw.Read(buff)
}

func (p *PortProxy) Write(buff []byte) (int, error) {
	return p.rw.Write(buff)
}

// Config sets the port configuration setting to the specified value
func (p *PortProxy) Config(setting, value string) error {
	return p.changeSettingsCB(setting, value)
}

// Close the port
func (p *PortProxy) Close() error {
	return p.closeCB()
}

// Monitor opens a communication port. It returns a PortProxy to communicate with the port and a PortDescriptor
// that describes the available configuration settings.
func Monitor(ctx context.Context, req *rpc.MonitorRequest) (*PortProxy, *pluggableMonitor.PortDescriptor, error) {
	pm := commands.GetPackageManager(req.GetInstance().GetId())
	if pm == nil {
		return nil, nil, &commands.InvalidInstanceError{}
	}

	m, err := findMonitorForProtocolAndBoard(pm, req.GetPort().GetProtocol(), req.GetFqbn())
	if err != nil {
		return nil, nil, err
	}

	if err := m.Run(); err != nil {
		return nil, nil, &commands.FailedMonitorError{Cause: err}
	}

	descriptor, err := m.Describe()
	if err != nil {
		return nil, nil, &commands.FailedMonitorError{Cause: err}
	}

	monIO, err := m.Open(req.GetPort().GetAddress(), req.Port.GetProtocol())
	if err != nil {
		return nil, nil, &commands.FailedMonitorError{Cause: err}
	}

	return &PortProxy{
		rw:               monIO,
		changeSettingsCB: m.Configure,
		closeCB: func() error {
			m.Close()
			return m.Quit()
		},
	}, descriptor, nil
}

func findMonitorForProtocolAndBoard(pm *packagemanager.PackageManager, protocol string, fqbn string) (*pluggableMonitor.PluggableMonitor, error) {
	if protocol == "" {
		return nil, &commands.MissingPortProtocolError{}
	}

	var monitorDepOrRecipe *cores.MonitorDependency

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
			monitorDepOrRecipe = mon
		}
	}

	if monitorDepOrRecipe == nil {
		// Otherwise look in all package for a suitable monitor
		for _, platformRel := range pm.InstalledPlatformReleases() {
			if mon, ok := platformRel.Monitors[protocol]; ok {
				monitorDepOrRecipe = mon
				break
			}
		}
	}

	if monitorDepOrRecipe == nil {
		return nil, &commands.NoMonitorAvailableForProtocolError{Protocol: protocol}
	}

	tool := pm.FindMonitorDependency(monitorDepOrRecipe)
	if tool == nil {
		return nil, &commands.MonitorNotFoundError{Monitor: monitorDepOrRecipe.String()}
	}

	return pluggableMonitor.New(
		monitorDepOrRecipe.Name,
		tool.InstallDir.Join(monitorDepOrRecipe.Name).String(),
	), nil
}
