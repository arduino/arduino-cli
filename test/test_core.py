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
import time
import platform
import pytest
import simplejson as json
import tempfile
import hashlib
from git import Repo
from pathlib import Path


def test_core_search(run_command, httpserver):
    # Set up the server to serve our custom index file
    test_index = Path(__file__).parent / "testdata" / "test_index.json"
    httpserver.expect_request("/test_index.json").respond_with_data(test_index.read_text())

    url = httpserver.url_for("/test_index.json")
    assert run_command(f"core update-index --additional-urls={url}")
    # search a specific core
    result = run_command("core search avr")
    assert result.ok
    assert 2 < len(result.stdout.splitlines())
    result = run_command("core search avr --format json")
    assert result.ok
    data = json.loads(result.stdout)
    assert 0 < len(data)
    # additional URL
    result = run_command("core search test_core --format json --additional-urls={}".format(url))
    assert result.ok
    data = json.loads(result.stdout)
    assert 1 == len(data)
    # show all versions
    result = run_command("core search test_core --all --format json --additional-urls={}".format(url))
    assert result.ok
    data = json.loads(result.stdout)
    assert 2 == len(data)

    def get_platforms(stdout):
        data = json.loads(stdout)
        platforms = {p["id"]: [] for p in data}
        for p in data:
            platforms[p["id"]].append(p["latest"])
        return platforms

    # Search all Retrokit platforms
    result = run_command(f"core search retrokit --all --additional-urls={url} --format json")
    assert result.ok
    platforms = get_platforms(result.stdout)
    assert "1.0.5" in platforms["Retrokits-RK002:arm"]
    assert "1.0.6" in platforms["Retrokits-RK002:arm"]

    # Search using Retrokit Package Maintainer
    result = run_command(f"core search Retrokits-RK002 --all --additional-urls={url} --format json")
    assert result.ok
    platforms = get_platforms(result.stdout)
    assert "1.0.5" in platforms["Retrokits-RK002:arm"]
    assert "1.0.6" in platforms["Retrokits-RK002:arm"]

    # Search using the Retrokit Platform name
    result = run_command(f"core search rk002 --all --additional-urls={url} --format json")
    assert result.ok
    platforms = get_platforms(result.stdout)
    assert "1.0.5" in platforms["Retrokits-RK002:arm"]
    assert "1.0.6" in platforms["Retrokits-RK002:arm"]

    # Search using board names
    result = run_command(f"core search myboard --all --additional-urls={url} --format json")
    assert result.ok
    platforms = get_platforms(result.stdout)
    assert "1.2.3" in platforms["Package:x86"]

    def run_search(search_args, expected_ids):
        res = run_command(f"core search --format json {search_args}")
        assert res.ok
        data = json.loads(res.stdout)
        platform_ids = [p["id"] for p in data]
        for platform_id in expected_ids:
            assert platform_id in platform_ids

    run_search("mkr1000", ["arduino:samd"])
    run_search("mkr 1000", ["arduino:samd"])

    run_search("yún", ["arduino:avr"])
    run_search("yùn", ["arduino:avr"])
    run_search("yun", ["arduino:avr"])

    run_search("nano", ["arduino:avr", "arduino:megaavr", "arduino:samd", "arduino:mbed_nano"])
    run_search("nano 33", ["arduino:samd", "arduino:mbed_nano"])
    run_search("nano ble", ["arduino:mbed_nano"])
    run_search("ble", ["arduino:mbed_nano"])
    run_search("ble nano", ["arduino:mbed_nano"])


def test_core_search_no_args(run_command, httpserver):
    """
    This tests `core search` with and without additional URLs in case no args
    are passed (i.e. all results are shown).
    """
    # Set up the server to serve our custom index file
    test_index = Path(__file__).parent / "testdata" / "test_index.json"
    httpserver.expect_request("/test_index.json").respond_with_data(test_index.read_text())

    # update custom index and install test core (installed cores affect `core search`)
    url = httpserver.url_for("/test_index.json")
    assert run_command(f"core update-index --additional-urls={url}")
    assert run_command(f"core install test:x86 --additional-urls={url}")

    # list all with no additional urls, ensure the test core won't show up
    result = run_command("core search")
    assert result.ok
    num_platforms = 0
    lines = [l.strip().split() for l in result.stdout.strip().splitlines()]
    # The header is printed on the first lines
    assert ["test:x86", "2.0.0", "test_core"] in lines
    header_index = lines.index(["ID", "Version", "Name"])
    # We use black to format and flake8 to lint .py files but they disagree on certain
    # things like this one, thus we ignore this specific flake8 rule and stand by black
    # opinion.
    # We ignore this specific case because ignoring it globally would probably cause more
    # issue. For more info about the rule see: https://www.flake8rules.com/rules/E203.html
    num_platforms = len(lines[header_index + 1 :])  # noqa: E203

    # same thing in JSON format, also check the number of platforms found is the same
    result = run_command("core search --format json")
    assert result.ok
    platforms = json.loads(result.stdout)
    assert 1 == len([e for e in platforms if e.get("name") == "test_core"])
    assert len(platforms) == num_platforms

    # list all with additional urls, check the test core is there
    result = run_command(f"core search --additional-urls={url}")
    assert result.ok
    num_platforms = 0
    lines = [l.strip().split() for l in result.stdout.strip().splitlines()]
    # The header is printed on the first lines
    assert ["test:x86", "2.0.0", "test_core"] in lines
    header_index = lines.index(["ID", "Version", "Name"])
    # We use black to format and flake8 to lint .py files but they disagree on certain
    # things like this one, thus we ignore this specific flake8 rule and stand by black
    # opinion.
    # We ignore this specific case because ignoring it globally would probably cause more
    # issue. For more info about the rule see: https://www.flake8rules.com/rules/E203.html
    num_platforms = len(lines[header_index + 1 :])  # noqa: E203

    # same thing in JSON format, also check the number of platforms found is the same
    result = run_command(f"core search --format json --additional-urls={url}")
    assert result.ok
    platforms = json.loads(result.stdout)
    assert 1 == len([e for e in platforms if e.get("name") == "test_core"])
    assert len(platforms) == num_platforms


def test_core_updateindex_url_not_found(run_command, httpserver):
    assert run_command("core update-index")

    # Brings up a local server to fake a failure
    httpserver.expect_request("/test_index.json").respond_with_data(status=404)
    url = httpserver.url_for("/test_index.json")

    result = run_command(f"core update-index --additional-urls={url}")
    assert result.failed
    lines = [l.strip() for l in result.stderr.splitlines()]
    assert f"Error updating index: downloading index {url}: 404 NOT FOUND" in lines


def test_core_updateindex_internal_server_error(run_command, httpserver):
    assert run_command("core update-index")

    # Brings up a local server to fake a failure
    httpserver.expect_request("/test_index.json").respond_with_data(status=500)
    url = httpserver.url_for("/test_index.json")

    result = run_command(f"core update-index --additional-urls={url}")
    assert result.failed
    lines = [l.strip() for l in result.stderr.splitlines()]
    assert f"Error updating index: downloading index {url}: 500 INTERNAL SERVER ERROR" in lines


def test_core_install_without_updateindex(run_command):
    # Missing "core update-index"
    # Download samd core pinned to 1.8.6
    result = run_command("core install arduino:samd@1.8.6")
    assert result.ok
    assert "Updating index: package_index.json downloaded" in result.stdout


@pytest.mark.skipif(
    platform.system() == "Windows", reason="core fails with fatal error: bits/c++config.h: No such file or directory",
)
def test_core_install_esp32(run_command, data_dir):
    # update index
    url = "https://dl.espressif.com/dl/package_esp32_index.json"
    assert run_command(f"core update-index --additional-urls={url}")
    # install 3rd-party core
    assert run_command(f"core install esp32:esp32@1.0.4 --additional-urls={url}")
    # create a sketch and compile to double check the core was successfully installed
    sketch_name = "test_core_install_esp32"
    sketch_path = os.path.join(data_dir, sketch_name)
    assert run_command(f"sketch new {sketch_path}")
    assert run_command(f"compile -b esp32:esp32:esp32 {sketch_path}")
    # prevent regressions for https://github.com/arduino/arduino-cli/issues/163
    sketch_path_md5 = hashlib.md5(sketch_path.encode()).hexdigest().upper()
    build_dir = Path(tempfile.gettempdir(), f"arduino-sketch-{sketch_path_md5}")
    assert (build_dir / f"{sketch_name}.ino.partitions.bin").exists()


def test_core_download(run_command, downloads_dir):
    assert run_command("core update-index")

    # Download a specific core version
    assert run_command("core download arduino:avr@1.6.16")
    assert os.path.exists(os.path.join(downloads_dir, "packages", "avr-1.6.16.tar.bz2"))

    # Wrong core version
    result = run_command("core download arduino:avr@69.42.0")
    assert result.failed

    # Wrong core
    result = run_command("core download bananas:avr")
    assert result.failed


def _in(jsondata, name, version=None):
    installed_cores = json.loads(jsondata)
    for c in installed_cores:
        if name == c.get("id"):
            if version is None:
                return True
            elif version == c.get("installed"):
                return True
    return False


def test_core_install(run_command):
    assert run_command("core update-index")

    # Install a specific core version
    assert run_command("core install arduino:avr@1.6.16")
    result = run_command("core list --format json")
    assert result.ok
    assert _in(result.stdout, "arduino:avr", "1.6.16")

    # Replace it with a more recent one
    assert run_command("core install arduino:avr@1.6.17")
    result = run_command("core list --format json")
    assert result.ok
    assert _in(result.stdout, "arduino:avr", "1.6.17")

    # Confirm core is listed as "updatable"
    result = run_command("core list --updatable --format json")
    assert result.ok
    assert _in(result.stdout, "arduino:avr", "1.6.17")

    # Upgrade the core to latest version
    assert run_command("core upgrade arduino:avr")
    result = run_command("core list --format json")
    assert result.ok
    assert not _in(result.stdout, "arduino:avr", "1.6.17")
    # double check the code isn't updatable anymore
    result = run_command("core list --updatable --format json")
    assert result.ok
    assert not _in(result.stdout, "arduino:avr")


def test_core_uninstall(run_command):
    assert run_command("core update-index")
    assert run_command("core install arduino:avr")
    result = run_command("core list --format json")
    assert result.ok
    assert _in(result.stdout, "arduino:avr")
    assert run_command("core uninstall arduino:avr")
    result = run_command("core list --format json")
    assert result.ok
    assert not _in(result.stdout, "arduino:avr")


def test_core_uninstall_tool_dependency_removal(run_command, data_dir):
    # These platforms both have a dependency on the arduino:avr-gcc@7.3.0-atmel3.6.1-arduino5 tool
    # arduino:avr@1.8.2 has a dependency on arduino:avrdude@6.3.0-arduino17
    assert run_command("core install arduino:avr@1.8.2")
    # arduino:megaavr@1.8.4 has a dependency on arduino:avrdude@6.3.0-arduino16
    assert run_command("core install arduino:megaavr@1.8.4")
    assert run_command("core uninstall arduino:avr")

    arduino_tools_path = Path(data_dir, "packages", "arduino", "tools")

    avr_gcc_binaries_path = arduino_tools_path.joinpath("avr-gcc", "7.3.0-atmel3.6.1-arduino5", "bin")
    # The tool arduino:avr-gcc@7.3.0-atmel3.6.1-arduino5 that is a dep of another installed platform should remain
    assert avr_gcc_binaries_path.joinpath("avr-gcc").exists() or avr_gcc_binaries_path.joinpath("avr-gcc.exe").exists()

    avrdude_binaries_path = arduino_tools_path.joinpath("avrdude", "6.3.0-arduino17", "bin")
    # The tool arduino:avrdude@6.3.0-arduino17 that is only a dep of arduino:avr should have been removed
    assert (
        avrdude_binaries_path.joinpath("avrdude").exists() or avrdude_binaries_path.joinpath("avrdude.exe").exists()
    ) is False


def test_core_zipslip(run_command):
    url = "https://raw.githubusercontent.com/arduino/arduino-cli/master/test/testdata/test_index.json"
    assert run_command("core update-index --additional-urls={}".format(url))

    # Install a core and check if malicious content has been extracted.
    run_command("core install zipslip:x86 --additional-urls={}".format(url))
    assert os.path.exists("/tmp/evil.txt") is False


def test_core_broken_install(run_command):
    url = "https://raw.githubusercontent.com/arduino/arduino-cli/master/test/testdata/test_index.json"
    assert run_command("core update-index --additional-urls={}".format(url))
    assert not run_command("core install brokenchecksum:x86 --additional-urls={}".format(url))


def test_core_install_creates_installed_json(run_command, data_dir):
    assert run_command("core update-index")
    assert run_command("core install arduino:avr@1.6.23")

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


def test_core_update_with_local_url(run_command):
    test_index = str(Path(__file__).parent / "testdata" / "test_index.json")
    if platform.system() == "Windows":
        test_index = f"/{test_index}".replace("\\", "/")

    res = run_command(f'core update-index --additional-urls="file://{test_index}"')
    assert res.ok
    assert "Updating index: test_index.json downloaded" in res.stdout


def test_core_search_manually_installed_cores_not_printed(run_command, data_dir):
    assert run_command("core update-index")

    # Verifies only cores in board manager are shown
    res = run_command("core search --format json")
    assert res.ok
    cores = json.loads(res.stdout)
    num_cores = len(cores)
    assert num_cores > 0

    # Manually installs a core in sketchbooks hardware folder
    git_url = "https://github.com/arduino/ArduinoCore-avr.git"
    repo_dir = Path(data_dir, "hardware", "arduino-beta-development", "avr")
    assert Repo.clone_from(git_url, repo_dir, multi_options=["-b 1.8.3"])

    # Verifies manually installed core is not shown
    res = run_command("core search --format json")
    assert res.ok
    cores = json.loads(res.stdout)
    assert num_cores == len(cores)
    mapped = {core["id"]: core for core in cores}
    core_id = "arduino-beta-development:avr"
    assert core_id not in mapped


def test_core_list_all_manually_installed_core(run_command, data_dir):
    assert run_command("core update-index")

    # Verifies only cores in board manager are shown
    res = run_command("core list --all --format json")
    assert res.ok
    cores = json.loads(res.stdout)
    num_cores = len(cores)
    assert num_cores > 0

    # Manually installs a core in sketchbooks hardware folder
    git_url = "https://github.com/arduino/ArduinoCore-avr.git"
    repo_dir = Path(data_dir, "hardware", "arduino-beta-development", "avr")
    assert Repo.clone_from(git_url, repo_dir, multi_options=["-b 1.8.3"])

    # Verifies manually installed core is shown
    res = run_command("core list --all --format json")
    assert res.ok
    cores = json.loads(res.stdout)
    assert num_cores + 1 == len(cores)
    mapped = {core["id"]: core for core in cores}
    expected_core_id = "arduino-beta-development:avr"
    assert expected_core_id in mapped
    assert "Arduino AVR Boards" == mapped[expected_core_id]["name"]
    assert "1.8.3" == mapped[expected_core_id]["latest"]


def test_core_list_updatable_all_flags(run_command, data_dir):
    assert run_command("core update-index")

    # Verifies only cores in board manager are shown
    res = run_command("core list --all --updatable --format json")
    assert res.ok
    cores = json.loads(res.stdout)
    num_cores = len(cores)
    assert num_cores > 0

    # Manually installs a core in sketchbooks hardware folder
    git_url = "https://github.com/arduino/ArduinoCore-avr.git"
    repo_dir = Path(data_dir, "hardware", "arduino-beta-development", "avr")
    assert Repo.clone_from(git_url, repo_dir, multi_options=["-b 1.8.3"])

    # Verifies using both --updatable and --all flags --all takes precedence
    res = run_command("core list --all --updatable --format json")
    assert res.ok
    cores = json.loads(res.stdout)
    assert num_cores + 1 == len(cores)
    mapped = {core["id"]: core for core in cores}
    expected_core_id = "arduino-beta-development:avr"
    assert expected_core_id in mapped
    assert "Arduino AVR Boards" == mapped[expected_core_id]["name"]
    assert "1.8.3" == mapped[expected_core_id]["latest"]


def test_core_upgrade_removes_unused_tools(run_command, data_dir):
    assert run_command("update")

    # Installs a core
    assert run_command("core install arduino:avr@1.8.2")

    # Verifies expected tool is installed
    tool_path = Path(data_dir, "packages", "arduino", "tools", "avr-gcc", "7.3.0-atmel3.6.1-arduino5")
    assert tool_path.exists()

    # Upgrades core
    assert run_command("core upgrade arduino:avr")

    # Verifies tool is uninstalled since it's not used by newer core version
    assert not tool_path.exists()


def test_core_install_removes_unused_tools(run_command, data_dir):
    assert run_command("update")

    # Installs a core
    assert run_command("core install arduino:avr@1.8.2")

    # Verifies expected tool is installed
    tool_path = Path(data_dir, "packages", "arduino", "tools", "avr-gcc", "7.3.0-atmel3.6.1-arduino5")
    assert tool_path.exists()

    # Installs newer version of already installed core
    assert run_command("core install arduino:avr@1.8.3")

    # Verifies tool is uninstalled since it's not used by newer core version
    assert not tool_path.exists()


def test_core_list_with_installed_json(run_command, data_dir):
    assert run_command("update")

    # Install core
    url = "https://adafruit.github.io/arduino-board-index/package_adafruit_index.json"
    assert run_command(f"core update-index --additional-urls={url}")
    assert run_command(f"core install adafruit:avr@1.4.13 --additional-urls={url}")

    # Verifies installed core is correctly found and name is set
    res = run_command("core list --format json")
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
    res = run_command("core list --format json")
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
    assert run_command("update")

    # Verifies index update is not run
    res = run_command("core search")
    assert res.ok
    assert "Updating index" not in res.stdout

    # Change edit time of package index file
    index_file = Path(data_dir, "package_index.json")
    date = datetime.datetime.now() - datetime.timedelta(hours=25)
    mod_time = time.mktime(date.timetuple())
    os.utime(index_file, (mod_time, mod_time))

    # Verifies index update is run
    res = run_command("core search")
    assert res.ok
    assert "Updating index" in res.stdout

    # Verifies index update is not run again
    res = run_command("core search")
    assert res.ok
    assert "Updating index" not in res.stdout


def test_core_search_sorted_results(run_command, httpserver):
    # Set up the server to serve our custom index file
    test_index = Path(__file__).parent / "testdata" / "test_index.json"
    httpserver.expect_request("/test_index.json").respond_with_data(test_index.read_text())

    # update custom index
    url = httpserver.url_for("/test_index.json")
    assert run_command(f"core update-index --additional-urls={url}")

    # This is done only to avoid index update output when calling core search
    # since it automatically updates them if they're outdated and it makes it
    # harder to parse the list of cores
    assert run_command("core search")

    # list all with additional url specified
    result = run_command(f"core search --additional-urls={url}")
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
    result = run_command(f"core search --additional-urls={url} --format=json")
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
    assert run_command(f"core update-index --additional-urls={url}")

    # install some core for testing
    assert run_command(f"core install test:x86 Retrokits-RK002:arm Package:x86 --additional-urls={url}")

    # list all with additional url specified
    result = run_command(f"core list --additional-urls={url}")
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
    result = run_command(f"core list --additional-urls={url} --format=json")
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
    assert run_command(f"core update-index --additional-urls={url}")

    # install some core for testing
    assert run_command(f"core install Package:x86 --additional-urls={url}")

    installed_json_file = Path(data_dir, "packages", "Package", "hardware", "x86", "1.2.3", "installed.json")
    assert installed_json_file.exists()
    installed_json = json.load(installed_json_file.open("r"))
    platform = installed_json["packages"][0]["platforms"][0]
    del platform["deprecated"]
    installed_json["packages"][0]["platforms"][0] = platform
    with open(installed_json_file, "w") as f:
        json.dump(installed_json, f)

    # test same behaviour with json output
    result = run_command(f"core list --additional-urls={url} --format=json")
    assert result.ok

    platforms = json.loads(result.stdout)
    assert len(platforms) == 1
    assert platforms[0]["deprecated"]
