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
import json
import yaml


def test_set_string_with_multiple_arguments(run_command):
    # Create a config file
    assert run_command(["config", "init", "--dest-dir", "."])

    # Verifies default state
    result = run_command(["config", "dump", "--format", "json"])
    assert result.ok
    settings_json = json.loads(result.stdout)
    assert "info" == settings_json["logging"]["level"]

    # Tries to change value
    res = run_command(["config", "set", "logging.level", "trace", "debug"])
    assert res.failed
    assert "Can't set multiple values in key logging.level" in res.stderr


def test_set_bool_with_single_argument(run_command):
    # Create a config file
    assert run_command(["config", "init", "--dest-dir", "."])

    # Verifies default state
    result = run_command(["config", "dump", "--format", "json"])
    assert result.ok
    settings_json = json.loads(result.stdout)
    assert not settings_json["library"]["enable_unsafe_install"]

    # Changes value
    assert run_command(["config", "set", "library.enable_unsafe_install", "true"])

    # Verifies value is changed
    result = run_command(["config", "dump", "--format", "json"])
    assert result.ok
    settings_json = json.loads(result.stdout)
    assert settings_json["library"]["enable_unsafe_install"]


def test_set_bool_with_multiple_arguments(run_command):
    # Create a config file
    assert run_command(["config", "init", "--dest-dir", "."])

    # Verifies default state
    result = run_command(["config", "dump", "--format", "json"])
    assert result.ok
    settings_json = json.loads(result.stdout)
    assert not settings_json["library"]["enable_unsafe_install"]

    # Changes value'
    res = run_command(["config", "set", "library.enable_unsafe_install", "true", "foo"])
    assert res.failed
    assert "Can't set multiple values in key library.enable_unsafe_install" in res.stderr


def test_delete(run_command, working_dir):
    # Create a config file
    assert run_command(["config", "init", "--dest-dir", "."])

    # Verifies default state
    result = run_command(["config", "dump", "--format", "json"])
    assert result.ok
    settings_json = json.loads(result.stdout)
    assert not settings_json["library"]["enable_unsafe_install"]

    # Delete config key
    assert run_command(["config", "delete", "library.enable_unsafe_install"])

    # Verifies value is not found, we read directly from file instead of using
    # the dump command since that would still print the deleted value if it has
    # a default
    config_file = Path(working_dir, "arduino-cli.yaml")
    config_lines = config_file.open().readlines()
    assert "enable_unsafe_install" not in config_lines

    # Verifies default state
    result = run_command(["config", "dump", "--format", "json"])
    assert result.ok
    settings_json = json.loads(result.stdout)
    assert [] == settings_json["board_manager"]["additional_urls"]

    # Delete config key and sub keys
    assert run_command(["config", "delete", "board_manager"])

    # Verifies value is not found, we read directly from file instead of using
    # the dump command since that would still print the deleted value if it has
    # a default
    config_file = Path(working_dir, "arduino-cli.yaml")
    config_lines = config_file.open().readlines()
    assert "additional_urls" not in config_lines
    assert "board_manager" not in config_lines
