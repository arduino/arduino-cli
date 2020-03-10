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


@pytest.mark.skipif(running_on_ci(), reason="VMs have no serial ports")
def test_board_list(run_command):
    result = run_command("core update-index")
    assert result.ok
    result = run_command("board list --format json")
    assert result.ok
    # check is a valid json and contains a list of ports
    ports = json.loads(result.stdout)
    assert isinstance(ports, list)
    for port in ports:
        assert "protocol" in port
        assert "protocol_label" in port


@pytest.mark.skipif(running_on_ci(), reason="VMs have no serial ports")
def test_board_listall(run_command):
    assert run_command("core update-index")
    result = run_command("board listall")
    assert result.ok
    assert ["Board", "Name", "FQBN"] == result.stdout.splitlines()[0].strip().split()
