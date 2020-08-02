## Arduino Uno/Mega/Duemilanove is not detected

When you run `arduino-cli board list`, your board doesn't show up. Possible causes:

- Your board is a cheaper clone, or
- It mounts a USB2Serial converter like FT232 or CH340: these chips always report the same USB VID/PID to the operating
  system, so the only thing we know is that the board mounts that specific USB2Serial chip, but we don’t know which
  board that chip is on.

## What's the FQBN string?

For a deeper understanding of how FQBN works, you should understand the [Arduino platform specification][0].

## How to set multiple board options?

Additional board options have to be separated by commas (instead of colon):

`$ arduino-cli compile --fqbn "esp8266:esp8266:generic:xtal=160,baud=57600" TestSketch`

## Where is the Serial Monitor?

Being this a command line tool we believe it's up to the user to choose their preferred way of interacting with the
serial connection. If we were to integrate it into the CLI we'd end up putting a tool inside a tool, and this would be
something that we're trying to avoid.

There are many excellent serial terminals to chose from. On Linux or macOS, you may already have [screen][screen]
installed. On Windows, a good choice for command line usage is Plink, included with [PuTTY][putty].

## Additional assistance

If your question wasn't answered, feel free to ask on [Arduino CLI's forum board][1].

[0]: platform-specification.md
[1]: https://forum.arduino.cc/index.php?board=145.0
[screen]: https://www.gnu.org/software/screen/manual/screen.html
[putty]: https://www.chiark.greenend.org.uk/~sgtatham/putty/
