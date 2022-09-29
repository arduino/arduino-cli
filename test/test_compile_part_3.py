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

import shutil
from git import Repo
from pathlib import Path


# def test_compile_with_fully_precompiled_library(run_command, data_dir):
#    assert run_command(["update"])
#
#    assert run_command(["core", "install", "arduino:mbed@1.3.1"])
#    fqbn = "arduino:mbed:nano33ble"
#
#    # Install fully precompiled library
#    # For more information see:
#    # https://arduino.github.io/arduino-cli/latest/library-specification/#precompiled-binaries
#    assert run_command(["lib", "install", "Arduino_TensorFlowLite@2.1.1-ALPHA-precompiled"])
#    sketch_folder = Path(data_dir, "libraries", "Arduino_TensorFlowLite", "examples", "hello_world")
#
#    # Install example dependency
#    # assert run_command("lib install Arduino_LSM9DS1")#
#
#    # Compile and verify dependencies detection for fully precompiled library is skipped
#    result = run_command(["compile", "-b", fqbn, sketch_folder, "-v"])
#    assert result.ok
#    assert "Skipping dependencies detection for precompiled library Arduino_TensorFlowLite" in result.stdout
