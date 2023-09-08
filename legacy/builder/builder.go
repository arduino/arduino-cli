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
	"reflect"
	"time"

	"github.com/arduino/arduino-cli/arduino/builder/preprocessor"
	"github.com/arduino/arduino-cli/i18n"
	"github.com/arduino/arduino-cli/legacy/builder/phases"
	"github.com/arduino/arduino-cli/legacy/builder/types"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

var tr = i18n.Tr

const DEFAULT_DEBUG_LEVEL = 5
const DEFAULT_WARNINGS_LEVEL = "none"

type Builder struct{}

func (s *Builder) Run(ctx *types.Context) error {
	if err := ctx.BuildPath.MkdirAll(); err != nil {
		return err
	}

	var _err, mainErr error
	commands := []types.Command{
		containerBuildOptions(ctx),

		types.BareCommand(func(ctx *types.Context) error {
			return recipeByPrefixSuffixRunner(ctx, "recipe.hooks.prebuild", ".pattern", false)
		}),

		types.BareCommand(func(ctx *types.Context) error {
			ctx.LineOffset, _err = ctx.Builder.PrepareSketchBuildPath(ctx.SourceOverride, ctx.SketchBuildPath)
			return _err
		}),

		logIfVerbose(false, tr("Detecting libraries used...")),
		findIncludes(ctx),

		&WarnAboutArchIncompatibleLibraries{},

		logIfVerbose(false, tr("Generating function prototypes...")),
		types.BareCommand(PreprocessSketch),

		logIfVerbose(false, tr("Compiling sketch...")),

		types.BareCommand(func(ctx *types.Context) error {
			return recipeByPrefixSuffixRunner(ctx, "recipe.hooks.sketch.prebuild", ".pattern", false)
		}),

		types.BareCommand(func(ctx *types.Context) error {
			sketchObjectFiles, err := phases.SketchBuilder(
				ctx.SketchBuildPath,
				ctx.BuildProperties,
				ctx.SketchLibrariesDetector.IncludeFolders(),
				ctx.OnlyUpdateCompilationDatabase,
				ctx.Verbose,
				ctx.CompilationDatabase,
				ctx.Jobs,
				ctx.WarningsLevel,
				ctx.Stdout, ctx.Stderr,
				func(msg string) { ctx.Info(msg) },
				func(data []byte) { ctx.WriteStdout(data) },
				func(data []byte) { ctx.WriteStderr(data) },
				&ctx.Progress, ctx.ProgressCB,
			)
			if err != nil {
				return err
			}
			ctx.SketchObjectFiles = sketchObjectFiles
			return nil
		}),

		types.BareCommand(func(ctx *types.Context) error {
			return recipeByPrefixSuffixRunner(ctx, "recipe.hooks.sketch.postbuild", ".pattern", true)
		}),

		logIfVerbose(false, tr("Compiling libraries...")),
		types.BareCommand(func(ctx *types.Context) error {
			return recipeByPrefixSuffixRunner(ctx, "recipe.hooks.libraries.prebuild", ".pattern", false)
		}),

		types.BareCommand(func(ctx *types.Context) error {
			return UnusedCompiledLibrariesRemover(
				ctx.LibrariesBuildPath,
				ctx.SketchLibrariesDetector.ImportedLibraries(),
			)
		}),

		types.BareCommand(func(ctx *types.Context) error {
			librariesObjectFiles, err := phases.LibrariesBuilder(
				ctx.LibrariesBuildPath,
				ctx.BuildProperties,
				ctx.SketchLibrariesDetector.IncludeFolders(),
				ctx.SketchLibrariesDetector.ImportedLibraries(),
				ctx.Verbose,
				ctx.OnlyUpdateCompilationDatabase,
				ctx.CompilationDatabase,
				ctx.Jobs,
				ctx.WarningsLevel,
				ctx.Stdout,
				ctx.Stderr,
				func(msg string) { ctx.Info(msg) },
				func(data []byte) { ctx.WriteStdout(data) },
				func(data []byte) { ctx.WriteStderr(data) },
				&ctx.Progress, ctx.ProgressCB,
			)
			if err != nil {
				return err
			}
			ctx.LibrariesObjectFiles = librariesObjectFiles

			return nil
		}),
		types.BareCommand(func(ctx *types.Context) error {
			return recipeByPrefixSuffixRunner(ctx, "recipe.hooks.libraries.postbuild", ".pattern", true)
		}),

		logIfVerbose(false, tr("Compiling core...")),
		types.BareCommand(func(ctx *types.Context) error {
			return recipeByPrefixSuffixRunner(ctx, "recipe.hooks.core.prebuild", ".pattern", false)
		}),

		types.BareCommand(func(ctx *types.Context) error {
			objectFiles, archiveFile, err := phases.CoreBuilder(
				ctx.BuildPath, ctx.CoreBuildPath, ctx.Builder.CoreBuildCachePath(),
				ctx.BuildProperties,
				ctx.ActualPlatform,
				ctx.Verbose, ctx.OnlyUpdateCompilationDatabase, ctx.Clean,
				ctx.CompilationDatabase,
				ctx.Jobs,
				ctx.WarningsLevel,
				ctx.Stdout, ctx.Stderr,
				func(msg string) { ctx.Info(msg) },
				func(data []byte) { ctx.WriteStdout(data) },
				func(data []byte) { ctx.WriteStderr(data) },
				&ctx.Progress, ctx.ProgressCB,
			)

			ctx.CoreObjectsFiles = objectFiles
			ctx.CoreArchiveFilePath = archiveFile

			return err
		}),

		types.BareCommand(func(ctx *types.Context) error {
			return recipeByPrefixSuffixRunner(ctx, "recipe.hooks.core.postbuild", ".pattern", true)
		}),

		logIfVerbose(false, tr("Linking everything together...")),
		types.BareCommand(func(ctx *types.Context) error {
			return recipeByPrefixSuffixRunner(ctx, "recipe.hooks.linking.prelink", ".pattern", false)
		}),

		types.BareCommand(func(ctx *types.Context) error {
			verboseInfoOut, err := phases.Linker(
				ctx.OnlyUpdateCompilationDatabase,
				ctx.Verbose,
				ctx.SketchObjectFiles,
				ctx.LibrariesObjectFiles,
				ctx.CoreObjectsFiles,
				ctx.CoreArchiveFilePath,
				ctx.BuildPath,
				ctx.BuildProperties,
				ctx.Stdout,
				ctx.Stderr,
				ctx.WarningsLevel,
			)
			if ctx.Verbose {
				ctx.Info(string(verboseInfoOut))
			}
			return err
		}),

		types.BareCommand(func(ctx *types.Context) error {
			return recipeByPrefixSuffixRunner(ctx, "recipe.hooks.linking.postlink", ".pattern", true)
		}),

		types.BareCommand(func(ctx *types.Context) error {
			return recipeByPrefixSuffixRunner(ctx, "recipe.hooks.objcopy.preobjcopy", ".pattern", false)
		}),
		types.BareCommand(func(ctx *types.Context) error {
			return recipeByPrefixSuffixRunner(ctx, "recipe.objcopy.", ".pattern", true)
		}),
		types.BareCommand(func(ctx *types.Context) error {
			return recipeByPrefixSuffixRunner(ctx, "recipe.hooks.objcopy.postobjcopy", ".pattern", true)
		}),

		types.BareCommand(func(ctx *types.Context) error {
			return MergeSketchWithBootloader(
				ctx.OnlyUpdateCompilationDatabase, ctx.Verbose,
				ctx.BuildPath, ctx.Sketch, ctx.BuildProperties,
				func(s string) { ctx.Info(s) },
				func(s string) { ctx.Warn(s) },
			)
		}),

		types.BareCommand(func(ctx *types.Context) error {
			return recipeByPrefixSuffixRunner(ctx, "recipe.hooks.postbuild", ".pattern", true)
		}),
	}

	ctx.Progress.AddSubSteps(len(commands) + 5)
	defer ctx.Progress.RemoveSubSteps()

	for _, command := range commands {
		PrintRingNameIfDebug(ctx, command)
		err := command.Run(ctx)
		if err != nil {
			mainErr = errors.WithStack(err)
			break
		}
		ctx.Progress.CompleteStep()
		ctx.PushProgress()
	}

	if ctx.CompilationDatabase != nil {
		ctx.CompilationDatabase.SaveToFile()
	}

	var otherErr error
	commands = []types.Command{
		types.BareCommand(func(ctx *types.Context) error {
			ctx.SketchLibrariesDetector.PrintUsedAndNotUsedLibraries(mainErr != nil)
			return nil
		}),

		types.BareCommand(func(ctx *types.Context) error {
			infoOut, _ := PrintUsedLibrariesIfVerbose(ctx.Verbose, ctx.SketchLibrariesDetector.ImportedLibraries())
			ctx.Info(string(infoOut))
			return nil
		}),

		&ExportProjectCMake{SketchError: mainErr != nil},

		types.BareCommand(func(ctx *types.Context) error {
			executableSectionsSize, err := phases.Sizer(
				ctx.OnlyUpdateCompilationDatabase, mainErr != nil, ctx.Verbose,
				ctx.BuildProperties,
				ctx.Stdout, ctx.Stderr,
				func(msg string) { ctx.Info(msg) },
				func(msg string) { ctx.Warn(msg) },
				ctx.WarningsLevel,
			)
			ctx.ExecutableSectionsSize = executableSectionsSize
			return err
		}),
	}
	for _, command := range commands {
		PrintRingNameIfDebug(ctx, command)
		err := command.Run(ctx)
		if err != nil {
			otherErr = errors.WithStack(err)
			break
		}
		ctx.Progress.CompleteStep()
		ctx.PushProgress()
	}

	if mainErr != nil {
		return mainErr
	}

	return otherErr
}

func PreprocessSketch(ctx *types.Context) error {
	preprocessorImpl := preprocessor.PreprocessSketchWithCtags
	normalOutput, verboseOutput, err := preprocessorImpl(
		ctx.Sketch, ctx.BuildPath, ctx.SketchLibrariesDetector.IncludeFolders(), ctx.LineOffset,
		ctx.BuildProperties, ctx.OnlyUpdateCompilationDatabase)
	if ctx.Verbose {
		ctx.WriteStdout(verboseOutput)
	} else {
		ctx.WriteStdout(normalOutput)
	}
	return err
}

type Preprocess struct{}

func (s *Preprocess) Run(ctx *types.Context) error {
	if err := ctx.BuildPath.MkdirAll(); err != nil {
		return err
	}

	var _err error
	commands := []types.Command{
		containerBuildOptions(ctx),

		types.BareCommand(func(ctx *types.Context) error {
			return recipeByPrefixSuffixRunner(ctx, "recipe.hooks.prebuild", ".pattern", false)
		}),

		types.BareCommand(func(ctx *types.Context) error {
			ctx.LineOffset, _err = ctx.Builder.PrepareSketchBuildPath(ctx.SourceOverride, ctx.SketchBuildPath)
			return _err
		}),

		findIncludes(ctx),

		&WarnAboutArchIncompatibleLibraries{},

		types.BareCommand(PreprocessSketch),
	}

	if err := runCommands(ctx, commands); err != nil {
		return err
	}

	// Output arduino-preprocessed source
	preprocessedSketch, err := ctx.SketchBuildPath.Join(ctx.Sketch.MainFile.Base() + ".cpp").ReadFile()
	if err != nil {
		return err
	}
	ctx.WriteStdout(preprocessedSketch)
	return nil
}

func runCommands(ctx *types.Context, commands []types.Command) error {
	ctx.Progress.AddSubSteps(len(commands))
	defer ctx.Progress.RemoveSubSteps()

	for _, command := range commands {
		PrintRingNameIfDebug(ctx, command)
		err := command.Run(ctx)
		if err != nil {
			return errors.WithStack(err)
		}
		ctx.Progress.CompleteStep()
		ctx.PushProgress()
	}
	return nil
}

func PrintRingNameIfDebug(ctx *types.Context, command types.Command) {
	logrus.Debugf("Ts: %d - Running: %s", time.Now().Unix(), reflect.Indirect(reflect.ValueOf(command)).Type().Name())
}

func RunBuilder(ctx *types.Context) error {
	return runCommands(ctx, []types.Command{&Builder{}})
}

func RunPreprocess(ctx *types.Context) error {
	command := Preprocess{}
	return command.Run(ctx)
}

func findIncludes(ctx *types.Context) types.BareCommand {
	return types.BareCommand(func(ctx *types.Context) error {
		return ctx.SketchLibrariesDetector.FindIncludes(
			ctx.BuildPath,
			ctx.BuildProperties.GetPath("build.core.path"),
			ctx.BuildProperties.GetPath("build.variant.path"),
			ctx.SketchBuildPath,
			ctx.Sketch,
			ctx.LibrariesBuildPath,
			ctx.BuildProperties,
			ctx.TargetPlatform.Platform.Architecture,
		)
	})
}

func logIfVerbose(warn bool, msg string) types.BareCommand {
	return types.BareCommand(func(ctx *types.Context) error {
		if !ctx.Verbose {
			return nil
		}
		if warn {
			ctx.Warn(msg)
		} else {
			ctx.Info(msg)
		}
		return nil
	})
}

func recipeByPrefixSuffixRunner(ctx *types.Context, prefix, suffix string, skipIfOnlyUpdatingCompilationDatabase bool) error {
	return RecipeByPrefixSuffixRunner(
		prefix, suffix, skipIfOnlyUpdatingCompilationDatabase,
		ctx.OnlyUpdateCompilationDatabase, ctx.Verbose,
		ctx.BuildProperties, ctx.Stdout, ctx.Stderr,
		func(msg string) { ctx.Info(msg) },
	)
}

func containerBuildOptions(ctx *types.Context) types.BareCommand {
	return types.BareCommand(func(ctx *types.Context) error {
		// TODO here we can pass only the properties we're reading from the
		// ctx.BuildProperties
		buildOptionsJSON, buildOptionsJSONPrevious, infoMessage, err := ContainerBuildOptions(
			ctx.HardwareDirs, ctx.BuiltInToolsDirs, ctx.OtherLibrariesDirs,
			ctx.BuiltInLibrariesDirs, ctx.BuildPath, ctx.Sketch, ctx.CustomBuildProperties,
			ctx.FQBN.String(), ctx.Clean, ctx.BuildProperties,
		)
		if infoMessage != "" {
			ctx.Info(infoMessage)
		}
		if err != nil {
			return err
		}

		ctx.BuildOptionsJson = buildOptionsJSON
		ctx.BuildOptionsJsonPrevious = buildOptionsJSONPrevious

		return nil
	})
}
