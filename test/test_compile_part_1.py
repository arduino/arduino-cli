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

import os
import platform
import tempfile
import hashlib
from pathlib import Path
import simplejson as json

import pytest

from .common import running_on_ci


def test_compile_with_multiple_build_property_flags(run_command, data_dir, copy_sketch, working_dir):
    # Init the environment explicitly
    assert run_command(["core", "update-index"])

    # Install Arduino AVR Boards
    assert run_command(["core", "install", "arduino:avr@1.8.3"])

    sketch_path = copy_sketch("sketch_with_multiple_defines")
    fqbn = "arduino:avr:uno"

    # Compile using multiple build properties separated by a space
    res = run_command(
        [
            "compile",
            "-b",
            fqbn,
            '--build-property=compiler.cpp.extra_flags=\\"-DPIN=2 -DSSID=\\"This is a String\\"\\"',
            sketch_path,
            "--verbose",
            "--clean",
        ]
    )
    assert res.failed

    # Compile using multiple build properties separated by a space and properly quoted
    res = run_command(
        [
            "compile",
            "-b",
            fqbn,
            '--build-property=compiler.cpp.extra_flags=-DPIN=2 \\"-DSSID=\\"This is a String\\"\\"',
            sketch_path,
            "--verbose",
            "--clean",
        ]
    )
    assert res.ok
    assert '-DPIN=2 "-DSSID=\\"This is a String\\""' in res.stdout

    # Tries compilation using multiple build properties separated by a comma
    res = run_command(
        [
            "compile",
            "-b",
            fqbn,
            '--build-property=compiler.cpp.extra_flags=\\"-DPIN=2,-DSSID=\\"This is a String\\"\\"',
            sketch_path,
            "--verbose",
            "--clean",
        ]
    )
    assert res.failed

    res = run_command(
        [
            "compile",
            "-b",
            fqbn,
            '--build-property=compiler.cpp.extra_flags=\\"-DPIN=2\\"',
            '--build-property=compiler.cpp.extra_flags=\\"-DSSID=\\"This is a String\\"\\"',
            sketch_path,
            "--verbose",
            "--clean",
        ]
    )
    assert res.failed
    assert "-DPIN=2" not in res.stdout
    assert '-DSSID=\\"This is a String\\"' in res.stdout

    res = run_command(
        [
            "compile",
            "-b",
            fqbn,
            '--build-property=compiler.cpp.extra_flags=\\"-DPIN=2\\"',
            '--build-property=build.extra_flags=\\"-DSSID=\\"hello world\\"\\"',
            sketch_path,
            "--verbose",
            "--clean",
        ]
    )
    assert res.ok
    assert "-DPIN=2" in res.stdout
    assert '-DSSID=\\"hello world\\"' in res.stdout
