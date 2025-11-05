Introduced in Arduino IDE 1.6.4, Boards Manager makes it easy to install and update Arduino platforms. In order to
provide Boards Manager installation support for a platform, a JSON formatted index file must be published. This is the
specification for that file.

Boards Manager functionality is provided by [Arduino CLI](getting-started.md#adding-3rd-party-cores) and
[Arduino IDE](https://docs.arduino.cc/learn/starting-guide/cores).

## Naming of the JSON index file

Many different index files coming from different vendors may be in use, so each vendor should name their own index file
in a way that won't conflict with others. The file must be named as follows:

`package_YOURNAME_PACKAGENAME_index.json`

The prefix `package_` and the postfix `_index.json` are **mandatory** (otherwise the index file is not recognised by the
Arduino development software) while the choice of `YOURNAME_PACKAGENAME` is left to the packager. We suggest using a
domain name owned by the packager. For example:

`package_arduino.cc_index.json`

or

`package_example.com_avr_boards_index.json`

The index URL is periodically checked for updates, so expect a constant flow of downloads (proportional to the number of
active users).

## JSON Index file contents

The root of the JSON index is an array of `packages`:

```json
{
  "packages": [PACKAGE_XXXX]
}
```

3rd party vendors should use a single `PACKAGE_XXXX` that is a dictionary map with the vendor's metadata, a list of
`PLATFORMS` and a list of `TOOLS`. For example:

<!-- prettier-ignore -->
```json
    {
      "name": "arduino",
      "maintainer": "Arduino LLC",
      "websiteURL": "http://www.arduino.cc/",
      "email": "packages@arduino.cc",

      "platforms": [PLATFORM_AVR, PLATFORM_ARM, PLATFORM_XXXXX, PLATFORM_YYYYY],

      "tools": [
        TOOLS_COMPILER_AVR,
        TOOLS_UPLOADER_AVR,
        TOOLS_COMPILER_ARM,
        TOOLS_XXXXXXX,
        TOOLS_YYYYYYY
      ]
    }

```

The metadata fields are:

- `name`: the folder used for the installed cores. The
  [vendor folder](platform-specification.md#hardware-folders-structure) name of the installed package is determined by
  this field
  - The value must not contain any characters other than the letters `A`-`Z` and `a`-`z`, numbers (`0`-`9`), underscores
    (`_`), dashes (`-`), and dots (`.`).
- `maintainer`: the extended name of the vendor that is displayed on the Arduino IDE Boards Manager GUI
- `websiteURL`: the URL to the vendor's website, appears on the Arduino IDE Boards Manager as a "More info" link
- `email`: the email of the vendor/maintainer

Now, before looking at `PLATFORMS`, let's explore first how `TOOLS` are made.

### Tools definitions

Each tool describes a binary distribution of a command line tool. A tool can be:

- a compiler toolchain
- an uploader
- a file preprocessor
- a debugger
- a program that performs a firmware upgrade
- a [pluggable discovery](pluggable-discovery-specification.md)
- a [pluggable monitor](pluggable-monitor-specification.md)

basically anything that can run on the user's host PC and do something useful.

For example, Arduino uses two command line tools for the AVR boards: avr-gcc (the compiler) and avrdude (the uploader).

Tools are mapped as JSON in this way:

```json
        {
          "name": "avr-gcc",
          "version": "7.3.0-atmel3.6.1-arduino7",
          "systems": [
            {
              "size": "34683056",
              "checksum": "SHA-256:3903553d035da59e33cff9941b857c3cb379cb0638105dfdf69c97f0acc8e7b5",
              "host": "arm-linux-gnueabihf",
              "archiveFileName": "avr-gcc-7.3.0-atmel3.6.1-arduino7-arm-linux-gnueabihf.tar.bz2",
              "url": "http://downloads.arduino.cc/tools/avr-gcc-7.3.0-atmel3.6.1-arduino7-arm-linux-gnueabihf.tar.bz2"
            },
            {
              "size": "38045723",
              "checksum": "SHA-256:03d322b9df6da17289e9e7c6233c34a8535d9c645c19efc772ba19e56914f339",
              "host": "aarch64-linux-gnu",
              "archiveFileName": "avr-gcc-7.3.0-atmel3.6.1-arduino7-aarch64-pc-linux-gnu.tar.bz2",
              "url": "http://downloads.arduino.cc/tools/avr-gcc-7.3.0-atmel3.6.1-arduino7-aarch64-pc-linux-gnu.tar.bz2"
            },
            {
              "size": "36684546",
              "checksum": "SHA-256:f6ed2346953fcf88df223469088633eb86de997fa27ece117fd1ef170d69c1f8",
              "host": "x86_64-apple-darwin14",
              "archiveFileName": "avr-gcc-7.3.0-atmel3.6.1-arduino7-x86_64-apple-darwin14.tar.bz2",
              "url": "http://downloads.arduino.cc/tools/avr-gcc-7.3.0-atmel3.6.1-arduino7-x86_64-apple-darwin14.tar.bz2"
            },
            {
              "size": "52519412",
              "checksum": "SHA-256:a54f64755fff4cb792a1495e5defdd789902a2a3503982e81b898299cf39800e",
              "host": "i686-mingw32",
              "archiveFileName": "avr-gcc-7.3.0-atmel3.6.1-arduino7-i686-w64-mingw32.zip",
              "url": "http://downloads.arduino.cc/tools/avr-gcc-7.3.0-atmel3.6.1-arduino7-i686-w64-mingw32.zip"
            },
            {
              "size": "37176991",
              "checksum": "SHA-256:954bbffb33545bcdcd473af993da2980bf32e8461ff55a18e0eebc7b2ef69a4c",
              "host": "i686-linux-gnu",
              "archiveFileName": "avr-gcc-7.3.0-atmel3.6.1-arduino7-i686-pc-linux-gnu.tar.bz2",
              "url": "http://downloads.arduino.cc/tools/avr-gcc-7.3.0-atmel3.6.1-arduino7-i686-pc-linux-gnu.tar.bz2"
            },
            {
              "size": "37630618",
              "checksum": "SHA-256:bd8c37f6952a2130ac9ee32c53f6a660feb79bee8353c8e289eb60fdcefed91e",
              "host": "x86_64-linux-gnu",
              "archiveFileName": "avr-gcc-7.3.0-atmel3.6.1-arduino7-x86_64-pc-linux-gnu.tar.bz2",
              "url": "http://downloads.arduino.cc/tools/avr-gcc-7.3.0-atmel3.6.1-arduino7-x86_64-pc-linux-gnu.tar.bz2"
            }
          ]
        },
```

The field `name` and `version` are respectively the name and version of the tool. Each tool is uniquely identified by
the triple (`packager`, `name`, `version`). `packager` (AKA "vendor") is defined by the `name` value of the tool's
package. There can be many different versions of the same tool available at the same time, for example:

- (`arduino`, `avr-gcc`, `5.4.0-atmel3.6.1-arduino2`)
- (`arduino`, `avr-gcc`, `7.3.0-atmel3.6.1-arduino5`)
- (`arduino`, `avr-gcc`, `7.3.0-atmel3.6.1-arduino7`)
- (`arduino`, `avrdude`, `5.11`)
- (`arduino`, `avrdude`, `6.0`)
- (`arduino`, `avrdude`, `6.1`)
- .....

The `systems` field lists all available [Tools Flavours](#tools-flavours-available-builds-made-for-different-os).

The other fields are:

- `url`: the download URL of the tool's archive
- `archiveFileName`: the name of the file saved to disk after the download (some web servers don't provide the filename
  through the HTTP request)
- `size`: the size of the archive in bytes
- `checksum`: the checksum of the archive, used to check if the file has been corrupted. The format is
  `ALGORITHM:CHECKSUM`, currently `MD5`, `SHA-1`,`SHA-256` algorithm are supported, we recommend `SHA-256`. On \*nix or
  macOS you can use the command `shasum -a 256 filename` to generate SHA-256 checksums. There are free options for
  Windows, including md5deep. There are also online utilities for generating checksums.

#### Tools flavours (available builds made for different OS)

Each tool version may come in different build flavours for different OS. Each flavour is listed under the `systems`
array. The IDE will take care to install the right flavour for the user's OS by matching the `host` value with the
following table or fail if a needed flavour is missing.

| OS flavour      | `host` regexp                          | suggested `host` value              |
| --------------- | -------------------------------------- | ----------------------------------- |
| Linux 32        | `i[3456]86-.*linux-gnu`                | `i686-linux-gnu`                    |
| Linux 64        | `x86_64-.*linux-gnu`                   | `x86_64-linux-gnu`                  |
| Linux Arm       | `arm.*-linux-gnueabihf`                | `arm-linux-gnueabihf`               |
| Linux Arm64     | `(aarch64\|arm64)-linux-gnu`           | `aarch64-linux-gnu`                 |
| Linux RISC-V 64 | `riscv64-linux-gnu`                    | `riscv64-linux-gnu`                 |
| Windows 32      | `i[3456]86-.*(mingw32\|cygwin)`        | `i686-mingw32` or `i686-cygwin`     |
| Windows 64      | `(amd64\|x86_64)-.*(mingw32\|cygwin)`  | `x86_64-mingw32` or `x86_64-cygwin` |
| Windows Arm64   | `(aarch64\|arm64)-.*(mingw32\|cygwin)` | `arm64-mingw32` or `arm64-cygwin`   |
| MacOSX 32       | `i[3456]86-apple-darwin.*`             | `i686-apple-darwin`                 |
| MacOSX 64       | `x86_64-apple-darwin.*`                | `x86_64-apple-darwin`               |
| MacOSX Arm64    | `arm64-apple-darwin.*`                 | `arm64-apple-darwin`                |
| FreeBSD 32      | `i?[3456]86-freebsd[0-9]*`             | `i686-freebsd`                      |
| FreeBSD 64      | `amd64-freebsd[0-9]*`                  | `amd64-freebsd`                     |
| FreeBSD Arm     | `arm.*-freebsd[0-9]*`                  | `arm-freebsd`                       |

The `host` value is matched with the regexp, this means that a more specific value for the `host` field is allowed (for
example you may write `x86_64-apple-darwin14.1` for MacOSX instead of the suggested `x86_64-apple-darwin`), by the way,
we recommend to keep it simple and stick to the suggested value in the table.

Some OS allows to run different flavours:

| The OS...     | ...may also run builds for |
| ------------- | -------------------------- |
| Windows 64    | Windows 32                 |
| Windows Arm64 | Windows 32 or Windows 64   |
| MacOSX 64     | MacOSX 32                  |
| MacOSX Arm64  | MacOSX 64 or MacOSX 32     |

This is taken into account when the tools are downloaded (for example if we are on a Windows 64 machine and the needed
tool is available only for the Windows 32 flavour, then the Windows 32 flavour will be downloaded and used).

For completeness, the previous example `avr-gcc` comes with builds for:

- ARM Linux 32 (`arm-linux-gnueabihf`),
- ARM Linux 64 (`aarch64-linux-gnu`),
- MacOSX 64 (`x86_64-apple-darwin14`),
- Windows 32 (`i686-mingw32`),
- Linux 32 (`i686-linux-gnu`),
- Linux 64 (`x86_64-linux-gnu`)
- MacOSX Arm64 will use the MacOSX 64 flavour
- Windows 64 will use the Windows 32 flavour
- Windows Arm64 will use the Windows 32 flavour

Note: this information is not used to select the toolchain during compilation. If you want a specific version to be
used, you should use the notation `{runtime.tools.TOOLNAME-VERSION.path}` in the platform.txt.

### Platforms definitions

Finally, let's see how `PLATFORMS` are made.

```json
        {
          "name": "Arduino AVR Boards",
          "architecture": "avr",
          "version": "1.6.6",
          "category": "Arduino",
          "help": {
            "online": "http://www.arduino.cc/en/Reference/HomePage"
          },
          "url": "http://downloads.arduino.cc/cores/avr-1.6.6.tar.bz2",
          "archiveFileName": "avr-1.6.6.tar.bz2",
          "checksum": "SHA-256:08ad5db4978ebea22344edc5d77dce0923d8a644da7a14dc8072e883c76058d8",
          "size": "4876916",
          "boards": [
            {"name": "Arduino Yún"},
            {"name": "Arduino Uno"},
            {"name": "Arduino Diecimila"},
            {"name": "Arduino Nano"},
            {"name": "Arduino Mega"},
            {"name": "Arduino MegaADK"},
            {"name": "Arduino Leonardo"},
          ],
          "toolsDependencies": [
            { "packager": "arduino", "name": "avr-gcc", "version": "4.8.1-arduino5" },
            { "packager": "arduino", "name": "avrdude", "version": "6.0.1-arduino5" }
          ],
          "discoveryDependencies": [
            { "packager": "arduino", "name": "serial-discovery" },
            { "packager": "arduino", "name": "mdns-discovery" }
          ],
          "monitorDependencies": [
            { "packager": "arduino", "name": "serial-monitor" }
          ]
        },
```

Each PLATFORM describes a core for a specific architecture. The fields needed are:

- `name`: the extended name of the platform that is displayed on the Boards Manager GUI
- `architecture`: is the architecture of the platform (avr, sam, etc...). It must match the architecture of the core as
  explained in the [Arduino platform specification](platform-specification.md#hardware-folders-structure)
  - The value must not contain any characters other than the letters `A`-`Z` and `a`-`z`, numbers (`0`-`9`), underscores
    (`_`), dashes (`-`), and dots (`.`).
- `version`: the version of the platform.
- `deprecated`: (optional) setting to `true` causes the platform to be moved to the bottom of all Boards Manager and
  [`arduino-cli core`](https://arduino.github.io/arduino-cli/latest/commands/arduino-cli_core/) listings and marked
  "DEPRECATED".
- `category`: this field is reserved, a 3rd party core must set it to `Contributed`
- `help`/`online`: is a URL that is displayed on the Arduino IDE's Boards Manager as an "Online Help" link
- `url`, `archiveFileName`, `size` and `checksum`: metadata of the core archive file. The meaning is the same as for the
  TOOLS
- `boards`: the list of boards supported (note: just the names to display on the Arduino IDE's Boards Manager GUI! the
  real boards definitions are inside `boards.txt` inside the core archive file)
- `toolsDependencies`: the tools needed by this platform. They will be installed by Boards Manager along with the
  platform. Each tool is referenced by the triple (`packager`, `name`, `version`) as previously said. Note that you can
  reference tools available in other packages as well, even if no platform of that package is installed.
- `discoveryDependencies`: the Pluggable Discoveries needed by this platform. These are [tools](#tools-definitions),
  defined exactly like the ones referenced in `toolsDependencies`. Unlike `toolsDependencies`, discoveries are
  referenced by the pair (`packager`, `name`). The `version` is not specified because the latest installed discovery
  tool will always be used. Like `toolsDependencies` they will be installed by Boards Manager along with the platform
  and can reference tools available in other packages as well, even if no platform of that package is installed.
- `monitorDependencies`: the Pluggable Monitors needed by this platform. These are [tools](#tools-definitions), defined
  exactly like the ones referenced in `toolsDependencies`. Unlike `toolsDependencies`, monitors are referenced by the
  pair (`packager`, `name`). The `version` is not specified because the latest installed monitor tool will always be
  used. Like `toolsDependencies` they will be installed by Boards Manager along with the platform and can reference
  tools available in other packages as well, even if no platform of that package is installed.

The `version` field is validated by both Arduino IDE and [JSemVer](https://github.com/zafarkhaja/jsemver). Here are the
rules Arduino IDE follows for parsing versions
([source](https://github.com/arduino/Arduino/blob/master/arduino-core/src/cc/arduino/contributions/VersionHelper.java)):

- Split the version at the `-` character and continue with the first part.
- If there are no dots (`.`), parse `version` as an integer and form a Version from that integer using
  `Version.forIntegers`
- If there is one dot, split `version` into two, parse each part as an integer, and form a Version from those integers
  using `Version.forIntegers`
- Otherwise, simply parse `version` into a Version using `Version.valueOf`

Note: if you miss a bracket in the JSON index, then add the URL to your Preferences, and open Boards Manager it can
cause the Arduino IDE to no longer load until you have deleted the file from your arduino15 folder.

#### How a tool's path is determined in platform.txt

When the IDE needs a tool, it downloads the corresponding archive file and unpacks the content into a private folder
that can be referenced from `platform.txt` using one of the following properties:

- `{runtime.tools.TOOLNAME-VERSION.path}`
- `{runtime.tools.TOOLNAME.path}`

For example, to obtain the avr-gcc 4.8.1 folder we can use `{runtime.tools.avr-gcc-4.8.1.path}` or
`{runtime.tools.avr-gcc.path}`.

In general the same tool may be provided by different packagers (for example the Arduino packager may provide an
`arduino:avr-gcc` and another 3rd party packager may provide their own `3rdparty:avr-gcc`). The rules to disambiguate
are as follows:

- The property `{runtime.tools.TOOLNAME.path}` points, in order of priority, to:
  1. the tool, version and packager specified via `toolsDependencies` in the `package_index.json`
  1. the highest version of the tool provided by the packager of the current platform
  1. the highest version of the tool provided by the packager of the referenced platform used for compile (see
     ["Referencing another core, variant or tool"](platform-specification.md#referencing-another-core-variant-or-tool)
     for more info)
  1. the highest version of the tool provided by any other packager (in case of tie, the first packager in alphabetical
     order wins)

- The property `{runtime.tools.TOOLNAME-VERSION.path}` points, in order of priority, to:
  1. the tool and version provided by the packager of the current platform
  1. the tool and version provided by the packager of the referenced platform used for compile (see
     ["Referencing another core, variant or tool"](platform-specification.md#referencing-another-core-variant-or-tool)
     for more info)
  1. the tool and version provided by any other packager (in case of tie, the first packager in alphabetical order wins)

### Example JSON index file

```json
{
  "packages": [
    {
      "name": "myboard",
      "maintainer": "Jane Developer",
      "websiteURL": "https://github.com/janedeveloper/myboard",
      "email": "jane@janedeveloper.org",
      "help": {
        "online": "http://janedeveloper.org/forum/myboard"
      },
      "platforms": [
        {
          "name": "My Board",
          "architecture": "avr",
          "version": "1.0.0",
          "category": "Contributed",
          "help": {
            "online": "http://janedeveloper.org/forum/myboard"
          },
          "url": "https://janedeveloper.github.io/myboard/myboard-1.0.0.zip",
          "archiveFileName": "myboard-1.0.0.zip",
          "checksum": "SHA-256:ec3ff8a1dc96d3ba6f432b9b837a35fd4174a34b3d2927de1d51010e8b94f9f1",
          "size": "15005",
          "boards": [{ "name": "My Board" }, { "name": "My Board Pro" }],
          "toolsDependencies": [
            {
              "packager": "arduino",
              "name": "avr-gcc",
              "version": "4.8.1-arduino5"
            },
            {
              "packager": "arduino",
              "name": "avrdude",
              "version": "6.0.1-arduino5"
            }
          ]
        },
        {
          "name": "My Board",
          "architecture": "avr",
          "version": "1.0.1",
          "category": "Contributed",
          "help": {
            "online": "http://janedeveloper.org/forum/myboard"
          },
          "url": "https://janedeveloper.github.io/myboard/myboard-1.0.1.zip",
          "archiveFileName": "myboard-1.0.1.zip",
          "checksum": "SHA-256:9c86ee28a7ce9fe33e8b07ec643316131e0031b0d22e63bb398902a5fdadbca9",
          "size": "15125",
          "boards": [{ "name": "My Board" }, { "name": "My Board Pro" }],
          "toolsDependencies": [
            {
              "packager": "arduino",
              "name": "avr-gcc",
              "version": "4.8.1-arduino5"
            },
            {
              "packager": "arduino",
              "name": "avrdude",
              "version": "6.0.1-arduino5"
            }
          ]
        }
      ],
      "tools": []
    }
  ]
}
```

In the example there is one `PACKAGE`, My Board. The package is compatible with the AVR architecture. There are two
versions of the `PACKAGE`, 1.0.0 and 1.0.1. No `TOOLS` needed to be installed so that section was left empty.

Here is the Boards Manager entry created by the example: ![Boards Manager screenshot](img/boards-manager-screenshot.png)

## Archive structure

It must contain a single folder in the root. All files and `__MACOSX` folder present in the root will be ignored.

Valid structure

```
.
└── avr/
    ├── bootloaders
    ├── cores
    ├── firmwares
    ├── libraries
    ├── variants
    ├── boards.txt
    ├── platform.txt
    └── programmers.txt
```

Invalid structure:

```
.
├── avr/
│   ├── ...
│   ├── boards.txt
│   ├── platform.txt
│   └── programmers.txt
├── folder2
└── folder3
```

**Note**: the folder structure of the core archive is slightly different from the standard manually installed Arduino
IDE 1.5+ compatible hardware folder structure. You must remove the architecture folder(e.g., `avr` or `arm`), moving all
the files and folders within the architecture folder up a level.

### Installation

The installation archives contain the Board support files.

Supported formats are `.zip`, `.tar.bz2`, and `.tar.gz`. Starting from Arduino CLI >=0.30.0 support for `.tar.xz`, and
`.tar.zst` has been added, by the way, if you want to keep compatibility with older versions of Arduino IDE and Arduino
CLI we recommend using one of the older formats.

The folder structure of the core archive is slightly different from the standard manually installed Arduino IDE 1.5+
compatible hardware folder structure. You must remove the architecture folder(e.g., `avr` or `arm`), moving all the
files and folders within the architecture folder up a level.

---

After adding Boards Manager support for your boards, please share the JSON index file URL on the
[Unofficial list of 3rd party boards support urls](https://github.com/arduino/Arduino/wiki/Unofficial-list-of-3rd-party-boards-support-urls).
