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

package profiles_test

import (
	"testing"

	"github.com/arduino/arduino-cli/internal/integrationtest"
	"github.com/arduino/go-paths-helper"
	"github.com/stretchr/testify/require"
	"go.bug.st/testifyjson/requirejson"
)

func TestCompileWithProfiles(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	// Init the environment explicitly
	_, _, err := cli.Run("core", "update-index")
	require.NoError(t, err)

	// copy sketch_with_profile into the working directory
	sketchPath := cli.CopySketch("sketch_with_profile")

	// use profile without a required library -> should fail
	_, _, err = cli.Run("lib", "install", "Arduino_JSON")
	require.NoError(t, err)
	_, _, err = cli.Run("compile", "-m", "avr1", sketchPath.String())
	require.Error(t, err)

	// use profile with the required library -> should succeed
	_, _, err = cli.Run("compile", "-m", "avr2", sketchPath.String())
	require.NoError(t, err)
}

func TestBuilderDidNotCatchLibsFromUnusedPlatforms(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	// Init the environment explicitly
	_, _, err := cli.Run("core", "update-index")
	require.NoError(t, err)

	// copy sketch into the working directory
	sketchPath := cli.CopySketch("sketch_with_error_including_wire")

	// install two platforms with the Wire library bundled
	_, _, err = cli.Run("core", "install", "arduino:avr")
	require.NoError(t, err)
	_, _, err = cli.Run("core", "install", "arduino:samd")
	require.NoError(t, err)

	// compile for AVR
	stdout, stderr, err := cli.Run("compile", "-b", "arduino:avr:uno", sketchPath.String())
	require.Error(t, err)

	// check that the library resolver did not take the SAMD bundled Wire library into account
	require.NotContains(t, string(stdout), "samd")
	require.NotContains(t, string(stderr), "samd")
}

func TestCompileWithDefaultProfile(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	// Init the environment explicitly
	_, _, err := cli.Run("core", "update-index")
	require.NoError(t, err)

	// Installa core/libs globally
	_, _, err = cli.Run("core", "install", "arduino:avr@1.8.3")
	require.NoError(t, err)

	// copy sketch_with_profile into the working directory
	sketchWithoutDefProfilePath := cli.CopySketch("sketch_without_default_profile")
	sketchWithDefProfilePath := cli.CopySketch("sketch_with_default_profile")

	{
		// no default profile -> error missing FQBN
		_, _, err := cli.Run("compile", sketchWithoutDefProfilePath.String(), "--json")
		require.Error(t, err)
	}
	{
		// specified fbqn -> compile with specified FQBN and use global installation
		stdout, _, err := cli.Run("compile", "-b", "arduino:avr:nano", sketchWithoutDefProfilePath.String(), "--json")
		require.NoError(t, err)
		jsonOut := requirejson.Parse(t, stdout)
		jsonOut.Query(".builder_result.build_platform").MustContain(`{"id":"arduino:avr", "version":"1.8.3"}`)
		jsonOut.Query(".builder_result.build_properties").MustContain(`[ "build.fqbn=arduino:avr:nano" ]`)
	}
	{
		// specified profile -> use the specified profile
		stdout, _, err := cli.Run("compile", "--profile", "avr1", sketchWithoutDefProfilePath.String(), "--json")
		require.NoError(t, err)
		jsonOut := requirejson.Parse(t, stdout)
		jsonOut.Query(".builder_result.build_platform").MustContain(`{"id":"arduino:avr", "version":"1.8.4"}`)
		jsonOut.Query(".builder_result.build_properties").MustContain(`[ "build.fqbn=arduino:avr:uno" ]`)
	}
	{
		// specified profile and fqbn -> use the specified profile and override fqbn
		stdout, _, err := cli.Run("compile", "--profile", "avr1", "-b", "arduino:avr:nano", sketchWithoutDefProfilePath.String(), "--json")
		require.NoError(t, err)
		jsonOut := requirejson.Parse(t, stdout)
		jsonOut.Query(".builder_result.build_platform").MustContain(`{"id":"arduino:avr", "version":"1.8.4"}`)
		jsonOut.Query(".builder_result.build_properties").MustContain(`[ "build.fqbn=arduino:avr:nano" ]`)
	}

	{
		// default profile -> use default profile
		stdout, _, err := cli.Run("compile", sketchWithDefProfilePath.String(), "--json")
		require.NoError(t, err)
		jsonOut := requirejson.Parse(t, stdout)
		jsonOut.Query(".builder_result.build_platform").MustContain(`{"id":"arduino:avr", "version":"1.8.5"}`)
		jsonOut.Query(".builder_result.build_properties").MustContain(`[ "build.fqbn=arduino:avr:leonardo" ]`)
	}
	{
		// default profile, specified fbqn -> use default profile, override fqbn
		stdout, _, err := cli.Run("compile", "-b", "arduino:avr:nano", sketchWithDefProfilePath.String(), "--json")
		require.NoError(t, err)
		jsonOut := requirejson.Parse(t, stdout)
		jsonOut.Query(".builder_result.build_platform").MustContain(`{"id":"arduino:avr", "version":"1.8.5"}`)
		jsonOut.Query(".builder_result.build_properties").MustContain(`[ "build.fqbn=arduino:avr:nano" ]`)
	}
	{
		// default profile, specified different profile -> use the specified profile
		stdout, _, err := cli.Run("compile", "--profile", "avr1", sketchWithDefProfilePath.String(), "--json")
		require.NoError(t, err)
		jsonOut := requirejson.Parse(t, stdout)
		jsonOut.Query(".builder_result.build_platform").MustContain(`{"id":"arduino:avr", "version":"1.8.4"}`)
		jsonOut.Query(".builder_result.build_properties").MustContain(`[ "build.fqbn=arduino:avr:uno" ]`)
	}
	{
		// default profile, specified different profile and fqbn -> use the specified profile and override fqbn
		stdout, _, err := cli.Run("compile", "--profile", "avr1", "-b", "arduino:avr:nano", sketchWithDefProfilePath.String(), "--json")
		require.NoError(t, err)
		jsonOut := requirejson.Parse(t, stdout)
		jsonOut.Query(".builder_result.build_platform").MustContain(`{"id":"arduino:avr", "version":"1.8.4"}`)
		jsonOut.Query(".builder_result.build_properties").MustContain(`[ "build.fqbn=arduino:avr:nano" ]`)
	}
}

func createTempSketch(t *testing.T, cli *integrationtest.ArduinoCLI, sketchName string) *paths.Path {
	sketchDir := cli.SketchbookDir().Join(sketchName)
	t.Cleanup(func() { require.NoError(t, sketchDir.RemoveAll()) })
	_, _, err := cli.Run("sketch", "new", sketchDir.String())
	require.NoError(t, err)
	return sketchDir
}

func TestProfileCreate(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	// Init the environment explicitly
	_, _, err := cli.Run("core", "update-index")
	require.NoError(t, err)
	_, _, err = cli.Run("core", "install", "arduino:avr@1.8.6")
	require.NoError(t, err)

	t.Run("WithInvalidSketchDir", func(t *testing.T) {
		invalidSketchDir := cli.SketchbookDir().Join("tempSketch")

		_, stderr, err := cli.Run("profile", "create", invalidSketchDir.String(), "-m", "test", "-b", "arduino:avr:uno")
		require.Error(t, err)
		require.Contains(t, string(stderr), "no such file or directory")

		require.NoError(t, invalidSketchDir.MkdirAll())
		t.Cleanup(func() { require.NoError(t, invalidSketchDir.RemoveAll()) })

		_, stderr, err = cli.Run("profile", "create", invalidSketchDir.String(), "-m", "test", "-b", "arduino:avr:uno")
		require.Error(t, err)
		require.Contains(t, string(stderr), "main file missing from sketch")
	})

	t.Run("WithNotInstalledPlatform", func(t *testing.T) {
		sketchDir := createTempSketch(t, cli, "TestSketch")
		_, stderr, err := cli.Run("profile", "create", sketchDir.String(), "-m", "uno", "-b", "arduino:samd:zero")
		require.Error(t, err)
		require.Contains(t, string(stderr), "platform not installed")
	})

	t.Run("WithoutSketchYAML", func(t *testing.T) {
		sketchDir := createTempSketch(t, cli, "TestSketch")
		projectFile := sketchDir.Join("sketch.yaml")

		stdout, _, err := cli.Run("profile", "create", sketchDir.String(), "-m", "test", "-b", "arduino:avr:uno")
		require.NoError(t, err)
		require.Contains(t, string(stdout), "Project file created in: "+projectFile.String())
		require.FileExists(t, projectFile.String())
		fileContent, err := projectFile.ReadFile()
		require.NoError(t, err)
		require.Contains(t, string(fileContent), "profiles:\n  test:\n")

		t.Run("AddNewProfile", func(t *testing.T) {
			// Add a new profile
			_, _, err := cli.Run("profile", "create", sketchDir.String(), "-m", "uno", "-b", "arduino:avr:uno")
			require.NoError(t, err)
			fileContent, err := projectFile.ReadFile()
			require.NoError(t, err)
			require.Contains(t, string(fileContent), "  uno:\n    fqbn: arduino:avr:uno\n    platforms:\n      - platform: arduino:avr (1.8.6)\n")
			require.NotContains(t, string(fileContent), "default_profile: uno")
		})

		t.Run("AddAndSetDefaultProfile", func(t *testing.T) {
			// Add a new profile and set it as default
			_, _, err := cli.Run("profile", "create", sketchDir.String(), "-m", "leonardo", "-b", "arduino:avr:leonardo", "--set-default")
			require.NoError(t, err)
			fileContent, err := projectFile.ReadFile()
			require.NoError(t, err)
			require.Contains(t, string(fileContent), "  leonardo:\n    fqbn: arduino:avr:leonardo\n    platforms:\n      - platform: arduino:avr (1.8.6)\n")
			require.Contains(t, string(fileContent), "default_profile: leonardo")
		})

		t.Run("WrongFQBN", func(t *testing.T) {
			// Adding a profile with an incorrect FQBN should return an error
			_, stderr, err := cli.Run("profile", "create", sketchDir.String(), "-m", "wrong_fqbn", "-b", "foo:bar")
			require.Error(t, err)
			require.Contains(t, string(stderr), "Invalid FQBN")
		})

		t.Run("MissingFQBN", func(t *testing.T) {
			// Add a profile with no FQBN should return an error
			_, _, err := cli.Run("profile", "create", sketchDir.String(), "-m", "Uno")
			require.Error(t, err)
		})

		t.Run("AlreadyExistingProfile", func(t *testing.T) {
			// Adding a profile with a name that already exists should return an error
			_, stderr, err := cli.Run("profile", "create", sketchDir.String(), "-m", "uno", "-b", "arduino:avr:uno")
			require.Error(t, err)
			require.Contains(t, string(stderr), "Profile 'uno' already exists")
		})
	})
}

func TestProfileLib(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	// Init the environment explicitly
	_, _, err := cli.Run("core", "update-index")
	require.NoError(t, err)
	_, _, err = cli.Run("core", "install", "arduino:avr")
	require.NoError(t, err)

	t.Run("AddLibToDefaultProfile", func(t *testing.T) {
		sk := createTempSketch(t, cli, "AddLibSketch")

		_, _, err = cli.Run("profile", "create", sk.String(), "-m", "uno", "-b", "arduino:avr:uno", "--set-default")
		require.NoError(t, err)

		out, _, err := cli.Run("profile", "lib", "add", "Arduino_Modulino@0.7.0", "--sketch-path", sk.String(), "--json")
		require.NoError(t, err)
		requirejson.Parse(t, out).Query(".added_libraries").MustContain(`
			[
				{	"kind": "index",
					"library": {"name": "Arduino_Modulino",	"version": "0.7.0" }
				}
			]`)

		fileContent, err := sk.Join("sketch.yaml").ReadFile()
		require.NoError(t, err)
		require.Contains(t, string(fileContent), "      - Arduino_Modulino (0.7.0)\n")
		// dependency added as well
		require.Contains(t, string(fileContent), "      - dependency: Arduino_LSM6DSOX (")

		t.Run("ChangeLibVersionToDefaultProfile", func(t *testing.T) {
			out, _, err := cli.Run("profile", "lib", "add", "Arduino_Modulino@0.6.0", "--sketch-path", sk.String(), "--json")
			require.NoError(t, err)
			outjson := requirejson.Parse(t, out)
			outjson.Query(".added_libraries").MustContain(`
			[
				{	"kind": "index",
					"library": {"name": "Arduino_Modulino",	"version": "0.6.0"}
				}
			]`)
			outjson.Query(".skipped_libraries").MustContain(`
			[
				{	"kind": "index",
					"library": {"name":"Arduino_LSM6DSOX"}
				}
			]`)

			fileContent, err := sk.Join("sketch.yaml").ReadFile()
			require.NoError(t, err)
			require.Contains(t, string(fileContent), "      - Arduino_Modulino (0.6.0)\n")
		})

		t.Run("RemoveLibFromDefaultProfile", func(t *testing.T) {
			_, _, err = cli.Run("profile", "lib", "remove", "Arduino_Modulino", "--sketch-path", sk.String())
			require.NoError(t, err)
			fileContent, err := sk.Join("sketch.yaml").ReadFile()
			require.NoError(t, err)
			require.NotContains(t, string(fileContent), "Arduino_Modulino")
			// dependency removed as well
			require.NotContains(t, string(fileContent), "Arduino_LSM6DSOX")
		})

		t.Run("AddInexistentLibToDefaultProfile", func(t *testing.T) {
			_, stderr, err := cli.Run("profile", "lib", "add", "foobar123", "--sketch-path", sk.String())
			require.Error(t, err)
			require.Equal(t, "Error adding foobar123: Library 'foobar123@latest' not found\n", string(stderr))
		})

		t.Run("RemoveLibNotInProfile", func(t *testing.T) {
			_, stderr, err := cli.Run("profile", "lib", "remove", "Arduino_JSON", "--sketch-path", sk.String())
			require.Error(t, err)
			require.Equal(t, "Error removing library Arduino_JSON from the profile: could not remove library: Library 'Arduino_JSON' not found\n", string(stderr))
		})
	})

}

func TestProfileLibAddRemoveFromSpecificProfile(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	// Init the environment explicitly
	_, _, err := cli.Run("core", "update-index")
	require.NoError(t, err)
	_, _, err = cli.Run("core", "install", "arduino:avr")
	require.NoError(t, err)
	sk := createTempSketch(t, cli, "Simple")

	_, _, err = cli.Run("profile", "create", sk.String(), "-m", "uno", "-b", "arduino:avr:uno")
	require.NoError(t, err)
	// Add a second profile
	_, _, err = cli.Run("profile", "create", sk.String(), "-m", "my_profile", "-b", "arduino:avr:uno")
	require.NoError(t, err)

	// Add library to a specific profile
	_, _, err = cli.Run("profile", "lib", "add", "Arduino_Modulino@0.7.0", "-m", "my_profile", "--sketch-path", sk.String(), "--no-deps")
	require.NoError(t, err)
	fileContent, err := sk.Join("sketch.yaml").ReadFile()
	require.NoError(t, err)
	require.Contains(t, string(fileContent), "  my_profile:\n    fqbn: arduino:avr:uno\n    platforms:\n      - platform: arduino:avr (1.8.6)\n    libraries:\n      - Arduino_Modulino (0.7.0)\n")

	// Remove library from a specific profile
	_, _, err = cli.Run("profile", "lib", "remove", "Arduino_Modulino", "-m", "my_profile", "--sketch-path", sk.String())
	require.NoError(t, err)
	fileContent, err = sk.Join("sketch.yaml").ReadFile()
	require.NoError(t, err)
	require.NotContains(t, string(fileContent), "Arduino_Modulino")
}

func TestProfileSetDefault(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	// Init the environment explicitly
	_, _, err := cli.Run("core", "update-index")
	require.NoError(t, err)
	_, _, err = cli.Run("core", "install", "arduino:avr")
	require.NoError(t, err)
	sk := createTempSketch(t, cli, "Simple")

	// Create two profiles and set both as default (the second one should override the first)
	_, _, err = cli.Run("profile", "create", sk.String(), "-m", "my_profile", "-b", "arduino:avr:uno", "--set-default")
	require.NoError(t, err)
	_, _, err = cli.Run("profile", "create", sk.String(), "-m", "uno", "-b", "arduino:avr:uno", "--set-default")
	require.NoError(t, err)

	fileContent, err := sk.Join("sketch.yaml").ReadFileAsLines()
	require.NoError(t, err)
	require.Contains(t, fileContent, "default_profile: uno")
	require.NotContains(t, fileContent, "default_profile: my_profile")

	// Change default profile, and test JSON output
	out, _, err := cli.Run("profile", "set-default", "my_profile", "--sketch-path", sk.String(), "--json")
	require.NoError(t, err)
	fileContent, err = sk.Join("sketch.yaml").ReadFileAsLines()
	require.NoError(t, err)
	require.NotContains(t, fileContent, "default_profile: uno")
	require.Contains(t, fileContent, "default_profile: my_profile")
	requirejson.Parse(t, out).Query(".default_profile").MustEqual(`"my_profile"`)

	// Changing to an inexistent profile returns an error
	_, stderr, err := cli.Run("profile", "set-default", "inexistent_profile", "--sketch-path", sk.String())
	require.Error(t, err)
	require.Equal(t, "Cannot set inexistent_profile as default profile: Profile 'inexistent_profile' not found\n", string(stderr))
}
