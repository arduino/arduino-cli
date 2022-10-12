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


def test_sketch_archive_case_mismatch_fails(run_command, data_dir):
    sketch_name = "ArchiveSketchCaseMismatch"
    sketch_path = Path(data_dir, sketch_name)

    assert run_command(["sketch", "new", sketch_path])

    # Rename main .ino file so casing is different from sketch name
    Path(sketch_path, f"{sketch_name}.ino").rename(sketch_path / f"{sketch_name.lower()}.ino")

    res = run_command(["sketch", "archive", sketch_path])
    assert res.failed
    assert "Error archiving: Can't open sketch:" in res.stderr
