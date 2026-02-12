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
	"slices"
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
	"github.com/sirupsen/logrus"
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
	preRunner                     *runner.Runner
	detectedChangeInLibraries     bool
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
	logrus.Tracef("[LD] LIBRARY: %s", library.Name)
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

// IncludeFoldersChanged returns true if the include folders list changed
// from the previous compile.
func (l *SketchLibrariesDetector) IncludeFoldersChanged() bool {
	return l.detectedChangeInLibraries
}

// addIncludeFolder add the given folder to the include path.
func (l *SketchLibrariesDetector) addIncludeFolder(folder *paths.Path) {
	logrus.Tracef("[LD] INCLUDE-PATH: %s", folder.String())
	l.includeFolders = append(l.includeFolders, folder)
	l.cache.ExpectAddedIncludePath(folder)
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
	jobs int,
) error {
	logrus.Debug("Finding required libraries for the sketch.")
	defer func() {
		logrus.Debugf("Library detection completed. Found %d required libraries.", len(l.importedLibraries))
	}()

	err := l.findIncludes(ctx, buildPath, buildCorePath, buildVariantPath, sketchBuildPath, sketch, librariesBuildPath, buildProperties, platformArch, jobs)
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
	jobs int,
) error {
	librariesResolutionCachePath := buildPath.Join("libraries.cache")
	var cachedIncludeFolders paths.PathList
	if librariesResolutionCachePath.Exist() {
		d, err := librariesResolutionCachePath.ReadFile()
		if err != nil {
			return err
		}
		if err := json.Unmarshal(d, &cachedIncludeFolders); err != nil {
			return err
		}
	}
	if l.useCachedLibrariesResolution && librariesResolutionCachePath.Exist() {
		l.includeFolders = cachedIncludeFolders
		if l.logger.VerbosityLevel() == logger.VerbosityVerbose {
			l.logger.Info("Using cached library discovery: " + librariesResolutionCachePath.String())
		}
		return nil
	}

	cachePath := buildPath.Join("includes.cache")
	if err := l.cache.Load(cachePath); err != nil {
		l.logger.Warn(i18n.Tr("Failed to load library discovery cache: %[1]s", err))
	}

	// Pre-run cache entries
	l.preRunner = runner.New(ctx, jobs)
	for _, entry := range l.cache.EntriesAhead() {
		if entry.CompileTask != nil {
			upToDate, _ := entry.Compile.ObjFileIsUpToDate(logrus.WithField("runner", "prerun"))
			if !upToDate {
				_ = entry.Compile.PrepareBuildPath()
				l.preRunner.Enqueue(entry.CompileTask)
			}
		}
	}
	defer func() {
		if l.preRunner != nil {
			l.preRunner.Cancel()
		}
	}()

	l.addIncludeFolder(buildCorePath)
	if buildVariantPath != nil {
		l.addIncludeFolder(buildVariantPath)
	}

	sourceFileQueue := &uniqueSourceFileQueue{}

	if !l.useCachedLibrariesResolution {
		sketch := sketch
		mergedfile, err := l.makeSourceFile(sketchBuildPath, sketchBuildPath, paths.New(sketch.MainFile.Base()+".cpp"))
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

			// Create a new pre-runner if the previous one was cancelled
			if l.preRunner == nil {
				l.preRunner = runner.New(ctx, jobs)
				// Push in the remainder of the queue
				for _, sourceFile := range *sourceFileQueue {
					l.preRunner.Enqueue(l.gccPreprocessTask(sourceFile, buildProperties))
				}
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
	} else if err := librariesResolutionCachePath.WriteFile(d); err != nil {
		return err
	}
	l.detectedChangeInLibraries = !slices.Equal(
		cachedIncludeFolders.AsStrings(),
		l.includeFolders.AsStrings())
	return nil
}

func (l *SketchLibrariesDetector) gccPreprocessTask(sourceFile sourceFile, buildProperties *properties.Map) *runner.Task {
	// Libraries may require the "utility" directory to be added to the include
	// search path, but only for the source code of the library, so we temporary
	// copy the current search path list and add the library' utility directory
	// if needed.
	includeFolders := l.includeFolders.Clone()
	if extraInclude := sourceFile.ExtraIncludePath; extraInclude != nil {
		includeFolders = append(includeFolders, extraInclude)
	}

	_ = sourceFile.PrepareBuildPath()
	return preprocessor.GCC(sourceFile.SourcePath, paths.NullPath(), includeFolders, buildProperties, sourceFile.DepfilePath)
}

func (l *SketchLibrariesDetector) findMissingIncludesInCompilationUnit(
	ctx context.Context,
	sourceFileQueue *uniqueSourceFileQueue,
	buildProperties *properties.Map,
	librariesBuildPath *paths.Path,
	platformArch string,
) error {
	sourceFile := sourceFileQueue.Pop()
	sourcePath := sourceFile.SourcePath

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
	unchanged, err := sourceFile.ObjFileIsUpToDate(logrus.WithField("runner", "main"))
	if err != nil {
		return err
	}

	first := true
	for {
		preprocTask := l.gccPreprocessTask(sourceFile, buildProperties)
		l.cache.ExpectCompile(sourceFile, preprocTask)

		var preprocErr error
		var preprocResult *runner.Result

		var missingIncludeH string
		if entry := l.cache.Peek(); unchanged && entry != nil && entry.MissingIncludeH != nil {
			missingIncludeH = *entry.MissingIncludeH
			logrus.Tracef("[LD] COMPILE-CACHE: %s", sourceFile.SourcePath)
			if first && l.logger.VerbosityLevel() == logger.VerbosityVerbose {
				l.logger.Info(i18n.Tr("Using cached library dependencies for file: %[1]s", sourcePath))
			}
			first = false
		} else {
			logrus.Tracef("[LD] COMPILE: %s", sourceFile.SourcePath)
			if l.preRunner != nil {
				if r := l.preRunner.Results(preprocTask); r != nil {
					preprocResult = r
					preprocErr = preprocResult.Error
				}
			}
			if preprocResult == nil {
				// The pre-runner missed this task, maybe the cache is outdated
				// or maybe the source code changed.

				// Stop the pre-runner
				if l.preRunner != nil {
					l.preRunner.Cancel()
					l.preRunner = nil
				}

				// Run the actual preprocessor
				preprocResult = preprocTask.Run(ctx)
				preprocErr = preprocResult.Error
			}
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

		logrus.Tracef("[LD] MISSING: %s", missingIncludeH)
		l.cache.ExpectMissingIncludeH(missingIncludeH)

		if missingIncludeH == "" {
			// No missing includes found, we're done
			return nil
		}

		library := l.resolveLibrary(missingIncludeH, platformArch)
		if library == nil {
			// Library could not be resolved, show error

			// If preprocess result came from cache, run the preprocessor to obtain the actual error to show
			if preprocErr == nil || len(preprocResult.Stderr) == 0 {
				preprocResult = preprocTask.Run(ctx)
				preprocErr = preprocResult.Error
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
	logrus.Tracef("[LD] SCAN: %s (recurse=%v)", folder, recurse)

	sourceFileExtensions := []string{}
	for k := range globals.SourceFilesValidExtensions {
		sourceFileExtensions = append(sourceFileExtensions, k)
	}
	filePaths, err := utils.FindFilesInFolder(folder, recurse, sourceFileExtensions...)
	if err != nil {
		return err
	}

	for _, filePath := range filePaths {
		sourceFile, err := l.makeSourceFile(sourceDir, buildDir, filePath, extraIncludePath...)
		if err != nil {
			return err
		}
		sourceFileQueue.Push(sourceFile)
	}

	return nil
}

// makeSourceFile create a sourceFile object for the given source file path.
// The given sourceFilePath can be absolute, or relative within the sourceRoot root folder.
func (l *SketchLibrariesDetector) makeSourceFile(sourceRoot, buildRoot, sourceFilePath *paths.Path, extraIncludePaths ...*paths.Path) (sourceFile, error) {
	if len(extraIncludePaths) > 1 {
		panic("only one extra include path allowed")
	}
	var extraIncludePath *paths.Path
	if len(extraIncludePaths) > 0 {
		extraIncludePath = extraIncludePaths[0]
	}

	if sourceFilePath.IsAbs() {
		var err error
		sourceFilePath, err = sourceRoot.RelTo(sourceFilePath)
		if err != nil {
			return sourceFile{}, err
		}
	}
	return sourceFile{
		SourcePath:       sourceRoot.JoinPath(sourceFilePath),
		DepfilePath:      buildRoot.Join(fmt.Sprintf("%s.libsdetect.d", sourceFilePath)),
		ExtraIncludePath: extraIncludePath,
	}, nil
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
	builtInLibrariesDir *paths.Path,
	customLibraryDirs paths.PathList, // libraryDirs paths.PathList,
	librariesDirs paths.PathList, // otherLibrariesDirs paths.PathList,
	buildPlatform *cores.PlatformRelease,
	targetPlatform *cores.PlatformRelease,
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

		if builtInLibrariesDir != nil {
			if err := builtInLibrariesDir.ToAbs(); err != nil {
				return nil, nil, nil, err
			}
			lmb.AddLibrariesDir(librariesmanager.LibrariesDir{
				Path:     builtInLibrariesDir,
				Location: libraries.IDEBuiltIn,
			})
		}

		if buildPlatform != targetPlatform {
			lmb.AddLibrariesDir(librariesmanager.LibrariesDir{
				PlatformRelease: buildPlatform,
				Path:            buildPlatform.GetLibrariesDir(),
				Location:        libraries.ReferencedPlatformBuiltIn,
			})
		}
		lmb.AddLibrariesDir(librariesmanager.LibrariesDir{
			PlatformRelease: targetPlatform,
			Path:            targetPlatform.GetLibrariesDir(),
			Location:        libraries.PlatformBuiltIn,
		})

		librariesFolders := librariesDirs
		if err := librariesFolders.ToAbs(); err != nil {
			return nil, nil, nil, err
		}
		for _, folder := range librariesFolders {
			lmb.AddLibrariesDir(librariesmanager.LibrariesDir{
				Path:     folder,
				Location: libraries.User, // XXX: Should be libraries.Unmanaged?
			})
		}

		for _, dir := range customLibraryDirs {
			lmb.AddLibrariesDir(librariesmanager.LibrariesDir{
				Path:            dir,
				Location:        libraries.Unmanaged,
				IsSingleLibrary: true,
			})
		}

		newLm, libsLoadingWarnings := lmb.Build()
		for _, status := range libsLoadingWarnings {
			verboseOut.Write([]byte(status.Message()))
		}
		lm = newLm
	}

	allLibs := lm.FindAllInstalled()
	resolver := librariesresolver.NewCppResolver(allLibs, targetPlatform, buildPlatform)
	return lm, resolver, verboseOut.Bytes(), nil
}
