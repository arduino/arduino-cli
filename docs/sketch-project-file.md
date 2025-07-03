Sketch metadata is defined in a file named `sketch.yaml`. This file is in YAML format.

## Build profiles

Arduino CLI provides support for reproducible builds through the use of build profiles.

A profile is a complete description of all the resources needed to build a sketch. The sketch project file may contain
multiple profiles.

Each profile will define:

- The board FQBN
- The programmer to use
- The target core platform name and version (with the 3rd party platform index URL if needed)
- A possible core platform name and version, that is a dependency of the target core platform (with the 3rd party
  platform index URL if needed)
- A list of libraries used in the sketch. Each library could be:
  - a library taken from the Arduino Libraries Index
  - a library installed anywhere in the filesystem
- The port and protocol to upload the sketch and monitor the board

The format of the file is the following:

```
profiles:
  <PROFILE_NAME>:
    notes: <USER_NOTES>
    fqbn: <FQBN>
    programmer: <PROGRAMMER>
    platforms:
      - platform: <PLATFORM> [(<PLATFORM_VERSION>)]
        platform_index_url: <3RD_PARTY_PLATFORM_URL>
      - platform: <PLATFORM_DEPENDENCY> [(<PLATFORM_DEPENDENCY_VERSION>)]
        platform_index_url: <3RD_PARTY_PLATFORM_DEPENDENCY_URL>
    libraries:
      - <INDEX_LIB_NAME> (<INDEX_LIB_VERSION>)
      - dir: <LOCAL_LIB_PATH>
    port: <PORT_NAME>
    port_config:
      <PORT_SETTING_NAME>: <PORT_SETTING_VALUE>
      ...
    protocol: <PORT_PROTOCOL>
  ...more profiles here...
```

There is an optional `profiles:` section containing all the profiles. Each field in a profile is mandatory (unless noted
otherwise below). The available fields are:

- `<PROFILE_NAME>` is the profile identifier, it’s a user-defined field, and the allowed characters are alphanumerics,
  underscore `_`, dot `.`, and dash `-`.
- `<PLATFORM>` is the target core platform identifier, for example, `arduino:avr` or `adafruit:samd`.
- `<PLATFORM_VERSION>` is the target core platform version required.
- `<3RD_PARTY_PLATFORM_URL>` is the index URL to download the target core platform (also known as “Additional Boards
  Manager URLs” in the Arduino IDE). This field can be omitted for the official `arduino:*` platforms.
- `<PLATFORM_DEPENDENCY>`, `<PLATFORM_DEPENDENCY_VERSION>`, and `<3RD_PARTY_PLATFORM_DEPENDENCY_URL>` contains the same
  information as `<PLATFORM>`, `<PLATFORM_VERSION>`, and `<3RD_PARTY_PLATFORM_URL>` respectively but for the core
  platform dependency of the main core platform. These fields are optional.
- `libraries:` is a section where the required libraries to build the project are defined. This section is optional.
  - `<INDEX_LIB_NAME> (<INDEX_LIB_VERSION>)` represents a library from the Arduino Libraries Index, for example,
    `MyLib (1.0.0)`.
  - `dir: <LOCAL_LIB_PATH>` represents a library installed in the filesystem and `<LOCAL_LIB_PATH>` is the path to the
    library. The path could be absolute or relative to the sketch folder. This option is available since Arduino CLI
    1.3.0.
- `<USER_NOTES>` is a free text string available to the developer to add comments. This field is optional.
- `<PROGRAMMER>` is the programmer that will be used. This field is optional.

The following fields are available since Arduino CLI 1.1.0:

- `<PORT_NAME>` is the port that will be used to upload and monitor the board (unless explicitly set otherwise). This
  field is optional.
- `port_config` section with `<PORT_SETTING_NAME>` and `<PORT_SETTING_VALUE>` defines the port settings that will be
  used in the `monitor` command. Typically is used to set the baudrate for the serial port (for example
  `baudrate: 115200`) but any setting/value can be specified. Multiple settings can be set. These fields are optional.
- `<PORT_PROTOCOL>` is the protocol for the port used to upload and monitor the board. This field is optional.

#### Using a system-installed platform.

The fields `<PLATFORM_VERSION>` and `<PLATFORM_DEPENDENCY_VERSION>` are optional, if they are omitted, the sketch
compilation will use the platforms installed system-wide. This could be helpful during the development of a platform
(where a specific release is not yet available), or if a specific version of a platform is not a strict requirement.

#### An example of a complete project file.

A complete example of a sketch project file may be the following:

```
profiles:
  nanorp:
    fqbn: arduino:mbed_nano:nanorp2040connect
    platforms:
      - platform: arduino:mbed_nano (2.1.0)
    libraries:
      - ArduinoIoTCloud (1.0.2)
      - Arduino_ConnectionHandler (0.6.4)
      - TinyDHT sensor library (1.1.0)

  another_profile_name:
    notes: testing the limit of the AVR platform, may be unstable
    fqbn: arduino:avr:uno
    platforms:
      - platform: arduino:avr (1.8.4)
    libraries:
      - VitconMQTT (1.0.1)
      - Arduino_ConnectionHandler (0.6.4)
      - TinyDHT sensor library (1.1.0)
    port: /dev/ttyACM0
    port_config:
      baudrate: 115200

  tiny:
    notes: testing the very limit of the AVR platform, it will be very unstable
    fqbn: attiny:avr:ATtinyX5:cpu=attiny85,clock=internal16
    platforms:
      - platform: attiny:avr (1.0.2)
        platform_index_url: https://raw.githubusercontent.com/damellis/attiny/ide-1.6.x-boards-manager/package_damellis_attiny_index.json
      - platform: arduino:avr (1.8.3)
    libraries:
      - ArduinoIoTCloud (1.0.2)
      - Arduino_ConnectionHandler (0.6.4)
      - TinyDHT sensor library (1.1.0)

  feather:
    fqbn: adafruit:samd:adafruit_feather_m0
    platforms:
      - platform: adafruit:samd (1.6.0)
        platform_index_url: https://adafruit.github.io/arduino-board-index/package_adafruit_index.json
    libraries:
      - ArduinoIoTCloud (1.0.2)
      - Arduino_ConnectionHandler (0.6.4)
      - TinyDHT sensor library (1.1.0)

default_profile: nanorp
```

### Building a sketch

When a sketch project file is present, it can be leveraged to compile the sketch with the `--profile/-m` flag in the
`compile` command:

```
arduino-cli compile --profile nanorp
```

In this case, the sketch will be compiled using the core platform and libraries specified in the nanorp profile. If a
core platform or a library is missing it will be automatically downloaded and installed on the fly in an isolated
directory inside the data folder. The dedicated storage is not accessible to the user and is meant as a "cache" of the
resources used to build the sketch.

When using the profile-based build, the globally installed platforms and libraries are excluded from the compile and can
not be used in any way. In other words, the build is isolated from the system and will rely only on the resources
specified in the profile: this will ensure that the build is portable and reproducible independently from the platforms
and libraries installed in the system.

### Using a default profile

If a `default_profile` is specified in the `sketch.yaml` then the “classic” compile command:

```
arduino-cli compile [sketch]
```

will, instead, trigger a profile-based build using the default profile indicated in the `sketch.yaml`.

## Default flags for Arduino CLI usage

The sketch project file may be used to set the default value for some command line flags of the Arduino CLI, in
particular:

- The `default_fqbn` key sets the default value for the `--fqbn` flag
- The `default_programmer` key sets the default value for the `--programmer` flag
- The `default_port` key sets the default value for the `--port` flag
- The `default_port_config` key sets the default values for the `--config` flag in the `monitor` command (available
  since Arduino CLI 1.1.0)
- The `default_protocol` key sets the default value for the `--protocol` flag
- The `default_profile` key sets the default value for the `--profile` flag

For example:

```
default_fqbn: arduino:samd:mkr1000
default_programmer: atmel_ice
default_port: /dev/ttyACM0
default_port_config:
  baudrate: 115200
default_protocol: serial
default_profile: myprofile
```

With this configuration set, it is not necessary to specify the `--fqbn`, `--programmer`, `--port`, `--protocol` or
`--profile` flags to the [`arduino-cli compile`](commands/arduino-cli_compile.md),
[`arduino-cli upload`](commands/arduino-cli_upload.md) or [`arduino-cli debug`](commands/arduino-cli_debug.md) commands
when compiling, uploading or debugging the sketch. Moreover in the `monitor` command it is not necessary to specify the
`--config baudrate=115200` to communicate with the monitor port of the board.
