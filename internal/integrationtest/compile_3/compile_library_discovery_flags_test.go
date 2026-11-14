// This file is part of arduino-cli.
//
// Copyright (C) Arduino s.r.l. and/or its affiliated companies
//
// This software is released under the GNU General Public License version 3,
// which covers the main part of arduino-cli.
// The terms of this license can be found at:
// https://www.gnu.org/licenses/gpl-3.0.en.html

package compile_test

import (
	"encoding/json"
	"testing"

	"github.com/arduino/arduino-cli/internal/integrationtest"
	"github.com/stretchr/testify/require"
)

func TestCompileLibraryDiscoveryFlags(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	_, _, err := cli.Run("core", "update-index")
	require.NoError(t, err)
	_, _, err = cli.Run("core", "install", "arduino:avr")
	require.NoError(t, err)

	sketchPath := cli.CopySketch("sketch_with_discovery_test_library")
	libraryDir := sketchPath.Join("libraries", "DiscoveryTestLib")

	// Deliberately NOT using --show-properties: that flag returns before
	// Builder.preprocess() ever runs, so build.library_discovery_flags
	// (set inside preprocess()) would never appear — the same reason
	// {includes} is absent from --show-properties output too. A real
	// compile does run preprocess(), and its build properties (including
	// ours) are always captured into the RPC response's build_properties
	// field regardless of --show-properties; --json is how this test
	// observes that field.
	//
	// Uses --library (singular, path to a single library's root folder),
	// not --libraries (plural, path to a folder of libraries): the former
	// sets Location == libraries.Unmanaged, the latter Location ==
	// libraries.User — deliberately picking --library here to exercise the
	// IN_SPECIFIED_PATH suffix.
	stdout, stderr, err := cli.Run("compile",
		"--fqbn", "arduino:avr:uno",
		"--library", libraryDir.String(),
		"--json",
		sketchPath.String())
	require.NoError(t, err)
	require.Empty(t, stderr)

	// Decode only build_properties, not the shared cliCompileResponse type
	// (compile_show_properties_test.go): that type wraps the full
	// rpc.BuilderResult, whose UsedLibraries[].Location enum field cannot
	// be unmarshaled by encoding/json from protobuf-JSON's string encoding
	// once a sketch actually resolves a library, which this fixture does
	// (that file's own fixture is library-free, so it never hits this).
	var resp struct {
		BuilderResult struct {
			BuildProperties []string `json:"build_properties"`
		} `json:"builder_result"`
	}
	require.NoError(t, json.Unmarshal(stdout, &resp))
	// The fixture library is provided via --library, i.e.
	// Location == libraries.Unmanaged, hence IN_SPECIFIED_PATH.
	require.Contains(t, resp.BuilderResult.BuildProperties, "build.library_discovery_flags=-DFOUND_DISCOVERYTESTLIB_LIB=0x010203 -DFOUND_DISCOVERYTESTLIB_LIB_IN_SPECIFIED_PATH=1")
}
