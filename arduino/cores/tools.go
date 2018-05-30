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
	"regexp"
	"runtime"
	"strings"

	properties "github.com/arduino/go-properties-map"
	"github.com/bcmi-labs/arduino-cli/arduino/resources"

	"github.com/blang/semver"
)

// Tool represents a single Tool, part of a Package.
type Tool struct {
	Name     string                  `json:"name,required"` // The Name of the Tool.
	Releases map[string]*ToolRelease `json:"releases"`      //Maps Version to Release.
	Package  *Package                `json:"-"`
}

// ToolRelease represents a single release of a tool
type ToolRelease struct {
	Version  string     `json:"version,required"` // The version number of this Release.
	Flavours []*Flavour `json:"systems"`          // Maps OS to Flavour
	Tool     *Tool      `json:"-"`
	Folder   string     `json:"-"`
}

// Flavour represents a flavour of a Tool version.
type Flavour struct {
	OS       string `json:"os,required"` // The OS Supported by this flavour.
	Resource *resources.DownloadResource
}

// GetOrCreateRelease returns the ToolRelease object with the specified version
// or creates a new one if not found
func (tool *Tool) GetOrCreateRelease(version string) *ToolRelease {
	if release, ok := tool.Releases[version]; ok {
		return release
	}
	release := &ToolRelease{
		Version: version,
		Tool:    tool,
	}
	tool.Releases[version] = release
	return release
}

// GetRelease returns the specified release corresponding the provided version,
// or nil if not found.
func (tool *Tool) GetRelease(version string) *ToolRelease {
	return tool.Releases[version]
}

// GetAllReleasesVersions returns all the version numbers in this Core Package.
func (tool *Tool) GetAllReleasesVersions() semver.Versions {
	releases := tool.Releases
	versions := make(semver.Versions, 0, len(releases))
	for _, release := range releases {
		temp, err := semver.Make(release.Version)
		if err == nil {
			versions = append(versions, temp)
		}
	}

	return versions
}

// LatestRelease obtains latest version of a core package.
func (tool *Tool) LatestRelease() *ToolRelease {
	return tool.GetRelease(tool.latestReleaseVersion())
}

// latestReleaseVersion obtains latest version number.
func (tool *Tool) latestReleaseVersion() string {
	versions := tool.GetAllReleasesVersions()
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

// GetLatestInstalled returns the latest installed ToolRelease for the Tool, or nil if no releases are installed.
func (tool *Tool) GetLatestInstalled() *ToolRelease {
	var latest *ToolRelease
	for _, release := range tool.Releases {
		if release.IsInstalled() {
			if latest == nil {
				latest = release
			}
			latestVer, _ := semver.Make(latest.Version)
			releaseVer, _ := semver.Make(release.Version)
			if latestVer.LT(releaseVer) {
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
	return tr.Folder != ""
}

func (tr *ToolRelease) String() string {
	return tr.Tool.String() + "@" + tr.Version
}

// RuntimeProperties returns the runtime properties for this tool
func (tr *ToolRelease) RuntimeProperties() properties.Map {
	return properties.Map{
		"runtime.tools." + tr.Tool.Name + ".path":                    tr.Folder,
		"runtime.tools." + tr.Tool.Name + "-" + tr.Version + ".path": tr.Folder,
	}
}

var (
	// Raspberry PI, BBB or other ARM based host
	// PI: "arm-linux-gnueabihf"
	// Arch-linux on PI2: "armv7l-unknown-linux-gnueabihf"
	// Raspbian on PI2: "arm-linux-gnueabihf"
	// Ubuntu Mate on PI2: "arm-linux-gnueabihf"
	// Debian 7.9 on BBB: "arm-linux-gnueabihf"
	// Raspbian on PI Zero: "arm-linux-gnueabihf"
	regexpArmLinux = regexp.MustCompile("arm.*-linux-gnueabihf")
	regexpAmd64    = regexp.MustCompile("x86_64-.*linux-gnu")
	regexpi386     = regexp.MustCompile("i[3456]86-.*linux-gnu")
	regexpWindows  = regexp.MustCompile("i[3456]86-.*(mingw32|cygwin)")
	regexpMac64Bit = regexp.MustCompile("(i[3456]86|x86_64)-apple-darwin.*")
	regexpmac32Bit = regexp.MustCompile("i[3456]86-apple-darwin.*")
	regexpArmBSD   = regexp.MustCompile("arm.*-freebsd[0-9]*")
)

func (f *Flavour) isCompatibleWithCurrentMachine() bool {
	osName := runtime.GOOS
	osArch := runtime.GOARCH

	if f.OS == "all" {
		return true
	}

	if strings.Contains(osName, "linux") {
		if osArch == "arm" {
			return regexpArmLinux.MatchString(f.OS)
		} else if strings.Contains(osArch, "amd64") {
			return regexpAmd64.MatchString(f.OS)
		} else {
			return regexpi386.MatchString(f.OS)
		}
	} else if strings.Contains(osName, "windows") {
		return regexpWindows.MatchString(f.OS)
	} else if strings.Contains(osName, "darwin") {
		if strings.Contains(osArch, "x84_64") {
			return regexpMac64Bit.MatchString(f.OS)
		}
		return regexpmac32Bit.MatchString(f.OS)
	} else if strings.Contains(osName, "freebsd") {
		if osArch == "arm" {
			return regexpArmBSD.MatchString(f.OS)
		}
		genericFreeBSDexp := regexp.MustCompile(fmt.Sprintf("%s-freebsd[0-9]*", osArch))
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
