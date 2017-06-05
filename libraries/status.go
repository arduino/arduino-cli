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

import "github.com/pmylund/sortutil"

// StatusContext keeps the current status of the libraries in the system
// (the list of libraries, revisions, installed paths, etc.)
type StatusContext struct {
	Libraries map[string]*Library
}

// Library represents a library in the system
type Library struct {
	Name          string
	Author        string
	Maintainer    string
	Sentence      string
	Paragraph     string
	Website       string
	Category      string
	Architectures []string
	Types         []string
	Releases      []*Release
	Installed     *Release
	Latest        *Release
}

// Release represents a release of a library
type Release struct {
	Version         string
	URL             string
	ArchiveFileName string
	Size            int
	Checksum        string
}

// Versions returns an array of all versions available of the library
func (l *Library) Versions() []string {
	res := make([]string, len(l.Releases))
	i := 0
	for _, r := range l.Releases {
		res[i] = r.Version
		i++
	}
	sortutil.CiAsc(res)
	return res
}

// AddLibrary adds an IndexRelease to the status context
func (l *StatusContext) AddLibrary(indexLib *IndexRelease) {
	name := indexLib.Name
	if l.Libraries[name] == nil {
		l.Libraries[name] = indexLib.ExtractLibrary()
	} else {
		release := indexLib.ExtractRelease()
		lib := l.Libraries[name]
		lib.Releases = append(lib.Releases, release)
		if lib.Latest.Version < release.Version {
			lib.Latest = release
		}
	}
}

// Names returns an array with all the names of the registered libraries
func (l *StatusContext) Names() []string {
	res := make([]string, len(l.Libraries))
	i := 0
	for n := range l.Libraries {
		res[i] = n
		i++
	}
	sortutil.CiAsc(res)
	return res
}

func CreateStatusContextFromIndex(index *Index, sketchbookPaths []string, corePaths []string) (*StatusContext, error) {
	// Start with an empty status context
	libraries := StatusContext{
		Libraries: map[string]*Library{},
	}
	for _, lib := range index.Libraries {
		// Add all indexed libraries in the status context
		libraries.AddLibrary(&lib)
	}
	return &libraries, nil
}
