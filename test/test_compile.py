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
import pytest
import json
import os

from .common import running_on_ci


def test_compile_without_fqbn(run_command):
    # Init the environment explicitly
    result = run_command("core update-index")
    assert result.ok

    # Download latest AVR
    result = run_command("core install arduino:avr")
    assert result.ok

    # Build sketch without FQBN
    result = run_command("compile")
    assert result.failed


def test_compile_with_simple_sketch(run_command, data_dir):
    # Init the environment explicitly
    result = run_command("core update-index")
    assert result.ok

    # # Download latest AVR
    result = run_command("core install arduino:avr")
    assert result.ok

    sketch_name = "CompileIntegrationTest"
    sketch_path = os.path.join(data_dir, sketch_name)
    fqbn = "arduino:avr:uno"

    # Create a test sketch
    result = run_command("sketch new {}".format(sketch_name))
    assert result.ok
    assert "Sketch created in: {}".format(sketch_path) in result.stdout

    # Build sketch for arduino:avr:uno
    log_file_name = "compile.log"
    log_file_path = os.path.join(data_dir, log_file_name)
    result = run_command(
        "compile -b {fqbn} {sketch_path} --log-format json --log-file {log_file} --log-level trace".format(
            fqbn=fqbn, sketch_path=sketch_path, log_file=log_file_path))
    assert result.ok

    # let's test from the logs if the hex file produced by successful compile is moved to our sketch folder
    log_json = open(log_file_path, 'r')
    json_log_lines = log_json.readlines()
    assert is_message_in_json_log_lines("Executing `arduino compile`", json_log_lines)
    assert is_message_in_json_log_lines(
        "Compile {sketch} for {fqbn} successful".format(sketch=sketch_name,
                                                        fqbn=fqbn),
        json_log_lines)


@pytest.mark.skipif(running_on_ci(), reason="VMs have no serial ports")
def test_compile_and_compile_combo(run_command, data_dir):
    # Init the environment explicitly
    result = run_command("core update-index")
    assert result.ok

    # Install required core(s)
    result = run_command("core install arduino:avr")
    # result = run_command("core install arduino:samd")
    assert result.ok

    # Create a test sketch
    sketch_name = "CompileAndUploadIntegrationTest"
    sketch_path = os.path.join(data_dir, sketch_name)
    result = run_command("sketch new CompileAndUploadIntegrationTest")
    assert result.ok
    assert "Sketch created in: {}".format(sketch_path) in result.stdout

    #
    # Build a list of detected boards to test, if any.
    #
    result = run_command("board list --format json")
    assert result.ok

    #
    # The `board list --format json` returns a JSON that looks like to the following:
    #
    # [
    #     {
    #       "address": "/dev/cu.usbmodem14201",
    #       "protocol": "serial",
    #       "protocol_label": "Serial Port (USB)",
    #       "boards": [
    #         {
    #           "name": "Arduino NANO 33 IoT",
    #           "FQBN": "arduino:samd:nano_33_iot"
    #         }
    #       ]
    #     }
    #   ]

    detected_boards = []

    ports = json.loads(result.stdout)
    assert isinstance(ports, list)
    for port in ports:
        boards = port.get('boards')
        assert isinstance(boards, list)
        for board in boards:
            detected_boards.append(dict(address=port.get('address'), fqbn=board.get('FQBN')))

    assert len(detected_boards) >= 1, "There are no boards available for testing"

    # Build sketch for each detected board
    for board in detected_boards:
        log_file_name = "{fqbn}-compile.log".format(fqbn=board.get('fqbn'))
        log_file_path = os.path.join(data_dir, log_file_name)
        result = run_command(
            "compile -b {fqbn} --upload -p {address} {sketch_path} --log-format json --log-file {log_file} --log-level trace".format(
                fqbn=board.get('fqbn'),
                address=board.get('address'),
                sketch_path=sketch_path,
                log_file=log_file_path
            )
        )
        assert result.ok
        # check from the logs if the bin file were uploaded on the current board
        log_json = open(log_file_path, 'r')
        json_log_lines = log_json.readlines()
        assert is_message_in_json_log_lines("Executing `arduino compile`", json_log_lines)
        assert is_message_in_json_log_lines(
            "Compile {sketch} for {fqbn} successful".format(sketch=sketch_name,
                                                            fqbn=board.get(
                                                                'fqbn')),
            json_log_lines)
        assert is_message_in_json_log_lines("Executing `arduino upload`", json_log_lines)
        assert is_message_in_json_log_lines(
            "Upload {sketch} on {fqbn} successful".format(sketch=sketch_name,
                                                          fqbn=board.get(
                                                              'fqbn')),
            json_log_lines)


def is_message_in_json_log_lines(message, log_json_lines):
    return len([index for index, entry in enumerate(log_json_lines) if json.loads(entry).get("msg") == message]) == 1
