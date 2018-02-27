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

	"github.com/bcmi-labs/arduino-cli/common"
	"github.com/bcmi-labs/arduino-cli/common/formatter/output"
	"github.com/bcmi-labs/arduino-cli/configs"
	"github.com/bcmi-labs/arduino-cli/task"
	"github.com/sirupsen/logrus"
)

// IsCached returns a bool representing if the release has already been downloaded
func IsCached(release *DownloadResource) bool {
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
//     resource -> The resource to download.
//     label -> Name used to identify the type of the Item downloaded (library, core, tool)
//     fileDownloadFilter -> (optional) is called before the default download behavior, to override the handling of
// 							 the result of the download (i.e. decide how to copy the download to the file
// 							 or do something weird during the operation)
//     progressChangedHandler -> (optional) is invoked at each progress change.
//   RETURNS:
//     error if any
func downloadRelease(item *DownloadResource, label string, fileDownloadFilter FileDownloadFilter,
	progressChangedHandler common.DownloadPackageProgressChangedHandler) error {
	if item == nil {
		return errors.New("Cannot accept nil release")
	}
	initialData, err := OpenLocalArchiveForDownload(item)
	if err != nil {
		return fmt.Errorf("Cannot get Archive file of this release : %s", err)
	}
	defer initialData.Close()

	// Filter the default file download handler with the external one, if available
	filterDefaultDownloadHandler := func() common.HandleDownloadPackageResultFunc {
		// Call the external filter first
		if fileDownloadFilter != nil {
			return func(source io.Reader, initialData *os.File, initialSize int) error {
				source, err = fileDownloadFilter(source, initialData, initialSize)
				if err != nil {
					return err
				}
				// Copy the file content
				return common.DefaultDownloadHandlerFunc(source, initialData, initialSize)
			}
		}

		return nil
	}

	err = common.DownloadPackage(item.URL, initialData,
		item.Size, filterDefaultDownloadHandler(), progressChangedHandler)
	if err != nil {
		return err
	}
	err = checkLocalArchive(item)
	if err != nil {
		return errors.New("Archive has been downloaded, but it seems corrupted. Try again to redownload it")
	}
	return nil
}

// downloadTask returns the wrapper to download something without installing it.
func downloadTask(item *DownloadResource, fileDownloadProgressHandler FileDownloadFilter, label string,
	progressChangedHandler common.DownloadPackageProgressChangedHandler) task.Wrapper {
	return task.Wrapper{
		Task: func() task.Result {
			err := downloadRelease(item, label, fileDownloadProgressHandler, progressChangedHandler)
			if err != nil {
				return task.Result{
					Error: err,
				}
			}
			return task.Result{}
		},
	}
}

// FileDownloadFilter defines a function which acts as a filter to handle the download of a file.
// It receives a source Reader, the destination File and the initial point were the file should be filled in.
// With that information it can actively act both on the download stream and on the target file.
// A typical use may include wrapping the source and return a new one, in order to intercept reading to the download stream.
type FileDownloadFilter func(source io.Reader, initialData *os.File, initialSize int) (io.Reader, error)

// decorateDefaultDownloadHandler defines an handler that is made aware of
// the progress of the ParallelDownload.
// For this to work, the handler is notified of each new download task and it
// is expected to generate a FileDownloadFilter to track down the progress of the single file.
// The starting and final moment of the whole parallel download process are also reported to the handler.
type ParallelDownloadProgressHandler interface {
	// OnNewDownloadTask is called when a new download task is added to the queue of the ParallelDownload.
	// This method is supposed to return an optional FileDownloadFilter, which is passed to the downstream
	// download logic (a typical decorator uses the source Reader to intercept the reading progress and
	// use it in some way).
	OnNewDownloadTask(fileName string, fileSize int64) FileDownloadFilter
	// OnProgressChanged is called at each download progress change, giving information for a specific
	// fileName, reporting its total fileSize and the part downloadedSoFar (both in bytes)
	OnProgressChanged(fileName string, fileSize int64, downloadedSoFar int64)
	// OnDownloadStarted is called just before the download of the multiple tasks starts
	OnDownloadStarted()
	// OnDownloadStarted is called just after the download of the multiple tasks ends
	OnDownloadFinished()
}

// ParallelDownload executes multiple releases downloads in parallel and fills properly results.
//
//   forced is used to force download if cached.
//   OkStatus is used to tell the overlying process result ("Downloaded", "Installed", etc...)
//	 An optional progressHandler can be passed in order to be notified of the status of the download.
//   DOES NOT RETURN since will append results to the provided refResults; use refResults.Results() to get them.
func ParallelDownload(items []DownloadItem, forced bool, OkStatus string, refResults *[]output.ProcessResult,
	label string, progressHandler ParallelDownloadProgressHandler) {

	// TODO (l.biava): Future improvements envision this utility as an object (say a Builder)
	// to simplify the passing of all those parameters, the progress handling closure, the outputResults
	// internally populated, etc.

	itemC := len(items)
	tasks := make(map[string]task.Wrapper, itemC)
	paths := make(map[string]string, itemC)

	logrus.Info(fmt.Sprintf("Initiating parallel download of %d tasks", itemC))

	for _, item := range items {
		cached := IsCached(item.Resource)
		releaseNotNil := item.Resource != nil
		itemName := item.Name
		if forced || releaseNotNil && (!cached || checkLocalArchive(item.Resource) != nil) {
			var fileDownloadFilter FileDownloadFilter
			// Notify the progress handlers, while retrieving the optional fileDownloadFilter
			if progressHandler != nil {
				fileDownloadFilter = progressHandler.OnNewDownloadTask(itemName, int64(item.Resource.Size))
			}
			paths[itemName], _ = ArchivePath(item.Resource) // if the release exists the archivepath always exists

			// Forward the per-file progress handler, if available
			// WARNING: This is using a closure on itemName!
			getProgressHandler := func(fileName string) common.DownloadPackageProgressChangedHandler {
				if progressHandler != nil {
					return func(fileSize int64, downloadedSoFar int64) {
						progressHandler.OnProgressChanged(fileName, fileSize, downloadedSoFar)
					}
				}
				return nil
			}

			tasks[itemName] = downloadTask(item.Resource, fileDownloadFilter, label, getProgressHandler(itemName))
		} else if !forced && releaseNotNil && cached {
			// Consider OK
			path, _ := ArchivePath(item.Resource)
			*refResults = append(*refResults, output.ProcessResult{
				ItemName: itemName,
				Status:   OkStatus,
				Path:     path,
			})
		}
	}

	if len(tasks) > 0 {
		if progressHandler != nil {
			progressHandler.OnDownloadStarted()
		}

		results := task.ExecuteParallelFromMap(tasks)

		if progressHandler != nil {
			progressHandler.OnDownloadFinished()
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

// ArchivePath returns the fullPath of the Archive of this release.
func ArchivePath(r *DownloadResource) (string, error) {
	staging, err := configs.DownloadCacheFolder(r.CachePath).Get()
	if err != nil {
		return "", err
	}
	return filepath.Join(staging, r.ArchiveFileName), nil
}

// OpenLocalArchiveForDownload open local archive if present
// used to resume download. Creates an empty file if not found.
func OpenLocalArchiveForDownload(r *DownloadResource) (*os.File, error) {
	path, err := ArchivePath(r)
	if err != nil {
		return nil, err
	}
	stats, err := os.Stat(path)
	if os.IsNotExist(err) || err == nil && stats.Size() >= r.Size {
		return os.Create(path)
	}
	return os.OpenFile(path, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
}
