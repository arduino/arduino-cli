# This file is part of arduino-cli.
#
# Copyright 2020 ARDUINO SA (http://www.arduino.cc/)
#
# This software is released under the GNU General Public License version 3,
# which covers the main part of arduino-cli.
# The terms of this license can be found at:
# https://www.gnu.org/licenses/gpl-3.0.en.html
#
# You can be released from the requirements of the above licenses by purchasing
# a commercial license. Buying such a license is mandatory if you want to modify or
# otherwise use the software for commercial activities involving the Arduino
# software without disclosing the source code of your own applications. To purchase
# a commercial license, send an email to license@arduino.cc.

import tempfile
import sys
import hashlib
import pytest
from pathlib import Path


def generate_build_dir(sketch_path):
    sketch_path_md5 = hashlib.md5(bytes(sketch_path)).hexdigest().upper()
    build_dir = Path(tempfile.gettempdir(), f"arduino-sketch-{sketch_path_md5}")
    build_dir.mkdir(parents=True, exist_ok=True)
    return build_dir.resolve()


indexes = [
    "https://github.com/stm32duino/BoardManagerFiles/raw/main/package_stmicroelectronics_index.json",
    "https://adafruit.github.io/arduino-board-index/package_adafruit_index.json",
    "https://raw.githubusercontent.com/espressif/arduino-esp32/gh-pages/package_esp32_index.json",
    "http://arduino.esp8266.com/stable/package_esp8266com_index.json",
    "https://github.com/sonydevworld/spresense-arduino-compatible/releases/download/generic/package_spresense_index.json",
]

cores_to_install = [
    "STMicroelectronics:stm32@2.2.0",
    "arduino:avr@1.8.3",
    "adafruit:avr@1.4.13",
    "arduino:samd@1.8.11",
    "esp32:esp32@1.0.6",
    "esp8266:esp8266@3.0.2",
    "SPRESENSE:spresense@2.0.2",
]

testdata = [
    (
        "STMicroelectronics:stm32:Nucleo_32:pnum=NUCLEO_F031K6,upload_method=serialMethod",
        "/dev/ttyACM0",
        "",
        {
            "darwin": '"" sh '
            '"{data_dir}/packages/STMicroelectronics/tools/STM32Tools/2.1.1/stm32CubeProg.sh" '
            '1 "{build_dir}/{sketch_name}.ino.bin" ttyACM0 -s\n',
            "linux": '"" sh '
            '"{data_dir}/packages/STMicroelectronics/tools/STM32Tools/2.1.1/stm32CubeProg.sh" '
            '1 "{build_dir}/{sketch_name}.ino.bin" ttyACM0 -s\n',
            "win32": '"{data_dir}/packages/STMicroelectronics/tools/STM32Tools/2.1.1/win/busybox.exe" '
            "sh "
            '"{data_dir}/packages/STMicroelectronics/tools/STM32Tools/2.1.1/stm32CubeProg.sh" '
            '1 "{build_dir}/{sketch_name}.ino.bin" ttyACM0 -s\n',
        },
    ),
    (
        "arduino:avr:uno",
        "/dev/ttyACM0",
        "",
        '"{data_dir}/packages/arduino/tools/avrdude/6.3.0-arduino17/bin/avrdude" '
        '"-C{data_dir}/packages/arduino/tools/avrdude/6.3.0-arduino17/etc/avrdude.conf" '
        '-v -V -patmega328p -carduino "-P/dev/ttyACM0" -b115200 -D '
        '"-Uflash:w:{build_dir}/{sketch_name}.ino.hex:i"\n',
    ),
    (
        "arduino:avr:uno",
        "",
        "usbasp",
        '"{data_dir}/packages/arduino/tools/avrdude/6.3.0-arduino17/bin/avrdude" '
        '"-C{data_dir}/packages/arduino/tools/avrdude/6.3.0-arduino17/etc/avrdude.conf" '
        "-v -V -patmega328p -cusbasp -Pusb "
        '"-Uflash:w:{build_dir}/{sketch_name}.ino.hex:i"\n',
    ),
    (
        "arduino:avr:uno",
        "/dev/ttyACM0",
        "avrisp",
        '"{data_dir}/packages/arduino/tools/avrdude/6.3.0-arduino17/bin/avrdude" '
        '"-C{data_dir}/packages/arduino/tools/avrdude/6.3.0-arduino17/etc/avrdude.conf" '
        "-v -V -patmega328p -cstk500v1 -P/dev/ttyACM0 "
        '"-Uflash:w:{build_dir}/{sketch_name}.ino.hex:i"\n',
    ),
    (
        "arduino:avr:uno",
        "/dev/ttyACM0",
        "arduinoasisp",
        '"{data_dir}/packages/arduino/tools/avrdude/6.3.0-arduino17/bin/avrdude" '
        '"-C{data_dir}/packages/arduino/tools/avrdude/6.3.0-arduino17/etc/avrdude.conf" '
        "-v -V -patmega328p -cstk500v1 -P/dev/ttyACM0 -b19200 "
        '"-Uflash:w:{build_dir}/{sketch_name}.ino.hex:i"\n',
    ),
    (
        "arduino:avr:nano",
        "/dev/ttyACM0",
        "",
        '"{data_dir}/packages/arduino/tools/avrdude/6.3.0-arduino17/bin/avrdude" '
        '"-C{data_dir}/packages/arduino/tools/avrdude/6.3.0-arduino17/etc/avrdude.conf" '
        '-v -V -patmega328p -carduino "-P/dev/ttyACM0" -b115200 -D '
        '"-Uflash:w:{build_dir}/{sketch_name}.ino.hex:i"\n',
    ),
    (
        "arduino:avr:nano",
        "",
        "usbasp",
        '"{data_dir}/packages/arduino/tools/avrdude/6.3.0-arduino17/bin/avrdude" '
        '"-C{data_dir}/packages/arduino/tools/avrdude/6.3.0-arduino17/etc/avrdude.conf" '
        "-v -V -patmega328p -cusbasp -Pusb "
        '"-Uflash:w:{build_dir}/{sketch_name}.ino.hex:i"\n',
    ),
    (
        "arduino:avr:nano",
        "/dev/ttyACM0",
        "avrisp",
        '"{data_dir}/packages/arduino/tools/avrdude/6.3.0-arduino17/bin/avrdude" '
        '"-C{data_dir}/packages/arduino/tools/avrdude/6.3.0-arduino17/etc/avrdude.conf" '
        "-v -V -patmega328p -cstk500v1 -P/dev/ttyACM0 "
        '"-Uflash:w:{build_dir}/{sketch_name}.ino.hex:i"\n',
    ),
    (
        "arduino:avr:nano",
        "/dev/ttyACM0",
        "arduinoasisp",
        '"{data_dir}/packages/arduino/tools/avrdude/6.3.0-arduino17/bin/avrdude" '
        '"-C{data_dir}/packages/arduino/tools/avrdude/6.3.0-arduino17/etc/avrdude.conf" '
        "-v -V -patmega328p -cstk500v1 -P/dev/ttyACM0 -b19200 "
        '"-Uflash:w:{build_dir}/{sketch_name}.ino.hex:i"\n',
    ),
    (
        "arduino:avr:nano:cpu=atmega328old",
        "/dev/ttyACM0",
        "",
        '"{data_dir}/packages/arduino/tools/avrdude/6.3.0-arduino17/bin/avrdude" '
        '"-C{data_dir}/packages/arduino/tools/avrdude/6.3.0-arduino17/etc/avrdude.conf" '
        '-v -V -patmega328p -carduino "-P/dev/ttyACM0" -b57600 -D '
        '"-Uflash:w:{build_dir}/{sketch_name}.ino.hex:i"\n',
    ),
    (
        "arduino:avr:nano:cpu=atmega328old",
        "",
        "usbasp",
        '"{data_dir}/packages/arduino/tools/avrdude/6.3.0-arduino17/bin/avrdude" '
        '"-C{data_dir}/packages/arduino/tools/avrdude/6.3.0-arduino17/etc/avrdude.conf" '
        "-v -V -patmega328p -cusbasp -Pusb "
        '"-Uflash:w:{build_dir}/{sketch_name}.ino.hex:i"\n',
    ),
    (
        "arduino:avr:nano:cpu=atmega328old",
        "/dev/ttyACM0",
        "avrisp",
        '"{data_dir}/packages/arduino/tools/avrdude/6.3.0-arduino17/bin/avrdude" '
        '"-C{data_dir}/packages/arduino/tools/avrdude/6.3.0-arduino17/etc/avrdude.conf" '
        "-v -V -patmega328p -cstk500v1 -P/dev/ttyACM0 "
        '"-Uflash:w:{build_dir}/{sketch_name}.ino.hex:i"\n',
    ),
    (
        "arduino:avr:nano:cpu=atmega328old",
        "/dev/ttyACM0",
        "arduinoasisp",
        '"{data_dir}/packages/arduino/tools/avrdude/6.3.0-arduino17/bin/avrdude" '
        '"-C{data_dir}/packages/arduino/tools/avrdude/6.3.0-arduino17/etc/avrdude.conf" '
        "-v -V -patmega328p -cstk500v1 -P/dev/ttyACM0 -b19200 "
        '"-Uflash:w:{build_dir}/{sketch_name}.ino.hex:i"\n',
    ),
    (
        "arduino:avr:mega",
        "/dev/ttyACM0",
        "",
        '"{data_dir}/packages/arduino/tools/avrdude/6.3.0-arduino17/bin/avrdude" '
        '"-C{data_dir}/packages/arduino/tools/avrdude/6.3.0-arduino17/etc/avrdude.conf" '
        '-v -V -patmega2560 -cwiring "-P/dev/ttyACM0" -b115200 -D '
        '"-Uflash:w:{build_dir}/{sketch_name}.ino.hex:i"\n',
    ),
    (
        "arduino:avr:mega:cpu=atmega1280",
        "/dev/ttyACM0",
        "",
        '"{data_dir}/packages/arduino/tools/avrdude/6.3.0-arduino17/bin/avrdude" '
        '"-C{data_dir}/packages/arduino/tools/avrdude/6.3.0-arduino17/etc/avrdude.conf" '
        '-v -V -patmega1280 -carduino "-P/dev/ttyACM0" -b57600 -D '
        '"-Uflash:w:{build_dir}/{sketch_name}.ino.hex:i"\n',
    ),
    (
        "arduino:avr:diecimila",
        "/dev/ttyACM0",
        "",
        '"{data_dir}/packages/arduino/tools/avrdude/6.3.0-arduino17/bin/avrdude" '
        '"-C{data_dir}/packages/arduino/tools/avrdude/6.3.0-arduino17/etc/avrdude.conf" '
        '-v -V -patmega328p -carduino "-P/dev/ttyACM0" -b57600 -D '
        '"-Uflash:w:{build_dir}/{sketch_name}.ino.hex:i"\n',
    ),
    (
        "arduino:avr:leonardo",
        "/dev/ttyACM0",
        "",
        "Performing 1200-bps touch reset on serial port /dev/ttyACM0\n"
        "Waiting for upload port...\n"
        "No upload port found, using /dev/ttyACM0 as fallback\n"
        '"{data_dir}/packages/arduino/tools/avrdude/6.3.0-arduino17/bin/avrdude" '
        '"-C{data_dir}/packages/arduino/tools/avrdude/6.3.0-arduino17/etc/avrdude.conf" '
        '-v -V -patmega32u4 -cavr109 "-P/dev/ttyACM0" -b57600 -D '
        '"-Uflash:w:{build_dir}/{sketch_name}.ino.hex:i"\n',
    ),
    (
        "arduino:avr:leonardo",
        "/dev/ttyACM999",
        "",
        "Performing 1200-bps touch reset on serial port /dev/ttyACM999\n"
        "Waiting for upload port...\n"
        "Upload port found on /dev/ttyACM9990\n"
        '"{data_dir}/packages/arduino/tools/avrdude/6.3.0-arduino17/bin/avrdude" '
        '"-C{data_dir}/packages/arduino/tools/avrdude/6.3.0-arduino17/etc/avrdude.conf" '
        '-v -V -patmega32u4 -cavr109 "-P/dev/ttyACM9990" -b57600 -D '
        '"-Uflash:w:{build_dir}/{sketch_name}.ino.hex:i"\n',
    ),
    (
        "arduino:avr:micro",
        "/dev/ttyACM0",
        "",
        "Performing 1200-bps touch reset on serial port /dev/ttyACM0\n"
        "Waiting for upload port...\n"
        "No upload port found, using /dev/ttyACM0 as fallback\n"
        '"{data_dir}/packages/arduino/tools/avrdude/6.3.0-arduino17/bin/avrdude" '
        '"-C{data_dir}/packages/arduino/tools/avrdude/6.3.0-arduino17/etc/avrdude.conf" '
        '-v -V -patmega32u4 -cavr109 "-P/dev/ttyACM0" -b57600 -D '
        '"-Uflash:w:{build_dir}/{sketch_name}.ino.hex:i"\n',
    ),
    (
        "arduino:avr:micro",
        "/dev/ttyACM999",
        "",
        "Performing 1200-bps touch reset on serial port /dev/ttyACM999\n"
        "Waiting for upload port...\n"
        "Upload port found on /dev/ttyACM9990\n"
        '"{data_dir}/packages/arduino/tools/avrdude/6.3.0-arduino17/bin/avrdude" '
        '"-C{data_dir}/packages/arduino/tools/avrdude/6.3.0-arduino17/etc/avrdude.conf" '
        '-v -V -patmega32u4 -cavr109 "-P/dev/ttyACM9990" -b57600 -D '
        '"-Uflash:w:{build_dir}/{sketch_name}.ino.hex:i"\n',
    ),
    (
        "arduino:avr:circuitplay32u4cat",
        "/dev/ttyACM0",
        "",
        "Performing 1200-bps touch reset on serial port /dev/ttyACM0\n"
        "Waiting for upload port...\n"
        "No upload port found, using /dev/ttyACM0 as fallback\n"
        '"{data_dir}/packages/arduino/tools/avrdude/6.3.0-arduino17/bin/avrdude" '
        '"-C{data_dir}/packages/arduino/tools/avrdude/6.3.0-arduino17/etc/avrdude.conf" '
        '-v -V -patmega32u4 -cavr109 "-P/dev/ttyACM0" -b57600 -D '
        '"-Uflash:w:{build_dir}/{sketch_name}.ino.hex:i"\n',
    ),
    (
        "arduino:avr:circuitplay32u4cat",
        "/dev/ttyACM999",
        "",
        "Performing 1200-bps touch reset on serial port /dev/ttyACM999\n"
        "Waiting for upload port...\n"
        "Upload port found on /dev/ttyACM9990\n"
        '"{data_dir}/packages/arduino/tools/avrdude/6.3.0-arduino17/bin/avrdude" '
        '"-C{data_dir}/packages/arduino/tools/avrdude/6.3.0-arduino17/etc/avrdude.conf" '
        '-v -V -patmega32u4 -cavr109 "-P/dev/ttyACM9990" -b57600 -D '
        '"-Uflash:w:{build_dir}/{sketch_name}.ino.hex:i"\n',
    ),
    (
        "arduino:avr:gemma",
        "/dev/ttyACM0",
        "usbGemma",
        '"{data_dir}/packages/arduino/tools/avrdude/6.3.0-arduino17/bin/avrdude" '
        '"-C{data_dir}/packages/arduino/hardware/avr/1.8.3/bootloaders/gemma/avrdude.conf" '
        "-v -V -pattiny85 -carduinogemma  "
        '"-Uflash:w:{build_dir}/{sketch_name}.ino.hex:i"\n',
    ),
    (
        "arduino:avr:gemma",
        "",
        "usbGemma",
        '"{data_dir}/packages/arduino/tools/avrdude/6.3.0-arduino17/bin/avrdude" '
        '"-C{data_dir}/packages/arduino/hardware/avr/1.8.3/bootloaders/gemma/avrdude.conf" '
        "-v -V -pattiny85 -carduinogemma  "
        '"-Uflash:w:{build_dir}/{sketch_name}.ino.hex:i"\n',
    ),
    (
        "arduino:avr:unowifi",
        "/dev/ttyACM0",
        "",
        '"{data_dir}/packages/arduino/tools/avrdude/6.3.0-arduino17/bin/avrdude" '
        '"-C{data_dir}/packages/arduino/tools/avrdude/6.3.0-arduino17/etc/avrdude.conf" '
        '-v -V -patmega328p -carduino "-P/dev/ttyACM0" -b115200 -D '
        '"-Uflash:w:{build_dir}/{sketch_name}.ino.hex:i"\n',
    ),
    (
        "arduino:avr:yun",
        "/dev/ttyACM0",
        "",
        "Performing 1200-bps touch reset on serial port /dev/ttyACM0\n"
        "Waiting for upload port...\n"
        "No upload port found, using /dev/ttyACM0 as fallback\n"
        '"{data_dir}/packages/arduino/tools/avrdude/6.3.0-arduino17/bin/avrdude" '
        '"-C{data_dir}/packages/arduino/tools/avrdude/6.3.0-arduino17/etc/avrdude.conf" '
        '-v -V -patmega32u4 -cavr109 "-P/dev/ttyACM0" -b57600 -D '
        '"-Uflash:w:{build_dir}/{sketch_name}.ino.hex:i"\n',
    ),
    (
        "arduino:avr:yun",
        "/dev/ttyACM999",
        "",
        "Performing 1200-bps touch reset on serial port /dev/ttyACM999\n"
        "Waiting for upload port...\n"
        "Upload port found on /dev/ttyACM9990\n"
        '"{data_dir}/packages/arduino/tools/avrdude/6.3.0-arduino17/bin/avrdude" '
        '"-C{data_dir}/packages/arduino/tools/avrdude/6.3.0-arduino17/etc/avrdude.conf" '
        '-v -V -patmega32u4 -cavr109 "-P/dev/ttyACM9990" -b57600 -D '
        '"-Uflash:w:{build_dir}/{sketch_name}.ino.hex:i"\n',
    ),
    (
        "adafruit:avr:circuitplay32u4cat",
        "/dev/ttyACM0",
        "",
        "Performing 1200-bps touch reset on serial port /dev/ttyACM0\n"
        "Waiting for upload port...\n"
        "No upload port found, using /dev/ttyACM0 as fallback\n"
        '"{data_dir}/packages/arduino/tools/avrdude/6.3.0-arduino17/bin/avrdude" '
        '"-C{data_dir}/packages/arduino/tools/avrdude/6.3.0-arduino17/etc/avrdude.conf" '
        "-v -patmega32u4 -cavr109 -P/dev/ttyACM0 -b57600 -D "
        '"-Uflash:w:{build_dir}/{sketch_name}.ino.hex:i"\n',
    ),
    (
        "adafruit:avr:circuitplay32u4cat",
        "/dev/ttyACM999",
        "",
        "Performing 1200-bps touch reset on serial port /dev/ttyACM999\n"
        "Waiting for upload port...\n"
        "Upload port found on /dev/ttyACM9990\n"
        '"{data_dir}/packages/arduino/tools/avrdude/6.3.0-arduino17/bin/avrdude" '
        '"-C{data_dir}/packages/arduino/tools/avrdude/6.3.0-arduino17/etc/avrdude.conf" '
        "-v -patmega32u4 -cavr109 -P/dev/ttyACM9990 -b57600 -D "
        '"-Uflash:w:{build_dir}/{sketch_name}.ino.hex:i"\n',
    ),
    (
        "adafruit:avr:flora8",
        "/dev/ttyACM0",
        "",
        "Performing 1200-bps touch reset on serial port /dev/ttyACM0\n"
        "Waiting for upload port...\n"
        "No upload port found, using /dev/ttyACM0 as fallback\n"
        '"{data_dir}/packages/arduino/tools/avrdude/6.3.0-arduino17/bin/avrdude" '
        '"-C{data_dir}/packages/arduino/tools/avrdude/6.3.0-arduino17/etc/avrdude.conf" '
        "-v -patmega32u4 -cavr109 -P/dev/ttyACM0 -b57600 -D "
        '"-Uflash:w:{build_dir}/{sketch_name}.ino.hex:i"\n',
    ),
    (
        "adafruit:avr:flora8",
        "/dev/ttyACM999",
        "",
        "Performing 1200-bps touch reset on serial port /dev/ttyACM999\n"
        "Waiting for upload port...\n"
        "Upload port found on /dev/ttyACM9990\n"
        '"{data_dir}/packages/arduino/tools/avrdude/6.3.0-arduino17/bin/avrdude" '
        '"-C{data_dir}/packages/arduino/tools/avrdude/6.3.0-arduino17/etc/avrdude.conf" '
        "-v -patmega32u4 -cavr109 -P/dev/ttyACM9990 -b57600 -D "
        '"-Uflash:w:{build_dir}/{sketch_name}.ino.hex:i"\n',
    ),
    (
        "adafruit:avr:gemma",
        "/dev/ttyACM0",
        "usbGemma",
        '"{data_dir}/packages/arduino/tools/avrdude/6.3.0-arduino17/bin/avrdude" '
        '"-C{data_dir}/packages/adafruit/hardware/avr/1.4.13/bootloaders/gemma/avrdude.conf" '
        "-v -pattiny85 -carduinogemma  "
        '"-Uflash:w:{build_dir}/{sketch_name}.ino.hex:i"\n',
    ),
    (
        "adafruit:avr:gemma",
        "",
        "usbGemma",
        '"{data_dir}/packages/arduino/tools/avrdude/6.3.0-arduino17/bin/avrdude" '
        '"-C{data_dir}/packages/adafruit/hardware/avr/1.4.13/bootloaders/gemma/avrdude.conf" '
        "-v -pattiny85 -carduinogemma  "
        '"-Uflash:w:{build_dir}/{sketch_name}.ino.hex:i"\n',
    ),
    (
        "adafruit:avr:itsybitsy32u4_3V",
        "/dev/ttyACM0",
        "",
        "Performing 1200-bps touch reset on serial port /dev/ttyACM0\n"
        "Waiting for upload port...\n"
        "No upload port found, using /dev/ttyACM0 as fallback\n"
        '"{data_dir}/packages/arduino/tools/avrdude/6.3.0-arduino17/bin/avrdude" '
        '"-C{data_dir}/packages/arduino/tools/avrdude/6.3.0-arduino17/etc/avrdude.conf" '
        "-v -patmega32u4 -cavr109 -P/dev/ttyACM0 -b57600 -D "
        '"-Uflash:w:{build_dir}/{sketch_name}.ino.hex:i"\n',
    ),
    (
        "adafruit:avr:itsybitsy32u4_3V",
        "/dev/ttyACM999",
        "",
        "Performing 1200-bps touch reset on serial port /dev/ttyACM999\n"
        "Waiting for upload port...\n"
        "Upload port found on /dev/ttyACM9990\n"
        '"{data_dir}/packages/arduino/tools/avrdude/6.3.0-arduino17/bin/avrdude" '
        '"-C{data_dir}/packages/arduino/tools/avrdude/6.3.0-arduino17/etc/avrdude.conf" '
        "-v -patmega32u4 -cavr109 -P/dev/ttyACM9990 -b57600 -D "
        '"-Uflash:w:{build_dir}/{sketch_name}.ino.hex:i"\n',
    ),
    (
        "adafruit:avr:itsybitsy32u4_5V",
        "/dev/ttyACM0",
        "",
        "Performing 1200-bps touch reset on serial port /dev/ttyACM0\n"
        "Waiting for upload port...\n"
        "No upload port found, using /dev/ttyACM0 as fallback\n"
        '"{data_dir}/packages/arduino/tools/avrdude/6.3.0-arduino17/bin/avrdude" '
        '"-C{data_dir}/packages/arduino/tools/avrdude/6.3.0-arduino17/etc/avrdude.conf" '
        "-v -patmega32u4 -cavr109 -P/dev/ttyACM0 -b57600 -D "
        '"-Uflash:w:{build_dir}/{sketch_name}.ino.hex:i"\n',
    ),
    (
        "adafruit:avr:itsybitsy32u4_5V",
        "/dev/ttyACM999",
        "",
        "Performing 1200-bps touch reset on serial port /dev/ttyACM999\n"
        "Waiting for upload port...\n"
        "Upload port found on /dev/ttyACM9990\n"
        '"{data_dir}/packages/arduino/tools/avrdude/6.3.0-arduino17/bin/avrdude" '
        '"-C{data_dir}/packages/arduino/tools/avrdude/6.3.0-arduino17/etc/avrdude.conf" '
        "-v -patmega32u4 -cavr109 -P/dev/ttyACM9990 -b57600 -D "
        '"-Uflash:w:{build_dir}/{sketch_name}.ino.hex:i"\n',
    ),
    (
        "adafruit:avr:metro",
        "/dev/ttyACM0",
        "",
        '"{data_dir}/packages/arduino/tools/avrdude/6.3.0-arduino17/bin/avrdude" '
        '"-C{data_dir}/packages/arduino/tools/avrdude/6.3.0-arduino17/etc/avrdude.conf" '
        "-v -patmega328p -carduino -P/dev/ttyACM0 -b115200 -D "
        '"-Uflash:w:{build_dir}/{sketch_name}.ino.hex:i"\n',
    ),
    (
        "adafruit:avr:trinket3",
        "",
        "usbasp",
        '"{data_dir}/packages/arduino/tools/avrdude/6.3.0-arduino17/bin/avrdude" '
        '"-C{data_dir}/packages/arduino/tools/avrdude/6.3.0-arduino17/etc/avrdude.conf" '
        "-v -pattiny85 -cusbasp -Pusb "
        '"-Uflash:w:{build_dir}/{sketch_name}.ino.hex:i"\n',
    ),
    (
        "adafruit:avr:trinket3",
        "/dev/ttyACM0",
        "avrisp",
        '"{data_dir}/packages/arduino/tools/avrdude/6.3.0-arduino17/bin/avrdude" '
        '"-C{data_dir}/packages/arduino/tools/avrdude/6.3.0-arduino17/etc/avrdude.conf" '
        "-v -pattiny85 -cstk500v1 -P/dev/ttyACM0 "
        '"-Uflash:w:{build_dir}/{sketch_name}.ino.hex:i"\n',
    ),
    (
        "adafruit:avr:trinket3",
        "/dev/ttyACM0",
        "arduinoasisp",
        '"{data_dir}/packages/arduino/tools/avrdude/6.3.0-arduino17/bin/avrdude" '
        '"-C{data_dir}/packages/arduino/tools/avrdude/6.3.0-arduino17/etc/avrdude.conf" '
        "-v -pattiny85 -cstk500v1 -P/dev/ttyACM0 -b19200 "
        '"-Uflash:w:{build_dir}/{sketch_name}.ino.hex:i"\n',
    ),
    (
        "arduino:samd:arduino_zero_edbg",
        "",
        "",
        {
            "darwin": '"{data_dir}/packages/arduino/tools/openocd/0.10.0-arduino7/bin/openocd" '
            "-d2 -s "
            '"{data_dir}/packages/arduino/tools/openocd/0.10.0-arduino7/share/openocd/scripts/" '
            "-f "
            '"{data_dir}/packages/arduino/hardware/samd/1.8.11/variants/arduino_zero/openocd_scripts/arduino_zero.cfg" '
            '-c "telnet_port disabled; program '
            "{{{build_dir}/{sketch_name}.ino.bin}} verify reset 0x2000; "
            'shutdown"\n',
            "linux": '"{data_dir}/packages/arduino/tools/openocd/0.10.0-arduino7/bin/openocd" '
            "-d2 -s "
            '"{data_dir}/packages/arduino/tools/openocd/0.10.0-arduino7/share/openocd/scripts/" '
            "-f "
            '"{data_dir}/packages/arduino/hardware/samd/1.8.11/variants/arduino_zero/openocd_scripts/arduino_zero.cfg" '
            '-c "telnet_port disabled; program '
            "{{{build_dir}/{sketch_name}.ino.bin}} verify reset 0x2000; "
            'shutdown"\n',
            "win32": '"{data_dir}/packages/arduino/tools/openocd/0.10.0-arduino7/bin/openocd.exe" '
            "-d2 -s "
            '"{data_dir}/packages/arduino/tools/openocd/0.10.0-arduino7/share/openocd/scripts/" '
            "-f "
            '"{data_dir}/packages/arduino/hardware/samd/1.8.11/variants/arduino_zero/openocd_scripts/arduino_zero.cfg" '
            '-c "telnet_port disabled; program '
            "{{{build_dir}/{sketch_name}.ino.bin}} verify reset 0x2000; "
            'shutdown"\n',
        },
    ),
    (
        "arduino:samd:adafruit_circuitplayground_m0",
        "/dev/ttyACM0",
        "",
        {
            "darwin": "Performing 1200-bps touch reset on serial port "
            "/dev/ttyACM0\n"
            "Waiting for upload port...\n"
            "No upload port found, using /dev/ttyACM0 as fallback\n"
            '"{data_dir}/packages/arduino/tools/bossac/1.7.0-arduino3/bossac" '
            "-i -d --port=ttyACM0 -U true -i -e -w -v "
            '"{build_dir}/{sketch_name}.ino.bin" -R\n',
            "linux": "Performing 1200-bps touch reset on serial port /dev/ttyACM0\n"
            "Waiting for upload port...\n"
            "No upload port found, using /dev/ttyACM0 as fallback\n"
            '"{data_dir}/packages/arduino/tools/bossac/1.7.0-arduino3/bossac" '
            "-i -d --port=ttyACM0 -U true -i -e -w -v "
            '"{build_dir}/{sketch_name}.ino.bin" -R\n',
            "win32": "Performing 1200-bps touch reset on serial port /dev/ttyACM0\n"
            "Waiting for upload port...\n"
            "No upload port found, using /dev/ttyACM0 as fallback\n"
            '"{data_dir}/packages/arduino/tools/bossac/1.7.0-arduino3/bossac.exe" '
            "-i -d --port=ttyACM0 -U true -i -e -w -v "
            '"{build_dir}/{sketch_name}.ino.bin" -R\n',
        },
    ),
    (
        "arduino:samd:adafruit_circuitplayground_m0",
        "/dev/ttyACM999",
        "",
        {
            "darwin": "Performing 1200-bps touch reset on serial port "
            "/dev/ttyACM999\n"
            "Waiting for upload port...\n"
            "Upload port found on /dev/ttyACM9990\n"
            '"{data_dir}/packages/arduino/tools/bossac/1.7.0-arduino3/bossac" '
            "-i -d --port=ttyACM9990 -U true -i -e -w -v "
            '"{build_dir}/{sketch_name}.ino.bin" -R\n',
            "linux": "Performing 1200-bps touch reset on serial port "
            "/dev/ttyACM999\n"
            "Waiting for upload port...\n"
            "Upload port found on /dev/ttyACM9990\n"
            '"{data_dir}/packages/arduino/tools/bossac/1.7.0-arduino3/bossac" '
            "-i -d --port=ttyACM9990 -U true -i -e -w -v "
            '"{build_dir}/{sketch_name}.ino.bin" -R\n',
            "win32": "Performing 1200-bps touch reset on serial port "
            "/dev/ttyACM999\n"
            "Waiting for upload port...\n"
            "Upload port found on /dev/ttyACM9990\n"
            '"{data_dir}/packages/arduino/tools/bossac/1.7.0-arduino3/bossac.exe" '
            "-i -d --port=ttyACM9990 -U true -i -e -w -v "
            '"{build_dir}/{sketch_name}.ino.bin" -R\n',
        },
    ),
    (
        "arduino:samd:mkrfox1200",
        "/dev/ttyACM0",
        "",
        {
            "darwin": "Performing 1200-bps touch reset on serial port "
            "/dev/ttyACM0\n"
            "Waiting for upload port...\n"
            "No upload port found, using /dev/ttyACM0 as fallback\n"
            '"{data_dir}/packages/arduino/tools/bossac/1.7.0-arduino3/bossac" '
            "-i -d --port=ttyACM0 -U true -i -e -w -v "
            '"{build_dir}/{sketch_name}.ino.bin" -R\n',
            "linux": "Performing 1200-bps touch reset on serial port /dev/ttyACM0\n"
            "Waiting for upload port...\n"
            "No upload port found, using /dev/ttyACM0 as fallback\n"
            '"{data_dir}/packages/arduino/tools/bossac/1.7.0-arduino3/bossac" '
            "-i -d --port=ttyACM0 -U true -i -e -w -v "
            '"{build_dir}/{sketch_name}.ino.bin" -R\n',
            "win32": "Performing 1200-bps touch reset on serial port /dev/ttyACM0\n"
            "Waiting for upload port...\n"
            "No upload port found, using /dev/ttyACM0 as fallback\n"
            '"{data_dir}/packages/arduino/tools/bossac/1.7.0-arduino3/bossac.exe" '
            "-i -d --port=ttyACM0 -U true -i -e -w -v "
            '"{build_dir}/{sketch_name}.ino.bin" -R\n',
        },
    ),
    (
        "arduino:samd:mkrfox1200",
        "/dev/ttyACM999",
        "",
        {
            "darwin": "Performing 1200-bps touch reset on serial port "
            "/dev/ttyACM999\n"
            "Waiting for upload port...\n"
            "Upload port found on /dev/ttyACM9990\n"
            '"{data_dir}/packages/arduino/tools/bossac/1.7.0-arduino3/bossac" '
            "-i -d --port=ttyACM9990 -U true -i -e -w -v "
            '"{build_dir}/{sketch_name}.ino.bin" -R\n',
            "linux": "Performing 1200-bps touch reset on serial port "
            "/dev/ttyACM999\n"
            "Waiting for upload port...\n"
            "Upload port found on /dev/ttyACM9990\n"
            '"{data_dir}/packages/arduino/tools/bossac/1.7.0-arduino3/bossac" '
            "-i -d --port=ttyACM9990 -U true -i -e -w -v "
            '"{build_dir}/{sketch_name}.ino.bin" -R\n',
            "win32": "Performing 1200-bps touch reset on serial port "
            "/dev/ttyACM999\n"
            "Waiting for upload port...\n"
            "Upload port found on /dev/ttyACM9990\n"
            '"{data_dir}/packages/arduino/tools/bossac/1.7.0-arduino3/bossac.exe" '
            "-i -d --port=ttyACM9990 -U true -i -e -w -v "
            '"{build_dir}/{sketch_name}.ino.bin" -R\n',
        },
    ),
    (
        "arduino:samd:mkrgsm1400",
        "/dev/ttyACM0",
        "",
        {
            "darwin": "Performing 1200-bps touch reset on serial port "
            "/dev/ttyACM0\n"
            "Waiting for upload port...\n"
            "No upload port found, using /dev/ttyACM0 as fallback\n"
            '"{data_dir}/packages/arduino/tools/bossac/1.7.0-arduino3/bossac" '
            "-i -d --port=ttyACM0 -U true -i -e -w -v "
            '"{build_dir}/{sketch_name}.ino.bin" -R\n',
            "linux": "Performing 1200-bps touch reset on serial port /dev/ttyACM0\n"
            "Waiting for upload port...\n"
            "No upload port found, using /dev/ttyACM0 as fallback\n"
            '"{data_dir}/packages/arduino/tools/bossac/1.7.0-arduino3/bossac" '
            "-i -d --port=ttyACM0 -U true -i -e -w -v "
            '"{build_dir}/{sketch_name}.ino.bin" -R\n',
            "win32": "Performing 1200-bps touch reset on serial port /dev/ttyACM0\n"
            "Waiting for upload port...\n"
            "No upload port found, using /dev/ttyACM0 as fallback\n"
            '"{data_dir}/packages/arduino/tools/bossac/1.7.0-arduino3/bossac.exe" '
            "-i -d --port=ttyACM0 -U true -i -e -w -v "
            '"{build_dir}/{sketch_name}.ino.bin" -R\n',
        },
    ),
    (
        "arduino:samd:mkrgsm1400",
        "/dev/ttyACM999",
        "",
        {
            "darwin": "Performing 1200-bps touch reset on serial port "
            "/dev/ttyACM999\n"
            "Waiting for upload port...\n"
            "Upload port found on /dev/ttyACM9990\n"
            '"{data_dir}/packages/arduino/tools/bossac/1.7.0-arduino3/bossac" '
            "-i -d --port=ttyACM9990 -U true -i -e -w -v "
            '"{build_dir}/{sketch_name}.ino.bin" -R\n',
            "linux": "Performing 1200-bps touch reset on serial port "
            "/dev/ttyACM999\n"
            "Waiting for upload port...\n"
            "Upload port found on /dev/ttyACM9990\n"
            '"{data_dir}/packages/arduino/tools/bossac/1.7.0-arduino3/bossac" '
            "-i -d --port=ttyACM9990 -U true -i -e -w -v "
            '"{build_dir}/{sketch_name}.ino.bin" -R\n',
            "win32": "Performing 1200-bps touch reset on serial port "
            "/dev/ttyACM999\n"
            "Waiting for upload port...\n"
            "Upload port found on /dev/ttyACM9990\n"
            '"{data_dir}/packages/arduino/tools/bossac/1.7.0-arduino3/bossac.exe" '
            "-i -d --port=ttyACM9990 -U true -i -e -w -v "
            '"{build_dir}/{sketch_name}.ino.bin" -R\n',
        },
    ),
    (
        "arduino:samd:mkrvidor4000",
        "/dev/ttyACM0",
        "",
        {
            "darwin": "Performing 1200-bps touch reset on serial port "
            "/dev/ttyACM0\n"
            "Waiting for upload port...\n"
            "No upload port found, using /dev/ttyACM0 as fallback\n"
            '"{data_dir}/packages/arduino/tools/bossac/1.7.0-arduino3/bossac" '
            "-i -d --port=ttyACM0 -I -U true -i -e -w "
            '"{build_dir}/{sketch_name}.ino.bin" -R\n',
            "linux": "Performing 1200-bps touch reset on serial port /dev/ttyACM0\n"
            "Waiting for upload port...\n"
            "No upload port found, using /dev/ttyACM0 as fallback\n"
            '"{data_dir}/packages/arduino/tools/bossac/1.7.0-arduino3/bossac" '
            "-i -d --port=ttyACM0 -I -U true -i -e -w "
            '"{build_dir}/{sketch_name}.ino.bin" -R\n',
            "win32": "Performing 1200-bps touch reset on serial port /dev/ttyACM0\n"
            "Waiting for upload port...\n"
            "No upload port found, using /dev/ttyACM0 as fallback\n"
            '"{data_dir}/packages/arduino/tools/bossac/1.7.0-arduino3/bossac.exe" '
            "-i -d --port=ttyACM0 -I -U true -i -e -w "
            '"{build_dir}/{sketch_name}.ino.bin" -R\n',
        },
    ),
    (
        "arduino:samd:mkrvidor4000",
        "/dev/ttyACM999",
        "",
        {
            "darwin": "Performing 1200-bps touch reset on serial port "
            "/dev/ttyACM999\n"
            "Waiting for upload port...\n"
            "Upload port found on /dev/ttyACM9990\n"
            '"{data_dir}/packages/arduino/tools/bossac/1.7.0-arduino3/bossac" '
            "-i -d --port=ttyACM9990 -I -U true -i -e -w "
            '"{build_dir}/{sketch_name}.ino.bin" -R\n',
            "linux": "Performing 1200-bps touch reset on serial port "
            "/dev/ttyACM999\n"
            "Waiting for upload port...\n"
            "Upload port found on /dev/ttyACM9990\n"
            '"{data_dir}/packages/arduino/tools/bossac/1.7.0-arduino3/bossac" '
            "-i -d --port=ttyACM9990 -I -U true -i -e -w "
            '"{build_dir}/{sketch_name}.ino.bin" -R\n',
            "win32": "Performing 1200-bps touch reset on serial port "
            "/dev/ttyACM999\n"
            "Waiting for upload port...\n"
            "Upload port found on /dev/ttyACM9990\n"
            '"{data_dir}/packages/arduino/tools/bossac/1.7.0-arduino3/bossac.exe" '
            "-i -d --port=ttyACM9990 -I -U true -i -e -w "
            '"{build_dir}/{sketch_name}.ino.bin" -R\n',
        },
    ),
    (
        "arduino:samd:mkrwan1310",
        "/dev/ttyACM0",
        "",
        {
            "darwin": "Performing 1200-bps touch reset on serial port "
            "/dev/ttyACM0\n"
            "Waiting for upload port...\n"
            "No upload port found, using /dev/ttyACM0 as fallback\n"
            '"{data_dir}/packages/arduino/tools/bossac/1.7.0-arduino3/bossac" '
            "-i -d --port=ttyACM0 -U true -i -e -w -v "
            '"{build_dir}/{sketch_name}.ino.bin" -R\n',
            "linux": "Performing 1200-bps touch reset on serial port /dev/ttyACM0\n"
            "Waiting for upload port...\n"
            "No upload port found, using /dev/ttyACM0 as fallback\n"
            '"{data_dir}/packages/arduino/tools/bossac/1.7.0-arduino3/bossac" '
            "-i -d --port=ttyACM0 -U true -i -e -w -v "
            '"{build_dir}/{sketch_name}.ino.bin" -R\n',
            "win32": "Performing 1200-bps touch reset on serial port /dev/ttyACM0\n"
            "Waiting for upload port...\n"
            "No upload port found, using /dev/ttyACM0 as fallback\n"
            '"{data_dir}/packages/arduino/tools/bossac/1.7.0-arduino3/bossac.exe" '
            "-i -d --port=ttyACM0 -U true -i -e -w -v "
            '"{build_dir}/{sketch_name}.ino.bin" -R\n',
        },
    ),
    (
        "arduino:samd:mkrwan1310",
        "/dev/ttyACM999",
        "",
        {
            "darwin": "Performing 1200-bps touch reset on serial port "
            "/dev/ttyACM999\n"
            "Waiting for upload port...\n"
            "Upload port found on /dev/ttyACM9990\n"
            '"{data_dir}/packages/arduino/tools/bossac/1.7.0-arduino3/bossac" '
            "-i -d --port=ttyACM9990 -U true -i -e -w -v "
            '"{build_dir}/{sketch_name}.ino.bin" -R\n',
            "linux": "Performing 1200-bps touch reset on serial port "
            "/dev/ttyACM999\n"
            "Waiting for upload port...\n"
            "Upload port found on /dev/ttyACM9990\n"
            '"{data_dir}/packages/arduino/tools/bossac/1.7.0-arduino3/bossac" '
            "-i -d --port=ttyACM9990 -U true -i -e -w -v "
            '"{build_dir}/{sketch_name}.ino.bin" -R\n',
            "win32": "Performing 1200-bps touch reset on serial port "
            "/dev/ttyACM999\n"
            "Waiting for upload port...\n"
            "Upload port found on /dev/ttyACM9990\n"
            '"{data_dir}/packages/arduino/tools/bossac/1.7.0-arduino3/bossac.exe" '
            "-i -d --port=ttyACM9990 -U true -i -e -w -v "
            '"{build_dir}/{sketch_name}.ino.bin" -R\n',
        },
    ),
    (
        "arduino:samd:mkrwifi1010",
        "/dev/ttyACM0",
        "",
        {
            "darwin": "Performing 1200-bps touch reset on serial port "
            "/dev/ttyACM0\n"
            "Waiting for upload port...\n"
            "No upload port found, using /dev/ttyACM0 as fallback\n"
            '"{data_dir}/packages/arduino/tools/bossac/1.7.0-arduino3/bossac" '
            "-i -d --port=ttyACM0 -U true -i -e -w -v "
            '"{build_dir}/{sketch_name}.ino.bin" -R\n',
            "linux": "Performing 1200-bps touch reset on serial port /dev/ttyACM0\n"
            "Waiting for upload port...\n"
            "No upload port found, using /dev/ttyACM0 as fallback\n"
            '"{data_dir}/packages/arduino/tools/bossac/1.7.0-arduino3/bossac" '
            "-i -d --port=ttyACM0 -U true -i -e -w -v "
            '"{build_dir}/{sketch_name}.ino.bin" -R\n',
            "win32": "Performing 1200-bps touch reset on serial port /dev/ttyACM0\n"
            "Waiting for upload port...\n"
            "No upload port found, using /dev/ttyACM0 as fallback\n"
            '"{data_dir}/packages/arduino/tools/bossac/1.7.0-arduino3/bossac.exe" '
            "-i -d --port=ttyACM0 -U true -i -e -w -v "
            '"{build_dir}/{sketch_name}.ino.bin" -R\n',
        },
    ),
    (
        "arduino:samd:mkrwifi1010",
        "/dev/ttyACM999",
        "",
        {
            "darwin": "Performing 1200-bps touch reset on serial port "
            "/dev/ttyACM999\n"
            "Waiting for upload port...\n"
            "Upload port found on /dev/ttyACM9990\n"
            '"{data_dir}/packages/arduino/tools/bossac/1.7.0-arduino3/bossac" '
            "-i -d --port=ttyACM9990 -U true -i -e -w -v "
            '"{build_dir}/{sketch_name}.ino.bin" -R\n',
            "linux": "Performing 1200-bps touch reset on serial port "
            "/dev/ttyACM999\n"
            "Waiting for upload port...\n"
            "Upload port found on /dev/ttyACM9990\n"
            '"{data_dir}/packages/arduino/tools/bossac/1.7.0-arduino3/bossac" '
            "-i -d --port=ttyACM9990 -U true -i -e -w -v "
            '"{build_dir}/{sketch_name}.ino.bin" -R\n',
            "win32": "Performing 1200-bps touch reset on serial port "
            "/dev/ttyACM999\n"
            "Waiting for upload port...\n"
            "Upload port found on /dev/ttyACM9990\n"
            '"{data_dir}/packages/arduino/tools/bossac/1.7.0-arduino3/bossac.exe" '
            "-i -d --port=ttyACM9990 -U true -i -e -w -v "
            '"{build_dir}/{sketch_name}.ino.bin" -R\n',
        },
    ),
    (
        "arduino:samd:mkr1000",
        "/dev/ttyACM0",
        "",
        {
            "darwin": "Performing 1200-bps touch reset on serial port "
            "/dev/ttyACM0\n"
            "Waiting for upload port...\n"
            "No upload port found, using /dev/ttyACM0 as fallback\n"
            '"{data_dir}/packages/arduino/tools/bossac/1.7.0-arduino3/bossac" '
            "-i -d --port=ttyACM0 -U true -i -e -w -v "
            '"{build_dir}/{sketch_name}.ino.bin" -R\n',
            "linux": "Performing 1200-bps touch reset on serial port /dev/ttyACM0\n"
            "Waiting for upload port...\n"
            "No upload port found, using /dev/ttyACM0 as fallback\n"
            '"{data_dir}/packages/arduino/tools/bossac/1.7.0-arduino3/bossac" '
            "-i -d --port=ttyACM0 -U true -i -e -w -v "
            '"{build_dir}/{sketch_name}.ino.bin" -R\n',
            "win32": "Performing 1200-bps touch reset on serial port /dev/ttyACM0\n"
            "Waiting for upload port...\n"
            "No upload port found, using /dev/ttyACM0 as fallback\n"
            '"{data_dir}/packages/arduino/tools/bossac/1.7.0-arduino3/bossac.exe" '
            "-i -d --port=ttyACM0 -U true -i -e -w -v "
            '"{build_dir}/{sketch_name}.ino.bin" -R\n',
        },
    ),
    (
        "arduino:samd:mkr1000",
        "/dev/ttyACM999",
        "",
        {
            "darwin": "Performing 1200-bps touch reset on serial port "
            "/dev/ttyACM999\n"
            "Waiting for upload port...\n"
            "Upload port found on /dev/ttyACM9990\n"
            '"{data_dir}/packages/arduino/tools/bossac/1.7.0-arduino3/bossac" '
            "-i -d --port=ttyACM9990 -U true -i -e -w -v "
            '"{build_dir}/{sketch_name}.ino.bin" -R\n',
            "linux": "Performing 1200-bps touch reset on serial port "
            "/dev/ttyACM999\n"
            "Waiting for upload port...\n"
            "Upload port found on /dev/ttyACM9990\n"
            '"{data_dir}/packages/arduino/tools/bossac/1.7.0-arduino3/bossac" '
            "-i -d --port=ttyACM9990 -U true -i -e -w -v "
            '"{build_dir}/{sketch_name}.ino.bin" -R\n',
            "win32": "Performing 1200-bps touch reset on serial port "
            "/dev/ttyACM999\n"
            "Waiting for upload port...\n"
            "Upload port found on /dev/ttyACM9990\n"
            '"{data_dir}/packages/arduino/tools/bossac/1.7.0-arduino3/bossac.exe" '
            "-i -d --port=ttyACM9990 -U true -i -e -w -v "
            '"{build_dir}/{sketch_name}.ino.bin" -R\n',
        },
    ),
    (
        "arduino:samd:mkrzero",
        "/dev/ttyACM0",
        "",
        {
            "darwin": "Performing 1200-bps touch reset on serial port "
            "/dev/ttyACM0\n"
            "Waiting for upload port...\n"
            "No upload port found, using /dev/ttyACM0 as fallback\n"
            '"{data_dir}/packages/arduino/tools/bossac/1.7.0-arduino3/bossac" '
            "-i -d --port=ttyACM0 -U true -i -e -w -v "
            '"{build_dir}/{sketch_name}.ino.bin" -R\n',
            "linux": "Performing 1200-bps touch reset on serial port /dev/ttyACM0\n"
            "Waiting for upload port...\n"
            "No upload port found, using /dev/ttyACM0 as fallback\n"
            '"{data_dir}/packages/arduino/tools/bossac/1.7.0-arduino3/bossac" '
            "-i -d --port=ttyACM0 -U true -i -e -w -v "
            '"{build_dir}/{sketch_name}.ino.bin" -R\n',
            "win32": "Performing 1200-bps touch reset on serial port /dev/ttyACM0\n"
            "Waiting for upload port...\n"
            "No upload port found, using /dev/ttyACM0 as fallback\n"
            '"{data_dir}/packages/arduino/tools/bossac/1.7.0-arduino3/bossac.exe" '
            "-i -d --port=ttyACM0 -U true -i -e -w -v "
            '"{build_dir}/{sketch_name}.ino.bin" -R\n',
        },
    ),
    (
        "arduino:samd:mkrzero",
        "/dev/ttyACM999",
        "",
        {
            "darwin": "Performing 1200-bps touch reset on serial port "
            "/dev/ttyACM999\n"
            "Waiting for upload port...\n"
            "Upload port found on /dev/ttyACM9990\n"
            '"{data_dir}/packages/arduino/tools/bossac/1.7.0-arduino3/bossac" '
            "-i -d --port=ttyACM9990 -U true -i -e -w -v "
            '"{build_dir}/{sketch_name}.ino.bin" -R\n',
            "linux": "Performing 1200-bps touch reset on serial port "
            "/dev/ttyACM999\n"
            "Waiting for upload port...\n"
            "Upload port found on /dev/ttyACM9990\n"
            '"{data_dir}/packages/arduino/tools/bossac/1.7.0-arduino3/bossac" '
            "-i -d --port=ttyACM9990 -U true -i -e -w -v "
            '"{build_dir}/{sketch_name}.ino.bin" -R\n',
            "win32": "Performing 1200-bps touch reset on serial port "
            "/dev/ttyACM999\n"
            "Waiting for upload port...\n"
            "Upload port found on /dev/ttyACM9990\n"
            '"{data_dir}/packages/arduino/tools/bossac/1.7.0-arduino3/bossac.exe" '
            "-i -d --port=ttyACM9990 -U true -i -e -w -v "
            '"{build_dir}/{sketch_name}.ino.bin" -R\n',
        },
    ),
    (
        "arduino:samd:nano_33_iot",
        "/dev/ttyACM0",
        "",
        {
            "darwin": "Performing 1200-bps touch reset on serial port "
            "/dev/ttyACM0\n"
            "Waiting for upload port...\n"
            "No upload port found, using /dev/ttyACM0 as fallback\n"
            '"{data_dir}/packages/arduino/tools/bossac/1.7.0-arduino3/bossac" '
            "-i -d --port=ttyACM0 -U true -i -e -w -v "
            '"{build_dir}/{sketch_name}.ino.bin" -R\n',
            "linux": "Performing 1200-bps touch reset on serial port /dev/ttyACM0\n"
            "Waiting for upload port...\n"
            "No upload port found, using /dev/ttyACM0 as fallback\n"
            '"{data_dir}/packages/arduino/tools/bossac/1.7.0-arduino3/bossac" '
            "-i -d --port=ttyACM0 -U true -i -e -w -v "
            '"{build_dir}/{sketch_name}.ino.bin" -R\n',
            "win32": "Performing 1200-bps touch reset on serial port /dev/ttyACM0\n"
            "Waiting for upload port...\n"
            "No upload port found, using /dev/ttyACM0 as fallback\n"
            '"{data_dir}/packages/arduino/tools/bossac/1.7.0-arduino3/bossac.exe" '
            "-i -d --port=ttyACM0 -U true -i -e -w -v "
            '"{build_dir}/{sketch_name}.ino.bin" -R\n',
        },
    ),
    (
        "arduino:samd:nano_33_iot",
        "/dev/ttyACM999",
        "",
        {
            "darwin": "Performing 1200-bps touch reset on serial port "
            "/dev/ttyACM999\n"
            "Waiting for upload port...\n"
            "Upload port found on /dev/ttyACM9990\n"
            '"{data_dir}/packages/arduino/tools/bossac/1.7.0-arduino3/bossac" '
            "-i -d --port=ttyACM9990 -U true -i -e -w -v "
            '"{build_dir}/{sketch_name}.ino.bin" -R\n',
            "linux": "Performing 1200-bps touch reset on serial port "
            "/dev/ttyACM999\n"
            "Waiting for upload port...\n"
            "Upload port found on /dev/ttyACM9990\n"
            '"{data_dir}/packages/arduino/tools/bossac/1.7.0-arduino3/bossac" '
            "-i -d --port=ttyACM9990 -U true -i -e -w -v "
            '"{build_dir}/{sketch_name}.ino.bin" -R\n',
            "win32": "Performing 1200-bps touch reset on serial port "
            "/dev/ttyACM999\n"
            "Waiting for upload port...\n"
            "Upload port found on /dev/ttyACM9990\n"
            '"{data_dir}/packages/arduino/tools/bossac/1.7.0-arduino3/bossac.exe" '
            "-i -d --port=ttyACM9990 -U true -i -e -w -v "
            '"{build_dir}/{sketch_name}.ino.bin" -R\n',
        },
    ),
    (
        "arduino:samd:arduino_zero_native",
        "/dev/ttyACM0",
        "",
        {
            "darwin": "Performing 1200-bps touch reset on serial port "
            "/dev/ttyACM0\n"
            "Waiting for upload port...\n"
            "No upload port found, using /dev/ttyACM0 as fallback\n"
            '"{data_dir}/packages/arduino/tools/bossac/1.7.0-arduino3/bossac" '
            "-i -d --port=ttyACM0 -U true -i -e -w -v "
            '"{build_dir}/{sketch_name}.ino.bin" -R\n',
            "linux": "Performing 1200-bps touch reset on serial port /dev/ttyACM0\n"
            "Waiting for upload port...\n"
            "No upload port found, using /dev/ttyACM0 as fallback\n"
            '"{data_dir}/packages/arduino/tools/bossac/1.7.0-arduino3/bossac" '
            "-i -d --port=ttyACM0 -U true -i -e -w -v "
            '"{build_dir}/{sketch_name}.ino.bin" -R\n',
            "win32": "Performing 1200-bps touch reset on serial port /dev/ttyACM0\n"
            "Waiting for upload port...\n"
            "No upload port found, using /dev/ttyACM0 as fallback\n"
            '"{data_dir}/packages/arduino/tools/bossac/1.7.0-arduino3/bossac.exe" '
            "-i -d --port=ttyACM0 -U true -i -e -w -v "
            '"{build_dir}/{sketch_name}.ino.bin" -R\n',
        },
    ),
    (
        "arduino:samd:arduino_zero_native",
        "/dev/ttyACM999",
        "",
        {
            "darwin": "Performing 1200-bps touch reset on serial port "
            "/dev/ttyACM999\n"
            "Waiting for upload port...\n"
            "Upload port found on /dev/ttyACM9990\n"
            '"{data_dir}/packages/arduino/tools/bossac/1.7.0-arduino3/bossac" '
            "-i -d --port=ttyACM9990 -U true -i -e -w -v "
            '"{build_dir}/{sketch_name}.ino.bin" -R\n',
            "linux": "Performing 1200-bps touch reset on serial port "
            "/dev/ttyACM999\n"
            "Waiting for upload port...\n"
            "Upload port found on /dev/ttyACM9990\n"
            '"{data_dir}/packages/arduino/tools/bossac/1.7.0-arduino3/bossac" '
            "-i -d --port=ttyACM9990 -U true -i -e -w -v "
            '"{build_dir}/{sketch_name}.ino.bin" -R\n',
            "win32": "Performing 1200-bps touch reset on serial port "
            "/dev/ttyACM999\n"
            "Waiting for upload port...\n"
            "Upload port found on /dev/ttyACM9990\n"
            '"{data_dir}/packages/arduino/tools/bossac/1.7.0-arduino3/bossac.exe" '
            "-i -d --port=ttyACM9990 -U true -i -e -w -v "
            '"{build_dir}/{sketch_name}.ino.bin" -R\n',
        },
    ),
    (
        "esp32:esp32:esp32",
        "/dev/ttyACM0",
        "",
        {
            "darwin": '"{data_dir}/packages/esp32/tools/esptool_py/3.0.0/esptool" '
            '--chip esp32 --port "/dev/ttyACM0" --baud 921600  --before '
            "default_reset --after hard_reset write_flash -z "
            "--flash_mode dio --flash_freq 80m --flash_size detect "
            "0xe000 "
            '"{data_dir}/packages/esp32/hardware/esp32/1.0.6/tools/partitions/boot_app0.bin" '
            "0x1000 "
            '"{data_dir}/packages/esp32/hardware/esp32/1.0.6/tools/sdk/bin/bootloader_qio_80m.bin" '
            '0x10000 "{build_dir}/{sketch_name}.ino.bin" 0x8000 '
            '"{build_dir}/{sketch_name}.ino.partitions.bin"\n',
            "linux": "python "
            '"{data_dir}/packages/esp32/tools/esptool_py/3.0.0/esptool.py" '
            '--chip esp32 --port "/dev/ttyACM0" --baud 921600  --before '
            "default_reset --after hard_reset write_flash -z --flash_mode "
            "dio --flash_freq 80m --flash_size detect 0xe000 "
            '"{data_dir}/packages/esp32/hardware/esp32/1.0.6/tools/partitions/boot_app0.bin" '
            "0x1000 "
            '"{data_dir}/packages/esp32/hardware/esp32/1.0.6/tools/sdk/bin/bootloader_qio_80m.bin" '
            '0x10000 "{build_dir}/{sketch_name}.ino.bin" 0x8000 '
            '"{build_dir}/{sketch_name}.ino.partitions.bin"\n',
            "win32": '"{data_dir}/packages/esp32/tools/esptool_py/3.0.0/esptool.exe" '
            '--chip esp32 --port "/dev/ttyACM0" --baud 921600  --before '
            "default_reset --after hard_reset write_flash -z --flash_mode "
            "dio --flash_freq 80m --flash_size detect 0xe000 "
            '"{data_dir}/packages/esp32/hardware/esp32/1.0.6/tools/partitions/boot_app0.bin" '
            "0x1000 "
            '"{data_dir}/packages/esp32/hardware/esp32/1.0.6/tools/sdk/bin/bootloader_qio_80m.bin" '
            '0x10000 "{build_dir}/{sketch_name}.ino.bin" 0x8000 '
            '"{build_dir}/{sketch_name}.ino.partitions.bin"\n',
        },
    ),
    (
        "esp32:esp32:esp32:PSRAM=enabled,PartitionScheme=no_ota,CPUFreq=80,FlashMode=dio,FlashFreq=40,FlashSize=8M,UploadSpeed=230400,DebugLevel=info",
        "/dev/ttyACM0",
        "",
        {
            "darwin": '"{data_dir}/packages/esp32/tools/esptool_py/3.0.0/esptool" '
            '--chip esp32 --port "/dev/ttyACM0" --baud 230400  --before '
            "default_reset --after hard_reset write_flash -z "
            "--flash_mode dio --flash_freq 40m --flash_size detect "
            "0xe000 "
            '"{data_dir}/packages/esp32/hardware/esp32/1.0.6/tools/partitions/boot_app0.bin" '
            "0x1000 "
            '"{data_dir}/packages/esp32/hardware/esp32/1.0.6/tools/sdk/bin/bootloader_dio_40m.bin" '
            '0x10000 "{build_dir}/{sketch_name}.ino.bin" 0x8000 '
            '"{build_dir}/{sketch_name}.ino.partitions.bin"\n',
            "linux": "python "
            '"{data_dir}/packages/esp32/tools/esptool_py/3.0.0/esptool.py" '
            '--chip esp32 --port "/dev/ttyACM0" --baud 230400  --before '
            "default_reset --after hard_reset write_flash -z --flash_mode "
            "dio --flash_freq 40m --flash_size detect 0xe000 "
            '"{data_dir}/packages/esp32/hardware/esp32/1.0.6/tools/partitions/boot_app0.bin" '
            "0x1000 "
            '"{data_dir}/packages/esp32/hardware/esp32/1.0.6/tools/sdk/bin/bootloader_dio_40m.bin" '
            '0x10000 "{build_dir}/{sketch_name}.ino.bin" 0x8000 '
            '"{build_dir}/{sketch_name}.ino.partitions.bin"\n',
            "win32": '"{data_dir}/packages/esp32/tools/esptool_py/3.0.0/esptool.exe" '
            '--chip esp32 --port "/dev/ttyACM0" --baud 230400  --before '
            "default_reset --after hard_reset write_flash -z --flash_mode "
            "dio --flash_freq 40m --flash_size detect 0xe000 "
            '"{data_dir}/packages/esp32/hardware/esp32/1.0.6/tools/partitions/boot_app0.bin" '
            "0x1000 "
            '"{data_dir}/packages/esp32/hardware/esp32/1.0.6/tools/sdk/bin/bootloader_dio_40m.bin" '
            '0x10000 "{build_dir}/{sketch_name}.ino.bin" 0x8000 '
            '"{build_dir}/{sketch_name}.ino.partitions.bin"\n',
        },
    ),
    (
        "esp32:esp32:esp32thing",
        "/dev/ttyACM0",
        "",
        {
            "darwin": '"{data_dir}/packages/esp32/tools/esptool_py/3.0.0/esptool" '
            '--chip esp32 --port "/dev/ttyACM0" --baud 921600  --before '
            "default_reset --after hard_reset write_flash -z "
            "--flash_mode dio --flash_freq 80m --flash_size detect "
            "0xe000 "
            '"{data_dir}/packages/esp32/hardware/esp32/1.0.6/tools/partitions/boot_app0.bin" '
            "0x1000 "
            '"{data_dir}/packages/esp32/hardware/esp32/1.0.6/tools/sdk/bin/bootloader_dio_80m.bin" '
            '0x10000 "{build_dir}/{sketch_name}.ino.bin" 0x8000 '
            '"{build_dir}/{sketch_name}.ino.partitions.bin"\n',
            "linux": "python "
            '"{data_dir}/packages/esp32/tools/esptool_py/3.0.0/esptool.py" '
            '--chip esp32 --port "/dev/ttyACM0" --baud 921600  --before '
            "default_reset --after hard_reset write_flash -z --flash_mode "
            "dio --flash_freq 80m --flash_size detect 0xe000 "
            '"{data_dir}/packages/esp32/hardware/esp32/1.0.6/tools/partitions/boot_app0.bin" '
            "0x1000 "
            '"{data_dir}/packages/esp32/hardware/esp32/1.0.6/tools/sdk/bin/bootloader_dio_80m.bin" '
            '0x10000 "{build_dir}/{sketch_name}.ino.bin" 0x8000 '
            '"{build_dir}/{sketch_name}.ino.partitions.bin"\n',
            "win32": '"{data_dir}/packages/esp32/tools/esptool_py/3.0.0/esptool.exe" '
            '--chip esp32 --port "/dev/ttyACM0" --baud 921600  --before '
            "default_reset --after hard_reset write_flash -z --flash_mode "
            "dio --flash_freq 80m --flash_size detect 0xe000 "
            '"{data_dir}/packages/esp32/hardware/esp32/1.0.6/tools/partitions/boot_app0.bin" '
            "0x1000 "
            '"{data_dir}/packages/esp32/hardware/esp32/1.0.6/tools/sdk/bin/bootloader_dio_80m.bin" '
            '0x10000 "{build_dir}/{sketch_name}.ino.bin" 0x8000 '
            '"{build_dir}/{sketch_name}.ino.partitions.bin"\n',
        },
    ),
    (
        "esp8266:esp8266:generic",
        "/dev/ttyACM0",
        "",
        '"{data_dir}/packages/esp8266/tools/python3/3.7.2-post1/python3" -I '
        '"{data_dir}/packages/esp8266/hardware/esp8266/3.0.2/tools/upload.py" '
        '--chip esp8266 --port "/dev/ttyACM0" --baud "115200" ""  --before '
        "default_reset --after hard_reset write_flash 0x0 "
        '"{build_dir}/{sketch_name}.ino.bin"\n',
    ),
    (
        "esp8266:esp8266:generic:xtal=160,vt=heap,mmu=3216,ResetMethod=nodtr_nosync,CrystalFreq=40,FlashFreq=20,eesz=2M,baud=57600",
        "/dev/ttyACM0",
        "",
        '"{data_dir}/packages/esp8266/tools/python3/3.7.2-post1/python3" -I '
        '"{data_dir}/packages/esp8266/hardware/esp8266/3.0.2/tools/upload.py" '
        '--chip esp8266 --port "/dev/ttyACM0" --baud "57600" ""  --before '
        "no_reset_no_sync --after soft_reset write_flash 0x0 "
        '"{build_dir}/{sketch_name}.ino.bin"\n',
    ),
    (
        "SPRESENSE:spresense:spresense",
        "/dev/ttyACM0",
        "",
        {
            "darwin": '"{data_dir}/packages/SPRESENSE/tools/spresense-tools/2.0.2/flash_writer/macosx/flash_writer" '
            '-s -c "/dev/ttyACM0"  -d -n "{build_dir}/{sketch_name}.ino.spk"',
            "linux": '"{data_dir}/packages/SPRESENSE/tools/spresense-tools/2.0.2/flash_writer/linux/flash_writer" '
            '-s -c "/dev/ttyACM0"  -d -n "{build_dir}/{sketch_name}.ino.spk"',
            "win32": '"{data_dir}/packages/SPRESENSE/tools/spresense-tools/2.0.2/flash_writer/windows/flash_writer.exe" '
            '-s -c "/dev/ttyACM0"  -d -n "{build_dir}/{sketch_name}.ino.spk"',
        },
    ),
]


@pytest.mark.parametrize("fqbn, upload_port, programmer, output", testdata)
def test_upload_sketch(
    run_command,
    session_data_dir,
    downloads_dir,
    fqbn,
    upload_port,
    programmer,
    output,
):
    env = {
        "ARDUINO_DATA_DIR": session_data_dir,
        "ARDUINO_DOWNLOADS_DIR": downloads_dir,
        "ARDUINO_SKETCHBOOK_DIR": session_data_dir,
    }

    # Install everything just once
    if not Path(session_data_dir, "packages").is_dir():
        assert run_command(["config", "init", "--overwrite"], custom_env=env)
        for package_index in indexes:
            assert run_command(["config", "add", "board_manager.additional_urls", package_index], custom_env=env)
        assert run_command(["update"], custom_env=env)

        for d in cores_to_install:
            assert run_command(["core", "install", d], custom_env=env)

    # Create a sketch
    sketch_name = "TestSketchForUpload"
    sketch_path = Path(session_data_dir, sketch_name)
    assert run_command(["sketch", "new", sketch_path], custom_env=env)

    # Fake compilation, we just need the folder to exist
    build_dir = generate_build_dir(sketch_path)
    programmer_arg = ["-P", programmer] if programmer else []
    port_arg = ["-p", upload_port] if upload_port else []
    res = run_command(
        ["upload"] + port_arg + programmer_arg + ["-b", fqbn, sketch_path, "--dry-run", "-v"], custom_env=env
    )
    assert res.ok

    if isinstance(output, str):
        out = output
    else:
        out = output[sys.platform]
    expected_output = out.format(
        data_dir=session_data_dir,
        upload_port=upload_port,
        build_dir=build_dir,
        sketch_name=sketch_name,
    ).replace("\\", "/")

    expected_output in res.stdout.replace("\\", "/").replace("\r", "")
