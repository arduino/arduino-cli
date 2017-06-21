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

package libraries

import (
	"strconv"

	"github.com/pmylund/sortutil"

	"strings"
)

// Library represents a library in the system
type Library struct {
	Name          string              `json:"name"`
	Author        string              `json:"author"`
	Maintainer    string              `json:"maintainer"`
	Sentence      string              `json:"sentence"`
	Paragraph     string              `json:"paragraph"`
	Website       string              `json:"website"`
	Category      string              `json:"category"`
	Architectures []string            `json:"architectures"`
	Types         []string            `json:"types"`
	Releases      map[string]*Release `json:"releases"`
	Installed     *Release            `json:"installed"`
}

// Release represents a release of a library
type Release struct {
	Version         string `json:"version"`
	URL             string `json:"url"`
	ArchiveFileName string `json:"archiveFileName"`
	Size            int    `json:"size"`
	Checksum        string `json:"checksum"`
}

// Versions returns an array of all versions available of the library
func (l *Library) Versions() []string {
	res := make([]string, len(l.Releases))
	i := 0
	for version := range l.Releases {
		res[i] = version
		i++
	}
	sortutil.CiAsc(res)
	return res
}

// GetVersion returns the Release corresponding to the specified version, or
// nil if not found.
func (l *Library) GetVersion(version string) *Release {
	return l.Releases[version]
}

// Latest obtains the latest version of a library.
func (l *Library) Latest() *Release {
	return l.GetVersion(l.latestVersion())
}

// latestVersion obtains latest version number.
//
// It uses lexicographics to compare version strings.
func (l *Library) latestVersion() string {
	versions := l.Versions()
	if len(versions) > 0 {
		return versions[0]
	}
	return ""
}

func (r *Release) String() string {
	res := "  Release: " + r.Version + "\n"
	res += "    URL: " + r.URL + "\n"
	res += "    ArchiveFileName: " + r.ArchiveFileName + "\n"
	res += "    Size: " + strconv.Itoa(r.Size) + "\n"
	res += "    Checksum: " + r.Checksum + "\n"
	return res
}

func (l *Library) String() string {
	res := "Name: " + l.Name + "\n"
	res += "  Author: " + l.Author + "\n"
	res += "  Maintainer: " + l.Maintainer + "\n"
	res += "  Sentence: " + l.Sentence + "\n"
	res += "  Paragraph: " + l.Paragraph + "\n"
	res += "  Website: " + l.Website + "\n"
	res += "  Category: " + l.Category + "\n"
	res += "  Architecture: " + strings.Join(l.Architectures, ", ") + "\n"
	res += "  Types: " + strings.Join(l.Types, ", ") + "\n"
	res += "  Versions: " + strings.Join(l.Versions(), ", ") + "\n"
	return res
}
