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

import pytest
import simplejson as json
from invoke import Local
from invoke.context import Context

from .common import Board


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
def working_dir(tmpdir_factory):
    """
    A tmp folder to work in
    will be created before running each test and deleted
    at the end, this way all the tests work in isolation.
    """
    return str(tmpdir_factory.mktemp("ArduinoTestWork"))


@pytest.fixture(scope="function")
def run_command(pytestconfig, data_dir, downloads_dir, working_dir):
    """
    Provide a wrapper around invoke's `run` API so that every test
    will work in the same temporary folder.

    Useful reference:
        http://docs.pyinvoke.org/en/1.4/api/runners.html#invoke.runners.Result
    """
    cli_path = os.path.join(str(pytestconfig.rootdir), "..", "arduino-cli")
    env = {
        "ARDUINO_DATA_DIR": data_dir,
        "ARDUINO_DOWNLOADS_DIR": downloads_dir,
        "ARDUINO_SKETCHBOOK_DIR": data_dir,
    }
    os.makedirs(os.path.join(data_dir, "packages"))

    def _run(cmd_string):
        cli_full_line = "{} {}".format(cli_path, cmd_string)
        run_context = Context()
        with run_context.cd(working_dir):
            return run_context.run(
                cli_full_line, echo=False, hide=True, warn=True, env=env
            )

    return _run


@pytest.fixture(scope="function")
def daemon_runner(pytestconfig, data_dir, downloads_dir, working_dir):
    """
    Provide an invoke's `Local` object that has started the arduino-cli in daemon mode.
    This way is simple to start and kill the daemon when the test is finished
    via the kill() function

    Useful reference:
        http://docs.pyinvoke.org/en/1.4/api/runners.html#invoke.runners.Local
        http://docs.pyinvoke.org/en/1.4/api/runners.html
    """
    cli_full_line = os.path.join(str(pytestconfig.rootdir), "..", "arduino-cli daemon")
    env = {
        "ARDUINO_DATA_DIR": data_dir,
        "ARDUINO_DOWNLOADS_DIR": downloads_dir,
        "ARDUINO_SKETCHBOOK_DIR": data_dir,
    }
    os.makedirs(os.path.join(data_dir, "packages"))
    run_context = Context()
    run_context.cd(working_dir)
    # Local Class is the implementation of a Runner abstract class
    runner = Local(run_context)
    runner.run(
        cli_full_line, echo=False, hide=True, warn=True, env=env, asynchronous=True
    )

    return runner


@pytest.fixture(scope="function")
def detected_boards(run_command):
    """
    This fixture provides a list of all the boards attached to the host.
    This fixture will parse the JSON output of `arduino-cli board list --format json`
    to extract all the connected boards data.

    :returns a list `Board` objects.
    """
    assert run_command("core update-index")
    result = run_command("board list --format json")
    assert result.ok

    detected_boards = []
    for port in json.loads(result.stdout):
        for board in port.get("boards", []):
            fqbn = board.get("FQBN")
            package, architecture, _id = fqbn.split(":")
            detected_boards.append(
                Board(
                    address=port.get("address"),
                    fqbn=fqbn,
                    package=package,
                    architecture=architecture,
                    id=_id,
                    core="{}:{}".format(package, architecture),
                )
            )

    return detected_boards
