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

	"github.com/arduino/arduino-cli/arduino/builder/preprocessor"
	"github.com/arduino/arduino-cli/legacy/builder"
	"github.com/arduino/arduino-cli/legacy/builder/types"
	paths "github.com/arduino/go-paths-helper"
	"github.com/stretchr/testify/require"
)

// TODO add them in the compile_4

func TestTryBuild040(t *testing.T) {
	tryBuild(t, paths.New("sketch_with_externC_multiline", "sketch_with_externC_multiline.ino"))
}

func TestTryBuild041(t *testing.T) {
	tryBuild(t, paths.New("sketch_with_multiline_template", "sketch_with_multiline_template.ino"))
}

func TestTryBuild042(t *testing.T) {
	tryBuild(t, paths.New("sketch_with_fake_function_pointer", "sketch_with_fake_function_pointer.ino"))
}

func TestTryBuild036(t *testing.T) {
	ctx := makeDefaultContext()
	tryBuildWithContext(t, ctx, "arduino:samd:arduino_zero_native", paths.New("sketch_fastleds", "sketch_fastleds.ino"))
}

func TestTryBuild039(t *testing.T) {
	ctx := makeDefaultContext()
	tryBuildWithContext(t, ctx, "arduino:samd:arduino_zero_native", paths.New("sketch12", "sketch12.ino"))
}

func makeDefaultContext() *types.Context {
	preprocessor.DebugPreprocessor = true
	return &types.Context{
		HardwareDirs:         paths.NewPathList(filepath.Join("..", "hardware"), "downloaded_hardware", "downloaded_board_manager_stuff"),
		BuiltInToolsDirs:     paths.NewPathList("downloaded_tools"),
		BuiltInLibrariesDirs: paths.New("downloaded_libraries"),
		OtherLibrariesDirs:   paths.NewPathList("libraries"),
	}
}

func tryBuild(t *testing.T, sketchLocation *paths.Path) {
	tryBuildWithContext(t, makeDefaultContext(), "arduino:avr:leonardo", sketchLocation)
}

func tryBuildWithContext(t *testing.T, ctx *types.Context, fqbn string, sketchLocation *paths.Path) {
	ctx = prepareBuilderTestContext(t, ctx, sketchLocation, fqbn)
	defer cleanUpBuilderTestContext(t, ctx)

	err := builder.RunBuilder(ctx)
	require.NoError(t, err, "Build error for "+sketchLocation.String())
}
