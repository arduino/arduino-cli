/*
 * This file is part of arduino-cli.
 *
 * Copyright 2018 ARDUINO SA (http://www.arduino.cc/)
 *
 * This software is released under the GNU General Public License version 3,
 * which covers the main part of arduino-cli.
 * The terms of this license can be found at:
 * https://www.gnu.org/licenses/gpl-3.0.en.html
 *
 * You can be released from the requirements of the above licenses by purchasing
 * a commercial license. Buying such a license is mandatory if you want to modify or
 * otherwise use the software for commercial activities involving the Arduino
 * software without disclosing the source code of your own applications. To purchase
 * a commercial license, send an email to license@arduino.cc.
 */

package packageindex

import (
	"encoding/json"
	"fmt"

	"github.com/arduino/arduino-cli/arduino/cores"
	"github.com/arduino/arduino-cli/arduino/resources"
	"github.com/arduino/go-paths-helper"
	"go.bug.st/relaxed-semver"
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
	Version          *semver.Version       `json:"version,required"`
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
	Packager string                 `json:"packager,required"`
	Name     string                 `json:"name,required"`
	Version  *semver.RelaxedVersion `json:"version,required"`
}

// indexToolRelease represents a single Tool from package_index.json file.
type indexToolRelease struct {
	Name    string                    `json:"name,required"`
	Version *semver.RelaxedVersion    `json:"version,required"`
	Systems []indexToolReleaseFlavour `json:"systems,required"`
}

// indexToolReleaseFlavour represents a single tool flavor in the package_index.json file.
type indexToolReleaseFlavour struct {
	OS              string      `json:"host,required"`
	URL             string      `json:"url,required"`
	ArchiveFileName string      `json:"archiveFileName,required"`
	Size            json.Number `json:"size,required"`
	Checksum        string      `json:"checksum,required"`
}

// indexBoard represents a single Board as written in package_index.json file.
type indexBoard struct {
	Name string         `json:"name"`
	ID   []indexBoardID `json:"id"`
}

type indexBoardID struct {
	USB string `json:"usb"`
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

func (inPlatformRelease indexPlatformRelease) extractPlatformIn(outPackage *cores.Package) error {
	outPlatform := outPackage.GetOrCreatePlatform(inPlatformRelease.Architecture)
	// FIXME: shall we use the Name and Category of the latest release? or maybe move Name and Category in PlatformRelease?
	outPlatform.Name = inPlatformRelease.Name
	outPlatform.Category = inPlatformRelease.Category

	size, err := inPlatformRelease.Size.Int64()
	if err != nil {
		return fmt.Errorf("invalid platform archive size: %s", err)
	}
	outPlatformRelease, err := outPlatform.GetOrCreateRelease(inPlatformRelease.Version)
	if err != nil {
		return fmt.Errorf("creating release: %s", err)
	}
	outPlatformRelease.Resource = &resources.DownloadResource{
		ArchiveFileName: inPlatformRelease.ArchiveFileName,
		Checksum:        inPlatformRelease.Checksum,
		Size:            size,
		URL:             inPlatformRelease.URL,
		CachePath:       "packages",
	}
	outPlatformRelease.BoardsManifest = inPlatformRelease.extractBoardsManifest()
	if deps, err := inPlatformRelease.extractDeps(); err == nil {
		outPlatformRelease.Dependencies = deps
	} else {
		return fmt.Errorf("invalid tool dependencies: %s", err)
	}
	return nil
}

func (inPlatformRelease indexPlatformRelease) extractDeps() (cores.ToolDependencies, error) {
	ret := make(cores.ToolDependencies, len(inPlatformRelease.ToolDependencies))
	for i, dep := range inPlatformRelease.ToolDependencies {
		ret[i] = &cores.ToolDependency{
			ToolName:     dep.Name,
			ToolVersion:  dep.Version,
			ToolPackager: dep.Packager,
		}
	}
	return ret, nil
}

func (inPlatformRelease indexPlatformRelease) extractBoardsManifest() []*cores.BoardManifest {
	boards := make([]*cores.BoardManifest, len(inPlatformRelease.Boards))
	for i, board := range inPlatformRelease.Boards {
		manifest := &cores.BoardManifest{Name: board.Name}
		for _, id := range board.ID {
			if id.USB != "" {
				manifest.ID = append(manifest.ID, &cores.BoardManifestID{USB: id.USB})
			}
		}
		boards[i] = manifest
	}
	return boards
}

func (inToolRelease indexToolRelease) extractToolIn(outPackage *cores.Package) {
	outTool := outPackage.GetOrCreateTool(inToolRelease.Name)

	outToolRelease := outTool.GetOrCreateRelease(inToolRelease.Version)
	outToolRelease.Flavors = inToolRelease.extractFlavours()
}

// extractFlavours extracts a map[OS]Flavor object from an indexToolRelease entry.
func (inToolRelease indexToolRelease) extractFlavours() []*cores.Flavor {
	ret := make([]*cores.Flavor, len(inToolRelease.Systems))
	for i, flavour := range inToolRelease.Systems {
		size, _ := flavour.Size.Int64()
		ret[i] = &cores.Flavor{
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
func LoadIndex(jsonIndexFile *paths.Path) (*Index, error) {
	buff, err := jsonIndexFile.ReadFile()
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
