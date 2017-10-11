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
	"fmt"
	"strings"

	"github.com/bcmi-labs/arduino-cli/cmd/formatter"
	"github.com/bcmi-labs/arduino-cli/task"
)

// actionOnItems pretty prints info about an action on one or more items.
func actionOnItems(itemPluralName string, actionPastParticiple string, itemOK []string, itemFails map[string]string) {
	msg := ""
	if len(itemFails) > 0 {
		if len(itemOK) > 0 {
			msg += fmt.Sprintf("The following %s were succesfully %s:\n", itemPluralName, actionPastParticiple)
			msg += fmt.Sprintln(strings.Join(itemOK, " "))
			msg += "However, t"
		} else { //UGLYYYY but it works
			msg += "T"
		}
		msg += fmt.Sprintf("he the following %s were not %s and failed :", itemPluralName, actionPastParticiple)
		for item, failure := range itemFails {
			msg += fmt.Sprintf("%-10s -%s\n", item, failure)
		}
	} else {
		msg += fmt.Sprintf("All %s successfully installed\n", itemPluralName)
	}
	formatter.Print(msg)
}

// DownloadFileIndex shows info regarding the download of a missing (or corrupted) file index.
// uses DownloadFunc to download the file.
func DownloadFileIndex(downloadFunc func() error) task.Wrapper {
	return task.Wrapper{
		BeforeMessage: "Downloading from download.arduino.cc",
		AfterMessage:  "Index File downloaded",
		ErrorMessage:  "Can't download index file, check your network connection.",
		Task: task.Task(func() task.Result {
			return task.Result{
				Result: nil,
				Error:  downloadFunc(),
			}
		}),
	}
}

//corruptedIndexFixResults executes a generic index fix procedure, made by a download and parse task.
func corruptedIndexFixResults(downloadTask, parseTask task.Wrapper) []task.Result {
	subTasks := []task.Wrapper{downloadTask, parseTask}
	wrapper := indexFixWrapperSkeleton()
	wrapper.Task = task.CreateSequence(subTasks, []bool{false, false}).Task()
	return wrapper.Execute().Result.([]task.Result)
}

// indexParseWrapperSkeleton provides an empty skeleton for a task wrapper regarding index (core index, lib index) error fixing tasks,
// which will be assigned later.
func indexFixWrapperSkeleton() task.Wrapper {
	return task.Wrapper{
		BeforeMessage: "Cannot parse index file, it may be corrupted.",
	}
}

// indexParseWrapperSkeleton provides an empty skeleton for a task wrapper regarding index (core index, lib index) parsing tasks,
// which will be assigned later.
func indexParseWrapperSkeleton() task.Wrapper {
	return task.Wrapper{
		BeforeMessage: "Parsing downloaded index file",
		ErrorMessage:  "Cannot parse index file",
		Task:          nil,
	}
}
