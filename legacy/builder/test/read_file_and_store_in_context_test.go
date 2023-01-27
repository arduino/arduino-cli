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

	"github.com/arduino/arduino-cli/legacy/builder"
	"github.com/arduino/arduino-cli/legacy/builder/types"
	paths "github.com/arduino/go-paths-helper"
	"github.com/stretchr/testify/require"
)

func TestReadFileAndStoreInContext(t *testing.T) {
	filePath, err := os.CreateTemp("", "test")
	NoError(t, err)

	file := paths.New(filePath.Name())
	defer file.RemoveAll()

	file.WriteFile([]byte("test test\nciao"))

	ctx := &types.Context{}

	command := &builder.ReadFileAndStoreInContext{FileToRead: file, Target: &ctx.SourceGccMinusE}
	err = command.Run(ctx)
	NoError(t, err)

	require.Equal(t, "test test\nciao", ctx.SourceGccMinusE)
}
