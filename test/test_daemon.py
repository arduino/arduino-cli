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
import subprocess
import time
import pytest
import requests
import yaml
from prometheus_client.parser import text_string_to_metric_families

@pytest.mark.timeout(30)
def test_telemetry_prometheus_endpoint(pytestconfig, data_dir, downloads_dir):

    # Use raw subprocess here due to a missing functionality in pinned pyinvoke ver,
    # in order to launch and detach the cli process in daemon mode
    cli_path = os.path.join(str(pytestconfig.rootdir), "..", "arduino-cli")
    env = os.environ.copy()
    env["ARDUINO_DATA_DIR"] = data_dir
    env["ARDUINO_DOWNLOADS_DIR"] = downloads_dir
    env["ARDUINO_SKETCHBOOK_DIR"] = data_dir
    daemon = subprocess.Popen([cli_path, "daemon"], env=env)


    # wait for and then parse repertory file
    repertory_file = os.path.join(data_dir, "repertory.yaml")
    while not os.path.exists(repertory_file):
        time.sleep(1)
    with open(repertory_file, 'r') as stream:
        repertory = yaml.safe_load(stream)

        # Check if :9090/metrics endpoint is alive,
        # telemetry is enabled by default in daemon mode
        metrics = requests.get("http://localhost:9090/metrics").text
        family = next(text_string_to_metric_families(metrics))
        sample = family.samples[0]
        assert repertory["installation"]["id"] == sample.labels["installationID"]
    #add a fixture here!
    daemon.kill()
