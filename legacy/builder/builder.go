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
	"github.com/arduino/arduino-cli/legacy/builder/utils"
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
		&ContainerSetupHardwareToolsLibsSketchAndProps{},

		&ContainerBuildOptions{},

		&RecipeByPrefixSuffixRunner{Prefix: "recipe.hooks.prebuild", Suffix: ".pattern"},

		types.BareCommand(func(ctx *types.Context) error {
			ctx.LineOffset, _err = ctx.Builder.PrepareSketchBuildPath(ctx.SourceOverride, ctx.SketchBuildPath)
			return _err
		}),

		utils.LogIfVerbose(false, tr("Detecting libraries used...")),
		&ContainerFindIncludes{},

		&WarnAboutArchIncompatibleLibraries{},

		utils.LogIfVerbose(false, tr("Generating function prototypes...")),
		types.BareCommand(PreprocessSketch),

		utils.LogIfVerbose(false, tr("Compiling sketch...")),
		&RecipeByPrefixSuffixRunner{Prefix: "recipe.hooks.sketch.prebuild", Suffix: ".pattern"},
		&phases.SketchBuilder{},
		&RecipeByPrefixSuffixRunner{Prefix: "recipe.hooks.sketch.postbuild", Suffix: ".pattern", SkipIfOnlyUpdatingCompilationDatabase: true},

		utils.LogIfVerbose(false, tr("Compiling libraries...")),
		&RecipeByPrefixSuffixRunner{Prefix: "recipe.hooks.libraries.prebuild", Suffix: ".pattern"},
		&UnusedCompiledLibrariesRemover{},
		&phases.LibrariesBuilder{},
		&RecipeByPrefixSuffixRunner{Prefix: "recipe.hooks.libraries.postbuild", Suffix: ".pattern", SkipIfOnlyUpdatingCompilationDatabase: true},

		utils.LogIfVerbose(false, tr("Compiling core...")),
		&RecipeByPrefixSuffixRunner{Prefix: "recipe.hooks.core.prebuild", Suffix: ".pattern"},
		&phases.CoreBuilder{},
		&RecipeByPrefixSuffixRunner{Prefix: "recipe.hooks.core.postbuild", Suffix: ".pattern", SkipIfOnlyUpdatingCompilationDatabase: true},

		utils.LogIfVerbose(false, tr("Linking everything together...")),
		&RecipeByPrefixSuffixRunner{Prefix: "recipe.hooks.linking.prelink", Suffix: ".pattern"},
		&phases.Linker{},
		&RecipeByPrefixSuffixRunner{Prefix: "recipe.hooks.linking.postlink", Suffix: ".pattern", SkipIfOnlyUpdatingCompilationDatabase: true},

		&RecipeByPrefixSuffixRunner{Prefix: "recipe.hooks.objcopy.preobjcopy", Suffix: ".pattern"},
		&RecipeByPrefixSuffixRunner{Prefix: "recipe.objcopy.", Suffix: ".pattern", SkipIfOnlyUpdatingCompilationDatabase: true},
		&RecipeByPrefixSuffixRunner{Prefix: "recipe.hooks.objcopy.postobjcopy", Suffix: ".pattern", SkipIfOnlyUpdatingCompilationDatabase: true},

		&MergeSketchWithBootloader{},

		&RecipeByPrefixSuffixRunner{Prefix: "recipe.hooks.postbuild", Suffix: ".pattern", SkipIfOnlyUpdatingCompilationDatabase: true},
	}

	ctx.Progress.AddSubSteps(len(commands) + 4)
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
		&PrintUsedAndNotUsedLibraries{SketchError: mainErr != nil},

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
	if ctx.UseArduinoPreprocessor {
		preprocessorImpl = preprocessor.PreprocessSketchWithArduinoPreprocessor
	}
	normalOutput, verboseOutput, err := preprocessorImpl(
		ctx.Sketch, ctx.BuildPath, ctx.IncludeFolders, ctx.LineOffset,
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
		&ContainerSetupHardwareToolsLibsSketchAndProps{},

		&ContainerBuildOptions{},

		&RecipeByPrefixSuffixRunner{Prefix: "recipe.hooks.prebuild", Suffix: ".pattern"},

		types.BareCommand(func(ctx *types.Context) error {
			ctx.LineOffset, _err = ctx.Builder.PrepareSketchBuildPath(ctx.SourceOverride, ctx.SketchBuildPath)
			return _err
		}),

		&ContainerFindIncludes{},

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

func RunParseHardware(ctx *types.Context) error {
	commands := []types.Command{
		&ContainerSetupHardwareToolsLibsSketchAndProps{},
	}
	return runCommands(ctx, commands)
}

func RunPreprocess(ctx *types.Context) error {
	command := Preprocess{}
	return command.Run(ctx)
}
