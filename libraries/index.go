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

package libraries

import (
	"encoding/json"
	"io/ioutil"
	"path/filepath"

	"github.com/bcmi-labs/arduino-cli/common"
)

// Index represents the content of a library_index.json file
type Index struct {
	Libraries []indexRelease `json:"libraries"`
}

// indexRelease is an entry of a library_index.json
type indexRelease struct {
	Name            string   `json:"name"`
	Version         string   `json:"version"`
	Author          string   `json:"author"`
	Maintainer      string   `json:"maintainer"`
	Sentence        string   `json:"sentence"`
	Paragraph       string   `json:"paragraph"`
	Website         string   `json:"website"`
	Category        string   `json:"category"`
	Architectures   []string `json:"architectures"`
	Types           []string `json:"types"`
	URL             string   `json:"url"`
	ArchiveFileName string   `json:"archiveFileName"`
	Size            int      `json:"size"`
	Checksum        string   `json:"checksum"`
}

//IndexPath returns the path of the index file for libraries.
func IndexPath() (string, error) {
	baseFolder, err := common.GetDefaultArduinoFolder()
	if err != nil {
		return "", err
	}
	return filepath.Join(baseFolder, "library_index.json"), nil
}

// LoadLibrariesIndex reads a library_index.json from a file and returns
// the corresponding LibrariesIndex structure.
func LoadLibrariesIndex() (*Index, error) {
	libFile, err := IndexPath()
	if err != nil {
		return nil, err
	}

	libBuff, err := ioutil.ReadFile(libFile)
	if err != nil {
		return nil, err
	}

	var index Index
	err = json.Unmarshal(libBuff, &index)
	if err != nil {
		return nil, err
	}

	return &index, nil
}

// extractRelease create a new Release with the information contained
// in this index element
func (indexLib *indexRelease) extractRelease() *Release {
	return &Release{
		Version:         indexLib.Version,
		URL:             indexLib.URL,
		ArchiveFileName: indexLib.ArchiveFileName,
		Size:            indexLib.Size,
		Checksum:        indexLib.Checksum,
	}

}

// extractLibrary create a new Library with the information contained
// in this index element.
func (indexLib *indexRelease) extractLibrary() *Library {
	release := indexLib.extractRelease()
	return &Library{
		Name:          indexLib.Name,
		Author:        indexLib.Author,
		Maintainer:    indexLib.Maintainer,
		Sentence:      indexLib.Sentence,
		Paragraph:     indexLib.Paragraph,
		Website:       indexLib.Website,
		Category:      indexLib.Category,
		Architectures: indexLib.Architectures,
		Types:         indexLib.Types,
		Releases:      map[string]*Release{release.Version: release},
	}
}
