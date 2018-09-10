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

package builder_utils

import (
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/arduino/arduino-builder/constants"
	"github.com/arduino/arduino-builder/i18n"
	"github.com/arduino/arduino-builder/types"
	"github.com/arduino/arduino-builder/utils"
	"github.com/arduino/go-paths-helper"
	"github.com/arduino/go-properties-map"
)

func PrintProgressIfProgressEnabledAndMachineLogger(ctx *types.Context) {

	if !ctx.Progress.PrintEnabled {
		return
	}

	log := ctx.GetLogger()
	if log.Name() == "machine" {
		log.Println(constants.LOG_LEVEL_INFO, constants.MSG_PROGRESS, strconv.FormatFloat(ctx.Progress.Progress, 'f', 2, 32))
		ctx.Progress.Progress += ctx.Progress.Steps
	}
}

func CompileFilesRecursive(ctx *types.Context, sourcePath *paths.Path, buildPath *paths.Path, buildProperties properties.Map, includes []string) (paths.PathList, error) {
	objectFiles, err := CompileFiles(ctx, sourcePath, false, buildPath, buildProperties, includes)
	if err != nil {
		return nil, i18n.WrapError(err)
	}

	folders, err := utils.ReadDirFiltered(sourcePath.String(), utils.FilterDirs)
	if err != nil {
		return nil, i18n.WrapError(err)
	}

	for _, folder := range folders {
		subFolderObjectFiles, err := CompileFilesRecursive(ctx, sourcePath.Join(folder.Name()), buildPath.Join(folder.Name()), buildProperties, includes)
		if err != nil {
			return nil, i18n.WrapError(err)
		}
		objectFiles.AddAll(subFolderObjectFiles)
	}

	return objectFiles, nil
}

func CompileFiles(ctx *types.Context, sourcePath *paths.Path, recurse bool, buildPath *paths.Path, buildProperties properties.Map, includes []string) (paths.PathList, error) {
	sObjectFiles, err := compileFilesWithExtensionWithRecipe(ctx, sourcePath, recurse, buildPath, buildProperties, includes, ".S", constants.RECIPE_S_PATTERN)
	if err != nil {
		return nil, i18n.WrapError(err)
	}
	cObjectFiles, err := compileFilesWithExtensionWithRecipe(ctx, sourcePath, recurse, buildPath, buildProperties, includes, ".c", constants.RECIPE_C_PATTERN)
	if err != nil {
		return nil, i18n.WrapError(err)
	}
	cppObjectFiles, err := compileFilesWithExtensionWithRecipe(ctx, sourcePath, recurse, buildPath, buildProperties, includes, ".cpp", constants.RECIPE_CPP_PATTERN)
	if err != nil {
		return nil, i18n.WrapError(err)
	}
	objectFiles := paths.NewPathList()
	objectFiles.AddAll(sObjectFiles)
	objectFiles.AddAll(cObjectFiles)
	objectFiles.AddAll(cppObjectFiles)
	return objectFiles, nil
}

func compileFilesWithExtensionWithRecipe(ctx *types.Context, sourcePath *paths.Path, recurse bool, buildPath *paths.Path, buildProperties properties.Map, includes []string, extension string, recipe string) (paths.PathList, error) {
	sources, err := findFilesInFolder(sourcePath, extension, recurse)
	if err != nil {
		return nil, i18n.WrapError(err)
	}
	return compileFilesWithRecipe(ctx, sourcePath, sources, buildPath, buildProperties, includes, recipe)
}

func findFilesInFolder(sourcePath *paths.Path, extension string, recurse bool) (paths.PathList, error) {
	files, err := utils.ReadDirFiltered(sourcePath.String(), utils.FilterFilesWithExtensions(extension))
	if err != nil {
		return nil, i18n.WrapError(err)
	}
	var sources paths.PathList
	for _, file := range files {
		sources = append(sources, sourcePath.Join(file.Name()))
	}

	if recurse {
		folders, err := utils.ReadDirFiltered(sourcePath.String(), utils.FilterDirs)
		if err != nil {
			return nil, i18n.WrapError(err)
		}

		for _, folder := range folders {
			otherSources, err := findFilesInFolder(sourcePath.Join(folder.Name()), extension, recurse)
			if err != nil {
				return nil, i18n.WrapError(err)
			}
			sources = append(sources, otherSources...)
		}
	}

	return sources, nil
}

func findAllFilesInFolder(sourcePath string, recurse bool) ([]string, error) {
	files, err := utils.ReadDirFiltered(sourcePath, utils.FilterFiles())
	if err != nil {
		return nil, i18n.WrapError(err)
	}
	var sources []string
	for _, file := range files {
		sources = append(sources, filepath.Join(sourcePath, file.Name()))
	}

	if recurse {
		folders, err := utils.ReadDirFiltered(sourcePath, utils.FilterDirs)
		if err != nil {
			return nil, i18n.WrapError(err)
		}

		for _, folder := range folders {
			otherSources, err := findAllFilesInFolder(filepath.Join(sourcePath, folder.Name()), recurse)
			if err != nil {
				return nil, i18n.WrapError(err)
			}
			sources = append(sources, otherSources...)
		}
	}

	return sources, nil
}

func compileFilesWithRecipe(ctx *types.Context, sourcePath *paths.Path, sources paths.PathList, buildPath *paths.Path, buildProperties properties.Map, includes []string, recipe string) (paths.PathList, error) {
	objectFiles := paths.NewPathList()
	if len(sources) == 0 {
		return objectFiles, nil
	}
	objectFilesChan := make(chan *paths.Path)
	errorsChan := make(chan error)
	doneChan := make(chan struct{})

	ctx.Progress.Steps = ctx.Progress.Steps / float64(len(sources))
	var wg sync.WaitGroup
	wg.Add(len(sources))

	for _, source := range sources {
		go func(source *paths.Path) {
			defer wg.Done()
			PrintProgressIfProgressEnabledAndMachineLogger(ctx)
			objectFile, err := compileFileWithRecipe(ctx, sourcePath, source, buildPath, buildProperties, includes, recipe)
			if err != nil {
				errorsChan <- err
			} else {
				objectFilesChan <- objectFile
			}
		}(source)
	}

	go func() {
		wg.Wait()
		doneChan <- struct{}{}
	}()

	for {
		select {
		case objectFile := <-objectFilesChan:
			objectFiles.Add(objectFile)
		case err := <-errorsChan:
			return nil, i18n.WrapError(err)
		case <-doneChan:
			close(objectFilesChan)
			for objectFile := range objectFilesChan {
				objectFiles.Add(objectFile)
			}
			objectFiles.Sort()
			return objectFiles, nil
		}
	}
}

func compileFileWithRecipe(ctx *types.Context, sourcePath *paths.Path, source *paths.Path, buildPath *paths.Path, buildProperties properties.Map, includes []string, recipe string) (*paths.Path, error) {
	logger := ctx.GetLogger()
	properties := buildProperties.Clone()
	properties[constants.BUILD_PROPERTIES_COMPILER_WARNING_FLAGS] = properties[constants.BUILD_PROPERTIES_COMPILER_WARNING_FLAGS+"."+ctx.WarningsLevel]
	properties[constants.BUILD_PROPERTIES_INCLUDES] = strings.Join(includes, constants.SPACE)
	properties[constants.BUILD_PROPERTIES_SOURCE_FILE] = source.String()
	relativeSource, err := sourcePath.RelTo(source)
	if err != nil {
		return nil, i18n.WrapError(err)
	}
	properties[constants.BUILD_PROPERTIES_OBJECT_FILE] = buildPath.JoinPath(relativeSource).String() + ".o"

	err = paths.New(properties[constants.BUILD_PROPERTIES_OBJECT_FILE]).Parent().MkdirAll()
	if err != nil {
		return nil, i18n.WrapError(err)
	}

	objIsUpToDate, err := ObjFileIsUpToDate(ctx, properties.GetPath(constants.BUILD_PROPERTIES_SOURCE_FILE), properties.GetPath(constants.BUILD_PROPERTIES_OBJECT_FILE), buildPath.Join(relativeSource.String()+".d"))
	if err != nil {
		return nil, i18n.WrapError(err)
	}

	if !objIsUpToDate {
		_, _, err = ExecRecipe(ctx, properties, recipe, false /* stdout */, utils.ShowIfVerbose /* stderr */, utils.Show)
		if err != nil {
			return nil, i18n.WrapError(err)
		}
	} else if ctx.Verbose {
		logger.Println(constants.LOG_LEVEL_INFO, constants.MSG_USING_PREVIOUS_COMPILED_FILE, properties[constants.BUILD_PROPERTIES_OBJECT_FILE])
	}

	return paths.New(properties[constants.BUILD_PROPERTIES_OBJECT_FILE]), nil
}

func ObjFileIsUpToDate(ctx *types.Context, sourceFile, objectFile, dependencyFile *paths.Path) (bool, error) {
	logger := ctx.GetLogger()
	debugLevel := ctx.DebugLevel
	if debugLevel >= 20 {
		logger.Fprintln(os.Stdout, constants.LOG_LEVEL_DEBUG, "Checking previous results for {0} (result = {1}, dep = {2})", sourceFile, objectFile, dependencyFile)
	}
	if objectFile == nil || dependencyFile == nil {
		if debugLevel >= 20 {
			logger.Fprintln(os.Stdout, constants.LOG_LEVEL_DEBUG, "Not found: nil")
		}
		return false, nil
	}

	sourceFile = sourceFile.Clean()
	sourceFileStat, err := sourceFile.Stat()
	if err != nil {
		return false, i18n.WrapError(err)
	}

	objectFile = objectFile.Clean()
	objectFileStat, err := objectFile.Stat()
	if err != nil {
		if os.IsNotExist(err) {
			if debugLevel >= 20 {
				logger.Fprintln(os.Stdout, constants.LOG_LEVEL_DEBUG, "Not found: {0}", objectFile)
			}
			return false, nil
		} else {
			return false, i18n.WrapError(err)
		}
	}

	dependencyFile = dependencyFile.Clean()
	dependencyFileStat, err := dependencyFile.Stat()
	if err != nil {
		if os.IsNotExist(err) {
			if debugLevel >= 20 {
				logger.Fprintln(os.Stdout, constants.LOG_LEVEL_DEBUG, "Not found: {0}", dependencyFile)
			}
			return false, nil
		} else {
			return false, i18n.WrapError(err)
		}
	}

	if sourceFileStat.ModTime().After(objectFileStat.ModTime()) {
		if debugLevel >= 20 {
			logger.Fprintln(os.Stdout, constants.LOG_LEVEL_DEBUG, "{0} newer than {1}", sourceFile, objectFile)
		}
		return false, nil
	}
	if sourceFileStat.ModTime().After(dependencyFileStat.ModTime()) {
		if debugLevel >= 20 {
			logger.Fprintln(os.Stdout, constants.LOG_LEVEL_DEBUG, "{0} newer than {1}", sourceFile, dependencyFile)
		}
		return false, nil
	}

	rows, err := dependencyFile.ReadFileAsLines()
	if err != nil {
		return false, i18n.WrapError(err)
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
		if debugLevel >= 20 {
			logger.Fprintln(os.Stdout, constants.LOG_LEVEL_DEBUG, "No colon in first line of depfile")
		}
		return false, nil
	}
	objFileInDepFile := firstRow[:len(firstRow)-1]
	if objFileInDepFile != objectFile.String() {
		if debugLevel >= 20 {
			logger.Fprintln(os.Stdout, constants.LOG_LEVEL_DEBUG, "Depfile is about different file: {0}", objFileInDepFile)
		}
		return false, nil
	}

	rows = rows[1:]
	for _, row := range rows {
		depStat, err := os.Stat(row)
		if err != nil && !os.IsNotExist(err) {
			// There is probably a parsing error of the dep file
			// Ignore the error and trigger a full rebuild anyway
			if debugLevel >= 20 {
				logger.Fprintln(os.Stdout, constants.LOG_LEVEL_DEBUG, "Failed to read: {0}", row)
				logger.Fprintln(os.Stdout, constants.LOG_LEVEL_DEBUG, i18n.WrapError(err).Error())
			}
			return false, nil
		}
		if os.IsNotExist(err) {
			if debugLevel >= 20 {
				logger.Fprintln(os.Stdout, constants.LOG_LEVEL_DEBUG, "Not found: {0}", row)
			}
			return false, nil
		}
		if depStat.ModTime().After(objectFileStat.ModTime()) {
			if debugLevel >= 20 {
				logger.Fprintln(os.Stdout, constants.LOG_LEVEL_DEBUG, "{0} newer than {1}", row, objectFile)
			}
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

func ArchiveCompiledFiles(ctx *types.Context, buildPath *paths.Path, archiveFile *paths.Path, objectFilesToArchive paths.PathList, buildProperties properties.Map) (*paths.Path, error) {
	logger := ctx.GetLogger()
	archiveFilePath := buildPath.JoinPath(archiveFile)

	rebuildArchive := false

	if archiveFileStat, err := archiveFilePath.Stat(); err == nil {

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
			err = archiveFilePath.Remove()
			if err != nil {
				return nil, i18n.WrapError(err)
			}
		} else {
			if ctx.Verbose {
				logger.Println(constants.LOG_LEVEL_INFO, constants.MSG_USING_PREVIOUS_COMPILED_FILE, archiveFilePath)
			}
			return archiveFilePath, nil
		}
	}

	for _, objectFile := range objectFilesToArchive {
		properties := buildProperties.Clone()
		properties[constants.BUILD_PROPERTIES_ARCHIVE_FILE] = archiveFilePath.Base()
		properties[constants.BUILD_PROPERTIES_ARCHIVE_FILE_PATH] = archiveFilePath.String()
		properties[constants.BUILD_PROPERTIES_OBJECT_FILE] = objectFile.String()

		_, _, err := ExecRecipe(ctx, properties, constants.RECIPE_AR_PATTERN, false /* stdout */, utils.ShowIfVerbose /* stderr */, utils.Show)
		if err != nil {
			return nil, i18n.WrapError(err)
		}
	}

	return archiveFilePath, nil
}

func ExecRecipe(ctx *types.Context, buildProperties properties.Map, recipe string, removeUnsetProperties bool, stdout int, stderr int) ([]byte, []byte, error) {
	// See util.ExecCommand for stdout/stderr arguments
	command, err := PrepareCommandForRecipe(ctx, buildProperties, recipe, removeUnsetProperties)
	if err != nil {
		return nil, nil, i18n.WrapError(err)
	}

	return utils.ExecCommand(ctx, command, stdout, stderr)
}

const COMMANDLINE_LIMIT = 30000

func PrepareCommandForRecipe(ctx *types.Context, buildProperties properties.Map, recipe string, removeUnsetProperties bool) (*exec.Cmd, error) {
	logger := ctx.GetLogger()
	pattern := buildProperties[recipe]
	if pattern == "" {
		return nil, i18n.ErrorfWithLogger(logger, constants.MSG_PATTERN_MISSING, recipe)
	}

	var err error
	commandLine := buildProperties.ExpandPropsInString(pattern)
	if removeUnsetProperties {
		commandLine = properties.DeleteUnexpandedPropsFromString(commandLine)
	}

	relativePath := ""

	if len(commandLine) > COMMANDLINE_LIMIT {
		relativePath = buildProperties["build.path"]
	}

	command, err := utils.PrepareCommand(commandLine, logger, relativePath)
	if err != nil {
		return nil, i18n.WrapError(err)
	}

	return command, nil
}

// GetCachedCoreArchiveFileName returns the filename to be used to store
// the global cached core.a.
func GetCachedCoreArchiveFileName(fqbn string, coreFolder *paths.Path) string {
	fqbnToUnderscore := strings.Replace(fqbn, ":", "_", -1)
	fqbnToUnderscore = strings.Replace(fqbnToUnderscore, "=", "_", -1)
	if absCoreFolder, err := coreFolder.Abs(); err == nil {
		coreFolder = absCoreFolder
	} // silently continue if absolute path can't be detected
	hash := utils.MD5Sum([]byte(coreFolder.String()))
	realName := "core_" + fqbnToUnderscore + "_" + hash + ".a"
	if len(realName) > 100 {
		// avoid really long names, simply hash the final part
		realName = "core_" + utils.MD5Sum([]byte(fqbnToUnderscore+"_"+hash)) + ".a"
	}
	return realName
}
