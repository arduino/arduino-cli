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
	"os"
	"path/filepath"
	"strings"

	"github.com/arduino/arduino-cli/arduino/libraries"
	"github.com/arduino/arduino-cli/legacy/builder/builder_utils"
	"github.com/arduino/arduino-cli/legacy/builder/constants"
	"github.com/arduino/arduino-cli/legacy/builder/types"
	"github.com/arduino/arduino-cli/legacy/builder/utils"
	"github.com/arduino/go-paths-helper"
	"github.com/arduino/go-properties-orderedmap"
	"github.com/pkg/errors"
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
		return errors.WithStack(err)
	}

	objectFiles, err := compileLibraries(ctx, libs, librariesBuildPath, buildProperties, includes)
	if err != nil {
		return errors.WithStack(err)
	}

	ctx.LibrariesObjectFiles = objectFiles
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
	logger.Fprintln(os.Stdout, constants.LOG_LEVEL_INFO, "Library {0} has been declared precompiled:", library.Name)

	// Try directory with full fpuSpecs first, if available
	if len(fpuSpecs) > 0 {
		fpuSpecs = strings.TrimRight(fpuSpecs, "-")
		fullPrecompDir := library.SourceDir.Join(mcu).Join(fpuSpecs)
		if fullPrecompDir.Exist() {
			logger.Fprintln(os.Stdout, constants.LOG_LEVEL_INFO, "Using precompiled library in {0}", fullPrecompDir)
			return fullPrecompDir
		}
		logger.Fprintln(os.Stdout, constants.LOG_LEVEL_INFO, "Precompiled library in \"{0}\" not found", fullPrecompDir)
	}

	precompDir := library.SourceDir.Join(mcu)
	if precompDir.Exist() {
		logger.Fprintln(os.Stdout, constants.LOG_LEVEL_INFO, "Using precompiled library in {0}", precompDir)
		return precompDir
	}
	logger.Fprintln(os.Stdout, constants.LOG_LEVEL_INFO, "Precompiled library in \"{0}\" not found", precompDir)
	return nil
}

func fixLDFLAG(ctx *types.Context, library *libraries.Library, path *paths.Path) {
	// find all library names in the folder and prepend -l
	filePaths := []string{}
	libsCmd := library.LDflags + " "
	extensions := func(ext string) bool {
		return PRECOMPILED_LIBRARIES_VALID_EXTENSIONS_DYNAMIC[ext] || PRECOMPILED_LIBRARIES_VALID_EXTENSIONS_STATIC[ext]
	}
	utils.FindFilesInFolder(&filePaths, path.String(), extensions, false)
	for _, lib := range filePaths {
		name := strings.TrimSuffix(filepath.Base(lib), filepath.Ext(lib))
		// strip "lib" first occurrence
		if strings.HasPrefix(name, "lib") {
			name = strings.Replace(name, "lib", "", 1)
			libsCmd += "-l" + name + " "
		}
	}

	currLDFlags := ctx.BuildProperties.Get(constants.BUILD_PROPERTIES_COMPILER_LIBRARIES_LDFLAGS)
	ctx.BuildProperties.Set(constants.BUILD_PROPERTIES_COMPILER_LIBRARIES_LDFLAGS, currLDFlags+"\"-L"+path.String()+"\" "+libsCmd+" ")
}

func compileLibraries(ctx *types.Context, libraries libraries.List, buildPath *paths.Path, buildProperties *properties.Map, includes []string) (paths.PathList, error) {
	ctx.Progress.AddSubSteps(len(libraries))
	defer ctx.Progress.RemoveSubSteps()

	objectFiles := paths.NewPathList()
	for _, library := range libraries {
		libraryObjectFiles, err := compileLibrary(ctx, library, buildPath, buildProperties, includes)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		objectFiles = append(objectFiles, libraryObjectFiles...)

		ctx.Progress.CompleteStep()
		builder_utils.PrintProgressIfProgressEnabledAndMachineLogger(ctx)
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
		return nil, errors.WithStack(err)
	}

	objectFiles := paths.NewPathList()

	if library.Precompiled {
		if precompiledPath := findExpectedPrecompiledLibFolder(ctx, library); precompiledPath != nil {
			// Add required LD flags
			fixLDFLAG(ctx, library, precompiledPath)

			// TODO: This codepath is just taken for .a with unusual names that would
			// be ignored by -L / -l methods.
			// Should we force precompiled libraries to start with "lib" ?
			extensions := func(ext string) bool { return PRECOMPILED_LIBRARIES_VALID_EXTENSIONS_STATIC[ext] }
			filePaths := []string{}
			err := utils.FindFilesInFolder(&filePaths, precompiledPath.String(), extensions, false)
			if err != nil {
				return nil, errors.WithStack(err)
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
			return nil, errors.WithStack(err)
		}
		if library.DotALinkage {
			archiveFile, err := builder_utils.ArchiveCompiledFiles(ctx, libraryBuildPath, paths.New(library.Name+".a"), libObjectFiles, buildProperties)
			if err != nil {
				return nil, errors.WithStack(err)
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
			return nil, errors.WithStack(err)
		}
		objectFiles.AddAll(libObjectFiles)

		if library.UtilityDir != nil {
			utilityBuildPath := libraryBuildPath.Join("utility")
			utilityObjectFiles, err := builder_utils.CompileFiles(ctx, library.UtilityDir, false, utilityBuildPath, buildProperties, includes)
			if err != nil {
				return nil, errors.WithStack(err)
			}
			objectFiles.AddAll(utilityObjectFiles)
		}
	}

	return objectFiles, nil
}
