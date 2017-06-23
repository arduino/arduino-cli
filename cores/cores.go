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

package cores

import (
	"fmt"
	"strings"

	"github.com/pmylund/sortutil"
)

// Core represents a core package.
type Core struct {
	Name         string              // The name of the Core Package.
	Architecture string              // The name of the architecture of this package.
	Category     string              // The category which this core package belongs to.
	Releases     map[string]*Release // The Releases of this core package, labeled by version.
}

// Release represents a release of a core package.
type Release struct {
	Version         string
	ArchiveFileName string
	Checksum        string
	Size            int64
	Boards          []string
}

// GetVersion returns the specified release corresponding the provided version,
// or nil if not found.
func (core *Core) GetVersion(version string) *Release {
	return core.Releases[version]
}

// Versions returns all the version numbers in this Core Package.
func (core *Core) Versions() []string {
	versions := make([]string, len(core.Releases))
	i := 0
	for version := range core.Releases {
		versions[i] = version
		i++
	}
	sortutil.CiAsc(versions)
	return versions
}

// Latest obtains latest version of a core package.
func (core *Core) Latest() *Release {
	return core.GetVersion(core.latestVersion())
}

// latestVersion obtains latest version number.
//
// It uses lexicographics to compare version strings.
func (core *Core) latestVersion() string {
	versions := core.Versions()
	if len(versions) > 0 {
		return versions[0]
	}
	return ""
}

func (core *Core) String(verbosity int) (res string) {
	if verbosity > 0 {
		res += fmt.Sprintf("Name        : %s\n", core.Name)
		res += fmt.Sprintf("Architecture: %s\n", core.Architecture)
		res += fmt.Sprintf("Category    : %s\n", core.Category)
		if verbosity > 1 {
			res += "Releases:\n"
			for _, release := range core.Releases {
				res += fmt.Sprintf("%s\n", release.String())
			}
		} else {
			res += fmt.Sprintf("Releases    : %s", core.Versions())
		}
	} else {
		res = fmt.Sprintf("%s\n", core.Name)
	}
	return
}

func (release *Release) String() string {
	res := fmt.Sprintf("Version           : %s\n", release.Version)
	res += fmt.Sprintf("Boards            : \n%s\n", strings.Join(release.Boards, ", "))
	res += fmt.Sprintf("Archive File Name : %s\n", release.ArchiveFileName)
	res += fmt.Sprintf("Checksum          : %s\n", release.Checksum)
	res += fmt.Sprintf("File Size         : %d\n", release.Size)
	return res
}
