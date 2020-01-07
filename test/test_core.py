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
    # list all
    result = run_command("core search")
    assert result.ok
    # filter out empty lines and subtract 1 for the header line
    platforms_count = len([l for l in result.stdout.splitlines() if l]) - 1
    result = run_command("core search --format json")
    assert result.ok
    assert len(json.loads(result.stdout)) == platforms_count
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
