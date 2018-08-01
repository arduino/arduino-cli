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

import (
	"fmt"
	"strings"
)

//ProcessResult contains info about a completed process.
type ProcessResult struct {
	ItemName string `json:"name,required"`
	Status   string `json:"status,omitempty"`
	Error    string `json:"error,omitempty"`
	Path     string `json:"path,omitempty"`
}

// String returns a string representation of the object.
//   EXAMPLE:
//   ToolName - ErrorText: Error explaining why failed
//   ToolName - StatusText: PATH = /path/to/result/dir
func (lr ProcessResult) String() string {
	ret := lr.ItemName
	if lr.Status != "" {
		ret += fmt.Sprint(" - ", lr.Status)
	}
	if lr.Error != "" {
		ret += fmt.Sprint(" - ", lr.Error)
	}
	return strings.TrimSpace(ret)
}
