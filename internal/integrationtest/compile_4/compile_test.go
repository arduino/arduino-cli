// This file is part of arduino-cli.
//
// Copyright 2023 ARDUINO SA (http://www.arduino.cc/)
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

package compile_test

import (
	"path/filepath"
	"testing"

	"github.com/arduino/arduino-cli/internal/integrationtest"
	"github.com/arduino/go-paths-helper"
	"github.com/stretchr/testify/require"
)

func TestCompileOfProblematicSketches(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	// Install Arduino AVR Boards
	_, _, err := cli.Run("core", "install", "arduino:avr@1.8.6")
	require.NoError(t, err)

	// Install Libraries required for tests
	_, _, err = cli.Run("lib", "install", "Bridge@1.6.1")
	require.NoError(t, err)

	integrationtest.CLISubtests{
		{"SketchWithInlineFunction", tryBuildAvrLeonardo},
		{"SketchWithConst", tryBuildAvrLeonardo},
		{"SketchWithFunctionSignatureInsideIfdef", tryBuildAvrLeonardo},
		{"SketchWithOldLibrary", tryBuildAvrLeonardo},
		{"SketchWithoutFunctions", tryPreprocessAvrLeonardo},
		{"SketchWithConfig", tryBuildAvrLeonardo},
		{"SketchWithUsbcon", tryBuildAvrLeonardo},
		//{"SketchWithTypename", tryBuildAvrLeonardo}, // XXX: Failing sketch, typename not supported
		{"SketchWithMacosxGarbage", tryBuildAvrLeonardo},
		{"SketchWithNamespace", tryBuildAvrLeonardo},
		{"SketchWithDefaultArgs", tryBuildAvrLeonardo},
		{"SketchWithClass", tryBuildAvrLeonardo},
		{"SketchWithBackupFiles", tryBuildAvrLeonardo},
		{"SketchWithSubfolders", tryBuildAvrLeonardo},
	}.Run(t, env, cli)
}

func tryBuildAvrLeonardo(t *testing.T, env *integrationtest.Environment, cli *integrationtest.ArduinoCLI) {
	subTestName := filepath.Base(t.Name())
	sketchPath, err := paths.New("testdata", subTestName).Abs()
	require.NoError(t, err)
	libsPath, err := paths.New("testdata", "libraries").Abs()
	require.NoError(t, err)
	_, _, err = cli.Run("compile", "-b", "arduino:avr:leonardo", "--libraries", libsPath.String(), sketchPath.String())
	require.NoError(t, err)
}

func tryPreprocessAvrLeonardo(t *testing.T, env *integrationtest.Environment, cli *integrationtest.ArduinoCLI) {
	subTestName := filepath.Base(t.Name())
	sketchPath, err := paths.New("testdata", subTestName).Abs()
	require.NoError(t, err)
	_, _, err = cli.Run("compile", "-b", "arduino:avr:leonardo", "--preprocess", sketchPath.String())
	require.NoError(t, err)
}
