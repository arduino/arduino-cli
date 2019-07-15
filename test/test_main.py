from invoke import run, Responder
import os

this_test_path = os.path.dirname(os.path.realpath(__file__))
cli_path = os.path.join(this_test_path, '..', 'arduino-cli')

def run_help():
    help = run(cli_path + ' help', pty=True)
    return help


def test_command_help():
    print(run_help())
    assert True == True
