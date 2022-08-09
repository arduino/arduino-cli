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

package core_test

import (
	"testing"

	"github.com/arduino/arduino-cli/internal/integrationtest"
	"github.com/arduino/go-paths-helper"
	"github.com/stretchr/testify/require"
	"go.bug.st/testifyjson/requirejson"
)

func TestCorrectHandlingOfPlatformVersionProperty(t *testing.T) {
	// See: https://github.com/arduino/arduino-cli/issues/1823
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	// Copy test platform
	testPlatform := paths.New("testdata", "issue_1823", "DxCore-dev")
	require.NoError(t, testPlatform.CopyDirTo(cli.SketchbookDir().Join("hardware", "DxCore-dev")))

	// Trigger problematic call
	out, _, err := cli.Run("core", "list", "--format", "json")
	require.NoError(t, err)
	requirejson.Contains(t, out, `[{"id":"DxCore-dev:megaavr","installed":"1.4.10","name":"DxCore"}]`)
}
