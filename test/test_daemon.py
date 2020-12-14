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
import time

import pytest
import requests
import yaml
from prometheus_client.parser import text_string_to_metric_families
from requests.adapters import HTTPAdapter
from requests.packages.urllib3.util.retry import Retry


@pytest.mark.timeout(60)
def test_metrics_prometheus_endpoint(daemon_runner, data_dir):
    # Wait for the inventory file to be created and then parse it
    # in order to check the generated ids
    inventory_file = os.path.join(data_dir, "inventory.yaml")
    while not os.path.exists(inventory_file):
        time.sleep(1)
    with open(inventory_file, "r") as stream:
        inventory = yaml.safe_load(stream)

        # Check if :9090/metrics endpoint is alive,
        # metrics is enabled by default in daemon mode
        s = requests.Session()
        retries = Retry(total=3, backoff_factor=1, status_forcelist=[500, 502, 503, 504])
        s.mount("http://", HTTPAdapter(max_retries=retries))
        metrics = s.get("http://localhost:9090/metrics").text
        family = next(text_string_to_metric_families(metrics))
        sample = family.samples[0]
        assert inventory["installation"]["id"] == sample.labels["installationID"]
