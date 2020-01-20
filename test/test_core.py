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
import pytest
import simplejson as json

from .common import running_on_ci


def test_core_search(run_command):
    url = "https://raw.githubusercontent.com/arduino/arduino-cli/master/test/testdata/test_index.json"
    assert run_command("core update-index --additional-urls={}".format(url))
    # search a specific core
    result = run_command("core search avr")
    assert result.ok
    assert 2 < len(result.stdout.splitlines())
    result = run_command("core search avr --format json")
    assert result.ok
    data = json.loads(result.stdout)
    assert 0 < len(data)
    # additional URL
    result = run_command(
        "core search test_core --format json --additional-urls={}".format(url)
    )
    assert result.ok
    data = json.loads(result.stdout)
    assert 1 == len(data)
    # show all versions
    result = run_command(
        "core search test_core --all --format json --additional-urls={}".format(url)
    )
    assert result.ok
    data = json.loads(result.stdout)
    assert 2 == len(data)


def test_core_search_no_args(run_command):
    """
    This tests `core search` with and without additional URLs in case no args
    are passed (i.e. all results are shown).
    """
    # update custom index and install test core (installed cores affect `core search`)
    url = "https://raw.githubusercontent.com/arduino/arduino-cli/massi/506/test/testdata/test_index.json"
    assert run_command("core update-index --additional-urls={}".format(url))
    assert run_command("core install test:x86 --additional-urls={}".format(url))

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
    result = run_command("core search --additional-urls={}".format(url))
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
    result = run_command("core search --format json --additional-urls={}".format(url))
    assert result.ok
    found = False
    platforms = json.loads(result.stdout)
    for elem in platforms:
        if elem.get("Name") == "test_core":
            found = True
            break
    assert found
    assert len(platforms) == num_platforms
