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
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/url"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/arduino/arduino-cli/internal/arduino/globals"
	"github.com/arduino/arduino-cli/internal/arduino/resources"
	"github.com/arduino/arduino-cli/internal/arduino/utils"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	paths "github.com/arduino/go-paths-helper"
	properties "github.com/arduino/go-properties-orderedmap"
	semver "go.bug.st/relaxed-semver"
)

// Platform represents a platform package.
type Platform struct {
	Architecture      string                                       // The name of the architecture of this package.
	Releases          map[semver.NormalizedString]*PlatformRelease // The Releases of this platform, labeled by version.
	Package           *Package                                     `json:"-"`
	ManuallyInstalled bool                                         // true if the Platform exists due to a manually installed release
	Deprecated        bool                                         // true if the latest PlatformRelease of this Platform has been deprecated
	Indexed           bool                                         // true if the Platform has been indexed from additional-urls
	Latest            *semver.Version                              `json:"-"`
}

// PlatformReleaseHelp represents the help URL for this Platform release
type PlatformReleaseHelp struct {
	Online string `json:"-"`
}

// PlatformRelease represents a release of a plaform package.
type PlatformRelease struct {
	Name                    string
	Category                string
	Resource                *resources.DownloadResource
	Version                 *semver.Version
	BoardsManifest          []*BoardManifest
	ToolDependencies        ToolDependencies
	DiscoveryDependencies   DiscoveryDependencies
	MonitorDependencies     MonitorDependencies
	Deprecated              bool
	Help                    PlatformReleaseHelp           `json:"-"`
	Platform                *Platform                     `json:"-"`
	Properties              *properties.Map               `json:"-"`
	Boards                  map[string]*Board             `json:"-"`
	orderedBoards           []*Board                      `json:"-"` // The Boards of this platform, in the order they are defined in the boards.txt file.
	Programmers             map[string]*Programmer        `json:"-"`
	Menus                   *properties.Map               `json:"-"`
	InstallDir              *paths.Path                   `json:"-"`
	Timestamps              *TimestampsStore              // Contains the timestamps of the files used to build this PlatformRelease
	IsTrusted               bool                          `json:"-"`
	PluggableDiscoveryAware bool                          `json:"-"` // true if the Platform supports pluggable discovery (no compatibility layer required)
	Monitors                map[string]*MonitorDependency `json:"-"`
	MonitorsDevRecipes      map[string]string             `json:"-"`
	Compatible              bool                          `json:"-"` // true if at all ToolDependencies are available for the current OS/ARCH.
}

// TimestampsStore is a generic structure to store timestamps
type TimestampsStore struct {
	timestamps map[*paths.Path]*time.Time
}

// NewTimestampsStore creates a new TimestampsStore
func NewTimestampsStore() *TimestampsStore {
	return &TimestampsStore{
		timestamps: map[*paths.Path]*time.Time{},
	}
}

// AddFile adds a file to the TimestampsStore
func (t *TimestampsStore) AddFile(path *paths.Path) {
	if info, err := path.Stat(); err != nil {
		t.timestamps[path] = nil // Save a missing file with a nil timestamp
	} else {
		modtime := info.ModTime()
		t.timestamps[path] = &modtime
	}
}

// Dirty returns true if one of the files stored in the TimestampsStore has been
// changed after being added to the store.
func (t *TimestampsStore) Dirty() bool {
	for path, timestamp := range t.timestamps {
		if info, err := path.Stat(); err != nil {
			if timestamp != nil {
				return true
			}
		} else {
			if timestamp == nil || info.ModTime() != *timestamp {
				return true
			}
		}
	}
	return false
}

// Dirty returns true if one of the files of this PlatformRelease has been changed
// (it means that the PlatformRelease should be rebuilt to be used correctly).
func (release *PlatformRelease) Dirty() bool {
	return release.Timestamps.Dirty()
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

// InternalUniqueIdentifier returns the unique identifier for this object
func (dep *ToolDependency) InternalUniqueIdentifier(platformIndexURL *url.URL) string {
	h := sha256.New()
	h.Write([]byte(dep.String()))
	if platformIndexURL != nil {
		h.Write([]byte(platformIndexURL.String()))
	}
	res := dep.String() + "_" + hex.EncodeToString(h.Sum([]byte{}))[:16]
	return utils.SanitizeName(res)
}

// DiscoveryDependencies is a list of DiscoveryDependency
type DiscoveryDependencies []*DiscoveryDependency

// Sort the DiscoveryDependencies by name.
func (d DiscoveryDependencies) Sort() {
	sort.Slice(d, func(i, j int) bool {
		if d[i].Packager != d[j].Packager {
			return d[i].Packager < d[j].Packager
		}
		return d[i].Name < d[j].Name
	})
}

// DiscoveryDependency identifies a specific discovery, version is omitted
// since the latest version will always be used
type DiscoveryDependency struct {
	Name     string
	Packager string
}

func (d *DiscoveryDependency) String() string {
	return fmt.Sprintf("%s:%s", d.Packager, d.Name)
}

// MonitorDependencies is a list of MonitorDependency
type MonitorDependencies []*MonitorDependency

// Sort the DiscoveryDependencies by name.
func (d MonitorDependencies) Sort() {
	sort.Slice(d, func(i, j int) bool {
		if d[i].Packager != d[j].Packager {
			return d[i].Packager < d[j].Packager
		}
		return d[i].Name < d[j].Name
	})
}

// MonitorDependency identifies a specific monitor, version is omitted
// since the latest version will always be used
type MonitorDependency struct {
	Name     string
	Packager string
}

func (d *MonitorDependency) String() string {
	return fmt.Sprintf("%s:%s", d.Packager, d.Name)
}

// GetOrCreateRelease returns the specified release corresponding the provided version,
// or creates a new one if not found.
func (platform *Platform) GetOrCreateRelease(version *semver.Version) *PlatformRelease {
	if version == nil {
		version = semver.MustParse("")
	}
	tag := version.NormalizedString()
	if release, ok := platform.Releases[tag]; ok {
		return release
	}
	release := &PlatformRelease{
		Version:     version,
		Boards:      map[string]*Board{},
		Properties:  properties.NewMap(),
		Programmers: map[string]*Programmer{},
		Platform:    platform,
		Timestamps:  NewTimestampsStore(),
	}
	platform.Releases[tag] = release
	return release
}

// GetManuallyInstalledRelease returns (*PlatformRelease, true) if the Platform has
// a manually installed release or (nil, false) otherwise.
func (platform *Platform) GetManuallyInstalledRelease() (*PlatformRelease, bool) {
	res, ok := platform.Releases[semver.MustParse("").NormalizedString()]
	return res, ok
}

// FindReleaseWithVersion returns the specified release corresponding the provided version,
// or nil if not found.
func (platform *Platform) FindReleaseWithVersion(version *semver.Version) *PlatformRelease {
	// use as an fmt.Stringer
	return platform.Releases[version.NormalizedString()]
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

// GetLatestCompatibleRelease returns the latest compatible release of this platform, or nil if no
// compatible releases are available.
func (platform *Platform) GetLatestCompatibleRelease() *PlatformRelease {
	var maximum *PlatformRelease
	for _, release := range platform.Releases {
		if !release.IsCompatible() {
			continue
		}
		if maximum == nil || release.Version.GreaterThan(maximum.Version) {
			maximum = release
		}
	}
	return maximum
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

// GetAllCompatibleReleasesVersions returns all the version numbers in this Platform Package that contains compatible tools.
func (platform *Platform) GetAllCompatibleReleasesVersions() []*semver.Version {
	versions := []*semver.Version{}
	for _, release := range platform.Releases {
		if !release.IsCompatible() {
			continue
		}
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
	maximum := versions[0]

	for i := 1; i < len(versions); i++ {
		if versions[i].GreaterThan(maximum) {
			maximum = versions[i]
		}
	}
	return maximum
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
	release.orderedBoards = append(release.orderedBoards, board)
	return board
}

// GetBoards returns the boards in this platforms in the order they
// are defined in the platform.txt file.
func (release *PlatformRelease) GetBoards() []*Board {
	return release.orderedBoards
}

// RequiresToolRelease returns true if the PlatformRelease requires the
// toolReleased passed as parameter
func (release *PlatformRelease) RequiresToolRelease(toolRelease *ToolRelease) bool {
	for _, toolDep := range release.ToolDependencies {
		if toolDep.ToolName == toolRelease.Tool.Name &&
			toolDep.ToolPackager == toolRelease.Tool.Package.Name &&
			toolDep.ToolVersion.Equal(toolRelease.Version) {
			return true
		}
	}
	for _, discovery := range release.DiscoveryDependencies {
		if discovery.Name == toolRelease.Tool.Name &&
			discovery.Packager == toolRelease.Tool.Package.Name &&
			// We always want the latest discovery version available
			toolRelease.Version.Equal(toolRelease.Tool.LatestRelease().Version) {
			return true
		}
	}
	for _, monitor := range release.MonitorDependencies {
		if monitor.Name == toolRelease.Tool.Name &&
			monitor.Packager == toolRelease.Tool.Package.Name &&
			// We always want the latest monitor version available
			toolRelease.Version.Equal(toolRelease.Tool.LatestRelease().Version) {
			return true
		}
	}
	return false
}

// RuntimeProperties returns the runtime properties for this PlatformRelease
func (release *PlatformRelease) RuntimeProperties() *properties.Map {
	res := properties.NewMap()
	if release.InstallDir != nil {
		res.SetPath("runtime.platform.path", release.InstallDir)
		res.SetPath("runtime.hardware.path", release.InstallDir.Join(".."))
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

// ToRPCPlatformReference creates a gRPC PlatformReference message out of this PlatformRelease
func (release *PlatformRelease) ToRPCPlatformReference() *rpc.InstalledPlatformReference {
	defaultURLPrefix := globals.DefaultIndexURL
	// TODO: create a IndexURL object to factorize this
	defaultURLPrefix = strings.TrimSuffix(defaultURLPrefix, filepath.Ext(defaultURLPrefix))
	defaultURLPrefix = strings.TrimSuffix(defaultURLPrefix, filepath.Ext(defaultURLPrefix)) // removes .tar.bz2

	url := release.Platform.Package.URL
	if strings.HasPrefix(url, defaultURLPrefix) {
		url = ""
	}
	return &rpc.InstalledPlatformReference{
		Id:         release.Platform.String(),
		Version:    release.Version.String(),
		InstallDir: release.InstallDir.String(),
		PackageUrl: url,
	}
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
		Name:      release.Name,
	})
}

// HasMetadata returns true if the PlatformRelease installation dir contains the installed.json file
func (release *PlatformRelease) HasMetadata() bool {
	if release.InstallDir == nil {
		return false
	}

	installedJSONPath := release.InstallDir.Join("installed.json")
	return installedJSONPath.Exist()
}

// IsCompatible returns true if all the tools dependencies of a PlatformRelease
// are available in the current OS/ARCH.
func (release *PlatformRelease) IsCompatible() bool {
	return release.Compatible
}
