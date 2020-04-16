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

package resources

import (
	"crypto"
	"encoding/hex"
	"testing"

	"github.com/arduino/go-paths-helper"
	"github.com/stretchr/testify/require"
	"go.bug.st/downloader"
)

func TestDownloadAndChecksums(t *testing.T) {
	tmp, err := paths.MkTempDir("", "")
	require.NoError(t, err)
	defer tmp.RemoveAll()
	testFile := tmp.Join("cache", "index.html")

	r := &DownloadResource{
		ArchiveFileName: "index.html",
		CachePath:       "cache",
		Checksum:        "SHA-256:e021e1a223d03069d5f08dea25a58ca445a7376d9bdf980f756034f118449e66",
		Size:            1119,
		URL:             "https://downloads.arduino.cc/index.html",
	}
	digest, err := hex.DecodeString("e021e1a223d03069d5f08dea25a58ca445a7376d9bdf980f756034f118449e66")
	require.NoError(t, err)

	downloadAndTestChecksum := func() {
		d, err := r.Download(tmp, &downloader.Config{})
		require.NoError(t, err)
		err = d.Run()
		require.NoError(t, err)

		data, err := testFile.ReadFile()
		require.NoError(t, err)
		algo := crypto.SHA256.New()
		algo.Write(data)
		require.EqualValues(t, digest, algo.Sum(nil))
	}

	// Normal download
	downloadAndTestChecksum()

	// Download with cached file
	d, err := r.Download(tmp, &downloader.Config{})
	require.NoError(t, err)
	require.Nil(t, d)

	// Download if cached file has data in excess (redownload)
	data, err := testFile.ReadFile()
	require.NoError(t, err)
	data = append(data, []byte("123123123")...)
	err = testFile.WriteFile(data)
	require.NoError(t, err)
	downloadAndTestChecksum()

	// Download if cached file has less data (resume)
	data, err = testFile.ReadFile()
	require.NoError(t, err)
	err = testFile.WriteFile(data[:1000])
	require.NoError(t, err)
	downloadAndTestChecksum()

	r.Checksum = "BOH:12312312312313123123123123123123"
	_, err = r.TestLocalArchiveChecksum(tmp)
	require.Error(t, err)

	r.Checksum = "MD5 667cf48afcc12c38c8c1637947a04224"
	_, err = r.TestLocalArchiveChecksum(tmp)
	require.Error(t, err)

	r.Checksum = "MD5:zmxcbzxmncbzxmnbczxmnbczxmnbcnnb"
	_, err = r.TestLocalArchiveChecksum(tmp)
	require.Error(t, err)

	r.Checksum = "SHA-1:c007e47637cc6ad6176e7d94aeffc232ee34c1c1"
	res, err := r.TestLocalArchiveChecksum(tmp)
	require.NoError(t, err)
	require.True(t, res)

	r.Checksum = "MD5:2e388576eefd92a15967868d5f566f29"
	res, err = r.TestLocalArchiveChecksum(tmp)
	require.NoError(t, err)
	require.True(t, res)

	r.Checksum = "MD5:12312312312312312312313131231231"
	res, err = r.TestLocalArchiveChecksum(tmp)
	require.NoError(t, err)
	require.False(t, res)

	_, err = r.TestLocalArchiveChecksum(paths.New("/not-existent"))
	require.Error(t, err)

	r.ArchiveFileName = "not-existent.zip"
	_, err = r.TestLocalArchiveChecksum(tmp)
	require.Error(t, err)
}
