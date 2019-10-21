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

/*

Include detection

This code is responsible for figuring out what libraries the current
sketch needs an populates both Context.ImportedLibraries with a list of
Library objects, and Context.IncludeFolders with a list of folders to
put on the include path.

Simply put, every #include in a source file pulls in the library that
provides that source file. This includes source files in the selected
libraries, so libraries can recursively include other libraries as well.

To implement this, the gcc preprocessor is used. A queue is created
containing, at first, the source files in the sketch. Each of the files
in the queue is processed in turn by running the preprocessor on it. If
the preprocessor provides an error, the output is examined to see if the
error is a missing header file originating from a #include directive.

The filename is extracted from that #include directive, and a library is
found that provides it. If multiple libraries provide the same file, one
is slected (how this selection works is not described here, see the
ResolveLibrary function for that). The library selected in this way is
added to the include path through Context.IncludeFolders and the list of
libraries to include in the link through Context.ImportedLibraries.

Furthermore, all of the library source files are added to the queue, to
be processed as well. When the preprocessor completes without showing an
#include error, processing of the file is complete and it advances to
the next. When no library can be found for a included filename, an error
is shown and the process is aborted.

Caching

Since this process is fairly slow (requiring at least one invocation of
the preprocessor per source file), its results are cached.

Just caching the complete result (i.e. the resulting list of imported
libraries) seems obvious, but such a cache is hard to invalidate. Making
a list of all the source and header files used to create the list and
check if any of them changed is probably feasible, but this would also
require caching the full list of libraries to invalidate the cache when
the include to library resolution might have a different result. Another
downside of a complete cache is that any changes requires re-running
everything, even if no includes were actually changed.

Instead, caching happens by keeping a sort of "journal" of the steps in
the include detection, essentially tracing each file processed and each
include path entry added. The cache is used by retracing these steps:
The include detection process is executed normally, except that instead
of running the preprocessor, the include filenames are (when possible)
read from the cache. Then, the include file to library resolution is
again executed normally. The results are checked against the cache and
as long as the results match, the cache is considered valid.

When a source file (or any of the files it includes, as indicated by the
.d file) is changed, the preprocessor is executed as normal for the
file, ignoring any includes from the cache. This does not, however,
invalidate the cache: If the results from the preprocessor match the
entries in the cache, the cache remains valid and can again be used for
the next (unchanged) file.

The cache file uses the JSON format and contains a list of entries. Each
entry represents a discovered library and contains:
 - Sourcefile: The source file that the include was found in
 - Include: The included filename found
 - Includepath: The addition to the include path

There are also some special entries:
 - When adding the initial include path entries, such as for the core
   and variant paths.  These are not discovered, so the Sourcefile and
   Include fields will be empty.
 - When a file contains no (more) missing includes, an entry with an
   empty Include and IncludePath is generated.

*/

package builder

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"time"

	"github.com/arduino/arduino-cli/arduino/libraries"
	"github.com/arduino/arduino-cli/arduino/sketch"
	"github.com/arduino/arduino-cli/legacy/builder/builder_utils"
	"github.com/arduino/arduino-cli/legacy/builder/types"
	"github.com/arduino/arduino-cli/legacy/builder/utils"
	"github.com/arduino/go-paths-helper"
	"github.com/pkg/errors"
)

type ContainerFindIncludes struct{}

func (s *ContainerFindIncludes) Run(ctx *types.Context) error {
	err := s.findIncludes(ctx)
	if err != nil && ctx.OnlyUpdateCompilationDatabase {
		ctx.Info(
			fmt.Sprintf("%s: %s",
				tr("An error occurred detecting libraries"),
				tr("the compilation database may be incomplete or inaccurate")))
		return nil
	}
	return err
}

func (s *ContainerFindIncludes) findIncludes(ctx *types.Context) error {
	finder := &CppIncludesFinder{
		ctx: ctx,
	}
	if err := finder.DetectLibraries(); err != nil {
		return err
	}
	if err := runCommand(ctx, &FailIfImportedLibraryIsWrong{}); err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// CppIncludesFinder implements an algorithm to automatically detect
// libraries used in a sketch and a way to cache this result for
// increasing detection speed on already processed sketches.
type CppIncludesFinder struct {
	ctx    *types.Context
	cache  *includeCache
	sketch *sketch.Sketch
	queue  *UniqueSourceFileQueue
}

func (f *CppIncludesFinder) DetectLibraries() error {
	f.cache = loadCacheFrom(f.ctx.BuildPath.Join("includes.cache"))
	f.sketch = f.ctx.Sketch
	f.queue = &UniqueSourceFileQueue{}

	f.appendIncludeFolder(nil, "", f.ctx.BuildProperties.GetPath("build.core.path"))
	if f.ctx.BuildProperties.Get("build.variant.path") != "" {
		f.appendIncludeFolder(nil, "", f.ctx.BuildProperties.GetPath("build.variant.path"))
	}

	mergedfile, err := MakeSourceFile(f.ctx, nil, paths.New(f.sketch.MainFile.Base()+".cpp"))
	if err != nil {
		return errors.WithStack(err)
	}
	f.queue.Push(mergedfile)

	f.queueSourceFilesFromFolder(nil, f.ctx.SketchBuildPath, false /* recurse */)
	srcSubfolderPath := f.ctx.SketchBuildPath.Join("src")
	if srcSubfolderPath.IsDir() {
		f.queueSourceFilesFromFolder(nil, srcSubfolderPath, true /* recurse */)
	}

	for !f.queue.Empty() {
		if err := f.findIncludesUntilDone(f.queue.Pop()); err != nil {
			f.cache.Remove()
			return errors.WithStack(err)
		}
	}

	// Finalize the cache
	f.cache.ExpectEnd()
	if err := f.cache.WriteToFile(); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// Append the given folder to the include path and match or append it to
// the cache. sourceFilePath and include indicate the source of this
// include (e.g. what #include line in what file it was resolved from)
// and should be the empty string for the default include folders, like
// the core or variant.
func (f *CppIncludesFinder) appendIncludeFolder(sourceFilePath *paths.Path, include string, folder *paths.Path) {
	f.ctx.IncludeFolders = append(f.ctx.IncludeFolders, folder)
	f.cache.ExpectEntry(sourceFilePath, include, folder)
}

func runCommand(ctx *types.Context, command types.Command) error {
	PrintRingNameIfDebug(ctx, command)
	err := command.Run(ctx)
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

type includeCacheEntry struct {
	Sourcefile  *paths.Path
	Include     string
	Includepath *paths.Path
}

func (entry *includeCacheEntry) String() string {
	return fmt.Sprintf("SourceFile: %s; Include: %s; IncludePath: %s",
		entry.Sourcefile, entry.Include, entry.Includepath)
}

func (entry *includeCacheEntry) Equals(other *includeCacheEntry) bool {
	return entry.String() == other.String()
}

type includeCache struct {
	// Are the cache contents valid so far?
	valid bool

	// The file where to save the cache
	cacheFilePath *paths.Path

	// Index into entries of the next entry to be processed. Unused
	// when the cache is invalid.
	next    int
	entries []*includeCacheEntry
}

// Return the next cache entry. Should only be called when the cache is
// valid and a next entry is available (the latter can be checked with
// ExpectFile). Does not advance the cache.
func (cache *includeCache) Next() *includeCacheEntry {
	return cache.entries[cache.next]
}

// Check that the next cache entry is about the given file. If it is
// not, or no entry is available, the cache is invalidated. Does not
// advance the cache.
func (cache *includeCache) ExpectFile(sourcefile *paths.Path) {
	if cache.valid && (cache.next >= len(cache.entries) || !cache.Next().Sourcefile.EqualsTo(sourcefile)) {
		cache.valid = false
		cache.entries = cache.entries[:cache.next]
	}
}

// Check that the next entry matches the given values. If so, advance
// the cache. If not, the cache is invalidated. If the cache is
// invalidated, or was already invalid, an entry with the given values
// is appended.
func (cache *includeCache) ExpectEntry(sourcefile *paths.Path, include string, librarypath *paths.Path) {
	entry := &includeCacheEntry{Sourcefile: sourcefile, Include: include, Includepath: librarypath}
	if cache.valid {
		if cache.next < len(cache.entries) && cache.Next().Equals(entry) {
			cache.next++
		} else {
			cache.valid = false
			cache.entries = cache.entries[:cache.next]
		}
	}

	if !cache.valid {
		cache.entries = append(cache.entries, entry)
	}
}

// Check that the cache is completely consumed. If not, the cache is
// invalidated.
func (cache *includeCache) ExpectEnd() {
	if cache.valid && cache.next < len(cache.entries) {
		cache.valid = false
		cache.entries = cache.entries[:cache.next]
	}
}

// Remove removes the cache file from disk.
func (cache *includeCache) Remove() error {
	return cache.cacheFilePath.Remove()
}

// loadCacheFrom read the cache from the given file
func loadCacheFrom(path *paths.Path) *includeCache {
	result := &includeCache{
		cacheFilePath: path,
		valid:         false,
	}
	if bytes, err := path.ReadFile(); err != nil {
		return result
	} else if err = json.Unmarshal(bytes, &result.entries); err != nil {
		return result
	}
	result.valid = true
	return result
}

// WriteToFile the cache file if it is invalidated. If the
// cache is still valid, just update the timestamps of the file.
func (cache *includeCache) WriteToFile() error {
	// If the cache was still valid all the way, just touch its file
	// (in case any source file changed without influencing the
	// includes). If it was invalidated, overwrite the cache with
	// the new contents.
	if cache.valid {
		cache.cacheFilePath.Chtimes(time.Now(), time.Now())
	} else {
		if bytes, err := json.MarshalIndent(cache.entries, "", "  "); err != nil {
			return errors.WithStack(err)
		} else if err := cache.cacheFilePath.WriteFile(bytes); err != nil {
			return errors.WithStack(err)
		}
	}
	return nil
}

func (f *CppIncludesFinder) findIncludesUntilDone(sourceFile SourceFile) error {
	sourcePath := sourceFile.SourcePath(f.ctx)
	targetFilePath := paths.NullPath()
	depPath := sourceFile.DepfilePath(f.ctx)
	objPath := sourceFile.ObjectPath(f.ctx)

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
	unchanged, err := builder_utils.ObjFileIsUpToDate(f.ctx, sourcePath, objPath, depPath)
	if err != nil {
		return errors.WithStack(err)
	}

	first := true
	for {
		var include string
		f.cache.ExpectFile(sourcePath)

		includes := f.ctx.IncludeFolders

		var preproc_err error
		var preproc_stderr []byte
		if unchanged && f.cache.valid {
			include = f.cache.Next().Include
			if first && f.ctx.Verbose {
				f.ctx.Info(tr("Using cached library dependencies for file: %[1]s", sourcePath))
			}
		} else {
			preproc_stderr, preproc_err = GCCPreprocRunnerForDiscoveringIncludes(f.ctx, sourcePath, targetFilePath, includes)
			// Unwrap error and see if it is an ExitError.
			_, is_exit_error := errors.Cause(preproc_err).(*exec.ExitError)
			if preproc_err == nil {
				// Preprocessor successful, done
				include = ""
			} else if !is_exit_error || preproc_stderr == nil {
				// Ignore ExitErrors (e.g. gcc returning
				// non-zero status), but bail out on
				// other errors
				return errors.WithStack(preproc_err)
			} else {
				include = IncludesFinderWithRegExp(string(preproc_stderr))
				if include == "" && f.ctx.Verbose {
					f.ctx.Info(tr("Error while detecting libraries included by %[1]s", sourcePath))
				}
			}
		}

		if include == "" {
			// No missing includes found, we're done
			f.cache.ExpectEntry(sourcePath, "", nil)
			return nil
		}

		library := ResolveLibrary(f.ctx, include)
		if library == nil {
			// Library could not be resolved, show error
			// err := runCommand(ctx, &GCCPreprocRunner{SourceFilePath: sourcePath, TargetFileName: paths.New(constants.FILE_CTAGS_TARGET_FOR_GCC_MINUS_E), Includes: includes})
			// return errors.WithStack(err)
			if preproc_err == nil || preproc_stderr == nil {
				// Filename came from cache, so run preprocessor to obtain error to show
				preproc_stderr, preproc_err = GCCPreprocRunnerForDiscoveringIncludes(f.ctx, sourcePath, targetFilePath, includes)
				if preproc_err == nil {
					// If there is a missing #include in the cache, but running
					// gcc does not reproduce that, there is something wrong.
					// Returning an error here will cause the cache to be
					// deleted, so hopefully the next compilation will succeed.
					return errors.New(tr("Internal error in cache"))
				}
			}
			f.ctx.Stderr.Write(preproc_stderr)
			return errors.WithStack(preproc_err)
		}

		// Add this library to the list of libraries, the
		// include path and queue its source files for further
		// include scanning
		f.ctx.ImportedLibraries = append(f.ctx.ImportedLibraries, library)
		f.appendIncludeFolder(sourcePath, include, library.SourceDir)
		if library.UtilityDir != nil {
			// TODO: Use library.SourceDirs() instead?
			includes = append(includes, library.UtilityDir)
		}
		sourceDirs := library.SourceDirs()
		for _, sourceDir := range sourceDirs {
			if library.Precompiled && library.PrecompiledWithSources {
				// Fully precompiled libraries should have no dependencies
				// to avoid ABI breakage
				if f.ctx.Verbose {
					f.ctx.Info(tr("Skipping dependencies detection for precompiled library %[1]s", library.Name))
				}
			} else {
				f.queueSourceFilesFromFolder(library, sourceDir.Dir, sourceDir.Recurse)
			}
		}
		first = false
	}
}

func (f *CppIncludesFinder) queueSourceFilesFromFolder(lib *libraries.Library, folder *paths.Path, recurse bool) error {
	extensions := func(ext string) bool { return ADDITIONAL_FILE_VALID_EXTENSIONS_NO_HEADERS[ext] }

	filePaths := []string{}
	err := utils.FindFilesInFolder(&filePaths, folder.String(), extensions, recurse)
	if err != nil {
		return errors.WithStack(err)
	}

	for _, filePath := range filePaths {
		sourceFile, err := MakeSourceFile(f.ctx, lib, paths.New(filePath))
		if err != nil {
			return errors.WithStack(err)
		}
		f.queue.Push(sourceFile)
	}

	return nil
}

type SourceFile struct {
	// Library pointer that this source file lives in or nil if not part of a library
	Library *libraries.Library

	// Path to the source file within the sketch/library root folder
	RelativePath *paths.Path
}

// Create a SourceFile containing the given source file path within the
// given origin. The given path can be absolute, or relative within the
// origin's root source folder
func MakeSourceFile(ctx *types.Context, lib *libraries.Library, path *paths.Path) (SourceFile, error) {
	if path.IsAbs() {
		var err error
		path, err = sourceRoot(ctx, lib).RelTo(path)
		if err != nil {
			return SourceFile{}, err
		}
	}
	return SourceFile{Library: lib, RelativePath: path}, nil
}

// Return the build root for the given origin, where build products will
// be placed. Any directories inside SourceFile.RelativePath will be
// appended here.
func buildRoot(ctx *types.Context, lib *libraries.Library) *paths.Path {
	if lib == nil {
		return ctx.SketchBuildPath
	}
	return ctx.LibrariesBuildPath.Join(lib.Name)
}

// Return the source root for the given origin, where its source files
// can be found. Prepending this to SourceFile.RelativePath will give
// the full path to that source file.
func sourceRoot(ctx *types.Context, lib *libraries.Library) *paths.Path {
	if lib == nil {
		return ctx.SketchBuildPath
	}
	return lib.SourceDir
}

func (f *SourceFile) SourcePath(ctx *types.Context) *paths.Path {
	return sourceRoot(ctx, f.Library).JoinPath(f.RelativePath)
}

func (f *SourceFile) ObjectPath(ctx *types.Context) *paths.Path {
	return buildRoot(ctx, f.Library).Join(f.RelativePath.String() + ".o")
}

func (f *SourceFile) DepfilePath(ctx *types.Context) *paths.Path {
	return buildRoot(ctx, f.Library).Join(f.RelativePath.String() + ".d")
}

type UniqueSourceFileQueue struct {
	queue []SourceFile
	curr  int
}

func (q *UniqueSourceFileQueue) Len() int {
	return len(q.queue) - q.curr
}

func (q *UniqueSourceFileQueue) Push(value SourceFile) {
	if !q.Contains(value) {
		q.queue = append(q.queue, value)
	}
}

func (q *UniqueSourceFileQueue) Pop() SourceFile {
	res := q.queue[q.curr]
	q.curr++
	return res
}

func (q *UniqueSourceFileQueue) Empty() bool {
	return q.Len() == 0
}

func (q *UniqueSourceFileQueue) Contains(target SourceFile) bool {
	for _, elem := range q.queue {
		if elem.Library == target.Library && elem.RelativePath.EqualsTo(target.RelativePath) {
			return true
		}
	}
	return false
}
