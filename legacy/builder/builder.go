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

	"github.com/arduino/arduino-cli/arduino/builder/sizer"
	"github.com/arduino/arduino-cli/i18n"
	"github.com/arduino/arduino-cli/legacy/builder/types"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

var tr = i18n.Tr

const DEFAULT_DEBUG_LEVEL = 5

type Builder struct{}

func (s *Builder) Run(ctx *types.Context) error {
	if err := ctx.Builder.GetBuildPath().MkdirAll(); err != nil {
		return err
	}

	var _err, mainErr error
	commands := []types.Command{
		containerBuildOptions(ctx),

		types.BareCommand(func(ctx *types.Context) error {
			return recipeByPrefixSuffixRunner(ctx, "recipe.hooks.prebuild", ".pattern", false)
		}),

		types.BareCommand(func(ctx *types.Context) error {
			ctx.LineOffset, _err = ctx.Builder.PrepareSketchBuildPath()
			return _err
		}),

		logIfVerbose(false, tr("Detecting libraries used...")),
		findIncludes(ctx),

		warnAboutArchIncompatibleLibraries(ctx),

		logIfVerbose(false, tr("Generating function prototypes...")),
		preprocessSketchCommand(ctx),

		logIfVerbose(false, tr("Compiling sketch...")),

		types.BareCommand(func(ctx *types.Context) error {
			return recipeByPrefixSuffixRunner(ctx, "recipe.hooks.sketch.prebuild", ".pattern", false)
		}),

		types.BareCommand(func(ctx *types.Context) error {
			sketchObjectFiles, err := ctx.Builder.BuildSketch(
				ctx.SketchLibrariesDetector.IncludeFolders(),
				ctx.OnlyUpdateCompilationDatabase,
				ctx.CompilationDatabase,
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
			return ctx.Builder.RemoveUnusedCompiledLibraries(
				ctx.SketchLibrariesDetector.ImportedLibraries(),
			)
		}),

		types.BareCommand(func(ctx *types.Context) error {
			librariesObjectFiles, err := ctx.Builder.BuildLibraries(
				ctx.SketchLibrariesDetector.IncludeFolders(),
				ctx.SketchLibrariesDetector.ImportedLibraries(),
				ctx.OnlyUpdateCompilationDatabase,
				ctx.CompilationDatabase,
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
			objectFiles, archiveFile, err := ctx.Builder.BuildCore(
				ctx.ActualPlatform,
				ctx.OnlyUpdateCompilationDatabase,
				ctx.CompilationDatabase,
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
			return ctx.Builder.Link(
				ctx.OnlyUpdateCompilationDatabase,
				ctx.SketchObjectFiles,
				ctx.LibrariesObjectFiles,
				ctx.CoreObjectsFiles,
				ctx.CoreArchiveFilePath,
			)
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
			return ctx.Builder.MergeSketchWithBootloader(ctx.OnlyUpdateCompilationDatabase)
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
			ctx.Builder.PrintUsedLibraries(ctx.SketchLibrariesDetector.ImportedLibraries())
			return nil
		}),

		types.BareCommand(func(ctx *types.Context) error {
			return ctx.Builder.ExportProjectCMake(
				mainErr != nil,
				ctx.SketchLibrariesDetector.ImportedLibraries(),
				ctx.SketchLibrariesDetector.IncludeFolders(),
				ctx.LineOffset,
				ctx.OnlyUpdateCompilationDatabase,
			)
		}),

		types.BareCommand(func(ctx *types.Context) error {
			executableSectionsSize, err := sizer.Size(
				ctx.OnlyUpdateCompilationDatabase, mainErr != nil,
				ctx.Builder.GetBuildProperties(),
				ctx.BuilderLogger,
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

func preprocessSketchCommand(ctx *types.Context) types.BareCommand {
	return func(ctx *types.Context) error {
		return ctx.Builder.PreprocessSketch(ctx.SketchLibrariesDetector.IncludeFolders(), ctx.LineOffset, ctx.OnlyUpdateCompilationDatabase)
	}
}

type Preprocess struct{}

func (s *Preprocess) Run(ctx *types.Context) error {
	if err := ctx.Builder.GetBuildPath().MkdirAll(); err != nil {
		return err
	}

	var _err error
	commands := []types.Command{
		containerBuildOptions(ctx),

		types.BareCommand(func(ctx *types.Context) error {
			return recipeByPrefixSuffixRunner(ctx, "recipe.hooks.prebuild", ".pattern", false)
		}),

		types.BareCommand(func(ctx *types.Context) error {
			ctx.LineOffset, _err = ctx.Builder.PrepareSketchBuildPath()
			return _err
		}),

		findIncludes(ctx),

		warnAboutArchIncompatibleLibraries(ctx),

		preprocessSketchCommand(ctx),
	}

	if err := runCommands(ctx, commands); err != nil {
		return err
	}

	// Output arduino-preprocessed source
	preprocessedSketch, err := ctx.Builder.GetSketchBuildPath().Join(ctx.Builder.Sketch().MainFile.Base() + ".cpp").ReadFile()
	if err != nil {
		return err
	}
	ctx.BuilderLogger.WriteStdout(preprocessedSketch)
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
			ctx.Builder.GetBuildPath(),
			ctx.Builder.GetBuildProperties().GetPath("build.core.path"),
			ctx.Builder.GetBuildProperties().GetPath("build.variant.path"),
			ctx.Builder.GetSketchBuildPath(),
			ctx.Builder.Sketch(),
			ctx.Builder.GetLibrariesBuildPath(),
			ctx.Builder.GetBuildProperties(),
			ctx.TargetPlatform.Platform.Architecture,
		)
	})
}

func logIfVerbose(warn bool, msg string) types.BareCommand {
	return types.BareCommand(func(ctx *types.Context) error {
		if !ctx.BuilderLogger.Verbose() {
			return nil
		}
		if warn {
			ctx.BuilderLogger.Warn(msg)
		} else {
			ctx.BuilderLogger.Info(msg)
		}
		return nil
	})
}

func recipeByPrefixSuffixRunner(ctx *types.Context, prefix, suffix string, skipIfOnlyUpdatingCompilationDatabase bool) error {
	return ctx.Builder.RunRecipe(
		prefix, suffix, skipIfOnlyUpdatingCompilationDatabase,
		ctx.OnlyUpdateCompilationDatabase,
	)
}

func containerBuildOptions(ctx *types.Context) types.BareCommand {
	return types.BareCommand(func(ctx *types.Context) error {
		return ctx.Builder.BuildOptionsManager.WipeBuildPath()
	})
}

func warnAboutArchIncompatibleLibraries(ctx *types.Context) types.BareCommand {
	return types.BareCommand(func(ctx *types.Context) error {
		ctx.Builder.WarnAboutArchIncompatibleLibraries(
			ctx.TargetPlatform,
			ctx.SketchLibrariesDetector.ImportedLibraries(),
		)
		return nil
	})
}
