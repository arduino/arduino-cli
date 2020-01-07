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

package commands

import (
	"github.com/arduino/arduino-cli/arduino/cores"
	"github.com/arduino/arduino-cli/arduino/cores/packagemanager"
	"github.com/arduino/arduino-cli/arduino/resources"
	semver "go.bug.st/relaxed-semver"
)

func loadBuiltinCtagsMetadata(pm *packagemanager.PackageManager) {
	builtinPackage := pm.Packages.GetOrCreatePackage("builtin")
	ctagsTool := builtinPackage.GetOrCreateTool("ctags")
	ctagsRel := ctagsTool.GetOrCreateRelease(semver.ParseRelaxed("5.8-arduino11"))
	ctagsRel.Flavors = []*cores.Flavor{
		{
			OS: "i686-pc-linux-gnu",
			Resource: &resources.DownloadResource{
				ArchiveFileName: "ctags-5.8-arduino11-pm-i686-pc-linux-gnu.tar.bz2",
				URL:             "https://downloads.arduino.cc/tools/ctags-5.8-arduino11-pm-i686-pc-linux-gnu.tar.bz2",
				Size:            106930,
				Checksum:        "SHA-256:3e219116f54d9035f6c49c600d0bb358710dcce139798ad237de0eef7998d0e2",
				CachePath:       "tools",
			},
		},
		{
			OS: "x86_64-pc-linux-gnu",
			Resource: &resources.DownloadResource{
				ArchiveFileName: "ctags-5.8-arduino11-pm-x86_64-pc-linux-gnu.tar.bz2",
				URL:             "https://downloads.arduino.cc/tools/ctags-5.8-arduino11-pm-x86_64-pc-linux-gnu.tar.bz2",
				Size:            111604,
				Checksum:        "SHA-256:62b514f3aaf37b5429ef703853bb46365fb91b4754c1916d085bf134004886e3",
				CachePath:       "tools",
			},
		},
		{
			OS: "i686-mingw32",
			Resource: &resources.DownloadResource{
				ArchiveFileName: "ctags-5.8-arduino11-pm-i686-mingw32.zip",
				URL:             "https://downloads.arduino.cc/tools/ctags-5.8-arduino11-pm-i686-mingw32.zip",
				Size:            116455,
				Checksum:        "SHA-256:106c9f074a3e2ec55bd8a461c1522bb4c90488275f061c3d51942862c99b8ba7",
				CachePath:       "tools",
			},
		},
		{
			OS: "x86_64-apple-darwin",
			Resource: &resources.DownloadResource{
				ArchiveFileName: "ctags-5.8-arduino11-pm-x86_64-apple-darwin.zip",
				URL:             "https://downloads.arduino.cc/tools/ctags-5.8-arduino11-pm-x86_64-apple-darwin.zip",
				Size:            118296,
				Checksum:        "SHA-256:bf74807260ddf0e1fc94d67e9cd8a41b7c0a2f0bee03e254e6975139b37ef250",
				CachePath:       "tools",
			},
		},
		{
			OS: "arm-linux-gnueabihf",
			Resource: &resources.DownloadResource{
				ArchiveFileName: "ctags-5.8-arduino11-pm-armv6-linux-gnueabihf.tar.bz2",
				URL:             "https://downloads.arduino.cc/tools/ctags-5.8-arduino11-pm-armv6-linux-gnueabihf.tar.bz2",
				Size:            95271,
				Checksum:        "SHA-256:098e8dc6a7dc0ddf09ef28e6e2e81d2ae181d12f41fb182dd78ff1387d4eb285",
				CachePath:       "tools",
			},
		},
		{
			OS: "aarch64-linux-gnu",
			Resource: &resources.DownloadResource{
				ArchiveFileName: "ctags-5.8-arduino11-pm-aarch64-linux-gnu.tar.bz2",
				URL:             "https://downloads.arduino.cc/tools/ctags-5.8-arduino11-pm-aarch64-linux-gnu.tar.bz2",
				Size:            100819,
				Checksum:        "SHA-256:89b6f8adb7b7124ebe2eb30970ea77c754cd2239e0d8a6b0102ae2a36416c6ef",
				CachePath:       "tools",
			},
		},
	}
}

var ctagsVersion = semver.ParseRelaxed("5.8-arduino11")

func getBuiltinCtagsTool(pm *packagemanager.PackageManager) (*cores.ToolRelease, error) {
	loadBuiltinCtagsMetadata(pm)
	return pm.Package("builtin").Tool("ctags").Release(ctagsVersion).Get()
}
