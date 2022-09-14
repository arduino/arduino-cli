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

import tempfile
import hashlib
from pathlib import Path
import simplejson as json


def test_compile_with_export_binaries_flag(run_command, data_dir):
    # Init the environment explicitly
    run_command(["core", "update-index"])

    # Download latest AVR
    run_command(["core", "install", "arduino:avr"])

    sketch_name = "CompileWithExportBinariesFlag"
    sketch_path = Path(data_dir, sketch_name)
    fqbn = "arduino:avr:uno"

    # Create a test sketch
    assert run_command(["sketch", "new", sketch_path])

    # Test the --output-dir flag with absolute path
    result = run_command(["compile", "-b", fqbn, sketch_path, "--export-binaries"])
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
    run_command(["core", "update-index"])

    # Download latest AVR
    run_command(["core", "install", "arduino:avr"])

    sketch_name = "CompileWithBuildPath"
    sketch_path = Path(data_dir, sketch_name)
    fqbn = "arduino:avr:uno"

    # Create a test sketch
    result = run_command(["sketch", "new", sketch_path])
    assert result.ok
    assert f"Sketch created in: {sketch_path}" in result.stdout

    # Test the --build-path flag with absolute path
    build_path = Path(data_dir, "test_dir", "build_dir")
    result = run_command(["compile", "-b", fqbn, sketch_path, "--build-path", build_path])
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
    run_command(["core", "update-index"])

    # Download latest AVR
    run_command(["core", "install", "arduino:avr"])

    sketch_name = "CompileWithExportBinariesEnvVar"
    sketch_path = Path(data_dir, sketch_name)
    fqbn = "arduino:avr:uno"

    # Create a test sketch
    assert run_command(["sketch", "new", sketch_path])

    env = {
        "ARDUINO_DATA_DIR": data_dir,
        "ARDUINO_DOWNLOADS_DIR": downloads_dir,
        "ARDUINO_SKETCHBOOK_DIR": data_dir,
        "ARDUINO_SKETCH_ALWAYS_EXPORT_BINARIES": "true",
    }
    # Test compilation with export binaries env var set
    result = run_command(["compile", "-b", fqbn, sketch_path], custom_env=env)
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
    run_command(["core", "update-index"])

    # Download latest AVR
    run_command(["core", "install", "arduino:avr"])

    sketch_name = "CompileWithExportBinariesConfig"
    sketch_path = Path(data_dir, sketch_name)
    fqbn = "arduino:avr:uno"

    # Create a test sketch
    assert run_command(["sketch", "new", sketch_path])

    # Create settings with export binaries set to true
    env = {
        "ARDUINO_DATA_DIR": data_dir,
        "ARDUINO_DOWNLOADS_DIR": downloads_dir,
        "ARDUINO_SKETCHBOOK_DIR": data_dir,
        "ARDUINO_SKETCH_ALWAYS_EXPORT_BINARIES": "true",
    }
    assert run_command(["config", "init", "--dest-dir", "."], custom_env=env)

    # Test compilation with export binaries env var set
    result = run_command(["compile", "-b", fqbn, sketch_path])
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
    run_command(["core", "update-index"])

    # Download latest AVR
    run_command(["core", "install", "arduino:avr"])

    sketch_name = "CompileWithInvalidURL"
    sketch_path = Path(data_dir, sketch_name)
    fqbn = "arduino:avr:uno"

    # Create a test sketch
    assert run_command(["sketch", "new", sketch_path])

    # Create settings with custom invalid URL
    assert run_command(
        ["config", "init", "--dest-dir", ".", "--additional-urls", "https://example.com/package_example_index.json"]
    )

    # Verifies compilation fails cause of missing local index file
    res = run_command(["compile", "-b", fqbn, sketch_path])
    assert res.ok
    lines = [l.strip() for l in res.stderr.splitlines()]
    assert "Error initializing instance: Loading index file: loading json index file" in lines[0]
    expected_index_file = Path(data_dir, "package_example_index.json")
    assert f"loading json index file {expected_index_file}: " + f"open {expected_index_file}:" in lines[-1]


def test_compile_with_custom_libraries(run_command, copy_sketch):
    # Creates config with additional URL to install necessary core
    url = "http://arduino.esp8266.com/stable/package_esp8266com_index.json"
    assert run_command(["config", "init", "--dest-dir", ".", "--additional-urls", url])

    # Init the environment explicitly
    assert run_command(["update"])

    # Install core to compile
    assert run_command(["core", "install", "esp8266:esp8266"])

    sketch_path = copy_sketch("sketch_with_multiple_custom_libraries")
    fqbn = "esp8266:esp8266:nodemcu:xtal=80,vt=heap,eesz=4M1M,wipe=none,baud=115200"

    first_lib = Path(sketch_path, "libraries1")
    second_lib = Path(sketch_path, "libraries2")
    # This compile command has been taken from this issue:
    # https://github.com/arduino/arduino-cli/issues/973
    assert run_command(["compile", "--libraries", first_lib, "--libraries", second_lib, "-b", fqbn, sketch_path])


def test_compile_with_archives_and_long_paths(run_command):
    # Creates config with additional URL to install necessary core
    url = "http://arduino.esp8266.com/stable/package_esp8266com_index.json"
    assert run_command(["config", "init", "--dest-dir", ".", "--additional-urls", url])

    # Init the environment explicitly
    assert run_command(["update"])

    # Install core to compile
    assert run_command(["core", "install", "esp8266:esp8266@2.7.4"])

    # Install test library
    assert run_command(["lib", "install", "ArduinoIoTCloud"])

    result = run_command(["lib", "examples", "ArduinoIoTCloud", "--format", "json"])
    assert result.ok
    lib_output = json.loads(result.stdout)
    sketch_path = Path(lib_output[0]["library"]["install_dir"], "examples", "ArduinoIoTCloud-Advanced")

    assert run_command(["compile", "-b", "esp8266:esp8266:huzzah", sketch_path])


def test_compile_with_precompiled_library(run_command, data_dir):
    assert run_command(["update"])

    assert run_command(["core", "install", "arduino:samd@1.8.11"])
    fqbn = "arduino:samd:mkrzero"

    # Install precompiled library
    # For more information see:
    # https://arduino.github.io/arduino-cli/latest/library-specification/#precompiled-binaries
    assert run_command(["lib", "install", "BSEC Software Library@1.5.1474"])
    sketch_folder = Path(data_dir, "libraries", "BSEC_Software_Library", "examples", "basic")

    # Compile and verify dependencies detection for fully precompiled library is not skipped
    result = run_command(["compile", "-b", fqbn, sketch_folder, "-v"])
    assert result.ok
    assert "Skipping dependencies detection for precompiled library BSEC Software Library" not in result.stdout
