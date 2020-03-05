## Arduino Uno/Mega/Duemilanove is not detected

When you run `arduino-cli board list`, your board doesn't show up. Possible causes:

-  Your board is a cheaper clone, or
-  It mounts a USB2Serial converter like FT232 or CH320: these chips
   always reports the same USB VID/PID to the operating system, so the
   only thing that we know is that the board mounts that specific
   USB2Serial chip, but we don’t know which board is.

##  What's the FQBN string?

For a deeper understanding of how FQBN works, you should understand
Arduino Hardware specification. You can find more information in this
[arduino/Arduino wiki page][0].


[0]: https://github.com/arduino/Arduino/wiki/Arduino-IDE-1.5-3rd-party-Hardware-specification
