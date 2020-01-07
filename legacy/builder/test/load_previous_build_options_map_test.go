// This file is part of arduino-cli.
//
// Copyright 2020 ARDUINO SA (http://www.arduino.cc/)
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

package test

import (
	"testing"

	"github.com/arduino/arduino-cli/legacy/builder"
	"github.com/arduino/arduino-cli/legacy/builder/constants"
	"github.com/arduino/arduino-cli/legacy/builder/types"
	"github.com/stretchr/testify/require"
)

func TestLoadPreviousBuildOptionsMap(t *testing.T) {
	ctx := &types.Context{}

	buildPath := SetupBuildPath(t, ctx)
	defer buildPath.RemoveAll()

	err := buildPath.Join(constants.BUILD_OPTIONS_FILE).WriteFile([]byte("test"))
	NoError(t, err)

	command := builder.LoadPreviousBuildOptionsMap{}
	err = command.Run(ctx)
	NoError(t, err)

	require.Equal(t, "test", ctx.BuildOptionsJsonPrevious)
}

func TestLoadPreviousBuildOptionsMapMissingFile(t *testing.T) {
	ctx := &types.Context{}

	buildPath := SetupBuildPath(t, ctx)
	defer buildPath.RemoveAll()

	command := builder.LoadPreviousBuildOptionsMap{}
	err := command.Run(ctx)
	NoError(t, err)

	require.Empty(t, ctx.BuildOptionsJsonPrevious)
}
