/*
 * This file is part of Arduino Builder.
 *
 * Arduino Builder is free software; you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation; either version 2 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program; if not, write to the Free Software
 * Foundation, Inc., 51 Franklin St, Fifth Floor, Boston, MA  02110-1301  USA
 *
 * As a special exception, you may use this file as part of a free software
 * library without restriction.  Specifically, if other files instantiate
 * templates or use macros or inline functions from this file, or you compile
 * this file and link it with other files to produce an executable, this
 * file does not by itself cause the resulting executable to be covered by
 * the GNU General Public License.  This exception does not however
 * invalidate any other reasons why the executable file might be covered by
 * the GNU General Public License.
 *
 * Copyright 2015 Arduino LLC (http://www.arduino.cc/)
 */

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
	ctx.ImportedLibraries = []*libraries.Library{&libraries.Library{Name: "Bridge"}}

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
	ctx.ImportedLibraries = []*libraries.Library{&libraries.Library{Name: "Bridge"}}

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
