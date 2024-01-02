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

	"github.com/arduino/arduino-cli/commands/internal/instances"
	"github.com/arduino/arduino-cli/internal/arduino/libraries/librariesindex"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	semver "go.bug.st/relaxed-semver"
)

// LibrarySearch FIXMEDOC
func LibrarySearch(ctx context.Context, req *rpc.LibrarySearchRequest) (*rpc.LibrarySearchResponse, error) {
	li, err := instances.GetLibrariesIndex(req.GetInstance())
	if err != nil {
		return nil, err
	}
	return searchLibrary(req, li), nil
}

func searchLibrary(req *rpc.LibrarySearchRequest, li *librariesindex.Index) *rpc.LibrarySearchResponse {
	res := []*rpc.SearchedLibrary{}
	query := req.GetSearchArgs()
	matcher := MatcherFromQueryString(query)

	for _, lib := range li.Libraries {
		if matcher(lib) {
			res = append(res, indexLibraryToRPCSearchLibrary(lib, req.GetOmitReleasesDetails()))
		}
	}

	// get a sorted slice of results
	sort.Slice(res, func(i, j int) bool {
		// Sort by name, but bubble up exact matches
		equalsI := strings.EqualFold(res[i].GetName(), query)
		equalsJ := strings.EqualFold(res[j].GetName(), query)
		if equalsI && !equalsJ {
			return true
		} else if !equalsI && equalsJ {
			return false
		}
		return res[i].GetName() < res[j].GetName()
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

func getLibraryDependenciesParameter(deps []*librariesindex.Dependency) []*rpc.LibraryDependency {
	res := []*rpc.LibraryDependency{}
	for _, dep := range deps {
		res = append(res, &rpc.LibraryDependency{
			Name:              dep.GetName(),
			VersionConstraint: dep.GetConstraint().String(),
		})
	}
	return res
}
