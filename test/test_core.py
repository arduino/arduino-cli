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


def test_core_search_manually_installed_cores_not_printed(run_command, data_dir):
    assert run_command(["core", "update-index"])

    # Verifies only cores in board manager are shown
    res = run_command(["core", "search", "--format", "json"])
    assert res.ok
    cores = json.loads(res.stdout)
    num_cores = len(cores)
    assert num_cores > 0

    # Manually installs a core in sketchbooks hardware folder
    git_url = "https://github.com/arduino/ArduinoCore-avr.git"
    repo_dir = Path(data_dir, "hardware", "arduino-beta-development", "avr")
    assert Repo.clone_from(git_url, repo_dir, multi_options=["-b 1.8.3"])

    # Verifies manually installed core is not shown
    res = run_command(["core", "search", "--format", "json"])
    assert res.ok
    cores = json.loads(res.stdout)
    assert num_cores == len(cores)
    mapped = {core["id"]: core for core in cores}
    core_id = "arduino-beta-development:avr"
    assert core_id not in mapped


def test_core_list_all_manually_installed_core(run_command, data_dir):
    assert run_command(["core", "update-index"])

    # Verifies only cores in board manager are shown
    res = run_command(["core", "list", "--all", "--format", "json"])
    assert res.ok
    cores = json.loads(res.stdout)
    num_cores = len(cores)
    assert num_cores > 0

    # Manually installs a core in sketchbooks hardware folder
    git_url = "https://github.com/arduino/ArduinoCore-avr.git"
    repo_dir = Path(data_dir, "hardware", "arduino-beta-development", "avr")
    assert Repo.clone_from(git_url, repo_dir, multi_options=["-b 1.8.3"])

    # Verifies manually installed core is shown
    res = run_command(["core", "list", "--all", "--format", "json"])
    assert res.ok
    cores = json.loads(res.stdout)
    assert num_cores + 1 == len(cores)
    mapped = {core["id"]: core for core in cores}
    expected_core_id = "arduino-beta-development:avr"
    assert expected_core_id in mapped
    assert "Arduino AVR Boards" == mapped[expected_core_id]["name"]
    assert "1.8.3" == mapped[expected_core_id]["latest"]


def test_core_list_updatable_all_flags(run_command, data_dir):
    assert run_command(["core", "update-index"])

    # Verifies only cores in board manager are shown
    res = run_command(["core", "list", "--all", "--updatable", "--format", "json"])
    assert res.ok
    cores = json.loads(res.stdout)
    num_cores = len(cores)
    assert num_cores > 0

    # Manually installs a core in sketchbooks hardware folder
    git_url = "https://github.com/arduino/ArduinoCore-avr.git"
    repo_dir = Path(data_dir, "hardware", "arduino-beta-development", "avr")
    assert Repo.clone_from(git_url, repo_dir, multi_options=["-b 1.8.3"])

    # Verifies using both --updatable and --all flags --all takes precedence
    res = run_command(["core", "list", "--all", "--updatable", "--format", "json"])
    assert res.ok
    cores = json.loads(res.stdout)
    assert num_cores + 1 == len(cores)
    mapped = {core["id"]: core for core in cores}
    expected_core_id = "arduino-beta-development:avr"
    assert expected_core_id in mapped
    assert "Arduino AVR Boards" == mapped[expected_core_id]["name"]
    assert "1.8.3" == mapped[expected_core_id]["latest"]


def test_core_upgrade_removes_unused_tools(run_command, data_dir):
    assert run_command(["update"])

    # Installs a core
    assert run_command(["core", "install", "arduino:avr@1.8.2"])

    # Verifies expected tool is installed
    tool_path = Path(data_dir, "packages", "arduino", "tools", "avr-gcc", "7.3.0-atmel3.6.1-arduino5")
    assert tool_path.exists()

    # Upgrades core
    assert run_command(["core", "upgrade", "arduino:avr"])

    # Verifies tool is uninstalled since it's not used by newer core version
    assert not tool_path.exists()


def test_core_install_removes_unused_tools(run_command, data_dir):
    assert run_command(["update"])

    # Installs a core
    assert run_command(["core", "install", "arduino:avr@1.8.2"])

    # Verifies expected tool is installed
    tool_path = Path(data_dir, "packages", "arduino", "tools", "avr-gcc", "7.3.0-atmel3.6.1-arduino5")
    assert tool_path.exists()

    # Installs newer version of already installed core
    assert run_command(["core", "install", "arduino:avr@1.8.3"])

    # Verifies tool is uninstalled since it's not used by newer core version
    assert not tool_path.exists()


def test_core_list_with_installed_json(run_command, data_dir):
    assert run_command(["update"])

    # Install core
    url = "https://adafruit.github.io/arduino-board-index/package_adafruit_index.json"
    assert run_command(["core", "update-index", f"--additional-urls={url}"])
    assert run_command(["core", "install", "adafruit:avr@1.4.13", f"--additional-urls={url}"])

    # Verifies installed core is correctly found and name is set
    res = run_command(["core", "list", "--format", "json"])
    assert res.ok
    cores = json.loads(res.stdout)
    mapped = {core["id"]: core for core in cores}
    assert len(mapped) == 1
    assert "adafruit:avr" in mapped
    assert mapped["adafruit:avr"]["name"] == "Adafruit AVR Boards"

    # Deletes installed.json file, this file stores information about the core,
    # that is used mostly when removing package indexes and their cores are still installed;
    # this way we don't lose much information about it.
    # It might happen that the user has old cores installed before the addition of
    # the installed.json file so we need to handle those cases.
    installed_json = Path(data_dir, "packages", "adafruit", "hardware", "avr", "1.4.13", "installed.json")
    installed_json.unlink()

    # Verifies installed core is still found and name is set
    res = run_command(["core", "list", "--format", "json"])
    assert res.ok
    cores = json.loads(res.stdout)
    mapped = {core["id"]: core for core in cores}
    assert len(mapped) == 1
    assert "adafruit:avr" in mapped
    # Name for this core changes since if there's installed.json file we read it from
    # platform.txt, turns out that this core has different names used in different files
    # thus the change.
    assert mapped["adafruit:avr"]["name"] == "Adafruit Boards"


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


def test_core_search_sorted_results(run_command, httpserver):
    # Set up the server to serve our custom index file
    test_index = Path(__file__).parent / "testdata" / "test_index.json"
    httpserver.expect_request("/test_index.json").respond_with_data(test_index.read_text())

    # update custom index
    url = httpserver.url_for("/test_index.json")
    assert run_command(["core", "update-index", f"--additional-urls={url}"])

    # This is done only to avoid index update output when calling core search
    # since it automatically updates them if they're outdated and it makes it
    # harder to parse the list of cores
    assert run_command(["core", "search"])

    # list all with additional url specified
    result = run_command(["core", "search", f"--additional-urls={url}"])
    assert result.ok

    lines = [l.strip().split(maxsplit=2) for l in result.stdout.strip().splitlines()][1:]
    not_deprecated = [l for l in lines if not l[2].startswith("[DEPRECATED]")]
    deprecated = [l for l in lines if l[2].startswith("[DEPRECATED]")]

    # verify that results are already sorted correctly
    assert not_deprecated == sorted(not_deprecated, key=lambda tokens: tokens[2].lower())
    assert deprecated == sorted(deprecated, key=lambda tokens: tokens[2].lower())

    # verify that deprecated platforms are the last ones
    assert lines == not_deprecated + deprecated

    # test same behaviour with json output
    result = run_command(["core", "search", f"--additional-urls={url}", "--format=json"])
    assert result.ok

    platforms = json.loads(result.stdout)
    not_deprecated = [p for p in platforms if not p.get("deprecated")]
    deprecated = [p for p in platforms if p.get("deprecated")]

    # verify that results are already sorted correctly
    assert not_deprecated == sorted(not_deprecated, key=lambda keys: keys["name"].lower())
    assert deprecated == sorted(deprecated, key=lambda keys: keys["name"].lower())
    # verify that deprecated platforms are the last ones
    assert platforms == not_deprecated + deprecated


def test_core_list_sorted_results(run_command, httpserver):
    # Set up the server to serve our custom index file
    test_index = Path(__file__).parent / "testdata" / "test_index.json"
    httpserver.expect_request("/test_index.json").respond_with_data(test_index.read_text())

    # update custom index
    url = httpserver.url_for("/test_index.json")
    assert run_command(["core", "update-index", f"--additional-urls={url}"])

    # install some core for testing
    assert run_command(
        ["core", "install", "test:x86", "Retrokits-RK002:arm", "Package:x86", f"--additional-urls={url}"]
    )

    # list all with additional url specified
    result = run_command(["core", "list", f"--additional-urls={url}"])
    assert result.ok

    lines = [l.strip().split(maxsplit=3) for l in result.stdout.strip().splitlines()][1:]
    assert len(lines) == 3
    not_deprecated = [l for l in lines if not l[3].startswith("[DEPRECATED]")]
    deprecated = [l for l in lines if l[3].startswith("[DEPRECATED]")]

    # verify that results are already sorted correctly
    assert not_deprecated == sorted(not_deprecated, key=lambda tokens: tokens[3].lower())
    assert deprecated == sorted(deprecated, key=lambda tokens: tokens[3].lower())

    # verify that deprecated platforms are the last ones
    assert lines == not_deprecated + deprecated

    # test same behaviour with json output
    result = run_command(["core", "list", f"--additional-urls={url}", "--format=json"])
    assert result.ok

    platforms = json.loads(result.stdout)
    assert len(platforms) == 3
    not_deprecated = [p for p in platforms if not p.get("deprecated")]
    deprecated = [p for p in platforms if p.get("deprecated")]

    # verify that results are already sorted correctly
    assert not_deprecated == sorted(not_deprecated, key=lambda keys: keys["name"].lower())
    assert deprecated == sorted(deprecated, key=lambda keys: keys["name"].lower())
    # verify that deprecated platforms are the last ones
    assert platforms == not_deprecated + deprecated


def test_core_list_deprecated_platform_with_installed_json(run_command, httpserver, data_dir):
    # Set up the server to serve our custom index file
    test_index = Path(__file__).parent / "testdata" / "test_index.json"
    httpserver.expect_request("/test_index.json").respond_with_data(test_index.read_text())

    # update custom index
    url = httpserver.url_for("/test_index.json")
    assert run_command(["core", "update-index", f"--additional-urls={url}"])

    # install some core for testing
    assert run_command(["core", "install", "Package:x86", f"--additional-urls={url}"])

    installed_json_file = Path(data_dir, "packages", "Package", "hardware", "x86", "1.2.3", "installed.json")
    assert installed_json_file.exists()
    installed_json = json.load(installed_json_file.open("r"))
    platform = installed_json["packages"][0]["platforms"][0]
    del platform["deprecated"]
    installed_json["packages"][0]["platforms"][0] = platform
    with open(installed_json_file, "w") as f:
        json.dump(installed_json, f)

    # test same behaviour with json output
    result = run_command(["core", "list", f"--additional-urls={url}", "--format=json"])
    assert result.ok

    platforms = json.loads(result.stdout)
    assert len(platforms) == 1
    assert platforms[0]["deprecated"]


def test_core_list_platform_without_platform_txt(run_command, data_dir):
    assert run_command(["update"])

    # Verifies no core is installed
    res = run_command(["core", "list", "--format", "json"])
    assert res.ok
    cores = json.loads(res.stdout)
    assert len(cores) == 0

    # Simulates creation of a new core in the sketchbook hardware folder
    # without a platforms.txt
    test_boards_txt = Path(__file__).parent / "testdata" / "boards.local.txt"
    boards_txt = Path(data_dir, "hardware", "some-packager", "some-arch", "boards.txt")
    boards_txt.parent.mkdir(parents=True, exist_ok=True)
    boards_txt.touch()
    boards_txt.write_bytes(test_boards_txt.read_bytes())

    # Verifies no core is installed
    res = run_command(["core", "list", "--format", "json"])
    assert res.ok
    cores = json.loads(res.stdout)
    assert len(cores) == 1
    core = cores[0]
    assert core["id"] == "some-packager:some-arch"
    assert core["name"] == "some-packager-some-arch"


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


def test_core_with_missing_custom_board_options_is_loaded(run_command, data_dir):
    test_platform_name = "platform_with_missing_custom_board_options"
    platform_install_dir = Path(data_dir, "hardware", "arduino-beta-dev", test_platform_name)
    platform_install_dir.mkdir(parents=True)

    # Install platform in Sketchbook hardware dir
    shutil.copytree(
        Path(__file__).parent / "testdata" / test_platform_name,
        platform_install_dir,
        dirs_exist_ok=True,
    )

    assert run_command(["update"])

    res = run_command(["core", "list", "--format", "json"])
    assert res.ok

    cores = json.loads(res.stdout)
    mapped = {core["id"]: core for core in cores}
    assert len(mapped) == 1
    # Verifies platform is loaded except excluding board with missing options
    assert "arduino-beta-dev:platform_with_missing_custom_board_options" in mapped
    boards = {b["fqbn"]: b for b in mapped["arduino-beta-dev:platform_with_missing_custom_board_options"]["boards"]}
    assert len(boards) == 2
    # Verify board with malformed options is not loaded
    assert "arduino-beta-dev:platform_with_missing_custom_board_options:nessuno" in boards
    # Verify other board is loaded
    assert "arduino-beta-dev:platform_with_missing_custom_board_options:altra" in boards


def test_core_list_outdated_core(run_command):
    assert run_command(["update"])

    # Install an old core version
    assert run_command(["core", "install", "arduino:samd@1.8.6"])

    res = run_command(["core", "list", "--format", "json"])

    data = json.loads(res.stdout)
    assert len(data) == 1
    samd_core = data[0]
    assert samd_core["installed"] == "1.8.6"
    installed_version = semver.parse_version_info(samd_core["installed"])
    latest_version = semver.parse_version_info(samd_core["latest"])
    # Installed version must be older than latest
    assert installed_version.compare(latest_version) == -1


def test_core_loading_package_manager(run_command, data_dir):
    # Create empty architecture folder (this condition is normally produced by `core uninstall`)
    (Path(data_dir) / "packages" / "foovendor" / "hardware" / "fooarch").mkdir(parents=True)

    result = run_command(["core", "list", "--all", "--format", "json"])
    assert result.ok  # this should not make the cli crash


def test_core_index_without_checksum(run_command):
    assert run_command(["config", "init", "--dest-dir", "."])
    url = "https://raw.githubusercontent.com/keyboardio/ArduinoCore-GD32-Keyboardio/ae5938af2f485910729e7d27aa233032a1cb4734/package_gd32_index.json"  # noqa: E501
    assert run_command(["config", "add", "board_manager.additional_urls", url])

    assert run_command(["core", "update-index"])
    result = run_command(["core", "list", "--all"])
    assert result.ok  # this should not make the cli crash