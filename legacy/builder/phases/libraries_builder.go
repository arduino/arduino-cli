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
	"strings"

	"github.com/arduino/arduino-cli/arduino/builder/cpp"
	"github.com/arduino/arduino-cli/arduino/builder/utils"
	"github.com/arduino/arduino-cli/arduino/libraries"
	f "github.com/arduino/arduino-cli/internal/algorithms"
	"github.com/arduino/arduino-cli/legacy/builder/constants"
	"github.com/arduino/arduino-cli/legacy/builder/types"
	"github.com/arduino/go-paths-helper"
	"github.com/arduino/go-properties-orderedmap"
	"github.com/pkg/errors"
)

var FLOAT_ABI_CFLAG = "float-abi"
var FPU_CFLAG = "fpu"

type LibrariesBuilder struct{}

func (s *LibrariesBuilder) Run(ctx *types.Context) error {
	librariesBuildPath := ctx.LibrariesBuildPath
	buildProperties := ctx.BuildProperties
	includesFolders := ctx.SketchLibrariesDetector.IncludeFolders()
	includes := f.Map(includesFolders.AsStrings(), cpp.WrapWithHyphenI)
	libs := ctx.SketchLibrariesDetector.ImportedLibraries()

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

func directoryContainsFile(folder *paths.Path) bool {
	if files, err := folder.ReadDir(); err == nil {
		files.FilterOutDirs()
		return len(files) > 0
	}
	return false
}

func findExpectedPrecompiledLibFolder(ctx *types.Context, library *libraries.Library) *paths.Path {
	mcu := ctx.BuildProperties.Get(constants.BUILD_PROPERTIES_BUILD_MCU)
	// Add fpu specifications if they exist
	// To do so, resolve recipe.cpp.o.pattern,
	// search for -mfpu=xxx -mfloat-abi=yyy and add to a subfolder
	command, _ := utils.PrepareCommandForRecipe(ctx.BuildProperties, "recipe.cpp.o.pattern", true)
	fpuSpecs := ""
	for _, el := range command.GetArgs() {
		if strings.Contains(el, FPU_CFLAG) {
			toAdd := strings.Split(el, "=")
			if len(toAdd) > 1 {
				fpuSpecs += strings.TrimSpace(toAdd[1]) + "-"
				break
			}
		}
	}
	for _, el := range command.GetArgs() {
		if strings.Contains(el, FLOAT_ABI_CFLAG) {
			toAdd := strings.Split(el, "=")
			if len(toAdd) > 1 {
				fpuSpecs += strings.TrimSpace(toAdd[1]) + "-"
				break
			}
		}
	}

	ctx.Info(tr("Library %[1]s has been declared precompiled:", library.Name))

	// Try directory with full fpuSpecs first, if available
	if len(fpuSpecs) > 0 {
		fpuSpecs = strings.TrimRight(fpuSpecs, "-")
		fullPrecompDir := library.SourceDir.Join(mcu).Join(fpuSpecs)
		if fullPrecompDir.Exist() && directoryContainsFile(fullPrecompDir) {
			ctx.Info(tr("Using precompiled library in %[1]s", fullPrecompDir))
			return fullPrecompDir
		}
		ctx.Info(tr(`Precompiled library in "%[1]s" not found`, fullPrecompDir))
	}

	precompDir := library.SourceDir.Join(mcu)
	if precompDir.Exist() && directoryContainsFile(precompDir) {
		ctx.Info(tr("Using precompiled library in %[1]s", precompDir))
		return precompDir
	}
	ctx.Info(tr(`Precompiled library in "%[1]s" not found`, precompDir))
	return nil
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
		objectFiles.AddAll(libraryObjectFiles)

		ctx.Progress.CompleteStep()
		ctx.PushProgress()
	}

	return objectFiles, nil
}

func compileLibrary(ctx *types.Context, library *libraries.Library, buildPath *paths.Path, buildProperties *properties.Map, includes []string) (paths.PathList, error) {
	if ctx.Verbose {
		ctx.Info(tr(`Compiling library "%[1]s"`, library.Name))
	}
	libraryBuildPath := buildPath.Join(library.DirName)

	if err := libraryBuildPath.MkdirAll(); err != nil {
		return nil, errors.WithStack(err)
	}

	objectFiles := paths.NewPathList()

	if library.Precompiled {
		coreSupportPrecompiled := ctx.BuildProperties.ContainsKey("compiler.libraries.ldflags")
		precompiledPath := findExpectedPrecompiledLibFolder(ctx, library)

		if !coreSupportPrecompiled {
			ctx.Info(tr("The platform does not support '%[1]s' for precompiled libraries.", "compiler.libraries.ldflags"))
		} else if precompiledPath != nil {
			// Find all libraries in precompiledPath
			libs, err := precompiledPath.ReadDir()
			if err != nil {
				return nil, errors.WithStack(err)
			}

			// Add required LD flags
			libsCmd := library.LDflags + " "
			dynAndStaticLibs := libs.Clone()
			dynAndStaticLibs.FilterSuffix(".a", ".so")
			for _, lib := range dynAndStaticLibs {
				name := strings.TrimSuffix(lib.Base(), lib.Ext())
				if strings.HasPrefix(name, "lib") {
					libsCmd += "-l" + name[3:] + " "
				}
			}

			currLDFlags := ctx.BuildProperties.Get("compiler.libraries.ldflags")
			ctx.BuildProperties.Set("compiler.libraries.ldflags", currLDFlags+" \"-L"+precompiledPath.String()+"\" "+libsCmd+" ")

			// TODO: This codepath is just taken for .a with unusual names that would
			// be ignored by -L / -l methods.
			// Should we force precompiled libraries to start with "lib" ?
			staticLibs := libs.Clone()
			staticLibs.FilterSuffix(".a")
			for _, lib := range staticLibs {
				if !strings.HasPrefix(lib.Base(), "lib") {
					objectFiles.Add(lib)
				}
			}

			if library.PrecompiledWithSources {
				return objectFiles, nil
			}
		}
	}

	if library.Layout == libraries.RecursiveLayout {
		libObjectFiles, err := utils.CompileFilesRecursive(
			library.SourceDir, libraryBuildPath, buildProperties, includes,
			ctx.OnlyUpdateCompilationDatabase,
			ctx.CompilationDatabase,
			ctx.Jobs,
			ctx.Verbose,
			ctx.WarningsLevel,
			ctx.Stdout, ctx.Stderr,
			func(msg string) { ctx.Info(msg) },
			func(data []byte) { ctx.WriteStdout(data) },
			func(data []byte) { ctx.WriteStderr(data) },
			&ctx.Progress, ctx.ProgressCB,
		)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		if library.DotALinkage {
			archiveFile, verboseInfo, err := utils.ArchiveCompiledFiles(
				libraryBuildPath, paths.New(library.DirName+".a"), libObjectFiles, buildProperties,
				ctx.OnlyUpdateCompilationDatabase, ctx.Verbose,
				ctx.Stdout, ctx.Stderr,
			)
			if ctx.Verbose {
				ctx.Info(string(verboseInfo))
			}
			if err != nil {
				return nil, errors.WithStack(err)
			}
			objectFiles.Add(archiveFile)
		} else {
			objectFiles.AddAll(libObjectFiles)
		}
	} else {
		if library.UtilityDir != nil {
			includes = append(includes, cpp.WrapWithHyphenI(library.UtilityDir.String()))
		}
		libObjectFiles, err := utils.CompileFiles(
			library.SourceDir, libraryBuildPath, buildProperties, includes,
			ctx.OnlyUpdateCompilationDatabase,
			ctx.CompilationDatabase,
			ctx.Jobs,
			ctx.Verbose,
			ctx.WarningsLevel,
			ctx.Stdout, ctx.Stderr,
			func(msg string) { ctx.Info(msg) },
			func(data []byte) { ctx.WriteStdout(data) },
			func(data []byte) { ctx.WriteStderr(data) },
			&ctx.Progress, ctx.ProgressCB,
		)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		objectFiles.AddAll(libObjectFiles)

		if library.UtilityDir != nil {
			utilityBuildPath := libraryBuildPath.Join("utility")
			utilityObjectFiles, err := utils.CompileFiles(
				library.UtilityDir, utilityBuildPath, buildProperties, includes,
				ctx.OnlyUpdateCompilationDatabase,
				ctx.CompilationDatabase,
				ctx.Jobs,
				ctx.Verbose,
				ctx.WarningsLevel,
				ctx.Stdout, ctx.Stderr,
				func(msg string) { ctx.Info(msg) },
				func(data []byte) { ctx.WriteStdout(data) },
				func(data []byte) { ctx.WriteStderr(data) },
				&ctx.Progress, ctx.ProgressCB,
			)
			if err != nil {
				return nil, errors.WithStack(err)
			}
			objectFiles.AddAll(utilityObjectFiles)
		}
	}

	return objectFiles, nil
}
