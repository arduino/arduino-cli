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

package libraries

import (
	"fmt"

	"github.com/arduino/arduino-cli/arduino/cores"
	"github.com/arduino/go-paths-helper"
	semver "go.bug.st/relaxed-semver"
)

var MandatoryProperties = []string{"name", "version", "author", "maintainer"}
var OptionalProperties = []string{"sentence", "paragraph", "url"}
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

	InstallDir        *paths.Path
	SourceDir         *paths.Path
	UtilityDir        *paths.Path
	Location          LibraryLocation
	ContainerPlatform *cores.PlatformRelease `json:""`
	Layout            LibraryLayout
	RealName          string
	DotALinkage       bool
	Precompiled       bool
	LDflags           string
	IsLegacy          bool
	Version           *semver.Version
	License           string
	Properties        map[string]string
}

func (library *Library) String() string {
	if library.Version.String() == "" {
		return library.Name
	}
	return library.Name + "@" + library.Version.String()
}

// SupportsAnyArchitectureIn returns true if any of the following is true:
// - the library supports at least one of the given architectures
// - the library is architecture independent
// - the library doesn't specify any `architecture` field in library.properties
func (library *Library) SupportsAnyArchitectureIn(archs ...string) bool {
	if len(library.Architectures) == 0 {
		return true
	}
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
	return library.IsOptimizedForArchitecture("*")
}

// PriorityForArchitecture returns an integer that represents the
// priority this lib has for the specified architecture based on
// his location and the architectures directly supported (as exposed
// on the `architecture` field of the `library.properties`)
// This function returns an integer between 0 and 255, higher means
// higher priority.
func (library *Library) PriorityForArchitecture(arch string) uint8 {
	bonus := uint8(0)

	// Bonus for core-optimized libraries
	if library.IsOptimizedForArchitecture(arch) {
		bonus = 0x10
	}

	switch library.Location {
	case IDEBuiltIn:
		return bonus + 0x00
	case ReferencedPlatformBuiltIn:
		return bonus + 0x01
	case PlatformBuiltIn:
		return bonus + 0x02
	case Sketchbook:
		return bonus + 0x03
	}
	panic(fmt.Sprintf("Invalid library location: %d", library.Location))
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
