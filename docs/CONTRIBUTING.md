# How to contribute

First of all, thanks for contributing! This document provides some basic guidelines for contributing to this repository.

There are several ways you can get involved:

| Type of contribution                              | Contribution method                                     |
| ------------------------------------------------- | ------------------------------------------------------- |
| - Support request<br/>- Question<br/>- Discussion | Post on the [Arduino Forum][forum]                      |
| - Bug report<br/>- Feature request                | Issue report (read the [issue guidelines][issues])      |
| Beta testing                                      | Try out the [nightly build][nightly]                    |
| - Bug fix<br/>- Enhancement                       | Pull Request (read the [pull request guidelines][prs])  |
| Translations for Arduino CLI                      | [transifex][translate]                                  |
| Monetary                                          | - [Donate][donate]<br/>- [Buy official products][store] |

## Issue Reports

Do you need help or have a question about using Arduino CLI? Support requests should be made to Arduino CLI's dedicated
board in the [Arduino forum][forum].

High quality bug reports and feature requests are valuable contributions to the Arduino CLI project.

### Before reporting an issue

- Give the [nightly build][nightly] a test drive to see if your issue was already resolved.
- Search [existing pull requests and issues][issue-tracker] to see if it was already reported. If you have additional
  information to provide about an existing issue, please comment there. You can use the [Reactions feature][reactions]
  if you only want to express support.

### Qualities of an excellent report

- The issue title should be descriptive. Vague titles make it difficult to understand the purpose of the issue, which
  might cause your issue to be overlooked.
- Provide a full set of steps necessary to reproduce the issue. Demonstration code or commands should be complete and
  simplified to the minimum necessary to reproduce the issue.
- Be responsive. We may need you to provide additional information in order to investigate and resolve the issue.
- If you find a solution to your problem, please comment on your issue report with an explanation of how you were able
  to fix it and close the issue.

## Pull Requests

To propose improvements or fix a bug, feel free to submit a PR.

### Legal requirements

Before we can accept your contributions you have to sign the [Contributor License Agreement][0]

### Pull request checklist

In order to ease code reviews and have your contributions merged faster, here is a list of items you can check before
submitting a PR:

- Create small PRs that are narrowly focused on addressing a single concern.
- PR titles indirectly become part of the CHANGELOG so it's crucial to provide a good record of **what** change is being
  made in the title; **why** it was made will go in the PR description, along with a link to a GitHub issue if it
  exists.
- <a id="breaking"></a> If the PR contains a breaking change, please start the commit message and PR title with the
  string **[breaking]**. Don't forget to describe in the PR description and in the [`UPGRADING.md`][upgrading-file] file
  what changes users might need to make in their workflow or application due to this PR. A breaking change is a change
  that forces users to change their code, command-line invocations, build scripts or data files when upgrading from an
  older version of Arduino CLI.
- Write tests for the code you wrote.
- Open your PR against the `master` branch.
- Maintain **clean commit history** and use **meaningful commit messages**. PRs with messy commit history are difficult
  to review and require a lot of work to be merged.
- Your PR must pass all CI tests before we will merge it. If you're seeing an error and don't think it's your fault, it
  may not be! The reviewer will help you if there are test failures that seem not related to the change you are making.

### Prerequisites

To build the Arduino CLI from sources you need the following tools to be available in your local environment:

- [Go][1] version 1.17 or later
- [Taskfile][2] to help you run the most common tasks from the command line

If you want to run integration tests you will also need:

- A serial port with an Arduino board attached
- A working [Python][3] environment, version 3.8 or later

If you're working on the gRPC interface you will also have to:

- download the [protoc][6] compiler
- run `go get -u github.com/golang/protobuf/protoc-gen-go`

### Building the source code

From the project folder root, just run:

```shell
task build
```

The project uses Go modules so dependencies will be downloaded automatically. At the end of the build, you should find
an `arduino-cli` executable in the same folder.

### Running the tests

There are several checks and test suites in place to ensure the code works as expected and is written in a way that's
consistent across the whole codebase. To avoid pushing changes that will cause the CI system to fail, you can run most
of the tests locally.

To ensure code style is consistent, run:

```shell
task check
```

To run unit tests:

```shell
task go:test
```

To run integration tests (these will take some time and require special setup, see following paragraph):

```shell
task go:test-integration
```

#### Running only some tests

By default, all tests from all go packages are run. To run only unit tests from one or more specific packages, you can
set the TARGETS environment variable, e.g.:

```
TARGETS=./arduino/cores/packagemanager task go:test
```

Alternatively, to run only some specific test(s), you can specify a regex to match against the test function name:

```
TEST_REGEX='^TestTryBuild.*' task go:test
```

Both can be combined as well, typically to run only a specific test:

```
TEST_REGEX='^TestFindBoardWithFQBN$' TARGETS=./arduino/cores/packagemanager task go:test
```

### Integration tests

Being a command line interface, Arduino CLI is heavily interactive and it has to stay consistent in accepting the user
input and providing the expected output and proper exit codes. On top of this, many Arduino CLI features involve
communicating with external devices, most likely through a serial port, so unit tests can only go so far in giving us
confidence that the code is working.

For these reasons, in addition to regular unit tests the project has a suite of integration tests that actually run
Arduino CLI in a different process and assess the options are correctly understood and the output is what we expect.

##### Hardware requirements for running the full suite of integration tests:

An Arduino board attached to a serial port. The board must:

- Use one of the VID/PID pairs used by Arduino or their partners (as is the case with all modern official Arduino boards
  except the classic Nano).
- Accept uploads using the FQBN associated with that VID/PID (which will be the case unless you have installed a custom
  bootloader or removed the bootloader).

Note that running the integration tests will result in a sketch being uploaded to every attached Arduino board meeting
the above requirements.

##### Software requirements for running integration tests:

A working Python environment. Chances are that you already have Python installed in your system, if this is not the case
you can [download][3] the official distribution or use the package manager provided by your Operating System.

Some dependencies need to be installed before running the tests and to avoid polluting your global Python environment
with dependencies that might be only used by the Arduino CLI, to do so we use [Poetry][poetry-website]. First you need
to install it (you might need to `sudo` the following command):

```shell
pip3 install --user poetry
```

For more installation options read the [official documentation][poetry-docs].

#### Running tests

After the software requirements have been installed you should be able to run the tests with:

```shell
task go:test-integration
```

This will automatically install the necessary dependencies, if not already installed, and run the integration tests
automatically.

To run specific modules you must run `pytest` from the virtual environment created by Poetry.

```shell
poetry run pytest test/test_lib.py
```

To run very specific test functions:

```shell
poetry run pytest test/test_lib.py::test_list
```

You can avoid writing the `poetry run` prefix each time by creating a new shell inside the virtual environment:

```shell
poetry shell
pytest test_lib.py
pytest test_lib.py::test_list
```

#### Linting and formatting

When editing any Python file in the project remember to run linting checks with:

```shell
task python:lint
```

This will run `flake8` automatically and return any error in the code formatting, if not already installed it will also
install integration tests dependencies.

In case of linting errors you should be able to solve most of them by automatically formatting with:

```shell
task python:format
```

### Dependency license metadata

Metadata about the license types of all dependencies is cached in the repository. To update this cache, run the
following command from the repository root folder:

```
task general:cache-dep-licenses
```

The necessary **Licensed** tool can be installed by following
[these instructions](https://github.com/github/licensed#as-an-executable).

#### Configuration files formatting

To keep the configurations tidy and in order we use [Prettier][prettier-website] to automatically format all YAML files
in the project. Keeping and enforcing a formatting standard helps everyone make small PRs and avoids the introduction of
formatting changes made by unconfigured editors.

There are several ways to run Prettier. If you're using Visual Studio Code you can easily use the [`prettier-vscode`
extension][prettier-vscode-extension] to automatically format as you write.

Otherwise you can use the following tasks. To do so you'll need to install `npm` if not already installed. Check the
[official documentation][npm-install-docs] to learn how to install `npm` for your platform.

Ensure the formatting is compliant by running the command:

```shell
task general:format-prettier
```

When opening a new Pull Request, checks are automatically run to verify that configuration files are correctly
formatted. In case of failures we might ask you to update the PR with correct formatting.

### Working on docs

Documentation is provided to final users in form of static HTML content generated from a tool called [MkDocs][9] and
hosted on [GitHub Pages][7].

#### Local development

Most of the documentation consists of static content written over several Markdown files under the `docs` folder at the
root of this git repository but some other content is dynamically generated from the CI pipelines - this is the case
with the command line reference and the gRPC interface, for example.

If you want to check out how the documentation would look after some local changes, you might need to reproduce what
happens in the CI, generating the full documentation website from your personal computer. To run the docs toolchain
locally, you need to have a few dependencies and tools installed:

- [Go][1] version 1.17 or later
- [Taskfile][2] to help you run the most common tasks from the command line
- A working [Python][3] environment, see [this paragraph](#integration-tests) if you need to setup one

Before running the toolchain, perform the following operations from the root of the git repository (if you have a Python
virtual environment, activate it before proceeding):

- go get -u github.com/pseudomuto/protoc-gen-doc/cmd/protoc-gen-doc
- poetry install

When working on docs, you can launch a command that will take care of generating the docs, build the static website and
start a local server you can later access with a web browser to see a preview of your changes. From the root of the git
repository run:

```shell
task website:serve
```

If you don't see any error, hit http://127.0.0.1:8000 with your browser to navigate the generated docs.

#### Docs publishing

The present git repository has a special branch called `gh-pages` that contains the generated HTML code for the docs
website; every time a change is pushed to this special branch, GitHub automatically triggers a deployment to pull the
change and publish a new version of the website. Do not open Pull Requests to push changes to the `gh-pages` branch,
that will be done exclusively from the CI.

#### Docs formatting

To keep the documentation tidy and in order we use [Prettier][prettier-website] to automatically format all Markdown
files in the project. Keeping and enforcing a formatting standard helps everyone make small PRs and avoids the
introduction of formatting changes made by unconfigured editors.

There are several ways to run Prettier. If you're using Visual Studio Code you can easily use the [`prettier-vscode`
extension][prettier-vscode-extension] to automatically format as you write.

Otherwise you can use the following tasks. To do so you'll need to install `npm` if not already installed. Check the
[official documentation][npm-install-docs] to learn how to install `npm` for your platform.

Ensure the formatting is compliant by running the command:

```shell
task general:format-prettier
```

When opening a new Pull Request, checks are automatically run to verify that documentation is correctly formatted. In
case of failures we might ask you to update the PR with correct formatting.

#### Docs automation

In order to avoid unwanted changes to the public website hosting the Arduino CLI documentation, only Mike is allowed to
push changes to the `gh-pages` branch, and this only happens from within the CI, in a workflow named [Deploy
Website][11].

Details on the documentation publishing system are available [here][12].

### Internationalization (i18n)

In order to support i18n in the CLI, any messages that are intended to be translated should be wrapped in a call to
`i18n.Tr`. This call allows us to build a catalog of translatable strings, replacing the reference string at runtime
with the localized value.

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

### About `easyjson` golang library

We use the hi-performance `easyjson` library to parse the large JSON index files for libraries and platforms. To obtain
the best performance we must do some code generation, this is done via `task go:easyjson-generate`. If you ever touch
source code using the `easyjson` library, make sure to re-run the `go:easyjson-generate` task to see if there are
changes in the generated code.

### Additional settings

If you need to push a commit that's only shipping documentation changes or example files, thus a complete no-op for the
test suite, please start the commit message with the string **[skip ci]** to skip the build and give that slot to
someone else who does need it.

If your PR doesn't need to be included in the changelog, please start the commit message and PR title with the string
**[skip changelog]**

[0]: https://cla-assistant.io/arduino/arduino-cli
[1]: https://go.dev/doc/install
[2]: https://taskfile.dev/#/installation
[3]: https://www.python.org/downloads/
[6]: https://github.com/protocolbuffers/protobuf/releases
[7]: https://pages.github.com/
[9]: https://www.mkdocs.org/
[11]: https://github.com/arduino/arduino-cli/blob/master/.github/workflows/deploy-cobra-mkdocs-versioned-poetry.yml
[12]:
  https://github.com/arduino/tooling-project-assets/blob/main/workflow-templates/deploy-cobra-mkdocs-versioned-poetry.md
[forum]: https://forum.arduino.cc/index.php?board=145.0
[issues]: #issue-reports
[nightly]: https://arduino.github.io/arduino-cli/latest/installation/#nightly-builds
[prs]: #pull-requests
[translate]: https://www.transifex.com/arduino-1/arduino-cli/
[donate]: https://www.arduino.cc/en/Main/Contribute
[store]: https://store.arduino.cc
[issue-tracker]: https://github.com/arduino/arduino-cli/issues?q=
[reactions]: https://github.com/blog/2119-add-reactions-to-pull-requests-issues-and-comments
[prettier-website]: https://prettier.io/
[prettier-vscode-extension]: https://github.com/prettier/prettier-vscode
[npm-install-docs]: https://docs.npmjs.com/downloading-and-installing-node-js-and-npm
[poetry-website]: https://python-poetry.org/
[poetry-docs]: https://python-poetry.org/docs/
[upgrading-file]: UPGRADING.md
