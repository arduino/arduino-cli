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
	"bytes"
	"fmt"
	"strings"
	"time"

	"github.com/arduino/arduino-cli/arduino/cores"
	"github.com/arduino/arduino-cli/arduino/libraries"
	"github.com/arduino/arduino-cli/arduino/libraries/librariesmanager"
	"github.com/arduino/arduino-cli/arduino/libraries/librariesresolver"
	"github.com/arduino/go-paths-helper"
	"github.com/pkg/errors"
)

type libraryResolutionResult struct {
	Library          *libraries.Library
	NotUsedLibraries []*libraries.Library
}

// SketchLibrariesDetector todo
type SketchLibrariesDetector struct {
	librariesManager             *librariesmanager.LibrariesManager
	librariesResolver            *librariesresolver.Cpp
	useCachedLibrariesResolution bool
	verbose                      bool
	verboseInfoFn                func(msg string)
	verboseWarnFn                func(msg string)
	importedLibraries            libraries.List
	librariesResolutionResults   map[string]libraryResolutionResult
}

// NewSketchLibrariesDetector todo
func NewSketchLibrariesDetector(
	lm *librariesmanager.LibrariesManager,
	libsResolver *librariesresolver.Cpp,
	verbose, useCachedLibrariesResolution bool,
	verboseInfoFn func(msg string),
	verboseWarnFn func(msg string),
) *SketchLibrariesDetector {
	return &SketchLibrariesDetector{
		librariesManager:             lm,
		librariesResolver:            libsResolver,
		useCachedLibrariesResolution: useCachedLibrariesResolution,
		librariesResolutionResults:   map[string]libraryResolutionResult{},
		verbose:                      verbose,
		verboseInfoFn:                verboseInfoFn,
		verboseWarnFn:                verboseWarnFn,
		importedLibraries:            libraries.List{},
	}
}

// ResolveLibrary todo
func (l *SketchLibrariesDetector) ResolveLibrary(header, platformArch string) *libraries.Library {
	importedLibraries := l.importedLibraries
	candidates := l.librariesResolver.AlternativesFor(header)

	if l.verbose {
		l.verboseInfoFn(tr("Alternatives for %[1]s: %[2]s", header, candidates))
		l.verboseInfoFn(fmt.Sprintf("ResolveLibrary(%s)", header))
		l.verboseInfoFn(fmt.Sprintf("  -> %s: %s", tr("candidates"), candidates))
	}

	if len(candidates) == 0 {
		return nil
	}

	for _, candidate := range candidates {
		if importedLibraries.Contains(candidate) {
			return nil
		}
	}

	selected := l.librariesResolver.ResolveFor(header, platformArch)
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

	candidates.Remove(selected)
	l.librariesResolutionResults[header] = libraryResolutionResult{
		Library:          selected,
		NotUsedLibraries: candidates,
	}

	return selected
}

// ImportedLibraries todo
func (l *SketchLibrariesDetector) ImportedLibraries() libraries.List {
	// TODO understand if we have to do a deepcopy
	return l.importedLibraries
}

// AppendImportedLibraries todo should rename this, probably after refactoring the
// container_find_includes command.
func (l *SketchLibrariesDetector) AppendImportedLibraries(library *libraries.Library) {
	l.importedLibraries = append(l.importedLibraries, library)
}

// UseCachedLibrariesResolution todo
func (l *SketchLibrariesDetector) UseCachedLibrariesResolution() bool {
	return l.useCachedLibrariesResolution
}

// PrintUsedAndNotUsedLibraries todo
func (l *SketchLibrariesDetector) PrintUsedAndNotUsedLibraries(sketchError bool) {
	// Print this message:
	// - as warning, when the sketch didn't compile
	// - as info, when verbose is on
	// - otherwise, output nothing
	if !sketchError && !l.verbose {
		return
	}

	res := ""
	for header, libResResult := range l.librariesResolutionResults {
		if len(libResResult.NotUsedLibraries) == 0 {
			continue
		}
		res += fmt.Sprintln(tr(`Multiple libraries were found for "%[1]s"`, header))
		res += fmt.Sprintln("  " + tr("Used: %[1]s", libResResult.Library.InstallDir))
		for _, notUsedLibrary := range libResResult.NotUsedLibraries {
			res += fmt.Sprintln("  " + tr("Not used: %[1]s", notUsedLibrary.InstallDir))
		}
	}
	res = strings.TrimSpace(res)
	if sketchError {
		l.verboseWarnFn(res)
	} else {
		l.verboseInfoFn(res)
	}
	// todo why?? should we remove this?
	time.Sleep(100 * time.Millisecond)
}

// AppendIncludeFolder todo should rename this, probably after refactoring the
// container_find_includes command.
//func (l *SketchLibrariesDetector) AppendIncludeFolder(ctx *types.Context, cache *includeCache, sourceFilePath *paths.Path, include string, folder *paths.Path) {
//	ctx.IncludeFolders = append(ctx.IncludeFolders, folder)
//	cache.ExpectEntry(sourceFilePath, include, folder)
//}

// LibrariesLoader todo
func LibrariesLoader(
	useCachedLibrariesResolution bool,
	librariesManager *librariesmanager.LibrariesManager,
	builtInLibrariesDirs *paths.Path, libraryDirs, otherLibrariesDirs paths.PathList,
	actualPlatform, targetPlatform *cores.PlatformRelease,
) (*librariesmanager.LibrariesManager, *librariesresolver.Cpp, []byte, error) {
	verboseOut := &bytes.Buffer{}
	lm := librariesManager
	if useCachedLibrariesResolution {
		// Since we are using the cached libraries resolution
		// the library manager is not needed.
		lm = librariesmanager.NewLibraryManager(nil, nil)
	}
	if librariesManager == nil {
		lm = librariesmanager.NewLibraryManager(nil, nil)

		builtInLibrariesFolders := builtInLibrariesDirs
		if builtInLibrariesFolders != nil {
			if err := builtInLibrariesFolders.ToAbs(); err != nil {
				return nil, nil, nil, errors.WithStack(err)
			}
			lm.AddLibrariesDir(builtInLibrariesFolders, libraries.IDEBuiltIn)
		}

		if actualPlatform != targetPlatform {
			lm.AddPlatformReleaseLibrariesDir(actualPlatform, libraries.ReferencedPlatformBuiltIn)
		}
		lm.AddPlatformReleaseLibrariesDir(targetPlatform, libraries.PlatformBuiltIn)

		librariesFolders := otherLibrariesDirs
		if err := librariesFolders.ToAbs(); err != nil {
			return nil, nil, nil, errors.WithStack(err)
		}
		for _, folder := range librariesFolders {
			lm.AddLibrariesDir(folder, libraries.User)
		}

		for _, status := range lm.RescanLibraries() {
			// With the refactoring of the initialization step of the CLI we changed how
			// errors are returned when loading platforms and libraries, that meant returning a list of
			// errors instead of a single one to enhance the experience for the user.
			// I have no intention right now to start a refactoring of the legacy package too, so
			// here's this shitty solution for now.
			// When we're gonna refactor the legacy package this will be gone.
			verboseOut.Write([]byte(status.Message()))
		}

		for _, dir := range libraryDirs {
			// Libraries specified this way have top priority
			if err := lm.LoadLibraryFromDir(dir, libraries.Unmanaged); err != nil {
				return nil, nil, nil, errors.WithStack(err)
			}
		}
	}

	resolver := librariesresolver.NewCppResolver()
	if err := resolver.ScanIDEBuiltinLibraries(lm); err != nil {
		return nil, nil, nil, errors.WithStack(err)
	}
	if err := resolver.ScanUserAndUnmanagedLibraries(lm); err != nil {
		return nil, nil, nil, errors.WithStack(err)
	}
	if err := resolver.ScanPlatformLibraries(lm, targetPlatform); err != nil {
		return nil, nil, nil, errors.WithStack(err)
	}
	if actualPlatform != targetPlatform {
		if err := resolver.ScanPlatformLibraries(lm, actualPlatform); err != nil {
			return nil, nil, nil, errors.WithStack(err)
		}
	}
	return lm, resolver, verboseOut.Bytes(), nil
}
