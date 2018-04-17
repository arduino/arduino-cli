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

package libraries

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/bcmi-labs/arduino-cli/arduino/releases"

	"github.com/bcmi-labs/arduino-cli/configs"
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

func (l *Library) String() string {
	return l.Name
}

// InstalledRelease returns the installed release of the library.
func (l *Library) InstalledRelease() (*Release, error) {
	libFolder, err := configs.LibrariesFolder.Get()
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
	Version  string `json:"version"`
	Resource *releases.DownloadResource
	Library  *Library
}

func (r *Release) String() string {
	return r.Library.String() + "@" + r.Version
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
