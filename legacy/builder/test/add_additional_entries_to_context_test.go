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
	paths "github.com/arduino/go-paths-helper"
	"github.com/stretchr/testify/require"
)

func TestAddAdditionalEntriesToContextNoBuildPath(t *testing.T) {
	ctx := &types.Context{}

	command := builder.AddAdditionalEntriesToContext{}
	NoError(t, command.Run(ctx))

	require.Empty(t, ctx.PreprocPath)
	require.Empty(t, ctx.SketchBuildPath)
	require.Empty(t, ctx.LibrariesBuildPath)
	require.Empty(t, ctx.CoreBuildPath)

	require.NotNil(t, ctx.WarningsLevel)

	require.True(t, ctx.CollectedSourceFiles.Empty())

	require.Equal(t, 0, len(ctx.LibrariesResolutionResults))
}

func TestAddAdditionalEntriesToContextWithBuildPath(t *testing.T) {
	ctx := &types.Context{}
	ctx.BuildPath = paths.New("folder")

	command := builder.AddAdditionalEntriesToContext{}
	NoError(t, command.Run(ctx))

	require.Equal(t, Abs(t, paths.New("folder", constants.FOLDER_PREPROC)), ctx.PreprocPath)
	require.Equal(t, Abs(t, paths.New("folder", constants.FOLDER_SKETCH)), ctx.SketchBuildPath)
	require.Equal(t, Abs(t, paths.New("folder", "libraries")), ctx.LibrariesBuildPath)
	require.Equal(t, Abs(t, paths.New("folder", constants.FOLDER_CORE)), ctx.CoreBuildPath)

	require.NotNil(t, ctx.WarningsLevel)

	require.True(t, ctx.CollectedSourceFiles.Empty())

	require.Equal(t, 0, len(ctx.LibrariesResolutionResults))
}
