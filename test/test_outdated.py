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


def test_outdated(run_command):
    # Updates index for cores and libraries
    assert run_command("core update-index")
    assert run_command("lib update-index")

    # Installs an outdated core and library
    assert run_command("core install --skip-post-install arduino:avr@1.6.3")
    assert run_command("lib install USBHost@1.0.0")

    # Installs latest version of a core and a library
    assert run_command("core install --skip-post-install arduino:samd")
    assert run_command("lib install ArduinoJson")

    # Verifies only outdated cores and libraries are returned
    result = run_command("outdated")
    assert result.ok
    lines = [l.strip() for l in result.stdout.splitlines()]
    assert lines[1].startswith("Arduino AVR Boards")
    assert lines[4].startswith("USBHost")
