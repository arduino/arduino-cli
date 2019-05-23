//
// This file is part of arduino-cli.
//
// Copyright 2018 ARDUINO SA (http://www.arduino.cc/)
//
// This software is released under the GNU General Public License version 3,
// which covers the main part of arduino-cli.
// The terms of this license can be found at:
// https://www.gnu.org/licenses/gpl-3.0.en.html
//
// You can be released from the requirements of the above licenses by purchasing
// a commercial license. Buying such a license is mandatory if you want to modify or
// otherwise use the software for commercial activities involving the Arduino
// software without disclosing the source code of your own applications. To purchase
// a commercial license, send an email to license@arduino.cc.
//

package commands

import (
	"fmt"

	"github.com/arduino/arduino-cli/arduino/cores"
	"github.com/arduino/arduino-cli/arduino/cores/packagemanager"
	"github.com/arduino/arduino-cli/arduino/discovery"
	"github.com/arduino/arduino-cli/arduino/resources"
	semver "go.bug.st/relaxed-semver"
)

var serialDiscoveryVersion = semver.ParseRelaxed("0.5.0")

func loadBuiltinSerialDiscoveryMetadata(pm *packagemanager.PackageManager) {
	builtinPackage := pm.GetPackages().GetOrCreatePackage("builtin")
	ctagsTool := builtinPackage.GetOrCreateTool("serial-discovery")
	ctagsRel := ctagsTool.GetOrCreateRelease(serialDiscoveryVersion)
	ctagsRel.Flavors = []*cores.Flavor{
		// {
		// 	OS: "i686-pc-linux-gnu",
		// 	Resource: &resources.DownloadResource{
		// 		ArchiveFileName: "serial-discovery-1.0.0-i686-pc-linux-gnu.tar.bz2",
		// 		URL:             "https://downloads.arduino.cc/tools/serial-discovery-1.0.0-i686-pc-linux-gnu.tar.bz2",
		// 		Size:            ,
		// 		Checksum:        "SHA-256:",
		// 		CachePath:       "tools",
		// 	},
		// },
		{
			OS: "x86_64-pc-linux-gnu",
			Resource: &resources.DownloadResource{
				ArchiveFileName: "serial-discovery-0.5.0-x86_64-pc-linux-gnu.tar.bz2",
				URL:             "https://downloads.arduino.cc/tools/serial-discovery-0.5.0-x86_64-pc-linux-gnu.tar.bz2",
				Size:            1507380,
				Checksum:        "SHA-256:473cdd9e9f189cfd507b1f6c312d767513da11ec87cdbff1610153d6285e15ce",
				CachePath:       "tools",
			},
		},
		// {
		// 	OS: "i686-mingw32",
		// 	Resource: &resources.DownloadResource{
		// 		ArchiveFileName: "serial-discovery-1.0.0-i686-mingw32.zip",
		// 		URL:             "https://downloads.arduino.cc/tools/serial-discovery-1.0.0-i686-mingw32.zip",
		// 		Size:            ,
		// 		Checksum:        "SHA-256:",
		// 		CachePath:       "tools",
		// 	},
		// },
		// {
		// 	OS: "x86_64-apple-darwin",
		// 	Resource: &resources.DownloadResource{
		// 		ArchiveFileName: "serial-discovery-1.0.0-x86_64-apple-darwin.zip",
		// 		URL:             "https://downloads.arduino.cc/tools/serial-discovery-1.0.0-x86_64-apple-darwin.zip",
		// 		Size:            ,
		// 		Checksum:        "SHA-256:",
		// 		CachePath:       "tools",
		// 	},
		// },
		// {
		// 	OS: "arm-linux-gnueabihf",
		// 	Resource: &resources.DownloadResource{
		// 		ArchiveFileName: "serial-discovery-1.0.0-armv6-linux-gnueabihf.tar.bz2",
		// 		URL:             "https://downloads.arduino.cc/tools/serial-discovery-1.0.0-armv6-linux-gnueabihf.tar.bz2",
		// 		Size:            ,
		// 		Checksum:        "SHA-256:",
		// 		CachePath:       "tools",
		// 	},
		// },
	}
}

func getBuiltinSerialDiscoveryTool(pm *packagemanager.PackageManager) (*cores.ToolRelease, error) {
	return pm.Package("builtin").Tool("serial-discovery").Release(serialDiscoveryVersion).Get()
}

func newBuiltinSerialDiscovery(pm *packagemanager.PackageManager) (*discovery.Discovery, error) {
	t, err := getBuiltinSerialDiscoveryTool(pm)
	if err != nil {
		return nil, err
	}
	if !t.IsInstalled() {
		return nil, fmt.Errorf("missing serial-discovery tool")
	}
	return discovery.NewFromCommandLine(t.InstallDir.Join("serial-discovery").String())
}
