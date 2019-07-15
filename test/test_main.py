from invoke import run, Responder
import os
# import json

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
    # Accept a list of arguments
    # Resource: http://docs.pyinvoke.org/en/1.2/api/runners.html#invoke.runners.Runner
    # result = run(cli_line(*args), echo=False)  # , hide='out')
    print("Running: {}".format(cli_line(*args)))
    result = run(cli_line(*args), echo=False, hide='out')
    return result


def test_command_help():
    result = run_command('help')
    assert result.ok
    assert result.stderr == ''
    assert 'Usage' in result.stdout


def test_command_lib_list():
    result = run_command('lib list')
    assert result.ok
    assert result.stderr == ''
    result = run_command('lib list', '--format json')
    assert '{}' == result.stdout
    # assert json.loads('{}') == json.loads(result.stdout)


def test_command_lib_install():
    libs = ['\"AzureIoTProtocol_MQTT\"', '\"CMMC MQTT Connector\"', '\"WiFiNINA\"']
    result1 = run_command('lib install {}'.format(' '.join(libs)))
    result2 = run_command('lib install {}'.format(' '.join(libs)))
    # Installation should be idempotent
    assert result1 == result2
