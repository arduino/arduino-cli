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

	//boardPackage, platformRelease, board, buildProperties, buildPlatformRelease, err := pm.ResolveFQBN(fqbn)
	boardPackage, _, board, _, _, err := pm.ResolveFQBN(fqbn)
	if err != nil {
		return nil, fmt.Errorf("loading board data: %s", err)
	}

	details := &rpc.BoardDetailsResp{}
	details.Name = board.Name()
	details.Fqbn = board.FQBN()
	details.PropertiesId = board.BoardID
	details.Official = fqbn.Package == "arduino"
	details.Version = board.PlatformRelease.Version.String()

	details.Package = &rpc.Package{
		Name:       boardPackage.Name,
		Maintainer: boardPackage.Maintainer,
		WebsiteURL: boardPackage.WebsiteURL,
		Email:      boardPackage.Email,
		Help:       &rpc.Help{Online: boardPackage.Help.Online},
		Url:        "TBD",
	}

	details.Platform = &rpc.BoardPlatform{
		Architecture:         board.PlatformRelease.Platform.Architecture,
		Category:             board.PlatformRelease.Platform.Category,
		Url:                  board.PlatformRelease.Resource.URL,
		ArchiveFileName:      board.PlatformRelease.Resource.ArchiveFileName,
		Checksum:             board.PlatformRelease.Resource.Checksum,
		Size:                 board.PlatformRelease.Resource.Size,
		Name:                 board.PlatformRelease.Platform.Name,
	}


	details.IdentificationPref = []*rpc.IdentificationPref{}
	vids := board.Properties.SubTree("vid")
	pids := board.Properties.SubTree("pid")
	for id, vid := range vids.AsMap() {
		if pid, ok := pids.GetOk(id); ok {
			idPref := rpc.IdentificationPref{UsbID: &rpc.USBID{VID: vid, PID: pid}}
			details.IdentificationPref = append(details.IdentificationPref, &idPref)
		}
	}

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

	details.ToolsDependencies = []*rpc.ToolsDependencies{}
	for _, reqTool := range board.PlatformRelease.Dependencies {
		details.ToolsDependencies = append(details.ToolsDependencies, &rpc.ToolsDependencies{
			Name:     reqTool.ToolName,
			Packager: reqTool.ToolPackager,
			Version:  reqTool.ToolVersion.String(),
		})
	}

	return details, nil
}

