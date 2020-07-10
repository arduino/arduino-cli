Introduced in Arduino IDE 1.6.4, Boards Manager makes it easy to install and update Arduino platforms. In order to
provide Boards Manager installation support for a platform, a JSON formatted index file must be published. This is the
specification for that file.

Boards Manager functionality is provided by [Arduino CLI](getting-started.md#adding-3rd-party-cores),
[Arduino IDE](https://www.arduino.cc/en/Guide/Cores), and Arduino Pro IDE.

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

The index URL is periodically checked for updates so expect a constant flow of downloads (proportional to the number of
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
- `maintainer`: the extended name of the vendor that is displayed on the Arduino IDE/Pro IDE's Boards Manager GUI
- `websiteURL`: the URL to the vendor's website, appears on the Arduino IDE/Pro IDE's Boards Manager as a "More info"
  link
- `email`: the email of the vendor/maintainer

Now, before looking at `PLATFORMS`, let's explore first how `TOOLS` are made.

### Tools definitions

Each tool describes a binary distribution of a command line tool. A tool can be:

- a compiler toolchain
- an uploader
- a file preprocessor
- a debugger
- a program that performs a firmware upgrade

basically anything that can run on the user's host PC and do something useful.

For example, Arduino uses two command line tools for the AVR boards: avr-gcc (the compiler) and avrdude (the uploader).

Tools are mapped as JSON in this way:

```json
        {
          "name": "avr-gcc",
          "version": "4.8.1-arduino5",
          "systems": [
            {
              "host": "i386-apple-darwin11",
              "url": "http://downloads.arduino.cc/tools/avr-gcc-4.8.1-arduino5-i386-apple-darwin11.tar.bz2",
              "archiveFileName": "avr-gcc-4.8.1-arduino5-i386-apple-darwin11.tar.bz2",
              "size": "24437400",
              "checksum": "SHA-256:111b3ef00d737d069eb237a8933406cbb928e4698689e24663cffef07688a901"
            },
            {
              "host": "x86_64-linux-gnu",
              "url": "http://downloads.arduino.cc/tools/avr-gcc-4.8.1-arduino5-x86_64-pc-linux-gnu.tar.bz2",
              "archiveFileName": "avr-gcc-4.8.1-arduino5-x86_64-pc-linux-gnu.tar.bz2",
              "size": "27093036",
              "checksum": "SHA-256:9054fcc174397a419ba56c4ce1bfcbcad275a6a080cc144905acc9b0351ee9cc"
            },
            {
              "host": "i686-linux-gnu",
              "url": "http://downloads.arduino.cc/tools/avr-gcc-4.8.1-arduino5-i686-pc-linux-gnu.tar.bz2",
              "archiveFileName": "avr-gcc-4.8.1-arduino5-i686-pc-linux-gnu.tar.bz2",
              "size": "25882375",
              "checksum": "SHA-256:7648b7f549b37191da0b0be53bae791b652f82ac3cb4e7877f85075aaf32141f"
            },
            {
              "host": "i686-mingw32",
              "url": "http://downloads.arduino.cc/tools/avr-gcc-4.8.1-arduino5-i686-mingw32.zip",
              "archiveFileName": "avr-gcc-4.8.1-arduino5-i686-mingw32.zip",
              "size": "46044779",
              "checksum": "SHA-256:d4303226a7b41d3c445d901b5aa5903458def3fc7b7ff4ffef37cabeb37d424d"
            }
          ]
        },
```

The field `name` and `version` are respectively the name and version of the tool. Each tool is uniquely identified by
the triple (`packager`, `name`, `version`). `packager` (AKA "vendor") is defined by the `name` value of the tool's
package. There can be many different versions of the same tool available at the same time, for example:

- (`arduino`, `avr-gcc`, `4.8.1-arduino2`)
- (`arduino`, `avr-gcc`, `4.8.1-arduino3`)
- (`arduino`, `avr-gcc`, `4.8.1-arduino5`)
- (`arduino`, `avrdude`, `5.11`)
- (`arduino`, `avrdude`, `6.0`)
- (`arduino`, `avrdude`, `6.1`)
- .....

Each tool version may come in different build flavours for different OS. Each flavour is listed under the `systems`
array. In the example above `avr-gcc` comes with builds for:

- Linux 64-bit (`x86_64-linux-gnu`),
- Linux 32-bit (`i686-linux-gnu`),
- Windows (`i686-mingw32`),
- Mac (`i386-apple-darwin11`)

The IDE will take care to install the right flavour based on the `host` value, or fail if a needed flavour is
missing.<br> Note that this information is not used to select the toolchain during compilation. If you want this
specific version to be used, you should use the notation {runtime.tools.TOOLNAME-VERSION.path} in the platform.txt.

The other fields are:

- `url`: the download URL of the tool's archive
- `archiveFileName`: the name of the file saved to disk after the download (some web servers don't provide the filename
  through the HTTP request)
- `size`: the size of the archive in bytes
- `checksum`: the checksum of the archive, used to check if the file has been corrupted. The format is
  `ALGORITHM:CHECKSUM`, currently `MD5`, `SHA-1`,`SHA-256` algorithm are supported, we recommend `SHA-256`. On \*nix or
  MacOSX you may be able to use the command `shasum -a 256 filename` to generate SHA-256 checksums. There are many free
  options for Windows including md5deep, there are also online utilities for generating checksums.

##### How a tool's path is determined in platform.txt

When the IDE needs a tool it downloads the corresponding archive file and unpacks the content into a private folder that
can be referenced from `platform.txt` using one of the following properties:

- `{runtime.tools.TOOLNAME-VERSION.path}`
- `{runtime.tools.TOOLNAME.path}`

For example to obtain the avr-gcc 4.8.1 folder we can use `{runtime.tools.avr-gcc-4.8.1-arduino5.path}` or
`{runtime.tools.avr-gcc.path}`.

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
          ]
        },
```

Each PLATFORM describes a core for a specific architecture. The fields needed are:

- `name`: the extended name of the platform that is displayed on the Boards Manager GUI
- `architecture`: is the architecture of the platform (avr, sam, etc...). It must match the architecture of the core as
  explained in the [Arduino platform specification](platform-specification.md#hardware-folders-structure)
- `version`: the version of the platform.
- `category`: this field is reserved, a 3rd party core must set it to `Contributed`
- `help`/`online`: is a URL that is displayed on the Arduino IDE's Boards Manager as an "Online Help" link
- `url`, `archiveFileName`, `size` and `checksum`: metadata of the core archive file. The meaning is the same as for the
  TOOLS
- `boards`: the list of boards supported (note: just the names to display on the Arduino IDE and Arduino Pro IDE's
  Boards Manager GUI! the real boards definitions are inside `boards.txt` inside the core archive file)
- `toolsDependencies`: the tools needed by this core. Each tool is referenced by the triple (`packager`, `name`,
  `version`) as previously said. Note that you can reference tools available in other packages as well.

The `version` field is validated by both Arduino IDE and [JSemVer](https://github.com/zafarkhaja/jsemver). Here are the
rules Arduino IDE follows for parsing versions
([source](https://github.com/arduino/Arduino/blob/master/arduino-core/src/cc/arduino/contributions/VersionHelper.java)):

- Split the version at the - character and continue with the first part.
- If there are no dots (`.`), parse `version` as an integer and form a Version from that integer using
  `Version.forIntegers`
- If there is one dot, split `version` into two, parse each part as an integer, and form a Version from those integers
  using `Version.forIntegers`
- Otherwise, simply parse `version` into a Version using `Version.valueOf`

Note: if you miss a bracket in the JSON index, then add the URL to your Preferences, and open Boards Manager it can
cause the Arduino IDE to no longer load until you have deleted the file from your arduino15 folder.

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
versions of the `PACKAGE`, 1.0.0 and 1.0.1. No `TOOLS` needed to be installed so that section was left blank.

Here is the Boards Manager entry created by the example: ![Boards Manager screenshot](img/boards-manager-screenshot.png)

## Installation archive structure

The installation archives contain the Board support files. Supported formats are .zip, .tar.bz2, and .tar.gz.

The folder structure of the core archive is slightly different from the standard manually installed Arduino IDE 1.5+
compatible hardware folder structure. You must remove the architecture folder(e.g., `avr` or `arm`), moving all the
files and folders within the architecture folder up a level.

---

After adding Boards Manager support for your boards, please share the JSON index file URL on the
[Unofficial list of 3rd party boards support urls](https://github.com/arduino/Arduino/wiki/Unofficial-list-of-3rd-party-boards-support-urls).
