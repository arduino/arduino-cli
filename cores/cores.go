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
	"os"
	"strings"

	"github.com/bcmi-labs/arduino-cli/common/releases"

	"github.com/blang/semver"
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
	URL             string
	Dependencies    ToolDependencies // The Dependency entries to load tools.
}

// ToolDependencies is a set of tuples representing summary data of a tool dependency set.
type ToolDependencies []*ToolDependency

// ToolDependency is a tuple representing summary data of a tool.
type ToolDependency struct {
	ToolName     string
	ToolVersion  string
	ToolPackager string
}

// GetVersion returns the specified release corresponding the provided version,
// or nil if not found.
func (core *Core) GetVersion(version string) *Release {
	if version == "latest" {
		return core.GetVersion(core.latestVersion())
	}
	return core.Releases[version]
}

// Versions returns all the version numbers in this Core Package.
func (core *Core) Versions() semver.Versions {
	versions := make(semver.Versions, 0, len(core.Releases))
	for _, release := range core.Releases {
		temp, err := semver.Make(release.Version)
		if err == nil {
			versions = append(versions, temp)
		}
	}

	return versions
}

// latestVersion obtains latest version number.
//
// It uses lexicographics to compare version strings.
func (core *Core) latestVersion() string {
	versions := core.Versions()
	if len(versions) > 0 {
		max := versions[0]
		for i := 1; i < len(versions); i++ {
			if versions[i].GT(max) {
				max = versions[i]
			}
		}
		return fmt.Sprint(max)
	}
	return ""
}

func (core *Core) String() string {
	res := fmt.Sprintln("Name        :", core.Name) +
		fmt.Sprintln("Architecture:", core.Architecture) +
		fmt.Sprintln("Category    :", core.Category)
	if core.Releases != nil && len(core.Releases) > 0 {
		res += "Releases:\n"
		for _, release := range core.Releases {
			res += fmt.Sprintln(release)
		}
	}
	return res
}

func (release *Release) String() string {
	return fmt.Sprintln("  Version           : ", release.Version) +
		fmt.Sprintln("  Boards            :") +
		fmt.Sprintln(strings.Join(release.Boards, ", ")) +
		fmt.Sprintln("  Archive File Name :", release.ArchiveFileName) +
		fmt.Sprintln("  Checksum          :", release.Checksum) +
		fmt.Sprintln("  File Size         :", release.Size) +
		fmt.Sprintln("  URL               :", release.URL)
}

// OpenLocalArchiveForDownload Creates an empty file if not found.
func (release Release) OpenLocalArchiveForDownload() (*os.File, error) {
	path, err := releases.ArchivePath(release)
	if err != nil {
		return nil, err
	}
	stats, err := os.Stat(path)
	if os.IsNotExist(err) || err == nil && stats.Size() >= release.Size {
		return os.Create(path)
	} else if err != nil {
		return nil, err
	}
	return os.OpenFile(path, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
}

// Implementation of Release interface

// ExpectedChecksum returns the expected checksum for this release.
func (release Release) ExpectedChecksum() string {
	return release.Checksum
}

// ArchiveName returns the archive file name (not the path)
func (release Release) ArchiveName() string {
	return release.ArchiveFileName
}

// ArchiveSize returns the archive size.
func (release Release) ArchiveSize() int64 {
	return release.Size
}

// ArchiveURL returns the archive URL.
func (release Release) ArchiveURL() string {
	return release.URL
}
