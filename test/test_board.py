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


def test_board_search(run_command, data_dir):
    assert run_command(["update"])

    res = run_command(["board", "search", "--format", "json"])
    assert res.ok
    data = json.loads(res.stdout)
    # Verifies boards are returned
    assert len(data) > 0
    # Verifies no board has FQBN set since no platform is installed
    assert len([board["fqbn"] for board in data if "fqbn" in board]) == 0
    names = [board["name"] for board in data if "name" in board]
    assert "Arduino Uno" in names
    assert "Arduino Yún" in names
    assert "Arduino Zero" in names
    assert "Arduino Nano 33 BLE" in names
    assert "Arduino Portenta H7" in names

    # Search in non installed boards
    res = run_command(["board", "search", "--format", "json", "nano", "33"])
    assert res.ok
    data = json.loads(res.stdout)
    # Verifies boards are returned
    assert len(data) > 0
    # Verifies no board has FQBN set since no platform is installed
    assert len([board["fqbn"] for board in data if "fqbn" in board]) == 0
    names = [board["name"] for board in data if "name" in board]
    assert "Arduino Nano 33 BLE" in names
    assert "Arduino Nano 33 IoT" in names

    # Install a platform from index
    assert run_command(["core", "install", "arduino:avr@1.8.3"])

    res = run_command(["board", "search", "--format", "json"])
    assert res.ok
    data = json.loads(res.stdout)
    assert len(data) > 0
    # Verifies some FQBNs are now returned after installing a platform
    assert len([board["fqbn"] for board in data if "fqbn" in board]) == 26
    installed_boards = {board["fqbn"]: board for board in data if "fqbn" in board}
    assert "arduino:avr:uno" in installed_boards
    assert "Arduino Uno" == installed_boards["arduino:avr:uno"]["name"]
    assert "arduino:avr:yun" in installed_boards
    assert "Arduino Yún" == installed_boards["arduino:avr:yun"]["name"]

    res = run_command(["board", "search", "--format", "json", "arduino", "yun"])
    assert res.ok
    data = json.loads(res.stdout)
    assert len(data) > 0
    installed_boards = {board["fqbn"]: board for board in data if "fqbn" in board}
    assert "arduino:avr:yun" in installed_boards
    assert "Arduino Yún" == installed_boards["arduino:avr:yun"]["name"]

    # Manually installs a core in sketchbooks hardware folder
    git_url = "https://github.com/arduino/ArduinoCore-samd.git"
    repo_dir = Path(data_dir, "hardware", "arduino-beta-development", "samd")
    assert Repo.clone_from(git_url, repo_dir, multi_options=["-b 1.8.11"])

    res = run_command(["board", "search", "--format", "json"])
    assert res.ok
    data = json.loads(res.stdout)
    assert len(data) > 0
    # Verifies some FQBNs are now returned after installing a platform
    assert len([board["fqbn"] for board in data if "fqbn" in board]) == 43
    installed_boards = {board["fqbn"]: board for board in data if "fqbn" in board}
    assert "arduino:avr:uno" in installed_boards
    assert "Arduino Uno" == installed_boards["arduino:avr:uno"]["name"]
    assert "arduino:avr:yun" in installed_boards
    assert "Arduino Yún" == installed_boards["arduino:avr:yun"]["name"]
    assert "arduino-beta-development:samd:mkrwifi1010" in installed_boards
    assert "Arduino MKR WiFi 1010" == installed_boards["arduino-beta-development:samd:mkrwifi1010"]["name"]
    assert "arduino-beta-development:samd:mkr1000" in installed_boards
    assert "Arduino MKR1000" == installed_boards["arduino-beta-development:samd:mkr1000"]["name"]
    assert "arduino-beta-development:samd:mkrzero" in installed_boards
    assert "Arduino MKRZERO" == installed_boards["arduino-beta-development:samd:mkrzero"]["name"]
    assert "arduino-beta-development:samd:nano_33_iot" in installed_boards
    assert "Arduino NANO 33 IoT" == installed_boards["arduino-beta-development:samd:nano_33_iot"]["name"]
    assert "arduino-beta-development:samd:arduino_zero_native" in installed_boards

    res = run_command(["board", "search", "--format", "json", "mkr1000"])
    assert res.ok
    data = json.loads(res.stdout)
    assert len(data) > 0
    # Verifies some FQBNs are now returned after installing a platform
    installed_boards = {board["fqbn"]: board for board in data if "fqbn" in board}
    assert "arduino-beta-development:samd:mkr1000" in installed_boards
    assert "Arduino MKR1000" == installed_boards["arduino-beta-development:samd:mkr1000"]["name"]


def test_board_attach_without_sketch_json(run_command, data_dir):
    run_command(["update"])

    sketch_name = "BoardAttachWithoutSketchJson"
    sketch_path = Path(data_dir, sketch_name)
    fqbn = "arduino:avr:uno"

    # Create a test sketch
    assert run_command(["sketch", "new", sketch_path])

    assert run_command(["board", "attach", "-b", fqbn, sketch_path])


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
