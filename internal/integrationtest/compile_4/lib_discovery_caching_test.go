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
	"go.bug.st/testifyjson/requirejson"
)

func TestLibDiscoveryCache(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	t.Cleanup(env.CleanUp)

	// Install Arduino AVR Boards
	_, _, err := cli.Run("core", "install", "arduino:avr@1.8.6")
	require.NoError(t, err)

	// Copy the testdata sketchbook
	testdata, err := paths.New("testdata", "libraries_discovery_caching").Abs()
	require.NoError(t, err)
	sketchbook := cli.SketchbookDir()
	require.NoError(t, sketchbook.RemoveAll())
	require.NoError(t, testdata.CopyDirTo(cli.SketchbookDir()))

	buildpath, err := paths.MkTempDir("", "tmpbuildpath")
	require.NoError(t, err)
	t.Cleanup(func() { buildpath.RemoveAll() })

	{
		sketchA := sketchbook.Join("SketchA")
		{
			outjson, _, err := cli.Run("compile", "-v", "-b", "arduino:avr:uno", "--build-path", buildpath.String(), "--json", sketchA.String())
			require.NoError(t, err)
			j := requirejson.Parse(t, outjson)
			j.MustContain(`{"builder_result":{
		"used_libraries": [
			{ "name": "LibA" },
        	{ "name": "LibB" }
    	],
	}}`)
		}

		// Update SketchA
		require.NoError(t, sketchA.Join("SketchA.ino").WriteFile([]byte(`
#include <LibC.h>
#include <LibA.h>
void setup() {}
void loop() {libAFunction();}
`)))

		{
			// This compile should FAIL!
			outjson, _, err := cli.Run("compile", "-v", "-b", "arduino:avr:uno", "--build-path", buildpath.String(), "--json", sketchA.String())
			require.Error(t, err)
			j := requirejson.Parse(t, outjson)
			j.MustContain(`{
"builder_result":{
	"used_libraries": [
		{ "name": "LibC" },
		{ "name": "LibA" }
	],
    "diagnostics": [
		{
			"severity": "ERROR",
			"message": "'libAFunction' was not declared in this scope\n void loop() {libAFunction();}\n              ^~~~~~~~~~~~"
		}
	]
}}`)
			j.Query(".compiler_out").MustContain(`"The list of included libraries has been changed... rebuilding all libraries."`)
		}

		{
			// This compile should FAIL!
			outjson, _, err := cli.Run("compile", "-v", "-b", "arduino:avr:uno", "--build-path", buildpath.String(), "--json", sketchA.String())
			require.Error(t, err)
			j := requirejson.Parse(t, outjson)
			j.MustContain(`{
"builder_result":{
	"used_libraries": [
		{ "name": "LibC" },
		{ "name": "LibA" }
	],
    "diagnostics": [
		{
			"severity": "ERROR",
			"message": "'libAFunction' was not declared in this scope\n void loop() {libAFunction();}\n              ^~~~~~~~~~~~"
		}
	]
}}`)
			j.Query(".compiler_out").MustNotContain(`"The list of included libraries has changed... rebuilding all libraries."`)
		}
	}
}
