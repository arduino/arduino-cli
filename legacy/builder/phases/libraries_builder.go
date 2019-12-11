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

package phases

import (
	"os"
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
var FLOAT_ABI_CFLAG = "float-abi"
var FPU_CFLAG = "fpu"

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

func findExpectedPrecompiledLibFolder(ctx *types.Context, library *libraries.Library) *paths.Path {
	mcu := ctx.BuildProperties.Get(constants.BUILD_PROPERTIES_BUILD_MCU)
	// Add fpu specifications if they exist
	// To do so, resolve recipe.cpp.o.pattern,
	// search for -mfpu=xxx -mfloat-abi=yyy and add to a subfolder
	command, _ := builder_utils.PrepareCommandForRecipe(ctx, ctx.BuildProperties, constants.RECIPE_CPP_PATTERN, true)
	fpuSpecs := ""
	for _, el := range strings.Split(command.String(), " ") {
		if strings.Contains(el, FPU_CFLAG) {
			toAdd := strings.Split(el, "=")
			if len(toAdd) > 1 {
				fpuSpecs += strings.TrimSpace(toAdd[1]) + "-"
				break
			}
		}
	}
	for _, el := range strings.Split(command.String(), " ") {
		if strings.Contains(el, FLOAT_ABI_CFLAG) {
			toAdd := strings.Split(el, "=")
			if len(toAdd) > 1 {
				fpuSpecs += strings.TrimSpace(toAdd[1]) + "-"
				break
			}
		}
	}

	logger := ctx.GetLogger()
	if len(fpuSpecs) > 0 {
		fpuSpecs = strings.TrimRight(fpuSpecs, "-")
		if library.SourceDir.Join(mcu).Join(fpuSpecs).Exist() {
			return library.SourceDir.Join(mcu).Join(fpuSpecs)
		} else {
			// we are unsure, compile from sources
			logger.Fprintln(os.Stdout, constants.LOG_LEVEL_INFO,
				constants.MSG_PRECOMPILED_LIBRARY_NOT_FOUND_FOR, library.Name, library.SourceDir.Join(mcu).Join(fpuSpecs))
			return nil
		}
	}

	if library.SourceDir.Join(mcu).Exist() {
		return library.SourceDir.Join(mcu)
	}

	logger.Fprintln(os.Stdout, constants.LOG_LEVEL_INFO,
		constants.MSG_PRECOMPILED_LIBRARY_NOT_FOUND_FOR, library.Name, library.SourceDir.Join(mcu))

	return nil
}

func fixLDFLAGforPrecompiledLibraries(ctx *types.Context, libs libraries.List) error {

	for _, library := range libs {
		if library.Precompiled {
			// add library src path to compiler.c.elf.extra_flags
			// use library.Name as lib name and srcPath/{mcpu} as location
			path := findExpectedPrecompiledLibFolder(ctx, library)
			if path == nil {
				break
			}
			// find all library names in the folder and prepend -l
			filePaths := []string{}
			libs_cmd := library.LDflags + " "
			extensions := func(ext string) bool {
				return PRECOMPILED_LIBRARIES_VALID_EXTENSIONS_DYNAMIC[ext] || PRECOMPILED_LIBRARIES_VALID_EXTENSIONS_STATIC[ext]
			}
			utils.FindFilesInFolder(&filePaths, path.String(), extensions, false)
			for _, lib := range filePaths {
				name := strings.TrimSuffix(filepath.Base(lib), filepath.Ext(lib))
				// strip "lib" first occurrence
				if strings.HasPrefix(name, "lib") {
					name = strings.Replace(name, "lib", "", 1)
					libs_cmd += "-l" + name + " "
				}
			}

			currLDFlags := ctx.BuildProperties.Get(constants.BUILD_PROPERTIES_COMPILER_LIBRARIES_LDFLAGS)
			ctx.BuildProperties.Set(constants.BUILD_PROPERTIES_COMPILER_LIBRARIES_LDFLAGS, currLDFlags+"\"-L"+path.String()+"\" "+libs_cmd+" ")
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
		precompiledPath := findExpectedPrecompiledLibFolder(ctx, library)
		if precompiledPath != nil {
			// TODO: This codepath is just taken for .a with unusual names that would
			// be ignored by -L / -l methods.
			// Should we force precompiled libraries to start with "lib" ?
			err := utils.FindFilesInFolder(&filePaths, precompiledPath.String(), extensions, false)
			if err != nil {
				return nil, i18n.WrapError(err)
			}
			for _, path := range filePaths {
				if !strings.HasPrefix(filepath.Base(path), "lib") {
					objectFiles.Add(paths.New(path))
				}
			}
			return objectFiles, nil
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
