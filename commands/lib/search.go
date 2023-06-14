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

package lib

import (
	"context"
	"sort"
	"strings"

	"github.com/arduino/arduino-cli/arduino"
	"github.com/arduino/arduino-cli/arduino/libraries/librariesindex"
	"github.com/arduino/arduino-cli/arduino/libraries/librariesmanager"
	"github.com/arduino/arduino-cli/arduino/utils"
	"github.com/arduino/arduino-cli/commands"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	semver "go.bug.st/relaxed-semver"
)

// LibrarySearch FIXMEDOC
func LibrarySearch(ctx context.Context, req *rpc.LibrarySearchRequest) (*rpc.LibrarySearchResponse, error) {
	lm := commands.GetLibraryManager(req)
	if lm == nil {
		return nil, &arduino.InvalidInstanceError{}
	}
	return searchLibrary(req, lm), nil
}

func searchLibrary(req *rpc.LibrarySearchRequest, lm *librariesmanager.LibrariesManager) *rpc.LibrarySearchResponse {
	res := []*rpc.SearchedLibrary{}
	query := req.GetSearchArgs()
	if query == "" {
		query = req.GetQuery()
	}
	queryTerms := utils.SearchTermsFromQueryString(query)

	for _, lib := range lm.Index.Libraries {
		toTest := lib.Name + " " +
			lib.Latest.Paragraph + " " +
			lib.Latest.Sentence + " " +
			lib.Latest.Author + " "
		for _, include := range lib.Latest.ProvidesIncludes {
			toTest += include + " "
		}

		if utils.Match(toTest, queryTerms) {
			res = append(res, indexLibraryToRPCSearchLibrary(lib, req.GetOmitReleasesDetails()))
		}
	}

	// get a sorted slice of results
	sort.Slice(res, func(i, j int) bool {
		// Sort by name, but bubble up exact matches
		equalsI := strings.EqualFold(res[i].Name, query)
		equalsJ := strings.EqualFold(res[j].Name, query)
		if equalsI && !equalsJ {
			return true
		} else if !equalsI && equalsJ {
			return false
		}
		return res[i].Name < res[j].Name
	})

	return &rpc.LibrarySearchResponse{Libraries: res, Status: rpc.LibrarySearchStatus_LIBRARY_SEARCH_STATUS_SUCCESS}
}

// indexLibraryToRPCSearchLibrary converts a librariindex.Library to rpc.SearchLibrary
func indexLibraryToRPCSearchLibrary(lib *librariesindex.Library, omitReleasesDetails bool) *rpc.SearchedLibrary {
	var releases map[string]*rpc.LibraryRelease
	if !omitReleasesDetails {
		releases = map[string]*rpc.LibraryRelease{}
		for _, rel := range lib.Releases {
			releases[rel.Version.String()] = getLibraryParameters(rel)
		}
	}

	versions := semver.List{}
	for _, rel := range lib.Releases {
		versions = append(versions, rel.Version)
	}
	sort.Sort(versions)

	versionsString := []string{}
	for _, v := range versions {
		versionsString = append(versionsString, v.String())
	}

	return &rpc.SearchedLibrary{
		Name:              lib.Name,
		Releases:          releases,
		Latest:            getLibraryParameters(lib.Latest),
		AvailableVersions: versionsString,
	}
}

// getLibraryParameters FIXMEDOC
func getLibraryParameters(rel *librariesindex.Release) *rpc.LibraryRelease {
	return &rpc.LibraryRelease{
		Author:           rel.Author,
		Version:          rel.Version.String(),
		Maintainer:       rel.Maintainer,
		Sentence:         rel.Sentence,
		Paragraph:        rel.Paragraph,
		Website:          rel.Website,
		Category:         rel.Category,
		Architectures:    rel.Architectures,
		Types:            rel.Types,
		License:          rel.License,
		ProvidesIncludes: rel.ProvidesIncludes,
		Dependencies:     getLibraryDependenciesParameter(rel.GetDependencies()),
		Resources: &rpc.DownloadResource{
			Url:             rel.Resource.URL,
			ArchiveFilename: rel.Resource.ArchiveFileName,
			Checksum:        rel.Resource.Checksum,
			Size:            rel.Resource.Size,
			CachePath:       rel.Resource.CachePath,
		},
	}
}

func getLibraryDependenciesParameter(deps []semver.Dependency) []*rpc.LibraryDependency {
	res := []*rpc.LibraryDependency{}
	for _, dep := range deps {
		res = append(res, &rpc.LibraryDependency{
			Name:              dep.GetName(),
			VersionConstraint: dep.GetConstraint().String(),
		})
	}
	return res
}
