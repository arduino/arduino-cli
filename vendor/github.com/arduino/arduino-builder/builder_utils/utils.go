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
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/arduino/arduino-builder/constants"
	"github.com/arduino/arduino-builder/i18n"
	"github.com/arduino/arduino-builder/utils"
	"github.com/arduino/go-properties-map"
)

func CompileFilesRecursive(objectFiles []string, sourcePath string, buildPath string, buildProperties properties.Map, includes []string, verbose bool, warningsLevel string, logger i18n.Logger) ([]string, error) {
	objectFiles, err := CompileFiles(objectFiles, sourcePath, false, buildPath, buildProperties, includes, verbose, warningsLevel, logger)
	if err != nil {
		return nil, i18n.WrapError(err)
	}

	folders, err := utils.ReadDirFiltered(sourcePath, utils.FilterDirs)
	if err != nil {
		return nil, i18n.WrapError(err)
	}

	for _, folder := range folders {
		objectFiles, err = CompileFilesRecursive(objectFiles, filepath.Join(sourcePath, folder.Name()), filepath.Join(buildPath, folder.Name()), buildProperties, includes, verbose, warningsLevel, logger)
		if err != nil {
			return nil, i18n.WrapError(err)
		}
	}

	return objectFiles, nil
}

func CompileFiles(objectFiles []string, sourcePath string, recurse bool, buildPath string, buildProperties properties.Map, includes []string, verbose bool, warningsLevel string, logger i18n.Logger) ([]string, error) {
	objectFiles, err := compileFilesWithExtensionWithRecipe(objectFiles, sourcePath, recurse, buildPath, buildProperties, includes, ".S", constants.RECIPE_S_PATTERN, verbose, warningsLevel, logger)
	if err != nil {
		return nil, i18n.WrapError(err)
	}
	objectFiles, err = compileFilesWithExtensionWithRecipe(objectFiles, sourcePath, recurse, buildPath, buildProperties, includes, ".c", constants.RECIPE_C_PATTERN, verbose, warningsLevel, logger)
	if err != nil {
		return nil, i18n.WrapError(err)
	}
	objectFiles, err = compileFilesWithExtensionWithRecipe(objectFiles, sourcePath, recurse, buildPath, buildProperties, includes, ".cpp", constants.RECIPE_CPP_PATTERN, verbose, warningsLevel, logger)
	if err != nil {
		return nil, i18n.WrapError(err)
	}
	return objectFiles, nil
}

func compileFilesWithExtensionWithRecipe(objectFiles []string, sourcePath string, recurse bool, buildPath string, buildProperties properties.Map, includes []string, extension string, recipe string, verbose bool, warningsLevel string, logger i18n.Logger) ([]string, error) {
	sources, err := findFilesInFolder(sourcePath, extension, recurse)
	if err != nil {
		return nil, i18n.WrapError(err)
	}
	return compileFilesWithRecipe(objectFiles, sourcePath, sources, buildPath, buildProperties, includes, recipe, verbose, warningsLevel, logger)
}

func findFilesInFolder(sourcePath string, extension string, recurse bool) ([]string, error) {
	files, err := utils.ReadDirFiltered(sourcePath, utils.FilterFilesWithExtensions(extension))
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
			otherSources, err := findFilesInFolder(filepath.Join(sourcePath, folder.Name()), extension, recurse)
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

func compileFilesWithRecipe(objectFiles []string, sourcePath string, sources []string, buildPath string, buildProperties properties.Map, includes []string, recipe string, verbose bool, warningsLevel string, logger i18n.Logger) ([]string, error) {
	for _, source := range sources {
		objectFile, err := compileFileWithRecipe(sourcePath, source, buildPath, buildProperties, includes, recipe, verbose, warningsLevel, logger)
		if err != nil {
			return nil, i18n.WrapError(err)
		}

		objectFiles = append(objectFiles, objectFile)
	}
	return objectFiles, nil
}

func compileFileWithRecipe(sourcePath string, source string, buildPath string, buildProperties properties.Map, includes []string, recipe string, verbose bool, warningsLevel string, logger i18n.Logger) (string, error) {
	properties := buildProperties.Clone()
	properties[constants.BUILD_PROPERTIES_COMPILER_WARNING_FLAGS] = properties[constants.BUILD_PROPERTIES_COMPILER_WARNING_FLAGS+"."+warningsLevel]
	properties[constants.BUILD_PROPERTIES_INCLUDES] = strings.Join(includes, constants.SPACE)
	properties[constants.BUILD_PROPERTIES_SOURCE_FILE] = source
	relativeSource, err := filepath.Rel(sourcePath, source)
	if err != nil {
		return "", i18n.WrapError(err)
	}
	properties[constants.BUILD_PROPERTIES_OBJECT_FILE] = filepath.Join(buildPath, relativeSource+".o")

	err = utils.EnsureFolderExists(filepath.Dir(properties[constants.BUILD_PROPERTIES_OBJECT_FILE]))
	if err != nil {
		return "", i18n.WrapError(err)
	}

	objIsUpToDate, err := ObjFileIsUpToDate(properties[constants.BUILD_PROPERTIES_SOURCE_FILE], properties[constants.BUILD_PROPERTIES_OBJECT_FILE], filepath.Join(buildPath, relativeSource+".d"))
	if err != nil {
		return "", i18n.WrapError(err)
	}

	if !objIsUpToDate {
		_, err = ExecRecipe(properties, recipe, false, verbose, verbose, logger)
		if err != nil {
			return "", i18n.WrapError(err)
		}
	} else if verbose {
		logger.Println(constants.LOG_LEVEL_INFO, constants.MSG_USING_PREVIOUS_COMPILED_FILE, properties[constants.BUILD_PROPERTIES_OBJECT_FILE])
	}

	return properties[constants.BUILD_PROPERTIES_OBJECT_FILE], nil
}

func ObjFileIsUpToDate(sourceFile, objectFile, dependencyFile string) (bool, error) {
	sourceFile = filepath.Clean(sourceFile)
	objectFile = filepath.Clean(objectFile)
	dependencyFile = filepath.Clean(dependencyFile)

	sourceFileStat, err := os.Stat(sourceFile)
	if err != nil {
		return false, i18n.WrapError(err)
	}

	objectFileStat, err := os.Stat(objectFile)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		} else {
			return false, i18n.WrapError(err)
		}
	}

	dependencyFileStat, err := os.Stat(dependencyFile)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		} else {
			return false, i18n.WrapError(err)
		}
	}

	if sourceFileStat.ModTime().After(objectFileStat.ModTime()) {
		return false, nil
	}
	if sourceFileStat.ModTime().After(dependencyFileStat.ModTime()) {
		return false, nil
	}

	rows, err := utils.ReadFileToRows(dependencyFile)
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
		return false, nil
	}
	objFileInDepFile := firstRow[:len(firstRow)-1]
	if objFileInDepFile != objectFile {
		return false, nil
	}

	rows = rows[1:]
	for _, row := range rows {
		depStat, err := os.Stat(row)
		if err != nil && !os.IsNotExist(err) {
			// There is probably a parsing error of the dep file
			// Ignore the error and trigger a full rebuild anyway
			return false, nil
		}
		if os.IsNotExist(err) {
			return false, nil
		}
		if depStat.ModTime().After(objectFileStat.ModTime()) {
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

func CoreOrReferencedCoreHasChanged(corePath, targetCorePath, targetFile string) bool {

	targetFileStat, err := os.Stat(targetFile)
	if err == nil {
		files, err := findAllFilesInFolder(corePath, true)
		if err != nil {
			return true
		}
		for _, file := range files {
			fileStat, err := os.Stat(file)
			if err != nil || fileStat.ModTime().After(targetFileStat.ModTime()) {
				return true
			}
		}
		if targetCorePath != constants.EMPTY_STRING && !strings.EqualFold(corePath, targetCorePath) {
			return CoreOrReferencedCoreHasChanged(targetCorePath, constants.EMPTY_STRING, targetFile)
		}
		return false
	}
	return true
}

func ArchiveCompiledFiles(buildPath string, archiveFile string, objectFiles []string, buildProperties properties.Map, verbose bool, logger i18n.Logger) (string, error) {
	archiveFilePath := filepath.Join(buildPath, archiveFile)

	rebuildArchive := false

	if archiveFileStat, err := os.Stat(archiveFilePath); err == nil {

		for _, objectFile := range objectFiles {
			objectFileStat, _ := os.Stat(objectFile)
			if objectFileStat.ModTime().After(archiveFileStat.ModTime()) {
				// need to rebuild the archive
				rebuildArchive = true
				break
			}
		}

		// something changed, rebuild the core archive
		if rebuildArchive {
			err = os.Remove(archiveFilePath)
			if err != nil {
				return "", i18n.WrapError(err)
			}
		} else {
			if verbose {
				logger.Println(constants.LOG_LEVEL_INFO, constants.MSG_USING_PREVIOUS_COMPILED_FILE, archiveFilePath)
			}
			return archiveFilePath, nil
		}
	}

	for _, objectFile := range objectFiles {
		properties := buildProperties.Clone()
		properties[constants.BUILD_PROPERTIES_ARCHIVE_FILE] = filepath.Base(archiveFilePath)
		properties[constants.BUILD_PROPERTIES_ARCHIVE_FILE_PATH] = archiveFilePath
		properties[constants.BUILD_PROPERTIES_OBJECT_FILE] = objectFile

		_, err := ExecRecipe(properties, constants.RECIPE_AR_PATTERN, false, verbose, verbose, logger)
		if err != nil {
			return "", i18n.WrapError(err)
		}
	}

	return archiveFilePath, nil
}

func ExecRecipe(properties properties.Map, recipe string, removeUnsetProperties bool, echoCommandLine bool, echoOutput bool, logger i18n.Logger) ([]byte, error) {
	command, err := PrepareCommandForRecipe(properties, recipe, removeUnsetProperties, echoCommandLine, logger)
	if err != nil {
		return nil, i18n.WrapError(err)
	}

	if echoOutput {
		command.Stdout = os.Stdout
	}

	command.Stderr = os.Stderr

	if echoOutput {
		err := command.Run()
		return nil, i18n.WrapError(err)
	}

	bytes, err := command.Output()
	return bytes, i18n.WrapError(err)
}

func PrepareCommandForRecipe(buildProperties properties.Map, recipe string, removeUnsetProperties bool, echoCommandLine bool, logger i18n.Logger) (*exec.Cmd, error) {
	pattern := buildProperties[recipe]
	if pattern == constants.EMPTY_STRING {
		return nil, i18n.ErrorfWithLogger(logger, constants.MSG_PATTERN_MISSING, recipe)
	}

	var err error
	commandLine := buildProperties.ExpandPropsInString(pattern)
	if removeUnsetProperties {
		commandLine = properties.DeleteUnexpandedPropsFromString(commandLine)
	}

	command, err := utils.PrepareCommand(commandLine, logger)
	if err != nil {
		return nil, i18n.WrapError(err)
	}

	if echoCommandLine {
		fmt.Println(commandLine)
	}

	return command, nil
}

func ExecRecipeCollectStdErr(buildProperties properties.Map, recipe string, removeUnsetProperties bool, echoCommandLine bool, logger i18n.Logger) (string, error) {
	command, err := PrepareCommandForRecipe(buildProperties, recipe, removeUnsetProperties, echoCommandLine, logger)
	if err != nil {
		return "", i18n.WrapError(err)
	}

	buffer := &bytes.Buffer{}
	command.Stderr = buffer
	command.Run()
	return string(buffer.Bytes()), nil
}

func RemoveHyphenMDDFlagFromGCCCommandLine(buildProperties properties.Map) {
	buildProperties[constants.BUILD_PROPERTIES_COMPILER_CPP_FLAGS] = strings.Replace(buildProperties[constants.BUILD_PROPERTIES_COMPILER_CPP_FLAGS], "-MMD", "", -1)
}

// CopyFile copies the contents of the file named src to the file named
// by dst. The file will be created if it does not already exist. If the
// destination file exists, all it's contents will be replaced by the contents
// of the source file. The file mode will be copied from the source and
// the copied data is synced/flushed to stable storage.
func CopyFile(src, dst string) (err error) {
	in, err := os.Open(src)
	if err != nil {
		return
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return
	}
	defer func() {
		if e := out.Close(); e != nil {
			err = e
		}
	}()

	_, err = io.Copy(out, in)
	if err != nil {
		return
	}

	err = out.Sync()
	if err != nil {
		return
	}

	si, err := os.Stat(src)
	if err != nil {
		return
	}
	err = os.Chmod(dst, si.Mode())
	if err != nil {
		return
	}

	return
}

// GetCachedCoreArchiveFileName returns the filename to be used to store
// the global cached core.a.
func GetCachedCoreArchiveFileName(fqbn, coreFolder string) string {
	fqbnToUnderscore := strings.Replace(fqbn, ":", "_", -1)
	fqbnToUnderscore = strings.Replace(fqbnToUnderscore, "=", "_", -1)
	if absCoreFolder, err := filepath.Abs(coreFolder); err == nil {
		coreFolder = absCoreFolder
	} // silently continue if absolute path can't be detected
	hash := utils.MD5Sum([]byte(coreFolder))
	return "core_" + fqbnToUnderscore + "_" + hash + ".a"
}
