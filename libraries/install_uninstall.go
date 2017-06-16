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
 * Copyright 2017 BCMI LABS SA (http://www.arduino.cc/)
 */

package libraries

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/bcmi-labs/arduino-cli/common"
)

var install func(*zip.Reader, string) error = common.Unzip

// Uninstall a library means remove its directory.
var Uninstall func(libraryFolder string) error = os.RemoveAll

// DownloadAndInstall downloads a library and installs it to its specified location.
func DownloadAndInstall(library *Library) error {
	libFolder, err := common.GetDefaultLibFolder()
	if err != nil {
		return fmt.Errorf("Cannot get Lib destination directory")
	}

	stagingFolder, err := getDownloadCacheFolder()
	cacheFilePath := filepath.Join(stagingFolder, fmt.Sprintf("%s-%s.zip", library.Name, library.Latest.Version))

	var zipArchive *zip.Reader

	_, err = os.Stat(cacheFilePath)
	if os.IsNotExist(err) {
		zipArchive, err = DownloadAndCache(library)
		if err != nil {
			return err
		}
	} else {
		fmt.Printf("%s library found in cache downloads ... using cached zip archive\n", library.Name)
		content, err := ioutil.ReadFile(cacheFilePath)
		if err != nil {
			return err
		}

		zipArchive, err = zip.NewReader(bytes.NewReader(content), int64(len(content)))
		if err != nil {
			return err
		}
	}

	err = install(zipArchive, libFolder)
	if err != nil {
		return err
	}

	//add postinstall here? for verbose maybe

	return nil
}

// prepareInstall move a downloaded library to a cache folder, before installation.
func prepareInstall(library *Library, body []byte) (*zip.Reader, error) {
	reader := bytes.NewReader(body)

	archive, err := zip.NewReader(reader, int64(reader.Len()))
	if err != nil {
		return nil, fmt.Errorf("Cannot read downloaded archive")
	}

	// if I can read the archive I save it to staging folder.
	stagingFolder, err := getDownloadCacheFolder()
	if err != nil {
		return nil, fmt.Errorf("Cannot get download cache folder")
	}

	err = ioutil.WriteFile(filepath.Join(stagingFolder, fmt.Sprintf("%s-%s.zip", library.Name, library.Latest.Version)), body, 0666)
	if err != nil {
		return nil, fmt.Errorf("Cannot write download to cache folder, %s", err.Error())
	}
	return archive, nil
}

// getLibFolder returns the destination folder of the downloaded specified library.
// It creates the folder if does not find it.
func getLibFolder(library *Library) (string, error) {
	baseFolder, err := common.GetDefaultLibFolder()
	if err != nil {
		return "", err
	}

	libFolder := filepath.Join(baseFolder, fmt.Sprintf("%s-%s", library.Name, library.Latest.Version))
	return common.GetFolder(libFolder, "library")
}
