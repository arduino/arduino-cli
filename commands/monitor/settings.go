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

	"github.com/arduino/arduino-cli/commands/cmderrors"
	"github.com/arduino/arduino-cli/commands/internal/instances"
	pluggableMonitor "github.com/arduino/arduino-cli/internal/arduino/monitor"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
)

// EnumerateMonitorPortSettings returns a description of the configuration settings of a monitor port
func EnumerateMonitorPortSettings(ctx context.Context, req *rpc.EnumerateMonitorPortSettingsRequest) (*rpc.EnumerateMonitorPortSettingsResponse, error) {
	pme, release, err := instances.GetPackageManagerExplorer(req.GetInstance())
	if err != nil {
		return nil, err
	}
	defer release()

	m, boardSettings, err := findMonitorAndSettingsForProtocolAndBoard(pme, req.GetPortProtocol(), req.GetFqbn())
	if err != nil {
		return nil, err
	}

	if err := m.Run(); err != nil {
		return nil, &cmderrors.FailedMonitorError{Cause: err}
	}
	defer m.Quit()

	desc, err := m.Describe()
	if err != nil {
		return nil, &cmderrors.FailedMonitorError{Cause: err}
	}

	// Apply default settings for this board and protocol
	for setting, value := range boardSettings.AsMap() {
		if param, ok := desc.ConfigurationParameters[setting]; ok {
			for _, v := range param.Values {
				if v == value {
					param.Selected = value
					break
				}
			}
		}
	}

	return &rpc.EnumerateMonitorPortSettingsResponse{Settings: convert(desc)}, nil
}

func convert(desc *pluggableMonitor.PortDescriptor) []*rpc.MonitorPortSettingDescriptor {
	res := []*rpc.MonitorPortSettingDescriptor{}
	for settingID, descriptor := range desc.ConfigurationParameters {
		res = append(res, &rpc.MonitorPortSettingDescriptor{
			SettingId:  settingID,
			Label:      descriptor.Label,
			Type:       descriptor.Type,
			EnumValues: descriptor.Values,
			Value:      descriptor.Selected,
		})
	}
	return res
}
