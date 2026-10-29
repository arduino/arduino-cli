// This file is part of arduino-cli.
//
// Copyright 2026 ARDUINO SA (http://www.arduino.cc/)
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

package utils

import (
	"os"
	"runtime"
	"testing"

	"github.com/arduino/go-paths-helper"
	"github.com/stretchr/testify/require"
)

func TestCopyFileAtomically(t *testing.T) {
	dir := paths.New(t.TempDir())
	src := dir.Join("src")
	dst := dir.Join("dst")
	require.NoError(t, src.WriteFile([]byte("new")))
	require.NoError(t, src.Chmod(0644))

	// Copy to a missing dst
	require.NoError(t, CopyFileAtomically(src, dst))
	data, err := dst.ReadFile()
	require.NoError(t, err)
	require.Equal(t, "new", string(data))
	if runtime.GOOS != "windows" {
		info, err := dst.Stat()
		require.NoError(t, err)
		require.Equal(t, os.FileMode(0644), info.Mode().Perm())
	}

	// Replace an existing dst
	require.NoError(t, dst.WriteFile([]byte("old content")))
	require.NoError(t, CopyFileAtomically(src, dst))
	data, err = dst.ReadFile()
	require.NoError(t, err)
	require.Equal(t, "new", string(data))

	// A missing src leaves dst and no temp file
	require.Error(t, CopyFileAtomically(dir.Join("missing"), dst))
	files, err := dir.ReadDir()
	require.NoError(t, err)
	require.Len(t, files, 2)
	data, err = dst.ReadFile()
	require.NoError(t, err)
	require.Equal(t, "new", string(data))
}
