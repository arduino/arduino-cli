## Configuration keys

* `board_manager`
    * `additional_urls` - the URLs to any additional Board Manager package index
    files needed for your boards platforms.
* `daemon` - options related to running Arduino CLI as a [gRPC] server.
    * `port` - TCP port used for gRPC client connections.
* `directories` - directories used by Arduino CLI.
    * `data` - directory used to store Board/Library Manager index files and
    Board Manager platform installations.
    * `downloads` - directory used to stage downloaded archives during
    Board/Library Manager installations.
    * `user` - the equivalent of the Arduino IDE's
    ["sketchbook" directory][sketchbook directory]. Library Manager
    installations are made to the `libraries` subdirectory of the user
    directory.
* `logging` - configuration options for Arduino CLI's logs.
    * `file` - path to the file where logs will be written.
    * `format` - output format for the logs. Allowed values are `text` or
    `json`.
    * `level` - messages with this level and above will be logged. Valid levels
    are: `trace`, `debug`, `info`, `warn`, `error`, `fatal`, `panic`.
* `telemetry` - settings related to the collection of data used for continued
improvement of Arduino CLI.
    * `addr` - TCP port used for telemetry communication.
    * `enabled` - controls the use of telemetry.

## Configuration methods

Arduino CLI may be configured in three ways:

1. Command line flags
1. Environment variables
1. Configuration file

If a configuration option is configured by multiple methods, the value set by
the method highest on the above list overwrites the ones below it.

If a configuration option is not set, Arduino CLI uses a default value.

[`arduino-cli config dump`][arduino-cli config dump] displays the current
configuration values.

### Command line flags

Arduino CLI's command line flags are documented in the command line help and the
[Arduino CLI command reference].

#### Example

Setting an additional Board Manager URL using the
[`--additional-urls`][arduino-cli global flags] command line flag:

```shell
$ arduino-cli core update-index --additional-urls https://downloads.arduino.cc/packages/package_staging_index.json
```

### Environment variables

All configuration options can be set via environment variables. The variable
names start with `ARDUINO`, followed by the configuration key names, with each
component separated by `_`. For example, the `ARDUINO_DIRECTORIES_USER`
environment variable sets the `directories.user` configuration option.

On Linux or macOS, you can use the [`export` command][export command] to set
environment variables. On Windows cmd, you can use the
[`set` command][set command].

#### Example

Setting an additional Board Manager URL using the
`ARDUINO_BOARD_MANAGER_ADDITIONAL_URLS` environment variable:

```sh
$ export ARDUINO_BOARD_MANAGER_ADDITIONAL_URLS=https://downloads.arduino.cc/packages/package_staging_index.json
```

### Configuration file

[`arduino-cli config init`][arduino-cli config init] creates or updates a
configuration file with the current configuration settings.

This allows saving the options set by command line flags or environment
variables. For example:

```sh
arduino-cli config init --additional-urls https://downloads.arduino.cc/packages/package_staging_index.json
```

#### File name

The configuration file must be named `arduino-cli`, with the appropriate file
extension for the file's format.

#### Supported formats

`arduino-cli config init` creates a YAML file, however a variety of common
formats are supported:

* [JSON]
* [TOML]
* [YAML]
* [Java properties file]
* [HCL]
* envfile
* [INI]

#### Locations

Configuration files in the following locations are recognized by Arduino CLI:

1. Location specified by the [`--config-file`][Arduino CLI command reference]
command line flag
1. Current working directory
1. Any parent directory of the current working directory (more immediate parents
having higher precedence)
1. Arduino CLI data directory (as configured by `directories.data`)

If multiple configuration files are present, the one highest on the above list
is used. Configuration files are not combined.

The location of the active configuration file can be determined by running the
command:

```sh
arduino-cli config dump --verbose
```

#### Example

Setting an additional Board Manager URL using a YAML format configuration file:

```yaml
board_manager:
  additional_urls:
  - https://downloads.arduino.cc/packages/package_staging_index.json
```

Doing the same using a TOML format file:

```toml
[board_manager]
additional_urls = [ "https://downloads.arduino.cc/packages/package_staging_index.json" ]
```


[gRPC]: https://grpc.io
[sketchbook directory]: sketch-specification.md#sketchbook
[arduino-cli config dump]: ../commands/arduino-cli_config_dump
[Arduino CLI command reference]: ../commands/arduino-cli
[arduino-cli global flags]: ../commands/arduino-cli_config/#options-inherited-from-parent-commands
[export command]: https://ss64.com/bash/export.html
[set command]: https://docs.microsoft.com/en-us/windows-server/administration/windows-commands/set_1
[arduino-cli config init]: ../commands/arduino-cli_config_init
[JSON]: https://www.json.org
[TOML]: https://github.com/toml-lang/toml
[YAML]: https://en.wikipedia.org/wiki/YAML
[Java properties file]: https://en.wikipedia.org/wiki/.properties
[HCL]: https://github.com/hashicorp/hcl
[INI]: https://en.wikipedia.org/wiki/INI_file
