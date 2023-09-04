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

package builder_utils

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	bUtils "github.com/arduino/arduino-cli/arduino/builder/utils"
	"github.com/arduino/arduino-cli/arduino/globals"
	"github.com/arduino/arduino-cli/executils"
	"github.com/arduino/arduino-cli/i18n"
	"github.com/arduino/arduino-cli/legacy/builder/constants"
	"github.com/arduino/arduino-cli/legacy/builder/types"
	"github.com/arduino/arduino-cli/legacy/builder/utils"
	"github.com/arduino/go-paths-helper"
	"github.com/arduino/go-properties-orderedmap"
	"github.com/pkg/errors"
)

var tr = i18n.Tr

// DirContentIsOlderThan returns true if the content of the given directory is
// older than target file. If extensions are given, only the files with these
// extensions are tested.
func DirContentIsOlderThan(dir *paths.Path, target *paths.Path, extensions ...string) (bool, error) {
	targetStat, err := target.Stat()
	if err != nil {
		return false, err
	}
	targetModTime := targetStat.ModTime()

	files, err := bUtils.FindFilesInFolder(dir, true, extensions...)
	if err != nil {
		return false, err
	}
	for _, file := range files {
		file, err := file.Stat()
		if err != nil {
			return false, err
		}
		if file.ModTime().After(targetModTime) {
			return false, nil
		}
	}
	return true, nil
}

func CompileFiles(ctx *types.Context, sourceDir *paths.Path, buildPath *paths.Path, buildProperties *properties.Map, includes []string) (paths.PathList, error) {
	return compileFiles(ctx, sourceDir, false, buildPath, buildProperties, includes)
}

func CompileFilesRecursive(ctx *types.Context, sourceDir *paths.Path, buildPath *paths.Path, buildProperties *properties.Map, includes []string) (paths.PathList, error) {
	return compileFiles(ctx, sourceDir, true, buildPath, buildProperties, includes)
}

func compileFiles(ctx *types.Context, sourceDir *paths.Path, recurse bool, buildPath *paths.Path, buildProperties *properties.Map, includes []string) (paths.PathList, error) {
	validExtensions := []string{}
	for ext := range globals.SourceFilesValidExtensions {
		validExtensions = append(validExtensions, ext)
	}

	sources, err := bUtils.FindFilesInFolder(sourceDir, recurse, validExtensions...)
	if err != nil {
		return nil, err
	}

	ctx.Progress.AddSubSteps(len(sources))
	defer ctx.Progress.RemoveSubSteps()

	objectFiles := paths.NewPathList()
	var objectFilesMux sync.Mutex
	if len(sources) == 0 {
		return objectFiles, nil
	}
	var errorsList []error
	var errorsMux sync.Mutex

	queue := make(chan *paths.Path)
	job := func(source *paths.Path) {
		recipe := fmt.Sprintf("recipe%s.o.pattern", source.Ext())
		if !buildProperties.ContainsKey(recipe) {
			recipe = fmt.Sprintf("recipe%s.o.pattern", globals.SourceFilesValidExtensions[source.Ext()])
		}
		objectFile, err := compileFileWithRecipe(ctx, sourceDir, source, buildPath, buildProperties, includes, recipe)
		if err != nil {
			errorsMux.Lock()
			errorsList = append(errorsList, err)
			errorsMux.Unlock()
		} else {
			objectFilesMux.Lock()
			objectFiles.Add(objectFile)
			objectFilesMux.Unlock()
		}
	}

	// Spawn jobs runners
	var wg sync.WaitGroup
	jobs := ctx.Jobs
	if jobs == 0 {
		jobs = runtime.NumCPU()
	}
	for i := 0; i < jobs; i++ {
		wg.Add(1)
		go func() {
			for source := range queue {
				job(source)
			}
			wg.Done()
		}()
	}

	// Feed jobs until error or done
	for _, source := range sources {
		errorsMux.Lock()
		gotError := len(errorsList) > 0
		errorsMux.Unlock()
		if gotError {
			break
		}
		queue <- source

		ctx.Progress.CompleteStep()
		ctx.PushProgress()
	}
	close(queue)
	wg.Wait()
	if len(errorsList) > 0 {
		// output the first error
		return nil, errors.WithStack(errorsList[0])
	}
	objectFiles.Sort()
	return objectFiles, nil
}

func compileFileWithRecipe(ctx *types.Context, sourcePath *paths.Path, source *paths.Path, buildPath *paths.Path, buildProperties *properties.Map, includes []string, recipe string) (*paths.Path, error) {
	properties := buildProperties.Clone()
	properties.Set(constants.BUILD_PROPERTIES_COMPILER_WARNING_FLAGS, properties.Get(constants.BUILD_PROPERTIES_COMPILER_WARNING_FLAGS+"."+ctx.WarningsLevel))
	properties.Set(constants.BUILD_PROPERTIES_INCLUDES, strings.Join(includes, constants.SPACE))
	properties.SetPath("source_file", source)
	relativeSource, err := sourcePath.RelTo(source)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	depsFile := buildPath.Join(relativeSource.String() + ".d")
	objectFile := buildPath.Join(relativeSource.String() + ".o")

	properties.SetPath(constants.BUILD_PROPERTIES_OBJECT_FILE, objectFile)
	err = objectFile.Parent().MkdirAll()
	if err != nil {
		return nil, errors.WithStack(err)
	}

	objIsUpToDate, err := bUtils.ObjFileIsUpToDate(source, objectFile, depsFile)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	command, err := PrepareCommandForRecipe(properties, recipe, false)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	if ctx.CompilationDatabase != nil {
		ctx.CompilationDatabase.Add(source, command)
	}
	if !objIsUpToDate && !ctx.OnlyUpdateCompilationDatabase {
		// Since this compile could be multithreaded, we first capture the command output
		stdout, stderr, err := utils.ExecCommand(ctx, command, utils.Capture, utils.Capture)
		// and transfer all at once at the end...
		if ctx.Verbose {
			ctx.WriteStdout(stdout)
		}
		ctx.WriteStderr(stderr)

		// ...and then return the error
		if err != nil {
			return nil, errors.WithStack(err)
		}
	} else if ctx.Verbose {
		if objIsUpToDate {
			ctx.Info(tr("Using previously compiled file: %[1]s", objectFile))
		} else {
			ctx.Info(tr("Skipping compile of: %[1]s", objectFile))
		}
	}

	return objectFile, nil
}

func ArchiveCompiledFiles(ctx *types.Context, buildPath *paths.Path, archiveFile *paths.Path, objectFilesToArchive paths.PathList, buildProperties *properties.Map) (*paths.Path, error) {
	archiveFilePath := buildPath.JoinPath(archiveFile)

	if ctx.OnlyUpdateCompilationDatabase {
		if ctx.Verbose {
			ctx.Info(tr("Skipping archive creation of: %[1]s", archiveFilePath))
		}
		return archiveFilePath, nil
	}

	if archiveFileStat, err := archiveFilePath.Stat(); err == nil {
		rebuildArchive := false
		for _, objectFile := range objectFilesToArchive {
			objectFileStat, err := objectFile.Stat()
			if err != nil || objectFileStat.ModTime().After(archiveFileStat.ModTime()) {
				// need to rebuild the archive
				rebuildArchive = true
				break
			}
		}

		// something changed, rebuild the core archive
		if rebuildArchive {
			if err := archiveFilePath.Remove(); err != nil {
				return nil, errors.WithStack(err)
			}
		} else {
			if ctx.Verbose {
				ctx.Info(tr("Using previously compiled file: %[1]s", archiveFilePath))
			}
			return archiveFilePath, nil
		}
	}

	for _, objectFile := range objectFilesToArchive {
		properties := buildProperties.Clone()
		properties.Set(constants.BUILD_PROPERTIES_ARCHIVE_FILE, archiveFilePath.Base())
		properties.SetPath(constants.BUILD_PROPERTIES_ARCHIVE_FILE_PATH, archiveFilePath)
		properties.SetPath(constants.BUILD_PROPERTIES_OBJECT_FILE, objectFile)

		command, err := PrepareCommandForRecipe(properties, constants.RECIPE_AR_PATTERN, false)
		if err != nil {
			return nil, errors.WithStack(err)
		}

		_, _, err = utils.ExecCommand(ctx, command, utils.ShowIfVerbose /* stdout */, utils.Show /* stderr */)
		if err != nil {
			return nil, errors.WithStack(err)
		}
	}

	return archiveFilePath, nil
}

const COMMANDLINE_LIMIT = 30000

func PrepareCommandForRecipe(buildProperties *properties.Map, recipe string, removeUnsetProperties bool) (*executils.Process, error) {
	pattern := buildProperties.Get(recipe)
	if pattern == "" {
		return nil, errors.Errorf(tr("%[1]s pattern is missing"), recipe)
	}

	commandLine := buildProperties.ExpandPropsInString(pattern)
	if removeUnsetProperties {
		commandLine = properties.DeleteUnexpandedPropsFromString(commandLine)
	}

	parts, err := properties.SplitQuotedString(commandLine, `"'`, false)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	// if the overall commandline is too long for the platform
	// try reducing the length by making the filenames relative
	// and changing working directory to build.path
	var relativePath string
	if len(commandLine) > COMMANDLINE_LIMIT {
		relativePath = buildProperties.Get("build.path")
		for i, arg := range parts {
			if _, err := os.Stat(arg); os.IsNotExist(err) {
				continue
			}
			rel, err := filepath.Rel(relativePath, arg)
			if err == nil && !strings.Contains(rel, "..") && len(rel) < len(arg) {
				parts[i] = rel
			}
		}
	}

	command, err := executils.NewProcess(nil, parts...)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	if relativePath != "" {
		command.SetDir(relativePath)
	}

	return command, nil
}
