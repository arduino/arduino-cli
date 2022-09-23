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

package libraries

import (
	"fmt"

	"github.com/arduino/arduino-cli/arduino/cores"
	"github.com/arduino/arduino-cli/arduino/globals"
	"github.com/arduino/arduino-cli/i18n"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	paths "github.com/arduino/go-paths-helper"
	properties "github.com/arduino/go-properties-orderedmap"
	semver "go.bug.st/relaxed-semver"
)

// MandatoryProperties FIXMEDOC
var MandatoryProperties = []string{"name", "version", "author", "maintainer"}

// OptionalProperties FIXMEDOC
var OptionalProperties = []string{"sentence", "paragraph", "url"}

// ValidCategories FIXMEDOC
var ValidCategories = map[string]bool{
	"Display":             true,
	"Communication":       true,
	"Signal Input/Output": true,
	"Sensors":             true,
	"Device Control":      true,
	"Timing":              true,
	"Data Storage":        true,
	"Data Processing":     true,
	"Other":               true,
	"Uncategorized":       true,
}

var tr = i18n.Tr

// Library represents a library in the system
type Library struct {
	Name          string
	Author        string
	Maintainer    string
	Sentence      string
	Paragraph     string
	Website       string
	Category      string
	Architectures []string

	Types []string `json:"types,omitempty"`

	InstallDir             *paths.Path
	CanonicalName          string
	SourceDir              *paths.Path
	UtilityDir             *paths.Path
	Location               LibraryLocation
	ContainerPlatform      *cores.PlatformRelease `json:""`
	Layout                 LibraryLayout
	DotALinkage            bool
	Precompiled            bool
	PrecompiledWithSources bool
	LDflags                string
	IsLegacy               bool
	Version                *semver.Version
	License                string
	Properties             *properties.Map
	Examples               paths.PathList
	declaredHeaders        []string
	sourceHeaders          []string
	CompatibleWith         map[string]bool
}

func (library *Library) String() string {
	if library.Version.String() == "" {
		return library.Name
	}
	return library.Name + "@" + library.Version.String()
}

// ToRPCLibrary converts this library into an rpc.Library
func (library *Library) ToRPCLibrary() (*rpc.Library, error) {
	pathOrEmpty := func(p *paths.Path) string {
		if p == nil {
			return ""
		}
		return p.String()
	}
	platformOrEmpty := func(p *cores.PlatformRelease) string {
		if p == nil {
			return ""
		}
		return p.String()
	}

	// If the the "includes" property is empty or not included in the "library.properties" file
	// we search for headers by reading the library files directly
	headers := library.DeclaredHeaders()
	if len(headers) == 0 {
		var err error
		headers, err = library.SourceHeaders()
		if err != nil {
			return nil, fmt.Errorf(tr("reading library headers: %w"), err)
		}
	}

	return &rpc.Library{
		Name:              library.Name,
		Author:            library.Author,
		Maintainer:        library.Maintainer,
		Sentence:          library.Sentence,
		Paragraph:         library.Paragraph,
		Website:           library.Website,
		Category:          library.Category,
		Architectures:     library.Architectures,
		Types:             library.Types,
		InstallDir:        pathOrEmpty(library.InstallDir),
		SourceDir:         pathOrEmpty(library.SourceDir),
		UtilityDir:        pathOrEmpty(library.UtilityDir),
		Location:          library.Location.ToRPCLibraryLocation(),
		ContainerPlatform: platformOrEmpty(library.ContainerPlatform),
		Layout:            library.Layout.ToRPCLibraryLayout(),
		DotALinkage:       library.DotALinkage,
		Precompiled:       library.Precompiled,
		LdFlags:           library.LDflags,
		IsLegacy:          library.IsLegacy,
		Version:           library.Version.String(),
		License:           library.License,
		Examples:          library.Examples.AsStrings(),
		ProvidesIncludes:  headers,
		CompatibleWith:    library.CompatibleWith,
	}, nil
}

// SupportsAnyArchitectureIn returns true if any of the following is true:
// - the library supports at least one of the given architectures
// - the library is architecture independent
// - the library doesn't specify any `architecture` field in library.properties
func (library *Library) SupportsAnyArchitectureIn(archs ...string) bool {
	if library.IsArchitectureIndependent() {
		return true
	}
	for _, arch := range archs {
		if arch == "*" || library.IsOptimizedForArchitecture(arch) {
			return true
		}
	}
	return false
}

// IsOptimizedForArchitecture returns true if the library declares to be
// explicitly compatible for a specific architecture (the `architecture` field
// in library.properties contains the architecture passed as parameter)
func (library *Library) IsOptimizedForArchitecture(arch string) bool {
	for _, libArch := range library.Architectures {
		if libArch == arch {
			return true
		}
	}
	return false
}

// IsArchitectureIndependent returns true if the library declares to be
// compatible with all architectures (the `architecture` field in
// library.properties contains the `*` item)
func (library *Library) IsArchitectureIndependent() bool {
	return library.IsOptimizedForArchitecture("*") || library.Architectures == nil || len(library.Architectures) == 0
}

// SourceDir represents a source dir of a library
type SourceDir struct {
	Dir     *paths.Path
	Recurse bool
}

// SourceDirs return all the source directories of a library
func (library *Library) SourceDirs() []SourceDir {
	dirs := []SourceDir{}
	dirs = append(dirs,
		SourceDir{
			Dir:     library.SourceDir,
			Recurse: library.Layout == RecursiveLayout,
		})
	if library.UtilityDir != nil {
		dirs = append(dirs,
			SourceDir{
				Dir:     library.UtilityDir,
				Recurse: false,
			})
	}
	return dirs
}

// LocationPriorityFor returns a number representing the location priority for the given library
// using the given platform and referenced-platform. Higher value means higher priority.
func (library *Library) LocationPriorityFor(platformRelease, refPlatformRelease *cores.PlatformRelease) int {
	if library.Location == IDEBuiltIn {
		return 1
	} else if library.ContainerPlatform == refPlatformRelease {
		return 2
	} else if library.ContainerPlatform == platformRelease {
		return 3
	} else if library.Location == User {
		return 4
	}
	return 0
}

// DeclaredHeaders returns the C++ headers that the library declares in library.properties
func (library *Library) DeclaredHeaders() []string {
	if library.declaredHeaders == nil {
		library.declaredHeaders = []string{}
	}
	return library.declaredHeaders
}

// SourceHeaders returns all the C++ headers in the library even if not declared in library.properties
func (library *Library) SourceHeaders() ([]string, error) {
	if library.sourceHeaders == nil {
		cppHeaders, err := library.SourceDir.ReadDir()
		if err != nil {
			return nil, fmt.Errorf(tr("reading lib src dir: %s"), err)
		}
		headerExtensions := []string{}
		for k := range globals.HeaderFilesValidExtensions {
			headerExtensions = append(headerExtensions, k)
		}
		cppHeaders.FilterSuffix(headerExtensions...)
		res := []string{}
		for _, cppHeader := range cppHeaders {
			res = append(res, cppHeader.Base())
		}
		library.sourceHeaders = res
	}
	return library.sourceHeaders, nil
}
