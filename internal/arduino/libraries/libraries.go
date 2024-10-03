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
	"encoding/binary"
	"errors"
	"fmt"
	"io"

	"github.com/arduino/arduino-cli/internal/arduino/cores"
	"github.com/arduino/arduino-cli/internal/arduino/globals"
	"github.com/arduino/arduino-cli/internal/i18n"
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
	DirName                string
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
	InDevelopment          bool
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

func (library *Library) MarshalBinary(out io.Writer, prefix *paths.Path) error {
	writeString := func(in string) error {
		inBytes := []byte(in)
		if err := binary.Write(out, binary.NativeEndian, uint16(len(inBytes))); err != nil {
			return err
		}
		_, err := out.Write(inBytes)
		return err
	}
	writeStringArray := func(in []string) error {
		if err := binary.Write(out, binary.NativeEndian, uint16(len(in))); err != nil {
			return err
		}
		for _, i := range in {
			if err := writeString(i); err != nil {
				return err
			}
		}
		return nil
	}
	writeMap := func(in map[string]bool) error {
		if err := binary.Write(out, binary.NativeEndian, uint16(len(in))); err != nil {
			return err
		}
		for k, v := range in {
			if err := writeString(k); err != nil {
				return err
			}
			if err := binary.Write(out, binary.NativeEndian, v); err != nil {
				return err
			}
		}
		return nil
	}
	writeProperties := func(in *properties.Map) error {
		keys := in.Keys()
		if err := binary.Write(out, binary.NativeEndian, uint16(len(keys))); err != nil {
			return err
		}
		for _, k := range keys {
			v := in.Get(k)
			if err := writeString(k); err != nil {
				return err
			}
			if err := writeString(v); err != nil {
				return err
			}
		}
		return nil
	}
	writePath := func(in *paths.Path) error {
		if in == nil {
			return writeString("")
		} else if p, err := in.RelFrom(prefix); err != nil {
			return err
		} else {
			return writeString(p.String())
		}
	}
	writePathList := func(in []*paths.Path) error {
		if err := binary.Write(out, binary.NativeEndian, uint16(len(in))); err != nil {
			return err
		}
		for _, p := range in {
			if err := writePath(p); err != nil {
				return err
			}
		}
		return nil
	}
	if err := writeString(library.Name); err != nil {
		return err
	}
	if err := writeString(library.Author); err != nil {
		return err
	}
	if err := writeString(library.Maintainer); err != nil {
		return err
	}
	if err := writeString(library.Sentence); err != nil {
		return err
	}
	if err := writeString(library.Paragraph); err != nil {
		return err
	}
	if err := writeString(library.Website); err != nil {
		return err
	}
	if err := writeString(library.Category); err != nil {
		return err
	}
	if err := writeStringArray(library.Architectures); err != nil {
		return err
	}
	if err := writeStringArray(library.Types); err != nil {
		return err
	}
	if err := writePath(library.InstallDir); err != nil {
		return err
	}
	if err := writeString(library.DirName); err != nil {
		return err
	}
	if err := writePath(library.SourceDir); err != nil {
		return err
	}
	if err := writePath(library.UtilityDir); err != nil {
		return err
	}
	if err := binary.Write(out, binary.NativeEndian, int32(library.Location)); err != nil {
		return err
	}
	// library.ContainerPlatform      *cores.PlatformRelease `json:""`
	if err := binary.Write(out, binary.NativeEndian, int32(library.Layout)); err != nil {
		return err
	}
	if err := binary.Write(out, binary.NativeEndian, library.DotALinkage); err != nil {
		return err
	}
	if err := binary.Write(out, binary.NativeEndian, library.Precompiled); err != nil {
		return err
	}
	if err := binary.Write(out, binary.NativeEndian, library.PrecompiledWithSources); err != nil {
		return err
	}
	if err := writeString(library.LDflags); err != nil {
		return err
	}
	if err := binary.Write(out, binary.NativeEndian, library.IsLegacy); err != nil {
		return err
	}
	if err := binary.Write(out, binary.NativeEndian, library.InDevelopment); err != nil {
		return err
	}
	if err := writeString(library.Version.String()); err != nil {
		return err
	}
	if err := writeString(library.License); err != nil {
		return err
	}
	if err := writePathList(library.Examples); err != nil {
		return err
	}
	if err := writeStringArray(library.declaredHeaders); err != nil {
		return err
	}
	if err := writeStringArray(library.sourceHeaders); err != nil {
		return err
	}
	if err := writeMap(library.CompatibleWith); err != nil {
		return err
	}
	if err := writeProperties(library.Properties); err != nil {
		return err
	}

	return nil
}

func (library *Library) UnmarshalBinary(in io.Reader, prefix *paths.Path) error {
	readString := func() (string, error) {
		var len uint16
		if err := binary.Read(in, binary.NativeEndian, &len); err != nil {
			return "", err
		}
		res := make([]byte, len)
		if _, err := in.Read(res); err != nil {
			return "", err
		}
		return string(res), nil
	}
	readStringArray := func() ([]string, error) {
		var len uint16
		if err := binary.Read(in, binary.NativeEndian, &len); err != nil {
			return nil, err
		}
		if len == 0 {
			return nil, nil
		}
		res := make([]string, len)
		for i := range res {
			var err error
			res[i], err = readString()
			if err != nil {
				return nil, err
			}
		}
		return res, nil
	}
	readMap := func() (map[string]bool, error) {
		var len uint16
		if err := binary.Read(in, binary.NativeEndian, &len); err != nil {
			return nil, err
		}
		res := map[string]bool{}
		for i := uint16(0); i < len; i++ {
			k, err := readString()
			if err != nil {
				return nil, err
			}
			var v bool
			if err := binary.Read(in, binary.NativeEndian, &v); err != nil {
				return nil, err
			}
			res[k] = v
		}
		return res, nil
	}
	readPath := func() (*paths.Path, error) {
		if p, err := readString(); err != nil {
			return nil, err
		} else if p == "" {
			return nil, nil
		} else {
			return prefix.Join(p), nil
		}
	}
	readPathList := func() (paths.PathList, error) {
		var len uint16
		if err := binary.Read(in, binary.NativeEndian, &len); err != nil {
			return nil, err
		}
		list := paths.NewPathList()
		for range len {
			if p, err := readPath(); err != nil {
				return nil, err
			} else {
				list.Add(p)
			}
		}
		return list, nil
	}
	readProperties := func() (*properties.Map, error) {
		var len uint16
		if err := binary.Read(in, binary.NativeEndian, &len); err != nil {
			return nil, err
		}
		props := properties.NewMap()
		for range len {
			if k, err := readString(); err != nil {
				return nil, err
			} else if v, err := readString(); err != nil {
				return nil, err
			} else {
				props.Set(k, v)
			}
		}
		return props, nil
	}
	var err error
	library.Name, err = readString()
	if err != nil {
		return err
	}
	library.Author, err = readString()
	if err != nil {
		return err
	}
	library.Maintainer, err = readString()
	if err != nil {
		return err
	}
	library.Sentence, err = readString()
	if err != nil {
		return err
	}
	library.Paragraph, err = readString()
	if err != nil {
		return err
	}
	library.Website, err = readString()
	if err != nil {
		return err
	}
	library.Category, err = readString()
	if err != nil {
		return err
	}
	library.Architectures, err = readStringArray()
	if err != nil {
		return err
	}
	library.Types, err = readStringArray()
	if err != nil {
		return err
	}
	library.InstallDir, err = readPath()
	if err != nil {
		return err
	}
	library.DirName, err = readString()
	if err != nil {
		return err
	}
	library.SourceDir, err = readPath()
	if err != nil {
		return err
	}
	library.UtilityDir, err = readPath()
	if err != nil {
		return err
	}
	var location int32
	if err := binary.Read(in, binary.NativeEndian, &location); err != nil {
		return err
	}
	library.Location = LibraryLocation(location)
	// library.ContainerPlatform      *cores.PlatformRelease `json:""`
	var layout int32
	if err := binary.Read(in, binary.NativeEndian, &layout); err != nil {
		return err
	}
	library.Layout = LibraryLayout(layout)
	if err := binary.Read(in, binary.NativeEndian, &library.DotALinkage); err != nil {
		return err
	}
	if err := binary.Read(in, binary.NativeEndian, &library.Precompiled); err != nil {
		return err
	}
	if err := binary.Read(in, binary.NativeEndian, &library.PrecompiledWithSources); err != nil {
		return err
	}
	library.LDflags, err = readString()
	if err != nil {
		return err
	}
	if err := binary.Read(in, binary.NativeEndian, &library.IsLegacy); err != nil {
		return err
	}
	if err := binary.Read(in, binary.NativeEndian, &library.InDevelopment); err != nil {
		return err
	}
	version, err := readString()
	if err != nil {
		return err
	}
	library.Version = semver.MustParse(version)
	library.License, err = readString()
	if err != nil {
		return err
	}
	library.Examples, err = readPathList()
	if err != nil {
		return err
	}
	library.declaredHeaders, err = readStringArray()
	if err != nil {
		return err
	}
	library.sourceHeaders, err = readStringArray()
	if err != nil {
		return err
	}
	library.CompatibleWith, err = readMap()
	if err != nil {
		return err
	}
	library.Properties, err = readProperties()
	if err != nil {
		return err
	}
	return nil
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

	// If the "includes" property is empty or not included in the "library.properties" file
	// we search for headers by reading the library files directly
	headers := library.DeclaredHeaders()
	if len(headers) == 0 {
		var err error
		headers, err = library.SourceHeaders()
		if err != nil {
			return nil, fmt.Errorf("%s: %w", i18n.Tr("reading library headers"), err)
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
		InDevelopment:     library.InDevelopment,
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

// IsCompatibleWith returns true if the library declares compatibility with
// the given architecture. If this function returns false, the library may still
// be compatible with the given architecture, but it's not explicitly declared.
func (library *Library) IsCompatibleWith(arch string) bool {
	return library.IsArchitectureIndependent() || library.IsOptimizedForArchitecture(arch)
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
			return nil, errors.New(i18n.Tr("reading library source directory: %s", err))
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
