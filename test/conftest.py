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

import pytest
from invoke import run


@pytest.fixture(scope="function")
def data_dir(tmpdir_factory):
    """
    A tmp folder will be created before running
    each test and deleted at the end, this way all the
    tests work in isolation.
    """
    return str(tmpdir_factory.mktemp("ArduinoTest"))


@pytest.fixture(scope="session")
def downloads_dir(tmpdir_factory):
    """
    To save time and bandwidth, all the tests will access
    the same download cache folder.
    """
    return str(tmpdir_factory.mktemp("ArduinoTest"))


@pytest.fixture(scope="function")
def run_command(data_dir, downloads_dir):
    """
    Provide a wrapper around invoke's `run` API so that every test
    will work in the same temporary folder.

    Useful reference:
        http://docs.pyinvoke.org/en/1.2/api/runners.html#invoke.runners.Result
    """
    cli_path = os.path.join(pytest.config.rootdir, "..", "arduino-cli")
    env = {
        "ARDUINO_DATA_DIR": data_dir,
        "ARDUINO_DOWNLOADS_DIR": downloads_dir,
        "ARDUINO_SKETCHBOOK_DIR": data_dir,
    }

    def _run(cmd_string):
        cli_full_line = "{} {}".format(cli_path, cmd_string)
        return run(cli_full_line, echo=False, hide=True, warn=True, env=env)

    return _run
