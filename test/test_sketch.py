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
from pathlib import Path


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


def test_sketch_archive_relative_sketch_path_with_absolute_zip_path_with_include_build_dir_flag(
    run_command, copy_sketch, working_dir
):
    copy_sketch("sketch_simple")
    # Creates a folder where to save the zip
    archives_folder = f"{working_dir}/my_archives/"
    Path(archives_folder).mkdir()

    result = run_command(["sketch", "archive", "./sketch_simple", archives_folder, "--include-build-dir"])
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

    result = run_command(
        ["sketch", "archive", "./sketch_simple", "./my_archives/my_custom_sketch", "--include-build-dir"]
    )
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

    result = run_command(
        ["sketch", "archive", "./sketch_simple", "./my_archives/my_custom_sketch.zip", "--include-build-dir"]
    )
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

    result = run_command(
        ["sketch", "archive", "./sketch_simple", f"{archives_folder}/my_custom_sketch", "--include-build-dir"]
    )
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

    result = run_command(
        ["sketch", "archive", "./sketch_simple", f"{archives_folder}/my_custom_sketch.zip", "--include-build-dir"]
    )
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

    result = run_command(["sketch", "archive", f"{working_dir}/sketch_simple", "./my_archives", "--include-build-dir"])
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
        ["sketch", "archive", f"{working_dir}/sketch_simple", archives_folder, "--include-build-dir"],
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
        ["sketch", "archive", f"{working_dir}/sketch_simple", "./my_archives/my_custom_sketch", "--include-build-dir"]
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
        [
            "sketch",
            "archive",
            f"{working_dir}/sketch_simple",
            "./my_archives/my_custom_sketch.zip",
            "--include-build-dir",
        ]
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
        [
            "sketch",
            "archive",
            f"{working_dir}/sketch_simple",
            f"{archives_folder}/my_custom_sketch",
            "--include-build-dir",
        ],
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
        [
            "sketch",
            "archive",
            f"{working_dir}/sketch_simple",
            f"{archives_folder}/my_custom_sketch.zip",
            "--include-build-dir",
        ],
        copy_sketch("sketch_simple"),
    )
    assert result.ok

    archive = zipfile.ZipFile(f"{archives_folder}/my_custom_sketch.zip")
    archive_files = archive.namelist()

    verify_zip_contains_sketch_including_build_dir(archive_files)

    archive.close()


def test_sketch_archive_with_pde_main_file(run_command, copy_sketch, working_dir):
    sketch_name = "sketch_pde_main_file"
    sketch_dir = copy_sketch(sketch_name)
    sketch_file = Path(sketch_dir, f"{sketch_name}.pde")
    res = run_command(["sketch", "archive"], sketch_dir)
    assert res.ok
    assert "Sketches with .pde extension are deprecated, please rename the following files to .ino" in res.stderr
    assert str(sketch_file.relative_to(sketch_dir)) in res.stderr

    archive = zipfile.ZipFile(f"{working_dir}/{sketch_name}.zip")
    archive_files = archive.namelist()

    assert f"{sketch_name}/{sketch_name}.pde" in archive_files

    archive.close()


def test_sketch_archive_with_multiple_main_files(run_command, copy_sketch, working_dir):
    sketch_name = "sketch_multiple_main_files"
    sketch_dir = copy_sketch(sketch_name)
    sketch_file = Path(sketch_dir, f"{sketch_name}.pde")
    res = run_command(["sketch", "archive"], sketch_dir)
    assert res.failed
    assert "Sketches with .pde extension are deprecated, please rename the following files to .ino" in res.stderr
    assert str(sketch_file.relative_to(sketch_dir)) in res.stderr
    assert "Error archiving: Can't open sketch: multiple main sketch files found" in res.stderr


def test_sketch_archive_case_mismatch_fails(run_command, data_dir):
    sketch_name = "ArchiveSketchCaseMismatch"
    sketch_path = Path(data_dir, sketch_name)

    assert run_command(["sketch", "new", sketch_path])

    # Rename main .ino file so casing is different from sketch name
    Path(sketch_path, f"{sketch_name}.ino").rename(sketch_path / f"{sketch_name.lower()}.ino")

    res = run_command(["sketch", "archive", sketch_path])
    assert res.failed
    assert "Error archiving: Can't open sketch:" in res.stderr
