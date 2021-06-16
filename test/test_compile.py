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
import shutil
from git import Repo
from pathlib import Path
import simplejson as json

import pytest

from .common import running_on_ci


def test_compile_without_fqbn(run_command):
    # Init the environment explicitly
    run_command("core update-index")

    # Install Arduino AVR Boards
    run_command("core install arduino:avr@1.8.3")

    # Build sketch without FQBN
    result = run_command("compile")
    assert result.failed


def test_compile_with_simple_sketch(run_command, data_dir, working_dir):
    # Init the environment explicitly
    run_command("core update-index")

    # Download latest AVR
    run_command("core install arduino:avr")

    sketch_name = "CompileIntegrationTest"
    sketch_path = Path(data_dir, sketch_name)
    fqbn = "arduino:avr:uno"

    # Create a test sketch
    result = run_command(f"sketch new {sketch_path}")
    assert result.ok
    assert f"Sketch created in: {sketch_path}" in result.stdout

    # Build sketch for arduino:avr:uno
    result = run_command(f"compile -b {fqbn} {sketch_path}")
    assert result.ok

    # Build sketch for arduino:avr:uno with json output
    result = run_command(f"compile -b {fqbn} {sketch_path} --format json")
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
    run_command("core update-index")

    # Install Arduino AVR Boards
    run_command("core install arduino:avr@1.8.3")

    # Create a test sketch
    sketch_path = os.path.join(data_dir, "test_output_flag_default_path")
    fqbn = "arduino:avr:uno"
    result = run_command("sketch new {}".format(sketch_path))
    assert result.ok

    # Test the --output-dir flag defaulting to current working dir
    result = run_command("compile -b {fqbn} {sketch_path} --output-dir test".format(fqbn=fqbn, sketch_path=sketch_path))
    assert result.ok
    target = os.path.join(working_dir, "test")
    assert os.path.exists(target) and os.path.isdir(target)


def test_compile_with_sketch_with_symlink_selfloop(run_command, data_dir):
    # Init the environment explicitly
    run_command("core update-index")

    # Install Arduino AVR Boards
    run_command("core install arduino:avr@1.8.3")

    sketch_name = "CompileIntegrationTestSymlinkSelfLoop"
    sketch_path = os.path.join(data_dir, sketch_name)
    fqbn = "arduino:avr:uno"

    # Create a test sketch
    result = run_command("sketch new {}".format(sketch_path))
    assert result.ok
    assert "Sketch created in: {}".format(sketch_path) in result.stdout

    # create a symlink that loops on himself
    loop_file_path = os.path.join(sketch_path, "loop")
    os.symlink(loop_file_path, loop_file_path)

    # Build sketch for arduino:avr:uno
    result = run_command("compile -b {fqbn} {sketch_path}".format(fqbn=fqbn, sketch_path=sketch_path))
    # The assertion is a bit relaxed in this case because win behaves differently from macOs and linux
    # returning a different error detailed message
    assert "Error during sketch processing" in result.stderr
    assert not result.ok

    sketch_name = "CompileIntegrationTestSymlinkDirLoop"
    sketch_path = os.path.join(data_dir, sketch_name)
    fqbn = "arduino:avr:uno"

    # Create a test sketch
    result = run_command("sketch new {}".format(sketch_path))
    assert result.ok
    assert "Sketch created in: {}".format(sketch_path) in result.stdout

    # create a symlink that loops on the upper level
    loop_dir_path = os.path.join(sketch_path, "loop_dir")
    os.mkdir(loop_dir_path)
    loop_dir_symlink_path = os.path.join(loop_dir_path, "loop_dir_symlink")
    os.symlink(loop_dir_path, loop_dir_symlink_path)

    # Build sketch for arduino:avr:uno
    result = run_command("compile -b {fqbn} {sketch_path}".format(fqbn=fqbn, sketch_path=sketch_path))
    # The assertion is a bit relaxed also in this case because macOS behaves differently from win and linux:
    # the cli does not follow recursively the symlink til breaking
    assert "Error during sketch processing" in result.stderr
    assert not result.ok


def test_compile_blacklisted_sketchname(run_command, data_dir):
    """
    Compile should ignore folders named `RCS`, `.git` and the likes, but
    it should be ok for a sketch to be named like RCS.ino
    """
    # Init the environment explicitly
    run_command("core update-index")

    # Install Arduino AVR Boards
    run_command("core install arduino:avr@1.8.3")

    sketch_name = "RCS"
    sketch_path = os.path.join(data_dir, sketch_name)
    fqbn = "arduino:avr:uno"

    # Create a test sketch
    result = run_command("sketch new {}".format(sketch_path))
    assert result.ok
    assert "Sketch created in: {}".format(sketch_path) in result.stdout

    # Build sketch for arduino:avr:uno
    result = run_command("compile -b {fqbn} {sketch_path}".format(fqbn=fqbn, sketch_path=sketch_path))
    assert result.ok


def test_compile_without_precompiled_libraries(run_command, data_dir):
    # Init the environment explicitly
    url = "https://adafruit.github.io/arduino-board-index/package_adafruit_index.json"
    assert run_command(f"core update-index --additional-urls={url}")
    assert run_command(f"core install arduino:mbed@1.3.1 --additional-urls={url}")

    # Precompiled version of Arduino_TensorflowLite
    assert run_command("lib install Arduino_LSM9DS1")
    assert run_command("lib install Arduino_TensorflowLite@2.1.1-ALPHA-precompiled")

    sketch_path = Path(data_dir, "libraries", "Arduino_TensorFlowLite", "examples", "hello_world")
    assert run_command(f"compile -b arduino:mbed:nano33ble {sketch_path}")

    assert run_command(f"core install arduino:samd@1.8.7 --additional-urls={url}")
    assert run_command(f"core install adafruit:samd@1.6.4 --additional-urls={url}")
    # should work on adafruit too after https://github.com/arduino/arduino-cli/pull/1134
    assert run_command(f"compile -b adafruit:samd:adafruit_feather_m4 {sketch_path}")

    # Non-precompiled version of Arduino_TensorflowLite
    assert run_command("lib install Arduino_TensorflowLite@2.1.0-ALPHA")
    assert run_command(f"compile -b arduino:mbed:nano33ble {sketch_path}")
    assert run_command(f"compile -b adafruit:samd:adafruit_feather_m4 {sketch_path}")

    # Bosch sensor library
    assert run_command('lib install "BSEC Software Library@1.5.1474"')
    sketch_path = Path(data_dir, "libraries", "BSEC_Software_Library", "examples", "basic")
    assert run_command(f"compile -b arduino:samd:mkr1000 {sketch_path}")
    assert run_command(f"compile -b arduino:mbed:nano33ble {sketch_path}")

    # USBBlaster library
    assert run_command('lib install "USBBlaster@1.0.0"')
    sketch_path = Path(data_dir, "libraries", "USBBlaster", "examples", "USB_Blaster")
    assert run_command(f"compile -b arduino:samd:mkrvidor4000 {sketch_path}")


def test_compile_with_build_properties_flag(run_command, data_dir, copy_sketch):
    # Init the environment explicitly
    assert run_command("core update-index")

    # Install Arduino AVR Boards
    assert run_command("core install arduino:avr@1.8.3")

    sketch_path = copy_sketch("sketch_with_single_string_define")
    fqbn = "arduino:avr:uno"

    # Compile using a build property with quotes
    res = run_command(
        f"compile -b {fqbn} "
        + '--build-properties="build.extra_flags=\\"-DMY_DEFINE=\\"hello world\\"\\"" '
        + f"{sketch_path} --verbose --clean"
    )
    assert res.failed
    assert "Flag --build-properties has been deprecated, please use --build-property instead." not in res.stderr

    # Try again with quotes
    res = run_command(
        f"compile -b {fqbn} "
        + '--build-properties="build.extra_flags=-DMY_DEFINE=\\"hello\\"" '
        + f"{sketch_path} --verbose --clean"
    )
    assert res.failed
    assert "Flag --build-properties has been deprecated, please use --build-property instead." not in res.stderr

    # Try without quotes
    sketch_path = copy_sketch("sketch_with_single_int_define")
    res = run_command(
        f"compile -b {fqbn} "
        + '--build-properties="build.extra_flags=-DMY_DEFINE=1" '
        + f"{sketch_path} --verbose --clean"
    )
    assert res.ok
    assert "Flag --build-properties has been deprecated, please use --build-property instead." in res.stderr
    assert "-DMY_DEFINE=1" in res.stdout

    sketch_path = copy_sketch("sketch_with_multiple_int_defines")
    res = run_command(
        f"compile -b {fqbn} "
        + '--build-properties="build.extra_flags=-DFIRST_PIN=1,compiler.cpp.extra_flags=-DSECOND_PIN=2" '
        + f"{sketch_path} --verbose --clean"
    )
    assert res.ok
    assert "Flag --build-properties has been deprecated, please use --build-property instead." in res.stderr
    assert "-DFIRST_PIN=1" in res.stdout
    assert "-DSECOND_PIN=2" in res.stdout


def test_compile_with_build_property_containing_quotes(run_command, data_dir, copy_sketch):
    # Init the environment explicitly
    assert run_command("core update-index")

    # Install Arduino AVR Boards
    assert run_command("core install arduino:avr@1.8.3")

    sketch_path = copy_sketch("sketch_with_single_string_define")
    fqbn = "arduino:avr:uno"

    # Compile using a build property with quotes
    res = run_command(
        f"compile -b {fqbn} "
        + '--build-property="build.extra_flags=\\"-DMY_DEFINE=\\"hello world\\"\\"" '
        + f"{sketch_path} --verbose"
    )
    assert res.ok
    assert '-DMY_DEFINE=\\"hello world\\"' in res.stdout


def test_compile_with_multiple_build_property_flags(run_command, data_dir, copy_sketch, working_dir):
    # Init the environment explicitly
    assert run_command("core update-index")

    # Install Arduino AVR Boards
    assert run_command("core install arduino:avr@1.8.3")

    sketch_path = copy_sketch("sketch_with_multiple_defines")
    fqbn = "arduino:avr:uno"

    # Compile using multiple build properties separated by a space
    res = run_command(
        f"compile -b {fqbn} "
        + '--build-property="compiler.cpp.extra_flags=\\"-DPIN=2 -DSSID=\\"This is a String\\"\\"" '
        + f"{sketch_path} --verbose --clean"
    )
    assert res.failed

    # Compile using multiple build properties separated by a space and properly quoted
    res = run_command(
        f"compile -b {fqbn} "
        + '--build-property="compiler.cpp.extra_flags=-DPIN=2 \\"-DSSID=\\"This is a String\\"\\"" '
        + f"{sketch_path} --verbose --clean"
    )
    assert res.ok
    assert '-DPIN=2 "-DSSID=\\"This is a String\\""' in res.stdout

    # Tries compilation using multiple build properties separated by a comma
    res = run_command(
        f"compile -b {fqbn} "
        + '--build-property="compiler.cpp.extra_flags=\\"-DPIN=2,-DSSID=\\"This is a String\\"\\"\\" '
        + f"{sketch_path} --verbose --clean"
    )
    assert res.failed

    res = run_command(
        f"compile -b {fqbn} "
        + '--build-property="compiler.cpp.extra_flags=\\"-DPIN=2\\"" '
        + '--build-property="compiler.cpp.extra_flags=\\"-DSSID=\\"This is a String\\"\\"" '
        + f"{sketch_path} --verbose --clean"
    )
    assert res.failed
    assert "-DPIN=2" not in res.stdout
    assert '-DSSID=\\"This is a String\\"' in res.stdout

    res = run_command(
        f"compile -b {fqbn} "
        + '--build-property="compiler.cpp.extra_flags=\\"-DPIN=2\\"" '
        + '--build-property="build.extra_flags=\\"-DSSID=\\"hello world\\"\\"" '
        + f"{sketch_path} --verbose --clean"
    )
    assert res.ok
    assert "-DPIN=2" in res.stdout
    assert '-DSSID=\\"hello world\\"' in res.stdout


def test_compile_with_output_dir_flag(run_command, data_dir):
    # Init the environment explicitly
    run_command("core update-index")

    # Download latest AVR
    run_command("core install arduino:avr")

    sketch_name = "CompileWithOutputDir"
    sketch_path = Path(data_dir, sketch_name)
    fqbn = "arduino:avr:uno"

    # Create a test sketch
    result = run_command(f"sketch new {sketch_path}")
    assert result.ok
    assert f"Sketch created in: {sketch_path}" in result.stdout

    # Test the --output-dir flag with absolute path
    output_dir = Path(data_dir, "test_dir", "output_dir")
    result = run_command(f"compile -b {fqbn} {sketch_path} --output-dir {output_dir}")
    assert result.ok

    # Verifies expected binaries have been built
    sketch_path_md5 = hashlib.md5(bytes(sketch_path)).hexdigest().upper()
    build_dir = Path(tempfile.gettempdir(), f"arduino-sketch-{sketch_path_md5}")
    assert (build_dir / f"{sketch_name}.ino.eep").exists()
    assert (build_dir / f"{sketch_name}.ino.elf").exists()
    assert (build_dir / f"{sketch_name}.ino.hex").exists()
    assert (build_dir / f"{sketch_name}.ino.with_bootloader.bin").exists()
    assert (build_dir / f"{sketch_name}.ino.with_bootloader.hex").exists()

    # Verifies binaries are exported when --output-dir flag is specified
    assert output_dir.exists()
    assert output_dir.is_dir()
    assert (output_dir / f"{sketch_name}.ino.eep").exists()
    assert (output_dir / f"{sketch_name}.ino.elf").exists()
    assert (output_dir / f"{sketch_name}.ino.hex").exists()
    assert (output_dir / f"{sketch_name}.ino.with_bootloader.bin").exists()
    assert (output_dir / f"{sketch_name}.ino.with_bootloader.hex").exists()


def test_compile_with_export_binaries_flag(run_command, data_dir):
    # Init the environment explicitly
    run_command("core update-index")

    # Download latest AVR
    run_command("core install arduino:avr")

    sketch_name = "CompileWithExportBinariesFlag"
    sketch_path = Path(data_dir, sketch_name)
    fqbn = "arduino:avr:uno"

    # Create a test sketch
    assert run_command("sketch new {}".format(sketch_path))

    # Test the --output-dir flag with absolute path
    result = run_command(f"compile -b {fqbn} {sketch_path} --export-binaries")
    assert result.ok
    assert Path(sketch_path, "build").exists()
    assert Path(sketch_path, "build").is_dir()

    # Verifies binaries are exported when --export-binaries flag is set
    assert (sketch_path / "build" / fqbn.replace(":", ".") / f"{sketch_name}.ino.eep").exists()
    assert (sketch_path / "build" / fqbn.replace(":", ".") / f"{sketch_name}.ino.elf").exists()
    assert (sketch_path / "build" / fqbn.replace(":", ".") / f"{sketch_name}.ino.hex").exists()
    assert (sketch_path / "build" / fqbn.replace(":", ".") / f"{sketch_name}.ino.with_bootloader.bin").exists()
    assert (sketch_path / "build" / fqbn.replace(":", ".") / f"{sketch_name}.ino.with_bootloader.hex").exists()


def test_compile_with_custom_build_path(run_command, data_dir):
    # Init the environment explicitly
    run_command("core update-index")

    # Download latest AVR
    run_command("core install arduino:avr")

    sketch_name = "CompileWithBuildPath"
    sketch_path = Path(data_dir, sketch_name)
    fqbn = "arduino:avr:uno"

    # Create a test sketch
    result = run_command(f"sketch new {sketch_path}")
    assert result.ok
    assert f"Sketch created in: {sketch_path}" in result.stdout

    # Test the --build-path flag with absolute path
    build_path = Path(data_dir, "test_dir", "build_dir")
    result = run_command(f"compile -b {fqbn} {sketch_path} --build-path {build_path}")
    print(result.stderr)
    assert result.ok

    # Verifies expected binaries have been built to build_path
    assert build_path.exists()
    assert build_path.is_dir()
    assert (build_path / f"{sketch_name}.ino.eep").exists()
    assert (build_path / f"{sketch_name}.ino.elf").exists()
    assert (build_path / f"{sketch_name}.ino.hex").exists()
    assert (build_path / f"{sketch_name}.ino.with_bootloader.bin").exists()
    assert (build_path / f"{sketch_name}.ino.with_bootloader.hex").exists()

    # Verifies there are no binaries in temp directory
    sketch_path_md5 = hashlib.md5(bytes(sketch_path)).hexdigest().upper()
    build_dir = Path(tempfile.gettempdir(), f"arduino-sketch-{sketch_path_md5}")
    assert not (build_dir / f"{sketch_name}.ino.eep").exists()
    assert not (build_dir / f"{sketch_name}.ino.elf").exists()
    assert not (build_dir / f"{sketch_name}.ino.hex").exists()
    assert not (build_dir / f"{sketch_name}.ino.with_bootloader.bin").exists()
    assert not (build_dir / f"{sketch_name}.ino.with_bootloader.hex").exists()


def test_compile_with_export_binaries_env_var(run_command, data_dir, downloads_dir):
    # Init the environment explicitly
    run_command("core update-index")

    # Download latest AVR
    run_command("core install arduino:avr")

    sketch_name = "CompileWithExportBinariesEnvVar"
    sketch_path = Path(data_dir, sketch_name)
    fqbn = "arduino:avr:uno"

    # Create a test sketch
    assert run_command("sketch new {}".format(sketch_path))

    env = {
        "ARDUINO_DATA_DIR": data_dir,
        "ARDUINO_DOWNLOADS_DIR": downloads_dir,
        "ARDUINO_SKETCHBOOK_DIR": data_dir,
        "ARDUINO_SKETCH_ALWAYS_EXPORT_BINARIES": "true",
    }
    # Test compilation with export binaries env var set
    result = run_command(f"compile -b {fqbn} {sketch_path}", custom_env=env)
    assert result.ok
    assert Path(sketch_path, "build").exists()
    assert Path(sketch_path, "build").is_dir()

    # Verifies binaries are exported when export binaries env var is set
    assert (sketch_path / "build" / fqbn.replace(":", ".") / f"{sketch_name}.ino.eep").exists()
    assert (sketch_path / "build" / fqbn.replace(":", ".") / f"{sketch_name}.ino.elf").exists()
    assert (sketch_path / "build" / fqbn.replace(":", ".") / f"{sketch_name}.ino.hex").exists()
    assert (sketch_path / "build" / fqbn.replace(":", ".") / f"{sketch_name}.ino.with_bootloader.bin").exists()
    assert (sketch_path / "build" / fqbn.replace(":", ".") / f"{sketch_name}.ino.with_bootloader.hex").exists()


def test_compile_with_export_binaries_config(run_command, data_dir, downloads_dir):
    # Init the environment explicitly
    run_command("core update-index")

    # Download latest AVR
    run_command("core install arduino:avr")

    sketch_name = "CompileWithExportBinariesConfig"
    sketch_path = Path(data_dir, sketch_name)
    fqbn = "arduino:avr:uno"

    # Create a test sketch
    assert run_command("sketch new {}".format(sketch_path))

    # Create settings with export binaries set to true
    env = {
        "ARDUINO_DATA_DIR": data_dir,
        "ARDUINO_DOWNLOADS_DIR": downloads_dir,
        "ARDUINO_SKETCHBOOK_DIR": data_dir,
        "ARDUINO_SKETCH_ALWAYS_EXPORT_BINARIES": "true",
    }
    assert run_command("config init --dest-dir .", custom_env=env)

    # Test compilation with export binaries env var set
    result = run_command(f"compile -b {fqbn} {sketch_path}")
    assert result.ok
    assert Path(sketch_path, "build").exists()
    assert Path(sketch_path, "build").is_dir()

    # Verifies binaries are exported when export binaries env var is set
    assert (sketch_path / "build" / fqbn.replace(":", ".") / f"{sketch_name}.ino.eep").exists()
    assert (sketch_path / "build" / fqbn.replace(":", ".") / f"{sketch_name}.ino.elf").exists()
    assert (sketch_path / "build" / fqbn.replace(":", ".") / f"{sketch_name}.ino.hex").exists()
    assert (sketch_path / "build" / fqbn.replace(":", ".") / f"{sketch_name}.ino.with_bootloader.bin").exists()
    assert (sketch_path / "build" / fqbn.replace(":", ".") / f"{sketch_name}.ino.with_bootloader.hex").exists()


def test_compile_with_invalid_url(run_command, data_dir):
    # Init the environment explicitly
    run_command("core update-index")

    # Download latest AVR
    run_command("core install arduino:avr")

    sketch_name = "CompileWithInvalidURL"
    sketch_path = Path(data_dir, sketch_name)
    fqbn = "arduino:avr:uno"

    # Create a test sketch
    assert run_command(f'sketch new "{sketch_path}"')

    # Create settings with custom invalid URL
    assert run_command("config init --dest-dir . --additional-urls https://example.com/package_example_index.json")

    # Verifies compilation fails cause of missing local index file
    res = run_command(f'compile -b {fqbn} "{sketch_path}"')
    assert res.ok
    lines = [l.strip() for l in res.stderr.splitlines()]
    assert "Error initializing instance: Loading index file: loading json index file" in lines[0]
    expected_index_file = Path(data_dir, "package_example_index.json")
    assert f"loading json index file {expected_index_file}: " + f"open {expected_index_file}:" in lines[-1]


def test_compile_with_custom_libraries(run_command, copy_sketch):
    # Creates config with additional URL to install necessary core
    url = "http://arduino.esp8266.com/stable/package_esp8266com_index.json"
    assert run_command(f"config init --dest-dir . --additional-urls {url}")

    # Init the environment explicitly
    assert run_command("update")

    # Install core to compile
    assert run_command("core install esp8266:esp8266")

    sketch_path = copy_sketch("sketch_with_multiple_custom_libraries")
    fqbn = "esp8266:esp8266:nodemcu:xtal=80,vt=heap,eesz=4M1M,wipe=none,baud=115200"

    first_lib = Path(sketch_path, "libraries1")
    second_lib = Path(sketch_path, "libraries2")
    # This compile command has been taken from this issue:
    # https://github.com/arduino/arduino-cli/issues/973
    assert run_command(f"compile --libraries {first_lib},{second_lib} -b {fqbn} {sketch_path}")


def test_compile_with_archives_and_long_paths(run_command):
    # Creates config with additional URL to install necessary core
    url = "http://arduino.esp8266.com/stable/package_esp8266com_index.json"
    assert run_command(f"config init --dest-dir . --additional-urls {url}")

    # Init the environment explicitly
    assert run_command("update")

    # Install core to compile
    assert run_command("core install esp8266:esp8266@2.7.4")

    # Install test library
    assert run_command("lib install ArduinoIoTCloud")

    result = run_command("lib examples ArduinoIoTCloud --format json")
    assert result.ok
    lib_output = json.loads(result.stdout)
    sketch_path = Path(lib_output[0]["library"]["install_dir"], "examples", "ArduinoIoTCloud-Advanced")

    assert run_command(f"compile -b esp8266:esp8266:huzzah {sketch_path}")


def test_compile_with_precompiled_library(run_command, data_dir):
    assert run_command("update")

    assert run_command("core install arduino:samd@1.8.11")
    fqbn = "arduino:samd:mkrzero"

    # Install precompiled library
    # For more information see:
    # https://arduino.github.io/arduino-cli/latest/library-specification/#precompiled-binaries
    assert run_command('lib install "BSEC Software Library@1.5.1474"')
    sketch_folder = Path(data_dir, "libraries", "BSEC_Software_Library", "examples", "basic")

    # Compile and verify dependencies detection for fully precompiled library is not skipped
    result = run_command(f"compile -b {fqbn} {sketch_folder} -v")
    assert result.ok
    assert "Skipping dependencies detection for precompiled library BSEC Software Library" not in result.stdout


def test_compile_with_fully_precompiled_library(run_command, data_dir):
    assert run_command("update")

    assert run_command("core install arduino:mbed@1.3.1")
    fqbn = "arduino:mbed:nano33ble"

    # Install fully precompiled library
    # For more information see:
    # https://arduino.github.io/arduino-cli/latest/library-specification/#precompiled-binaries
    assert run_command("lib install Arduino_TensorFlowLite@2.1.1-ALPHA-precompiled")
    sketch_folder = Path(data_dir, "libraries", "Arduino_TensorFlowLite", "examples", "hello_world")

    # Install example dependency
    # assert run_command("lib install Arduino_LSM9DS1")

    # Compile and verify dependencies detection for fully precompiled library is skipped
    result = run_command(f"compile -b {fqbn} {sketch_folder} -v")
    assert result.ok
    assert "Skipping dependencies detection for precompiled library Arduino_TensorFlowLite" in result.stdout


def test_compile_sketch_with_pde_extension(run_command, data_dir):
    # Init the environment explicitly
    assert run_command("update")

    # Install core to compile
    assert run_command("core install arduino:avr@1.8.3")

    sketch_name = "CompilePdeSketch"
    sketch_path = Path(data_dir, sketch_name)
    fqbn = "arduino:avr:uno"

    # Create a test sketch
    assert run_command(f"sketch new {sketch_path}")

    # Renames sketch file to pde
    sketch_file = Path(sketch_path, f"{sketch_name}.ino").rename(sketch_path / f"{sketch_name}.pde")

    # Build sketch from folder
    res = run_command(f"compile --clean -b {fqbn} {sketch_path}")
    assert res.ok
    assert "Sketches with .pde extension are deprecated, please rename the following files to .ino:" in res.stderr
    assert str(sketch_file) in res.stderr

    # Build sketch from file
    res = run_command(f"compile --clean -b {fqbn} {sketch_file}")
    assert res.ok
    assert "Sketches with .pde extension are deprecated, please rename the following files to .ino" in res.stderr
    assert str(sketch_file) in res.stderr


def test_compile_sketch_with_multiple_main_files(run_command, data_dir):
    # Init the environment explicitly
    assert run_command("update")

    # Install core to compile
    assert run_command("core install arduino:avr@1.8.3")

    sketch_name = "CompileSketchMultipleMainFiles"
    sketch_path = Path(data_dir, sketch_name)
    fqbn = "arduino:avr:uno"

    # Create a test sketch
    assert run_command(f"sketch new {sketch_path}")

    # Copy .ino sketch file to .pde
    sketch_ino_file = Path(sketch_path, f"{sketch_name}.ino")
    sketch_pde_file = Path(sketch_path / f"{sketch_name}.pde")
    shutil.copyfile(sketch_ino_file, sketch_pde_file)

    # Build sketch from folder
    res = run_command(f"compile --clean -b {fqbn} {sketch_path}")
    assert res.failed
    assert "Error during build: opening sketch: multiple main sketch files found" in res.stderr

    # Build sketch from .ino file
    res = run_command(f"compile --clean -b {fqbn} {sketch_ino_file}")
    assert res.failed
    assert "Error during build: opening sketch: multiple main sketch files found" in res.stderr

    # Build sketch from .pde file
    res = run_command(f"compile --clean -b {fqbn} {sketch_pde_file}")
    assert res.failed
    assert "Error during build: opening sketch: multiple main sketch files found" in res.stderr


def test_compile_sketch_case_mismatch_fails(run_command, data_dir):
    # Init the environment explicitly
    assert run_command("update")

    # Install core to compile
    assert run_command("core install arduino:avr@1.8.3")

    sketch_name = "CompileSketchCaseMismatch"
    sketch_path = Path(data_dir, sketch_name)
    fqbn = "arduino:avr:uno"

    assert run_command(f"sketch new {sketch_path}")

    # Rename main .ino file so casing is different from sketch name
    sketch_main_file = Path(sketch_path, f"{sketch_name}.ino").rename(sketch_path / f"{sketch_name.lower()}.ino")

    # Verifies compilation fails when:
    # * Compiling with sketch path
    res = run_command(f"compile --clean -b {fqbn} {sketch_path}")
    assert res.failed
    assert "Error during build: opening sketch: no valid sketch found" in res.stderr
    # * Compiling with sketch main file
    res = run_command(f"compile --clean -b {fqbn} {sketch_main_file}")
    assert res.failed
    assert "Error during build: opening sketch: no valid sketch found" in res.stderr
    # * Compiling in sketch path
    res = run_command(f"compile --clean -b {fqbn}", custom_working_dir=sketch_path)
    assert res.failed
    assert "Error during build: opening sketch: no valid sketch found" in res.stderr


def test_compile_with_only_compilation_database_flag(run_command, data_dir):
    assert run_command("update")

    assert run_command("core install arduino:avr@1.8.3")

    sketch_name = "CompileSketchOnlyCompilationDatabaseFlag"
    sketch_path = Path(data_dir, sketch_name)
    fqbn = "arduino:avr:uno"

    assert run_command(f"sketch new {sketch_path}")

    # Verifies no binaries exist
    build_path = Path(sketch_path, "build")
    assert not build_path.exists()

    # Compile with both --export-binaries and --only-compilation-database flags
    assert run_command(f"compile --export-binaries --only-compilation-database --clean -b {fqbn} {sketch_path}")

    # Verifies no binaries are exported
    assert not build_path.exists()

    # Verifies no binaries exist
    build_path = Path(data_dir, "export-dir")
    assert not build_path.exists()

    # Compile by setting the --output-dir flag and --only-compilation-database flags
    assert run_command(f"compile --output-dir {build_path} --only-compilation-database --clean -b {fqbn} {sketch_path}")

    # Verifies no binaries are exported
    assert not build_path.exists()


def test_compile_using_platform_local_txt(run_command, data_dir):
    assert run_command("update")

    assert run_command("core install arduino:avr@1.8.3")

    sketch_name = "CompileSketchUsingPlatformLocalTxt"
    sketch_path = Path(data_dir, sketch_name)
    fqbn = "arduino:avr:uno"

    assert run_command(f"sketch new {sketch_path}")

    # Verifies compilation works without issues
    assert run_command(f"compile --clean -b {fqbn} {sketch_path}")

    # Overrides default platform compiler with an unexisting one
    platform_local_txt = Path(data_dir, "packages", "arduino", "hardware", "avr", "1.8.3", "platform.local.txt")
    platform_local_txt.write_text("compiler.c.cmd=my-compiler-that-does-not-exist")

    # Verifies compilation now fails because compiler is not found
    res = run_command(f"compile --clean -b {fqbn} {sketch_path}")
    assert res.failed
    assert "my-compiler-that-does-not-exist" in res.stderr


def test_compile_using_boards_local_txt(run_command, data_dir):
    assert run_command("update")

    assert run_command("core install arduino:avr@1.8.3")

    sketch_name = "CompileSketchUsingBoardsLocalTxt"
    sketch_path = Path(data_dir, sketch_name)
    # Use a made up board
    fqbn = "arduino:avr:nessuno"

    assert run_command(f"sketch new {sketch_path}")

    # Verifies compilation fails because board doesn't exist
    res = run_command(f"compile --clean -b {fqbn} {sketch_path}")
    assert res.failed
    assert "Error during build: Error resolving FQBN: board arduino:avr@1.8.3:nessuno not found" in res.stderr

    # Use custom boards.local.txt with made arduino:avr:nessuno board
    boards_local_txt = Path(data_dir, "packages", "arduino", "hardware", "avr", "1.8.3", "boards.local.txt")
    shutil.copyfile(Path(__file__).parent / "testdata" / "boards.local.txt", boards_local_txt)

    assert run_command(f"compile --clean -b {fqbn} {sketch_path}")


def test_compile_manually_installed_platform(run_command, data_dir):
    assert run_command("update")

    sketch_name = "CompileSketchManuallyInstalledPlatformUsingPlatformLocalTxt"
    sketch_path = Path(data_dir, sketch_name)
    fqbn = "arduino-beta-development:avr:uno"
    assert run_command(f"sketch new {sketch_path}")

    # Manually installs a core in sketchbooks hardware folder
    git_url = "https://github.com/arduino/ArduinoCore-avr.git"
    repo_dir = Path(data_dir, "hardware", "arduino-beta-development", "avr")
    assert Repo.clone_from(git_url, repo_dir, multi_options=["-b 1.8.3"])

    # Installs also the same core via CLI so all the necessary tools are installed
    assert run_command("core install arduino:avr@1.8.3")

    # Verifies compilation works without issues
    assert run_command(f"compile --clean -b {fqbn} {sketch_path}")


def test_compile_manually_installed_platform_using_platform_local_txt(run_command, data_dir):
    assert run_command("update")

    sketch_name = "CompileSketchManuallyInstalledPlatformUsingPlatformLocalTxt"
    sketch_path = Path(data_dir, sketch_name)
    fqbn = "arduino-beta-development:avr:uno"
    assert run_command(f"sketch new {sketch_path}")

    # Manually installs a core in sketchbooks hardware folder
    git_url = "https://github.com/arduino/ArduinoCore-avr.git"
    repo_dir = Path(data_dir, "hardware", "arduino-beta-development", "avr")
    assert Repo.clone_from(git_url, repo_dir, multi_options=["-b 1.8.3"])

    # Installs also the same core via CLI so all the necessary tools are installed
    assert run_command("core install arduino:avr@1.8.3")

    # Verifies compilation works without issues
    assert run_command(f"compile --clean -b {fqbn} {sketch_path}")

    # Overrides default platform compiler with an unexisting one
    platform_local_txt = Path(repo_dir, "platform.local.txt")
    platform_local_txt.write_text("compiler.c.cmd=my-compiler-that-does-not-exist")

    # Verifies compilation now fails because compiler is not found
    res = run_command(f"compile --clean -b {fqbn} {sketch_path}")
    assert res.failed
    assert "my-compiler-that-does-not-exist" in res.stderr


def test_compile_manually_installed_platform_using_boards_local_txt(run_command, data_dir):
    assert run_command("update")

    sketch_name = "CompileSketchManuallyInstalledPlatformUsingBoardsLocalTxt"
    sketch_path = Path(data_dir, sketch_name)
    fqbn = "arduino-beta-development:avr:nessuno"
    assert run_command(f"sketch new {sketch_path}")

    # Manually installs a core in sketchbooks hardware folder
    git_url = "https://github.com/arduino/ArduinoCore-avr.git"
    repo_dir = Path(data_dir, "hardware", "arduino-beta-development", "avr")
    assert Repo.clone_from(git_url, repo_dir, multi_options=["-b 1.8.3"])

    # Installs also the same core via CLI so all the necessary tools are installed
    assert run_command("core install arduino:avr@1.8.3")

    # Verifies compilation fails because board doesn't exist
    res = run_command(f"compile --clean -b {fqbn} {sketch_path}")
    assert res.failed
    assert (
        "Error during build: Error resolving FQBN: board arduino-beta-development:avr@1.8.3:nessuno not found"
        in res.stderr
    )

    # Use custom boards.local.txt with made arduino:avr:nessuno board
    boards_local_txt = Path(repo_dir, "boards.local.txt")
    shutil.copyfile(Path(__file__).parent / "testdata" / "boards.local.txt", boards_local_txt)

    assert run_command(f"compile --clean -b {fqbn} {sketch_path}")


def test_compile_with_library(run_command, data_dir):
    assert run_command("update")

    assert run_command("core install arduino:avr@1.8.3")

    sketch_name = "CompileSketchWithWiFi101Dependency"
    sketch_path = Path(data_dir, sketch_name)
    fqbn = "arduino:avr:uno"
    # Create new sketch and add library include
    assert run_command(f"sketch new {sketch_path}")
    sketch_file = sketch_path / f"{sketch_name}.ino"
    lines = []
    with open(sketch_file, "r") as f:
        lines = f.readlines()
    lines = ["#include <WiFi101.h>\n"] + lines
    with open(sketch_file, "w") as f:
        f.writelines(lines)

    # Manually installs a library
    git_url = "https://github.com/arduino-libraries/WiFi101.git"
    lib_path = Path(data_dir, "my-libraries", "WiFi101")
    assert Repo.clone_from(git_url, lib_path, multi_options=["-b 0.16.1"])

    res = run_command(f"compile -b {fqbn} {sketch_path} --library {lib_path} -v")
    assert res.ok
    assert "WiFi101" in res.stdout


def test_compile_with_library_priority(run_command, data_dir):
    assert run_command("update")

    assert run_command("core install arduino:avr@1.8.3")

    sketch_name = "CompileSketchWithLibraryPriority"
    sketch_path = Path(data_dir, sketch_name)
    fqbn = "arduino:avr:uno"

    # Manually installs a library
    git_url = "https://github.com/arduino-libraries/WiFi101.git"
    manually_install_lib_path = Path(data_dir, "my-libraries", "WiFi101")
    assert Repo.clone_from(git_url, manually_install_lib_path, multi_options=["-b 0.16.1"])

    # Install the same library we installed manually
    assert run_command("lib install WiFi101")

    # Create new sketch and add library include
    assert run_command(f"sketch new {sketch_path}")
    sketch_file = sketch_path / f"{sketch_name}.ino"
    lines = []
    with open(sketch_file, "r") as f:
        lines = f.readlines()
    lines = ["#include <WiFi101.h>"] + lines
    with open(sketch_file, "w") as f:
        f.writelines(lines)

    res = run_command(f"compile -b {fqbn} {sketch_path} --library {manually_install_lib_path} -v")
    assert res.ok
    cli_installed_lib_path = Path(data_dir, "libraries", "WiFi101")
    expected_output = [
        'Multiple libraries were found for "WiFi101.h"',
        f" Used: {manually_install_lib_path}",
        f" Not used: {cli_installed_lib_path}",
    ]
    assert "\n".join(expected_output) in res.stdout


def test_recompile_with_different_library(run_command, data_dir):
    assert run_command("update")

    assert run_command("core install arduino:avr@1.8.3")

    sketch_name = "RecompileCompileSketchWithDifferentLibrary"
    sketch_path = Path(data_dir, sketch_name)
    fqbn = "arduino:avr:uno"

    # Install library
    assert run_command("lib install WiFi101")

    # Manually installs the same library already installed
    git_url = "https://github.com/arduino-libraries/WiFi101.git"
    manually_install_lib_path = Path(data_dir, "my-libraries", "WiFi101")
    assert Repo.clone_from(git_url, manually_install_lib_path, multi_options=["-b 0.16.1"])

    # Create new sketch and add library include
    assert run_command(f"sketch new {sketch_path}")
    sketch_file = sketch_path / f"{sketch_name}.ino"
    lines = []
    with open(sketch_file, "r") as f:
        lines = f.readlines()
    lines = ["#include <WiFi101.h>"] + lines
    with open(sketch_file, "w") as f:
        f.writelines(lines)

    sketch_path_md5 = hashlib.md5(bytes(sketch_path)).hexdigest().upper()
    build_dir = Path(tempfile.gettempdir(), f"arduino-sketch-{sketch_path_md5}")

    # Compile sketch using library not managed by CLI
    res = run_command(f"compile -b {fqbn} --library {manually_install_lib_path} {sketch_path} -v")
    assert res.ok
    obj_path = build_dir / "libraries" / "WiFi101" / "WiFi.cpp.o"
    assert f"Using previously compiled file: {obj_path}" not in res.stdout

    # Compile again using library installed from CLI
    res = run_command(f"compile -b {fqbn} {sketch_path} -v")
    assert res.ok
    obj_path = build_dir / "libraries" / "WiFi101" / "WiFi.cpp.o"
    assert f"Using previously compiled file: {obj_path}" not in res.stdout


def test_compile_with_conflicting_libraries_include(run_command, data_dir, copy_sketch):
    assert run_command("update")

    assert run_command("core install arduino:avr@1.8.3")

    # Install conflicting libraries
    git_url = "https://github.com/pstolarz/OneWireNg.git"
    one_wire_ng_lib_path = Path(data_dir, "libraries", "onewireng_0_8_1")
    assert Repo.clone_from(git_url, one_wire_ng_lib_path, multi_options=["-b 0.8.1"])

    git_url = "https://github.com/PaulStoffregen/OneWire.git"
    one_wire_lib_path = Path(data_dir, "libraries", "onewire_2_3_5")
    assert Repo.clone_from(git_url, one_wire_lib_path, multi_options=["-b v2.3.5"])

    sketch_path = copy_sketch("sketch_with_conflicting_libraries_include")
    fqbn = "arduino:avr:uno"

    res = run_command(f"compile -b {fqbn} {sketch_path} --verbose")
    assert res.ok
    expected_output = [
        'Multiple libraries were found for "OneWire.h"',
        f" Used: {one_wire_lib_path}",
        f" Not used: {one_wire_ng_lib_path}",
    ]
    assert "\n".join(expected_output) in res.stdout


def test_compile_with_invalid_build_options_json(run_command, data_dir):
    assert run_command("update")

    assert run_command("core install arduino:avr@1.8.3")

    sketch_name = "CompileInvalidBuildOptionsJson"
    sketch_path = Path(data_dir, sketch_name)
    fqbn = "arduino:avr:uno"

    # Create a test sketch
    assert run_command(f"sketch new {sketch_path}")

    # Get the build directory
    sketch_path_md5 = hashlib.md5(bytes(sketch_path)).hexdigest().upper()
    build_dir = Path(tempfile.gettempdir(), f"arduino-sketch-{sketch_path_md5}")

    assert run_command(f"compile -b {fqbn} {sketch_path} --verbose")

    # Breaks the build.options.json file
    build_options_json = build_dir / "build.options.json"
    with open(build_options_json, "w") as f:
        f.write("invalid json")

    assert run_command(f"compile -b {fqbn} {sketch_path} --verbose")


def test_compile_with_esp32_bundled_libraries(run_command, data_dir, copy_sketch):
    # Some esp cores have have bundled libraries that are optimize for that architecture,
    # it might happen that if the user has a library with the same name installed conflicts
    # can ensue and the wrong library is used for compilation, thus it fails.
    # This happens because for "historical" reasons these platform have their "name" key
    # in the "library.properties" flag suffixed with "(esp32)" or similar even though that
    # doesn't respect the libraries specification.
    # https://arduino.github.io/arduino-cli/latest/library-specification/#libraryproperties-file-format
    #
    # The reason those libraries have these suffixes is to avoid an annoying bug in the Java IDE
    # that would have caused the libraries that are both bundled with the core and the Java IDE to be
    # always marked as updatable. For more info see: https://github.com/arduino/Arduino/issues/4189
    assert run_command("update")

    # Update index with esp32 core and install it
    url = "https://dl.espressif.com/dl/package_esp32_index.json"
    core_version = "1.0.6"
    assert run_command(f"core update-index --additional-urls={url}")
    assert run_command(f"core install esp32:esp32@{core_version} --additional-urls={url}")

    # Install a library with the same name as one bundled with the core
    assert run_command("lib install SD")

    sketch_path = copy_sketch("sketch_with_sd_library")
    fqbn = "esp32:esp32:esp32"

    res = run_command(f"compile -b {fqbn} {sketch_path} --verbose")
    assert res.failed

    core_bundled_lib_path = Path(data_dir, "packages", "esp32", "hardware", "esp32", core_version, "libraries", "SD")
    cli_installed_lib_path = Path(data_dir, "libraries", "SD")
    expected_output = [
        'Multiple libraries were found for "SD.h"',
        f" Used: {core_bundled_lib_path}",
        f" Not used: {cli_installed_lib_path}",
    ]
    assert "\n".join(expected_output) not in res.stdout


def test_compile_with_esp8266_bundled_libraries(run_command, data_dir, copy_sketch):
    # Some esp cores have have bundled libraries that are optimize for that architecture,
    # it might happen that if the user has a library with the same name installed conflicts
    # can ensue and the wrong library is used for compilation, thus it fails.
    # This happens because for "historical" reasons these platform have their "name" key
    # in the "library.properties" flag suffixed with "(esp32)" or similar even though that
    # doesn't respect the libraries specification.
    # https://arduino.github.io/arduino-cli/latest/library-specification/#libraryproperties-file-format
    #
    # The reason those libraries have these suffixes is to avoid an annoying bug in the Java IDE
    # that would have caused the libraries that are both bundled with the core and the Java IDE to be
    # always marked as updatable. For more info see: https://github.com/arduino/Arduino/issues/4189
    assert run_command("update")

    # Update index with esp8266 core and install it
    url = "http://arduino.esp8266.com/stable/package_esp8266com_index.json"
    core_version = "2.7.4"
    assert run_command(f"core update-index --additional-urls={url}")
    assert run_command(f"core install esp8266:esp8266@{core_version} --additional-urls={url}")

    # Install a library with the same name as one bundled with the core
    assert run_command("lib install SD")

    sketch_path = copy_sketch("sketch_with_sd_library")
    fqbn = "esp8266:esp8266:generic"

    res = run_command(f"compile -b {fqbn} {sketch_path} --verbose")
    assert res.failed

    core_bundled_lib_path = Path(
        data_dir, "packages", "esp8266", "hardware", "esp8266", core_version, "libraries", "SD"
    )
    cli_installed_lib_path = Path(data_dir, "libraries", "SD")
    expected_output = [
        'Multiple libraries were found for "SD.h"',
        f" Used: {core_bundled_lib_path}",
        f" Not used: {cli_installed_lib_path}",
    ]
    assert "\n".join(expected_output) not in res.stdout
