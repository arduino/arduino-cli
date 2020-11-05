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
import zipfile
from pathlib import Path

import pytest

from test.common import running_on_ci


@pytest.mark.skipif(
    running_on_ci() and platform.system() == "Windows",
    reason="Test disabled on Github Actions Win VM until tmpdir inconsistent behavior bug is fixed",
)
def test_sketch_new(run_command, working_dir):
    # Create a test sketch in current directory
    current_path = working_dir
    sketch_name = "SketchNewIntegrationTest"
    current_sketch_path = os.path.join(current_path, sketch_name)
    result = run_command("sketch new {}".format(sketch_name))
    assert result.ok
    assert "Sketch created in: {}".format(current_sketch_path) in result.stdout
    assert os.path.isfile(os.path.join(current_sketch_path, sketch_name + ".ino"))

    # Create a test sketch in current directory but using an absolute path
    sketch_name = "SketchNewIntegrationTestAbsolute"
    current_sketch_path = os.path.join(current_path, sketch_name)
    result = run_command("sketch new {}".format(current_sketch_path))
    assert result.ok
    assert "Sketch created in: {}".format(current_sketch_path) in result.stdout
    assert os.path.isfile(os.path.join(current_sketch_path, sketch_name + ".ino"))

    # Create a test sketch in current directory subpath but using an absolute path
    sketch_name = "SketchNewIntegrationTestSubpath"
    sketch_subpath = os.path.join("subpath", sketch_name)
    current_sketch_path = os.path.join(current_path, sketch_subpath)
    result = run_command("sketch new {}".format(sketch_subpath))
    assert result.ok
    assert "Sketch created in: {}".format(current_sketch_path) in result.stdout
    assert os.path.isfile(os.path.join(current_sketch_path, sketch_name + ".ino"))

    # Create a test sketch in current directory using .ino extension
    sketch_name = "SketchNewIntegrationTestDotIno"
    current_sketch_path = os.path.join(current_path, sketch_name)
    result = run_command("sketch new {}".format(sketch_name + ".ino"))
    assert result.ok
    assert "Sketch created in: {}".format(current_sketch_path) in result.stdout
    assert os.path.isfile(os.path.join(current_sketch_path, sketch_name + ".ino"))


def verify_zip_contains_sketch_excluding_build_dir(files):
    assert "sketch_simple/doc.txt" in files
    assert "sketch_simple/header.h" in files
    assert "sketch_simple/merged_sketch.txt" in files
    assert "sketch_simple/old.pde" in files
    assert "sketch_simple/other.ino" in files
    assert "sketch_simple/s_file.S" in files
    assert "sketch_simple/sketch_simple.ino" in files
    assert "sketch_simple/src/helper.h" in files
    assert "sketch_simple/build/adafruit.samd.adafruit_feather_m0/sketch_simple.ino.hex" not in files
    assert "sketch_simple/build/adafruit.samd.adafruit_feather_m0/sketch_simple.ino.map" not in files
    assert "sketch_simple/build/arduino.avr.uno/sketch_simple.ino.eep" not in files
    assert "sketch_simple/build/arduino.avr.uno/sketch_simple.ino.hex" not in files
    assert "sketch_simple/build/arduino.avr.uno/sketch_simple.ino.with_bootloader.hex" not in files


def verify_zip_contains_sketch_including_build_dir(files):
    assert "sketch_simple/doc.txt" in files
    assert "sketch_simple/header.h" in files
    assert "sketch_simple/merged_sketch.txt" in files
    assert "sketch_simple/old.pde" in files
    assert "sketch_simple/other.ino" in files
    assert "sketch_simple/s_file.S" in files
    assert "sketch_simple/sketch_simple.ino" in files
    assert "sketch_simple/src/helper.h" in files
    assert "sketch_simple/build/adafruit.samd.adafruit_feather_m0/sketch_simple.ino.hex" in files
    assert "sketch_simple/build/adafruit.samd.adafruit_feather_m0/sketch_simple.ino.map" in files
    assert "sketch_simple/build/arduino.avr.uno/sketch_simple.ino.eep" in files
    assert "sketch_simple/build/arduino.avr.uno/sketch_simple.ino.hex" in files
    assert "sketch_simple/build/arduino.avr.uno/sketch_simple.ino.with_bootloader.hex" in files


def test_sketch_archive_no_args(run_command, copy_sketch, working_dir):
    result = run_command("sketch archive", copy_sketch("sketch_simple"))
    print(result.stderr)
    assert result.ok

    archive = zipfile.ZipFile(f"{working_dir}/sketch_simple.zip")
    archive_files = archive.namelist()

    verify_zip_contains_sketch_excluding_build_dir(archive_files)

    archive.close()


def test_sketch_archive_dot_arg(run_command, copy_sketch, working_dir):
    result = run_command("sketch archive .", copy_sketch("sketch_simple"))
    assert result.ok

    archive = zipfile.ZipFile(f"{working_dir}/sketch_simple.zip")
    archive_files = archive.namelist()

    verify_zip_contains_sketch_excluding_build_dir(archive_files)

    archive.close()


def test_sketch_archive_dot_arg_relative_zip_path(run_command, copy_sketch, working_dir):
    # Creates a folder where to save the zip
    archives_folder = f"{working_dir}/my_archives/"
    Path(archives_folder).mkdir()

    result = run_command("sketch archive . ../my_archives", copy_sketch("sketch_simple"))
    assert result.ok

    archive = zipfile.ZipFile(f"{working_dir}/my_archives/sketch_simple.zip")
    archive_files = archive.namelist()

    verify_zip_contains_sketch_excluding_build_dir(archive_files)

    archive.close()


def test_sketch_archive_dot_arg_absolute_zip_path(run_command, copy_sketch, working_dir):
    # Creates a folder where to save the zip
    archives_folder = f"{working_dir}/my_archives/"
    Path(archives_folder).mkdir()

    result = run_command(f'sketch archive . "{archives_folder}"', copy_sketch("sketch_simple"))
    assert result.ok

    archive = zipfile.ZipFile(f"{archives_folder}/sketch_simple.zip")
    archive_files = archive.namelist()

    verify_zip_contains_sketch_excluding_build_dir(archive_files)

    archive.close()


def test_sketch_archive_dot_arg_relative_zip_path_and_name_without_extension(run_command, copy_sketch, working_dir):
    # Creates a folder where to save the zip
    archives_folder = f"{working_dir}/my_archives/"
    Path(archives_folder).mkdir()

    result = run_command("sketch archive . ../my_archives/my_custom_sketch", copy_sketch("sketch_simple"))
    assert result.ok

    archive = zipfile.ZipFile(f"{working_dir}/my_archives/my_custom_sketch.zip")
    archive_files = archive.namelist()

    verify_zip_contains_sketch_excluding_build_dir(archive_files)

    archive.close()


def test_sketch_archive_dot_arg_absolute_zip_path_and_name_without_extension(run_command, copy_sketch, working_dir):
    # Creates a folder where to save the zip
    archives_folder = f"{working_dir}/my_archives/"
    Path(archives_folder).mkdir()

    result = run_command(f'sketch archive . "{archives_folder}/my_custom_sketch"', copy_sketch("sketch_simple"))
    assert result.ok

    archive = zipfile.ZipFile(f"{archives_folder}/my_custom_sketch.zip")
    archive_files = archive.namelist()

    verify_zip_contains_sketch_excluding_build_dir(archive_files)

    archive.close()


def test_sketch_archive_dot_arg_custom_zip_path_and_name_with_extension(run_command, copy_sketch, working_dir):
    # Creates a folder where to save the zip
    archives_folder = f"{working_dir}/my_archives/"
    Path(archives_folder).mkdir()

    result = run_command(f'sketch archive . "{archives_folder}/my_custom_sketch.zip"', copy_sketch("sketch_simple"))
    assert result.ok

    archive = zipfile.ZipFile(f"{archives_folder}/my_custom_sketch.zip")
    archive_files = archive.namelist()

    verify_zip_contains_sketch_excluding_build_dir(archive_files)

    archive.close()


def test_sketch_archive_relative_sketch_path(run_command, copy_sketch, working_dir):
    copy_sketch("sketch_simple")
    result = run_command("sketch archive ./sketch_simple")
    assert result.ok

    archive = zipfile.ZipFile(f"{working_dir}/sketch_simple.zip")
    archive_files = archive.namelist()

    verify_zip_contains_sketch_excluding_build_dir(archive_files)

    archive.close()


def test_sketch_archive_absolute_sketch_path(run_command, copy_sketch, working_dir):
    result = run_command(f'sketch archive "{working_dir}/sketch_simple"', copy_sketch("sketch_simple"))
    assert result.ok

    archive = zipfile.ZipFile(f"{working_dir}/sketch_simple.zip")
    archive_files = archive.namelist()

    verify_zip_contains_sketch_excluding_build_dir(archive_files)

    archive.close()


def test_sketch_archive_relative_sketch_path_with_relative_zip_path(run_command, copy_sketch, working_dir):
    copy_sketch("sketch_simple")
    # Creates a folder where to save the zip
    archives_folder = f"{working_dir}/my_archives/"
    Path(archives_folder).mkdir()

    result = run_command("sketch archive ./sketch_simple ./my_archives")
    assert result.ok

    archive = zipfile.ZipFile(f"{working_dir}/my_archives/sketch_simple.zip")
    archive_files = archive.namelist()

    verify_zip_contains_sketch_excluding_build_dir(archive_files)

    archive.close()


def test_sketch_archive_relative_sketch_path_with_absolute_zip_path(run_command, copy_sketch, working_dir):
    copy_sketch("sketch_simple")
    # Creates a folder where to save the zip
    archives_folder = f"{working_dir}/my_archives/"
    Path(archives_folder).mkdir()

    result = run_command(f'sketch archive ./sketch_simple "{archives_folder}"')
    assert result.ok

    archive = zipfile.ZipFile(f"{archives_folder}/sketch_simple.zip")
    archive_files = archive.namelist()

    verify_zip_contains_sketch_excluding_build_dir(archive_files)

    archive.close()


def test_sketch_archive_relative_sketch_path_with_relative_zip_path_and_name_without_extension(
    run_command, copy_sketch, working_dir
):
    copy_sketch("sketch_simple")
    # Creates a folder where to save the zip
    archives_folder = f"{working_dir}/my_archives/"
    Path(archives_folder).mkdir()

    result = run_command("sketch archive ./sketch_simple ./my_archives/my_custom_sketch")
    assert result.ok

    archive = zipfile.ZipFile(f"{working_dir}/my_archives/my_custom_sketch.zip")
    archive_files = archive.namelist()

    verify_zip_contains_sketch_excluding_build_dir(archive_files)

    archive.close()


def test_sketch_archive_relative_sketch_path_with_relative_zip_path_and_name_with_extension(
    run_command, copy_sketch, working_dir
):
    copy_sketch("sketch_simple")
    # Creates a folder where to save the zip
    archives_folder = f"{working_dir}/my_archives/"
    Path(archives_folder).mkdir()

    result = run_command("sketch archive ./sketch_simple ./my_archives/my_custom_sketch.zip")
    assert result.ok

    archive = zipfile.ZipFile(f"{working_dir}/my_archives/my_custom_sketch.zip")
    archive_files = archive.namelist()

    verify_zip_contains_sketch_excluding_build_dir(archive_files)

    archive.close()


def test_sketch_archive_relative_sketch_path_with_absolute_zip_path_and_name_without_extension(
    run_command, copy_sketch, working_dir
):
    copy_sketch("sketch_simple")
    # Creates a folder where to save the zip
    archives_folder = f"{working_dir}/my_archives/"
    Path(archives_folder).mkdir()

    result = run_command(f'sketch archive ./sketch_simple "{archives_folder}/my_custom_sketch"')
    assert result.ok

    archive = zipfile.ZipFile(f"{archives_folder}/my_custom_sketch.zip")
    archive_files = archive.namelist()

    verify_zip_contains_sketch_excluding_build_dir(archive_files)

    archive.close()


def test_sketch_archive_relative_sketch_path_with_absolute_zip_path_and_name_with_extension(
    run_command, copy_sketch, working_dir
):
    copy_sketch("sketch_simple")
    # Creates a folder where to save the zip
    archives_folder = f"{working_dir}/my_archives/"
    Path(archives_folder).mkdir()

    result = run_command(f'sketch archive ./sketch_simple "{archives_folder}/my_custom_sketch.zip"')
    assert result.ok

    archive = zipfile.ZipFile(f"{archives_folder}/my_custom_sketch.zip")
    archive_files = archive.namelist()

    verify_zip_contains_sketch_excluding_build_dir(archive_files)

    archive.close()


def test_sketch_archive_absolute_sketch_path_with_relative_zip_path(run_command, copy_sketch, working_dir):
    copy_sketch("sketch_simple")
    # Creates a folder where to save the zip
    archives_folder = f"{working_dir}/my_archives/"
    Path(archives_folder).mkdir()

    result = run_command(f'sketch archive "{working_dir}/sketch_simple" ./my_archives')
    assert result.ok

    archive = zipfile.ZipFile(f"{working_dir}/my_archives/sketch_simple.zip")
    archive_files = archive.namelist()

    verify_zip_contains_sketch_excluding_build_dir(archive_files)

    archive.close()


def test_sketch_archive_absolute_sketch_path_with_absolute_zip_path(run_command, copy_sketch, working_dir):
    # Creates a folder where to save the zip
    archives_folder = f"{working_dir}/my_archives/"
    Path(archives_folder).mkdir()

    result = run_command(
        f'sketch archive "{working_dir}/sketch_simple" "{archives_folder}"', copy_sketch("sketch_simple")
    )
    assert result.ok

    archive = zipfile.ZipFile(f"{archives_folder}/sketch_simple.zip")
    archive_files = archive.namelist()

    verify_zip_contains_sketch_excluding_build_dir(archive_files)

    archive.close()


def test_sketch_archive_absolute_sketch_path_with_relative_zip_path_and_name_without_extension(
    run_command, copy_sketch, working_dir
):
    copy_sketch("sketch_simple")
    # Creates a folder where to save the zip
    archives_folder = f"{working_dir}/my_archives/"
    Path(archives_folder).mkdir()

    result = run_command(f'sketch archive "{working_dir}/sketch_simple" ./my_archives/my_custom_sketch')
    assert result.ok

    archive = zipfile.ZipFile(f"{working_dir}/my_archives/my_custom_sketch.zip")
    archive_files = archive.namelist()

    verify_zip_contains_sketch_excluding_build_dir(archive_files)

    archive.close()


def test_sketch_archive_absolute_sketch_path_with_relative_zip_path_and_name_with_extension(
    run_command, copy_sketch, working_dir
):
    copy_sketch("sketch_simple")
    # Creates a folder where to save the zip
    archives_folder = f"{working_dir}/my_archives/"
    Path(archives_folder).mkdir()

    result = run_command(f'sketch archive "{working_dir}/sketch_simple" ./my_archives/my_custom_sketch.zip')
    assert result.ok

    archive = zipfile.ZipFile(f"{working_dir}/my_archives/my_custom_sketch.zip")
    archive_files = archive.namelist()

    verify_zip_contains_sketch_excluding_build_dir(archive_files)

    archive.close()


def test_sketch_archive_absolute_sketch_path_with_absolute_zip_path_and_name_without_extension(
    run_command, copy_sketch, working_dir
):
    # Creates a folder where to save the zip
    archives_folder = f"{working_dir}/my_archives/"
    Path(archives_folder).mkdir()

    result = run_command(
        f'sketch archive "{working_dir}/sketch_simple" "{archives_folder}/my_custom_sketch"',
        copy_sketch("sketch_simple"),
    )
    assert result.ok

    archive = zipfile.ZipFile(f"{archives_folder}/my_custom_sketch.zip")
    archive_files = archive.namelist()

    verify_zip_contains_sketch_excluding_build_dir(archive_files)

    archive.close()


def test_sketch_archive_absolute_sketch_path_with_absolute_zip_path_and_name_with_extension(
    run_command, copy_sketch, working_dir
):
    # Creates a folder where to save the zip
    archives_folder = f"{working_dir}/my_archives/"
    Path(archives_folder).mkdir()

    result = run_command(
        f'sketch archive "{working_dir}/sketch_simple" "{archives_folder}/my_custom_sketch.zip"',
        copy_sketch("sketch_simple"),
    )
    assert result.ok

    archive = zipfile.ZipFile(f"{archives_folder}/my_custom_sketch.zip")
    archive_files = archive.namelist()

    verify_zip_contains_sketch_excluding_build_dir(archive_files)

    archive.close()


def test_sketch_archive_no_args_with_include_build_dir_flag(run_command, copy_sketch, working_dir):
    result = run_command("sketch archive --include-build-dir", copy_sketch("sketch_simple"))
    assert result.ok

    archive = zipfile.ZipFile(f"{working_dir}/sketch_simple.zip")
    archive_files = archive.namelist()

    verify_zip_contains_sketch_including_build_dir(archive_files)

    archive.close()


def test_sketch_archive_dot_arg_with_include_build_dir_flag(run_command, copy_sketch, working_dir):
    result = run_command("sketch archive . --include-build-dir", copy_sketch("sketch_simple"))
    assert result.ok

    archive = zipfile.ZipFile(f"{working_dir}/sketch_simple.zip")
    archive_files = archive.namelist()

    verify_zip_contains_sketch_including_build_dir(archive_files)

    archive.close()


def test_sketch_archive_dot_arg_relative_zip_path_with_include_build_dir_flag(run_command, copy_sketch, working_dir):
    # Creates a folder where to save the zip
    archives_folder = f"{working_dir}/my_archives/"
    Path(archives_folder).mkdir()

    result = run_command("sketch archive . ../my_archives --include-build-dir", copy_sketch("sketch_simple"))
    assert result.ok

    archive = zipfile.ZipFile(f"{working_dir}/my_archives/sketch_simple.zip")
    archive_files = archive.namelist()

    verify_zip_contains_sketch_including_build_dir(archive_files)

    archive.close()


def test_sketch_archive_dot_arg_absolute_zip_path_with_include_build_dir_flag(run_command, copy_sketch, working_dir):
    # Creates a folder where to save the zip
    archives_folder = f"{working_dir}/my_archives/"
    Path(archives_folder).mkdir()

    result = run_command(f'sketch archive . "{archives_folder}" --include-build-dir', copy_sketch("sketch_simple"))
    assert result.ok

    archive = zipfile.ZipFile(f"{archives_folder}/sketch_simple.zip")
    archive_files = archive.namelist()

    verify_zip_contains_sketch_including_build_dir(archive_files)

    archive.close()


def test_sketch_archive_dot_arg_relative_zip_path_and_name_without_extension_with_include_build_dir_flag(
    run_command, copy_sketch, working_dir
):
    # Creates a folder where to save the zip
    archives_folder = f"{working_dir}/my_archives/"
    Path(archives_folder).mkdir()

    result = run_command(
        "sketch archive . ../my_archives/my_custom_sketch --include-build-dir", copy_sketch("sketch_simple")
    )
    assert result.ok

    archive = zipfile.ZipFile(f"{working_dir}/my_archives/my_custom_sketch.zip")
    archive_files = archive.namelist()

    verify_zip_contains_sketch_including_build_dir(archive_files)

    archive.close()


def test_sketch_archive_dot_arg_absolute_zip_path_and_name_without_extension_with_include_build_dir_flag(
    run_command, copy_sketch, working_dir
):
    # Creates a folder where to save the zip
    archives_folder = f"{working_dir}/my_archives/"
    Path(archives_folder).mkdir()

    result = run_command(
        f'sketch archive . "{archives_folder}/my_custom_sketch" --include-build-dir', copy_sketch("sketch_simple")
    )
    assert result.ok

    archive = zipfile.ZipFile(f"{archives_folder}/my_custom_sketch.zip")
    archive_files = archive.namelist()

    verify_zip_contains_sketch_including_build_dir(archive_files)

    archive.close()


def test_sketch_archive_dot_arg_custom_zip_path_and_name_with_extension_with_include_build_dir_flag(
    run_command, copy_sketch, working_dir
):
    # Creates a folder where to save the zip
    archives_folder = f"{working_dir}/my_archives/"
    Path(archives_folder).mkdir()

    result = run_command(
        f'sketch archive . "{archives_folder}/my_custom_sketch.zip" --include-build-dir', copy_sketch("sketch_simple")
    )
    assert result.ok

    archive = zipfile.ZipFile(f"{archives_folder}/my_custom_sketch.zip")
    archive_files = archive.namelist()

    verify_zip_contains_sketch_including_build_dir(archive_files)

    archive.close()


def test_sketch_archive_relative_sketch_path_with_include_build_dir_flag(run_command, copy_sketch, working_dir):
    copy_sketch("sketch_simple")
    result = run_command("sketch archive ./sketch_simple --include-build-dir")
    assert result.ok

    archive = zipfile.ZipFile(f"{working_dir}/sketch_simple.zip")
    archive_files = archive.namelist()

    verify_zip_contains_sketch_including_build_dir(archive_files)

    archive.close()


def test_sketch_archive_absolute_sketch_path_with_include_build_dir_flag(run_command, copy_sketch, working_dir):
    result = run_command(
        f'sketch archive "{working_dir}/sketch_simple" --include-build-dir', copy_sketch("sketch_simple")
    )
    assert result.ok

    archive = zipfile.ZipFile(f"{working_dir}/sketch_simple.zip")
    archive_files = archive.namelist()

    verify_zip_contains_sketch_including_build_dir(archive_files)

    archive.close()


def test_sketch_archive_relative_sketch_path_with_relative_zip_path_with_include_build_dir_flag(
    run_command, copy_sketch, working_dir
):
    copy_sketch("sketch_simple")
    # Creates a folder where to save the zip
    archives_folder = f"{working_dir}/my_archives/"
    Path(archives_folder).mkdir()

    result = run_command("sketch archive ./sketch_simple ./my_archives --include-build-dir")
    assert result.ok

    archive = zipfile.ZipFile(f"{working_dir}/my_archives/sketch_simple.zip")
    archive_files = archive.namelist()

    verify_zip_contains_sketch_including_build_dir(archive_files)

    archive.close()


def test_sketch_archive_relative_sketch_path_with_absolute_zip_path_with_include_build_dir_flag(
    run_command, copy_sketch, working_dir
):
    copy_sketch("sketch_simple")
    # Creates a folder where to save the zip
    archives_folder = f"{working_dir}/my_archives/"
    Path(archives_folder).mkdir()

    result = run_command(f'sketch archive ./sketch_simple "{archives_folder}" --include-build-dir')
    assert result.ok

    archive = zipfile.ZipFile(f"{archives_folder}/sketch_simple.zip")
    archive_files = archive.namelist()

    verify_zip_contains_sketch_including_build_dir(archive_files)

    archive.close()


def test_sketch_archive_relative_sketch_path_with_relative_zip_path_and_name_without_extension_with_include_build_dir_flag(  # noqa
    run_command, copy_sketch, working_dir
):
    copy_sketch("sketch_simple")
    # Creates a folder where to save the zip
    archives_folder = f"{working_dir}/my_archives/"
    Path(archives_folder).mkdir()

    result = run_command("sketch archive ./sketch_simple ./my_archives/my_custom_sketch --include-build-dir")
    assert result.ok

    archive = zipfile.ZipFile(f"{working_dir}/my_archives/my_custom_sketch.zip")
    archive_files = archive.namelist()

    verify_zip_contains_sketch_including_build_dir(archive_files)

    archive.close()


def test_sketch_archive_relative_sketch_path_with_relative_zip_path_and_name_with_extension_with_include_build_dir_flag(
    run_command, copy_sketch, working_dir
):
    copy_sketch("sketch_simple")
    # Creates a folder where to save the zip
    archives_folder = f"{working_dir}/my_archives/"
    Path(archives_folder).mkdir()

    result = run_command("sketch archive ./sketch_simple ./my_archives/my_custom_sketch.zip --include-build-dir")
    assert result.ok

    archive = zipfile.ZipFile(f"{working_dir}/my_archives/my_custom_sketch.zip")
    archive_files = archive.namelist()

    verify_zip_contains_sketch_including_build_dir(archive_files)

    archive.close()


def test_sketch_archive_relative_sketch_path_with_absolute_zip_path_and_name_without_extension_with_include_build_dir_flag(  # noqa
    run_command, copy_sketch, working_dir
):
    copy_sketch("sketch_simple")
    # Creates a folder where to save the zip
    archives_folder = f"{working_dir}/my_archives/"
    Path(archives_folder).mkdir()

    result = run_command(f'sketch archive ./sketch_simple "{archives_folder}/my_custom_sketch" --include-build-dir')
    assert result.ok

    archive = zipfile.ZipFile(f"{archives_folder}/my_custom_sketch.zip")
    archive_files = archive.namelist()

    verify_zip_contains_sketch_including_build_dir(archive_files)

    archive.close()


def test_sketch_archive_relative_sketch_path_with_absolute_zip_path_and_name_with_extension_with_include_build_dir_flag(
    run_command, copy_sketch, working_dir
):
    copy_sketch("sketch_simple")
    # Creates a folder where to save the zip
    archives_folder = f"{working_dir}/my_archives/"
    Path(archives_folder).mkdir()

    result = run_command(f'sketch archive ./sketch_simple "{archives_folder}/my_custom_sketch.zip" --include-build-dir')
    assert result.ok

    archive = zipfile.ZipFile(f"{archives_folder}/my_custom_sketch.zip")
    archive_files = archive.namelist()

    verify_zip_contains_sketch_including_build_dir(archive_files)

    archive.close()


def test_sketch_archive_absolute_sketch_path_with_relative_zip_path_with_include_build_dir_flag(
    run_command, copy_sketch, working_dir
):
    copy_sketch("sketch_simple")
    # Creates a folder where to save the zip
    archives_folder = f"{working_dir}/my_archives/"
    Path(archives_folder).mkdir()

    result = run_command(f'sketch archive "{working_dir}/sketch_simple" ./my_archives --include-build-dir')
    assert result.ok

    archive = zipfile.ZipFile(f"{working_dir}/my_archives/sketch_simple.zip")
    archive_files = archive.namelist()

    verify_zip_contains_sketch_including_build_dir(archive_files)

    archive.close()


def test_sketch_archive_absolute_sketch_path_with_absolute_zip_path_with_include_build_dir_flag(
    run_command, copy_sketch, working_dir
):
    # Creates a folder where to save the zip
    archives_folder = f"{working_dir}/my_archives/"
    Path(archives_folder).mkdir()

    result = run_command(
        f'sketch archive "{working_dir}/sketch_simple" "{archives_folder}" --include-build-dir',
        copy_sketch("sketch_simple"),
    )
    assert result.ok

    archive = zipfile.ZipFile(f"{archives_folder}/sketch_simple.zip")
    archive_files = archive.namelist()

    verify_zip_contains_sketch_including_build_dir(archive_files)

    archive.close()


def test_sketch_archive_absolute_sketch_path_with_relative_zip_path_and_name_without_extension_with_include_build_dir_flag(  # noqa
    run_command, copy_sketch, working_dir
):
    copy_sketch("sketch_simple")
    # Creates a folder where to save the zip
    archives_folder = f"{working_dir}/my_archives/"
    Path(archives_folder).mkdir()

    result = run_command(
        f'sketch archive "{working_dir}/sketch_simple" ./my_archives/my_custom_sketch --include-build-dir'
    )
    assert result.ok

    archive = zipfile.ZipFile(f"{working_dir}/my_archives/my_custom_sketch.zip")
    archive_files = archive.namelist()

    verify_zip_contains_sketch_including_build_dir(archive_files)

    archive.close()


def test_sketch_archive_absolute_sketch_path_with_relative_zip_path_and_name_with_extension_with_include_build_dir_flag(
    run_command, copy_sketch, working_dir
):
    copy_sketch("sketch_simple")
    # Creates a folder where to save the zip
    archives_folder = f"{working_dir}/my_archives/"
    Path(archives_folder).mkdir()

    result = run_command(
        f'sketch archive "{working_dir}/sketch_simple" ./my_archives/my_custom_sketch.zip --include-build-dir'
    )
    assert result.ok

    archive = zipfile.ZipFile(f"{working_dir}/my_archives/my_custom_sketch.zip")
    archive_files = archive.namelist()

    verify_zip_contains_sketch_including_build_dir(archive_files)

    archive.close()


def test_sketch_archive_absolute_sketch_path_with_absolute_zip_path_and_name_without_extension_with_include_build_dir_flag(  # noqa
    run_command, copy_sketch, working_dir
):
    # Creates a folder where to save the zip
    archives_folder = f"{working_dir}/my_archives/"
    Path(archives_folder).mkdir()

    result = run_command(
        f'sketch archive "{working_dir}/sketch_simple" "{archives_folder}/my_custom_sketch" --include-build-dir',
        copy_sketch("sketch_simple"),
    )
    assert result.ok

    archive = zipfile.ZipFile(f"{archives_folder}/my_custom_sketch.zip")
    archive_files = archive.namelist()

    verify_zip_contains_sketch_including_build_dir(archive_files)

    archive.close()


def test_sketch_archive_absolute_sketch_path_with_absolute_zip_path_and_name_with_extension_with_include_build_dir_flag(
    run_command, copy_sketch, working_dir
):
    # Creates a folder where to save the zip
    archives_folder = f"{working_dir}/my_archives/"
    Path(archives_folder).mkdir()

    result = run_command(
        f'sketch archive "{working_dir}/sketch_simple" "{archives_folder}/my_custom_sketch.zip" --include-build-dir',
        copy_sketch("sketch_simple"),
    )
    assert result.ok

    archive = zipfile.ZipFile(f"{archives_folder}/my_custom_sketch.zip")
    archive_files = archive.namelist()

    verify_zip_contains_sketch_including_build_dir(archive_files)

    archive.close()
