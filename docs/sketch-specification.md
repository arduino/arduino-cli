This is the specification for Arduino sketches.

The programs that run on Arduino boards are called "sketches". This term was inherited from
[Processing](https://processing.org/), upon which the Arduino IDE and the core API were based.

## Sketch folders and files

The sketch root folder name and code file names must start with a basic letter (`A`-`Z` or `a`-`z`) or number (`0`-`9`),
followed by basic letters, numbers, underscores (`_`), dots (`.`) and dashes (`-`). The maximum length is 63 characters.

Support for names starting with a number was added in Arduino IDE 1.8.4.

### Sketch root folder

Because many Arduino sketches only contain a single .ino file, it's easy to think of that file as the sketch. However,
it is the folder that is the sketch. The reason is that sketches may consist of multiple code files and the folder is
what groups those files into a single program.

### Primary sketch file

Every sketch must contain a `.ino` file with a file name matching the sketch root folder name.

`.pde` is also supported but **deprecated** and will be removed in the future, using the `.ino` extension is strongly
recommended.

### Additional code files

Sketches may consist of multiple code files.

The following extensions are supported:

- .ino - [Arduino language](https://www.arduino.cc/reference/en/) files.
- .pde - Alternate extension for Arduino language files. This file extension is also used by Processing sketches. .ino
  is recommended to avoid confusion. **`.pde` extension is deprecated and will be removed in the future.**
- .cpp - C++ files.
- .c - C Files.
- .S - Assembly language files.
- .h - Header files.
- .tpp, .ipp - Header files (available from Arduino CLI 0.19.0).

For information about how each of these files and other parts of the sketch are used during compilation, see the
[Sketch build process documentation](sketch-build-process.md).

### `src` subfolder

The contents of the `src` subfolder are compiled recursively. Unlike the code files in the sketch root folder, these
files are not shown as tabs in the IDEs.

This is useful for files you don't want to expose to the sketch user via the IDE's interface. It can be used to bundle
libraries with the sketch in order to make it a self-contained project.

Arduino language files under the `src` folder are not supported.

- In Arduino IDE 1.6.5-r5 and older, no recursive compilation was done.
- In Arduino IDE 1.6.6 - 1.6.9, recursive compilation was done of all subfolders of the sketch folder.
- In Arduino IDE 1.6.10 and newer, recursive compilation is limited to the `src` subfolder of the sketch folder.

### `data` subfolder

The `data` folder is used to add additional files to the sketch, which will not be compiled.

Files added to the sketch via the Arduino IDE's **Sketch > Add File...** are placed in the `data` folder.

The Arduino IDE's **File > Save As...** only copies the code files in the sketch root folder and the full contents of
the `data` folder, so any non-code files outside the `data` folder are stripped.

### Metadata

#### `sketch.json`

Arduino CLI and Arduino Web Editor use a file named sketch.json, located in the sketch root folder, to store sketch
metadata.

The `cpu` key contains the board configuration information. This can be set via
[`arduino-cli board attach`](commands/arduino-cli_board_attach.md) or by selecting a board in the Arduino Web Editor
while the sketch is open. With this configuration set, it is not necessary to specify the `--fqbn` or `--port` flags to
the [`arduino-cli compile`](commands/arduino-cli_compile.md) or [`arduino-cli upload`](commands/arduino-cli_upload.md)
commands when compiling or uploading the sketch.

The `included_libs` key defines the library versions the Arduino Web Editor uses when the sketch is compiled. This is
Arduino Web Editor specific because all versions of all the Library Manager libraries are pre-installed in Arduino Web
Editor, while only one version of each library may be installed when using the other Arduino development software.

#### Sketch project file

This is an optional file named `sketch.yaml`, located in the root folder of the sketch.

Inside the sketch project file the user can define one or more "profiles": each profile is a description of all the
resources needed to build the sketch (platform and libraries each pinned to a specific version).

For more information see the [sketch project file](sketch-project-file.md) documentation.

### Secrets

Arduino Web Editor has a
["Secret tab" feature](https://create.arduino.cc/projecthub/Arduino_Genuino/store-your-sensitive-data-safely-when-sharing-a-sketch-e7d0f0)
that makes it easy to share sketches without accidentally exposing sensitive data (e.g., passwords or tokens). The
Arduino Web Editor automatically generates macros for any identifier in the sketch which starts with `SECRET_` and
contains all uppercase characters.

When you download a sketch from Arduino Web Editor that contains a Secret tab, the empty `#define` directives for the
secrets are in a file named arduino_secrets.h, with an `#include` directive to that file at the top of the primary
sketch file. This is hidden when viewing the sketch in Arduino Web Editor.

### Documentation

Image and text files in common formats which are present in the sketch root folder are displayed in tabs in the Arduino
Web Editor.

### Sketch file structure example

```
Foo
|_ arduino_secrets.h
|_ Abc.ino
|_ Def.cpp
|_ Def.h
|_ Foo.ino
|_ Ghi.c
|_ Ghi.h
|_ Jkl.h
|_ Jkl.S
|_ sketch.json
|_ data
|  |_ Schematic.pdf
|_ src
   |_ SomeLib
      |_ library.properties
      |_ src
         |_ SomeLib.h
         |_ SomeLib.cpp
```

## Sketchbook

The Arduino IDE provides a "sketchbook" folder (analogous to Arduino CLI's "user directory"). In addition to being the
place where user libraries and manually installed platforms are installed, the sketchbook is a convenient place to store
sketches. Sketches in the sketchbook folder appear under the Arduino IDE's **File > Sketchbook** menu. However, there is
no requirement to store sketches in the sketchbook folder.

## Library/Boards Manager links

**(available in Arduino IDE >=1.6.9 <2.x || >=2.0.1)**

A URI in a comment in the form:

```text
http://librarymanager[/TYPE_FILTER[/TOPIC_FILTER]][#SEARCH_KEYWORDS]
```

will open a search in
[Library Manager](https://docs.arduino.cc/software/ide-v1/tutorials/installing-libraries#using-the-library-manager) when
clicked in the Arduino IDE.

A URI in a comment in the form:

```text
http://boardsmanager[/TYPE_FILTER][#SEARCH_KEYWORDS]
```

will open a search in [Boards Manager](https://docs.arduino.cc/learn/starting-guide/cores) when clicked in the Arduino
IDE.

These links can be used to offer the user an easy way to install dependencies of the sketch.

The search field will be populated with the `SEARCH_KEYWORDS` fragment component of the URI. Any characters other than
`A`-`Z`, `a`-`z`, `0`-`9`, and `:` are treated as spaces by the search algorithm, which allows multiple keywords to be
specified via the URI.

---

**(available in Arduino IDE >=2.0.1)**

The "**Type**" filter will be set to the optional `TYPE_FILTER` path component of the URI.

The Library Manager "**Topic**" filter will be set to the optional `TOPIC_FILTER` path component of the URI.

Unlike the `SEARCH_KEYWORDS` fragment, spaces and reserved characters in these components must be
[percent-encoded](https://en.wikipedia.org/wiki/Percent-encoding) (e.g., `Signal%20Input%2FOutput`).

Although the filter setting is not supported by previous IDE versions, URIs containing these path components still
function in all other respects.

---

This feature is only available when using the Arduino IDE, so be sure to provide supplementary documentation to help the
users of other development software install the sketch dependencies.

### Example

```c++
// install the Arduino SAMD Boards platform to add support for your MKR WiFi 1010 board
// if using the Arduino IDE, click here: http://boardsmanager/Arduino#SAMD

// install the WiFiNINA library via Library Manager
// if using the Arduino IDE, click here: http://librarymanager/Arduino/Communication#WiFiNINA
#include <WiFiNINA.h>
```

## See also

- [Sketch build process documentation](sketch-build-process.md)
- [Style guide for example sketches](https://docs.arduino.cc/learn/contributions/arduino-writing-style-guide)
