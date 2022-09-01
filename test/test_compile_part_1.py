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
import platform
import tempfile
import hashlib
from pathlib import Path
import simplejson as json

import pytest

from .common import running_on_ci


def test_compile_with_simple_sketch(run_command, data_dir, working_dir):
    # Init the environment explicitly
    run_command(["core", "update-index"])

    # Download latest AVR
    run_command(["core", "install", "arduino:avr"])

    sketch_name = "CompileIntegrationTest"
    sketch_path = Path(data_dir, sketch_name)
    fqbn = "arduino:avr:uno"

    # Create a test sketch
    result = run_command(["sketch", "new", sketch_path])
    assert result.ok
    assert f"Sketch created in: {sketch_path}" in result.stdout

    # Build sketch for arduino:avr:uno
    result = run_command(["compile", "-b", fqbn, sketch_path])
    assert result.ok

    # Build sketch for arduino:avr:uno with json output
    result = run_command(["compile", "-b", fqbn, sketch_path, "--format", "json"])
    assert result.ok
    # check is a valid json and contains requested data
    compile_output = json.loads(result.stdout)
    assert compile_output["compiler_out"] != ""
    assert compile_output["compiler_err"] == ""

    # Verifies expected binaries have been built
    sketch_path_md5 = hashlib.md5(bytes(sketch_path)).hexdigest().upper()
    build_dir = Path(tempfile.gettempdir(), f"arduino-sketch-{sketch_path_md5}")
    assert (build_dir / f"{sketch_name}.ino.eep").exists()
    assert (build_dir / f"{sketch_name}.ino.elf").exists()
    assert (build_dir / f"{sketch_name}.ino.hex").exists()
    assert (build_dir / f"{sketch_name}.ino.with_bootloader.bin").exists()
    assert (build_dir / f"{sketch_name}.ino.with_bootloader.hex").exists()

    # Verifies binaries are not exported by default to Sketch folder
    sketch_build_dir = Path(sketch_path, "build", fqbn.replace(":", "."))
    assert not (sketch_build_dir / f"{sketch_name}.ino.eep").exists()
    assert not (sketch_build_dir / f"{sketch_name}.ino.elf").exists()
    assert not (sketch_build_dir / f"{sketch_name}.ino.hex").exists()
    assert not (sketch_build_dir / f"{sketch_name}.ino.with_bootloader.bin").exists()
    assert not (sketch_build_dir / f"{sketch_name}.ino.with_bootloader.hex").exists()


@pytest.mark.skipif(
    running_on_ci() and platform.system() == "Windows",
    reason="Test disabled on Github Actions Win VM until tmpdir inconsistent behavior bug is fixed",
)
def test_output_flag_default_path(run_command, data_dir, working_dir):
    # Init the environment explicitly
    run_command(["core", "update-index"])

    # Install Arduino AVR Boards
    run_command(["core", "install", "arduino:avr@1.8.3"])

    # Create a test sketch
    sketch_path = os.path.join(data_dir, "test_output_flag_default_path")
    fqbn = "arduino:avr:uno"
    result = run_command(["sketch", "new", sketch_path])
    assert result.ok

    # Test the --output-dir flag defaulting to current working dir
    result = run_command(["compile", "-b", fqbn, sketch_path, "--output-dir", "test"])
    assert result.ok
    target = os.path.join(working_dir, "test")
    assert os.path.exists(target) and os.path.isdir(target)


def test_compile_with_sketch_with_symlink_selfloop(run_command, data_dir):
    # Init the environment explicitly
    run_command(["core", "update-index"])

    # Install Arduino AVR Boards
    run_command(["core", "install", "arduino:avr@1.8.3"])

    sketch_name = "CompileIntegrationTestSymlinkSelfLoop"
    sketch_path = os.path.join(data_dir, sketch_name)
    fqbn = "arduino:avr:uno"

    # Create a test sketch
    result = run_command(["sketch", "new", sketch_path])
    assert result.ok
    assert "Sketch created in: {}".format(sketch_path) in result.stdout

    # create a symlink that loops on himself
    loop_file_path = os.path.join(sketch_path, "loop")
    os.symlink(loop_file_path, loop_file_path)

    # Build sketch for arduino:avr:uno
    result = run_command(["compile", "-b", fqbn, sketch_path])
    # The assertion is a bit relaxed in this case because win behaves differently from macOs and linux
    # returning a different error detailed message
    assert "Error opening sketch:" in result.stderr
    assert not result.ok

    sketch_name = "CompileIntegrationTestSymlinkDirLoop"
    sketch_path = os.path.join(data_dir, sketch_name)
    fqbn = "arduino:avr:uno"

    # Create a test sketch
    result = run_command(["sketch", "new", sketch_path])
    assert result.ok
    assert "Sketch created in: {}".format(sketch_path) in result.stdout

    # create a symlink that loops on the upper level
    loop_dir_path = os.path.join(sketch_path, "loop_dir")
    os.mkdir(loop_dir_path)
    loop_dir_symlink_path = os.path.join(loop_dir_path, "loop_dir_symlink")
    os.symlink(loop_dir_path, loop_dir_symlink_path)

    # Build sketch for arduino:avr:uno
    result = run_command(["compile", "-b", fqbn, sketch_path])
    # The assertion is a bit relaxed in this case because win behaves differently from macOs and linux
    # returning a different error detailed message
    assert "Error opening sketch:" in result.stderr
    assert not result.ok


def test_compile_blacklisted_sketchname(run_command, data_dir):
    """
    Compile should ignore folders named `RCS`, `.git` and the likes, but
    it should be ok for a sketch to be named like RCS.ino
    """
    # Init the environment explicitly
    run_command(["core", "update-index"])

    # Install Arduino AVR Boards
    run_command(["core", "install", "arduino:avr@1.8.3"])

    sketch_name = "RCS"
    sketch_path = os.path.join(data_dir, sketch_name)
    fqbn = "arduino:avr:uno"

    # Create a test sketch
    result = run_command(["sketch", "new", sketch_path])
    assert result.ok
    assert "Sketch created in: {}".format(sketch_path) in result.stdout

    # Build sketch for arduino:avr:uno
    result = run_command(["compile", "-b", fqbn, sketch_path])
    assert result.ok


def test_compile_without_precompiled_libraries(run_command, data_dir):
    # Init the environment explicitly
    url = "https://adafruit.github.io/arduino-board-index/package_adafruit_index.json"
    assert run_command(["core", "update-index", f"--additional-urls={url}"])
    assert run_command(["core", "install", "arduino:mbed@1.3.1", f"--additional-urls={url}"])

    #    # Precompiled version of Arduino_TensorflowLite
    #    assert run_command(["lib", "install", "Arduino_LSM9DS1"])
    #    assert run_command(["lib", "install", "Arduino_TensorflowLite@2.1.1-ALPHA-precompiled"])
    #
    #    sketch_path = Path(data_dir, "libraries", "Arduino_TensorFlowLite", "examples", "hello_world")
    #    assert run_command(["compile", "-b", "arduino:mbed:nano33ble", sketch_path])

    assert run_command(["core", "install", "arduino:samd@1.8.7", f"--additional-urls={url}"])
    #    assert run_command(["core", "install", "adafruit:samd@1.6.4", f"--additional-urls={url}"])
    #    # should work on adafruit too after https://github.com/arduino/arduino-cli/pull/1134
    #    assert run_command(["compile", "-b", "adafruit:samd:adafruit_feather_m4", sketch_path])
    #
    #    # Non-precompiled version of Arduino_TensorflowLite
    #    assert run_command(["lib", "install", "Arduino_TensorflowLite@2.1.0-ALPHA"])
    #    assert run_command(["compile", "-b", "arduino:mbed:nano33ble", sketch_path])
    #    assert run_command(["compile", "-b", "adafruit:samd:adafruit_feather_m4", sketch_path])

    # Bosch sensor library
    assert run_command(["lib", "install", "BSEC Software Library@1.5.1474"])
    sketch_path = Path(data_dir, "libraries", "BSEC_Software_Library", "examples", "basic")
    assert run_command(["compile", "-b", "arduino:samd:mkr1000", sketch_path])
    assert run_command(["compile", "-b", "arduino:mbed:nano33ble", sketch_path])

    # USBBlaster library
    assert run_command(["lib", "install", "USBBlaster@1.0.0"])
    sketch_path = Path(data_dir, "libraries", "USBBlaster", "examples", "USB_Blaster")
    assert run_command(["compile", "-b", "arduino:samd:mkrvidor4000", sketch_path])


def test_compile_with_build_properties_flag(run_command, data_dir, copy_sketch):
    # Init the environment explicitly
    assert run_command(["core", "update-index"])

    # Install Arduino AVR Boards
    assert run_command(["core", "install", "arduino:avr@1.8.3"])

    sketch_path = copy_sketch("sketch_with_single_string_define")
    fqbn = "arduino:avr:uno"

    # Compile using a build property with quotes
    res = run_command(
        [
            "compile",
            "-b",
            fqbn,
            '--build-properties="build.extra_flags=\\"-DMY_DEFINE=\\"hello world\\"\\""',
            sketch_path,
            "--verbose",
            "--clean",
        ]
    )
    assert res.failed
    assert "Flag --build-properties has been deprecated, please use --build-property instead." not in res.stderr

    # Try again with quotes
    res = run_command(
        [
            "compile",
            "-b",
            fqbn,
            '--build-properties="build.extra_flags=-DMY_DEFINE=\\"hello\\""',
            sketch_path,
            "--verbose",
            "--clean",
        ]
    )
    assert res.failed
    assert "Flag --build-properties has been deprecated, please use --build-property instead." not in res.stderr

    # Try without quotes
    sketch_path = copy_sketch("sketch_with_single_int_define")
    res = run_command(
        [
            "compile",
            "-b",
            fqbn,
            '--build-properties="build.extra_flags=-DMY_DEFINE=1"',
            sketch_path,
            "--verbose",
            "--clean",
        ]
    )
    assert res.ok
    assert "Flag --build-properties has been deprecated, please use --build-property instead." in res.stderr
    assert "-DMY_DEFINE=1" in res.stdout

    sketch_path = copy_sketch("sketch_with_multiple_int_defines")
    res = run_command(
        [
            "compile",
            "-b",
            fqbn,
            '--build-properties="build.extra_flags=-DFIRST_PIN=1,compiler.cpp.extra_flags=-DSECOND_PIN=2"',
            sketch_path,
            "--verbose",
            "--clean",
        ]
    )
    assert res.ok
    assert "Flag --build-properties has been deprecated, please use --build-property instead." in res.stderr
    assert "-DFIRST_PIN=1" in res.stdout
    assert "-DSECOND_PIN=2" in res.stdout


def test_compile_with_build_property_containing_quotes(run_command, data_dir, copy_sketch):
    # Init the environment explicitly
    assert run_command(["core", "update-index"])

    # Install Arduino AVR Boards
    assert run_command(["core", "install", "arduino:avr@1.8.3"])

    sketch_path = copy_sketch("sketch_with_single_string_define")
    fqbn = "arduino:avr:uno"

    # Compile using a build property with quotes
    res = run_command(
        [
            "compile",
            "-b",
            fqbn,
            '--build-property=build.extra_flags=\\"-DMY_DEFINE=\\"hello world\\"\\"',
            sketch_path,
            "--verbose",
        ]
    )
    assert res.ok
    assert '-DMY_DEFINE=\\"hello world\\"' in res.stdout


def test_compile_with_multiple_build_property_flags(run_command, data_dir, copy_sketch, working_dir):
    # Init the environment explicitly
    assert run_command(["core", "update-index"])

    # Install Arduino AVR Boards
    assert run_command(["core", "install", "arduino:avr@1.8.3"])

    sketch_path = copy_sketch("sketch_with_multiple_defines")
    fqbn = "arduino:avr:uno"

    # Compile using multiple build properties separated by a space
    res = run_command(
        [
            "compile",
            "-b",
            fqbn,
            '--build-property=compiler.cpp.extra_flags=\\"-DPIN=2 -DSSID=\\"This is a String\\"\\"',
            sketch_path,
            "--verbose",
            "--clean",
        ]
    )
    assert res.failed

    # Compile using multiple build properties separated by a space and properly quoted
    res = run_command(
        [
            "compile",
            "-b",
            fqbn,
            '--build-property=compiler.cpp.extra_flags=-DPIN=2 \\"-DSSID=\\"This is a String\\"\\"',
            sketch_path,
            "--verbose",
            "--clean",
        ]
    )
    assert res.ok
    assert '-DPIN=2 "-DSSID=\\"This is a String\\""' in res.stdout

    # Tries compilation using multiple build properties separated by a comma
    res = run_command(
        [
            "compile",
            "-b",
            fqbn,
            '--build-property=compiler.cpp.extra_flags=\\"-DPIN=2,-DSSID=\\"This is a String\\"\\"',
            sketch_path,
            "--verbose",
            "--clean",
        ]
    )
    assert res.failed

    res = run_command(
        [
            "compile",
            "-b",
            fqbn,
            '--build-property=compiler.cpp.extra_flags=\\"-DPIN=2\\"',
            '--build-property=compiler.cpp.extra_flags=\\"-DSSID=\\"This is a String\\"\\"',
            sketch_path,
            "--verbose",
            "--clean",
        ]
    )
    assert res.failed
    assert "-DPIN=2" not in res.stdout
    assert '-DSSID=\\"This is a String\\"' in res.stdout

    res = run_command(
        [
            "compile",
            "-b",
            fqbn,
            '--build-property=compiler.cpp.extra_flags=\\"-DPIN=2\\"',
            '--build-property=build.extra_flags=\\"-DSSID=\\"hello world\\"\\"',
            sketch_path,
            "--verbose",
            "--clean",
        ]
    )
    assert res.ok
    assert "-DPIN=2" in res.stdout
    assert '-DSSID=\\"hello world\\"' in res.stdout
