package packagemanager_test

import (
	"encoding/json"
	"testing"

	"github.com/arduino/arduino-cli/arduino/cores/packageindex"
	"github.com/arduino/arduino-cli/arduino/cores/packagemanager"
	"github.com/arduino/arduino-cli/cli/output"
	"github.com/arduino/arduino-cli/commands"
	"github.com/arduino/go-paths-helper"
	"github.com/stretchr/testify/require"
	semver "go.bug.st/relaxed-semver"
)

func TestInstallPlatform(t *testing.T) {
	dataDir := paths.New("testdata", "data_dir_1")
	packageDir := paths.TempDir().Join("test", "packages")
	downloadDir := paths.TempDir().Join("test", "staging")
	tmpDir := paths.TempDir().Join("test", "tmp")
	packageDir.MkdirAll()
	downloadDir.MkdirAll()
	tmpDir.MkdirAll()
	defer paths.TempDir().Join("test").RemoveAll()

	pm := packagemanager.NewPackageManager(dataDir, packageDir, downloadDir, tmpDir)
	pm.LoadPackageIndexFromFile(dataDir.Join("package_index.json"))

	platformRelease, tools, err := pm.FindPlatformReleaseDependencies(&packagemanager.PlatformReference{
		Package:              "arduino",
		PlatformArchitecture: "avr",
		PlatformVersion:      semver.MustParse("1.6.23"),
	})
	require.NotNil(t, platformRelease)
	require.NotNil(t, tools)
	require.Nil(t, err)

	downloaderConfig, err := commands.GetDownloaderConfig()
	require.NotNil(t, downloaderConfig)
	require.Nil(t, err)
	downloader, err := pm.DownloadPlatformRelease(platformRelease, downloaderConfig)
	require.NotNil(t, downloader)
	require.Nil(t, err)
	err = commands.Download(downloader, platformRelease.String(), output.NewNullDownloadProgressCB())
	require.Nil(t, err)

	err = pm.InstallPlatform(platformRelease)
	require.Nil(t, err)

	destDir := packageDir.Join("arduino", "hardware", "avr", "1.6.23")
	require.True(t, destDir.IsDir())

	installedJSON := destDir.Join("installed.json")
	require.True(t, installedJSON.Exist())

	bt, err := installedJSON.ReadFile()
	require.Nil(t, err)

	index := &packageindex.Index{}
	err = json.Unmarshal(bt, index)
	require.Nil(t, err)

	require.Equal(t, 1, len(index.Packages))
	indexPackage := index.Packages[0]
	require.Equal(t, "arduino", indexPackage.Name)
	require.Equal(t, "Arduino", indexPackage.Maintainer)
	require.Equal(t, "http://www.arduino.cc/", indexPackage.WebsiteURL)
	require.Equal(t, "packages@arduino.cc", indexPackage.Email)
	require.Equal(t, 1, len(indexPackage.Platforms))
	require.Equal(t, "http://www.arduino.cc/en/Reference/HomePage", indexPackage.Help.Online)

	indexPackageTools := indexPackage.Tools
	require.Equal(t, 44, len(indexPackageTools))
	// Just check one to verifies it's actually correct
	require.Contains(t, indexPackageTools, &packageindex.IndexToolRelease{
		Name:    "arm-none-eabi-gcc",
		Version: semver.ParseRelaxed("4.8.3-2014q1"),
		Systems: []packageindex.IndexToolReleaseFlavour{
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
			{
				OS:              "x86_64-apple-darwin",
				URL:             "http://downloads.arduino.cc/gcc-arm-none-eabi-4.8.3-2014q1-mac.tar.gz",
				ArchiveFileName: "gcc-arm-none-eabi-4.8.3-2014q1-mac.tar.gz",
				Size:            "52518522",
				Checksum:        "SHA-256:3598acf21600f17a8e4a4e8e193dc422b894dc09384759b270b2ece5facb59c2",
			},
			{
				OS:              "x86_64-pc-linux-gnu",
				URL:             "http://downloads.arduino.cc/gcc-arm-none-eabi-4.8.3-2014q1-linux64.tar.gz",
				ArchiveFileName: "gcc-arm-none-eabi-4.8.3-2014q1-linux64.tar.gz",
				Size:            "51395093",
				Checksum:        "SHA-256:d23f6626148396d6ec42a5b4d928955a703e0757829195fa71a939e5b86eecf6",
			},
			{
				OS:              "i686-pc-linux-gnu",
				URL:             "http://downloads.arduino.cc/gcc-arm-none-eabi-4.8.3-2014q1-linux32.tar.gz",
				ArchiveFileName: "gcc-arm-none-eabi-4.8.3-2014q1-linux32.tar.gz",
				Size:            "51029223",
				Checksum:        "SHA-256:ba1994235f69c526c564f65343f22ddbc9822b2ea8c5ee07dd79d89f6ace2498",
			},
		}})

	indexPlatformRelease := indexPackage.Platforms[0]
	require.Equal(t, "Arduino AVR Boards", indexPlatformRelease.Name)
	require.Equal(t, "avr", indexPlatformRelease.Architecture)
	require.Equal(t, "1.6.23", indexPlatformRelease.Version.String())
	require.Equal(t, "Arduino", indexPlatformRelease.Category)
	require.Equal(t, "http://www.arduino.cc/en/Reference/HomePage", indexPlatformRelease.Help.Online)
	require.Equal(t, "http://downloads.arduino.cc/cores/avr-1.6.23.tar.bz2", indexPlatformRelease.URL)
	require.Equal(t, "avr-1.6.23.tar.bz2", indexPlatformRelease.ArchiveFileName)
	require.Equal(t, "SHA-256:18618d7f256f26cd77c35f4c888d5d1b2334f07925094fdc99ac3188722284aa", indexPlatformRelease.Checksum)
	require.Equal(t, json.Number("5001988"), indexPlatformRelease.Size)
	require.Contains(t, indexPlatformRelease.Boards, packageindex.IndexBoard{Name: "Arduino Yún"})
	require.Contains(t, indexPlatformRelease.Boards, packageindex.IndexBoard{Name: "Arduino/Genuino Uno"})
	require.Contains(t, indexPlatformRelease.Boards, packageindex.IndexBoard{Name: "Arduino Uno WiFi"})
	require.Contains(t, indexPlatformRelease.Boards, packageindex.IndexBoard{Name: "Arduino Diecimila"})
	require.Contains(t, indexPlatformRelease.Boards, packageindex.IndexBoard{Name: "Arduino Nano"})
	require.Contains(t, indexPlatformRelease.Boards, packageindex.IndexBoard{Name: "Arduino/Genuino Mega"})
	require.Contains(t, indexPlatformRelease.Boards, packageindex.IndexBoard{Name: "Arduino MegaADK"})
	require.Contains(t, indexPlatformRelease.Boards, packageindex.IndexBoard{Name: "Arduino Leonardo"})
	require.Contains(t, indexPlatformRelease.Boards, packageindex.IndexBoard{Name: "Arduino Leonardo Ethernet"})
	require.Contains(t, indexPlatformRelease.Boards, packageindex.IndexBoard{Name: "Arduino/Genuino Micro"})
	require.Contains(t, indexPlatformRelease.Boards, packageindex.IndexBoard{Name: "Arduino Esplora"})
	require.Contains(t, indexPlatformRelease.Boards, packageindex.IndexBoard{Name: "Arduino Mini"})
	require.Contains(t, indexPlatformRelease.Boards, packageindex.IndexBoard{Name: "Arduino Ethernet"})
	require.Contains(t, indexPlatformRelease.Boards, packageindex.IndexBoard{Name: "Arduino Fio"})
	require.Contains(t, indexPlatformRelease.Boards, packageindex.IndexBoard{Name: "Arduino BT"})
	require.Contains(t, indexPlatformRelease.Boards, packageindex.IndexBoard{Name: "Arduino LilyPadUSB"})
	require.Contains(t, indexPlatformRelease.Boards, packageindex.IndexBoard{Name: "Arduino Lilypad"})
	require.Contains(t, indexPlatformRelease.Boards, packageindex.IndexBoard{Name: "Arduino Pro"})
	require.Contains(t, indexPlatformRelease.Boards, packageindex.IndexBoard{Name: "Arduino ATMegaNG"})
	require.Contains(t, indexPlatformRelease.Boards, packageindex.IndexBoard{Name: "Arduino Robot Control"})
	require.Contains(t, indexPlatformRelease.Boards, packageindex.IndexBoard{Name: "Arduino Robot Motor"})
	require.Contains(t, indexPlatformRelease.Boards, packageindex.IndexBoard{Name: "Arduino Gemma"})
	require.Contains(t, indexPlatformRelease.Boards, packageindex.IndexBoard{Name: "Adafruit Circuit Playground"})
	require.Contains(t, indexPlatformRelease.Boards, packageindex.IndexBoard{Name: "Arduino Yún Mini"})
	require.Contains(t, indexPlatformRelease.Boards, packageindex.IndexBoard{Name: "Arduino Industrial 101"})
	require.Contains(t, indexPlatformRelease.Boards, packageindex.IndexBoard{Name: "Linino One"})
	require.Contains(t, indexPlatformRelease.ToolDependencies,
		packageindex.IndexToolDependency{
			Packager: "arduino",
			Name:     "avr-gcc",
			Version:  semver.ParseRelaxed("5.4.0-atmel3.6.1-arduino2"),
		})
	require.Contains(t, indexPlatformRelease.ToolDependencies,
		packageindex.IndexToolDependency{
			Packager: "arduino",
			Name:     "avrdude",
			Version:  semver.ParseRelaxed("6.3.0-arduino14"),
		})
	require.Contains(t, indexPlatformRelease.ToolDependencies,
		packageindex.IndexToolDependency{
			Packager: "arduino",
			Name:     "arduinoOTA",
			Version:  semver.ParseRelaxed("1.2.1"),
		})
}
