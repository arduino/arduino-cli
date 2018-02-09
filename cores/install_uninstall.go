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
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/bcmi-labs/arduino-cli/common"
	"github.com/bcmi-labs/arduino-cli/common/releases"
	"github.com/codeclysm/extract"
)

// DirPermissions is the default permission for create directories.
// respects umask on linux.
var DirPermissions os.FileMode = 0777

// Install installs a specific release of a core.
func Install(packager, arch string, release releases.Release) error {
	if release == nil {
		return errors.New("Not existing version of the core")
	}

	arduinoFolder, err := common.ArduinoHomeFolder.Get()
	if err != nil {
		return err
	}
	tempFolder := filepath.Join(arduinoFolder, "tmp", "packages",
		fmt.Sprintf("core-%d", time.Now().Unix()))

	coresFolder, err := common.CoresFolder(packager).Get()
	if err != nil {
		return err
	}

	cacheFilePath, err := releases.ArchivePath(release)
	if err != nil {
		return err
	}

	destCoresDirParent := filepath.Join(coresFolder, arch)
	destCoresDir := filepath.Join(destCoresDirParent, release.VersionName())

	defer func() {
		//cleaning empty directories
		if empty, _ := IsDirEmpty(destCoresDir); empty {
			os.RemoveAll(destCoresDir)
		}
		if empty, _ := IsDirEmpty(destCoresDirParent); empty {
			os.RemoveAll(destCoresDirParent)
		}
	}()

	err = os.MkdirAll(tempFolder, DirPermissions)
	if err != nil {
		return err
	}
	defer os.RemoveAll(tempFolder)

	err = os.MkdirAll(destCoresDirParent, DirPermissions)
	if err != nil {
		return err
	}

	file, err := os.Open(cacheFilePath)
	if err != nil {
		return err
	}
	defer file.Close()

	err = extract.Archive(file, tempFolder, nil)
	if err != nil {
		return err
	}

	root := coreRealRoot(tempFolder)
	if root == "invalid" {
		return errors.New("invalid archive structure")
	}

	err = os.Rename(root, destCoresDir)
	if err != nil {
		return err
	}

	err = createPackageFile(destCoresDir)
	if err != nil {
		return err
	}

	return nil
}

// InstallTool installs a specific release of a tool.
func InstallTool(packager, name string, release releases.Release) error {
	if release == nil {
		return errors.New("Not existing version of the tool")
	}

	arduinoFolder, err := common.ArduinoHomeFolder.Get()
	if err != nil {
		return err
	}
	tempFolder := filepath.Join(arduinoFolder, "tmp", "tools",
		fmt.Sprintf("tool-%d", time.Now().Unix()))

	toolsFolder, err := common.ToolsFolder(packager).Get()
	if err != nil {
		return err
	}

	cacheFilePath, err := releases.ArchivePath(release)
	if err != nil {
		return err
	}

	destToolsDirParent := filepath.Join(toolsFolder, name)
	destToolsDir := filepath.Join(destToolsDirParent, release.VersionName())

	defer func() {
		//cleaning empty directories
		if empty, _ := IsDirEmpty(destToolsDir); empty {
			os.RemoveAll(destToolsDir)
		}
		if empty, _ := IsDirEmpty(destToolsDirParent); empty {
			os.RemoveAll(destToolsDirParent)
		}
	}()

	err = os.MkdirAll(tempFolder, DirPermissions)
	if err != nil {
		return err
	}
	defer os.RemoveAll(tempFolder)

	err = os.MkdirAll(destToolsDirParent, DirPermissions)
	if err != nil {
		return err
	}

	file, err := os.Open(cacheFilePath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Ignore the top level directory inside archives. E.g. not "avr/bin/avr-g++"", but "bin/avr-g++"".
	var shift = func(path string) string {
		parts := strings.Split(path, string(filepath.Separator))
		parts = parts[1:]
		return strings.Join(parts, string(filepath.Separator))
	}
	err = extract.Archive(file, tempFolder, shift)
	if err != nil {
		return err
	}

	root := toolRealRoot(tempFolder)

	err = os.Rename(root, destToolsDir)
	if err != nil {
		return err
	}

	err = createPackageFile(destToolsDir)
	if err != nil {
		return err
	}

	return nil
}

// IsDirEmpty returns if the directory specified by path is empty,
// and an error if occurred.
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

func coreRealRoot(root string) string {
	realRoot := "invalid"
	filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if strings.HasSuffix(path, "platform.txt") {
			realRoot = filepath.Dir(path)
			return errors.New("stopped, ok") //error put to stop the search of the root
		}
		return nil
	})
	return realRoot
}

func toolRealRoot(root string) string {
	realRoot := root
	filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil || !info.IsDir() {
			return nil //ignore this step
		}
		dir, err := os.Open(path)
		if err != nil {
			return nil
		}
		defer dir.Close()
		_, err = dir.Readdir(3)
		if err == io.EOF { // read 3 files failed with EOF, dir has 2 files or more.
			realRoot = path
			return errors.New("stopped, ok") //error put to stop the search of the root
		}
		return nil
	})
	return realRoot
}
