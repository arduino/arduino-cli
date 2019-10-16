# arduino-cli

![Workflow Status](https://github.com/Arduino/arduino-cli/workflows/test/badge.svg)
[![codecov](https://codecov.io/gh/arduino/arduino-cli/branch/master/graph/badge.svg)](https://codecov.io/gh/arduino/arduino-cli)

`arduino-cli` is an all-in-one solution that provides builder, boards/library manager, uploader,
discovery and many other tools needed to use any Arduino compatible board and platforms.

This software is currently under active development: anything can change at any time, API and UI
must be considered unstable.

## How to contribute

Contributions are welcome!

Please read the document [How to contribute](CONTRIBUTING.md) which will guide you through how to
build the source code, run the tests, and contribute your changes to the project.

:sparkles: Thanks to all our [contributors](https://github.com/arduino/arduino-cli/graphs/contributors)! :sparkles:

## How to install

### Get the latest package

You have several options to install the latest version of the Arduino CLI
on your system.

#### Install via Homebrew (macOS/Linux)

The Arduino CLI is available as a Homebrew formula since version `0.5.0`:

```console
brew update
brew install arduino-cli
```

#### Use the install script

The easiest way to get the latest version of `arduino-cli` on any supported
platform is using the `install.sh` script:

```console
curl -fsSL https://raw.githubusercontent.com/arduino/arduino-cli/master/install.sh | sh
```

The script will install `arduino-cli` at `$PWD/bin`, if you want to target a different directory,
for example `~/local/bin`, set the `BINDIR` environment variable like this:

```console
curl -fsSL https://raw.githubusercontent.com/arduino/arduino-cli/master/install.sh | BINDIR=~/local/bin sh
```

#### Download the latest packages from Arduino CDN

In order to get the latest stable release for your platform you can use the
following links:

- [Linux 64 bit](https://downloads.arduino.cc/arduino-cli/arduino-cli_latest_Linux_64bit.tar.gz)
- [Linux 32 bit](https://downloads.arduino.cc/arduino-cli/arduino-cli_latest_Linux_32bit.tar.gz)
- [Linux ARM 64 bit](https://downloads.arduino.cc/arduino-cli/arduino-cli_latest_Linux_ARM64.tar.gz)
- [Linux ARM 32 bit](https://downloads.arduino.cc/arduino-cli/arduino-cli_latest_Linux_ARMv7.tar.gz)
- [Windows 64 bit](https://downloads.arduino.cc/arduino-cli/arduino-cli_latest_Windows_64bit.zip)
- [Windows 32 bit](https://downloads.arduino.cc/arduino-cli/arduino-cli_latest_Windows_32bit.zip)
- [Mac OSX](https://downloads.arduino.cc/arduino-cli/arduino-cli_latest_macOS_64bit.tar.gz)

These links return a `302: Found` response, redirecting to latest generated builds by replacing `latest` with the latest
available stable release. Once downloaded, place the executable `arduino-cli` into a directory which is in your `PATH`.

**Deprecation notice:** Links in the form `http://downloads.arduino.cc/arduino-cli/arduino-cli-latest-<platform>.tar.bz2`
won't be further updated. That URL will provide arduino-cli 0.3.7-alpha.preview, regardless of further releases.

#### Download the latest package from the release page on GitHub

Alternatively you can download one of the pre-built binaries for the supported
platforms from the
[release page](https://github.com/arduino/arduino-cli/releases). Once downloaded,
place the executable `arduino-cli` into a directory which is in your `PATH`.

### Get a nightly build

These builds are generated once a day from `master` branch starting at 01:00 GMT.
In order to get the latest nightly build for your platform, use the following links:

- [Linux 64 bit](https://downloads.arduino.cc/arduino-cli/nightly/arduino-cli_nightly-latest_Linux_64bit.tar.gz)
- [Linux 32 bit](https://downloads.arduino.cc/arduino-cli/nightly/arduino-cli_nightly-latest_Linux_32bit.tar.gz)
- [Linux ARM 64 bit](https://downloads.arduino.cc/arduino-cli/nightly/arduino-cli_nightly-latest_Linux_ARM64.tar.gz)
- [Linux ARM 32 bit](https://downloads.arduino.cc/arduino-cli/nightly/arduino-cli_nightly-latest_Linux_ARMv7.tar.gz)
- [Windows 64 bit](https://downloads.arduino.cc/arduino-cli/nightly/arduino-cli_nightly-latest_Windows_64bit.zip)
- [Windows 32 bit](https://downloads.arduino.cc/arduino-cli/nightly/arduino-cli_nightly-latest_Windows_32bit.zip)
- [Mac OSX](https://downloads.arduino.cc/arduino-cli/nightly/arduino-cli_nightly-latest_macOS_64bit.tar.gz)

These links return a `302: Found` response, redirecting to latest generated builds by replacing `latest` with the latest
available build date, using the format YYYYMMDD (i.e for 2019/Aug/06 `latest` is replaced with `20190806` )

Checksums for the nightly builds are available at
`https://downloads.arduino.cc/arduino-cli/nightly/nightly-<DATE>-checksums.txt`

Once downloaded, place the executable `arduino-cli` into a directory which is in your `PATH`.

### Build from source using Docker

If you don't have a working Golang environment or if you want to build
`arduino-cli` targeting different platforms, you can use Docker to get a binary
directly from sources. From the project folder run:

```console
docker run -v $PWD:/arduino-cli -w /arduino-cli -e PACKAGE_NAME_PREFIX='snapshot' arduino/arduino-cli:builder-1 goreleaser --rm-dist --snapshot --skip-publish
```

Once the build is over, you will find a `./dist/` folder containing the packages built out of
the current source tree.

### Build from source

If you're familiar with Golang or if you want to contribute to the project, you will probably
build the `arduino-cli` locally with your Go compiler. Please refer to the
[contributing](CONTRIBUTING.md) doc for setup instructions.

## Getting Started

The goal of the Arduino CLI is to be used by either including it in Makefile or in any kind of
script for the Command Line. The Arduino CLI aims to replace the majority of features the Arduino
IDE has without the graphical UI.

### Step 1. Create a new sketch

The command will create a new empty sketch named `MyFirstSketch` in the current directory

```console
$ arduino-cli sketch new MyFirstSketch
Sketch created in: /home/luca/MyFirstSketch

$ cat /home/luca/MyFirstSketch/MyFirstSketch.ino
void setup() {
}

void loop() {
}
```

### Step 2. Modify your sketch

Use your favourite file editor or IDE to modify the .ino file, in this example
under: `$HOME/MyFirstSketch/MyFirstSketch.ino`
and change the file to look like this one:

```C
void setup() {
  pinMode(LED_BUILTIN, OUTPUT);
}

void loop() {
  digitalWrite(LED_BUILTIN, HIGH);
  delay(1000);
  digitalWrite(LED_BUILTIN, LOW);
  delay(1000);
}
```

### Step 3. Connect the board to your PC

If you are running a fresh install of the arduino-cli you probably need to update the platform
indexes by running:

```console
$ arduino-cli core update-index
Updating index: package_index.json downloaded
```

Now, just connect the board to your PCs by using the USB cable. In this example we will use the
MKR1000 board:

```console
$ arduino-cli board list
Port         Type              Board Name              FQBN                 Core
/dev/ttyACM1 Serial Port (USB) Arduino/Genuino MKR1000 arduino:samd:mkr1000 arduino:samd
```

the board has been discovered but we need the correct core to program it, let's
install it!

### Step 4. Install the core for your board

From the output of the `board list` command, the right platform for the Arduino
MKR1000 is `arduino:samd`, we can install it with:

```console
$ arduino-cli core install arduino:samd
Downloading tools...
arduino:arm-none-eabi-gcc@4.8.3-2014q1 downloaded
arduino:bossac@1.7.0 downloaded
arduino:openocd@0.9.0-arduino6-static downloaded
arduino:CMSIS@4.5.0 downloaded
arduino:CMSIS-Atmel@1.1.0 downloaded
arduino:arduinoOTA@1.2.0 downloaded
Downloading cores...
arduino:samd@1.6.19 downloaded
Installing tools...
Installing platforms...
Results:
arduino:samd@1.6.19 - Installed
arduino:arm-none-eabi-gcc@4.8.3-2014q1 - Installed
arduino:bossac@1.7.0 - Installed
arduino:openocd@0.9.0-arduino6-static - Installed
arduino:CMSIS@4.5.0 - Installed
arduino:CMSIS-Atmel@1.1.0 - Installed
arduino:arduinoOTA@1.2.0 - Installed
```

Now verify we have installed the core properly by running

```console
$ arduino-cli core list
ID              Installed       Latest  Name
arduino:samd    1.6.19          1.6.19  Arduino SAMD Boards (32-bits ARM Cortex-M0+)
```

If the board is not detected for any reason, you can list all the supported boards
with `arduino-cli board listall` and also search for a specific board:

```console
$ arduino-cli board listall mkr
Board Name              FQBN
Arduino MKR FOX 1200    arduino:samd:mkrfox1200
Arduino MKR GSM 1400    arduino:samd:mkrgsm1400
Arduino MKR WAN 1300    arduino:samd:mkrwan1300
Arduino MKR WiFi 1010   arduino:samd:mkrwifi1010
Arduino MKRZERO         arduino:samd:mkrzero
Arduino/Genuino MKR1000 arduino:samd:mkr1000
```

Great! Now we are ready to compile and upload the sketch.

#### Adding 3rd party cores

To add 3rd party core packages add a link of the additional package to the file `arduino-cli.yaml`

If you want to add the ESP8266 core, for example:

```yaml
board_manager:
  additional_urls:
    - http://arduino.esp8266.com/stable/package_esp8266com_index.json
```

And then run:

```console
$ arduino-cli core update-index
Updating index: package_index.json downloaded
Updating index: package_esp8266com_index.json downloaded
Updating index: package_index.json downloaded

$ arduino-cli core search esp8266
ID              Version Name
esp8266:esp8266 2.5.2   esp8266
```

Alternatively, you can pass the `--additional-urls` to any command involving the additional cores:

```console
$ arduino-cli core update-index --additional-urls http://arduino.esp8266.com/stable/package_esp8266com_index.json
$
$ arduino-cli core search esp8266 --additional-urls http://arduino.esp8266.com/stable/package_esp8266com_index.json
ID              Version Name
esp8266:esp8266 2.5.2   esp8266
```

### Step 5. Compile the sketch

To compile the sketch we have to run the `compile` command with the proper FQBN we just got in the
previous command.

```console
$ arduino-cli compile --fqbn arduino:samd:mkr1000 MyFirstSketch
Sketch uses 9600 bytes (3%) of program storage space. Maximum is 262144 bytes.
```

### Step 6. Upload your sketch

We can finally upload the sketch and see our board blinking, we now have to specify the serial port
used by our board other than the FQBN:

```console
$ arduino-cli upload -p /dev/ttyACM0 --fqbn arduino:samd:mkr1000 MyFirstSketch
No new serial port detected.
Atmel SMART device 0x10010005 found
Device       : ATSAMD21G18A
Chip ID      : 10010005
Version      : v2.0 [Arduino:XYZ] Dec 20 2016 15:36:43
Address      : 8192
Pages        : 3968
Page Size    : 64 bytes
Total Size   : 248KB
Planes       : 1
Lock Regions : 16
Locked       : none
Security     : false
Boot Flash   : true
BOD          : true
BOR          : true
Arduino      : FAST_CHIP_ERASE
Arduino      : FAST_MULTI_PAGE_WRITE
Arduino      : CAN_CHECKSUM_MEMORY_BUFFER
Erase flash
done in 0.784 seconds

Write 9856 bytes to flash (154 pages)
[==============================] 100% (154/154 pages)
done in 0.069 seconds

Verify 9856 bytes of flash with checksum.
Verify successful
done in 0.009 seconds
CPU reset.
```

### Step 7. Add libraries

Now we can try to add a useful library to our sketch. We can at first look at the name of a library,
our favourite one is the wifi101, here the command to get more info:

```console
$ arduino-cli lib search wifi101
Name: "WiFi101OTA"
  Author:  Arduino
  Maintainer:  Arduino <info@arduino.cc>
  Sentence:  Update sketches to your board over WiFi
  Paragraph:  Requires an SD card and SAMD board
  Website:  http://www.arduino.cc/en/Reference/WiFi101OTA
  Category:  Other
  Architecture:  samd
  Types:  Arduino
  Versions:  [1.0.2, 1.0.0, 1.0.1]
Name: "WiFi101"
  Author:  Arduino
  Maintainer:  Arduino <info@arduino.cc>
  Sentence:  Network driver for ATMEL WINC1500 module (used on Arduino/Genuino Wifi Shield 101 and MKR1000 boards)
  Paragraph:  This library implements a network driver for devices based on the ATMEL WINC1500 wifi module
  Website:  http://www.arduino.cc/en/Reference/WiFi101
  Category:  Communication
  Architecture:  *
  Types:  Arduino
  Versions:  [0.5.0, 0.6.0, 0.10.0, 0.11.0, 0.11.1, 0.11.2, 0.12.0, 0.15.2, 0.8.0, 0.9.0, 0.12.1, 0.14.1, 0.14.4, 0.14.5, 0.15.1, 0.7.0, 0.14.0, 0.14.2, 0.14.3, 0.9.1, 0.13.0, 0.15.0, 0.5.1]
```

We are now ready to install it! Please be sure to use the full name of the lib as specified in the
"Name:" section previously seen:

```console
$ arduino-cli lib install "WiFi101"
Downloading libraries...
WiFi101@0.15.2 downloaded
Installed WiFi101@0.15.2
```

## Inline Help

`arduino-cli` is a container of commands, to see the full list just run:

```console
$ arduino-cli
Arduino Command Line Interface (arduino-cli).

Usage:
  arduino-cli [command]

Examples:
arduino <command> [flags...]

Available Commands:
  board         Arduino board commands.
  compile       Compiles Arduino sketches.
  config        Arduino Configuration Commands.
  core          Arduino Core operations.
  help          Help about any command
  lib           Arduino commands about libraries.
  sketch        Arduino CLI Sketch Commands.
  upload        Upload Arduino sketches.
  version       Shows version number of Arduino CLI.
  ....
```

Each command has his own specific help that can be obtained with the `help` command, for example:

```console
$ arduino-cli help core
Arduino Core operations.

Usage:
  arduino-cli core [command]

Examples:
arduino-cli core update-index # to update the package index file.

Available Commands:
  download     Downloads one or more cores and corresponding tool dependencies.
  install      Installs one or more cores and corresponding tool dependencies.
  list         Shows the list of installed cores.
  update-index Updates the index of cores.

Flags:
  -h, --help   help for core

Global Flags:
      --config-file string   The custom config file (if not specified the default one will be used).
      --debug                Enables debug output (super verbose, used to debug the CLI).
      --format string        The output format, can be [text|json]. (default "text")

Use "arduino-cli core [command] --help" for more information about a command.
```

## FAQ

### Why the Arduino Uno/Mega/Duemilanove is not detected when I run `arduino-cli board list`?

Because:

- Your board is a cheaper clone, or
- It mounts a USB2Serial converter like FT232 or CH320: these chips always reports the same USB
VID/PID to the operating system, so the only thing that we know is that the board mounts that
specific USB2Serial chip, but we don't know which board is.

### What is the core for the Uno/Mega/Nano/Duemilanove?

`arduino:avr`

### What is the FQBN for ...?

- Arduino UNO: `arduino:avr:uno`
- Arduino Mega: `arduino:avr:mega`
- Arduino Nano: `arduino:avr:nano` or `arduino:avr:nano:cpu=atmega328old` if you have the old bootloader

### How can I find the core/FQBN for a board?

Update the core index to have latest boards informations:

```console
$ arduino-cli core update-index
Updating index: package_index.json downloaded
```

See: <https://github.com/arduino/arduino-cli#step-4-find-and-install-the-right-core>

Further help can be found in [this comment](https://github.com/arduino/arduino-cli/issues/138#issuecomment-459169051) in [#138](https://github.com/arduino/arduino-cli/issues/138).

For a deeper understanding of how FQBN works, you should understand Arduino Hardware specification.
You can find more information in this [arduino/Arduino wiki page](https://github.com/arduino/Arduino/wiki/Arduino-IDE-1.5-3rd-party-Hardware-specification)

## Using the gRPC interface

The [client_example](./client_example) folder contains a sample program that
shows how to use gRPC interface of the CLI.
