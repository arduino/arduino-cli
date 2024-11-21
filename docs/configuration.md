## Configuration keys

- `board_manager`
  - `additional_urls` - the URLs to any additional Boards Manager package index files needed for your boards platforms.
- `daemon` - options related to running Arduino CLI as a [gRPC] server.
  - `port` - TCP port used for gRPC client connections.
- `directories` - directories used by Arduino CLI.
  - `data` - directory used to store Boards/Library Manager index files and Boards Manager platform installations.
  - `downloads` - directory used to stage downloaded archives during Boards/Library Manager installations.
  - `user` - the equivalent of the Arduino IDE's ["sketchbook" directory][sketchbook directory]. Library Manager
    installations are made to the `libraries` subdirectory of the user directory. Users can manually install 3rd party
    platforms in the `hardware` subdirecotry of the user directory.
  - `builtin.libraries` - the libraries in this directory will be available to all platforms without the need for the
    user to install them, but with the lowest priority over other installed libraries with the same name, it's the
    equivalent of the Arduino IDE's bundled libraries directory.
- `library` - configuration options relating to Arduino libraries.
  - `enable_unsafe_install` - set to `true` to enable the use of the `--git-url` and `--zip-file` flags with
    [`arduino-cli lib install`][arduino cli lib install]. These are considered "unsafe" installation methods because
    they allow installing files that have not passed through the Library Manager submission process.
- `locale` - the language used by Arduino CLI to communicate to the user, the parameter is the language identifier in
  the standard POSIX format `<language>[_<TERRITORY>[.<encoding>]]` (for example `it` or `it_IT`, or `it_IT.UTF-8`).
- `logging` - configuration options for Arduino CLI's logs.
  - `file` - path to the file where logs will be written.
  - `format` - output format for the logs. Allowed values are `text` or `json`.
  - `level` - messages with this level and above will be logged. Valid levels are: `trace`, `debug`, `info`, `warn`,
    `error`, `fatal`, `panic`.
- `metrics` - settings related to the collection of data used for continued improvement of Arduino CLI.
  - `addr` - TCP port used for metrics communication.
  - `enabled` - controls the use of metrics.
- `output` - settings related to text output.
  - `no_color` - ANSI color escape codes are added by default to the output. Set to `true` to disable colored text
    output.
- `sketch` - configuration options relating to [Arduino sketches][sketch specification].
  - `always_export_binaries` - set to `true` to make [`arduino-cli compile`][arduino-cli compile] always save binaries
    to the sketch folder. This is the equivalent of using the [`--export-binaries`][arduino-cli compile options] flag.
- `updater` - configuration options related to Arduino CLI updates
  - `enable_notification` - set to `false` to disable notifications of new Arduino CLI releases, defaults to `true`
- `build_cache` configuration options related to the compilation cache
  - `path` - the path to the build cache, default is `$TMP/arduino`.
  - `extra_paths` - a list of paths to look for precompiled artifacts if not found on `build_cache.path` setting.
  - `compilations_before_purge` - interval, in number of compilations, at which the cache is purged, defaults to `10`.
    When `0` the cache is never purged.
  - `ttl` - cache expiration time of build folders. If the cache is hit by a compilation the corresponding build files
    lifetime is renewed. The value format must be a valid input for
    [time.ParseDuration()](https://pkg.go.dev/time#ParseDuration), defaults to `720h` (30 days).
- `network` - configuration options related to the network connection.
  - `proxy` - URL of the proxy server.

### Default directories

The following are the default directories selected by the Arduino CLI if alternatives are not specified in the
configuration file.

- The `directories.data` default is OS-dependent:

  - on Linux (and other Unix-based OS) is: `{HOME}/.arduino15`
  - on Windows is: `{HOME}/AppData/Local/Arduino15`
  - on MacOS is: `{HOME}/Library/Arduino15`

- The `directories.download` default is `{directories.data}/staging`. If the value of `{directories.data}` is changed in
  the configuration the user-specified value will be used.

- The `directories.user` default is OS-dependent:
  - on Linux (and other Unix-based OS) is: `{HOME}/Arduino`
  - on Windows is: `{DOCUMENTS}/Arduino`
  - on MacOS is: `{HOME}/Documents/Arduino`

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

`ARDUINO_BOARD_MANAGER_ADDITIONAL_URLS` environment variables can be a list of space-separated URLs.

#### Example

Setting an additional Boards Manager URL using the `ARDUINO_BOARD_MANAGER_ADDITIONAL_URLS` environment variable:

```sh
$ export ARDUINO_BOARD_MANAGER_ADDITIONAL_URLS=https://downloads.arduino.cc/packages/package_staging_index.json
```

Setting multiple additional Boards Manager URLs using the `ARDUINO_BOARD_MANAGER_ADDITIONAL_URLS` environment variable:

```sh
$ export ARDUINO_BOARD_MANAGER_ADDITIONAL_URLS="https://downloads.arduino.cc/packages/package_staging_index.json https://downloads.arduino.cc/packages/package_mbed_index.json"
```

### Configuration file

[`arduino-cli config init`][arduino-cli config init] creates a new empty configuration file.

This allows saving the options set by command line flags or environment variables. For example:

```sh
arduino-cli config init --additional-urls https://downloads.arduino.cc/packages/package_staging_index.json
```

#### Locations

The default configuration file is named `arduino-cli.yaml`. The configuration file is searched in the following
locations, in order of priority:

1. Location specified by the [`--config-file`][arduino cli command reference] command line flag
1. Location specified by the `ARDUINO_CONFIG_FILE` environment variable
1. Location specified by the `ARDUINO_DIRECTORIES_DATA` environment variable

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

#### JSON schema

The configuration file [JSON schema][configuration-schema] can be used to independently validate the file content. This
schema should be considered unstable in this version.

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
[configuration-schema]: ./configuration.schema.json
