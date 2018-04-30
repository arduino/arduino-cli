![Build status](https://drone.arduino.cc/api/badges/bcmi-labs/arduino-cli/status.svg)

# arduino-cli

Arduino CLI is a tool to access all Arduino Create API from Command Line.
It implements all functions provided by web version of Arduino Create.

### How to build from source

* You should have a recent Go compiler installed.
* Run `go get github.com/bcmi-labs/arduino-cli`

```bash
go build -o arduino-cli
./arduino-cli ARGS
```

You may want to copy the binary into a directory which is in your `PATH` environment variable
(such as `/usr/local/bin/`) or add the binary's directory to it.

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

#### Contributing

To contribute to this project:

* `git clone` this repository.
* Create a new branch with the name `feature-to-implement` or `bug-to-fix`.
* Code your contribution and push to your branch.
* Do a Pull Request.
