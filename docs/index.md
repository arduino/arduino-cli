Arduino CLI is an all-in-one solution that provides builder, boards/library manager,
uploader, discovery and many other tools needed to use any Arduino compatible board
and platforms.

## Installation

You have several options to install the latest version of the Arduino
CLI on your system.

### Install via Homebrew (macOS/Linux)

The Arduino CLI is available as a Homebrew formula since version
`0.5.0`:

```sh
brew update
brew install arduino-cli
```

### Use the install script

The easiest way to get the latest version of the Arduino CLI on any
supported platform is using the `install.sh` script:

```sh
curl -fsSL https://raw.githubusercontent.com/arduino/arduino-cli/master/install.sh | sh
```

The script will install `arduino-cli` at `$PWD/bin` but if you want to target a
different directory, for example `~/local/bin`, set the `BINDIR` environment
variable like this:

```sh
curl -fsSL https://raw.githubusercontent.com/arduino/arduino-cli/master/install.sh | BINDIR=~/local/bin sh
```

### Download the latest packages

You can download the latest version of the pre-built binaries for the supported
platforms from the [release page](https://github.com/arduino/arduino-cli/releases)
or following the links in the following table. Once downloaded, extract the
binary `arduino-cli` into a directory that's is in your `PATH`.

Platform  |                    |                    |
--------- | ------------------ | ------------------ |
Linux     | [Linux 32 bit]     | [Linux 64 bit]     |
Linux ARM | [Linux ARM 32 bit] | [Linux ARM 64 bit] |
Windows   | [Windows 32 bit]   | [Windows 64 bit]   |
Mac OSX   |                    | [Mac OSX]          |

[Linux 64 bit]: https://downloads.arduino.cc/arduino-cli/arduino-cli_latest_Linux_64bit.tar.gz
[Linux 32 bit]: https://downloads.arduino.cc/arduino-cli/arduino-cli_latest_Linux_32bit.tar.gz
[Linux ARM 64 bit]: https://downloads.arduino.cc/arduino-cli/arduino-cli_latest_Linux_ARM64.tar.gz
[Linux ARM 32 bit]: https://downloads.arduino.cc/arduino-cli/arduino-cli_latest_Linux_ARMv7.tar.gz
[Windows 64 bit]: https://downloads.arduino.cc/arduino-cli/arduino-cli_latest_Windows_64bit.zip
[Windows 32 bit]: https://downloads.arduino.cc/arduino-cli/arduino-cli_latest_Windows_32bit.zip
[Mac OSX]: https://downloads.arduino.cc/arduino-cli/arduino-cli_latest_macOS_64bit.tar.gz

> **Deprecation notice**: links in the form
  `http://downloads.arduino.cc/arduino-cli/arduino-cli-latest-<platform>.tar.bz2`
  won’t be further updated. That URL will provide the version
  `0.3.7-alpha.preview`, regardless of further releases.

### Download a nightly build

These builds are generated everyday at 01:00 GMT from the `master` branch and
should be considered unstable. In order to get the latest nightly build
available for the supported platform, use the following links:

Platform  |                            |                            |
--------- | -------------------------- | -------------------------- |
Linux     | [Nightly Linux 32 bit]     | [Nightly Linux 64 bit]     |
Linux ARM | [Nightly Linux ARM 32 bit] | [Nightly Linux ARM 64 bit] |
Windows   | [Nightly Windows 32 bit]   | [Nightly Windows 64 bit]   |
Mac OSX   |                            | [Mac OSX]                  |

[Nightly Linux 64 bit]: https://downloads.arduino.cc/arduino-cli/nightly/arduino-cli_nightly-latest_Linux_64bit.tar.gz
[Nightly Windows 32 bit]: https://downloads.arduino.cc/arduino-cli/nightly/arduino-cli_nightly-latest_Linux_32bit.tar.gz
[Nightly Linux ARM 64 bit]: https://downloads.arduino.cc/arduino-cli/nightly/arduino-cli_nightly-latest_Linux_ARM64.tar.gz
[Nightly Linux ARM 32 bit]: https://downloads.arduino.cc/arduino-cli/nightly/arduino-cli_nightly-latest_Linux_ARMv7.tar.gz
[Nightly Windows 64 bit]: https://downloads.arduino.cc/arduino-cli/nightly/arduino-cli_nightly-latest_Windows_64bit.zip
[Nightly Windows 32 bit]: https://downloads.arduino.cc/arduino-cli/nightly/arduino-cli_nightly-latest_Windows_32bit.zip
[Nightly Mac OSX]: https://downloads.arduino.cc/arduino-cli/nightly/arduino-cli_nightly-latest_macOS_64bit.tar.gz

> These links return a `302: Found` response, redirecting to latest
  generated builds by replacing `latest` with the latest available build
  date, using the format YYYYMMDD (i.e for 2019/Aug/06 `latest` is
  replaced with `20190806` )

Checksums for the nightly builds are available at
`https://downloads.arduino.cc/arduino-cli/nightly/nightly-<DATE>-checksums.txt`

Once downloaded, extract the executable `arduino-cli` into a directory
which is in your ``PATH``.

### Build from source

If you’re familiar with Golang or if you want to contribute to the
project, you will probably build the `arduino-cli` locally with your
Go toolchain. Please refer to the [contributing] document for setup instructions.

If you don’t have a working Golang environment or if you want to build
`arduino-cli` targeting different platforms, you can use Docker to get
a binary directly from sources. From the project folder run:

```sh
docker run -v $PWD:/arduino-cli -w /arduino-cli -e PACKAGE_NAME_PREFIX='snapshot' arduino/arduino-cli:builder-1 goreleaser --rm-dist --snapshot --skip-publish
```

Once the build is over, you will find a `./dist/` folder containing the packages
built out of the current source tree.

## Getting started

`arduino-cli` is a container of commands and each command has its own
dedicated help text that can be shown with the `help` command like this:

```console
$ arduino-cli help core
Arduino Core operations.

Usage:
    arduino-cli core [command]

Examples:
    ./arduino-cli core update-index

Available Commands:
    download     Downloads one or more cores and corresponding tool dependencies.
    install      Installs one or more cores and corresponding tool dependencies.
    list         Shows the list of installed platforms.
    search       Search for a core in the package index.
    uninstall    Uninstalls one or more cores and corresponding tool dependencies if no more used.
    update-index Updates the index of cores.
    upgrade      Upgrades one or all installed platforms to the latest version.

Flags:
    -h, --help   help for core

Global Flags:
        --additional-urls strings   Additional URLs for the board manager.
        --config-file string        The custom config file (if not specified the default will be used).
        --format string             The output format, can be [text|json]. (default "text")
        --log-file string           Path to the file where logs will be written.
        --log-format string         The output format for the logs, can be [text|json].
        --log-level string          Messages with this level and above will be logged.
    -v, --verbose                   Print the logs on the standard output.

Use "arduino-cli core [command] --help" for more information about a command.
```

Follow the [Getting started guide](/getting-started/) to see how to use the most
common CLI commands available.


[contributing]: https://github.com/arduino/arduino-cli/blob/master/CONTRIBUTING.md
