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
import platform

import simplejson as json
import pytest
import shutil
from git import Repo
from pathlib import Path
import tempfile
import requests
import zipfile
import io
import re


# Util function to download library from URL
def download_lib(url, download_dir):
    tmp = Path(tempfile.TemporaryDirectory().name)
    tmp.mkdir(parents=True, exist_ok=True)
    regex = re.compile(r"^(.*)-[0-9]+.[0-9]+.[0-9]")
    response = requests.get(url)
    # Download and unzips library removing version suffix
    with zipfile.ZipFile(io.BytesIO(response.content)) as thezip:
        for zipinfo in thezip.infolist():
            with thezip.open(zipinfo) as f:
                dest_dir = tmp / regex.sub("\\g<1>", zipinfo.filename)
                if zipinfo.is_dir():
                    dest_dir.mkdir(parents=True, exist_ok=True)
                else:
                    dest_dir.write_bytes(f.read())

    # Recreates zip with folder without version suffix
    z = zipfile.ZipFile(download_dir, "w")
    for f in tmp.glob("**/*"):
        z.write(f, arcname=f.relative_to(tmp))
    z.close()


def test_install_git_url_and_zip_path_flags_visibility(run_command, data_dir, downloads_dir):
    # Verifies installation fail because flags are not found
    git_url = "https://github.com/arduino-libraries/WiFi101.git"
    res = run_command(["lib", "install", "--git-url", git_url])
    assert res.failed
    assert "--git-url and --zip-path are disabled by default, for more information see:" in res.stderr

    # Download library
    url = "https://github.com/arduino-libraries/AudioZero/archive/refs/tags/1.1.1.zip"
    zip_path = Path(downloads_dir, "libraries", "AudioZero.zip")
    zip_path.parent.mkdir(parents=True, exist_ok=True)
    download_lib(url, zip_path)

    res = run_command(["lib", "install", "--zip-path", zip_path])
    assert res.failed
    assert "--git-url and --zip-path are disabled by default, for more information see:" in res.stderr

    env = {
        "ARDUINO_DATA_DIR": data_dir,
        "ARDUINO_DOWNLOADS_DIR": downloads_dir,
        "ARDUINO_SKETCHBOOK_DIR": data_dir,
        "ARDUINO_ENABLE_UNSAFE_LIBRARY_INSTALL": "true",
    }
    # Verifies installation is successful when flags are enabled with env var
    res = run_command(["lib", "install", "--git-url", git_url], custom_env=env)
    assert res.ok
    assert "--git-url and --zip-path flags allow installing untrusted files, use it at your own risk." in res.stdout

    res = run_command(["lib", "install", "--zip-path", zip_path], custom_env=env)
    assert res.ok
    assert "--git-url and --zip-path flags allow installing untrusted files, use it at your own risk." in res.stdout

    # Uninstall libraries to install them again
    assert run_command(["lib", "uninstall", "WiFi101", "AudioZero"])

    # Verifies installation is successful when flags are enabled with settings file
    assert run_command(["config", "init", "--dest-dir", "."], custom_env=env)

    res = run_command(["lib", "install", "--git-url", git_url])
    assert res.ok
    assert "--git-url and --zip-path flags allow installing untrusted files, use it at your own risk." in res.stdout

    res = run_command(["lib", "install", "--zip-path", zip_path])
    assert res.ok
    assert "--git-url and --zip-path flags allow installing untrusted files, use it at your own risk." in res.stdout


def test_install_with_zip_path(run_command, data_dir, downloads_dir):
    # Initialize configs to enable --zip-path flag
    env = {
        "ARDUINO_DATA_DIR": data_dir,
        "ARDUINO_DOWNLOADS_DIR": downloads_dir,
        "ARDUINO_SKETCHBOOK_DIR": data_dir,
        "ARDUINO_ENABLE_UNSAFE_LIBRARY_INSTALL": "true",
    }
    assert run_command(["config", "init", "--dest-dir", "."], custom_env=env)

    # Download a specific lib version
    # Download library
    url = "https://github.com/arduino-libraries/AudioZero/archive/refs/tags/1.1.1.zip"
    zip_path = Path(downloads_dir, "libraries", "AudioZero.zip")
    zip_path.parent.mkdir(parents=True, exist_ok=True)
    download_lib(url, zip_path)

    lib_install_dir = Path(data_dir, "libraries", "AudioZero")
    # Verifies library is not already installed
    assert not lib_install_dir.exists()

    # Test zip-path install
    res = run_command(["lib", "install", "--zip-path", zip_path])
    assert res.ok
    assert "--git-url and --zip-path flags allow installing untrusted files, use it at your own risk." in res.stdout

    # Verifies library is installed in expected path
    assert lib_install_dir.exists()
    files = list(lib_install_dir.glob("**/*"))
    assert lib_install_dir / "examples" / "SimpleAudioPlayerZero" / "SimpleAudioPlayerZero.ino" in files
    assert lib_install_dir / "src" / "AudioZero.h" in files
    assert lib_install_dir / "src" / "AudioZero.cpp" in files
    assert lib_install_dir / "keywords.txt" in files
    assert lib_install_dir / "library.properties" in files
    assert lib_install_dir / "README.adoc" in files

    # Reinstall library
    assert run_command(["lib", "install", "--zip-path", zip_path])

    # Verifies library remains installed
    assert lib_install_dir.exists()
    files = list(lib_install_dir.glob("**/*"))
    assert lib_install_dir / "examples" / "SimpleAudioPlayerZero" / "SimpleAudioPlayerZero.ino" in files
    assert lib_install_dir / "src" / "AudioZero.h" in files
    assert lib_install_dir / "src" / "AudioZero.cpp" in files
    assert lib_install_dir / "keywords.txt" in files
    assert lib_install_dir / "library.properties" in files
    assert lib_install_dir / "README.adoc" in files


def test_update_index(run_command):
    result = run_command(["lib", "update-index"])
    assert result.ok
    lines = [l.strip() for l in result.stdout.splitlines()]
    assert "Downloading index: library_index.tar.bz2 downloaded" in lines


def test_uninstall(run_command):
    libs = ["Arduino_BQ24195", "WiFiNINA"]
    assert run_command(["lib", "install"] + libs)

    result = run_command(["lib", "uninstall"] + libs)
    assert result.ok


def test_uninstall_spaces(run_command):
    key = "LiquidCrystal I2C"
    assert run_command(["lib", "install", key])
    assert run_command(["lib", "uninstall", key])
    result = run_command(["lib", "list", "--format", "json"])
    assert result.ok
    assert len(json.loads(result.stdout)) == 0


def test_lib_ops_caseinsensitive(run_command):
    """
    This test is supposed to (un)install the following library,
    As you can see the name is all caps:

    Name: "PCM"
      Author: David Mellis <d.mellis@bcmi-labs.cc>, Michael Smith <michael@hurts.ca>
      Maintainer: David Mellis <d.mellis@bcmi-labs.cc>
      Sentence: Playback of short audio samples.
      Paragraph: These samples are encoded directly in the Arduino sketch as an array of numbers.
      Website: http://highlowtech.org/?p=1963
      Category: Signal Input/Output
      Architecture: avr
      Types: Contributed
      Versions: [1.0.0]
    """
    key = "pcm"
    assert run_command(["lib", "install", key])
    assert run_command(["lib", "uninstall", key])
    result = run_command(["lib", "list", "--format", "json"])
    assert result.ok
    assert len(json.loads(result.stdout)) == 0


def test_search(run_command):
    assert run_command(["update"])

    result = run_command(["lib", "search", "--names"])
    assert result.ok
    lines = [l.strip() for l in result.stdout.strip().splitlines()]
    assert "Downloading index: library_index.tar.bz2 downloaded" in lines
    libs = [l[6:].strip('"') for l in lines if "Name:" in l]

    expected = {"WiFi101", "WiFi101OTA", "Firebase Arduino based on WiFi101", "WiFi101_Generic"}
    assert expected == {lib for lib in libs if "WiFi101" in lib}

    result = run_command(["lib", "search", "--names", "--format", "json"])
    assert result.ok
    libs_json = json.loads(result.stdout)
    assert len(libs) == len(libs_json.get("libraries"))

    result = run_command(["lib", "search", "--names"])
    assert result.ok

    def run_search(search_args, expected_libraries):
        res = run_command(["lib", "search", "--names", "--format", "json"] + search_args.split(" "))
        assert res.ok
        data = json.loads(res.stdout)
        libraries = [l["name"] for l in data["libraries"]]
        for l in expected_libraries:
            assert l in libraries

    run_search("Arduino_MKRIoTCarrier", ["Arduino_MKRIoTCarrier"])
    run_search("Arduino mkr iot carrier", ["Arduino_MKRIoTCarrier"])
    run_search("mkr iot carrier", ["Arduino_MKRIoTCarrier"])
    run_search("mkriotcarrier", ["Arduino_MKRIoTCarrier"])

    run_search(
        "dht",
        ["DHT sensor library", "DHT sensor library for ESPx", "DHT12", "SimpleDHT", "TinyDHT sensor library", "SDHT"],
    )
    run_search("dht11", ["DHT sensor library", "DHT sensor library for ESPx", "SimpleDHT", "SDHT"])
    run_search("dht12", ["DHT12", "DHT12 sensor library", "SDHT"])
    run_search("dht22", ["DHT sensor library", "DHT sensor library for ESPx", "SimpleDHT", "SDHT"])
    run_search("dht sensor", ["DHT sensor library", "DHT sensor library for ESPx", "SimpleDHT", "SDHT"])
    run_search("sensor dht", [])

    run_search("arduino json", ["ArduinoJson", "Arduino_JSON"])
    run_search("arduinojson", ["ArduinoJson"])
    run_search("json", ["ArduinoJson", "Arduino_JSON"])


def test_search_paragraph(run_command):
    """
    Search for a string that's only present in the `paragraph` field
    within the index file.
    """
    assert run_command(["lib", "update-index"])
    result = run_command(["lib", "search", "A simple and efficient JSON library", "--names", "--format", "json"])
    assert result.ok
    data = json.loads(result.stdout)
    libraries = [l["name"] for l in data["libraries"]]
    assert "ArduinoJson" in libraries


def test_lib_list_with_updatable_flag(run_command):
    # Init the environment explicitly
    run_command(["lib", "update-index"])

    # No libraries to update
    result = run_command(["lib", "list", "--updatable"])
    assert result.ok
    assert "" == result.stderr
    assert "No libraries update is available." in result.stdout.strip()
    # No library to update in json
    result = run_command(["lib", "list", "--updatable", "--format", "json"])
    assert result.ok
    assert "" == result.stderr
    assert 0 == len(json.loads(result.stdout))

    # Install outdated library
    assert run_command(["lib", "install", "ArduinoJson@6.11.0"])
    # Install latest version of library
    assert run_command(["lib", "install", "WiFi101"])

    res = run_command(["lib", "list", "--updatable"])
    assert res.ok
    assert "" == res.stderr
    # lines = res.stdout.strip().splitlines()
    lines = [l.strip().split(maxsplit=4) for l in res.stdout.strip().splitlines()]
    assert 2 == len(lines)
    assert ["Name", "Installed", "Available", "Location", "Description"] in lines
    line = lines[1]
    assert "ArduinoJson" == line[0]
    assert "6.11.0" == line[1]
    # Verifies available version is not equal to installed one and not empty
    assert "6.11.0" != line[2]
    assert "" != line[2]
    assert "An efficient and elegant JSON library..." == line[4]

    # Look at the JSON output
    res = run_command(["lib", "list", "--updatable", "--format", "json"], hide=True)
    assert res.ok
    assert "" == res.stderr
    data = json.loads(res.stdout)
    assert 1 == len(data)
    # be sure data contains the available version
    assert "6.11.0" == data[0]["library"]["version"]
    assert "6.11.0" != data[0]["release"]["version"]
    assert "" != data[0]["release"]["version"]


def test_install_with_git_url_from_current_directory(run_command, downloads_dir, data_dir):
    assert run_command(["update"])

    env = {
        "ARDUINO_DATA_DIR": data_dir,
        "ARDUINO_DOWNLOADS_DIR": downloads_dir,
        "ARDUINO_SKETCHBOOK_DIR": data_dir,
        "ARDUINO_ENABLE_UNSAFE_LIBRARY_INSTALL": "true",
    }

    lib_install_dir = Path(data_dir, "libraries", "WiFi101")
    # Verifies library is not installed
    assert not lib_install_dir.exists()

    # Clone repository locally
    git_url = "https://github.com/arduino-libraries/WiFi101.git"
    repo_dir = Path(data_dir, "WiFi101")
    assert Repo.clone_from(git_url, repo_dir)

    assert run_command(["lib", "install", "--git-url", "."], custom_working_dir=repo_dir, custom_env=env)

    # Verifies library is installed to correct folder
    assert lib_install_dir.exists()


@pytest.mark.skipif(
    platform.system() == "Windows",
    reason="Using a file uri as git url doesn't work on Windows, "
    + "this must be removed when this issue is fixed: https://github.com/go-git/go-git/issues/247",
)
def test_install_with_git_url_local_file_uri(run_command, downloads_dir, data_dir):
    assert run_command(["update"])

    env = {
        "ARDUINO_DATA_DIR": data_dir,
        "ARDUINO_DOWNLOADS_DIR": downloads_dir,
        "ARDUINO_SKETCHBOOK_DIR": data_dir,
        "ARDUINO_ENABLE_UNSAFE_LIBRARY_INSTALL": "true",
    }

    lib_install_dir = Path(data_dir, "libraries", "WiFi101")
    # Verifies library is not installed
    assert not lib_install_dir.exists()

    # Clone repository locally
    git_url = "https://github.com/arduino-libraries/WiFi101.git"
    repo_dir = Path(data_dir, "WiFi101")
    assert Repo.clone_from(git_url, repo_dir)

    assert run_command(["lib", "install", "--git-url", repo_dir.as_uri()], custom_env=env)

    # Verifies library is installed
    assert lib_install_dir.exists()


def test_install_with_git_local_url(run_command, downloads_dir, data_dir):
    assert run_command(["update"])

    env = {
        "ARDUINO_DATA_DIR": data_dir,
        "ARDUINO_DOWNLOADS_DIR": downloads_dir,
        "ARDUINO_SKETCHBOOK_DIR": data_dir,
        "ARDUINO_ENABLE_UNSAFE_LIBRARY_INSTALL": "true",
    }

    lib_install_dir = Path(data_dir, "libraries", "WiFi101")
    # Verifies library is not installed
    assert not lib_install_dir.exists()

    # Clone repository locally
    git_url = "https://github.com/arduino-libraries/WiFi101.git"
    repo_dir = Path(data_dir, "WiFi101")
    assert Repo.clone_from(git_url, repo_dir)

    assert run_command(["lib", "install", "--git-url", repo_dir], custom_env=env)

    # Verifies library is installed
    assert lib_install_dir.exists()


def test_install_with_git_url_relative_path(run_command, downloads_dir, data_dir):
    assert run_command(["update"])

    env = {
        "ARDUINO_DATA_DIR": data_dir,
        "ARDUINO_DOWNLOADS_DIR": downloads_dir,
        "ARDUINO_SKETCHBOOK_DIR": data_dir,
        "ARDUINO_ENABLE_UNSAFE_LIBRARY_INSTALL": "true",
    }

    lib_install_dir = Path(data_dir, "libraries", "WiFi101")
    # Verifies library is not installed
    assert not lib_install_dir.exists()

    # Clone repository locally
    git_url = "https://github.com/arduino-libraries/WiFi101.git"
    repo_dir = Path(data_dir, "WiFi101")
    assert Repo.clone_from(git_url, repo_dir)

    assert run_command(["lib", "install", "--git-url", "./WiFi101"], custom_working_dir=data_dir, custom_env=env)

    # Verifies library is installed
    assert lib_install_dir.exists()


def test_install_with_git_url_does_not_create_git_repo(run_command, downloads_dir, data_dir):
    assert run_command(["update"])

    env = {
        "ARDUINO_DATA_DIR": data_dir,
        "ARDUINO_DOWNLOADS_DIR": downloads_dir,
        "ARDUINO_SKETCHBOOK_DIR": data_dir,
        "ARDUINO_ENABLE_UNSAFE_LIBRARY_INSTALL": "true",
    }

    lib_install_dir = Path(data_dir, "libraries", "WiFi101")
    # Verifies library is not installed
    assert not lib_install_dir.exists()

    # Clone repository locally
    git_url = "https://github.com/arduino-libraries/WiFi101.git"
    repo_dir = Path(data_dir, "WiFi101")
    assert Repo.clone_from(git_url, repo_dir)

    assert run_command(["lib", "install", "--git-url", repo_dir], custom_env=env)

    # Verifies installed library is not a git repository
    assert not Path(lib_install_dir, ".git").exists()


def test_install_with_git_url_multiple_libraries(run_command, downloads_dir, data_dir):
    assert run_command(["update"])

    env = {
        "ARDUINO_DATA_DIR": data_dir,
        "ARDUINO_DOWNLOADS_DIR": downloads_dir,
        "ARDUINO_SKETCHBOOK_DIR": data_dir,
        "ARDUINO_ENABLE_UNSAFE_LIBRARY_INSTALL": "true",
    }

    wifi_install_dir = Path(data_dir, "libraries", "WiFi101")
    ble_install_dir = Path(data_dir, "libraries", "ArduinoBLE")
    # Verifies libraries are not installed
    assert not wifi_install_dir.exists()
    assert not ble_install_dir.exists()

    wifi_url = "https://github.com/arduino-libraries/WiFi101.git"
    ble_url = "https://github.com/arduino-libraries/ArduinoBLE.git"

    assert run_command(["lib", "install", "--git-url", wifi_url, ble_url], custom_env=env)

    # Verifies library are installed
    assert wifi_install_dir.exists()
    assert ble_install_dir.exists()


def test_install_with_zip_path_multiple_libraries(run_command, downloads_dir, data_dir):
    assert run_command(["update"])

    env = {
        "ARDUINO_DATA_DIR": data_dir,
        "ARDUINO_DOWNLOADS_DIR": downloads_dir,
        "ARDUINO_SKETCHBOOK_DIR": data_dir,
        "ARDUINO_ENABLE_UNSAFE_LIBRARY_INSTALL": "true",
    }

    # Downloads zip to be installed later
    wifi_zip_path = Path(downloads_dir, "libraries", "WiFi101-0.16.1.zip")
    ble_zip_path = Path(downloads_dir, "libraries", "ArduinoBLE-1.1.3.zip")
    download_lib("https://github.com/arduino-libraries/WiFi101/archive/refs/tags/0.16.1.zip", wifi_zip_path)
    download_lib("https://github.com/arduino-libraries/ArduinoBLE/archive/refs/tags/1.1.3.zip", ble_zip_path)

    wifi_install_dir = Path(data_dir, "libraries", "WiFi101")
    ble_install_dir = Path(data_dir, "libraries", "ArduinoBLE")

    # Verifies libraries are not installed
    assert not wifi_install_dir.exists()
    assert not ble_install_dir.exists()

    assert run_command(["lib", "install", "--zip-path", wifi_zip_path, ble_zip_path], custom_env=env)

    # Verifies library are installed
    assert wifi_install_dir.exists()
    assert ble_install_dir.exists()


def test_lib_examples(run_command, data_dir):
    assert run_command(["update"])

    assert run_command(["lib", "install", "Arduino_JSON@0.1.0"])

    res = run_command(["lib", "examples", "Arduino_JSON", "--format", "json"])
    assert res.ok
    data = json.loads(res.stdout)
    assert len(data) == 1
    examples = data[0]["examples"]

    assert str(Path(data_dir, "libraries", "Arduino_JSON", "examples", "JSONArray")) in examples
    assert str(Path(data_dir, "libraries", "Arduino_JSON", "examples", "JSONKitchenSink")) in examples
    assert str(Path(data_dir, "libraries", "Arduino_JSON", "examples", "JSONObject")) in examples


def test_lib_examples_with_pde_file(run_command, data_dir):
    assert run_command(["update"])

    assert run_command(["lib", "install", "Encoder@1.4.1"])

    res = run_command(["lib", "examples", "Encoder", "--format", "json"])
    assert res.ok
    data = json.loads(res.stdout)
    assert len(data) == 1
    examples = data[0]["examples"]

    assert str(Path(data_dir, "libraries", "Encoder", "examples", "Basic")) in examples
    assert str(Path(data_dir, "libraries", "Encoder", "examples", "NoInterrupts")) in examples
    assert str(Path(data_dir, "libraries", "Encoder", "examples", "SpeedTest")) in examples
    assert str(Path(data_dir, "libraries", "Encoder", "examples", "TwoKnobs")) in examples


def test_lib_examples_with_case_mismatch(run_command, data_dir):
    assert run_command(["update"])

    assert run_command(["lib", "install", "WiFiManager@2.0.3-alpha"])

    res = run_command(["lib", "examples", "WiFiManager", "--format", "json"])
    assert res.ok
    data = json.loads(res.stdout)
    assert len(data) == 1
    examples = data[0]["examples"]

    assert len(examples) == 14

    examples_path = Path(data_dir, "libraries", "WiFiManager", "examples")
    # Verifies sketches with correct casing are listed
    assert str(examples_path / "Advanced") in examples
    assert str(examples_path / "AutoConnect" / "AutoConnectWithFeedbackLED") in examples
    assert str(examples_path / "AutoConnect" / "AutoConnectWithFSParameters") in examples
    assert str(examples_path / "AutoConnect" / "AutoConnectWithFSParametersAndCustomIP") in examples
    assert str(examples_path / "Basic") in examples
    assert str(examples_path / "DEV" / "OnDemandConfigPortal") in examples
    assert str(examples_path / "NonBlocking" / "AutoConnectNonBlocking") in examples
    assert str(examples_path / "NonBlocking" / "AutoConnectNonBlockingwParams") in examples
    assert str(examples_path / "Old_examples" / "AutoConnectWithFeedback") in examples
    assert str(examples_path / "Old_examples" / "AutoConnectWithReset") in examples
    assert str(examples_path / "Old_examples" / "AutoConnectWithStaticIP") in examples
    assert str(examples_path / "Old_examples" / "AutoConnectWithTimeout") in examples
    assert str(examples_path / "OnDemand" / "OnDemandConfigPortal") in examples
    assert str(examples_path / "ParamsChildClass") in examples

    # Verifies sketches with wrong casing are not returned
    assert str(examples_path / "NonBlocking" / "OnDemandNonBlocking") not in examples
    assert str(examples_path / "OnDemand" / "OnDemandWebPortal") not in examples


def test_lib_list_using_library_with_invalid_version(run_command, data_dir):
    assert run_command(["update"])

    # Install a library
    assert run_command(["lib", "install", "WiFi101@0.16.1"])

    # Verifies library is correctly returned
    res = run_command(["lib", "list", "--format", "json"])
    assert res.ok
    data = json.loads(res.stdout)
    assert len(data) == 1
    assert "0.16.1" == data[0]["library"]["version"]

    # Changes the version of the currently installed library so that it's
    # invalid
    lib_path = Path(data_dir, "libraries", "WiFi101")
    Path(lib_path, "library.properties").write_text("name=WiFi101\nversion=1.0001")

    # Verifies version is now empty
    res = run_command(["lib", "list", "--format", "json"])
    assert res.ok
    data = json.loads(res.stdout)
    assert len(data) == 1
    assert "version" not in data[0]["library"]


def test_lib_upgrade_using_library_with_invalid_version(run_command, data_dir):
    assert run_command(["update"])

    # Install a library
    assert run_command(["lib", "install", "WiFi101@0.16.1"])

    # Verifies library is correctly returned
    res = run_command(["lib", "list", "--format", "json"])
    assert res.ok
    data = json.loads(res.stdout)
    assert len(data) == 1
    assert "0.16.1" == data[0]["library"]["version"]

    # Changes the version of the currently installed library so that it's
    # invalid
    lib_path = Path(data_dir, "libraries", "WiFi101")
    Path(lib_path, "library.properties").write_text("name=WiFi101\nversion=1.0001")

    # Verifies version is now empty
    res = run_command(["lib", "list", "--format", "json"])
    assert res.ok
    data = json.loads(res.stdout)
    assert len(data) == 1
    assert "version" not in data[0]["library"]

    # Upgrade library
    assert run_command(["lib", "upgrade", "WiFi101"])

    # Verifies library has been updated
    res = run_command(["lib", "list", "--format", "json"])
    assert res.ok
    data = json.loads(res.stdout)
    assert len(data) == 1
    assert "" != data[0]["library"]["version"]


def test_install_zip_lib_with_macos_metadata(run_command, data_dir, downloads_dir):
    # Initialize configs to enable --zip-path flag
    env = {
        "ARDUINO_DATA_DIR": data_dir,
        "ARDUINO_DOWNLOADS_DIR": downloads_dir,
        "ARDUINO_SKETCHBOOK_DIR": data_dir,
        "ARDUINO_ENABLE_UNSAFE_LIBRARY_INSTALL": "true",
    }
    assert run_command(["config", "init", "--dest-dir", "."], custom_env=env)

    lib_install_dir = Path(data_dir, "libraries", "fake-lib")
    # Verifies library is not already installed
    assert not lib_install_dir.exists()

    zip_path = Path(__file__).parent / "testdata" / "fake-lib.zip"
    # Test zip-path install
    res = run_command(["lib", "install", "--zip-path", zip_path])
    assert res.ok
    assert "--git-url and --zip-path flags allow installing untrusted files, use it at your own risk." in res.stdout

    # Verifies library is installed in expected path
    assert lib_install_dir.exists()
    files = list(lib_install_dir.glob("**/*"))
    assert lib_install_dir / "library.properties" in files
    assert lib_install_dir / "src" / "fake-lib.h" in files

    # Reinstall library
    assert run_command(["lib", "install", "--zip-path", zip_path])

    # Verifies library remains installed
    assert lib_install_dir.exists()
    files = list(lib_install_dir.glob("**/*"))
    assert lib_install_dir / "library.properties" in files
    assert lib_install_dir / "src" / "fake-lib.h" in files


def test_install_zip_invalid_library(run_command, data_dir, downloads_dir):
    # Initialize configs to enable --zip-path flag
    env = {
        "ARDUINO_DATA_DIR": data_dir,
        "ARDUINO_DOWNLOADS_DIR": downloads_dir,
        "ARDUINO_SKETCHBOOK_DIR": data_dir,
        "ARDUINO_ENABLE_UNSAFE_LIBRARY_INSTALL": "true",
    }
    assert run_command(["config", "init", "--dest-dir", "."], custom_env=env)

    lib_install_dir = Path(data_dir, "libraries", "lib-without-header")
    # Verifies library is not already installed
    assert not lib_install_dir.exists()

    zip_path = Path(__file__).parent / "testdata" / "lib-without-header.zip"
    # Test zip-path install
    res = run_command(["lib", "install", "--zip-path", zip_path])
    assert res.failed
    assert "library not valid" in res.stderr

    lib_install_dir = Path(data_dir, "libraries", "lib-without-properties")
    # Verifies library is not already installed
    assert not lib_install_dir.exists()

    zip_path = Path(__file__).parent / "testdata" / "lib-without-properties.zip"
    # Test zip-path install
    res = run_command(["lib", "install", "--zip-path", zip_path])
    assert res.failed
    assert "library not valid" in res.stderr


def test_install_git_invalid_library(run_command, data_dir, downloads_dir):
    # Initialize configs to enable --zip-path flag
    env = {
        "ARDUINO_DATA_DIR": data_dir,
        "ARDUINO_DOWNLOADS_DIR": downloads_dir,
        "ARDUINO_SKETCHBOOK_DIR": data_dir,
        "ARDUINO_ENABLE_UNSAFE_LIBRARY_INSTALL": "true",
    }
    assert run_command(["config", "init", "--dest-dir", "."], custom_env=env)

    # Create fake library repository
    repo_dir = Path(data_dir, "lib-without-header")
    with Repo.init(repo_dir) as repo:
        lib_properties = Path(repo_dir, "library.properties")
        lib_properties.touch()
        repo.index.add([str(lib_properties)])
        repo.index.commit("First commit")

    lib_install_dir = Path(data_dir, "libraries", "lib-without-header")
    # Verifies library is not already installed
    assert not lib_install_dir.exists()

    res = run_command(["lib", "install", "--git-url", repo_dir], custom_env=env)
    assert res.failed
    assert "library not valid" in res.stderr
    assert not lib_install_dir.exists()

    # Create another fake library repository
    repo_dir = Path(data_dir, "lib-without-properties")
    with Repo.init(repo_dir) as repo:
        lib_header = Path(repo_dir, "src", "lib-without-properties.h")
        lib_header.parent.mkdir(parents=True, exist_ok=True)
        lib_header.touch()
        repo.index.add([str(lib_header)])
        repo.index.commit("First commit")

    lib_install_dir = Path(data_dir, "libraries", "lib-without-properties")
    # Verifies library is not already installed
    assert not lib_install_dir.exists()

    res = run_command(["lib", "install", "--git-url", repo_dir], custom_env=env)
    assert res.failed
    assert "library not valid" in res.stderr
    assert not lib_install_dir.exists()


def test_upgrade_does_not_try_to_upgrade_bundled_core_libraries_in_sketchbook(run_command, data_dir):
    test_platform_name = "platform_with_bundled_library"
    platform_install_dir = Path(data_dir, "hardware", "arduino-beta-dev", test_platform_name)
    platform_install_dir.mkdir(parents=True)

    # Install platform in Sketchbook hardware dir
    shutil.copytree(
        Path(__file__).parent / "testdata" / test_platform_name,
        platform_install_dir,
        dirs_exist_ok=True,
    )

    assert run_command(["update"])

    # Install latest version of library identical to one
    # bundled with test platform
    assert run_command(["lib", "install", "USBHost"])

    res = run_command(["lib", "list", "--all", "--format", "json"])
    assert res.ok
    libs = json.loads(res.stdout)
    assert len(libs) == 2
    # Verify both libraries have the same name
    assert libs[0]["library"]["name"] == "USBHost"
    assert libs[1]["library"]["name"] == "USBHost"

    res = run_command(["lib", "upgrade"])
    assert res.ok
    # Empty output means nothing has been updated as expected
    assert res.stdout == ""


def test_upgrade_does_not_try_to_upgrade_bundled_core_libraries(run_command, data_dir):
    test_platform_name = "platform_with_bundled_library"
    platform_install_dir = Path(data_dir, "packages", "arduino", "hardware", "arch", "4.2.0")
    platform_install_dir.mkdir(parents=True)

    # Simulate installation of a platform with arduino-cli
    shutil.copytree(
        Path(__file__).parent / "testdata" / test_platform_name,
        platform_install_dir,
        dirs_exist_ok=True,
    )

    assert run_command(["update"])

    # Install latest version of library identical to one
    # bundled with test platform
    assert run_command(["lib", "install", "USBHost"])

    res = run_command(["lib", "list", "--all", "--format", "json"])
    assert res.ok
    libs = json.loads(res.stdout)
    assert len(libs) == 2
    # Verify both libraries have the same name
    assert libs[0]["library"]["name"] == "USBHost"
    assert libs[1]["library"]["name"] == "USBHost"

    res = run_command(["lib", "upgrade"])
    assert res.ok
    # Empty output means nothing has been updated as expected
    assert res.stdout == ""
