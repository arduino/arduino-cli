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
import os
import time

import pytest

from .common import running_on_ci, parse_json_traces

# Skip this module when running in CI environments
pytestmark = pytest.mark.skipif(running_on_ci(), reason="VMs have no serial ports")


def test_upload(run_command, data_dir, detected_boards):
    # Init the environment explicitly
    run_command("core update-index")

    for board in detected_boards:
        # Download core
        run_command(f"core install {board.core}")
        # Create a sketch
        sketch_name = "foo"
        sketch_path = os.path.join(data_dir, sketch_name)
        fqbn = board.fqbn
        address = board.address
        assert run_command(f"sketch new {sketch_path}")
        # Build sketch
        assert run_command(f"compile -b {fqbn} {sketch_path}")
        # Upload without port must fail
        result = run_command(f"upload -b {fqbn} {sketch_path}")
        assert result.failed
        # Upload
        res = run_command(f"upload -b {fqbn} -p {address} {sketch_path}")
        print(res.stderr)
        assert res

        # multiple uploads requires some pauses
        time.sleep(2)
        # Upload using --input-dir reusing standard sketch "build" folder artifacts
        fqbn_path = fqbn.replace(":", ".")
        assert run_command(f"upload -b {fqbn} -p {address} --input-dir {sketch_path}/build/{fqbn_path} {sketch_path}")

        # multiple uploads requires some pauses
        time.sleep(2)
        # Upload using --input-file reusing standard sketch "build" folder artifacts
        assert run_command(
            f"upload -b {fqbn} -p {address} --input-file {sketch_path}/build/{fqbn_path}/{sketch_name}.ino.bin"
        )


def test_upload_after_attach(run_command, data_dir, detected_boards):
    # Init the environment explicitly
    run_command("core update-index")

    for board in detected_boards:
        # Download core
        run_command(f"core install {board.core}")
        # Create a sketch
        sketch_path = os.path.join(data_dir, "foo")
        assert run_command("sketch new {}".format(sketch_path))
        assert run_command(
            "board attach serial://{port} {sketch_path}".format(port=board.address, sketch_path=sketch_path)
        )
        # Build sketch
        assert run_command("compile {sketch_path}".format(sketch_path=sketch_path))
        # Upload
        assert run_command("upload  {sketch_path}".format(sketch_path=sketch_path))


def test_compile_and_upload_combo(run_command, data_dir, detected_boards, wait_for_board):
    # Init the environment explicitly
    run_command("core update-index")

    # Install required core(s)
    run_command("core install arduino:avr@1.8.3")
    run_command("core install arduino:samd@1.8.6")

    # Create a test sketch
    sketch_name = "CompileAndUploadIntegrationTest"
    sketch_path = os.path.join(data_dir, sketch_name)
    sketch_main_file = os.path.join(sketch_path, sketch_name + ".ino")
    result = run_command("sketch new {}".format(sketch_path))
    assert result.ok
    assert "Sketch created in: {}".format(sketch_path) in result.stdout

    # Build sketch for each detected board
    for board in detected_boards:
        log_file_name = "{fqbn}-compile.log".format(fqbn=board.fqbn.replace(":", "-"))
        log_file_path = os.path.join(data_dir, log_file_name)
        command_log_flags = "--log-format json --log-file {} --log-level trace".format(log_file_path)

        def run_test(s):
            wait_for_board()
            result = run_command(f"compile -b {board.fqbn} --upload -p {board.address} {s} {command_log_flags}")
            print(result.stderr)
            assert result.ok

            # check from the logs if the bin file were uploaded on the current board
            log_json = open(log_file_path, "r")
            traces = parse_json_traces(log_json.readlines())
            assert f"Compile {sketch_path} for {board.fqbn} started" in traces
            assert f"Compile {sketch_name} for {board.fqbn} successful" in traces
            assert f"Upload {sketch_path} on {board.fqbn} started" in traces
            assert "Upload successful" in traces

        run_test(sketch_path)
        run_test(sketch_main_file)
