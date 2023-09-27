// This file is part of arduino-cli.
//
// Copyright 2022 ARDUINO SA (http://www.arduino.cc/)
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
	"testing"

	"github.com/arduino/arduino-cli/internal/integrationtest"
	"github.com/arduino/go-paths-helper"
	"github.com/stretchr/testify/require"
)

func TestCompileAndUploadRuntimeProperties(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	// https://github.com/arduino/arduino-cli/issues/1971
	sketchbookHardwareDir := cli.SketchbookDir().Join("hardware")
	require.NoError(t, sketchbookHardwareDir.MkdirAll())

	// Copy test platform
	testPlatform := paths.New("..", "testdata", "foo")
	require.NoError(t, testPlatform.CopyDirTo(sketchbookHardwareDir.Join("foo")))

	// Install dependencies of the demo platform
	_, _, err := cli.Run("core", "update-index")
	require.NoError(t, err)
	_, _, err = cli.Run("core", "install", "arduino:avr")
	require.NoError(t, err)

	// Check compile runtime properties expansion
	bareMinimum := cli.CopySketch("bare_minimum")
	stdout, _, err := cli.Run("compile", "--fqbn", "foo:avr:bar", "-v", bareMinimum.String())
	require.NoError(t, err)
	require.Contains(t, string(stdout), "PREBUILD-runtime.hardware.path="+sketchbookHardwareDir.String())

	// Check upload runtime properties expansion
	stdout, _, err = cli.Run("upload", "--fqbn", "foo:avr:bar", bareMinimum.String())
	require.NoError(t, err)
	require.Contains(t, string(stdout), "UPLOAD-runtime.hardware.path="+sketchbookHardwareDir.String())
}
