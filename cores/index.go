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

package cores

import (
	"encoding/json"
	"io/ioutil"

	"github.com/bcmi-labs/arduino-cli/configs"
)

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

func (pack indexPackage) extractPackage() *Package {
	p := &Package{
		Name:       pack.Name,
		Maintainer: pack.Maintainer,
		WebsiteURL: pack.WebsiteURL,
		Email:      pack.Email,
		Plaftorms:  map[string]*Platform{},
		Tools:      map[string]*Tool{},
	}

	for _, tool := range pack.Tools {
		name := tool.Name
		if p.Tools[name] == nil {
			p.Tools[name] = tool.extractTool()
		} else {
			p.Tools[name].Releases[tool.Version] = tool.extractRelease()
		}
	}

	for _, platform := range pack.Platforms {
		name := platform.Architecture
		if p.Plaftorms[name] == nil {
			p.Plaftorms[name] = platform.extractPlatform()
		} else {
			release := platform.extractRelease()
			p.Plaftorms[name].Releases[release.Version] = release
		}
	}

	return p
}

func (release indexPlatformRelease) extractPlatform() *Platform {
	return &Platform{
		Name:         release.Name,
		Architecture: release.Architecture,
		Category:     release.Category,
		Releases:     map[string]*PlatformRelease{release.Version: release.extractRelease()},
	}
}

func (release indexPlatformRelease) extractRelease() *PlatformRelease {
	return &PlatformRelease{
		Version:         release.Version,
		ArchiveFileName: release.ArchiveFileName,
		Checksum:        release.Checksum,
		Size:            release.Size,
		URL:             release.URL,
		Boards:          release.extractBoards(),
		Dependencies:    release.extractDeps(),
	}
}

func (release indexPlatformRelease) extractDeps() ToolDependencies {
	ret := make(ToolDependencies, len(release.ToolDependencies))
	for i, dep := range release.ToolDependencies {
		ret[i] = &ToolDependency{
			ToolName:     dep.Name,
			ToolVersion:  dep.Version,
			ToolPackager: dep.Packager,
		}
	}
	return ret
}

func (release indexPlatformRelease) extractBoards() []string {
	boards := make([]string, len(release.BoardsNames))
	for i, board := range release.BoardsNames {
		boards[i] = board.Name
	}
	return boards
}

// extractTool extracts a Tool object from an indexToolRelease entry.
func (itr indexToolRelease) extractTool() *Tool {
	return &Tool{
		Name: itr.Name,
		Releases: map[string]*ToolRelease{
			itr.Version: itr.extractRelease(),
		},
	}
}

// extractRelease extracts a ToolRelease object from an indexToolRelease entry.
func (itr indexToolRelease) extractRelease() *ToolRelease {
	return &ToolRelease{
		Version:  itr.Version,
		Flavours: itr.extractFlavours(),
	}
}

// extractFlavours extracts a map[OS]Flavour object from an indexToolRelease entry.
func (itr indexToolRelease) extractFlavours() []*Flavour {
	ret := make([]*Flavour, len(itr.Systems))
	for i, flavour := range itr.Systems {
		ret[i] = &Flavour{
			OS:              flavour.OS,
			ArchiveFileName: flavour.ArchiveFileName,
			Checksum:        flavour.Checksum,
			Size:            flavour.Size,
			URL:             flavour.URL,
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
