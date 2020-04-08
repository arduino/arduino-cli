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
import simplejson as json


def test_list(run_command):
    # Init the environment explicitly
    assert run_command("core update-index")

    # When output is empty, nothing is printed out, no matter the output format
    result = run_command("lib list")
    assert result.ok
    assert "" == result.stderr
    assert "No libraries installed." == result.stdout.strip()
    result = run_command("lib list --format json")
    assert result.ok
    assert "" == result.stderr
    assert 0 == len(json.loads(result.stdout))

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

    # Test failing-install of library with wrong dependency
    # (https://github.com/arduino/arduino-cli/issues/534)
    result = run_command("lib install MD_Parola@3.2.0")
    assert (
        "Error resolving dependencies for MD_Parola@3.2.0: dependency 'MD_MAX72xx' is not available"
        in result.stderr
    )

def test_update_index(run_command):
    result = run_command("lib update-index")
    assert result.ok
    assert (
        "Updating index: library_index.json downloaded"
        == result.stdout.splitlines()[-1].strip()
    )


def test_uninstall(run_command):
    libs = ['"AzureIoTProtocol_MQTT"', '"WiFiNINA"']
    assert run_command("lib install {}".format(" ".join(libs)))

    result = run_command("lib uninstall {}".format(" ".join(libs)))
    assert result.ok

def test_uninstall_spaces(run_command):
    key = '"LiquidCrystal I2C"'
    assert run_command("lib install {}".format(key))
    assert run_command("lib uninstall {}".format(key))
    result = run_command("lib list --format json")
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
    key = 'pcm'
    assert run_command("lib install {}".format(key))
    assert run_command("lib uninstall {}".format(key))
    result = run_command("lib list --format json")
    assert result.ok
    assert len(json.loads(result.stdout)) == 0

def test_search(run_command):
    assert run_command("lib update-index")

    result = run_command("lib search --names")
    assert result.ok
    out_lines = result.stdout.strip().splitlines()
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
    assert len(libs_json.get("libraries")) >= 1


def test_search_paragraph(run_command):
    """
    Search for a string that's only present in the `paragraph` field
    within the index file.
    """
    assert run_command("lib update-index")
    result = run_command(
        'lib search "A simple and efficient JSON library" --format json'
    )
    assert result.ok
    libs_json = json.loads(result.stdout)
    assert 1 == len(libs_json.get("libraries"))
