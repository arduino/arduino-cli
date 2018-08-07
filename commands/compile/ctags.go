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
 * Copyright 2017-2018 ARDUINO AG (http://www.arduino.cc/)
 */

package compile

import (
	"github.com/bcmi-labs/arduino-cli/arduino/cores"
	"github.com/bcmi-labs/arduino-cli/arduino/cores/packagemanager"
	"github.com/bcmi-labs/arduino-cli/arduino/resources"
	"go.bug.st/relaxed-semver"
)

func loadBuiltinCtagsMetadata(pm *packagemanager.PackageManager) {
	builtinPackage := pm.GetPackages().GetOrCreatePackage("builtin")
	ctagsTool := builtinPackage.GetOrCreateTool("ctags")
	ctagsRel := ctagsTool.GetOrCreateRelease(semver.ParseRelaxed("5.8-arduino11"))
	ctagsRel.Flavours = []*cores.Flavour{
		&cores.Flavour{
			OS: "i686-pc-linux-gnu",
			Resource: &resources.DownloadResource{
				ArchiveFileName: "ctags-5.8-arduino11-pm-i686-pc-linux-gnu.tar.bz2",
				URL:             "http://downloads.arduino.cc/tools/ctags-5.8-arduino11-pm-i686-pc-linux-gnu.tar.bz2",
				Size:            106930,
				Checksum:        "SHA-256:3e219116f54d9035f6c49c600d0bb358710dcce139798ad237de0eef7998d0e2",
				CachePath:       "tools",
			},
		},
		&cores.Flavour{
			OS: "x86_64-pc-linux-gnu",
			Resource: &resources.DownloadResource{
				ArchiveFileName: "ctags-5.8-arduino11-pm-x86_64-pc-linux-gnu.tar.bz2",
				URL:             "http://downloads.arduino.cc/tools/ctags-5.8-arduino11-pm-x86_64-pc-linux-gnu.tar.bz2",
				Size:            111604,
				Checksum:        "SHA-256:62b514f3aaf37b5429ef703853bb46365fb91b4754c1916d085bf134004886e3",
				CachePath:       "tools",
			},
		},
		&cores.Flavour{
			OS: "i686-mingw32",
			Resource: &resources.DownloadResource{
				ArchiveFileName: "ctags-5.8-arduino11-pm-i686-mingw32.zip",
				URL:             "http://downloads.arduino.cc/tools/ctags-5.8-arduino11-pm-i686-mingw32.zip",
				Size:            116455,
				Checksum:        "SHA-256:106c9f074a3e2ec55bd8a461c1522bb4c90488275f061c3d51942862c99b8ba7",
				CachePath:       "tools",
			},
		},
		&cores.Flavour{
			OS: "x86_64-apple-darwin",
			Resource: &resources.DownloadResource{
				ArchiveFileName: "ctags-5.8-arduino11-pm-x86_64-apple-darwin.zip",
				URL:             "http://downloads.arduino.cc/tools/ctags-5.8-arduino11-pm-x86_64-apple-darwin.zip",
				Size:            107801,
				Checksum:        "SHA-256:e8c5db45867fb5987ad0fc429d8bbbcf5549f4b7c2b5ade863e76ac77255144d",
				CachePath:       "tools",
			},
		},
		&cores.Flavour{
			OS: "arm-linux-gnueabihf",
			Resource: &resources.DownloadResource{
				ArchiveFileName: "ctags-5.8-arduino11-pm-armv6-linux-gnueabihf.tar.bz2",
				URL:             "http://downloads.arduino.cc/tools/ctags-5.8-arduino11-pm-armv6-linux-gnueabihf.tar.bz2",
				Size:            95271,
				Checksum:        "SHA-256:098e8dc6a7dc0ddf09ef28e6e2e81d2ae181d12f41fb182dd78ff1387d4eb285",
				CachePath:       "tools",
			},
		},
	}
}

var ctagsVersion = semver.ParseRelaxed("5.8-arduino11")

func getBuiltinCtagsTool(pm *packagemanager.PackageManager) (*cores.ToolRelease, error) {
	return pm.Package("builtin").Tool("ctags").Release(ctagsVersion).Get()
}
