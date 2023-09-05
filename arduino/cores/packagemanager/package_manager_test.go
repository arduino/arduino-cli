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

package packagemanager

import (
	"fmt"
	"net/url"
	"os"
	"runtime"
	"strings"
	"testing"

	"github.com/arduino/arduino-cli/arduino/cores"
	"github.com/arduino/arduino-cli/configuration"
	"github.com/arduino/go-paths-helper"
	"github.com/arduino/go-properties-orderedmap"
	"github.com/stretchr/testify/require"
	semver "go.bug.st/relaxed-semver"
)

var customHardware = paths.New("testdata", "custom_hardware")
var dataDir1 = paths.New("testdata", "data_dir_1")

// Intended to be used alongside dataDir1
var extraHardware = paths.New("testdata", "extra_hardware")

func TestFindBoardWithFQBN(t *testing.T) {
	pmb := NewBuilder(customHardware, customHardware, customHardware, customHardware, "test")
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
	pmb := NewBuilder(nil, nil, nil, nil, "test")
	// Hardware from main packages directory
	pmb.LoadHardwareFromDirectory(dataDir1.Join("packages"))
	// This contains the arduino:avr core
	pmb.LoadHardwareFromDirectory(customHardware)
	// This contains the referenced:avr core
	pmb.LoadHardwareFromDirectory(extraHardware)
	pm := pmb.Build()
	pme, release := pm.NewExplorer()
	defer release()

	t.Run("NormalizeFQBN", func(t *testing.T) {
		testNormalization := func(in, expected string) {
			fqbn, err := cores.ParseFQBN(in)
			require.Nil(t, err)
			require.NotNil(t, fqbn)
			normalized, err := pme.NormalizeFQBN(fqbn)
			if expected == "ERROR" {
				require.Error(t, err)
				require.Nil(t, normalized)
			} else {
				require.NoError(t, err)
				require.NotNil(t, normalized)
				require.Equal(t, expected, normalized.String())
			}
		}
		testNormalization("arduino:avr:mega", "arduino:avr:mega")
		testNormalization("arduino:avr:mega:cpu=atmega2560", "arduino:avr:mega")
		testNormalization("arduino:avr:mega:cpu=atmega1280", "arduino:avr:mega:cpu=atmega1280")
		testNormalization("esp8266:esp8266:generic:baud=57600,wipe=sdk", "esp8266:esp8266:generic:baud=57600,wipe=sdk")
		testNormalization("esp8266:esp8266:generic:baud=115200,wipe=sdk", "esp8266:esp8266:generic:wipe=sdk")
		testNormalization("arduino:avr:mega:cpu=nonexistent", "ERROR")
		testNormalization("arduino:avr:mega:nonexistent=blah", "ERROR")
	})

	t.Run("BoardAndBuildPropertiesArduinoUno", func(t *testing.T) {
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

		require.Equal(t, "arduino", pkg.Name)
		require.Equal(t, "avr", platformRelease.Platform.Architecture)
		require.Equal(t, "uno", board.BoardID)
		require.Equal(t, "atmega328p", props.Get("build.mcu"))
	})

	t.Run("BoardAndBuildPropertiesArduinoMega", func(t *testing.T) {
		fqbn, err := cores.ParseFQBN("arduino:avr:mega")
		require.Nil(t, err)
		require.NotNil(t, fqbn)
		pkg, platformRelease, board, props, buildPlatformRelease, err := pme.ResolveFQBN(fqbn)
		require.Nil(t, err)
		require.Equal(t, pkg, platformRelease.Platform.Package)
		require.NotNil(t, platformRelease)
		require.NotNil(t, platformRelease.Platform)
		require.Equal(t, platformRelease.Platform.String(), "arduino:avr")
		require.NotNil(t, board)
		require.Equal(t, board.Name(), "Arduino Mega or Mega 2560")
		require.NotNil(t, props)
		require.Equal(t, platformRelease, buildPlatformRelease)
	})

	t.Run("BoardAndBuildPropertiesArduinoMegaWithNonDefaultCpuOption", func(t *testing.T) {
		fqbn, err := cores.ParseFQBN("arduino:avr:mega:cpu=atmega1280")
		require.Nil(t, err)
		require.NotNil(t, fqbn)
		pkg, platformRelease, board, props, buildPlatformRelease, err := pme.ResolveFQBN(fqbn)
		require.Nil(t, err)
		require.Equal(t, pkg, platformRelease.Platform.Package)
		require.NotNil(t, platformRelease)
		require.NotNil(t, platformRelease.Platform)
		require.Equal(t, platformRelease, buildPlatformRelease)

		require.Equal(t, "arduino", pkg.Name)
		require.Equal(t, "avr", platformRelease.Platform.Architecture)
		require.Equal(t, "mega", board.BoardID)
		require.Equal(t, "atmega1280", props.Get("build.mcu"))
		require.Equal(t, "AVR_MEGA", props.Get("build.board"))
	})

	t.Run("BoardAndBuildPropertiesArduinoMegaWithDefaultCpuOption", func(t *testing.T) {
		fqbn, err := cores.ParseFQBN("arduino:avr:mega:cpu=atmega2560")
		require.Nil(t, err)
		require.NotNil(t, fqbn)
		pkg, platformRelease, board, props, buildPlatformRelease, err := pme.ResolveFQBN(fqbn)
		require.Nil(t, err)
		require.Equal(t, pkg, platformRelease.Platform.Package)
		require.NotNil(t, platformRelease)
		require.NotNil(t, platformRelease.Platform)
		require.Equal(t, platformRelease, buildPlatformRelease)

		require.Equal(t, "arduino", pkg.Name)
		require.Equal(t, "avr", platformRelease.Platform.Architecture)
		require.Equal(t, "mega", board.BoardID)
		require.Equal(t, "atmega2560", props.Get("build.mcu"))
		require.Equal(t, "AVR_MEGA2560", props.Get("build.board"))

	})

	t.Run("BoardAndBuildPropertiesForReferencedArduinoUno", func(t *testing.T) {
		// Test a board referenced from the main AVR arduino platform
		fqbn, err := cores.ParseFQBN("referenced:avr:uno")
		require.Nil(t, err)
		require.NotNil(t, fqbn)
		pkg, platformRelease, board, props, buildPlatformRelease, err := pme.ResolveFQBN(fqbn)
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
	})

	t.Run("BoardAndBuildPropertiesForArduinoDue", func(t *testing.T) {
		fqbn, err := cores.ParseFQBN("arduino:sam:arduino_due_x")
		require.Nil(t, err)
		require.NotNil(t, fqbn)
		pkg, platformRelease, board, props, buildPlatformRelease, err := pme.ResolveFQBN(fqbn)
		require.Nil(t, err)
		require.Equal(t, pkg, platformRelease.Platform.Package)
		require.Equal(t, platformRelease, buildPlatformRelease)

		require.Equal(t, "arduino", pkg.Name)
		require.Equal(t, "sam", platformRelease.Platform.Architecture)
		require.Equal(t, "arduino_due_x", board.BoardID)
		require.Equal(t, "cortex-m3", props.Get("build.mcu"))
	})

	t.Run("BoardAndBuildPropertiesForCustomArduinoYun", func(t *testing.T) {
		fqbn, err := cores.ParseFQBN("my_avr_platform:avr:custom_yun")
		require.Nil(t, err)
		require.NotNil(t, fqbn)
		pkg, platformRelease, board, props, buildPlatformRelease, err := pme.ResolveFQBN(fqbn)
		require.Nil(t, err)
		require.Equal(t, pkg, platformRelease.Platform.Package)
		require.NotEqual(t, platformRelease, buildPlatformRelease)

		require.Equal(t, "my_avr_platform", pkg.Name)
		require.Equal(t, "avr", platformRelease.Platform.Architecture)
		require.Equal(t, "custom_yun", board.BoardID)
		require.Equal(t, "atmega32u4", props.Get("build.mcu"))
		require.Equal(t, "AVR_YUN", props.Get("build.board"))
	})

	t.Run("BoardAndBuildPropertiesForWatterotCore", func(t *testing.T) {
		fqbn, err := cores.ParseFQBN("watterott:avr:attiny841:core=spencekonde,info=info")
		require.Nil(t, err)
		require.NotNil(t, fqbn)
		pkg, platformRelease, board, props, buildPlatformRelease, err := pme.ResolveFQBN(fqbn)
		require.Nil(t, err)
		require.Equal(t, pkg, platformRelease.Platform.Package)
		require.Equal(t, platformRelease, buildPlatformRelease)

		require.Equal(t, "watterott", pkg.Name)
		require.Equal(t, "avr", platformRelease.Platform.Architecture)
		require.Equal(t, "attiny841", board.BoardID)
		require.Equal(t, "tiny841", props.Get("build.core"))
		require.Equal(t, "tiny14", props.Get("build.variant"))
	})

	t.Run("BoardAndBuildPropertiesForReferencedFeatherM0", func(t *testing.T) {
		// Test a board referenced from the Adafruit SAMD core (this tests
		// deriving where the package and core name are different)
		fqbn, err := cores.ParseFQBN("referenced:samd:feather_m0")
		require.Nil(t, err)
		require.NotNil(t, fqbn)
		pkg, platformRelease, board, props, buildPlatformRelease, err := pme.ResolveFQBN(fqbn)
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
	})

	t.Run("BoardAndBuildPropertiesForNonExistentPackage", func(t *testing.T) {
		// Test a board referenced from a non-existent package
		fqbn, err := cores.ParseFQBN("referenced:avr:dummy_invalid_package")
		require.Nil(t, err)
		require.NotNil(t, fqbn)
		pkg, platformRelease, board, props, buildPlatformRelease, err := pme.ResolveFQBN(fqbn)
		require.NotNil(t, err)
		require.Equal(t, pkg, platformRelease.Platform.Package)
		require.NotNil(t, platformRelease)
		require.NotNil(t, platformRelease.Platform)
		require.Equal(t, platformRelease.Platform.String(), "referenced:avr")
		require.NotNil(t, board)
		require.Equal(t, board.Name(), "Referenced dummy with invalid package")
		require.Nil(t, props)
		require.Nil(t, buildPlatformRelease)
	})

	t.Run("BoardAndBuildPropertiesForNonExistentArchitecture", func(t *testing.T) {
		// Test a board referenced from a non-existent platform/architecture
		fqbn, err := cores.ParseFQBN("referenced:avr:dummy_invalid_platform")
		require.Nil(t, err)
		require.NotNil(t, fqbn)
		pkg, platformRelease, board, props, buildPlatformRelease, err := pme.ResolveFQBN(fqbn)
		require.NotNil(t, err)
		require.Equal(t, pkg, platformRelease.Platform.Package)
		require.NotNil(t, platformRelease)
		require.NotNil(t, platformRelease.Platform)
		require.Equal(t, platformRelease.Platform.String(), "referenced:avr")
		require.NotNil(t, board)
		require.Equal(t, board.Name(), "Referenced dummy with invalid platform")
		require.Nil(t, props)
		require.Nil(t, buildPlatformRelease)
	})

	t.Run("BoardAndBuildPropertiesForNonExistentCore", func(t *testing.T) {
		// Test a board referenced from a non-existent core
		// Note that ResolveFQBN does not actually check this currently
		fqbn, err := cores.ParseFQBN("referenced:avr:dummy_invalid_core")
		require.Nil(t, err)
		require.NotNil(t, fqbn)
		pkg, platformRelease, board, props, buildPlatformRelease, err := pme.ResolveFQBN(fqbn)
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
	})

	t.Run("AddBuildBoardPropertyIfMissing", func(t *testing.T) {
		fqbn, err := cores.ParseFQBN("my_avr_platform:avr:mymega")
		require.Nil(t, err)
		require.NotNil(t, fqbn)
		pkg, platformRelease, board, props, buildPlatformRelease, err := pme.ResolveFQBN(fqbn)
		require.Nil(t, err)
		require.Equal(t, pkg, platformRelease.Platform.Package)
		require.Equal(t, platformRelease, buildPlatformRelease)

		require.Equal(t, "my_avr_platform", pkg.Name)
		require.NotNil(t, platformRelease)
		require.NotNil(t, platformRelease.Platform)
		require.Equal(t, "avr", platformRelease.Platform.Architecture)
		require.Equal(t, "mymega", board.BoardID)
		require.Equal(t, "atmega2560", props.Get("build.mcu"))
		require.Equal(t, "AVR_MYMEGA", props.Get("build.board"))
	})

	t.Run("AddBuildBoardPropertyIfNotMissing", func(t *testing.T) {
		fqbn, err := cores.ParseFQBN("my_avr_platform:avr:mymega:cpu=atmega1280")
		require.Nil(t, err)
		require.NotNil(t, fqbn)
		pkg, platformRelease, board, props, buildPlatformRelease, err := pme.ResolveFQBN(fqbn)
		require.Nil(t, err)
		require.Equal(t, pkg, platformRelease.Platform.Package)
		require.Equal(t, platformRelease, buildPlatformRelease)

		require.Equal(t, "my_avr_platform", pkg.Name)
		require.Equal(t, "avr", platformRelease.Platform.Architecture)
		require.Equal(t, "mymega", board.BoardID)
		require.Equal(t, "atmega1280", props.Get("build.mcu"))
		require.Equal(t, "MYMEGA1280", props.Get("build.board"))
	})
}

func TestBoardOptionsFunctions(t *testing.T) {
	pmb := NewBuilder(customHardware, customHardware, customHardware, customHardware, "test")
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
	pmb := NewBuilder(dataDir1, dataDir1.Join("packages"), nil, nil, "")
	_ = pmb.LoadHardwareFromDirectories(paths.NewPathList(dataDir1.Join("packages").String()))
	pm := pmb.Build()
	pme, release := pm.NewExplorer()
	defer release()

	pl := pme.FindPlatform("arduino", "avr")
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
	t.Setenv("ARDUINO_DATA_DIR", dataDir1.String())
	configuration.Settings = configuration.Init("")
	pmb := NewBuilder(
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

	testPlatform := pme.FindPlatformRelease(&PlatformReference{
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
	pmb := NewBuilder(customHardware, customHardware, customHardware, customHardware, "test")
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
	pmb := NewBuilder(customHardware, customHardware, customHardware, customHardware, "test")
	pmb.LoadHardwareFromDirectory(customHardware)
	pm := pmb.Build()

	// Creates another PackageManager but don't load the hardware
	emptyPmb := NewBuilder(customHardware, customHardware, customHardware, customHardware, "test")
	emptyPm := emptyPmb.Build()

	// Verifies they're not equal
	require.NotEqual(t, pm, emptyPm)

	// Clear the first PackageManager that contains loaded hardware
	emptyPmb.BuildIntoExistingPackageManager(pm)

	// the discovery manager is maintained
	require.NotEqual(t, pm.discoveryManager, emptyPm.discoveryManager)
	// Verifies all other fields are assigned to target
	pm.discoveryManager = emptyPm.discoveryManager
	require.Equal(t, pm, emptyPm)
}

func TestFindToolsRequiredFromPlatformRelease(t *testing.T) {
	// Create all the necessary data to load discoveries
	fakePath, err := paths.TempDir().MkTempDir("fake-path")
	require.NoError(t, err)
	defer fakePath.RemoveAll()

	pmb := NewBuilder(fakePath, fakePath, fakePath, fakePath, "test")
	pack := pmb.GetOrCreatePackage("arduino")

	{
		// some tool
		tool := pack.GetOrCreateTool("some-tool")
		toolRelease := tool.GetOrCreateRelease(semver.ParseRelaxed("4.2.0"))
		// We set this to fake the tool is installed
		toolRelease.InstallDir = fakePath
		f, err := toolRelease.InstallDir.Join(toolRelease.Tool.Name + ".exe").Create()
		require.NoError(t, err)
		require.NoError(t, f.Close())
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
		f, err := toolRelease.InstallDir.Join(toolRelease.Tool.Name + ".exe").Create()
		require.NoError(t, err)
		require.NoError(t, f.Close())
	}

	{
		// ble-discovery tool
		tool := pack.GetOrCreateTool("ble-discovery")
		toolRelease := tool.GetOrCreateRelease(semver.ParseRelaxed("1.0.0"))
		// We set this to fake the tool is installed
		toolRelease.InstallDir = fakePath
		f, err := toolRelease.InstallDir.Join(toolRelease.Tool.Name + ".exe").Create()
		require.NoError(t, err)
		require.NoError(t, f.Close())
		tool.GetOrCreateRelease(semver.ParseRelaxed("0.1.0"))
	}

	{
		// serial-discovery tool
		tool := pack.GetOrCreateTool("serial-discovery")
		tool.GetOrCreateRelease(semver.ParseRelaxed("1.0.0"))
		toolRelease := tool.GetOrCreateRelease(semver.ParseRelaxed("0.1.0"))
		// We set this to fake the tool is installed
		toolRelease.InstallDir = fakePath
		f, err := toolRelease.InstallDir.Join(toolRelease.Tool.Name + ".exe").Create()
		require.NoError(t, err)
		require.NoError(t, f.Close())
	}

	{
		// ble-monitor tool
		tool := pack.GetOrCreateTool("ble-monitor")
		toolRelease := tool.GetOrCreateRelease(semver.ParseRelaxed("1.0.0"))
		// We set this to fake the tool is installed
		toolRelease.InstallDir = fakePath
		f, err := toolRelease.InstallDir.Join(toolRelease.Tool.Name + ".exe").Create()
		require.NoError(t, err)
		require.NoError(t, f.Close())
		tool.GetOrCreateRelease(semver.ParseRelaxed("0.1.0"))
	}

	{
		// serial-monitor tool
		tool := pack.GetOrCreateTool("serial-monitor")
		tool.GetOrCreateRelease(semver.ParseRelaxed("1.0.0"))
		toolRelease := tool.GetOrCreateRelease(semver.ParseRelaxed("0.1.0"))
		// We set this to fake the tool is installed
		toolRelease.InstallDir = fakePath
		f, err := toolRelease.InstallDir.Join(toolRelease.Tool.Name + ".exe").Create()
		require.NoError(t, err)
		require.NoError(t, f.Close())
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
	pmb := NewBuilder(nil, nil, nil, nil, "test")
	pmb.LoadPackageIndexFromFile(paths.New("testdata", "package_tooltest_index.json"))
	pm := pmb.Build()
	pme, release := pm.NewExplorer()
	defer release()

	pl, tools, err := pme.FindPlatformReleaseDependencies(&PlatformReference{Package: "test", PlatformArchitecture: "avr"})
	require.NoError(t, err)
	require.NotNil(t, pl)
	require.Len(t, tools, 3)
	require.Equal(t, "[test:some-tool@0.42.0 test:discovery-tool@0.42.0 test:monitor-tool@0.42.0]", fmt.Sprint(tools))
}

func TestLegacyPackageConversionToPluggableDiscovery(t *testing.T) {
	// Pass nil, since these paths are only used for installing
	pmb := NewBuilder(nil, nil, nil, nil, "test")
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

func TestVariantAndCoreSelection(t *testing.T) {
	// Pass nil, since these paths are only used for installing
	pmb := NewBuilder(nil, nil, nil, nil, "test")
	// Hardware from main packages directory
	pmb.LoadHardwareFromDirectory(dataDir1.Join("packages"))
	pm := pmb.Build()
	pme, release := pm.NewExplorer()
	defer release()

	requireSameFile := func(f1, f2 *paths.Path) {
		require.True(t, f1.EquivalentTo(f2), "%s must be equivalent to %s", f1, f2)
	}

	// build.core test suite
	t.Run("CoreWithoutSubstitutions", func(t *testing.T) {
		fqbn, err := cores.ParseFQBN("test2:avr:test")
		require.NoError(t, err)
		require.NotNil(t, fqbn)
		_, _, _, buildProps, _, err := pme.ResolveFQBN(fqbn)
		require.NoError(t, err)
		require.Equal(t, "arduino", buildProps.Get("build.core"))
		requireSameFile(buildProps.GetPath("build.core.path"), dataDir1.Join("packages", "test2", "hardware", "avr", "1.0.0", "cores", "arduino"))
	})
	t.Run("CoreWithSubstitutions", func(t *testing.T) {
		fqbn, err := cores.ParseFQBN("test2:avr:test2")
		require.NoError(t, err)
		require.NotNil(t, fqbn)
		_, _, _, buildProps, _, err := pme.ResolveFQBN(fqbn)
		require.NoError(t, err)
		require.Equal(t, "{my_core}", buildProps.Get("build.core"))
		require.Equal(t, "arduino", buildProps.Get("my_core"))
		requireSameFile(buildProps.GetPath("build.core.path"), dataDir1.Join("packages", "test2", "hardware", "avr", "1.0.0", "cores", "arduino"))
	})
	t.Run("CoreWithSubstitutionsAndDefaultOption", func(t *testing.T) {
		fqbn, err := cores.ParseFQBN("test2:avr:test3")
		require.NoError(t, err)
		require.NotNil(t, fqbn)
		_, _, _, buildProps, _, err := pme.ResolveFQBN(fqbn)
		require.NoError(t, err)
		require.Equal(t, "{my_core}", buildProps.Get("build.core"))
		require.Equal(t, "arduino", buildProps.Get("my_core"))
		requireSameFile(buildProps.GetPath("build.core.path"), dataDir1.Join("packages", "test2", "hardware", "avr", "1.0.0", "cores", "arduino"))
	})
	t.Run("CoreWithSubstitutionsAndNonDefaultOption", func(t *testing.T) {
		fqbn, err := cores.ParseFQBN("test2:avr:test3:core=referenced")
		require.NoError(t, err)
		require.NotNil(t, fqbn)
		_, _, _, buildProps, _, err := pme.ResolveFQBN(fqbn)
		require.NoError(t, err)
		require.Equal(t, "{my_core}", buildProps.Get("build.core"))
		require.Equal(t, "arduino:arduino", buildProps.Get("my_core"))
		requireSameFile(buildProps.GetPath("build.core.path"), dataDir1.Join("packages", "arduino", "hardware", "avr", "1.8.3", "cores", "arduino"))
	})

	// build.variant test suite
	t.Run("VariantWithoutSubstitutions", func(t *testing.T) {
		fqbn, err := cores.ParseFQBN("test2:avr:test4")
		require.NoError(t, err)
		require.NotNil(t, fqbn)
		_, _, _, buildProps, _, err := pme.ResolveFQBN(fqbn)
		require.NoError(t, err)
		require.Equal(t, "standard", buildProps.Get("build.variant"))
		requireSameFile(buildProps.GetPath("build.variant.path"), dataDir1.Join("packages", "test2", "hardware", "avr", "1.0.0", "variants", "standard"))
	})
	t.Run("VariantWithSubstitutions", func(t *testing.T) {
		fqbn, err := cores.ParseFQBN("test2:avr:test5")
		require.NoError(t, err)
		require.NotNil(t, fqbn)
		_, _, _, buildProps, _, err := pme.ResolveFQBN(fqbn)
		require.NoError(t, err)
		require.Equal(t, "{my_variant}", buildProps.Get("build.variant"))
		require.Equal(t, "standard", buildProps.Get("my_variant"))
		requireSameFile(buildProps.GetPath("build.variant.path"), dataDir1.Join("packages", "test2", "hardware", "avr", "1.0.0", "variants", "standard"))
	})
	t.Run("VariantWithSubstitutionsAndDefaultOption", func(t *testing.T) {
		fqbn, err := cores.ParseFQBN("test2:avr:test6")
		require.NoError(t, err)
		require.NotNil(t, fqbn)
		_, _, _, buildProps, _, err := pme.ResolveFQBN(fqbn)
		require.NoError(t, err)
		require.Equal(t, "{my_variant}", buildProps.Get("build.variant"))
		require.Equal(t, "standard", buildProps.Get("my_variant"))
		requireSameFile(buildProps.GetPath("build.variant.path"), dataDir1.Join("packages", "test2", "hardware", "avr", "1.0.0", "variants", "standard"))
	})
	t.Run("VariantWithSubstitutionsAndNonDefaultOption", func(t *testing.T) {
		fqbn, err := cores.ParseFQBN("test2:avr:test6:variant=referenced")
		require.NoError(t, err)
		require.NotNil(t, fqbn)
		_, _, _, buildProps, _, err := pme.ResolveFQBN(fqbn)
		require.NoError(t, err)
		require.Equal(t, "{my_variant}", buildProps.Get("build.variant"))
		require.Equal(t, "arduino:standard", buildProps.Get("my_variant"))
		requireSameFile(buildProps.GetPath("build.variant.path"), dataDir1.Join("packages", "arduino", "hardware", "avr", "1.8.3", "variants", "standard"))
	})
}

func TestRunScript(t *testing.T) {
	pmb := NewBuilder(nil, nil, nil, nil, "test")
	pm := pmb.Build()
	pme, release := pm.NewExplorer()
	defer release()

	// prepare dummy post install script
	dir := paths.New(t.TempDir())

	type Test struct {
		testName   string
		scriptName string
	}

	tests := []Test{
		{
			testName:   "PostInstallScript",
			scriptName: "post_install",
		},
		{
			testName:   "PreUninstallScript",
			scriptName: "pre_uninstall",
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			var scriptPath *paths.Path
			var err error
			if runtime.GOOS == "windows" {
				scriptPath = dir.Join(test.scriptName + ".bat")

				err = scriptPath.WriteFile([]byte(
					`@echo off
					echo sent in stdout
					echo sent in stderr 1>&2`))
			} else {
				scriptPath = dir.Join(test.scriptName + ".sh")
				err = scriptPath.WriteFile([]byte(
					`#!/bin/sh
					echo "sent in stdout"
					echo "sent in stderr" 1>&2`))
			}
			require.NoError(t, err)
			err = os.Chmod(scriptPath.String(), 0777)
			require.NoError(t, err)
			stdout, stderr, err := pme.RunPreOrPostScript(dir, test.scriptName)
			require.NoError(t, err)

			// `HasPrefix` because windows seem to add a trailing space at the end
			require.Equal(t, "sent in stdout", strings.Trim(string(stdout), "\n\r "))
			require.Equal(t, "sent in stderr", strings.Trim(string(stderr), "\n\r "))
		})
	}
}
