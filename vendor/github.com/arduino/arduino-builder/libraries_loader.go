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
	"os"
	"path/filepath"
	"strings"

	"github.com/arduino/arduino-builder/constants"
	"github.com/arduino/arduino-builder/i18n"
	"github.com/arduino/arduino-builder/types"
	"github.com/arduino/arduino-builder/utils"
	"github.com/arduino/go-properties-map"
)

type LibrariesLoader struct{}

func (s *LibrariesLoader) Run(ctx *types.Context) error {
	builtInLibrariesFolders := ctx.BuiltInLibrariesFolders
	builtInLibrariesFolders, err := utils.AbsolutizePaths(builtInLibrariesFolders)
	if err != nil {
		return i18n.WrapError(err)
	}
	sortedLibrariesFolders := []string{}
	sortedLibrariesFolders = utils.AppendIfNotPresent(sortedLibrariesFolders, builtInLibrariesFolders...)

	platform := ctx.TargetPlatform
	debugLevel := ctx.DebugLevel
	logger := ctx.GetLogger()

	actualPlatform := ctx.ActualPlatform
	if actualPlatform != platform {
		sortedLibrariesFolders = appendPathToLibrariesFolders(sortedLibrariesFolders, filepath.Join(actualPlatform.Folder, constants.FOLDER_LIBRARIES))
	}

	sortedLibrariesFolders = appendPathToLibrariesFolders(sortedLibrariesFolders, filepath.Join(platform.Folder, constants.FOLDER_LIBRARIES))

	librariesFolders := ctx.OtherLibrariesFolders
	librariesFolders, err = utils.AbsolutizePaths(librariesFolders)
	if err != nil {
		return i18n.WrapError(err)
	}
	sortedLibrariesFolders = utils.AppendIfNotPresent(sortedLibrariesFolders, librariesFolders...)

	ctx.LibrariesFolders = sortedLibrariesFolders

	var libraries []*types.Library
	for _, libraryFolder := range sortedLibrariesFolders {
		subFolders, err := utils.ReadDirFiltered(libraryFolder, utils.FilterDirs)
		if err != nil {
			return i18n.WrapError(err)
		}
		for _, subFolder := range subFolders {
			library, err := makeLibrary(filepath.Join(libraryFolder, subFolder.Name()), debugLevel, logger)
			if err != nil {
				return i18n.WrapError(err)
			}
			libraries = append(libraries, library)
		}
	}

	ctx.Libraries = libraries

	headerToLibraries := make(map[string][]*types.Library)
	for _, library := range libraries {
		headers, err := utils.ReadDirFiltered(library.SrcFolder, utils.FilterFilesWithExtensions(".h", ".hpp", ".hh"))
		if err != nil {
			return i18n.WrapError(err)
		}
		for _, header := range headers {
			headerFileName := header.Name()
			headerToLibraries[headerFileName] = append(headerToLibraries[headerFileName], library)
		}
	}

	ctx.HeaderToLibraries = headerToLibraries

	return nil
}

func makeLibrary(libraryFolder string, debugLevel int, logger i18n.Logger) (*types.Library, error) {
	if _, err := os.Stat(filepath.Join(libraryFolder, constants.LIBRARY_PROPERTIES)); os.IsNotExist(err) {
		return makeLegacyLibrary(libraryFolder)
	}
	return makeNewLibrary(libraryFolder, debugLevel, logger)
}

func addUtilityFolder(library *types.Library) {
	utilitySourcePath := filepath.Join(library.Folder, constants.LIBRARY_FOLDER_UTILITY)
	stat, err := os.Stat(utilitySourcePath)
	if err == nil && stat.IsDir() {
		library.UtilityFolder = utilitySourcePath
	}
}

func makeNewLibrary(libraryFolder string, debugLevel int, logger i18n.Logger) (*types.Library, error) {
	libProperties, err := properties.Load(filepath.Join(libraryFolder, constants.LIBRARY_PROPERTIES))
	if err != nil {
		return nil, i18n.WrapError(err)
	}

	if libProperties[constants.LIBRARY_MAINTAINER] == constants.EMPTY_STRING && libProperties[constants.LIBRARY_EMAIL] != constants.EMPTY_STRING {
		libProperties[constants.LIBRARY_MAINTAINER] = libProperties[constants.LIBRARY_EMAIL]
	}

	for _, propName := range LIBRARY_NOT_SO_MANDATORY_PROPERTIES {
		if libProperties[propName] == constants.EMPTY_STRING {
			libProperties[propName] = "-"
		}
	}

	library := &types.Library{}
	library.Folder = libraryFolder
	if stat, err := os.Stat(filepath.Join(libraryFolder, constants.LIBRARY_FOLDER_SRC)); err == nil && stat.IsDir() {
		library.Layout = types.LIBRARY_RECURSIVE
		library.SrcFolder = filepath.Join(libraryFolder, constants.LIBRARY_FOLDER_SRC)
	} else {
		library.Layout = types.LIBRARY_FLAT
		library.SrcFolder = libraryFolder
		addUtilityFolder(library)
	}

	subFolders, err := utils.ReadDirFiltered(libraryFolder, utils.FilterDirs)
	if err != nil {
		return nil, i18n.WrapError(err)
	}

	if debugLevel >= 0 {
		for _, subFolder := range subFolders {
			if utils.IsSCCSOrHiddenFile(subFolder) {
				if !utils.IsSCCSFile(subFolder) && utils.IsHiddenFile(subFolder) {
					logger.Fprintln(os.Stdout, constants.LOG_LEVEL_WARN, constants.MSG_WARNING_SPURIOUS_FILE_IN_LIB, filepath.Base(subFolder.Name()), libProperties[constants.LIBRARY_NAME])
				}
			}
		}
	}

	if libProperties[constants.LIBRARY_ARCHITECTURES] == constants.EMPTY_STRING {
		libProperties[constants.LIBRARY_ARCHITECTURES] = constants.LIBRARY_ALL_ARCHS
	}
	library.Archs = []string{}
	for _, arch := range strings.Split(libProperties[constants.LIBRARY_ARCHITECTURES], ",") {
		library.Archs = append(library.Archs, strings.TrimSpace(arch))
	}

	libProperties[constants.LIBRARY_CATEGORY] = strings.TrimSpace(libProperties[constants.LIBRARY_CATEGORY])
	if !LIBRARY_CATEGORIES[libProperties[constants.LIBRARY_CATEGORY]] {
		logger.Fprintln(os.Stdout, constants.LOG_LEVEL_WARN, constants.MSG_WARNING_LIB_INVALID_CATEGORY, libProperties[constants.LIBRARY_CATEGORY], libProperties[constants.LIBRARY_NAME], constants.LIB_CATEGORY_UNCATEGORIZED)
		libProperties[constants.LIBRARY_CATEGORY] = constants.LIB_CATEGORY_UNCATEGORIZED
	}
	library.Category = libProperties[constants.LIBRARY_CATEGORY]

	if libProperties[constants.LIBRARY_LICENSE] == constants.EMPTY_STRING {
		libProperties[constants.LIBRARY_LICENSE] = constants.LIB_LICENSE_UNSPECIFIED
	}
	library.License = libProperties[constants.LIBRARY_LICENSE]

	library.Name = filepath.Base(libraryFolder)
	library.RealName = strings.TrimSpace(libProperties[constants.LIBRARY_NAME])
	library.Version = strings.TrimSpace(libProperties[constants.LIBRARY_VERSION])
	library.Author = strings.TrimSpace(libProperties[constants.LIBRARY_AUTHOR])
	library.Maintainer = strings.TrimSpace(libProperties[constants.LIBRARY_MAINTAINER])
	library.Sentence = strings.TrimSpace(libProperties[constants.LIBRARY_SENTENCE])
	library.Paragraph = strings.TrimSpace(libProperties[constants.LIBRARY_PARAGRAPH])
	library.URL = strings.TrimSpace(libProperties[constants.LIBRARY_URL])
	library.IsLegacy = false
	library.DotALinkage = strings.TrimSpace(libProperties[constants.LIBRARY_DOT_A_LINKAGE]) == "true"
	library.Precompiled = strings.TrimSpace(libProperties[constants.LIBRARY_PRECOMPILED]) == "true"
	library.LDflags = strings.TrimSpace(libProperties[constants.LIBRARY_LDFLAGS])
	library.Properties = libProperties

	return library, nil
}

func makeLegacyLibrary(libraryFolder string) (*types.Library, error) {
	library := &types.Library{
		Folder:    libraryFolder,
		SrcFolder: libraryFolder,
		Layout:    types.LIBRARY_FLAT,
		Name:      filepath.Base(libraryFolder),
		Archs:     []string{constants.LIBRARY_ALL_ARCHS},
		IsLegacy:  true,
	}
	addUtilityFolder(library)
	return library, nil
}

func appendPathToLibrariesFolders(librariesFolders []string, newLibrariesFolder string) []string {
	if stat, err := os.Stat(newLibrariesFolder); os.IsNotExist(err) || !stat.IsDir() {
		return librariesFolders
	}

	return utils.AppendIfNotPresent(librariesFolders, newLibrariesFolder)
}
