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

import shutil
from git import Repo
from pathlib import Path


def test_compile_with_fully_precompiled_library(run_command, data_dir):
    assert run_command(["update"])

    assert run_command(["core", "install", "arduino:mbed@1.3.1"])
    fqbn = "arduino:mbed:nano33ble"

    # Install fully precompiled library
    # For more information see:
    # https://arduino.github.io/arduino-cli/latest/library-specification/#precompiled-binaries
    assert run_command(["lib", "install", "Arduino_TensorFlowLite@2.1.1-ALPHA-precompiled"])
    sketch_folder = Path(data_dir, "libraries", "Arduino_TensorFlowLite", "examples", "hello_world")

    # Install example dependency
    # assert run_command("lib install Arduino_LSM9DS1")

    # Compile and verify dependencies detection for fully precompiled library is skipped
    result = run_command(["compile", "-b", fqbn, sketch_folder, "-v"])
    assert result.ok
    assert "Skipping dependencies detection for precompiled library Arduino_TensorFlowLite" in result.stdout


def test_compile_sketch_with_pde_extension(run_command, data_dir):
    # Init the environment explicitly
    assert run_command(["update"])

    # Install core to compile
    assert run_command(["core", "install", "arduino:avr@1.8.3"])

    sketch_name = "CompilePdeSketch"
    sketch_path = Path(data_dir, sketch_name)
    fqbn = "arduino:avr:uno"

    # Create a test sketch
    assert run_command(["sketch", "new", sketch_path])

    # Renames sketch file to pde
    sketch_file = Path(sketch_path, f"{sketch_name}.ino").rename(sketch_path / f"{sketch_name}.pde")

    # Build sketch from folder
    res = run_command(["compile", "--clean", "-b", fqbn, sketch_path])
    assert res.ok
    assert "Sketches with .pde extension are deprecated, please rename the following files to .ino:" in res.stderr
    assert str(sketch_file) in res.stderr

    # Build sketch from file
    res = run_command(["compile", "--clean", "-b", fqbn, sketch_file])
    assert res.ok
    assert "Sketches with .pde extension are deprecated, please rename the following files to .ino" in res.stderr
    assert str(sketch_file) in res.stderr


def test_compile_sketch_with_multiple_main_files(run_command, data_dir):
    # Init the environment explicitly
    assert run_command(["update"])

    # Install core to compile
    assert run_command(["core", "install", "arduino:avr@1.8.3"])

    sketch_name = "CompileSketchMultipleMainFiles"
    sketch_path = Path(data_dir, sketch_name)
    fqbn = "arduino:avr:uno"

    # Create a test sketch
    assert run_command(["sketch", "new", sketch_path])

    # Copy .ino sketch file to .pde
    sketch_ino_file = Path(sketch_path, f"{sketch_name}.ino")
    sketch_pde_file = Path(sketch_path / f"{sketch_name}.pde")
    shutil.copyfile(sketch_ino_file, sketch_pde_file)

    # Build sketch from folder
    res = run_command(["compile", "--clean", "-b", fqbn, sketch_path])
    assert res.failed
    assert "Error opening sketch: multiple main sketch files found" in res.stderr

    # Build sketch from .ino file
    res = run_command(["compile", "--clean", "-b", fqbn, sketch_ino_file])
    assert res.failed
    assert "Error opening sketch: multiple main sketch files found" in res.stderr

    # Build sketch from .pde file
    res = run_command(["compile", "--clean", "-b", fqbn, sketch_pde_file])
    assert res.failed
    assert "Error opening sketch: multiple main sketch files found" in res.stderr


def test_compile_sketch_case_mismatch_fails(run_command, data_dir):
    # Init the environment explicitly
    assert run_command(["update"])

    # Install core to compile
    assert run_command(["core", "install", "arduino:avr@1.8.3"])

    sketch_name = "CompileSketchCaseMismatch"
    sketch_path = Path(data_dir, sketch_name)
    fqbn = "arduino:avr:uno"

    assert run_command(["sketch", "new", sketch_path])

    # Rename main .ino file so casing is different from sketch name
    sketch_main_file = Path(sketch_path, f"{sketch_name}.ino").rename(sketch_path / f"{sketch_name.lower()}.ino")

    # Verifies compilation fails when:
    # * Compiling with sketch path
    res = run_command(["compile", "--clean", "-b", fqbn, sketch_path])
    assert res.failed
    assert "Error opening sketch: no valid sketch found" in res.stderr
    # * Compiling with sketch main file
    res = run_command(["compile", "--clean", "-b", fqbn, sketch_main_file])
    assert res.failed
    assert "Error opening sketch: no valid sketch found" in res.stderr
    # * Compiling in sketch path
    res = run_command(["compile", "--clean", "-b", fqbn], custom_working_dir=sketch_path)
    assert res.failed
    assert "Error opening sketch: no valid sketch found" in res.stderr


def test_compile_with_only_compilation_database_flag(run_command, data_dir):
    assert run_command(["update"])

    assert run_command(["core", "install", "arduino:avr@1.8.3"])

    sketch_name = "CompileSketchOnlyCompilationDatabaseFlag"
    sketch_path = Path(data_dir, sketch_name)
    fqbn = "arduino:avr:uno"

    assert run_command(["sketch", "new", sketch_path])

    # Verifies no binaries exist
    build_path = Path(sketch_path, "build")
    assert not build_path.exists()

    # Compile with both --export-binaries and --only-compilation-database flags
    assert run_command(
        ["compile", "--export-binaries", "--only-compilation-database", "--clean", "-b", fqbn, sketch_path]
    )

    # Verifies no binaries are exported
    assert not build_path.exists()

    # Verifies no binaries exist
    build_path = Path(data_dir, "export-dir")
    assert not build_path.exists()

    # Compile by setting the --output-dir flag and --only-compilation-database flags
    assert run_command(
        ["compile", "--output-dir", build_path, "--only-compilation-database", "--clean", "-b", fqbn, sketch_path]
    )

    # Verifies no binaries are exported
    assert not build_path.exists()


def test_compile_using_platform_local_txt(run_command, data_dir):
    assert run_command(["update"])

    assert run_command(["core", "install", "arduino:avr@1.8.3"])

    sketch_name = "CompileSketchUsingPlatformLocalTxt"
    sketch_path = Path(data_dir, sketch_name)
    fqbn = "arduino:avr:uno"

    assert run_command(["sketch", "new", sketch_path])

    # Verifies compilation works without issues
    assert run_command(["compile", "--clean", "-b", fqbn, sketch_path])

    # Overrides default platform compiler with an unexisting one
    platform_local_txt = Path(data_dir, "packages", "arduino", "hardware", "avr", "1.8.3", "platform.local.txt")
    platform_local_txt.write_text("compiler.c.cmd=my-compiler-that-does-not-exist")

    # Verifies compilation now fails because compiler is not found
    res = run_command(["compile", "--clean", "-b", fqbn, sketch_path])
    assert res.failed
    assert "my-compiler-that-does-not-exist" in res.stderr


def test_compile_using_boards_local_txt(run_command, data_dir):
    assert run_command(["update"])

    assert run_command(["core", "install", "arduino:avr@1.8.3"])

    sketch_name = "CompileSketchUsingBoardsLocalTxt"
    sketch_path = Path(data_dir, sketch_name)
    # Use a made up board
    fqbn = "arduino:avr:nessuno"

    assert run_command(["sketch", "new", sketch_path])

    # Verifies compilation fails because board doesn't exist
    res = run_command(["compile", "--clean", "-b", fqbn, sketch_path])
    assert res.failed
    assert "Error during build: Error resolving FQBN: board arduino:avr:nessuno not found" in res.stderr

    # Use custom boards.local.txt with made arduino:avr:nessuno board
    boards_local_txt = Path(data_dir, "packages", "arduino", "hardware", "avr", "1.8.3", "boards.local.txt")
    shutil.copyfile(Path(__file__).parent / "testdata" / "boards.local.txt", boards_local_txt)

    assert run_command(["compile", "--clean", "-b", fqbn, sketch_path])


def test_compile_manually_installed_platform(run_command, data_dir):
    assert run_command(["update"])

    sketch_name = "CompileSketchManuallyInstalledPlatformUsingPlatformLocalTxt"
    sketch_path = Path(data_dir, sketch_name)
    fqbn = "arduino-beta-development:avr:uno"
    assert run_command(["sketch", "new", sketch_path])

    # Manually installs a core in sketchbooks hardware folder
    git_url = "https://github.com/arduino/ArduinoCore-avr.git"
    repo_dir = Path(data_dir, "hardware", "arduino-beta-development", "avr")
    assert Repo.clone_from(git_url, repo_dir, multi_options=["-b 1.8.3"])

    # Installs also the same core via CLI so all the necessary tools are installed
    assert run_command(["core", "install", "arduino:avr@1.8.3"])

    # Verifies compilation works without issues
    assert run_command(["compile", "--clean", "-b", fqbn, sketch_path])


def test_compile_manually_installed_platform_using_platform_local_txt(run_command, data_dir):
    assert run_command(["update"])

    sketch_name = "CompileSketchManuallyInstalledPlatformUsingPlatformLocalTxt"
    sketch_path = Path(data_dir, sketch_name)
    fqbn = "arduino-beta-development:avr:uno"
    assert run_command(["sketch", "new", sketch_path])

    # Manually installs a core in sketchbooks hardware folder
    git_url = "https://github.com/arduino/ArduinoCore-avr.git"
    repo_dir = Path(data_dir, "hardware", "arduino-beta-development", "avr")
    assert Repo.clone_from(git_url, repo_dir, multi_options=["-b 1.8.3"])

    # Installs also the same core via CLI so all the necessary tools are installed
    assert run_command(["core", "install", "arduino:avr@1.8.3"])

    # Verifies compilation works without issues
    assert run_command(["compile", "--clean", "-b", fqbn, sketch_path])

    # Overrides default platform compiler with an unexisting one
    platform_local_txt = Path(repo_dir, "platform.local.txt")
    platform_local_txt.write_text("compiler.c.cmd=my-compiler-that-does-not-exist")

    # Verifies compilation now fails because compiler is not found
    res = run_command(["compile", "--clean", "-b", fqbn, sketch_path])
    assert res.failed
    assert "my-compiler-that-does-not-exist" in res.stderr
