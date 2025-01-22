## Arduino Nano/Uno/Mega is not detected

When you run [`arduino-cli board list`][arduino cli board list], your board doesn't show up. Possible causes:

- Your board is a cheaper derivative, or
- It's a board, such the classic Nano, that uses a USB to serial converter like FTDI FT232 or CH340. These chips always
  report the same USB VID/PID to the operating system, so the only thing we know is that the board mounts that specific
  USB2Serial chip, but we don’t know which board that chip is on.

## What's the FQBN string?

FQBN stands for Fully Qualified Board Name. It has the following format:
`VENDOR:ARCHITECTURE:BOARD_ID[:MENU_ID=OPTION_ID[,MENU2_ID=OPTION_ID ...]]`, with each `MENU_ID=OPTION_ID` being an
optional key-value pair configuration. Each field accepts letters (`A-Z` or `a-z`), numbers (`0-9`), underscores (`_`),
dashes(`-`) and dots(`.`). The special character `=` is accepted in the configuration value. The `VENDOR` and
`ARCHITECTURE` parts can be empty. For a deeper understanding of how FQBN works, you should understand the [Arduino
platform specification][0].

## How to set multiple board options?

Additional board options have to be separated by commas (instead of colon):

`$ arduino-cli compile --fqbn "esp8266:esp8266:generic:xtal=160,baud=57600" TestSketch`

## Where is the Serial Monitor?

The serial monitor is available through the [monitor command][monitor command]. By the way, the functionality provided
by this command is very limited and you may want to look for other tools if you need more advanced functionality.

There are many excellent serial terminals to chose from. On Linux or macOS, you may already have `screen` installed. On
Windows, a good choice for command line usage is Plink, included with [PuTTY][putty].

## How to change monitor configuration?

[Configuration parameters][configuration parameters] of the monitor can be obtained by executing the following command:

`$ arduino-cli monitor -p <port> --describe`

These parameters can be modified by passing a list of `<key>=<desiredValue>` pairs to the `--config` flag. For example,
when using a serial port, the monitor baud rate can be set to 4800 with the following command:

`$ arduino-cli monitor -p <port> --config baudrate=4800`

## "Permission denied" error in sketch upload

This problem might happen on some Linux systems, and can be solved by setting up serial port permissions. First, search
for the port your board is connected to, with the command:

`$ arduino-cli board list`

Then add your user to the group with the following command, replacing `<username>` with your username and `<group>` with
your group name. Logging out and in again is necessary for the changes to take effect.

`$ sudo usermod -a -G <group> <username>`

## Additional assistance

If your question wasn't answered, feel free to ask on [Arduino CLI's forum board][1].

[arduino cli board list]: commands/arduino-cli_board_list.md
[0]: platform-specification.md
[1]: https://forum.arduino.cc/c/software/arduino-cli/89
[putty]: https://www.chiark.greenend.org.uk/~sgtatham/putty/
[monitor command]: commands/arduino-cli_monitor.md
[configuration parameters]: pluggable-monitor-specification.md#describe-command
