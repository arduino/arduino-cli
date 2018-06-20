/*
 * This file is part of arduino-cli.
 *
 * arduino-cli is free software; you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation; either version 2 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program; if not, write to the Free Software
 * Foundation, Inc., 51 Franklin St, Fifth Floor, Boston, MA  02110-1301  USA
 *
 * As a special exception, you may use this file as part of a free software
 * library without restriction.  Specifically, if other files instantiate
 * templates or use macros or inline functions from this file, or you compile
 * this file and link it with other files to produce an executable, this
 * file does not by itself cause the resulting executable to be covered by
 * the GNU General Public License.  This exception does not however
 * invalidate any other reasons why the executable file might be covered by
 * the GNU General Public License.
 *
 * Copyright 2017 ARDUINO AG (http://www.arduino.cc/)
 */

package packageindex

import (
	"encoding/json"
	"io/ioutil"

	"github.com/bcmi-labs/arduino-cli/arduino/resources"

	"github.com/bcmi-labs/arduino-cli/arduino/cores"
)

// Index represents Cores and Tools struct as seen from package_index.json file.
type Index struct {
	Packages []*indexPackage `json:"packages"`
}

// indexPackage represents a single entry from package_index.json file.
type indexPackage struct {
	Name       string                  `json:"name,required"`
	Maintainer string                  `json:"maintainer,required"`
	WebsiteURL string                  `json:"websiteUrl"`
	Email      string                  `json:"email"`
	Platforms  []*indexPlatformRelease `json:"platforms,required"`
	Tools      []*indexToolRelease     `json:"tools,required"`
	Help       indexHelp               `json:"help,omitempty"`
}

// indexPlatformRelease represents a single Core Platform from package_index.json file.
type indexPlatformRelease struct {
	Name             string                `json:"name,required"`
	Architecture     string                `json:"architecture"`
	Version          string                `json:"version,required"`
	Category         string                `json:"category"`
	URL              string                `json:"url"`
	ArchiveFileName  string                `json:"archiveFileName,required"`
	Checksum         string                `json:"checksum,required"`
	Size             json.Number           `json:"size,required"`
	Boards           []indexBoard          `json:"boards"`
	Help             indexHelp             `json:"help,omitempty"`
	ToolDependencies []indexToolDependency `json:"toolsDependencies,required"`
}

// indexToolDependency represents a single dependency of a core from a tool.
type indexToolDependency struct {
	Packager string `json:"packager,required"`
	Name     string `json:"name,required"`
	Version  string `json:"version,required"`
}

// indexToolRelease represents a single Tool from package_index.json file.
type indexToolRelease struct {
	Name    string                    `json:"name,required"`
	Version string                    `json:"version,required"`
	Systems []indexToolReleaseFlavour `json:"systems,required"`
}

// indexToolReleaseFlavour represents a single tool flavour in the package_index.json file.
type indexToolReleaseFlavour struct {
	OS              string      `json:"host,required"`
	URL             string      `json:"url,required"`
	ArchiveFileName string      `json:"archiveFileName,required"`
	Size            json.Number `json:"size,required"`
	Checksum        string      `json:"checksum,required"`
}

// indexBoard represents a single Board as written in package_index.json file.
type indexBoard struct {
	Name string `json:"name"`
}

type indexHelp struct {
	Online string `json:"online,omitempty"`
}

// MergeIntoPackages converts the Index data into a cores.Packages and merge them
// with the existing conents of the cores.Packages passed as parameter.
func (index Index) MergeIntoPackages(outPackages *cores.Packages) {
	for _, inPackage := range index.Packages {
		inPackage.extractPackageIn(outPackages)
	}
}

func (inPackage indexPackage) extractPackageIn(outPackages *cores.Packages) {
	outPackage := outPackages.GetOrCreatePackage(inPackage.Name)
	outPackage.Maintainer = inPackage.Maintainer
	outPackage.WebsiteURL = inPackage.WebsiteURL
	outPackage.Email = inPackage.Email

	for _, inTool := range inPackage.Tools {
		inTool.extractToolIn(outPackage)
	}

	for _, inPlatform := range inPackage.Platforms {
		inPlatform.extractPlatformIn(outPackage)
	}
}

func (inPlatformRelease indexPlatformRelease) extractPlatformIn(outPackage *cores.Package) {
	outPlatform := outPackage.GetOrCreatePlatform(inPlatformRelease.Architecture)
	// FIXME: shall we use the Name and Category of the latest release? or maybe move Name and Category in PlatformRelease?
	outPlatform.Name = inPlatformRelease.Name
	outPlatform.Category = inPlatformRelease.Category

	size, _ := inPlatformRelease.Size.Int64()
	outPlatformRelease := outPlatform.GetOrCreateRelease(inPlatformRelease.Version)
	outPlatformRelease.Resource = &resources.DownloadResource{
		ArchiveFileName: inPlatformRelease.ArchiveFileName,
		Checksum:        inPlatformRelease.Checksum,
		Size:            size,
		URL:             inPlatformRelease.URL,
		CachePath:       "packages",
	}
	outPlatformRelease.BoardsManifest = inPlatformRelease.extractBoardsManifest()
	outPlatformRelease.Dependencies = inPlatformRelease.extractDeps()
}

func (inPlatformRelease indexPlatformRelease) extractDeps() cores.ToolDependencies {
	ret := make(cores.ToolDependencies, len(inPlatformRelease.ToolDependencies))
	for i, dep := range inPlatformRelease.ToolDependencies {
		ret[i] = &cores.ToolDependency{
			ToolName:     dep.Name,
			ToolVersion:  dep.Version,
			ToolPackager: dep.Packager,
		}
	}
	return ret
}

func (inPlatformRelease indexPlatformRelease) extractBoardsManifest() []*cores.BoardManifest {
	boards := make([]*cores.BoardManifest, len(inPlatformRelease.Boards))
	for i, board := range inPlatformRelease.Boards {
		boards[i] = &cores.BoardManifest{Name: board.Name}
	}
	return boards
}

func (inToolRelease indexToolRelease) extractToolIn(outPackage *cores.Package) {
	outTool := outPackage.GetOrCreateTool(inToolRelease.Name)

	outToolRelease := outTool.GetOrCreateRelease(inToolRelease.Version)
	outToolRelease.Flavours = inToolRelease.extractFlavours()
}

// extractFlavours extracts a map[OS]Flavour object from an indexToolRelease entry.
func (inToolRelease indexToolRelease) extractFlavours() []*cores.Flavour {
	ret := make([]*cores.Flavour, len(inToolRelease.Systems))
	for i, flavour := range inToolRelease.Systems {
		size, _ := flavour.Size.Int64()
		ret[i] = &cores.Flavour{
			OS: flavour.OS,
			Resource: &resources.DownloadResource{
				ArchiveFileName: flavour.ArchiveFileName,
				Checksum:        flavour.Checksum,
				Size:            size,
				URL:             flavour.URL,
				CachePath:       "packages",
			},
		}
	}
	return ret
}

// LoadIndex reads a package_index.json from a file and returns the corresponding Index structure.
func LoadIndex(jsonIndexPath string) (*Index, error) {
	buff, err := ioutil.ReadFile(jsonIndexPath)
	if err != nil {
		return nil, err
	}
	//fmt.Println(string(buff))
	var index Index
	err = json.Unmarshal(buff, &index)
	if err != nil {
		return nil, err
	}

	return &index, nil
}
