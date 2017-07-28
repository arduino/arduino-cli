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

package libraries

import (
	"bufio"
	"errors"

	"strings"

	"io/ioutil"

	"path/filepath"

	"os"

	"fmt"

	"github.com/bcmi-labs/arduino-cli/common"
	"github.com/bcmi-labs/arduino-cli/common/releases"
	"github.com/blang/semver"
)

// Library represents a library in the system
type Library struct {
	Name          string              `json:"name,required"`
	Author        string              `json:"author,omitempty"`
	Maintainer    string              `json:"maintainer,omitempty"`
	Sentence      string              `json:"sentence,omitempty"`
	Paragraph     string              `json:"paragraph,omitempty"`
	Website       string              `json:"website,omitempty"`
	Category      string              `json:"category,omitempty"`
	Architectures []string            `json:"architectures,omitempty"`
	Types         []string            `json:"types,omitempty"`
	Releases      map[string]*Release `json:"releases,omitempty"`
}

// InstalledRelease returns the installed release of the library.
func (l *Library) InstalledRelease() (*Release, error) {
	libFolder, err := common.GetDefaultLibFolder()
	if err != nil {
		return nil, err
	}
	files, err := ioutil.ReadDir(libFolder)
	if err != nil {
		return nil, err
	}
	// reached := false
	// QUESTION : what to do if i find multiple versions installed? @cmaglie
	// purpose : use reached variable to create a list of installed versions,
	//           it may be useful if someone accidentally put in libFolder multiple
	//           versions, to allow them to be deleted.
	for _, file := range files {
		name := strings.Replace(file.Name(), "_", " ", -1)
		// found
		// QUESTION : what to do if i find multiple versions installed? @cmaglie
		if file.IsDir() {
			// try to read library.properties
			content, err := os.Open(filepath.Join(libFolder, file.Name(), "library.properties"))
			if err != nil && strings.Contains(name, l.Name) {
				// use folder name
				version := strings.SplitN(name, "-", 2)[1] // split only once, useful for libName-1.0.0-pre-alpha/beta versions.
				return l.Releases[version], nil
			}
			defer content.Close()

			scanner := bufio.NewScanner(content)
			fields := make(map[string]string, 20)

			for scanner.Scan() {
				line := strings.SplitN(scanner.Text(), "=", 1)
				if len(line) == 2 {
					fields[line[0]] = line[1]
				}
			}
			if scanner.Err() != nil && strings.Contains(name, l.Name) {
				// use folder name
				version := strings.SplitN(name, "-", 2)[1] // split only once, useful for libName-1.0.0-pre-alpha/beta versions.
				return l.Releases[version], nil
			}

			_, nameExists := fields["name"]
			version, versionExists := fields["version"]
			if nameExists && versionExists {
				return l.GetVersion(version), nil
			} else if strings.Contains(name, l.Name) {
				// use folder name
				version = strings.SplitN(name, "-", 2)[1] // split only once, useful for libName-1.0.0-pre-alpha/beta versions.
				return l.Releases[version], nil
			}
		}
	}
	return nil, nil // no error, but not found
}

// Release represents a release of a library
type Release struct {
	Version         string `json:"version"`
	URL             string `json:"url"`
	ArchiveFileName string `json:"archiveFileName"`
	Size            int64  `json:"size"`
	Checksum        string `json:"checksum"`
}

// OpenLocalArchiveForDownload reads the data from the local archive if present,
// and returns the []byte of the file content. Used by resume Download.
// Creates an empty file if not found.
func (r Release) OpenLocalArchiveForDownload() (*os.File, error) {
	path, err := r.ArchivePath()
	if err != nil {
		return nil, err
	}
	stats, err := os.Stat(path)
	if os.IsNotExist(err) || err == nil && stats.Size() >= r.ArchiveSize() {
		return os.Create(path)
	}
	return os.OpenFile(path, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
}

// ArchivePath returns the fullPath of the Archive of this release.
func (r Release) ArchivePath() (string, error) {
	staging, err := getDownloadCacheFolder()
	if err != nil {
		return "", err
	}
	return filepath.Join(staging, r.ArchiveFileName), nil
}

// CheckLocalArchive check for integrity of the local archive.
func (r Release) CheckLocalArchive() error {
	archivePath, err := r.ArchivePath()
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
	if stats.Size() > r.ArchiveSize() {
		return errors.New("Archive size does not match with specification of this release, assuming corruption")
	}
	if !r.checksumMatches() {
		return errors.New("Checksum does not match, assuming corruption")
	}
	return nil
}

func (r Release) checksumMatches() bool {
	return releases.Match(r)
}

// Versions returns an array of all versions available of the library
func (l Library) Versions() semver.Versions {
	res := make(semver.Versions, len(l.Releases))
	i := 0
	for version := range l.Releases {
		temp, err := semver.Make(version)
		if err == nil {
			res[i] = temp
			i++
		}
	}
	//sortutil.CiAsc(res)
	return res
}

// GetVersion returns the Release corresponding to the specified version, or
// nil if not found.
//
// If version == "latest" then release.Version contains the latest version.
func (l Library) GetVersion(version string) *Release {
	if version == "latest" {
		return l.Releases[l.latestVersion()]
	}
	return l.Releases[version]
}

// Latest obtains the latest version of a library.
func (l Library) Latest() *Release {
	return l.GetVersion(l.latestVersion())
}

// latestVersion obtains latest version number.
//
// It uses lexicographics to compare version strings.
func (l *Library) latestVersion() string {
	versions := l.Versions()
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

func (r *Release) String() string {
	return fmt.Sprintln("  Release: "+fmt.Sprint(r.Version)) +
		fmt.Sprintln("    URL: "+r.URL) +
		fmt.Sprintln("    ArchiveFileName: "+r.ArchiveFileName) +
		fmt.Sprintln("    Size: ", r.ArchiveSize()) +
		fmt.Sprintln("    Checksum: ", r.Checksum)
}

func (l Library) String() string {
	return fmt.Sprintf("Name: \"%s\"\n", l.Name) +
		fmt.Sprintln("  Author: ", l.Author) +
		fmt.Sprintln("  Maintainer: ", l.Maintainer) +
		fmt.Sprintln("  Sentence: ", l.Sentence) +
		fmt.Sprintln("  Paragraph: ", l.Paragraph) +
		fmt.Sprintln("  Website: ", l.Website) +
		fmt.Sprintln("  Category: ", l.Category) +
		fmt.Sprintln("  Architecture: ", strings.Join(l.Architectures, ", ")) +
		fmt.Sprintln("  Types: ", strings.Join(l.Types, ", ")) +
		fmt.Sprintln("  Versions: ", strings.Replace(fmt.Sprint(l.Versions()), " ", ", ", -1))
}

// Release interface implementation

// ArchiveSize returns the archive size.
func (r Release) ArchiveSize() int64 {
	return r.Size
}

// ArchiveURL returns the archive URL.
func (r Release) ArchiveURL() string {
	return r.URL
}

// GetDownloadCacheFolder returns the cache folder of this release.
// Mostly this is based on the type of release (library, core, tool)
// In this case returns libraries cache folder.
func (r Release) GetDownloadCacheFolder() (string, error) {
	return getDownloadCacheFolder()
}

// ArchiveName returns the archive file name (not the path).
func (r Release) ArchiveName() string {
	return r.ArchiveFileName
}

// ExpectedChecksum returns the expected checksum for this release.
func (r Release) ExpectedChecksum() string {
	return r.Checksum
}
