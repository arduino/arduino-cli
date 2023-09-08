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
	"io"
	"strings"

	"github.com/arduino/arduino-cli/arduino/builder"
	"github.com/arduino/arduino-cli/arduino/builder/cpp"
	"github.com/arduino/arduino-cli/arduino/builder/progress"
	"github.com/arduino/arduino-cli/arduino/builder/utils"
	"github.com/arduino/arduino-cli/arduino/libraries"
	f "github.com/arduino/arduino-cli/internal/algorithms"
	"github.com/arduino/arduino-cli/legacy/builder/constants"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	"github.com/arduino/go-paths-helper"
	"github.com/arduino/go-properties-orderedmap"
	"github.com/pkg/errors"
)

var FLOAT_ABI_CFLAG = "float-abi"
var FPU_CFLAG = "fpu"

func LibrariesBuilder(
	librariesBuildPath *paths.Path,
	buildProperties *properties.Map,
	includesFolders paths.PathList,
	importedLibraries libraries.List,
	verbose, onlyUpdateCompilationDatabase bool,
	compilationDatabase *builder.CompilationDatabase,
	jobs int,
	warningsLevel string,
	stdoutWriter, stderrWriter io.Writer,
	verboseInfoFn func(msg string),
	verboseStdoutFn, verboseStderrFn func(data []byte),
	progress *progress.Struct, progressCB rpc.TaskProgressCB,
) (paths.PathList, error) {
	includes := f.Map(includesFolders.AsStrings(), cpp.WrapWithHyphenI)
	libs := importedLibraries

	if err := librariesBuildPath.MkdirAll(); err != nil {
		return nil, errors.WithStack(err)
	}

	librariesObjectFiles, err := compileLibraries(
		libs, librariesBuildPath, buildProperties, includes,
		verbose, onlyUpdateCompilationDatabase,
		compilationDatabase,
		jobs,
		warningsLevel,
		stdoutWriter, stderrWriter,
		verboseInfoFn,
		verboseStdoutFn, verboseStderrFn,
		progress, progressCB,
	)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return librariesObjectFiles, nil
}

func directoryContainsFile(folder *paths.Path) bool {
	if files, err := folder.ReadDir(); err == nil {
		files.FilterOutDirs()
		return len(files) > 0
	}
	return false
}

func findExpectedPrecompiledLibFolder(
	library *libraries.Library,
	buildProperties *properties.Map,
	verboseInfoFn func(msg string),
) *paths.Path {
	mcu := buildProperties.Get(constants.BUILD_PROPERTIES_BUILD_MCU)
	// Add fpu specifications if they exist
	// To do so, resolve recipe.cpp.o.pattern,
	// search for -mfpu=xxx -mfloat-abi=yyy and add to a subfolder
	command, _ := utils.PrepareCommandForRecipe(buildProperties, "recipe.cpp.o.pattern", true)
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

	verboseInfoFn(tr("Library %[1]s has been declared precompiled:", library.Name))

	// Try directory with full fpuSpecs first, if available
	if len(fpuSpecs) > 0 {
		fpuSpecs = strings.TrimRight(fpuSpecs, "-")
		fullPrecompDir := library.SourceDir.Join(mcu).Join(fpuSpecs)
		if fullPrecompDir.Exist() && directoryContainsFile(fullPrecompDir) {
			verboseInfoFn(tr("Using precompiled library in %[1]s", fullPrecompDir))
			return fullPrecompDir
		}
		verboseInfoFn(tr(`Precompiled library in "%[1]s" not found`, fullPrecompDir))
	}

	precompDir := library.SourceDir.Join(mcu)
	if precompDir.Exist() && directoryContainsFile(precompDir) {
		verboseInfoFn(tr("Using precompiled library in %[1]s", precompDir))
		return precompDir
	}
	verboseInfoFn(tr(`Precompiled library in "%[1]s" not found`, precompDir))
	return nil
}

func compileLibraries(
	libraries libraries.List, buildPath *paths.Path, buildProperties *properties.Map, includes []string,
	verbose, onlyUpdateCompilationDatabase bool,
	compilationDatabase *builder.CompilationDatabase,
	jobs int,
	warningsLevel string,
	stdoutWriter, stderrWriter io.Writer,
	verboseInfoFn func(msg string),
	verboseStdoutFn, verboseStderrFn func(data []byte),
	progress *progress.Struct, progressCB rpc.TaskProgressCB,
) (paths.PathList, error) {
	progress.AddSubSteps(len(libraries))
	defer progress.RemoveSubSteps()

	objectFiles := paths.NewPathList()
	for _, library := range libraries {
		libraryObjectFiles, err := compileLibrary(
			library, buildPath, buildProperties, includes,
			verbose, onlyUpdateCompilationDatabase,
			compilationDatabase,
			jobs,
			warningsLevel,
			stdoutWriter, stderrWriter,
			verboseInfoFn, verboseStdoutFn, verboseStderrFn,
			progress, progressCB,
		)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		objectFiles.AddAll(libraryObjectFiles)

		progress.CompleteStep()
		// PushProgress
		if progressCB != nil {
			progressCB(&rpc.TaskProgress{
				Percent:   progress.Progress,
				Completed: progress.Progress >= 100.0,
			})
		}
	}

	return objectFiles, nil
}

func compileLibrary(
	library *libraries.Library, buildPath *paths.Path, buildProperties *properties.Map, includes []string,
	verbose, onlyUpdateCompilationDatabase bool,
	compilationDatabase *builder.CompilationDatabase,
	jobs int,
	warningsLevel string,
	stdoutWriter, stderrWriter io.Writer,
	verboseInfoFn func(msg string),
	verboseStdoutFn, verboseStderrFn func(data []byte),
	progress *progress.Struct, progressCB rpc.TaskProgressCB,
) (paths.PathList, error) {
	if verbose {
		verboseInfoFn(tr(`Compiling library "%[1]s"`, library.Name))
	}
	libraryBuildPath := buildPath.Join(library.DirName)

	if err := libraryBuildPath.MkdirAll(); err != nil {
		return nil, errors.WithStack(err)
	}

	objectFiles := paths.NewPathList()

	if library.Precompiled {
		coreSupportPrecompiled := buildProperties.ContainsKey("compiler.libraries.ldflags")
		precompiledPath := findExpectedPrecompiledLibFolder(
			library,
			buildProperties,
			verboseInfoFn,
		)

		if !coreSupportPrecompiled {
			verboseInfoFn(tr("The platform does not support '%[1]s' for precompiled libraries.", "compiler.libraries.ldflags"))
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

			currLDFlags := buildProperties.Get("compiler.libraries.ldflags")
			buildProperties.Set("compiler.libraries.ldflags", currLDFlags+" \"-L"+precompiledPath.String()+"\" "+libsCmd+" ")

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
			onlyUpdateCompilationDatabase,
			compilationDatabase,
			jobs,
			verbose,
			warningsLevel,
			stdoutWriter, stderrWriter,
			verboseInfoFn, verboseStdoutFn, verboseStderrFn,
			progress, progressCB,
		)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		if library.DotALinkage {
			archiveFile, verboseInfo, err := utils.ArchiveCompiledFiles(
				libraryBuildPath, paths.New(library.DirName+".a"), libObjectFiles, buildProperties,
				onlyUpdateCompilationDatabase, verbose,
				stdoutWriter, stderrWriter,
			)
			if verbose {
				verboseInfoFn(string(verboseInfo))
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
			onlyUpdateCompilationDatabase,
			compilationDatabase,
			jobs,
			verbose,
			warningsLevel,
			stdoutWriter, stderrWriter,
			verboseInfoFn, verboseStdoutFn, verboseStderrFn,
			progress, progressCB,
		)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		objectFiles.AddAll(libObjectFiles)

		if library.UtilityDir != nil {
			utilityBuildPath := libraryBuildPath.Join("utility")
			utilityObjectFiles, err := utils.CompileFiles(
				library.UtilityDir, utilityBuildPath, buildProperties, includes,
				onlyUpdateCompilationDatabase,
				compilationDatabase,
				jobs,
				verbose,
				warningsLevel,
				stdoutWriter, stderrWriter,
				verboseInfoFn, verboseStdoutFn, verboseStderrFn,
				progress, progressCB,
			)
			if err != nil {
				return nil, errors.WithStack(err)
			}
			objectFiles.AddAll(utilityObjectFiles)
		}
	}

	return objectFiles, nil
}
