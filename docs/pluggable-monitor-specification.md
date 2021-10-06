Monitor tools are a special kind of tool used to let the user communicate with the supported boards. A platform
developer can create their own tools following the specification below. These tools must be in the form of command line
executables that can be launched as a subprocess.

They will communicate to the parent process via stdin/stdout, in particular a monitor tool will accept commands as plain
text strings from stdin and will send answers back in JSON format on stdout. Each tool will implement the commands to
open and control communication ports for a specific protocol as specified in this document. The actual I/O data stream
from the communication port will be transferred to the parent process through a separate channel via TCP/IP.

### Pluggable monitor API via stdin/stdout

All the commands listed in this specification must be implemented in the monitor tool.

After startup, the tool will just stay idle waiting for commands. The available commands are: `HELLO`, `DESCRIBE`,
`CONFIGURE`, `OPEN`, `CLOSE` and `QUIT`.

After each command the client always expects a response from the monitor. The monitor must not introduce any delay and
must respond to all commands as fast as possible.

#### HELLO command

`HELLO` **must be the first command sent** to the monitor to tell the name of the client/IDE and the version of the
pluggable monitor protocol that the client/IDE supports. The syntax of the command is:

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

- if the client/IDE supports the same or a more recent version of the protocol than the monitor tool, then the
  client/IDE should go into a compatibility mode and use the protocol level supported by the monitor tool.
- if the monitor tool supports a more recent version of the protocol than the client/IDE, then the monitor tool should
  downgrade itself into compatibility mode and report a `protocolVersion` that is less than or equal to the one
  supported by the client/IDE.
- if the monitor tool cannot go into compatibility mode, it must report the protocol version supported (even if greater
  than the version supported by the client/IDE) and the client/IDE may decide to terminate the monitor tool or produce
  an error/warning.

#### DESCRIBE command

The `DESCRIBE` command returns a description of the communication port. The description will have metadata about the
port configuration, and which parameters are available to the user.

```JSON
{
  "event": "describe",
  "message": "ok",
  "port_description": {
    "protocol": "serial",
    "configuration_parameters": {
      "baudrate": {
        "label": "Baudrate",
        "type": "enum",
        "values": [
          "300", "600", "750", "1200", "2400", "4800", "9600",
          "19200", "38400", "57600", "115200", "230400", "460800",
          "500000", "921600", "1000000", "2000000"
        ],
        "selected": "9600"
      },
      "parity": {
        "label": "Parity",
        "type": "enum",
        "values": [ "N", "E", "O", "M", "S" ],
        "selected": "N"
      },
      "bits": {
        "label": "Data bits",
        "type": "enum",
        "values": [ "5", "6", "7", "8", "9" ],
        "selected": "8"
      },
      "stop_bits": {
        "label": "Stop bits",
        "type": "enum",
        "values": [ "1", "1.5", "2" ],
        "selected": "1"
      }
    }
  }
}
```

The field `protocol` is the board port protocol identifier, it must match with the corresponding protocol identifier for
a pluggable discovery tool.

`configuration_parameters` is a key/value map that enumerates the available port parameters.

Each parameter has a unique name (`baudrate`, `parity`, etc...), a `type` (in this case only `enum` is allowed but more
types may be added in the future if needed), and the `selected` value for each parameter.

The parameter name can not contain spaces, the allowed characters are alphanumerics, underscore `_`, dot `.`, and dash
`-`.

The `enum` types must have a list of possible `values`.

The client/IDE may expose these configuration values to the user via a config file or a GUI, in this case the `label`
field may be used for a user readable description of the parameter.

#### CONFIGURE command

The `CONFIGURE` command sets configuration parameters for the communication port. The parameters can be changed one at a
time and the syntax is:

`CONFIGURE <PARAMETER_NAME> <VALUE>`

The response to the command is:

```JSON
{
  "event": "configure",
  "message": "ok",
}
```

or if there is an error:

```JSON
{
  "event": "configure",
  "error": true,
  "message": "invalid value for parameter baudrate: 123456"
}
```

The currently selected parameters or their default value may be obtained using the `DESCRIBE` command.

#### OPEN command

The `OPEN` command opens a communication port with the board, the data exchanged with the board will be transferred to
the Client/IDE via TCP/IP.

The Client/IDE must first TCP-Listen to a randomly selected TCP port and send the address to connect it to the monitor
tool as part of the `OPEN` command. The syntax of the `OPEN` command is:

`OPEN <CLIENT_TCPIP_ADDRESS> <BOARD_PORT>`

For example, let's suppose that the Client/IDE wants to communicate with the serial port `/dev/ttyACM0` using an
hypotetical `serial-monitor` tool, then the sequence of actions to perform will be the following:

1. the Client/IDE must first listen to a random TCP port (let's suppose it chose `32123`)
1. the Client/IDE runs the `serial-monitor` tool and initialize it with the `HELLO` command
1. the Client/IDE sends the command `OPEN 127.0.0.1:32123 /dev/ttyACM0` to the monitor tool
1. the monitor tool opens `/dev/ttyACM0`
1. the monitor tool connects via TCP/IP to `127.0.0.1:32123` and start streaming data back and forth

The answer to the `OPEN` command is:

```JSON
{
  "event": "open",
  "message": "ok"
}
```

If the monitor tool cannot communicate with the board, or if the tool can not connect back to the TCP port, or if any
other error condition happens:

```JSON
{
  "event": "open",
  "error": true,
  "message": "unknown port /dev/ttyACM23"
}
```

The board port will be opened using the parameters previously set through the `CONFIGURE` command.

Once the port is opened, it may be unexpectedly closed at any time due to hardware failure, or because the Client/IDE
closes the TCP/IP connection, etc. In this case an asynchronous `port_closed` message must be generated from the monitor
tool:

```JSON
{
  "event": "port_closed",
  "message": "serial port disappeared!"
}
```

or

```JSON
{
  "event": "port_closed",
  "message": "lost TCP/IP connection with the client!"
}
```

#### CLOSE command

The `CLOSE` command will close the currently opened port and close the TCP/IP connection used to communicate with the
Client/IDE. The answer to the command is:

```JSON
{
  "event": "close",
  "message": "ok"
}
```

or in case of error

```JSON
{
  "event": "close",
  "error": true,
  "message": "port already closed"
}
```

#### QUIT command

The `QUIT` command terminates the monitor. The response to `QUIT` is:

```JSON
{
  "eventType": "quit",
  "message": "OK"
}
```

after this output the monitor exits. This command is supposed to always succeed.

#### Invalid commands

If the client sends an invalid or malformed command, the monitor should answer with:

```JSON
{
  "eventType": "command_error",
  "error": true,
  "message": "Unknown command XXXX"
}
```
