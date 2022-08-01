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
	"github.com/spf13/viper"
)

// HardwareDirectories returns all paths that may contains hardware packages.
func HardwareDirectories(settings *viper.Viper) paths.PathList {
	res := paths.PathList{}

	if settings.IsSet("directories.Data") {
		packagesDir := PackagesDir(Settings)
		if packagesDir.IsDir() {
			res.Add(packagesDir)
		}
	}

	if settings.IsSet("directories.User") {
		skDir := paths.New(settings.GetString("directories.User"))
		hwDir := skDir.Join("hardware")
		if hwDir.IsDir() {
			res.Add(hwDir)
		}
	}

	return res
}

// BuiltinToolsDirectories returns all paths that may contains bundled-tools.
func BuiltinToolsDirectories(settings *viper.Viper) paths.PathList {
	return paths.NewPathList(settings.GetStringSlice("directories.builtin.Tools")...)
}

// IDEBuiltinLibrariesDir returns the IDE-bundled libraries paths. Usually
// one of these directories is present in the Arduino IDE.
func IDEBuiltinLibrariesDir(settings *viper.Viper) paths.PathList {
	return paths.NewPathList(Settings.GetStringSlice("directories.builtin.Libraries")...)
}

// LibrariesDir returns the full path to the user directory containing
// custom libraries
func LibrariesDir(settings *viper.Viper) *paths.Path {
	return paths.New(settings.GetString("directories.User")).Join("libraries")
}

// PackagesDir returns the full path to the packages folder
func PackagesDir(settings *viper.Viper) *paths.Path {
	return paths.New(settings.GetString("directories.Data")).Join("packages")
}

// ProfilesCacheDir returns the full path to the profiles cache directory
// (it contains all the platforms and libraries used to compile a sketch
// using profiles)
func ProfilesCacheDir(settings *viper.Viper) *paths.Path {
	return paths.New(settings.GetString("directories.Data")).Join("internal")
}
