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

package builder

import (
	"os/exec"
	"testing"

	"github.com/arduino/go-paths-helper"
	"github.com/stretchr/testify/require"
)

func TestCompilationDatabase(t *testing.T) {
	tmpfile, err := paths.WriteToTempFile([]byte{}, nil, "")
	require.NoError(t, err)
	defer tmpfile.Remove()

	cmd := exec.Command("gcc", "arg1", "arg2")
	db := NewCompilationDatabase(tmpfile)
	db.Add(paths.New("test"), cmd)
	db.SaveToFile()

	db2, err := LoadCompilationDatabase(tmpfile)
	require.NoError(t, err)
	require.Equal(t, db, db2)
	require.Len(t, db2.Contents, 1)
	require.Equal(t, db2.Contents[0].File, "test")
	require.Equal(t, db2.Contents[0].Command, "")
	require.Equal(t, db2.Contents[0].Arguments, []string{"gcc", "arg1", "arg2"})
	cwd, err := paths.Getwd()
	require.NoError(t, err)
	require.Equal(t, db2.Contents[0].Directory, cwd.String())
}
