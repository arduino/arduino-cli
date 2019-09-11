# This file is part of arduino-cli.
#
# Copyright 2019 ARDUINO SA (http://www.arduino.cc/)
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

import pytest

from test.common import running_on_ci


@pytest.mark.skipif(running_on_ci() and platform.system() == "Windows",
                    reason="Test disabled on Github Actions Win VM until tmpdir inconsistent behavior bug is fixed")
def test_sketch_new(run_command, working_dir):
    # Create a test sketch in current directory
    current_path = working_dir
    sketch_name = "SketchNewIntegrationTest"
    current_sketch_path = os.path.join(current_path, sketch_name)
    result = run_command("sketch new {}".format(sketch_name))
    assert result.ok
    assert "Sketch created in: {}".format(current_sketch_path) in result.stdout
    assert os.path.isfile(os.path.join(current_sketch_path, sketch_name + ".ino"))

    # Create a test sketch in current directory but using an absolute path
    sketch_name = "SketchNewIntegrationTestAbsolute"
    current_sketch_path = os.path.join(current_path, sketch_name)
    result = run_command("sketch new {}".format(current_sketch_path))
    assert result.ok
    assert "Sketch created in: {}".format(current_sketch_path) in result.stdout
    assert os.path.isfile(os.path.join(current_sketch_path, sketch_name + ".ino"))

    # Create a test sketch in current directory subpath but using an absolute path
    sketch_name = "SketchNewIntegrationTestSubpath"
    sketch_subpath = os.path.join("subpath", sketch_name)
    current_sketch_path = os.path.join(current_path, sketch_subpath)
    result = run_command("sketch new {}".format(sketch_subpath))
    assert result.ok
    assert "Sketch created in: {}".format(current_sketch_path) in result.stdout
    assert os.path.isfile(os.path.join(current_sketch_path, sketch_name + ".ino"))

    # Create a test sketch in current directory using .ino extension
    sketch_name = "SketchNewIntegrationTestDotIno"
    current_sketch_path = os.path.join(current_path, sketch_name)
    result = run_command("sketch new {}".format(sketch_name + ".ino"))
    assert result.ok
    assert "Sketch created in: {}".format(current_sketch_path) in result.stdout
    assert os.path.isfile(os.path.join(current_sketch_path, sketch_name + ".ino"))
