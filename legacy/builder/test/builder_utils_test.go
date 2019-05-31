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
	"io/ioutil"
	"testing"
	"time"

	"github.com/arduino/arduino-cli/legacy/builder/builder_utils"
	"github.com/arduino/arduino-cli/legacy/builder/types"
	paths "github.com/arduino/go-paths-helper"
	"github.com/stretchr/testify/require"
)

func sleep(t *testing.T) {
	dur, err := time.ParseDuration("1s")
	NoError(t, err)
	time.Sleep(dur)
}

func tempFile(t *testing.T, prefix string) *paths.Path {
	file, err := ioutil.TempFile("", prefix)
	file.Close()
	NoError(t, err)
	return paths.New(file.Name())
}

func TestObjFileIsUpToDateObjMissing(t *testing.T) {
	ctx := &types.Context{}

	sourceFile := tempFile(t, "source")
	defer sourceFile.RemoveAll()

	upToDate, err := builder_utils.ObjFileIsUpToDate(ctx, sourceFile, nil, nil)
	NoError(t, err)
	require.False(t, upToDate)
}

func TestObjFileIsUpToDateDepMissing(t *testing.T) {
	ctx := &types.Context{}

	sourceFile := tempFile(t, "source")
	defer sourceFile.RemoveAll()

	objFile := tempFile(t, "obj")
	defer objFile.RemoveAll()

	upToDate, err := builder_utils.ObjFileIsUpToDate(ctx, sourceFile, objFile, nil)
	NoError(t, err)
	require.False(t, upToDate)
}

func TestObjFileIsUpToDateObjOlder(t *testing.T) {
	ctx := &types.Context{}

	objFile := tempFile(t, "obj")
	defer objFile.RemoveAll()
	depFile := tempFile(t, "dep")
	defer depFile.RemoveAll()

	sleep(t)

	sourceFile := tempFile(t, "source")
	defer sourceFile.RemoveAll()

	upToDate, err := builder_utils.ObjFileIsUpToDate(ctx, sourceFile, objFile, depFile)
	NoError(t, err)
	require.False(t, upToDate)
}

func TestObjFileIsUpToDateObjNewer(t *testing.T) {
	ctx := &types.Context{}

	sourceFile := tempFile(t, "source")
	defer sourceFile.RemoveAll()

	sleep(t)

	objFile := tempFile(t, "obj")
	defer objFile.RemoveAll()
	depFile := tempFile(t, "dep")
	defer depFile.RemoveAll()

	upToDate, err := builder_utils.ObjFileIsUpToDate(ctx, sourceFile, objFile, depFile)
	NoError(t, err)
	require.True(t, upToDate)
}

func TestObjFileIsUpToDateDepIsNewer(t *testing.T) {
	ctx := &types.Context{}

	sourceFile := tempFile(t, "source")
	defer sourceFile.RemoveAll()

	sleep(t)

	objFile := tempFile(t, "obj")
	defer objFile.RemoveAll()
	depFile := tempFile(t, "dep")
	defer depFile.RemoveAll()

	sleep(t)

	headerFile := tempFile(t, "header")
	defer headerFile.RemoveAll()

	data := objFile.String() + ": \\\n\t" + sourceFile.String() + " \\\n\t" + headerFile.String()
	depFile.WriteFile([]byte(data))

	upToDate, err := builder_utils.ObjFileIsUpToDate(ctx, sourceFile, objFile, depFile)
	NoError(t, err)
	require.False(t, upToDate)
}

func TestObjFileIsUpToDateDepIsOlder(t *testing.T) {
	ctx := &types.Context{}

	sourceFile := tempFile(t, "source")
	defer sourceFile.RemoveAll()

	headerFile := tempFile(t, "header")
	defer headerFile.RemoveAll()

	sleep(t)

	objFile := tempFile(t, "obj")
	defer objFile.RemoveAll()
	depFile := tempFile(t, "dep")
	defer depFile.RemoveAll()

	res := objFile.String() + ": \\\n\t" + sourceFile.String() + " \\\n\t" + headerFile.String()
	depFile.WriteFile([]byte(res))

	upToDate, err := builder_utils.ObjFileIsUpToDate(ctx, sourceFile, objFile, depFile)
	NoError(t, err)
	require.True(t, upToDate)
}

func TestObjFileIsUpToDateDepIsWrong(t *testing.T) {
	ctx := &types.Context{}

	sourceFile := tempFile(t, "source")
	defer sourceFile.RemoveAll()

	sleep(t)

	objFile := tempFile(t, "obj")
	defer objFile.RemoveAll()
	depFile := tempFile(t, "dep")
	defer depFile.RemoveAll()

	sleep(t)

	headerFile := tempFile(t, "header")
	defer headerFile.RemoveAll()

	res := sourceFile.String() + ": \\\n\t" + sourceFile.String() + " \\\n\t" + headerFile.String()
	depFile.WriteFile([]byte(res))

	upToDate, err := builder_utils.ObjFileIsUpToDate(ctx, sourceFile, objFile, depFile)
	NoError(t, err)
	require.False(t, upToDate)
}
