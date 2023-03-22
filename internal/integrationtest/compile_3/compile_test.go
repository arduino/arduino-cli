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
	"github.com/arduino/go-properties-orderedmap"
	"github.com/stretchr/testify/require"
	"go.bug.st/testifyjson/requirejson"
)

func TestRuntimeToolPropertiesGeneration(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	// Run update-index with our test index
	_, _, err := cli.Run("core", "install", "arduino:avr@1.8.5")
	require.NoError(t, err)

	// Install test data into datadir
	testdata := paths.New("testdata", "platforms_with_conflicting_tools")
	hardwareDir := cli.DataDir().Join("packages")
	err = testdata.Join("alice").CopyDirTo(hardwareDir.Join("alice"))
	require.NoError(t, err)
	err = testdata.Join("bob").CopyDirTo(hardwareDir.Join("bob"))
	require.NoError(t, err)

	sketch, err := paths.New("testdata", "bare_minimum").Abs()
	require.NoError(t, err)

	// As seen in https://github.com/arduino/arduino-cli/issues/73 the map randomess
	// may make the function fail half of the times. Repeating the test 3 times
	// greatly increases the chances to trigger the bad case.
	for i := 0; i < 3; i++ {
		stdout, _, err := cli.Run("compile", "-b", "alice:avr:alice", "--show-properties", sketch.String())
		require.NoError(t, err)
		res, err := properties.LoadFromBytes(stdout)
		require.NoError(t, err)
		// the tools coming from the same packager are selected
		require.True(t, res.GetPath("runtime.tools.avr-gcc.path").EquivalentTo(hardwareDir.Join("alice", "tools", "avr-gcc", "50.0.0")))
		require.True(t, res.GetPath("runtime.tools.avrdude.path").EquivalentTo(hardwareDir.Join("alice", "tools", "avrdude", "1.0.0")))

		stdout, _, err = cli.Run("compile", "-b", "bob:avr:bob", "--show-properties", sketch.String())
		require.NoError(t, err)
		res, err = properties.LoadFromBytes(stdout)
		require.NoError(t, err)
		// the latest version available are selected
		require.True(t, res.GetPath("runtime.tools.avr-gcc.path").EquivalentTo(hardwareDir.Join("alice", "tools", "avr-gcc", "50.0.0")))
		require.True(t, res.GetPath("runtime.tools.avrdude.path").EquivalentTo(hardwareDir.Join("arduino", "tools", "avrdude", "6.3.0-arduino17")))

		stdout, _, err = cli.Run("compile", "-b", "arduino:avr:uno", "--show-properties", sketch.String())
		require.NoError(t, err)
		res, err = properties.LoadFromBytes(stdout)
		require.NoError(t, err)
		// the selected tools are listed as platform dependencies from the index.json
		require.True(t, res.GetPath("runtime.tools.avr-gcc.path").EquivalentTo(hardwareDir.Join("arduino", "tools", "avr-gcc", "7.3.0-atmel3.6.1-arduino7")))
		require.True(t, res.GetPath("runtime.tools.avrdude.path").EquivalentTo(hardwareDir.Join("arduino", "tools", "avrdude", "6.3.0-arduino17")))
	}
}

func TestCompileBuildPathInsideSketch(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	_, _, err := cli.Run("core", "update-index")
	require.NoError(t, err)

	_, _, err = cli.Run("core", "install", "arduino:avr")
	require.NoError(t, err)

	sketch := "sketchSimple"
	_, _, err = cli.Run("sketch", "new", sketch)
	require.NoError(t, err)

	cli.SetWorkingDir(cli.WorkingDir().Join(sketch))
	// Compile the sketch creating the build directory inside the sketch directory
	_, _, err = cli.Run("compile", "-b", "arduino:avr:mega", "--build-path", "build-mega")
	require.NoError(t, err)

	// Compile again using the same build path
	_, _, err = cli.Run("compile", "-b", "arduino:avr:mega", "--build-path", "build-mega")
	require.NoError(t, err)
}

func TestCompilerErrOutput(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	// Run update-index with our test index
	_, _, err := cli.Run("core", "install", "arduino:avr@1.8.5")
	require.NoError(t, err)

	// prepare sketch
	sketch, err := paths.New("testdata", "blink_with_wrong_cpp").Abs()
	require.NoError(t, err)

	// Run compile and catch err stream
	out, _, err := cli.Run("compile", "-b", "arduino:avr:uno", "--format", "json", sketch.String())
	require.Error(t, err)
	compilerErr := requirejson.Parse(t, out).Query(".compiler_err")
	compilerErr.MustContain(`"error"`)
}

func TestCompileRelativeLibraryPath(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	// Initialize configs to enable --zip-path flag
	_, _, err := cli.Run("config", "init", "--dest-dir", ".")
	require.NoError(t, err)
	_, _, err = cli.Run("config", "set", "library.enable_unsafe_install", "true", "--config-file", "arduino-cli.yaml")
	require.NoError(t, err)
	configFile := cli.WorkingDir().Join("arduino-cli.yaml")

	_, _, err = cli.Run("core", "install", "arduino:avr")
	require.NoError(t, err)

	// Install library and its dependencies
	zipPath, err := paths.New("..", "testdata", "FooLib.zip").Abs()
	require.NoError(t, err)
	// Manually install the library and move into one of the example's directories
	FooLib := cli.WorkingDir().Join("FooLib")
	err = paths.New("..", "testdata", "FooLib").CopyDirTo(FooLib)
	require.NoError(t, err)
	cli.SetWorkingDir(FooLib.Join("examples", "FooSketch"))

	// Compile using a relative path to the library
	_, _, err = cli.Run("compile", "-b", "arduino:avr:uno", "--library", "../../")
	require.NoError(t, err)

	// Install the same library using lib install and compile again using the relative path.
	// The manually installed library should be chosen
	_, _, err = cli.Run("lib", "install", "--zip-path", zipPath.String(), "--config-file", configFile.String())
	require.NoError(t, err)
	stdout, _, err := cli.Run("compile", "-b", "arduino:avr:uno", "--library", "../../", "-v")
	require.NoError(t, err)
	require.Contains(t, string(stdout), "Multiple libraries were found for \"FooLib.h\"")
	require.Contains(t, string(stdout), "Used: "+FooLib.String())
	require.Contains(t, string(stdout), "Not used: "+cli.SketchbookDir().Join("libraries", "FooLib").String())
}
