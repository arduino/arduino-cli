/*
 * This file is part of arduino-cli.
 *
 * arduino-cli is free software; you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation; either version 2 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program; if not, write to the Free Software
 * Foundation, Inc., 51 Franklin St, Fifth Floor, Boston, MA  02110-1301  USA
 *
 * As a special exception, you may use this file as part of a free software
 * library without restriction.  Specifically, if other files instantiate
 * templates or use macros or inline functions from this file, or you compile
 * this file and link it with other files to produce an executable, this
 * file does not by itself cause the resulting executable to be covered by
 * the GNU General Public License.  This exception does not however
 * invalidate any other reasons why the executable file might be covered by
 * the GNU General Public License.
 *
 * Copyright 2017 ARDUINO AG (http://www.arduino.cc/)
 */

package cores

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/bcmi-labs/arduino-cli/common/releases"
	"github.com/bcmi-labs/arduino-cli/configs"
	"github.com/codeclysm/extract"
)

// installResource installs the specified resource in three steps:
// - the archive is unpacked in a temporary subfolder of tempPath
// - there should be only one root folder in the unpacked content
// - the only root folder is moved/renamed to/as the destination directory
// Note that tempPath and destDir must be on the same filesystem partition
// otherwise the last step will fail.
func installResource(tempPath string, destDir string, release *releases.DownloadResource) error {
	// Create a temporary folder to extract package
	if err := os.MkdirAll(tempPath, 0777); err != nil {
		return fmt.Errorf("creating temp dir for extraction: %s", err)
	}
	tempDir, err := ioutil.TempDir(tempPath, "package-")
	if err != nil {
		return fmt.Errorf("creating temp dir for extraction: %s", err)
	}
	defer os.RemoveAll(tempDir)

	// Obtain the archive path and open it
	archivePath, err := release.ArchivePath()
	if err != nil {
		return fmt.Errorf("getting archive path: %s", err)
	}
	file, err := os.Open(archivePath)
	if err != nil {
		return fmt.Errorf("opening archive file: %s", err)
	}
	defer file.Close()

	// Extract into temp directory
	if err := extract.Archive(file, tempDir, nil); err != nil {
		return fmt.Errorf("extracting archive: %s", err)
	}

	// Check package content and find package root dir
	root, err := findPackageRoot(tempDir)
	if err != nil {
		return fmt.Errorf("searching package root dir: %s", err)
	}

	// Ensure container dir exists
	destDirParent := filepath.Dir(destDir)
	if err := os.MkdirAll(destDirParent, 0777); err != nil {
		return err
	}
	defer func() {
		if empty, err := IsDirEmpty(destDirParent); err == nil && empty {
			os.RemoveAll(destDirParent)
		}
	}()

	// Move/rename the extracted root directory in the destination directory
	if err := os.Rename(root, destDir); err != nil {
		return err
	}

	// Create a package file
	if err := createPackageFile(destDir); err != nil {
		return err
	}

	return nil
}

// InstallPlatform installs a specific release of a platform.
func InstallPlatform(platformRelease *PlatformRelease) error {
	coreDir, err := configs.CoresFolder(platformRelease.Platform.Package.Name).Get()
	if err != nil {
		return fmt.Errorf("getting platforms dir: %s", err)
	}

	dataDir, err := configs.ArduinoDataFolder.Get()
	if err != nil {
		return fmt.Errorf("getting data dir: %s", err)
	}

	return installResource(
		filepath.Join(dataDir, "tmp"),
		filepath.Join(coreDir, platformRelease.Platform.Architecture, platformRelease.Version),
		platformRelease.Resource)
}

// InstallTool installs a specific release of a tool.
func InstallTool(toolRelease *ToolRelease) error {
	toolResource := toolRelease.GetCompatibleFlavour()
	if toolResource == nil {
		return fmt.Errorf("no compatible version of %s tools found for the current os", toolRelease.Tool.Name)
	}

	toolDir, err := configs.ToolsFolder(toolRelease.Tool.Package.Name).Get()
	if err != nil {
		return fmt.Errorf("gettin tools dir: %s", err)
	}

	dataDir, err := configs.ArduinoDataFolder.Get()
	if err != nil {
		return fmt.Errorf("getting data dir: %s", err)
	}

	return installResource(
		filepath.Join(dataDir, "tmp"),
		filepath.Join(toolDir, toolRelease.Tool.Name, toolRelease.Version),
		toolResource)
}

// IsDirEmpty returns true if the directory specified by path is empty.
func IsDirEmpty(path string) (bool, error) {
	f, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer f.Close()

	// read in ONLY one file
	_, err = f.Readdir(1)

	// and if the file is EOF... well, the dir is empty.
	if err == io.EOF {
		return true, nil
	}
	return false, err
}

func findPackageRoot(parent string) (string, error) {
	files, err := ioutil.ReadDir(parent)
	if err != nil {
		return "", fmt.Errorf("reading package root dir: %s", err)
	}
	root := ""
	for _, fileInfo := range files {
		if !fileInfo.IsDir() {
			continue
		}
		if root == "" {
			root = fileInfo.Name()
		} else {
			return "", fmt.Errorf("no unique root dir in archive, found '%s' and '%s'", root, fileInfo.Name())
		}
	}
	return filepath.Join(parent, root), nil
}
