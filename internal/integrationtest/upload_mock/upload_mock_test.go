// This file is part of arduino-cli.
//
// Copyright 2022 ARDUINO SA (http://www.arduino.cc/)
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

package upload_mock_test

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"runtime"
	"strings"
	"testing"

	"github.com/arduino/arduino-cli/internal/integrationtest"
	"github.com/arduino/go-paths-helper"
	"github.com/stretchr/testify/require"
)

type parameters struct {
	Fqbn       string
	UploadPort string
	Programmer string
	Output     string
}

type parametersMap struct {
	Fqbn       string
	UploadPort string
	Programmer string
	Output     map[string]string
}

func TestUploadSketch(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	t.Cleanup(env.CleanUp)

	indexes := []string{
		"https://github.com/stm32duino/BoardManagerFiles/raw/main/package_stmicroelectronics_index.json",
		"https://adafruit.github.io/arduino-board-index/package_adafruit_index.json",
		"https://raw.githubusercontent.com/espressif/arduino-esp32/gh-pages/package_esp32_index.json",
		"http://arduino.esp8266.com/stable/package_esp8266com_index.json",
		"https://github.com/sonydevworld/spresense-arduino-compatible/releases/download/generic/package_spresense_index.json",
	}

	coresToInstall := []string{
		"STMicroelectronics:stm32@2.2.0",
		"arduino:avr@1.8.3",
		"adafruit:avr@1.4.13",
		"arduino:samd@1.8.11",
		"esp32:esp32@1.0.6",
		"esp8266:esp8266@3.0.2",
		"SPRESENSE:spresense@2.0.2",
	}

	testParameters := []parameters{
		{
			Fqbn:       "arduino:avr:uno",
			UploadPort: "/dev/ttyACM0",
			Programmer: "",
			Output:     "\"{data_dir}/packages/arduino/tools/avrdude/6.3.0-arduino17/bin/avrdude\" \"-C{data_dir}/packages/arduino/tools/avrdude/6.3.0-arduino17/etc/avrdude.conf\" -v -V -patmega328p -carduino \"-P/dev/ttyACM0\" -b115200 -D \"-Uflash:w:{build_dir}/{sketch_name}.ino.hex:i\"\n",
		},
		{
			Fqbn:       "arduino:avr:uno",
			UploadPort: "",
			Programmer: "usbasp",
			Output:     "\"{data_dir}/packages/arduino/tools/avrdude/6.3.0-arduino17/bin/avrdude\" \"-C{data_dir}/packages/arduino/tools/avrdude/6.3.0-arduino17/etc/avrdude.conf\" -v -V -patmega328p -cusbasp -Pusb \"-Uflash:w:{build_dir}/{sketch_name}.ino.hex:i\"\n",
		},
		{
			Fqbn:       "arduino:avr:uno",
			UploadPort: "/dev/ttyACM0",
			Programmer: "avrisp",
			Output:     "\"{data_dir}/packages/arduino/tools/avrdude/6.3.0-arduino17/bin/avrdude\" \"-C{data_dir}/packages/arduino/tools/avrdude/6.3.0-arduino17/etc/avrdude.conf\" -v -V -patmega328p -cstk500v1 -P/dev/ttyACM0 \"-Uflash:w:{build_dir}/{sketch_name}.ino.hex:i\"\n",
		},
		{
			Fqbn:       "arduino:avr:uno",
			UploadPort: "/dev/ttyACM0",
			Programmer: "arduinoasisp",
			Output:     "\"{data_dir}/packages/arduino/tools/avrdude/6.3.0-arduino17/bin/avrdude\" \"-C{data_dir}/packages/arduino/tools/avrdude/6.3.0-arduino17/etc/avrdude.conf\" -v -V -patmega328p -cstk500v1 -P/dev/ttyACM0 -b19200 \"-Uflash:w:{build_dir}/{sketch_name}.ino.hex:i\"\n",
		},
		{
			Fqbn:       "arduino:avr:nano",
			UploadPort: "/dev/ttyACM0",
			Programmer: "",
			Output:     "\"{data_dir}/packages/arduino/tools/avrdude/6.3.0-arduino17/bin/avrdude\" \"-C{data_dir}/packages/arduino/tools/avrdude/6.3.0-arduino17/etc/avrdude.conf\" -v -V -patmega328p -carduino \"-P/dev/ttyACM0\" -b115200 -D \"-Uflash:w:{build_dir}/{sketch_name}.ino.hex:i\"\n",
		},
		{
			Fqbn:       "arduino:avr:nano",
			UploadPort: "",
			Programmer: "usbasp",
			Output:     "\"{data_dir}/packages/arduino/tools/avrdude/6.3.0-arduino17/bin/avrdude\" \"-C{data_dir}/packages/arduino/tools/avrdude/6.3.0-arduino17/etc/avrdude.conf\" -v -V -patmega328p -cusbasp -Pusb \"-Uflash:w:{build_dir}/{sketch_name}.ino.hex:i\"\n",
		},
		{
			Fqbn:       "arduino:avr:nano",
			UploadPort: "/dev/ttyACM0",
			Programmer: "avrisp",
			Output:     "\"{data_dir}/packages/arduino/tools/avrdude/6.3.0-arduino17/bin/avrdude\" \"-C{data_dir}/packages/arduino/tools/avrdude/6.3.0-arduino17/etc/avrdude.conf\" -v -V -patmega328p -cstk500v1 -P/dev/ttyACM0 \"-Uflash:w:{build_dir}/{sketch_name}.ino.hex:i\"\n",
		},
		{
			Fqbn:       "arduino:avr:nano",
			UploadPort: "/dev/ttyACM0",
			Programmer: "arduinoasisp",
			Output:     "\"{data_dir}/packages/arduino/tools/avrdude/6.3.0-arduino17/bin/avrdude\" \"-C{data_dir}/packages/arduino/tools/avrdude/6.3.0-arduino17/etc/avrdude.conf\" -v -V -patmega328p -cstk500v1 -P/dev/ttyACM0 -b19200 \"-Uflash:w:{build_dir}/{sketch_name}.ino.hex:i\"\n",
		},
		{
			Fqbn:       "arduino:avr:nano:cpu=atmega328old",
			UploadPort: "/dev/ttyACM0",
			Programmer: "",
			Output:     "\"{data_dir}/packages/arduino/tools/avrdude/6.3.0-arduino17/bin/avrdude\" \"-C{data_dir}/packages/arduino/tools/avrdude/6.3.0-arduino17/etc/avrdude.conf\" -v -V -patmega328p -carduino \"-P/dev/ttyACM0\" -b57600 -D \"-Uflash:w:{build_dir}/{sketch_name}.ino.hex:i\"\n",
		},
		{
			Fqbn:       "arduino:avr:nano:cpu=atmega328old",
			UploadPort: "",
			Programmer: "usbasp",
			Output:     "\"{data_dir}/packages/arduino/tools/avrdude/6.3.0-arduino17/bin/avrdude\" \"-C{data_dir}/packages/arduino/tools/avrdude/6.3.0-arduino17/etc/avrdude.conf\" -v -V -patmega328p -cusbasp -Pusb \"-Uflash:w:{build_dir}/{sketch_name}.ino.hex:i\"\n",
		},
		{
			Fqbn:       "arduino:avr:nano:cpu=atmega328old",
			UploadPort: "/dev/ttyACM0",
			Programmer: "avrisp",
			Output:     "\"{data_dir}/packages/arduino/tools/avrdude/6.3.0-arduino17/bin/avrdude\" \"-C{data_dir}/packages/arduino/tools/avrdude/6.3.0-arduino17/etc/avrdude.conf\" -v -V -patmega328p -cstk500v1 -P/dev/ttyACM0 \"-Uflash:w:{build_dir}/{sketch_name}.ino.hex:i\"\n",
		},
		{
			Fqbn:       "arduino:avr:nano:cpu=atmega328old",
			UploadPort: "/dev/ttyACM0",
			Programmer: "arduinoasisp",
			Output:     "\"{data_dir}/packages/arduino/tools/avrdude/6.3.0-arduino17/bin/avrdude\" \"-C{data_dir}/packages/arduino/tools/avrdude/6.3.0-arduino17/etc/avrdude.conf\" -v -V -patmega328p -cstk500v1 -P/dev/ttyACM0 -b19200 \"-Uflash:w:{build_dir}/{sketch_name}.ino.hex:i\"\n",
		},
		{
			Fqbn:       "arduino:avr:mega",
			UploadPort: "/dev/ttyACM0",
			Programmer: "",
			Output:     "\"{data_dir}/packages/arduino/tools/avrdude/6.3.0-arduino17/bin/avrdude\" \"-C{data_dir}/packages/arduino/tools/avrdude/6.3.0-arduino17/etc/avrdude.conf\" -v -V -patmega2560 -cwiring \"-P/dev/ttyACM0\" -b115200 -D \"-Uflash:w:{build_dir}/{sketch_name}.ino.hex:i\"\n",
		},
		{
			Fqbn:       "arduino:avr:mega:cpu=atmega1280",
			UploadPort: "/dev/ttyACM0",
			Programmer: "",
			Output:     "\"{data_dir}/packages/arduino/tools/avrdude/6.3.0-arduino17/bin/avrdude\" \"-C{data_dir}/packages/arduino/tools/avrdude/6.3.0-arduino17/etc/avrdude.conf\" -v -V -patmega1280 -carduino \"-P/dev/ttyACM0\" -b57600 -D \"-Uflash:w:{build_dir}/{sketch_name}.ino.hex:i\"\n",
		},
		{
			Fqbn:       "arduino:avr:diecimila",
			UploadPort: "/dev/ttyACM0",
			Programmer: "",
			Output:     "\"{data_dir}/packages/arduino/tools/avrdude/6.3.0-arduino17/bin/avrdude\" \"-C{data_dir}/packages/arduino/tools/avrdude/6.3.0-arduino17/etc/avrdude.conf\" -v -V -patmega328p -carduino \"-P/dev/ttyACM0\" -b57600 -D \"-Uflash:w:{build_dir}/{sketch_name}.ino.hex:i\"\n",
		},
		{
			Fqbn:       "arduino:avr:leonardo",
			UploadPort: "/dev/ttyACM999",
			Programmer: "",
			Output:     "Performing 1200-bps touch reset on serial port /dev/ttyACM999\nWaiting for upload port...\nUpload port found on /dev/ttyACM9990\n\"{data_dir}/packages/arduino/tools/avrdude/6.3.0-arduino17/bin/avrdude\" \"-C{data_dir}/packages/arduino/tools/avrdude/6.3.0-arduino17/etc/avrdude.conf\" -v -V -patmega32u4 -cavr109 \"-P/dev/ttyACM9990\" -b57600 -D \"-Uflash:w:{build_dir}/{sketch_name}.ino.hex:i\"\n",
		},
		{
			Fqbn:       "arduino:avr:leonardo",
			UploadPort: "",
			Programmer: "",
			Output:     "Skipping 1200-bps touch reset: no serial port selected!\nWaiting for upload port...\nUpload port found on newport\n\"{data_dir}/packages/arduino/tools/avrdude/6.3.0-arduino17/bin/avrdude\" \"-C{data_dir}/packages/arduino/tools/avrdude/6.3.0-arduino17/etc/avrdude.conf\" -v -V -patmega32u4 -cavr109 \"-Pnewport\" -b57600 -D \"-Uflash:w:{build_dir}/{sketch_name}.ino.hex:i\"\n",
		},
		{
			Fqbn:       "arduino:avr:micro",
			UploadPort: "/dev/ttyACM999",
			Programmer: "",
			Output:     "Performing 1200-bps touch reset on serial port /dev/ttyACM999\nWaiting for upload port...\nUpload port found on /dev/ttyACM9990\n\"{data_dir}/packages/arduino/tools/avrdude/6.3.0-arduino17/bin/avrdude\" \"-C{data_dir}/packages/arduino/tools/avrdude/6.3.0-arduino17/etc/avrdude.conf\" -v -V -patmega32u4 -cavr109 \"-P/dev/ttyACM9990\" -b57600 -D \"-Uflash:w:{build_dir}/{sketch_name}.ino.hex:i\"\n",
		},
		{
			Fqbn:       "arduino:avr:micro",
			UploadPort: "",
			Programmer: "",
			Output:     "Skipping 1200-bps touch reset: no serial port selected!\nWaiting for upload port...\nUpload port found on newport\n\"{data_dir}/packages/arduino/tools/avrdude/6.3.0-arduino17/bin/avrdude\" \"-C{data_dir}/packages/arduino/tools/avrdude/6.3.0-arduino17/etc/avrdude.conf\" -v -V -patmega32u4 -cavr109 \"-Pnewport\" -b57600 -D \"-Uflash:w:{build_dir}/{sketch_name}.ino.hex:i\"\n",
		},
		{
			Fqbn:       "arduino:avr:circuitplay32u4cat",
			UploadPort: "/dev/ttyACM999",
			Programmer: "",
			Output:     "Performing 1200-bps touch reset on serial port /dev/ttyACM999\nWaiting for upload port...\nUpload port found on /dev/ttyACM9990\n\"{data_dir}/packages/arduino/tools/avrdude/6.3.0-arduino17/bin/avrdude\" \"-C{data_dir}/packages/arduino/tools/avrdude/6.3.0-arduino17/etc/avrdude.conf\" -v -V -patmega32u4 -cavr109 \"-P/dev/ttyACM9990\" -b57600 -D \"-Uflash:w:{build_dir}/{sketch_name}.ino.hex:i\"\n",
		},
		{
			Fqbn:       "arduino:avr:gemma",
			UploadPort: "/dev/ttyACM0",
			Programmer: "usbGemma",
			Output:     "\"{data_dir}/packages/arduino/tools/avrdude/6.3.0-arduino17/bin/avrdude\" \"-C{data_dir}/packages/arduino/hardware/avr/1.8.3/bootloaders/gemma/avrdude.conf\" -v -V -pattiny85 -carduinogemma  \"-Uflash:w:{build_dir}/{sketch_name}.ino.hex:i\"\n",
		},
		{
			Fqbn:       "arduino:avr:gemma",
			UploadPort: "",
			Programmer: "usbGemma",
			Output:     "\"{data_dir}/packages/arduino/tools/avrdude/6.3.0-arduino17/bin/avrdude\" \"-C{data_dir}/packages/arduino/hardware/avr/1.8.3/bootloaders/gemma/avrdude.conf\" -v -V -pattiny85 -carduinogemma  \"-Uflash:w:{build_dir}/{sketch_name}.ino.hex:i\"\n",
		},
		{
			Fqbn:       "arduino:avr:unowifi",
			UploadPort: "/dev/ttyACM0",
			Programmer: "",
			Output:     "\"{data_dir}/packages/arduino/tools/avrdude/6.3.0-arduino17/bin/avrdude\" \"-C{data_dir}/packages/arduino/tools/avrdude/6.3.0-arduino17/etc/avrdude.conf\" -v -V -patmega328p -carduino \"-P/dev/ttyACM0\" -b115200 -D \"-Uflash:w:{build_dir}/{sketch_name}.ino.hex:i\"\n",
		},
		{
			Fqbn:       "arduino:avr:yun",
			UploadPort: "/dev/ttyACM999",
			Programmer: "",
			Output:     "Performing 1200-bps touch reset on serial port /dev/ttyACM999\nWaiting for upload port...\nUpload port found on /dev/ttyACM9990\n\"{data_dir}/packages/arduino/tools/avrdude/6.3.0-arduino17/bin/avrdude\" \"-C{data_dir}/packages/arduino/tools/avrdude/6.3.0-arduino17/etc/avrdude.conf\" -v -V -patmega32u4 -cavr109 \"-P/dev/ttyACM9990\" -b57600 -D \"-Uflash:w:{build_dir}/{sketch_name}.ino.hex:i\"\n",
		},
		{
			Fqbn:       "adafruit:avr:circuitplay32u4cat",
			UploadPort: "/dev/ttyACM999",
			Programmer: "",
			Output:     "Performing 1200-bps touch reset on serial port /dev/ttyACM999\nWaiting for upload port...\nUpload port found on /dev/ttyACM9990\n\"{data_dir}/packages/arduino/tools/avrdude/6.3.0-arduino17/bin/avrdude\" \"-C{data_dir}/packages/arduino/tools/avrdude/6.3.0-arduino17/etc/avrdude.conf\" -v -patmega32u4 -cavr109 -P/dev/ttyACM9990 -b57600 -D \"-Uflash:w:{build_dir}/{sketch_name}.ino.hex:i\"\n",
		},
		{
			Fqbn:       "adafruit:avr:flora8",
			UploadPort: "/dev/ttyACM999",
			Programmer: "",
			Output:     "Performing 1200-bps touch reset on serial port /dev/ttyACM999\nWaiting for upload port...\nUpload port found on /dev/ttyACM9990\n\"{data_dir}/packages/arduino/tools/avrdude/6.3.0-arduino17/bin/avrdude\" \"-C{data_dir}/packages/arduino/tools/avrdude/6.3.0-arduino17/etc/avrdude.conf\" -v -patmega32u4 -cavr109 -P/dev/ttyACM9990 -b57600 -D \"-Uflash:w:{build_dir}/{sketch_name}.ino.hex:i\"\n",
		},
		{
			Fqbn:       "adafruit:avr:gemma",
			UploadPort: "/dev/ttyACM0",
			Programmer: "usbGemma",
			Output:     "\"{data_dir}/packages/arduino/tools/avrdude/6.3.0-arduino17/bin/avrdude\" \"-C{data_dir}/packages/adafruit/hardware/avr/1.4.13/bootloaders/gemma/avrdude.conf\" -v -pattiny85 -carduinogemma  \"-Uflash:w:{build_dir}/{sketch_name}.ino.hex:i\"\n",
		},
		{
			Fqbn:       "adafruit:avr:gemma",
			UploadPort: "",
			Programmer: "usbGemma",
			Output:     "\"{data_dir}/packages/arduino/tools/avrdude/6.3.0-arduino17/bin/avrdude\" \"-C{data_dir}/packages/adafruit/hardware/avr/1.4.13/bootloaders/gemma/avrdude.conf\" -v -pattiny85 -carduinogemma  \"-Uflash:w:{build_dir}/{sketch_name}.ino.hex:i\"\n",
		},
		{
			Fqbn:       "adafruit:avr:itsybitsy32u4_3V",
			UploadPort: "/dev/ttyACM999",
			Programmer: "",
			Output:     "Performing 1200-bps touch reset on serial port /dev/ttyACM999\nWaiting for upload port...\nUpload port found on /dev/ttyACM9990\n\"{data_dir}/packages/arduino/tools/avrdude/6.3.0-arduino17/bin/avrdude\" \"-C{data_dir}/packages/arduino/tools/avrdude/6.3.0-arduino17/etc/avrdude.conf\" -v -patmega32u4 -cavr109 -P/dev/ttyACM9990 -b57600 -D \"-Uflash:w:{build_dir}/{sketch_name}.ino.hex:i\"\n",
		},
		{
			Fqbn:       "adafruit:avr:itsybitsy32u4_5V",
			UploadPort: "/dev/ttyACM999",
			Programmer: "",
			Output:     "Performing 1200-bps touch reset on serial port /dev/ttyACM999\nWaiting for upload port...\nUpload port found on /dev/ttyACM9990\n\"{data_dir}/packages/arduino/tools/avrdude/6.3.0-arduino17/bin/avrdude\" \"-C{data_dir}/packages/arduino/tools/avrdude/6.3.0-arduino17/etc/avrdude.conf\" -v -patmega32u4 -cavr109 -P/dev/ttyACM9990 -b57600 -D \"-Uflash:w:{build_dir}/{sketch_name}.ino.hex:i\"\n",
		},
		{
			Fqbn:       "adafruit:avr:metro",
			UploadPort: "/dev/ttyACM0",
			Programmer: "",
			Output:     "\"{data_dir}/packages/arduino/tools/avrdude/6.3.0-arduino17/bin/avrdude\" \"-C{data_dir}/packages/arduino/tools/avrdude/6.3.0-arduino17/etc/avrdude.conf\" -v -patmega328p -carduino -P/dev/ttyACM0 -b115200 -D \"-Uflash:w:{build_dir}/{sketch_name}.ino.hex:i\"\n",
		},
		{
			Fqbn:       "adafruit:avr:trinket3",
			UploadPort: "",
			Programmer: "usbasp",
			Output:     "\"{data_dir}/packages/arduino/tools/avrdude/6.3.0-arduino17/bin/avrdude\" \"-C{data_dir}/packages/arduino/tools/avrdude/6.3.0-arduino17/etc/avrdude.conf\" -v -pattiny85 -cusbasp -Pusb \"-Uflash:w:{build_dir}/{sketch_name}.ino.hex:i\"\n",
		},
		{
			Fqbn:       "adafruit:avr:trinket3",
			UploadPort: "/dev/ttyACM0",
			Programmer: "avrisp",
			Output:     "\"{data_dir}/packages/arduino/tools/avrdude/6.3.0-arduino17/bin/avrdude\" \"-C{data_dir}/packages/arduino/tools/avrdude/6.3.0-arduino17/etc/avrdude.conf\" -v -pattiny85 -cstk500v1 -P/dev/ttyACM0 \"-Uflash:w:{build_dir}/{sketch_name}.ino.hex:i\"\n",
		},
		{
			Fqbn:       "adafruit:avr:trinket3",
			UploadPort: "/dev/ttyACM0",
			Programmer: "arduinoasisp",
			Output:     "\"{data_dir}/packages/arduino/tools/avrdude/6.3.0-arduino17/bin/avrdude\" \"-C{data_dir}/packages/arduino/tools/avrdude/6.3.0-arduino17/etc/avrdude.conf\" -v -pattiny85 -cstk500v1 -P/dev/ttyACM0 -b19200 \"-Uflash:w:{build_dir}/{sketch_name}.ino.hex:i\"\n",
		},
		{
			Fqbn:       "esp8266:esp8266:generic",
			UploadPort: "/dev/ttyACM0",
			Programmer: "",
			Output:     "\"{data_dir}/packages/esp8266/tools/python3/3.7.2-post1/python3\" -I \"{data_dir}/packages/esp8266/hardware/esp8266/3.0.2/tools/upload.py\" --chip esp8266 --port \"/dev/ttyACM0\" --baud \"115200\" \"\"  --before default_reset --after hard_reset write_flash 0x0 \"{build_dir}/{sketch_name}.ino.bin\"\n",
		},
		{
			Fqbn:       "esp8266:esp8266:generic:xtal=160,vt=heap,mmu=3216,ResetMethod=nodtr_nosync,CrystalFreq=40,FlashFreq=20,eesz=2M,baud=57600",
			UploadPort: "/dev/ttyACM0",
			Programmer: "",
			Output:     "\"{data_dir}/packages/esp8266/tools/python3/3.7.2-post1/python3\" -I \"{data_dir}/packages/esp8266/hardware/esp8266/3.0.2/tools/upload.py\" --chip esp8266 --port \"/dev/ttyACM0\" --baud \"57600\" \"\"  --before no_reset_no_sync --after soft_reset write_flash 0x0 \"{build_dir}/{sketch_name}.ino.bin\"\n",
		},
	}

	testParametersMap := []parametersMap{
		{
			Fqbn:       "arduino:avr:leonardo",
			UploadPort: "/dev/ttyACM0",
			Programmer: "",
			Output: map[string]string{
				"linux":  "Performing 1200-bps touch reset on serial port /dev/ttyACM0\nWaiting for upload port...\nNo upload port found, using /dev/ttyACM0 as fallback\n\"{data_dir}/packages/arduino/tools/avrdude/6.3.0-arduino17/bin/avrdude\" \"-C{data_dir}/packages/arduino/tools/avrdude/6.3.0-arduino17/etc/avrdude.conf\" -v -V -patmega32u4 -cavr109 \"-P/dev/ttyACM0\" -b57600 -D \"-Uflash:w:{build_dir}/{sketch_name}.ino.hex:i\"\n",
				"darwin": "Performing 1200-bps touch reset on serial port /dev/ttyACM0\nWaiting for upload port...\nNo upload port found, using /dev/ttyACM0 as fallback\n\"{data_dir}/packages/arduino/tools/avrdude/6.3.0-arduino17/bin/avrdude\" \"-C{data_dir}/packages/arduino/tools/avrdude/6.3.0-arduino17/etc/avrdude.conf\" -v -V -patmega32u4 -cavr109 \"-P/dev/ttyACM0\" -b57600 -D \"-Uflash:w:{build_dir}/{sketch_name}.ino.hex:i\"\n",
				"win32":  "Performing 1200-bps touch reset on serial port /dev/ttyACM0\nWaiting for upload port...\nNo upload port found, using /dev/ttyACM0 as fallback\n\"{data_dir}/packages/arduino/tools/avrdude/6.3.0-arduino17/bin/avrdude\" \"-C{data_dir}/packages/arduino/tools/avrdude/6.3.0-arduino17/etc/avrdude.conf\" -v -V -patmega32u4 -cavr109 \"-P/dev/ttyACM0\" -b57600 -D \"-Uflash:w:{build_dir}/{sketch_name}.ino.hex:i\"\n",
			},
		},
		{
			Fqbn:       "arduino:avr:micro",
			UploadPort: "/dev/ttyACM0",
			Programmer: "",
			Output: map[string]string{
				"linux":  "Performing 1200-bps touch reset on serial port /dev/ttyACM0\nWaiting for upload port...\nNo upload port found, using /dev/ttyACM0 as fallback\n\"{data_dir}/packages/arduino/tools/avrdude/6.3.0-arduino17/bin/avrdude\" \"-C{data_dir}/packages/arduino/tools/avrdude/6.3.0-arduino17/etc/avrdude.conf\" -v -V -patmega32u4 -cavr109 \"-P/dev/ttyACM0\" -b57600 -D \"-Uflash:w:{build_dir}/{sketch_name}.ino.hex:i\"\n",
				"darwin": "Performing 1200-bps touch reset on serial port /dev/ttyACM0\nWaiting for upload port...\nNo upload port found, using /dev/ttyACM0 as fallback\n\"{data_dir}/packages/arduino/tools/avrdude/6.3.0-arduino17/bin/avrdude\" \"-C{data_dir}/packages/arduino/tools/avrdude/6.3.0-arduino17/etc/avrdude.conf\" -v -V -patmega32u4 -cavr109 \"-P/dev/ttyACM0\" -b57600 -D \"-Uflash:w:{build_dir}/{sketch_name}.ino.hex:i\"\n",
				"win32":  "Performing 1200-bps touch reset on serial port /dev/ttyACM0\nWaiting for upload port...\nNo upload port found, using /dev/ttyACM0 as fallback\n\"{data_dir}/packages/arduino/tools/avrdude/6.3.0-arduino17/bin/avrdude\" \"-C{data_dir}/packages/arduino/tools/avrdude/6.3.0-arduino17/etc/avrdude.conf\" -v -V -patmega32u4 -cavr109 \"-P/dev/ttyACM0\" -b57600 -D \"-Uflash:w:{build_dir}/{sketch_name}.ino.hex:i\"\n",
			},
		},
		{
			Fqbn:       "arduino:avr:circuitplay32u4cat",
			UploadPort: "/dev/ttyACM0",
			Programmer: "",
			Output: map[string]string{
				"linux":  "Performing 1200-bps touch reset on serial port /dev/ttyACM0\nWaiting for upload port...\nNo upload port found, using /dev/ttyACM0 as fallback\n\"{data_dir}/packages/arduino/tools/avrdude/6.3.0-arduino17/bin/avrdude\" \"-C{data_dir}/packages/arduino/tools/avrdude/6.3.0-arduino17/etc/avrdude.conf\" -v -V -patmega32u4 -cavr109 \"-P/dev/ttyACM0\" -b57600 -D \"-Uflash:w:{build_dir}/{sketch_name}.ino.hex:i\"\n",
				"darwin": "Performing 1200-bps touch reset on serial port /dev/ttyACM0\nWaiting for upload port...\nNo upload port found, using /dev/ttyACM0 as fallback\n\"{data_dir}/packages/arduino/tools/avrdude/6.3.0-arduino17/bin/avrdude\" \"-C{data_dir}/packages/arduino/tools/avrdude/6.3.0-arduino17/etc/avrdude.conf\" -v -V -patmega32u4 -cavr109 \"-P/dev/ttyACM0\" -b57600 -D \"-Uflash:w:{build_dir}/{sketch_name}.ino.hex:i\"\n",
				"win32":  "Performing 1200-bps touch reset on serial port /dev/ttyACM0\nWaiting for upload port...\nNo upload port found, using /dev/ttyACM0 as fallback\n\"{data_dir}/packages/arduino/tools/avrdude/6.3.0-arduino17/bin/avrdude\" \"-C{data_dir}/packages/arduino/tools/avrdude/6.3.0-arduino17/etc/avrdude.conf\" -v -V -patmega32u4 -cavr109 \"-P/dev/ttyACM0\" -b57600 -D \"-Uflash:w:{build_dir}/{sketch_name}.ino.hex:i\"\n",
			},
		},
		{
			Fqbn:       "arduino:avr:yun",
			UploadPort: "/dev/ttyACM0",
			Programmer: "",
			Output: map[string]string{
				"linux":  "Performing 1200-bps touch reset on serial port /dev/ttyACM0\nWaiting for upload port...\nNo upload port found, using /dev/ttyACM0 as fallback\n\"{data_dir}/packages/arduino/tools/avrdude/6.3.0-arduino17/bin/avrdude\" \"-C{data_dir}/packages/arduino/tools/avrdude/6.3.0-arduino17/etc/avrdude.conf\" -v -V -patmega32u4 -cavr109 \"-P/dev/ttyACM0\" -b57600 -D \"-Uflash:w:{build_dir}/{sketch_name}.ino.hex:i\"\n",
				"darwin": "Performing 1200-bps touch reset on serial port /dev/ttyACM0\nWaiting for upload port...\nNo upload port found, using /dev/ttyACM0 as fallback\n\"{data_dir}/packages/arduino/tools/avrdude/6.3.0-arduino17/bin/avrdude\" \"-C{data_dir}/packages/arduino/tools/avrdude/6.3.0-arduino17/etc/avrdude.conf\" -v -V -patmega32u4 -cavr109 \"-P/dev/ttyACM0\" -b57600 -D \"-Uflash:w:{build_dir}/{sketch_name}.ino.hex:i\"\n",
				"win32":  "Performing 1200-bps touch reset on serial port /dev/ttyACM0\nWaiting for upload port...\nNo upload port found, using /dev/ttyACM0 as fallback\n\"{data_dir}/packages/arduino/tools/avrdude/6.3.0-arduino17/bin/avrdude\" \"-C{data_dir}/packages/arduino/tools/avrdude/6.3.0-arduino17/etc/avrdude.conf\" -v -V -patmega32u4 -cavr109 \"-P/dev/ttyACM0\" -b57600 -D \"-Uflash:w:{build_dir}/{sketch_name}.ino.hex:i\"\n",
			},
		},
		{
			Fqbn:       "adafruit:avr:circuitplay32u4cat",
			UploadPort: "/dev/ttyACM0",
			Programmer: "",
			Output: map[string]string{
				"linux":  "Performing 1200-bps touch reset on serial port /dev/ttyACM0\nWaiting for upload port...\nNo upload port found, using /dev/ttyACM0 as fallback\n\"{data_dir}/packages/arduino/tools/avrdude/6.3.0-arduino17/bin/avrdude\" \"-C{data_dir}/packages/arduino/tools/avrdude/6.3.0-arduino17/etc/avrdude.conf\" -v -patmega32u4 -cavr109 -P/dev/ttyACM0 -b57600 -D \"-Uflash:w:{build_dir}/{sketch_name}.ino.hex:i\"\n",
				"darwin": "Performing 1200-bps touch reset on serial port /dev/ttyACM0\nWaiting for upload port...\nNo upload port found, using /dev/ttyACM0 as fallback\n\"{data_dir}/packages/arduino/tools/avrdude/6.3.0-arduino17/bin/avrdude\" \"-C{data_dir}/packages/arduino/tools/avrdude/6.3.0-arduino17/etc/avrdude.conf\" -v -patmega32u4 -cavr109 -P/dev/ttyACM0 -b57600 -D \"-Uflash:w:{build_dir}/{sketch_name}.ino.hex:i\"\n",
				"win32":  "Performing 1200-bps touch reset on serial port /dev/ttyACM0\nWaiting for upload port...\nNo upload port found, using /dev/ttyACM0 as fallback\n\"{data_dir}/packages/arduino/tools/avrdude/6.3.0-arduino17/bin/avrdude\" \"-C{data_dir}/packages/arduino/tools/avrdude/6.3.0-arduino17/etc/avrdude.conf\" -v -patmega32u4 -cavr109 -P/dev/ttyACM0 -b57600 -D \"-Uflash:w:{build_dir}/{sketch_name}.ino.hex:i\"\n",
			},
		},
		{
			Fqbn:       "adafruit:avr:flora8",
			UploadPort: "/dev/ttyACM0",
			Programmer: "",
			Output: map[string]string{
				"linux":  "Performing 1200-bps touch reset on serial port /dev/ttyACM0\nWaiting for upload port...\nNo upload port found, using /dev/ttyACM0 as fallback\n\"{data_dir}/packages/arduino/tools/avrdude/6.3.0-arduino17/bin/avrdude\" \"-C{data_dir}/packages/arduino/tools/avrdude/6.3.0-arduino17/etc/avrdude.conf\" -v -patmega32u4 -cavr109 -P/dev/ttyACM0 -b57600 -D \"-Uflash:w:{build_dir}/{sketch_name}.ino.hex:i\"\n",
				"darwin": "Performing 1200-bps touch reset on serial port /dev/ttyACM0\nWaiting for upload port...\nNo upload port found, using /dev/ttyACM0 as fallback\n\"{data_dir}/packages/arduino/tools/avrdude/6.3.0-arduino17/bin/avrdude\" \"-C{data_dir}/packages/arduino/tools/avrdude/6.3.0-arduino17/etc/avrdude.conf\" -v -patmega32u4 -cavr109 -P/dev/ttyACM0 -b57600 -D \"-Uflash:w:{build_dir}/{sketch_name}.ino.hex:i\"\n",
				"win32":  "Performing 1200-bps touch reset on serial port /dev/ttyACM0\nWaiting for upload port...\nNo upload port found, using /dev/ttyACM0 as fallback\n\"{data_dir}/packages/arduino/tools/avrdude/6.3.0-arduino17/bin/avrdude\" \"-C{data_dir}/packages/arduino/tools/avrdude/6.3.0-arduino17/etc/avrdude.conf\" -v -patmega32u4 -cavr109 -P/dev/ttyACM0 -b57600 -D \"-Uflash:w:{build_dir}/{sketch_name}.ino.hex:i\"\n",
			},
		},
		{
			Fqbn:       "adafruit:avr:itsybitsy32u4_3V",
			UploadPort: "/dev/ttyACM0",
			Programmer: "",
			Output: map[string]string{
				"linux":  "Performing 1200-bps touch reset on serial port /dev/ttyACM0\nWaiting for upload port...\nNo upload port found, using /dev/ttyACM0 as fallback\n\"{data_dir}/packages/arduino/tools/avrdude/6.3.0-arduino17/bin/avrdude\" \"-C{data_dir}/packages/arduino/tools/avrdude/6.3.0-arduino17/etc/avrdude.conf\" -v -patmega32u4 -cavr109 -P/dev/ttyACM0 -b57600 -D \"-Uflash:w:{build_dir}/{sketch_name}.ino.hex:i\"\n",
				"darwin": "Performing 1200-bps touch reset on serial port /dev/ttyACM0\nWaiting for upload port...\nNo upload port found, using /dev/ttyACM0 as fallback\n\"{data_dir}/packages/arduino/tools/avrdude/6.3.0-arduino17/bin/avrdude\" \"-C{data_dir}/packages/arduino/tools/avrdude/6.3.0-arduino17/etc/avrdude.conf\" -v -patmega32u4 -cavr109 -P/dev/ttyACM0 -b57600 -D \"-Uflash:w:{build_dir}/{sketch_name}.ino.hex:i\"\n",
				"win32":  "Performing 1200-bps touch reset on serial port /dev/ttyACM0\nWaiting for upload port...\nNo upload port found, using /dev/ttyACM0 as fallback\n\"{data_dir}/packages/arduino/tools/avrdude/6.3.0-arduino17/bin/avrdude\" \"-C{data_dir}/packages/arduino/tools/avrdude/6.3.0-arduino17/etc/avrdude.conf\" -v -patmega32u4 -cavr109 -P/dev/ttyACM0 -b57600 -D \"-Uflash:w:{build_dir}/{sketch_name}.ino.hex:i\"\n",
			},
		},
		{
			Fqbn:       "adafruit:avr:itsybitsy32u4_5V",
			UploadPort: "/dev/ttyACM0",
			Programmer: "",
			Output: map[string]string{
				"linux":  "Performing 1200-bps touch reset on serial port /dev/ttyACM0\nWaiting for upload port...\nNo upload port found, using /dev/ttyACM0 as fallback\n\"{data_dir}/packages/arduino/tools/avrdude/6.3.0-arduino17/bin/avrdude\" \"-C{data_dir}/packages/arduino/tools/avrdude/6.3.0-arduino17/etc/avrdude.conf\" -v -patmega32u4 -cavr109 -P/dev/ttyACM0 -b57600 -D \"-Uflash:w:{build_dir}/{sketch_name}.ino.hex:i\"\n",
				"darwin": "Performing 1200-bps touch reset on serial port /dev/ttyACM0\nWaiting for upload port...\nNo upload port found, using /dev/ttyACM0 as fallback\n\"{data_dir}/packages/arduino/tools/avrdude/6.3.0-arduino17/bin/avrdude\" \"-C{data_dir}/packages/arduino/tools/avrdude/6.3.0-arduino17/etc/avrdude.conf\" -v -patmega32u4 -cavr109 -P/dev/ttyACM0 -b57600 -D \"-Uflash:w:{build_dir}/{sketch_name}.ino.hex:i\"\n",
				"win32":  "Performing 1200-bps touch reset on serial port /dev/ttyACM0\nWaiting for upload port...\nNo upload port found, using /dev/ttyACM0 as fallback\n\"{data_dir}/packages/arduino/tools/avrdude/6.3.0-arduino17/bin/avrdude\" \"-C{data_dir}/packages/arduino/tools/avrdude/6.3.0-arduino17/etc/avrdude.conf\" -v -patmega32u4 -cavr109 -P/dev/ttyACM0 -b57600 -D \"-Uflash:w:{build_dir}/{sketch_name}.ino.hex:i\"\n",
			},
		},
		{
			Fqbn:       "STMicroelectronics:stm32:Nucleo_32:pnum=NUCLEO_F031K6,upload_method=serialMethod",
			UploadPort: "/dev/ttyACM0",
			Programmer: "",
			Output: map[string]string{
				"darwin": "\"\" sh \"{data_dir}/packages/STMicroelectronics/tools/STM32Tools/2.1.1/stm32CubeProg.sh\" 1 \"{build_dir}/{sketch_name}.ino.bin\" ttyACM0 -s\n",
				"linux":  "\"\" sh \"{data_dir}/packages/STMicroelectronics/tools/STM32Tools/2.1.1/stm32CubeProg.sh\" 1 \"{build_dir}/{sketch_name}.ino.bin\" ttyACM0 -s\n",
				"win32":  "\"{data_dir}/packages/STMicroelectronics/tools/STM32Tools/2.1.1/win/busybox.exe\" \"sh \"\"{data_dir}/packages/STMicroelectronics/tools/STM32Tools/2.1.1/stm32CubeProg.sh\" 1 \"{build_dir}/{sketch_name}.ino.bin\" ttyACM0 -s\n",
			},
		},
		{
			Fqbn:       "arduino:samd:arduino_zero_edbg",
			UploadPort: "",
			Programmer: "",
			Output: map[string]string{
				"darwin": "\"{data_dir}/packages/arduino/tools/openocd/0.10.0-arduino7/bin/openocd\" -d2 -s \"{data_dir}/packages/arduino/tools/openocd/0.10.0-arduino7/share/openocd/scripts/\" -f \"{data_dir}/packages/arduino/hardware/samd/1.8.11/variants/arduino_zero/openocd_scripts/arduino_zero.cfg\" -c \"telnet_port disabled; program {{build_dir}/{sketch_name}.ino.bin} verify reset 0x2000; shutdown\"\n",
				"linux":  "\"{data_dir}/packages/arduino/tools/openocd/0.10.0-arduino7/bin/openocd\" -d2 -s \"{data_dir}/packages/arduino/tools/openocd/0.10.0-arduino7/share/openocd/scripts/\" -f \"{data_dir}/packages/arduino/hardware/samd/1.8.11/variants/arduino_zero/openocd_scripts/arduino_zero.cfg\" -c \"telnet_port disabled; program {{build_dir}/{sketch_name}.ino.bin} verify reset 0x2000; shutdown\"\n",
				"win32":  "\"{data_dir}/packages/arduino/tools/openocd/0.10.0-arduino7/bin/openocd.exe\" -d2 -s \"{data_dir}/packages/arduino/tools/openocd/0.10.0-arduino7/share/openocd/scripts/\" -f \"{data_dir}/packages/arduino/hardware/samd/1.8.11/variants/arduino_zero/openocd_scripts/arduino_zero.cfg\" -c \"telnet_port disabled; program {{build_dir}/{sketch_name}.ino.bin} verify reset 0x2000; shutdown\"\n",
			},
		},
		{
			Fqbn:       "arduino:samd:adafruit_circuitplayground_m0",
			UploadPort: "/dev/ttyACM0",
			Programmer: "",
			Output: map[string]string{
				"darwin": "Performing 1200-bps touch reset on serial port /dev/ttyACM0\nWaiting for upload port...\nNo upload port found, using /dev/ttyACM0 as fallback\n\"{data_dir}/packages/arduino/tools/bossac/1.7.0-arduino3/bossac\" -i -d --port=ttyACM0 -U true -i -e -w -v \"{build_dir}/{sketch_name}.ino.bin\" -R\n",
				"linux":  "Performing 1200-bps touch reset on serial port /dev/ttyACM0\nWaiting for upload port...\nNo upload port found, using /dev/ttyACM0 as fallback\n\"{data_dir}/packages/arduino/tools/bossac/1.7.0-arduino3/bossac\" -i -d --port=ttyACM0 -U true -i -e -w -v \"{build_dir}/{sketch_name}.ino.bin\" -R\n",
				"win32":  "Performing 1200-bps touch reset on serial port /dev/ttyACM0\nWaiting for upload port...\nNo upload port found, using /dev/ttyACM0 as fallback\n\"{data_dir}/packages/arduino/tools/bossac/1.7.0-arduino3/bossac.exe\" -i -d --port=ttyACM0 -U true -i -e -w -v \"{build_dir}/{sketch_name}.ino.bin\" -R\n",
			},
		},
		{
			Fqbn:       "arduino:samd:adafruit_circuitplayground_m0",
			UploadPort: "/dev/ttyACM999",
			Programmer: "",
			Output: map[string]string{
				"darwin": "Performing 1200-bps touch reset on serial port /dev/ttyACM999\nWaiting for upload port...\nUpload port found on /dev/ttyACM9990\n\"{data_dir}/packages/arduino/tools/bossac/1.7.0-arduino3/bossac\" -i -d --port=ttyACM9990 -U true -i -e -w -v \"{build_dir}/{sketch_name}.ino.bin\" -R\n",
				"linux":  "Performing 1200-bps touch reset on serial port /dev/ttyACM999\nWaiting for upload port...\nUpload port found on /dev/ttyACM9990\n\"{data_dir}/packages/arduino/tools/bossac/1.7.0-arduino3/bossac\" -i -d --port=ttyACM9990 -U true -i -e -w -v \"{build_dir}/{sketch_name}.ino.bin\" -R\n",
				"win32":  "Performing 1200-bps touch reset on serial port /dev/ttyACM999\nWaiting for upload port...\nUpload port found on /dev/ttyACM9990\n\"{data_dir}/packages/arduino/tools/bossac/1.7.0-arduino3/bossac.exe\" -i -d --port=ttyACM9990 -U true -i -e -w -v \"{build_dir}/{sketch_name}.ino.bin\" -R\n",
			},
		},
		{
			Fqbn:       "arduino:samd:mkrfox1200",
			UploadPort: "/dev/ttyACM0",
			Programmer: "",
			Output: map[string]string{
				"darwin": "Performing 1200-bps touch reset on serial port /dev/ttyACM0\nWaiting for upload port...\nNo upload port found, using /dev/ttyACM0 as fallback\n\"{data_dir}/packages/arduino/tools/bossac/1.7.0-arduino3/bossac\" -i -d --port=ttyACM0 -U true -i -e -w -v \"{build_dir}/{sketch_name}.ino.bin\" -R\n",
				"linux":  "Performing 1200-bps touch reset on serial port /dev/ttyACM0\nWaiting for upload port...\nNo upload port found, using /dev/ttyACM0 as fallback\n\"{data_dir}/packages/arduino/tools/bossac/1.7.0-arduino3/bossac\" -i -d --port=ttyACM0 -U true -i -e -w -v \"{build_dir}/{sketch_name}.ino.bin\" -R\n",
				"win32":  "Performing 1200-bps touch reset on serial port /dev/ttyACM0\nWaiting for upload port...\nNo upload port found, using /dev/ttyACM0 as fallback\n\"{data_dir}/packages/arduino/tools/bossac/1.7.0-arduino3/bossac.exe\" -i -d --port=ttyACM0 -U true -i -e -w -v \"{build_dir}/{sketch_name}.ino.bin\" -R\n",
			},
		},
		{
			Fqbn:       "arduino:samd:mkrfox1200",
			UploadPort: "/dev/ttyACM999",
			Programmer: "",
			Output: map[string]string{
				"darwin": "Performing 1200-bps touch reset on serial port /dev/ttyACM999\nWaiting for upload port...\nUpload port found on /dev/ttyACM9990\n\"{data_dir}/packages/arduino/tools/bossac/1.7.0-arduino3/bossac\" -i -d --port=ttyACM9990 -U true -i -e -w -v \"{build_dir}/{sketch_name}.ino.bin\" -R\n",
				"linux":  "Performing 1200-bps touch reset on serial port /dev/ttyACM999\nWaiting for upload port...\nUpload port found on /dev/ttyACM9990\n\"{data_dir}/packages/arduino/tools/bossac/1.7.0-arduino3/bossac\" -i -d --port=ttyACM9990 -U true -i -e -w -v \"{build_dir}/{sketch_name}.ino.bin\" -R\n",
				"win32":  "Performing 1200-bps touch reset on serial port /dev/ttyACM999\nWaiting for upload port...\nUpload port found on /dev/ttyACM9990\n\"{data_dir}/packages/arduino/tools/bossac/1.7.0-arduino3/bossac.exe\" -i -d --port=ttyACM9990 -U true -i -e -w -v \"{build_dir}/{sketch_name}.ino.bin\" -R\n",
			},
		},
		{
			Fqbn:       "arduino:samd:mkrgsm1400",
			UploadPort: "/dev/ttyACM0",
			Programmer: "",
			Output: map[string]string{
				"darwin": "Performing 1200-bps touch reset on serial port /dev/ttyACM0\nWaiting for upload port...\nNo upload port found, using /dev/ttyACM0 as fallback\n\"{data_dir}/packages/arduino/tools/bossac/1.7.0-arduino3/bossac\" -i -d --port=ttyACM0 -U true -i -e -w -v \"{build_dir}/{sketch_name}.ino.bin\" -R\n",
				"linux":  "Performing 1200-bps touch reset on serial port /dev/ttyACM0\nWaiting for upload port...\nNo upload port found, using /dev/ttyACM0 as fallback\n\"{data_dir}/packages/arduino/tools/bossac/1.7.0-arduino3/bossac\" -i -d --port=ttyACM0 -U true -i -e -w -v \"{build_dir}/{sketch_name}.ino.bin\" -R\n",
				"win32":  "Performing 1200-bps touch reset on serial port /dev/ttyACM0\nWaiting for upload port...\nNo upload port found, using /dev/ttyACM0 as fallback\n\"{data_dir}/packages/arduino/tools/bossac/1.7.0-arduino3/bossac.exe\" -i -d --port=ttyACM0 -U true -i -e -w -v \"{build_dir}/{sketch_name}.ino.bin\" -R\n",
			},
		},
		{
			Fqbn:       "arduino:samd:mkrgsm1400",
			UploadPort: "/dev/ttyACM999",
			Programmer: "",
			Output: map[string]string{
				"darwin": "Performing 1200-bps touch reset on serial port /dev/ttyACM999\nWaiting for upload port...\nUpload port found on /dev/ttyACM9990\n\"{data_dir}/packages/arduino/tools/bossac/1.7.0-arduino3/bossac\" -i -d --port=ttyACM9990 -U true -i -e -w -v \"{build_dir}/{sketch_name}.ino.bin\" -R\n",
				"linux":  "Performing 1200-bps touch reset on serial port /dev/ttyACM999\nWaiting for upload port...\nUpload port found on /dev/ttyACM9990\n\"{data_dir}/packages/arduino/tools/bossac/1.7.0-arduino3/bossac\" -i -d --port=ttyACM9990 -U true -i -e -w -v \"{build_dir}/{sketch_name}.ino.bin\" -R\n",
				"win32":  "Performing 1200-bps touch reset on serial port /dev/ttyACM999\nWaiting for upload port...\nUpload port found on /dev/ttyACM9990\n\"{data_dir}/packages/arduino/tools/bossac/1.7.0-arduino3/bossac.exe\" -i -d --port=ttyACM9990 -U true -i -e -w -v \"{build_dir}/{sketch_name}.ino.bin\" -R\n",
			},
		},
		{
			Fqbn:       "arduino:samd:mkrvidor4000",
			UploadPort: "/dev/ttyACM0",
			Programmer: "",
			Output: map[string]string{
				"darwin": "Performing 1200-bps touch reset on serial port /dev/ttyACM0\nWaiting for upload port...\nNo upload port found, using /dev/ttyACM0 as fallback\n\"{data_dir}/packages/arduino/tools/bossac/1.7.0-arduino3/bossac\" -i -d --port=ttyACM0 -I -U true -i -e -w \"{build_dir}/{sketch_name}.ino.bin\" -R\n",
				"linux":  "Performing 1200-bps touch reset on serial port /dev/ttyACM0\nWaiting for upload port...\nNo upload port found, using /dev/ttyACM0 as fallback\n\"{data_dir}/packages/arduino/tools/bossac/1.7.0-arduino3/bossac\" -i -d --port=ttyACM0 -I -U true -i -e -w \"{build_dir}/{sketch_name}.ino.bin\" -R\n",
				"win32":  "Performing 1200-bps touch reset on serial port /dev/ttyACM0\nWaiting for upload port...\nNo upload port found, using /dev/ttyACM0 as fallback\n\"{data_dir}/packages/arduino/tools/bossac/1.7.0-arduino3/bossac.exe\" -i -d --port=ttyACM0 -I -U true -i -e -w \"{build_dir}/{sketch_name}.ino.bin\" -R\n",
			},
		},
		{
			Fqbn:       "arduino:samd:mkrvidor4000",
			UploadPort: "/dev/ttyACM999",
			Programmer: "",
			Output: map[string]string{
				"darwin": "Performing 1200-bps touch reset on serial port /dev/ttyACM999\nWaiting for upload port...\nUpload port found on /dev/ttyACM9990\n\"{data_dir}/packages/arduino/tools/bossac/1.7.0-arduino3/bossac\" -i -d --port=ttyACM9990 -I -U true -i -e -w \"{build_dir}/{sketch_name}.ino.bin\" -R\n",
				"linux":  "Performing 1200-bps touch reset on serial port /dev/ttyACM999\nWaiting for upload port...\nUpload port found on /dev/ttyACM9990\n\"{data_dir}/packages/arduino/tools/bossac/1.7.0-arduino3/bossac\" -i -d --port=ttyACM9990 -I -U true -i -e -w \"{build_dir}/{sketch_name}.ino.bin\" -R\n",
				"win32":  "Performing 1200-bps touch reset on serial port /dev/ttyACM999\nWaiting for upload port...\nUpload port found on /dev/ttyACM9990\n\"{data_dir}/packages/arduino/tools/bossac/1.7.0-arduino3/bossac.exe\" -i -d --port=ttyACM9990 -I -U true -i -e -w \"{build_dir}/{sketch_name}.ino.bin\" -R\n",
			},
		},
		{
			Fqbn:       "arduino:samd:mkrwan1310",
			UploadPort: "/dev/ttyACM0",
			Programmer: "",
			Output: map[string]string{
				"darwin": "Performing 1200-bps touch reset on serial port /dev/ttyACM0\nWaiting for upload port...\nNo upload port found, using /dev/ttyACM0 as fallback\n\"{data_dir}/packages/arduino/tools/bossac/1.7.0-arduino3/bossac\" -i -d --port=ttyACM0 -U true -i -e -w -v \"{build_dir}/{sketch_name}.ino.bin\" -R\n",
				"linux":  "Performing 1200-bps touch reset on serial port /dev/ttyACM0\nWaiting for upload port...\nNo upload port found, using /dev/ttyACM0 as fallback\n\"{data_dir}/packages/arduino/tools/bossac/1.7.0-arduino3/bossac\" -i -d --port=ttyACM0 -U true -i -e -w -v \"{build_dir}/{sketch_name}.ino.bin\" -R\n",
				"win32":  "Performing 1200-bps touch reset on serial port /dev/ttyACM0\nWaiting for upload port...\nNo upload port found, using /dev/ttyACM0 as fallback\n\"{data_dir}/packages/arduino/tools/bossac/1.7.0-arduino3/bossac.exe\" -i -d --port=ttyACM0 -U true -i -e -w -v \"{build_dir}/{sketch_name}.ino.bin\" -R\n",
			},
		},
		{
			Fqbn:       "arduino:samd:mkrwan1310",
			UploadPort: "/dev/ttyACM999",
			Programmer: "",
			Output: map[string]string{
				"darwin": "Performing 1200-bps touch reset on serial port /dev/ttyACM999\nWaiting for upload port...\nUpload port found on /dev/ttyACM9990\n\"{data_dir}/packages/arduino/tools/bossac/1.7.0-arduino3/bossac\" -i -d --port=ttyACM9990 -U true -i -e -w -v \"{build_dir}/{sketch_name}.ino.bin\" -R\n",
				"linux":  "Performing 1200-bps touch reset on serial port /dev/ttyACM999\nWaiting for upload port...\nUpload port found on /dev/ttyACM9990\n\"{data_dir}/packages/arduino/tools/bossac/1.7.0-arduino3/bossac\" -i -d --port=ttyACM9990 -U true -i -e -w -v \"{build_dir}/{sketch_name}.ino.bin\" -R\n",
				"win32":  "Performing 1200-bps touch reset on serial port /dev/ttyACM999\nWaiting for upload port...\nUpload port found on /dev/ttyACM9990\n\"{data_dir}/packages/arduino/tools/bossac/1.7.0-arduino3/bossac.exe\" -i -d --port=ttyACM9990 -U true -i -e -w -v \"{build_dir}/{sketch_name}.ino.bin\" -R\n",
			},
		},
		{
			Fqbn:       "arduino:samd:mkrwifi1010",
			UploadPort: "/dev/ttyACM0",
			Programmer: "",
			Output: map[string]string{
				"darwin": "Performing 1200-bps touch reset on serial port /dev/ttyACM0\nWaiting for upload port...\nNo upload port found, using /dev/ttyACM0 as fallback\n\"{data_dir}/packages/arduino/tools/bossac/1.7.0-arduino3/bossac\" -i -d --port=ttyACM0 -U true -i -e -w -v \"{build_dir}/{sketch_name}.ino.bin\" -R\n",
				"linux":  "Performing 1200-bps touch reset on serial port /dev/ttyACM0\nWaiting for upload port...\nNo upload port found, using /dev/ttyACM0 as fallback\n\"{data_dir}/packages/arduino/tools/bossac/1.7.0-arduino3/bossac\" -i -d --port=ttyACM0 -U true -i -e -w -v \"{build_dir}/{sketch_name}.ino.bin\" -R\n",
				"win32":  "Performing 1200-bps touch reset on serial port /dev/ttyACM0\nWaiting for upload port...\nNo upload port found, using /dev/ttyACM0 as fallback\n\"{data_dir}/packages/arduino/tools/bossac/1.7.0-arduino3/bossac.exe\" -i -d --port=ttyACM0 -U true -i -e -w -v \"{build_dir}/{sketch_name}.ino.bin\" -R\n",
			},
		},
		{
			Fqbn:       "arduino:samd:mkrwifi1010",
			UploadPort: "/dev/ttyACM999",
			Programmer: "",
			Output: map[string]string{
				"darwin": "Performing 1200-bps touch reset on serial port /dev/ttyACM999\nWaiting for upload port...\nUpload port found on /dev/ttyACM9990\n\"{data_dir}/packages/arduino/tools/bossac/1.7.0-arduino3/bossac\" -i -d --port=ttyACM9990 -U true -i -e -w -v \"{build_dir}/{sketch_name}.ino.bin\" -R\n",
				"linux":  "Performing 1200-bps touch reset on serial port /dev/ttyACM999\nWaiting for upload port...\nUpload port found on /dev/ttyACM9990\n\"{data_dir}/packages/arduino/tools/bossac/1.7.0-arduino3/bossac\" -i -d --port=ttyACM9990 -U true -i -e -w -v \"{build_dir}/{sketch_name}.ino.bin\" -R\n",
				"win32":  "Performing 1200-bps touch reset on serial port /dev/ttyACM999\nWaiting for upload port...\nUpload port found on /dev/ttyACM9990\n\"{data_dir}/packages/arduino/tools/bossac/1.7.0-arduino3/bossac.exe\" -i -d --port=ttyACM9990 -U true -i -e -w -v \"{build_dir}/{sketch_name}.ino.bin\" -R\n",
			},
		},
		{
			Fqbn:       "arduino:samd:mkr1000",
			UploadPort: "/dev/ttyACM0",
			Programmer: "",
			Output: map[string]string{
				"darwin": "Performing 1200-bps touch reset on serial port /dev/ttyACM0\nWaiting for upload port...\nNo upload port found, using /dev/ttyACM0 as fallback\n\"{data_dir}/packages/arduino/tools/bossac/1.7.0-arduino3/bossac\" -i -d --port=ttyACM0 -U true -i -e -w -v \"{build_dir}/{sketch_name}.ino.bin\" -R\n",
				"linux":  "Performing 1200-bps touch reset on serial port /dev/ttyACM0\nWaiting for upload port...\nNo upload port found, using /dev/ttyACM0 as fallback\n\"{data_dir}/packages/arduino/tools/bossac/1.7.0-arduino3/bossac\" -i -d --port=ttyACM0 -U true -i -e -w -v \"{build_dir}/{sketch_name}.ino.bin\" -R\n",
				"win32":  "Performing 1200-bps touch reset on serial port /dev/ttyACM0\nWaiting for upload port...\nNo upload port found, using /dev/ttyACM0 as fallback\n\"{data_dir}/packages/arduino/tools/bossac/1.7.0-arduino3/bossac.exe\" -i -d --port=ttyACM0 -U true -i -e -w -v \"{build_dir}/{sketch_name}.ino.bin\" -R\n",
			},
		},
		{
			Fqbn:       "arduino:samd:mkr1000",
			UploadPort: "",
			Programmer: "",
			Output: map[string]string{
				"darwin": "Skipping 1200-bps touch reset: no serial port selected!\nWaiting for upload port...\nUpload port found on newport\n\"{data_dir}/packages/arduino/tools/bossac/1.7.0-arduino3/bossac\" -i -d --port=newport -U true -i -e -w -v \"{build_dir}/{sketch_name}.ino.bin\" -R\n",
				"linux":  "Skipping 1200-bps touch reset: no serial port selected!\nWaiting for upload port...\nUpload port found on newport\n\"{data_dir}/packages/arduino/tools/bossac/1.7.0-arduino3/bossac\" -i -d --port=newport -U true -i -e -w -v \"{build_dir}/{sketch_name}.ino.bin\" -R\n",
				"win32":  "Skipping 1200-bps touch reset: no serial port selected!\nWaiting for upload port...\nUpload port found on newport\n\"{data_dir}/packages/arduino/tools/bossac/1.7.0-arduino3/bossac.exe\" -i -d --port=newport -U true -i -e -w -v \"{build_dir}/{sketch_name}.ino.bin\" -R\n",
			},
		},
		{
			Fqbn:       "arduino:samd:mkr1000",
			UploadPort: "/dev/ttyACM999",
			Programmer: "",
			Output: map[string]string{
				"darwin": "Performing 1200-bps touch reset on serial port /dev/ttyACM999\nWaiting for upload port...\nUpload port found on /dev/ttyACM9990\n\"{data_dir}/packages/arduino/tools/bossac/1.7.0-arduino3/bossac\" -i -d --port=ttyACM9990 -U true -i -e -w -v \"{build_dir}/{sketch_name}.ino.bin\" -R\n",
				"linux":  "Performing 1200-bps touch reset on serial port /dev/ttyACM999\nWaiting for upload port...\nUpload port found on /dev/ttyACM9990\n\"{data_dir}/packages/arduino/tools/bossac/1.7.0-arduino3/bossac\" -i -d --port=ttyACM9990 -U true -i -e -w -v \"{build_dir}/{sketch_name}.ino.bin\" -R\n",
				"win32":  "Performing 1200-bps touch reset on serial port /dev/ttyACM999\nWaiting for upload port...\nUpload port found on /dev/ttyACM9990\n\"{data_dir}/packages/arduino/tools/bossac/1.7.0-arduino3/bossac.exe\" -i -d --port=ttyACM9990 -U true -i -e -w -v \"{build_dir}/{sketch_name}.ino.bin\" -R\n",
			},
		},
		{
			Fqbn:       "arduino:samd:mkrzero",
			UploadPort: "/dev/ttyACM0",
			Programmer: "",
			Output: map[string]string{
				"darwin": "Performing 1200-bps touch reset on serial port /dev/ttyACM0\nWaiting for upload port...\nNo upload port found, using /dev/ttyACM0 as fallback\n\"{data_dir}/packages/arduino/tools/bossac/1.7.0-arduino3/bossac\" -i -d --port=ttyACM0 -U true -i -e -w -v \"{build_dir}/{sketch_name}.ino.bin\" -R\n",
				"linux":  "Performing 1200-bps touch reset on serial port /dev/ttyACM0\nWaiting for upload port...\nNo upload port found, using /dev/ttyACM0 as fallback\n\"{data_dir}/packages/arduino/tools/bossac/1.7.0-arduino3/bossac\" -i -d --port=ttyACM0 -U true -i -e -w -v \"{build_dir}/{sketch_name}.ino.bin\" -R\n",
				"win32":  "Performing 1200-bps touch reset on serial port /dev/ttyACM0\nWaiting for upload port...\nNo upload port found, using /dev/ttyACM0 as fallback\n\"{data_dir}/packages/arduino/tools/bossac/1.7.0-arduino3/bossac.exe\" -i -d --port=ttyACM0 -U true -i -e -w -v \"{build_dir}/{sketch_name}.ino.bin\" -R\n",
			},
		},
		{
			Fqbn:       "arduino:samd:mkrzero",
			UploadPort: "/dev/ttyACM999",
			Programmer: "",
			Output: map[string]string{
				"darwin": "Performing 1200-bps touch reset on serial port /dev/ttyACM999\nWaiting for upload port...\nUpload port found on /dev/ttyACM9990\n\"{data_dir}/packages/arduino/tools/bossac/1.7.0-arduino3/bossac\" -i -d --port=ttyACM9990 -U true -i -e -w -v \"{build_dir}/{sketch_name}.ino.bin\" -R\n",
				"linux":  "Performing 1200-bps touch reset on serial port /dev/ttyACM999\nWaiting for upload port...\nUpload port found on /dev/ttyACM9990\n\"{data_dir}/packages/arduino/tools/bossac/1.7.0-arduino3/bossac\" -i -d --port=ttyACM9990 -U true -i -e -w -v \"{build_dir}/{sketch_name}.ino.bin\" -R\n",
				"win32":  "Performing 1200-bps touch reset on serial port /dev/ttyACM999\nWaiting for upload port...\nUpload port found on /dev/ttyACM9990\n\"{data_dir}/packages/arduino/tools/bossac/1.7.0-arduino3/bossac.exe\" -i -d --port=ttyACM9990 -U true -i -e -w -v \"{build_dir}/{sketch_name}.ino.bin\" -R\n",
			},
		},
		{
			Fqbn:       "arduino:samd:nano_33_iot",
			UploadPort: "/dev/ttyACM0",
			Programmer: "",
			Output: map[string]string{
				"darwin": "Performing 1200-bps touch reset on serial port /dev/ttyACM0\nWaiting for upload port...\nNo upload port found, using /dev/ttyACM0 as fallback\n\"{data_dir}/packages/arduino/tools/bossac/1.7.0-arduino3/bossac\" -i -d --port=ttyACM0 -U true -i -e -w -v \"{build_dir}/{sketch_name}.ino.bin\" -R\n",
				"linux":  "Performing 1200-bps touch reset on serial port /dev/ttyACM0\nWaiting for upload port...\nNo upload port found, using /dev/ttyACM0 as fallback\n\"{data_dir}/packages/arduino/tools/bossac/1.7.0-arduino3/bossac\" -i -d --port=ttyACM0 -U true -i -e -w -v \"{build_dir}/{sketch_name}.ino.bin\" -R\n",
				"win32":  "Performing 1200-bps touch reset on serial port /dev/ttyACM0\nWaiting for upload port...\nNo upload port found, using /dev/ttyACM0 as fallback\n\"{data_dir}/packages/arduino/tools/bossac/1.7.0-arduino3/bossac.exe\" -i -d --port=ttyACM0 -U true -i -e -w -v \"{build_dir}/{sketch_name}.ino.bin\" -R\n",
			},
		},
		{
			Fqbn:       "arduino:samd:nano_33_iot",
			UploadPort: "/dev/ttyACM999",
			Programmer: "",
			Output: map[string]string{
				"darwin": "Performing 1200-bps touch reset on serial port /dev/ttyACM999\nWaiting for upload port...\nUpload port found on /dev/ttyACM9990\n\"{data_dir}/packages/arduino/tools/bossac/1.7.0-arduino3/bossac\" -i -d --port=ttyACM9990 -U true -i -e -w -v \"{build_dir}/{sketch_name}.ino.bin\" -R\n",
				"linux":  "Performing 1200-bps touch reset on serial port /dev/ttyACM999\nWaiting for upload port...\nUpload port found on /dev/ttyACM9990\n\"{data_dir}/packages/arduino/tools/bossac/1.7.0-arduino3/bossac\" -i -d --port=ttyACM9990 -U true -i -e -w -v \"{build_dir}/{sketch_name}.ino.bin\" -R\n",
				"win32":  "Performing 1200-bps touch reset on serial port /dev/ttyACM999\nWaiting for upload port...\nUpload port found on /dev/ttyACM9990\n\"{data_dir}/packages/arduino/tools/bossac/1.7.0-arduino3/bossac.exe\" -i -d --port=ttyACM9990 -U true -i -e -w -v \"{build_dir}/{sketch_name}.ino.bin\" -R\n",
			},
		},
		{
			Fqbn:       "arduino:samd:arduino_zero_native",
			UploadPort: "/dev/ttyACM0",
			Programmer: "",
			Output: map[string]string{
				"darwin": "Performing 1200-bps touch reset on serial port /dev/ttyACM0\nWaiting for upload port...\nNo upload port found, using /dev/ttyACM0 as fallback\n\"{data_dir}/packages/arduino/tools/bossac/1.7.0-arduino3/bossac\" -i -d --port=ttyACM0 -U true -i -e -w -v \"{build_dir}/{sketch_name}.ino.bin\" -R\n",
				"linux":  "Performing 1200-bps touch reset on serial port /dev/ttyACM0\nWaiting for upload port...\nNo upload port found, using /dev/ttyACM0 as fallback\n\"{data_dir}/packages/arduino/tools/bossac/1.7.0-arduino3/bossac\" -i -d --port=ttyACM0 -U true -i -e -w -v \"{build_dir}/{sketch_name}.ino.bin\" -R\n",
				"win32":  "Performing 1200-bps touch reset on serial port /dev/ttyACM0\nWaiting for upload port...\nNo upload port found, using /dev/ttyACM0 as fallback\n\"{data_dir}/packages/arduino/tools/bossac/1.7.0-arduino3/bossac.exe\" -i -d --port=ttyACM0 -U true -i -e -w -v \"{build_dir}/{sketch_name}.ino.bin\" -R\n",
			},
		},
		{
			Fqbn:       "arduino:samd:arduino_zero_native",
			UploadPort: "/dev/ttyACM999",
			Programmer: "",
			Output: map[string]string{
				"darwin": "Performing 1200-bps touch reset on serial port /dev/ttyACM999\nWaiting for upload port...\nUpload port found on /dev/ttyACM9990\n\"{data_dir}/packages/arduino/tools/bossac/1.7.0-arduino3/bossac\" -i -d --port=ttyACM9990 -U true -i -e -w -v \"{build_dir}/{sketch_name}.ino.bin\" -R\n",
				"linux":  "Performing 1200-bps touch reset on serial port /dev/ttyACM999\nWaiting for upload port...\nUpload port found on /dev/ttyACM9990\n\"{data_dir}/packages/arduino/tools/bossac/1.7.0-arduino3/bossac\" -i -d --port=ttyACM9990 -U true -i -e -w -v \"{build_dir}/{sketch_name}.ino.bin\" -R\n",
				"win32":  "Performing 1200-bps touch reset on serial port /dev/ttyACM999\nWaiting for upload port...\nUpload port found on /dev/ttyACM9990\n\"{data_dir}/packages/arduino/tools/bossac/1.7.0-arduino3/bossac.exe\" -i -d --port=ttyACM9990 -U true -i -e -w -v \"{build_dir}/{sketch_name}.ino.bin\" -R\n",
			},
		},
		{
			Fqbn:       "esp32:esp32:esp32",
			UploadPort: "/dev/ttyACM0",
			Programmer: "",
			Output: map[string]string{
				"darwin": "\"{data_dir}/packages/esp32/tools/esptool_py/3.0.0/esptool\" --chip esp32 --port \"/dev/ttyACM0\" --baud 921600  --before default_reset --after hard_reset write_flash -z --flash_mode dio --flash_freq 80m --flash_size detect 0xe000 \"{data_dir}/packages/esp32/hardware/esp32/1.0.6/tools/partitions/boot_app0.bin\" 0x1000 \"{data_dir}/packages/esp32/hardware/esp32/1.0.6/tools/sdk/bin/bootloader_qio_80m.bin\" 0x10000 \"{build_dir}/{sketch_name}.ino.bin\" 0x8000 \"{build_dir}/{sketch_name}.ino.partitions.bin\"\n",
				"linux":  "python \"{data_dir}/packages/esp32/tools/esptool_py/3.0.0/esptool.py\" --chip esp32 --port \"/dev/ttyACM0\" --baud 921600  --before default_reset --after hard_reset write_flash -z --flash_mode dio --flash_freq 80m --flash_size detect 0xe000 \"{data_dir}/packages/esp32/hardware/esp32/1.0.6/tools/partitions/boot_app0.bin\" 0x1000 \"{data_dir}/packages/esp32/hardware/esp32/1.0.6/tools/sdk/bin/bootloader_qio_80m.bin\" 0x10000 \"{build_dir}/{sketch_name}.ino.bin\" 0x8000 \"{build_dir}/{sketch_name}.ino.partitions.bin\"\n",
				"win32":  "\"{data_dir}/packages/esp32/tools/esptool_py/3.0.0/esptool.exe\" --chip esp32 --port \"/dev/ttyACM0\" --baud 921600  --before default_reset --after hard_reset write_flash -z --flash_mode dio --flash_freq 80m --flash_size detect 0xe000 \"{data_dir}/packages/esp32/hardware/esp32/1.0.6/tools/partitions/boot_app0.bin\" 0x1000 \"{data_dir}/packages/esp32/hardware/esp32/1.0.6/tools/sdk/bin/bootloader_qio_80m.bin\" 0x10000 \"{build_dir}/{sketch_name}.ino.bin\" 0x8000 \"{build_dir}/{sketch_name}.ino.partitions.bin\"\n",
			},
		},
		{
			Fqbn:       "esp32:esp32:esp32:PSRAM=enabled,PartitionScheme=no_ota,CPUFreq=80,FlashMode=dio,FlashFreq=40,FlashSize=8M,UploadSpeed=230400,DebugLevel=info",
			UploadPort: "/dev/ttyACM0",
			Programmer: "",
			Output: map[string]string{
				"darwin": "\"{data_dir}/packages/esp32/tools/esptool_py/3.0.0/esptool\" --chip esp32 --port \"/dev/ttyACM0\" --baud 230400  --before default_reset --after hard_reset write_flash -z --flash_mode dio --flash_freq 40m --flash_size detect 0xe000 \"{data_dir}/packages/esp32/hardware/esp32/1.0.6/tools/partitions/boot_app0.bin\" 0x1000 \"{data_dir}/packages/esp32/hardware/esp32/1.0.6/tools/sdk/bin/bootloader_dio_40m.bin\" 0x10000 \"{build_dir}/{sketch_name}.ino.bin\" 0x8000 \"{build_dir}/{sketch_name}.ino.partitions.bin\"\n",
				"linux":  "python \"{data_dir}/packages/esp32/tools/esptool_py/3.0.0/esptool.py\" --chip esp32 --port \"/dev/ttyACM0\" --baud 230400  --before default_reset --after hard_reset write_flash -z --flash_mode dio --flash_freq 40m --flash_size detect 0xe000 \"{data_dir}/packages/esp32/hardware/esp32/1.0.6/tools/partitions/boot_app0.bin\" 0x1000 \"{data_dir}/packages/esp32/hardware/esp32/1.0.6/tools/sdk/bin/bootloader_dio_40m.bin\" 0x10000 \"{build_dir}/{sketch_name}.ino.bin\" 0x8000 \"{build_dir}/{sketch_name}.ino.partitions.bin\"\n",
				"win32":  "\"{data_dir}/packages/esp32/tools/esptool_py/3.0.0/esptool.exe\" --chip esp32 --port \"/dev/ttyACM0\" --baud 230400  --before default_reset --after hard_reset write_flash -z --flash_mode dio --flash_freq 40m --flash_size detect 0xe000 \"{data_dir}/packages/esp32/hardware/esp32/1.0.6/tools/partitions/boot_app0.bin\" 0x1000 \"{data_dir}/packages/esp32/hardware/esp32/1.0.6/tools/sdk/bin/bootloader_dio_40m.bin\" 0x10000 \"{build_dir}/{sketch_name}.ino.bin\" 0x8000 \"{build_dir}/{sketch_name}.ino.partitions.bin\"\n",
			},
		},
		{
			Fqbn:       "esp32:esp32:esp32thing",
			UploadPort: "/dev/ttyACM0",
			Programmer: "",
			Output: map[string]string{
				"darwin": "\"{data_dir}/packages/esp32/tools/esptool_py/3.0.0/esptool\" --chip esp32 --port \"/dev/ttyACM0\" --baud 921600  --before default_reset --after hard_reset write_flash -z --flash_mode dio --flash_freq 80m --flash_size detect 0xe000 \"{data_dir}/packages/esp32/hardware/esp32/1.0.6/tools/partitions/boot_app0.bin\" 0x1000 \"{data_dir}/packages/esp32/hardware/esp32/1.0.6/tools/sdk/bin/bootloader_dio_80m.bin\" 0x10000 \"{build_dir}/{sketch_name}.ino.bin\" 0x8000 \"{build_dir}/{sketch_name}.ino.partitions.bin\"\n",
				"linux":  "python \"{data_dir}/packages/esp32/tools/esptool_py/3.0.0/esptool.py\" --chip esp32 --port \"/dev/ttyACM0\" --baud 921600  --before default_reset --after hard_reset write_flash -z --flash_mode dio --flash_freq 80m --flash_size detect 0xe000 \"{data_dir}/packages/esp32/hardware/esp32/1.0.6/tools/partitions/boot_app0.bin\" 0x1000 \"{data_dir}/packages/esp32/hardware/esp32/1.0.6/tools/sdk/bin/bootloader_dio_80m.bin\" 0x10000 \"{build_dir}/{sketch_name}.ino.bin\" 0x8000 \"{build_dir}/{sketch_name}.ino.partitions.bin\"\n",
				"win32":  "\"{data_dir}/packages/esp32/tools/esptool_py/3.0.0/esptool.exe\" --chip esp32 --port \"/dev/ttyACM0\" --baud 921600  --before default_reset --after hard_reset write_flash -z --flash_mode dio --flash_freq 80m --flash_size detect 0xe000 \"{data_dir}/packages/esp32/hardware/esp32/1.0.6/tools/partitions/boot_app0.bin\" 0x1000 \"{data_dir}/packages/esp32/hardware/esp32/1.0.6/tools/sdk/bin/bootloader_dio_80m.bin\" 0x10000 \"{build_dir}/{sketch_name}.ino.bin\" 0x8000 \"{build_dir}/{sketch_name}.ino.partitions.bin\"\n",
			},
		},
		{
			Fqbn:       "SPRESENSE:spresense:spresense",
			UploadPort: "/dev/ttyACM0",
			Programmer: "",
			Output: map[string]string{
				"darwin": "\"{data_dir}/packages/SPRESENSE/tools/spresense-tools/2.0.2/flash_writer/macosx/flash_writer\" -s -c \"/dev/ttyACM0\"  -d -n \"{build_dir}/{sketch_name}.ino.spk\"",
				"linux":  "\"{data_dir}/packages/SPRESENSE/tools/spresense-tools/2.0.2/flash_writer/linux/flash_writer\" -s -c \"/dev/ttyACM0\"  -d -n \"{build_dir}/{sketch_name}.ino.spk\"",
				"win32":  "\"{data_dir}/packages/SPRESENSE/tools/spresense-tools/2.0.2/flash_writer/windows/flash_writer.exe\" -s -c \"/dev/ttyACM0\"  -d -n \"{build_dir}/{sketch_name}.ino.spk\"",
			},
		},
	}

	if cli.DataDir().Join("packages").NotExist() {
		_, _, err := cli.Run("config", "init", "--overwrite")
		require.NoError(t, err)
		for _, v := range indexes {
			_, _, err := cli.Run("config", "add", "board_manager.additional_urls", v)
			require.NoError(t, err)
		}
		_, _, err = cli.Run("update")
		require.NoError(t, err)
		for _, v := range coresToInstall {
			_, _, err := cli.Run("core", "install", v)
			require.NoError(t, err)
		}
	}

	sketchName := "TestSketchForUpload"
	sketchPath := cli.SketchbookDir().Join(sketchName)
	_, _, err := cli.Run("sketch", "new", sketchPath.String())
	require.NoError(t, err)
	buildDir := generateBuildDir(sketchPath, t)
	t.Cleanup(func() { buildDir.RemoveAll() })

	for i, _test := range testParameters {
		test := _test
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			t.Parallel()
			var stdout []byte
			if test.Programmer != "" {
				if test.UploadPort != "" {
					stdout, _, err = cli.Run("upload", "-p", test.UploadPort, "-P", test.Programmer, "-b", test.Fqbn, sketchPath.String(), "--dry-run", "-v")
					require.NoError(t, err)
				} else {
					stdout, _, err = cli.Run("upload", "-P", test.Programmer, "-b", test.Fqbn, sketchPath.String(), "--dry-run", "-v")
					require.NoError(t, err)
				}
			} else {
				if test.UploadPort != "" {
					stdout, _, err = cli.Run("upload", "-p", test.UploadPort, "-b", test.Fqbn, sketchPath.String(), "--dry-run", "-v")
					require.NoError(t, err)
				} else {
					stdout, _, err = cli.Run("upload", "-b", test.Fqbn, sketchPath.String(), "--dry-run", "-v")
					require.NoError(t, err)
				}
			}
			r := strings.NewReplacer("{data_dir}", cli.DataDir().String(), "{upload_port", test.UploadPort,
				"{build_dir}", buildDir.String(), "{sketch_name}", sketchName)
			expectedOut := strings.ReplaceAll(r.Replace(test.Output), "\\", "/")
			require.Contains(t, strings.ReplaceAll(string(stdout), "\\", "/"), expectedOut)
		})
	}

	for i, _test := range testParametersMap {
		test := _test
		t.Run(fmt.Sprintf("WithMap%d", i), func(t *testing.T) {
			t.Parallel()
			var stdout []byte
			if test.Programmer != "" {
				if test.UploadPort != "" {
					stdout, _, err = cli.Run("upload", "-p", test.UploadPort, "-P", test.Programmer, "-b", test.Fqbn, sketchPath.String(), "--dry-run", "-v")
					require.NoError(t, err)
				} else {
					stdout, _, err = cli.Run("upload", "-P", test.Programmer, "-b", test.Fqbn, sketchPath.String(), "--dry-run", "-v")
					require.NoError(t, err)
				}
			} else {
				if test.UploadPort != "" {
					stdout, _, err = cli.Run("upload", "-p", test.UploadPort, "-b", test.Fqbn, sketchPath.String(), "--dry-run", "-v")
					require.NoError(t, err)
				} else {
					stdout, _, err = cli.Run("upload", "-b", test.Fqbn, sketchPath.String(), "--dry-run", "-v")
					require.NoError(t, err)
				}
			}
			out := test.Output[runtime.GOOS]
			r := strings.NewReplacer("{data_dir}", cli.DataDir().String(), "{upload_port", test.UploadPort,
				"{build_dir}", buildDir.String(), "{sketch_name}", sketchName)
			expectedOut := strings.ReplaceAll(r.Replace(out), "\\", "/")
			require.Contains(t, strings.ReplaceAll(string(stdout), "\\", "/"), expectedOut)
		})
	}
}

func generateBuildDir(sketchPath *paths.Path, t *testing.T) *paths.Path {
	md5 := md5.Sum(([]byte(sketchPath.String())))
	sketchPathMd5 := strings.ToUpper(hex.EncodeToString(md5[:]))
	buildDir := paths.TempDir().Join("arduino", "sketches", sketchPathMd5)
	require.NoError(t, buildDir.MkdirAll())
	require.NoError(t, buildDir.ToAbs())
	return buildDir
}

func TestUploadWithInputDirFlag(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	_, _, err := cli.Run("core", "install", "arduino:mbed_opta")
	require.NoError(t, err)

	sketchPath := cli.SketchbookDir().Join("TestSketchForUpload")
	_, _, err = cli.Run("sketch", "new", sketchPath.String())
	require.NoError(t, err)

	// Create a fake build directory
	buildDir := sketchPath.Join("build")
	require.NoError(t, buildDir.MkdirAll())
	require.NoError(t, buildDir.Join("TestSketchForUpload.ino.bin").WriteFile(nil))
	require.NoError(t, buildDir.Join("TestSketchForUpload.ino.elf").WriteFile(nil))
	require.NoError(t, buildDir.Join("TestSketchForUpload.ino.hex").WriteFile(nil))
	require.NoError(t, buildDir.Join("TestSketchForUpload.ino.map").WriteFile(nil))

	// Test with input-dir flag
	_, _, err = cli.Run(
		"upload",
		"-b", "arduino:mbed_opta:opta",
		"-i", buildDir.String(),
		"-t",
		"-p", "/dev/ttyACM0",
		"--dry-run", "-v",
		sketchPath.String())
	require.NoError(t, err)

	// Test with input-dir flag and no sketch
	_, _, err = cli.Run(
		"upload",
		"-b", "arduino:mbed_opta:opta",
		"-i", buildDir.String(),
		"-t",
		"-p", "/dev/ttyACM0",
		"--dry-run", "-v")
	require.NoError(t, err)
}
