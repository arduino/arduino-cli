The process the Arduino development software uses to build a sketch. More useful information can be found in the
[Arduino platform specification](platform-specification.md). Note that the following refers specifically to the build
process for AVR targets. Other architectures follow a similar build process, but may use other tools and compilers.

## Overview

A number of things have to happen for your Arduino code to get onto the Arduino board. First, the Arduino development
software performs some minor pre-processing to turn your sketch into a C++ program. Next, dependencies of the sketch are
located. It then gets passed to a compiler (e.g, `avr-gcc`), which turns the human readable code into machine readable
instructions (or object files). Then your code gets combined with (linked against) the standard Arduino libraries that
provide basic functions like `digitalWrite()` or `Serial.print()`. The result is a single Intel hex file, which contains
the specific bytes that need to be written to the program memory of the chip on the Arduino board. This file is then
uploaded to the board: transmitted over the USB or serial connection via the bootloader already on the chip or with
external programming hardware.

## Pre-Processing

The Arduino development software performs a few transformations to your sketch before passing it to the
compiler (e.g., `avr-gcc`):

- All .ino and .pde files in the sketch folder (shown in the Arduino IDE as tabs with no extension) are concatenated
  together, starting with the file that matches the folder name followed by the others in alphabetical order. The .cpp
  filename extension is then added to the resulting file.
- If not already present, `#include <Arduino.h>` is added to the sketch. This header file (found in the core folder for
  the currently selected board) includes all the definitions needed for the standard Arduino core.
- Prototypes are generated for all function definitions in .ino/.pde files that don't already have prototypes. In some
  rare cases, prototype generation may fail for some functions. To work around this, you can provide your own prototypes
  for these functions.
- `#line` directives are added to make warning or error messages reflect the original sketch layout.

No pre-processing is done to files in a sketch with any extension other than .ino or .pde. Additionally, .h files in the
sketch are not automatically #included from the main sketch file. Further, if you want to call functions defined in a .c
file from a .cpp file (like one generated from your sketch), you'll need to wrap its declarations in an `extern "C" {}`
block that is defined only inside of C++ files.

## Dependency Resolution

The sketch is scanned recursively for dependencies. There are predefined include search paths:

1. Core library folder (as defined by [`{build.core}`](platform-specification.md#boardstxt))
1. Variant folder (as defined by [`{build.variant}`](platform-specification.md#boardstxt))
1. Standard system directories (e.g., [`{runtime.tools.avr-gcc.path}/avr/include`](platform-specification.md#tools))
1. Include search paths added to resolve prior dependencies

If the dependency is not present in any of those locations, the installed libraries are then searched (see the
[**Location Priority** table](#location-priority) below for library locations). For information on the allowed library
sub-folder structures see [the Arduino library specification](library-specification.md#source-code). `-I` options are
generated for the path to each library dependency and appended to the
[`includes` property](platform-specification.md#recipes-to-compile-source-code), to be used in
[compilation recipes](platform-specification.md#recipes-to-compile-source-code) in platform.txt.

If multiple libraries contain a file that matches the `#include` directive, the priority is determined by applying the
following rules, one by one in this order, until a rule determines a winner:

1. A library that is architecture compatible wins against a library that is not architecture compatible (see
   [**Architecture Matching**](#architecture-matching))
1. A library with both [library name](#library-name-priority) and [folder name](#folder-name-priority) matching the
   include wins
1. A library that has better "library name priority" or "folder name priority" wins (see
   [**Library Name Priority**](#library-name-priority) and [**Folder Name Priority**](#folder-name-priority))
1. A library that is architecture optimized wins against a library that is not architecture optimized (see
   [**Architecture Matching**](#architecture-matching))
1. A library that has a better "location priority" wins (see [**Location Priority**](#location-priority))
1. A library that has a folder name with a better score using the "closest-match" algorithm wins
1. A library that has a folder name that comes first in alphanumeric order wins

### Architecture Matching

A library is considered **compatible** with architecture `X` if the `architectures` field in
[library.properties](library-specification.md#library-metadata):

- explicitly contains the architecture `X`
- contains the catch-all `*`
- is not specified at all.

A library is considered **optimized** for architecture `X` only if the `architectures` field in library.properties
explicitly contains the architecture `X`. This means that a library that is optimized for architecture `X` is also
compatible with it.

Examples:

| `architectures` field in `library.properties` | Compatible with `avr` | Optimized for `avr` |
| --------------------------------------------- | --------------------- | ------------------- |
| not specified                                 | YES                   | NO                  |
| `architectures=*`                             | YES                   | NO                  |
| `architectures=avr`                           | YES                   | YES                 |
| `architectures=*,avr`                         | YES                   | YES                 |
| `architectures=*,esp8266`                     | YES                   | NO                  |
| `architectures=avr,esp8266`                   | YES                   | YES                 |
| `architectures=samd`                          | NO                    | NO                  |

### Library Name Priority

A library's name is defined by the [library.properties](library-specification.md#libraryproperties-file-format) `name`
field. That value is sanitized by replacing spaces with `_` before comparing it to the file name of the include.

The "library name priority" is determined as follows (in order of highest to lowest priority):

| Rule                                                                      | Example for `Arduino_Low_Power.h` |
| ------------------------------------------------------------------------- | --------------------------------- |
| The library name matches the include 100%                                 | `Arduino Low Power`               |
| The library name matches the include 100%, except with a `-master` suffix | `Arduino Low Power-master`        |
| The library name has a matching prefix                                    | `Arduino Low Power Whatever`      |
| The library name has a matching suffix                                    | `Awesome Arduino Low Power`       |
| The library name contains the include                                     | `The Arduino Low Power Lib`       |

### Folder Name Priority

The "folder name priority" is determined as follows (in order of highest to lowest priority):

| Rule                                                                     | Example for `Servo.h`       |
| ------------------------------------------------------------------------ | --------------------------- |
| The folder name matches the include 100%                                 | `Servo`                     |
| The folder name matches the include 100%, except with a `-master` suffix | `Servo-master`              |
| The folder name has a matching prefix                                    | `ServoWhatever`             |
| The folder name has a matching suffix                                    | `AwesomeServo`              |
| The folder name contains the include                                     | `AnAwesomeServoForWhatever` |

### Location Priority

The "location priority" is determined as follows (in order of highest to lowest priority):

1. The library is specified using the [`--library` option](commands/arduino-cli_compile.md#options) of
   `arduino-cli compile`
1. The library is under a custom libraries path specified via the
   [`--libraries` option](commands/arduino-cli_compile.md#options) of `arduino-cli compile` (in decreasing order of
   priority when multiple custom paths are defined)
1. The library is under the `libraries` subfolder of the IDE's sketchbook or Arduino CLI's user directory
1. The library is bundled with the board platform/core
   ([`{runtime.platform.path}/libraries`](platform-specification.md#global-predefined-properties))
1. The library is bundled with the [referenced](platform-specification.md#referencing-another-core-variant-or-tool)
   board platform/core
1. The library is bundled with the Arduino IDE
   ([`{runtime.ide.path}/libraries`](platform-specification.md#global-predefined-properties))
   - This location is only used by Arduino CLI when it's located in the Arduino IDE installation folder

#### Location priorities in Arduino Web Editor

The location priorities system works in the same manner in [Arduino Web Editor](https://create.arduino.cc/editor), but
its cloud-based nature may make the locations of libraries less obvious.

1. **Custom**: the imported libraries, shown under the **Libraries > Custom** tab.
   - These libraries are under `/tmp/\<some number>/custom`
1. **Pinned**: libraries that were [associated with the sketch](sketch-specification.md#metadata) by choosing a specific
   version from the library's "Include" dropdown menu.
   - These libraries are under `/tmp/\<some number>/pinned`
   - Note: clicking the "Include" button does not result in the library being pinned to the sketch.
1. **[Platform bundled](platform-specification.md#platform-bundled-libraries)**: these are listed under the
   **Libraries > Default** tab, but with "for \<architecture name\>" appended to the library name (e.g., "SPI for AVR").
   - These libraries are under `/home/builder/.arduino15/packages`
   1. [Board platform](platform-specification.md#platform-terminology) bundled
   1. [Core platform](platform-specification.md#platform-terminology) bundled
1. **Built-in**:
   - The non-platform bundled libraries listed under the **Libraries > Default** tab.
   - Libraries listed under **Libraries > Library Manager**.
   - These libraries are under `/home/builder/opt/libraries/latest`

## Compilation

Sketches are compiled by platform-specific versions of `gcc` and `g++` according to the variables in the boards.txt file of the selected board's
[platform](platform-specification.md).

The sketch is built in a temporary directory in the system-wide temporary directory (e.g. /tmp on Linux).

Files taken as source files for the build process are .S, .c and .cpp files (including the .cpp file generated from the
sketch's .ino and .pde files during the sketch pre-processing step). Source files of the target are compiled and output
with .o extensions to this build directory, as are the main sketch files and any other source files in the sketch and
any source files in any libraries which are `#include`d in the sketch.

Before compiling a source file, an attempt is made to reuse the previously compiled .o file, which speeds up the build
process. A special .d (dependency) file provides a list of all other files included by the source. The compile step is
skipped if the .o and .d files exist and have timestamps newer than the source and all the dependent files. If the
source or any dependent file has been modified, or any error occurs verifying the files, the compiler is run normally,
writing a new .o & .d file. After a new board is selected from the IDE's Board menu, all source files are rebuilt on the
next compile.

These .o files are then linked together into a static library and the main sketch file is linked against this library.
Only the parts of the library needed for your sketch are included in the final .hex file, reducing the size of most
sketches.

The .hex file is the final output of the compilation which is then uploaded to the board.

If verbose output during compilation is enabled, the complete command line of each external command executed as part of
the build process will be printed in the console.

## Uploading

Sketches are uploaded by a platform-specific upload tool (e.g., avrdude). The upload process is also controlled by variables in the boards and main preferences
files. See the [Arduino platform specification](platform-specification.md) page for details.

If verbose output during upload is enabled, debugging information will be output to the console, including the upload tool's
command lines and verbose output.
