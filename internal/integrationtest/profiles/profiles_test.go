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

func TestInitProfile(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	// Init the environment explicitly
	_, _, err := cli.Run("core", "update-index")
	require.NoError(t, err)

	_, _, err = cli.Run("sketch", "new", cli.SketchbookDir().Join("Simple").String())
	require.NoError(t, err)

	_, _, err = cli.Run("core", "install", "arduino:avr")
	require.NoError(t, err)

	integrationtest.CLISubtests{
		{"NoProfile", initNoProfile},
		{"ProfileCorrectFQBN", initWithCorrectFqbn},
		{"ProfileWrongFQBN", initWithWrongFqbn},
		{"ProfileMissingFQBN", initMissingFqbn},
		{"ExistingProfile", initExistingProfile},
		{"SetDefaultProfile", initSetDefaultProfile},
	}.Run(t, env, cli)
}

func initNoProfile(t *testing.T, env *integrationtest.Environment, cli *integrationtest.ArduinoCLI) {
	projectFile := cli.SketchbookDir().Join("Simple", "sketch.yaml")
	// Create an empty project file
	stdout, _, err := cli.Run("profile", "init", cli.SketchbookDir().Join("Simple").String())
	require.NoError(t, err)
	require.Contains(t, string(stdout), "Project file created in: "+projectFile.String())
	require.FileExists(t, projectFile.String())
	fileContent, err := projectFile.ReadFile()
	require.NoError(t, err)
	require.Equal(t, "profiles:\n", string(fileContent))
}

func initWithCorrectFqbn(t *testing.T, env *integrationtest.Environment, cli *integrationtest.ArduinoCLI) {
	projectFile := cli.SketchbookDir().Join("Simple", "sketch.yaml")
	// Add a profile with a correct FQBN
	_, _, err := cli.Run("profile", "init", cli.SketchbookDir().Join("Simple").String(), "-m", "Uno", "-b", "arduino:avr:uno")
	require.NoError(t, err)
	require.FileExists(t, projectFile.String())
	fileContent, err := projectFile.ReadFile()
	require.NoError(t, err)
	require.Equal(t, "profiles:\n  Uno:\n    fqbn: arduino:avr:uno\n    platforms:\n      - platform: arduino:avr (1.8.6)\n    libraries:\n\ndefault_profile: Uno\n", string(fileContent))
}

func initWithWrongFqbn(t *testing.T, env *integrationtest.Environment, cli *integrationtest.ArduinoCLI) {
	// Adding a profile with an incorrect FQBN should return an error
	_, stderr, err := cli.Run("profile", "init", cli.SketchbookDir().Join("Simple").String(), "-m", "wrong_fqbn", "-b", "foo:bar")
	require.Error(t, err)
	require.Contains(t, string(stderr), "Invalid FQBN")
}

func initMissingFqbn(t *testing.T, env *integrationtest.Environment, cli *integrationtest.ArduinoCLI) {
	// Add a profile with no FQBN should return an error
	_, stderr, err := cli.Run("profile", "init", cli.SketchbookDir().Join("Simple").String(), "-m", "Uno")
	require.Error(t, err)
	require.Contains(t, string(stderr), "Missing FQBN (Fully Qualified Board Name)")
}

func initExistingProfile(t *testing.T, env *integrationtest.Environment, cli *integrationtest.ArduinoCLI) {
	// Adding a profile with a name that already exists should return an error
	_, stderr, err := cli.Run("profile", "init", cli.SketchbookDir().Join("Simple").String(), "-m", "Uno", "-b", "arduino:avr:uno")
	require.Error(t, err)
	require.Contains(t, string(stderr), "Profile 'Uno' already exists")
}

func initSetDefaultProfile(t *testing.T, env *integrationtest.Environment, cli *integrationtest.ArduinoCLI) {
	// Adding a profile with a name that already exists should return an error
	_, _, err := cli.Run("profile", "init", cli.SketchbookDir().Join("Simple").String(), "-m", "new_profile", "-b", "arduino:avr:uno", "--default")
	require.NoError(t, err)
	fileContent, err := cli.SketchbookDir().Join("Simple", "sketch.yaml").ReadFileAsLines()
	require.NoError(t, err)
	require.Contains(t, fileContent, "default_profile: new_profile")
}

func TestInitProfileMissingSketchFile(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	// Init the environment explicitly
	_, _, err := cli.Run("core", "update-index")
	require.NoError(t, err)

	_, stderr, err := cli.Run("profile", "init", cli.SketchbookDir().Join("Simple").String())
	require.Error(t, err)
	require.Contains(t, string(stderr), "no such file or directory")

	err = cli.SketchbookDir().Join("Simple").MkdirAll()
	require.NoError(t, err)
	_, stderr, err = cli.Run("profile", "init", cli.SketchbookDir().Join("Simple").String())
	require.Error(t, err)
	require.Contains(t, string(stderr), "main file missing from sketch")
}

func TestInitProfilePlatformNotInstalled(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	// Init the environment explicitly
	_, _, err := cli.Run("core", "update-index")
	require.NoError(t, err)

	_, _, err = cli.Run("sketch", "new", cli.SketchbookDir().Join("Simple").String())
	require.NoError(t, err)

	// Adding a profile with a name that already exists should return an error
	_, stderr, err := cli.Run("profile", "init", cli.SketchbookDir().Join("Simple").String(), "-m", "Uno", "-b", "arduino:avr:uno")
	require.Error(t, err)
	require.Contains(t, string(stderr), "platform not installed")
}

func TestProfileLib(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	// Init the environment explicitly
	_, _, err := cli.Run("core", "update-index")
	require.NoError(t, err)

	_, _, err = cli.Run("sketch", "new", cli.SketchbookDir().Join("Simple").String())
	require.NoError(t, err)

	_, _, err = cli.Run("core", "install", "arduino:avr")
	require.NoError(t, err)

	_, _, err = cli.Run("profile", "init", cli.SketchbookDir().Join("Simple").String(), "-m", "Uno", "-b", "arduino:avr:uno")
	require.NoError(t, err)

	integrationtest.CLISubtests{
		{"AddLibToDefaultProfile", addLibToDefaultProfile},
		{"ChangeLibVersionDefaultProfile", changeLibVersionDefaultProfile},
		{"RemoveLibFromDefaultProfile", removeLibFromDefaultProfile},
		{"AddInexistentLibToDefaultProfile", addInexistentLibToDefaultProfile},
		{"RemoveLibNotInDefaultProfile", removeLibNotInDefaultProfile},
	}.Run(t, env, cli)
}

func addLibToDefaultProfile(t *testing.T, env *integrationtest.Environment, cli *integrationtest.ArduinoCLI) {
	_, _, err := cli.Run("profile", "lib", "add", "Modulino@0.5.0", "--dest-dir", cli.SketchbookDir().Join("Simple").String())
	require.NoError(t, err)
	fileContent, err := cli.SketchbookDir().Join("Simple", "sketch.yaml").ReadFile()
	require.NoError(t, err)
	require.Equal(t, "profiles:\n  Uno:\n    fqbn: arduino:avr:uno\n    platforms:\n      - platform: arduino:avr (1.8.6)\n    libraries:\n      - Modulino (0.5.0)\n\ndefault_profile: Uno\n", string(fileContent))
}

func changeLibVersionDefaultProfile(t *testing.T, env *integrationtest.Environment, cli *integrationtest.ArduinoCLI) {
	fileContent, err := cli.SketchbookDir().Join("Simple", "sketch.yaml").ReadFile()
	require.NoError(t, err)
	require.Equal(t, "profiles:\n  Uno:\n    fqbn: arduino:avr:uno\n    platforms:\n      - platform: arduino:avr (1.8.6)\n    libraries:\n      - Modulino (0.5.0)\n\ndefault_profile: Uno\n", string(fileContent))

	_, _, err = cli.Run("profile", "lib", "add", "Modulino@0.4.0", "--dest-dir", cli.SketchbookDir().Join("Simple").String())
	require.NoError(t, err)
	fileContent, err = cli.SketchbookDir().Join("Simple", "sketch.yaml").ReadFile()
	require.NoError(t, err)
	require.Equal(t, "profiles:\n  Uno:\n    fqbn: arduino:avr:uno\n    platforms:\n      - platform: arduino:avr (1.8.6)\n    libraries:\n      - Modulino (0.4.0)\n\ndefault_profile: Uno\n", string(fileContent))
}

func removeLibFromDefaultProfile(t *testing.T, env *integrationtest.Environment, cli *integrationtest.ArduinoCLI) {
	_, _, err := cli.Run("profile", "lib", "remove", "Modulino", "--dest-dir", cli.SketchbookDir().Join("Simple").String())
	require.NoError(t, err)
	fileContent, err := cli.SketchbookDir().Join("Simple", "sketch.yaml").ReadFile()
	require.NoError(t, err)
	require.Equal(t, "profiles:\n  Uno:\n    fqbn: arduino:avr:uno\n    platforms:\n      - platform: arduino:avr (1.8.6)\n    libraries:\n\ndefault_profile: Uno\n", string(fileContent))
}

func addInexistentLibToDefaultProfile(t *testing.T, env *integrationtest.Environment, cli *integrationtest.ArduinoCLI) {
	_, stderr, err := cli.Run("profile", "lib", "add", "foobar", "--dest-dir", cli.SketchbookDir().Join("Simple").String())
	require.Error(t, err)
	require.Equal(t, "Error adding foobar to the profile : Library 'foobar@latest' not found\n", string(stderr))
}

func removeLibNotInDefaultProfile(t *testing.T, env *integrationtest.Environment, cli *integrationtest.ArduinoCLI) {
	_, stderr, err := cli.Run("profile", "lib", "remove", "Arduino_JSON", "--dest-dir", cli.SketchbookDir().Join("Simple").String())
	require.Error(t, err)
	require.Equal(t, "Error removing Arduino_JSON from the profile : Library 'Arduino_JSON' not found\n", string(stderr))
}

func TestProfileLibSpecificProfile(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	// Init the environment explicitly
	_, _, err := cli.Run("core", "update-index")
	require.NoError(t, err)

	_, _, err = cli.Run("sketch", "new", cli.SketchbookDir().Join("Simple").String())
	require.NoError(t, err)

	_, _, err = cli.Run("core", "install", "arduino:avr")
	require.NoError(t, err)

	_, _, err = cli.Run("profile", "init", cli.SketchbookDir().Join("Simple").String(), "-m", "Uno", "-b", "arduino:avr:uno")
	require.NoError(t, err)

	// Add a second profile
	_, _, err = cli.Run("profile", "init", cli.SketchbookDir().Join("Simple").String(), "-m", "my_profile", "-b", "arduino:avr:uno")
	require.NoError(t, err)

	// Add library to a specific profile
	_, _, err = cli.Run("profile", "lib", "add", "Modulino@0.5.0", "-m", "my_profile", "--dest-dir", cli.SketchbookDir().Join("Simple").String())
	require.NoError(t, err)
	fileContent, err := cli.SketchbookDir().Join("Simple", "sketch.yaml").ReadFile()
	require.NoError(t, err)
	require.Contains(t, string(fileContent), "  my_profile:\n    fqbn: arduino:avr:uno\n    platforms:\n      - platform: arduino:avr (1.8.6)\n    libraries:\n      - Modulino (0.5.0)\n")

	// Remove library from a specific profile
	_, _, err = cli.Run("profile", "lib", "remove", "Modulino", "-m", "my_profile", "--dest-dir", cli.SketchbookDir().Join("Simple").String())
	require.NoError(t, err)
	fileContent, err = cli.SketchbookDir().Join("Simple", "sketch.yaml").ReadFile()
	require.NoError(t, err)
	require.NotContains(t, string(fileContent), "- Modulino (0.5.0)")
}

func TestProfileSetDefault(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	// Init the environment explicitly
	_, _, err := cli.Run("core", "update-index")
	require.NoError(t, err)

	_, _, err = cli.Run("sketch", "new", cli.SketchbookDir().Join("Simple").String())
	require.NoError(t, err)

	_, _, err = cli.Run("core", "install", "arduino:avr")
	require.NoError(t, err)

	_, _, err = cli.Run("profile", "init", cli.SketchbookDir().Join("Simple").String(), "-m", "Uno", "-b", "arduino:avr:uno")
	require.NoError(t, err)

	// Add a second profile
	_, _, err = cli.Run("profile", "init", cli.SketchbookDir().Join("Simple").String(), "-m", "my_profile", "-b", "arduino:avr:uno")
	require.NoError(t, err)
	fileContent, err := cli.SketchbookDir().Join("Simple", "sketch.yaml").ReadFileAsLines()
	require.NoError(t, err)
	require.Contains(t, fileContent, "default_profile: Uno")
	require.NotContains(t, fileContent, "default_profile: my_profile")

	// Change default profile
	_, _, err = cli.Run("profile", "set-default", "my_profile", "--dest-dir", cli.SketchbookDir().Join("Simple").String())
	require.NoError(t, err)
	fileContent, err = cli.SketchbookDir().Join("Simple", "sketch.yaml").ReadFileAsLines()
	require.NoError(t, err)
	require.NotContains(t, fileContent, "default_profile: Uno")
	require.Contains(t, fileContent, "default_profile: my_profile")

	// Changing to an inexistent profile returns an error
	_, stderr, err := cli.Run("profile", "set-default", "inexistent_profile", "--dest-dir", cli.SketchbookDir().Join("Simple").String())
	require.Error(t, err)
	require.Equal(t, "Cannot set inexistent_profile as default profile: Profile 'inexistent_profile' not found\n", string(stderr))
}
