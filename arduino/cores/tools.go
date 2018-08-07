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
	"regexp"
	"runtime"

	"github.com/arduino/go-paths-helper"
	properties "github.com/arduino/go-properties-map"
	"github.com/bcmi-labs/arduino-cli/arduino/resources"
	"go.bug.st/relaxed-semver"
)

// Tool represents a single Tool, part of a Package.
type Tool struct {
	Name     string                  `json:"name,required"` // The Name of the Tool.
	Releases map[string]*ToolRelease `json:"releases"`      //Maps Version to Release.
	Package  *Package                `json:"-"`
}

// ToolRelease represents a single release of a tool
type ToolRelease struct {
	Version    *semver.RelaxedVersion `json:"version,required"` // The version number of this Release.
	Flavours   []*Flavour             `json:"systems"`          // Maps OS to Flavour
	Tool       *Tool                  `json:"-"`
	InstallDir *paths.Path            `json:"-"`
}

// Flavour represents a flavour of a Tool version.
type Flavour struct {
	OS       string `json:"os,required"` // The OS Supported by this flavour.
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

// GetRelease returns the specified release corresponding the provided version,
// or nil if not found.
func (tool *Tool) GetRelease(version *semver.RelaxedVersion) *ToolRelease {
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
	if latest := tool.latestReleaseVersion(); latest == nil {
		return nil
	} else {
		return tool.GetRelease(latest)
	}
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
func (tr *ToolRelease) RuntimeProperties() properties.Map {
	return properties.Map{
		"runtime.tools." + tr.Tool.Name + ".path":                             tr.InstallDir.String(),
		"runtime.tools." + tr.Tool.Name + "-" + tr.Version.String() + ".path": tr.InstallDir.String(),
	}
}

var (
	regexpArmLinux = regexp.MustCompile("arm.*-linux-gnueabihf")
	regexpAmd64    = regexp.MustCompile("x86_64-.*linux-gnu")
	regexpi386     = regexp.MustCompile("i[3456]86-.*linux-gnu")
	regexpWindows  = regexp.MustCompile("i[3456]86-.*(mingw32|cygwin)")
	regexpMac64Bit = regexp.MustCompile("(i[3456]86|x86_64)-apple-darwin.*")
	regexpmac32Bit = regexp.MustCompile("i[3456]86-apple-darwin.*")
	regexpArmBSD   = regexp.MustCompile("arm.*-freebsd[0-9]*")
)

func (f *Flavour) isCompatibleWithCurrentMachine() bool {
	return f.isCompatibleWith(runtime.GOOS, runtime.GOARCH)
}

func (f *Flavour) isCompatibleWith(osName, osArch string) bool {
	if f.OS == "all" {
		return true
	}

	switch osName + "," + osArch {
	case "linux,arm", "linux,armbe":
		return regexpArmLinux.MatchString(f.OS)
	case "linux,amd64":
		return regexpAmd64.MatchString(f.OS)
	case "linux,i386":
		return regexpi386.MatchString(f.OS)
	case "windows,i386", "windows,amd64":
		return regexpWindows.MatchString(f.OS)
	case "darwin,amd64":
		return regexpmac32Bit.MatchString(f.OS) || regexpMac64Bit.MatchString(f.OS)
	case "darwin,i386":
		return regexpmac32Bit.MatchString(f.OS)
	case "freebsd,arm":
		return regexpArmBSD.MatchString(f.OS)
	case "freebsd,i386", "freebsd,amd64":
		genericFreeBSDexp := regexp.MustCompile(osArch + "%s-freebsd[0-9]*")
		return genericFreeBSDexp.MatchString(f.OS)
	}
	return false
}

// GetCompatibleFlavour returns the downloadable resource compatible with the running O.S.
func (tr *ToolRelease) GetCompatibleFlavour() *resources.DownloadResource {
	for _, flavour := range tr.Flavours {
		if flavour.isCompatibleWithCurrentMachine() {
			return flavour.Resource
		}
	}
	return nil
}
