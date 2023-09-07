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
		&ContainerBuildOptions{},

		&RecipeByPrefixSuffixRunner{Prefix: "recipe.hooks.prebuild", Suffix: ".pattern"},

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
		&RecipeByPrefixSuffixRunner{Prefix: "recipe.hooks.sketch.prebuild", Suffix: ".pattern"},
		&phases.SketchBuilder{},
		&RecipeByPrefixSuffixRunner{Prefix: "recipe.hooks.sketch.postbuild", Suffix: ".pattern", SkipIfOnlyUpdatingCompilationDatabase: true},

		logIfVerbose(false, tr("Compiling libraries...")),
		&RecipeByPrefixSuffixRunner{Prefix: "recipe.hooks.libraries.prebuild", Suffix: ".pattern"},
		&UnusedCompiledLibrariesRemover{},
		&phases.LibrariesBuilder{},
		&RecipeByPrefixSuffixRunner{Prefix: "recipe.hooks.libraries.postbuild", Suffix: ".pattern", SkipIfOnlyUpdatingCompilationDatabase: true},

		logIfVerbose(false, tr("Compiling core...")),
		&RecipeByPrefixSuffixRunner{Prefix: "recipe.hooks.core.prebuild", Suffix: ".pattern"},

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

		&RecipeByPrefixSuffixRunner{Prefix: "recipe.hooks.core.postbuild", Suffix: ".pattern", SkipIfOnlyUpdatingCompilationDatabase: true},

		logIfVerbose(false, tr("Linking everything together...")),
		&RecipeByPrefixSuffixRunner{Prefix: "recipe.hooks.linking.prelink", Suffix: ".pattern"},

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

		&RecipeByPrefixSuffixRunner{Prefix: "recipe.hooks.linking.postlink", Suffix: ".pattern", SkipIfOnlyUpdatingCompilationDatabase: true},

		&RecipeByPrefixSuffixRunner{Prefix: "recipe.hooks.objcopy.preobjcopy", Suffix: ".pattern"},
		&RecipeByPrefixSuffixRunner{Prefix: "recipe.objcopy.", Suffix: ".pattern", SkipIfOnlyUpdatingCompilationDatabase: true},
		&RecipeByPrefixSuffixRunner{Prefix: "recipe.hooks.objcopy.postobjcopy", Suffix: ".pattern", SkipIfOnlyUpdatingCompilationDatabase: true},

		&MergeSketchWithBootloader{},

		&RecipeByPrefixSuffixRunner{Prefix: "recipe.hooks.postbuild", Suffix: ".pattern", SkipIfOnlyUpdatingCompilationDatabase: true},
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

		&PrintUsedLibrariesIfVerbose{},

		&ExportProjectCMake{SketchError: mainErr != nil},

		&phases.Sizer{SketchError: mainErr != nil},
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
		&ContainerBuildOptions{},

		&RecipeByPrefixSuffixRunner{Prefix: "recipe.hooks.prebuild", Suffix: ".pattern"},

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
