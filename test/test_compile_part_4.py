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
