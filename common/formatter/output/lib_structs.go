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

import (
	"fmt"
	"strings"

	"github.com/bcmi-labs/arduino-cli/arduino/libraries/librariesindex"
)

// VersionResult represents the output of the version commands.
type VersionResult struct {
	CommandName string `json:"command,required"`
	Version     string `json:"version,required"`
}

func (vr VersionResult) String() string {
	return fmt.Sprintf("%s ver.%s", vr.CommandName, vr.Version)
}

// VersionFullInfo represents the output of a verbose request of version of a command.
type VersionFullInfo struct {
	Versions []VersionResult `json:"versions,required"`
}

// LibProcessResults represent the result of a process on libraries.
type LibProcessResults struct {
	Libraries map[string]ProcessResult `json:"libraries,required"`
}

// CoreProcessResults represent the result of a process on cores or tools.
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
	Libraries []*librariesindex.Library `json:"libraries,required"`
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
			fmt.Sprintln("  Author: ", l.Latest.Author) +
			fmt.Sprintln("  Maintainer: ", l.Latest.Maintainer) +
			fmt.Sprintln("  Sentence: ", l.Latest.Sentence) +
			fmt.Sprintln("  Paragraph: ", l.Latest.Paragraph) +
			fmt.Sprintln("  Website: ", l.Latest.Website) +
			fmt.Sprintln("  Category: ", l.Latest.Category) +
			fmt.Sprintln("  Architecture: ", strings.Join(l.Latest.Architectures, ", ")) +
			fmt.Sprintln("  Types: ", strings.Join(l.Latest.Types, ", ")) +
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
