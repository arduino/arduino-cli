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

package types

import (
	"testing"

	"github.com/arduino/arduino-cli/arduino/cores"
	paths "github.com/arduino/go-paths-helper"
	"github.com/stretchr/testify/require"
)

func TestInjectBuildOption(t *testing.T) {
	fqbn, err := cores.ParseFQBN("aaa:bbb:ccc")
	require.NoError(t, err)

	{
		ctx := &Context{
			HardwareDirs:          paths.NewPathList("aaa", "bbb"),
			BuiltInToolsDirs:      paths.NewPathList("ccc", "ddd"),
			BuiltInLibrariesDirs:  paths.New("eee"),
			OtherLibrariesDirs:    paths.NewPathList("fff", "ggg"),
			SketchLocation:        paths.New("hhh"),
			FQBN:                  fqbn,
			ArduinoAPIVersion:     "iii",
			CustomBuildProperties: []string{"jjj", "kkk"},
			OptimizationFlags:     "lll",
		}
		newCtx := &Context{}
		newCtx.InjectBuildOptions(ctx.ExtractBuildOptions())
		require.Equal(t, ctx, newCtx)
	}
	{
		ctx := &Context{
			HardwareDirs:          paths.NewPathList("aaa", "bbb"),
			BuiltInToolsDirs:      paths.NewPathList("ccc", "ddd"),
			OtherLibrariesDirs:    paths.NewPathList("fff", "ggg"),
			SketchLocation:        paths.New("hhh"),
			FQBN:                  fqbn,
			ArduinoAPIVersion:     "iii",
			CustomBuildProperties: []string{"jjj", "kkk"},
			OptimizationFlags:     "lll",
		}
		newCtx := &Context{}
		newCtx.InjectBuildOptions(ctx.ExtractBuildOptions())
		require.Equal(t, ctx, newCtx)
	}
}
