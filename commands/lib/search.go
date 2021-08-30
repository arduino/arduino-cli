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

	"github.com/arduino/arduino-cli/arduino/libraries/librariesindex"
	"github.com/arduino/arduino-cli/arduino/libraries/librariesmanager"
	"github.com/arduino/arduino-cli/arduino/utils"
	"github.com/arduino/arduino-cli/commands"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	semver "go.bug.st/relaxed-semver"
)

// LibrarySearch FIXMEDOC
func LibrarySearch(ctx context.Context, req *rpc.LibrarySearchRequest) (*rpc.LibrarySearchResponse, error) {
	lm := commands.GetLibraryManager(req.GetInstance().GetId())
	if lm == nil {
		return nil, &commands.InvalidInstanceError{}
	}
	return searchLibrary(req, lm), nil
}

func searchLibrary(req *rpc.LibrarySearchRequest, lm *librariesmanager.LibrariesManager) *rpc.LibrarySearchResponse {
	res := []*rpc.SearchedLibrary{}
	status := rpc.LibrarySearchStatus_LIBRARY_SEARCH_STATUS_SUCCESS

	for _, lib := range lm.Index.Libraries {
		toTest := []string{lib.Name, lib.Latest.Paragraph, lib.Latest.Sentence}
		if !utils.MatchAny(req.GetQuery(), toTest) {
			continue
		}
		res = append(res, indexLibraryToRPCSearchLibrary(lib))
	}

	return &rpc.LibrarySearchResponse{Libraries: res, Status: status}
}

// indexLibraryToRPCSearchLibrary converts a librariindex.Library to rpc.SearchLibrary
func indexLibraryToRPCSearchLibrary(lib *librariesindex.Library) *rpc.SearchedLibrary {
	releases := map[string]*rpc.LibraryRelease{}
	for str, rel := range lib.Releases {
		releases[str] = getLibraryParameters(rel)
	}
	latest := getLibraryParameters(lib.Latest)

	return &rpc.SearchedLibrary{
		Name:     lib.Name,
		Releases: releases,
		Latest:   latest,
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
