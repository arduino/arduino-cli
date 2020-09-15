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
import collections
import json


Board = collections.namedtuple("Board", "address fqbn package architecture id core")


def running_on_ci():
    """
    Returns whether the program is running on a CI environment
    """
    val = os.getenv("APPVEYOR") or os.getenv("DRONE") or os.getenv("GITHUB_WORKFLOW")
    return val is not None


def parse_json_traces(log_json_lines):
    trace_entries = []
    for entry in log_json_lines:
        entry = json.loads(entry)
        if entry.get("level") == "trace":
            trace_entries.append(entry.get("msg"))
    return trace_entries
