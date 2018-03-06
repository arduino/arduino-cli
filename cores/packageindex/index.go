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
 *it
 * Copyright 2017 ARDUINO AG (http://www.arduino.cc/)
 */

package packageindex

import (
	"encoding/json"
	"io/ioutil"

	"github.com/bcmi-labs/arduino-cli/common/releases"

	"github.com/bcmi-labs/arduino-cli/common"
	"github.com/bcmi-labs/arduino-cli/configs"
	"github.com/bcmi-labs/arduino-cli/cores"
)

// packageIndexURL contains the index URL for core packages.
const packageIndexURL = "https://downloads.arduino.cc/packages/package_index.json"

// coreIndexPath returns the path of the index file for libraries.
var coreIndexPath = configs.IndexPath("package_index.json")

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
	Size             int64                 `json:"size,required,string"`
	BoardsNames      []indexBoardName      `json:"boards"`
	Help             indexHelp             `json:"help,omitempty"`
	ToolDependencies []indexToolDependency `json:"toolsDependencies, required"`
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
	OS              string `json:"host,required"`
	URL             string `json:"url,required"`
	ArchiveFileName string `json:"archiveFileName,required"`
	Size            int64  `json:"size,required,string"`
	Checksum        string `json:"checksum,required"`
}

// indexBoardName represents a single Board as written in package_index.json file.
type indexBoardName struct {
	Name string
}

type indexHelp struct {
	Online string `json:"online,omitempty"`
}

// CreateStatusContext creates a status context from index data.
func (index Index) CreateStatusContext() *cores.Packages {
	outPackages := cores.NewPackages()

	for _, inPackage := range index.Packages {
		inPackage.extractPackageIn(outPackages)
	}
	return outPackages
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

	outPlatformRelease := outPlatform.GetOrCreateRelease(inPlatformRelease.Version)
	outPlatformRelease.Resource = &releases.DownloadResource{
		ArchiveFileName: inPlatformRelease.ArchiveFileName,
		Checksum:        inPlatformRelease.Checksum,
		Size:            inPlatformRelease.Size,
		URL:             inPlatformRelease.URL,
		CachePath:       "packages",
	}
	outPlatformRelease.BoardNames = inPlatformRelease.extractBoards()
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

func (inPlatformRelease indexPlatformRelease) extractBoards() []string {
	boards := make([]string, len(inPlatformRelease.BoardsNames))
	for i, board := range inPlatformRelease.BoardsNames {
		boards[i] = board.Name
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
		ret[i] = &cores.Flavour{
			OS: flavour.OS,
			Resource: &releases.DownloadResource{
				ArchiveFileName: flavour.ArchiveFileName,
				Checksum:        flavour.Checksum,
				Size:            flavour.Size,
				URL:             flavour.URL,
				CachePath:       "packages",
			},
		}
	}
	return ret
}

// LoadIndex reads a package_index.json from a file and returns
// the corresponding Index structure.
func LoadIndex(index *Index) error {
	coreFile, err := coreIndexPath.Get()
	if err != nil {
		return err
	}

	buff, err := ioutil.ReadFile(coreFile)
	if err != nil {
		return err
	}
	//fmt.Println(string(buff))
	err = json.Unmarshal(buff, index)
	if err != nil {
		return err
	}

	return nil
}

// DownloadDefaultPackageIndexFile downloads the core packages index file from arduino repository.
func DownloadDefaultPackageIndexFile() error {
	return common.DownloadIndex(coreIndexPath, packageIndexURL)
}
