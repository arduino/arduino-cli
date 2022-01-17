## Configuration keys

- `board_manager`
  - `additional_urls` - the URLs to any additional Boards Manager package index files needed for your boards platforms.
- `directories` - directories used by Arduino CLI.
  - `data` - directory used to store Boards/Library Manager index files and Boards Manager platform installations.
  - `downloads` - directory used to stage downloaded archives during Boards/Library Manager installations.
  - `user` - the equivalent of the Arduino IDE's ["sketchbook" directory][sketchbook directory]. Library Manager
    installations are made to the `libraries` subdirectory of the user directory.
- `library` - configuration options relating to Arduino libraries.
  - `enable_unsafe_install` - set to `true` to enable the use of the `--git-url` and `--zip-file` flags with
    [`arduino-cli lib install`][arduino cli lib install]. These are considered "unsafe" installation methods because
    they allow installing files that have not passed through the Library Manager submission process.
- `logging` - configuration options for Arduino CLI's logs.
  - `file` - path to the file where logs will be written.
  - `format` - output format for the logs. Allowed values are `text` or `json`.
  - `level` - messages with this level and above will be logged. Valid levels are: `trace`, `debug`, `info`, `warn`,
    `error`, `fatal`, `panic`.
- `metrics` - settings related to the collection of data used for continued improvement of Arduino CLI.
  - `addr` - TCP port used for metrics communication.
  - `enabled` - controls the use of metrics.
- `sketch` - configuration options relating to [Arduino sketches][sketch specification].
  - `always_export_binaries` - set to `true` to make [`arduino-cli compile`][arduino-cli compile] always save binaries
    to the sketch folder. This is the equivalent of using the [`--export-binaries`][arduino-cli compile options] flag.
- `updater` - configuration options related to Arduino CLI updates
  - `enable_notification` - set to `false` to disable notifications of new Arduino CLI releases, defaults to `true`

## Configuration methods

Arduino CLI may be configured in three ways:

1. Command line flags
1. Environment variables
1. Configuration file

If a configuration option is configured by multiple methods, the value set by the method highest on the above list
overwrites the ones below it.

If a configuration option is not set, Arduino CLI uses a default value.

[`arduino-cli config dump`][arduino-cli config dump] displays the current configuration values.

### Command line flags

Arduino CLI's command line flags are documented in the command line help and the [Arduino CLI command reference].

#### Example

Setting an additional Boards Manager URL using the [`--additional-urls`][arduino-cli global flags] command line flag:

```shell
$ arduino-cli core update-index --additional-urls https://downloads.arduino.cc/packages/package_staging_index.json
```

### Environment variables

All configuration options can be set via environment variables. The variable names start with `ARDUINO`, followed by the
configuration key names, with each component separated by `_`. For example, the `ARDUINO_DIRECTORIES_USER` environment
variable sets the `directories.user` configuration option.

On Linux or macOS, you can use the [`export` command][export command] to set environment variables. On Windows cmd, you
can use the [`set` command][set command].

#### Example

Setting an additional Boards Manager URL using the `ARDUINO_BOARD_MANAGER_ADDITIONAL_URLS` environment variable:

```sh
$ export ARDUINO_BOARD_MANAGER_ADDITIONAL_URLS=https://downloads.arduino.cc/packages/package_staging_index.json
```

### Configuration file

[`arduino-cli config init`][arduino-cli config init] creates or updates a configuration file with the current
configuration settings.

This allows saving the options set by command line flags or environment variables. For example:

```sh
arduino-cli config init --additional-urls https://downloads.arduino.cc/packages/package_staging_index.json
```

#### File name

The configuration file must be named `arduino-cli`, with the appropriate file extension for the file's format.

#### Supported formats

`arduino-cli config init` creates a YAML file, however a variety of common formats are supported:

- [JSON]
- [TOML]
- [YAML]
- [Java properties file]
- [HCL]
- envfile
- [INI]

#### Locations

Configuration files in the following locations are recognized by Arduino CLI:

1. Location specified by the [`--config-file`][arduino cli command reference] command line flag
1. Current working directory
1. Any parent directory of the current working directory (more immediate parents having higher precedence)
1. Arduino CLI data directory (as configured by `directories.data`)

If multiple configuration files are present, the one highest on the above list is used. Configuration files are not
combined.

The location of the active configuration file can be determined by running the command:

```sh
arduino-cli config dump --verbose
```

#### Example

Setting an additional Boards Manager URL using a YAML format configuration file:

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

## Daemon

The `daemon` mode may be configured in two ways:

1. Command line flags
1. Configuration file

If a configuration option is configured by multiple methods, the value set by the method highest on the above list
overwrites the ones below it.

If a configuration option is not set, Arduino CLI uses a default value.

### Command line flags

Arduino CLI's `daemon` mode flags are documented in the [Arduino CLI `daemon` command reference][arduino-cli daemon].

#### Example

Starting the `daemon` mode with a different port using the [`--port`][arduino-cli daemon flags] flag

```shell
$ arduino-cli daemon --port 12345
```

<a id="daemon-configuration-file"></a>

### Configuration file

The `arduino-cli daemon` mode uses a different configuration file from the other commands and it must be managed by the
user that starts the `daemon` process. The Arduino CLI doesn't offer any way of managing those configurations and won't
search for a configuration file if not explicitly specified by the user.

It can be easily set like this:

```sh
arduino-cli daemon --config-file /path/to/settings/daemon-settings.yaml
```

#### Keys

The `arduino-cli daemon` mode is completely separated from the command line settings and it uses a different
configuration from the other commands:

- `ip`: IP used to listen for gRPC connections
- `port`: Port used listen for gRPC connections
- `daemonize`: True to run daemon process in background
- `debug`: True to enable debug logging of gRPC calls
- `debug-filter`: List of gRPC calls to log when in debug mode
- `verbose`: True to print logs in stdout
- `format`: Stdout output format
- `no-color`: True to disable color output to stdout and stderr
- `log-level`: Messages with this level and above will be logged
- `log-file`: Path to the file where logs will be written
- `log-format`: Output format for the logs

#### File name

The configuration file name can be anything but the file extension must be appropriate for the file's format.

#### Supported formats

The `daemon` mode `--config-file` flag supports a variety of common formats much like when running the Arduino CLI as a
command line tool:

- [JSON]
- [TOML]
- [YAML]
- [Java properties file]
- [HCL]
- envfile
- [INI]

#### Example

Setting a custom port with debug enabled and custom filters using a YAML format configuration file:

```yaml
port: "12345"
debug: true
debug-filter:
  - "Create"
  - "PlatformInstall"
  - "UpdateIndex"
```

Doing the same using a TOML format file:

```toml
port = "12345"
debug = true
debug-filter = [
    "Create",
    "PlatformInstall",
    "UpdateIndex"
]
```

[grpc]: https://grpc.io
[sketchbook directory]: sketch-specification.md#sketchbook
[arduino cli lib install]: commands/arduino-cli_lib_install.md
[sketch specification]: sketch-specification.md
[arduino-cli compile]: commands/arduino-cli_compile.md
[arduino-cli compile options]: commands/arduino-cli_compile.md#options
[arduino-cli config dump]: commands/arduino-cli_config_dump.md
[arduino cli command reference]: commands/arduino-cli.md
[arduino-cli global flags]: commands/arduino-cli_config.md#options-inherited-from-parent-commands
[export command]: https://ss64.com/bash/export.html
[set command]: https://docs.microsoft.com/en-us/windows-server/administration/windows-commands/set_1
[arduino-cli config init]: commands/arduino-cli_config_init.md
[json]: https://www.json.org
[toml]: https://github.com/toml-lang/toml
[yaml]: https://en.wikipedia.org/wiki/YAML
[java properties file]: https://en.wikipedia.org/wiki/.properties
[hcl]: https://github.com/hashicorp/hcl
[ini]: https://en.wikipedia.org/wiki/INI_file
[arduino-cli daemon]: commands/arduino-cli_daemon.md
[arduino-cli daemon flags]: commands/arduino-cli_daemon.md#options
