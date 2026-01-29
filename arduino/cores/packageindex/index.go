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

package packageindex

import (
	"encoding/json"
	"fmt"

	"github.com/arduino/arduino-cli/arduino/cores"
	"github.com/arduino/arduino-cli/arduino/resources"
	"github.com/arduino/arduino-cli/arduino/security"
	"github.com/arduino/arduino-cli/i18n"
	"github.com/arduino/go-paths-helper"
	easyjson "github.com/mailru/easyjson"
	"github.com/sirupsen/logrus"
	semver "go.bug.st/relaxed-semver"
)

// Index represents Cores and Tools struct as seen from package_index.json file.
//
//easyjson:json
type Index struct {
	Packages  []*indexPackage `json:"packages"`
	IsTrusted bool
}

// indexPackage represents a single entry from package_index.json file.
//
//easyjson:json
type indexPackage struct {
	Name       string                  `json:"name"`
	Maintainer string                  `json:"maintainer"`
	WebsiteURL string                  `json:"websiteUrl"`
	URL        string                  `json:"Url"`
	Email      string                  `json:"email"`
	Platforms  []*indexPlatformRelease `json:"platforms"`
	Tools      []*indexToolRelease     `json:"tools"`
	Help       indexHelp               `json:"help,omitempty"`
}

// indexPlatformRelease represents a single Core Platform from package_index.json file.
//
//easyjson:json
type indexPlatformRelease struct {
	Name                  string                     `json:"name"`
	Architecture          string                     `json:"architecture"`
	Version               *semver.Version            `json:"version"`
	Deprecated            bool                       `json:"deprecated"`
	Category              string                     `json:"category"`
	URL                   string                     `json:"url"`
	ArchiveFileName       string                     `json:"archiveFileName"`
	Checksum              string                     `json:"checksum"`
	Size                  json.Number                `json:"size"`
	Boards                []indexBoard               `json:"boards"`
	Help                  indexHelp                  `json:"help,omitempty"`
	ToolDependencies      []indexToolDependency      `json:"toolsDependencies"`
	DiscoveryDependencies []indexDiscoveryDependency `json:"discoveryDependencies"`
	MonitorDependencies   []indexMonitorDependency   `json:"monitorDependencies"`
}

// indexToolDependency represents a single dependency of a core from a tool.
//
//easyjson:json
type indexToolDependency struct {
	Packager string                 `json:"packager"`
	Name     string                 `json:"name"`
	Version  *semver.RelaxedVersion `json:"version"`
}

// indexDiscoveryDependency represents a single dependency of a core from a pluggable discovery tool.
//
//easyjson:json
type indexDiscoveryDependency struct {
	Packager string `json:"packager"`
	Name     string `json:"name"`
}

// indexMonitorDependency represents a single dependency of a core from a pluggable monitor tool.
//
//easyjson:json
type indexMonitorDependency struct {
	Packager string `json:"packager"`
	Name     string `json:"name"`
}

// indexToolRelease represents a single Tool from package_index.json file.
//
//easyjson:json
type indexToolRelease struct {
	Name    string                    `json:"name"`
	Version *semver.RelaxedVersion    `json:"version"`
	Systems []indexToolReleaseFlavour `json:"systems"`
}

// indexToolReleaseFlavour represents a single tool flavor in the package_index.json file.
//
//easyjson:json
type indexToolReleaseFlavour struct {
	OS              string      `json:"host"`
	URL             string      `json:"url"`
	ArchiveFileName string      `json:"archiveFileName"`
	Size            json.Number `json:"size"`
	Checksum        string      `json:"checksum"`
}

// indexBoard represents a single Board as written in package_index.json file.
//
//easyjson:json
type indexBoard struct {
	Name string         `json:"name"`
	ID   []indexBoardID `json:"id,omitempty"`
}

// indexBoardID represents the ID of a single board. i.e. uno, yun, diecimila, micro and the likes
//
//easyjson:json
type indexBoardID struct {
	USB string `json:"usb"`
}

// indexHelp represents the help URL
//
//easyjson:json
type indexHelp struct {
	Online string `json:"online,omitempty"`
}

var tr = i18n.Tr

// MergeIntoPackages converts the Index data into a cores.Packages and merge them
// with the existing contents of the cores.Packages passed as parameter.
func (index Index) MergeIntoPackages(outPackages cores.Packages) {
	for _, inPackage := range index.Packages {
		inPackage.extractPackageIn(outPackages, index.IsTrusted)
	}
}

// IndexFromPlatformRelease creates an Index that contains a single indexPackage
// which in turn contains a single indexPlatformRelease converted from the one
// passed as argument
func IndexFromPlatformRelease(pr *cores.PlatformRelease) Index {
	boards := []indexBoard{}
	for _, manifest := range pr.BoardsManifest {
		board := indexBoard{
			Name: manifest.Name,
		}
		for _, id := range manifest.ID {
			if id.USB != "" {
				board.ID = []indexBoardID{{USB: id.USB}}
			}
		}
		boards = append(boards, board)
	}

	tools := []indexToolDependency{}
	for _, t := range pr.ToolDependencies {
		tools = append(tools, indexToolDependency{
			Packager: t.ToolPackager,
			Name:     t.ToolName,
			Version:  t.ToolVersion,
		})
	}

	discoveries := []indexDiscoveryDependency{}
	for _, d := range pr.DiscoveryDependencies {
		discoveries = append(discoveries, indexDiscoveryDependency{
			Packager: d.Packager,
			Name:     d.Name,
		})
	}

	monitors := []indexMonitorDependency{}
	for _, m := range pr.MonitorDependencies {
		monitors = append(monitors, indexMonitorDependency{
			Packager: m.Packager,
			Name:     m.Name,
		})
	}

	packageTools := []*indexToolRelease{}
	for name, tool := range pr.Platform.Package.Tools {
		for _, toolRelease := range tool.Releases {
			flavours := []indexToolReleaseFlavour{}
			for _, flavour := range toolRelease.Flavors {
				flavours = append(flavours, indexToolReleaseFlavour{
					OS:              flavour.OS,
					URL:             flavour.Resource.URL,
					ArchiveFileName: flavour.Resource.ArchiveFileName,
					Size:            json.Number(fmt.Sprintf("%d", flavour.Resource.Size)),
					Checksum:        flavour.Resource.Checksum,
				})
			}
			packageTools = append(packageTools, &indexToolRelease{
				Name:    name,
				Version: toolRelease.Version,
				Systems: flavours,
			})
		}
	}

	return Index{
		IsTrusted: pr.IsTrusted,
		Packages: []*indexPackage{
			{
				Name:       pr.Platform.Package.Name,
				Maintainer: pr.Platform.Package.Maintainer,
				WebsiteURL: pr.Platform.Package.WebsiteURL,
				URL:        pr.Platform.Package.URL,
				Email:      pr.Platform.Package.Email,
				Platforms: []*indexPlatformRelease{{
					Name:                  pr.Platform.Name,
					Architecture:          pr.Platform.Architecture,
					Version:               pr.Version,
					Deprecated:            pr.Platform.Deprecated,
					Category:              pr.Platform.Category,
					URL:                   pr.Resource.URL,
					ArchiveFileName:       pr.Resource.ArchiveFileName,
					Checksum:              pr.Resource.Checksum,
					Size:                  json.Number(fmt.Sprintf("%d", pr.Resource.Size)),
					Boards:                boards,
					Help:                  indexHelp{Online: pr.Help.Online},
					ToolDependencies:      tools,
					DiscoveryDependencies: discoveries,
					MonitorDependencies:   monitors,
				}},
				Tools: packageTools,
				Help:  indexHelp{Online: pr.Platform.Package.Help.Online},
			},
		},
	}
}

func (inPackage indexPackage) extractPackageIn(outPackages cores.Packages, trusted bool) {
	outPackage := outPackages.GetOrCreatePackage(inPackage.Name)
	outPackage.Maintainer = inPackage.Maintainer
	outPackage.WebsiteURL = inPackage.WebsiteURL
	outPackage.URL = inPackage.URL
	outPackage.Email = inPackage.Email
	outPackage.Help = cores.PackageHelp{Online: inPackage.Help.Online}

	for _, inTool := range inPackage.Tools {
		inTool.extractToolIn(outPackage)
	}

	for _, inPlatform := range inPackage.Platforms {
		inPlatform.extractPlatformIn(outPackage, trusted)
	}
}

func (inPlatformRelease indexPlatformRelease) extractPlatformIn(outPackage *cores.Package, trusted bool) error {
	outPlatform := outPackage.GetOrCreatePlatform(inPlatformRelease.Architecture)
	// FIXME: shall we use the Name and Category of the latest release? or maybe move Name and Category in PlatformRelease?
	outPlatform.Name = inPlatformRelease.Name
	outPlatform.Category = inPlatformRelease.Category
	// If the Platform is installed before deprecation installed.json file does not include "deprecated" field.
	// The installed.json is read during loading phase of an installed Platform, if the deprecated field is not found
	// the package_index.json field would be overwritten and the deprecation info would be lost.
	// This check prevents that behaviour.
	if !outPlatform.Deprecated {
		outPlatform.Deprecated = inPlatformRelease.Deprecated
	}

	size, err := inPlatformRelease.Size.Int64()
	if err != nil {
		return fmt.Errorf("invalid platform archive size: %s", err)
	}
	outPlatformRelease := outPlatform.GetOrCreateRelease(inPlatformRelease.Version)
	outPlatformRelease.IsTrusted = trusted
	outPlatformRelease.Resource = &resources.DownloadResource{
		ArchiveFileName: inPlatformRelease.ArchiveFileName,
		Checksum:        inPlatformRelease.Checksum,
		Size:            size,
		URL:             inPlatformRelease.URL,
		CachePath:       "packages",
	}
	outPlatformRelease.Help = cores.PlatformReleaseHelp{Online: inPlatformRelease.Help.Online}
	outPlatformRelease.BoardsManifest = inPlatformRelease.extractBoardsManifest()
	outPlatformRelease.ToolDependencies = inPlatformRelease.extractToolDependencies()
	outPlatformRelease.DiscoveryDependencies = inPlatformRelease.extractDiscoveryDependencies()
	outPlatformRelease.MonitorDependencies = inPlatformRelease.extractMonitorDependencies()
	return nil
}

func (inPlatformRelease indexPlatformRelease) extractToolDependencies() cores.ToolDependencies {
	res := make(cores.ToolDependencies, len(inPlatformRelease.ToolDependencies))
	for i, tool := range inPlatformRelease.ToolDependencies {
		res[i] = &cores.ToolDependency{
			ToolName:     tool.Name,
			ToolVersion:  tool.Version,
			ToolPackager: tool.Packager,
		}
	}
	return res
}

func (inPlatformRelease indexPlatformRelease) extractDiscoveryDependencies() cores.DiscoveryDependencies {
	res := make(cores.DiscoveryDependencies, len(inPlatformRelease.DiscoveryDependencies))
	for i, discovery := range inPlatformRelease.DiscoveryDependencies {
		res[i] = &cores.DiscoveryDependency{
			Name:     discovery.Name,
			Packager: discovery.Packager,
		}
	}
	return res
}

func (inPlatformRelease indexPlatformRelease) extractMonitorDependencies() cores.MonitorDependencies {
	res := make(cores.MonitorDependencies, len(inPlatformRelease.MonitorDependencies))
	for i, discovery := range inPlatformRelease.MonitorDependencies {
		res[i] = &cores.MonitorDependency{
			Name:     discovery.Name,
			Packager: discovery.Packager,
		}
	}
	return res
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
	var index Index
	err = easyjson.Unmarshal(buff, &index)
	if err != nil {
		return nil, err
	}

	jsonSignatureFile := jsonIndexFile.Parent().Join(jsonIndexFile.Base() + ".sig")
	trusted, _, err := security.VerifyArduinoDetachedSignature(jsonIndexFile, jsonSignatureFile)
	if err != nil {
		logrus.
			WithField("index", jsonIndexFile).
			WithField("signatureFile", jsonSignatureFile).
			WithError(err).Infof("Checking signature")
	} else {
		logrus.
			WithField("index", jsonIndexFile).
			WithField("signatureFile", jsonSignatureFile).
			WithField("trusted", trusted).Infof("Checking signature")
		index.IsTrusted = trusted
	}
	return &index, nil
}

// LoadIndexNoSign reads a package_index.json from a file and returns the corresponding Index structure.
func LoadIndexNoSign(jsonIndexFile *paths.Path) (*Index, error) {
	buff, err := jsonIndexFile.ReadFile()
	if err != nil {
		return nil, err
	}
	var index Index
	err = easyjson.Unmarshal(buff, &index)
	if err != nil {
		return nil, err
	}

	index.IsTrusted = true

	return &index, nil
}
