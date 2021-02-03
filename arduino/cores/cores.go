// This file is part of arduino-cli.
//
// Copyright 2020 ARDUINO SA (http://www.arduino.cc/)
//
// This software is released under the GNU General Public License version 3,
// which covers the main part of arduino-cli.
// The terms of this license can be found at:
// https://www.gnu.org/licenses/gpl-3.0.en.html
//
// You can be released from the requirements of the above licenses by purchasing
// a commercial license. Buying such a license is mandatory if you want to
// modify or otherwise use the software for commercial activities involving the
// Arduino software without disclosing the source code of your own applications.
// To purchase a commercial license, send an email to license@arduino.cc.

package cores

import (
	"encoding/json"
	"sort"
	"strings"

	"github.com/arduino/arduino-cli/arduino/resources"
	paths "github.com/arduino/go-paths-helper"
	properties "github.com/arduino/go-properties-orderedmap"
	semver "go.bug.st/relaxed-semver"
)

// Platform represents a platform package.
type Platform struct {
	Architecture      string // The name of the architecture of this package.
	Name              string
	Category          string
	Releases          map[string]*PlatformRelease // The Releases of this platform, labeled by version.
	Package           *Package                    `json:"-"`
	ManuallyInstalled bool                        // true if the Platform has been installed without the CLI
}

// PlatformReleaseHelp represents the help URL for this Platform release
type PlatformReleaseHelp struct {
	Online string `json:"-"`
}

// PlatformRelease represents a release of a plaform package.
type PlatformRelease struct {
	Resource       *resources.DownloadResource
	Version        *semver.Version
	BoardsManifest []*BoardManifest
	Dependencies   ToolDependencies       // The Dependency entries to load tools.
	Help           PlatformReleaseHelp    `json:"-"`
	Platform       *Platform              `json:"-"`
	Properties     *properties.Map        `json:"-"`
	Boards         map[string]*Board      `json:"-"`
	Programmers    map[string]*Programmer `json:"-"`
	Menus          *properties.Map        `json:"-"`
	InstallDir     *paths.Path            `json:"-"`
	IsIDEBundled   bool                   `json:"-"`
	IsTrusted      bool                   `json:"-"`
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

// Sort sorts the ToolDependencies by name and (if multiple instance of the same
// tool is required) by version.
func (deps ToolDependencies) Sort() {
	sort.Slice(deps, func(i, j int) bool {
		if deps[i].ToolPackager != deps[j].ToolPackager {
			return deps[i].ToolPackager < deps[j].ToolPackager
		}
		if deps[i].ToolName != deps[j].ToolName {
			return deps[i].ToolName < deps[j].ToolName
		}
		return deps[i].ToolVersion.LessThan(deps[j].ToolVersion)
	})
}

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
func (platform *Platform) GetOrCreateRelease(version *semver.Version) *PlatformRelease {
	tag := ""
	if version != nil {
		tag = version.String()
	}
	if release, ok := platform.Releases[tag]; ok {
		return release
	}
	release := &PlatformRelease{
		Version:     version,
		Boards:      map[string]*Board{},
		Properties:  properties.NewMap(),
		Programmers: map[string]*Programmer{},
		Platform:    platform,
	}
	platform.Releases[tag] = release
	return release
}

// FindReleaseWithVersion returns the specified release corresponding the provided version,
// or nil if not found.
func (platform *Platform) FindReleaseWithVersion(version *semver.Version) *PlatformRelease {
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
	return platform.FindReleaseWithVersion(latestVersion)
}

// GetAllReleases returns all the releases of this platform, or an empty
// slice if no releases are available
func (platform *Platform) GetAllReleases() []*PlatformRelease {
	retVal := []*PlatformRelease{}
	for _, v := range platform.GetAllReleasesVersions() {
		retVal = append(retVal, platform.FindReleaseWithVersion(v))
	}

	return retVal
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

// GetAllInstalled returns all installed PlatformRelease
func (platform *Platform) GetAllInstalled() []*PlatformRelease {
	res := []*PlatformRelease{}
	if platform.Releases != nil {
		for _, release := range platform.Releases {
			if release.IsInstalled() {
				res = append(res, release)
			}
		}

	}
	return res
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
		BoardID:         boardID,
		Properties:      properties.NewMap(),
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
			toolDep.ToolVersion.Equal(toolRelease.Version) {
			return true
		}
	}
	return false
}

// RuntimeProperties returns the runtime properties for this PlatformRelease
func (release *PlatformRelease) RuntimeProperties() *properties.Map {
	res := properties.NewMap()
	if release.InstallDir != nil {
		res.Set("runtime.platform.path", release.InstallDir.String())
	}

	return res
}

// GetLibrariesDir returns the path to the core libraries or nil if not
// present
func (release *PlatformRelease) GetLibrariesDir() *paths.Path {
	if release.InstallDir != nil {
		libDir := release.InstallDir.Join("libraries")
		if libDir.IsDir() {
			return libDir
		}
	}

	return nil
}

// IsInstalled returns true if the PlatformRelease is installed
func (release *PlatformRelease) IsInstalled() bool {
	return release.InstallDir != nil
}

func (release *PlatformRelease) String() string {
	version := ""
	if release.Version != nil {
		version = release.Version.String()
	}
	return release.Platform.String() + "@" + version
}

// MarshalJSON provides a more user friendly serialization for
// PlatformRelease objects.
func (release *PlatformRelease) MarshalJSON() ([]byte, error) {
	latestStr := ""
	latest := release.Platform.GetLatestRelease()
	if latest != nil {
		latestStr = latest.Version.String()
	}

	return json.Marshal(&struct {
		ID        string `json:"ID,omitempty"`
		Installed string `json:"Installed,omitempty"`
		Latest    string `json:"Latest,omitempty"`
		Name      string `json:"Name,omitempty"`
	}{
		ID:        release.Platform.String(),
		Installed: release.Version.String(),
		Latest:    latestStr,
		Name:      release.Platform.Name,
	})
}
