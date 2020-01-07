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

var boardUno = &Board{
	BoardID: "uno",
	Properties: properties.NewFromHashmap(map[string]string{
		"name":                      "Arduino/Genuino Uno",
		"vid.0":                     "0x2341",
		"pid.0":                     "0x0043",
		"vid.1":                     "0x2341",
		"pid.1":                     "0x0001",
		"vid.2":                     "0x2A03",
		"pid.2":                     "0x0043",
		"vid.3":                     "0x2341",
		"pid.3":                     "0x0243",
		"upload.tool":               "avrdude",
		"upload.protocol":           "arduino",
		"upload.maximum_size":       "32256",
		"upload.maximum_data_size":  "2048",
		"upload.speed":              "115200",
		"bootloader.tool":           "avrdude",
		"bootloader.low_fuses":      "0xFF",
		"bootloader.high_fuses":     "0xDE",
		"bootloader.extended_fuses": "0xFD",
		"bootloader.unlock_bits":    "0x3F",
		"bootloader.lock_bits":      "0x0F",
		"bootloader.file":           "optiboot/optiboot_atmega328.hex",
		"build.mcu":                 "atmega328p",
		"build.f_cpu":               "16000000L",
		"build.board":               "AVR_UNO",
		"build.core":                "arduino",
		"build.variant":             "standard",
	}),
	PlatformRelease: &PlatformRelease{
		Platform: &Platform{
			Architecture: "avr",
			Package: &Package{
				Name: "arduino",
			},
		},
	},
}

var boardMega = &Board{
	BoardID: "mega",
	Properties: properties.NewFromHashmap(map[string]string{
		"name":                                "Arduino/Genuino Mega or Mega 2560",
		"vid.0":                               "0x2341",
		"pid.0":                               "0x0010",
		"vid.1":                               "0x2341",
		"pid.1":                               "0x0042",
		"vid.2":                               "0x2A03",
		"pid.2":                               "0x0010",
		"vid.3":                               "0x2A03",
		"pid.3":                               "0x0042",
		"vid.4":                               "0x2341",
		"pid.4":                               "0x0210",
		"vid.5":                               "0x2341",
		"pid.5":                               "0x0242",
		"upload.tool":                         "avrdude",
		"upload.maximum_data_size":            "8192",
		"bootloader.tool":                     "avrdude",
		"bootloader.low_fuses":                "0xFF",
		"bootloader.unlock_bits":              "0x3F",
		"bootloader.lock_bits":                "0x0F",
		"build.f_cpu":                         "16000000L",
		"build.core":                          "arduino",
		"build.variant":                       "mega",
		"build.board":                         "AVR_MEGA2560",
		"menu.cpu.atmega2560":                 "ATmega2560 (Mega 2560)",
		"menu.cpu.atmega2560.upload.protocol": "wiring",
		"menu.cpu.atmega2560.upload.maximum_size":       "253952",
		"menu.cpu.atmega2560.upload.speed":              "115200",
		"menu.cpu.atmega2560.bootloader.high_fuses":     "0xD8",
		"menu.cpu.atmega2560.bootloader.extended_fuses": "0xFD",
		"menu.cpu.atmega2560.bootloader.file":           "stk500v2/stk500boot_v2_mega2560.hex",
		"menu.cpu.atmega2560.build.mcu":                 "atmega2560",
		"menu.cpu.atmega2560.build.board":               "AVR_MEGA2560",
		"menu.cpu.atmega1280":                           "ATmega1280",
		"menu.cpu.atmega1280.upload.protocol":           "arduino",
		"menu.cpu.atmega1280.upload.maximum_size":       "126976",
		"menu.cpu.atmega1280.upload.speed":              "57600",
		"menu.cpu.atmega1280.bootloader.high_fuses":     "0xDA",
		"menu.cpu.atmega1280.bootloader.extended_fuses": "0xF5",
		"menu.cpu.atmega1280.bootloader.file":           "atmega/ATmegaBOOT_168_atmega1280.hex",
		"menu.cpu.atmega1280.build.mcu":                 "atmega1280",
		"menu.cpu.atmega1280.build.board":               "AVR_MEGA",
	}),
	PlatformRelease: &PlatformRelease{
		Platform: &Platform{
			Architecture: "avr",
			Package: &Package{
				Name: "arduino",
			},
		},
	},
}

var boardWatterottTiny841 = &Board{
	BoardID: "attiny841",
	Properties: properties.NewFromHashmap(map[string]string{
		"name":                                "ATtiny841 (8 MHz)",
		"menu.core.arduino":                   "Standard Arduino",
		"menu.core.arduino.build.core":        "arduino:arduino",
		"menu.core.arduino.build.variant":     "tiny14",
		"menu.core.spencekonde":               "ATtiny841 (by Spence Konde)",
		"menu.core.spencekonde.build.core":    "tiny841",
		"menu.core.spencekonde.build.variant": "tiny14",
		"menu.info.info":                      "Press Reset, when Uploading is shown.",
		"vid.0":                               "0x16D0",
		"pid.0":                               "0x0753",
		"bootloader.tool":                     "avrdude",
		"bootloader.low_fuses":                "0xE2",
		"bootloader.high_fuses":               "0xDD",
		"bootloader.extended_fuses":           "0xFE",
		"bootloader.unlock_bits":              "0xFF",
		"bootloader.lock_bits":                "0xFF",
		"bootloader.file":                     "micronucleus-t841.hex",
		"upload.tool":                         "micronucleus",
		"upload.protocol":                     "usb",
		"upload.wait_for_upload_port":         "false",
		"upload.use_1200bps_touch":            "false",
		"upload.disable_flushing":             "false",
		"upload.maximum_size":                 "6500",
		"build.mcu":                           "attiny841",
		"build.f_cpu":                         "8000000L",
		"build.board":                         "AVR_ATTINY841",
	}),
	PlatformRelease: &PlatformRelease{
		Platform: &Platform{
			Architecture: "avr",
			Package: &Package{
				Name: "watterott",
			},
		},
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
	expConf2560 := properties.NewFromHashmap(map[string]string{
		"bootloader.extended_fuses": "0xFD",
		"bootloader.file":           "stk500v2/stk500boot_v2_mega2560.hex",
		"bootloader.high_fuses":     "0xD8",
		"bootloader.lock_bits":      "0x0F",
		"bootloader.low_fuses":      "0xFF",
		"bootloader.tool":           "avrdude",
		"bootloader.unlock_bits":    "0x3F",
		"build.board":               "AVR_MEGA2560",
		"build.core":                "arduino",
		"build.f_cpu":               "16000000L",
		"build.mcu":                 "atmega2560",
		"build.variant":             "mega",
		"menu.cpu.atmega1280":       "ATmega1280",
		"menu.cpu.atmega1280.bootloader.extended_fuses": "0xF5",
		"menu.cpu.atmega1280.bootloader.file":           "atmega/ATmegaBOOT_168_atmega1280.hex",
		"menu.cpu.atmega1280.bootloader.high_fuses":     "0xDA",
		"menu.cpu.atmega1280.build.board":               "AVR_MEGA",
		"menu.cpu.atmega1280.build.mcu":                 "atmega1280",
		"menu.cpu.atmega1280.upload.maximum_size":       "126976",
		"menu.cpu.atmega1280.upload.protocol":           "arduino",
		"menu.cpu.atmega1280.upload.speed":              "57600",
		"menu.cpu.atmega2560":                           "ATmega2560 (Mega 2560)",
		"menu.cpu.atmega2560.bootloader.extended_fuses": "0xFD",
		"menu.cpu.atmega2560.bootloader.file":           "stk500v2/stk500boot_v2_mega2560.hex",
		"menu.cpu.atmega2560.bootloader.high_fuses":     "0xD8",
		"menu.cpu.atmega2560.build.board":               "AVR_MEGA2560",
		"menu.cpu.atmega2560.build.mcu":                 "atmega2560",
		"menu.cpu.atmega2560.upload.maximum_size":       "253952",
		"menu.cpu.atmega2560.upload.protocol":           "wiring",
		"menu.cpu.atmega2560.upload.speed":              "115200",
		"name":                                          "Arduino/Genuino Mega or Mega 2560",
		"pid.0":                                         "0x0010",
		"pid.1":                                         "0x0042",
		"pid.2":                                         "0x0010",
		"pid.3":                                         "0x0042",
		"pid.4":                                         "0x0210",
		"pid.5":                                         "0x0242",
		"upload.maximum_data_size":                      "8192",
		"upload.maximum_size":                           "253952",
		"upload.protocol":                               "wiring",
		"upload.speed":                                  "115200",
		"upload.tool":                                   "avrdude",
		"vid.0":                                         "0x2341",
		"vid.1":                                         "0x2341",
		"vid.2":                                         "0x2A03",
		"vid.3":                                         "0x2A03",
		"vid.4":                                         "0x2341",
		"vid.5":                                         "0x2341",
	})

	conf2560, err := boardMega.GeneratePropertiesForConfiguration("cpu=atmega2560")
	require.NoError(t, err, "generating cpu=atmega2560 configuration")
	require.EqualValues(t, expConf2560.AsMap(), conf2560.AsMap(), "configuration for cpu=atmega2560")

	expConf1280 := properties.NewFromHashmap(map[string]string{
		"bootloader.extended_fuses": "0xF5",
		"bootloader.file":           "atmega/ATmegaBOOT_168_atmega1280.hex",
		"bootloader.high_fuses":     "0xDA",
		"bootloader.lock_bits":      "0x0F",
		"bootloader.low_fuses":      "0xFF",
		"bootloader.tool":           "avrdude",
		"bootloader.unlock_bits":    "0x3F",
		"build.board":               "AVR_MEGA",
		"build.core":                "arduino",
		"build.f_cpu":               "16000000L",
		"build.mcu":                 "atmega1280",
		"build.variant":             "mega",
		"menu.cpu.atmega1280":       "ATmega1280",
		"menu.cpu.atmega1280.bootloader.extended_fuses": "0xF5",
		"menu.cpu.atmega1280.bootloader.file":           "atmega/ATmegaBOOT_168_atmega1280.hex",
		"menu.cpu.atmega1280.bootloader.high_fuses":     "0xDA",
		"menu.cpu.atmega1280.build.board":               "AVR_MEGA",
		"menu.cpu.atmega1280.build.mcu":                 "atmega1280",
		"menu.cpu.atmega1280.upload.maximum_size":       "126976",
		"menu.cpu.atmega1280.upload.protocol":           "arduino",
		"menu.cpu.atmega1280.upload.speed":              "57600",
		"menu.cpu.atmega2560":                           "ATmega2560 (Mega 2560)",
		"menu.cpu.atmega2560.bootloader.extended_fuses": "0xFD",
		"menu.cpu.atmega2560.bootloader.file":           "stk500v2/stk500boot_v2_mega2560.hex",
		"menu.cpu.atmega2560.bootloader.high_fuses":     "0xD8",
		"menu.cpu.atmega2560.build.board":               "AVR_MEGA2560",
		"menu.cpu.atmega2560.build.mcu":                 "atmega2560",
		"menu.cpu.atmega2560.upload.maximum_size":       "253952",
		"menu.cpu.atmega2560.upload.protocol":           "wiring",
		"menu.cpu.atmega2560.upload.speed":              "115200",
		"name":                                          "Arduino/Genuino Mega or Mega 2560",
		"pid.0":                                         "0x0010",
		"pid.1":                                         "0x0042",
		"pid.2":                                         "0x0010",
		"pid.3":                                         "0x0042",
		"pid.4":                                         "0x0210",
		"pid.5":                                         "0x0242",
		"upload.maximum_data_size":                      "8192",
		"upload.maximum_size":                           "126976",
		"upload.protocol":                               "arduino",
		"upload.speed":                                  "57600",
		"upload.tool":                                   "avrdude",
		"vid.0":                                         "0x2341",
		"vid.1":                                         "0x2341",
		"vid.2":                                         "0x2A03",
		"vid.3":                                         "0x2A03",
		"vid.4":                                         "0x2341",
		"vid.5":                                         "0x2341",
	})
	conf1280, err := boardMega.GeneratePropertiesForConfiguration("cpu=atmega1280")
	require.NoError(t, err, "generating cpu=atmega1280 configuration")
	require.EqualValues(t, expConf1280.AsMap(), conf1280.AsMap(), "configuration for cpu=atmega1280")

	_, err = boardMega.GeneratePropertiesForConfiguration("cpu=atmegassss")
	require.Error(t, err, "generating cpu=atmegassss configuration")

	_, err = boardUno.GeneratePropertiesForConfiguration("cpu=atmega1280")
	require.Error(t, err, "generating cpu=atmega1280 configuration")

	expWatterott := properties.NewFromHashmap(map[string]string{
		"bootloader.extended_fuses":           "0xFE",
		"bootloader.file":                     "micronucleus-t841.hex",
		"bootloader.high_fuses":               "0xDD",
		"bootloader.lock_bits":                "0xFF",
		"bootloader.low_fuses":                "0xE2",
		"bootloader.tool":                     "avrdude",
		"bootloader.unlock_bits":              "0xFF",
		"build.board":                         "AVR_ATTINY841",
		"build.core":                          "tiny841",
		"build.f_cpu":                         "8000000L",
		"build.mcu":                           "attiny841",
		"build.variant":                       "tiny14",
		"menu.core.arduino":                   "Standard Arduino",
		"menu.core.arduino.build.core":        "arduino:arduino",
		"menu.core.arduino.build.variant":     "tiny14",
		"menu.core.spencekonde":               "ATtiny841 (by Spence Konde)",
		"menu.core.spencekonde.build.core":    "tiny841",
		"menu.core.spencekonde.build.variant": "tiny14",
		"menu.info.info":                      "Press Reset, when Uploading is shown.",
		"name":                                "ATtiny841 (8 MHz)",
		"pid.0":                               "0x0753",
		"upload.disable_flushing":             "false",
		"upload.maximum_size":                 "6500",
		"upload.protocol":                     "usb",
		"upload.tool":                         "micronucleus",
		"upload.use_1200bps_touch":            "false",
		"upload.wait_for_upload_port":         "false",
		"vid.0":                               "0x16D0",
	})
	confWatterott, err := boardWatterottTiny841.GeneratePropertiesForConfiguration("core=spencekonde,info=info")
	require.NoError(t, err, "generating core=spencekonde,info=info configuration")
	require.EqualValues(t, expWatterott.AsMap(), confWatterott.AsMap(), "generating core=spencekonde,info=info configuration")

	// data, err := json.MarshalIndent(prop, "", "  ")
	// require.NoError(t, err, "marshaling result")
	// fmt.Print(string(data))
}
