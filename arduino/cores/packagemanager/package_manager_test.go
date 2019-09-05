/*
 * This file is part of arduino-cli.
 *
 * Copyright 2018 ARDUINO SA (http://www.arduino.cc/)
 *
 * This software is released under the GNU General Public License version 3,
 * which covers the main part of arduino-cli.
 * The terms of this license can be found at:
 * https://www.gnu.org/licenses/gpl-3.0.en.html
 *
 * You can be released from the requirements of the above licenses by purchasing
 * a commercial license. Buying such a license is mandatory if you want to modify or
 * otherwise use the software for commercial activities involving the Arduino
 * software without disclosing the source code of your own applications. To purchase
 * a commercial license, send an email to license@arduino.cc.
 */

package packagemanager_test

import (
	"net/url"
	"testing"

	"github.com/arduino/arduino-cli/arduino/cores"
	"github.com/arduino/arduino-cli/arduino/cores/packagemanager"
	"github.com/arduino/arduino-cli/configs"
	"github.com/arduino/go-paths-helper"
	"github.com/arduino/go-properties-orderedmap"
	"github.com/stretchr/testify/require"
	semver "go.bug.st/relaxed-semver"
)

var customHardware = paths.New("testdata", "custom_hardware")
var dataDir1 = paths.New("testdata", "data_dir_1")

func TestFindBoardWithFQBN(t *testing.T) {
	pm := packagemanager.NewPackageManager(customHardware, customHardware, customHardware, customHardware)
	pm.LoadHardwareFromDirectory(customHardware)

	board, err := pm.FindBoardWithFQBN("arduino:avr:uno")
	require.Nil(t, err)
	require.NotNil(t, board)
	require.Equal(t, board.Name(), "Arduino/Genuino Uno")

	board, err = pm.FindBoardWithFQBN("arduino:avr:mega")
	require.Nil(t, err)
	require.NotNil(t, board)
	require.Equal(t, board.Name(), "Arduino/Genuino Mega or Mega 2560")
}

func TestBoardOptionsFunctions(t *testing.T) {
	pm := packagemanager.NewPackageManager(customHardware, customHardware, customHardware, customHardware)
	pm.LoadHardwareFromDirectory(customHardware)

	nano, err := pm.FindBoardWithFQBN("arduino:avr:nano")
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

	esp8266, err := pm.FindBoardWithFQBN("esp8266:esp8266:generic")
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

func TestFindToolsRequiredForBoard(t *testing.T) {
	pm := packagemanager.NewPackageManager(
		dataDir1,
		dataDir1.Join("packages"),
		dataDir1.Join("staging"),
		dataDir1)
	conf := &configs.Configuration{
		DataDir: dataDir1,
	}
	loadIndex := func(addr string) {
		res, err := url.Parse(addr)
		require.NoError(t, err)
		require.NoError(t, pm.LoadPackageIndex(res))
	}
	loadIndex("https://dl.espressif.com/dl/package_esp32_index.json")
	loadIndex("http://arduino.esp8266.com/stable/package_esp8266com_index.json")
	loadIndex("https://adafruit.github.io/arduino-board-index/package_adafruit_index.json")
	require.NoError(t, pm.LoadHardware(conf))
	esp32, err := pm.FindBoardWithFQBN("esp32:esp32:esp32")
	require.NoError(t, err)
	esptool231 := pm.FindToolDependency(&cores.ToolDependency{
		ToolPackager: "esp32",
		ToolName:     "esptool",
		ToolVersion:  semver.ParseRelaxed("2.3.1"),
	})
	require.NotNil(t, esptool231)
	esptool0413 := pm.FindToolDependency(&cores.ToolDependency{
		ToolPackager: "esp8266",
		ToolName:     "esptool",
		ToolVersion:  semver.ParseRelaxed("0.4.13"),
	})
	require.NotNil(t, esptool0413)

	testConflictingToolsInDifferentPackages := func() {
		tools, err := pm.FindToolsRequiredForBoard(esp32)
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

	feather, err := pm.FindBoardWithFQBN("adafruit:samd:adafruit_feather_m0_express")
	require.NoError(t, err)
	require.NotNil(t, feather)
	featherTools, err := pm.FindToolsRequiredForBoard(feather)
	require.NoError(t, err)
	require.NotNil(t, featherTools)

	// Test when a package index requires two different version of the same tool
	// See: https://github.com/arduino/arduino-cli/issues/166#issuecomment-528295989
	bossac17 := pm.FindToolDependency(&cores.ToolDependency{
		ToolPackager: "arduino",
		ToolName:     "bossac",
		ToolVersion:  semver.ParseRelaxed("1.7.0"),
	})
	require.NotNil(t, bossac17)
	bossac18 := pm.FindToolDependency(&cores.ToolDependency{
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
