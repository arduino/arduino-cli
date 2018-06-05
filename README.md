![Build status](https://drone.arduino.cc/api/badges/bcmi-labs/arduino-cli/status.svg)

# arduino-cli

arduino-cli provides all the tooling needed to use Arduino compatible platforms and boards.
It implements all functions provided by Arduino IDE and Arduino Create.

### How to build from source

* You should have a recent Go compiler installed.
* Run `go get -u github.com/bcmi-labs/arduino-cli`
* The `arduino-cli` executable will be produced in `$GOPATH/bin/arduino-cli`

You may want to copy the executable into a directory which is in your `PATH` environment variable
(such as `/usr/local/bin/`).

#### Usage: an example

A general call is
```bash
arduino-cli COMMAND
```

To see the full list of commands, call one of the following:

```bash
arduino-cli help [COMMAND]
arduino-cli [COMMAND] -h
arduino-cli [COMMAND] --help
```

