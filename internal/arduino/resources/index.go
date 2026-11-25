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
	"errors"
	"net/url"
	"path"
	"strings"

	"github.com/arduino/arduino-cli/commands/cmderrors"
	"github.com/arduino/arduino-cli/internal/arduino/httpclient"
	"github.com/arduino/arduino-cli/internal/arduino/security"
	"github.com/arduino/arduino-cli/internal/arduino/utils"
	"github.com/arduino/arduino-cli/internal/i18n"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	"github.com/arduino/go-paths-helper"
	"github.com/codeclysm/extract/v4"
	"github.com/sirupsen/logrus"
	"go.bug.st/downloader/v3"
)

// IndexResource is a reference to an index file URL with an optional signature.
type IndexResource struct {
	URL                          *url.URL
	SignatureURL                 *url.URL
	EnforceSignatureVerification bool
}

// IndexFileName returns the index file name as it is saved in data dir (package_xxx_index.json).
// If the URL contains a fragment (e.g. #package_foo_index.json), that fragment is used as-is as
// the local filename, allowing users to provide a meaningful name for URLs whose path provides none.
// The returned name is guaranteed to be a plain file name without any directory component.
func (res *IndexResource) IndexFileName() (string, error) {
	filename, err := res.indexFileName()
	if err != nil {
		return "", err
	}
	if err := ValidateFileName(filename); err != nil {
		return "", &cmderrors.InvalidURLError{Cause: err}
	}
	return filename, nil
}

func (res *IndexResource) indexFileName() (string, error) {
	if res.URL.Fragment != "" {
		return res.URL.Fragment, nil
	}
	filename := path.Base(res.URL.Path) // == package_index.json[.gz] || packacge_index.tar.bz2
	if filename == "." || filename == "" || filename == "/" {
		return "", &cmderrors.InvalidURLError{}
	}
	switch {
	case strings.HasSuffix(filename, ".json"):
		return filename, nil
	case strings.HasSuffix(filename, ".gz"):
		return strings.TrimSuffix(filename, ".gz"), nil
	case strings.HasSuffix(filename, ".tar.bz2"):
		return strings.TrimSuffix(filename, ".tar.bz2") + ".json", nil
	}
	return filename + ".json", nil
}

// Download will download the index and possibly check the signature using the Arduino's public key.
// If the file is in .gz format it will be unpacked first.
func (res *IndexResource) Download(ctx context.Context, destDir *paths.Path, downloadCB rpc.DownloadProgressCB, config downloader.Config) error {
	// Create destination directory
	if err := destDir.MkdirAll(); err != nil {
		return &cmderrors.PermissionDeniedError{Message: i18n.Tr("Can't create data directory %s", destDir), Cause: err}
	}

	// Create a temp dir to stage all downloads
	tmp, err := paths.MkTempDir("", "library_index_download")
	if err != nil {
		return &cmderrors.TempDirCreationFailedError{Cause: err}
	}
	defer tmp.RemoveAll()

	// Download index file
	// Strip any fragment from the request URL — fragments are client-side hints, not sent to the server.
	reqURL := *res.URL
	reqURL.Fragment = ""
	reqURL.RawFragment = ""
	downloadFileName := path.Base(res.URL.Path) // == package_index.json[.gz] || package_index.tar.bz2
	indexFileName, err := res.IndexFileName()   // == package_index.json (or fragment-specified name)
	if err != nil {
		return err
	}
	// Ensure the final destination can not escape destDir
	indexPath, err := safeJoin(destDir, indexFileName)
	if err != nil {
		return &cmderrors.InvalidURLError{Cause: err}
	}
	// Use a fixed name for the temporary download: downloadFileName comes from the URL
	// and is used only to detect the format and in progress messages.
	tmpIndexPath := tmp.Join("downloaded_index")
	config.DoNotResumeDownload = true // Disable resuming downloads for index files
	if err := httpclient.DownloadFile(ctx, tmpIndexPath, reqURL.String(), "", i18n.Tr("Downloading index: %s", downloadFileName), downloadCB, config); err != nil {
		return &cmderrors.FailedDownloadError{Message: i18n.Tr("Error downloading index '%s'", &reqURL), Cause: err}
	}

	var signaturePath, tmpSignaturePath *paths.Path
	hasSignature := false

	// Expand the index if it is compressed
	if strings.HasSuffix(downloadFileName, ".tar.bz2") {
		signatureFileName := indexFileName + ".sig"
		if signaturePath, err = safeJoin(destDir, signatureFileName); err != nil {
			return &cmderrors.InvalidURLError{Cause: err}
		}

		// .tar.bz2 archive may contain both index and signature

		// Extract archive in a tmp/archive subdirectory
		f, err := tmpIndexPath.Open()
		if err != nil {
			return &cmderrors.PermissionDeniedError{Message: i18n.Tr("Error opening %s", tmpIndexPath), Cause: err}
		}
		defer f.Close()
		tmpArchivePath := tmp.Join("archive")
		_ = tmpArchivePath.MkdirAll()
		if err := extract.Bz2(ctx, f, tmpArchivePath.String(), nil); err != nil {
			return &cmderrors.PermissionDeniedError{Message: i18n.Tr("Error extracting %s", tmpIndexPath), Cause: err}
		}

		// Look for index.json
		if tmpIndexPath, err = safeJoin(tmpArchivePath, indexFileName); err != nil {
			return &cmderrors.InvalidURLError{Cause: err}
		}
		if !tmpIndexPath.Exist() {
			return &cmderrors.NotFoundError{Message: i18n.Tr("Invalid archive: file %[1]s not found in archive %[2]s", indexFileName, tmpArchivePath.Base())}
		}

		// Look for signature
		if t, err := safeJoin(tmpArchivePath, signatureFileName); err == nil && t.Exist() {
			tmpSignaturePath = t
			hasSignature = true
		} else {
			logrus.Infof("No signature %s found in package index archive %s", signatureFileName, tmpArchivePath.Base())
		}
	} else if strings.HasSuffix(downloadFileName, ".gz") {
		tmpUnzippedIndexPath := tmp.Join("unzipped_index")
		if err := paths.GUnzip(tmpIndexPath, tmpUnzippedIndexPath); err != nil {
			return &cmderrors.PermissionDeniedError{Message: i18n.Tr("Error extracting %s", indexFileName), Cause: err}
		}
		tmpIndexPath = tmpUnzippedIndexPath
	}

	// Check the signature if needed
	if res.SignatureURL != nil {
		// Compose signature URL
		signatureFileName := path.Base(res.SignatureURL.Path)

		// Download signature
		if signaturePath, err = safeJoin(destDir, signatureFileName); err != nil {
			return &cmderrors.InvalidURLError{Cause: err}
		}
		if tmpSignaturePath, err = safeJoin(tmp, signatureFileName); err != nil {
			return &cmderrors.InvalidURLError{Cause: err}
		}
		config.DoNotResumeDownload = true // Disable resuming downloads for index files
		if err := httpclient.DownloadFile(ctx, tmpSignaturePath, res.SignatureURL.String(), "", i18n.Tr("Downloading index signature: %s", signatureFileName), downloadCB, config); err != nil {
			return &cmderrors.FailedDownloadError{Message: i18n.Tr("Error downloading index signature '%s'", res.SignatureURL), Cause: err}
		}

		hasSignature = true
	}

	if hasSignature {
		// Check signature...
		if valid, _, err := security.VerifyArduinoDetachedSignature(tmpIndexPath, tmpSignaturePath); err != nil {
			return &cmderrors.PermissionDeniedError{Message: i18n.Tr("Error verifying signature"), Cause: err}
		} else if !valid {
			return &cmderrors.SignatureVerificationFailedError{File: res.URL.String()}
		}
	} else {
		if res.EnforceSignatureVerification {
			return &cmderrors.PermissionDeniedError{Message: i18n.Tr("Error verifying signature"), Cause: errors.New(i18n.Tr("missing signature"))}
		}
	}

	// TODO: Implement a ResourceValidator
	// if !validate(tmpIndexPath) { return error }

	// Make a backup copy of old index and signature so the defer function can rollback in case of errors.
	oldIndex := tmp.Join("old_index")
	if indexPath.Exist() {
		if err := indexPath.CopyTo(oldIndex); err != nil {
			return &cmderrors.PermissionDeniedError{Message: i18n.Tr("Error saving downloaded index"), Cause: err}
		}
		defer oldIndex.CopyTo(indexPath) // will silently fail in case of success
	}
	oldSignature := tmp.Join("old_signature")
	if oldSignature.Exist() {
		if err := signaturePath.CopyTo(oldSignature); err != nil {
			return &cmderrors.PermissionDeniedError{Message: i18n.Tr("Error saving downloaded index signature"), Cause: err}
		}
		defer oldSignature.CopyTo(signaturePath) // will silently fail in case of success
	}
	if err := tmpIndexPath.CopyTo(indexPath); err != nil {
		return &cmderrors.PermissionDeniedError{Message: i18n.Tr("Error saving downloaded index"), Cause: err}
	}
	if err := utils.SyncFile(indexPath); err != nil {
		return &cmderrors.PermissionDeniedError{Message: i18n.Tr("Error saving downloaded index"), Cause: err}
	}
	if hasSignature {
		if err := tmpSignaturePath.CopyTo(signaturePath); err != nil {
			return &cmderrors.PermissionDeniedError{Message: i18n.Tr("Error saving downloaded index signature"), Cause: err}
		}
		if err := utils.SyncFile(signaturePath); err != nil {
			return &cmderrors.PermissionDeniedError{Message: i18n.Tr("Error saving downloaded index signature"), Cause: err}
		}
	}
	_ = oldIndex.Remove()
	_ = oldSignature.Remove()
	return nil
}
