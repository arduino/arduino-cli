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
import json
import os

import semver
import yaml


def test_version(run_command):
    result = run_command(["version"])
    assert result.ok
    assert "Version:" in result.stdout
    assert "Commit:" in result.stdout
    assert "" == result.stderr

    result = run_command(["version", "--format", "json"])
    assert result.ok
    parsed_out = json.loads(result.stdout)
    assert parsed_out.get("Application", False) == "arduino-cli"
    version = parsed_out.get("VersionString", False)
    assert semver.VersionInfo.isvalid(version=version) or "git-snapshot" in version or "nightly" in version
    assert isinstance(parsed_out.get("Commit", False), str)


def test_log_options(run_command, data_dir):
    """
    using `version` as a test command
    """

    # no logs
    out_lines = run_command(["version"]).stdout.strip().split("\n")
    assert len(out_lines) == 1

    # plain text logs on stdoud
    out_lines = run_command(["version", "-v"]).stdout.strip().split("\n")
    assert len(out_lines) > 1
    assert out_lines[0].startswith("\x1b[36mINFO\x1b[0m")  # account for the colors

    # plain text logs on file
    log_file = os.path.join(data_dir, "log.txt")
    run_command(["version", "--log-file", log_file])
    with open(log_file) as f:
        lines = f.readlines()
        assert lines[0].startswith('time="')  # file format is different from console

    # json on stdout
    out_lines = run_command(["version", "-v", "--log-format", "JSON"]).stdout.strip().split("\n")
    lg = json.loads(out_lines[0])
    assert "level" in lg

    # json on file
    log_file = os.path.join(data_dir, "log.json")
    run_command(["version", "--log-format", "json", "--log-file", log_file])
    with open(log_file) as f:
        for line in f.readlines():
            json.loads(line)


def test_inventory_creation(run_command, data_dir):
    """
    using `version` as a test command
    """

    # no logs
    out_lines = run_command(["version"]).stdout.strip().split("\n")
    assert len(out_lines) == 1

    # parse inventory file
    inventory_file = os.path.join(data_dir, "inventory.yaml")
    with open(inventory_file, "r") as stream:
        inventory = yaml.safe_load(stream)
        assert "installation" in inventory
