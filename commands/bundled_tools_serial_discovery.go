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
		{
			OS: "i686-pc-linux-gnu",
			Resource: &resources.DownloadResource{
				ArchiveFileName: "serial-discovery-linux32-v1.0.0.tar.bz2",
				URL:             "https://downloads.arduino.cc/tools/serial-discovery-linux32-v1.0.0.tar.bz2",
				Size:            1469113,
				Checksum:        "SHA-256:35d96977844ad8d5ca9363e1ae5794450e5f7cf3d29ce7fdfe656b59e7fff725",
				CachePath:       "tools",
			},
		},
		{
			OS: "x86_64-pc-linux-gnu",
			Resource: &resources.DownloadResource{
				ArchiveFileName: "serial-discovery-linux64-v1.0.0.tar.bz2",
				URL:             "https://downloads.arduino.cc/tools/serial-discovery-linux64-v1.0.0.tar.bz2",
				Size:            1503971,
				Checksum:        "SHA-256:1a870d4d823ea6ebec403f63b10a1dbc9c623a6efea5cfa9141fa20045b731e2",
				CachePath:       "tools",
			},
		},
		{
			OS: "i686-mingw32",
			Resource: &resources.DownloadResource{
				ArchiveFileName: "serial-discovery-windows-v1.0.0.zip",
				URL:             "https://downloads.arduino.cc/tools/serial-discovery-windows-v1.0.0.zip",
				Size:            1512379,
				Checksum:        "SHA-256:b956128ab27a3a883c938d17cad640ba396876472f2ed25d8e661f12f5d0f584",
				CachePath:       "tools",
			},
		},
		{
			OS: "x86_64-apple-darwin",
			Resource: &resources.DownloadResource{
				ArchiveFileName: "serial-discovery-macosx-v1.0.0.tar.bz2",
				URL:             "https://downloads.arduino.cc/tools/serial-discovery-macosx-v1.0.0.tar.bz2",
				Size:            746132,
				Checksum:        "SHA-256:fcff1b972b70a73cd738facc6d99174d8323293b60c12149c8f6f3084fb2170e",
				CachePath:       "tools",
			},
		},
		{
			OS: "arm-linux-gnueabihf",
			Resource: &resources.DownloadResource{
				ArchiveFileName: "serial-discovery-linuxarm-v1.0.0.tar.bz2",
				URL:             "https://downloads.arduino.cc/tools/serial-discovery-linuxarm-v1.0.0.tar.bz2",
				Size:            1395174,
				Checksum:        "SHA-256:f196765caa62d38475208c27b3b516e61427d5d3a8ddc6e863acb4e4a3984701",
				CachePath:       "tools",
			},
		},
		{
			OS: "arm64-linux-gnueabihf",
			Resource: &resources.DownloadResource{
				ArchiveFileName: "serial-discovery-linuxarm64-v1.0.0.tar.bz2",
				URL:             "https://downloads.arduino.cc/tools/serial-discovery-linuxarm64-v1.0.0.tar.bz2",
				Size:            1402706,
				Checksum:        "SHA-256:c87010ed670254c06ac7abbc4daf7446e4e17f1945a75fc2602dd5930835dd25",
				CachePath:       "tools",
			},
		},
	}
}

func getBuiltinSerialDiscoveryTool(pm *packagemanager.PackageManager) (*cores.ToolRelease, error) {
	loadBuiltinSerialDiscoveryMetadata(pm)
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
