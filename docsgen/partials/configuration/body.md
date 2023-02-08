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
