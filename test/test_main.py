# This file is part of arduino-cli.

# Copyright 2019 ARDUINO SA (http://www.arduino.cc/)

# This software is released under the GNU General Public License version 3,
# which covers the main part of arduino-cli.
# The terms of this license can be found at:
# https://www.gnu.org/licenses/gpl-3.0.en.html

# You can be released from the requirements of the above licenses by purchasing
# a commercial license. Buying such a license is mandatory if you want to modify or
# otherwise use the software for commercial activities involving the Arduino
# software without disclosing the source code of your own applications. To purchase
# a commercial license, send an email to license@arduino.cc.
import json
import semver


def test_help(run_command):
    result = run_command("help")
    assert result.ok
    assert result.stderr == ""
    assert "Usage" in result.stdout


def test_version(run_command):
    result = run_command("version")
    assert result.ok
    assert "Version:" in result.stdout
    assert "Commit:" in result.stdout
    assert "" == result.stderr

    result = run_command("version --format json")
    assert result.ok
    parsed_out = json.loads(result.stdout)
    assert parsed_out.get("Application", False) == "arduino-cli"
    assert isinstance(semver.parse(parsed_out.get("VersionString", False)), dict)
    assert isinstance(parsed_out.get("Commit", False), str)
