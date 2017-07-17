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
 * Copyright 2017 BCMI LABS SA (http://www.arduino.cc/)
 */

package cores

import (
	"encoding/json"
	"io/ioutil"
	"path/filepath"

	"github.com/bcmi-labs/arduino-cli/common"
)

// Index represents Cores and Tools struct as seen from package_index.json file.
type Index struct {
	Packages []*indexPackage `json:"packages"`
}

//IndexPath returns the path of the index file for libraries.
func IndexPath() (string, error) {
	baseFolder, err := common.GetDefaultArduinoFolder()
	if err != nil {
		return "", err
	}
	return filepath.Join(baseFolder, "package_index.json"), nil
}

//indexPackage represents a single entry from package_index.json file.
type indexPackage struct {
	Name       string              `json:"name"`
	Maintainer string              `json:"maintainer"`
	WebsiteURL string              `json:"websiteUrl"`
	Email      string              `json:"email"`
	Platforms  []*indexCoreRelease `json:"platforms"`
	//Tools      []*indexToolRelease `json:"tools"`
} //TODO: help : { online : "address" } is not in all package managers

// indexCoreRelease represents a single Core Platform from package_index.json file.
type indexCoreRelease struct {
	Name            string              `json:"name"`
	Architecture    string              `json:"architecture"`
	Version         string              `json:"version"`
	Category        string              `json:"category"`
	URL             string              `json:"url"`
	ArchiveFileName string              `json:"archiveFileName"`
	Checksum        string              `json:"checksum"`
	Size            int64               `json:"size"`
	Boards          []indexBoardRelease `json:"boards"`
}

type indexBoardRelease struct {
	Name string
}

func (packag indexPackage) extractPackage() (pm *Package) {
	pm = &Package{
		Name:       packag.Name,
		Maintainer: packag.Maintainer,
		WebsiteURL: packag.WebsiteURL,
		Email:      packag.Email,
		Cores:      make(map[string]*Core, len(packag.Platforms)),
	}
	for _, core := range packag.Platforms {
		pm.AddCore(core)
	}
	return
}

func (release *indexCoreRelease) extractCore() *Core {
	return &Core{
		Name:         release.Name,
		Architecture: release.Architecture,
		Category:     release.Category,
		Releases:     map[string]*Release{release.Version: release.extractRelease()},
	}
}

func (release *indexCoreRelease) extractRelease() *Release {
	return &Release{
		Version:         release.Version,
		ArchiveFileName: release.ArchiveFileName,
		Checksum:        release.Checksum,
		Size:            release.Size,
		Boards:          release.extractBoards(),
	}
}

func (release *indexCoreRelease) extractBoards() []string {
	boards := make([]string, 0, len(release.Boards))
	for i, board := range release.Boards {
		boards[i] = board.Name
	}
	return boards
}

// LoadPackagesIndex reads a package_index.json from a file and returns
// the corresponding Index structure.
func LoadPackagesIndex() (common.Index, error) {
	libFile, err := IndexPath()
	if err != nil {
		return nil, err
	}

	buff, err := ioutil.ReadFile(libFile)
	if err != nil {
		return nil, err
	}

	var index Index
	err = json.Unmarshal(buff, &index)
	if err != nil {
		return nil, err
	}

	return index, nil
}
