# This file is part of arduino-cli.

# Copyright 2019 ARDUINO SA (http://www.arduino.cc/)

# This software is released under the GNU General Public License version 3,
# which covers the main part of arduino-cli.
# The terms of this license can be found at:
# https://www.gnu.org/licenses/gpl-3.0.en.html

# You can be released from the requirements of the above licenses by purchasing
# a commercial license. Buying such a license is mandatory if you want to modify or
# otherwise use the software for commercial activities involving the Arduino
# software without disclosing the source code of your own applications. To purchase
# a commercial license, send an email to license@arduino.cc.
import os

import pytest
from invoke import run


@pytest.fixture(scope="session")
def sketchbook_path(tmpdir_factory):
    """
    A tmp folder will be created before running
    the tests and deleted at the end.
    """
    fn = tmpdir_factory.mktemp('ArduinoTest')
    return fn


@pytest.fixture(scope="session")
def runner(sketchbook_path):
    return Runner(sketchbook_path)


class Runner:
    def __init__(self, sketchbook_path):
        self.sketchbook_path = None
        self.cli_path = os.path.join(pytest.config.rootdir, '..', 'arduino-cli')

    def _cli_line(self, *args):
        # Accept a list of arguments cli_line('lib list --format json')
        # Return a full command line string e.g. 'arduino-cli help --format json'
        cli_full_line = ' '.join([self.cli_path, ' '.join(str(arg) for arg in args), "--sketchbook-path {}".format(self.sketchbook_path)])
        return cli_full_line
    
    def run(self, *args):
        return run(self._cli_line(*args), echo=False, hide=True, warn=True)
