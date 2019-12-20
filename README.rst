arduino-cli
===========

|Tests passing| |Nightly build| |codecov|

``arduino-cli`` is an all-in-one solution that provides builder,
boards/library manager, uploader, discovery and many other tools needed
to use any Arduino compatible board and platforms.

  This software is currently under active development: anything can change
  at any time, API and UI must be considered unstable until we release version
  1.0.0.

.. contents:: **Table of Contents**
    :backlinks: none
    :depth: 2

How to contribute
-----------------

Contributions are welcome!

Please read the document `How to contribute <CONTRIBUTING.md>`__ which
will guide you through how to build the source code, run the tests, and
contribute your changes to the project.

`:sparkles:` Thanks to all our `contributors <https://github.com/arduino/arduino-cli/graphs/contributors>`__! `:sparkles:`

How to install
--------------

Get the latest version
~~~~~~~~~~~~~~~~~~~~~~

You have several options to install the latest version of the Arduino
CLI on your system.

Install via Homebrew (macOS/Linux)
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

The Arduino CLI is available as a Homebrew formula since version
``0.5.0``:

.. code:: console

   brew update
   brew install arduino-cli

Use the install script
^^^^^^^^^^^^^^^^^^^^^^

The easiest way to get the latest version of ``arduino-cli`` on any
supported platform is using the ``install.sh`` script:

.. code:: console

   curl -fsSL https://raw.githubusercontent.com/arduino/arduino-cli/master/install.sh | sh

The script will install ``arduino-cli`` at ``$PWD/bin``, if you want to
target a different directory, for example ``~/local/bin``, set the
``BINDIR`` environment variable like this:

.. code:: console

   curl -fsSL https://raw.githubusercontent.com/arduino/arduino-cli/master/install.sh | BINDIR=~/local/bin sh

Download the latest packages
^^^^^^^^^^^^^^^^^^^^^^^^^^^^

You can download the latest version of the pre-built binaries for the supported
platforms from the `release page <https://github.com/arduino/arduino-cli/releases>`__
or following the links in the following table. Once downloaded, extract the
binary ``arduino-cli`` into a directory which is in your ``PATH``.

+---------------+---------------------+---------------------+
| **Linux**     | `Linux 32 bit`_     | `Linux 64 bit`_     |
+---------------+---------------------+---------------------+
| **Linux ARM** | `Linux ARM 32 bit`_ | `Linux ARM 64 bit`_ |
+---------------+---------------------+---------------------+
| **Windows**   | `Windows 32 bit`_   | `Windows 64 bit`_   |
+---------------+---------------------+---------------------+
| **Mac OSX**   |                     | `Mac OSX`_          |
+---------------+---------------------+---------------------+

.. _`Linux 64 bit`: https://downloads.arduino.cc/arduino-cli/arduino-cli_latest_Linux_64bit.tar.gz
.. _`Linux 32 bit`: https://downloads.arduino.cc/arduino-cli/arduino-cli_latest_Linux_32bit.tar.gz
.. _`Linux ARM 64 bit`: https://downloads.arduino.cc/arduino-cli/arduino-cli_latest_Linux_ARM64.tar.gz
.. _`Linux ARM 32 bit`: https://downloads.arduino.cc/arduino-cli/arduino-cli_latest_Linux_ARMv7.tar.gz
.. _`Windows 64 bit`: https://downloads.arduino.cc/arduino-cli/arduino-cli_latest_Windows_64bit.zip
.. _`Windows 32 bit`: https://downloads.arduino.cc/arduino-cli/arduino-cli_latest_Windows_32bit.zip
.. _`Mac OSX`: https://downloads.arduino.cc/arduino-cli/arduino-cli_latest_macOS_64bit.tar.gz

  Deprecation notice: links in the form
  ``http://downloads.arduino.cc/arduino-cli/arduino-cli-latest-<platform>.tar.bz2``
  won’t be further updated. That URL will provide the version
  ``0.3.7-alpha.preview``, regardless of further releases.

Get a nightly build
~~~~~~~~~~~~~~~~~~~

These builds are generated everyday at 01:00 GMT from the ``master`` branch and
should be considered unstable. In order to get the latest nightly build
available for the supported platform, use the following links:

+---------------+-----------------------------+-----------------------------+
| **Linux**     | `Nightly Linux 32 bit`_     | `Nightly Linux 64 bit`_     |
+---------------+-----------------------------+-----------------------------+
| **Linux ARM** | `Nightly Linux ARM 32 bit`_ | `Nightly Linux ARM 64 bit`_ |
+---------------+-----------------------------+-----------------------------+
| **Windows**   | `Nightly Windows 32 bit`_   | `Nightly Windows 64 bit`_   |
+---------------+-----------------------------+-----------------------------+
| **Mac OSX**   |                             | `Nightly Mac OSX`_          |
+---------------+-----------------------------+-----------------------------+

.. _`Nightly Linux 64 bit`: https://downloads.arduino.cc/arduino-cli/nightly/arduino-cli_nightly-latest_Linux_64bit.tar.gz
.. _`Nightly Linux 32 bit`: https://downloads.arduino.cc/arduino-cli/nightly/arduino-cli_nightly-latest_Linux_32bit.tar.gz
.. _`Nightly Linux ARM 64 bit`: https://downloads.arduino.cc/arduino-cli/nightly/arduino-cli_nightly-latest_Linux_ARM64.tar.gz
.. _`Nightly Linux ARM 32 bit`: https://downloads.arduino.cc/arduino-cli/nightly/arduino-cli_nightly-latest_Linux_ARMv7.tar.gz
.. _`Nightly Windows 64 bit`: https://downloads.arduino.cc/arduino-cli/nightly/arduino-cli_nightly-latest_Windows_64bit.zip
.. _`Nightly Windows 32 bit`: https://downloads.arduino.cc/arduino-cli/nightly/arduino-cli_nightly-latest_Windows_32bit.zip
.. _`Nightly Mac OSX`: https://downloads.arduino.cc/arduino-cli/nightly/arduino-cli_nightly-latest_macOS_64bit.tar.gz

These links return a ``302: Found`` response, redirecting to latest
generated builds by replacing ``latest`` with the latest available build
date, using the format YYYYMMDD (i.e for 2019/Aug/06 ``latest`` is
replaced with ``20190806`` )

Checksums for the nightly builds are available at
``https://downloads.arduino.cc/arduino-cli/nightly/nightly-<DATE>-checksums.txt``

Once downloaded, extract the executable ``arduino-cli`` into a directory
which is in your ``PATH``.

Build from source
~~~~~~~~~~~~~~~~~

If you’re familiar with Golang or if you want to contribute to the
project, you will probably build the ``arduino-cli`` locally with your
Go compiler. Please refer to the `contributing <CONTRIBUTING.md>`__ doc
for setup instructions.

If you don’t have a working Golang environment or if you want to build
``arduino-cli`` targeting different platforms, you can use Docker to get
a binary directly from sources. From the project folder run:

.. code:: console

   docker run -v $PWD:/arduino-cli -w /arduino-cli -e PACKAGE_NAME_PREFIX='snapshot' arduino/arduino-cli:builder-1 goreleaser --rm-dist --snapshot --skip-publish

Once the build is over, you will find a ``./dist/`` folder containing
the packages built out of the current source tree.

How to use
----------

Despite there's no feature parity at the moment, Arduino CLI provides many of
the features you can find in the Arduino IDE, let's see some examples.

Create a configuration file
~~~~~~~~~~~~~~~~~~~~~~~~~~~

Arduino CLI doesn't strictly require a configuration file to work because the
command line interface any possible functionality. However, having one
can spare you a lot of typing when issuing a command, so let's create it
right ahead with:

.. code:: console

  $ arduino-cli config init
  Config file written: /home/luca/.arduino15/arduino-cli.yaml

If you inspect ``arduino-cli.yaml`` contents, you'll find out the available
options with their respective default values.

Create a new sketch
~~~~~~~~~~~~~~~~~~~

To create a new sketch named ``MyFirstSketch`` in the current directory, run
the following command:

.. code:: console

  $ arduino-cli sketch new MyFirstSketch
  Sketch created in: /home/luca/MyFirstSketch

A sketch is a folder containing assets like source files and libraries; the
``new`` command creates for you a .ino file called ``MyFirstSketch.ino``
containing Arduino boilerplate code:

.. code:: console

    $ cat $HOME/MyFirstSketch/MyFirstSketch.ino
    void setup() {
    }

    void loop() {
    }

At this point you can use your favourite file editor or IDE to open the
file ``$HOME/MyFirstSketch/MyFirstSketch.ino`` and change the code like this:

.. code:: c

   void setup() {
     pinMode(LED_BUILTIN, OUTPUT);
   }

   void loop() {
     digitalWrite(LED_BUILTIN, HIGH);
     delay(1000);
     digitalWrite(LED_BUILTIN, LOW);
     delay(1000);
   }

Connect the board to your PC
~~~~~~~~~~~~~~~~~~~~~~~~~~~~

The first thing to do upon a fresh install is to update the local cache of
available platforms and libraries by running:

.. code:: console

   $ arduino-cli core update-index
   Updating index: package_index.json downloaded

After connecting the board to your PCs by using the USB cable, you should be
able to check whether it's been recognized by running:

.. code:: console

   $ arduino-cli board list
   Port         Type              Board Name              FQBN                 Core
   /dev/ttyACM1 Serial Port (USB) Arduino/Genuino MKR1000 arduino:samd:mkr1000 arduino:samd

In this example, the MKR1000 board was recognized and from the output of the
command you see the platform core called ``arduino:samd`` is the one that needs
to be installed to make it work.

If you see an ``Unknown`` board listed, uploading
should still work as long as you identify the platform core and use the correct
FQBN string. When a board is not detected for whatever reason, you can list all
the supported boards and their FQBN strings by running the following:

.. code:: console

   $ arduino-cli board listall mkr
   Board Name              FQBN
   Arduino MKR FOX 1200    arduino:samd:mkrfox1200
   Arduino MKR GSM 1400    arduino:samd:mkrgsm1400
   Arduino MKR WAN 1300    arduino:samd:mkrwan1300
   Arduino MKR WiFi 1010   arduino:samd:mkrwifi1010
   Arduino MKRZERO         arduino:samd:mkrzero
   Arduino/Genuino MKR1000 arduino:samd:mkr1000

Install the core for your board
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

To install the ``arduino:samd`` platform core, run the following:

.. code:: console

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

Now verify we have installed the core properly by running:

.. code:: console

   $ arduino-cli core list
   ID              Installed       Latest  Name
   arduino:samd    1.6.19          1.6.19  Arduino SAMD Boards (32-bits ARM Cortex-M0+)

Great! Now we are ready to compile and upload the sketch.

Adding 3rd party cores
^^^^^^^^^^^^^^^^^^^^^^

If your board requires 3rd party core packages to work, you can pass a link to
the the additional package index file with the ``--additional-urls`` option to
any command that require a platform core to work:

.. code:: console

   $ arduino-cli core search esp8266 --additional-urls http://arduino.esp8266.com/stable/package_esp8266com_index.json
   ID              Version Name
   esp8266:esp8266 2.5.2   esp8266

To avoid passing the ``--additional-urls`` option every time you run a command,
you can list the URLs to additional package indexes in the Arduino CLI
configuration file.

For example, to add the ESP8266 core, edit the configration file and change the
``board_manager`` settings as follows:

.. code:: yaml

   board_manager:
     additional_urls:
       - http://arduino.esp8266.com/stable/package_esp8266com_index.json

From now on, commands supporting custom cores will automatically use the
additional URL from the configuration file:

.. code:: console

   $ arduino-cli core update-index
   Updating index: package_index.json downloaded
   Updating index: package_esp8266com_index.json downloaded
   Updating index: package_index.json downloaded

   $ arduino-cli core search esp8266
   ID              Version Name
   esp8266:esp8266 2.5.2   esp8266

Compile and upload the sketch
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

To compile the sketch you run the ``compile`` command passing the proper FQBN
string:

.. code:: console

   $ arduino-cli compile --fqbn arduino:samd:mkr1000 MyFirstSketch
   Sketch uses 9600 bytes (3%) of program storage space. Maximum is 262144 bytes.

To upload the sketch to your board, run the following command, this time also
providing the serial port where the board is connected:

.. code:: console

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

Add libraries
~~~~~~~~~~~~~

If you need to add more functionalities to your sketch, chances are some of the
libraries available in the Arduino ecosystem already provide what you need.
For example, if you need a debouncing strategy to better handle button inputs,
you can try searching for the ``debouncer`` keyword:

.. code:: console

  $ arduino-cli lib search debouncer
    Name: "Debouncer"
      Author: hideakitai
      Maintainer: hideakitai
      Sentence: Debounce library for Arduino
      Paragraph: Debounce library for Arduino
      Website: https://github.com/hideakitai
      Category: Timing
      Architecture: *
      Types: Contributed
      Versions: [0.1.0]
    Name: "FTDebouncer"
      Author: Ubi de Feo
      Maintainer: Ubi de Feo, Sebastian Hunkeler
      Sentence: An efficient, low footprint, fast pin debouncing library for Arduino
      Paragraph: This pin state supervisor manages debouncing of buttons and handles transitions between LOW and HIGH state, calling a function and notifying your code of which pin has been activated or deactivated.
      Website: https://github.com/ubidefeo/FTDebouncer
      Category: Uncategorized
      Architecture: *
      Types: Contributed
      Versions: [1.3.0]
    Name: "SoftTimer"
      Author: Balazs Kelemen <prampec+arduino@gmail.com>
      Maintainer: Balazs Kelemen <prampec+arduino@gmail.com>
      Sentence: SoftTimer is a lightweight pseudo multitasking solution for Arduino.
      Paragraph: SoftTimer enables higher level Arduino programing, yet easy to use, and lightweight. You are often faced with the problem that you need to do multiple tasks at the same time. In SoftTimer, the programmer creates Tasks that runs periodically. This library comes with a collection of handy tools like blinker, pwm, debouncer.
      Website: https://github.com/prampec/arduino-softtimer
      Category: Timing
      Architecture: *
      Types: Contributed
      Versions: [3.0.0, 3.1.0, 3.1.1, 3.1.2, 3.1.3, 3.1.5, 3.2.0]

Our favourite is ``FTDebouncer``, can install it by running:

.. code:: console

    $ arduino-cli lib install FTDebouncer
      FTDebouncer depends on FTDebouncer@1.3.0
      Downloading FTDebouncer@1.3.0...
      FTDebouncer@1.3.0 downloaded
      Installing FTDebouncer@1.3.0...
      Installed FTDebouncer@1.3.0

Getting help
------------

``arduino-cli`` is a container of commands and each command has its own
dedicated help text that can be shown with the ``help`` command like this:

.. code:: console

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

Troubleshooting
---------------

  Arduino Uno/Mega/Duemilanove is not detected when you
  run ``arduino-cli board list``

Possible causes:

-  Your board is a cheaper clone, or
-  It mounts a USB2Serial converter like FT232 or CH320: these chips
   always reports the same USB VID/PID to the operating system, so the
   only thing that we know is that the board mounts that specific
   USB2Serial chip, but we don’t know which board is.

  What's the FQBN string?

For a deeper understanding of how FQBN works, you should understand
Arduino Hardware specification. You can find more information in this
`arduino/Arduino wiki
page <https://github.com/arduino/Arduino/wiki/Arduino-IDE-1.5-3rd-party-Hardware-specification>`__

Using the gRPC interface
------------------------

The `client_example <./client_example>`__ folder contains a sample
program that shows how to use gRPC interface of the CLI.

.. |Tests passing| image:: https://github.com/Arduino/arduino-cli/workflows/test/badge.svg
   :target: https://github.com/Arduino/arduino-cli/actions?workflow=test
.. |Nightly build| image:: https://github.com/Arduino/arduino-cli/workflows/nightly/badge.svg
   :target: https://github.com/Arduino/arduino-cli/actions?workflow=nightly
.. |codecov| image:: https://codecov.io/gh/arduino/arduino-cli/branch/master/graph/badge.svg
   :target: https://codecov.io/gh/arduino/arduino-cli
