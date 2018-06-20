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

	"github.com/arduino/go-paths-helper"

	properties "github.com/arduino/go-properties-map"
	"github.com/bcmi-labs/arduino-cli/arduino/resources"

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
	Resource       *resources.DownloadResource
	Version        string
	BoardsManifest []*BoardManifest
	Dependencies   ToolDependencies // The Dependency entries to load tools.
	Platform       *Platform        `json:"-"`

	Properties  properties.Map            `json:"-"`
	Boards      map[string]*Board         `json:"-"`
	Programmers map[string]properties.Map `json:"-"`
	Menus       map[string]string         `json:"-"`
	Folder      string                    `json:"-"`
}

// BoardManifest contains information about a board. These metadata are usually
// provided by the package_index.json
type BoardManifest struct {
	Name string `json:"-"`
}

// ToolDependencies is a set of tool dependency
type ToolDependencies []*ToolDependency

// ToolDependency is a tuple that uniquely identifies a specific version of a Tool
type ToolDependency struct {
	ToolName     string
	ToolVersion  string
	ToolPackager string
}

func (dep *ToolDependency) String() string {
	return dep.ToolPackager + ":" + dep.ToolName + "@" + dep.ToolVersion
}

// GetOrCreateRelease returns the specified release corresponding the provided version,
// or creates a new one if not found.
func (platform *Platform) GetOrCreateRelease(version string) *PlatformRelease {
	if release, ok := platform.Releases[version]; ok {
		return release
	}
	release := &PlatformRelease{
		Version:     version,
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
		return platform.GetRelease(platform.latestReleaseVersion())
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

// latestReleaseVersion obtains latest version number.
func (platform *Platform) latestReleaseVersion() string {
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
	return platform.Package.Name + ":" + platform.Architecture
}

// GetOrCreateBoard returns the Board object with the specified boardID
// or creates a new one if not found
func (release *PlatformRelease) GetOrCreateBoard(boardID string) *Board {
	if board, ok := release.Boards[boardID]; ok {
		return board
	}
	board := &Board{
		BoardId:         boardID,
		Properties:      properties.Map{},
		PlatformRelease: release,
	}
	release.Boards[boardID] = board
	return board
}

// RuntimeProperties returns the runtime properties for this PlatformRelease
func (release *PlatformRelease) RuntimeProperties() properties.Map {
	return properties.Map{
		"runtime.platform.path": release.Folder,
	}
}

// GetLibrariesDir returns the path to the core libraries or nil if not
// present
func (release *PlatformRelease) GetLibrariesDir() *paths.Path {
	libDir := paths.New(release.Folder).Join("libraries")
	if isDir, _ := libDir.IsDir(); isDir {
		return libDir
	}
	return nil
}

func (release *PlatformRelease) String() string {
	return release.Platform.String() + "@" + release.Version
}
