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
import shutil
import json
from pathlib import Path

import pytest

from .common import running_on_ci, parse_json_traces

# Skip this module when running in CI environments
pytestmark = pytest.mark.skipif(running_on_ci(), reason="VMs have no serial ports")


def test_upload_after_attach(run_command, data_dir, detected_boards):
    # Init the environment explicitly
    run_command(["core", "update-index"])

    for board in detected_boards:
        # Download core
        run_command(["core", "install", board.core])
        # Create a sketch
        sketch_path = os.path.join(data_dir, "foo")
        assert run_command(["sketch", "new", sketch_path])
        assert run_command(["board", "attach", "-p", board.address, sketch_path])
        # Build sketch
        assert run_command(["compile", sketch_path])
        # Upload
        assert run_command(["upload", sketch_path])


def test_compile_and_upload_combo_sketch_with_pde_extension(run_command, data_dir, detected_boards, wait_for_board):
    assert run_command(["update"])

    sketch_name = "CompileAndUploadPdeSketch"
    sketch_path = Path(data_dir, sketch_name)

    # Create a test sketch
    assert run_command(["sketch", "new", sketch_path])

    # Renames sketch file to pde
    sketch_file = Path(sketch_path, f"{sketch_name}.ino").rename(sketch_path / f"{sketch_name}.pde")

    for board in detected_boards:
        # Install core
        core = ":".join(board.fqbn.split(":")[:2])
        assert run_command(["core", "install", core])

        # Build sketch and upload from folder
        wait_for_board()
        res = run_command(["compile", "--clean", "-b", board.fqbn, "-u", "-p", board.address, sketch_path])
        assert res.ok
        assert "Sketches with .pde extension are deprecated, please rename the following files to .ino" in res.stderr
        assert str(sketch_file) in res.stderr

        # Build sketch and upload from file
        wait_for_board()
        res = run_command(["compile", "--clean", "-b", board.fqbn, "-u", "-p", board.address, sketch_file])
        assert res.ok
        assert "Sketches with .pde extension are deprecated, please rename the following files to .ino" in res.stderr
        assert str(sketch_file) in res.stderr


def test_upload_sketch_with_pde_extension(run_command, data_dir, detected_boards, wait_for_board):
    assert run_command(["update"])

    sketch_name = "UploadPdeSketch"
    sketch_path = Path(data_dir, sketch_name)

    # Create a test sketch
    assert run_command(["sketch", "new", sketch_path])

    # Renames sketch file to pde
    sketch_file = Path(sketch_path, f"{sketch_name}.ino").rename(sketch_path / f"{sketch_name}.pde")

    for board in detected_boards:
        # Install core
        core = ":".join(board.fqbn.split(":")[:2])
        assert run_command(["core", "install", core])

        # Compile sketch first
        res = run_command(["compile", "--clean", "-b", board.fqbn, sketch_path, "--format", "json"])
        assert res.ok
        data = json.loads(res.stdout)
        build_dir = Path(data["builder_result"]["build_path"])

        # Upload from sketch folder
        wait_for_board()
        assert run_command(["upload", "-b", board.fqbn, "-p", board.address, sketch_path])

        # Upload from sketch file
        wait_for_board()
        assert run_command(["upload", "-b", board.fqbn, "-p", board.address, sketch_file])

        wait_for_board()
        res = run_command(["upload", "-b", board.fqbn, "-p", board.address, "--input-dir", build_dir])
        assert (
            "Sketches with .pde extension are deprecated, please rename the following files to .ino:" not in res.stderr
        )

        # Upload from binary file
        wait_for_board()
        # We don't need a specific file when using the --input-file flag to upload since
        # it's just used to calculate the directory, so it's enough to get a random file
        # that's inside that directory
        binary_file = next(build_dir.glob(f"{sketch_name}.pde.*"))
        res = run_command(["upload", "-b", board.fqbn, "-p", board.address, "--input-file", binary_file])
        assert (
            "Sketches with .pde extension are deprecated, please rename the following files to .ino:" not in res.stderr
        )


def test_upload_with_input_dir_containing_multiple_binaries(run_command, data_dir, detected_boards, wait_for_board):
    # This tests verifies the behaviour outlined in this issue:
    # https://github.com/arduino/arduino-cli/issues/765#issuecomment-699678646
    assert run_command(["update"])

    # Create a two different sketches
    sketch_one_name = "UploadMultipleBinariesSketchOne"
    sketch_one_path = Path(data_dir, sketch_one_name)
    assert run_command(["sketch", "new", sketch_one_path])

    sketch_two_name = "UploadMultipleBinariesSketchTwo"
    sketch_two_path = Path(data_dir, sketch_two_name)
    assert run_command(["sketch", "new", sketch_two_path])

    for board in detected_boards:
        # Install core
        core = ":".join(board.fqbn.split(":")[:2])
        assert run_command(["core", "install", core])

        # Compile both sketches and copy binaries in the same directory same build directory
        res = run_command(["compile", "--clean", "-b", board.fqbn, sketch_one_path, "--format", "json"])
        assert res.ok
        data = json.loads(res.stdout)
        build_dir_one = Path(data["builder_result"]["build_path"])
        res = run_command(["compile", "--clean", "-b", board.fqbn, sketch_two_path, "--format", "json"])
        assert res.ok
        data = json.loads(res.stdout)
        build_dir_two = Path(data["builder_result"]["build_path"])

        # Copy binaries to same folder
        binaries_dir = Path(data_dir, "build", "BuiltBinaries")
        shutil.copytree(build_dir_one, binaries_dir, dirs_exist_ok=True)
        shutil.copytree(build_dir_two, binaries_dir, dirs_exist_ok=True)

        wait_for_board()
        # Verifies upload fails because multiple binaries are found
        res = run_command(["upload", "-b", board.fqbn, "-p", board.address, "--input-dir", binaries_dir])
        assert res.failed
        assert (
            "Error during Upload: "
            + "Error finding build artifacts: "
            + "autodetect build artifact: "
            + "multiple build artifacts found:"
            in res.stderr
        )

        # Copy binaries to folder with same name of a sketch
        binaries_dir = Path(data_dir, "build", "UploadMultipleBinariesSketchOne")
        shutil.copytree(build_dir_one, binaries_dir, dirs_exist_ok=True)
        shutil.copytree(build_dir_two, binaries_dir, dirs_exist_ok=True)

        wait_for_board()
        # Verifies upload is successful using the binaries with the same name of the containing folder
        res = run_command(["upload", "-b", board.fqbn, "-p", board.address, "--input-dir", binaries_dir])
        assert (
            "Sketches with .pde extension are deprecated, please rename the following files to .ino:" not in res.stderr
        )


def test_compile_and_upload_combo_sketch_with_mismatched_casing(run_command, data_dir, detected_boards, wait_for_board):
    assert run_command(["update"])

    # Create a sketch
    sketch_name = "CompileUploadComboMismatchCasing"
    sketch_path = Path(data_dir, sketch_name)
    assert run_command(["sketch", "new", sketch_path])

    # Rename main .ino file so casing is different from sketch name
    Path(sketch_path, f"{sketch_name}.ino").rename(sketch_path / f"{sketch_name.lower()}.ino")

    for board in detected_boards:
        # Install core
        core = ":".join(board.fqbn.split(":")[:2])
        assert run_command(["core", "install", core])

        # Try to compile
        res = run_command(["compile", "--clean", "-b", board.fqbn, "-u", "-p", board.address, sketch_path])
        assert res.failed
        assert "Error opening sketch:" in res.stderr


def test_upload_sketch_with_mismatched_casing(run_command, data_dir, detected_boards, wait_for_board):
    assert run_command(["update"])

    # Create a sketch
    sketch_name = "UploadMismatchCasing"
    sketch_path = Path(data_dir, sketch_name)
    assert run_command(["sketch", "new", sketch_path])

    # Rename main .ino file so casing is different from sketch name
    Path(sketch_path, f"{sketch_name}.ino").rename(sketch_path / f"{sketch_name.lower()}.ino")

    for board in detected_boards:
        # Install core
        core = ":".join(board.fqbn.split(":")[:2])
        assert run_command(["core", "install", core])

        # Tries to upload given sketch, it has not been compiled but it fails even before
        # searching for binaries since the sketch is not valid
        res = run_command(["upload", "-b", board.fqbn, "-p", board.address, sketch_path])
        assert res.failed
        assert "Error during Upload:" in res.stderr


def test_upload_to_port_with_board_autodetect(run_command, data_dir, detected_boards):
    assert run_command(["update"])

    # Create a sketch
    sketch_name = "SketchSimple"
    sketch_path = Path(data_dir, sketch_name)
    assert run_command(["sketch", "new", sketch_path])

    for board in detected_boards:
        # Install core
        core = ":".join(board.fqbn.split(":")[:2])
        assert run_command(["core", "install", core])

        assert run_command(["compile", "-b", board.fqbn, sketch_path])

        res = run_command(["upload", "-p", board.address, sketch_path])
        assert res.ok


def test_compile_and_upload_to_port_with_board_autodetect(run_command, data_dir, detected_boards):
    assert run_command(["update"])

    # Create a sketch
    sketch_name = "SketchSimple"
    sketch_path = Path(data_dir, sketch_name)
    assert run_command(["sketch", "new", sketch_path])

    for board in detected_boards:
        # Install core
        core = ":".join(board.fqbn.split(":")[:2])
        assert run_command(["core", "install", core])

        res = run_command(["compile", "-u", "-p", board.address, sketch_path])
        assert res.ok
