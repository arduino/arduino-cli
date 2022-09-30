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
import shutil
from git import Repo
from pathlib import Path


def test_compile_manually_installed_platform_using_boards_local_txt(run_command, data_dir):
    assert run_command(["update"])

    sketch_name = "CompileSketchManuallyInstalledPlatformUsingBoardsLocalTxt"
    sketch_path = Path(data_dir, sketch_name)
    fqbn = "arduino-beta-development:avr:nessuno"
    assert run_command(["sketch", "new", sketch_path])

    # Manually installs a core in sketchbooks hardware folder
    git_url = "https://github.com/arduino/ArduinoCore-avr.git"
    repo_dir = Path(data_dir, "hardware", "arduino-beta-development", "avr")
    assert Repo.clone_from(git_url, repo_dir, multi_options=["-b 1.8.3"])

    # Installs also the same core via CLI so all the necessary tools are installed
    assert run_command(["core", "install", "arduino:avr@1.8.3"])

    # Verifies compilation fails because board doesn't exist
    res = run_command(["compile", "--clean", "-b", fqbn, sketch_path])
    assert res.failed
    assert (
        "Error during build: Error resolving FQBN: board arduino-beta-development:avr:nessuno not found" in res.stderr
    )

    # Use custom boards.local.txt with made arduino:avr:nessuno board
    boards_local_txt = Path(repo_dir, "boards.local.txt")
    shutil.copyfile(Path(__file__).parent / "testdata" / "boards.local.txt", boards_local_txt)

    assert run_command(["compile", "--clean", "-b", fqbn, sketch_path])


def test_recompile_with_different_library(run_command, data_dir):
    assert run_command(["update"])

    assert run_command(["core", "install", "arduino:avr@1.8.3"])

    sketch_name = "RecompileCompileSketchWithDifferentLibrary"
    sketch_path = Path(data_dir, sketch_name)
    fqbn = "arduino:avr:uno"

    # Install library
    assert run_command(["lib", "install", "WiFi101"])

    # Manually installs the same library already installed
    git_url = "https://github.com/arduino-libraries/WiFi101.git"
    manually_install_lib_path = Path(data_dir, "my-libraries", "WiFi101")
    assert Repo.clone_from(git_url, manually_install_lib_path, multi_options=["-b 0.16.1"])

    # Create new sketch and add library include
    assert run_command(["sketch", "new", sketch_path])
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
    res = run_command(["compile", "-b", fqbn, "--library", manually_install_lib_path, sketch_path, "-v"])
    assert res.ok
    obj_path = build_dir / "libraries" / "WiFi101" / "WiFi.cpp.o"
    assert f"Using previously compiled file: {obj_path}" not in res.stdout

    # Compile again using library installed from CLI
    res = run_command(["compile", "-b", fqbn, sketch_path, "-v"])
    assert res.ok
    obj_path = build_dir / "libraries" / "WiFi101" / "WiFi.cpp.o"
    assert f"Using previously compiled file: {obj_path}" not in res.stdout


def test_compile_with_conflicting_libraries_include(run_command, data_dir, copy_sketch):
    assert run_command(["update"])

    assert run_command(["core", "install", "arduino:avr@1.8.3"])

    # Install conflicting libraries
    git_url = "https://github.com/pstolarz/OneWireNg.git"
    one_wire_ng_lib_path = Path(data_dir, "libraries", "onewireng_0_8_1")
    assert Repo.clone_from(git_url, one_wire_ng_lib_path, multi_options=["-b 0.8.1"])

    git_url = "https://github.com/PaulStoffregen/OneWire.git"
    one_wire_lib_path = Path(data_dir, "libraries", "onewire_2_3_5")
    assert Repo.clone_from(git_url, one_wire_lib_path, multi_options=["-b v2.3.5"])

    sketch_path = copy_sketch("sketch_with_conflicting_libraries_include")
    fqbn = "arduino:avr:uno"

    res = run_command(["compile", "-b", fqbn, sketch_path, "--verbose"])
    assert res.ok
    expected_output = [
        'Multiple libraries were found for "OneWire.h"',
        f"  Used: {one_wire_lib_path}",
        f"  Not used: {one_wire_ng_lib_path}",
    ]
    assert "\n".join(expected_output) in res.stdout


def test_compile_with_invalid_build_options_json(run_command, data_dir):
    assert run_command(["update"])

    assert run_command(["core", "install", "arduino:avr@1.8.3"])

    sketch_name = "CompileInvalidBuildOptionsJson"
    sketch_path = Path(data_dir, sketch_name)
    fqbn = "arduino:avr:uno"

    # Create a test sketch
    assert run_command(["sketch", "new", sketch_path])

    # Get the build directory
    sketch_path_md5 = hashlib.md5(bytes(sketch_path)).hexdigest().upper()
    build_dir = Path(tempfile.gettempdir(), f"arduino-sketch-{sketch_path_md5}")

    assert run_command(["compile", "-b", fqbn, sketch_path, "--verbose"])

    # Breaks the build.options.json file
    build_options_json = build_dir / "build.options.json"
    with open(build_options_json, "w") as f:
        f.write("invalid json")

    assert run_command(["compile", "-b", fqbn, sketch_path, "--verbose"])


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
    assert run_command(["update"])

    # Update index with esp32 core and install it
    url = "https://raw.githubusercontent.com/espressif/arduino-esp32/gh-pages/package_esp32_index.json"
    core_version = "1.0.6"
    assert run_command(["core", "update-index", f"--additional-urls={url}"])
    assert run_command(["core", "install", f"esp32:esp32@{core_version}", f"--additional-urls={url}"])

    # Install a library with the same name as one bundled with the core
    assert run_command(["lib", "install", "SD"])

    sketch_path = copy_sketch("sketch_with_sd_library")
    fqbn = "esp32:esp32:esp32"

    res = run_command(["compile", "-b", fqbn, sketch_path, "--verbose"])
    assert res.failed

    core_bundled_lib_path = Path(data_dir, "packages", "esp32", "hardware", "esp32", core_version, "libraries", "SD")
    cli_installed_lib_path = Path(data_dir, "libraries", "SD")
    expected_output = [
        'Multiple libraries were found for "SD.h"',
        f"  Used: {core_bundled_lib_path}",
        f"  Not used: {cli_installed_lib_path}",
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
    assert run_command(["update"])

    # Update index with esp8266 core and install it
    url = "http://arduino.esp8266.com/stable/package_esp8266com_index.json"
    core_version = "2.7.4"
    assert run_command(["core", "update-index", f"--additional-urls={url}"])
    assert run_command(["core", "install", f"esp8266:esp8266@{core_version}", f"--additional-urls={url}"])

    # Install a library with the same name as one bundled with the core
    assert run_command(["lib", "install", "SD"])

    sketch_path = copy_sketch("sketch_with_sd_library")
    fqbn = "esp8266:esp8266:generic"

    res = run_command(["compile", "-b", fqbn, sketch_path, "--verbose"])
    assert res.failed

    core_bundled_lib_path = Path(
        data_dir, "packages", "esp8266", "hardware", "esp8266", core_version, "libraries", "SD"
    )
    cli_installed_lib_path = Path(data_dir, "libraries", "SD")
    expected_output = [
        'Multiple libraries were found for "SD.h"',
        f"  Used: {core_bundled_lib_path}",
        f"  Not used: {cli_installed_lib_path}",
    ]
    assert "\n".join(expected_output) not in res.stdout


def test_generate_compile_commands_json_resilience(run_command, data_dir, copy_sketch):
    assert run_command(["update"])

    # check it didn't fail with esp32@2.0.1 that has a prebuild hook that must run:
    # https://github.com/arduino/arduino-cli/issues/1547
    url = "https://raw.githubusercontent.com/espressif/arduino-esp32/gh-pages/package_esp32_index.json"
    assert run_command(["core", "update-index", f"--additional-urls={url}"])
    assert run_command(["core", "install", "esp32:esp32@2.0.1", f"--additional-urls={url}"])
    sketch_path = copy_sketch("sketch_simple")
    assert run_command(["compile", "-b", "esp32:esp32:featheresp32", "--only-compilation-database", sketch_path])

    # check it didn't fail on a sketch with a missing include
    sketch_path = copy_sketch("sketch_with_missing_include")
    assert run_command(["compile", "-b", "esp32:esp32:featheresp32", "--only-compilation-database", sketch_path])


def test_compile_sketch_with_tpp_file_include(run_command, copy_sketch):
    assert run_command(["update"])

    # Download latest AVR
    run_command(["core", "install", "arduino:avr"])

    sketch_name = "sketch_with_tpp_file_include"
    sketch_path = copy_sketch(sketch_name)
    fqbn = "arduino:avr:uno"

    assert run_command(["compile", "-b", fqbn, sketch_path, "--verbose"])


def test_compile_sketch_with_ipp_file_include(run_command, copy_sketch):
    assert run_command(["update"])

    # Download latest AVR
    run_command(["core", "install", "arduino:avr"])

    sketch_name = "sketch_with_ipp_file_include"
    sketch_path = copy_sketch(sketch_name)
    fqbn = "arduino:avr:uno"

    assert run_command(["compile", "-b", fqbn, sketch_path, "--verbose"])


def test_compile_with_relative_build_path(run_command, data_dir, copy_sketch):
    assert run_command(["update"])

    run_command(["core", "install", "arduino:avr@1.8.3"])

    sketch_name = "sketch_simple"
    sketch_path = copy_sketch(sketch_name)
    fqbn = "arduino:avr:uno"

    build_path = Path("..", "build_path")
    working_dir = Path(data_dir, "working_dir")
    working_dir.mkdir()
    assert run_command(
        ["compile", "-b", fqbn, "--build-path", build_path, sketch_path, "-v"],
        custom_working_dir=working_dir,
    )

    absolute_build_path = Path(data_dir, "build_path")
    built_files = [f.name for f in absolute_build_path.glob("*")]
    assert f"{sketch_name}.ino.eep" in built_files
    assert f"{sketch_name}.ino.elf" in built_files
    assert f"{sketch_name}.ino.hex" in built_files
    assert f"{sketch_name}.ino.with_bootloader.bin" in built_files
    assert f"{sketch_name}.ino.with_bootloader.hex" in built_files
    assert "build.options.json" in built_files
    assert "compile_commands.json" in built_files
    assert "core" in built_files
    assert "includes.cache" in built_files
    assert "libraries" in built_files
    assert "preproc" in built_files
    assert "sketch" in built_files


def test_compile_without_upload_and_fqbn(run_command, data_dir):
    assert run_command(["update"])

    # Create a sketch
    sketch_name = "SketchSimple"
    sketch_path = Path(data_dir, sketch_name)
    assert run_command(["sketch", "new", sketch_path])

    res = run_command(["compile", sketch_path])
    assert res.failed
    assert "Missing FQBN (Fully Qualified Board Name)" in res.stderr


def test_compile_non_installed_platform_with_wrong_packager_and_arch(run_command, data_dir):
    assert run_command(["update"])

    # Create a sketch
    sketch_name = "SketchSimple"
    sketch_path = Path(data_dir, sketch_name)
    assert run_command(["sketch", "new", sketch_path])

    # Compile with wrong packager
    res = run_command(["compile", "-b", "wrong:avr:uno", sketch_path])
    assert res.failed
    assert "Error during build: Platform 'wrong:avr' not found: platform not installed" in res.stderr
    assert "Platform wrong:avr is not found in any known index" in res.stderr

    # Compile with wrong arch
    res = run_command(["compile", "-b", "arduino:wrong:uno", sketch_path])
    assert res.failed
    assert "Error during build: Platform 'arduino:wrong' not found: platform not installed" in res.stderr
    assert "Platform arduino:wrong is not found in any known index" in res.stderr


def test_compile_with_known_platform_not_installed(run_command, data_dir):
    assert run_command(["update"])

    # Create a sketch
    sketch_name = "SketchSimple"
    sketch_path = Path(data_dir, sketch_name)
    assert run_command(["sketch", "new", sketch_path])

    # Try to compile using a platform found in the index but not installed
    res = run_command(["compile", "-b", "arduino:avr:uno", sketch_path])
    assert res.failed
    assert "Error during build: Platform 'arduino:avr' not found: platform not installed" in res.stderr
    # Verifies command to fix error is shown to user
    assert "Try running `arduino-cli core install arduino:avr`" in res.stderr


def test_compile_with_fake_secure_boot_core(run_command, data_dir):
    assert run_command(["update"])

    assert run_command(["core", "install", "arduino:avr@1.8.3"])

    sketch_name = "SketchSimple"
    sketch_path = Path(data_dir, sketch_name)
    fqbn = "arduino:avr:uno"

    assert run_command(["sketch", "new", sketch_path])

    # Verifies compilation works
    assert run_command(["compile", "--clean", "-b", fqbn, sketch_path])

    # Overrides default platform adding secure_boot support using platform.local.txt
    avr_platform_path = Path(data_dir, "packages", "arduino", "hardware", "avr", "1.8.3", "platform.local.txt")
    test_platform_name = "platform_with_secure_boot"
    shutil.copyfile(
        Path(__file__).parent / "testdata" / test_platform_name / "platform.local.txt",
        avr_platform_path,
    )

    # Overrides default board adding secure boot support using board.local.txt
    avr_board_path = Path(data_dir, "packages", "arduino", "hardware", "avr", "1.8.3", "boards.local.txt")
    shutil.copyfile(
        Path(__file__).parent / "testdata" / test_platform_name / "boards.local.txt",
        avr_board_path,
    )

    # Verifies compilation works with secure boot disabled
    res = run_command(["compile", "--clean", "-b", fqbn + ":security=none", sketch_path, "-v"])
    assert res.ok
    assert "echo exit" in res.stdout

    # Verifies compilation works with secure boot enabled
    res = run_command(["compile", "--clean", "-b", fqbn + ":security=sien", sketch_path, "-v"])
    assert res.ok
    assert "Default_Keys/default-signing-key.pem" in res.stdout
    assert "Default_Keys/default-encrypt-key.pem" in res.stdout

    # Verifies compilation does not work with secure boot enabled and using only one flag
    res = run_command(
        [
            "compile",
            "--clean",
            "-b",
            fqbn + ":security=sien",
            sketch_path,
            "--keys-keychain",
            data_dir,
            "-v",
        ]
    )
    assert res.failed
    assert "Flag --sign-key is mandatory when used in conjunction with flag --keys-keychain" in res.stderr

    # Verifies compilation works with secure boot enabled and when overriding the sign key and encryption key used
    keys_dir = Path(data_dir, "keys_dir")
    keys_dir.mkdir()
    sign_key_path = Path(keys_dir, "my-sign-key.pem")
    sign_key_path.touch()
    encrypt_key_path = Path(keys_dir, "my-encrypt-key.pem")
    encrypt_key_path.touch()
    res = run_command(
        [
            "compile",
            "--clean",
            "-b",
            fqbn + ":security=sien",
            sketch_path,
            "--keys-keychain",
            keys_dir,
            "--sign-key",
            "my-sign-key.pem",
            "--encrypt-key",
            "my-encrypt-key.pem",
            "-v",
        ]
    )
    assert res.ok
    assert "my-sign-key.pem" in res.stdout
    assert "my-encrypt-key.pem" in res.stdout
