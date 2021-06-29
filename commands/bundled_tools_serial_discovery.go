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
	"fmt"

	"github.com/arduino/arduino-cli/arduino/cores"
	"github.com/arduino/arduino-cli/arduino/cores/packagemanager"
	"github.com/arduino/arduino-cli/arduino/resources"
	semver "go.bug.st/relaxed-semver"
)

var (
	serialDiscoveryVersion = semver.ParseRelaxed("1.3.0-rc1")
	serialDiscoveryFlavors = []*cores.Flavor{
		{
			OS: "i686-pc-linux-gnu",
			Resource: &resources.DownloadResource{
				ArchiveFileName: fmt.Sprintf("serial-discovery_v%s_Linux_32bit.tar.bz2", serialDiscoveryVersion),
				URL:             fmt.Sprintf("https://downloads.arduino.cc/discovery/serial-discovery/serial-discovery_v%s_Linux_32bit.tar.gz", serialDiscoveryVersion),
				Size:            1633143,
				Checksum:        "SHA-256:2fb17882018f3eefeaf933673cbc42cea83ce739503880ccc7f9cf521de0e513",
				CachePath:       "tools",
			},
		},
		{
			OS: "x86_64-pc-linux-gnu",
			Resource: &resources.DownloadResource{
				ArchiveFileName: fmt.Sprintf("serial-discovery_v%s_Linux_64bit.tar.bz2", serialDiscoveryVersion),
				URL:             fmt.Sprintf("https://downloads.arduino.cc/discovery/serial-discovery/serial-discovery_v%s_Linux_64bit.tar.gz", serialDiscoveryVersion),
				Size:            1688362,
				Checksum:        "SHA-256:e0e55ea9c5e05f12af5d89dc3a69d63e12211f54122b4bf45a7cab9f0a6f89e5",
				CachePath:       "tools",
			},
		},
		{
			OS: "i686-mingw32",
			Resource: &resources.DownloadResource{
				ArchiveFileName: fmt.Sprintf("serial-discovery_v%s_Windows_32bit.zip", serialDiscoveryVersion),
				URL:             fmt.Sprintf("https://downloads.arduino.cc/discovery/serial-discovery/serial-discovery_v%s_Windows_32bit.zip", serialDiscoveryVersion),
				Size:            1742668,
				Checksum:        "SHA-256:4acfe521d6fc3b29643ab69ced246d7dd20637772fc79fc3e509829c18290d90",
				CachePath:       "tools",
			},
		},
		{
			OS: "x86_64-mingw32",
			Resource: &resources.DownloadResource{
				ArchiveFileName: fmt.Sprintf("serial-discovery_v%s_Windows_64bit.zip", serialDiscoveryVersion),
				URL:             fmt.Sprintf("https://downloads.arduino.cc/discovery/serial-discovery/serial-discovery_v%s_Windows_64bit.zip", serialDiscoveryVersion),
				Size:            1709333,
				Checksum:        "SHA-256:82b2edea04f7c97b98cbb04de95ec48be95de64fa5f196d730dc824d7558b952",
				CachePath:       "tools",
			},
		},
		{
			OS: "x86_64-apple-darwin",
			Resource: &resources.DownloadResource{
				ArchiveFileName: fmt.Sprintf("serial-discovery_v%s_macOS_64bit.tar.bz2", serialDiscoveryVersion),
				URL:             fmt.Sprintf("https://downloads.arduino.cc/discovery/serial-discovery/serial-discovery_v%s_macOS_64bit.tar.gz", serialDiscoveryVersion),
				Size:            964596,
				Checksum:        "SHA-256:ec4be0f5c1ed6af3f31bb01ed6a5433274a76a1dc7cb68d39813b2b0475d7337",
				CachePath:       "tools",
			},
		},
		{
			OS: "arm-linux-gnueabihf",
			Resource: &resources.DownloadResource{
				ArchiveFileName: fmt.Sprintf("serial-discovery_v%s_Linux_ARMv6.tar.bz2", serialDiscoveryVersion),
				URL:             fmt.Sprintf("https://downloads.arduino.cc/discovery/serial-discovery/serial-discovery_v%s_Linux_ARMv6.tar.gz", serialDiscoveryVersion),
				Size:            1570847,
				Checksum:        "SHA-256:9341e2541ad41ee2cdaad1e8d851254c8bce63c937cdafd57db7d1439d8ced59",
				CachePath:       "tools",
			},
		},
		{
			OS: "arm64-linux-gnueabihf",
			Resource: &resources.DownloadResource{
				ArchiveFileName: fmt.Sprintf("serial-discovery_v%s_Linux_ARM64.tar.bz2", serialDiscoveryVersion),
				URL:             fmt.Sprintf("https://downloads.arduino.cc/discovery/serial-discovery/serial-discovery_v%s_Linux_ARM64.tar.gz", serialDiscoveryVersion),
				Size:            1580108,
				Checksum:        "SHA-256:1da38f94be8db69bbe26d6a95692b665f6bc9bf89aa62b58d4e4cfb0f7fd8733",
				CachePath:       "tools",
			},
		},
	}
)

func getBuiltinSerialDiscoveryTool(pm *packagemanager.PackageManager) *cores.ToolRelease {
	builtinPackage := pm.Packages.GetOrCreatePackage("builtin")
	serialDiscoveryTool := builtinPackage.GetOrCreateTool("serial-discovery")
	serialDiscoveryToolRel := serialDiscoveryTool.GetOrCreateRelease(serialDiscoveryVersion)
	serialDiscoveryToolRel.Flavors = serialDiscoveryFlavors
	return serialDiscoveryToolRel
}
