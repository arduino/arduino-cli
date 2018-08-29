/*
 * This file is part of arduino-cli.
 *
 * Copyright 2018 ARDUINO SA (http://www.arduino.cc/)
 *
 * This software is released under the GNU General Public License version 3,
 * which covers the main part of arduino-cli.
 * The terms of this license can be found at:
 * https://www.gnu.org/licenses/gpl-3.0.en.html
 *
 * You can be released from the requirements of the above licenses by purchasing
 * a commercial license. Buying such a license is mandatory if you want to modify or
 * otherwise use the software for commercial activities involving the Arduino
 * software without disclosing the source code of your own applications. To purchase
 * a commercial license, send an email to license@arduino.cc.
 */

package resources

import (
	"fmt"
	"os"

	paths "github.com/arduino/go-paths-helper"
	"github.com/codeclysm/extract"
)

// Install installs the resource in three steps:
// - the archive is unpacked in a temporary subdir of tempPath
// - there should be only one root dir in the unpacked content
// - the only root dir is moved/renamed to/as the destination directory
// Note that tempPath and destDir must be on the same filesystem partition
// otherwise the last step will fail.
func (release *DownloadResource) Install(downloadDir, tempPath, destDir *paths.Path) error {
	// Create a temporary dir to extract package
	if err := tempPath.MkdirAll(); err != nil {
		return fmt.Errorf("creating temp dir for extraction: %s", err)
	}
	tempDir, err := tempPath.MkTempDir("package-")
	if err != nil {
		return fmt.Errorf("creating temp dir for extraction: %s", err)
	}
	defer tempDir.RemoveAll()

	// Obtain the archive path and open it
	archivePath, err := release.ArchivePath(downloadDir)
	if err != nil {
		return fmt.Errorf("getting archive path: %s", err)
	}
	file, err := os.Open(archivePath.String())
	if err != nil {
		return fmt.Errorf("opening archive file: %s", err)
	}
	defer file.Close()

	// Extract into temp directory
	if err := extract.Archive(file, tempDir.String(), nil); err != nil {
		return fmt.Errorf("extracting archive: %s", err)
	}

	// Check package content and find package root dir
	root, err := findPackageRoot(tempDir)
	if err != nil {
		return fmt.Errorf("searching package root dir: %s", err)
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
	if isdir, _ := destDir.IsDir(); isdir {
		destDir.RemoveAll()
	}

	// Move/rename the extracted root directory in the destination directory
	if err := root.Rename(destDir); err != nil {
		return fmt.Errorf("moving extracted archive to destination dir: %s", err)
	}

	// TODO
	// // Create a package file
	// if err := createPackageFile(destDir); err != nil {
	// 	return err
	// }

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
		return nil, fmt.Errorf("reading package root dir: %s", err)
	}
	var root *paths.Path
	for _, file := range files {
		if isdir, _ := file.IsDir(); !isdir {
			continue
		}
		if root == nil {
			root = file
		} else {
			return nil, fmt.Errorf("no unique root dir in archive, found '%s' and '%s'", root, file)
		}
	}
	return root, nil
}
