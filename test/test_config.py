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


def test_init(run_command, data_dir, working_dir):
    result = run_command(["config", "init"])
    assert "" == result.stderr
    assert result.ok
    assert data_dir in result.stdout


def test_init_with_existing_custom_config(run_command, data_dir, working_dir, downloads_dir):
    result = run_command(["config", "init", "--additional-urls", "https://example.com"])
    assert result.ok
    assert data_dir in result.stdout

    config_file = open(Path(data_dir) / "arduino-cli.yaml", "r")
    configs = yaml.load(config_file.read(), Loader=yaml.FullLoader)
    config_file.close()
    assert ["https://example.com"] == configs["board_manager"]["additional_urls"]
    assert data_dir == configs["directories"]["data"]
    assert downloads_dir == configs["directories"]["downloads"]
    assert data_dir == configs["directories"]["user"]
    assert "" == configs["logging"]["file"]
    assert "text" == configs["logging"]["format"]
    assert "info" == configs["logging"]["level"]
    assert ":9090" == configs["metrics"]["addr"]
    assert configs["metrics"]["enabled"]

    config_file_path = Path(working_dir) / "config" / "test" / "config.yaml"
    assert not config_file_path.exists()
    result = run_command(["config", "init", "--dest-file", config_file_path])
    assert result.ok
    assert str(config_file_path) in result.stdout

    config_file = open(config_file_path, "r")
    configs = yaml.load(config_file.read(), Loader=yaml.FullLoader)
    config_file.close()
    assert [] == configs["board_manager"]["additional_urls"]
    assert data_dir == configs["directories"]["data"]
    assert downloads_dir == configs["directories"]["downloads"]
    assert data_dir == configs["directories"]["user"]
    assert "" == configs["logging"]["file"]
    assert "text" == configs["logging"]["format"]
    assert "info" == configs["logging"]["level"]
    assert ":9090" == configs["metrics"]["addr"]
    assert configs["metrics"]["enabled"]


def test_init_overwrite_existing_custom_file(run_command, data_dir, working_dir, downloads_dir):
    result = run_command(["config", "init", "--additional-urls", "https://example.com"])
    assert result.ok
    assert data_dir in result.stdout

    config_file = open(Path(data_dir) / "arduino-cli.yaml", "r")
    configs = yaml.load(config_file.read(), Loader=yaml.FullLoader)
    config_file.close()
    assert ["https://example.com"] == configs["board_manager"]["additional_urls"]
    assert data_dir == configs["directories"]["data"]
    assert downloads_dir == configs["directories"]["downloads"]
    assert data_dir == configs["directories"]["user"]
    assert "" == configs["logging"]["file"]
    assert "text" == configs["logging"]["format"]
    assert "info" == configs["logging"]["level"]
    assert ":9090" == configs["metrics"]["addr"]
    assert configs["metrics"]["enabled"]

    result = run_command(["config", "init", "--overwrite"])
    assert result.ok
    assert data_dir in result.stdout

    config_file = open(Path(data_dir) / "arduino-cli.yaml", "r")
    configs = yaml.load(config_file.read(), Loader=yaml.FullLoader)
    config_file.close()
    assert [] == configs["board_manager"]["additional_urls"]
    assert data_dir == configs["directories"]["data"]
    assert downloads_dir == configs["directories"]["downloads"]
    assert data_dir == configs["directories"]["user"]
    assert "" == configs["logging"]["file"]
    assert "text" == configs["logging"]["format"]
    assert "info" == configs["logging"]["level"]
    assert ":9090" == configs["metrics"]["addr"]
    assert configs["metrics"]["enabled"]


def test_init_dest_absolute_path(run_command, working_dir):
    dest = Path(working_dir) / "config" / "test"
    expected_config_file = dest / "arduino-cli.yaml"
    assert not expected_config_file.exists()
    result = run_command(["config", "init", "--dest-dir", dest])
    assert result.ok
    assert str(expected_config_file) in result.stdout
    assert expected_config_file.exists()


def test_init_dest_relative_path(run_command, working_dir):
    dest = Path(working_dir) / "config" / "test"
    expected_config_file = dest / "arduino-cli.yaml"
    assert not expected_config_file.exists()
    result = run_command(["config", "init", "--dest-dir", "config/test"])
    assert result.ok
    assert str(expected_config_file) in result.stdout
    assert expected_config_file.exists()


def test_init_dest_flag_with_overwrite_flag(run_command, working_dir):
    dest = Path(working_dir) / "config" / "test"

    expected_config_file = dest / "arduino-cli.yaml"
    assert not expected_config_file.exists()

    result = run_command(["config", "init", "--dest-dir", dest])
    assert result.ok
    assert expected_config_file.exists()

    result = run_command(["config", "init", "--dest-dir", dest])
    assert result.failed
    assert "Config file already exists, use --overwrite to discard the existing one." in result.stderr

    result = run_command(["config", "init", "--dest-dir", dest, "--overwrite"])
    assert result.ok
    assert str(expected_config_file) in result.stdout


def test_init_dest_and_config_file_flags(run_command, working_dir):
    result = run_command(["config", "init", "--dest-file", "some_other_path", "--dest-dir", "some_path"])
    assert result.failed
    assert "Can't use --dest-file and --dest-dir flags at the same time." in result.stderr


def test_init_config_file_flag_absolute_path(run_command, working_dir):
    config_file = Path(working_dir) / "config" / "test" / "config.yaml"
    assert not config_file.exists()
    result = run_command(["config", "init", "--dest-file", config_file])
    assert result.ok
    assert str(config_file) in result.stdout
    assert config_file.exists()


def test_init_config_file_flag_relative_path(run_command, working_dir):
    config_file = Path(working_dir) / "config.yaml"
    assert not config_file.exists()
    result = run_command(["config", "init", "--dest-file", "config.yaml"])
    assert result.ok
    assert str(config_file) in result.stdout
    assert config_file.exists()


def test_init_config_file_flag_with_overwrite_flag(run_command, working_dir):
    config_file = Path(working_dir) / "config" / "test" / "config.yaml"
    assert not config_file.exists()

    result = run_command(["config", "init", "--dest-file", config_file])
    assert result.ok
    assert config_file.exists()

    result = run_command(["config", "init", "--dest-file", config_file])
    assert result.failed
    assert "Config file already exists, use --overwrite to discard the existing one." in result.stderr

    result = run_command(["config", "init", "--dest-file", config_file, "--overwrite"])
    assert result.ok
    assert str(config_file) in result.stdout


def test_dump(run_command, data_dir, working_dir):
    # Create a config file first
    config_file = Path(working_dir) / "config" / "test" / "config.yaml"
    assert not config_file.exists()
    result = run_command(["config", "init", "--dest-file", config_file])
    assert result.ok
    assert config_file.exists()

    result = run_command(["config", "dump", "--config-file", config_file, "--format", "json"])
    assert result.ok
    settings_json = json.loads(result.stdout)
    assert [] == settings_json["board_manager"]["additional_urls"]

    result = run_command(["config", "init", "--additional-urls", "https://example.com"])
    assert result.ok
    config_file = Path(data_dir) / "arduino-cli.yaml"
    assert str(config_file) in result.stdout
    assert config_file.exists()

    result = run_command(["config", "dump", "--format", "json"])
    assert result.ok
    settings_json = json.loads(result.stdout)
    assert ["https://example.com"] == settings_json["board_manager"]["additional_urls"]


def test_dump_with_config_file_flag(run_command, working_dir):
    # Create a config file first
    config_file = Path(working_dir) / "config" / "test" / "config.yaml"
    assert not config_file.exists()
    result = run_command(["config", "init", "--dest-file", config_file, "--additional-urls=https://example.com"])
    assert result.ok
    assert config_file.exists()

    result = run_command(["config", "dump", "--config-file", config_file, "--format", "json"])
    assert result.ok
    settings_json = json.loads(result.stdout)
    assert ["https://example.com"] == settings_json["board_manager"]["additional_urls"]

    result = run_command(
        [
            "config",
            "dump",
            "--config-file",
            config_file,
            "--additional-urls=https://another-url.com",
            "--format",
            "json",
        ]
    )
    assert result.ok
    settings_json = json.loads(result.stdout)
    assert ["https://another-url.com"] == settings_json["board_manager"]["additional_urls"]


def test_add_remove_set_delete_on_unexisting_key(run_command):
    # Create a config file
    assert run_command(["config", "init", "--dest-dir", "."])

    res = run_command(["config", "add", "some.key", "some_value"])
    assert res.failed
    assert "Settings key doesn't exist" in res.stderr

    res = run_command(["config", "remove", "some.key", "some_value"])
    assert res.failed
    assert "Settings key doesn't exist" in res.stderr

    res = run_command(["config", "set", "some.key", "some_value"])
    assert res.failed
    assert "Settings key doesn't exist" in res.stderr

    res = run_command(["config", "delete", "some.key"])
    assert res.failed
    assert "Settings key doesn't exist" in res.stderr


def test_add_single_argument(run_command):
    # Create a config file
    assert run_command(["config", "init", "--dest-dir", "."])

    # Verifies no additional urls are present
    result = run_command(["config", "dump", "--format", "json"])
    assert result.ok
    settings_json = json.loads(result.stdout)
    assert [] == settings_json["board_manager"]["additional_urls"]

    # Adds one URL
    url = "https://example.com"
    assert run_command(["config", "add", "board_manager.additional_urls", url])

    # Verifies URL has been saved
    result = run_command(["config", "dump", "--format", "json"])
    assert result.ok
    settings_json = json.loads(result.stdout)
    assert ["https://example.com"] == settings_json["board_manager"]["additional_urls"]


def test_add_multiple_arguments(run_command):
    # Create a config file
    assert run_command(["config", "init", "--dest-dir", "."])

    # Verifies no additional urls are present
    result = run_command(["config", "dump", "--format", "json"])
    assert result.ok
    settings_json = json.loads(result.stdout)
    assert [] == settings_json["board_manager"]["additional_urls"]

    # Adds multiple URLs at the same time
    urls = [
        "https://example.com/package_example_index.json",
        "https://example.com/yet_another_package_example_index.json",
    ]
    assert run_command(["config", "add", "board_manager.additional_urls"] + urls)

    # Verifies URL has been saved
    result = run_command(["config", "dump", "--format", "json"])
    assert result.ok
    settings_json = json.loads(result.stdout)
    assert 2 == len(settings_json["board_manager"]["additional_urls"])
    assert urls[0] in settings_json["board_manager"]["additional_urls"]
    assert urls[1] in settings_json["board_manager"]["additional_urls"]


def test_add_on_unsupported_key(run_command):
    # Create a config file
    assert run_command(["config", "init", "--dest-dir", "."])

    # Verifies default value
    result = run_command(["config", "dump", "--format", "json"])
    assert result.ok
    settings_json = json.loads(result.stdout)
    assert "text" == settings_json["logging"]["format"]

    # Tries and fails to add a new item
    result = run_command(["config", "add", "logging.format", "json"])
    assert result.failed
    assert "The key 'logging.format' is not a list of items, can't add to it.\nMaybe use 'config set'?" in result.stderr

    # Verifies value is not changed
    result = run_command(["config", "dump", "--format", "json"])
    assert result.ok
    settings_json = json.loads(result.stdout)
    assert "text" == settings_json["logging"]["format"]


def test_remove_single_argument(run_command):
    # Create a config file
    assert run_command(["config", "init", "--dest-dir", "."])

    # Adds URLs
    urls = [
        "https://example.com/package_example_index.json",
        "https://example.com/yet_another_package_example_index.json",
    ]
    assert run_command(["config", "add", "board_manager.additional_urls"] + urls)

    # Verifies default state
    result = run_command(["config", "dump", "--format", "json"])
    assert result.ok
    settings_json = json.loads(result.stdout)
    assert 2 == len(settings_json["board_manager"]["additional_urls"])
    assert urls[0] in settings_json["board_manager"]["additional_urls"]
    assert urls[1] in settings_json["board_manager"]["additional_urls"]

    # Remove first URL
    assert run_command(["config", "remove", "board_manager.additional_urls", urls[0]])

    # Verifies URLs has been removed
    result = run_command(["config", "dump", "--format", "json"])
    assert result.ok
    settings_json = json.loads(result.stdout)
    assert ["https://example.com/yet_another_package_example_index.json"] == settings_json["board_manager"][
        "additional_urls"
    ]


def test_remove_multiple_arguments(run_command):
    # Create a config file
    assert run_command(["config", "init", "--dest-dir", "."])

    # Adds URLs
    urls = [
        "https://example.com/package_example_index.json",
        "https://example.com/yet_another_package_example_index.json",
    ]
    assert run_command(["config", "add", "board_manager.additional_urls"] + urls)

    # Verifies default state
    result = run_command(["config", "dump", "--format", "json"])
    assert result.ok
    settings_json = json.loads(result.stdout)
    assert 2 == len(settings_json["board_manager"]["additional_urls"])
    assert urls[0] in settings_json["board_manager"]["additional_urls"]
    assert urls[1] in settings_json["board_manager"]["additional_urls"]

    # Remove all URLs
    assert run_command(["config", "remove", "board_manager.additional_urls"] + urls)

    # Verifies all URLs have been removed
    result = run_command(["config", "dump", "--format", "json"])
    assert result.ok
    settings_json = json.loads(result.stdout)
    assert [] == settings_json["board_manager"]["additional_urls"]


def test_remove_on_unsupported_key(run_command):
    # Create a config file
    assert run_command(["config", "init", "--dest-dir", "."])

    # Verifies default value
    result = run_command(["config", "dump", "--format", "json"])
    assert result.ok
    settings_json = json.loads(result.stdout)
    assert "text" == settings_json["logging"]["format"]

    # Tries and fails to remove an item
    result = run_command(["config", "remove", "logging.format", "text"])
    assert result.failed
    assert (
        "The key 'logging.format' is not a list of items, can't remove from it.\nMaybe use 'config delete'?"
        in result.stderr
    )

    # Verifies value is not changed
    result = run_command(["config", "dump", "--format", "json"])
    assert result.ok
    settings_json = json.loads(result.stdout)
    assert "text" == settings_json["logging"]["format"]


def test_set_slice_with_single_argument(run_command):
    # Create a config file
    assert run_command(["config", "init", "--dest-dir", "."])

    # Verifies default state
    result = run_command(["config", "dump", "--format", "json"])
    assert result.ok
    settings_json = json.loads(result.stdout)
    assert [] == settings_json["board_manager"]["additional_urls"]

    # Set an URL in the list
    url = "https://example.com/package_example_index.json"
    assert run_command(["config", "set", "board_manager.additional_urls", url])

    # Verifies value is changed
    result = run_command(["config", "dump", "--format", "json"])
    assert result.ok
    settings_json = json.loads(result.stdout)
    assert [url] == settings_json["board_manager"]["additional_urls"]

    # Sets another URL
    url = "https://example.com/yet_another_package_example_index.json"
    assert run_command(["config", "set", "board_manager.additional_urls", url])

    # Verifies previous value is overwritten
    result = run_command(["config", "dump", "--format", "json"])
    assert result.ok
    settings_json = json.loads(result.stdout)
    assert [url] == settings_json["board_manager"]["additional_urls"]


def test_set_slice_with_multiple_arguments(run_command):
    # Create a config file
    assert run_command(["config", "init", "--dest-dir", "."])

    # Verifies default state
    result = run_command(["config", "dump", "--format", "json"])
    assert result.ok
    settings_json = json.loads(result.stdout)
    assert [] == settings_json["board_manager"]["additional_urls"]

    # Set some URLs in the list
    urls = [
        "https://example.com/first_package_index.json",
        "https://example.com/second_package_index.json",
    ]
    assert run_command(["config", "set", "board_manager.additional_urls"] + urls)

    # Verifies value is changed
    result = run_command(["config", "dump", "--format", "json"])
    assert result.ok
    settings_json = json.loads(result.stdout)
    assert 2 == len(settings_json["board_manager"]["additional_urls"])
    assert urls[0] in settings_json["board_manager"]["additional_urls"]
    assert urls[1] in settings_json["board_manager"]["additional_urls"]

    # Sets another set of URL
    urls = [
        "https://example.com/third_package_index.json",
        "https://example.com/fourth_package_index.json",
    ]
    assert run_command(["config", "set", "board_manager.additional_urls"] + urls)

    # Verifies previous value is overwritten
    result = run_command(["config", "dump", "--format", "json"])
    assert result.ok
    settings_json = json.loads(result.stdout)
    assert 2 == len(settings_json["board_manager"]["additional_urls"])
    assert urls[0] in settings_json["board_manager"]["additional_urls"]
    assert urls[1] in settings_json["board_manager"]["additional_urls"]


def test_set_string_with_single_argument(run_command):
    # Create a config file
    assert run_command(["config", "init", "--dest-dir", "."])

    # Verifies default state
    result = run_command(["config", "dump", "--format", "json"])
    assert result.ok
    settings_json = json.loads(result.stdout)
    assert "info" == settings_json["logging"]["level"]

    # Changes value
    assert run_command(["config", "set", "logging.level", "trace"])

    # Verifies value is changed
    result = run_command(["config", "dump", "--format", "json"])
    assert result.ok
    settings_json = json.loads(result.stdout)
    assert "trace" == settings_json["logging"]["level"]


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
