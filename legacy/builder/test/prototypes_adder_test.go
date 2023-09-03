// This file is part of arduino-cli.
//
// Copyright 2020 ARDUINO SA (http://www.arduino.cc/)
// Copyright 2015 Matthijs Kooijman
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

package test

import (
	"path/filepath"
	"strings"
	"testing"

	bldr "github.com/arduino/arduino-cli/arduino/builder"
	"github.com/arduino/arduino-cli/arduino/builder/cpp"
	"github.com/arduino/arduino-cli/legacy/builder"
	"github.com/arduino/arduino-cli/legacy/builder/types"
	paths "github.com/arduino/go-paths-helper"
	"github.com/stretchr/testify/require"
)

func loadPreprocessedSketch(t *testing.T, ctx *types.Context) string {
	res, err := ctx.SketchBuildPath.Join(ctx.Sketch.MainFile.Base() + ".cpp").ReadFile()
	NoError(t, err)
	return string(res)
}

func TestPrototypesAdderCharWithEscapedDoubleQuote(t *testing.T) {
	ctx := prepareBuilderTestContext(t, nil, paths.New("CharWithEscapedDoubleQuote", "CharWithEscapedDoubleQuote.ino"), "arduino:avr:leonardo")
	defer cleanUpBuilderTestContext(t, ctx)

	ctx.Verbose = true

	var _err error
	commands := []types.Command{
		&builder.ContainerSetupHardwareToolsLibsSketchAndProps{},
		types.BareCommand(func(ctx *types.Context) error {
			ctx.LineOffset, _err = bldr.PrepareSketchBuildPath(ctx.Sketch, ctx.SourceOverride, ctx.SketchBuildPath)
			return _err
		}),
		&builder.ContainerFindIncludes{},
	}
	for _, command := range commands {
		err := command.Run(ctx)
		NoError(t, err)
	}
	NoError(t, builder.PreprocessSketch(ctx))

	preprocessed := LoadAndInterpolate(t, filepath.Join("CharWithEscapedDoubleQuote", "CharWithEscapedDoubleQuote.preprocessed.txt"), ctx)
	preprocessedSketch := loadPreprocessedSketch(t, ctx)
	require.Equal(t, preprocessed, strings.Replace(preprocessedSketch, "\r\n", "\n", -1))
}

func TestPrototypesAdderIncludeBetweenMultilineComment(t *testing.T) {
	ctx := prepareBuilderTestContext(t, nil, paths.New("IncludeBetweenMultilineComment", "IncludeBetweenMultilineComment.ino"), "arduino:sam:arduino_due_x_dbg")
	defer cleanUpBuilderTestContext(t, ctx)

	ctx.Verbose = true

	var _err error
	commands := []types.Command{
		&builder.ContainerSetupHardwareToolsLibsSketchAndProps{},
		types.BareCommand(func(ctx *types.Context) error {
			ctx.LineOffset, _err = bldr.PrepareSketchBuildPath(ctx.Sketch, ctx.SourceOverride, ctx.SketchBuildPath)
			return _err
		}),
		&builder.ContainerFindIncludes{},
	}
	for _, command := range commands {
		err := command.Run(ctx)
		NoError(t, err)
	}
	NoError(t, builder.PreprocessSketch(ctx))

	preprocessed := LoadAndInterpolate(t, filepath.Join("IncludeBetweenMultilineComment", "IncludeBetweenMultilineComment.preprocessed.txt"), ctx)
	preprocessedSketch := loadPreprocessedSketch(t, ctx)
	require.Equal(t, preprocessed, strings.Replace(preprocessedSketch, "\r\n", "\n", -1))
}

func TestPrototypesAdderLineContinuations(t *testing.T) {
	ctx := prepareBuilderTestContext(t, nil, paths.New("LineContinuations", "LineContinuations.ino"), "arduino:avr:leonardo")
	defer cleanUpBuilderTestContext(t, ctx)

	ctx.Verbose = true

	var _err error
	commands := []types.Command{
		&builder.ContainerSetupHardwareToolsLibsSketchAndProps{},
		types.BareCommand(func(ctx *types.Context) error {
			ctx.LineOffset, _err = bldr.PrepareSketchBuildPath(ctx.Sketch, ctx.SourceOverride, ctx.SketchBuildPath)
			return _err
		}),
		&builder.ContainerFindIncludes{},
	}
	for _, command := range commands {
		err := command.Run(ctx)
		NoError(t, err)
	}
	NoError(t, builder.PreprocessSketch(ctx))

	preprocessed := LoadAndInterpolate(t, filepath.Join("LineContinuations", "LineContinuations.preprocessed.txt"), ctx)
	preprocessedSketch := loadPreprocessedSketch(t, ctx)
	require.Equal(t, preprocessed, strings.Replace(preprocessedSketch, "\r\n", "\n", -1))
}

func TestPrototypesAdderStringWithComment(t *testing.T) {
	ctx := prepareBuilderTestContext(t, nil, paths.New("StringWithComment", "StringWithComment.ino"), "arduino:avr:leonardo")
	defer cleanUpBuilderTestContext(t, ctx)

	ctx.Verbose = true

	var _err error
	commands := []types.Command{
		&builder.ContainerSetupHardwareToolsLibsSketchAndProps{},
		types.BareCommand(func(ctx *types.Context) error {
			ctx.LineOffset, _err = bldr.PrepareSketchBuildPath(ctx.Sketch, ctx.SourceOverride, ctx.SketchBuildPath)
			return _err
		}),
		&builder.ContainerFindIncludes{},
	}
	for _, command := range commands {
		err := command.Run(ctx)
		NoError(t, err)
	}
	NoError(t, builder.PreprocessSketch(ctx))

	preprocessed := LoadAndInterpolate(t, filepath.Join("StringWithComment", "StringWithComment.preprocessed.txt"), ctx)
	preprocessedSketch := loadPreprocessedSketch(t, ctx)
	require.Equal(t, preprocessed, strings.Replace(preprocessedSketch, "\r\n", "\n", -1))
}

func TestPrototypesAdderSketchWithStruct(t *testing.T) {
	ctx := prepareBuilderTestContext(t, nil, paths.New("SketchWithStruct", "SketchWithStruct.ino"), "arduino:avr:leonardo")
	defer cleanUpBuilderTestContext(t, ctx)

	ctx.Verbose = true

	var _err error
	commands := []types.Command{
		&builder.ContainerSetupHardwareToolsLibsSketchAndProps{},
		types.BareCommand(func(ctx *types.Context) error {
			ctx.LineOffset, _err = bldr.PrepareSketchBuildPath(ctx.Sketch, ctx.SourceOverride, ctx.SketchBuildPath)
			return _err
		}),
		&builder.ContainerFindIncludes{},
	}
	for _, command := range commands {
		err := command.Run(ctx)
		NoError(t, err)
	}
	NoError(t, builder.PreprocessSketch(ctx))

	preprocessed := LoadAndInterpolate(t, filepath.Join("SketchWithStruct", "SketchWithStruct.preprocessed.txt"), ctx)
	preprocessedSketch := loadPreprocessedSketch(t, ctx)
	obtained := strings.Replace(preprocessedSketch, "\r\n", "\n", -1)
	// ctags based preprocessing removes the space after "dostuff", but this is still OK
	// TODO: remove this exception when moving to a more powerful parser
	preprocessed = strings.Replace(preprocessed, "void dostuff (A_NEW_TYPE * bar);", "void dostuff(A_NEW_TYPE * bar);", 1)
	obtained = strings.Replace(obtained, "void dostuff (A_NEW_TYPE * bar);", "void dostuff(A_NEW_TYPE * bar);", 1)
	require.Equal(t, preprocessed, obtained)
}

func TestPrototypesAdderSketchNoFunctionsTwoFiles(t *testing.T) {
	sketchLocation := paths.New("sketch_no_functions_two_files", "sketch_no_functions_two_files.ino")
	quotedSketchLocation := cpp.QuoteString(Abs(t, sketchLocation).String())

	ctx := prepareBuilderTestContext(t, nil, paths.New("sketch_no_functions_two_files", "sketch_no_functions_two_files.ino"), "arduino:avr:leonardo")
	defer cleanUpBuilderTestContext(t, ctx)

	ctx.Verbose = true

	var _err error
	commands := []types.Command{
		&builder.ContainerSetupHardwareToolsLibsSketchAndProps{},
		types.BareCommand(func(ctx *types.Context) error {
			ctx.LineOffset, _err = bldr.PrepareSketchBuildPath(ctx.Sketch, ctx.SourceOverride, ctx.SketchBuildPath)
			return _err
		}),
		&builder.ContainerFindIncludes{},
	}
	for _, command := range commands {
		err := command.Run(ctx)
		NoError(t, err)
	}
	mergedSketch := loadPreprocessedSketch(t, ctx)
	NoError(t, builder.PreprocessSketch(ctx))

	preprocessedSketch := loadPreprocessedSketch(t, ctx)
	require.Contains(t, preprocessedSketch, "#include <Arduino.h>\n#line 1 "+quotedSketchLocation+"\n")
	require.Equal(t, mergedSketch, preprocessedSketch) // No prototypes added
}

func TestPrototypesAdderSketchWithSubstringFunctionMember(t *testing.T) {
	sketchLocation := paths.New("sketch_with_class_and_method_substring", "sketch_with_class_and_method_substring.ino")
	quotedSketchLocation := cpp.QuoteString(Abs(t, sketchLocation).String())

	ctx := prepareBuilderTestContext(t, nil, sketchLocation, "arduino:avr:uno")
	defer cleanUpBuilderTestContext(t, ctx)

	ctx.Verbose = true

	var _err error
	commands := []types.Command{
		&builder.ContainerSetupHardwareToolsLibsSketchAndProps{},
		types.BareCommand(func(ctx *types.Context) error {
			ctx.LineOffset, _err = bldr.PrepareSketchBuildPath(ctx.Sketch, ctx.SourceOverride, ctx.SketchBuildPath)
			return _err
		}),
		&builder.ContainerFindIncludes{},
	}
	for _, command := range commands {
		err := command.Run(ctx)
		NoError(t, err)
	}
	NoError(t, builder.PreprocessSketch(ctx))

	preprocessedSketch := loadPreprocessedSketch(t, ctx)
	require.Contains(t, preprocessedSketch, "class Foo {\nint blooper(int x) { return x+1; }\n};\n\nFoo foo;\n\n#line 7 "+quotedSketchLocation+"\nvoid setup();")
}
