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

package detector

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/arduino/arduino-cli/internal/arduino/builder/internal/diagnostics"
	"github.com/arduino/arduino-cli/internal/arduino/builder/internal/preprocessor"
	"github.com/arduino/arduino-cli/internal/arduino/builder/internal/runner"
	"github.com/arduino/arduino-cli/internal/arduino/builder/internal/utils"
	"github.com/arduino/arduino-cli/internal/arduino/builder/logger"
	"github.com/arduino/arduino-cli/internal/arduino/cores"
	"github.com/arduino/arduino-cli/internal/arduino/globals"
	"github.com/arduino/arduino-cli/internal/arduino/libraries"
	"github.com/arduino/arduino-cli/internal/arduino/libraries/librariesmanager"
	"github.com/arduino/arduino-cli/internal/arduino/libraries/librariesresolver"
	"github.com/arduino/arduino-cli/internal/arduino/sketch"
	"github.com/arduino/arduino-cli/internal/i18n"
	"github.com/arduino/go-paths-helper"
	"github.com/arduino/go-properties-orderedmap"
)

type libraryResolutionResult struct {
	Library          *libraries.Library
	NotUsedLibraries []*libraries.Library
}

// SketchLibrariesDetector todo
type SketchLibrariesDetector struct {
	librariesManager              *librariesmanager.LibrariesManager
	librariesResolver             *librariesresolver.Cpp
	useCachedLibrariesResolution  bool
	cache                         *detectorCache
	onlyUpdateCompilationDatabase bool
	importedLibraries             libraries.List
	librariesResolutionResults    map[string]libraryResolutionResult
	includeFolders                paths.PathList
	logger                        *logger.BuilderLogger
	diagnosticStore               *diagnostics.Store
}

// NewSketchLibrariesDetector todo
func NewSketchLibrariesDetector(
	lm *librariesmanager.LibrariesManager,
	libsResolver *librariesresolver.Cpp,
	useCachedLibrariesResolution bool,
	onlyUpdateCompilationDatabase bool,
	logger *logger.BuilderLogger,
	diagnosticStore *diagnostics.Store,
) *SketchLibrariesDetector {
	return &SketchLibrariesDetector{
		librariesManager:              lm,
		librariesResolver:             libsResolver,
		useCachedLibrariesResolution:  useCachedLibrariesResolution,
		cache:                         newDetectorCache(),
		librariesResolutionResults:    map[string]libraryResolutionResult{},
		importedLibraries:             libraries.List{},
		includeFolders:                paths.PathList{},
		onlyUpdateCompilationDatabase: onlyUpdateCompilationDatabase,
		logger:                        logger,
		diagnosticStore:               diagnosticStore,
	}
}

// ResolveLibrary todo
func (l *SketchLibrariesDetector) resolveLibrary(header, platformArch string) *libraries.Library {
	importedLibraries := l.importedLibraries
	candidates := l.librariesResolver.AlternativesFor(header)

	if l.logger.VerbosityLevel() == logger.VerbosityVerbose {
		l.logger.Info(i18n.Tr("Alternatives for %[1]s: %[2]s", header, candidates))
		l.logger.Info(fmt.Sprintf("ResolveLibrary(%s)", header))
		l.logger.Info(fmt.Sprintf("  -> %s: %s", i18n.Tr("candidates"), candidates))
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

// addAndBuildLibrary adds the given library to the imported libraries list and queues its source files
// for further processing.
func (l *SketchLibrariesDetector) addAndBuildLibrary(sourceFileQueue *uniqueSourceFileQueue, librariesBuildPath *paths.Path, library *libraries.Library) {
	l.importedLibraries = append(l.importedLibraries, library)
	if library.Precompiled && library.PrecompiledWithSources {
		// Fully precompiled libraries should have no dependencies to avoid ABI breakage
		if l.logger.VerbosityLevel() == logger.VerbosityVerbose {
			l.logger.Info(i18n.Tr("Skipping dependencies detection for precompiled library %[1]s", library.Name))
		}
	} else {
		for _, sourceDir := range library.SourceDirs() {
			l.queueSourceFilesFromFolder(
				sourceFileQueue,
				sourceDir.Dir, sourceDir.Recurse,
				library.SourceDir,
				librariesBuildPath.Join(library.DirName),
				library.UtilityDir)
		}
	}
}

// PrintUsedAndNotUsedLibraries todo
func (l *SketchLibrariesDetector) PrintUsedAndNotUsedLibraries(sketchError bool) {
	// Print this message:
	// - as warning, when the sketch didn't compile
	// - as info, when verbose is on
	// - otherwise, output nothing
	if !sketchError && l.logger.VerbosityLevel() != logger.VerbosityVerbose {
		return
	}

	res := ""
	for header, libResResult := range l.librariesResolutionResults {
		if len(libResResult.NotUsedLibraries) == 0 {
			continue
		}
		res += fmt.Sprintln(i18n.Tr(`Multiple libraries were found for "%[1]s"`, header))
		res += fmt.Sprintln("  " + i18n.Tr("Used: %[1]s", libResResult.Library.InstallDir))
		for _, notUsedLibrary := range libResResult.NotUsedLibraries {
			res += fmt.Sprintln("  " + i18n.Tr("Not used: %[1]s", notUsedLibrary.InstallDir))
		}
	}
	res = strings.TrimSpace(res)
	if sketchError {
		l.logger.Warn(res)
	} else {
		l.logger.Info(res)
	}
	// todo why?? should we remove this?
	time.Sleep(100 * time.Millisecond)
}

// IncludeFolders returns the list of include folders detected as needed.
func (l *SketchLibrariesDetector) IncludeFolders() paths.PathList {
	return l.includeFolders
}

// addIncludeFolder add the given folder to the include path.
func (l *SketchLibrariesDetector) addIncludeFolder(folder *paths.Path) {
	l.includeFolders = append(l.includeFolders, folder)
	l.cache.Expect(&detectorCacheEntry{AddedIncludePath: folder})
}

// FindIncludes todo
func (l *SketchLibrariesDetector) FindIncludes(
	ctx context.Context,
	buildPath *paths.Path,
	buildCorePath *paths.Path,
	buildVariantPath *paths.Path,
	sketchBuildPath *paths.Path,
	sketch *sketch.Sketch,
	librariesBuildPath *paths.Path,
	buildProperties *properties.Map,
	platformArch string,
) error {
	err := l.findIncludes(ctx, buildPath, buildCorePath, buildVariantPath, sketchBuildPath, sketch, librariesBuildPath, buildProperties, platformArch)
	if err != nil && l.onlyUpdateCompilationDatabase {
		l.logger.Info(
			fmt.Sprintf(
				"%s: %s",
				i18n.Tr("An error occurred detecting libraries"),
				i18n.Tr("the compilation database may be incomplete or inaccurate"),
			),
		)
		return nil
	}
	return err
}

func (l *SketchLibrariesDetector) findIncludes(
	ctx context.Context,
	buildPath *paths.Path,
	buildCorePath *paths.Path,
	buildVariantPath *paths.Path,
	sketchBuildPath *paths.Path,
	sketch *sketch.Sketch,
	librariesBuildPath *paths.Path,
	buildProperties *properties.Map,
	platformArch string,
) error {
	librariesResolutionCache := buildPath.Join("libraries.cache")
	if l.useCachedLibrariesResolution && librariesResolutionCache.Exist() {
		d, err := librariesResolutionCache.ReadFile()
		if err != nil {
			return err
		}
		if err := json.Unmarshal(d, &l.includeFolders); err != nil {
			return err
		}
		if l.logger.VerbosityLevel() == logger.VerbosityVerbose {
			l.logger.Info("Using cached library discovery: " + librariesResolutionCache.String())
		}
		return nil
	}

	cachePath := buildPath.Join("includes.cache")
	if err := l.cache.Load(cachePath); err != nil {
		l.logger.Warn(i18n.Tr("Failed to load library discovery cache: %[1]s", err))
	}

	l.addIncludeFolder(buildCorePath)
	if buildVariantPath != nil {
		l.addIncludeFolder(buildVariantPath)
	}

	sourceFileQueue := &uniqueSourceFileQueue{}

	if !l.useCachedLibrariesResolution {
		sketch := sketch
		mergedfile, err := makeSourceFile(sketchBuildPath, sketchBuildPath, paths.New(sketch.MainFile.Base()+".cpp"))
		if err != nil {
			return err
		}
		sourceFileQueue.Push(mergedfile)

		l.queueSourceFilesFromFolder(sourceFileQueue, sketchBuildPath, false /* recurse */, sketchBuildPath, sketchBuildPath)
		srcSubfolderPath := sketchBuildPath.Join("src")
		if srcSubfolderPath.IsDir() {
			l.queueSourceFilesFromFolder(sourceFileQueue, srcSubfolderPath, true /* recurse */, sketchBuildPath, sketchBuildPath)
		}

		allInstalledSorted := l.librariesManager.FindAllInstalled()
		allInstalledSorted.SortByName() // Sort libraries to ensure consistent ordering
		for _, library := range allInstalledSorted {
			if library.Location == libraries.Profile {
				l.logger.Info(i18n.Tr("The library %[1]s has been automatically added from sketch project.", library.Name))
				l.addAndBuildLibrary(sourceFileQueue, librariesBuildPath, library)
				l.addIncludeFolder(library.SourceDir)
			}
		}

		for !sourceFileQueue.Empty() {
			err := l.findMissingIncludesInCompilationUnit(ctx, sourceFileQueue, buildProperties, librariesBuildPath, platformArch)
			if err != nil {
				cachePath.Remove()
				return err
			}
		}

		// Finalize the cache
		if err := l.cache.Save(cachePath); err != nil {
			return err
		}
	}

	if err := l.failIfImportedLibraryIsWrong(); err != nil {
		return err
	}

	if d, err := json.Marshal(l.includeFolders); err != nil {
		return err
	} else if err := librariesResolutionCache.WriteFile(d); err != nil {
		return err
	}

	return nil
}

func (l *SketchLibrariesDetector) findMissingIncludesInCompilationUnit(
	ctx context.Context,
	sourceFileQueue *uniqueSourceFileQueue,
	buildProperties *properties.Map,
	librariesBuildPath *paths.Path,
	platformArch string,
) error {
	sourceFile := sourceFileQueue.Pop()
	sourcePath := sourceFile.SourcePath()
	depPath := sourceFile.DepfilePath()
	objPath := sourceFile.ObjectPath()

	// TODO: This should perhaps also compare against the
	// include.cache file timestamp. Now, it only checks if the file
	// changed after the object file was generated, but if it
	// changed between generating the cache and the object file,
	// this could show the file as unchanged when it really is
	// changed. Changing files during a build isn't really
	// supported, but any problems from it should at least be
	// resolved when doing another build, which is not currently the
	// case.
	// TODO: This reads the dependency file, but the actual building
	// does it again. Should the result be somehow cached? Perhaps
	// remove the object file if it is found to be stale?
	unchanged, err := utils.ObjFileIsUpToDate(sourcePath, objPath, depPath)
	if err != nil {
		return err
	}

	first := true
	for {
		l.cache.Expect(&detectorCacheEntry{CompilingSourcePath: sourcePath})

		// Libraries may require the "utility" directory to be added to the include
		// search path, but only for the source code of the library, so we temporary
		// copy the current search path list and add the library' utility directory
		// if needed.
		includeFolders := l.includeFolders
		if extraInclude := sourceFile.ExtraIncludePath(); extraInclude != nil {
			includeFolders = append(includeFolders, extraInclude)
		}

		var preprocErr error
		var preprocResult *runner.Result

		var missingIncludeH string
		if entry := l.cache.Peek(); unchanged && entry != nil && entry.MissingIncludeH != nil {
			missingIncludeH = *entry.MissingIncludeH
			if first && l.logger.VerbosityLevel() == logger.VerbosityVerbose {
				l.logger.Info(i18n.Tr("Using cached library dependencies for file: %[1]s", sourcePath))
			}
			first = false
		} else {
			preprocResult, preprocErr = preprocessor.GCC(ctx, sourcePath, paths.NullPath(), includeFolders, buildProperties)
			if l.logger.VerbosityLevel() == logger.VerbosityVerbose {
				l.logger.WriteStdout(preprocResult.Stdout)
			}
			// Unwrap error and see if it is an ExitError.
			var exitErr *exec.ExitError
			if preprocErr == nil {
				// Preprocessor successful, done
				missingIncludeH = ""
			} else if isExitErr := errors.As(preprocErr, &exitErr); !isExitErr || len(preprocResult.Stderr) == 0 {
				// Ignore ExitErrors (e.g. gcc returning non-zero status), but bail out on other errors
				return preprocErr
			} else {
				missingIncludeH = IncludesFinderWithRegExp(string(preprocResult.Stderr))
				if missingIncludeH == "" && l.logger.VerbosityLevel() == logger.VerbosityVerbose {
					l.logger.Info(i18n.Tr("Error while detecting libraries included by %[1]s", sourcePath))
				}
			}
		}

		l.cache.Expect(&detectorCacheEntry{MissingIncludeH: &missingIncludeH})

		if missingIncludeH == "" {
			// No missing includes found, we're done
			return nil
		}

		library := l.resolveLibrary(missingIncludeH, platformArch)
		if library == nil {
			// Library could not be resolved, show error

			// If preprocess result came from cache, run the preprocessor to obtain the actual error to show
			if preprocErr == nil || len(preprocResult.Stderr) == 0 {
				preprocResult, preprocErr = preprocessor.GCC(ctx, sourcePath, paths.NullPath(), includeFolders, buildProperties)
				if l.logger.VerbosityLevel() == logger.VerbosityVerbose {
					l.logger.WriteStdout(preprocResult.Stdout)
				}
				if preprocErr == nil {
					// If there is a missing #include in the cache, but running
					// gcc does not reproduce that, there is something wrong.
					// Returning an error here will cause the cache to be
					// deleted, so hopefully the next compilation will succeed.
					return errors.New(i18n.Tr("Internal error in cache"))
				}
			}
			l.diagnosticStore.Parse(preprocResult.Args, preprocResult.Stderr)
			l.logger.WriteStderr(preprocResult.Stderr)
			return preprocErr
		}

		// Add this library to the list of libraries, the
		// include path and queue its source files for further
		// include scanning
		l.addAndBuildLibrary(sourceFileQueue, librariesBuildPath, library)
		l.addIncludeFolder(library.SourceDir)
	}
}

func (l *SketchLibrariesDetector) queueSourceFilesFromFolder(
	sourceFileQueue *uniqueSourceFileQueue,
	folder *paths.Path,
	recurse bool,
	sourceDir *paths.Path,
	buildDir *paths.Path,
	extraIncludePath ...*paths.Path,
) error {
	sourceFileExtensions := []string{}
	for k := range globals.SourceFilesValidExtensions {
		sourceFileExtensions = append(sourceFileExtensions, k)
	}
	filePaths, err := utils.FindFilesInFolder(folder, recurse, sourceFileExtensions...)
	if err != nil {
		return err
	}

	for _, filePath := range filePaths {
		sourceFile, err := makeSourceFile(sourceDir, buildDir, filePath, extraIncludePath...)
		if err != nil {
			return err
		}
		sourceFileQueue.Push(sourceFile)
	}

	return nil
}

func (l *SketchLibrariesDetector) failIfImportedLibraryIsWrong() error {
	if len(l.importedLibraries) == 0 {
		return nil
	}

	for _, library := range l.importedLibraries {
		if !library.IsLegacy {
			if library.InstallDir.Join("arch").IsDir() {
				return errors.New(i18n.Tr("%[1]s folder is no longer supported! See %[2]s for more information", "'arch'", "http://goo.gl/gfFJzU"))
			}
			for _, propName := range libraries.MandatoryProperties {
				if !library.Properties.ContainsKey(propName) {
					return errors.New(i18n.Tr("Missing '%[1]s' from library in %[2]s", propName, library.InstallDir))
				}
			}
			if library.Layout == libraries.RecursiveLayout {
				if library.UtilityDir != nil {
					return errors.New(i18n.Tr("Library can't use both '%[1]s' and '%[2]s' folders. Double check in '%[3]s'.", "src", "utility", library.InstallDir))
				}
			}
		}
	}

	return nil
}

var includeRegexp = regexp.MustCompile(`(?ms)^\s*[0-9 |]*\s*#[ \t]*include\s*[<"](\S+)[">]`)

// IncludesFinderWithRegExp fixdoc
func IncludesFinderWithRegExp(source string) string {
	match := includeRegexp.FindStringSubmatch(source)
	if match != nil {
		return strings.TrimSpace(match[1])
	}
	return findIncludeForOldCompilers(source)
}

func findIncludeForOldCompilers(source string) string {
	lines := strings.Split(source, "\n")
	for _, line := range lines {
		splittedLine := strings.Split(line, ":")
		for i := range splittedLine {
			if strings.Contains(splittedLine[i], "fatal error") {
				return strings.TrimSpace(splittedLine[i+1])
			}
		}
	}
	return ""
}

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
		lm, _ = librariesmanager.NewBuilder().Build()
	}
	if librariesManager == nil {
		lmb := librariesmanager.NewBuilder()

		builtInLibrariesFolders := builtInLibrariesDirs
		if builtInLibrariesFolders != nil {
			if err := builtInLibrariesFolders.ToAbs(); err != nil {
				return nil, nil, nil, err
			}
			lmb.AddLibrariesDir(librariesmanager.LibrariesDir{
				Path:     builtInLibrariesFolders,
				Location: libraries.IDEBuiltIn,
			})
		}

		if actualPlatform != targetPlatform {
			lmb.AddLibrariesDir(librariesmanager.LibrariesDir{
				PlatformRelease: actualPlatform,
				Path:            actualPlatform.GetLibrariesDir(),
				Location:        libraries.ReferencedPlatformBuiltIn,
			})
		}
		lmb.AddLibrariesDir(librariesmanager.LibrariesDir{
			PlatformRelease: targetPlatform,
			Path:            targetPlatform.GetLibrariesDir(),
			Location:        libraries.PlatformBuiltIn,
		})

		librariesFolders := otherLibrariesDirs
		if err := librariesFolders.ToAbs(); err != nil {
			return nil, nil, nil, err
		}
		for _, folder := range librariesFolders {
			lmb.AddLibrariesDir(librariesmanager.LibrariesDir{
				Path:     folder,
				Location: libraries.User, // XXX: Should be libraries.Unmanaged?
			})
		}

		for _, dir := range libraryDirs {
			lmb.AddLibrariesDir(librariesmanager.LibrariesDir{
				Path:            dir,
				Location:        libraries.Unmanaged,
				IsSingleLibrary: true,
			})
		}

		newLm, libsLoadingWarnings := lmb.Build()
		for _, status := range libsLoadingWarnings {
			// With the refactoring of the initialization step of the CLI we changed how
			// errors are returned when loading platforms and libraries, that meant returning a list of
			// errors instead of a single one to enhance the experience for the user.
			// I have no intention right now to start a refactoring of the legacy package too, so
			// here's this shitty solution for now.
			// When we're gonna refactor the legacy package this will be gone.
			verboseOut.Write([]byte(status.Message()))
		}
		lm = newLm
	}

	allLibs := lm.FindAllInstalled()
	resolver := librariesresolver.NewCppResolver(allLibs, targetPlatform, actualPlatform)
	return lm, resolver, verboseOut.Bytes(), nil
}
