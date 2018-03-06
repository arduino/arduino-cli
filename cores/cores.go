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

package cores

import (
	"fmt"
	"strings"

	properties "github.com/arduino/go-properties-map"
	"github.com/bcmi-labs/arduino-cli/common/releases"

	"github.com/blang/semver"
)

// Platform represents a platform package.
type Platform struct {
	Architecture string // The name of the architecture of this package.
	Name         string
	Category     string
	Releases     map[string]*PlatformRelease // The Releases of this platform, labeled by version.
	Package      *Package                    `json:"-"`
}

// PlatformRelease represents a release of a plaform package.
type PlatformRelease struct {
	Resource     *releases.DownloadResource
	Version      string
	BoardNames   []string
	Dependencies ToolDependencies // The Dependency entries to load tools.
	Platform     *Platform        `json:"-"`

	Properties  properties.Map            `json:"-"`
	Boards      map[string]*Board         `json:"-"`
	Programmers map[string]properties.Map `json:"-"`
	Folder      string                    `json:"-"`
}

// ToolDependencies is a set of tool dependency
type ToolDependencies []*ToolDependency

// ToolDependency is a tuple that uniquely identifies a specific version of a Tool
type ToolDependency struct {
	ToolName     string
	ToolVersion  string
	ToolPackager string
}

// GetOrCreateRelease returns the specified release corresponding the provided version,
// or creates a new one if not found.
func (platform *Platform) GetOrCreateRelease(version string) *PlatformRelease {
	if release, ok := platform.Releases[version]; ok {
		return release
	}
	release := &PlatformRelease{
		Boards:      map[string]*Board{},
		Properties:  properties.Map{},
		Programmers: map[string]properties.Map{},
		Platform:    platform,
	}
	platform.Releases[version] = release
	return release
}

// GetRelease returns the specified release corresponding the provided version,
// or nil if not found.
func (platform *Platform) GetRelease(version string) *PlatformRelease {
	if version == "latest" {
		return platform.GetRelease(platform.latestRelease())
	}
	return platform.Releases[version]
}

// GetAllReleasesVersions returns all the version numbers in this Platform Package.
func (platform *Platform) GetAllReleasesVersions() semver.Versions {
	versions := make(semver.Versions, 0, len(platform.Releases))
	for _, release := range platform.Releases {
		temp, err := semver.Make(release.Version)
		if err == nil {
			versions = append(versions, temp)
		}
	}

	return versions
}

// latestRelease obtains latest version number.
func (platform *Platform) latestRelease() string {
	// TODO: Cache latest version using a field in Platform
	versions := platform.GetAllReleasesVersions()
	if len(versions) == 0 {
		return ""
	}
	max := versions[0]
	for i := 1; i < len(versions); i++ {
		if versions[i].GT(max) {
			max = versions[i]
		}
	}
	return fmt.Sprint(max)
}

// GetInstalled return one of the installed PlatformRelease
// TODO: This is a temporary method to help incremental transition from
// arduino-builder, it will be probably removed in the future
func (platform *Platform) GetInstalled() *PlatformRelease {
	for _, release := range platform.Releases {
		if release.Folder != "" {
			return release
		}
	}
	return nil
}

func (platform *Platform) String() string {
	res := fmt.Sprintln("Name        :", platform.Name) +
		fmt.Sprintln("Architecture:", platform.Architecture) +
		fmt.Sprintln("Category    :", platform.Category)
	if platform.Releases != nil && len(platform.Releases) > 0 {
		res += "Releases:\n"
		for _, release := range platform.Releases {
			res += fmt.Sprintln(release)
		}
	}
	return res
}

func (release *PlatformRelease) String() string {
	return fmt.Sprintln("  Version           : ", release.Version) +
		fmt.Sprintln("  Boards            :") +
		fmt.Sprintln(strings.Join(release.BoardNames, ", ")) +
		fmt.Sprintln("  Archive File Name :", release.Resource.ArchiveFileName) +
		fmt.Sprintln("  Checksum          :", release.Resource.Checksum) +
		fmt.Sprintln("  File Size         :", release.Resource.Size) +
		fmt.Sprintln("  URL               :", release.Resource.URL)
}
