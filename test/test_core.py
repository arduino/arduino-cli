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
import datetime
import shutil
import time
import platform
import pytest
import simplejson as json
import tempfile
import hashlib
from git import Repo
from pathlib import Path
import semver


@pytest.mark.skipif(
    platform.system() == "Windows",
    reason="core fails with fatal error: bits/c++config.h: No such file or directory",
)
def test_core_install_esp32(run_command, data_dir):
    # update index
    url = "https://raw.githubusercontent.com/espressif/arduino-esp32/gh-pages/package_esp32_index.json"
    assert run_command(["core", "update-index", f"--additional-urls={url}"])
    # install 3rd-party core
    assert run_command(["core", "install", "esp32:esp32@2.0.0", f"--additional-urls={url}"])
    # create a sketch and compile to double check the core was successfully installed
    sketch_name = "test_core_install_esp32"
    sketch_path = os.path.join(data_dir, sketch_name)
    assert run_command(["sketch", "new", sketch_path])
    assert run_command(["compile", "-b", "esp32:esp32:esp32", sketch_path])
    # prevent regressions for https://github.com/arduino/arduino-cli/issues/163
    sketch_path_md5 = hashlib.md5(sketch_path.encode()).hexdigest().upper()
    build_dir = Path(tempfile.gettempdir(), f"arduino-sketch-{sketch_path_md5}")
    assert (build_dir / f"{sketch_name}.ino.partitions.bin").exists()


def test_core_install_creates_installed_json(run_command, data_dir):
    assert run_command(["core", "update-index"])
    assert run_command(["core", "install", "arduino:avr@1.6.23"])

    installed_json_file = Path(data_dir, "packages", "arduino", "hardware", "avr", "1.6.23", "installed.json")
    assert installed_json_file.exists()
    installed_json = json.load(installed_json_file.open("r"))

    expected_installed_json = json.load((Path(__file__).parent / "testdata" / "installed.json").open("r"))

    def ordered(obj):
        if isinstance(obj, dict):
            return sorted({k: ordered(v) for k, v in obj.items()})
        if isinstance(obj, list):
            return sorted(ordered(x) for x in obj)
        else:
            return obj

    assert ordered(installed_json) == ordered(expected_installed_json)


def test_core_search_update_index_delay(run_command, data_dir):
    assert run_command(["update"])

    # Verifies index update is not run
    res = run_command(["core", "search"])
    assert res.ok
    assert "Downloading index" not in res.stdout

    # Change edit time of package index file
    index_file = Path(data_dir, "package_index.json")
    date = datetime.datetime.now() - datetime.timedelta(hours=25)
    mod_time = time.mktime(date.timetuple())
    os.utime(index_file, (mod_time, mod_time))

    # Verifies index update is run
    res = run_command(["core", "search"])
    assert res.ok
    assert "Downloading index" in res.stdout

    # Verifies index update is not run again
    res = run_command(["core", "search"])
    assert res.ok
    assert "Downloading index" not in res.stdout


@pytest.mark.skipif(
    platform.system() in ["Darwin", "Windows"],
    reason="macOS by default is case insensitive https://github.com/actions/virtual-environments/issues/865 "
    + "Windows too is case insensitive"
    + "https://stackoverflow.com/questions/7199039/file-paths-in-windows-environment-not-case-sensitive",
)
def test_core_download_multiple_platforms(run_command, data_dir):
    assert run_command(["update"])

    # Verifies no core is installed
    res = run_command(["core", "list", "--format", "json"])
    assert res.ok
    cores = json.loads(res.stdout)
    assert len(cores) == 0

    # Simulates creation of two new cores in the sketchbook hardware folder
    test_boards_txt = Path(__file__).parent / "testdata" / "boards.local.txt"
    boards_txt = Path(data_dir, "packages", "PACKAGER", "hardware", "ARCH", "1.0.0", "boards.txt")
    boards_txt.parent.mkdir(parents=True, exist_ok=True)
    boards_txt.touch()
    assert boards_txt.write_bytes(test_boards_txt.read_bytes())

    boards_txt1 = Path(data_dir, "packages", "packager", "hardware", "arch", "1.0.0", "boards.txt")
    boards_txt1.parent.mkdir(parents=True, exist_ok=True)
    boards_txt1.touch()
    assert boards_txt1.write_bytes(test_boards_txt.read_bytes())

    # Verifies the two cores are detected
    res = run_command(["core", "list", "--format", "json"])
    assert res.ok
    cores = json.loads(res.stdout)
    assert len(cores) == 2

    # Try to do an operation on the fake cores.
    # The cli should not allow it since optimizing the casing results in finding two cores
    res = run_command(["core", "upgrade", "Packager:Arch"])
    assert res.failed
    assert "Invalid argument passed: Found 2 platform for reference" in res.stderr


def test_core_index_without_checksum(run_command):
    assert run_command(["config", "init", "--dest-dir", "."])
    url = "https://raw.githubusercontent.com/keyboardio/ArduinoCore-GD32-Keyboardio/ae5938af2f485910729e7d27aa233032a1cb4734/package_gd32_index.json"  # noqa: E501
    assert run_command(["config", "add", "board_manager.additional_urls", url])

    assert run_command(["core", "update-index"])
    result = run_command(["core", "list", "--all"])
    assert result.ok  # this should not make the cli crash
