# Upgrading

Here you can find a list of migration guides to handle breaking changes between releases of the CLI.

## 0.14.0

### Changes in `debug` command

Previously it was required:

- To provide a debug command line recipe in `platform.txt` like `tools.reciped-id.debug.pattern=.....` that will start a
  `gdb` session for the selected board.
- To add a `debug.tool` definition in the `boards.txt` to recall that recipe, for example `myboard.debug.tool=recipe-id`

Now:

- Only the configuration needs to be supplied, the `arduino-cli` or the GUI tool will figure out how to call and setup
  the `gdb` session. An example of configuration is the following:

```
debug.executable={build.path}/{build.project_name}.elf
debug.toolchain=gcc
debug.toolchain.path={runtime.tools.arm-none-eabi-gcc-7-2017q4.path}/bin/
debug.toolchain.prefix=arm-none-eabi-
debug.server=openocd
debug.server.openocd.path={runtime.tools.openocd-0.10.0-arduino7.path}/bin/
debug.server.openocd.scripts_dir={runtime.tools.openocd-0.10.0-arduino7.path}/share/openocd/scripts/
debug.server.openocd.script={runtime.platform.path}/variants/{build.variant}/{build.openocdscript}
```

The `debug.server.XXXX` subkeys are optional and also "free text", this means that the configuration may be extended as
needed by the specific server. For now only `openocd` is supported. Anyway, if this change works, any other kind of
server may be fairly easily added.

The `debug.xxx=yyy` definitions above may be supplied and overlayed in the usual ways:

- on `platform.txt`: definition here will be shared through all boards in the platform
- on `boards.txt` as part of a board definition: they will override the global platform definitions
- on `programmers.txt`: they will override the boards and global platform definitions if the programmer is selected
