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

package packageindex

import (
	"testing"

	"github.com/arduino/arduino-cli/internal/arduino/cores"
	"github.com/arduino/arduino-cli/internal/arduino/resources"
	"github.com/arduino/go-paths-helper"
	"github.com/stretchr/testify/require"
	semver "go.bug.st/relaxed-semver"
)

func TestIndexParsing(t *testing.T) {
	semver.WarnInvalidVersionWhenParsingRelaxed = true

	list, err := paths.New("testdata").ReadDir()
	require.NoError(t, err)
	for _, indexFile := range list {
		if indexFile.Ext() != ".json" {
			continue
		}
		_, err := LoadIndex(indexFile)
		require.NoError(t, err)
	}
}

func TestIndexFromPlatformRelease(t *testing.T) {
	arduinoTools := map[string]*cores.Tool{
		"serial-discovery": {
			Name: "serial-discovery",
			Releases: map[semver.NormalizedString]*cores.ToolRelease{
				"1.0.0": {
					Version: semver.ParseRelaxed("1.0.0"),
					Flavors: []*cores.Flavor{
						{
							OS: "arm-linux-gnueabihf",
							Resource: &resources.DownloadResource{
								URL:             "some-serial-discovery-1.0.0-url",
								ArchiveFileName: "serial-discovery-1.0.0.tar.bz2",
								Checksum:        "SHA-256:some-serial-discovery-1.0.0-sha",
								Size:            201341,
							},
						},
						{
							OS: "i686-mingw32",
							Resource: &resources.DownloadResource{
								URL:             "some-serial-discovery-1.0.0-other-url",
								ArchiveFileName: "serial-discovery-1.0.0.tar.gz",
								Checksum:        "SHA-256:some-serial-discovery-1.0.0-other-sha",
								Size:            222918,
							},
						},
					},
				},
				"0.1.0": {
					Version: semver.ParseRelaxed("0.1.0"),
					Flavors: []*cores.Flavor{
						{
							OS: "arm-linux-gnueabihf",
							Resource: &resources.DownloadResource{
								URL:             "some-serial-discovery-0.1.0-url",
								ArchiveFileName: "serial-discovery-0.1.0.tar.bz2",
								Checksum:        "SHA-256:some-serial-discovery-0.1.0-sha",
								Size:            201341,
							},
						},
						{
							OS: "i686-mingw32",
							Resource: &resources.DownloadResource{
								URL:             "some-serial-discovery-0.1.0-other-url",
								ArchiveFileName: "serial-discovery-0.1.0.tar.gz",
								Checksum:        "SHA-256:some-serial-discovery-0.1.0-other-sha",
								Size:            222918,
							},
						},
					},
				},
			},
		},
		"ble-discovery": {
			Name: "ble-discovery",
			Releases: map[semver.NormalizedString]*cores.ToolRelease{
				"1.0.0": {
					Version: semver.ParseRelaxed("1.0.0"),
					Flavors: []*cores.Flavor{
						{
							OS: "arm-linux-gnueabihf",
							Resource: &resources.DownloadResource{
								URL:             "some-ble-discovery-1.0.0-url",
								ArchiveFileName: "ble-discovery-1.0.0.tar.bz2",
								Checksum:        "SHA-256:some-ble-discovery-1.0.0-sha",
								Size:            201341,
							},
						},
						{
							OS: "i686-mingw32",
							Resource: &resources.DownloadResource{
								URL:             "some-ble-discovery-1.0.0-other-url",
								ArchiveFileName: "ble-discovery-1.0.0.tar.gz",
								Checksum:        "SHA-256:some-ble-discovery-1.0.0-other-sha",
								Size:            222918,
							},
						},
					},
				},
				"0.1.0": {
					Version: semver.ParseRelaxed("0.1.0"),
					Flavors: []*cores.Flavor{
						{
							OS: "arm-linux-gnueabihf",
							Resource: &resources.DownloadResource{
								URL:             "some-ble-discovery-0.1.0-url",
								ArchiveFileName: "ble-discovery-0.1.0.tar.bz2",
								Checksum:        "SHA-256:some-ble-discovery-0.1.0-sha",
								Size:            201341,
							},
						},

						{
							OS: "i686-mingw32",
							Resource: &resources.DownloadResource{
								URL:             "some-ble-discovery-0.1.0-other-url",
								ArchiveFileName: "ble-discovery-0.1.0.tar.gz",
								Checksum:        "SHA-256:some-ble-discovery-0.1.0-other-sha",
								Size:            222918,
							},
						},
					},
				},
			},
		},
		"bossac": {
			Name: "bossac",
			Releases: map[semver.NormalizedString]*cores.ToolRelease{
				"1.6.1-arduino": {
					Version: semver.ParseRelaxed("1.6.1-arduino"),
					Flavors: []*cores.Flavor{
						{
							OS: "arm-linux-gnueabihf",
							Resource: &resources.DownloadResource{
								URL:             "http://downloads.arduino.cc/bossac-1.6.1-arduino-arm-linux-gnueabihf.tar.bz2",
								ArchiveFileName: "bossac-1.6.1-arduino-arm-linux-gnueabihf.tar.bz2",
								Checksum:        "SHA-256:8c4e63db982178919c824e7a35580dffc95c3426afa7285de3eb583982d4d391",
								Size:            201341,
							},
						},
						{
							OS: "i686-mingw32",
							Resource: &resources.DownloadResource{
								URL:             "http://downloads.arduino.cc/bossac-1.6.1-arduino-mingw32.tar.gz",
								ArchiveFileName: "bossac-1.6.1-arduino-mingw32.tar.gz",
								Checksum:        "SHA-256:d59f43e2e83a337d04c4ae88b195a4ee175b8d87fff4c43144d23412a4a9513b",
								Size:            222918,
							},
						},
					},
				},
				"1.7.0": {
					Version: semver.ParseRelaxed("1.7.0"),
					Flavors: []*cores.Flavor{
						{
							OS: "i686-mingw32",
							Resource: &resources.DownloadResource{
								URL:             "http://downloads.arduino.cc/bossac-1.7.0-arduino-mingw32.tar.bz2",
								ArchiveFileName: "bossac-1.7.0-arduino-mingw32.tar.bz2",
								Checksum:        "SHA-256:9ef7d11b4fabca0adc17102a0290957d5cc26ce46b422c3a5344722c80acc7b2",
								Size:            243066,
							},
						},
						{
							OS: "x86_64-apple-darwin",
							Resource: &resources.DownloadResource{
								URL:             "http://downloads.arduino.cc/bossac-1.7.0-arduino-x86_64-apple-darwin.tar.bz2",
								ArchiveFileName: "bossac-1.7.0-arduino-x86_64-apple-darwin.tar.bz2",
								Checksum:        "SHA-256:feac36ab38876c163dcf51bdbcfbed01554eede3d41c59a0e152e170fe5164d2",
								Size:            63822,
							},
						},
					},
				},
			},
		},
		"arm-none-eabi-gcc": {
			Name: "arm-none-eabi-gcc",
			Releases: map[semver.NormalizedString]*cores.ToolRelease{
				"4.8.3-2014q1": {
					Version: semver.ParseRelaxed("4.8.3-2014q1"),
					Flavors: []*cores.Flavor{
						{
							OS: "arm-linux-gnueabihf",
							Resource: &resources.DownloadResource{
								URL:             "http://downloads.arduino.cc/gcc-arm-none-eabi-4.8.3-2014q1-arm.tar.bz2",
								ArchiveFileName: "gcc-arm-none-eabi-4.8.3-2014q1-arm.tar.bz2",
								Checksum:        "SHA-256:ebe96b34c4f434667cab0187b881ed585e7c7eb990fe6b69be3c81ec7e11e845",
								Size:            44423906,
							},
						},
						{
							OS: "i686-mingw32",
							Resource: &resources.DownloadResource{
								URL:             "http://downloads.arduino.cc/gcc-arm-none-eabi-4.8.3-2014q1-windows.tar.gz",
								ArchiveFileName: "gcc-arm-none-eabi-4.8.3-2014q1-windows.tar.gz",
								Checksum:        "SHA-256:fd8c111c861144f932728e00abd3f7d1107e186eb9cd6083a54c7236ea78b7c2",
								Size:            84537449,
							},
						},
					},
				},
				"7-2017q4": {
					Version: semver.ParseRelaxed("7-2017q4"),
					Flavors: []*cores.Flavor{
						{
							OS: "arm-linux-gnueabihf",
							Resource: &resources.DownloadResource{
								URL:             "http://downloads.arduino.cc/tools/gcc-arm-none-eabi-4.8.3-2014q1-arm.tar.bz2",
								ArchiveFileName: "gcc-arm-none-eabi-4.8.3-2014q1-arm.tar.bz2",
								Checksum:        "SHA-256:ebe96b34c4f434667cab0187b881ed585e7c7eb990fe6b69be3c81ec7e11e845",
								Size:            44423906,
							},
						},
						{
							OS: "aarch64-linux-gnu",
							Resource: &resources.DownloadResource{
								URL:             "http://downloads.arduino.cc/tools/gcc-arm-none-eabi-7-2018-q2-update-linuxarm64.tar.bz2",
								ArchiveFileName: "gcc-arm-none-eabi-7-2018-q2-update-linuxarm64.tar.bz2",
								Checksum:        "SHA-256:6fb5752fb4d11012bd0a1ceb93a19d0641ff7cf29d289b3e6b86b99768e66f76",
								Size:            99558726,
							},
						},
					},
				},
			},
		},
	}

	avrPlatformRelease := &cores.PlatformRelease{
		Resource: &resources.DownloadResource{
			URL:             "http://downloads.arduino.cc/cores/avr-1.6.23.tar.bz2",
			ArchiveFileName: "avr-1.6.23.tar.bz2",
			Checksum:        "SHA-256:18618d7f256f26cd77c35f4c888d5d1b2334f07925094fdc99ac3188722284aa",
			Size:            5001988,
		},
		Version: semver.MustParse("1.8.3"),
		Help:    cores.PlatformReleaseHelp{Online: "http://www.arduino.cc/en/Reference/HomePage"},
		BoardsManifest: []*cores.BoardManifest{
			{Name: "Arduino Yún"},
			{Name: "Arduino/Genuino Uno"},
			{Name: "Arduino Uno WiFi"},
		},
		ToolDependencies: cores.ToolDependencies{
			{ToolPackager: "arduino", ToolName: "avr-gcc", ToolVersion: semver.ParseRelaxed("5.4.0-atmel3.6.1-arduino2")},
			{ToolPackager: "arduino", ToolName: "avrdude", ToolVersion: semver.ParseRelaxed("6.3.0-arduino14")},
			{ToolPackager: "arduino", ToolName: "arduinoOTA", ToolVersion: semver.ParseRelaxed("1.2.1")},
		},
		DiscoveryDependencies: cores.DiscoveryDependencies{
			{Packager: "arduino", Name: "ble-discovery"},
			{Packager: "arduino", Name: "serial-discovery"},
		},
		MonitorDependencies: cores.MonitorDependencies{
			{Packager: "arduino", Name: "ble-monitor"},
			{Packager: "arduino", Name: "serial-monitor"},
		},
		Name:     "Arduino AVR Boards",
		Category: "Arduino",
	}
	avrPlatform := &cores.Platform{
		Architecture: "avr",
		Releases: map[semver.NormalizedString]*cores.PlatformRelease{
			avrPlatformRelease.Version.NormalizedString(): avrPlatformRelease,
		},
	}
	avrPlatformRelease.Platform = avrPlatform

	arduinoPackage := &cores.Package{
		Name:       "arduino",
		Maintainer: "Arduino",
		WebsiteURL: "https://arduino.cc/",
		URL:        "",
		Email:      "packages@arduino.cc",
		Help:       cores.PackageHelp{Online: "http://www.arduino.cc/en/Reference/HomePage"},
		Tools:      arduinoTools,
		Platforms: map[string]*cores.Platform{
			"avr": avrPlatform,
		},
	}
	avrPlatform.Package = arduinoPackage

	expectedIndex := Index{
		IsTrusted: false,
		Packages: []*indexPackage{{
			Name:       "arduino",
			Maintainer: "Arduino",
			WebsiteURL: "https://arduino.cc/",
			URL:        "",
			Email:      "packages@arduino.cc",
			Help:       indexHelp{Online: "http://www.arduino.cc/en/Reference/HomePage"},
			Platforms: []*indexPlatformRelease{{
				Name:            "Arduino AVR Boards",
				Architecture:    "avr",
				Version:         semver.MustParse("1.8.3"),
				Category:        "Arduino",
				URL:             "http://downloads.arduino.cc/cores/avr-1.6.23.tar.bz2",
				ArchiveFileName: "avr-1.6.23.tar.bz2",
				Checksum:        "SHA-256:18618d7f256f26cd77c35f4c888d5d1b2334f07925094fdc99ac3188722284aa",
				Size:            "5001988",
				Boards: []indexBoard{
					{Name: "Arduino Yún"},
					{Name: "Arduino/Genuino Uno"},
					{Name: "Arduino Uno WiFi"},
				},
				Help: indexHelp{Online: "http://www.arduino.cc/en/Reference/HomePage"},
				ToolDependencies: []indexToolDependency{
					{Packager: "arduino", Name: "avr-gcc", Version: semver.ParseRelaxed("5.4.0-atmel3.6.1-arduino2")},
					{Packager: "arduino", Name: "avrdude", Version: semver.ParseRelaxed("6.3.0-arduino14")},
					{Packager: "arduino", Name: "arduinoOTA", Version: semver.ParseRelaxed("1.2.1")},
				},
				DiscoveryDependencies: []indexDiscoveryDependency{
					{Packager: "arduino", Name: "ble-discovery"},
					{Packager: "arduino", Name: "serial-discovery"},
				},
				MonitorDependencies: []indexMonitorDependency{
					{Packager: "arduino", Name: "ble-monitor"},
					{Packager: "arduino", Name: "serial-monitor"},
				},
			}},
			Tools: []*indexToolRelease{
				{
					Name:    "serial-discovery",
					Version: semver.ParseRelaxed("1.0.0"),
					Systems: []indexToolReleaseFlavour{
						{
							OS:              "arm-linux-gnueabihf",
							URL:             "some-serial-discovery-1.0.0-url",
							ArchiveFileName: "serial-discovery-1.0.0.tar.bz2",
							Checksum:        "SHA-256:some-serial-discovery-1.0.0-sha",
							Size:            "201341",
						},
						{
							OS:              "i686-mingw32",
							URL:             "some-serial-discovery-1.0.0-other-url",
							ArchiveFileName: "serial-discovery-1.0.0.tar.gz",
							Checksum:        "SHA-256:some-serial-discovery-1.0.0-other-sha",
							Size:            "222918",
						},
					},
				},
				{
					Name:    "serial-discovery",
					Version: semver.ParseRelaxed("0.1.0"),
					Systems: []indexToolReleaseFlavour{
						{
							OS:              "arm-linux-gnueabihf",
							URL:             "some-serial-discovery-0.1.0-url",
							ArchiveFileName: "serial-discovery-0.1.0.tar.bz2",
							Checksum:        "SHA-256:some-serial-discovery-0.1.0-sha",
							Size:            "201341",
						},
						{
							OS:              "i686-mingw32",
							URL:             "some-serial-discovery-0.1.0-other-url",
							ArchiveFileName: "serial-discovery-0.1.0.tar.gz",
							Checksum:        "SHA-256:some-serial-discovery-0.1.0-other-sha",
							Size:            "222918",
						},
					},
				},
				{
					Name:    "ble-discovery",
					Version: semver.ParseRelaxed("1.0.0"),
					Systems: []indexToolReleaseFlavour{
						{
							OS:              "arm-linux-gnueabihf",
							URL:             "some-ble-discovery-1.0.0-url",
							ArchiveFileName: "ble-discovery-1.0.0.tar.bz2",
							Checksum:        "SHA-256:some-ble-discovery-1.0.0-sha",
							Size:            "201341",
						},
						{
							OS:              "i686-mingw32",
							URL:             "some-ble-discovery-1.0.0-other-url",
							ArchiveFileName: "ble-discovery-1.0.0.tar.gz",
							Checksum:        "SHA-256:some-ble-discovery-1.0.0-other-sha",
							Size:            "222918",
						},
					},
				},
				{
					Name:    "ble-discovery",
					Version: semver.ParseRelaxed("0.1.0"),
					Systems: []indexToolReleaseFlavour{
						{
							OS:              "arm-linux-gnueabihf",
							URL:             "some-ble-discovery-0.1.0-url",
							ArchiveFileName: "ble-discovery-0.1.0.tar.bz2",
							Checksum:        "SHA-256:some-ble-discovery-0.1.0-sha",
							Size:            "201341",
						},
						{
							OS:              "i686-mingw32",
							URL:             "some-ble-discovery-0.1.0-other-url",
							ArchiveFileName: "ble-discovery-0.1.0.tar.gz",
							Checksum:        "SHA-256:some-ble-discovery-0.1.0-other-sha",
							Size:            "222918",
						},
					},
				},
				{
					Name:    "bossac",
					Version: semver.ParseRelaxed("1.6.1-arduino"),
					Systems: []indexToolReleaseFlavour{
						{
							OS:              "arm-linux-gnueabihf",
							URL:             "http://downloads.arduino.cc/bossac-1.6.1-arduino-arm-linux-gnueabihf.tar.bz2",
							ArchiveFileName: "bossac-1.6.1-arduino-arm-linux-gnueabihf.tar.bz2",
							Size:            "201341",
							Checksum:        "SHA-256:8c4e63db982178919c824e7a35580dffc95c3426afa7285de3eb583982d4d391",
						},
						{
							OS:              "i686-mingw32",
							URL:             "http://downloads.arduino.cc/bossac-1.6.1-arduino-mingw32.tar.gz",
							ArchiveFileName: "bossac-1.6.1-arduino-mingw32.tar.gz",
							Size:            "222918",
							Checksum:        "SHA-256:d59f43e2e83a337d04c4ae88b195a4ee175b8d87fff4c43144d23412a4a9513b",
						},
					},
				},
				{
					Name:    "bossac",
					Version: semver.ParseRelaxed("1.7.0"),
					Systems: []indexToolReleaseFlavour{
						{
							OS:              "i686-mingw32",
							URL:             "http://downloads.arduino.cc/bossac-1.7.0-arduino-mingw32.tar.bz2",
							ArchiveFileName: "bossac-1.7.0-arduino-mingw32.tar.bz2",
							Size:            "243066",
							Checksum:        "SHA-256:9ef7d11b4fabca0adc17102a0290957d5cc26ce46b422c3a5344722c80acc7b2",
						},
						{
							OS:              "x86_64-apple-darwin",
							URL:             "http://downloads.arduino.cc/bossac-1.7.0-arduino-x86_64-apple-darwin.tar.bz2",
							ArchiveFileName: "bossac-1.7.0-arduino-x86_64-apple-darwin.tar.bz2",
							Size:            "63822",
							Checksum:        "SHA-256:feac36ab38876c163dcf51bdbcfbed01554eede3d41c59a0e152e170fe5164d2",
						},
					},
				},
				{
					Name:    "arm-none-eabi-gcc",
					Version: semver.ParseRelaxed("4.8.3-2014q1"),
					Systems: []indexToolReleaseFlavour{
						{
							OS:              "arm-linux-gnueabihf",
							URL:             "http://downloads.arduino.cc/gcc-arm-none-eabi-4.8.3-2014q1-arm.tar.bz2",
							ArchiveFileName: "gcc-arm-none-eabi-4.8.3-2014q1-arm.tar.bz2",
							Size:            "44423906",
							Checksum:        "SHA-256:ebe96b34c4f434667cab0187b881ed585e7c7eb990fe6b69be3c81ec7e11e845",
						},
						{
							OS:              "i686-mingw32",
							URL:             "http://downloads.arduino.cc/gcc-arm-none-eabi-4.8.3-2014q1-windows.tar.gz",
							ArchiveFileName: "gcc-arm-none-eabi-4.8.3-2014q1-windows.tar.gz",
							Size:            "84537449",
							Checksum:        "SHA-256:fd8c111c861144f932728e00abd3f7d1107e186eb9cd6083a54c7236ea78b7c2",
						},
					},
				},
				{
					Name:    "arm-none-eabi-gcc",
					Version: semver.ParseRelaxed("7-2017q4"),
					Systems: []indexToolReleaseFlavour{
						{
							OS:              "arm-linux-gnueabihf",
							URL:             "http://downloads.arduino.cc/tools/gcc-arm-none-eabi-4.8.3-2014q1-arm.tar.bz2",
							ArchiveFileName: "gcc-arm-none-eabi-4.8.3-2014q1-arm.tar.bz2",
							Size:            "44423906",
							Checksum:        "SHA-256:ebe96b34c4f434667cab0187b881ed585e7c7eb990fe6b69be3c81ec7e11e845",
						},
						{
							OS:              "aarch64-linux-gnu",
							URL:             "http://downloads.arduino.cc/tools/gcc-arm-none-eabi-7-2018-q2-update-linuxarm64.tar.bz2",
							ArchiveFileName: "gcc-arm-none-eabi-7-2018-q2-update-linuxarm64.tar.bz2",
							Size:            "99558726",
							Checksum:        "SHA-256:6fb5752fb4d11012bd0a1ceb93a19d0641ff7cf29d289b3e6b86b99768e66f76",
						},
					},
				},
			},
		}},
	}

	in := IndexFromPlatformRelease(avrPlatformRelease)
	require.Equal(t, expectedIndex.IsTrusted, in.IsTrusted)
	require.Equal(t, len(expectedIndex.Packages), len(in.Packages))

	for i := range expectedIndex.Packages {
		expectedPackage := expectedIndex.Packages[i]
		indexPackage := in.Packages[i]
		require.Equal(t, expectedPackage.Name, indexPackage.Name)
		require.Equal(t, expectedPackage.Maintainer, indexPackage.Maintainer)
		require.Equal(t, expectedPackage.WebsiteURL, indexPackage.WebsiteURL)
		require.Equal(t, expectedPackage.Email, indexPackage.Email)
		require.Equal(t, expectedPackage.Help.Online, indexPackage.Help.Online)
		require.Equal(t, len(expectedPackage.Tools), len(indexPackage.Tools))
		require.ElementsMatch(t, expectedPackage.Tools, indexPackage.Tools)

		require.Equal(t, len(expectedPackage.Platforms), len(indexPackage.Platforms))
		for n := range expectedPackage.Platforms {
			expectedPlatform := expectedPackage.Platforms[n]
			indexPlatform := indexPackage.Platforms[n]
			require.Equal(t, expectedPlatform.Name, indexPlatform.Name)
			require.Equal(t, expectedPlatform.Architecture, indexPlatform.Architecture)
			require.Equal(t, expectedPlatform.Version.String(), indexPlatform.Version.String())
			require.Equal(t, expectedPlatform.Category, indexPlatform.Category)
			require.Equal(t, expectedPlatform.Help.Online, indexPlatform.Help.Online)
			require.Equal(t, expectedPlatform.URL, indexPlatform.URL)
			require.Equal(t, expectedPlatform.ArchiveFileName, indexPlatform.ArchiveFileName)
			require.Equal(t, expectedPlatform.Checksum, indexPlatform.Checksum)
			require.Equal(t, expectedPlatform.Size, indexPlatform.Size)
			require.ElementsMatch(t, expectedPlatform.Boards, indexPlatform.Boards)
			require.ElementsMatch(t, expectedPlatform.ToolDependencies, indexPlatform.ToolDependencies)
			require.ElementsMatch(t, expectedPlatform.DiscoveryDependencies, indexPlatform.DiscoveryDependencies)
			require.ElementsMatch(t, expectedPlatform.MonitorDependencies, indexPlatform.MonitorDependencies)
		}
	}
}
