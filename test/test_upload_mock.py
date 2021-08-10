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
from typing import Union


def generate_build_dir(sketch_path):
    sketch_path_md5 = hashlib.md5(bytes(sketch_path)).hexdigest().upper()
    build_dir = Path(tempfile.gettempdir(), f"arduino-sketch-{sketch_path_md5}")
    build_dir.mkdir(parents=True, exist_ok=True)
    return build_dir.resolve()


def generate_expected_output(
    output: str, upload_tools: Union[dict, str], data_dir: str, upload_port: str, build_dir: str, sketch_name: str
) -> str:
    if isinstance(upload_tools, str):
        tool = upload_tools
    else:
        tool = upload_tools[sys.platform]
    return output.format(
        tool_executable=tool, data_dir=data_dir, upload_port=upload_port, build_dir=build_dir, sketch_name=sketch_name,
    ).replace("\\", "/")


testdata = [
    (
        "",
        "arduino:avr:uno",
        "arduino:avr@1.8.3",
        [],
        "/dev/ttyACM0",
        '"{data_dir}/packages/arduino/tools/avrdude/6.3.0-arduino17/bin/avrdude"',
        "{tool_executable} "
        + '"-C{data_dir}/packages/arduino/tools/avrdude/6.3.0-arduino17/etc/avrdude.conf" '
        + '-v -V -patmega328p -carduino "-P{upload_port}" -b115200 -D "-Uflash:w:{build_dir}/{sketch_name}.ino.hex:i"',
    ),
    (
        "",
        "arduino:avr:leonardo",
        "arduino:avr@1.8.3",
        [],
        "/dev/ttyACM999",
        '"{data_dir}/packages/arduino/tools/avrdude/6.3.0-arduino17/bin/avrdude"',
        "{tool_executable} "
        + '"-C{data_dir}/packages/arduino/tools/avrdude/6.3.0-arduino17/etc/avrdude.conf" '
        + '-v -V -patmega32u4 -cavr109 "-P{upload_port}0" -b57600 -D "-Uflash:w:{build_dir}/{sketch_name}.ino.hex:i"',
    ),
    (
        "https://adafruit.github.io/arduino-board-index/package_adafruit_index.json",
        "adafruit:avr:flora8",
        "adafruit:avr@1.4.13",
        ["arduino:avr@1.8.3"],
        "/dev/ttyACM0",
        '"{data_dir}/packages/arduino/tools/avrdude/6.3.0-arduino17/bin/avrdude"',
        "{tool_executable} "
        + '"-C{data_dir}/packages/arduino/tools/avrdude/6.3.0-arduino17/etc/avrdude.conf" '
        + '-v -patmega32u4 -cavr109 -P{upload_port} -b57600 -D "-Uflash:w:{build_dir}/{sketch_name}.ino.hex:i"',
    ),
    (
        "https://adafruit.github.io/arduino-board-index/package_adafruit_index.json",
        "adafruit:avr:flora8",
        "adafruit:avr@1.4.13",
        ["arduino:avr@1.8.3"],
        "/dev/ttyACM999",
        '"{data_dir}/packages/arduino/tools/avrdude/6.3.0-arduino17/bin/avrdude"',
        "{tool_executable} "
        + '"-C{data_dir}/packages/arduino/tools/avrdude/6.3.0-arduino17/etc/avrdude.conf" '
        + '-v -patmega32u4 -cavr109 -P{upload_port}0 -b57600 -D "-Uflash:w:{build_dir}/{sketch_name}.ino.hex:i"',
    ),
    (
        "https://dl.espressif.com/dl/package_esp32_index.json",
        "esp32:esp32:esp32thing",
        "esp32:esp32@1.0.6",
        [],
        "/dev/ttyACM0",
        {
            "linux": 'python "{data_dir}/packages/esp32/tools/esptool_py/3.0.0/esptool.py"',
            "darwin": '"{data_dir}/packages/esp32/tools/esptool_py/3.0.0/esptool.py"',
            "win32": '"{data_dir}/packages/esp32/tools/esptool_py/3.0.0/esptool.exe"',
        },
        "{tool_executable} "
        + '--chip esp32 --port "{upload_port}" --baud 921600  --before default_reset '
        + "--after hard_reset write_flash -z --flash_mode dio --flash_freq 80m --flash_size detect 0xe000 "
        + '"{data_dir}/packages/esp32/hardware/esp32/1.0.6/tools/partitions/boot_app0.bin" 0x1000 '
        + '"{data_dir}/packages/esp32/hardware/esp32/1.0.6/tools/sdk/bin/bootloader_dio_80m.bin" 0x10000 '
        + '"{build_dir}/{sketch_name}.ino.bin" 0x8000 "{build_dir}/{sketch_name}.ino.partitions.bin"',
    ),
    (
        "http://arduino.esp8266.com/stable/package_esp8266com_index.json",
        "esp8266:esp8266:generic",
        "esp8266:esp8266@3.0.1",
        [],
        "/dev/ttyACM0",
        '"{data_dir}/packages/esp8266/tools/python3/3.7.2-post1/python3"',
        "{tool_executable} "
        + '"{data_dir}/packages/esp8266/hardware/esp8266/3.0.1/tools/upload.py" '
        + '--chip esp8266 --port "{upload_port}" --baud "115200" ""  '
        + "--before default_reset --after hard_reset write_flash 0x0 "
        + '"{build_dir}/{sketch_name}.ino.bin"',
    ),
]


@pytest.mark.parametrize("package_index, fqbn, core, deps, upload_port, upload_tools, output", testdata)
def test_upload_sketch(
    run_command,
    session_data_dir,
    downloads_dir,
    package_index,
    fqbn,
    core,
    core_dependencies,
    upload_port,
    upload_tools,
    output,
):
    env = {
        "ARDUINO_DATA_DIR": session_data_dir,
        "ARDUINO_DOWNLOADS_DIR": downloads_dir,
        "ARDUINO_SKETCHBOOK_DIR": session_data_dir,
    }

    if package_index:
        assert run_command("config init --overwrite", custom_env=env)
        assert run_command(f"config add board_manager.additional_urls {package_index}", custom_env=env)
        assert run_command("update", custom_env=env)

    assert run_command(f"core install {core}", custom_env=env)

    for d in core_dependencies:
        assert run_command(f"core install {d}", custom_env=env)

    # Create a sketch
    sketch_name = "TestSketchForUpload"
    sketch_path = Path(session_data_dir, sketch_name)
    assert run_command(f'sketch new "{sketch_path}"', custom_env=env)

    # Fake compilation, we just need the folder to exist
    build_dir = generate_build_dir(sketch_path)

    res = run_command(f'upload -p {upload_port} -b {fqbn} "{sketch_path}" --dry-run -v', custom_env=env)
    assert res.ok

    generate_expected_output(
        output=output,
        upload_tools=upload_tools,
        data_dir=session_data_dir,
        upload_port=upload_port,
        build_dir=build_dir,
        sketch_name=sketch_name,
    ) in res.stdout.replace("\\", "/")
