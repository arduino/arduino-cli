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
	"testing"

	"github.com/arduino/arduino-cli/legacy/builder"
	"github.com/arduino/arduino-cli/legacy/builder/types"
	paths "github.com/arduino/go-paths-helper"
)

func TestTryBuild001(t *testing.T) {
	tryBuild(t, "sketch_with_inline_function", "sketch_with_inline_function.ino")
}

func TestTryBuild002(t *testing.T) {
	tryBuild(t, "sketch_with_function_signature_inside_ifdef", "sketch_with_function_signature_inside_ifdef.ino")
}

func TestTryBuild003(t *testing.T) {
	tryPreprocess(t, "sketch_no_functions", "sketch_no_functions.ino")
}

func TestTryBuild004(t *testing.T) {
	tryBuild(t, "sketch_with_const", "sketch_with_const.ino")
}

func TestTryBuild005(t *testing.T) {
	tryBuild(t, "sketch_with_old_lib", "sketch_with_old_lib.ino")
}

func TestTryBuild006(t *testing.T) {
	tryBuild(t, "sketch_with_macosx_garbage", "sketch_with_macosx_garbage.ino")
}

func TestTryBuild007(t *testing.T) {
	tryBuild(t, "sketch_with_config", "sketch_with_config.ino")
}

// XXX: Failing sketch, typename not supported
//func TestTryBuild008(t *testing.T) {
//	tryBuild(t, "sketch_with_typename", "sketch.ino")
//}

func TestTryBuild009(t *testing.T) {
	tryBuild(t, "sketch_with_usbcon", "sketch_with_usbcon.ino")
}

func TestTryBuild010(t *testing.T) {
	tryBuild(t, "sketch_with_namespace", "sketch_with_namespace.ino")
}

func TestTryBuild011(t *testing.T) {
	tryBuild(t, "sketch_with_inline_function", "sketch_with_inline_function.ino")
}

func TestTryBuild012(t *testing.T) {
	tryBuild(t, "sketch_with_default_args", "sketch_with_default_args.ino")
}

func TestTryBuild013(t *testing.T) {
	tryBuild(t, "sketch_with_class", "sketch_with_class.ino")
}

func TestTryBuild014(t *testing.T) {
	tryBuild(t, "sketch_with_backup_files", "sketch_with_backup_files.ino")
}

func TestTryBuild015(t *testing.T) {
	tryBuild(t, "sketch_with_subfolders")
}

// This is a sketch that fails to build on purpose
//func TestTryBuild016(t *testing.T) {
//	tryBuild(t, "sketch_that_checks_if_SPI_has_transactions_and_includes_missing_Ethernet", "sketch.ino")
//}

func TestTryBuild017(t *testing.T) {
	tryPreprocess(t, "sketch_no_functions_two_files", "sketch_no_functions_two_files.ino")
}

func TestTryBuild018(t *testing.T) {
	tryBuild(t, "sketch_that_checks_if_SPI_has_transactions", "sketch_that_checks_if_SPI_has_transactions.ino")
}

func TestTryBuild019(t *testing.T) {
	tryBuild(t, "sketch_with_ifdef", "sketch_with_ifdef.ino")
}

func TestTryBuild020(t *testing.T) {
	ctx := makeDefaultContext(t)
	ctx.OtherLibrariesDirs = paths.NewPathList("dependent_libraries", "libraries")
	tryPreprocessWithContext(t, ctx, "sketch_with_dependend_libraries", "sketch_with_dependend_libraries.ino")
}

func TestTryBuild021(t *testing.T) {
	tryBuild(t, "sketch_with_function_pointer", "sketch_with_function_pointer.ino")
}

func TestTryBuild022(t *testing.T) {
	ctx := makeDefaultContext(t)
	ctx.FQBN = parseFQBN(t, "arduino:samd:arduino_zero_native")
	tryBuildWithContext(t, ctx, "sketch_usbhost", "sketch_usbhost.ino")
}

func TestTryBuild023(t *testing.T) {
	tryBuild(t, "sketch1", "sketch1.ino")
}

func TestTryBuild024(t *testing.T) {
	tryBuild(t, "SketchWithIfDef", "SketchWithIfDef.ino")
}

// The library for this sketch is missing
//func TestTryBuild025(t *testing.T) {
//	tryBuild(t, "sketch3", "Baladuino.ino")
//}

func TestTryBuild026(t *testing.T) {
	tryBuild(t, "CharWithEscapedDoubleQuote", "CharWithEscapedDoubleQuote.ino")
}

func TestTryBuild027(t *testing.T) {
	tryBuild(t, "IncludeBetweenMultilineComment", "IncludeBetweenMultilineComment.ino")
}

func TestTryBuild028(t *testing.T) {
	tryBuild(t, "LineContinuations", "LineContinuations.ino")
}

func TestTryBuild029(t *testing.T) {
	tryBuild(t, "StringWithComment", "StringWithComment.ino")
}

func TestTryBuild030(t *testing.T) {
	tryBuild(t, "SketchWithStruct", "SketchWithStruct.ino")
}

func TestTryBuild031(t *testing.T) {
	tryBuild(t, "sketch9", "sketch9.ino")
}

func TestTryBuild032(t *testing.T) {
	tryBuild(t, "sketch10", "sketch10.ino")
}

func TestTryBuild033(t *testing.T) {
	tryBuild(t, "sketch_that_includes_arduino_h", "sketch_that_includes_arduino_h.ino")
}

func TestTryBuild034(t *testing.T) {
	tryBuild(t, "sketch_with_static_asserts", "sketch_with_static_asserts.ino")
}

func TestTryBuild035(t *testing.T) {
	tryBuild(t, "sketch_with_enum_class", "sketch_with_enum_class.ino")
}

func TestTryBuild036(t *testing.T) {
	ctx := makeDefaultContext(t)
	ctx.FQBN = parseFQBN(t, "arduino:samd:arduino_zero_native")
	tryBuildWithContext(t, ctx, "sketch_fastleds", "sketch_fastleds.ino")
}

func TestTryBuild037(t *testing.T) {
	tryBuild(t, "sketch_with_externC", "sketch_with_externC.ino")
}

func TestTryBuild038(t *testing.T) {
	tryBuild(t, "sketch_with_multiline_prototypes", "sketch_with_multiline_prototypes.ino")
}

func TestTryBuild039(t *testing.T) {
	ctx := makeDefaultContext(t)
	ctx.FQBN = parseFQBN(t, "arduino:samd:arduino_zero_native")
	tryBuildWithContext(t, ctx, "sketch12", "sketch12.ino")
}

func TestTryBuild040(t *testing.T) {
	tryBuild(t, "sketch_with_externC_multiline", "sketch_with_externC_multiline.ino")
}

func TestTryBuild041(t *testing.T) {
	tryBuild(t, "sketch_with_multiline_template", "sketch_with_multiline_template.ino")
}

func TestTryBuild042(t *testing.T) {
	tryBuild(t, "sketch_with_fake_function_pointer", "sketch_with_fake_function_pointer.ino")
}

func makeDefaultContext(t *testing.T) *types.Context {
	DownloadCoresAndToolsAndLibraries(t)

	ctx := &types.Context{
		HardwareDirs:         paths.NewPathList(filepath.Join("..", "hardware"), "downloaded_hardware", "downloaded_board_manager_stuff"),
		BuiltInToolsDirs:     paths.NewPathList("downloaded_tools"),
		BuiltInLibrariesDirs: paths.New("downloaded_libraries"),
		OtherLibrariesDirs:   paths.NewPathList("libraries"),
		FQBN:                 parseFQBN(t, "arduino:avr:leonardo"),
		ArduinoAPIVersion:    "10607",
		Verbose:              true,
		DebugPreprocessor:    true,
	}
	buildPath := SetupBuildPath(t, ctx)
	defer buildPath.RemoveAll()

	return ctx
}

func tryBuild(t *testing.T, sketchPath ...string) {
	ctx := makeDefaultContext(t)
	tryBuildWithContext(t, ctx, sketchPath...)
}

func tryBuildWithContext(t *testing.T, ctx *types.Context, sketchPath ...string) {
	sketchLocation := paths.New(sketchPath...)
	ctx.Sketch = OpenSketch(t, sketchLocation)

	err := builder.RunBuilder(ctx)
	NoError(t, err, "Build error for "+sketchLocation.String())
}

func tryPreprocess(t *testing.T, sketchPath ...string) {
	ctx := makeDefaultContext(t)
	tryPreprocessWithContext(t, ctx, sketchPath...)
}

func tryPreprocessWithContext(t *testing.T, ctx *types.Context, sketchPath ...string) {
	sketchLocation := paths.New(sketchPath...)
	ctx.Sketch = OpenSketch(t, sketchLocation)

	err := builder.RunPreprocess(ctx)
	NoError(t, err, "Build error for "+sketchLocation.String())
}
