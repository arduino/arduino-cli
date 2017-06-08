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
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/arduino/arduino-cli/builder_client_helpers"
)

// Index represents the content of a library_index.json file
type Index struct {
	Libraries []IndexRelease `json:"libraries"`
}

// IndexRelease is an entry of a library_index.json
type IndexRelease struct {
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

// LoadLibrariesIndexFromFile reads a library_index.json from a file and returns
// the corresponding LibrariesIndex structure.
func LoadLibrariesIndexFromFile(libFile string) (*Index, error) {
	libBuff, err := ioutil.ReadFile(libFile)
	if err != nil {
		return nil, err
	}

	var index Index
	if err := json.Unmarshal(libBuff, &index); err != nil {
		return nil, err
	}

	return &index, nil
}

func DownloadLibrariesFileBase(saveToPath string) error {
	req, err := http.NewRequest("GET", "http://downloads.arduino.cc/libraries/library_index.json", nil)
	if err != nil {
		return err
	}

	client := http.DefaultClient
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(saveToPath, content, 0666)
	if err != nil {
		return err
	}
	return nil
}

func UpdateLocalFile(localFilePath string, libraries []Library) error {
	content, err := ioutil.ReadFile(localFilePath)
	if err != nil {
		return err
	}
	var index IndexRelease
	err = json.Unmarshal(content, &index)
	if err != nil {
		return err
	}

}

// DownloadLibrariesIndex downloads from arduino repository the libraries index.
func (indexLib *IndexRelease) DownloadLibrariesIndex() (*[]Library, error) {
	client := builderClient.New(nil)
	response, err := client.ListLibraries(nil, builderClient.ListLibrariesPath(), nil, nil, nil)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode != 200 {
		return nil, fmt.Errorf("Status Code not OK, request failed")
	}

	jsonResult, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	var result []Library
	err = json.Unmarshal(jsonResult, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// UpdateLocalFile updates local cache file using an Index (most of times downloaded from arduino archives)
func (indexLib *IndexRelease) UpdateLocalFile(localFilePath string, index *Index) error {
	if localFilePath == "" || index == nil {
		return fmt.Errorf("Invalid arguments, please specify valid localFilePath and index")
	}
	content, err := json.Marshal(index)

	if err != nil {
		return err
	}

	ioutil.WriteFile(localFilePath, content, 0666)
	return nil
}

// ExtractRelease create a new Release with the information contained
// in this index element
func (indexLib *IndexRelease) ExtractRelease() *Release {
	return &Release{
		Version:         indexLib.Version,
		URL:             indexLib.URL,
		ArchiveFileName: indexLib.ArchiveFileName,
		Size:            indexLib.Size,
		Checksum:        indexLib.Checksum,
	}
}

// ExtractLibrary create a new Library with the information contained
// in this index element
func (indexLib *IndexRelease) ExtractLibrary() *Library {
	release := indexLib.ExtractRelease()
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
		Releases:      []*Release{release},
		Latest:        release,
	}
}
