from invoke import run, Responder
import os

this_test_path = os.path.dirname(os.path.realpath(__file__))
# Calculate absolute path of the CLI
cli_path = os.path.join(this_test_path, '..', 'arduino-cli')

# Useful reference: 
# http://docs.pyinvoke.org/en/1.2/api/runners.html#invoke.runners.Result

def cli_line(*args):
    # Accept a list of arguments cli_line('lib list --format json')
    # Return a full command line string e.g. 'arduino-cli help --format json'
    cli_full_line = ' '.join([cli_path, ' '.join(str(arg) for arg in args)])
    print(cli_full_line)
    return cli_full_line


# def run_command(*args):
#     result = run(cli_line(*args), echo=False)
#     return result

def run_help():
    help = run(cli_line('help'), echo=False)
    return help


def run_lib_list():
    help = run(cli_line('lib list', '--format json'), pty=True)
    # return help
    return


def test_command_help():
    result = run_help()
    assert result.ok
    assert result.stderr == ''
    assert 'Usage' in result.stdout
