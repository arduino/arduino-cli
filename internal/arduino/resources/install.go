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
	"os"

	"github.com/arduino/arduino-cli/internal/i18n"
	paths "github.com/arduino/go-paths-helper"
	"github.com/codeclysm/extract/v3"
	"go.bug.st/cleanup"
)

// Install installs the resource in three steps:
// - the archive is unpacked in a temporary subdir of tempPath
// - there should be only one root dir in the unpacked content
// - the only root dir is moved/renamed to/as the destination directory
// Note that tempPath and destDir must be on the same filesystem partition
// otherwise the last step will fail.
func (release *DownloadResource) Install(downloadDir, tempPath, destDir *paths.Path) error {
	// Check the integrity of the package
	if ok, err := release.TestLocalArchiveIntegrity(downloadDir); err != nil {
		return errors.New(i18n.Tr("testing local archive integrity: %s", err))
	} else if !ok {
		return errors.New(i18n.Tr("checking local archive integrity"))
	}

	// Create a temporary dir to extract package
	if err := tempPath.MkdirAll(); err != nil {
		return errors.New(i18n.Tr("creating temp dir for extraction: %s", err))
	}
	tempDir, err := tempPath.MkTempDir("package-")
	if err != nil {
		return errors.New(i18n.Tr("creating temp dir for extraction: %s", err))
	}
	defer tempDir.RemoveAll()

	// Obtain the archive path and open it
	archivePath, err := release.ArchivePath(downloadDir)
	if err != nil {
		return errors.New(i18n.Tr("getting archive path: %s", err))
	}
	file, err := os.Open(archivePath.String())
	if err != nil {
		return errors.New(i18n.Tr("opening archive file: %s", err))
	}
	defer file.Close()

	// Extract into temp directory
	ctx, cancel := cleanup.InterruptableContext(context.Background())
	defer cancel()
	if err := extract.Archive(ctx, file, tempDir.String(), nil); err != nil {
		return errors.New(i18n.Tr("extracting archive: %s", err))
	}

	// Check package content and find package root dir
	root, err := findPackageRoot(tempDir)
	if err != nil {
		return errors.New(i18n.Tr("searching package root dir: %s", err))
	}

	// Ensure container dir exists
	destDirParent := destDir.Parent()
	if err := destDirParent.MkdirAll(); err != nil {
		return err
	}
	defer func() {
		if empty, err := IsDirEmpty(destDirParent); err == nil && empty {
			destDirParent.RemoveAll()
		}
	}()

	// If the destination dir already exists remove it
	if destDir.IsDir() {
		destDir.RemoveAll()
	}

	// Move/rename the extracted root directory in the destination directory
	if err := root.Rename(destDir); err != nil {
		// Copy the extracted root directory to the destination directory, if move failed
		if err := root.CopyDirTo(destDir); err != nil {
			return errors.New(i18n.Tr("moving extracted archive to destination dir: %s", err))
		}
	}

	return nil
}

// IsDirEmpty returns true if the directory specified by path is empty.
func IsDirEmpty(path *paths.Path) (bool, error) {
	files, err := path.ReadDir()
	if err != nil {
		return false, err
	}
	return len(files) == 0, nil
}

func findPackageRoot(parent *paths.Path) (*paths.Path, error) {
	files, err := parent.ReadDir()
	if err != nil {
		return nil, errors.New(i18n.Tr("reading package root dir: %s", err))
	}

	files.FilterDirs()
	files.FilterOutPrefix("__MACOSX")

	if len(files) == 0 {
		return nil, errors.New(i18n.Tr("files in archive must be placed in a subdirectory"))
	}
	if len(files) > 1 {
		return nil, errors.New(i18n.Tr("no unique root dir in archive, found '%[1]s' and '%[2]s'", files[0], files[1]))
	}

	return files[0], nil
}
