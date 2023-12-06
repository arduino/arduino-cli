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
	"bytes"
	"crypto"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"hash"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/arduino/arduino-cli/i18n"
	paths "github.com/arduino/go-paths-helper"
)

var tr = i18n.Tr

// TestLocalArchiveChecksum test if the checksum of the local archive match the checksum of the DownloadResource
func (r *DownloadResource) TestLocalArchiveChecksum(downloadDir *paths.Path) (bool, error) {
	if r.Checksum == "" {
		return false, fmt.Errorf(tr("missing checksum for: %s"), r.ArchiveFileName)
	}
	split := strings.SplitN(r.Checksum, ":", 2)
	if len(split) != 2 {
		return false, fmt.Errorf(tr("invalid checksum format: %s"), r.Checksum)
	}
	digest, err := hex.DecodeString(split[1])
	if err != nil {
		return false, fmt.Errorf(tr("invalid hash '%[1]s': %[2]s"), split[1], err)
	}

	// names based on: https://docs.oracle.com/javase/8/docs/technotes/guides/security/StandardNames.html#MessageDigest
	var algo hash.Hash
	switch split[0] {
	case "SHA-256":
		algo = crypto.SHA256.New()
	case "SHA-1":
		algo = crypto.SHA1.New()
	case "MD5":
		algo = crypto.MD5.New()
	default:
		return false, fmt.Errorf(tr("unsupported hash algorithm: %s"), split[0])
	}

	filePath, err := r.ArchivePath(downloadDir)
	if err != nil {
		return false, fmt.Errorf(tr("getting archive path: %s"), err)
	}

	file, err := os.Open(filePath.String())
	if err != nil {
		return false, fmt.Errorf(tr("opening archive file: %s"), err)
	}
	defer file.Close()
	if _, err := io.Copy(algo, file); err != nil {
		return false, fmt.Errorf(tr("computing hash: %s"), err)
	}

	if !bytes.Equal(algo.Sum(nil), digest) {
		return false, fmt.Errorf(tr("archive hash differs from hash in index"))
	}

	return true, nil
}

// TestLocalArchiveSize test if the local archive size match the DownloadResource size
func (r *DownloadResource) TestLocalArchiveSize(downloadDir *paths.Path) (bool, error) {
	filePath, err := r.ArchivePath(downloadDir)
	if err != nil {
		return false, fmt.Errorf(tr("getting archive path: %s"), err)
	}
	info, err := filePath.Stat()
	if err != nil {
		return false, fmt.Errorf(tr("getting archive info: %s"), err)
	}
	if info.Size() != r.Size {
		return false, fmt.Errorf("%s: %d != %d", tr("fetched archive size differs from size specified in index"), info.Size(), r.Size)
	}

	return true, nil
}

// TestLocalArchiveIntegrity checks for integrity of the local archive.
func (r *DownloadResource) TestLocalArchiveIntegrity(downloadDir *paths.Path) (bool, error) {
	if cached, err := r.IsCached(downloadDir); err != nil {
		return false, fmt.Errorf(tr("testing if archive is cached: %s"), err)
	} else if !cached {
		return false, nil
	}

	if ok, err := r.TestLocalArchiveSize(downloadDir); err != nil {
		return false, fmt.Errorf(tr("testing archive size: %s"), err)
	} else if !ok {
		return false, nil
	}

	ok, err := r.TestLocalArchiveChecksum(downloadDir)
	if err != nil {
		return false, fmt.Errorf(tr("testing archive checksum: %s"), err)
	}
	return ok, nil
}

const (
	filePermissions = 0644
	packageFileName = "package.json"
)

type packageFile struct {
	Checksum string `json:"checksum"`
}

func computeDirChecksum(root string) (string, error) {
	hash := sha256.New()
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || (info.Name() == packageFileName && filepath.Dir(path) == root) {
			return nil
		}
		f, err := os.Open(path)
		if err != nil {
			return nil
		}
		defer f.Close()
		if _, err := io.Copy(hash, f); err != nil {
			return fmt.Errorf(tr("failed to compute hash of file \"%s\""), info.Name())
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

// CheckDirChecksum reads checksum from the package.json and compares it with a recomputed value.
func CheckDirChecksum(root string) (bool, error) {
	packageJSON, err := os.ReadFile(filepath.Join(root, packageFileName))
	if err != nil {
		return false, err
	}
	var file packageFile
	json.Unmarshal(packageJSON, &file)
	checksum, err := computeDirChecksum(root)
	if err != nil {
		return false, err
	}
	if file.Checksum != checksum {
		return false, fmt.Errorf(tr("Checksum differs from checksum in package.json"))
	}

	return true, nil
}
