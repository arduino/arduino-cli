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
	"os"
	"path/filepath"
	"testing"

	"github.com/arduino/arduino-cli/legacy/builder"
	"github.com/arduino/arduino-cli/legacy/builder/types"
	paths "github.com/arduino/go-paths-helper"
	"github.com/stretchr/testify/require"
)

func TestGenerateBuildPathIfMissing(t *testing.T) {
	ctx := &types.Context{
		SketchLocation: paths.New("test"),
	}

	command := builder.GenerateBuildPathIfMissing{}
	err := command.Run(ctx)
	NoError(t, err)

	require.Equal(t, filepath.Join(os.TempDir(), "arduino-sketch-098F6BCD4621D373CADE4E832627B4F6"), ctx.BuildPath.String())
	_, err = os.Stat(filepath.Join(os.TempDir(), "arduino-sketch-098F6BCD4621D373CADE4E832627B4F6"))
	require.True(t, os.IsNotExist(err))
}

func TestGenerateBuildPathIfEmpty(t *testing.T) {
	ctx := &types.Context{
		SketchLocation: paths.New("test"),
	}

	createBuildPathIfMissing := builder.GenerateBuildPathIfMissing{}
	err := createBuildPathIfMissing.Run(ctx)
	NoError(t, err)

	require.Equal(t, filepath.Join(os.TempDir(), "arduino-sketch-098F6BCD4621D373CADE4E832627B4F6"), ctx.BuildPath.String())
	_, err = os.Stat(filepath.Join(os.TempDir(), "arduino-sketch-098F6BCD4621D373CADE4E832627B4F6"))
	require.True(t, os.IsNotExist(err))
}

func TestDontGenerateBuildPathIfPresent(t *testing.T) {
	ctx := &types.Context{}
	ctx.BuildPath = paths.New("test")

	createBuildPathIfMissing := builder.GenerateBuildPathIfMissing{}
	err := createBuildPathIfMissing.Run(ctx)
	NoError(t, err)

	require.Equal(t, ctx.BuildPath.String(), "test")
}

func TestGenerateBuildPathAndEnsureItExists(t *testing.T) {
	ctx := &types.Context{
		SketchLocation: paths.New("test"),
	}

	commands := []types.Command{
		&builder.GenerateBuildPathIfMissing{},
		&builder.EnsureBuildPathExists{},
	}

	for _, command := range commands {
		err := command.Run(ctx)
		NoError(t, err)
	}

	defer os.RemoveAll(filepath.Join(os.TempDir(), "arduino-sketch-098F6BCD4621D373CADE4E832627B4F6"))

	require.Equal(t, filepath.Join(os.TempDir(), "arduino-sketch-098F6BCD4621D373CADE4E832627B4F6"), ctx.BuildPath.String())
	_, err := os.Stat(filepath.Join(os.TempDir(), "arduino-sketch-098F6BCD4621D373CADE4E832627B4F6"))
	NoError(t, err)
}
