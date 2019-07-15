from invoke import run,Responder

#from pytest import capsys


def func(capfd):
    # help = run('arduino-cli help', hide=True, echo=False)
    # capture = py.io.StdCaptureFD(out=False, in_=False )
    #with capfd.disabled():
    help = run('arduino-cli help', pty=True)
    # out,err = capture.reset()
    return help

def test_command_help(capfd):
    print(func(capfd))
    assert True == True
