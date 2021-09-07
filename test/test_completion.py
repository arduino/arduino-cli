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


def test_completion_no_args(run_command):
    result = run_command(["completion"])
    assert not result.ok
    assert "Error: accepts 1 arg(s), received 0" in result.stderr
    assert result.stdout == ""


def test_completion_bash(run_command):
    result = run_command(["completion", "bash"])
    assert result.ok
    assert result.stderr == ""
    assert "_arduino-cli_root_command()" in result.stdout
    assert "__start_arduino-cli()" in result.stdout


def test_completion_zsh(run_command):
    result = run_command(["completion", "zsh"])
    assert result.ok
    assert result.stderr == ""
    assert "#compdef _arduino-cli arduino-cli" in result.stdout
    assert "_arduino-cli()" in result.stdout


def test_completion_fish(run_command):
    result = run_command(["completion", "fish"])
    assert result.ok
    assert result.stderr == ""
    assert "# fish completion for arduino-cli" in result.stdout
    assert "function __arduino_cli_perform_completion" in result.stdout


def test_completion_powershell(run_command):
    result = run_command(["completion", "powershell"])
    assert result.ok
    assert result.stderr == ""
    assert "# powershell completion for arduino-cli" in result.stdout
    assert "Register-ArgumentCompleter -CommandName 'arduino-cli' -ScriptBlock" in result.stdout


def test_completion_bash_no_desc(run_command):
    result = run_command(["completion", "bash", "--no-descriptions"])
    assert not result.ok
    assert result.stdout == ""
    assert "Error: command description is not supported by bash" in result.stderr


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
