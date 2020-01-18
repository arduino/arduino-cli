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
import platform


def test_cache_clean(run_command):
    """
    Clean the cache under arduino caching file directory which is
    "<Arduino configure file path>/staging"
    """
    result = run_command("cache clean")
    assert result.ok

    # Generate /staging directory
    result = run_command("lib list")
    assert result.ok

    result = run_command("cache clean")
    assert result.ok

    running_platform = platform.system()
    homeDir = os.path.expanduser("~")
    if running_platform == "Linux":
        assert not (os.path.isdir(homeDir + ".arduino15/staging"))
    elif running_platform == "Darwin":
        assert not (os.path.isdir(homeDir + "Library/Arduino15/staging"))
    elif running_platform == "Windows":
        assert not (os.path.isdir(homeDir + "Arduino15/staging"))
