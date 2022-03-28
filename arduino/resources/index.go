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
	"net/url"
	"path"
	"strings"

	"github.com/arduino/arduino-cli/arduino"
	"github.com/arduino/arduino-cli/arduino/httpclient"
	"github.com/arduino/arduino-cli/arduino/security"
	"github.com/arduino/go-paths-helper"
	"go.bug.st/downloader/v2"
)

// IndexResource is a reference to a package_index.json
type IndexResource struct {
	URL          *url.URL
	SignatureURL *url.URL
}

// Download will download the index and possibly check the signature using the Arduino's public key.
// If the file is in .gz format it will be unpacked first.
func (res *IndexResource) Download(destDir *paths.Path, downloadCB httpclient.DownloadProgressCB) error {
	// Create destination directory
	if err := destDir.MkdirAll(); err != nil {
		return &arduino.PermissionDeniedError{Message: tr("Can't create data directory %s", destDir), Cause: err}
	}

	// Create a temp dir to stage all downloads
	tmp, err := paths.MkTempDir("", "library_index_download")
	if err != nil {
		return &arduino.TempDirCreationFailedError{Cause: err}
	}
	defer tmp.RemoveAll()

	// Download index file
	indexFileName := path.Base(res.URL.Path) // == pakcage_index.json[.gz]
	tmpIndexPath := tmp.Join(indexFileName)
	if err := httpclient.DownloadFile(tmpIndexPath, res.URL.String(), tr("Downloading index: %s", indexFileName), downloadCB, nil, downloader.NoResume); err != nil {
		return &arduino.FailedDownloadError{Message: tr("Error downloading index '%s'", res.URL), Cause: err}
	}

	// Expand the index if it is compressed
	if strings.HasSuffix(indexFileName, ".gz") {
		indexFileName = strings.TrimSuffix(indexFileName, ".gz") // == pakcage_index.json
		tmpUnzippedIndexPath := tmp.Join(indexFileName)
		if err := paths.GUnzip(tmpIndexPath, tmpUnzippedIndexPath); err != nil {
			return &arduino.PermissionDeniedError{Message: tr("Error extracting %s", indexFileName), Cause: err}
		}
		tmpIndexPath = tmpUnzippedIndexPath
	}

	// Check the signature if needed
	var signaturePath, tmpSignaturePath *paths.Path
	if res.SignatureURL != nil {
		// Compose signature URL
		signatureFileName := path.Base(res.SignatureURL.Path)

		// Download signature
		signaturePath = destDir.Join(signatureFileName)
		tmpSignaturePath = tmp.Join(signatureFileName)
		if err := httpclient.DownloadFile(tmpSignaturePath, res.SignatureURL.String(), tr("Downloading index signature: %s", signatureFileName), downloadCB, nil, downloader.NoResume); err != nil {
			return &arduino.FailedDownloadError{Message: tr("Error downloading index signature '%s'", res.SignatureURL), Cause: err}
		}

		// Check signature...
		if valid, _, err := security.VerifyArduinoDetachedSignature(tmpIndexPath, tmpSignaturePath); err != nil {
			return &arduino.PermissionDeniedError{Message: tr("Error verifying signature"), Cause: err}
		} else if !valid {
			return &arduino.SignatureVerificationFailedError{File: res.URL.String()}
		}
	}

	// TODO: Implement a ResourceValidator
	// if !validate(tmpIndexPath) { return error }

	// Make a backup copy of old index and signature so the defer function can rollback in case of errors.
	indexPath := destDir.Join(indexFileName)
	oldIndex := tmp.Join("old_index")
	if indexPath.Exist() {
		if err := indexPath.CopyTo(oldIndex); err != nil {
			return &arduino.PermissionDeniedError{Message: tr("Error saving downloaded index"), Cause: err}
		}
		defer oldIndex.CopyTo(indexPath) // will silently fail in case of success
	}
	oldSignature := tmp.Join("old_signature")
	if oldSignature.Exist() {
		if err := signaturePath.CopyTo(oldSignature); err != nil {
			return &arduino.PermissionDeniedError{Message: tr("Error saving downloaded index signature"), Cause: err}
		}
		defer oldSignature.CopyTo(signaturePath) // will silently fail in case of success
	}
	if err := tmpIndexPath.CopyTo(indexPath); err != nil {
		return &arduino.PermissionDeniedError{Message: tr("Error saving downloaded index"), Cause: err}
	}
	if res.SignatureURL != nil {
		if err := tmpSignaturePath.CopyTo(signaturePath); err != nil {
			return &arduino.PermissionDeniedError{Message: tr("Error saving downloaded index signature"), Cause: err}
		}
	}
	oldIndex.Remove()
	oldSignature.Remove()
	return nil
}
