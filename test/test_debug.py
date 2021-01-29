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


from pathlib import Path


def test_debugger_starts(run_command, data_dir):
    # Init the environment explicitly
    assert run_command("core update-index")

    # Install cores
    assert run_command("core install arduino:samd")

    # Create sketch for testing
    sketch_name = "DebuggerStartTest"
    sketch_path = Path(data_dir, sketch_name)
    fqbn = "arduino:samd:mkr1000"

    assert run_command(f"sketch new {sketch_path}")

    # Build sketch
    assert run_command(f"compile -b {fqbn} {sketch_path}")

    programmer = "atmel_ice"
    # Starts debugger
    assert run_command(f"debug -b {fqbn} -P {programmer} {sketch_path} --info")


def test_debugger_with_pde_sketch_starts(run_command, data_dir):
    assert run_command("update")

    # Install core
    assert run_command("core install arduino:samd")

    # Create sketch for testing
    sketch_name = "DebuggerPdeSketchStartTest"
    sketch_path = Path(data_dir, sketch_name)
    fqbn = "arduino:samd:mkr1000"

    assert run_command(f"sketch new {sketch_path}")

    # Renames sketch file to pde
    Path(sketch_path, f"{sketch_name}.ino").rename(sketch_path / f"{sketch_name}.pde")

    # Build sketch
    assert run_command(f"compile -b {fqbn} {sketch_path}")

    programmer = "atmel_ice"
    # Starts debugger
    assert run_command(f"debug -b {fqbn} -P {programmer} {sketch_path} --info")
