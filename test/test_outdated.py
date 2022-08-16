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

def test_outdated_using_library_with_invalid_version(run_command, data_dir):
    assert run_command(["update"])

    # Install latest version of a library library
    assert run_command(["lib", "install", "WiFi101"])

    # Verifies library is correctly returned
    res = run_command(["outdated"])
    assert res.ok
    assert "WiFi101" not in res.stdout

    # Changes the version of the currently installed library so that it's
    # invalid
    lib_path = Path(data_dir, "libraries", "WiFi101")
    Path(lib_path, "library.properties").write_text("name=WiFi101\nversion=1.0001")

    # Verifies library is correctly returned
    res = run_command(["outdated"])
    assert res.ok
    lines = [l.strip().split() for l in res.stdout.splitlines()]
    assert "WiFi101" == lines[1][0]
