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
	"fmt"

	"github.com/arduino/arduino-builder/types"
	"github.com/bcmi-labs/arduino-cli/arduino/libraries"
)

func ResolveLibrary(ctx *types.Context, header string) *libraries.Library {
	resolver := ctx.LibrariesResolver
	importedLibraries := ctx.ImportedLibraries

	candidates := resolver.AlternativesFor(header)
	fmt.Printf("ResolveLibrary(%s)\n", header)
	fmt.Printf("  -> candidates: %s\n", candidates)

	if candidates == nil || len(candidates) == 0 {
		return nil
	}

	for _, candidate := range candidates {
		if importedLibraries.Contains(candidate) {
			return nil
		}
	}

	selected := resolver.ResolveFor(header, ctx.TargetPlatform.Platform.Architecture)
	if alreadyImported := importedLibraries.FindByName(selected.Name); alreadyImported != nil {
		selected = alreadyImported
	}

	ctx.LibrariesResolutionResults[header] = types.LibraryResolutionResult{
		Library:          selected,
		NotUsedLibraries: filterOutLibraryFrom(candidates, selected),
	}

	return selected
}

func filterOutLibraryFrom(libs libraries.List, libraryToRemove *libraries.Library) libraries.List {
	filteredOutLibraries := []*libraries.Library{}
	for _, lib := range libs {
		if lib != libraryToRemove {
			filteredOutLibraries = append(filteredOutLibraries, lib)
		}
	}
	return filteredOutLibraries
}
