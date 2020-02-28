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

package librariesindex

import (
	"encoding/json"
	"fmt"

	"github.com/arduino/arduino-cli/arduino/resources"
	"github.com/arduino/go-paths-helper"
	semver "go.bug.st/relaxed-semver"
)

type indexJSON struct {
	Libraries []indexRelease `json:"libraries"`
}

type indexRelease struct {
	Name             string             `json:"name,required"`
	Version          *semver.Version    `json:"version,required"`
	Author           string             `json:"author"`
	Maintainer       string             `json:"maintainer"`
	Sentence         string             `json:"sentence"`
	Paragraph        string             `json:"paragraph"`
	Website          string             `json:"website"`
	Category         string             `json:"category"`
	Architectures    []string           `json:"architectures"`
	Types            []string           `json:"types"`
	URL              string             `json:"url"`
	ArchiveFileName  string             `json:"archiveFileName"`
	Size             int64              `json:"size"`
	Checksum         string             `json:"checksum"`
	Dependencies     []*indexDependency `json:"dependencies,omitempty"`
	License          string             `json:"license"`
	ProvidesIncludes []string           `json:"providesIncludes"`
}

type indexDependency struct {
	Name    string `json:"name"`
	Version string `json:"version,omitempty"`
}

// LoadIndex reads a library_index.json and create the corresponding Index
func LoadIndex(indexFile *paths.Path) (*Index, error) {
	buff, err := indexFile.ReadFile()
	if err != nil {
		return nil, fmt.Errorf("reading library_index.json: %s", err)
	}

	var i indexJSON
	err = json.Unmarshal(buff, &i)
	if err != nil {
		return nil, fmt.Errorf("parsing library_index.json: %s", err)
	}

	return i.extractIndex()
}

func (i indexJSON) extractIndex() (*Index, error) {
	index := &Index{
		Libraries: map[string]*Library{},
	}
	for _, indexLib := range i.Libraries {
		indexLib.extractLibraryIn(index)
	}
	return index, nil
}

func (indexLib *indexRelease) extractLibraryIn(index *Index) {
	library, exist := index.Libraries[indexLib.Name]
	if !exist {
		library = &Library{
			Name:     indexLib.Name,
			Releases: map[string]*Release{},
		}
		index.Libraries[indexLib.Name] = library
	}
	indexLib.extractReleaseIn(library)
}

func (indexLib *indexRelease) extractReleaseIn(library *Library) {
	release := &Release{
		Version:       indexLib.Version,
		Author:        indexLib.Author,
		Maintainer:    indexLib.Maintainer,
		Sentence:      indexLib.Sentence,
		Paragraph:     indexLib.Paragraph,
		Website:       indexLib.Website,
		Category:      indexLib.Category,
		Architectures: indexLib.Architectures,
		Types:         indexLib.Types,
		Resource: &resources.DownloadResource{
			URL:             indexLib.URL,
			ArchiveFileName: indexLib.ArchiveFileName,
			Size:            indexLib.Size,
			Checksum:        indexLib.Checksum,
			CachePath:       "libraries",
		},
		Library:          library,
		Dependencies:     indexLib.extractDependencies(),
		License:          indexLib.License,
		ProvidesIncludes: indexLib.ProvidesIncludes,
	}
	library.Releases[indexLib.Version.String()] = release
	if library.Latest == nil || library.Latest.Version.LessThan(release.Version) {
		library.Latest = release
	}
}

func (indexLib *indexRelease) extractDependencies() []semver.Dependency {
	res := []semver.Dependency{}
	if indexLib.Dependencies == nil || len(indexLib.Dependencies) == 0 {
		return res
	}
	for _, indexDep := range indexLib.Dependencies {
		res = append(res, indexDep.extractDependency())
	}
	return res
}

func (indexDep *indexDependency) extractDependency() *Dependency {
	var constraint semver.Constraint
	if c, err := semver.ParseConstraint(indexDep.Version); err == nil {
		constraint = c
	} else {
		// FIXME: report invalid constraint
	}
	return &Dependency{
		Name:              indexDep.Name,
		VersionConstraint: constraint,
	}
}
