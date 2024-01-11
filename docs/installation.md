<!-- Source: https://github.com/arduino/tooling-project-assets/blob/main/other/installation-script/installation.md -->

Several options are available for installation of Arduino CLI. Instructions for each are provided below:

## Install via Homebrew (macOS/Linux)

The Arduino CLI is available as a Homebrew formula since version `0.5.0`:

```sh
brew update
brew install arduino-cli
```

### Command line completion

[Command line completion](command-line-completion.md#brew) files are already bundled in the homebrew installation.

## Use the install script

The script requires `sh`, which is always available on Linux and macOS. `sh` is not available by default on Windows,
though it is available as part of [Git for Windows](https://gitforwindows.org/) (Git Bash). If you don't have `sh`
available, use the ["Download" installation option](#download).

This script will install the latest version of Arduino CLI to `$PWD/bin`:

```
curl -fsSL https://raw.githubusercontent.com/arduino/arduino-cli/master/install.sh | sh
```

If you want to target a different directory, for example `~/local/bin`, set the `BINDIR` environment variable like this:

```
curl -fsSL https://raw.githubusercontent.com/arduino/arduino-cli/master/install.sh | BINDIR=~/local/bin sh
```

If you would like to use the `arduino-cli` command from any location, install Arduino CLI to a directory already in your
[`PATH`](https://en.wikipedia.org/wiki/PATH%5F%28variable%29) or add the Arduino CLI installation path to your `PATH`
environment variable.

If you want to download a specific Arduino CLI version, for example `0.9.0` or `nightly-latest`, pass the version number
as a parameter like this:

```
curl -fsSL https://raw.githubusercontent.com/arduino/arduino-cli/master/install.sh | sh -s 0.9.0
```

Arduino CLI checks for new releases every 24 hours. If you don't like this behaviour you can disable it by setting the
[`updater.enable_notification` config](configuration.md#configuration-keys) or the
[env var `ARDUINO_UPDATER_ENABLE_NOTIFICATION`](configuration.md#environment-variables) to `false`.

## Download

Pre-built binaries for all the supported platforms are available for download from the links below.

If you would like to use the `arduino-cli` command from any location, extract the downloaded file to a directory already
in your [`PATH`](https://en.wikipedia.org/wiki/PATH%5F%28variable%29) or add the Arduino CLI installation path to your
`PATH` environment variable.

### Latest release

| Platform    |                      |                        |
| ----------- | -------------------- | ---------------------- |
| Linux       | [32 bit][linux32]    | [64 bit][linux64]      |
| Linux ARM   | [32 bit][linuxarm32] | [64 bit][linuxarm64]   |
| Linux ARMv6 | [32 bit][linuxarmv6] |                        |
| Windows exe | [32 bit][windows32]  | [64 bit][windows64]    |
| Windows msi |                      | [64 bit][windowsmsi64] |
| macOS       |                      | [64 bit][macos64]      |
| macOS ARM   |                      | [64 bit][macosarm64]   |

[linux64]: https://downloads.arduino.cc/arduino-cli/arduino-cli_latest_Linux_64bit.tar.gz
[linux32]: https://downloads.arduino.cc/arduino-cli/arduino-cli_latest_Linux_32bit.tar.gz
[linuxarm64]: https://downloads.arduino.cc/arduino-cli/arduino-cli_latest_Linux_ARM64.tar.gz
[linuxarm32]: https://downloads.arduino.cc/arduino-cli/arduino-cli_latest_Linux_ARMv7.tar.gz
[linuxarmv6]: https://downloads.arduino.cc/arduino-cli/arduino-cli_latest_Linux_ARMv6.tar.gz
[windows64]: https://downloads.arduino.cc/arduino-cli/arduino-cli_latest_Windows_64bit.zip
[windows32]: https://downloads.arduino.cc/arduino-cli/arduino-cli_latest_Windows_32bit.zip
[windowsmsi64]: https://downloads.arduino.cc/arduino-cli/arduino-cli_latest_Windows_64bit.msi
[macos64]: https://downloads.arduino.cc/arduino-cli/arduino-cli_latest_macOS_64bit.tar.gz
[macosarm64]: https://downloads.arduino.cc/arduino-cli/arduino-cli_latest_macOS_ARM64.tar.gz

### Previous versions

These are available from the "Assets" sections on the [releases page](https://github.com/arduino/arduino-cli/releases).

### Nightly builds

These builds are generated every day at 01:00 GMT from the `master` branch and should be considered unstable. In order
to get the latest nightly build available for the supported platform, use the following links:

| Platform    |                              |                                |
| ----------- | ---------------------------- | ------------------------------ |
| Linux       | [32 bit][linux32-nightly]    | [64 bit][linux64-nightly]      |
| Linux ARM   | [32 bit][linuxarm32-nightly] | [64 bit][linuxarm64-nightly]   |
| Linux ARMv6 | [32 bit][linuxarmv6-nightly] |                                |
| Windows exe | [32 bit][windows32-nightly]  | [64 bit][windows64-nightly]    |
| Windows msi |                              | [64 bit][windowsmsi64-nightly] |
| macOS       |                              | [64 bit][macos64-nightly]      |
| macOS ARM   |                              | [64 bit][macosarm64-nightly]   |

[linux64-nightly]: https://downloads.arduino.cc/arduino-cli/nightly/arduino-cli_nightly-latest_Linux_64bit.tar.gz
[linux32-nightly]: https://downloads.arduino.cc/arduino-cli/nightly/arduino-cli_nightly-latest_Linux_32bit.tar.gz
[linuxarm64-nightly]: https://downloads.arduino.cc/arduino-cli/nightly/arduino-cli_nightly-latest_Linux_ARM64.tar.gz
[linuxarm32-nightly]: https://downloads.arduino.cc/arduino-cli/nightly/arduino-cli_nightly-latest_Linux_ARMv7.tar.gz
[linuxarmv6-nightly]: https://downloads.arduino.cc/arduino-cli/nightly/arduino-cli_nightly-latest_Linux_ARMv6.tar.gz
[windows64-nightly]: https://downloads.arduino.cc/arduino-cli/nightly/arduino-cli_nightly-latest_Windows_64bit.zip
[windows32-nightly]: https://downloads.arduino.cc/arduino-cli/nightly/arduino-cli_nightly-latest_Windows_32bit.zip
[windowsmsi64-nightly]: https://downloads.arduino.cc/arduino-cli/nightly/arduino-cli_nightly-latest_Windows_64bit.msi
[macos64-nightly]: https://downloads.arduino.cc/arduino-cli/nightly/arduino-cli_nightly-latest_macOS_64bit.tar.gz
[macosarm64-nightly]: https://downloads.arduino.cc/arduino-cli/nightly/arduino-cli_nightly-latest_macOS_ARM64.tar.gz

> These links return a `302: Found` response, redirecting to latest generated builds by replacing `latest` with the
> latest available build date, using the format YYYYMMDD (i.e for 2019-08-06 `latest` is replaced with `20190806` )

Checksums for the nightly builds are available at
`https://downloads.arduino.cc/arduino-cli/nightly/nightly-<DATE>-checksums.txt`

## Build from source

If you're familiar with Golang or if you want to contribute to the project, you will probably build Arduino CLI locally
with your Go toolchain. See the ["How to contribute"](CONTRIBUTING.md#building-the-source-code) page for instructions.
