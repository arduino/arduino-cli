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


def test_completion_bash_no_desc(run_command):
    result = run_command(["completion", "bash", "--no-descriptions"])
    assert result.ok
    assert result.stderr == ""
    assert "# bash completion V2 for arduino-cli" in result.stdout
    assert "__start_arduino-cli()" in result.stdout
    assert "__completeNoDesc" in result.stdout


def test_completion_zsh_no_desc(run_command):
    result = run_command(["completion", "zsh", "--no-descriptions"])
    assert result.ok
    assert result.stderr == ""
    assert "#compdef _arduino-cli arduino-cli" in result.stdout
    assert "_arduino-cli()" in result.stdout
    assert "__completeNoDesc" in result.stdout


def test_completion_fish_no_desc(run_command):
    result = run_command(["completion", "fish", "--no-descriptions"])
    assert result.ok
    assert result.stderr == ""
    assert "# fish completion for arduino-cli" in result.stdout
    assert "function __arduino_cli_perform_completion" in result.stdout
    assert "__completeNoDesc" in result.stdout


def test_completion_powershell_no_desc(run_command):
    result = run_command(["completion", "powershell", "--no-descriptions"])
    assert not result.ok
    assert result.stdout == ""
    assert "Error: command description is not supported by powershell" in result.stderr


# test if the completion suggestions returned are meaningful
# we use the __complete hidden command
# https://github.com/spf13/cobra/blob/master/shell_completions.md#debugging

# test static completions
def test_static_completions(run_command):
    result = run_command(
        [
            "__complete",
            "--format",
            "",
        ]
    )
    assert "json" in result.stdout
    result = run_command(
        [
            "__complete",
            "--log-format",
            "",
        ]
    )
    assert "json" in result.stdout
    result = run_command(
        [
            "__complete",
            "--log-level",
            "",
        ]
    )
    assert "trace" in result.stdout


# here we test if the completions coming from the core are working
def test_config_completion(run_command):
    result = run_command(["__complete", "config", "add", ""])
    assert "board_manager.additional_urls" in result.stdout
    result = run_command(["__complete", "config", "remove", ""])
    assert "board_manager.additional_urls" in result.stdout
    result = run_command(["__complete", "config", "delete", ""])
    assert "board_manager.additional_urls" in result.stdout
    result = run_command(["__complete", "config", "set", ""])
    assert "board_manager.additional_urls" in result.stdout


# here we test if the completions coming from the libs are working
def test_lib_completion(run_command):
    assert run_command(["lib", "update-index"])
    result = run_command(["__complete", "lib", "install", ""], hide=True)
    assert "WiFi101" in result.stdout
    result = run_command(["__complete", "lib", "download", ""], hide=True)
    assert "WiFi101" in result.stdout
    result = run_command(["__complete", "lib", "uninstall", ""], hide=True)
    assert "WiFi101" not in result.stdout  # not yet installed

    assert run_command(["lib", "install", "WiFi101"])
    result = run_command(["__complete", "lib", "uninstall", ""])
    assert "WiFi101" in result.stdout
    result = run_command(["__complete", "lib", "examples", ""])
    assert "WiFi101" in result.stdout
    result = run_command(["__complete", "lib", "deps", ""])
    assert "WiFi101" in result.stdout


# here we test if the completions coming from the core are working
def test_core_completion(run_command):
    assert run_command(["core", "update-index"])
    result = run_command(["__complete", "core", "install", ""])
    assert "arduino:avr" in result.stdout
    result = run_command(["__complete", "core", "download", ""])
    assert "arduino:avr" in result.stdout
    result = run_command(["__complete", "core", "uninstall", ""])
    assert "arduino:avr" not in result.stdout

    # we install a core because the provided completions comes from it
    assert run_command(["core", "install", "arduino:avr@1.8.3"])

    result = run_command(["__complete", "core", "uninstall", ""])
    assert "arduino:avr" in result.stdout

    result = run_command(["__complete", "board", "details", "-b", ""])
    assert "arduino:avr:uno" in result.stdout
    result = run_command(["__complete", "burn-bootloader", "-b", ""])
    assert "arduino:avr:uno" in result.stdout
    result = run_command(["__complete", "compile", "-b", ""])
    assert "arduino:avr:uno" in result.stdout
    result = run_command(["__complete", "debug", "-b", ""])
    assert "arduino:avr:uno" in result.stdout
    result = run_command(["__complete", "lib", "examples", "-b", ""])
    assert "arduino:avr:uno" in result.stdout
    result = run_command(["__complete", "upload", "-b", ""])
    assert "arduino:avr:uno" in result.stdout
    result = run_command(["__complete", "monitor", "-b", ""])
    assert "arduino:avr:uno" in result.stdout
    result = run_command(["__complete", "burn-bootloader", "-l", ""])
    assert "network" in result.stdout
    result = run_command(["__complete", "compile", "-l", ""])
    assert "network" in result.stdout
    result = run_command(["__complete", "debug", "-l", ""])
    assert "network" in result.stdout
    result = run_command(["__complete", "upload", "-l", ""])
    assert "network" in result.stdout
    result = run_command(["__complete", "monitor", "-l", ""])
    assert "network" in result.stdout
    result = run_command(["__complete", "burn-bootloader", "-P", ""])
    assert "atmel_ice" in result.stdout
    result = run_command(["__complete", "compile", "-P", ""])
    assert "atmel_ice" in result.stdout
    result = run_command(["__complete", "debug", "-P", ""])
    assert "atmel_ice" in result.stdout
    result = run_command(["__complete", "upload", "-P", ""])
    assert "atmel_ice" in result.stdout
