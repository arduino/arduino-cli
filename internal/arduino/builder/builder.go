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

package builder

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/arduino/arduino-cli/internal/arduino/builder/internal/compilation"
	"github.com/arduino/arduino-cli/internal/arduino/builder/internal/detector"
	"github.com/arduino/arduino-cli/internal/arduino/builder/internal/diagnostics"
	"github.com/arduino/arduino-cli/internal/arduino/builder/internal/progress"
	"github.com/arduino/arduino-cli/internal/arduino/builder/internal/utils"
	"github.com/arduino/arduino-cli/internal/arduino/builder/logger"
	"github.com/arduino/arduino-cli/internal/arduino/cores"
	"github.com/arduino/arduino-cli/internal/arduino/libraries"
	"github.com/arduino/arduino-cli/internal/arduino/libraries/librariesmanager"
	"github.com/arduino/arduino-cli/internal/arduino/sketch"
	"github.com/arduino/arduino-cli/internal/i18n"
	"github.com/arduino/arduino-cli/pkg/fqbn"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	"github.com/arduino/go-paths-helper"
	"github.com/arduino/go-properties-orderedmap"
)

// ErrSketchCannotBeLocatedInBuildPath fixdoc
var ErrSketchCannotBeLocatedInBuildPath = errors.New("sketch cannot be located in build path")

// Builder is a Sketch builder.
type Builder struct {
	ctx context.Context

	sketch          *sketch.Sketch
	buildProperties *properties.Map

	buildPath          *paths.Path
	sketchBuildPath    *paths.Path
	coreBuildPath      *paths.Path
	librariesBuildPath *paths.Path

	// Parallel processes
	jobs int

	// core related
	coreBuildCachePath       *paths.Path
	extraCoreBuildCachePaths paths.PathList

	logger *logger.BuilderLogger
	clean  bool

	// Source code overrides (filename -> content map).
	// The provided source data is used instead of reading it from disk.
	// The keys of the map are paths relative to sketch folder.
	sourceOverrides map[string]string

	// Set to true to skip build and produce only Compilation Database
	onlyUpdateCompilationDatabase bool
	// Compilation Database to build/update
	compilationDatabase *compilation.Database

	// Progress of all various steps
	Progress *progress.Struct

	// Sizer results
	executableSectionsSize ExecutablesFileSections

	// C++ Parsing
	lineOffset int

	targetPlatform *cores.PlatformRelease
	buildPlatform  *cores.PlatformRelease

	buildArtifacts *buildArtifacts

	buildOptions *buildOptions

	libsDetector *detector.SketchLibrariesDetector

	toolEnv []string

	diagnosticStore *diagnostics.Store
}

// buildArtifacts contains the result of various build
type buildArtifacts struct {
	// populated by BuildCore
	coreArchiveFilePath *paths.Path
	coreObjectsFiles    paths.PathList

	// populated by BuildLibraries
	librariesObjectFiles paths.PathList

	// populated by BuildSketch
	sketchObjectFiles paths.PathList
}

// NewBuilder creates a sketch Builder.
func NewBuilder(
	ctx context.Context,
	sk *sketch.Sketch,
	boardBuildProperties *properties.Map,
	buildPath *paths.Path,
	optimizeForDebug bool,
	coreBuildCachePath *paths.Path,
	extraCoreBuildCachePaths paths.PathList,
	jobs int,
	customBuildProperties []string,
	hardwareDirs paths.PathList,
	librariesDirs paths.PathList,
	builtInLibrariesDirs *paths.Path,
	fqbn *fqbn.FQBN,
	clean bool,
	sourceOverrides map[string]string,
	onlyUpdateCompilationDatabase bool,
	targetPlatform, buildPlatform *cores.PlatformRelease,
	useCachedLibrariesResolution bool,
	librariesManager *librariesmanager.LibrariesManager,
	customLibraryDirs paths.PathList,
	stdout, stderr io.Writer, verbosity logger.Verbosity, warningsLevel string,
	progressCB rpc.TaskProgressCB,
	toolEnv []string,
) (*Builder, error) {
	buildProperties := properties.NewMap()
	if boardBuildProperties != nil {
		buildProperties.Merge(boardBuildProperties)
	}
	if sk != nil {
		buildProperties.SetPath("sketch_path", sk.FullPath)
		buildProperties.Set("build.project_name", sk.MainFile.Base())
		buildProperties.SetPath("build.source.path", sk.FullPath)
	}
	if buildPath != nil {
		buildProperties.SetPath("build.path", buildPath)
	}
	if optimizeForDebug {
		if debugFlags, ok := buildProperties.GetOk("compiler.optimization_flags.debug"); ok {
			buildProperties.Set("compiler.optimization_flags", debugFlags)
		}
	} else {
		if releaseFlags, ok := buildProperties.GetOk("compiler.optimization_flags.release"); ok {
			buildProperties.Set("compiler.optimization_flags", releaseFlags)
		}
	}

	// Add user provided custom build properties
	if p, err := properties.LoadFromSlice(customBuildProperties); err == nil {
		buildProperties.Merge(p)
	} else {
		return nil, fmt.Errorf("invalid build properties: %w", err)
	}

	sketchBuildPath, err := buildPath.Join("sketch").Abs()
	if err != nil {
		return nil, err
	}
	librariesBuildPath, err := buildPath.Join("libraries").Abs()
	if err != nil {
		return nil, err
	}
	coreBuildPath, err := buildPath.Join("core").Abs()
	if err != nil {
		return nil, err
	}

	if buildPath.Canonical().EqualsTo(sk.FullPath.Canonical()) {
		return nil, ErrSketchCannotBeLocatedInBuildPath
	}

	log := logger.New(stdout, stderr, verbosity, warningsLevel)
	libsManager, libsResolver, libsLoadingWarnings, err := detector.LibrariesLoader(
		useCachedLibrariesResolution,
		librariesManager,
		builtInLibrariesDirs,
		customLibraryDirs,
		librariesDirs,
		buildPlatform,
		targetPlatform,
	)
	if err != nil {
		return nil, err
	}
	if log.VerbosityLevel() == logger.VerbosityVerbose {
		log.Warn(string(libsLoadingWarnings))
	}

	diagnosticStore := diagnostics.NewStore()
	b := &Builder{
		ctx:                           ctx,
		sketch:                        sk,
		buildProperties:               buildProperties,
		buildPath:                     buildPath,
		sketchBuildPath:               sketchBuildPath,
		coreBuildPath:                 coreBuildPath,
		librariesBuildPath:            librariesBuildPath,
		jobs:                          jobs,
		coreBuildCachePath:            coreBuildCachePath,
		extraCoreBuildCachePaths:      extraCoreBuildCachePaths,
		logger:                        log,
		clean:                         clean,
		sourceOverrides:               sourceOverrides,
		onlyUpdateCompilationDatabase: onlyUpdateCompilationDatabase,
		compilationDatabase:           compilation.NewDatabase(buildPath.Join("compile_commands.json")),
		Progress:                      progress.New(progressCB),
		executableSectionsSize:        []ExecutableSectionSize{},
		buildArtifacts:                &buildArtifacts{},
		targetPlatform:                targetPlatform,
		buildPlatform:                 buildPlatform,
		toolEnv:                       toolEnv,
		buildOptions: newBuildOptions(
			hardwareDirs,
			librariesDirs,
			builtInLibrariesDirs,
			buildPath,
			sk,
			customBuildProperties,
			fqbn,
			clean,
			buildProperties.Get("compiler.optimization_flags"),
			buildProperties.GetPath("runtime.platform.path"),
			buildProperties.GetPath("build.core.path"), // TODO can we buildCorePath ?
		),
		diagnosticStore: diagnosticStore,
		libsDetector: detector.NewSketchLibrariesDetector(
			libsManager, libsResolver,
			useCachedLibrariesResolution,
			onlyUpdateCompilationDatabase,
			log,
			diagnosticStore,
		),
	}
	return b, nil
}

// GetBuildProperties returns the build properties for running this build
func (b *Builder) GetBuildProperties() *properties.Map {
	return b.buildProperties
}

// GetBuildPath returns the build path
func (b *Builder) GetBuildPath() *paths.Path {
	return b.buildPath
}

// ExecutableSectionsSize fixdoc
func (b *Builder) ExecutableSectionsSize() ExecutablesFileSections {
	return b.executableSectionsSize
}

// ImportedLibraries fixdoc
func (b *Builder) ImportedLibraries() libraries.List {
	return b.libsDetector.ImportedLibraries()
}

// CompilerDiagnostics returns the parsed compiler diagnostics
func (b *Builder) CompilerDiagnostics() diagnostics.Diagnostics {
	return b.diagnosticStore.Diagnostics()
}

// Preprocess fixdoc
func (b *Builder) Preprocess() ([]byte, error) {
	b.Progress.AddSubSteps(6)
	defer b.Progress.RemoveSubSteps()

	if err := b.preprocess(); err != nil {
		return nil, err
	}

	// Return arduino-preprocessed source
	preprocessedSketch, err := b.sketchBuildPath.Join(b.sketch.MainFile.Base() + ".cpp").ReadFile()
	return preprocessedSketch, err
}

func (b *Builder) preprocess() error {
	if err := b.buildPath.MkdirAll(); err != nil {
		return err
	}

	if err := b.wipeBuildPathIfBuildOptionsChanged(); err != nil {
		return err
	}
	if err := b.createBuildOptionsJSON(); err != nil {
		return err
	}
	b.Progress.CompleteStep()

	if err := b.RunRecipe("recipe.hooks.prebuild", ".pattern", false); err != nil {
		return err
	}
	b.Progress.CompleteStep()

	if err := b.prepareSketchBuildPath(); err != nil {
		return err
	}
	b.Progress.CompleteStep()

	b.logIfVerbose(false, i18n.Tr("Detecting libraries used..."))
	err := b.libsDetector.FindIncludes(
		b.ctx,
		b.buildPath,
		b.buildProperties.GetPath("build.core.path"),
		b.buildProperties.GetPath("build.variant.path"),
		b.sketchBuildPath,
		b.sketch,
		b.librariesBuildPath,
		b.buildProperties,
		b.targetPlatform.Platform.Architecture,
	)
	if err != nil {
		return err
	}
	b.Progress.CompleteStep()

	b.warnAboutArchIncompatibleLibraries(b.libsDetector.ImportedLibraries())
	b.Progress.CompleteStep()

	b.logIfVerbose(false, i18n.Tr("Generating function prototypes..."))
	if err := b.preprocessSketch(b.libsDetector.IncludeFolders()); err != nil {
		return err
	}
	b.Progress.CompleteStep()

	return nil
}

func (b *Builder) logIfVerbose(warn bool, msg string) {
	if b.logger.VerbosityLevel() != logger.VerbosityVerbose {
		return
	}
	if warn {
		b.logger.Warn(msg)
		return
	}
	b.logger.Info(msg)
}

// Build fixdoc
func (b *Builder) Build() error {
	b.Progress.AddSubSteps(6 + // preprocess
		18 + // build
		1, // size
	)
	defer b.Progress.RemoveSubSteps()

	if err := b.preprocess(); err != nil {
		return err
	}

	buildErr := b.build()

	b.libsDetector.PrintUsedAndNotUsedLibraries(buildErr != nil)
	b.printUsedLibraries(b.libsDetector.ImportedLibraries())
	if buildErr != nil {
		return buildErr
	}

	if err := b.size(); err != nil {
		return err
	}
	b.Progress.CompleteStep()

	return nil
}

// Build fixdoc
func (b *Builder) build() error {
	b.logIfVerbose(false, i18n.Tr("Compiling sketch..."))
	if err := b.RunRecipe("recipe.hooks.sketch.prebuild", ".pattern", false); err != nil {
		return err
	}
	b.Progress.CompleteStep()

	if err := b.buildSketch(b.libsDetector.IncludeFolders()); err != nil {
		return err
	}
	b.Progress.CompleteStep()

	if err := b.RunRecipe("recipe.hooks.sketch.postbuild", ".pattern", true); err != nil {
		return err
	}
	b.Progress.CompleteStep()

	b.logIfVerbose(false, i18n.Tr("Compiling libraries..."))
	if err := b.RunRecipe("recipe.hooks.libraries.prebuild", ".pattern", false); err != nil {
		return err
	}
	b.Progress.CompleteStep()

	if err := b.removeUnusedCompiledLibraries(b.libsDetector.ImportedLibraries()); err != nil {
		return err
	}
	b.Progress.CompleteStep()

	if err := b.buildLibraries(b.libsDetector.IncludeFolders(), b.libsDetector.ImportedLibraries()); err != nil {
		return err
	}
	b.Progress.CompleteStep()

	if err := b.RunRecipe("recipe.hooks.libraries.postbuild", ".pattern", true); err != nil {
		return err
	}
	b.Progress.CompleteStep()

	b.logIfVerbose(false, i18n.Tr("Compiling core..."))
	if err := b.RunRecipe("recipe.hooks.core.prebuild", ".pattern", false); err != nil {
		return err
	}
	b.Progress.CompleteStep()

	if err := b.buildCore(); err != nil {
		return err
	}
	b.Progress.CompleteStep()

	if err := b.RunRecipe("recipe.hooks.core.postbuild", ".pattern", true); err != nil {
		return err
	}
	b.Progress.CompleteStep()

	b.logIfVerbose(false, i18n.Tr("Linking everything together..."))
	if err := b.RunRecipe("recipe.hooks.linking.prelink", ".pattern", false); err != nil {
		return err
	}
	b.Progress.CompleteStep()

	if err := b.link(); err != nil {
		return err
	}
	b.Progress.CompleteStep()

	if err := b.RunRecipe("recipe.hooks.linking.postlink", ".pattern", true); err != nil {
		return err
	}
	b.Progress.CompleteStep()

	if err := b.RunRecipe("recipe.hooks.objcopy.preobjcopy", ".pattern", false); err != nil {
		return err
	}
	b.Progress.CompleteStep()

	if err := b.RunRecipe("recipe.objcopy.", ".pattern", true); err != nil {
		return err
	}
	b.Progress.CompleteStep()

	if err := b.RunRecipe("recipe.hooks.objcopy.postobjcopy", ".pattern", true); err != nil {
		return err
	}
	b.Progress.CompleteStep()

	if err := b.mergeSketchWithBootloader(); err != nil {
		return err
	}
	b.Progress.CompleteStep()

	if err := b.RunRecipe("recipe.hooks.postbuild", ".pattern", true); err != nil {
		return err
	}
	b.Progress.CompleteStep()

	if b.compilationDatabase != nil {
		b.compilationDatabase.SaveToFile()
	}
	return nil
}

func (b *Builder) prepareCommandForRecipe(buildProperties *properties.Map, recipe string, removeUnsetProperties bool) (*paths.Process, error) {
	pattern := buildProperties.Get(recipe)
	if pattern == "" {
		return nil, errors.New(i18n.Tr("%[1]s pattern is missing", recipe))
	}

	commandLine := buildProperties.ExpandPropsInString(pattern)
	if removeUnsetProperties {
		commandLine = properties.DeleteUnexpandedPropsFromString(commandLine)
	}

	args, _ := properties.SplitQuotedString(commandLine, `"'`, false)

	// if the overall commandline is too long for the platform
	// try reducing the length by making the filenames relative
	// and changing working directory to build.path
	var relativePath string
	if len(commandLine) > 30000 {
		relativePath = buildProperties.Get("build.path")
		for i, arg := range args {
			if _, err := os.Stat(arg); os.IsNotExist(err) {
				continue
			}
			rel, err := filepath.Rel(relativePath, arg)
			if err == nil && !strings.Contains(rel, "..") && len(rel) < len(arg) {
				args[i] = rel
			}
		}
	}

	command, err := paths.NewProcess(b.toolEnv, args...)
	if err != nil {
		return nil, err
	}
	if relativePath != "" {
		command.SetDir(relativePath)
	}

	return command, nil
}

func (b *Builder) execCommand(command *paths.Process) error {
	if b.logger.VerbosityLevel() == logger.VerbosityVerbose {
		b.logger.Info(utils.PrintableCommand(command.GetArgs()))
		command.RedirectStdoutTo(b.logger.Stdout())
	}
	command.RedirectStderrTo(b.logger.Stderr())

	if err := command.Start(); err != nil {
		return err
	}

	return command.Wait()
}
