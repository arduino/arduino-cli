// This file is part of arduino-cli.
//
// Copyright 2026 ARDUINO SA (http://www.arduino.cc/)
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

package compile

import (
	"testing"

	"github.com/arduino/arduino-cli/internal/integrationtest"
	"github.com/arduino/go-paths-helper"
	"github.com/stretchr/testify/require"
	"go.bug.st/testifyjson/requirejson"
)

func TestCompileWithNotIncludedProfileLibs(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	t.Cleanup(env.CleanUp)

	sketch, err := paths.New("testdata", "SketchWithNotIncludedProfileLibs").Abs()
	require.NoError(t, err)

	// Test that the libraries specified in the profile but not actually used in the sketch are not included in the compilation result
	out, _, err := cli.RunWithContext(t.Context(), "compile", sketch.String(), "--json")
	require.NoError(t, err, "compilation should not fail")
	jsonout := requirejson.Parse(t, out)
	jsonout.Query(".builder_result.used_libraries").MustContain(`[{"name": "FlashStorage"}]`)
	jsonout.Query(".builder_result.used_libraries").MustNotContain(`[{"name": "ArduinoRS485"}]`)
}
