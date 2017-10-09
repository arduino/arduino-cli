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

package output

import "fmt"

// SketchSyncResult contains the result of an `arduino sketch sync` operation.
type SketchSyncResult struct {
	PushedSketches  []string          `json:"pushedSketches,required"`
	PulledSketches  []string          `json:"pulledSketches,required"`
	SkippedSketches []string          `json:"skippedSketches,required"`
	Errors          []SketchSyncError `json:"errors,required"`
}

// SketchSyncError represents an error during a `arduino sketch sync` operation.
type SketchSyncError struct {
	Sketch string `json:"sketch,required"`
	Error  error  `json:"error,required"`
}

// String returns a string representation of the object. For this object
// it is used to provide verbose info of a sync process.
func (ssr SketchSyncResult) String() string {
	totalSketches := len(ssr.SkippedSketches) + len(ssr.PushedSketches) + len(ssr.PulledSketches) + len(ssr.Errors)
	//this function iterates an array and if it's not empty pretty prints it.
	iterate := func(array []string, header string) string {
		ret := ""
		if len(array) > 0 {
			ret += fmt.Sprintln(header)
			for _, item := range array {
				ret += fmt.Sprintln(" -", item)
			}
		}
		return ret
	}

	ret := fmt.Sprintf("%d sketches synced:\n", totalSketches)

	ret += iterate(ssr.PushedSketches, "Pushed Sketches:")
	ret += iterate(ssr.PulledSketches, "Pulled Sketches:")
	ret += iterate(ssr.SkippedSketches, "Skipped Sketches:")

	if len(ssr.Errors) > 0 {
		ret += fmt.Sprintln("Errors:")
		for _, item := range ssr.Errors {
			ret += fmt.Sprintf(" - Sketch %s : %s\n", item.Sketch, item.Error)
		}
	}

	return ret
}
