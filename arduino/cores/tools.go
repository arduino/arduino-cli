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
	"regexp"
	"runtime"

	"github.com/arduino/arduino-cli/arduino/resources"
	"github.com/arduino/go-paths-helper"
	properties "github.com/arduino/go-properties-orderedmap"
	semver "go.bug.st/relaxed-semver"
)

// Tool represents a single Tool, part of a Package.
type Tool struct {
	Name     string                  `json:"name,required"` // The Name of the Tool.
	Releases map[string]*ToolRelease `json:"releases"`      // Maps Version to Release.
	Package  *Package                `json:"-"`
}

// ToolRelease represents a single release of a tool
type ToolRelease struct {
	Version    *semver.RelaxedVersion `json:"version,required"` // The version number of this Release.
	Flavors    []*Flavor              `json:"systems"`          // Maps OS to Flavor
	Tool       *Tool                  `json:"-"`
	InstallDir *paths.Path            `json:"-"`
}

// Flavor represents a flavor of a Tool version.
type Flavor struct {
	OS       string `json:"os,required"` // The OS Supported by this flavor.
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
	releases := tool.Releases
	versions := []*semver.RelaxedVersion{}
	for _, release := range releases {
		versions = append(versions, release.Version)
	}

	return versions
}

// LatestRelease obtains latest version of a core package.
func (tool *Tool) LatestRelease() *ToolRelease {
	latest := tool.latestReleaseVersion()
	if latest == nil {
		return nil
	}

	return tool.FindReleaseWithRelaxedVersion(latest)
}

// latestReleaseVersion obtains latest version number.
func (tool *Tool) latestReleaseVersion() *semver.RelaxedVersion {
	versions := tool.GetAllReleasesVersions()
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

// GetLatestInstalled returns the latest installed ToolRelease for the Tool, or nil if no releases are installed.
func (tool *Tool) GetLatestInstalled() *ToolRelease {
	var latest *ToolRelease
	for _, release := range tool.Releases {
		if release.IsInstalled() {
			if latest == nil {
				latest = release
			} else if latest.Version.LessThan(release.Version) {
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
	res.Set("runtime.tools."+tr.Tool.Name+".path", tr.InstallDir.String())
	res.Set("runtime.tools."+tr.Tool.Name+"-"+tr.Version.String()+".path", tr.InstallDir.String())
	return res
}

var (
	regexpArmLinux   = regexp.MustCompile("arm.*-linux-gnueabihf")
	regexpArm64Linux = regexp.MustCompile("(aarch64|arm64)-linux-gnu")
	regexpAmd64      = regexp.MustCompile("x86_64-.*linux-gnu")
	regexpi386       = regexp.MustCompile("i[3456]86-.*linux-gnu")
	regexpWindows    = regexp.MustCompile("i[3456]86-.*(mingw32|cygwin)")
	regexpMac64Bit   = regexp.MustCompile("(i[3456]86|x86_64)-apple-darwin.*")
	regexpmac32Bit   = regexp.MustCompile("i[3456]86-apple-darwin.*")
	regexpArmBSD     = regexp.MustCompile("arm.*-freebsd[0-9]*")
)

func (f *Flavor) isCompatibleWithCurrentMachine() bool {
	return f.isCompatibleWith(runtime.GOOS, runtime.GOARCH)
}

func (f *Flavor) isCompatibleWith(osName, osArch string) bool {
	if f.OS == "all" {
		return true
	}

	switch osName + "," + osArch {
	case "linux,arm", "linux,armbe":
		return regexpArmLinux.MatchString(f.OS)
	case "linux,arm64":
		return regexpArm64Linux.MatchString(f.OS)
	case "linux,amd64":
		return regexpAmd64.MatchString(f.OS)
	case "linux,386":
		return regexpi386.MatchString(f.OS)
	case "windows,386", "windows,amd64":
		return regexpWindows.MatchString(f.OS)
	case "darwin,amd64":
		return regexpmac32Bit.MatchString(f.OS) || regexpMac64Bit.MatchString(f.OS)
	case "darwin,386":
		return regexpmac32Bit.MatchString(f.OS)
	case "freebsd,arm":
		return regexpArmBSD.MatchString(f.OS)
	case "freebsd,386", "freebsd,amd64":
		genericFreeBSDexp := regexp.MustCompile(osArch + "%s-freebsd[0-9]*")
		return genericFreeBSDexp.MatchString(f.OS)
	}
	return false
}

// GetCompatibleFlavour returns the downloadable resource compatible with the running O.S.
func (tr *ToolRelease) GetCompatibleFlavour() *resources.DownloadResource {
	for _, flavour := range tr.Flavors {
		if flavour.isCompatibleWithCurrentMachine() {
			return flavour.Resource
		}
	}
	return nil
}
