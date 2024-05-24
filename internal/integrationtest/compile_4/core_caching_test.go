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
	"testing"

	"github.com/arduino/arduino-cli/internal/integrationtest"
	"github.com/arduino/go-paths-helper"
	"github.com/stretchr/testify/require"
)

func TestBuildCacheCoreWithExtraDirs(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	t.Cleanup(env.CleanUp)

	// Install Arduino AVR Boards
	_, _, err := cli.Run("core", "install", "arduino:avr@1.8.6")
	require.NoError(t, err)

	// Main core cache
	defaultCache := paths.TempDir().Join("arduino")
	cache1, err := paths.MkTempDir("", "core_cache")
	require.NoError(t, err)
	t.Cleanup(func() { cache1.RemoveAll() })
	cache2, err := paths.MkTempDir("", "extra_core_cache")
	require.NoError(t, err)
	t.Cleanup(func() { cache2.RemoveAll() })

	sketch, err := paths.New("testdata", "BareMinimum").Abs()
	require.NoError(t, err)

	{
		// Compile sketch with empty cache
		out, _, err := cli.Run("compile", "-v", "-b", "arduino:avr:uno", sketch.String())
		require.NoError(t, err)
		require.Contains(t, string(out), "Archiving built core (caching) in: "+defaultCache.String())

		// Check that the core cache is re-used
		out, _, err = cli.Run("compile", "-v", "-b", "arduino:avr:uno", sketch.String())
		require.NoError(t, err)
		require.Contains(t, string(out), "Using precompiled core: "+defaultCache.String())
	}

	{
		env := cli.GetDefaultEnv()
		env["ARDUINO_BUILD_CACHE_PATH"] = cache1.String()

		// Compile sketch with empty cache user-defined core cache
		out, _, err := cli.RunWithCustomEnv(env, "compile", "-v", "-b", "arduino:avr:uno", sketch.String())
		require.NoError(t, err)
		require.Contains(t, string(out), "Archiving built core (caching) in: "+cache1.String())

		// Check that the core cache is re-used with user-defined core cache
		out, _, err = cli.RunWithCustomEnv(env, "compile", "-v", "-b", "arduino:avr:uno", sketch.String())
		require.NoError(t, err)
		require.Contains(t, string(out), "Using precompiled core: "+cache1.String())

		// Clean run should rebuild and save in user-defined core cache
		out, _, err = cli.RunWithCustomEnv(env, "compile", "-v", "-b", "arduino:avr:uno", "--clean", sketch.String())
		require.NoError(t, err)
		require.Contains(t, string(out), "Archiving built core (caching) in: "+cache1.String())
	}

	{
		env := cli.GetDefaultEnv()
		env["ARDUINO_BUILD_CACHE_EXTRA_PATHS"] = cache1.String()

		// Both extra and default cache are full, should use the default one
		out, _, err := cli.RunWithCustomEnv(env, "compile", "-v", "-b", "arduino:avr:uno", sketch.String())
		require.NoError(t, err)
		require.Contains(t, string(out), "Using precompiled core: "+defaultCache.String())

		// Clean run, should rebuild and save in default cache
		out, _, err = cli.RunWithCustomEnv(env, "compile", "-v", "-b", "arduino:avr:uno", "--clean", sketch.String())
		require.NoError(t, err)
		require.Contains(t, string(out), "Archiving built core (caching) in: "+defaultCache.String())

		// Clean default cache
		require.NoError(t, defaultCache.RemoveAll())

		// Now, extra is full and default is empty, should use extra
		out, _, err = cli.RunWithCustomEnv(env, "compile", "-v", "-b", "arduino:avr:uno", sketch.String())
		require.NoError(t, err)
		require.Contains(t, string(out), "Using precompiled core: "+cache1.String())
	}

	{
		env := cli.GetDefaultEnv()
		env["ARDUINO_BUILD_CACHE_EXTRA_PATHS"] = cache1.String() // Populated
		env["ARDUINO_BUILD_CACHE_PATH"] = cache2.String()        // Empty

		// Extra cache is full, should use the cache1 (extra)
		out, _, err := cli.RunWithCustomEnv(env, "compile", "-v", "-b", "arduino:avr:uno", sketch.String())
		require.NoError(t, err)
		require.Contains(t, string(out), "Using precompiled core: "+cache1.String())

		// Clean run, should rebuild and save in cache2 (user defined default cache)
		out, _, err = cli.RunWithCustomEnv(env, "compile", "-v", "-b", "arduino:avr:uno", "--clean", sketch.String())
		require.NoError(t, err)
		require.Contains(t, string(out), "Archiving built core (caching) in: "+cache2.String())

		// Both caches are full, should use the cache2 (user defined default)
		out, _, err = cli.RunWithCustomEnv(env, "compile", "-v", "-b", "arduino:avr:uno", sketch.String())
		require.NoError(t, err)
		require.Contains(t, string(out), "Using precompiled core: "+cache2.String())
	}
}
