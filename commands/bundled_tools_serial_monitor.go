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
	serialMonitorVersion = semver.ParseRelaxed("0.9.1")
	serialMonitorFlavors = []*cores.Flavor{
		{
			OS: "i686-pc-linux-gnu",
			Resource: &resources.DownloadResource{
				ArchiveFileName: fmt.Sprintf("serial-monitor_v%s_Linux_32bit.tar.bz2", serialMonitorVersion),
				URL:             fmt.Sprintf("https://downloads.arduino.cc/monitor/serial-monitor/serial-monitor_v%s_Linux_32bit.tar.gz", serialMonitorVersion),
				Size:            1899387,
				Checksum:        "SHA-256:3939282c9c74dd259a0ebd66d959133efafc8b50fd800860d8c1f634615b665c",
				CachePath:       "tools",
			},
		},
		{
			OS: "x86_64-pc-linux-gnu",
			Resource: &resources.DownloadResource{
				ArchiveFileName: fmt.Sprintf("serial-monitor_v%s_Linux_64bit.tar.bz2", serialMonitorVersion),
				URL:             fmt.Sprintf("https://downloads.arduino.cc/monitor/serial-monitor/serial-monitor_v%s_Linux_64bit.tar.gz", serialMonitorVersion),
				Size:            1954589,
				Checksum:        "SHA-256:f121374fc33a66350381591816b2f2a0b0a108d70cf0ca01c59cc05186e6a5ce",
				CachePath:       "tools",
			},
		},
		{
			OS: "i686-mingw32",
			Resource: &resources.DownloadResource{
				ArchiveFileName: fmt.Sprintf("serial-monitor_v%s_Windows_32bit.zip", serialMonitorVersion),
				URL:             fmt.Sprintf("https://downloads.arduino.cc/monitor/serial-monitor/serial-monitor_v%s_Windows_32bit.zip", serialMonitorVersion),
				Size:            1956735,
				Checksum:        "SHA-256:15157e93618365cd959df57a9a25ccaa5a79d46a34f589e8711f571fe2e318e7",
				CachePath:       "tools",
			},
		},
		{
			OS: "x86_64-mingw32",
			Resource: &resources.DownloadResource{
				ArchiveFileName: fmt.Sprintf("serial-monitor_v%s_Windows_64bit.zip", serialMonitorVersion),
				URL:             fmt.Sprintf("https://downloads.arduino.cc/monitor/serial-monitor/serial-monitor_v%s_Windows_64bit.zip", serialMonitorVersion),
				Size:            1990791,
				Checksum:        "SHA-256:e45561908526e855a7b9284ee438d2503cb21f9a5421fd840c1f10cd46b10b25",
				CachePath:       "tools",
			},
		},
		{
			OS: "x86_64-apple-darwin",
			Resource: &resources.DownloadResource{
				ArchiveFileName: fmt.Sprintf("serial-monitor_v%s_macOS_64bit.tar.bz2", serialMonitorVersion),
				URL:             fmt.Sprintf("https://downloads.arduino.cc/monitor/serial-monitor/serial-monitor_v%s_macOS_64bit.tar.gz", serialMonitorVersion),
				Size:            1871195,
				Checksum:        "SHA-256:ebb4750e079ec893d89e9e256cd80b0e810a6cc17cd66189978f46246f52e14a",
				CachePath:       "tools",
			},
		},
		{
			OS: "arm-linux-gnueabihf",
			Resource: &resources.DownloadResource{
				ArchiveFileName: fmt.Sprintf("serial-monitor_v%s_Linux_ARMv6.tar.bz2", serialMonitorVersion),
				URL:             fmt.Sprintf("https://downloads.arduino.cc/monitor/serial-monitor/serial-monitor_v%s_Linux_ARMv6.tar.gz", serialMonitorVersion),
				Size:            1829212,
				Checksum:        "SHA-256:bd2cf410f7fbcb43dbe6ea9bdf265585de96bf7247cb425d050537ee59d16355",
				CachePath:       "tools",
			},
		},
		{
			OS: "arm64-linux-gnueabihf",
			Resource: &resources.DownloadResource{
				ArchiveFileName: fmt.Sprintf("serial-monitor_v%s_Linux_ARM64.tar.bz2", serialMonitorVersion),
				URL:             fmt.Sprintf("https://downloads.arduino.cc/monitor/serial-monitor/serial-monitor_v%s_Linux_ARM64.tar.gz", serialMonitorVersion),
				Size:            1837454,
				Checksum:        "SHA-256:be774d68c72fe7d79f9f6ec53f23e63e71793a15a60ed1210d51ec78c6fc0dc1",
				CachePath:       "tools",
			},
		},
	}
)

func getBuiltinSerialMonitorTool(pm *packagemanager.PackageManager) *cores.ToolRelease {
	builtinPackage := pm.Packages.GetOrCreatePackage("builtin")
	serialMonitorTool := builtinPackage.GetOrCreateTool("serial-monitor")
	serialMonitorToolRel := serialMonitorTool.GetOrCreateRelease(serialMonitorVersion)
	serialMonitorToolRel.Flavors = serialMonitorFlavors
	return serialMonitorToolRel
}
