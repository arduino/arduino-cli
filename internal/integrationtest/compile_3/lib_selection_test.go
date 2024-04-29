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
	"go.bug.st/testifyjson/requirejson"
)

func TestCompileLibrarySelection(t *testing.T) {
	// See: https://github.com/arduino/arduino-cli/issues/2106

	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	// Run update-index with our test index
	_, _, err := cli.Run("core", "install", "arduino:avr@1.8.5")
	require.NoError(t, err)

	// Prepare sketchbook and sketch
	sketchBook, err := paths.New("testdata", "sketchbook_for_testing_lib_priorities").Abs()
	require.NoError(t, err)
	vars := cli.GetDefaultEnv()
	vars["ARDUINO_DIRECTORIES_USER"] = sketchBook.String()

	sketch := sketchBook.Join("SketchUsingLibraryA")
	anotherLib := sketchBook.Join("libraries", "AnotherLibrary")

	// Perform two compile:
	// - the first should use LibraryA
	stdout, _, err := cli.RunWithCustomEnv(vars, "compile", "-b", "arduino:avr:mega", "--json", sketch.String())
	require.NoError(t, err)
	requirejson.Contains(t, stdout, `{
		"builder_result" : {
			"used_libraries" : [
				{ "name": "LibraryA" }
			]
		}
	}`)

	// - the second should use AnotherLibrary (because it was forced by --library)
	stdout, _, err = cli.RunWithCustomEnv(vars, "compile", "-b", "arduino:avr:mega", "--library", anotherLib.String(), "--json", sketch.String())
	require.NoError(t, err)
	requirejson.Contains(t, stdout, `{
		"builder_result" : {
			"used_libraries" : [
				{ "name": "AnotherLibrary" }
			]
		}
	}`)
}
