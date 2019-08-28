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
import simplejson as json


def test_list(run_command):
    # Init the environment explicitly
    assert run_command("core update-index")

    # When ouput is empty, nothing is printed out, no matter the output format
    result = run_command("lib list")
    assert result.ok
    assert "" == result.stderr
    assert "" == result.stdout
    result = run_command("lib list --format json")
    assert result.ok
    assert "" == result.stderr
    assert "" == result.stdout

    # Install something we can list at a version older than latest
    result = run_command("lib install ArduinoJson@6.11.0")
    assert result.ok

    # Look at the plain text output
    result = run_command("lib list")
    assert result.ok
    assert "" == result.stderr
    lines = result.stdout.strip().splitlines()
    assert 2 == len(lines)
    toks = [t.strip() for t in lines[1].split()]
    # be sure line contain the current version AND the available version
    assert "" != toks[1]
    assert "" != toks[2]

    # Look at the JSON output
    result = run_command("lib list --format json")
    assert result.ok
    assert "" == result.stderr
    data = json.loads(result.stdout)
    assert 1 == len(data)
    # be sure data contains the available version
    assert "" != data[0]["release"]["version"]


def test_install(run_command):
    libs = ['"AzureIoTProtocol_MQTT"', '"CMMC MQTT Connector"', '"WiFiNINA"']
    # Should be safe to run install multiple times
    assert run_command("lib install {}".format(" ".join(libs)))
    assert run_command("lib install {}".format(" ".join(libs)))


def test_update_index(run_command):
    result = run_command("lib update-index")
    assert result.ok
    assert (
        "Updating index: library_index.json downloaded"
        == result.stdout.splitlines()[-1].strip()
    )


def test_remove(run_command):
    libs = ['"AzureIoTProtocol_MQTT"', '"CMMC MQTT Connector"', '"WiFiNINA"']
    assert run_command("lib install {}".format(" ".join(libs)))

    result = run_command("lib uninstall {}".format(" ".join(libs)))
    assert result.ok


def test_search(run_command):
    assert run_command("lib update-index")

    result = run_command("lib search --names")
    assert result.ok
    out_lines = result.stdout.splitlines()
    # Create an array with just the name of the vars
    libs = []
    for line in out_lines:
        start = line.find('"') + 1
        libs.append(line[start:-1])

    expected = {"WiFi101", "WiFi101OTA", "Firebase Arduino based on WiFi101"}
    assert expected == {lib for lib in libs if "WiFi101" in lib}

    result = run_command("lib search --names --format json")
    assert result.ok
    libs_json = json.loads(result.stdout)
    assert len(libs) == len(libs_json.get("libraries"))

    result = run_command("lib search")
    assert result.ok

    # Search for a specific target
    result = run_command("lib search ArduinoJson --format json")
    assert result.ok
    libs_json = json.loads(result.stdout)
    assert 1 == len(libs_json.get("libraries"))
