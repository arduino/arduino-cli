# This file is part of arduino-cli.
#
# Copyright 2019 ARDUINO SA (http://www.arduino.cc/)
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
import collections
import os
import pytest
import invoke

from invoke.context import Context

Board = collections.namedtuple("Board", "address fqbn package architecture id core")


def running_on_ci():
    """
    Returns whether the program is running on a CI environment
    """
    val = os.getenv("APPVEYOR") or os.getenv("DRONE") or os.getenv("GITHUB_WORKFLOW")
    return val is not None


def build_runner(cli_path, env, working_dir):
    """
    Provide a wrapper around invoke's `run` API so that every test
    will work in its own temporary folder.

    Useful reference:
        http://docs.pyinvoke.org/en/1.2/api/runners.html#invoke.runners.Result

    :param cli_path: the path to the ``arduino-cli`` executable file.
    :param env: a ``dict`` with the environment variables to use.
    :param working_dir: the CWD where the command will be executed.

    :returns a runner function with the mechanic to run an ``arduino-cli`` instance
    with a given environment ``env`` in the directory ```working_dir`.
    """

    def _run(cmd_string):
        cli_full_line = "{} {}".format(cli_path, cmd_string)
        run_context = Context()
        with run_context.cd(working_dir):
            return invoke.run(cli_full_line, echo=False, hide=True, warn=True, env=env)

    return _run
