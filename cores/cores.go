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
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/bcmi-labs/arduino-cli/common/checksums"
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
	Dependencies    ToolDependencies // The Dependency entries to load tools.
}

// ToolDependencies is a set of tuples representing summary data of a tool.
type ToolDependencies []toolDependency

type toolDependency struct {
	ToolPackager string
	ToolName     string
	ToolVersion  string
}

// GetVersion returns the specified release corresponding the provided version,
// or nil if not found.
func (core *Core) GetVersion(version string) *Release {
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

// Latest obtains latest version of a core package.
func (core *Core) Latest() *Release {
	return core.GetVersion(core.latestVersion())
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
	res := fmt.Sprintln("Name        :", core.Name)
	res += fmt.Sprintln("Architecture:", core.Architecture)
	res += fmt.Sprintln("Category    :", core.Category)
	if core.Releases != nil && len(core.Releases) > 0 {
		res += "Releases:\n"
		for _, release := range core.Releases {
			res += fmt.Sprintln(release)
		}
	}
	return res
}

func (release *Release) String() string {
	res := fmt.Sprintln("  Version           : ", release.Version)
	res += fmt.Sprintln("  Boards            :")
	res += fmt.Sprintln(strings.Join(release.Boards, ", "))
	res += fmt.Sprintln("  Archive File Name :", release.ArchiveFileName)
	res += fmt.Sprintln("  Checksum          :", release.Checksum)
	res += fmt.Sprintln("  File Size         :", release.Size)
	return res
}

// OpenLocalArchiveForDownload Creates an empty file if not found.
func (release Release) OpenLocalArchiveForDownload() (*os.File, error) {
	path, err := release.ArchivePath()
	if err != nil {
		return nil, err
	}
	stats, err := os.Stat(path)
	if os.IsNotExist(err) || err == nil && stats.Size() >= release.Size {
		return os.Create(path)
	}
	return os.OpenFile(path, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
}

// ArchivePath returns the fullPath of the Archive of this release.
func (release Release) ArchivePath() (string, error) {
	staging, err := getDownloadCacheFolder()
	if err != nil {
		return "", err
	}
	return filepath.Join(staging, release.ArchiveFileName), nil
}

// CheckLocalArchive check for integrity of the local archive.
func (release Release) CheckLocalArchive() error {
	archivePath, err := release.ArchivePath()
	if err != nil {
		return err
	}
	stats, err := os.Stat(archivePath)
	if os.IsNotExist(err) {
		return errors.New("Archive does not exist")
	}
	if err != nil {
		return err
	}
	if stats.Size() > release.Size {
		return errors.New("Archive size does not match with specification of this release, assuming corruption")
	}
	if !release.checksumMatches() {
		return errors.New("Checksum does not match, assuming corruption")
	}
	return nil
}

func (release Release) checksumMatches() bool {
	return checksums.Match(release)
}

// ExpectedChecksum returns the expected checksum for this release.
func (release Release) ExpectedChecksum() string {
	return release.Checksum
}
