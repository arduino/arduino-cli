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

package configuration

import (
	"github.com/arduino/go-paths-helper"
)

// HardwareDirectories returns all paths that may contains hardware packages.
func (settings *Settings) HardwareDirectories() paths.PathList {
	res := paths.PathList{}

	if packagesDir := settings.PackagesDir(); packagesDir.IsDir() {
		res.Add(packagesDir)
	}

	if userDir, ok, _ := settings.GetStringOk("directories.user"); ok {
		if hwDir := paths.New(userDir, "hardware"); hwDir.IsDir() {
			res.Add(hwDir)
		}
	}

	return res
}

// IDEBuiltinLibrariesDir returns the IDE-bundled libraries path. Usually
// this directory is present in the Arduino IDE.
func (settings *Settings) IDEBuiltinLibrariesDir() *paths.Path {
	if builtinLibsDir, ok, _ := settings.GetStringOk("directories.builtin.libraries"); ok {
		return paths.New(builtinLibsDir)
	}
	return nil
}

// LibrariesDir returns the full path to the user directory containing
// custom libraries
func (settings *Settings) LibrariesDir() *paths.Path {
	return settings.UserDir().Join("libraries")
}

// UserDir returns the full path to the user directory
func (settings *Settings) UserDir() *paths.Path {
	if userDir, ok, _ := settings.GetStringOk("directories.user"); ok {
		return paths.New(userDir)
	}
	return paths.New(settings.Defaults.GetString("directories.user"))
}

// PackagesDir returns the full path to the packages folder
func (settings *Settings) PackagesDir() *paths.Path {
	return settings.DataDir().Join("packages")
}

// ProfilesCacheDir returns the full path to the profiles cache directory
// (it contains all the platforms and libraries used to compile a sketch
// using profiles)
func (settings *Settings) ProfilesCacheDir() *paths.Path {
	return settings.DataDir().Join("internal")
}

// DataDir returns the full path to the data directory
func (settings *Settings) DataDir() *paths.Path {
	if dataDir, ok, _ := settings.GetStringOk("directories.data"); ok {
		return paths.New(dataDir)
	}
	return paths.New(settings.Defaults.GetString("directories.data"))
}

// DownloadsDir returns the full path to the download cache directory
func (settings *Settings) DownloadsDir() *paths.Path {
	if downloadDir, ok, _ := settings.GetStringOk("directories.downloads"); ok {
		return paths.New(downloadDir)
	}
	return settings.DataDir().Join("staging")
}
