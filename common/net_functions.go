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

package common

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/bcmi-labs/arduino-cli/task"

	pb "gopkg.in/cheggaaa/pb.v1"
)

// DownloadPackageIndex is a function to download a generic index.
func DownloadPackageIndex(indexPathFunc func() (string, error), URL string) error {
	file, err := indexPathFunc()
	if err != nil {
		return err
	}

	req, err := http.NewRequest("GET", URL, nil)
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

	err = ioutil.WriteFile(file, content, 0666)
	if err != nil {
		return err
	}
	return nil
}

// DownloadPackage downloads a package from arduino repository, applying a label for the progress bar.
func DownloadPackage(URL string, downloadLabel string, progressBar *pb.ProgressBar, initialData *os.File, totalSize int64) error {
	client := http.DefaultClient

	if initialData == nil {
		return errors.New("Cannot fill a nil file pointer")
	}

	request, err := http.NewRequest("GET", URL, nil)
	if err != nil {
		return fmt.Errorf("Cannot create HTTP request: %s", err)
	}

	var initialSize int64
	stats, err := initialData.Stat()
	if err != nil {
		initialSize = 0
	} else {
		fileSize := stats.Size()
		if fileSize >= totalSize {
			initialSize = 0
		} else {
			initialSize = fileSize
		}
	}

	if initialSize > 0 {
		request.Header.Add("Range", fmt.Sprintf("bytes=%d-", initialSize))
	}

	response, err := client.Do(request)

	if err != nil {
		return fmt.Errorf("Cannot fetch %s. Response creation error", downloadLabel)
	} else if response.StatusCode != 200 &&
		response.StatusCode != 206 &&
		response.StatusCode != 416 {
		response.Body.Close()
		return fmt.Errorf("Cannot fetch %s. Source responded with a status %d code", downloadLabel, response.StatusCode)
	}
	defer response.Body.Close()

	source := response.Body
	if progressBar != nil {
		progressBar.Add(int(initialSize))
		source = progressBar.NewProxyReader(response.Body)
	}

	_, err = io.Copy(initialData, source)
	if err != nil {
		return fmt.Errorf("Cannot read response body %s", err)
	}
	return nil
}

// DownloadRelease downloads a generic release.
//
//   PARAMS:
//     name -> The name of the Item to download
//     release -> The release to download
//     progBar -> a progress bar, can be nil. If not nill progress is handled for that bar.
//     label -> Name used to identify the type of the Item downloaded (library, core, tool)
//   RETURNS:
//     error if any
func DownloadRelease(name string, release Release, progBar *pb.ProgressBar, label string) error {
	if release == nil {
		return errors.New("Cannot accept nil release")
	}

	initialData, err := release.OpenLocalArchiveForDownload()
	if err != nil {
		return fmt.Errorf("Cannot get Archive file of this release : %s", err)
	}
	defer initialData.Close()
	err = DownloadPackage(release.ArchiveURL(), fmt.Sprint(label, " ", name), progBar, initialData, release.ArchiveSize())
	if err != nil {
		return err
	}
	err = release.CheckLocalArchive()
	if err != nil {
		return errors.New("Archive has been downloaded, but it seems corrupted. Try again to redownload it")
	}
	return nil
}

// DownloadAndCache returns the wrapper to download something without installing it
func DownloadAndCache(name string, release Release, progBar *pb.ProgressBar) task.Wrapper {
	return task.Wrapper{
		Task: func() task.Result {
			err := DownloadRelease(name, release, progBar, "library")
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
