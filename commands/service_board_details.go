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

package commands

import (
	"context"

	"github.com/arduino/arduino-cli/commands/cmderrors"
	"github.com/arduino/arduino-cli/commands/internal/instances"
	"github.com/arduino/arduino-cli/internal/arduino/utils"
	"github.com/arduino/arduino-cli/pkg/fqbn"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
)

// BoardDetails returns all details for a board including tools and HW identifiers.
// This command basically gather al the information and translates it into the required grpc struct properties
func (s *arduinoCoreServerImpl) BoardDetails(ctx context.Context, req *rpc.BoardDetailsRequest) (*rpc.BoardDetailsResponse, error) {
	pme, release, err := instances.GetPackageManagerExplorer(req.GetInstance())
	if err != nil {
		return nil, err
	}
	defer release()

	fqbn, err := fqbn.Parse(req.GetFqbn())
	if err != nil {
		return nil, &cmderrors.InvalidFQBNError{Cause: err}
	}

	boardPackage, boardPlatformRelease, board, boardProperties, _, err := pme.ResolveFQBN(fqbn)
	if err != nil {
		return nil, &cmderrors.UnknownFQBNError{Cause: err}
	}

	details := &rpc.BoardDetailsResponse{}
	details.Name = board.Name()
	details.Fqbn = board.FQBN()
	details.PropertiesId = board.BoardID
	details.Official = fqbn.Packager == "arduino"
	details.Version = board.PlatformRelease.Version.String()
	details.IdentificationProperties = []*rpc.BoardIdentificationProperties{}
	for _, p := range board.GetIdentificationProperties() {
		details.IdentificationProperties = append(details.GetIdentificationProperties(), &rpc.BoardIdentificationProperties{
			Properties: p.AsMap(),
		})
	}
	for _, k := range boardProperties.Keys() {
		v := boardProperties.Get(k)
		details.BuildProperties = append(details.GetBuildProperties(), k+"="+v)
	}
	if !req.GetDoNotExpandBuildProperties() {
		details.BuildProperties, _ = utils.ExpandBuildProperties(details.GetBuildProperties())
	}

	details.Package = &rpc.Package{
		Name:       boardPackage.Name,
		Maintainer: boardPackage.Maintainer,
		WebsiteUrl: boardPackage.WebsiteURL,
		Email:      boardPackage.Email,
		Help:       &rpc.Help{Online: boardPackage.Help.Online},
		Url:        boardPackage.URL,
	}

	details.Platform = &rpc.BoardPlatform{
		Architecture: boardPlatformRelease.Platform.Architecture,
		Category:     boardPlatformRelease.Category,
		Name:         boardPlatformRelease.Name,
	}

	if boardPlatformRelease.Resource != nil {
		details.Platform.Url = boardPlatformRelease.Resource.URL
		details.Platform.ArchiveFilename = boardPlatformRelease.Resource.ArchiveFileName
		details.Platform.Checksum = boardPlatformRelease.Resource.Checksum
		details.Platform.Size = boardPlatformRelease.Resource.Size
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
			configOption.Values = append(configOption.GetValues(), configValue)
		}

		details.ConfigOptions = append(details.GetConfigOptions(), configOption)
	}

	details.ToolsDependencies = []*rpc.ToolsDependencies{}
	for _, tool := range boardPlatformRelease.ToolDependencies {
		toolRelease := pme.FindToolDependency(tool)
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
		details.ToolsDependencies = append(details.GetToolsDependencies(), &rpc.ToolsDependencies{
			Name:     tool.ToolName,
			Packager: tool.ToolPackager,
			Version:  tool.ToolVersion.String(),
			Systems:  systems,
		})
	}

	details.DefaultProgrammerId = board.GetDefaultProgrammerID()
	details.Programmers = []*rpc.Programmer{}
	for id, p := range boardPlatformRelease.Programmers {
		details.Programmers = append(details.GetProgrammers(), &rpc.Programmer{
			Platform: boardPlatformRelease.Name,
			Id:       id,
			Name:     p.Name,
		})
	}

	return details, nil
}
