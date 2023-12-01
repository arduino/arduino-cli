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

package buildcache

import (
	"testing"
	"time"

	"github.com/arduino/go-paths-helper"
	"github.com/stretchr/testify/require"
)

func Test_UpdateLastUsedFileNotExisting(t *testing.T) {
	testBuildDir := paths.New(t.TempDir(), "sketches", "xxx")
	require.NoError(t, testBuildDir.MkdirAll())
	timeBeforeUpdating := time.Unix(0, 0)
	requireCorrectUpdate(t, testBuildDir, timeBeforeUpdating)
}

func Test_UpdateLastUsedFileExisting(t *testing.T) {
	testBuildDir := paths.New(t.TempDir(), "sketches", "xxx")
	require.NoError(t, testBuildDir.MkdirAll())

	// create the file
	preExistingFile := testBuildDir.Join(lastUsedFileName)
	require.NoError(t, preExistingFile.WriteFile([]byte{}))
	timeBeforeUpdating := time.Now().Add(-time.Second)
	preExistingFile.Chtimes(time.Now(), timeBeforeUpdating)
	requireCorrectUpdate(t, testBuildDir, timeBeforeUpdating)
}

func requireCorrectUpdate(t *testing.T, dir *paths.Path, prevModTime time.Time) {
	_, err := New(dir.Parent()).GetOrCreate(dir.Base())
	require.NoError(t, err)
	expectedFile := dir.Join(lastUsedFileName)
	fileInfo, err := expectedFile.Stat()
	require.Nil(t, err)
	require.Greater(t, fileInfo.ModTime(), prevModTime)
}

func TestPurge(t *testing.T) {
	ttl := time.Minute

	dirToPurge := paths.New(t.TempDir(), "root")

	lastUsedTimesByDirPath := map[*paths.Path]time.Time{
		(dirToPurge.Join("old")):   time.Now().Add(-ttl - time.Hour),
		(dirToPurge.Join("fresh")): time.Now().Add(-ttl + time.Minute),
	}

	// create the metadata files
	for dirPath, lastUsedTime := range lastUsedTimesByDirPath {
		require.NoError(t, dirPath.MkdirAll())
		infoFilePath := dirPath.Join(lastUsedFileName).Canonical()
		require.NoError(t, infoFilePath.WriteFile([]byte{}))
		// make sure access time does not matter
		accesstime := time.Now()
		require.NoError(t, infoFilePath.Chtimes(accesstime, lastUsedTime))
	}

	New(dirToPurge).Purge(ttl)

	require.False(t, dirToPurge.Join("old").Exist())
	require.True(t, dirToPurge.Join("fresh").Exist())
}
