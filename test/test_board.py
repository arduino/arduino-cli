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
from pathlib import Path
from git import Repo
import os
import glob
import simplejson as json
import semver
import pytest

from .common import running_on_ci


def test_board_search_with_outdated_core(run_command):
    assert run_command(["update"])

    # Install an old core version
    assert run_command(["core", "install", "arduino:samd@1.8.6"])

    res = run_command(["board", "search", "arduino:samd:mkrwifi1010", "--format", "json"])

    data = json.loads(res.stdout)
    assert len(data) == 1
    board = data[0]
    assert board["name"] == "Arduino MKR WiFi 1010"
    assert board["fqbn"] == "arduino:samd:mkrwifi1010"
    samd_core = board["platform"]
    assert samd_core["id"] == "arduino:samd"
    installed_version = semver.parse_version_info(samd_core["installed"])
    latest_version = semver.parse_version_info(samd_core["latest"])
    # Installed version must be older than latest
    assert installed_version.compare(latest_version) == -1
