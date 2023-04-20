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

package builder

import (
	"fmt"

	"github.com/arduino/arduino-cli/arduino/libraries"
	"github.com/arduino/arduino-cli/legacy/builder/types"
)

func ResolveLibrary(ctx *types.Context, header string) *libraries.Library {
	resolver := ctx.LibrariesResolver
	importedLibraries := ctx.ImportedLibraries

	candidates := resolver.AlternativesFor(header)

	if ctx.Verbose {
		ctx.Info(tr("Alternatives for %[1]s: %[2]s", header, candidates))
		ctx.Info(fmt.Sprintf("ResolveLibrary(%s)", header))
		ctx.Info(fmt.Sprintf("  -> %s: %s", tr("candidates"), candidates))
	}

	if len(candidates) == 0 {
		return nil
	}

	for _, candidate := range candidates {
		if importedLibraries.Contains(candidate) {
			return nil
		}
	}

	selected := resolver.ResolveFor(header, ctx.TargetPlatform.Platform.Architecture)
	if alreadyImported := importedLibraries.FindByName(selected.Name); alreadyImported != nil {
		// Certain libraries might have the same name but be different.
		// This usually happens when the user includes two or more custom libraries that have
		// different header name but are stored in a parent folder with identical name, like
		// ./libraries1/Lib/lib1.h and ./libraries2/Lib/lib2.h
		// Without this check the library resolution would be stuck in a loop.
		// This behaviour has been reported in this issue:
		// https://github.com/arduino/arduino-cli/issues/973
		if selected == alreadyImported {
			selected = alreadyImported
		}
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
