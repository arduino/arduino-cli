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

	"github.com/arduino/arduino-cli/arduino/cores"
	"github.com/arduino/arduino-cli/commands"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
)

// Details returns all details for a board including tools and HW identifiers.
// This command basically gather al the information and translates it into the required grpc struct properties
func Details(ctx context.Context, req *rpc.BoardDetailsRequest) (*rpc.BoardDetailsResponse, error) {
	pm := commands.GetPackageManager(req.GetInstance().GetId())
	if pm == nil {
		return nil, &commands.InvalidInstanceError{}
	}

	fqbn, err := cores.ParseFQBN(req.GetFqbn())
	if err != nil {
		return nil, &commands.InvalidFQBNError{Cause: err}
	}

	boardPackage, boardPlatform, board, boardProperties, boardRefPlatform, err := pm.ResolveFQBN(fqbn)
	if err != nil {
		return nil, &commands.UnknownFQBNError{Cause: err}
	}

	details := &rpc.BoardDetailsResponse{}
	details.Name = board.Name()
	details.Fqbn = board.FQBN()
	details.PropertiesId = board.BoardID
	details.Official = fqbn.Package == "arduino"
	details.Version = board.PlatformRelease.Version.String()
	details.IdentificationProperties = []*rpc.BoardIdentificationProperties{}
	for _, p := range board.GetIdentificationProperties() {
		details.IdentificationProperties = append(details.IdentificationProperties, &rpc.BoardIdentificationProperties{
			Properties: p.AsMap(),
		})
	}

	details.DebuggingSupported = boardProperties.ContainsKey("debug.executable") ||
		boardPlatform.Properties.ContainsKey("debug.executable") ||
		(boardRefPlatform != nil && boardRefPlatform.Properties.ContainsKey("debug.executable")) ||
		// HOTFIX: Remove me when the `arduino:samd` core is updated
		boardPlatform.String() == "arduino:samd@1.8.9" ||
		boardPlatform.String() == "arduino:samd@1.8.8"

	details.Package = &rpc.Package{
		Name:       boardPackage.Name,
		Maintainer: boardPackage.Maintainer,
		WebsiteUrl: boardPackage.WebsiteURL,
		Email:      boardPackage.Email,
		Help:       &rpc.Help{Online: boardPackage.Help.Online},
		Url:        boardPackage.URL,
	}

	details.Platform = &rpc.BoardPlatform{
		Architecture: boardPlatform.Platform.Architecture,
		Category:     boardPlatform.Platform.Category,
		Name:         boardPlatform.Platform.Name,
	}

	if boardPlatform.Resource != nil {
		details.Platform.Url = boardPlatform.Resource.URL
		details.Platform.ArchiveFilename = boardPlatform.Resource.ArchiveFileName
		details.Platform.Checksum = boardPlatform.Resource.Checksum
		details.Platform.Size = boardPlatform.Resource.Size
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
	for _, tool := range boardPlatform.ToolDependencies {
		toolRelease := pm.FindToolDependency(tool)
		var systems []*rpc.Systems
		if toolRelease != nil {
			for _, f := range toolRelease.Flavors {
				systems = append(systems, &rpc.Systems{
					Checksum:        f.Resource.Checksum,
					Size:            f.Resource.Size,
					Host:            f.OS,
					ArchiveFilename: f.Resource.ArchiveFileName,
					Url:             f.Resource.URL,
				})
			}
		}
		details.ToolsDependencies = append(details.ToolsDependencies, &rpc.ToolsDependencies{
			Name:     tool.ToolName,
			Packager: tool.ToolPackager,
			Version:  tool.ToolVersion.String(),
			Systems:  systems,
		})
	}

	details.Programmers = []*rpc.Programmer{}
	for id, p := range boardPlatform.Programmers {
		details.Programmers = append(details.Programmers, &rpc.Programmer{
			Platform: boardPlatform.Platform.Name,
			Id:       id,
			Name:     p.Name,
		})
	}

	return details, nil
}
