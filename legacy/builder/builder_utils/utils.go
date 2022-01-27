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
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/arduino/arduino-cli/i18n"
	"github.com/arduino/arduino-cli/legacy/builder/constants"
	"github.com/arduino/arduino-cli/legacy/builder/types"
	"github.com/arduino/arduino-cli/legacy/builder/utils"
	"github.com/arduino/go-paths-helper"
	"github.com/arduino/go-properties-orderedmap"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

var tr = i18n.Tr

func CompileFilesRecursive(ctx *types.Context, sourcePath *paths.Path, buildPath *paths.Path, buildProperties *properties.Map, includes []string) (paths.PathList, error) {
	objectFiles, err := CompileFiles(ctx, sourcePath, false, buildPath, buildProperties, includes)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	folders, err := utils.ReadDirFiltered(sourcePath.String(), utils.FilterDirs)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	for _, folder := range folders {
		subFolderObjectFiles, err := CompileFilesRecursive(ctx, sourcePath.Join(folder.Name()), buildPath.Join(folder.Name()), buildProperties, includes)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		objectFiles.AddAll(subFolderObjectFiles)
	}

	return objectFiles, nil
}

func CompileFiles(ctx *types.Context, sourcePath *paths.Path, recurse bool, buildPath *paths.Path, buildProperties *properties.Map, includes []string) (paths.PathList, error) {
	sSources, err := findFilesInFolder(sourcePath, ".S", recurse)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	cSources, err := findFilesInFolder(sourcePath, ".c", recurse)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	cppSources, err := findFilesInFolder(sourcePath, ".cpp", recurse)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	ctx.Progress.AddSubSteps(len(sSources) + len(cSources) + len(cppSources))
	defer ctx.Progress.RemoveSubSteps()

	sObjectFiles, err := compileFilesWithRecipe(ctx, sourcePath, sSources, buildPath, buildProperties, includes, constants.RECIPE_S_PATTERN)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	cObjectFiles, err := compileFilesWithRecipe(ctx, sourcePath, cSources, buildPath, buildProperties, includes, constants.RECIPE_C_PATTERN)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	cppObjectFiles, err := compileFilesWithRecipe(ctx, sourcePath, cppSources, buildPath, buildProperties, includes, constants.RECIPE_CPP_PATTERN)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	objectFiles := paths.NewPathList()
	objectFiles.AddAll(sObjectFiles)
	objectFiles.AddAll(cObjectFiles)
	objectFiles.AddAll(cppObjectFiles)
	return objectFiles, nil
}

func findFilesInFolder(sourcePath *paths.Path, extension string, recurse bool) (paths.PathList, error) {
	files, err := utils.ReadDirFiltered(sourcePath.String(), utils.FilterFilesWithExtensions(extension))
	if err != nil {
		return nil, errors.WithStack(err)
	}
	var sources paths.PathList
	for _, file := range files {
		sources = append(sources, sourcePath.Join(file.Name()))
	}

	if recurse {
		folders, err := utils.ReadDirFiltered(sourcePath.String(), utils.FilterDirs)
		if err != nil {
			return nil, errors.WithStack(err)
		}

		for _, folder := range folders {
			otherSources, err := findFilesInFolder(sourcePath.Join(folder.Name()), extension, recurse)
			if err != nil {
				return nil, errors.WithStack(err)
			}
			sources = append(sources, otherSources...)
		}
	}

	return sources, nil
}

func findAllFilesInFolder(sourcePath string, recurse bool) ([]string, error) {
	files, err := utils.ReadDirFiltered(sourcePath, utils.FilterFiles())
	if err != nil {
		return nil, errors.WithStack(err)
	}
	var sources []string
	for _, file := range files {
		sources = append(sources, filepath.Join(sourcePath, file.Name()))
	}

	if recurse {
		folders, err := utils.ReadDirFiltered(sourcePath, utils.FilterDirs)
		if err != nil {
			return nil, errors.WithStack(err)
		}

		for _, folder := range folders {
			if !utils.IsSCCSOrHiddenFile(folder) {
				// Skip SCCS directories as they do not influence the build and can be very large
				otherSources, err := findAllFilesInFolder(filepath.Join(sourcePath, folder.Name()), recurse)
				if err != nil {
					return nil, errors.WithStack(err)
				}
				sources = append(sources, otherSources...)
			}
		}
	}

	return sources, nil
}

func compileFilesWithRecipe(ctx *types.Context, sourcePath *paths.Path, sources paths.PathList, buildPath *paths.Path, buildProperties *properties.Map, includes []string, recipe string) (paths.PathList, error) {
	objectFiles := paths.NewPathList()
	if len(sources) == 0 {
		return objectFiles, nil
	}
	var objectFilesMux sync.Mutex
	var errorsList []error
	var errorsMux sync.Mutex

	queue := make(chan *paths.Path)
	job := func(source *paths.Path) {
		objectFile, err := compileFileWithRecipe(ctx, sourcePath, source, buildPath, buildProperties, includes, recipe)
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
	properties.SetPath(constants.BUILD_PROPERTIES_SOURCE_FILE, source)
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

	objIsUpToDate, err := ObjFileIsUpToDate(ctx, source, objectFile, depsFile)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	command, err := PrepareCommandForRecipe(properties, recipe, false, ctx.PackageManager.GetEnvVarsForSpawnedProcess())
	if err != nil {
		return nil, errors.WithStack(err)
	}
	if ctx.CompilationDatabase != nil {
		ctx.CompilationDatabase.Add(source, command)
	}
	if !objIsUpToDate && !ctx.OnlyUpdateCompilationDatabase {
		_, _, err = utils.ExecCommand(ctx, command, utils.ShowIfVerbose /* stdout */, utils.Show /* stderr */)
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

func ObjFileIsUpToDate(ctx *types.Context, sourceFile, objectFile, dependencyFile *paths.Path) (bool, error) {
	logrus.Debugf("Checking previous results for %v (result = %v, dep = %v)", sourceFile, objectFile, dependencyFile)
	if objectFile == nil || dependencyFile == nil {
		logrus.Debugf("Not found: nil")
		return false, nil
	}

	sourceFile = sourceFile.Clean()
	sourceFileStat, err := sourceFile.Stat()
	if err != nil {
		return false, errors.WithStack(err)
	}

	objectFile = objectFile.Clean()
	objectFileStat, err := objectFile.Stat()
	if err != nil {
		if os.IsNotExist(err) {
			logrus.Debugf("Not found: %v", objectFile)
			return false, nil
		} else {
			return false, errors.WithStack(err)
		}
	}

	dependencyFile = dependencyFile.Clean()
	dependencyFileStat, err := dependencyFile.Stat()
	if err != nil {
		if os.IsNotExist(err) {
			logrus.Debugf("Not found: %v", dependencyFile)
			return false, nil
		} else {
			return false, errors.WithStack(err)
		}
	}

	if sourceFileStat.ModTime().After(objectFileStat.ModTime()) {
		logrus.Debugf("%v newer than %v", sourceFile, objectFile)
		return false, nil
	}
	if sourceFileStat.ModTime().After(dependencyFileStat.ModTime()) {
		logrus.Debugf("%v newer than %v", sourceFile, dependencyFile)
		return false, nil
	}

	rows, err := dependencyFile.ReadFileAsLines()
	if err != nil {
		return false, errors.WithStack(err)
	}

	rows = utils.Map(rows, removeEndingBackSlash)
	rows = utils.Map(rows, strings.TrimSpace)
	rows = utils.Map(rows, unescapeDep)
	rows = utils.Filter(rows, nonEmptyString)

	if len(rows) == 0 {
		return true, nil
	}

	firstRow := rows[0]
	if !strings.HasSuffix(firstRow, ":") {
		logrus.Debugf("No colon in first line of depfile")
		return false, nil
	}
	objFileInDepFile := firstRow[:len(firstRow)-1]
	if objFileInDepFile != objectFile.String() {
		logrus.Debugf("Depfile is about different file: %v", objFileInDepFile)
		return false, nil
	}

	// The first line of the depfile contains the path to the object file to generate.
	// The second line of the depfile contains the path to the source file.
	// All subsequent lines contain the header files necessary to compile the object file.

	// If we don't do this check it might happen that trying to compile a source file
	// that has the same name but a different path wouldn't recreate the object file.
	if sourceFile.String() != strings.Trim(rows[1], " ") {
		return false, nil
	}

	rows = rows[1:]
	for _, row := range rows {
		depStat, err := os.Stat(row)
		if err != nil && !os.IsNotExist(err) {
			// There is probably a parsing error of the dep file
			// Ignore the error and trigger a full rebuild anyway
			logrus.WithError(err).Debugf("Failed to read: %v", row)
			return false, nil
		}
		if os.IsNotExist(err) {
			logrus.Debugf("Not found: %v", row)
			return false, nil
		}
		if depStat.ModTime().After(objectFileStat.ModTime()) {
			logrus.Debugf("%v newer than %v", row, objectFile)
			return false, nil
		}
	}

	return true, nil
}

func unescapeDep(s string) string {
	s = strings.Replace(s, "\\ ", " ", -1)
	s = strings.Replace(s, "\\\t", "\t", -1)
	s = strings.Replace(s, "\\#", "#", -1)
	s = strings.Replace(s, "$$", "$", -1)
	s = strings.Replace(s, "\\\\", "\\", -1)
	return s
}

func removeEndingBackSlash(s string) string {
	if strings.HasSuffix(s, "\\") {
		s = s[:len(s)-1]
	}
	return s
}

func nonEmptyString(s string) bool {
	return s != constants.EMPTY_STRING
}

func CoreOrReferencedCoreHasChanged(corePath, targetCorePath, targetFile *paths.Path) bool {

	targetFileStat, err := targetFile.Stat()
	if err == nil {
		files, err := findAllFilesInFolder(corePath.String(), true)
		if err != nil {
			return true
		}
		for _, file := range files {
			fileStat, err := os.Stat(file)
			if err != nil || fileStat.ModTime().After(targetFileStat.ModTime()) {
				return true
			}
		}
		if targetCorePath != nil && !strings.EqualFold(corePath.String(), targetCorePath.String()) {
			return CoreOrReferencedCoreHasChanged(targetCorePath, nil, targetFile)
		}
		return false
	}
	return true
}

func TXTBuildRulesHaveChanged(corePath, targetCorePath, targetFile *paths.Path) bool {

	targetFileStat, err := targetFile.Stat()
	if err == nil {
		files, err := findAllFilesInFolder(corePath.String(), true)
		if err != nil {
			return true
		}
		for _, file := range files {
			// report changes only for .txt files
			if filepath.Ext(file) != ".txt" {
				continue
			}
			fileStat, err := os.Stat(file)
			if err != nil || fileStat.ModTime().After(targetFileStat.ModTime()) {
				return true
			}
		}
		if targetCorePath != nil && !corePath.EqualsTo(targetCorePath) {
			return TXTBuildRulesHaveChanged(targetCorePath, nil, targetFile)
		}
		return false
	}
	return true
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

		command, err := PrepareCommandForRecipe(properties, constants.RECIPE_AR_PATTERN, false, ctx.PackageManager.GetEnvVarsForSpawnedProcess())
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

func PrepareCommandForRecipe(buildProperties *properties.Map, recipe string, removeUnsetProperties bool, toolEnv []string) (*exec.Cmd, error) {
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
	command := exec.Command(parts[0], parts[1:]...)
	command.Env = append(command.Env, os.Environ()...)
	command.Env = append(command.Env, toolEnv...)

	// if the overall commandline is too long for the platform
	// try reducing the length by making the filenames relative
	// and changing working directory to build.path
	if len(commandLine) > COMMANDLINE_LIMIT {
		relativePath := buildProperties.Get("build.path")
		for i, arg := range command.Args {
			if _, err := os.Stat(arg); os.IsNotExist(err) {
				continue
			}
			rel, err := filepath.Rel(relativePath, arg)
			if err == nil && !strings.Contains(rel, "..") && len(rel) < len(arg) {
				command.Args[i] = rel
			}
		}
		command.Dir = relativePath
	}

	return command, nil
}
