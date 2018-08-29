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

package cores

import (
	"strings"

	"github.com/arduino/go-paths-helper"

	"github.com/arduino/arduino-cli/arduino/resources"
	properties "github.com/arduino/go-properties-map"
	"go.bug.st/relaxed-semver"
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
	Version        *semver.Version
	BoardsManifest []*BoardManifest
	Dependencies   ToolDependencies // The Dependency entries to load tools.
	Platform       *Platform        `json:"-"`

	Properties  properties.Map            `json:"-"`
	Boards      map[string]*Board         `json:"-"`
	Programmers map[string]properties.Map `json:"-"`
	Menus       map[string]string         `json:"-"`
	InstallDir  *paths.Path               `json:"-"`
}

// BoardManifest contains information about a board. These metadata are usually
// provided by the package_index.json
type BoardManifest struct {
	Name string             `json:"-"`
	ID   []*BoardManifestID `json:"-"`
}

// BoardManifestID contains information on how to identify a board. These metadata
// are usually provided by the package_index.json
type BoardManifestID struct {
	USB string `json:"-"`
}

// HasUsbID returns true if the BoardManifes contains the specified USB id as
// identification for this board. usbID should be in the format "0000:0000"
func (bm *BoardManifest) HasUsbID(vid, pid string) bool {
	usbID := strings.ToLower(vid + ":" + pid)
	for _, id := range bm.ID {
		if usbID == strings.ToLower(id.USB) {
			return true
		}
	}
	return false
}

// ToolDependencies is a set of tool dependency
type ToolDependencies []*ToolDependency

// ToolDependency is a tuple that uniquely identifies a specific version of a Tool
type ToolDependency struct {
	ToolName     string
	ToolVersion  *semver.RelaxedVersion
	ToolPackager string
}

func (dep *ToolDependency) String() string {
	return dep.ToolPackager + ":" + dep.ToolName + "@" + dep.ToolVersion.String()
}

// GetOrCreateRelease returns the specified release corresponding the provided version,
// or creates a new one if not found.
func (platform *Platform) GetOrCreateRelease(version *semver.Version) (*PlatformRelease, error) {
	tag := ""
	if version != nil {
		tag = version.String()
	}
	if release, ok := platform.Releases[tag]; ok {
		return release, nil
	}
	release := &PlatformRelease{
		Version:     version,
		Boards:      map[string]*Board{},
		Properties:  properties.Map{},
		Programmers: map[string]properties.Map{},
		Platform:    platform,
	}
	platform.Releases[tag] = release
	return release, nil
}

// GetReleaseRelaxedVersion returns the specified release corresponding the provided version,
// or nil if not found.
func (platform *Platform) GetReleaseVersion(version *semver.Version) *PlatformRelease {
	// use as an fmt.Stringer
	return platform.Releases[version.String()]
}

// GetLatestRelease returns the latest release of this platform, or nil if no releases
// are available
func (platform *Platform) GetLatestRelease() *PlatformRelease {
	latestVersion := platform.latestReleaseVersion()
	if latestVersion == nil {
		return nil
	}
	return platform.GetReleaseVersion(latestVersion)
}

// GetAllReleasesVersions returns all the version numbers in this Platform Package.
func (platform *Platform) GetAllReleasesVersions() []*semver.Version {
	versions := []*semver.Version{}
	for _, release := range platform.Releases {
		versions = append(versions, release.Version)
	}
	return versions
}

// latestReleaseVersion obtains latest version number, or nil if no release available
func (platform *Platform) latestReleaseVersion() *semver.Version {
	// TODO: Cache latest version using a field in Platform
	versions := platform.GetAllReleasesVersions()
	if len(versions) == 0 {
		return nil
	}
	max := versions[0]

	for i := 1; i < len(versions); i++ {
		if versions[i].GreaterThan(max) {
			max = versions[i]
		}
	}
	return max
}

// GetInstalled return one of the installed PlatformRelease
// TODO: This is a temporary method to help incremental transition from
// arduino-builder, it will be probably removed in the future
func (platform *Platform) GetInstalled() *PlatformRelease {
	for _, release := range platform.Releases {
		if release.InstallDir != nil {
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

// RequiresToolRelease returns true if the PlatformRelease requires the
// toolReleased passed as parameter
func (release *PlatformRelease) RequiresToolRelease(toolRelease *ToolRelease) bool {
	for _, toolDep := range release.Dependencies {
		if toolDep.ToolName == toolRelease.Tool.Name &&
			toolDep.ToolPackager == toolRelease.Tool.Package.Name &&
			toolDep.ToolVersion == toolRelease.Version {
			return true
		}
	}
	return false
}

// RuntimeProperties returns the runtime properties for this PlatformRelease
func (release *PlatformRelease) RuntimeProperties() properties.Map {
	return properties.Map{
		"runtime.platform.path": release.InstallDir.String(),
	}
}

// GetLibrariesDir returns the path to the core libraries or nil if not
// present
func (release *PlatformRelease) GetLibrariesDir() *paths.Path {
	libDir := release.InstallDir.Join("libraries")
	if isDir, _ := libDir.IsDir(); isDir {
		return libDir
	}
	return nil
}

func (release *PlatformRelease) String() string {
	version := ""
	if release.Version != nil {
		version = release.Version.String()
	}
	return release.Platform.String() + "@" + version
}
