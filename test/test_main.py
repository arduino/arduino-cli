from invoke import run, Responder
import os
import json

this_test_path = os.path.dirname(os.path.realpath(__file__))
# Calculate absolute path of the CLI
cli_path = os.path.join(this_test_path, '..', 'arduino-cli')

# Useful reference: 
# http://docs.pyinvoke.org/en/1.2/api/runners.html#invoke.runners.Result


def cli_line(*args):
    # Accept a list of arguments cli_line('lib list --format json')
    # Return a full command line string e.g. 'arduino-cli help --format json'
    cli_full_line = ' '.join([cli_path, ' '.join(str(arg) for arg in args)])
    # print(cli_full_line)
    return cli_full_line


def run_command(*args):
    result = run(cli_line(*args), echo=False, hide='out')
    return result


def test_command_help():
    result = run_command('help')
    assert result.ok
    assert result.stderr == ''
    assert 'Usage' in result.stdout
    # result.out


def test_command_lib_list():
    result = run_command('lib list')
    assert result.stderr == ''
    result = run_command('lib list', '--format json')
    assert '{}' == result.stdout


def test_command_lib_install():
    libs = ['\"AzureIoTProtocol_MQTT\"', '\"CMMC MQTT Connector\"', '\"WiFiNINA\"']
    # Should be safe to run install multiple times
    result_1 = run_command('lib install {}'.format(' '.join(libs)))
    assert result_1.ok
    result_2 = run_command('lib install {}'.format(' '.join(libs)))
    assert result_2.ok


def test_command_lib_remove():
    libs = ['\"AzureIoTProtocol_MQTT\"', '\"CMMC MQTT Connector\"', '\"WiFiNINA\"']
    result = run_command('lib uninstall {}'.format(' '.join(libs)))


def test_command_board_list():
    result = run_command('board list --format json')
    # check is a valid json and contains a list of ports
    ports = json.loads(result.stdout).get('ports')
    assert isinstance(ports, list)
    for port in ports:
        assert 'protocol' in port
        assert 'protocol_label' in port


def test_command_board_listall():
    result = run_command('board listall')
    assert ['Board', 'Name', 'FQBN'] == result.stdout.splitlines()[0].strip().split()

