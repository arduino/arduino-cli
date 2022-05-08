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


def test_upgrade(run_command):
    # Updates index for cores and libraries
    run_command(["core", "update-index"])
    run_command(["lib", "update-index"])

    # Installs an outdated core and library
    run_command(["core", "install", "arduino:avr@1.6.3"])
    assert run_command(["lib", "install", "USBHost@1.0.0"])

    # Installs latest version of a core and a library
    run_command(["core", "install", "arduino:samd"])
    assert run_command(["lib", "install", "ArduinoJson"])

    # Verifies outdated core and libraries are shown
    result = run_command(["outdated"])
    assert result.ok
    lines = result.stdout.splitlines()
    assert "Arduino AVR Boards" in lines[1]
    assert "USBHost" in lines[4]

    result = run_command(["upgrade"])
    assert result.ok

    # Verifies cores and libraries have been updated
    result = run_command(["outdated"])
    assert result.ok
    assert result.stdout == ""


def test_upgrade_using_library_with_invalid_version(run_command, data_dir):
    assert run_command(["update"])

    # Install latest version of a library
    assert run_command(["lib", "install", "WiFi101"])

    # Verifies library is not shown
    res = run_command(["outdated"])
    assert res.ok
    assert "WiFi101" not in res.stdout

    # Changes the version of the currently installed library so that it's
    # invalid
    lib_path = Path(data_dir, "libraries", "WiFi101")
    Path(lib_path, "library.properties").write_text("name=WiFi101\nversion=1.0001")

    # Verifies library gets upgraded
    res = run_command(["upgrade"])
    assert res.ok
    assert "WiFi101" in res.stdout


def test_upgrade_unused_core_tools_are_removed(run_command, data_dir):
    assert run_command(["update"])

    # Installs a core
    assert run_command(["core", "install", "arduino:avr@1.8.2"])

    # Verifies expected tool is installed
    tool_path = Path(data_dir, "packages", "arduino", "tools", "avr-gcc", "7.3.0-atmel3.6.1-arduino5")
    assert tool_path.exists()

    # Upgrades everything
    assert run_command(["upgrade"])

    # Verifies tool is uninstalled since it's not used by newer core version
    assert not tool_path.exists()
