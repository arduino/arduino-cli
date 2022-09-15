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
from pathlib import Path
import simplejson as json


def test_compile_with_custom_libraries(run_command, copy_sketch):
    # Creates config with additional URL to install necessary core
    url = "http://arduino.esp8266.com/stable/package_esp8266com_index.json"
    assert run_command(["config", "init", "--dest-dir", ".", "--additional-urls", url])

    # Init the environment explicitly
    assert run_command(["update"])

    # Install core to compile
    assert run_command(["core", "install", "esp8266:esp8266"])

    sketch_path = copy_sketch("sketch_with_multiple_custom_libraries")
    fqbn = "esp8266:esp8266:nodemcu:xtal=80,vt=heap,eesz=4M1M,wipe=none,baud=115200"

    first_lib = Path(sketch_path, "libraries1")
    second_lib = Path(sketch_path, "libraries2")
    # This compile command has been taken from this issue:
    # https://github.com/arduino/arduino-cli/issues/973
    assert run_command(["compile", "--libraries", first_lib, "--libraries", second_lib, "-b", fqbn, sketch_path])


def test_compile_with_archives_and_long_paths(run_command):
    # Creates config with additional URL to install necessary core
    url = "http://arduino.esp8266.com/stable/package_esp8266com_index.json"
    assert run_command(["config", "init", "--dest-dir", ".", "--additional-urls", url])

    # Init the environment explicitly
    assert run_command(["update"])

    # Install core to compile
    assert run_command(["core", "install", "esp8266:esp8266@2.7.4"])

    # Install test library
    assert run_command(["lib", "install", "ArduinoIoTCloud"])

    result = run_command(["lib", "examples", "ArduinoIoTCloud", "--format", "json"])
    assert result.ok
    lib_output = json.loads(result.stdout)
    sketch_path = Path(lib_output[0]["library"]["install_dir"], "examples", "ArduinoIoTCloud-Advanced")

    assert run_command(["compile", "-b", "esp8266:esp8266:huzzah", sketch_path])


def test_compile_with_precompiled_library(run_command, data_dir):
    assert run_command(["update"])

    assert run_command(["core", "install", "arduino:samd@1.8.11"])
    fqbn = "arduino:samd:mkrzero"

    # Install precompiled library
    # For more information see:
    # https://arduino.github.io/arduino-cli/latest/library-specification/#precompiled-binaries
    assert run_command(["lib", "install", "BSEC Software Library@1.5.1474"])
    sketch_folder = Path(data_dir, "libraries", "BSEC_Software_Library", "examples", "basic")

    # Compile and verify dependencies detection for fully precompiled library is not skipped
    result = run_command(["compile", "-b", fqbn, sketch_folder, "-v"])
    assert result.ok
    assert "Skipping dependencies detection for precompiled library BSEC Software Library" not in result.stdout
