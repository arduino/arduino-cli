## Arduino Uno/Mega/Duemilanove is not detected

When you run `arduino-cli board list`, your board doesn't show up. Possible causes:

-  Your board is a cheaper clone, or
-  It mounts a USB2Serial converter like FT232 or CH340: these chips
   always report the same USB VID/PID to the operating system, so the
   only thing we know is that the board mounts that specific
   USB2Serial chip, but we don’t know which board that chip is on.

##  What's the FQBN string?

For a deeper understanding of how FQBN works, you should understand
the [Arduino platform specification][0].

##  How to set multiple board options?

Additional board options have to be separated by commas (instead of colon):

`$ arduino-cli compile -b "LGT8fx Boards:avr:328:clock=16MHz,variant=modelP" test-sketch`


[0]: platform-specification.md
