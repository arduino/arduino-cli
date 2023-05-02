// This file is part of arduino-cli.
//
// Copyright 2023 ARDUINO SA (http://www.arduino.cc/)
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
	"os"
	"path"
	"path/filepath"
	"testing"

	"github.com/arduino/go-paths-helper"

	"github.com/stretchr/testify/require"
)

func TestInstallPlatform(t *testing.T) {
	t.Run("ignore __MACOSX folder", func(t *testing.T) {
		testFileName := "platform_with_root_and__MACOSX_folder.tar.bz2"
		testFilePath := filepath.Join("testdata/valid", testFileName)

		downloadDir, tempPath, destDir := paths.New(t.TempDir()), paths.New(t.TempDir()), paths.New(t.TempDir())

		// copy testfile in the download dir
		origin, err := os.ReadFile(testFilePath)
		require.NoError(t, err)
		require.NoError(t, os.WriteFile(path.Join(downloadDir.String(), testFileName), origin, 0644))

		r := &DownloadResource{
			ArchiveFileName: testFileName,
			Checksum:        "SHA-256:600ad56b6260352e0b2cee786f60749e778e179252a0594ba542f0bd1f8adee5",
			Size:            157,
		}

		require.NoError(t, r.Install(downloadDir, tempPath, destDir))
	})

	tests := []struct {
		testName         string
		downloadResource *DownloadResource
		error            string
	}{
		{
			testName: "multiple root folders not allowed",
			downloadResource: &DownloadResource{
				ArchiveFileName: "platform_with_multiple_root_folders.tar.bz2",
				Checksum:        "SHA-256:8b3fc6253c5ac2f3ba684eba0d62bb8a4ee93469fa822f81e2cd7d1b959c4044",
				Size:            148,
			},
			error: "no unique root dir in archive",
		},
		{
			testName: "root folder not present",
			downloadResource: &DownloadResource{
				ArchiveFileName: "platform_without_root_folder.tar.bz2",
				Checksum:        "SHA-256:bc00db9784e20f50d7a5fceccb6bd95ebff4a3e847aac88365b95a6851a24963",
				Size:            177,
			},
			error: "files in archive must be placed in a subdirectory",
		},
	}
	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			downloadDir, tempPath, destDir := paths.New(t.TempDir()), paths.New(t.TempDir()), paths.New(t.TempDir())
			testFileName := test.downloadResource.ArchiveFileName
			testFilePath := filepath.Join("testdata/invalid", testFileName)

			// copy testfile in the download dir
			origin, err := os.ReadFile(testFilePath)
			require.NoError(t, err)
			require.NoError(t, os.WriteFile(path.Join(downloadDir.String(), testFileName), origin, 0644))

			err = test.downloadResource.Install(downloadDir, tempPath, destDir)
			require.Error(t, err)
			require.Contains(t, err.Error(), test.error)
		})
	}
}
