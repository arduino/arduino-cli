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

package cores

import (
	"testing"

	properties "github.com/arduino/go-properties-orderedmap"
	"github.com/stretchr/testify/require"
)

var boardUnoProperties = properties.NewMap()

func init() {
	boardUnoProperties.Set("name", "Arduino/Genuino Uno")
	boardUnoProperties.Set("vid.0", "0x2341")
	boardUnoProperties.Set("pid.0", "0x0043")
	boardUnoProperties.Set("vid.1", "0x2341")
	boardUnoProperties.Set("pid.1", "0x0001")
	boardUnoProperties.Set("vid.2", "0x2A03")
	boardUnoProperties.Set("pid.2", "0x0043")
	boardUnoProperties.Set("vid.3", "0x2341")
	boardUnoProperties.Set("pid.3", "0x0243")
	boardUnoProperties.Set("upload.tool", "avrdude")
	boardUnoProperties.Set("upload.protocol", "arduino")
	boardUnoProperties.Set("upload.maximum_size", "32256")
	boardUnoProperties.Set("upload.maximum_data_size", "2048")
	boardUnoProperties.Set("upload.speed", "115200")
	boardUnoProperties.Set("bootloader.tool", "avrdude")
	boardUnoProperties.Set("bootloader.low_fuses", "0xFF")
	boardUnoProperties.Set("bootloader.high_fuses", "0xDE")
	boardUnoProperties.Set("bootloader.extended_fuses", "0xFD")
	boardUnoProperties.Set("bootloader.unlock_bits", "0x3F")
	boardUnoProperties.Set("bootloader.lock_bits", "0x0F")
	boardUnoProperties.Set("bootloader.file", "optiboot/optiboot_atmega328.hex")
	boardUnoProperties.Set("build.mcu", "atmega328p")
	boardUnoProperties.Set("build.f_cpu", "16000000L")
	boardUnoProperties.Set("build.board", "AVR_UNO")
	boardUnoProperties.Set("build.core", "arduino")
	boardUnoProperties.Set("build.variant", "standard")
}

var boardUno = &Board{
	BoardID:    "uno",
	Properties: boardUnoProperties,
	PlatformRelease: &PlatformRelease{
		Platform: &Platform{
			Architecture: "avr",
			Package: &Package{
				Name: "arduino",
			},
		},
		Menus: properties.NewMap(),
	},
}

var boardMegaProperties = properties.NewMap()

func init() {
	boardMegaProperties.Set("name", "Arduino/Genuino Mega or Mega 2560")
	boardMegaProperties.Set("vid.0", "0x2341")
	boardMegaProperties.Set("pid.0", "0x0010")
	boardMegaProperties.Set("vid.1", "0x2341")
	boardMegaProperties.Set("pid.1", "0x0042")
	boardMegaProperties.Set("vid.2", "0x2A03")
	boardMegaProperties.Set("pid.2", "0x0010")
	boardMegaProperties.Set("vid.3", "0x2A03")
	boardMegaProperties.Set("pid.3", "0x0042")
	boardMegaProperties.Set("vid.4", "0x2341")
	boardMegaProperties.Set("pid.4", "0x0210")
	boardMegaProperties.Set("vid.5", "0x2341")
	boardMegaProperties.Set("pid.5", "0x0242")
	boardMegaProperties.Set("upload.tool", "avrdude")
	boardMegaProperties.Set("upload.maximum_data_size", "8192")
	boardMegaProperties.Set("bootloader.tool", "avrdude")
	boardMegaProperties.Set("bootloader.low_fuses", "0xFF")
	boardMegaProperties.Set("bootloader.unlock_bits", "0x3F")
	boardMegaProperties.Set("bootloader.lock_bits", "0x0F")
	boardMegaProperties.Set("build.f_cpu", "16000000L")
	boardMegaProperties.Set("build.core", "arduino")
	boardMegaProperties.Set("build.variant", "mega")
	boardMegaProperties.Set("build.board", "AVR_MEGA2560")
	boardMegaProperties.Set("menu.cpu.atmega2560", "ATmega2560 (Mega 2560)")
	boardMegaProperties.Set("menu.cpu.atmega2560.upload.protocol", "wiring")
	boardMegaProperties.Set("menu.cpu.atmega2560.upload.maximum_size", "253952")
	boardMegaProperties.Set("menu.cpu.atmega2560.upload.speed", "115200")
	boardMegaProperties.Set("menu.cpu.atmega2560.bootloader.high_fuses", "0xD8")
	boardMegaProperties.Set("menu.cpu.atmega2560.bootloader.extended_fuses", "0xFD")
	boardMegaProperties.Set("menu.cpu.atmega2560.bootloader.file", "stk500v2/stk500boot_v2_mega2560.hex")
	boardMegaProperties.Set("menu.cpu.atmega2560.build.mcu", "atmega2560")
	boardMegaProperties.Set("menu.cpu.atmega2560.build.board", "AVR_MEGA2560")
	boardMegaProperties.Set("menu.cpu.atmega1280", "ATmega1280")
	boardMegaProperties.Set("menu.cpu.atmega1280.upload.protocol", "arduino")
	boardMegaProperties.Set("menu.cpu.atmega1280.upload.maximum_size", "126976")
	boardMegaProperties.Set("menu.cpu.atmega1280.upload.speed", "57600")
	boardMegaProperties.Set("menu.cpu.atmega1280.bootloader.high_fuses", "0xDA")
	boardMegaProperties.Set("menu.cpu.atmega1280.bootloader.extended_fuses", "0xF5")
	boardMegaProperties.Set("menu.cpu.atmega1280.bootloader.file", "atmega/ATmegaBOOT_168_atmega1280.hex")
	boardMegaProperties.Set("menu.cpu.atmega1280.build.mcu", "atmega1280")
	boardMegaProperties.Set("menu.cpu.atmega1280.build.board", "AVR_MEGA")
}

var boardMega = &Board{
	BoardID:    "mega",
	Properties: boardMegaProperties,

	PlatformRelease: &PlatformRelease{
		Platform: &Platform{
			Architecture: "avr",
			Package: &Package{
				Name: "arduino",
			},
		},
		Menus: properties.NewFromHashmap(map[string]string{
			"cpu": "Processor",
		}),
	},
}

var boardWatterottTiny841Properties = properties.NewMap()

func init() {
	boardWatterottTiny841Properties.Set("name", "ATtiny841 (8 MHz)")
	boardWatterottTiny841Properties.Set("menu.core.arduino", "Standard Arduino")
	boardWatterottTiny841Properties.Set("menu.core.arduino.build.core", "arduino:arduino")
	boardWatterottTiny841Properties.Set("menu.core.arduino.build.variant", "tiny14")
	boardWatterottTiny841Properties.Set("menu.core.spencekonde", "ATtiny841 (by Spence Konde)")
	boardWatterottTiny841Properties.Set("menu.core.spencekonde.build.core", "tiny841")
	boardWatterottTiny841Properties.Set("menu.core.spencekonde.build.variant", "tiny14")
	boardWatterottTiny841Properties.Set("menu.info.info", "Press Reset, when Uploading is shown.")
	boardWatterottTiny841Properties.Set("vid.0", "0x16D0")
	boardWatterottTiny841Properties.Set("pid.0", "0x0753")
	boardWatterottTiny841Properties.Set("bootloader.tool", "avrdude")
	boardWatterottTiny841Properties.Set("bootloader.low_fuses", "0xE2")
	boardWatterottTiny841Properties.Set("bootloader.high_fuses", "0xDD")
	boardWatterottTiny841Properties.Set("bootloader.extended_fuses", "0xFE")
	boardWatterottTiny841Properties.Set("bootloader.unlock_bits", "0xFF")
	boardWatterottTiny841Properties.Set("bootloader.lock_bits", "0xFF")
	boardWatterottTiny841Properties.Set("bootloader.file", "micronucleus-t841.hex")
	boardWatterottTiny841Properties.Set("upload.tool", "micronucleus")
	boardWatterottTiny841Properties.Set("upload.protocol", "usb")
	boardWatterottTiny841Properties.Set("upload.wait_for_upload_port", "false")
	boardWatterottTiny841Properties.Set("upload.use_1200bps_touch", "false")
	boardWatterottTiny841Properties.Set("upload.disable_flushing", "false")
	boardWatterottTiny841Properties.Set("upload.maximum_size", "6500")
	boardWatterottTiny841Properties.Set("build.mcu", "attiny841")
	boardWatterottTiny841Properties.Set("build.f_cpu", "8000000L")
	boardWatterottTiny841Properties.Set("build.board", "AVR_ATTINY841")

}

var boardWatterottTiny841 = &Board{
	BoardID:    "attiny841",
	Properties: boardWatterottTiny841Properties,
	PlatformRelease: &PlatformRelease{
		Platform: &Platform{
			Architecture: "avr",
			Package: &Package{
				Name: "watterott",
			},
		},
		Menus: properties.NewFromHashmap(map[string]string{
			"core": "Core",
			"info": "Info",
		}),
	},
}

func TestBoardPropertiesName(t *testing.T) {
	require.Equal(t, boardUno.Name(), "Arduino/Genuino Uno", "board name")
	require.Equal(t, boardMega.Name(), "Arduino/Genuino Mega or Mega 2560", "board name")
}

func TestBoardFQBN(t *testing.T) {
	require.Equal(t, boardUno.FQBN(), "arduino:avr:uno", "board FQBN")
	require.Equal(t, boardUno.String(), "arduino:avr:uno", "board to string")
	require.Equal(t, boardMega.FQBN(), "arduino:avr:mega", "board FQBN")
	require.Equal(t, boardMega.String(), "arduino:avr:mega", "board to string")
}

func TestBoard(t *testing.T) {
	require.True(t, boardUno.HasUsbID("0x2341", "0x0043"), "has usb 2341:0043")
	require.True(t, boardUno.HasUsbID("0x2341", "0x0001"), "has usb 2341:0001")
	require.True(t, boardUno.HasUsbID("0x2A03", "0x0043"), "has usb 2A03:0043")
	require.True(t, boardUno.HasUsbID("0x2341", "0x0243"), "has usb 2341:0243")
	require.False(t, boardUno.HasUsbID("0x1A03", "0x0001"), "has usb 1A03:0001")
	require.False(t, boardUno.HasUsbID("0x2A03", "0x0143"), "has usb 2A03:0143")

	require.True(t, boardMega.HasUsbID("0x2341", "0x0010"), "has usb 2341:0010")
	require.True(t, boardMega.HasUsbID("0x2341", "0x0042"), "has usb 2341:0042")
	require.True(t, boardMega.HasUsbID("0x2A03", "0x0010"), "has usb 2A03:0010")
	require.True(t, boardMega.HasUsbID("0x2A03", "0x0042"), "has usb 2A03:0042")
	require.False(t, boardMega.HasUsbID("0x1A03", "0x0042"), "has usb 1A03:0042")
	require.False(t, boardMega.HasUsbID("0x2A03", "0x0043"), "has usb 2A03:0043")
}

func TestBoardOptions(t *testing.T) {
	expConf2560 := properties.NewMap()
	expConf2560.Set("bootloader.extended_fuses", "0xFD")
	expConf2560.Set("bootloader.file", "stk500v2/stk500boot_v2_mega2560.hex")
	expConf2560.Set("bootloader.high_fuses", "0xD8")
	expConf2560.Set("bootloader.lock_bits", "0x0F")
	expConf2560.Set("bootloader.low_fuses", "0xFF")
	expConf2560.Set("bootloader.tool", "avrdude")
	expConf2560.Set("bootloader.unlock_bits", "0x3F")
	expConf2560.Set("build.board", "AVR_MEGA2560")
	expConf2560.Set("build.core", "arduino")
	expConf2560.Set("build.f_cpu", "16000000L")
	expConf2560.Set("build.mcu", "atmega2560")
	expConf2560.Set("build.variant", "mega")
	expConf2560.Set("menu.cpu.atmega1280", "ATmega1280")
	expConf2560.Set("menu.cpu.atmega1280.bootloader.extended_fuses", "0xF5")
	expConf2560.Set("menu.cpu.atmega1280.bootloader.file", "atmega/ATmegaBOOT_168_atmega1280.hex")
	expConf2560.Set("menu.cpu.atmega1280.bootloader.high_fuses", "0xDA")
	expConf2560.Set("menu.cpu.atmega1280.build.board", "AVR_MEGA")
	expConf2560.Set("menu.cpu.atmega1280.build.mcu", "atmega1280")
	expConf2560.Set("menu.cpu.atmega1280.upload.maximum_size", "126976")
	expConf2560.Set("menu.cpu.atmega1280.upload.protocol", "arduino")
	expConf2560.Set("menu.cpu.atmega1280.upload.speed", "57600")
	expConf2560.Set("menu.cpu.atmega2560", "ATmega2560 (Mega 2560)")
	expConf2560.Set("menu.cpu.atmega2560.bootloader.extended_fuses", "0xFD")
	expConf2560.Set("menu.cpu.atmega2560.bootloader.file", "stk500v2/stk500boot_v2_mega2560.hex")
	expConf2560.Set("menu.cpu.atmega2560.bootloader.high_fuses", "0xD8")
	expConf2560.Set("menu.cpu.atmega2560.build.board", "AVR_MEGA2560")
	expConf2560.Set("menu.cpu.atmega2560.build.mcu", "atmega2560")
	expConf2560.Set("menu.cpu.atmega2560.upload.maximum_size", "253952")
	expConf2560.Set("menu.cpu.atmega2560.upload.protocol", "wiring")
	expConf2560.Set("menu.cpu.atmega2560.upload.speed", "115200")
	expConf2560.Set("name", "Arduino/Genuino Mega or Mega 2560")
	expConf2560.Set("pid.0", "0x0010")
	expConf2560.Set("pid.1", "0x0042")
	expConf2560.Set("pid.2", "0x0010")
	expConf2560.Set("pid.3", "0x0042")
	expConf2560.Set("pid.4", "0x0210")
	expConf2560.Set("pid.5", "0x0242")
	expConf2560.Set("upload.maximum_data_size", "8192")
	expConf2560.Set("upload.maximum_size", "253952")
	expConf2560.Set("upload.protocol", "wiring")
	expConf2560.Set("upload.speed", "115200")
	expConf2560.Set("upload.tool", "avrdude")
	expConf2560.Set("vid.0", "0x2341")
	expConf2560.Set("vid.1", "0x2341")
	expConf2560.Set("vid.2", "0x2A03")
	expConf2560.Set("vid.3", "0x2A03")
	expConf2560.Set("vid.4", "0x2341")
	expConf2560.Set("vid.5", "0x2341")

	conf2560, err := boardMega.GeneratePropertiesForConfiguration("cpu=atmega2560")
	require.NoError(t, err, "generating cpu=atmega2560 configuration")
	require.EqualValues(t, expConf2560.AsMap(), conf2560.AsMap(), "configuration for cpu=atmega2560")
	require.EqualValues(t, map[string]string{"cpu": "Processor"}, boardMega.GetConfigOptions().AsMap())
	require.EqualValues(t, map[string]string{
		"atmega1280": "ATmega1280",
		"atmega2560": "ATmega2560 (Mega 2560)",
	}, boardMega.GetConfigOptionValues("cpu").AsMap())
	require.EqualValues(t, map[string]string{"cpu": "atmega2560"}, boardMega.defaultConfig.AsMap())
	expConf1280 := properties.NewMap()
	expConf1280.Set("bootloader.extended_fuses", "0xF5")
	expConf1280.Set("bootloader.file", "atmega/ATmegaBOOT_168_atmega1280.hex")
	expConf1280.Set("bootloader.high_fuses", "0xDA")
	expConf1280.Set("bootloader.lock_bits", "0x0F")
	expConf1280.Set("bootloader.low_fuses", "0xFF")
	expConf1280.Set("bootloader.tool", "avrdude")
	expConf1280.Set("bootloader.unlock_bits", "0x3F")
	expConf1280.Set("build.board", "AVR_MEGA")
	expConf1280.Set("build.core", "arduino")
	expConf1280.Set("build.f_cpu", "16000000L")
	expConf1280.Set("build.mcu", "atmega1280")
	expConf1280.Set("build.variant", "mega")
	expConf1280.Set("menu.cpu.atmega1280", "ATmega1280")
	expConf1280.Set("menu.cpu.atmega1280.bootloader.extended_fuses", "0xF5")
	expConf1280.Set("menu.cpu.atmega1280.bootloader.file", "atmega/ATmegaBOOT_168_atmega1280.hex")
	expConf1280.Set("menu.cpu.atmega1280.bootloader.high_fuses", "0xDA")
	expConf1280.Set("menu.cpu.atmega1280.build.board", "AVR_MEGA")
	expConf1280.Set("menu.cpu.atmega1280.build.mcu", "atmega1280")
	expConf1280.Set("menu.cpu.atmega1280.upload.maximum_size", "126976")
	expConf1280.Set("menu.cpu.atmega1280.upload.protocol", "arduino")
	expConf1280.Set("menu.cpu.atmega1280.upload.speed", "57600")
	expConf1280.Set("menu.cpu.atmega2560", "ATmega2560 (Mega 2560)")
	expConf1280.Set("menu.cpu.atmega2560.bootloader.extended_fuses", "0xFD")
	expConf1280.Set("menu.cpu.atmega2560.bootloader.file", "stk500v2/stk500boot_v2_mega2560.hex")
	expConf1280.Set("menu.cpu.atmega2560.bootloader.high_fuses", "0xD8")
	expConf1280.Set("menu.cpu.atmega2560.build.board", "AVR_MEGA2560")
	expConf1280.Set("menu.cpu.atmega2560.build.mcu", "atmega2560")
	expConf1280.Set("menu.cpu.atmega2560.upload.maximum_size", "253952")
	expConf1280.Set("menu.cpu.atmega2560.upload.protocol", "wiring")
	expConf1280.Set("menu.cpu.atmega2560.upload.speed", "115200")
	expConf1280.Set("name", "Arduino/Genuino Mega or Mega 2560")
	expConf1280.Set("pid.0", "0x0010")
	expConf1280.Set("pid.1", "0x0042")
	expConf1280.Set("pid.2", "0x0010")
	expConf1280.Set("pid.3", "0x0042")
	expConf1280.Set("pid.4", "0x0210")
	expConf1280.Set("pid.5", "0x0242")
	expConf1280.Set("upload.maximum_data_size", "8192")
	expConf1280.Set("upload.maximum_size", "126976")
	expConf1280.Set("upload.protocol", "arduino")
	expConf1280.Set("upload.speed", "57600")
	expConf1280.Set("upload.tool", "avrdude")
	expConf1280.Set("vid.0", "0x2341")
	expConf1280.Set("vid.1", "0x2341")
	expConf1280.Set("vid.2", "0x2A03")
	expConf1280.Set("vid.3", "0x2A03")
	expConf1280.Set("vid.4", "0x2341")
	expConf1280.Set("vid.5", "0x2341")
	conf1280, err := boardMega.GeneratePropertiesForConfiguration("cpu=atmega1280")
	require.NoError(t, err, "generating cpu=atmega1280 configuration")
	require.EqualValues(t, expConf1280.AsMap(), conf1280.AsMap(), "configuration for cpu=atmega1280")

	_, err = boardMega.GeneratePropertiesForConfiguration("cpu=atmegassss")
	require.Error(t, err, "generating cpu=atmegassss configuration")

	_, err = boardUno.GeneratePropertiesForConfiguration("cpu=atmega1280")
	require.Error(t, err, "generating cpu=atmega1280 configuration")

	expWatterott := properties.NewMap()
	expWatterott.Set("bootloader.extended_fuses", "0xFE")
	expWatterott.Set("bootloader.file", "micronucleus-t841.hex")
	expWatterott.Set("bootloader.high_fuses", "0xDD")
	expWatterott.Set("bootloader.lock_bits", "0xFF")
	expWatterott.Set("bootloader.low_fuses", "0xE2")
	expWatterott.Set("bootloader.tool", "avrdude")
	expWatterott.Set("bootloader.unlock_bits", "0xFF")
	expWatterott.Set("build.board", "AVR_ATTINY841")
	expWatterott.Set("build.core", "tiny841")
	expWatterott.Set("build.f_cpu", "8000000L")
	expWatterott.Set("build.mcu", "attiny841")
	expWatterott.Set("build.variant", "tiny14")
	expWatterott.Set("menu.core.arduino", "Standard Arduino")
	expWatterott.Set("menu.core.arduino.build.core", "arduino:arduino")
	expWatterott.Set("menu.core.arduino.build.variant", "tiny14")
	expWatterott.Set("menu.core.spencekonde", "ATtiny841 (by Spence Konde)")
	expWatterott.Set("menu.core.spencekonde.build.core", "tiny841")
	expWatterott.Set("menu.core.spencekonde.build.variant", "tiny14")
	expWatterott.Set("menu.info.info", "Press Reset, when Uploading is shown.")
	expWatterott.Set("name", "ATtiny841 (8 MHz)")
	expWatterott.Set("pid.0", "0x0753")
	expWatterott.Set("upload.disable_flushing", "false")
	expWatterott.Set("upload.maximum_size", "6500")
	expWatterott.Set("upload.protocol", "usb")
	expWatterott.Set("upload.tool", "micronucleus")
	expWatterott.Set("upload.use_1200bps_touch", "false")
	expWatterott.Set("upload.wait_for_upload_port", "false")
	expWatterott.Set("vid.0", "0x16D0")
	confWatterott, err := boardWatterottTiny841.GeneratePropertiesForConfiguration("core=spencekonde,info=info")
	require.NoError(t, err, "generating core=spencekonde,info=info configuration")
	require.EqualValues(t, expWatterott.AsMap(), confWatterott.AsMap(), "generating core=spencekonde,info=info configuration")

	// data, err := json.MarshalIndent(prop, "", "  ")
	// require.NoError(t, err, "marshaling result")
	// fmt.Print(string(data))
}

func TestOSSpecificBoardOptions(t *testing.T) {
	boardWihOSSpecificOptionProperties := properties.NewMap()
	boardWihOSSpecificOptionProperties.Set("menu.UploadSpeed.115200", "115200")
	boardWihOSSpecificOptionProperties.Set("menu.UploadSpeed.115200.upload.speed", "115200")
	boardWihOSSpecificOptionProperties.Set("menu.UploadSpeed.9600", "9600")
	boardWihOSSpecificOptionProperties.Set("menu.UploadSpeed.9600.upload.speed", "9600")
	boardWihOSSpecificOptionProperties.Set("menu.UploadSpeed.57600", "57600")
	boardWihOSSpecificOptionProperties.Set("menu.UploadSpeed.57600.upload.speed", "57600")
	boardWihOSSpecificOptionProperties.Set("menu.UploadSpeed.230400", "230400")
	boardWihOSSpecificOptionProperties.Set("menu.UploadSpeed.230400.macosx", "230400")
	boardWihOSSpecificOptionProperties.Set("menu.UploadSpeed.230400.upload.speed", "230400")
	boardWihOSSpecificOptionProperties.Set("menu.UploadSpeed.256000.windows", "256000")
	boardWihOSSpecificOptionProperties.Set("menu.UploadSpeed.256000.upload.speed", "256000")
	boardWihOSSpecificOptionProperties.Set("menu.UploadSpeed.460800", "460800")
	boardWihOSSpecificOptionProperties.Set("menu.UploadSpeed.460800.macosx", "460800")
	boardWihOSSpecificOptionProperties.Set("menu.UploadSpeed.460800.upload.speed", "460800")
	boardWihOSSpecificOptionProperties.Set("menu.UploadSpeed.512000.windows", "512000")
	boardWihOSSpecificOptionProperties.Set("menu.UploadSpeed.512000.upload.speed", "512000")
	boardWihOSSpecificOptionProperties.Set("menu.UploadSpeed.921600", "921600")
	boardWihOSSpecificOptionProperties.Set("menu.UploadSpeed.921600.upload.speed", "921600")

	boardWithOSSpecificOptions := &Board{
		BoardID:    "test",
		Properties: boardWihOSSpecificOptionProperties,
		PlatformRelease: &PlatformRelease{
			Platform: &Platform{
				Architecture: "test",
				Package: &Package{
					Name: "test",
				},
			},
			Menus: properties.NewFromHashmap(map[string]string{
				"UploadSpeed": "Upload Speed",
			}),
		},
	}

	_, err := boardWithOSSpecificOptions.GeneratePropertiesForConfiguration("UploadSpeed=256000")
	require.Error(t, err)
}

func TestBoardMatching(t *testing.T) {
	brd01 := &Board{
		Properties: properties.NewFromHashmap(map[string]string{
			"upload_port.pid": "0x0010",
			"upload_port.vid": "0x2341",
		}),
	}
	require.True(t, brd01.IsBoardMatchingIDProperties(properties.NewFromHashmap(map[string]string{
		"pid": "0x0010",
		"vid": "0x2341",
	})))
	require.False(t, brd01.IsBoardMatchingIDProperties(properties.NewFromHashmap(map[string]string{
		"pid": "xxx",
		"vid": "0x2341",
	})))
	require.False(t, brd01.IsBoardMatchingIDProperties(properties.NewFromHashmap(map[string]string{
		"pid": "0x0010",
	})))
	// Extra port properties are OK
	require.True(t, brd01.IsBoardMatchingIDProperties(properties.NewFromHashmap(map[string]string{
		"pid":    "0x0010",
		"vid":    "0x2341",
		"serial": "942947289347893247",
	})))

	// Indexed identifications
	brd02 := &Board{
		Properties: properties.NewFromHashmap(map[string]string{
			"upload_port.0.pid": "0x0010",
			"upload_port.0.vid": "0x2341",
			"upload_port.1.pid": "0x0042",
			"upload_port.1.vid": "0x2341",
			"upload_port.2.pid": "0x0010",
			"upload_port.2.vid": "0x2A03",
			"upload_port.3.pid": "0x0042",
			"upload_port.3.vid": "0x2A03",
			"upload_port.4.pid": "0x0210",
			"upload_port.4.vid": "0x2341",
			"upload_port.5.pid": "0x0242",
			"upload_port.5.vid": "0x2341",
		}),
	}
	require.True(t, brd02.IsBoardMatchingIDProperties(properties.NewFromHashmap(map[string]string{
		"pid": "0x0242",
		"vid": "0x2341",
	})))
	require.True(t, brd02.IsBoardMatchingIDProperties(properties.NewFromHashmap(map[string]string{
		"pid":    "0x0242",
		"vid":    "0x2341",
		"serial": "897439287289347",
	})))

	// Indexed starting from 1
	brd03 := &Board{
		Properties: properties.NewFromHashmap(map[string]string{
			"upload_port.1.pid": "0x0042",
			"upload_port.1.vid": "0x2341",
			"upload_port.2.pid": "0x0010",
			"upload_port.2.vid": "0x2A03",
			"upload_port.3.pid": "0x0042",
			"upload_port.3.vid": "0x2A03",
			"upload_port.4.pid": "0x0210",
			"upload_port.4.vid": "0x2341",
			"upload_port.5.pid": "0x0242",
			"upload_port.5.vid": "0x2341",
		}),
	}
	require.True(t, brd03.IsBoardMatchingIDProperties(properties.NewFromHashmap(map[string]string{
		"pid": "0x0242",
		"vid": "0x2341",
	})))
	require.True(t, brd03.IsBoardMatchingIDProperties(properties.NewFromHashmap(map[string]string{
		"pid":    "0x0242",
		"vid":    "0x2341",
		"serial": "897439287289347",
	})))

	// Mixed indentificiations (not-permitted)
	brd04 := &Board{
		Properties: properties.NewFromHashmap(map[string]string{
			"upload_port.pid":   "0x2222",
			"upload_port.vid":   "0x3333",
			"upload_port.0.pid": "0x0010",
			"upload_port.0.vid": "0x2341",
			"upload_port.1.pid": "0x0042",
			"upload_port.1.vid": "0x2341",
			"upload_port.2.pid": "0x0010",
			"upload_port.2.vid": "0x2A03",
			"upload_port.3.pid": "0x0042",
			"upload_port.3.vid": "0x2A03",
		}),
	}
	require.True(t, brd04.IsBoardMatchingIDProperties(properties.NewFromHashmap(map[string]string{
		"pid": "0x0042",
		"vid": "0x2341",
	})))
	require.True(t, brd04.IsBoardMatchingIDProperties(properties.NewFromHashmap(map[string]string{
		"pid":    "0x0042",
		"vid":    "0x2341",
		"serial": "897439287289347",
	})))
	require.False(t, brd04.IsBoardMatchingIDProperties(properties.NewFromHashmap(map[string]string{
		"pid": "0x2222",
		"vid": "0x3333",
	})))
	require.False(t, brd04.IsBoardMatchingIDProperties(properties.NewFromHashmap(map[string]string{
		"pid":    "0x2222",
		"vid":    "0x3333",
		"serial": "897439287289347",
	})))

	// Mixed protocols
	brd05 := &Board{
		Properties: properties.NewFromHashmap(map[string]string{
			"upload_port.0.pid":    "0x0010",
			"upload_port.0.vid":    "0x2341",
			"upload_port.1.pears":  "2",
			"upload_port.1.apples": "3",
			"upload_port.1.lemons": "X",
			"upload_port.2.pears":  "100",
			"upload_port.3.mac":    "0x0010",
			"upload_port.3.vid":    "0x2341",
		}),
	}
	require.True(t, brd05.IsBoardMatchingIDProperties(properties.NewFromHashmap(map[string]string{
		"pid": "0x0010",
		"vid": "0x2341",
	})))
	require.True(t, brd05.IsBoardMatchingIDProperties(properties.NewFromHashmap(map[string]string{
		"pears":  "2",
		"apples": "3",
		"lemons": "X",
	})))
	require.True(t, brd05.IsBoardMatchingIDProperties(properties.NewFromHashmap(map[string]string{
		"pears": "100",
	})))
	require.True(t, brd05.IsBoardMatchingIDProperties(properties.NewFromHashmap(map[string]string{
		"mac": "0x0010",
		"vid": "0x2341",
	})))
	require.False(t, brd05.IsBoardMatchingIDProperties(properties.NewFromHashmap(map[string]string{
		"pears": "2",
	})))
	require.True(t, brd05.IsBoardMatchingIDProperties(properties.NewFromHashmap(map[string]string{
		"pears":  "100",
		"apples": "300",
		"lemons": "XXX",
	})))
}

func TestBoardConfigMatching(t *testing.T) {
	brd01 := &Board{
		Properties: properties.NewFromHashmap(map[string]string{
			"upload_port.pid":                     "0x0010",
			"upload_port.vid":                     "0x2341",
			"menu.cpu.atmega1280":                 "ATmega1280",
			"menu.cpu.atmega1280.upload_port.cpu": "atmega1280",
			"menu.cpu.atmega1280.build_cpu":       "atmega1280",
			"menu.cpu.atmega2560":                 "ATmega2560",
			"menu.cpu.atmega2560.upload_port.cpu": "atmega2560",
			"menu.cpu.atmega2560.build_cpu":       "atmega2560",
			"menu.mem.1k":                         "1KB",
			"menu.mem.1k.upload_port.mem":         "1",
			"menu.mem.1k.build_mem":               "1024",
			"menu.mem.2k":                         "2KB",
			"menu.mem.2k.upload_port.1.mem":       "2",
			"menu.mem.2k.upload_port.2.ab":        "ef",
			"menu.mem.2k.upload_port.2.cd":        "gh",
			"menu.mem.2k.build_mem":               "2048",
		}),
		PlatformRelease: &PlatformRelease{
			Platform: &Platform{
				Architecture: "avr",
				Package: &Package{
					Name: "arduino",
				},
			},
			Menus: properties.NewFromHashmap(map[string]string{
				"cpu": "Processor",
				"mem": "Memory",
			}),
		},
	}

	type m map[string]string
	type Test struct {
		testName            string
		identificationProps map[string]string
		configOutput        map[string]string
	}

	tests := []Test{
		{"Simple",
			m{"pid": "0x0010", "vid": "0x2341"},
			m{}},
		{"WithConfig1",
			m{"pid": "0x0010", "vid": "0x2341", "cpu": "atmega2560"},
			m{"cpu": "atmega2560"}},
		{"WithConfig2",
			m{"pid": "0x0010", "vid": "0x2341", "cpu": "atmega1280"},
			m{"cpu": "atmega1280"}},
		{"WithDoubleConfig1",
			m{"pid": "0x0010", "vid": "0x2341", "cpu": "atmega1280", "mem": "1"},
			m{"cpu": "atmega1280", "mem": "1k"}},
		{"WithDoubleConfig2",
			m{"pid": "0x0010", "vid": "0x2341", "cpu": "atmega1280", "ab": "ef"},
			m{"cpu": "atmega1280"}},
		{"WithDoubleConfig3",
			m{"pid": "0x0010", "vid": "0x2341", "cpu": "atmega1280", "ab": "ef", "cd": "gh"},
			m{"cpu": "atmega1280", "mem": "2k"}},
		{"WithIncompleteIdentificationProps",
			m{"cpu": "atmega1280"},
			nil},
	}
	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			identificationProps := properties.NewFromHashmap(test.identificationProps)
			if test.configOutput != nil {
				require.True(t, brd01.IsBoardMatchingIDProperties(identificationProps))
				config := brd01.IdentifyBoardConfiguration(identificationProps)
				require.EqualValues(t, test.configOutput, config.AsMap())
			} else {
				require.False(t, brd01.IsBoardMatchingIDProperties(identificationProps))
			}
		})
	}
}
