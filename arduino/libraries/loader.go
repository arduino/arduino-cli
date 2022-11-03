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
	"strings"

	"github.com/arduino/arduino-cli/arduino/sketch"
	"github.com/arduino/go-paths-helper"
	properties "github.com/arduino/go-properties-orderedmap"
	"github.com/pkg/errors"
	semver "go.bug.st/relaxed-semver"
)

// Load loads a library from the given LibraryLocation
func Load(libDir *paths.Path, location LibraryLocation) (*Library, error) {
	if libDir.Join("library.properties").Exist() {
		return makeNewLibrary(libDir, location)
	}
	return makeLegacyLibrary(libDir, location)
}

func addUtilityDirectory(library *Library) {
	utilitySourcePath := library.InstallDir.Join("utility")
	if utilitySourcePath.IsDir() {
		library.UtilityDir = utilitySourcePath
	}
}

func makeNewLibrary(libraryDir *paths.Path, location LibraryLocation) (*Library, error) {
	libProperties, err := properties.Load(libraryDir.Join("library.properties").String())
	if err != nil {
		return nil, fmt.Errorf(tr("loading library.properties: %s"), err)
	}

	if libProperties.Get("maintainer") == "" && libProperties.Get("email") != "" {
		libProperties.Set("maintainer", libProperties.Get("email"))
	}

	for _, propName := range MandatoryProperties {
		if libProperties.Get(propName) == "" {
			libProperties.Set(propName, "-")
		}
	}

	commaSeparatedToList := func(in string) []string {
		res := []string{}
		for _, e := range strings.Split(in, ",") {
			res = append(res, strings.TrimSpace(e))
		}
		return res
	}

	library := &Library{}
	library.Location = location
	library.InstallDir = libraryDir.Canonical()
	if libraryDir.Join("src").Exist() {
		library.Layout = RecursiveLayout
		library.SourceDir = libraryDir.Join("src")
	} else {
		library.Layout = FlatLayout
		library.SourceDir = libraryDir
		addUtilityDirectory(library)
	}

	if libProperties.Get("architectures") == "" {
		libProperties.Set("architectures", "*")
	}
	library.Architectures = commaSeparatedToList(libProperties.Get("architectures"))

	libProperties.Set("category", strings.TrimSpace(libProperties.Get("category")))
	if !ValidCategories[libProperties.Get("category")] {
		libProperties.Set("category", "Uncategorized")
	}
	library.Category = libProperties.Get("category")

	if libProperties.Get("license") == "" {
		libProperties.Set("license", "Unspecified")
	}
	library.License = libProperties.Get("license")

	version := strings.TrimSpace(libProperties.Get("version"))
	if v, err := semver.Parse(version); err != nil {
		// FIXME: do it in linter?
		//fmt.Printf("invalid version %s for library in %s: %s", version, libraryDir, err)
	} else {
		library.Version = v
	}

	if includes := libProperties.Get("includes"); includes != "" {
		library.declaredHeaders = commaSeparatedToList(includes)
	}

	if err := addExamples(library); err != nil {
		return nil, errors.Errorf(tr("scanning examples: %s"), err)
	}
	library.DirName = libraryDir.Base()
	library.Name = strings.TrimSpace(libProperties.Get("name"))
	library.Author = strings.TrimSpace(libProperties.Get("author"))
	library.Maintainer = strings.TrimSpace(libProperties.Get("maintainer"))
	library.Sentence = strings.TrimSpace(libProperties.Get("sentence"))
	library.Paragraph = strings.TrimSpace(libProperties.Get("paragraph"))
	library.Website = strings.TrimSpace(libProperties.Get("url"))
	library.IsLegacy = false
	library.DotALinkage = libProperties.GetBoolean("dot_a_linkage")
	library.PrecompiledWithSources = libProperties.Get("precompiled") == "full"
	library.Precompiled = libProperties.Get("precompiled") == "true" || library.PrecompiledWithSources
	library.LDflags = strings.TrimSpace(libProperties.Get("ldflags"))
	library.Properties = libProperties
	library.InDevelopment = libraryDir.Join(".development").Exist()
	return library, nil
}

func makeLegacyLibrary(path *paths.Path, location LibraryLocation) (*Library, error) {
	library := &Library{
		InstallDir:    path.Canonical(),
		Location:      location,
		SourceDir:     path,
		Layout:        FlatLayout,
		Name:          path.Base(),
		DirName:       path.Base(),
		Architectures: []string{"*"},
		IsLegacy:      true,
		Version:       semver.MustParse(""),
		InDevelopment: path.Join(".development").Exist(),
	}
	if err := addExamples(library); err != nil {
		return nil, errors.Errorf(tr("scanning examples: %s"), err)
	}
	addUtilityDirectory(library)
	return library, nil
}

func addExamples(lib *Library) error {
	files, err := lib.InstallDir.ReadDir()
	if err != nil {
		return err
	}
	examples := paths.NewPathList()
	for _, file := range files {
		name := strings.ToLower(file.Base())
		if name != "example" && name != "examples" {
			continue
		}
		if !file.IsDir() {
			continue
		}
		if err := addExamplesToPathList(file, &examples); err != nil {
			return err
		}
		break
	}

	lib.Examples = examples
	return nil
}

func addExamplesToPathList(examplesPath *paths.Path, list *paths.PathList) error {
	files, err := examplesPath.ReadDir()
	if err != nil {
		return err
	}
	for _, file := range files {
		_, err := sketch.New(file)
		if err == nil {
			list.Add(file)
		} else if file.IsDir() {
			if err := addExamplesToPathList(file, list); err != nil {
				return err
			}
		}
	}
	return nil
}
