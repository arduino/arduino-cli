### Install via Homebrew (macOS/Linux)

The Arduino CLI is available as a Homebrew formula since version `0.5.0`:

```sh
brew update
brew install arduino-cli
```

#### Command line completion

[Command line completion](command-line-completion.md#brew) files are already bundled in the homebrew installation.

### Use the install script

The script requires `sh`. This is always available on Linux and macOS. `sh` is not available by default on Windows. The
script may be run on Windows by installing [Git for Windows], then running it from Git Bash.

This script will install the latest version of Arduino CLI to `$PWD/bin`:

```sh
curl -fsSL https://raw.githubusercontent.com/arduino/arduino-cli/master/install.sh | sh
```

If you want to target a different directory, for example `~/local/bin`, set the `BINDIR` environment variable like this:

```sh
curl -fsSL https://raw.githubusercontent.com/arduino/arduino-cli/master/install.sh | BINDIR=~/local/bin sh
```

If you would like to use the `arduino-cli` command from any location, install Arduino CLI to a directory already in your
`PATH` or add the Arduino CLI installation path to your `PATH` environment variable.

If you want to download a specific arduino-cli version, for example `0.9.0`, pass the version number as a parameter like
this:

```sh
curl -fsSL https://raw.githubusercontent.com/arduino/arduino-cli/master/install.sh | sh -s 0.9.0
```

### Download

Pre-built binaries for all the supported platforms are available for download from the links below.

If you would like to use the `arduino-cli` command from any location, extract the downloaded file to a directory already
in your `PATH` or add the Arduino CLI installation path to your `PATH` environment variable.

#### Latest packages

| Platform  |                    |                    |
| --------- | ------------------ | ------------------ |
| Linux     | [Linux 32 bit]     | [Linux 64 bit]     |
| Linux ARM | [Linux ARM 32 bit] | [Linux ARM 64 bit] |
| Windows   | [Windows 32 bit]   | [Windows 64 bit]   |
| Mac OSX   |                    | [Mac OSX]          |

[linux 64 bit]: https://downloads.arduino.cc/arduino-cli/arduino-cli_latest_Linux_64bit.tar.gz
[linux 32 bit]: https://downloads.arduino.cc/arduino-cli/arduino-cli_latest_Linux_32bit.tar.gz
[linux arm 64 bit]: https://downloads.arduino.cc/arduino-cli/arduino-cli_latest_Linux_ARM64.tar.gz
[linux arm 32 bit]: https://downloads.arduino.cc/arduino-cli/arduino-cli_latest_Linux_ARMv7.tar.gz
[windows 64 bit]: https://downloads.arduino.cc/arduino-cli/arduino-cli_latest_Windows_64bit.zip
[windows 32 bit]: https://downloads.arduino.cc/arduino-cli/arduino-cli_latest_Windows_32bit.zip
[mac osx]: https://downloads.arduino.cc/arduino-cli/arduino-cli_latest_macOS_64bit.tar.gz

> **Deprecation notice**: links in the form
> `http://downloads.arduino.cc/arduino-cli/arduino-cli-latest-<platform>.tar.bz2` won’t be further updated. That URL
> will provide the version `0.3.7-alpha.preview`, regardless of further releases.

#### Previous versions

These are available from the [releases page](https://github.com/arduino/arduino-cli/releases)

#### Nightly builds

These builds are generated every day at 01:00 GMT from the `master` branch and should be considered unstable. In order
to get the latest nightly build available for the supported platform, use the following links:

| Platform  |                            |                            |
| --------- | -------------------------- | -------------------------- |
| Linux     | [Nightly Linux 32 bit]     | [Nightly Linux 64 bit]     |
| Linux ARM | [Nightly Linux ARM 32 bit] | [Nightly Linux ARM 64 bit] |
| Windows   | [Nightly Windows 32 bit]   | [Nightly Windows 64 bit]   |
| Mac OSX   |                            | [Nightly Mac OSX]          |

[nightly linux 64 bit]: https://downloads.arduino.cc/arduino-cli/nightly/arduino-cli_nightly-latest_Linux_64bit.tar.gz
[nightly linux 32 bit]: https://downloads.arduino.cc/arduino-cli/nightly/arduino-cli_nightly-latest_Linux_32bit.tar.gz

<!-- prettier-ignore -->
[nightly linux arm 64 bit]: https://downloads.arduino.cc/arduino-cli/nightly/arduino-cli_nightly-latest_Linux_ARM64.tar.gz

<!-- prettier-ignore -->
[nightly linux arm 32 bit]: https://downloads.arduino.cc/arduino-cli/nightly/arduino-cli_nightly-latest_Linux_ARMv7.tar.gz
[nightly windows 64 bit]: https://downloads.arduino.cc/arduino-cli/nightly/arduino-cli_nightly-latest_Windows_64bit.zip
[nightly windows 32 bit]: https://downloads.arduino.cc/arduino-cli/nightly/arduino-cli_nightly-latest_Windows_32bit.zip
[nightly mac osx]: https://downloads.arduino.cc/arduino-cli/nightly/arduino-cli_nightly-latest_macOS_64bit.tar.gz

> These links return a `302: Found` response, redirecting to latest generated builds by replacing `latest` with the
> latest available build date, using the format YYYYMMDD (i.e for 2019/Aug/06 `latest` is replaced with `20190806` )

Checksums for the nightly builds are available at
`https://downloads.arduino.cc/arduino-cli/nightly/nightly-<DATE>-checksums.txt`

### Build from source

If you’re familiar with Golang or if you want to contribute to the project, you will probably build the `arduino-cli`
locally with your Go toolchain. Please refer to the [CONTRIBUTING] document for setup instructions.

If you don’t have a working Golang environment or if you want to build `arduino-cli` targeting different platforms, you
can use [Task][task-site] to get a binary directly from sources. From the project folder run:

```sh
task dist:all
```

Once the build is over, you will find a `./dist/` folder containing the packages built out of the current source tree.

[git for windows]: https://gitforwindows.org/
[contributing]: CONTRIBUTING.md
[task-site]: https://taskfile.dev/#/installation
