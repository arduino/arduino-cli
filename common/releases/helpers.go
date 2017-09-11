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

package releases

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/bcmi-labs/arduino-cli/cmd/formatter"
	"github.com/bcmi-labs/arduino-cli/cmd/output"
	"github.com/bcmi-labs/arduino-cli/common"
	"github.com/bcmi-labs/arduino-cli/task"
	pb "gopkg.in/cheggaaa/pb.v1"
)

// IsCached returns a bool representing if the release has already been downloaded
func IsCached(release Release) bool {
	archivePath, err := ArchivePath(release)
	if err != nil {
		return false
	}

	_, err = os.Stat(archivePath)
	return !os.IsNotExist(err)
}

// downloadRelease downloads a generic release.
//
//   PARAMS:
//     name -> The name of the Item to download.
//     release -> The release to download.
//     progBar -> a progress bar, can be nil. If not nill progress is handled for that bar.
//     label -> Name used to identify the type of the Item downloaded (library, core, tool)
//   RETURNS:
//     error if any
func downloadRelease(item DownloadItem, progBar *pb.ProgressBar, label string) error {
	if item.Release == nil {
		return errors.New("Cannot accept nil release")
	}
	initialData, err := OpenLocalArchiveForDownload(item.Release)
	if err != nil {
		return fmt.Errorf("Cannot get Archive file of this release : %s", err)
	}
	defer initialData.Close()
	// puts the progress bar
	err = common.DownloadPackage(item.Release.ArchiveURL(), initialData,
		item.Release.ArchiveSize(), handleWithProgressBarFunc(progBar))
	if err != nil {
		return err
	}
	err = checkLocalArchive(item.Release)
	if err != nil {
		return errors.New("Archive has been downloaded, but it seems corrupted. Try again to redownload it")
	}
	return nil
}

// downloadTask returns the wrapper to download something without installing it.
func downloadTask(item DownloadItem, progBar *pb.ProgressBar, label string) task.Wrapper {
	return task.Wrapper{
		Task: func() task.Result {
			err := downloadRelease(item, progBar, label)
			if err != nil {
				return task.Result{
					Error: err,
				}
			}
			return task.Result{}
		},
	}
}

// ParallelDownload executes multiple releases downloads in parallel and fills properly results.
//
//   forced is used to force download if cached.
//   OkStatus is used to tell the overlying process result ("Downloaded", "Installed", etc...)
//   DOES NOT RETURN because modified refResults array of results using pointer provided by refResults.Results().
func ParallelDownload(items []DownloadItem, forced bool, OkStatus string, verbosity int, refResults *[]output.ProcessResult, label string) {
	itemC := len(items)
	tasks := make(map[string]task.Wrapper, itemC)
	paths := make(map[string]string, itemC)

	textMode := formatter.IsCurrentFormat("text")

	var progressBars []*pb.ProgressBar
	if textMode {
		progressBars = make([]*pb.ProgressBar, 0, itemC)
	}

	for _, item := range items {
		cached := IsCached(item.Release)
		releaseNotNil := item.Release != nil
		if forced || releaseNotNil && (!cached || checkLocalArchive(item.Release) != nil) {
			var pBar *pb.ProgressBar
			if textMode {
				pBar = pb.StartNew(int(item.Release.ArchiveSize())).SetUnits(pb.U_BYTES).Prefix(fmt.Sprintf("%-20s", item.Name))
				progressBars = append(progressBars, pBar)
			}
			paths[item.Name], _ = ArchivePath(item.Release) //if the release exists the archivepath always exists
			tasks[item.Name] = downloadTask(item, pBar, label)
		} else if !forced && releaseNotNil && cached {
			//Consider OK
			path, _ := ArchivePath(item.Release)
			*refResults = append(*refResults, output.ProcessResult{
				ItemName: item.Name,
				Status:   OkStatus,
				Path:     path,
			})
		}
	}

	if len(tasks) > 0 {
		var pool *pb.Pool
		if textMode {
			pool, _ = pb.StartPool(progressBars...)
		}

		results := task.ExecuteParallelFromMap(tasks, verbosity)

		if textMode {
			pool.Stop()
		}

		for name, result := range results {
			if result.Error != nil {
				*refResults = append(*refResults, output.ProcessResult{
					ItemName: name,
					Error:    result.Error.Error(),
				})
			} else {
				*refResults = append(*refResults, output.ProcessResult{
					ItemName: name,
					Status:   OkStatus,
					Path:     paths[name],
				})
			}
		}
	}
}

func handleWithProgressBarFunc(progBar *pb.ProgressBar) func(io.Reader, *os.File, int) error {
	if progBar == nil {
		return nil
	}
	return func(source io.Reader, initialData *os.File, initialSize int) error {
		progBar.Add(int(initialSize))
		source = progBar.NewProxyReader(source)

		_, err := io.Copy(initialData, source)
		if err != nil {
			return fmt.Errorf("Cannot read response body %s", err)
		}
		return nil
	}
}

// ArchivePath returns the fullPath of the Archive of this release.
func ArchivePath(release Release) (string, error) {
	staging, err := release.GetDownloadCacheFolder()
	if err != nil {
		return "", err
	}
	return filepath.Join(staging, release.ArchiveName()), nil
}

// OpenLocalArchiveForDownload reads the data from the local archive if present,
// and returns the []byte of the file content. Used by resume Download.
// Creates an empty file if not found.
func OpenLocalArchiveForDownload(r Release) (*os.File, error) {
	path, err := ArchivePath(r)
	if err != nil {
		return nil, err
	}
	stats, err := os.Stat(path)
	if os.IsNotExist(err) || err == nil && stats.Size() >= r.ArchiveSize() {
		return os.Create(path)
	}
	return os.OpenFile(path, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
}
