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

package packagemanager_test

import (
	"bytes"
	"fmt"
	"io"
	"net/url"
	"os"
	"runtime"
	"strings"
	"testing"

	"github.com/arduino/arduino-cli/arduino/cores"
	"github.com/arduino/arduino-cli/arduino/cores/packagemanager"
	"github.com/arduino/arduino-cli/configuration"
	"github.com/arduino/go-paths-helper"
	"github.com/arduino/go-properties-orderedmap"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/writer"
	"github.com/stretchr/testify/require"
	semver "go.bug.st/relaxed-semver"
)

var customHardware = paths.New("testdata", "custom_hardware")
var dataDir1 = paths.New("testdata", "data_dir_1")

// Intended to be used alongside dataDir1
var extraHardware = paths.New("testdata", "extra_hardware")

func TestFindBoardWithFQBN(t *testing.T) {
	pmb := packagemanager.NewBuilder(customHardware, customHardware, customHardware, customHardware, "test")
	pmb.LoadHardwareFromDirectory(customHardware)
	pm := pmb.Build()
	pme, release := pm.NewExplorer()
	defer release()
	board, err := pme.FindBoardWithFQBN("arduino:avr:uno")
	require.Nil(t, err)
	require.NotNil(t, board)
	require.Equal(t, board.Name(), "Arduino/Genuino Uno")

	board, err = pme.FindBoardWithFQBN("arduino:avr:mega")
	require.Nil(t, err)
	require.NotNil(t, board)
	require.Equal(t, board.Name(), "Arduino/Genuino Mega or Mega 2560")
}

func TestResolveFQBN(t *testing.T) {
	// Pass nil, since these paths are only used for installing
	pmb := packagemanager.NewBuilder(nil, nil, nil, nil, "test")
	// Hardware from main packages directory
	pmb.LoadHardwareFromDirectory(dataDir1.Join("packages"))
	// This contains the arduino:avr core
	pmb.LoadHardwareFromDirectory(customHardware)
	// This contains the referenced:avr core
	pmb.LoadHardwareFromDirectory(extraHardware)
	pm := pmb.Build()
	pme, release := pm.NewExplorer()
	defer release()

	fqbn, err := cores.ParseFQBN("arduino:avr:uno")
	require.Nil(t, err)
	require.NotNil(t, fqbn)
	pkg, platformRelease, board, props, buildPlatformRelease, err := pme.ResolveFQBN(fqbn)
	require.Nil(t, err)
	require.Equal(t, pkg, platformRelease.Platform.Package)
	require.NotNil(t, platformRelease)
	require.NotNil(t, platformRelease.Platform)
	require.Equal(t, platformRelease.Platform.String(), "arduino:avr")
	require.NotNil(t, board)
	require.Equal(t, board.Name(), "Arduino Uno")
	require.NotNil(t, props)
	require.Equal(t, platformRelease, buildPlatformRelease)

	fqbn, err = cores.ParseFQBN("arduino:avr:mega")
	require.Nil(t, err)
	require.NotNil(t, fqbn)
	pkg, platformRelease, board, props, buildPlatformRelease, err = pme.ResolveFQBN(fqbn)
	require.Nil(t, err)
	require.Equal(t, pkg, platformRelease.Platform.Package)
	require.NotNil(t, platformRelease)
	require.NotNil(t, platformRelease.Platform)
	require.Equal(t, platformRelease.Platform.String(), "arduino:avr")
	require.NotNil(t, board)
	require.Equal(t, board.Name(), "Arduino Mega or Mega 2560")
	require.NotNil(t, props)
	require.Equal(t, platformRelease, buildPlatformRelease)

	// Test a board referenced from the main AVR arduino platform
	fqbn, err = cores.ParseFQBN("referenced:avr:uno")
	require.Nil(t, err)
	require.NotNil(t, fqbn)
	pkg, platformRelease, board, props, buildPlatformRelease, err = pme.ResolveFQBN(fqbn)
	require.Nil(t, err)
	require.Equal(t, pkg, platformRelease.Platform.Package)
	require.NotNil(t, platformRelease)
	require.NotNil(t, platformRelease.Platform)
	require.Equal(t, platformRelease.Platform.String(), "referenced:avr")
	require.NotNil(t, board)
	require.Equal(t, board.Name(), "Referenced Uno")
	require.NotNil(t, props)
	require.NotNil(t, buildPlatformRelease)
	require.NotNil(t, buildPlatformRelease.Platform)
	require.Equal(t, buildPlatformRelease.Platform.String(), "arduino:avr")

	// Test a board referenced from the Adafruit SAMD core (this tests
	// deriving where the package and core name are different)
	fqbn, err = cores.ParseFQBN("referenced:samd:feather_m0")
	require.Nil(t, err)
	require.NotNil(t, fqbn)
	pkg, platformRelease, board, props, buildPlatformRelease, err = pme.ResolveFQBN(fqbn)
	require.Nil(t, err)
	require.Equal(t, pkg, platformRelease.Platform.Package)
	require.NotNil(t, platformRelease)
	require.NotNil(t, platformRelease.Platform)
	require.Equal(t, platformRelease.Platform.String(), "referenced:samd")
	require.NotNil(t, board)
	require.Equal(t, board.Name(), "Referenced Feather M0")
	require.NotNil(t, props)
	require.NotNil(t, buildPlatformRelease)
	require.NotNil(t, buildPlatformRelease.Platform)
	require.Equal(t, buildPlatformRelease.Platform.String(), "adafruit:samd")

	// Test a board referenced from a non-existent package
	fqbn, err = cores.ParseFQBN("referenced:avr:dummy_invalid_package")
	require.Nil(t, err)
	require.NotNil(t, fqbn)
	pkg, platformRelease, board, props, buildPlatformRelease, err = pme.ResolveFQBN(fqbn)
	require.NotNil(t, err)
	require.Equal(t, pkg, platformRelease.Platform.Package)
	require.NotNil(t, platformRelease)
	require.NotNil(t, platformRelease.Platform)
	require.Equal(t, platformRelease.Platform.String(), "referenced:avr")
	require.NotNil(t, board)
	require.Equal(t, board.Name(), "Referenced dummy with invalid package")
	require.NotNil(t, props)
	require.Nil(t, buildPlatformRelease)

	// Test a board referenced from a non-existent platform/architecture
	fqbn, err = cores.ParseFQBN("referenced:avr:dummy_invalid_platform")
	require.Nil(t, err)
	require.NotNil(t, fqbn)
	pkg, platformRelease, board, props, buildPlatformRelease, err = pme.ResolveFQBN(fqbn)
	require.NotNil(t, err)
	require.Equal(t, pkg, platformRelease.Platform.Package)
	require.NotNil(t, platformRelease)
	require.NotNil(t, platformRelease.Platform)
	require.Equal(t, platformRelease.Platform.String(), "referenced:avr")
	require.NotNil(t, board)
	require.Equal(t, board.Name(), "Referenced dummy with invalid platform")
	require.NotNil(t, props)
	require.Nil(t, buildPlatformRelease)

	// Test a board referenced from a non-existent core
	// Note that ResolveFQBN does not actually check this currently
	fqbn, err = cores.ParseFQBN("referenced:avr:dummy_invalid_core")
	require.Nil(t, err)
	require.NotNil(t, fqbn)
	pkg, platformRelease, board, props, buildPlatformRelease, err = pme.ResolveFQBN(fqbn)
	require.Nil(t, err)
	require.Equal(t, pkg, platformRelease.Platform.Package)
	require.NotNil(t, platformRelease)
	require.NotNil(t, platformRelease.Platform)
	require.Equal(t, platformRelease.Platform.String(), "referenced:avr")
	require.NotNil(t, board)
	require.Equal(t, board.Name(), "Referenced dummy with invalid core")
	require.NotNil(t, props)
	require.NotNil(t, buildPlatformRelease)
	require.NotNil(t, buildPlatformRelease.Platform)
	require.Equal(t, buildPlatformRelease.Platform.String(), "arduino:avr")
}

func TestBoardOptionsFunctions(t *testing.T) {
	pmb := packagemanager.NewBuilder(customHardware, customHardware, customHardware, customHardware, "test")
	pmb.LoadHardwareFromDirectory(customHardware)
	pm := pmb.Build()
	pme, release := pm.NewExplorer()
	defer release()

	nano, err := pme.FindBoardWithFQBN("arduino:avr:nano")
	require.Nil(t, err)
	require.NotNil(t, nano)
	require.Equal(t, nano.Name(), "Arduino Nano")

	nanoOptions := nano.GetConfigOptions()
	require.Equal(t, "Processor", nanoOptions.Get("cpu"))
	require.Equal(t, 1, nanoOptions.Size())
	nanoCPUValues := nano.GetConfigOptionValues("cpu")

	expectedNanoCPUValues := properties.NewMap()
	expectedNanoCPUValues.Set("atmega328", "ATmega328P")
	expectedNanoCPUValues.Set("atmega328old", "ATmega328P (Old Bootloader)")
	expectedNanoCPUValues.Set("atmega168", "ATmega168")
	require.EqualValues(t, expectedNanoCPUValues, nanoCPUValues)

	esp8266, err := pme.FindBoardWithFQBN("esp8266:esp8266:generic")
	require.Nil(t, err)
	require.NotNil(t, esp8266)
	require.Equal(t, esp8266.Name(), "Generic ESP8266 Module")

	esp8266Options := esp8266.GetConfigOptions()
	require.Equal(t, 13, esp8266Options.Size())
	require.Equal(t, "Builtin Led", esp8266Options.Get("led"))
	require.Equal(t, "Upload Speed", esp8266Options.Get("UploadSpeed"))

	esp8266UploadSpeedValues := esp8266.GetConfigOptionValues("UploadSpeed")
	for k, v := range esp8266UploadSpeedValues.AsMap() {
		// Some option values are missing for a particular OS: check that only the available options are listed
		require.Equal(t, k, v)
	}
}

func TestBoardOrdering(t *testing.T) {
	pmb := packagemanager.NewBuilder(dataDir1, dataDir1.Join("packages"), nil, nil, "")
	_ = pmb.LoadHardwareFromDirectories(paths.NewPathList(dataDir1.Join("packages").String()))
	pm := pmb.Build()
	pme, release := pm.NewExplorer()
	defer release()

	pl := pme.FindPlatform(&packagemanager.PlatformReference{
		Package:              "arduino",
		PlatformArchitecture: "avr",
	})
	require.NotNil(t, pl)
	plReleases := pl.GetAllInstalled()
	require.NotEmpty(t, plReleases)
	avr := plReleases[0]
	res := []string{}
	for _, board := range avr.GetBoards() {
		res = append(res, board.Name())
	}
	expected := []string{
		"Arduino Yún",
		"Arduino Uno",
		"Arduino Duemilanove or Diecimila",
		"Arduino Nano",
		"Arduino Mega or Mega 2560",
		"Arduino Mega ADK",
		"Arduino Leonardo",
		"Arduino Leonardo ETH",
		"Arduino Micro",
		"Arduino Esplora",
		"Arduino Mini",
		"Arduino Ethernet",
		"Arduino Fio",
		"Arduino BT",
		"LilyPad Arduino USB",
		"LilyPad Arduino",
		"Arduino Pro or Pro Mini",
		"Arduino NG or older",
		"Arduino Robot Control",
		"Arduino Robot Motor",
		"Arduino Gemma",
		"Adafruit Circuit Playground",
		"Arduino Yún Mini",
		"Arduino Industrial 101",
		"Linino One",
		"Arduino Uno WiFi",
	}
	require.Equal(t, expected, res)
}

func TestFindToolsRequiredForBoard(t *testing.T) {
	os.Setenv("ARDUINO_DATA_DIR", dataDir1.String())
	configuration.Settings = configuration.Init("")
	pmb := packagemanager.NewBuilder(
		dataDir1,
		configuration.PackagesDir(configuration.Settings),
		configuration.DownloadsDir(configuration.Settings),
		dataDir1,
		"test",
	)

	loadIndex := func(addr string) {
		res, err := url.Parse(addr)
		require.NoError(t, err)
		require.NoError(t, pmb.LoadPackageIndex(res))
	}
	loadIndex("https://dl.espressif.com/dl/package_esp32_index.json")
	loadIndex("http://arduino.esp8266.com/stable/package_esp8266com_index.json")
	loadIndex("https://adafruit.github.io/arduino-board-index/package_adafruit_index.json")
	loadIndex("https://test.com/package_test_index.json") // this is not downloaded, it just picks the "local cached" file package_test_index.json

	// We ignore the errors returned since they might not be necessarily blocking
	// but just warnings for the user, like in the case a board is not loaded
	// because of malformed menus
	pmb.LoadHardware()
	pm := pmb.Build()
	pme, release := pm.NewExplorer()
	defer release()

	esp32, err := pme.FindBoardWithFQBN("esp32:esp32:esp32")
	require.NoError(t, err)
	esptool231 := pme.FindToolDependency(&cores.ToolDependency{
		ToolPackager: "esp32",
		ToolName:     "esptool",
		ToolVersion:  semver.ParseRelaxed("2.3.1"),
	})
	require.NotNil(t, esptool231)
	esptool0413 := pme.FindToolDependency(&cores.ToolDependency{
		ToolPackager: "esp8266",
		ToolName:     "esptool",
		ToolVersion:  semver.ParseRelaxed("0.4.13"),
	})
	require.NotNil(t, esptool0413)

	testPlatform := pme.FindPlatformRelease(&packagemanager.PlatformReference{
		Package:              "test",
		PlatformArchitecture: "avr",
		PlatformVersion:      semver.MustParse("1.1.0")})

	testConflictingToolsInDifferentPackages := func() {
		tools, err := pme.FindToolsRequiredForBuild(esp32.PlatformRelease, nil)
		require.NoError(t, err)
		require.Contains(t, tools, esptool231)
		require.NotContains(t, tools, esptool0413)
	}

	// As seen in https://github.com/arduino/arduino-cli/issues/73 the map randomess
	// may make the function fail half of the times. Repeating the test 10 times
	// greatly increases the chances to trigger the bad case.
	testConflictingToolsInDifferentPackages()
	testConflictingToolsInDifferentPackages()
	testConflictingToolsInDifferentPackages()
	testConflictingToolsInDifferentPackages()
	testConflictingToolsInDifferentPackages()
	testConflictingToolsInDifferentPackages()
	testConflictingToolsInDifferentPackages()
	testConflictingToolsInDifferentPackages()
	testConflictingToolsInDifferentPackages()
	testConflictingToolsInDifferentPackages()

	{
		// Test buildPlatform dependencies
		arduinoBossac180 := pme.FindToolDependency(&cores.ToolDependency{
			ToolPackager: "arduino",
			ToolName:     "bossac",
			ToolVersion:  semver.ParseRelaxed("1.8.0-48-gb176eee"),
		})
		require.NotNil(t, arduinoBossac180)
		testBossac175 := pme.FindToolDependency(&cores.ToolDependency{
			ToolPackager: "test",
			ToolName:     "bossac",
			ToolVersion:  semver.ParseRelaxed("1.7.5"),
		})
		require.NotNil(t, testBossac175)

		tools, err := pme.FindToolsRequiredForBuild(esp32.PlatformRelease, nil)
		require.NoError(t, err)
		require.Contains(t, tools, esptool231)
		require.NotContains(t, tools, esptool0413)
		// When building without testPlatform dependency, arduino:bossac should be selected
		// since it has the higher version
		require.NotContains(t, tools, testBossac175)
		require.Contains(t, tools, arduinoBossac180)

		tools, err = pme.FindToolsRequiredForBuild(esp32.PlatformRelease, testPlatform)
		require.NoError(t, err)
		require.Contains(t, tools, esptool231)
		require.NotContains(t, tools, esptool0413)
		// When building with testPlatform dependency, test:bossac should be selected
		// because it has dependency priority
		require.Contains(t, tools, testBossac175)
		require.NotContains(t, tools, arduinoBossac180)
	}

	feather, err := pme.FindBoardWithFQBN("adafruit:samd:adafruit_feather_m0_express")
	require.NoError(t, err)
	require.NotNil(t, feather)
	featherTools, err := pme.FindToolsRequiredForBuild(feather.PlatformRelease, nil)
	require.NoError(t, err)
	require.NotNil(t, featherTools)

	// Test when a package index requires two different version of the same tool
	// See: https://github.com/arduino/arduino-cli/issues/166#issuecomment-528295989
	bossac17 := pme.FindToolDependency(&cores.ToolDependency{
		ToolPackager: "arduino",
		ToolName:     "bossac",
		ToolVersion:  semver.ParseRelaxed("1.7.0"),
	})
	require.NotNil(t, bossac17)
	bossac18 := pme.FindToolDependency(&cores.ToolDependency{
		ToolPackager: "arduino",
		ToolName:     "bossac",
		ToolVersion:  semver.ParseRelaxed("1.8.0-48-gb176eee"),
	})
	require.NotNil(t, bossac18)
	require.Contains(t, featherTools, bossac17)
	require.Contains(t, featherTools, bossac18)

	// Check if the runtime variable is set correctly to the latest version
	uploadProperties := properties.NewMap()
	for _, requiredTool := range featherTools {
		uploadProperties.Merge(requiredTool.RuntimeProperties())
	}
	require.Equal(t, bossac18.InstallDir.String(), uploadProperties.Get("runtime.tools.bossac.path"))
}

func TestIdentifyBoard(t *testing.T) {
	pmb := packagemanager.NewBuilder(customHardware, customHardware, customHardware, customHardware, "test")
	pmb.LoadHardwareFromDirectory(customHardware)
	pm := pmb.Build()
	pme, release := pm.NewExplorer()
	defer release()

	identify := func(vid, pid string) []*cores.Board {
		return pme.IdentifyBoard(properties.NewFromHashmap(map[string]string{
			"vid": vid, "pid": pid,
		}))
	}
	require.Equal(t, "[arduino:avr:uno]", fmt.Sprintf("%v", identify("0x2341", "0x0001")))

	// Check indexed vid/pid format (vid.0/pid.0)
	require.Equal(t, "[test:avr:a]", fmt.Sprintf("%v", identify("0x9999", "0x0001")))
	require.Equal(t, "[test:avr:b]", fmt.Sprintf("%v", identify("0x9999", "0x0002")))
	require.Equal(t, "[test:avr:c]", fmt.Sprintf("%v", identify("0x9999", "0x0003")))
	require.Equal(t, "[test:avr:c]", fmt.Sprintf("%v", identify("0x9999", "0x0004")))
	// https://github.com/arduino/arduino-cli/issues/456
	require.Equal(t, "[test:avr:d]", fmt.Sprintf("%v", identify("0x9999", "0x0005")))
	// Check mixed case
	require.Equal(t, "[test:avr:e]", fmt.Sprintf("%v", identify("0xAB00", "0xcd00")))
	require.Equal(t, "[test:avr:e]", fmt.Sprintf("%v", identify("0xab00", "0xCD00")))
}

func TestPackageManagerClear(t *testing.T) {
	// Create a PackageManager and load the harware
	pmb := packagemanager.NewBuilder(customHardware, customHardware, customHardware, customHardware, "test")
	pmb.LoadHardwareFromDirectory(customHardware)
	pm := pmb.Build()

	// Creates another PackageManager but don't load the hardware
	emptyPmb := packagemanager.NewBuilder(customHardware, customHardware, customHardware, customHardware, "test")
	emptyPm := emptyPmb.Build()

	// Verifies they're not equal
	require.NotEqual(t, pm, emptyPm)

	// Clear the first PackageManager that contains loaded hardware
	emptyPmb.BuildIntoExistingPackageManager(pm)

	// Verifies both PackageManagers are now equal
	require.Equal(t, pm, emptyPm)
}

func TestFindToolsRequiredFromPlatformRelease(t *testing.T) {
	// Create all the necessary data to load discoveries
	fakePath := paths.New("fake-path")

	pmb := packagemanager.NewBuilder(fakePath, fakePath, fakePath, fakePath, "test")
	pack := pmb.GetOrCreatePackage("arduino")

	{
		// some tool
		tool := pack.GetOrCreateTool("some-tool")
		toolRelease := tool.GetOrCreateRelease(semver.ParseRelaxed("4.2.0"))
		// We set this to fake the tool is installed
		toolRelease.InstallDir = fakePath
	}

	{
		// some tool
		tool := pack.GetOrCreateTool("some-tool")
		toolRelease := tool.GetOrCreateRelease(semver.ParseRelaxed("5.6.7"))
		// We set this to fake the tool is installed
		toolRelease.InstallDir = fakePath
	}

	{
		// some other tool
		tool := pack.GetOrCreateTool("some-other-tool")
		toolRelease := tool.GetOrCreateRelease(semver.ParseRelaxed("6.6.6"))
		// We set this to fake the tool is installed
		toolRelease.InstallDir = fakePath
	}

	{
		// ble-discovery tool
		tool := pack.GetOrCreateTool("ble-discovery")
		toolRelease := tool.GetOrCreateRelease(semver.ParseRelaxed("1.0.0"))
		// We set this to fake the tool is installed
		toolRelease.InstallDir = fakePath
		tool.GetOrCreateRelease(semver.ParseRelaxed("0.1.0"))
	}

	{
		// serial-discovery tool
		tool := pack.GetOrCreateTool("serial-discovery")
		tool.GetOrCreateRelease(semver.ParseRelaxed("1.0.0"))
		toolRelease := tool.GetOrCreateRelease(semver.ParseRelaxed("0.1.0"))
		// We set this to fake the tool is installed
		toolRelease.InstallDir = fakePath
	}

	{
		// ble-monitor tool
		tool := pack.GetOrCreateTool("ble-monitor")
		toolRelease := tool.GetOrCreateRelease(semver.ParseRelaxed("1.0.0"))
		// We set this to fake the tool is installed
		toolRelease.InstallDir = fakePath
		tool.GetOrCreateRelease(semver.ParseRelaxed("0.1.0"))
	}

	{
		// serial-monitor tool
		tool := pack.GetOrCreateTool("serial-monitor")
		tool.GetOrCreateRelease(semver.ParseRelaxed("1.0.0"))
		toolRelease := tool.GetOrCreateRelease(semver.ParseRelaxed("0.1.0"))
		// We set this to fake the tool is installed
		toolRelease.InstallDir = fakePath
	}

	platform := pack.GetOrCreatePlatform("avr")
	release := platform.GetOrCreateRelease(semver.MustParse("1.0.0"))
	release.ToolDependencies = append(release.ToolDependencies, &cores.ToolDependency{
		ToolName:     "some-tool",
		ToolVersion:  semver.ParseRelaxed("4.2.0"),
		ToolPackager: "arduino",
	})
	release.ToolDependencies = append(release.ToolDependencies, &cores.ToolDependency{
		ToolName:     "some-other-tool",
		ToolVersion:  semver.ParseRelaxed("6.6.6"),
		ToolPackager: "arduino",
	})
	release.DiscoveryDependencies = append(release.DiscoveryDependencies, &cores.DiscoveryDependency{
		Name:     "ble-discovery",
		Packager: "arduino",
	})
	release.DiscoveryDependencies = append(release.DiscoveryDependencies, &cores.DiscoveryDependency{
		Name:     "serial-discovery",
		Packager: "arduino",
	})
	release.MonitorDependencies = append(release.MonitorDependencies, &cores.MonitorDependency{
		Name:     "ble-monitor",
		Packager: "arduino",
	})
	release.MonitorDependencies = append(release.MonitorDependencies, &cores.MonitorDependency{
		Name:     "serial-monitor",
		Packager: "arduino",
	})
	// We set this to fake the platform is installed
	release.InstallDir = fakePath

	pm := pmb.Build()
	pme, pmeRelease := pm.NewExplorer()
	defer pmeRelease()
	tools, err := pme.FindToolsRequiredFromPlatformRelease(release)
	require.NoError(t, err)
	require.Len(t, tools, 6)
}

func TestFindPlatformReleaseDependencies(t *testing.T) {
	pmb := packagemanager.NewBuilder(nil, nil, nil, nil, "test")
	pmb.LoadPackageIndexFromFile(paths.New("testdata", "package_tooltest_index.json"))
	pm := pmb.Build()
	pme, release := pm.NewExplorer()
	defer release()

	pl, tools, err := pme.FindPlatformReleaseDependencies(&packagemanager.PlatformReference{Package: "test", PlatformArchitecture: "avr"})
	require.NoError(t, err)
	require.NotNil(t, pl)
	require.Len(t, tools, 3)
	require.Equal(t, "[test:some-tool@0.42.0 test:discovery-tool@0.42.0 test:monitor-tool@0.42.0]", fmt.Sprint(tools))
}

func TestLegacyPackageConversionToPluggableDiscovery(t *testing.T) {
	// Pass nil, since these paths are only used for installing
	pmb := packagemanager.NewBuilder(nil, nil, nil, nil, "test")
	// Hardware from main packages directory
	pmb.LoadHardwareFromDirectory(dataDir1.Join("packages"))
	pm := pmb.Build()
	pme, release := pm.NewExplorer()
	defer release()

	{
		fqbn, err := cores.ParseFQBN("esp32:esp32:esp32")
		require.NoError(t, err)
		require.NotNil(t, fqbn)
		_, platformRelease, board, _, _, err := pme.ResolveFQBN(fqbn)
		require.NoError(t, err)

		require.Equal(t, "esptool__pluggable_network", board.Properties.Get("upload.tool.network"))
		require.Equal(t, "esp32", board.Properties.Get("upload_port.0.board"))
		platformProps := platformRelease.Properties
		require.Equal(t, "builtin:serial-discovery", platformProps.Get("pluggable_discovery.required.0"))
		require.Equal(t, "builtin:mdns-discovery", platformProps.Get("pluggable_discovery.required.1"))
		require.Equal(t, "builtin:serial-monitor", platformProps.Get("pluggable_monitor.required.serial"))
		require.Equal(t, "{runtime.tools.esptool.path}", platformProps.Get("tools.esptool__pluggable_network.path"))
		require.Contains(t, platformProps.Get("tools.esptool__pluggable_network.cmd"), "esptool")
		require.Contains(t, platformProps.Get("tools.esptool__pluggable_network.network_cmd"), "{runtime.platform.path}/tools/espota")
		require.Equal(t, "esp32", platformProps.Get("tools.esptool__pluggable_network.upload.protocol"))
		require.Equal(t, "", platformProps.Get("tools.esptool__pluggable_network.upload.params.verbose"))
		require.Equal(t, "", platformProps.Get("tools.esptool__pluggable_network.upload.params.quiet"))
		require.Equal(t, "Password", platformProps.Get("tools.esptool__pluggable_network.upload.field.password"))
		require.Equal(t, "true", platformProps.Get("tools.esptool__pluggable_network.upload.field.password.secret"))
		require.Equal(t, "{network_cmd} -i \"{upload.port.address}\" -p \"{upload.port.properties.port}\" \"--auth={upload.field.password}\" -f \"{build.path}/{build.project_name}.bin\"", platformProps.Get("tools.esptool__pluggable_network.upload.pattern"))
	}
	{
		fqbn, err := cores.ParseFQBN("esp8266:esp8266:generic")
		require.NoError(t, err)
		require.NotNil(t, fqbn)
		_, platformRelease, board, _, _, err := pme.ResolveFQBN(fqbn)
		require.NoError(t, err)
		require.Equal(t, "esptool__pluggable_network", board.Properties.Get("upload.tool.network"))
		require.Equal(t, "generic", board.Properties.Get("upload_port.0.board"))
		platformProps := platformRelease.Properties
		require.Equal(t, "builtin:serial-discovery", platformProps.Get("pluggable_discovery.required.0"))
		require.Equal(t, "builtin:mdns-discovery", platformProps.Get("pluggable_discovery.required.1"))
		require.Equal(t, "builtin:serial-monitor", platformProps.Get("pluggable_monitor.required.serial"))
		require.Equal(t, "", platformProps.Get("tools.esptool__pluggable_network.path"))
		require.Equal(t, "{runtime.tools.python3.path}/python3", platformProps.Get("tools.esptool__pluggable_network.cmd"))
		require.Equal(t, "{runtime.tools.python3.path}/python3", platformProps.Get("tools.esptool__pluggable_network.network_cmd"))
		require.Equal(t, "esp", platformProps.Get("tools.esptool__pluggable_network.upload.protocol"))
		require.Equal(t, "", platformProps.Get("tools.esptool__pluggable_network.upload.params.verbose"))
		require.Equal(t, "", platformProps.Get("tools.esptool__pluggable_network.upload.params.quiet"))
		require.Equal(t, "Password", platformProps.Get("tools.esptool__pluggable_network.upload.field.password"))
		require.Equal(t, "true", platformProps.Get("tools.esptool__pluggable_network.upload.field.password.secret"))
		require.Equal(t, "\"{network_cmd}\" -I \"{runtime.platform.path}/tools/espota.py\" -i \"{upload.port.address}\" -p \"{upload.port.properties.port}\" \"--auth={upload.field.password}\" -f \"{build.path}/{build.project_name}.bin\"", platformProps.Get("tools.esptool__pluggable_network.upload.pattern"))
	}
	{
		fqbn, err := cores.ParseFQBN("arduino:avr:uno")
		require.NoError(t, err)
		require.NotNil(t, fqbn)
		_, platformRelease, board, _, _, err := pme.ResolveFQBN(fqbn)
		require.NoError(t, err)
		require.Equal(t, "avrdude__pluggable_network", board.Properties.Get("upload.tool.network"))
		require.Equal(t, "uno", board.Properties.Get("upload_port.4.board"))
		platformProps := platformRelease.Properties
		require.Equal(t, "builtin:serial-discovery", platformProps.Get("pluggable_discovery.required.0"))
		require.Equal(t, "builtin:mdns-discovery", platformProps.Get("pluggable_discovery.required.1"))
		require.Equal(t, "builtin:serial-monitor", platformProps.Get("pluggable_monitor.required.serial"))
		require.Equal(t, `"{network_cmd}" -address {upload.port.address} -port {upload.port.properties.port} -sketch "{build.path}/{build.project_name}.hex" -upload {upload.port.properties.endpoint_upload} -sync {upload.port.properties.endpoint_sync} -reset {upload.port.properties.endpoint_reset} -sync_exp {upload.port.properties.sync_return}`, platformProps.Get("tools.avrdude__pluggable_network.upload.pattern"))
	}
}

type plainFormatter struct{}

func (f *plainFormatter) Format(e *logrus.Entry) ([]byte, error) {
	return []byte(e.Message), nil
}
func TestRunPostInstall(t *testing.T) {
	logger := logrus.StandardLogger()
	pmb := packagemanager.NewBuilder(nil, nil, nil, nil, "test")
	pm := pmb.Build()
	pme, release := pm.NewExplorer()
	defer release()

	// prepare dummy post install script
	dir := paths.New(t.TempDir())

	// capture logs for inspection
	infoBuffer, errorBuffer := bytes.NewBuffer([]byte{}), bytes.NewBuffer([]byte{})
	logger.SetOutput(io.Discard)
	logger.SetFormatter(&plainFormatter{})
	logger.AddHook(&writer.Hook{
		Writer: infoBuffer,
		LogLevels: []logrus.Level{
			logrus.InfoLevel,
		},
	})
	logger.AddHook(&writer.Hook{
		Writer: errorBuffer,
		LogLevels: []logrus.Level{
			logrus.ErrorLevel,
		},
	})

	var scriptPath *paths.Path
	var err error
	if runtime.GOOS == "windows" {
		scriptPath = dir.Join("post_install.bat")

		err = scriptPath.WriteFile([]byte(
			`@echo off
			echo sent in stdout
			echo sent in stderr 1>&2`))
	} else {
		scriptPath = dir.Join("post_install.sh")
		err = scriptPath.WriteFile([]byte(
			`#!/bin/sh
			echo "sent in stdout"
			echo "sent in stderr" 1>&2`))
	}
	require.NoError(t, err)
	err = os.Chmod(scriptPath.String(), 0777)
	require.NoError(t, err)
	pme.RunPostInstallScript(dir)
	// `HasPrefix` because windows seem to add a trailing space at the end

	require.Equal(t, "sent in stdout", strings.Trim(infoBuffer.String(), "\n\r "))
	require.Equal(t, "sent in stderr", strings.Trim(errorBuffer.String(), "\n\r "))
}
