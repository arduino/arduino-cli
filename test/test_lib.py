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
import platform

import simplejson as json
import pytest
import shutil
from git import Repo
from pathlib import Path
import tempfile
import requests
import zipfile
import io
import re


@pytest.mark.skipif(
    platform.system() == "Windows",
    reason="Using a file uri as git url doesn't work on Windows, "
    + "this must be removed when this issue is fixed: https://github.com/go-git/go-git/issues/247",
)
def test_install_with_git_url_local_file_uri(run_command, downloads_dir, data_dir):
    assert run_command(["update"])

    env = {
        "ARDUINO_DATA_DIR": data_dir,
        "ARDUINO_DOWNLOADS_DIR": downloads_dir,
        "ARDUINO_SKETCHBOOK_DIR": data_dir,
        "ARDUINO_ENABLE_UNSAFE_LIBRARY_INSTALL": "true",
    }

    lib_install_dir = Path(data_dir, "libraries", "WiFi101")
    # Verifies library is not installed
    assert not lib_install_dir.exists()

    # Clone repository locally
    git_url = "https://github.com/arduino-libraries/WiFi101.git"
    repo_dir = Path(data_dir, "WiFi101")
    assert Repo.clone_from(git_url, repo_dir)

    assert run_command(["lib", "install", "--git-url", repo_dir.as_uri()], custom_env=env)

    # Verifies library is installed
    assert lib_install_dir.exists()
