Discovery tools are a special kind of tool used to find supported boards, a platform developer can create their own
following the specification below. These tools must be in the form of executables that can be launched as a subprocess
using a `platform.txt` command line recipe. They communicate to the parent process via stdin/stdout, accepting commands
as plain text strings from stdin and sending answers back in JSON format on stdout. Each tool will implement the
commands to list and enumerate ports for a specific protocol as specified in this document.

### Pluggable discovery API via stdin/stdout

All the commands listed in this specification must be implemented in the discovery.

After startup, the tool will just stay idle waiting for commands. The available commands are: `HELLO`, `START`, `STOP`,
`QUIT`, `LIST` and `START_SYNC`.

After each command the client always expects a response from the discovery. The discovery must not introduce any delay
and must respond to all commands as fast as possible.

#### HELLO command

`HELLO` **must be the first command sent** to the discovery to tell the name of the client/IDE and the version of the
pluggable discovery protocol that the client/IDE supports. The syntax of the command is:

`HELLO <PROTOCOL_VERSION> "<USER_AGENT>"`

- `<PROTOCOL_VERSION>` is the maximum protocol version supported by the client/IDE (currently `1`)
- `<USER_AGENT>` is the name and version of the client. It must not contain double-quotes (`"`).

some examples:

- `HELLO 1 "Arduino IDE 1.8.13"`

- `HELLO 1 "arduino-cli 1.2.3"`

the response to the command is:

```JSON
{
  "eventType": "hello",
  "protocolVersion": 1,
  "message": "OK"
}
```

The `protocolVersion` field represents the protocol version that will be used in the rest of the communication. There
are three possible cases:

- if the client/IDE supports the same or a more recent version of the protocol than the discovery, then the client/IDE
  should go into a compatibility mode and use the protocol level supported by the discovery.
- if the discovery supports a more recent version of the protocol than the client/IDE: the discovery should downgrade
  itself into compatibility mode and report a `protocolVersion` that is less than or equal to the one supported by the
  client/IDE.
- if the discovery cannot go into compatibility mode, it must report the protocol version supported (even if greater
  than the version supported by the client/IDE) and the client/IDE may decide to terminate the discovery or produce an
  error/warning.

#### START command

The `START` command initializes and starts the discovery internal subroutines. This command must be called before
`LIST`. The response to the start command is:

```JSON
{
  "eventType": "start",
  "message": "OK"
}
```

If the discovery could not start, for any reason, it must report the error with:

```JSON
{
  "eventType": "start",
  "error": true,
  "message": "Permission error"
}
```

The `error` field must be set to `true` and the `message` field should contain a description of the error.

#### STOP command

The `STOP` command stops the discovery internal subroutines and possibly frees the internally used resources. This
command should be called if the client wants to pause the discovery for a while. The response to the command is:

```JSON
{
  "eventType": "stop",
  "message": "OK"
}
```

If an error occurs:

```JSON
{
  "eventType": "stop",
  "error": true,
  "message": "Resource busy"
}
```

The `error` field must be set to `true` and the `message` field should contain a description of the error.

#### QUIT command

The `QUIT` command terminates the discovery. The response to `QUIT` is:

```JSON
{
  "eventType": "quit",
  "message": "OK"
}
```

after this output the discovery exits. This command is supposed to always succeed.

#### LIST command

The `LIST` command executes an enumeration of the ports and returns a list of the available ports at the moment of the
call. The format of the response is the following:

```
{
  "eventType": "list",
  "ports": [
    {
      "address":       <-- THE ADDRESS OF THE PORT
      "label":         <-- HOW THE PORT IS DISPLAYED ON THE GUI
      "protocol":      <-- THE PROTOCOL USED BY THE BOARD
      "protocolLabel": <-- HOW THE PROTOCOL IS DISPLAYED ON THE GUI
      "properties": {
                       <-- A LIST OF PROPERTIES OF THE PORT
      }
    },
    {
      ...              <-- OTHER PORTS...
    }
  ]
}
```

The `ports` field contains a list of the available ports.

Each port has:

- an `address` (for example `/dev/ttyACM0` for serial ports or `192.168.10.100` for network ports)
- a `label` that is the human readable form of the `address` (it may be for example `ttyACM0` or
  `SSH on 192.168.10.100`)
- `protocol` is the protocol identifier (such as `serial` or `dfu` or `ssh`)
- `protocolLabel` is the `protocol` in human readable form (for example `Serial port` or `DFU USB` or `Network (ssh)`)
- `properties` is a list of key/value pairs that represent information relative to the specific port

To make the above more clear let's show an example output from the `serial-discovery` builtin in the Arduino CLI:

```JSON
{
  "eventType": "list",
  "ports": [
    {
      "address": "/dev/ttyACM0",
      "label": "ttyACM0",
      "protocol": "serial",
      "protocolLabel": "Serial Port (USB)",
      "properties": {
        "pid": "0x804e",
        "vid": "0x2341",
        "serialNumber": "EBEABFD6514D32364E202020FF10181E",
        "name": "ttyACM0"
      }
    }
  ]
}
```

In this case the serial port metadata comes from a USB serial converter. Inside the `properties` we have all the
properties of the port, and some of them may be useful for product identification (in this case only USB VID/PID is
useful to identify the board model).

The `LIST` command performs a one-shot polling of the ports. The discovery should answer as soon as reasonably possible,
without any additional delay.

Some discoveries may require some time to discover a new port (for example network protocols like MDNS, Bluetooth, etc.
require some seconds to receive the broadcasts from all available clients) in that case it is fine to answer with an
empty or incomplete list.

If an error occurs and the discovery can't complete the enumeration, it must report the error with:

```JSON
{
  "eventType": "list",
  "error": true,
  "message": "Resource busy"
}
```

The `error` field must be set to `true` and the `message` field should contain a description of the error.

#### START_SYNC command

The `START_SYNC` command puts the tool in "events" mode: the discovery will send `add` and `remove` events each time a
new port is detected or removed respectively. If the discovery goes into "events" mode successfully the response to this
command is:

```JSON
{
  "eventType": "start_sync",
  "message": "OK"
}
```

After this message the discovery will send `add` and `remove` events asynchronously (more on that later). If an error
occurs and the discovery can't go in "events" mode the error must be reported as:

```JSON
{
  "eventType": "start_sync",
  "error": true,
  "message": "Resource busy"
}
```

The `error` field must be set to `true` and the `message` field should contain a description of the error.

Once in "event" mode, the discovery is allowed to send `add` and `remove` messages asynchronously in realtime, this
means that the client must be able to handle these incoming messages at any moment.

The `add` event looks like the following:

```JSON
{
  "eventType": "add",
  "port": {
    "address": "/dev/ttyACM0",
    "label": "ttyACM0",
    "properties": {
      "pid": "0x804e",
      "vid": "0x2341",
      "serialNumber": "EBEABFD6514D32364E202020FF10181E",
      "name": "ttyACM0"
    },
    "protocol": "serial",
    "protocolLabel": "Serial Port (USB)"
  }
}
```

It basically provides the same information as the `list` event but for a single port. After calling `START_SYNC` an
initial burst of add events must be generated in sequence to report all the ports available at the moment of the start.

The `remove` event looks like the following:

```JSON
{
  "eventType": "remove",
  "port": {
    "address": "/dev/ttyACM0",
    "protocol": "serial"
  }
}
```

The content is straightforward, in this case only the `address` and `protocol` fields are reported.

If the information about a port needs to be updated the discovery may send a new `add` message for the same port address
and protocol without sending a `remove` first: this means that all the previous information about the port must be
discarded and replaced with the new one.

#### Invalid commands

If the client sends an invalid or malformed command, the discovery should answer with:

```JSON
{
  "eventType": "command_error",
  "error": true,
  "message": "Unknown command XXXX"
}
```

### State machine

A well behaved pluggable discovery tool must reflect the following state machine.

![Pluggable discovery state machine](img/pluggable-discovery-state-machine.png)

The arrows represent the commands outlined in the above sections, calling a command successfully assumes the state
changes.

A pluggable discovery state is Alive when the process has been started but no command has been executed. Dead means the
process has been stopped and no further commands can be received.

### Board identification

The `properties` associated to a port can be used to identify the board attached to that port. The algorithm is simple:

- each board listed in the platform file [`boards.txt`](platform-specification.md#boardstxt) may declare a set of
  `upload_port.*` properties
- if each `upload_port.*` property has a match in the `properties` set coming from the discovery then the board is a
  "candidate" board attached to that port.

Some port `properties` may not be precise enough to uniquely identify a board, in that case more boards may match the
same set of `properties`, that's why we called it "candidate".

Let's see an example to clarify things a bit, let's suppose that we have the following `properties` coming from the
serial discovery:

```
  "port": {
    "address": "/dev/ttyACM0",
    "properties": {
      "pid": "0x804e",
      "vid": "0x2341",
      "serialNumber": "EBEABFD6514D32364E202020FF10181E",
      "name": "ttyACM0"
    },
    ...
```

in this case we can use `vid` and `pid` to identify the board. The `serialNumber`, instead, is unique for that specific
instance of the board so it can't be used to identify the board model. Let's suppose we have the following `boards.txt`:

```
# Arduino Zero (Programming Port)
# ---------------------------------------
arduino_zero_edbg.name=Arduino Zero (Programming Port)
arduino_zero_edbg.upload_port.vid=0x03eb
arduino_zero_edbg.upload_port.pid=0x2157
[...CUT...]
# Arduino Zero (Native USB Port)
# --------------------------------------
arduino_zero_native.name=Arduino Zero (Native USB Port)
arduino_zero_native.upload_port.0.vid=0x2341
arduino_zero_native.upload_port.0.pid=0x804d
arduino_zero_native.upload_port.1.vid=0x2341
arduino_zero_native.upload_port.1.pid=0x004d
arduino_zero_native.upload_port.2.vid=0x2341
arduino_zero_native.upload_port.2.pid=0x824d
arduino_zero_native.upload_port.3.vid=0x2341
arduino_zero_native.upload_port.3.pid=0x024d
[...CUT...]
# Arduino MKR1000
# -----------------------
mkr1000.name=Arduino MKR1000
mkr1000.upload_port.0.vid=0x2341       <------- MATCHING IDs
mkr1000.upload_port.0.pid=0x804e       <------- MATCHING IDs
mkr1000.upload_port.1.vid=0x2341
mkr1000.upload_port.1.pid=0x004e
mkr1000.upload_port.2.vid=0x2341
mkr1000.upload_port.2.pid=0x824e
mkr1000.upload_port.3.vid=0x2341
mkr1000.upload_port.3.pid=0x024e
[...CUT...]
```

As we can see the only board that has the two properties matching is the `mkr1000`, in this case the CLI knows that the
board is surely an MKR1000.

Note that `vid` and `pid` properties are just free text key/value pairs: the discovery may return basically anything,
the board just needs to have the same properties defined in `boards.txt` as `upload_port.*` to be identified.

We can also specify multiple identification properties for the same board using the `.N` suffix, for example:

```
myboard.name=My Wonderful Arduino Compatible Board
myboard.upload_port.pears=20
myboard.upload_port.apples=30
```

will match on `pears=20, apples=30` but:

```
myboard.name=My Wonderful Arduino Compatible Board
myboard.upload_port.0.pears=20
myboard.upload_port.0.apples=30
myboard.upload_port.1.pears=30
myboard.upload_port.1.apples=40
```

will match on both `pears=20, apples=30` and `pears=30, apples=40` but not `pears=20, apples=40`, in that sense each
"set" of identification properties is independent from each other and cannot be mixed for port matching.

#### Identification of board options

[Custom board options](platform-specification.md#custom-board-options) can also be identified.

Identification property values are associated with a custom board option by the board definition in
[`boards.txt`](platform-specification.md#boardstxt). Two formats are available.

If only a single set of identification properties are associated with the option:

```
BOARD_ID.menu.MENU_ID.OPTION_ID.upload_port.PORT_PROPERTY_KEY=PORT_PROPERTY_VALUE
```

If one or more sets of identification properties are associated with the option, an index number is used for each set:

```
BOARD_ID.menu.MENU_ID.OPTION_ID.upload_port.SET_INDEX.PORT_PROPERTY_KEY=PORT_PROPERTY_VALUE
```

If multiple identification properties are associated within a set, all must match for the option to be identified.

Let's see an example to clarify it, in the following `boards.txt`:

```
myboard.upload_port.pid=0x0010
myboard.upload_port.vid=0x2341
myboard.menu.cpu.atmega1280=ATmega1280
myboard.menu.cpu.atmega1280.upload_port.c=atmega1280          <--- identification property for cpu=atmega1280
myboard.menu.cpu.atmega1280.build_cpu=atmega1280
myboard.menu.cpu.atmega2560=ATmega2560
myboard.menu.cpu.atmega2560.upload_port.c=atmega2560          <--- identification property for cpu=atmega2560
myboard.menu.cpu.atmega2560.build_cpu=atmega2560
myboard.menu.mem.1k=1KB
myboard.menu.mem.1k.upload_port.mem=1                         <--- identification property for mem=1k
myboard.menu.mem.1k.build_mem=1024
myboard.menu.mem.2k=2KB
myboard.menu.mem.2k.upload_port.1.mem=2                       <------ identification property for mem=2k (case 1)
myboard.menu.mem.2k.upload_port.2.ab=ef                       <---\
myboard.menu.mem.2k.upload_port.2.cd=gh                       <---+-- identification property for mem=2k (case 2)
myboard.menu.mem.2k.build_mem=2048
```

we have a board called `myboard` with two custom menu options `cpu` and `mem`.

A port with the following identification properties:

```
vid=0x0010
pid=0x2341
c=atmega2560
```

will be identified as FQBN `mypackage:avr:myboard:cpu=atmega2560` because of the property `c=atmega2560`.

A port with the following identification properties:

```
vid=0x0010
pid=0x2341
c=atmega2560
mem=2
```

will be identified as FQBN `mypackage:avr:myboard:cpu=atmega2560,mem=2k`.

A port with the following identification properties:

```
vid=0x0010
pid=0x2341
c=atmega2560
ab=ef
cd=gh
```

will be identified as FQBN `mypackage:avr:myboard:cpu=atmega2560,mem=2k` too (they will match the second identification
properties set for `mem=2k`).
