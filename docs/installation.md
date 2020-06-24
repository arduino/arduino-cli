### Install via Homebrew (macOS/Linux)

The Arduino CLI is available as a Homebrew formula since version
`0.5.0`:

```sh
brew update
brew install arduino-cli
```

### Use the install script

The script requires `sh`. This is always available on Linux and macOS. `sh` is
not available by default on Windows. The script may be run on Windows by
installing [Git for Windows], then running it from Git Bash.

This script will install the latest version of Arduino CLI to `$PWD/bin`:

```sh
curl -fsSL https://raw.githubusercontent.com/arduino/arduino-cli/master/install.sh | sh
```

If you want to target a different directory, for example `~/local/bin`, set the
`BINDIR` environment variable like this:

```sh
curl -fsSL https://raw.githubusercontent.com/arduino/arduino-cli/master/install.sh | BINDIR=~/local/bin sh
```

If you would like to use the `arduino-cli` command from any location, install
Arduino CLI to a directory already in your `PATH` or add the Arduino CLI
installation path to your `PATH` environment variable.

### Download

Pre-built binaries for all the supported platforms are available for download
from the links below.

If you would like to use the `arduino-cli` command from any location, extract
the downloaded file to a directory already in your `PATH` or add the Arduino CLI
installation path to your `PATH` environment variable.

#### Latest packages

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

#### Previous versions

These are available from the [releases page](https://github.com/arduino/arduino-cli/releases)

#### Nightly builds

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
[Nightly Linux 32 bit]: https://downloads.arduino.cc/arduino-cli/nightly/arduino-cli_nightly-latest_Linux_32bit.tar.gz
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

### Build from source

If you’re familiar with Golang or if you want to contribute to the
project, you will probably build the `arduino-cli` locally with your
Go toolchain. Please refer to the [CONTRIBUTING] document for setup instructions.

If you don’t have a working Golang environment or if you want to build
`arduino-cli` targeting different platforms, you can use Docker to get
a binary directly from sources. From the project folder run:

```sh
docker run -v $PWD:/arduino-cli -w /arduino-cli -e PACKAGE_NAME_PREFIX='snapshot' arduino/arduino-cli:builder-1 goreleaser --rm-dist --snapshot --skip-publish
```

Once the build is over, you will find a `./dist/` folder containing the packages
built out of the current source tree.

[Git for Windows]: https://gitforwindows.org/
[CONTRIBUTING]: CONTRIBUTING.md
