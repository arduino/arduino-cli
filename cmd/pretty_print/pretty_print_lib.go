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
	"github.com/bcmi-labs/arduino-cli/libraries"
	"github.com/bcmi-labs/arduino-cli/task"
)

// DownloadLibFileIndex shows info regarding the download of a missing (or corrupted) file index of libraries.
func DownloadLibFileIndex() task.Wrapper {
	return DownloadFileIndex(libraries.DownloadLibrariesFile)
}

// CorruptedLibIndexFix pretty prints messages regarding corrupted index fixes of libraries.
func CorruptedLibIndexFix(index libraries.Index, verbosity int) (libraries.StatusContext, error) {
	downloadTask := DownloadLibFileIndex()
	parseTask := libIndexParse(index, verbosity)

	result := corruptedIndexFixResults(downloadTask, parseTask, verbosity)
	ret, _ := result[1].Result.(libraries.StatusContext)
	err := result[1].Error
	return ret, err
}

// libIndexParse pretty prints info about parsing an index file of libraries.
func libIndexParse(index libraries.Index, verbosity int) task.Wrapper {
	ret := indexParseWrapperSkeleton()
	ret.Task = func() task.Result {
		err := libraries.LoadIndex(&index)
		if err != nil {
			return task.Result{
				Error: err,
			}
		}
		return task.Result{
			Result: index.CreateStatusContext(),
		}
	}
	return ret
}
