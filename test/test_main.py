from invoke import run,Responder

#from pytest import capsys


def func():
    help = run('arduino-cli help', pty=True)
    return help

def test_command_help():
    print(func())
    assert True == True
