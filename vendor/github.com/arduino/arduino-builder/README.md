## Arduino Builder [![Build Status](https://travis-ci.org/arduino/arduino-builder.svg?branch=master)](https://travis-ci.org/arduino/arduino-builder)

A command line tool for compiling Arduino sketches

This tool is able to parse [Arduino Hardware specifications](https://github.com/arduino/Arduino/wiki/Arduino-IDE-1.5-3rd-party-Hardware-specification), properly run `gcc` and produce compiled sketches.

An Arduino sketch differs from a standard C program in that it misses a `main` (provided by the Arduino core), function prototypes are not mandatory, and libraries inclusion is automagic (you just have to `#include` them).
This tool generates function prototypes and gathers library paths, providing `gcc` with all the needed `-I` params.

### Usage

* `-compile` or `-dump-prefs` or `-preprocess`: Optional. If omitted, defaults to `-compile`. `-dump-prefs` will just print all build preferences used, `-compile` will use those preferences to run the actual compiler, `-preprocess` will only print preprocessed code to stdout.

* `-hardware`: Mandatory. Folder containing Arduino platforms. An example is the `hardware` folder shipped with the Arduino IDE, or the `packages` folder created by Arduino Boards Manager. Can be specified multiple times. If conflicting hardware definitions are specified, the last one wins.

* `-tools`: Mandatory. Folder containing Arduino tools (`gcc`, `avrdude`...). An example is the `hardware/tools` folder shipped with the Arduino IDE, or the `packages` folder created by Arduino Boards Manager. Can be specified multiple times.

* `-libraries`: Optional. Folder containing Arduino libraries. An example is the `libraries` folder shipped with the Arduino IDE. Can be specified multiple times.

* `-fqbn`: Mandatory. Fully Qualified Board Name, e.g.: arduino:avr:uno

* `-build-path`: Optional. Folder where to save compiled files. If omitted, a folder will be created in the temporary folder specified by your OS.

* `-prefs=key=value`: Optional. It allows to override some build properties.

* `-warnings`: Optional, can be "none", "default", "more" and "all". Defaults to "none". Used to tell `gcc` which warning level to use (`-W` flag).

* `-verbose`: Optional, turns on verbose mode.

* `-quiet`: Optional, supresses almost every output.

* `-debug-level`: Optional, defaults to "5". Used for debugging. Set it to 10 when submitting an issue.

* `-core-api-version`: Optional, defaults to "10600". The version of the Arduino IDE which is using this tool.

* `-logger`: Optional, can be "human", "humantags" or "machine". Defaults to "human". If "humantags" the messages are qualified with a prefix that indicates their level (info, debug, error). If "machine", messages emitted will be in a format which the Arduino IDE understands and that it uses for I18N.

* `-version`: if specified, prints version and exits.

* `-build-options-file`: it specifies path to a local `build.options.json` file (see paragraph below), which allows you to omit specifying params such as `-hardware`, `-tools`, `-libraries`, `-fqbn`, `-pref` and `-ide-version`.

* `-vid-pid`: when specified, VID/PID specific build properties are used, if boards supports them.

Final mandatory parameter is the sketch to compile (of course).

### What is and how to use build.options.json file

Every time you run this tool, it will create a `build.options.json` file in build path. It's used to understand if build options (such as hardware folders, fqbn and so on) were changed when compiling the same sketch.
If they changed, the whole build path is wiped out. If they didn't change, previous compiled files will be reused if the corresponding source files didn't change as well.
You can save this file locally and use it instead of specifying `-hardware`, `-tools`, `-libraries`, `-fqbn`, `-pref` and `-ide-version`.

### Using it for continuously verify your libraries or cores

See [Doing continuous integration with arduino builder](https://github.com/arduino/arduino-builder/wiki/Doing-continuous-integration-with-arduino-builder/).

### Building from source

You need [a recent version of Go (>=1.8.0)](https://golang.org/).

To build, run the following commands:

```
go get github.com/go-errors/errors
go get github.com/stretchr/testify
go get github.com/jstemmer/go-junit-report
go get -u github.com/arduino/go-properties-map
go get -u github.com/arduino/go-timeutils
go get google.golang.org/grpc
go get github.com/golang/protobuf/proto
go get golang.org/x/net/context
go get github.com/fsnotify/fsnotify
go get github.com/arduino/arduino-builder
go build github.com/arduino/arduino-builder/arduino-builder
```

### TDD

In order to run the tests, type:

```
go test github.com/arduino/arduino-builder/...
```

This runs all tests, showing any failures and a summary at the end.
Add the -v option to show each test as it is being ran. Currently,
arduino-builder itself also generates copious output, even for
non-failing testcases and without -v, and testing does not stop at the
first failure, so you probably want to redirect test output so you can
scroll back to find any failures.

To run a single test, use the -run option, which accepts a regular
expression (see also go help testflag).

```
go test github.com/arduino/arduino-builder/... -run 'TestBuilderEmptySketch'
go test github.com/arduino/arduino-builder/... -run 'TestPrototypesAdder.*'
```

In jenkins, use
```
go test -v github.com/arduino/arduino-builder/... | bin/go-junit-report > report.xml
```

The first time you run the tests, some needed files (toolchains and
source files) will be downloaded, which needs about 1GB of space (at the
time of writing). If you have a slow connection, this download might
exceed the default 10 minute timeout for a single test. If you run into
this, add `-timeout 60m` or similar to the commandline to extend the
timeout. If you are running on slower system (like a rasbperry pi),
increasing the timeout might be needed as well.

### License and Copyright

`arduino-builder` is licensed under General Public License version 2, as published by the Free Software Foundation. See [LICENSE.txt](LICENSE.txt).

Copyright (C) 2017 Arduino AG and contributors

See https://www.arduino.cc/ and https://github.com/arduino/arduino-builder/graphs/contributors
