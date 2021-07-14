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
import hashlib
import pytest
from pathlib import Path


def generate_build_dir(sketch_path):
    sketch_path_md5 = hashlib.md5(bytes(sketch_path)).hexdigest().upper()
    build_dir = Path(tempfile.gettempdir(), f"arduino-sketch-{sketch_path_md5}")
    build_dir.mkdir(parents=True, exist_ok=True)
    return build_dir.resolve()


testdata = [
    (
        "arduino:avr:uno",
        "arduino:avr@1.8.3",
        "/dev/ttyACM0",
        '"{data_dir}/packages/arduino/tools/avrdude/6.3.0-arduino17/bin/avrdude" '
        + '"-C{data_dir}/packages/arduino/tools/avrdude/6.3.0-arduino17/etc/avrdude.conf" '
        + '-v -V -patmega328p -carduino "-P{upload_port}" -b115200 -D "-Uflash:w:{build_dir}/{sketch_name}.ino.hex:i"',
    ),
    (
        "arduino:avr:leonardo",
        "arduino:avr@1.8.3",
        "/dev/ttyACM999",
        '"{data_dir}/packages/arduino/tools/avrdude/6.3.0-arduino17/bin/avrdude" '
        + '"-C{data_dir}/packages/arduino/tools/avrdude/6.3.0-arduino17/etc/avrdude.conf" '
        + '-v -V -patmega32u4 -cavr109 "-P{upload_port}0" -b57600 -D "-Uflash:w:{build_dir}/{sketch_name}.ino.hex:i"',
    ),
]


@pytest.mark.parametrize("fqbn, core, upload_port, expected_output", testdata)
def test_upload_sketch(run_command, session_data_dir, downloads_dir, fqbn, core, upload_port, expected_output):
    env = {
        "ARDUINO_DATA_DIR": session_data_dir,
        "ARDUINO_DOWNLOADS_DIR": downloads_dir,
        "ARDUINO_SKETCHBOOK_DIR": session_data_dir,
    }
    assert run_command(f"core install {core}", custom_env=env)

    # Create a sketch
    sketch_name = "TestSketchForUpload"
    sketch_path = Path(session_data_dir, sketch_name)
    assert run_command(f'sketch new "{sketch_path}"', custom_env=env)

    # Fake compilation, we just need the folder to exist
    build_dir = generate_build_dir(sketch_path)

    res = run_command(f'upload -p {upload_port} -b {fqbn} "{sketch_path}" --dry-run -v', custom_env=env)
    assert res.ok

    assert expected_output.format(
        data_dir=session_data_dir, upload_port=upload_port, build_dir=build_dir, sketch_name=sketch_name
    ).replace("\\", "/") in res.stdout.replace("\\", "/")
