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
	"os"
	"testing"
	"time"

	"github.com/arduino/arduino-cli/arduino/builder/utils"
	paths "github.com/arduino/go-paths-helper"
	"github.com/stretchr/testify/require"
)

func tempFile(t *testing.T, prefix string) *paths.Path {
	file, err := os.CreateTemp("", prefix)
	file.Close()
	NoError(t, err)
	return paths.New(file.Name())
}

func TestObjFileIsUpToDateObjMissing(t *testing.T) {
	sourceFile := tempFile(t, "source")
	defer sourceFile.RemoveAll()

	upToDate, err := utils.ObjFileIsUpToDate(sourceFile, nil, nil)
	NoError(t, err)
	require.False(t, upToDate)
}

func TestObjFileIsUpToDateDepMissing(t *testing.T) {
	sourceFile := tempFile(t, "source")
	defer sourceFile.RemoveAll()

	objFile := tempFile(t, "obj")
	defer objFile.RemoveAll()

	upToDate, err := utils.ObjFileIsUpToDate(sourceFile, objFile, nil)
	NoError(t, err)
	require.False(t, upToDate)
}

func TestObjFileIsUpToDateObjOlder(t *testing.T) {
	objFile := tempFile(t, "obj")
	defer objFile.RemoveAll()
	depFile := tempFile(t, "dep")
	defer depFile.RemoveAll()

	time.Sleep(time.Second)

	sourceFile := tempFile(t, "source")
	defer sourceFile.RemoveAll()

	upToDate, err := utils.ObjFileIsUpToDate(sourceFile, objFile, depFile)
	NoError(t, err)
	require.False(t, upToDate)
}

func TestObjFileIsUpToDateObjNewer(t *testing.T) {
	sourceFile := tempFile(t, "source")
	defer sourceFile.RemoveAll()

	time.Sleep(time.Second)

	objFile := tempFile(t, "obj")
	defer objFile.RemoveAll()
	depFile := tempFile(t, "dep")
	defer depFile.RemoveAll()

	upToDate, err := utils.ObjFileIsUpToDate(sourceFile, objFile, depFile)
	NoError(t, err)
	require.True(t, upToDate)
}

func TestObjFileIsUpToDateDepIsNewer(t *testing.T) {
	sourceFile := tempFile(t, "source")
	defer sourceFile.RemoveAll()

	time.Sleep(time.Second)

	objFile := tempFile(t, "obj")
	defer objFile.RemoveAll()
	depFile := tempFile(t, "dep")
	defer depFile.RemoveAll()

	time.Sleep(time.Second)

	headerFile := tempFile(t, "header")
	defer headerFile.RemoveAll()

	data := objFile.String() + ": \\\n\t" + sourceFile.String() + " \\\n\t" + headerFile.String()
	depFile.WriteFile([]byte(data))

	upToDate, err := utils.ObjFileIsUpToDate(sourceFile, objFile, depFile)
	NoError(t, err)
	require.False(t, upToDate)
}

func TestObjFileIsUpToDateDepIsOlder(t *testing.T) {
	sourceFile := tempFile(t, "source")
	defer sourceFile.RemoveAll()

	headerFile := tempFile(t, "header")
	defer headerFile.RemoveAll()

	time.Sleep(time.Second)

	objFile := tempFile(t, "obj")
	defer objFile.RemoveAll()
	depFile := tempFile(t, "dep")
	defer depFile.RemoveAll()

	res := objFile.String() + ": \\\n\t" + sourceFile.String() + " \\\n\t" + headerFile.String()
	depFile.WriteFile([]byte(res))

	upToDate, err := utils.ObjFileIsUpToDate(sourceFile, objFile, depFile)
	NoError(t, err)
	require.True(t, upToDate)
}

func TestObjFileIsUpToDateDepIsWrong(t *testing.T) {
	sourceFile := tempFile(t, "source")
	defer sourceFile.RemoveAll()

	time.Sleep(time.Second)

	objFile := tempFile(t, "obj")
	defer objFile.RemoveAll()
	depFile := tempFile(t, "dep")
	defer depFile.RemoveAll()

	time.Sleep(time.Second)

	headerFile := tempFile(t, "header")
	defer headerFile.RemoveAll()

	res := sourceFile.String() + ": \\\n\t" + sourceFile.String() + " \\\n\t" + headerFile.String()
	depFile.WriteFile([]byte(res))

	upToDate, err := utils.ObjFileIsUpToDate(sourceFile, objFile, depFile)
	NoError(t, err)
	require.False(t, upToDate)
}
