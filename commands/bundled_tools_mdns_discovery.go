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
	mdnsDiscoveryVersion = semver.ParseRelaxed("0.9.2")
	mdnsDiscoveryFlavors = []*cores.Flavor{
		{
			OS: "i686-pc-linux-gnu",
			Resource: &resources.DownloadResource{
				ArchiveFileName: fmt.Sprintf("mdns-discovery_%s_Linux_32bit.tar.bz2", mdnsDiscoveryVersion),
				URL:             fmt.Sprintf("https://downloads.arduino.cc/discovery/mdns-discovery/mdns-discovery_%s_Linux_32bit.tar.gz", mdnsDiscoveryVersion),
				Size:            2417970,
				Checksum:        "SHA-256:34afe67745c7d7e8d435fb668aa3e85a93d313b7a8a4bc13f7e3c5fe48db0fa8",
				CachePath:       "tools",
			},
		},
		{
			OS: "x86_64-pc-linux-gnu",
			Resource: &resources.DownloadResource{
				ArchiveFileName: fmt.Sprintf("mdns-discovery_%s_Linux_64bit.tar.bz2", mdnsDiscoveryVersion),
				URL:             fmt.Sprintf("https://downloads.arduino.cc/discovery/mdns-discovery/mdns-discovery_%s_Linux_64bit.tar.gz", mdnsDiscoveryVersion),
				Size:            2499458,
				Checksum:        "SHA-256:e97d9553590559e4b15239c4f8c86c9bd310aca3f0cdfd26098c5878b463f2ec",
				CachePath:       "tools",
			},
		},
		{
			OS: "i686-mingw32",
			Resource: &resources.DownloadResource{
				ArchiveFileName: fmt.Sprintf("mdns-discovery_%s_Windows_32bit.zip", mdnsDiscoveryVersion),
				URL:             fmt.Sprintf("https://downloads.arduino.cc/discovery/mdns-discovery/mdns-discovery_%s_Windows_32bit.zip", mdnsDiscoveryVersion),
				Size:            2548123,
				Checksum:        "SHA-256:2493290856294f8bed3eed50c9ed4cda6a4a98b4c01cae6db7c1621f4db33afd",
				CachePath:       "tools",
			},
		},
		{
			OS: "x86_64-mingw32",
			Resource: &resources.DownloadResource{
				ArchiveFileName: fmt.Sprintf("mdns-discovery_%s_Windows_64bit.zip", mdnsDiscoveryVersion),
				URL:             fmt.Sprintf("https://downloads.arduino.cc/discovery/mdns-discovery/mdns-discovery_%s_Windows_64bit.zip", mdnsDiscoveryVersion),
				Size:            2603561,
				Checksum:        "SHA-256:26ef37b3f331cd6cfbfd234431fe95a0c4aec8d8889286084e0c9012cb72ff1e",
				CachePath:       "tools",
			},
		},
		{
			OS: "x86_64-apple-darwin",
			Resource: &resources.DownloadResource{
				ArchiveFileName: fmt.Sprintf("mdns-discovery_%s_macOS_64bit.tar.bz2", mdnsDiscoveryVersion),
				URL:             fmt.Sprintf("https://downloads.arduino.cc/discovery/mdns-discovery/mdns-discovery_%s_macOS_64bit.tar.gz", mdnsDiscoveryVersion),
				Size:            2458693,
				Checksum:        "SHA-256:51c7c1c7a7a81e1224cfd55f84497cfae278c28124d10619a0f7adae29c0f45b",
				CachePath:       "tools",
			},
		},
		{
			OS: "arm-linux-gnueabihf",
			Resource: &resources.DownloadResource{
				ArchiveFileName: fmt.Sprintf("mdns-discovery_%s_Linux_ARMv6.tar.bz2", mdnsDiscoveryVersion),
				URL:             fmt.Sprintf("https://downloads.arduino.cc/discovery/mdns-discovery/mdns-discovery_%s_Linux_ARMv6.tar.gz", mdnsDiscoveryVersion),
				Size:            2321304,
				Checksum:        "SHA-256:cc1096936abddb21af23fa10c435e8e9e37ec9df2c3d2c41d265d466b03de0af",
				CachePath:       "tools",
			},
		},
		{
			OS: "arm64-linux-gnueabihf",
			Resource: &resources.DownloadResource{
				ArchiveFileName: fmt.Sprintf("mdns-discovery_%s_Linux_ARM64.tar.bz2", mdnsDiscoveryVersion),
				URL:             fmt.Sprintf("https://downloads.arduino.cc/discovery/mdns-discovery/mdns-discovery_%s_Linux_ARM64.tar.gz", mdnsDiscoveryVersion),
				Size:            2328169,
				Checksum:        "SHA-256:820266009f7decf421c005d702e1b4da43ba2af03d2ddd73f5a3a91737774571",
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
