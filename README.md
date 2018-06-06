![Build status](https://drone.arduino.cc/api/badges/bcmi-labs/arduino-cli/status.svg)

# arduino-cli

`arduino-cli` is an all-in-one solution that provides builder, boards/library manager, uploader, discovery and many other tools needed to use any Arduino compatible board and platforms.

This software is currently in alpha state: new features will be added and some may be changed.

It will be soon used as a building block in the Arduino IDE and Arduino Create.

## How to install

### Download the latest stable release

This is not yet available until the first stable version is released.

### Build from source

* You should have a recent Go compiler installed.
* Run `go get -u github.com/bcmi-labs/arduino-cli`
* The `arduino-cli` executable will be produced in `$GOPATH/bin/arduino-cli`

You may want to copy the executable into a directory which is in your `PATH` environment variable
(such as `/usr/local/bin/`).

## Usage

`arduino-cli` is a container of commands, to see the full list just run:
```bash
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
  login         Creates default credentials for an Arduino Create Session.
  logout        Clears credentials for the Arduino Create Session.
  sketch        Arduino CLI Sketch Commands.
  upload        Upload Arduino sketches.
  validate      Validates Arduino installation.
  version       Shows version number of arduino CLI.
....
```

Each command has his own specific help that can be obtained with the `help` command, for example:

```bash
$ arduino-cli help core
Arduino Core operations.

Usage:
  arduino-cli core [command]

Examples:
arduino core update-index # to update the package index file.

Available Commands:
  download     Downloads one or more cores and corresponding tool dependencies.
  install      Installs one or more cores and corresponding tool dependencies.
  list         Shows the list of installed cores.
  update-index Updates the index of cores.

Flags:
  -h, --help   help for core

Global Flags:
      --config-file string   The custom config file (if not specified ./.cli-config.yml will be used). (default "/home/megabug/Workspace/go/src/github.com/bcmi-labs/arduino-cli/.cli-config.yml")
      --debug                Enables debug output (super verbose, used to debug the CLI).
      --format string        The output format, can be [text|json]. (default "text")

Use "arduino-cli core [command] --help" for more information about a command.

```

