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

	"github.com/arduino/arduino-cli/arduino/libraries"
	"github.com/arduino/arduino-cli/legacy/builder"
	"github.com/arduino/arduino-cli/legacy/builder/types"
	paths "github.com/arduino/go-paths-helper"
	"github.com/stretchr/testify/require"
)

func TestUnusedCompiledLibrariesRemover(t *testing.T) {
	temp, err := paths.MkTempDir("", "test")
	NoError(t, err)
	defer temp.RemoveAll()

	NoError(t, temp.Join("SPI").MkdirAll())
	NoError(t, temp.Join("Bridge").MkdirAll())
	NoError(t, temp.Join("dummy_file").WriteFile([]byte{}))

	ctx := &types.Context{}
	ctx.LibrariesBuildPath = temp
	ctx.ImportedLibraries = []*libraries.Library{{Name: "Bridge"}}

	cmd := builder.UnusedCompiledLibrariesRemover{}
	err = cmd.Run(ctx)
	NoError(t, err)

	exist, err := temp.Join("SPI").ExistCheck()
	require.NoError(t, err)
	require.False(t, exist)
	exist, err = temp.Join("Bridge").ExistCheck()
	NoError(t, err)
	require.True(t, exist)
	exist, err = temp.Join("dummy_file").ExistCheck()
	NoError(t, err)
	require.True(t, exist)
}

func TestUnusedCompiledLibrariesRemoverLibDoesNotExist(t *testing.T) {
	ctx := &types.Context{}
	ctx.LibrariesBuildPath = paths.TempDir().Join("test")
	ctx.ImportedLibraries = []*libraries.Library{{Name: "Bridge"}}

	cmd := builder.UnusedCompiledLibrariesRemover{}
	err := cmd.Run(ctx)
	NoError(t, err)
}

func TestUnusedCompiledLibrariesRemoverNoUsedLibraries(t *testing.T) {
	temp, err := paths.MkTempDir("", "test")
	NoError(t, err)
	defer temp.RemoveAll()

	NoError(t, temp.Join("SPI").MkdirAll())
	NoError(t, temp.Join("Bridge").MkdirAll())
	NoError(t, temp.Join("dummy_file").WriteFile([]byte{}))

	ctx := &types.Context{}
	ctx.LibrariesBuildPath = temp
	ctx.ImportedLibraries = []*libraries.Library{}

	cmd := builder.UnusedCompiledLibrariesRemover{}
	err = cmd.Run(ctx)
	NoError(t, err)

	exist, err := temp.Join("SPI").ExistCheck()
	require.NoError(t, err)
	require.False(t, exist)
	exist, err = temp.Join("Bridge").ExistCheck()
	require.NoError(t, err)
	require.False(t, exist)
	exist, err = temp.Join("dummy_file").ExistCheck()
	NoError(t, err)
	require.True(t, exist)
}
