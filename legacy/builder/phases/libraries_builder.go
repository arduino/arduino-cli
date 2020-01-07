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

package phases

import (
	"path/filepath"
	"strings"

	"github.com/arduino/arduino-cli/arduino/libraries"
	"github.com/arduino/arduino-cli/legacy/builder/builder_utils"
	"github.com/arduino/arduino-cli/legacy/builder/constants"
	"github.com/arduino/arduino-cli/legacy/builder/i18n"
	"github.com/arduino/arduino-cli/legacy/builder/types"
	"github.com/arduino/arduino-cli/legacy/builder/utils"
	"github.com/arduino/go-paths-helper"
	"github.com/arduino/go-properties-orderedmap"
)

var PRECOMPILED_LIBRARIES_VALID_EXTENSIONS_STATIC = map[string]bool{".a": true}
var PRECOMPILED_LIBRARIES_VALID_EXTENSIONS_DYNAMIC = map[string]bool{".so": true}

type LibrariesBuilder struct{}

func (s *LibrariesBuilder) Run(ctx *types.Context) error {
	librariesBuildPath := ctx.LibrariesBuildPath
	buildProperties := ctx.BuildProperties
	includes := utils.Map(ctx.IncludeFolders.AsStrings(), utils.WrapWithHyphenI)
	libs := ctx.ImportedLibraries

	if err := librariesBuildPath.MkdirAll(); err != nil {
		return i18n.WrapError(err)
	}

	objectFiles, err := compileLibraries(ctx, libs, librariesBuildPath, buildProperties, includes)
	if err != nil {
		return i18n.WrapError(err)
	}

	ctx.LibrariesObjectFiles = objectFiles

	// Search for precompiled libraries
	fixLDFLAGforPrecompiledLibraries(ctx, libs)

	return nil
}

func fixLDFLAGforPrecompiledLibraries(ctx *types.Context, libs libraries.List) error {

	for _, library := range libs {
		if library.Precompiled {
			// add library src path to compiler.c.elf.extra_flags
			// use library.Name as lib name and srcPath/{mcpu} as location
			mcu := ctx.BuildProperties.Get(constants.BUILD_PROPERTIES_BUILD_MCU)
			path := library.SourceDir.Join(mcu).String()
			// find all library names in the folder and prepend -l
			filePaths := []string{}
			libs_cmd := library.LDflags + " "
			extensions := func(ext string) bool { return PRECOMPILED_LIBRARIES_VALID_EXTENSIONS_DYNAMIC[ext] }
			utils.FindFilesInFolder(&filePaths, path, extensions, true)
			for _, lib := range filePaths {
				name := strings.TrimSuffix(filepath.Base(lib), filepath.Ext(lib))
				// strip "lib" first occurrence
				name = strings.Replace(name, "lib", "", 1)
				libs_cmd += "-l" + name + " "
			}

			currLDFlags := ctx.BuildProperties.Get(constants.BUILD_PROPERTIES_COMPILER_LIBRARIES_LDFLAGS)
			ctx.BuildProperties.Set(constants.BUILD_PROPERTIES_COMPILER_LIBRARIES_LDFLAGS, currLDFlags+"\"-L"+path+"\" "+libs_cmd+" ")
		}
	}
	return nil
}

func compileLibraries(ctx *types.Context, libraries libraries.List, buildPath *paths.Path, buildProperties *properties.Map, includes []string) (paths.PathList, error) {
	objectFiles := paths.NewPathList()
	for _, library := range libraries {
		libraryObjectFiles, err := compileLibrary(ctx, library, buildPath, buildProperties, includes)
		if err != nil {
			return nil, i18n.WrapError(err)
		}
		objectFiles = append(objectFiles, libraryObjectFiles...)
	}

	return objectFiles, nil
}

func compileLibrary(ctx *types.Context, library *libraries.Library, buildPath *paths.Path, buildProperties *properties.Map, includes []string) (paths.PathList, error) {
	logger := ctx.GetLogger()
	if ctx.Verbose {
		logger.Println(constants.LOG_LEVEL_INFO, "Compiling library \"{0}\"", library.Name)
	}
	libraryBuildPath := buildPath.Join(library.Name)

	if err := libraryBuildPath.MkdirAll(); err != nil {
		return nil, i18n.WrapError(err)
	}

	objectFiles := paths.NewPathList()

	if library.Precompiled {
		// search for files with PRECOMPILED_LIBRARIES_VALID_EXTENSIONS
		extensions := func(ext string) bool { return PRECOMPILED_LIBRARIES_VALID_EXTENSIONS_STATIC[ext] }

		filePaths := []string{}
		mcu := buildProperties.Get(constants.BUILD_PROPERTIES_BUILD_MCU)
		err := utils.FindFilesInFolder(&filePaths, library.SourceDir.Join(mcu).String(), extensions, true)
		if err != nil {
			return nil, i18n.WrapError(err)
		}
		for _, path := range filePaths {
			if strings.Contains(filepath.Base(path), library.RealName) {
				objectFiles.Add(paths.New(path))
			}
		}
	}

	if library.Layout == libraries.RecursiveLayout {
		libObjectFiles, err := builder_utils.CompileFilesRecursive(ctx, library.SourceDir, libraryBuildPath, buildProperties, includes)
		if err != nil {
			return nil, i18n.WrapError(err)
		}
		if library.DotALinkage {
			archiveFile, err := builder_utils.ArchiveCompiledFiles(ctx, libraryBuildPath, paths.New(library.Name+".a"), libObjectFiles, buildProperties)
			if err != nil {
				return nil, i18n.WrapError(err)
			}
			objectFiles.Add(archiveFile)
		} else {
			objectFiles.AddAll(libObjectFiles)
		}
	} else {
		if library.UtilityDir != nil {
			includes = append(includes, utils.WrapWithHyphenI(library.UtilityDir.String()))
		}
		libObjectFiles, err := builder_utils.CompileFiles(ctx, library.SourceDir, false, libraryBuildPath, buildProperties, includes)
		if err != nil {
			return nil, i18n.WrapError(err)
		}
		objectFiles.AddAll(libObjectFiles)

		if library.UtilityDir != nil {
			utilityBuildPath := libraryBuildPath.Join("utility")
			utilityObjectFiles, err := builder_utils.CompileFiles(ctx, library.UtilityDir, false, utilityBuildPath, buildProperties, includes)
			if err != nil {
				return nil, i18n.WrapError(err)
			}
			objectFiles.AddAll(utilityObjectFiles)
		}
	}

	return objectFiles, nil
}
