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
	"regexp"
	"runtime"

	"github.com/arduino/arduino-cli/arduino/resources"
	"github.com/arduino/go-paths-helper"
	properties "github.com/arduino/go-properties-orderedmap"
	semver "go.bug.st/relaxed-semver"
)

// Tool represents a single Tool, part of a Package.
type Tool struct {
	Name     string                  `json:"name"`     // The Name of the Tool.
	Releases map[string]*ToolRelease `json:"releases"` // Maps Version to Release.
	Package  *Package                `json:"-"`
}

// ToolRelease represents a single release of a tool
type ToolRelease struct {
	Version    *semver.RelaxedVersion `json:"version"` // The version number of this Release.
	Flavors    []*Flavor              `json:"systems"` // Maps OS to Flavor
	Tool       *Tool                  `json:"-"`
	InstallDir *paths.Path            `json:"-"`
}

// Flavor represents a flavor of a Tool version.
type Flavor struct {
	OS       string `json:"os"` // The OS Supported by this flavor.
	Resource *resources.DownloadResource
}

// GetOrCreateRelease returns the ToolRelease object with the specified version
// or creates a new one if not found
func (tool *Tool) GetOrCreateRelease(version *semver.RelaxedVersion) *ToolRelease {
	if release, ok := tool.Releases[version.String()]; ok {
		return release
	}
	release := &ToolRelease{
		Version: version,
		Tool:    tool,
	}
	tool.Releases[version.String()] = release
	return release
}

// FindReleaseWithRelaxedVersion returns the specified release corresponding the provided version,
// or nil if not found.
func (tool *Tool) FindReleaseWithRelaxedVersion(version *semver.RelaxedVersion) *ToolRelease {
	return tool.Releases[version.String()]
}

// GetAllReleasesVersions returns all the version numbers in this Core Package.
func (tool *Tool) GetAllReleasesVersions() []*semver.RelaxedVersion {
	versions := []*semver.RelaxedVersion{}
	for _, release := range tool.Releases {
		versions = append(versions, release.Version)
	}

	return versions
}

// LatestRelease obtains latest version of a core package.
func (tool *Tool) LatestRelease() *ToolRelease {
	var latest *ToolRelease
	for _, release := range tool.Releases {
		if latest == nil || release.Version.GreaterThan(latest.Version) {
			latest = release
		}
	}
	return latest
}

// GetLatestInstalled returns the latest installed ToolRelease for the Tool, or nil if no releases are installed.
func (tool *Tool) GetLatestInstalled() *ToolRelease {
	var latest *ToolRelease
	for _, release := range tool.Releases {
		if release.IsInstalled() {
			if latest == nil || latest.Version.LessThan(release.Version) {
				latest = release
			}
		}
	}
	return latest
}

func (tool *Tool) String() string {
	return tool.Package.Name + ":" + tool.Name
}

// IsInstalled returns true if the ToolRelease is installed
func (tr *ToolRelease) IsInstalled() bool {
	return tr.InstallDir != nil
}

func (tr *ToolRelease) String() string {
	return tr.Tool.String() + "@" + tr.Version.String()
}

// RuntimeProperties returns the runtime properties for this tool
func (tr *ToolRelease) RuntimeProperties() *properties.Map {
	res := properties.NewMap()
	if tr.IsInstalled() {
		res.Set("runtime.tools."+tr.Tool.Name+".path", tr.InstallDir.String())
		res.Set("runtime.tools."+tr.Tool.Name+"-"+tr.Version.String()+".path", tr.InstallDir.String())
	}
	return res
}

var (
	regexpLinuxArm   = regexp.MustCompile("arm.*-linux-gnueabihf")
	regexpLinuxArm64 = regexp.MustCompile("(aarch64|arm64)-linux-gnu")
	regexpLinux64    = regexp.MustCompile("x86_64-.*linux-gnu")
	regexpLinux32    = regexp.MustCompile("i[3456]86-.*linux-gnu")
	regexpWindows32  = regexp.MustCompile("i[3456]86-.*(mingw32|cygwin)")
	regexpWindows64  = regexp.MustCompile("(amd64|x86_64)-.*(mingw32|cygwin)")
	regexpMac64      = regexp.MustCompile("x86_64-apple-darwin.*")
	regexpMac32      = regexp.MustCompile("i[3456]86-apple-darwin.*")
	regexpMacArm64   = regexp.MustCompile("arm64-apple-darwin.*")
	regexpFreeBSDArm = regexp.MustCompile("arm.*-freebsd[0-9]*")
	regexpFreeBSD32  = regexp.MustCompile("i?[3456]86-freebsd[0-9]*")
	regexpFreeBSD64  = regexp.MustCompile("amd64-freebsd[0-9]*")
)

func (f *Flavor) isExactMatchWith(osName, osArch string) bool {
	if f.OS == "all" {
		return true
	}

	switch osName + "," + osArch {
	case "linux,arm", "linux,armbe":
		return regexpLinuxArm.MatchString(f.OS)
	case "linux,arm64":
		return regexpLinuxArm64.MatchString(f.OS)
	case "linux,amd64":
		return regexpLinux64.MatchString(f.OS)
	case "linux,386":
		return regexpLinux32.MatchString(f.OS)
	case "windows,386":
		return regexpWindows32.MatchString(f.OS)
	case "windows,amd64":
		return regexpWindows64.MatchString(f.OS)
	case "darwin,arm64":
		return regexpMacArm64.MatchString(f.OS)
	case "darwin,amd64":
		return regexpMac64.MatchString(f.OS)
	case "darwin,386":
		return regexpMac32.MatchString(f.OS)
	case "freebsd,arm":
		return regexpFreeBSDArm.MatchString(f.OS)
	case "freebsd,386":
		return regexpFreeBSD32.MatchString(f.OS)
	case "freebsd,amd64":
		return regexpFreeBSD64.MatchString(f.OS)
	}
	return false
}

func (f *Flavor) isCompatibleWith(osName, osArch string) (bool, int) {
	if f.isExactMatchWith(osName, osArch) {
		return true, 1000
	}

	switch osName + "," + osArch {
	case "windows,amd64":
		return regexpWindows32.MatchString(f.OS), 10
	case "darwin,amd64":
		return regexpMac32.MatchString(f.OS), 10
	case "darwin,arm64":
		// Compatibility guaranteed through Rosetta emulation
		if regexpMac64.MatchString(f.OS) {
			// Prefer amd64 version if available
			return true, 20
		}
		return regexpMac32.MatchString(f.OS), 10
	}

	return false, 0
}

// GetCompatibleFlavour returns the downloadable resource compatible with the running O.S.
func (tr *ToolRelease) GetCompatibleFlavour() *resources.DownloadResource {
	return tr.GetFlavourCompatibleWith(runtime.GOOS, runtime.GOARCH)
}

// GetFlavourCompatibleWith returns the downloadable resource compatible with the specified O.S.
func (tr *ToolRelease) GetFlavourCompatibleWith(osName, osArch string) *resources.DownloadResource {
	var resource *resources.DownloadResource
	priority := -1
	for _, flavour := range tr.Flavors {
		if comp, p := flavour.isCompatibleWith(osName, osArch); comp && p > priority {
			resource = flavour.Resource
			priority = p
		}
	}
	return resource
}
