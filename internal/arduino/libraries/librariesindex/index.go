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
	"sort"

	"github.com/arduino/arduino-cli/commands/cmderrors"
	"github.com/arduino/arduino-cli/internal/arduino/libraries"
	"github.com/arduino/arduino-cli/internal/arduino/resources"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	semver "go.bug.st/relaxed-semver"
)

// Index represents the list of libraries available for download
type Index struct {
	Libraries map[string]*Library
}

// EmptyIndex is an empty library index
var EmptyIndex = &Index{Libraries: map[string]*Library{}}

// Library is a library available for download
type Library struct {
	Name     string
	Releases map[semver.NormalizedString]*Release
	Latest   *Release `json:"-"`
	Index    *Index   `json:"-"`
}

// Release is a release of a library available for download
type Release struct {
	Author           string
	Version          *semver.Version
	Dependencies     []*Dependency
	Maintainer       string
	Sentence         string
	Paragraph        string
	Website          string
	Category         string
	Architectures    []string
	Types            []string
	Resource         *resources.DownloadResource
	License          string
	ProvidesIncludes []string

	Library *Library `json:"-"`
}

// ToRPCLibraryRelease transform this Release into a rpc.LibraryRelease
func (r *Release) ToRPCLibraryRelease() *rpc.LibraryRelease {
	return &rpc.LibraryRelease{
		Author:        r.Author,
		Version:       r.Version.String(),
		Maintainer:    r.Maintainer,
		Sentence:      r.Sentence,
		Paragraph:     r.Paragraph,
		Website:       r.Website,
		Category:      r.Category,
		Architectures: r.Architectures,
		Types:         r.Types,
	}
}

// GetName returns the name of this library.
func (r *Release) GetName() string {
	return r.Library.Name
}

// GetVersion returns the version of this library.
func (r *Release) GetVersion() *semver.Version {
	return r.Version
}

// GetDependencies returns the dependencies of this library.
func (r *Release) GetDependencies() []*Dependency {
	return r.Dependencies
}

// Dependency is a library dependency
type Dependency struct {
	Name              string
	VersionConstraint semver.Constraint
}

// GetName returns the name of the dependency
func (r *Dependency) GetName() string {
	return r.Name
}

// GetConstraint returns the version Constraint of the dependecy
func (r *Dependency) GetConstraint() semver.Constraint {
	return r.VersionConstraint
}

func (r *Release) String() string {
	return r.Library.Name + "@" + r.Version.String()
}

// FindRelease search a library Release in the index. Returns nil if the
// release is not found. If the version is not specified returns the latest
// version available.
func (idx *Index) FindRelease(name string, version *semver.Version) (*Release, error) {
	if library, exists := idx.Libraries[name]; exists {
		if version == nil {
			return library.Latest, nil
		}
		if release, exists := library.Releases[version.NormalizedString()]; exists {
			return release, nil
		}
	}
	if version == nil {
		return nil, &cmderrors.LibraryNotFoundError{Library: name + "@latest"}
	}
	return nil, &cmderrors.LibraryNotFoundError{Library: name + "@" + version.String()}
}

// FindIndexedLibrary search an indexed library that matches the provided
// installed library or nil if not found
func (idx *Index) FindIndexedLibrary(lib *libraries.Library) *Library {
	return idx.Libraries[lib.Name]
}

// FindLibraryUpdate check if an installed library may be updated using
// one of the indexed libraries. This function returns the Release to install
// to update the library if found, otherwise nil is returned.
func (idx *Index) FindLibraryUpdate(lib *libraries.Library) *Release {
	indexLib := idx.FindIndexedLibrary(lib)
	if indexLib == nil {
		return nil
	}
	// If a library.properties has an invalid version property, usually empty or malformed,
	// the latest available version is returned
	if lib.Version == nil || indexLib.Latest.Version.GreaterThan(lib.Version) {
		return indexLib.Latest
	}
	return nil
}

// ResolveDependencies resolve the dependencies of a library release and returns a
// possible solution (the set of library releases to install together with the library).
// An optional "override" releases may be passed if we want to exclude the same
// libraries from the index (for example if we want to keep an installed library).
func (idx *Index) ResolveDependencies(lib *Release, overrides []*Release) []*Release {
	resolver := semver.NewResolver[*Release]()

	overridden := map[string]bool{}
	for _, override := range overrides {
		resolver.AddRelease(override)
		overridden[override.GetName()] = true
	}

	// Create and populate the library resolver
	for libName, indexLib := range idx.Libraries {
		if _, ok := overridden[libName]; ok {
			continue
		}
		for _, indexLibRelease := range indexLib.Releases {
			resolver.AddRelease(indexLibRelease)
		}
	}

	// Perform lib resolution
	return resolver.Resolve(lib)
}

// Versions returns an array of all versions available of the library
func (library *Library) Versions() []*semver.Version {
	res := semver.List{}
	for _, release := range library.Releases {
		res = append(res, release.Version)
	}
	sort.Sort(res)
	return res
}
