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
	"iter"
	"sort"
	"strings"

	"github.com/arduino/arduino-cli/commands/cmderrors"
	"github.com/arduino/arduino-cli/internal/arduino/libraries"
	"github.com/arduino/arduino-cli/internal/arduino/resources"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	"github.com/arduino/go-paths-helper"
	semver "go.bug.st/relaxed-semver"
)

// Index represents the list of libraries available for download. The library
// data is not kept in memory: it is streamed from the backing
// library_index.json file on demand by the methods below.
type Index struct {
	indexFile *paths.Path
}

// EmptyIndex is an empty library index
var EmptyIndex = &Index{}

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

// ReleaseCompare compares two library releases by name, or by version if the names are equal.
func ReleaseCompare(r1, r2 *Release) int {
	if cmp := strings.Compare(r1.GetName(), r2.GetName()); cmp != 0 {
		return cmp
	}
	return r1.GetVersion().CompareTo(r2.GetVersion())
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

// AsReleaseReference converts this release into a ReleaseReference, which is a stripped-down version of Release used for dependency resolution.
func (r *Release) AsReleaseReference() *ReleaseReference {
	return &ReleaseReference{
		name:         r.GetName(),
		version:      r.GetVersion(),
		dependencies: r.GetDependencies(),
	}
}

// releases streams the releases in the backing index file. It yields nothing on
// an empty index.
func (idx *Index) releases() iter.Seq[*indexRelease] {
	if idx == nil || idx.indexFile == nil {
		return func(yield func(*indexRelease) bool) {}
	}
	return scanIndexFile(idx.indexFile)
}

// findLibraries collects the requested libraries (by name), with all their
// releases, in a single scan of the index file.
func (idx *Index) findLibraries(names map[string]bool) map[string]*Library {
	libs := map[string]*Library{}
	if len(names) == 0 {
		return libs
	}
	for r := range idx.releases() {
		if !names[r.Name] {
			continue
		}
		lib := libs[r.Name]
		if lib == nil {
			lib = &Library{Name: r.Name, Releases: map[semver.NormalizedString]*Release{}}
			libs[r.Name] = lib
		}
		r.extractReleaseIn(lib)
	}
	return libs
}

// findLibrary collects a single library (with all its releases) from the index.
func (idx *Index) findLibrary(name string) *Library {
	return idx.findLibraries(map[string]bool{name: true})[name]
}

// FindIndexedLibraries returns the indexed libraries matching the given names,
// loaded in a single scan of the index file.
func (idx *Index) FindIndexedLibraries(names []string) map[string]*Library {
	set := make(map[string]bool, len(names))
	for _, name := range names {
		set[name] = true
	}
	return idx.findLibraries(set)
}

// Libraries streams the index, yielding each library with all its releases
// populated. Only one library at a time is held in memory. This relies on the
// index grouping all the releases of a library contiguously (as produced by the
// Arduino library index generator).
func (idx *Index) Libraries() iter.Seq[*Library] {
	return func(yield func(*Library) bool) {
		var current *Library
		for r := range idx.releases() {
			if current != nil && current.Name != r.Name {
				if !yield(current) {
					return
				}
				current = nil
			}
			if current == nil {
				current = &Library{Name: r.Name, Releases: map[semver.NormalizedString]*Release{}}
			}
			r.extractReleaseIn(current)
		}
		if current != nil {
			yield(current)
		}
	}
}

// HasLibrary returns true if a library with the given name exists in the index.
func (idx *Index) HasLibrary(name string) bool {
	return idx.findLibrary(name) != nil
}

// FindRelease search a library Release in the index. Returns nil if the
// release is not found. If the version is not specified returns the latest
// version available.
func (idx *Index) FindRelease(name string, version *semver.Version) (*Release, error) {
	if library := idx.findLibrary(name); library != nil {
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
	return idx.findLibrary(lib.Name)
}

// FindLibraryUpdates maps each of the given installed libraries to an update
// if this is available in the index, or to nil otherwise.
// The lookup is performed for all libraries in a single scan of the index file.
func (idx *Index) FindLibraryUpdates(libs ...*libraries.Library) []*Release {
	names := make([]string, len(libs))
	for i, lib := range libs {
		names[i] = lib.Name
	}
	indexed := idx.FindIndexedLibraries(names)
	updates := []*Release{}
	for _, lib := range libs {
		var update *Release
		if indexLib := indexed[lib.Name]; indexLib != nil {
			// If a library.properties has an invalid version property, usually empty or malformed,
			// the latest available version is returned
			if lib.Version == nil || indexLib.Latest.Version.GreaterThan(lib.Version) {
				update = indexLib.Latest
			}
		}
		updates = append(updates, update)
	}
	return updates
}

// ReleaseReference is a stripped-down version of Release, used for dependency resolution.
// It implements the semver.Release interface, but does not include the large descriptive fields
// of a full Release.
type ReleaseReference struct {
	name         string
	version      *semver.Version
	dependencies []*Dependency
}

func (rel *ReleaseReference) GetDependencies() []*Dependency {
	return rel.dependencies
}

func (rel *ReleaseReference) GetName() string {
	return rel.name
}

func (rel *ReleaseReference) GetVersion() *semver.Version {
	return rel.version
}

func (rel *ReleaseReference) String() string {
	return rel.name + "@" + rel.version.String()
}

// ReleaseReferenceCompare compares two library releases reference by name, or by version if the names are equal.
func ReleaseReferenceCompare(r1, r2 *ReleaseReference) int {
	if cmp := strings.Compare(r1.GetName(), r2.GetName()); cmp != 0 {
		return cmp
	}
	return r1.GetVersion().CompareTo(r2.GetVersion())
}

// ResolveDependencies resolve the dependencies of a library release and returns a
// possible solution (the set of library releases to install together with the library).
// An optional "override" releases may be passed if we want to exclude the same
// libraries from the index (for example if we want to keep an installed library).
func (idx *Index) ResolveDependencies(lib *Release, overrides []*Release) []*ReleaseReference {
	resolver := semver.NewResolver[*ReleaseReference]()

	overridden := map[string]bool{}
	for _, override := range overrides {
		resolver.AddRelease(&ReleaseReference{
			name:         override.GetName(),
			version:      override.GetVersion(),
			dependencies: override.GetDependencies(),
		})
		overridden[override.GetName()] = true
	}

	// Create and populate the library resolver
	for indexLib := range idx.releases() {
		if _, ok := overridden[indexLib.Name]; ok {
			continue
		}
		resolver.AddRelease(&ReleaseReference{
			name:         indexLib.Name,
			version:      indexLib.Version,
			dependencies: indexLib.extractDependencies(),
		})
	}

	// Perform lib resolution
	return resolver.Resolve(&ReleaseReference{
		name:         lib.GetName(),
		version:      lib.GetVersion(),
		dependencies: lib.GetDependencies(),
	})
}

// ResolveReleaseReferences maps an array of ReleaseReference to their corresponding full Release objects in the index.
func (idx *Index) ResolveReleaseReferences(solution []*ReleaseReference) []*Release {
	names := map[string]bool{}
	for _, rel := range solution {
		names[rel.GetName()] = true
	}
	solutionLibs := idx.findLibraries(names)
	solutionReleases := make([]*Release, len(solution))
	for i, rel := range solution {
		solutionReleases[i] = solutionLibs[rel.GetName()].Releases[rel.GetVersion().NormalizedString()]
	}
	return solutionReleases
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
