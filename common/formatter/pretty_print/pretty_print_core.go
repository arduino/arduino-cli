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

package prettyPrints

import (
	"github.com/bcmi-labs/arduino-cli/cores"
	"github.com/bcmi-labs/arduino-cli/cores/packageindex"
	"github.com/bcmi-labs/arduino-cli/task"
)

// DownloadCoreFileIndex shows info regarding the download of a missing (or corrupted) file index of core packages.
func DownloadCoreFileIndex() task.Wrapper {
	return DownloadFileIndex(packageindex.DownloadPackagesFile)
}

// CorruptedCoreIndexFix pretty prints messages regarding corrupted index fixes of cores.
func CorruptedCoreIndexFix(index packageindex.Index) (cores.StatusContext, error) {
	downloadTask := DownloadCoreFileIndex()
	parseTask := coreIndexParse(index)

	result := corruptedIndexFixResults(downloadTask, parseTask)

	return result[1].Result.(cores.StatusContext), result[1].Error
}

// coreIndexParse pretty prints info about parsing an index file of libraries.
func coreIndexParse(index packageindex.Index) task.Wrapper {
	ret := indexParseWrapperSkeleton()
	ret.Task = task.Task(func() task.Result {
		err := packageindex.LoadIndex(&index) // I try again
		status := index.CreateStatusContext()
		return task.Result{
			Result: status,
			Error:  err,
		}
	})
	return ret
}
