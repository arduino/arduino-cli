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

//LibProcessResults represent the result of a process on libraries.
type LibProcessResults struct {
	Libraries map[string]ProcessResult `json:"libraries,required"`
}

//CoreProcessResults represent the result of a process on cores or tools.
type CoreProcessResults struct {
	Cores map[string]ProcessResult `json:"cores,omitempty"`
	Tools map[string]ProcessResult `json:"tools,omitempty"`
}

// String returns a string representation of the object.
func (cpr CoreProcessResults) String() string {
	ret := ""
	for _, cr := range cpr.Cores {
		ret += fmt.Sprintln(cr)
	}
	for _, tr := range cpr.Tools {
		ret += fmt.Sprintln(tr)
	}
	return ret
}

// LibSearchResults represents a set of results of a search of libraries.
type LibSearchResults struct {
	Libraries []*libraries.Library `json:"libraries,required"`
}

// String returns a string representation of the object.
func (lpr LibProcessResults) String() string {
	ret := ""
	for _, lr := range lpr.Libraries {
		ret += fmt.Sprintln(lr)
	}
	return strings.TrimSpace(ret)
}

func (vfi VersionFullInfo) String() string {
	ret := ""
	for _, vr := range vfi.Versions {
		ret += fmt.Sprintln(vr)
	}
	return strings.TrimSpace(ret)
}

// String returns a string representation of the object.
func (lsr LibSearchResults) String() string {
	ret := ""
	for _, l := range lsr.Libraries {
		ret += fmt.Sprintf("Name: \"%s\"\n", l.Name) +
			fmt.Sprintln("  Author: ", l.Author) +
			fmt.Sprintln("  Maintainer: ", l.Maintainer) +
			fmt.Sprintln("  Sentence: ", l.Sentence) +
			fmt.Sprintln("  Paragraph: ", l.Paragraph) +
			fmt.Sprintln("  Website: ", l.Website) +
			fmt.Sprintln("  Category: ", l.Category) +
			fmt.Sprintln("  Architecture: ", strings.Join(l.Architectures, ", ")) +
			fmt.Sprintln("  Types: ", strings.Join(l.Types, ", ")) +
			fmt.Sprintln("  Versions: ", strings.Replace(fmt.Sprint(l.Versions()), " ", ", ", -1))
	}
	return strings.TrimSpace(ret)
}

// func (r *Release) Dump() string {
// 	return fmt.Sprintln("  Release: "+fmt.Sprint(r.Version)) +
// 		fmt.Sprintln("    URL: "+r.Resource.URL) +
// 		fmt.Sprintln("    ArchiveFileName: "+r.Resource.ArchiveFileName) +
// 		fmt.Sprintln("    Size: ", r.Resource.Size) +
// 		fmt.Sprintln("    Checksum: ", r.Resource.Checksum)
// }

// Results returns a set of generic results, to allow them to be modified externally.
//
// -> ProcessResults interface.
func (lpr LibProcessResults) Results() map[string]ProcessResult {
	return lpr.Libraries
}
