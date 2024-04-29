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
