// This file is part of arduino-cli.
//
// Copyright 2020-2022 ARDUINO SA (http://www.arduino.cc/)
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

package sketch

import (
	"fmt"
	"testing"

	"github.com/arduino/go-paths-helper"
	"github.com/stretchr/testify/require"
)

func TestProjectFileLoading(t *testing.T) {
	sketchProj := paths.New("testdata", "SketchWithProfiles", "sketch.yml")
	proj, err := LoadProjectFile(sketchProj)
	require.NoError(t, err)
	fmt.Println(proj)
	golden, err := sketchProj.ReadFile()
	require.NoError(t, err)
	require.Equal(t, proj.AsYaml(), string(golden))
}
