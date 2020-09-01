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
import zipfile
import shutil
from pathlib import Path

import pytest


@pytest.fixture(scope="function")
def copy_sketch(working_dir):
    # Copies sketch for testing
    sketch_path = Path(__file__).parent / "testdata" / "sketch_simple"
    test_sketch_path = Path(working_dir) / "sketch_simple"
    shutil.copytree(sketch_path, test_sketch_path)
    yield str(test_sketch_path)


def test_archive_no_args(run_command, copy_sketch, working_dir):
    result = run_command("archive", copy_sketch)
    assert result.ok

    archive = zipfile.ZipFile(f"{working_dir}/sketch_simple.zip")
    archive_files = archive.namelist()

    assert "sketch_simple/doc.txt" in archive_files
    assert "sketch_simple/header.h" in archive_files
    assert "sketch_simple/merged_sketch.txt" in archive_files
    assert "sketch_simple/old.pde" in archive_files
    assert "sketch_simple/other.ino" in archive_files
    assert "sketch_simple/s_file.S" in archive_files
    assert "sketch_simple/sketch_simple.ino" in archive_files
    assert "sketch_simple/src/helper.h" in archive_files
    assert "sketch_simple/build/adafruit.samd.adafruit_feather_m0/sketch_simple.ino.hex" not in archive_files
    assert "sketch_simple/build/adafruit.samd.adafruit_feather_m0/sketch_simple.ino.map" not in archive_files
    assert "sketch_simple/build/arduino.avr.uno/sketch_simple.ino.eep" not in archive_files
    assert "sketch_simple/build/arduino.avr.uno/sketch_simple.ino.hex" not in archive_files
    assert "sketch_simple/build/arduino.avr.uno/sketch_simple.ino.with_bootloader.hex" not in archive_files

    archive.close()


def test_archive_dot_arg(run_command, copy_sketch, working_dir):
    result = run_command("archive .", copy_sketch)
    assert result.ok

    archive = zipfile.ZipFile(f"{working_dir}/sketch_simple.zip")
    archive_files = archive.namelist()

    assert "sketch_simple/doc.txt" in archive_files
    assert "sketch_simple/header.h" in archive_files
    assert "sketch_simple/merged_sketch.txt" in archive_files
    assert "sketch_simple/old.pde" in archive_files
    assert "sketch_simple/other.ino" in archive_files
    assert "sketch_simple/s_file.S" in archive_files
    assert "sketch_simple/sketch_simple.ino" in archive_files
    assert "sketch_simple/src/helper.h" in archive_files
    assert "sketch_simple/build/adafruit.samd.adafruit_feather_m0/sketch_simple.ino.hex" not in archive_files
    assert "sketch_simple/build/adafruit.samd.adafruit_feather_m0/sketch_simple.ino.map" not in archive_files
    assert "sketch_simple/build/arduino.avr.uno/sketch_simple.ino.eep" not in archive_files
    assert "sketch_simple/build/arduino.avr.uno/sketch_simple.ino.hex" not in archive_files
    assert "sketch_simple/build/arduino.avr.uno/sketch_simple.ino.with_bootloader.hex" not in archive_files

    archive.close()


def test_archive_dot_arg_relative_zip_path(run_command, copy_sketch, working_dir):
    # Creates a folder where to save the zip
    archives_folder = f"{working_dir}/my_archives/"
    Path(archives_folder).mkdir()

    result = run_command("archive . ../my_archives", copy_sketch)
    assert result.ok

    archive = zipfile.ZipFile(f"{working_dir}/my_archives/sketch_simple.zip")
    archive_files = archive.namelist()

    assert "sketch_simple/doc.txt" in archive_files
    assert "sketch_simple/header.h" in archive_files
    assert "sketch_simple/merged_sketch.txt" in archive_files
    assert "sketch_simple/old.pde" in archive_files
    assert "sketch_simple/other.ino" in archive_files
    assert "sketch_simple/s_file.S" in archive_files
    assert "sketch_simple/sketch_simple.ino" in archive_files
    assert "sketch_simple/src/helper.h" in archive_files
    assert "sketch_simple/build/adafruit.samd.adafruit_feather_m0/sketch_simple.ino.hex" not in archive_files
    assert "sketch_simple/build/adafruit.samd.adafruit_feather_m0/sketch_simple.ino.map" not in archive_files
    assert "sketch_simple/build/arduino.avr.uno/sketch_simple.ino.eep" not in archive_files
    assert "sketch_simple/build/arduino.avr.uno/sketch_simple.ino.hex" not in archive_files
    assert "sketch_simple/build/arduino.avr.uno/sketch_simple.ino.with_bootloader.hex" not in archive_files

    archive.close()


def test_archive_dot_arg_absolute_zip_path(run_command, copy_sketch, working_dir):
    # Creates a folder where to save the zip
    archives_folder = f"{working_dir}/my_archives/"
    Path(archives_folder).mkdir()

    result = run_command(f"archive . {archives_folder}", copy_sketch)
    assert result.ok

    archive = zipfile.ZipFile(f"{archives_folder}/sketch_simple.zip")
    archive_files = archive.namelist()

    assert "sketch_simple/doc.txt" in archive_files
    assert "sketch_simple/header.h" in archive_files
    assert "sketch_simple/merged_sketch.txt" in archive_files
    assert "sketch_simple/old.pde" in archive_files
    assert "sketch_simple/other.ino" in archive_files
    assert "sketch_simple/s_file.S" in archive_files
    assert "sketch_simple/sketch_simple.ino" in archive_files
    assert "sketch_simple/src/helper.h" in archive_files
    assert "sketch_simple/build/adafruit.samd.adafruit_feather_m0/sketch_simple.ino.hex" not in archive_files
    assert "sketch_simple/build/adafruit.samd.adafruit_feather_m0/sketch_simple.ino.map" not in archive_files
    assert "sketch_simple/build/arduino.avr.uno/sketch_simple.ino.eep" not in archive_files
    assert "sketch_simple/build/arduino.avr.uno/sketch_simple.ino.hex" not in archive_files
    assert "sketch_simple/build/arduino.avr.uno/sketch_simple.ino.with_bootloader.hex" not in archive_files

    archive.close()


def test_archive_dot_arg_relative_zip_path_and_name_without_extension(run_command, copy_sketch, working_dir):
    # Creates a folder where to save the zip
    archives_folder = f"{working_dir}/my_archives/"
    Path(archives_folder).mkdir()

    result = run_command("archive . ../my_archives/my_custom_sketch", copy_sketch)
    assert result.ok

    archive = zipfile.ZipFile(f"{working_dir}/my_archives/my_custom_sketch.zip")
    archive_files = archive.namelist()

    assert "sketch_simple/doc.txt" in archive_files
    assert "sketch_simple/header.h" in archive_files
    assert "sketch_simple/merged_sketch.txt" in archive_files
    assert "sketch_simple/old.pde" in archive_files
    assert "sketch_simple/other.ino" in archive_files
    assert "sketch_simple/s_file.S" in archive_files
    assert "sketch_simple/sketch_simple.ino" in archive_files
    assert "sketch_simple/src/helper.h" in archive_files
    assert "sketch_simple/build/adafruit.samd.adafruit_feather_m0/sketch_simple.ino.hex" not in archive_files
    assert "sketch_simple/build/adafruit.samd.adafruit_feather_m0/sketch_simple.ino.map" not in archive_files
    assert "sketch_simple/build/arduino.avr.uno/sketch_simple.ino.eep" not in archive_files
    assert "sketch_simple/build/arduino.avr.uno/sketch_simple.ino.hex" not in archive_files
    assert "sketch_simple/build/arduino.avr.uno/sketch_simple.ino.with_bootloader.hex" not in archive_files

    archive.close()


def test_archive_dot_arg_absolute_zip_path_and_name_without_extension(run_command, copy_sketch, working_dir):
    # Creates a folder where to save the zip
    archives_folder = f"{working_dir}/my_archives/"
    Path(archives_folder).mkdir()

    result = run_command(f"archive . {archives_folder}/my_custom_sketch", copy_sketch)
    assert result.ok

    archive = zipfile.ZipFile(f"{archives_folder}/my_custom_sketch.zip")
    archive_files = archive.namelist()

    assert "sketch_simple/doc.txt" in archive_files
    assert "sketch_simple/header.h" in archive_files
    assert "sketch_simple/merged_sketch.txt" in archive_files
    assert "sketch_simple/old.pde" in archive_files
    assert "sketch_simple/other.ino" in archive_files
    assert "sketch_simple/s_file.S" in archive_files
    assert "sketch_simple/sketch_simple.ino" in archive_files
    assert "sketch_simple/src/helper.h" in archive_files
    assert "sketch_simple/build/adafruit.samd.adafruit_feather_m0/sketch_simple.ino.hex" not in archive_files
    assert "sketch_simple/build/adafruit.samd.adafruit_feather_m0/sketch_simple.ino.map" not in archive_files
    assert "sketch_simple/build/arduino.avr.uno/sketch_simple.ino.eep" not in archive_files
    assert "sketch_simple/build/arduino.avr.uno/sketch_simple.ino.hex" not in archive_files
    assert "sketch_simple/build/arduino.avr.uno/sketch_simple.ino.with_bootloader.hex" not in archive_files

    archive.close()


def test_archive_dot_arg_custom_zip_path_and_name_with_extension(run_command, copy_sketch, working_dir):
    # Creates a folder where to save the zip
    archives_folder = f"{working_dir}/my_archives/"
    Path(archives_folder).mkdir()

    result = run_command(f"archive . {archives_folder}/my_custom_sketch.zip", copy_sketch)
    assert result.ok

    archive = zipfile.ZipFile(f"{archives_folder}/my_custom_sketch.zip")
    archive_files = archive.namelist()

    assert "sketch_simple/doc.txt" in archive_files
    assert "sketch_simple/header.h" in archive_files
    assert "sketch_simple/merged_sketch.txt" in archive_files
    assert "sketch_simple/old.pde" in archive_files
    assert "sketch_simple/other.ino" in archive_files
    assert "sketch_simple/s_file.S" in archive_files
    assert "sketch_simple/sketch_simple.ino" in archive_files
    assert "sketch_simple/src/helper.h" in archive_files
    assert "sketch_simple/build/adafruit.samd.adafruit_feather_m0/sketch_simple.ino.hex" not in archive_files
    assert "sketch_simple/build/adafruit.samd.adafruit_feather_m0/sketch_simple.ino.map" not in archive_files
    assert "sketch_simple/build/arduino.avr.uno/sketch_simple.ino.eep" not in archive_files
    assert "sketch_simple/build/arduino.avr.uno/sketch_simple.ino.hex" not in archive_files
    assert "sketch_simple/build/arduino.avr.uno/sketch_simple.ino.with_bootloader.hex" not in archive_files

    archive.close()


def test_archive_relative_sketch_path(run_command, copy_sketch, working_dir):
    result = run_command("archive ./sketch_simple")
    assert result.ok

    archive = zipfile.ZipFile(f"{working_dir}/sketch_simple.zip")
    archive_files = archive.namelist()

    assert "sketch_simple/doc.txt" in archive_files
    assert "sketch_simple/header.h" in archive_files
    assert "sketch_simple/merged_sketch.txt" in archive_files
    assert "sketch_simple/old.pde" in archive_files
    assert "sketch_simple/other.ino" in archive_files
    assert "sketch_simple/s_file.S" in archive_files
    assert "sketch_simple/sketch_simple.ino" in archive_files
    assert "sketch_simple/src/helper.h" in archive_files
    assert "sketch_simple/build/adafruit.samd.adafruit_feather_m0/sketch_simple.ino.hex" not in archive_files
    assert "sketch_simple/build/adafruit.samd.adafruit_feather_m0/sketch_simple.ino.map" not in archive_files
    assert "sketch_simple/build/arduino.avr.uno/sketch_simple.ino.eep" not in archive_files
    assert "sketch_simple/build/arduino.avr.uno/sketch_simple.ino.hex" not in archive_files
    assert "sketch_simple/build/arduino.avr.uno/sketch_simple.ino.with_bootloader.hex" not in archive_files

    archive.close()


def test_archive_absolute_sketch_path(run_command, copy_sketch, working_dir):
    result = run_command(f"archive {working_dir}/sketch_simple", copy_sketch)
    assert result.ok

    archive = zipfile.ZipFile(f"{working_dir}/sketch_simple.zip")
    archive_files = archive.namelist()

    assert "sketch_simple/doc.txt" in archive_files
    assert "sketch_simple/header.h" in archive_files
    assert "sketch_simple/merged_sketch.txt" in archive_files
    assert "sketch_simple/old.pde" in archive_files
    assert "sketch_simple/other.ino" in archive_files
    assert "sketch_simple/s_file.S" in archive_files
    assert "sketch_simple/sketch_simple.ino" in archive_files
    assert "sketch_simple/src/helper.h" in archive_files
    assert "sketch_simple/build/adafruit.samd.adafruit_feather_m0/sketch_simple.ino.hex" not in archive_files
    assert "sketch_simple/build/adafruit.samd.adafruit_feather_m0/sketch_simple.ino.map" not in archive_files
    assert "sketch_simple/build/arduino.avr.uno/sketch_simple.ino.eep" not in archive_files
    assert "sketch_simple/build/arduino.avr.uno/sketch_simple.ino.hex" not in archive_files
    assert "sketch_simple/build/arduino.avr.uno/sketch_simple.ino.with_bootloader.hex" not in archive_files

    archive.close()


def test_archive_relative_sketch_path_with_relative_zip_path(run_command, copy_sketch, working_dir):
    # Creates a folder where to save the zip
    archives_folder = f"{working_dir}/my_archives/"
    Path(archives_folder).mkdir()

    result = run_command("archive ./sketch_simple ./my_archives")
    assert result.ok

    archive = zipfile.ZipFile(f"{working_dir}/my_archives/sketch_simple.zip")
    archive_files = archive.namelist()

    assert "sketch_simple/doc.txt" in archive_files
    assert "sketch_simple/header.h" in archive_files
    assert "sketch_simple/merged_sketch.txt" in archive_files
    assert "sketch_simple/old.pde" in archive_files
    assert "sketch_simple/other.ino" in archive_files
    assert "sketch_simple/s_file.S" in archive_files
    assert "sketch_simple/sketch_simple.ino" in archive_files
    assert "sketch_simple/src/helper.h" in archive_files
    assert "sketch_simple/build/adafruit.samd.adafruit_feather_m0/sketch_simple.ino.hex" not in archive_files
    assert "sketch_simple/build/adafruit.samd.adafruit_feather_m0/sketch_simple.ino.map" not in archive_files
    assert "sketch_simple/build/arduino.avr.uno/sketch_simple.ino.eep" not in archive_files
    assert "sketch_simple/build/arduino.avr.uno/sketch_simple.ino.hex" not in archive_files
    assert "sketch_simple/build/arduino.avr.uno/sketch_simple.ino.with_bootloader.hex" not in archive_files

    archive.close()


def test_archive_relative_sketch_path_with_absolute_zip_path(run_command, copy_sketch, working_dir):
    # Creates a folder where to save the zip
    archives_folder = f"{working_dir}/my_archives/"
    Path(archives_folder).mkdir()

    result = run_command(f"archive ./sketch_simple {archives_folder}")
    assert result.ok

    archive = zipfile.ZipFile(f"{archives_folder}/sketch_simple.zip")
    archive_files = archive.namelist()

    assert "sketch_simple/doc.txt" in archive_files
    assert "sketch_simple/header.h" in archive_files
    assert "sketch_simple/merged_sketch.txt" in archive_files
    assert "sketch_simple/old.pde" in archive_files
    assert "sketch_simple/other.ino" in archive_files
    assert "sketch_simple/s_file.S" in archive_files
    assert "sketch_simple/sketch_simple.ino" in archive_files
    assert "sketch_simple/src/helper.h" in archive_files
    assert "sketch_simple/build/adafruit.samd.adafruit_feather_m0/sketch_simple.ino.hex" not in archive_files
    assert "sketch_simple/build/adafruit.samd.adafruit_feather_m0/sketch_simple.ino.map" not in archive_files
    assert "sketch_simple/build/arduino.avr.uno/sketch_simple.ino.eep" not in archive_files
    assert "sketch_simple/build/arduino.avr.uno/sketch_simple.ino.hex" not in archive_files
    assert "sketch_simple/build/arduino.avr.uno/sketch_simple.ino.with_bootloader.hex" not in archive_files

    archive.close()


def test_archive_relative_sketch_path_with_relative_zip_path_and_name_without_extension(
    run_command, copy_sketch, working_dir
):
    # Creates a folder where to save the zip
    archives_folder = f"{working_dir}/my_archives/"
    Path(archives_folder).mkdir()

    result = run_command("archive ./sketch_simple ./my_archives/my_custom_sketch")
    assert result.ok

    archive = zipfile.ZipFile(f"{working_dir}/my_archives/my_custom_sketch.zip")
    archive_files = archive.namelist()

    assert "sketch_simple/doc.txt" in archive_files
    assert "sketch_simple/header.h" in archive_files
    assert "sketch_simple/merged_sketch.txt" in archive_files
    assert "sketch_simple/old.pde" in archive_files
    assert "sketch_simple/other.ino" in archive_files
    assert "sketch_simple/s_file.S" in archive_files
    assert "sketch_simple/sketch_simple.ino" in archive_files
    assert "sketch_simple/src/helper.h" in archive_files
    assert "sketch_simple/build/adafruit.samd.adafruit_feather_m0/sketch_simple.ino.hex" not in archive_files
    assert "sketch_simple/build/adafruit.samd.adafruit_feather_m0/sketch_simple.ino.map" not in archive_files
    assert "sketch_simple/build/arduino.avr.uno/sketch_simple.ino.eep" not in archive_files
    assert "sketch_simple/build/arduino.avr.uno/sketch_simple.ino.hex" not in archive_files
    assert "sketch_simple/build/arduino.avr.uno/sketch_simple.ino.with_bootloader.hex" not in archive_files

    archive.close()


def test_archive_relative_sketch_path_with_relative_zip_path_and_name_with_extension(
    run_command, copy_sketch, working_dir
):
    # Creates a folder where to save the zip
    archives_folder = f"{working_dir}/my_archives/"
    Path(archives_folder).mkdir()

    result = run_command("archive ./sketch_simple ./my_archives/my_custom_sketch.zip")
    assert result.ok

    archive = zipfile.ZipFile(f"{working_dir}/my_archives/my_custom_sketch.zip")
    archive_files = archive.namelist()

    assert "sketch_simple/doc.txt" in archive_files
    assert "sketch_simple/header.h" in archive_files
    assert "sketch_simple/merged_sketch.txt" in archive_files
    assert "sketch_simple/old.pde" in archive_files
    assert "sketch_simple/other.ino" in archive_files
    assert "sketch_simple/s_file.S" in archive_files
    assert "sketch_simple/sketch_simple.ino" in archive_files
    assert "sketch_simple/src/helper.h" in archive_files
    assert "sketch_simple/build/adafruit.samd.adafruit_feather_m0/sketch_simple.ino.hex" not in archive_files
    assert "sketch_simple/build/adafruit.samd.adafruit_feather_m0/sketch_simple.ino.map" not in archive_files
    assert "sketch_simple/build/arduino.avr.uno/sketch_simple.ino.eep" not in archive_files
    assert "sketch_simple/build/arduino.avr.uno/sketch_simple.ino.hex" not in archive_files
    assert "sketch_simple/build/arduino.avr.uno/sketch_simple.ino.with_bootloader.hex" not in archive_files

    archive.close()


def test_archive_relative_sketch_path_with_absolute_zip_path_and_name_without_extension(
    run_command, copy_sketch, working_dir
):
    # Creates a folder where to save the zip
    archives_folder = f"{working_dir}/my_archives/"
    Path(archives_folder).mkdir()

    result = run_command(f"archive ./sketch_simple {archives_folder}/my_custom_sketch")
    assert result.ok

    archive = zipfile.ZipFile(f"{archives_folder}/my_custom_sketch.zip")
    archive_files = archive.namelist()

    assert "sketch_simple/doc.txt" in archive_files
    assert "sketch_simple/header.h" in archive_files
    assert "sketch_simple/merged_sketch.txt" in archive_files
    assert "sketch_simple/old.pde" in archive_files
    assert "sketch_simple/other.ino" in archive_files
    assert "sketch_simple/s_file.S" in archive_files
    assert "sketch_simple/sketch_simple.ino" in archive_files
    assert "sketch_simple/src/helper.h" in archive_files
    assert "sketch_simple/build/adafruit.samd.adafruit_feather_m0/sketch_simple.ino.hex" not in archive_files
    assert "sketch_simple/build/adafruit.samd.adafruit_feather_m0/sketch_simple.ino.map" not in archive_files
    assert "sketch_simple/build/arduino.avr.uno/sketch_simple.ino.eep" not in archive_files
    assert "sketch_simple/build/arduino.avr.uno/sketch_simple.ino.hex" not in archive_files
    assert "sketch_simple/build/arduino.avr.uno/sketch_simple.ino.with_bootloader.hex" not in archive_files

    archive.close()


def test_archive_relative_sketch_path_with_absolute_zip_path_and_name_with_extension(
    run_command, copy_sketch, working_dir
):
    # Creates a folder where to save the zip
    archives_folder = f"{working_dir}/my_archives/"
    Path(archives_folder).mkdir()

    result = run_command(f"archive ./sketch_simple {archives_folder}/my_custom_sketch.zip")
    assert result.ok

    archive = zipfile.ZipFile(f"{archives_folder}/my_custom_sketch.zip")
    archive_files = archive.namelist()

    assert "sketch_simple/doc.txt" in archive_files
    assert "sketch_simple/header.h" in archive_files
    assert "sketch_simple/merged_sketch.txt" in archive_files
    assert "sketch_simple/old.pde" in archive_files
    assert "sketch_simple/other.ino" in archive_files
    assert "sketch_simple/s_file.S" in archive_files
    assert "sketch_simple/sketch_simple.ino" in archive_files
    assert "sketch_simple/src/helper.h" in archive_files
    assert "sketch_simple/build/adafruit.samd.adafruit_feather_m0/sketch_simple.ino.hex" not in archive_files
    assert "sketch_simple/build/adafruit.samd.adafruit_feather_m0/sketch_simple.ino.map" not in archive_files
    assert "sketch_simple/build/arduino.avr.uno/sketch_simple.ino.eep" not in archive_files
    assert "sketch_simple/build/arduino.avr.uno/sketch_simple.ino.hex" not in archive_files
    assert "sketch_simple/build/arduino.avr.uno/sketch_simple.ino.with_bootloader.hex" not in archive_files

    archive.close()


def test_archive_absolute_sketch_path_with_relative_zip_path(run_command, copy_sketch, working_dir):
    # Creates a folder where to save the zip
    archives_folder = f"{working_dir}/my_archives/"
    Path(archives_folder).mkdir()

    result = run_command(f"archive {working_dir}/sketch_simple ./my_archives")
    assert result.ok

    archive = zipfile.ZipFile(f"{working_dir}/my_archives/sketch_simple.zip")
    archive_files = archive.namelist()

    assert "sketch_simple/doc.txt" in archive_files
    assert "sketch_simple/header.h" in archive_files
    assert "sketch_simple/merged_sketch.txt" in archive_files
    assert "sketch_simple/old.pde" in archive_files
    assert "sketch_simple/other.ino" in archive_files
    assert "sketch_simple/s_file.S" in archive_files
    assert "sketch_simple/sketch_simple.ino" in archive_files
    assert "sketch_simple/src/helper.h" in archive_files
    assert "sketch_simple/build/adafruit.samd.adafruit_feather_m0/sketch_simple.ino.hex" not in archive_files
    assert "sketch_simple/build/adafruit.samd.adafruit_feather_m0/sketch_simple.ino.map" not in archive_files
    assert "sketch_simple/build/arduino.avr.uno/sketch_simple.ino.eep" not in archive_files
    assert "sketch_simple/build/arduino.avr.uno/sketch_simple.ino.hex" not in archive_files
    assert "sketch_simple/build/arduino.avr.uno/sketch_simple.ino.with_bootloader.hex" not in archive_files

    archive.close()


def test_archive_absolute_sketch_path_with_absolute_zip_path(run_command, copy_sketch, working_dir):
    # Creates a folder where to save the zip
    archives_folder = f"{working_dir}/my_archives/"
    Path(archives_folder).mkdir()

    result = run_command(f"archive {working_dir}/sketch_simple {archives_folder}", copy_sketch)
    assert result.ok

    archive = zipfile.ZipFile(f"{archives_folder}/sketch_simple.zip")
    archive_files = archive.namelist()

    assert "sketch_simple/doc.txt" in archive_files
    assert "sketch_simple/header.h" in archive_files
    assert "sketch_simple/merged_sketch.txt" in archive_files
    assert "sketch_simple/old.pde" in archive_files
    assert "sketch_simple/other.ino" in archive_files
    assert "sketch_simple/s_file.S" in archive_files
    assert "sketch_simple/sketch_simple.ino" in archive_files
    assert "sketch_simple/src/helper.h" in archive_files
    assert "sketch_simple/build/adafruit.samd.adafruit_feather_m0/sketch_simple.ino.hex" not in archive_files
    assert "sketch_simple/build/adafruit.samd.adafruit_feather_m0/sketch_simple.ino.map" not in archive_files
    assert "sketch_simple/build/arduino.avr.uno/sketch_simple.ino.eep" not in archive_files
    assert "sketch_simple/build/arduino.avr.uno/sketch_simple.ino.hex" not in archive_files
    assert "sketch_simple/build/arduino.avr.uno/sketch_simple.ino.with_bootloader.hex" not in archive_files

    archive.close()


def test_archive_absolute_sketch_path_with_relative_zip_path_and_name_without_extension(
    run_command, copy_sketch, working_dir
):
    # Creates a folder where to save the zip
    archives_folder = f"{working_dir}/my_archives/"
    Path(archives_folder).mkdir()

    result = run_command(f"archive {working_dir}/sketch_simple ./my_archives/my_custom_sketch")
    assert result.ok

    archive = zipfile.ZipFile(f"{working_dir}/my_archives/my_custom_sketch.zip")
    archive_files = archive.namelist()

    assert "sketch_simple/doc.txt" in archive_files
    assert "sketch_simple/header.h" in archive_files
    assert "sketch_simple/merged_sketch.txt" in archive_files
    assert "sketch_simple/old.pde" in archive_files
    assert "sketch_simple/other.ino" in archive_files
    assert "sketch_simple/s_file.S" in archive_files
    assert "sketch_simple/sketch_simple.ino" in archive_files
    assert "sketch_simple/src/helper.h" in archive_files
    assert "sketch_simple/build/adafruit.samd.adafruit_feather_m0/sketch_simple.ino.hex" not in archive_files
    assert "sketch_simple/build/adafruit.samd.adafruit_feather_m0/sketch_simple.ino.map" not in archive_files
    assert "sketch_simple/build/arduino.avr.uno/sketch_simple.ino.eep" not in archive_files
    assert "sketch_simple/build/arduino.avr.uno/sketch_simple.ino.hex" not in archive_files
    assert "sketch_simple/build/arduino.avr.uno/sketch_simple.ino.with_bootloader.hex" not in archive_files

    archive.close()


def test_archive_absolute_sketch_path_with_relative_zip_path_and_name_with_extension(
    run_command, copy_sketch, working_dir
):
    # Creates a folder where to save the zip
    archives_folder = f"{working_dir}/my_archives/"
    Path(archives_folder).mkdir()

    result = run_command(f"archive {working_dir}/sketch_simple ./my_archives/my_custom_sketch.zip")
    assert result.ok

    archive = zipfile.ZipFile(f"{working_dir}/my_archives/my_custom_sketch.zip")
    archive_files = archive.namelist()

    assert "sketch_simple/doc.txt" in archive_files
    assert "sketch_simple/header.h" in archive_files
    assert "sketch_simple/merged_sketch.txt" in archive_files
    assert "sketch_simple/old.pde" in archive_files
    assert "sketch_simple/other.ino" in archive_files
    assert "sketch_simple/s_file.S" in archive_files
    assert "sketch_simple/sketch_simple.ino" in archive_files
    assert "sketch_simple/src/helper.h" in archive_files
    assert "sketch_simple/build/adafruit.samd.adafruit_feather_m0/sketch_simple.ino.hex" not in archive_files
    assert "sketch_simple/build/adafruit.samd.adafruit_feather_m0/sketch_simple.ino.map" not in archive_files
    assert "sketch_simple/build/arduino.avr.uno/sketch_simple.ino.eep" not in archive_files
    assert "sketch_simple/build/arduino.avr.uno/sketch_simple.ino.hex" not in archive_files
    assert "sketch_simple/build/arduino.avr.uno/sketch_simple.ino.with_bootloader.hex" not in archive_files

    archive.close()


def test_archive_absolute_sketch_path_with_absolute_zip_path_and_name_without_extension(
    run_command, copy_sketch, working_dir
):
    # Creates a folder where to save the zip
    archives_folder = f"{working_dir}/my_archives/"
    Path(archives_folder).mkdir()

    result = run_command(f"archive {working_dir}/sketch_simple {archives_folder}/my_custom_sketch", copy_sketch)
    assert result.ok

    archive = zipfile.ZipFile(f"{archives_folder}/my_custom_sketch.zip")
    archive_files = archive.namelist()

    assert "sketch_simple/doc.txt" in archive_files
    assert "sketch_simple/header.h" in archive_files
    assert "sketch_simple/merged_sketch.txt" in archive_files
    assert "sketch_simple/old.pde" in archive_files
    assert "sketch_simple/other.ino" in archive_files
    assert "sketch_simple/s_file.S" in archive_files
    assert "sketch_simple/sketch_simple.ino" in archive_files
    assert "sketch_simple/src/helper.h" in archive_files
    assert "sketch_simple/build/adafruit.samd.adafruit_feather_m0/sketch_simple.ino.hex" not in archive_files
    assert "sketch_simple/build/adafruit.samd.adafruit_feather_m0/sketch_simple.ino.map" not in archive_files
    assert "sketch_simple/build/arduino.avr.uno/sketch_simple.ino.eep" not in archive_files
    assert "sketch_simple/build/arduino.avr.uno/sketch_simple.ino.hex" not in archive_files
    assert "sketch_simple/build/arduino.avr.uno/sketch_simple.ino.with_bootloader.hex" not in archive_files

    archive.close()


def test_archive_absolute_sketch_path_with_absolute_zip_path_and_name_with_extension(
    run_command, copy_sketch, working_dir
):
    # Creates a folder where to save the zip
    archives_folder = f"{working_dir}/my_archives/"
    Path(archives_folder).mkdir()

    result = run_command(f"archive {working_dir}/sketch_simple {archives_folder}/my_custom_sketch.zip", copy_sketch)
    assert result.ok

    archive = zipfile.ZipFile(f"{archives_folder}/my_custom_sketch.zip")
    archive_files = archive.namelist()

    assert "sketch_simple/doc.txt" in archive_files
    assert "sketch_simple/header.h" in archive_files
    assert "sketch_simple/merged_sketch.txt" in archive_files
    assert "sketch_simple/old.pde" in archive_files
    assert "sketch_simple/other.ino" in archive_files
    assert "sketch_simple/s_file.S" in archive_files
    assert "sketch_simple/sketch_simple.ino" in archive_files
    assert "sketch_simple/src/helper.h" in archive_files
    assert "sketch_simple/build/adafruit.samd.adafruit_feather_m0/sketch_simple.ino.hex" not in archive_files
    assert "sketch_simple/build/adafruit.samd.adafruit_feather_m0/sketch_simple.ino.map" not in archive_files
    assert "sketch_simple/build/arduino.avr.uno/sketch_simple.ino.eep" not in archive_files
    assert "sketch_simple/build/arduino.avr.uno/sketch_simple.ino.hex" not in archive_files
    assert "sketch_simple/build/arduino.avr.uno/sketch_simple.ino.with_bootloader.hex" not in archive_files

    archive.close()


def test_archive_no_args_with_include_build_dir_flag(run_command, copy_sketch, working_dir):
    result = run_command("archive --include-build-dir", copy_sketch)
    assert result.ok

    archive = zipfile.ZipFile(f"{working_dir}/sketch_simple.zip")
    archive_files = archive.namelist()

    assert "sketch_simple/doc.txt" in archive_files
    assert "sketch_simple/header.h" in archive_files
    assert "sketch_simple/merged_sketch.txt" in archive_files
    assert "sketch_simple/old.pde" in archive_files
    assert "sketch_simple/other.ino" in archive_files
    assert "sketch_simple/s_file.S" in archive_files
    assert "sketch_simple/sketch_simple.ino" in archive_files
    assert "sketch_simple/src/helper.h" in archive_files
    assert "sketch_simple/build/adafruit.samd.adafruit_feather_m0/sketch_simple.ino.hex" in archive_files
    assert "sketch_simple/build/adafruit.samd.adafruit_feather_m0/sketch_simple.ino.map" in archive_files
    assert "sketch_simple/build/arduino.avr.uno/sketch_simple.ino.eep" in archive_files
    assert "sketch_simple/build/arduino.avr.uno/sketch_simple.ino.hex" in archive_files
    assert "sketch_simple/build/arduino.avr.uno/sketch_simple.ino.with_bootloader.hex" in archive_files

    archive.close()


def test_archive_dot_arg_with_include_build_dir_flag(run_command, copy_sketch, working_dir):
    result = run_command("archive . --include-build-dir", copy_sketch)
    assert result.ok

    archive = zipfile.ZipFile(f"{working_dir}/sketch_simple.zip")
    archive_files = archive.namelist()

    assert "sketch_simple/doc.txt" in archive_files
    assert "sketch_simple/header.h" in archive_files
    assert "sketch_simple/merged_sketch.txt" in archive_files
    assert "sketch_simple/old.pde" in archive_files
    assert "sketch_simple/other.ino" in archive_files
    assert "sketch_simple/s_file.S" in archive_files
    assert "sketch_simple/sketch_simple.ino" in archive_files
    assert "sketch_simple/src/helper.h" in archive_files
    assert "sketch_simple/build/adafruit.samd.adafruit_feather_m0/sketch_simple.ino.hex" in archive_files
    assert "sketch_simple/build/adafruit.samd.adafruit_feather_m0/sketch_simple.ino.map" in archive_files
    assert "sketch_simple/build/arduino.avr.uno/sketch_simple.ino.eep" in archive_files
    assert "sketch_simple/build/arduino.avr.uno/sketch_simple.ino.hex" in archive_files
    assert "sketch_simple/build/arduino.avr.uno/sketch_simple.ino.with_bootloader.hex" in archive_files

    archive.close()


def test_archive_dot_arg_relative_zip_path_with_include_build_dir_flag(run_command, copy_sketch, working_dir):
    # Creates a folder where to save the zip
    archives_folder = f"{working_dir}/my_archives/"
    Path(archives_folder).mkdir()

    result = run_command("archive . ../my_archives --include-build-dir", copy_sketch)
    assert result.ok

    archive = zipfile.ZipFile(f"{working_dir}/my_archives/sketch_simple.zip")
    archive_files = archive.namelist()

    assert "sketch_simple/doc.txt" in archive_files
    assert "sketch_simple/header.h" in archive_files
    assert "sketch_simple/merged_sketch.txt" in archive_files
    assert "sketch_simple/old.pde" in archive_files
    assert "sketch_simple/other.ino" in archive_files
    assert "sketch_simple/s_file.S" in archive_files
    assert "sketch_simple/sketch_simple.ino" in archive_files
    assert "sketch_simple/src/helper.h" in archive_files
    assert "sketch_simple/build/adafruit.samd.adafruit_feather_m0/sketch_simple.ino.hex" in archive_files
    assert "sketch_simple/build/adafruit.samd.adafruit_feather_m0/sketch_simple.ino.map" in archive_files
    assert "sketch_simple/build/arduino.avr.uno/sketch_simple.ino.eep" in archive_files
    assert "sketch_simple/build/arduino.avr.uno/sketch_simple.ino.hex" in archive_files
    assert "sketch_simple/build/arduino.avr.uno/sketch_simple.ino.with_bootloader.hex" in archive_files

    archive.close()


def test_archive_dot_arg_absolute_zip_path_with_include_build_dir_flag(run_command, copy_sketch, working_dir):
    # Creates a folder where to save the zip
    archives_folder = f"{working_dir}/my_archives/"
    Path(archives_folder).mkdir()

    result = run_command(f"archive . {archives_folder} --include-build-dir", copy_sketch)
    assert result.ok

    archive = zipfile.ZipFile(f"{archives_folder}/sketch_simple.zip")
    archive_files = archive.namelist()

    assert "sketch_simple/doc.txt" in archive_files
    assert "sketch_simple/header.h" in archive_files
    assert "sketch_simple/merged_sketch.txt" in archive_files
    assert "sketch_simple/old.pde" in archive_files
    assert "sketch_simple/other.ino" in archive_files
    assert "sketch_simple/s_file.S" in archive_files
    assert "sketch_simple/sketch_simple.ino" in archive_files
    assert "sketch_simple/src/helper.h" in archive_files
    assert "sketch_simple/build/adafruit.samd.adafruit_feather_m0/sketch_simple.ino.hex" in archive_files
    assert "sketch_simple/build/adafruit.samd.adafruit_feather_m0/sketch_simple.ino.map" in archive_files
    assert "sketch_simple/build/arduino.avr.uno/sketch_simple.ino.eep" in archive_files
    assert "sketch_simple/build/arduino.avr.uno/sketch_simple.ino.hex" in archive_files
    assert "sketch_simple/build/arduino.avr.uno/sketch_simple.ino.with_bootloader.hex" in archive_files

    archive.close()


def test_archive_dot_arg_relative_zip_path_and_name_without_extension_with_include_build_dir_flag(
    run_command, copy_sketch, working_dir
):
    # Creates a folder where to save the zip
    archives_folder = f"{working_dir}/my_archives/"
    Path(archives_folder).mkdir()

    result = run_command("archive . ../my_archives/my_custom_sketch --include-build-dir", copy_sketch)
    assert result.ok

    archive = zipfile.ZipFile(f"{working_dir}/my_archives/my_custom_sketch.zip")
    archive_files = archive.namelist()

    assert "sketch_simple/doc.txt" in archive_files
    assert "sketch_simple/header.h" in archive_files
    assert "sketch_simple/merged_sketch.txt" in archive_files
    assert "sketch_simple/old.pde" in archive_files
    assert "sketch_simple/other.ino" in archive_files
    assert "sketch_simple/s_file.S" in archive_files
    assert "sketch_simple/sketch_simple.ino" in archive_files
    assert "sketch_simple/src/helper.h" in archive_files
    assert "sketch_simple/build/adafruit.samd.adafruit_feather_m0/sketch_simple.ino.hex" in archive_files
    assert "sketch_simple/build/adafruit.samd.adafruit_feather_m0/sketch_simple.ino.map" in archive_files
    assert "sketch_simple/build/arduino.avr.uno/sketch_simple.ino.eep" in archive_files
    assert "sketch_simple/build/arduino.avr.uno/sketch_simple.ino.hex" in archive_files
    assert "sketch_simple/build/arduino.avr.uno/sketch_simple.ino.with_bootloader.hex" in archive_files

    archive.close()


def test_archive_dot_arg_absolute_zip_path_and_name_without_extension_with_include_build_dir_flag(
    run_command, copy_sketch, working_dir
):
    # Creates a folder where to save the zip
    archives_folder = f"{working_dir}/my_archives/"
    Path(archives_folder).mkdir()

    result = run_command(f"archive . {archives_folder}/my_custom_sketch --include-build-dir", copy_sketch)
    assert result.ok

    archive = zipfile.ZipFile(f"{archives_folder}/my_custom_sketch.zip")
    archive_files = archive.namelist()

    assert "sketch_simple/doc.txt" in archive_files
    assert "sketch_simple/header.h" in archive_files
    assert "sketch_simple/merged_sketch.txt" in archive_files
    assert "sketch_simple/old.pde" in archive_files
    assert "sketch_simple/other.ino" in archive_files
    assert "sketch_simple/s_file.S" in archive_files
    assert "sketch_simple/sketch_simple.ino" in archive_files
    assert "sketch_simple/src/helper.h" in archive_files
    assert "sketch_simple/build/adafruit.samd.adafruit_feather_m0/sketch_simple.ino.hex" in archive_files
    assert "sketch_simple/build/adafruit.samd.adafruit_feather_m0/sketch_simple.ino.map" in archive_files
    assert "sketch_simple/build/arduino.avr.uno/sketch_simple.ino.eep" in archive_files
    assert "sketch_simple/build/arduino.avr.uno/sketch_simple.ino.hex" in archive_files
    assert "sketch_simple/build/arduino.avr.uno/sketch_simple.ino.with_bootloader.hex" in archive_files

    archive.close()


def test_archive_dot_arg_custom_zip_path_and_name_with_extension_with_include_build_dir_flag(
    run_command, copy_sketch, working_dir
):
    # Creates a folder where to save the zip
    archives_folder = f"{working_dir}/my_archives/"
    Path(archives_folder).mkdir()

    result = run_command(f"archive . {archives_folder}/my_custom_sketch.zip --include-build-dir", copy_sketch)
    assert result.ok

    archive = zipfile.ZipFile(f"{archives_folder}/my_custom_sketch.zip")
    archive_files = archive.namelist()

    assert "sketch_simple/doc.txt" in archive_files
    assert "sketch_simple/header.h" in archive_files
    assert "sketch_simple/merged_sketch.txt" in archive_files
    assert "sketch_simple/old.pde" in archive_files
    assert "sketch_simple/other.ino" in archive_files
    assert "sketch_simple/s_file.S" in archive_files
    assert "sketch_simple/sketch_simple.ino" in archive_files
    assert "sketch_simple/src/helper.h" in archive_files
    assert "sketch_simple/build/adafruit.samd.adafruit_feather_m0/sketch_simple.ino.hex" in archive_files
    assert "sketch_simple/build/adafruit.samd.adafruit_feather_m0/sketch_simple.ino.map" in archive_files
    assert "sketch_simple/build/arduino.avr.uno/sketch_simple.ino.eep" in archive_files
    assert "sketch_simple/build/arduino.avr.uno/sketch_simple.ino.hex" in archive_files
    assert "sketch_simple/build/arduino.avr.uno/sketch_simple.ino.with_bootloader.hex" in archive_files

    archive.close()


def test_archive_relative_sketch_path_with_include_build_dir_flag(run_command, copy_sketch, working_dir):
    result = run_command("archive ./sketch_simple --include-build-dir")
    assert result.ok

    archive = zipfile.ZipFile(f"{working_dir}/sketch_simple.zip")
    archive_files = archive.namelist()

    assert "sketch_simple/doc.txt" in archive_files
    assert "sketch_simple/header.h" in archive_files
    assert "sketch_simple/merged_sketch.txt" in archive_files
    assert "sketch_simple/old.pde" in archive_files
    assert "sketch_simple/other.ino" in archive_files
    assert "sketch_simple/s_file.S" in archive_files
    assert "sketch_simple/sketch_simple.ino" in archive_files
    assert "sketch_simple/src/helper.h" in archive_files
    assert "sketch_simple/build/adafruit.samd.adafruit_feather_m0/sketch_simple.ino.hex" in archive_files
    assert "sketch_simple/build/adafruit.samd.adafruit_feather_m0/sketch_simple.ino.map" in archive_files
    assert "sketch_simple/build/arduino.avr.uno/sketch_simple.ino.eep" in archive_files
    assert "sketch_simple/build/arduino.avr.uno/sketch_simple.ino.hex" in archive_files
    assert "sketch_simple/build/arduino.avr.uno/sketch_simple.ino.with_bootloader.hex" in archive_files

    archive.close()


def test_archive_absolute_sketch_path_with_include_build_dir_flag(run_command, copy_sketch, working_dir):
    result = run_command(f"archive {working_dir}/sketch_simple --include-build-dir", copy_sketch)
    assert result.ok

    archive = zipfile.ZipFile(f"{working_dir}/sketch_simple.zip")
    archive_files = archive.namelist()

    assert "sketch_simple/doc.txt" in archive_files
    assert "sketch_simple/header.h" in archive_files
    assert "sketch_simple/merged_sketch.txt" in archive_files
    assert "sketch_simple/old.pde" in archive_files
    assert "sketch_simple/other.ino" in archive_files
    assert "sketch_simple/s_file.S" in archive_files
    assert "sketch_simple/sketch_simple.ino" in archive_files
    assert "sketch_simple/src/helper.h" in archive_files
    assert "sketch_simple/build/adafruit.samd.adafruit_feather_m0/sketch_simple.ino.hex" in archive_files
    assert "sketch_simple/build/adafruit.samd.adafruit_feather_m0/sketch_simple.ino.map" in archive_files
    assert "sketch_simple/build/arduino.avr.uno/sketch_simple.ino.eep" in archive_files
    assert "sketch_simple/build/arduino.avr.uno/sketch_simple.ino.hex" in archive_files
    assert "sketch_simple/build/arduino.avr.uno/sketch_simple.ino.with_bootloader.hex" in archive_files

    archive.close()


def test_archive_relative_sketch_path_with_relative_zip_path_with_include_build_dir_flag(
    run_command, copy_sketch, working_dir
):
    # Creates a folder where to save the zip
    archives_folder = f"{working_dir}/my_archives/"
    Path(archives_folder).mkdir()

    result = run_command("archive ./sketch_simple ./my_archives --include-build-dir")
    assert result.ok

    archive = zipfile.ZipFile(f"{working_dir}/my_archives/sketch_simple.zip")
    archive_files = archive.namelist()

    assert "sketch_simple/doc.txt" in archive_files
    assert "sketch_simple/header.h" in archive_files
    assert "sketch_simple/merged_sketch.txt" in archive_files
    assert "sketch_simple/old.pde" in archive_files
    assert "sketch_simple/other.ino" in archive_files
    assert "sketch_simple/s_file.S" in archive_files
    assert "sketch_simple/sketch_simple.ino" in archive_files
    assert "sketch_simple/src/helper.h" in archive_files
    assert "sketch_simple/build/adafruit.samd.adafruit_feather_m0/sketch_simple.ino.hex" in archive_files
    assert "sketch_simple/build/adafruit.samd.adafruit_feather_m0/sketch_simple.ino.map" in archive_files
    assert "sketch_simple/build/arduino.avr.uno/sketch_simple.ino.eep" in archive_files
    assert "sketch_simple/build/arduino.avr.uno/sketch_simple.ino.hex" in archive_files
    assert "sketch_simple/build/arduino.avr.uno/sketch_simple.ino.with_bootloader.hex" in archive_files

    archive.close()


def test_archive_relative_sketch_path_with_absolute_zip_path_with_include_build_dir_flag(
    run_command, copy_sketch, working_dir
):
    # Creates a folder where to save the zip
    archives_folder = f"{working_dir}/my_archives/"
    Path(archives_folder).mkdir()

    result = run_command(f"archive ./sketch_simple {archives_folder} --include-build-dir")
    assert result.ok

    archive = zipfile.ZipFile(f"{archives_folder}/sketch_simple.zip")
    archive_files = archive.namelist()

    assert "sketch_simple/doc.txt" in archive_files
    assert "sketch_simple/header.h" in archive_files
    assert "sketch_simple/merged_sketch.txt" in archive_files
    assert "sketch_simple/old.pde" in archive_files
    assert "sketch_simple/other.ino" in archive_files
    assert "sketch_simple/s_file.S" in archive_files
    assert "sketch_simple/sketch_simple.ino" in archive_files
    assert "sketch_simple/src/helper.h" in archive_files
    assert "sketch_simple/build/adafruit.samd.adafruit_feather_m0/sketch_simple.ino.hex" in archive_files
    assert "sketch_simple/build/adafruit.samd.adafruit_feather_m0/sketch_simple.ino.map" in archive_files
    assert "sketch_simple/build/arduino.avr.uno/sketch_simple.ino.eep" in archive_files
    assert "sketch_simple/build/arduino.avr.uno/sketch_simple.ino.hex" in archive_files
    assert "sketch_simple/build/arduino.avr.uno/sketch_simple.ino.with_bootloader.hex" in archive_files

    archive.close()


def test_archive_relative_sketch_path_with_relative_zip_path_and_name_without_extension_with_include_build_dir_flag(
    run_command, copy_sketch, working_dir
):
    # Creates a folder where to save the zip
    archives_folder = f"{working_dir}/my_archives/"
    Path(archives_folder).mkdir()

    result = run_command("archive ./sketch_simple ./my_archives/my_custom_sketch --include-build-dir")
    assert result.ok

    archive = zipfile.ZipFile(f"{working_dir}/my_archives/my_custom_sketch.zip")
    archive_files = archive.namelist()

    assert "sketch_simple/doc.txt" in archive_files
    assert "sketch_simple/header.h" in archive_files
    assert "sketch_simple/merged_sketch.txt" in archive_files
    assert "sketch_simple/old.pde" in archive_files
    assert "sketch_simple/other.ino" in archive_files
    assert "sketch_simple/s_file.S" in archive_files
    assert "sketch_simple/sketch_simple.ino" in archive_files
    assert "sketch_simple/src/helper.h" in archive_files
    assert "sketch_simple/build/adafruit.samd.adafruit_feather_m0/sketch_simple.ino.hex" in archive_files
    assert "sketch_simple/build/adafruit.samd.adafruit_feather_m0/sketch_simple.ino.map" in archive_files
    assert "sketch_simple/build/arduino.avr.uno/sketch_simple.ino.eep" in archive_files
    assert "sketch_simple/build/arduino.avr.uno/sketch_simple.ino.hex" in archive_files
    assert "sketch_simple/build/arduino.avr.uno/sketch_simple.ino.with_bootloader.hex" in archive_files

    archive.close()


def test_archive_relative_sketch_path_with_relative_zip_path_and_name_with_extension_with_include_build_dir_flag(
    run_command, copy_sketch, working_dir
):
    # Creates a folder where to save the zip
    archives_folder = f"{working_dir}/my_archives/"
    Path(archives_folder).mkdir()

    result = run_command("archive ./sketch_simple ./my_archives/my_custom_sketch.zip --include-build-dir")
    assert result.ok

    archive = zipfile.ZipFile(f"{working_dir}/my_archives/my_custom_sketch.zip")
    archive_files = archive.namelist()

    assert "sketch_simple/doc.txt" in archive_files
    assert "sketch_simple/header.h" in archive_files
    assert "sketch_simple/merged_sketch.txt" in archive_files
    assert "sketch_simple/old.pde" in archive_files
    assert "sketch_simple/other.ino" in archive_files
    assert "sketch_simple/s_file.S" in archive_files
    assert "sketch_simple/sketch_simple.ino" in archive_files
    assert "sketch_simple/src/helper.h" in archive_files
    assert "sketch_simple/build/adafruit.samd.adafruit_feather_m0/sketch_simple.ino.hex" in archive_files
    assert "sketch_simple/build/adafruit.samd.adafruit_feather_m0/sketch_simple.ino.map" in archive_files
    assert "sketch_simple/build/arduino.avr.uno/sketch_simple.ino.eep" in archive_files
    assert "sketch_simple/build/arduino.avr.uno/sketch_simple.ino.hex" in archive_files
    assert "sketch_simple/build/arduino.avr.uno/sketch_simple.ino.with_bootloader.hex" in archive_files

    archive.close()


def test_archive_relative_sketch_path_with_absolute_zip_path_and_name_without_extension_with_include_build_dir_flag(
    run_command, copy_sketch, working_dir
):
    # Creates a folder where to save the zip
    archives_folder = f"{working_dir}/my_archives/"
    Path(archives_folder).mkdir()

    result = run_command(f"archive ./sketch_simple {archives_folder}/my_custom_sketch --include-build-dir")
    assert result.ok

    archive = zipfile.ZipFile(f"{archives_folder}/my_custom_sketch.zip")
    archive_files = archive.namelist()

    assert "sketch_simple/doc.txt" in archive_files
    assert "sketch_simple/header.h" in archive_files
    assert "sketch_simple/merged_sketch.txt" in archive_files
    assert "sketch_simple/old.pde" in archive_files
    assert "sketch_simple/other.ino" in archive_files
    assert "sketch_simple/s_file.S" in archive_files
    assert "sketch_simple/sketch_simple.ino" in archive_files
    assert "sketch_simple/src/helper.h" in archive_files
    assert "sketch_simple/build/adafruit.samd.adafruit_feather_m0/sketch_simple.ino.hex" in archive_files
    assert "sketch_simple/build/adafruit.samd.adafruit_feather_m0/sketch_simple.ino.map" in archive_files
    assert "sketch_simple/build/arduino.avr.uno/sketch_simple.ino.eep" in archive_files
    assert "sketch_simple/build/arduino.avr.uno/sketch_simple.ino.hex" in archive_files
    assert "sketch_simple/build/arduino.avr.uno/sketch_simple.ino.with_bootloader.hex" in archive_files

    archive.close()


def test_archive_relative_sketch_path_with_absolute_zip_path_and_name_with_extension_with_include_build_dir_flag(
    run_command, copy_sketch, working_dir
):
    # Creates a folder where to save the zip
    archives_folder = f"{working_dir}/my_archives/"
    Path(archives_folder).mkdir()

    result = run_command(f"archive ./sketch_simple {archives_folder}/my_custom_sketch.zip --include-build-dir")
    assert result.ok

    archive = zipfile.ZipFile(f"{archives_folder}/my_custom_sketch.zip")
    archive_files = archive.namelist()

    assert "sketch_simple/doc.txt" in archive_files
    assert "sketch_simple/header.h" in archive_files
    assert "sketch_simple/merged_sketch.txt" in archive_files
    assert "sketch_simple/old.pde" in archive_files
    assert "sketch_simple/other.ino" in archive_files
    assert "sketch_simple/s_file.S" in archive_files
    assert "sketch_simple/sketch_simple.ino" in archive_files
    assert "sketch_simple/src/helper.h" in archive_files
    assert "sketch_simple/build/adafruit.samd.adafruit_feather_m0/sketch_simple.ino.hex" in archive_files
    assert "sketch_simple/build/adafruit.samd.adafruit_feather_m0/sketch_simple.ino.map" in archive_files
    assert "sketch_simple/build/arduino.avr.uno/sketch_simple.ino.eep" in archive_files
    assert "sketch_simple/build/arduino.avr.uno/sketch_simple.ino.hex" in archive_files
    assert "sketch_simple/build/arduino.avr.uno/sketch_simple.ino.with_bootloader.hex" in archive_files

    archive.close()


def test_archive_absolute_sketch_path_with_relative_zip_path_with_include_build_dir_flag(
    run_command, copy_sketch, working_dir
):
    # Creates a folder where to save the zip
    archives_folder = f"{working_dir}/my_archives/"
    Path(archives_folder).mkdir()

    result = run_command(f"archive {working_dir}/sketch_simple ./my_archives --include-build-dir")
    assert result.ok

    archive = zipfile.ZipFile(f"{working_dir}/my_archives/sketch_simple.zip")
    archive_files = archive.namelist()

    assert "sketch_simple/doc.txt" in archive_files
    assert "sketch_simple/header.h" in archive_files
    assert "sketch_simple/merged_sketch.txt" in archive_files
    assert "sketch_simple/old.pde" in archive_files
    assert "sketch_simple/other.ino" in archive_files
    assert "sketch_simple/s_file.S" in archive_files
    assert "sketch_simple/sketch_simple.ino" in archive_files
    assert "sketch_simple/src/helper.h" in archive_files
    assert "sketch_simple/build/adafruit.samd.adafruit_feather_m0/sketch_simple.ino.hex" in archive_files
    assert "sketch_simple/build/adafruit.samd.adafruit_feather_m0/sketch_simple.ino.map" in archive_files
    assert "sketch_simple/build/arduino.avr.uno/sketch_simple.ino.eep" in archive_files
    assert "sketch_simple/build/arduino.avr.uno/sketch_simple.ino.hex" in archive_files
    assert "sketch_simple/build/arduino.avr.uno/sketch_simple.ino.with_bootloader.hex" in archive_files

    archive.close()


def test_archive_absolute_sketch_path_with_absolute_zip_path_with_include_build_dir_flag(
    run_command, copy_sketch, working_dir
):
    # Creates a folder where to save the zip
    archives_folder = f"{working_dir}/my_archives/"
    Path(archives_folder).mkdir()

    result = run_command(f"archive {working_dir}/sketch_simple {archives_folder} --include-build-dir", copy_sketch)
    assert result.ok

    archive = zipfile.ZipFile(f"{archives_folder}/sketch_simple.zip")
    archive_files = archive.namelist()

    assert "sketch_simple/doc.txt" in archive_files
    assert "sketch_simple/header.h" in archive_files
    assert "sketch_simple/merged_sketch.txt" in archive_files
    assert "sketch_simple/old.pde" in archive_files
    assert "sketch_simple/other.ino" in archive_files
    assert "sketch_simple/s_file.S" in archive_files
    assert "sketch_simple/sketch_simple.ino" in archive_files
    assert "sketch_simple/src/helper.h" in archive_files
    assert "sketch_simple/build/adafruit.samd.adafruit_feather_m0/sketch_simple.ino.hex" in archive_files
    assert "sketch_simple/build/adafruit.samd.adafruit_feather_m0/sketch_simple.ino.map" in archive_files
    assert "sketch_simple/build/arduino.avr.uno/sketch_simple.ino.eep" in archive_files
    assert "sketch_simple/build/arduino.avr.uno/sketch_simple.ino.hex" in archive_files
    assert "sketch_simple/build/arduino.avr.uno/sketch_simple.ino.with_bootloader.hex" in archive_files

    archive.close()


def test_archive_absolute_sketch_path_with_relative_zip_path_and_name_without_extension_with_include_build_dir_flag(
    run_command, copy_sketch, working_dir
):
    # Creates a folder where to save the zip
    archives_folder = f"{working_dir}/my_archives/"
    Path(archives_folder).mkdir()

    result = run_command(f"archive {working_dir}/sketch_simple ./my_archives/my_custom_sketch --include-build-dir")
    assert result.ok

    archive = zipfile.ZipFile(f"{working_dir}/my_archives/my_custom_sketch.zip")
    archive_files = archive.namelist()

    assert "sketch_simple/doc.txt" in archive_files
    assert "sketch_simple/header.h" in archive_files
    assert "sketch_simple/merged_sketch.txt" in archive_files
    assert "sketch_simple/old.pde" in archive_files
    assert "sketch_simple/other.ino" in archive_files
    assert "sketch_simple/s_file.S" in archive_files
    assert "sketch_simple/sketch_simple.ino" in archive_files
    assert "sketch_simple/src/helper.h" in archive_files
    assert "sketch_simple/build/adafruit.samd.adafruit_feather_m0/sketch_simple.ino.hex" in archive_files
    assert "sketch_simple/build/adafruit.samd.adafruit_feather_m0/sketch_simple.ino.map" in archive_files
    assert "sketch_simple/build/arduino.avr.uno/sketch_simple.ino.eep" in archive_files
    assert "sketch_simple/build/arduino.avr.uno/sketch_simple.ino.hex" in archive_files
    assert "sketch_simple/build/arduino.avr.uno/sketch_simple.ino.with_bootloader.hex" in archive_files

    archive.close()


def test_archive_absolute_sketch_path_with_relative_zip_path_and_name_with_extension_with_include_build_dir_flag(
    run_command, copy_sketch, working_dir
):
    # Creates a folder where to save the zip
    archives_folder = f"{working_dir}/my_archives/"
    Path(archives_folder).mkdir()

    result = run_command(f"archive {working_dir}/sketch_simple ./my_archives/my_custom_sketch.zip --include-build-dir")
    assert result.ok

    archive = zipfile.ZipFile(f"{working_dir}/my_archives/my_custom_sketch.zip")
    archive_files = archive.namelist()

    assert "sketch_simple/doc.txt" in archive_files
    assert "sketch_simple/header.h" in archive_files
    assert "sketch_simple/merged_sketch.txt" in archive_files
    assert "sketch_simple/old.pde" in archive_files
    assert "sketch_simple/other.ino" in archive_files
    assert "sketch_simple/s_file.S" in archive_files
    assert "sketch_simple/sketch_simple.ino" in archive_files
    assert "sketch_simple/src/helper.h" in archive_files
    assert "sketch_simple/build/adafruit.samd.adafruit_feather_m0/sketch_simple.ino.hex" in archive_files
    assert "sketch_simple/build/adafruit.samd.adafruit_feather_m0/sketch_simple.ino.map" in archive_files
    assert "sketch_simple/build/arduino.avr.uno/sketch_simple.ino.eep" in archive_files
    assert "sketch_simple/build/arduino.avr.uno/sketch_simple.ino.hex" in archive_files
    assert "sketch_simple/build/arduino.avr.uno/sketch_simple.ino.with_bootloader.hex" in archive_files

    archive.close()


def test_archive_absolute_sketch_path_with_absolute_zip_path_and_name_without_extension_with_include_build_dir_flag(
    run_command, copy_sketch, working_dir
):
    # Creates a folder where to save the zip
    archives_folder = f"{working_dir}/my_archives/"
    Path(archives_folder).mkdir()

    result = run_command(
        f"archive {working_dir}/sketch_simple {archives_folder}/my_custom_sketch --include-build-dir", copy_sketch
    )
    assert result.ok

    archive = zipfile.ZipFile(f"{archives_folder}/my_custom_sketch.zip")
    archive_files = archive.namelist()

    assert "sketch_simple/doc.txt" in archive_files
    assert "sketch_simple/header.h" in archive_files
    assert "sketch_simple/merged_sketch.txt" in archive_files
    assert "sketch_simple/old.pde" in archive_files
    assert "sketch_simple/other.ino" in archive_files
    assert "sketch_simple/s_file.S" in archive_files
    assert "sketch_simple/sketch_simple.ino" in archive_files
    assert "sketch_simple/src/helper.h" in archive_files
    assert "sketch_simple/build/adafruit.samd.adafruit_feather_m0/sketch_simple.ino.hex" in archive_files
    assert "sketch_simple/build/adafruit.samd.adafruit_feather_m0/sketch_simple.ino.map" in archive_files
    assert "sketch_simple/build/arduino.avr.uno/sketch_simple.ino.eep" in archive_files
    assert "sketch_simple/build/arduino.avr.uno/sketch_simple.ino.hex" in archive_files
    assert "sketch_simple/build/arduino.avr.uno/sketch_simple.ino.with_bootloader.hex" in archive_files

    archive.close()


def test_archive_absolute_sketch_path_with_absolute_zip_path_and_name_with_extension_with_include_build_dir_flag(
    run_command, copy_sketch, working_dir
):
    # Creates a folder where to save the zip
    archives_folder = f"{working_dir}/my_archives/"
    Path(archives_folder).mkdir()

    result = run_command(
        f"archive {working_dir}/sketch_simple {archives_folder}/my_custom_sketch.zip --include-build-dir", copy_sketch
    )
    assert result.ok

    archive = zipfile.ZipFile(f"{archives_folder}/my_custom_sketch.zip")
    archive_files = archive.namelist()

    assert "sketch_simple/doc.txt" in archive_files
    assert "sketch_simple/header.h" in archive_files
    assert "sketch_simple/merged_sketch.txt" in archive_files
    assert "sketch_simple/old.pde" in archive_files
    assert "sketch_simple/other.ino" in archive_files
    assert "sketch_simple/s_file.S" in archive_files
    assert "sketch_simple/sketch_simple.ino" in archive_files
    assert "sketch_simple/src/helper.h" in archive_files
    assert "sketch_simple/build/adafruit.samd.adafruit_feather_m0/sketch_simple.ino.hex" in archive_files
    assert "sketch_simple/build/adafruit.samd.adafruit_feather_m0/sketch_simple.ino.map" in archive_files
    assert "sketch_simple/build/arduino.avr.uno/sketch_simple.ino.eep" in archive_files
    assert "sketch_simple/build/arduino.avr.uno/sketch_simple.ino.hex" in archive_files
    assert "sketch_simple/build/arduino.avr.uno/sketch_simple.ino.with_bootloader.hex" in archive_files

    archive.close()
