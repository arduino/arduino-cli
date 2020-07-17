# How to contribute

First of all, thanks for contributing!

This document provides some basic guidelines for contributing to this
repository. To propose improvements or fix a bug, feel free to submit a PR.

## Legal requirements

Before we can accept your contributions you have to sign the [Contributor License Agreement][0]

## Pull Requests

In order to ease code reviews and have your contributions merged faster, here is
a list of items you can check before submitting a PR:

- Create small PRs that are narrowly focused on addressing a single concern.
- PR titles indirectly become part of the CHANGELOG so it's crucial to provide a
  good record of **what** change is being made in the title; **why** it was made
  will go in the PR description, along with a link to a GitHub issue if it
  exists.
- Write tests for the code you wrote.
- Open your PR against the `master` branch.
- Maintain **clean commit history** and use **meaningful commit messages**.
  PRs with messy commit history are difficult to review and require a lot of
  work to be merged.
- Your PR must pass all CI tests before we will merge it. If you're seeing an
  error and don't think
  it's your fault, it may not be! The reviewer will help you if there are test
  failures that seem
  not related to the change you are making.

## Prerequisites

To build the Arduino CLI from sources you need the following tools to be
available in your local environment:

- [Go][1] version 1.12 or later
- [Taskfile][2] to help you run the most common tasks from the command line

If you want to run integration tests you will also need:

- A serial port with an Arduino board attached
- A working [Python][3] environment, version 3.8 or later

If you're working on the gRPC interface you will also have to:

- download the [protoc][6] compiler
- run `go get -u github.com/golang/protobuf/protoc-gen-go`

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

#### Hardware requirements for running the full suite of integration tests:

An Arduino board attached to a serial port. The board must:

- Use one of the VID/PID pairs used by Arduino or their partners (as is the case
  with all modern official Arduino boards except the classic Nano).
- Accept uploads using the FQBN associated with that VID/PID (which will be the
  case unless you have installed a custom bootloader or removed the bootloader).

Note that running the integration tests will result in a sketch being uploaded
to every attached Arduino board meeting the above requirements.

#### Software requirements for running integration tests:

A working Python environment. Chances are that you already
have Python installed in your system, if this is not the case you can
[download][3] the official distribution or use the package manager provided by
your Operating System.

Some dependencies need to be installed before running the tests and to avoid
polluting your global Python environment with dependencies that might be only
used by the Arduino CLI, to do so we use [Poetry][poetry-website]. First you need to install it (you might need to `sudo`
the following command):

```shell
pip3 install --user poetry
```

For more installation options read the [official documentation][poetry-docs].

After Poetry has been installed you should be able to run the tests with:

```shell
task test-integration
```

This will automatically install the necessary dependencies, if not already installed, and run the integration tests automatically.

When editing any Python file in the project remember to run linting checks with:

```shell
task python:check
```

This will run `flake8` automatically and return any error in the code formatting, if not already installed it will also install integration tests dependencies.

In case of linting errors you should be able to solve most of them by automatically formatting with:

```shell
task python:format
```

## Working on docs

Documentation is provided to final users in form of static HTML content generated
from a tool called [MkDocs][9] and hosted on [GitHub Pages][7].

### Local development

Most of the documentation consists of static content written over several
Markdown files under the `docs` folder at the root of this git repository but
some other content is dynamically generated from the CI pipelines - this is the
case with the command line reference and the gRPC interface, for example.

If you want to check out how the documentation would look after some local
changes, you might need to reproduce what happens in the CI, generating the full
documentation website from your personal computer. To run the docs toolchain
locally, you need to have a few dependencies and tools installed:

- [Go][1] version 1.12 or later
- [Taskfile][2] to help you run the most common tasks from the command line
- A working [Python][3] environment, see [this paragraph](#integration-tests)
  if you need to setup one

Before running the toolchain, perform the following operations from the root of
the git repository (if you have a Python virtual environment, activate it before
proceeding):

- go get -u github.com/pseudomuto/protoc-gen-doc/cmd/protoc-gen-doc
- pip install -r requirements_docs.txt

When working on docs, you can launch a command that will take care of
generating the docs, build the static website and start a local server you can
later access with a web browser to see a preview of your changes. From the root
of the git repository run:

```shell
task docs:serve
```

If you don't see any error, hit http://127.0.0.1:8000 with your browser to
navigate the generated docs.

### Docs publishing

The present git repository has a special branch called `gh-pages` that contains
the generated HTML code for the docs website; every time a change is pushed to
this special branch, GitHub automatically triggers a [deployment][8] to pull the
change and publish a new version of the website. Do not open Pull Requests to
push changes to the `gh-pages` branch, that will be done exclusively from the
CI.

### Docs versioning

In order to provide support for multiple Arduino CLI releases, Documentation is
versioned so that visitors can select which version of the documentation website
should be displayed. Unfortunately this feature isn't provided by GitHub pages
or MkDocs, so we had to implement it on top of the generation process.

Before delving into the details of the generation process, here follow some
requirements that were established to provide versioned documentation:

- A special version of the documentation called `dev` is provided to reflect the
  status of the Arduino CLI on the `master` branch - this includes unreleased
  features and bugfixes.
- Docs are versioned after the minor version of an Arduino CLI release. For
  example, Arduino CLI `0.99.1` and `0.99.2` will be both covered by
  documentation version `0.99`.
- The landing page of the documentation website will automatically redirect
  visitors to the most recently released version of the Arduino CLI.

To implement the requirements above, the execution of MkDocs is wrapped using a
CLI tool called [Mike][10] that does a few things for us:

- It runs MkDocs targeting subfolders named after the Arduino CLI version, e.g.
  documentation for version `0.10.1` can be found under the folder `0.10`.
- It injects an HTML control into the documentation website that lets visitors
  choose which version of the docs to browse from a dropdown list.
- It provides a redirect to a version we decide when visitors hit the landing
  page of the documentation website.
- It pushes generated contents to the `gh-pages` branch.

> **Note:** unless you're working on the generation process itself, you should
> never run Mike from a local environment, either directly or through the Task
> `docs:publish`. This might result in unwanted changes to the public website.

### Docs automation

In order to avoid unwanted changes to the public website hosting the Arduino
CLI documentation, only Mike is allowed to push changes to the `gh-pages` branch,
and this only happens from within the CI, in a workflow named [docs][11].

The CI is responsible for guessing which version of the Arduino CLI we're
building docs for, so that generated contents will be stored in the appropriate
section of the documentation website. Because this guessing might be fairly
complex, the logic is implemented in a Python script called [`build.py`][12].
The script will determine the version of the Arduino CLI that was modified in
the current commit (either `dev` or an official, numbered release) and whether
the redirect to the latest version that happens on the landing page should be
updated or not.

## Internationalization (i18n)

In order to support i18n in the CLI, any messages that are intended to be translated
should be wrapped in a call to `i18n.Tr`. This call allows us to build a catalog of
translatable strings, replacing the reference string at runtime with the localized value.

Adding or modifying these messages requires an i18n update, as this process creates the
reference catalog that are shared with translators. For that reason, the `task check`
command will fail if the catalog was not updated to sync with changes to the source code.

To update the catalog, execute the following command and commit the changes.

```shell
task i18n:update
```

To verify that the catalog is up-to-date, you may execute the command:

```shell
task i18n:check
```

Example usage:

```golang
package main

import (
  "fmt"
  "github.com/arduino/arduino-cli/i18n"
)

func main() {
  fmt.Println(i18n.Tr("Hello World!"))
}
```

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
[7]: https://pages.github.com/
[8]: https://github.com/arduino/arduino-cli/deployments?environment=github-pages#activity-log
[9]: https://www.mkdocs.org/
[10]: https://github.com/jimporter/mike
[11]: https://github.com/arduino/arduino-cli/blob/master/.github/workflows/docs.yaml
[12]: https://github.com/arduino/arduino-cli/blob/master/docs/build.py
