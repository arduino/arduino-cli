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
import shutil
import json
from pathlib import Path

import pytest

from .common import running_on_ci, parse_json_traces

# Skip this module when running in CI environments
pytestmark = pytest.mark.skipif(running_on_ci(), reason="VMs have no serial ports")


def test_upload_after_attach(run_command, data_dir, detected_boards):
    # Init the environment explicitly
    run_command(["core", "update-index"])

    for board in detected_boards:
        # Download core
        run_command(["core", "install", board.core])
        # Create a sketch
        sketch_path = os.path.join(data_dir, "foo")
        assert run_command(["sketch", "new", sketch_path])
        assert run_command(["board", "attach", "-p", board.address, sketch_path])
        # Build sketch
        assert run_command(["compile", sketch_path])
        # Upload
        assert run_command(["upload", sketch_path])


def test_upload_to_port_with_board_autodetect(run_command, data_dir, detected_boards):
    assert run_command(["update"])

    # Create a sketch
    sketch_name = "SketchSimple"
    sketch_path = Path(data_dir, sketch_name)
    assert run_command(["sketch", "new", sketch_path])

    for board in detected_boards:
        # Install core
        core = ":".join(board.fqbn.split(":")[:2])
        assert run_command(["core", "install", core])

        assert run_command(["compile", "-b", board.fqbn, sketch_path])

        res = run_command(["upload", "-p", board.address, sketch_path])
        assert res.ok


def test_compile_and_upload_to_port_with_board_autodetect(run_command, data_dir, detected_boards):
    assert run_command(["update"])

    # Create a sketch
    sketch_name = "SketchSimple"
    sketch_path = Path(data_dir, sketch_name)
    assert run_command(["sketch", "new", sketch_path])

    for board in detected_boards:
        # Install core
        core = ":".join(board.fqbn.split(":")[:2])
        assert run_command(["core", "install", core])

        res = run_command(["compile", "-u", "-p", board.address, sketch_path])
        assert res.ok
