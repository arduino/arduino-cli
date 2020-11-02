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
from pathlib import Path

import pytest

from .common import running_on_ci, parse_json_traces

# Skip this module when running in CI environments
pytestmark = pytest.mark.skipif(running_on_ci(), reason="VMs have no serial ports")


def test_upload(run_command, data_dir, detected_boards):
    # Init the environment explicitly
    run_command("core update-index")

    for board in detected_boards:
        # Download platform
        run_command(f"core install {board.core}")
        # Create a sketch
        sketch_name = f"TestUploadSketch{board.id}"
        sketch_path = Path(data_dir, sketch_name)
        fqbn = board.fqbn
        address = board.address
        assert run_command(f"sketch new {sketch_path}")
        # Build sketch
        assert run_command(f"compile -b {fqbn} {sketch_path}")

        # Verifies binaries are not exported
        assert not (sketch_path / "build").exists()

        # Upload without port must fail
        assert not run_command(f"upload -b {fqbn} {sketch_path}")

        # Upload
        assert run_command(f"upload -b {fqbn} -p {address} {sketch_path}")


def test_upload_with_input_dir_flag(run_command, data_dir, detected_boards):
    # Init the environment explicitly
    run_command("core update-index")

    for board in detected_boards:
        # Download board platform
        run_command(f"core install {board.core}")

        # Create sketch
        sketch_name = f"TestUploadInputDirSketch{board.id}"
        sketch_path = Path(data_dir, sketch_name)
        fqbn = board.fqbn
        address = board.address
        assert run_command(f"sketch new {sketch_path}")

        # Build sketch and export binaries to custom directory
        output_dir = Path(data_dir, "test_dir", sketch_name, "build")
        assert run_command(f"compile -b {fqbn} {sketch_path} --output-dir {output_dir}")

        # Upload with --input-dir flag
        assert run_command(f"upload -b {fqbn} -p {address} --input-dir {output_dir} {sketch_path}")


def test_upload_with_input_file_flag(run_command, data_dir, detected_boards):
    # Init the environment explicitly
    run_command("core update-index")

    for board in detected_boards:
        # Download board platform
        run_command(f"core install {board.core}")

        # Create sketch
        sketch_name = f"TestUploadInputFileSketch{board.id}"
        sketch_path = Path(data_dir, sketch_name)
        fqbn = board.fqbn
        address = board.address
        assert run_command(f"sketch new {sketch_path}")

        # Build sketch and export binaries to custom directory
        output_dir = Path(data_dir, "test_dir", sketch_name, "build")
        assert run_command(f"compile -b {fqbn} {sketch_path} --output-dir {output_dir}")

        # We don't need a specific file when using the --input-file flag to upload since
        # it's just used to calculate the directory, so it's enough to get a random file
        # that's inside that directory
        input_file = next(output_dir.glob(f"{sketch_name}.ino.*"))
        # Upload using --input-file
        assert run_command(f"upload -b {fqbn} -p {address} --input-file {input_file}")


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


def test_compile_and_upload_combo_with_custom_build_path(run_command, data_dir, detected_boards, wait_for_board):
    # Init the environment explicitly
    run_command("core update-index")

    # Install required core(s)
    run_command("core install arduino:avr@1.8.3")
    run_command("core install arduino:samd@1.8.6")

    sketch_name = "CompileAndUploadCustomBuildPathIntegrationTest"
    sketch_path = Path(data_dir, sketch_name)
    assert run_command(f"sketch new {sketch_path}")

    for board in detected_boards:
        fqbn_normalized = board.fqbn.replace(":", "-")
        log_file_name = f"{fqbn_normalized}-compile.log"
        log_file = Path(data_dir, log_file_name)
        command_log_flags = f"--log-format json --log-file {log_file} --log-level trace"

        wait_for_board()

        build_path = Path(data_dir, "test_dir", fqbn_normalized, "build_dir")
        result = run_command(
            f"compile -b {board.fqbn} "
            + f"--upload -p {board.address} "
            + f"--build-path {build_path} "
            + f"{sketch_path} {command_log_flags}"
        )
        print(result.stderr)
        assert result.ok

        # check from the logs if the bin file were uploaded on the current board
        log_json = open(log_file, "r")
        traces = parse_json_traces(log_json.readlines())
        assert f"Compile {sketch_path} for {board.fqbn} started" in traces
        assert f"Compile {sketch_name} for {board.fqbn} successful" in traces
        assert f"Upload {sketch_path} on {board.fqbn} started" in traces
        assert "Upload successful" in traces
