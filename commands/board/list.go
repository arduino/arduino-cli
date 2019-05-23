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
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/arduino/arduino-cli/arduino/discovery"
	"github.com/arduino/arduino-cli/cli"
	"github.com/arduino/arduino-cli/commands"
	"github.com/arduino/arduino-cli/common/formatter"
	"github.com/arduino/arduino-cli/rpc"
)

func BoardList(ctx context.Context, req *rpc.BoardListReq) (*rpc.BoardListResp, error) {
	pm := commands.GetPackageManager(req)
	if pm == nil {
		return nil, errors.New("invalid instance")
	}

	// Check for bultin serial-discovery tool
	loadBuiltinSerialDiscoveryMetadata(pm)
	serialDiscoveryTool, _ := getBuiltinSerialDiscoveryTool(pm)
	if !serialDiscoveryTool.IsInstalled() {
		formatter.Print("Downloading and installing missing tool: " + serialDiscoveryTool.String())
		commands.DownloadToolRelease(pm, serialDiscoveryTool, cli.OutputProgressBar())
		commands.InstallToolRelease(pm, serialDiscoveryTool, cli.OutputTaskProgress())

		if err := pm.LoadHardware(cli.Config); err != nil {
			formatter.PrintError(err, "Could not load hardware packages.")
			os.Exit(cli.ErrCoreConfig)
		}
		serialDiscoveryTool, _ = getBuiltinSerialDiscoveryTool(pm)
		if !serialDiscoveryTool.IsInstalled() {
			formatter.PrintErrorMessage("Missing serial-discovery tool.")
			os.Exit(cli.ErrCoreConfig)
		}
	}

	// TODO: move to 'commands' modules
	_, err := discovery.NewFromCommandLine(serialDiscoveryTool.InstallDir.Join("serial-discovery").String())
	if err != nil {
		formatter.PrintError(err, "Error setting up serial-discovery tool.")
		os.Exit(cli.ErrCoreConfig)
	}

	resp := &rpc.BoardListResp{Ports: []*rpc.DetectedPort{}}
	for _, disc := range commands.GetDiscoveries(req) {
		ports, err := disc.List()
		if err != nil {
			fmt.Printf("Error getting port list from discovery %s: %s\n", disc.ID, err)
			continue
		}
		for _, port := range ports {
			b := []*rpc.DetectedBoard{}
			for _, board := range pm.IdentifyBoard(port.IdentificationPrefs) {
				b = append(b, &rpc.DetectedBoard{
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
	}

	return resp, nil
}
