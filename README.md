# arduino-cli

[![Build Status](https://cloud.drone.io/api/badges/arduino/arduino-cli/status.svg)](https://cloud.drone.io/arduino/arduino-cli)
[![Build status](https://ci.appveyor.com/api/projects/status/obclpbgsafum2wml/branch/master?svg=true)](https://ci.appveyor.com/project/Arduino/arduino-cli/branch/master)

`arduino-cli` is an all-in-one solution that provides builder, boards/library manager, uploader,
discovery and many other tools needed to use any Arduino compatible board and platforms.

This software is currently in alpha state: new features will be added and some may be changed.

It will be soon used as a building block in the Arduino IDE and Arduino Create.

## How to contribute

Contributions are welcome!

Please read the document [How to contribute](CONTRIBUTING.md) which will guide you through how to
build the source code, run the tests, and contribute your changes to the project.

## How to install

### Download the latest stable release

This is **not yet available** until the first stable version is released.

#### Download the latest unstable "alpha" preview

Please note that these are **preview** builds, they may have bugs, some features may not work or may
be changed without notice, the latest preview version is `0.3.7-alpha.preview`:

- [Linux 64 bit](https://downloads.arduino.cc/arduino-cli/arduino-cli-latest-linux64.tar.bz2)
- [Linux 32 bit](https://downloads.arduino.cc/arduino-cli/arduino-cli-latest-linux32.tar.bz2)
- [Linux ARM 64 bit](https://downloads.arduino.cc/arduino-cli/arduino-cli-latest-linuxarm32.tar.bz2)
- [Linux ARM 32 bit](https://downloads.arduino.cc/arduino-cli/arduino-cli-latest-linuxarm64.tar.bz2)
- [Windows](https://downloads.arduino.cc/arduino-cli/arduino-cli-latest-windows.zip)
- [Mac OSX](https://downloads.arduino.cc/arduino-cli/arduino-cli-latest-macosx.zip)

Once downloaded, place the executable `arduino-cli` into a directory which is in your `PATH`
environment variable.

#### Download the nightly build

These builds are generated once a day from `master` branch starting at 23:00 UTC

- [Linux 64 bit](https://downloads.arduino.cc/arduino-cli/nightly/arduino-cli-nightly-latest-linux64.tar.bz2)
- [Linux 32 bit](https://downloads.arduino.cc/arduino-cli/nightly/arduino-cli-nightly-latest-linux32.tar.bz2)
- [Linux ARM 64 bit](https://downloads.arduino.cc/arduino-cli/nightly/arduino-cli-nightly-latest-linuxarm32.tar.bz2)
- [Linux ARM 32 bit](https://downloads.arduino.cc/arduino-cli/nightly/arduino-cli-nightly-latest-linuxarm64.tar.bz2)
- [Windows](https://downloads.arduino.cc/arduino-cli/nightly/arduino-cli-nightly-latest-windows.zip)
- [Mac OSX](https://downloads.arduino.cc/arduino-cli/nightly/arduino-cli-nightly-latest-macosx.zip)

Once downloaded, place the executable `arduino-cli` into a directory which is in your `PATH`
environment variable.

### Build the latest "bleeding-edge" from source

- You should have a recent Go compiler installed.
- Run `go get -u github.com/arduino/arduino-cli`
- The `arduino-cli` executable will be produced in `$GOPATH/bin/arduino-cli`

You may want to copy the executable into a directory which is in your `PATH` environment variable
(such as `/usr/local/bin/`).

## Usage

The goal of the Arduino CLI is to be used by either including it in Makefile or in any kind of
script for the Command Line. The Arduino CLI aims to replace the majority of features the Arduino
IDE has without the graphical UI.

## Getting Started

### Step 1. Create a new sketch

The command will create a new empty sketch named MyFirstSketch in the default directory under \$HOME/Arduino/

    $ arduino-cli sketch new MyFirstSketch
    Sketch created in: /home/luca/Arduino/MyFirstSketch

    $ cat /home/luca/Arduino/MyFirstSketch/MyFirstSketch.ino
    void setup() {
    }

    void loop() {
    }

### Step 2. Modify your sketch

Use your favourite file editor or IDE to modify the .ino file under: `$HOME/Arduino/MyFirstSketch/MyFirstSketch.ino`
and change the file to look like this one:

    void setup() {
      pinMode(LED_BUILTIN, OUTPUT);
    }

    void loop() {
      digitalWrite(LED_BUILTIN, HIGH);
      delay(1000);
      digitalWrite(LED_BUILTIN, LOW);
      delay(1000);
    }

### Step 3. Connect the board to your PC

If you are running a fresh install of the arduino-cli you probably need to update the platform
indexes by running:

    $ arduino-cli core update-index
    Updating index: package_index.json downloaded

Now, just connect the board to your PCs by using the USB cable. In this example we will use the
MKR1000 board:

    $ arduino-cli board list
    FQBN    Port            ID              Board Name
            /dev/ttyACM0    2341:804E       unknown

the board has been discovered but we do not have the correct core to program it yet.
Let's install it!

### Step 4. Find and install the right core

We have to look at the core available with the `core search` command. It will provide a list of
available cores matching the name arduino:

    $ arduino-cli core search arduino
    Searching for platforms matching 'arduino'

    ID              Version Installed       Name
    Intel:arc32     2.0.2   No              Intel Curie Boards
    arduino:avr     1.6.21  No              Arduino AVR Boards
    arduino:nrf52   1.0.2   No              Arduino nRF52 Boards
    arduino:sam     1.6.11  No              Arduino SAM Boards (32-bits ARM Cortex-M3)
    arduino:samd    1.6.18  No              Arduino SAMD Boards (32-bits ARM Cortex-M0+)
    arduino:stm32f4 1.0.1   No              Arduino STM32F4 Boards
    littleBits:avr  1.0.0   No              littleBits Arduino AVR Modules

If you're unsure you can try to refine the search with the board name

    $ arduino-cli core search mkr1000
    Searching for platforms matching 'mkr1000'

    ID              Version Installed   Name
    arduino:samd    1.6.19  No          Arduino SAMD Boards (32-bits ARM Cortex-M0+)

So, the right platform for the Arduino MKR1000 is arduino:samd, now we can install it

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

Now verify we have installed the core properly by running

    $ arduino-cli core list
    ID              Installed       Latest  Name
    arduino:samd    1.6.19          1.6.19  Arduino SAMD Boards (32-bits ARM Cortex-M0+)

We can finally check if the board is now recognized as a MKR1000

    $ arduino-cli board list
    FQBN                    Port            ID              Board Name
    arduino:samd:mkr1000    /dev/ttyACM0    2341:804E       Arduino/Genuino MKR1000

If the board is not detected for any reason, you can list all the supported boards
with `arduino-cli board listall` and also search for a specific board:

    $ arduino-cli board listall mkr
    Board Name              FQBN
    Arduino MKR FOX 1200    arduino:samd:mkrfox1200
    Arduino MKR GSM 1400    arduino:samd:mkrgsm1400
    Arduino MKR WAN 1300    arduino:samd:mkrwan1300
    Arduino MKR WiFi 1010   arduino:samd:mkrwifi1010
    Arduino MKRZERO         arduino:samd:mkrzero
    Arduino/Genuino MKR1000 arduino:samd:mkr1000

Great! Now we have the Board FQBN (Fully Qualified Board Name) `arduino:samd:mkr1000`
and the Board Name look good, we are ready to compile and upload the sketch

#### Adding 3rd party cores

To add 3rd party core packages add a link of the additional package to the file `arduino-cli.yaml`

If you want to add the ESP8266 core, for example:

    board_manager:
      additional_urls:
        - http://arduino.esp8266.com/stable/package_esp8266com_index.json

And then run:

    arduino-cli core update-index
    arduino-cli core install esp8266:esp8266

### Step 5. Compile the sketch

To compile the sketch we have to run the `compile` command with the proper FQBN we just got in the
previous command.

    $ arduino-cli compile --fqbn arduino:samd:mkr1000 Arduino/MyFirstSketch
    Sketch uses 9600 bytes (3%) of program storage space. Maximum is 262144 bytes.

### Step 6. Upload your sketch

We can finally upload the sketch and see our board blinking, we now have to specify the serial port
used by our board other than the FQBN:

    $ arduino-cli upload -p /dev/ttyACM0 --fqbn arduino:samd:mkr1000 Arduino/MyFirstSketch
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

### Step 7. Add libraries

Now we can try to add a useful library to our sketch. We can at first look at the name of a library,
our favourite one is the wifi101, here the command to get more info:

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

We are now ready to install it! Please be sure to use the full name of the lib as specified in the
"Name:" section previously seen:

    $ arduino-cli lib install "WiFi101"
    Downloading libraries...
    WiFi101@0.15.2 downloaded
    Installed WiFi101@0.15.2

## Inline Help

`arduino-cli` is a container of commands, to see the full list just run:

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

Each command has his own specific help that can be obtained with the `help` command, for example:

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

    $ arduino-cli core update-index
    Updating index: package_index.json downloaded

See: <https://github.com/arduino/arduino-cli#step-4-find-and-install-the-right-core>

Further help can be found in [this comment](https://github.com/arduino/arduino-cli/issues/138#issuecomment-459169051) in [#138](https://github.com/arduino/arduino-cli/issues/138).

For a deeper understanding of how FQBN works, you should understand Arduino Hardware specification.
You can find more information in this [arduino/Arduino wiki page](https://github.com/arduino/Arduino/wiki/Arduino-IDE-1.5-3rd-party-Hardware-specification)
