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
import json
import os

import pytest

from .common import build_runner, Board


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
    Run the ``arduino-cli`` command to perform a the real test on the CLI.
    """
    cli_path = os.path.join(pytestconfig.rootdir, "..", "arduino-cli")
    env = {
        "ARDUINO_DATA_DIR": data_dir,
        "ARDUINO_DOWNLOADS_DIR": downloads_dir,
        "ARDUINO_SKETCHBOOK_DIR": data_dir,
    }

    return build_runner(cli_path, env, working_dir)


@pytest.fixture(scope="session")
def _run_session_command(pytestconfig, tmpdir_factory, downloads_dir):
    """
    Run the ``arduino-cli`` command to collect general metadata and store it in
    a `session` scope for the tests.
    """
    cli_path = os.path.join(pytestconfig.rootdir, "..", "arduino-cli")
    data_dir = tmpdir_factory.mktemp("SessionDataDir")
    env = {
        "ARDUINO_DATA_DIR": data_dir,
        "ARDUINO_DOWNLOADS_DIR": downloads_dir,
        "ARDUINO_SKETCHBOOK_DIR": data_dir,
    }
    working_dir = tmpdir_factory.mktemp("SessionTestWork")

    return build_runner(cli_path, env, working_dir)


@pytest.fixture(scope="session")
def detected_boards(_run_session_command):
    """This fixture provides a list of all the boards attached to the host.

    This fixture will parse the JSON output of the ``arduino-cli board list --format json``
    command to extract all the connected boards data.

    :returns a list ``Board`` objects.
    """

    result = _run_session_command("core update-index")
    assert result.ok

    result = _run_session_command("board list --format json")
    assert result.ok

    detected_boards = []

    ports = json.loads(result.stdout)
    assert isinstance(ports, list)
    for port in ports:
        boards = port.get('boards', [])
        assert isinstance(boards, list)
        for board in boards:
            fqbn = board.get('FQBN')
            package, architecture, _id = fqbn.split(":")
            detected_boards.append(
                Board(
                    address=port.get('address'),
                    fqbn=fqbn,
                    package=package,
                    architecture=architecture,
                    id=_id,
                    core="{}:{}".format(package, architecture)
                )
            )

    assert len(detected_boards) >= 1, "There are no boards available for testing"

    return detected_boards
