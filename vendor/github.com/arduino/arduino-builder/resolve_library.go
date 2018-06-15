/*
 * This file is part of Arduino Builder.
 *
 * Arduino Builder is free software; you can redistribute it and/or modify
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
 * Copyright 2015 Arduino LLC (http://www.arduino.cc/)
 */

package builder

import (
	"path/filepath"
	"strings"

	"github.com/arduino/arduino-builder/constants"
	"github.com/arduino/arduino-builder/types"
	"github.com/arduino/arduino-builder/utils"
	"github.com/bcmi-labs/arduino-cli/arduino/cores"
	"github.com/bcmi-labs/arduino-cli/arduino/libraries"
)

func ResolveLibrary(ctx *types.Context, header string) *libraries.Library {
	headerToLibraries := ctx.HeaderToLibraries
	platforms := []*cores.PlatformRelease{ctx.ActualPlatform, ctx.TargetPlatform}
	libraryResolutionResults := ctx.LibrariesResolutionResults
	importedLibraries := ctx.ImportedLibraries

	libs := append([]*libraries.Library{}, headerToLibraries[header]...)

	if libs == nil || len(libs) == 0 {
		return nil
	}

	if importedLibraryContainsOneOfCandidates(importedLibraries, libs) {
		return nil
	}

	if len(libs) == 1 {
		return libs[0]
	}

	reverse(libs)

	var library *libraries.Library

	for _, platform := range platforms {
		if platform != nil {
			library = findBestLibraryWithHeader(header, librariesCompatibleWithPlatform(libs, platform, true))
		}
	}

	if library == nil {
		library = findBestLibraryWithHeader(header, libs)
	}

	if library == nil {
		// reorder libraries to promote fully compatible ones
		for _, platform := range platforms {
			if platform != nil {
				libs = append(librariesCompatibleWithPlatform(libs, platform, false), libs...)
			}
		}
		library = libs[0]
	}

	library = useAlreadyImportedLibraryWithSameNameIfExists(library, importedLibraries)

	libraryResolutionResults[header] = types.LibraryResolutionResult{
		Library:          library,
		NotUsedLibraries: filterOutLibraryFrom(libs, library),
	}

	return library
}

//facepalm. sort.Reverse needs an Interface that implements Len/Less/Swap. It's a slice! What else for reversing it?!?
func reverse(data []*libraries.Library) {
	for i, j := 0, len(data)-1; i < j; i, j = i+1, j-1 {
		data[i], data[j] = data[j], data[i]
	}
}

func importedLibraryContainsOneOfCandidates(imported []*libraries.Library, candidates []*libraries.Library) bool {
	for _, i := range imported {
		for _, j := range candidates {
			if i == j {
				return true
			}
		}
	}
	return false
}

func useAlreadyImportedLibraryWithSameNameIfExists(library *libraries.Library, imported []*libraries.Library) *libraries.Library {
	for _, lib := range imported {
		if lib.Name == library.Name {
			return lib
		}
	}
	return library
}

func filterOutLibraryFrom(libs []*libraries.Library, libraryToRemove *libraries.Library) []*libraries.Library {
	filteredOutLibraries := []*libraries.Library{}
	for _, lib := range libs {
		if lib != libraryToRemove {
			filteredOutLibraries = append(filteredOutLibraries, lib)
		}
	}
	return filteredOutLibraries
}

func libraryCompatibleWithPlatform(library *libraries.Library, platform *cores.PlatformRelease) (bool, bool) {
	if len(library.Architectures) == 0 {
		return true, true
	}
	if utils.SliceContains(library.Architectures, constants.LIBRARY_ALL_ARCHS) {
		return true, true
	}
	return utils.SliceContains(library.Architectures, platform.Platform.Architecture), false
}

func libraryCompatibleWithAllPlatforms(library *libraries.Library) bool {
	if utils.SliceContains(library.Architectures, constants.LIBRARY_ALL_ARCHS) {
		return true
	}
	return false
}

func librariesCompatibleWithPlatform(libs []*libraries.Library, platform *cores.PlatformRelease, reorder bool) []*libraries.Library {
	var compatibleLibraries []*libraries.Library
	for _, library := range libs {
		compatible, generic := libraryCompatibleWithPlatform(library, platform)
		if compatible {
			if !generic && len(compatibleLibraries) != 0 && libraryCompatibleWithAllPlatforms(compatibleLibraries[0]) && reorder == true {
				//priority inversion
				compatibleLibraries = append([]*libraries.Library{library}, compatibleLibraries...)
			} else {
				compatibleLibraries = append(compatibleLibraries, library)
			}
		}
	}

	return compatibleLibraries
}

func findBestLibraryWithHeader(header string, libs []*libraries.Library) *libraries.Library {
	headerName := strings.Replace(header, filepath.Ext(header), constants.EMPTY_STRING, -1)

	var library *libraries.Library
	for _, headerName := range []string{headerName, strings.ToLower(headerName)} {
		library = findLibWithName(headerName, libs)
		if library != nil {
			return library
		}
		library = findLibWithName(headerName+"-master", libs)
		if library != nil {
			return library
		}
		library = findLibWithNameStartingWith(headerName, libs)
		if library != nil {
			return library
		}
		library = findLibWithNameEndingWith(headerName, libs)
		if library != nil {
			return library
		}
		library = findLibWithNameContaining(headerName, libs)
		if library != nil {
			return library
		}
	}

	return nil
}

func findLibWithName(name string, libraries []*libraries.Library) *libraries.Library {
	for _, library := range libraries {
		if simplifyName(library.Name) == simplifyName(name) {
			return library
		}
	}
	return nil
}

func findLibWithNameStartingWith(name string, libraries []*libraries.Library) *libraries.Library {
	for _, library := range libraries {
		if strings.HasPrefix(simplifyName(library.Name), simplifyName(name)) {
			return library
		}
	}
	return nil
}

func findLibWithNameEndingWith(name string, libraries []*libraries.Library) *libraries.Library {
	for _, library := range libraries {
		if strings.HasSuffix(simplifyName(library.Name), simplifyName(name)) {
			return library
		}
	}
	return nil
}

func findLibWithNameContaining(name string, libraries []*libraries.Library) *libraries.Library {
	for _, library := range libraries {
		if strings.Contains(simplifyName(library.Name), simplifyName(name)) {
			return library
		}
	}
	return nil
}

func simplifyName(name string) string {
	return strings.ToLower(strings.Replace(name, "_", " ", -1))
}
