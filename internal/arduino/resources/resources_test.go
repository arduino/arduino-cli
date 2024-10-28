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
	"context"
	"crypto"
	"encoding/hex"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"testing"

	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	"github.com/arduino/go-paths-helper"
	"github.com/stretchr/testify/require"
	"go.bug.st/downloader/v2"
)

func TestDownloadAndChecksums(t *testing.T) {
	testFileName := "core.zip"
	tmp, err := paths.MkTempDir("", "")
	require.NoError(t, err)
	defer tmp.RemoveAll()
	testFile := tmp.Join("cache", testFileName)

	// taken from test/testdata/test_index.json
	r := &DownloadResource{
		ArchiveFileName: testFileName,
		CachePath:       "cache",
		Checksum:        "SHA-256:6a338cf4d6d501176a2d352c87a8d72ac7488b8c5b82cdf2a4e2cef630391092",
		Size:            486,
		URL:             "https://raw.githubusercontent.com/arduino/arduino-cli/master/internal/integrationtest/testdata/core.zip",
	}
	digest, err := hex.DecodeString("6a338cf4d6d501176a2d352c87a8d72ac7488b8c5b82cdf2a4e2cef630391092")
	require.NoError(t, err)

	downloadAndTestChecksum := func() {
		err := r.Download(context.Background(), tmp, downloader.Config{}, "", func(*rpc.DownloadProgress) {}, "")
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
	err = r.Download(context.Background(), tmp, downloader.Config{}, "", func(*rpc.DownloadProgress) {}, "")
	require.NoError(t, err)

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
	err = testFile.WriteFile(data[:100])
	require.NoError(t, err)
	downloadAndTestChecksum()

	require.NoError(t, r.TestLocalArchiveSize(tmp))
	r.Size = 500
	require.NoError(t, r.TestLocalArchiveSize(tmp))

	r.Checksum = ""
	_, err = r.TestLocalArchiveChecksum(tmp)
	require.Error(t, err)

	r.Checksum = "BOH:12312312312313123123123123123123"
	_, err = r.TestLocalArchiveChecksum(tmp)
	require.Error(t, err)

	r.Checksum = "MD5 667cf48afcc12c38c8c1637947a04224"
	_, err = r.TestLocalArchiveChecksum(tmp)
	require.Error(t, err)

	r.Checksum = "MD5:zmxcbzxmncbzxmnbczxmnbczxmnbcnnb"
	_, err = r.TestLocalArchiveChecksum(tmp)
	require.Error(t, err)

	r.Checksum = "SHA-1:16e1495aff482f2650733e1661d5f7c69040ec13"
	res, err := r.TestLocalArchiveChecksum(tmp)
	require.NoError(t, err)
	require.True(t, res)

	r.Checksum = "MD5:3bcc3aab6842ff124df6a5cfba678ca1"
	res, err = r.TestLocalArchiveChecksum(tmp)
	require.NoError(t, err)
	require.True(t, res)

	_, err = r.TestLocalArchiveChecksum(paths.New("/not-existent"))
	require.Error(t, err)

	r.ArchiveFileName = "not-existent.zip"
	_, err = r.TestLocalArchiveChecksum(tmp)
	require.Error(t, err)
}

func TestIndexDownloadAndSignatureWithinArchive(t *testing.T) {
	ctx := context.Background()
	// Spawn test webserver
	mux := http.NewServeMux()
	fs := http.FileServer(http.Dir("testdata"))
	mux.Handle("/", fs)
	server := &http.Server{Handler: mux}
	ln, err := net.Listen("tcp", "127.0.0.1:")
	require.NoError(t, err)
	defer ln.Close()
	go server.Serve(ln)

	validIdxURL, err := url.Parse("http://" + ln.Addr().String() + "/valid/package_index.tar.bz2")
	require.NoError(t, err)
	idxResource := &IndexResource{URL: validIdxURL}
	destDir, err := paths.MkTempDir("", "")
	require.NoError(t, err)
	defer destDir.RemoveAll()
	err = idxResource.Download(ctx, destDir, func(curr *rpc.DownloadProgress) {}, downloader.GetDefaultConfig())
	require.NoError(t, err)
	require.True(t, destDir.Join("package_index.json").Exist())
	require.True(t, destDir.Join("package_index.json.sig").Exist())

	invalidIdxURL, err := url.Parse("http://" + ln.Addr().String() + "/invalid/package_index.tar.bz2")
	require.NoError(t, err)
	invIdxResource := &IndexResource{URL: invalidIdxURL}
	invDestDir, err := paths.MkTempDir("", "")
	require.NoError(t, err)
	defer invDestDir.RemoveAll()
	err = invIdxResource.Download(ctx, invDestDir, func(curr *rpc.DownloadProgress) {}, downloader.GetDefaultConfig())
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid signature")
	require.False(t, invDestDir.Join("package_index.json").Exist())
	require.False(t, invDestDir.Join("package_index.json.sig").Exist())
}

func TestIndexFileName(t *testing.T) {
	tests := []struct {
		url      string
		expected string
	}{
		{url: "package_index.json", expected: "package_index.json"},
		{url: "package_index.json.gz", expected: "package_index.json"},
		{url: "package_index.tar.bz2", expected: "package_index.json"},
		// https://github.com/arduino/arduino-cli/issues/2345
		{url: "package_arduino.cc_index.json", expected: "package_arduino.cc_index.json"},
		{url: "package_arduino.cc_index.json.gz", expected: "package_arduino.cc_index.json"},
		{url: "package_arduino.cc_index.tar.bz2", expected: "package_arduino.cc_index.json"},
		{url: "http://drazzy.com/package_drazzy.com_index.json", expected: "package_drazzy.com_index.json"},
	}
	for _, tc := range tests {
		ir := IndexResource{URL: &url.URL{Path: tc.url}}
		name, err := ir.IndexFileName()
		require.NoError(t, err, fmt.Sprintf("error trying url: %v", tc))
		require.Equal(t, tc.expected, name)
	}
}
