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

	"github.com/arduino/arduino-cli/arduino/cores"
	"github.com/arduino/arduino-cli/commands"
	rpc "github.com/arduino/arduino-cli/rpc/commands"
)

// Details FIXMEDOC
func Details(ctx context.Context, req *rpc.BoardDetailsReq) (*rpc.BoardDetailsResp, error) {
	pm := commands.GetPackageManager(req.GetInstance().GetId())
	if pm == nil {
		return nil, errors.New("invalid instance")
	}

	fqbn, err := cores.ParseFQBN(req.GetFqbn())
	if err != nil {
		return nil, fmt.Errorf("parsing fqbn: %s", err)
	}

	_, _, board, _, _, err := pm.ResolveFQBN(fqbn)
	if err != nil {
		return nil, fmt.Errorf("loading board data: %s", err)
	}

	details := &rpc.BoardDetailsResp{}
	details.Name = board.Name()
	details.ConfigOptions = []*rpc.ConfigOption{}
	options := board.GetConfigOptions()
	for _, option := range options.Keys() {
		configOption := &rpc.ConfigOption{}
		configOption.Option = option
		configOption.OptionLabel = options.Get(option)
		selected, hasSelected := fqbn.Configs.GetOk(option)

		values := board.GetConfigOptionValues(option)
		for i, value := range values.Keys() {
			configValue := &rpc.ConfigValue{}
			if hasSelected && value == selected {
				configValue.Selected = true
			} else if !hasSelected && i == 0 {
				configValue.Selected = true
			}

			configValue.Value = value
			configValue.ValueLabel = values.Get(value)
			configOption.Values = append(configOption.Values, configValue)
		}

		details.ConfigOptions = append(details.ConfigOptions, configOption)
	}

	details.RequiredTools = []*rpc.RequiredTool{}
	for _, reqTool := range board.PlatformRelease.Dependencies {
		details.RequiredTools = append(details.RequiredTools, &rpc.RequiredTool{
			Name:     reqTool.ToolName,
			Packager: reqTool.ToolPackager,
			Version:  reqTool.ToolVersion.String(),
		})
	}

	return details, nil
}
