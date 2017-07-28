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

package releases

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/bcmi-labs/arduino-cli/cmd/formatter"
	"github.com/bcmi-labs/arduino-cli/cmd/output"
	"github.com/bcmi-labs/arduino-cli/common"
	"github.com/bcmi-labs/arduino-cli/task"
	pb "gopkg.in/cheggaaa/pb.v1"
)

// IsCached returns a bool representing if the release has already been downloaded
func IsCached(release Release) bool {
	stagingFolder, err := release.GetDownloadCacheFolder()
	if err != nil {
		return false
	}

	_, err = os.Stat(filepath.Join(stagingFolder, release.ArchiveName()))
	return !os.IsNotExist(err)
}

// ParseArgs parses a sequence of "item@version" tokens and returns a Name-Version slice.
//
// If version is not present it is assumed as "latest" version.
func ParseArgs(args []string) []NameVersionPair {
	ret := make([]NameVersionPair, 0, len(args))
	for _, item := range args {
		tokens := strings.SplitN(item, "@", 2)
		var version string
		if len(tokens) == 2 {
			version = tokens[1]
		} else {
			version = "latest"
		}
		ret = append(ret, NameVersionPair{
			Name:    tokens[0],
			Version: version,
		})
	}
	return ret
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

	initialData, err := item.Release.OpenLocalArchiveForDownload()
	if err != nil {
		return fmt.Errorf("Cannot get Archive file of this release : %s", err)
	}
	defer initialData.Close()
	err = common.DownloadPackage(item.Release.ArchiveURL(), fmt.Sprint(label, " ", item.Name), progBar, initialData, item.Release.ArchiveSize())
	if err != nil {
		return err
	}
	err = item.Release.CheckLocalArchive()
	if err != nil {
		return errors.New("Archive has been downloaded, but it seems corrupted. Try again to redownload it")
	}
	return nil
}

// downloadAndCache returns the wrapper to download something without installing it
func downloadAndCache(item DownloadItem, progBar *pb.ProgressBar) task.Wrapper {
	return task.Wrapper{
		Task: func() task.Result {
			err := downloadRelease(item, progBar, "library")
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

// ParallelDownload executes multiple releases downloads in parallel and fills properly results.
//
//   forced is used to force download if cached.
//   OkStatus is used to tell the overlying process result ("Downloaded", "Installed", etc...)
//   DOES NOT RETURN because modified refResults array of results using pointer provided by refResults.Results().
func ParallelDownload(items []DownloadItem, forced bool, OkStatus string, verbosity int, refResults *[]output.ProcessResult) {
	itemC := len(items)
	tasks := make(map[string]task.Wrapper, itemC)
	progressBars := make([]*pb.ProgressBar, 0, itemC)
	textMode := formatter.IsCurrentFormat("text")
	for _, item := range items {
		cached := IsCached(item.Release)
		releaseNotNil := item.Release != nil
		if forced || releaseNotNil && (!cached || item.Release.CheckLocalArchive() != nil) {
			var pBar *pb.ProgressBar
			if textMode {
				pBar = pb.StartNew(int(item.Release.ArchiveSize())).SetUnits(pb.U_BYTES).Prefix(fmt.Sprintf("%-20s", item.Name))
				progressBars = append(progressBars, pBar)
			}
			tasks[item.Name] = downloadAndCache(item, pBar)
		} else if !forced && releaseNotNil && cached {
			//Consider OK
			*refResults = append(*refResults, output.ProcessResult{
				ItemName: item.Name,
				Status:   OkStatus,
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
				})
			}
		}
	}
}
