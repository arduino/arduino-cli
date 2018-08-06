/*
 * This file is part of arduino-cli.
 *
 * Copyright 2018 ARDUINO SA (http://www.arduino.cc/)
 *
 * This software is released under the GNU General Public License version 3,
 * which covers the main part of arduino-cli.
 * The terms of this license can be found at:
 * https://www.gnu.org/licenses/gpl-3.0.en.html
 *
 * You can be released from the requirements of the above licenses by purchasing
 * a commercial license. Buying such a license is mandatory if you want to modify or
 * otherwise use the software for commercial activities involving the Arduino
 * software without disclosing the source code of your own applications. To purchase
 * a commercial license, send an email to license@arduino.cc.
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
