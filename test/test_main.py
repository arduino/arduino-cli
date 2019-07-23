from invoke import run, Responder, exceptions
import os
import json
import pytest
import semver
from datetime import datetime


def test_command_help(run_command):
    result = run_command('help')
    assert result.ok
    assert result.stderr == ''
    assert 'Usage' in result.stdout


def test_command_lib_list(run_command):
    """
    When ouput is empty, nothing is printed out, no matter the output format
    """
    result = run_command('lib list')
    assert result.ok
    assert '' == result.stderr
    result = run_command('lib list --format json')
    assert '' == result.stdout


def test_command_lib_install(run_command):
    libs = ['\"AzureIoTProtocol_MQTT\"', '\"CMMC MQTT Connector\"', '\"WiFiNINA\"']
    # Should be safe to run install multiple times
    result_1 = run_command('lib install {}'.format(' '.join(libs)))
    assert result_1.ok
    result_2 = run_command('lib install {}'.format(' '.join(libs)))
    assert result_2.ok

def test_command_lib_update_index(run_command):
    result = run_command('lib update-index')
    assert result.ok
    assert 'Updating index: library_index.json downloaded' == result.stdout.splitlines()[-1].strip()

def test_command_lib_remove(run_command):
    libs = ['\"AzureIoTProtocol_MQTT\"', '\"CMMC MQTT Connector\"', '\"WiFiNINA\"']
    result = run_command('lib uninstall {}'.format(' '.join(libs)))
    assert result.ok

@pytest.mark.slow
def test_command_lib_search(run_command):
    result = run_command('lib search')
    assert result.ok
    out_lines = result.stdout.splitlines()
    libs = []
    # Create an array with just the name of the vars
    for line in out_lines:
        if 'Name: ' in line:
            libs.append(line.split()[1].strip('\"'))
    number_of_libs = len(libs)
    assert sorted(libs) == libs
    assert ['WiFi101', 'WiFi101OTA'] == [lib for lib in libs if 'WiFi101' in lib]
    result = run_command('lib search --format json')
    assert result.ok
    libs_found_from_json = json.loads(result.stdout)
    number_of_libs_from_json = len(libs_found_from_json.get('libraries'))
    assert number_of_libs == number_of_libs_from_json


def test_command_board_list(run_command):
    result = run_command('core update-index')
    assert result.ok
    result = run_command('board list --format json')
    assert result.ok
    # check is a valid json and contains a list of ports
    ports = json.loads(result.stdout).get('ports')
    assert isinstance(ports, list)
    for port in ports:
        assert 'protocol' in port
        assert 'protocol_label' in port


def test_command_board_listall(run_command):
    result = run_command('board listall')
    assert result.ok
    assert ['Board', 'Name', 'FQBN'] == result.stdout.splitlines()[0].strip().split()


def test_command_version(run_command):
    result = run_command('version --format json')
    assert result.ok
    parsed_out = json.loads(result.stdout)

    assert parsed_out.get('Application', False) == 'arduino-cli'
    assert isinstance(semver.parse(parsed_out.get('VersionString', False)), dict)
    assert isinstance(parsed_out.get('Commit', False), str)
