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
import platform
import pytest
import simplejson as json
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

    # Search all Retrokit platforms
    result = run_command(f"core search retrokit --all --additional-urls={url}")
    assert result.ok
    assert 3 == len(result.stdout.strip().splitlines())
    lines = [l.split(maxsplit=2) for l in result.stdout.strip().splitlines()]
    assert ["Retrokits-RK002:arm", "1.0.5", "RK002"] in lines
    assert ["Retrokits-RK002:arm", "1.0.6", "RK002"] in lines

    # Search using Retrokit Package Maintainer
    result = run_command(f"core search Retrokits-RK002 --all --additional-urls={url}")
    assert result.ok
    assert 3 == len(result.stdout.strip().splitlines())
    lines = [l.split(maxsplit=2) for l in result.stdout.strip().splitlines()]
    assert ["Retrokits-RK002:arm", "1.0.5", "RK002"] in lines
    assert ["Retrokits-RK002:arm", "1.0.6", "RK002"] in lines

    # Search using the Retrokit Platform name
    result = run_command(f"core search rk002 --all --additional-urls={url}")
    assert result.ok
    assert 3 == len(result.stdout.strip().splitlines())
    lines = [l.split(maxsplit=2) for l in result.stdout.strip().splitlines()]
    assert ["Retrokits-RK002:arm", "1.0.5", "RK002"] in lines
    assert ["Retrokits-RK002:arm", "1.0.6", "RK002"] in lines

    # Search using a board name
    result = run_command(f"core search myboard --all --additional-urls={url}")
    assert result.ok
    assert 2 == len(result.stdout.strip().splitlines())
    lines = [l.split(maxsplit=2) for l in result.stdout.strip().splitlines()]
    assert ["Package:x86", "1.2.3", "Platform"] in lines


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
    for l in result.stdout.splitlines()[1:]:  # ignore the header on first line
        if l:  # ignore empty lines
            assert not l.startswith("test:x86")
            num_platforms += 1

    # same thing in JSON format, also check the number of platforms found is the same
    result = run_command("core search --format json")
    assert result.ok
    platforms = json.loads(result.stdout)
    for elem in platforms:
        assert elem.get("Name") != "test_core"
    assert len(platforms) == num_platforms

    # list all with additional urls, check the test core is there
    result = run_command(f"core search --additional-urls={url}")
    assert result.ok
    num_platforms = 0
    found = False
    for l in result.stdout.splitlines()[1:]:  # ignore the header on first line
        if l:  # ignore empty lines
            if l.startswith("test:x86"):
                found = True
            num_platforms += 1
    assert found

    # same thing in JSON format, also check the number of platforms found is the same
    result = run_command(f"core search --format json --additional-urls={url}")
    assert result.ok
    found = False
    platforms = json.loads(result.stdout)
    for elem in platforms:
        if elem.get("Name") == "test_core":
            found = True
            break
    assert found
    assert len(platforms) == num_platforms


def test_core_updateindex_invalid_url(run_command):
    url = "http://www.invalid-domain-asjkdakdhadjkh.com/package_example_index.json"
    result = run_command("core update-index --additional-urls={}".format(url))
    assert result.failed


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
    assert run_command("core update-index --additional-urls={}".format(url))
    # install 3rd-party core
    assert run_command("core install esp32:esp32@1.0.4 --additional-urls={}".format(url))
    # create a sketch and compile to double check the core was successfully installed
    sketch_path = os.path.join(data_dir, "test_core_install_esp32")
    assert run_command("sketch new {}".format(sketch_path))
    assert run_command("compile -b esp32:esp32:esp32 {}".format(sketch_path))
    # prevent regressions for https://github.com/arduino/arduino-cli/issues/163
    assert os.path.exists(
        os.path.join(sketch_path, "build/esp32.esp32.esp32/test_core_install_esp32.ino.partitions.bin",)
    )


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
        if name == c.get("ID"):
            if version is None:
                return True
            elif version == c.get("Installed"):
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
