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
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
)

func TestCompileWithFullyPrecompiledLibrary(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	_, _, err := cli.Run("update")
	require.NoError(t, err)

	_, _, err = cli.Run("core", "install", "arduino:mbed@1.3.1")
	require.NoError(t, err)
	fqbn := "arduino:mbed:nano33ble"

	// Create settings with library unsafe install set to true
	envVar := cli.GetDefaultEnv()
	envVar["ARDUINO_LIBRARY_ENABLE_UNSAFE_INSTALL"] = "true"
	_, _, err = cli.RunWithCustomEnv(envVar, "config", "init", "--dest-dir", ".")
	require.NoError(t, err)

	// Install fully precompiled library
	// For more information see:
	// https://arduino.github.io/arduino-cli/latest/library-specification/#precompiled-binaries
	wd, err := paths.Getwd()
	require.NoError(t, err)
	_, _, err = cli.Run("lib", "install", "--zip-path", wd.Parent().Join("testdata", "Arduino_TensorFlowLite-2.1.0-ALPHA-precompiled.zip").String())
	require.NoError(t, err)
	sketchFolder := cli.SketchbookDir().Join("libraries", "Arduino_TensorFlowLite-2.1.0-ALPHA-precompiled", "examples", "hello_world")

	// Install example dependency
	_, _, err = cli.Run("lib", "install", "Arduino_LSM9DS1")
	require.NoError(t, err)

	// Compile and verify dependencies detection for fully precompiled library is skipped
	stdout, _, err := cli.Run("compile", "-b", fqbn, sketchFolder.String(), "-v")
	require.NoError(t, err)
	require.Contains(t, string(stdout), "Skipping dependencies detection for precompiled library Arduino_TensorFlowLite")
}

func TestCompileManuallyInstalledPlatform(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	_, _, err := cli.Run("update")
	require.NoError(t, err)

	sketchName := "CompileSketchManuallyInstalledPlatformUsingPlatformLocalTxt"
	sketchPath := cli.SketchbookDir().Join(sketchName)
	fqbn := "arduino-beta-development:avr:uno"
	_, _, err = cli.Run("sketch", "new", sketchPath.String())
	require.NoError(t, err)

	// Manually installs a core in sketchbooks hardware folder
	gitUrl := "https://github.com/arduino/ArduinoCore-avr.git"
	repoDir := cli.SketchbookDir().Join("hardware", "arduino-beta-development", "avr")
	_, err = git.PlainClone(repoDir.String(), false, &git.CloneOptions{
		URL:           gitUrl,
		ReferenceName: plumbing.NewTagReferenceName("1.8.3"),
	})
	require.NoError(t, err)

	// Installs also the same core via CLI so all the necessary tools are installed
	_, _, err = cli.Run("core", "install", "arduino:avr@1.8.3")
	require.NoError(t, err)

	// Verifies compilation works without issues
	_, _, err = cli.Run("compile", "--clean", "-b", fqbn, sketchPath.String())
	require.NoError(t, err)
}

func TestCompileManuallyInstalledPlatformUsingPlatformLocalTxt(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	_, _, err := cli.Run("update")
	require.NoError(t, err)

	sketchName := "CompileSketchManuallyInstalledPlatformUsingPlatformLocalTxt"
	sketchPath := cli.SketchbookDir().Join(sketchName)
	fqbn := "arduino-beta-development:avr:uno"
	_, _, err = cli.Run("sketch", "new", sketchPath.String())
	require.NoError(t, err)

	// Manually installs a core in sketchbooks hardware folder
	gitUrl := "https://github.com/arduino/ArduinoCore-avr.git"
	repoDir := cli.SketchbookDir().Join("hardware", "arduino-beta-development", "avr")
	_, err = git.PlainClone(repoDir.String(), false, &git.CloneOptions{
		URL:           gitUrl,
		ReferenceName: plumbing.NewTagReferenceName("1.8.3"),
	})
	require.NoError(t, err)

	// Installs also the same core via CLI so all the necessary tools are installed
	_, _, err = cli.Run("core", "install", "arduino:avr@1.8.3")
	require.NoError(t, err)

	// Verifies compilation works without issues
	_, _, err = cli.Run("compile", "--clean", "-b", fqbn, sketchPath.String())
	require.NoError(t, err)

	// Overrides default platform compiler with an unexisting one
	platformLocalTxt := repoDir.Join("platform.local.txt")
	platformLocalTxt.WriteFile([]byte("compiler.c.cmd=my-compiler-that-does-not-exist"))

	// Verifies compilation now fails because compiler is not found
	_, stderr, err := cli.Run("compile", "--clean", "-b", fqbn, sketchPath.String())
	require.Error(t, err)
	require.Contains(t, string(stderr), "my-compiler-that-does-not-exist")
}
