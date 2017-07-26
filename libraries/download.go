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
			err := common.DownloadRelease(library.Name, library.GetVersion(version), progBar, "library")
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

// downloadLatest downloads Latest version of a library.
func downloadLatest(library *Library, progBar *pb.ProgressBar, label string) ([]byte, error) {
	return nil, common.DownloadRelease(library.Name, library.GetVersion(library.latestVersion()), progBar, "library")
}

// DownloadLibrariesFile downloads the lib file from arduino repository.
func DownloadLibrariesFile() error {
	return common.DownloadPackageIndexFunc(IndexPath, LibraryIndexURL)
}

// getDownloadCacheFolder gets the folder where temp installs are stored
// until installation complete (libraries).
func getDownloadCacheFolder() (string, error) {
	return common.GetDownloadCacheFolder("libraries")
}
