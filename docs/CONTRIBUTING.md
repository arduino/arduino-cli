# How to contribute

First of all, thanks for contributing!

This document provides some basic guidelines for contributing to this
repository. To propose improvements or fix a bug, feel free to submit a PR.

## Legal requirements

Before we can accept your contributions you have to sign the [Contributor License Agreement][0]

## Prerequisites

To build the Arduino CLI from sources you need the following tools to be
available in your local enviroment:

* [Go][1] version 1.12 or later
* [Taskfile][2] to help you run the most common tasks from the command line

If you want to run integration tests you will also need:

* A serial port with an Arduino device attached
* A working [Python][3] environment, version 3.5 or later

If you're working on the gRPC interface you will also have to:

* download the [protoc][6] compiler
* run `go get -U github.com/golang/protobuf/protoc-gen-go`

## Building the source code

From the project folder root, just run:

```shell
task build
```

The project uses Go modules so dependencies will be downloaded automatically;
at the end of the build, you should find an `arduino-cli` executable in the
same folder.

## Running the tests

There are several checks and test suites in place to ensure the code works as
expected and is written in a way that's consistent across the whole codebase.
To avoid pushing changes that will cause the CI system to fail, you can run most
of the tests locally.

To ensure code style is consistent, run:

```shell
task check
```

To run unit tests:

```shell
task test-unit
```

To run integration tests (these will take some time and require special setup,
see following paragraph):

```shell
task test-integration
```
### Running only some tests

By default, all tests from all go packages are run. To run only unit
tests from one or more specific packages, you can set the TARGETS
environment variable, e.g.:

    TARGETS=./arduino/cores/packagemanager task test-unit

Alternatively, to run only some specific test(s), you can specify a regex
to match against the test function name:

    TEST_REGEX='^TestTryBuild.*' task test-unit

Both can be combined as well, typically to run only a specific test:

    TEST_REGEX='^TestFindBoardWithFQBN$' TARGETS=./arduino/cores/packagemanager task test-unit

### Integration tests

Being a command line interface, Arduino CLI is heavily interactive and it has to
stay consistent in accepting the user input and providing the expected output
and proper exit codes. On top of this, many Arduino CLI features involve
communicating with external devices, most likely through a serial
port, so unit tests can only go so far in giving us confidence that the code is
working.

For these reasons, in addition to regular unit tests the project has a suite of
integration tests that actually run Arduino CLI in a different process and
assess the options are correctly understood and the output is what we expect.

To run the full suite of integration tests you need an Arduino device attached
to a serial port and a working Python environment. Chances are that you already
have Python installed in your system, if this is not the case you can
[download][3] the official distribution or use the package manager provided by
your Operating System.

Some dependencies need to be installed before running the tests and to avoid
polluting your global Python enviroment with dependencies that might be only
used by the Arduino CLI, you can use a [virtual environment][4]. There are many
ways to manage virtual environments, for example you can use a productivity tool
called [hatch][5]. First you need to install it (you might need to `sudo`
the following command):

```shell
pip3 install --user hatch
```

Then you can create a virtual environment to be used while working on Arduino
CLI:

```shell
hatch env arduino-cli
```

At this point the virtual environment was created and you need to make it active
every time you open a new terminal session with the following command:

```shell
hatch shell arduino-cli
```

From now on, every package installed by Python will be confined to the
`arduino-cli` virtual environment, so you can proceed installing the
dependencies required with:

```shell
pip install -r test/requirements.txt
```

If the last step was successful, you should be able to run the tests with:

```shell
task test-integration
```

## Working on docs

Documentation consists of several Markdown files stored under the `docs` folder
at the root of the repo. Some of those files are automatically generated in the
CI pipeline that builds the documentation website so you won't find them in the
git repository and you need to generate them locally.

If you're working on docs and your changes are not trivial, you might want to
preview the documentation website locally, before opening a Pull Request. To run
the docs toolchain locally you need to have:

* [Go][1] version 1.12 or later
* [Taskfile][2] to help you run the most common tasks from the command line
* A working [Python][3] environment, version 3.5 or later

Before running the toolchain, perform the following operations:

* go get -u github.com/pseudomuto/protoc-gen-doc/cmd/protoc-gen-doc

When working on docs, you can launch a command that will take care of
generating the docs, build the static website and start a local server you can
access with your browser to see a preview of your changes - to launch this
command do:

```shell
task docs:serve
```

If you don't see any error, hit http://127.0.0.1:8000 with your browser.

## Pull Requests

In order to ease code reviews and have your contributions merged faster, here is
a list of items you can check before submitting a PR:

* Create small PRs that are narrowly focused on addressing a single concern.
* PR titles indirectly become part of the CHANGELOG so it's crucial to provide a
  good record of **what** change is being made in the title; **why** it was made
  will go in the PR description, along with a link to a GitHub issue if it
  exists.
* Write tests for the code you wrote.
* Open your PR against the `master` branch.
* Maintain **clean commit history** and use **meaningful commit messages**.
  PRs with messy commit history are difficult to review and require a lot of
  work to be merged.
* Your PR must pass all CI tests before we will merge it. If you're seeing an
  error and don't think
  it's your fault, it may not be! The reviewer will help you if there are test
  failures that seem
  not related to the change you are making.

## Additional settings

If you need to push a commit that's only shipping documentation changes or
example files, thus a complete no-op for the test suite, please start the commit
message with the string **[skip ci]** to skip the build and give that slot to
someone else who does need it.

If your PR doesn't need to be included in the changelog, please start the PR
title with the string **[skip changelog]**

[0]: https://cla-assistant.io/arduino/arduino-cli
[1]: https://golang.org/doc/install
[2]: https://taskfile.dev/#/installation
[3]: https://www.python.org/downloads/
[4]: https://docs.python.org/3/tutorial/venv.html
[5]: https://github.com/ofek/hatch
[6]: https://github.com/protocolbuffers/protobuf/releases