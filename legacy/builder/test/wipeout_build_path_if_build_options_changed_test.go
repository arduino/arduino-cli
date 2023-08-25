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
	"github.com/arduino/arduino-cli/legacy/builder/types"
	"github.com/stretchr/testify/require"
)

func TestWipeoutBuildPathIfBuildOptionsChanged(t *testing.T) {
	ctx := &types.Context{}

	buildPath := SetupBuildPath(t, ctx)
	defer buildPath.RemoveAll()

	ctx.BuildOptionsJsonPrevious = "{ \"old\":\"old\" }"
	ctx.BuildOptionsJson = "{ \"new\":\"new\" }"

	buildPath.Join("should_be_deleted.txt").Truncate()

	commands := []types.Command{
		&builder.WipeoutBuildPathIfBuildOptionsChanged{},
	}

	for _, command := range commands {
		err := command.Run(ctx)
		NoError(t, err)
	}

	exist, err := buildPath.ExistCheck()
	NoError(t, err)
	require.True(t, exist)

	files, err := buildPath.ReadDir()
	NoError(t, err)
	require.Equal(t, 0, len(files))

	exist, err = buildPath.Join("should_be_deleted.txt").ExistCheck()
	require.NoError(t, err)
	require.False(t, exist)
}

func TestWipeoutBuildPathIfBuildOptionsChangedNoPreviousBuildOptions(t *testing.T) {
	ctx := &types.Context{}

	buildPath := SetupBuildPath(t, ctx)
	defer buildPath.RemoveAll()

	ctx.BuildOptionsJson = "{ \"new\":\"new\" }"

	require.NoError(t, buildPath.Join("should_not_be_deleted.txt").Truncate())

	commands := []types.Command{
		&builder.WipeoutBuildPathIfBuildOptionsChanged{},
	}

	for _, command := range commands {
		err := command.Run(ctx)
		NoError(t, err)
	}

	exist, err := buildPath.ExistCheck()
	NoError(t, err)
	require.True(t, exist)

	files, err := buildPath.ReadDir()
	NoError(t, err)
	require.Equal(t, 1, len(files))

	exist, err = buildPath.Join("should_not_be_deleted.txt").ExistCheck()
	NoError(t, err)
	require.True(t, exist)
}
