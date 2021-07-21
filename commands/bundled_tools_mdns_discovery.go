// This file is part of arduino-cli.
//
// Copyright 2021 ARDUINO SA (http://www.arduino.cc/)
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
	"fmt"

	"github.com/arduino/arduino-cli/arduino/cores"
	"github.com/arduino/arduino-cli/arduino/cores/packagemanager"
	"github.com/arduino/arduino-cli/arduino/resources"
	semver "go.bug.st/relaxed-semver"
)

var (
	mdnsDiscoveryVersion = semver.ParseRelaxed("0.9.1")
	mdnsDiscoveryFlavors = []*cores.Flavor{
		{
			OS: "i686-pc-linux-gnu",
			Resource: &resources.DownloadResource{
				ArchiveFileName: fmt.Sprintf("mdns-discovery_%s_Linux_32bit.tar.bz2", mdnsDiscoveryVersion),
				URL:             fmt.Sprintf("https://downloads.arduino.cc/discovery/mdns-discovery/mdns-discovery_%s_Linux_32bit.tar.gz", mdnsDiscoveryVersion),
				Size:            2414966,
				Checksum:        "SHA-256:47e9e35544c3d0f105926ec31ed3ded1f45cf0d50f4679bd23586bab710102fa",
				CachePath:       "tools",
			},
		},
		{
			OS: "x86_64-pc-linux-gnu",
			Resource: &resources.DownloadResource{
				ArchiveFileName: fmt.Sprintf("mdns-discovery_%s_Linux_64bit.tar.bz2", mdnsDiscoveryVersion),
				URL:             fmt.Sprintf("https://downloads.arduino.cc/discovery/mdns-discovery/mdns-discovery_%s_Linux_64bit.tar.gz", mdnsDiscoveryVersion),
				Size:            2498393,
				Checksum:        "SHA-256:014d3acd8803660ae26545e3539006d92023a16065a0c585a8ca80aabed94734",
				CachePath:       "tools",
			},
		},
		{
			OS: "i686-mingw32",
			Resource: &resources.DownloadResource{
				ArchiveFileName: fmt.Sprintf("mdns-discovery_%s_Windows_32bit.zip", mdnsDiscoveryVersion),
				URL:             fmt.Sprintf("https://downloads.arduino.cc/discovery/mdns-discovery/mdns-discovery_%s_Windows_32bit.zip", mdnsDiscoveryVersion),
				Size:            2546719,
				Checksum:        "SHA-256:fa55b412ae9d4aeaafaf3903a804b332cf8f9056c038e78369eaacdb238c9d89",
				CachePath:       "tools",
			},
		},
		{
			OS: "x86_64-mingw32",
			Resource: &resources.DownloadResource{
				ArchiveFileName: fmt.Sprintf("mdns-discovery_%s_Windows_64bit.zip", mdnsDiscoveryVersion),
				URL:             fmt.Sprintf("https://downloads.arduino.cc/discovery/mdns-discovery/mdns-discovery_%s_Windows_64bit.zip", mdnsDiscoveryVersion),
				Size:            2601308,
				Checksum:        "SHA-256:c89220e8e28ad9622a9c896f706207588aa6d3875b07a7b0bd6322e1dde0c642",
				CachePath:       "tools",
			},
		},
		{
			OS: "x86_64-apple-darwin",
			Resource: &resources.DownloadResource{
				ArchiveFileName: fmt.Sprintf("mdns-discovery_%s_macOS_64bit.tar.bz2", mdnsDiscoveryVersion),
				URL:             fmt.Sprintf("https://downloads.arduino.cc/discovery/mdns-discovery/mdns-discovery_%s_macOS_64bit.tar.gz", mdnsDiscoveryVersion),
				Size:            2458193,
				Checksum:        "SHA-256:e376408e32f79394d1124875c1e2378fe717808aef7da9a6a25086f39bc49be1",
				CachePath:       "tools",
			},
		},
		{
			OS: "arm-linux-gnueabihf",
			Resource: &resources.DownloadResource{
				ArchiveFileName: fmt.Sprintf("mdns-discovery_%s_Linux_ARMv6.tar.bz2", mdnsDiscoveryVersion),
				URL:             fmt.Sprintf("https://downloads.arduino.cc/discovery/mdns-discovery/mdns-discovery_%s_Linux_ARMv6.tar.gz", mdnsDiscoveryVersion),
				Size:            2320366,
				Checksum:        "SHA-256:531478083a6f77bff419e98fb29aaf1f72c85d14f2265799ea96e33d153c88bb",
				CachePath:       "tools",
			},
		},
		{
			OS: "arm64-linux-gnueabihf",
			Resource: &resources.DownloadResource{
				ArchiveFileName: fmt.Sprintf("mdns-discovery_%s_Linux_ARM64.tar.bz2", mdnsDiscoveryVersion),
				URL:             fmt.Sprintf("https://downloads.arduino.cc/discovery/mdns-discovery/mdns-discovery_%s_Linux_ARM64.tar.gz", mdnsDiscoveryVersion),
				Size:            2326606,
				Checksum:        "SHA-256:7b30a8c6b770e53d2579bca1443da1fbf6d182224b77b32e06cee6d08cac4595",
				CachePath:       "tools",
			},
		},
	}
)

func getBuiltinMDNSDiscoveryTool(pm *packagemanager.PackageManager) *cores.ToolRelease {
	builtinPackage := pm.Packages.GetOrCreatePackage("builtin")
	mdnsDiscoveryTool := builtinPackage.GetOrCreateTool("mdns-discovery")
	mdnsDiscoveryToolRel := mdnsDiscoveryTool.GetOrCreateRelease(mdnsDiscoveryVersion)
	mdnsDiscoveryToolRel.Flavors = mdnsDiscoveryFlavors
	return mdnsDiscoveryToolRel
}
