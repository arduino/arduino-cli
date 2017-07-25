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

package output

import (
	"fmt"
	"strings"

	"github.com/bcmi-labs/arduino-cli/libraries"
)

// VersionResult represents the output of the version commands.
type VersionResult struct {
	CommandName string `json:"command,required"`
	Version     string `json:"version,required"`
}

func (vr VersionResult) String() string {
	return fmt.Sprintf("%s ver.%s", vr.CommandName, vr.Version)
}

//VersionFullInfo represents the output of a verbose request of version of a command.
type VersionFullInfo struct {
	Versions []VersionResult `json:"versions,required"`
}

func (vfi VersionFullInfo) String() string {
	ret := ""
	for _, vr := range vfi.Versions {
		ret += fmt.Sprintln(vr)
	}
	return strings.TrimSpace(ret)
}

//LibProcessResults represent the result of a process on libraries.
type LibProcessResults struct {
	Libraries []LibProcessResult `json:"libraries,required"`
}

// String returns a string representation of the object.
func (lpr LibProcessResults) String() string {
	ret := ""
	for _, lr := range lpr.Libraries {
		ret += fmt.Sprintln(lr)
	}
	return strings.TrimSpace(ret)
}

//LibProcessResult contains info about a completed process.
type LibProcessResult struct {
	LibraryName string `json:"name,required"`
	Status      string `json:"status,omitempty"`
	Error       string `json:"error,omitempty"`
	Path        string `json:"path,omitempty"`
}

// String returns a string representation of the object.
func (lr LibProcessResult) String() string {
	if lr.Error != "" {
		return strings.TrimSpace(fmt.Sprintf("%s - Error : %s", lr.LibraryName, lr.Error))
	}
	return strings.TrimSpace(fmt.Sprintf("%s - %s", lr.LibraryName, lr.Status))
}

//LibSearchResults represents a result of a search of libraries.
type LibSearchResults struct {
	Libraries []interface{} `json:"searchResults,required"`
}

// String returns a string representation of the object.
func (lsr LibSearchResults) String() string {
	ret := fmt.Sprintln("Search results:")
	for _, lib := range lsr.Libraries {
		ret += fmt.Sprintln(lib)
		// if the single cell is a library I may have higher verbosity.
		libVal, isLib := lib.(libraries.Library)
		if isLib && libVal.Releases != nil {
			for _, release := range libVal.Releases {
				ret += fmt.Sprintln(release)
			}
		}
	}
	return strings.TrimSpace(ret)
}
