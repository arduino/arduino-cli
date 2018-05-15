/*
 * This file is part of arduino-cli.
 *
 * arduino-cli is free software; you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation; either version 2 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program; if not, write to the Free Software
 * Foundation, Inc., 51 Franklin St, Fifth Floor, Boston, MA  02110-1301  USA
 *
 * As a special exception, you may use this file as part of a free software
 * library without restriction.  Specifically, if other files instantiate
 * templates or use macros or inline functions from this file, or you compile
 * this file and link it with other files to produce an executable, this
 * file does not by itself cause the resulting executable to be covered by
 * the GNU General Public License.  This exception does not however
 * invalidate any other reasons why the executable file might be covered by
 * the GNU General Public License.
 *
 * Copyright 2018 ARDUINO AG (http://www.arduino.cc/)
 */

package libraries

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	properties "github.com/arduino/go-properties-map"
)

func Load(libraryFolder string) (*Library, error) {
	if _, err := os.Stat(filepath.Join(libraryFolder, "library.properties")); os.IsNotExist(err) {
		return makeLegacyLibrary(libraryFolder)
	}
	return makeNewLibrary(libraryFolder)
}

func addUtilityFolder(library *Library) {
	utilitySourcePath := filepath.Join(library.Folder, "utility")
	stat, err := os.Stat(utilitySourcePath)
	if err == nil && stat.IsDir() {
		library.UtilityFolder = utilitySourcePath
	}
}

func makeNewLibrary(libraryFolder string) (*Library, error) {
	libProperties, err := properties.Load(filepath.Join(libraryFolder, "library.properties"))
	if err != nil {
		return nil, fmt.Errorf("loading library.properties: %s", err)
	}

	if libProperties["maintainer"] == "" && libProperties["email"] != "" {
		libProperties["maintainer"] = libProperties["email"]
	}

	for _, propName := range MandatoryProperties {
		if libProperties[propName] == "" {
			libProperties[propName] = "-"
		}
	}

	library := &Library{}
	library.Folder = libraryFolder
	if stat, err := os.Stat(filepath.Join(libraryFolder, "src")); err == nil && stat.IsDir() {
		library.Layout = RecursiveLayout
		library.SrcFolder = filepath.Join(libraryFolder, "src")
	} else {
		library.Layout = FlatLayout
		library.SrcFolder = libraryFolder
		addUtilityFolder(library)
	}

	if libProperties["architectures"] == "" {
		libProperties["architectures"] = "*"
	}
	library.Architectures = []string{}
	for _, arch := range strings.Split(libProperties["architectures"], ",") {
		library.Architectures = append(library.Architectures, strings.TrimSpace(arch))
	}

	libProperties["category"] = strings.TrimSpace(libProperties["category"])
	if !ValidCategories[libProperties["category"]] {
		libProperties["category"] = "Uncategorized"
	}
	library.Category = libProperties["category"]

	if libProperties["license"] == "" {
		libProperties["license"] = "Unspecified"
	}
	library.License = libProperties["license"]

	library.Name = filepath.Base(libraryFolder)
	library.RealName = strings.TrimSpace(libProperties["name"])
	library.Version = strings.TrimSpace(libProperties["version"])
	library.Author = strings.TrimSpace(libProperties["author"])
	library.Maintainer = strings.TrimSpace(libProperties["maintainer"])
	library.Sentence = strings.TrimSpace(libProperties["sentence"])
	library.Paragraph = strings.TrimSpace(libProperties["paragraph"])
	library.Website = strings.TrimSpace(libProperties["url"])
	library.IsLegacy = false
	library.DotALinkage = strings.TrimSpace(libProperties["dot_a_linkage"]) == "true"
	library.Precompiled = strings.TrimSpace(libProperties["precompiled"]) == "true"
	library.LDflags = strings.TrimSpace(libProperties["ldflags"])
	library.Properties = libProperties

	return library, nil
}

func makeLegacyLibrary(libraryFolder string) (*Library, error) {
	library := &Library{
		Folder:        libraryFolder,
		SrcFolder:     libraryFolder,
		Layout:        FlatLayout,
		Name:          filepath.Base(libraryFolder),
		Architectures: []string{"*"},
		IsLegacy:      true,
	}
	addUtilityFolder(library)
	return library, nil
}
