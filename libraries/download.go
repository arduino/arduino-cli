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
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"

	"errors"

	"github.com/bcmi-labs/arduino-cli/common"
	"github.com/bcmi-labs/arduino-cli/task"
	"gopkg.in/cheggaaa/pb.v1"
)

const (
	// LibraryIndexURL is the URL where to get library index.
	LibraryIndexURL string = "http://downloads.arduino.cc/libraries/library_index.json"
)

// DownloadAndCache downloads a library without installing it
func DownloadAndCache(library *Library, progBar *pb.ProgressBar, version string) task.Wrapper {
	return task.Wrapper{
		Task: func() task.Result {
			err := downloadRelease(library, progBar, version)
			if err != nil {
				return task.Result{
					Result: nil,
					Error:  err,
				}
			}
			return task.Result{
				Result: nil,
				Error:  nil,
			}
		},
	}
}

// DownloadLatest downloads Latest version of a library.
func downloadLatest(library *Library, progBar *pb.ProgressBar) ([]byte, error) {
	return nil, downloadRelease(library, progBar, library.latestVersion())
}

func downloadRelease(library *Library, progBar *pb.ProgressBar, version string) error {
	release := library.GetVersion(version)
	if release == nil {
		return errors.New("Invalid version number")
	}
	initialData, err := release.OpenLocalArchiveForDownload()
	if err != nil {
		return fmt.Errorf("Cannot get Archive file of this release : %s", err)
	}
	defer initialData.Close()
	err = common.DownloadPackage(release.URL, fmt.Sprintf("library %s", library.Name), progBar, initialData, release.Size)
	if err != nil {
		return err
	}
	err = release.CheckLocalArchive()
	if err != nil {
		return errors.New("Archive has been downloaded, but it seems corrupted")
	}
	return nil
}

// DownloadLibrariesFile downloads the lib file from arduino repository.
func DownloadLibrariesFile() error {
	libFile, err := IndexPath()
	if err != nil {
		return err
	}

	req, err := http.NewRequest("GET", LibraryIndexURL, nil)
	if err != nil {
		return err
	}

	client := http.DefaultClient
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(libFile, content, 0666)
	if err != nil {
		return err
	}
	return nil
}

// getDownloadCacheFolder gets the folder where temp installs are stored until installation complete (libraries).
func getDownloadCacheFolder() (string, error) {
	libFolder, err := common.GetDefaultArduinoFolder()
	if err != nil {
		return "", err
	}

	stagingFolder := filepath.Join(libFolder, "staging", "libraries")
	return common.GetFolder(stagingFolder, "libraries cache")
}
